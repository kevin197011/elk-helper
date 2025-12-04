// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// LarkClient handles Lark webhook notifications
type LarkClient struct {
	webhookURL string
	httpClient *http.Client
}

// NewLarkClient creates a new Lark client
func NewLarkClient(webhookURL string) *LarkClient {
	return &LarkClient{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendAlert sends alert message with logs to Lark
func (lc *LarkClient) SendAlert(ruleName, indexName string, logs []map[string]interface{}, fromTime, toTime time.Time, retryTimes int) error {
	if len(logs) == 0 {
		return nil
	}

	message := lc.buildMessage(ruleName, indexName, logs, fromTime, toTime)

	for attempt := 1; attempt <= retryTimes; attempt++ {
		body, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		req, err := http.NewRequest("POST", lc.webhookURL, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := lc.httpClient.Do(req)
		if err != nil {
			if attempt < retryTimes {
				waitTime := time.Duration(1<<uint(attempt)) * time.Second
				time.Sleep(waitTime)
				continue
			}
			return fmt.Errorf("failed to send to Lark after %d attempts: %w", retryTimes, err)
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var result map[string]interface{}
		if err := json.Unmarshal(respBody, &result); err != nil {
			if attempt < retryTimes {
				waitTime := time.Duration(1<<uint(attempt)) * time.Second
				time.Sleep(waitTime)
				continue
			}
			return fmt.Errorf("failed to parse Lark response: %w", err)
		}

		if resp.StatusCode == http.StatusOK {
			if code, ok := result["code"].(float64); ok && code == 0 {
				return nil
			}
		}

		if attempt < retryTimes {
			waitTime := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(waitTime)
		} else {
			return fmt.Errorf("lark API error: %v", result)
		}
	}

	return fmt.Errorf("failed to send to Lark after %d attempts", retryTimes)
}

func (lc *LarkClient) buildMessage(ruleName, indexName string, logs []map[string]interface{}, fromTime, toTime time.Time) map[string]interface{} {
	elements := []map[string]interface{}{
		{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("**📋 规则名称**\n%s", ruleName),
			},
		},
		{
			"tag": "div",
			"fields": []map[string]interface{}{
				{
					"is_short": true,
					"text": map[string]interface{}{
						"tag":     "lark_md",
						"content": fmt.Sprintf("**⏰ 时间范围**\n%s\n%s", formatTime(fromTime), formatTime(toTime)),
					},
				},
				{
					"is_short": true,
					"text": map[string]interface{}{
						"tag":     "lark_md",
						"content": fmt.Sprintf("**🔔 告警数量**\n%d 条", len(logs)),
					},
				},
			},
		},
		{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("**📊 索引名称**\n`%s`", indexName),
			},
		},
		{
			"tag": "hr",
		},
	}

	// Show summary of logs in table format (max 3 samples)
	if len(logs) > 0 {
		elements = append(elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": "**📝 日志摘要**",
			},
		})

		// Show up to 3 log samples with key fields only
		displayCount := len(logs)
		if displayCount > 3 {
			displayCount = 3
		}

		// Build table content
		tableContent := "| # | 状态码 | 时间 | URL | CF Ray |\n| --- | --- | --- | --- | --- |"

		for i := 0; i < displayCount; i++ {
			log := logs[i]
			row := lc.extractTableRow(i+1, log)
			tableContent += "\n" + row
		}

		elements = append(elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": tableContent,
			},
		})

		// If there are more logs, show count
		if len(logs) > 3 {
			elements = append(elements, map[string]interface{}{
				"tag": "div",
				"text": map[string]interface{}{
					"tag":     "lark_md",
					"content": fmt.Sprintf("**... 还有 %d 条日志 ...**", len(logs)-3),
				},
			})
		}
	}

	// Add note and @all
	elements = append(elements, map[string]interface{}{
		"tag": "hr",
	})
	elements = append(elements, map[string]interface{}{
		"tag": "note",
		"elements": []map[string]interface{}{
			{
				"tag":     "plain_text",
				"content": "💡 完整日志详情请登录 ELK Helper 系统查看",
			},
		},
	})
	elements = append(elements, map[string]interface{}{
		"tag": "div",
		"text": map[string]interface{}{
			"tag":     "lark_md",
			"content": "<at id=all></at>",
		},
	})

	return map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"config": map[string]interface{}{
				"wide_screen_mode": true,
			},
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"tag":     "plain_text",
					"content": "🚨 ELK 告警",
				},
				"template": "red",
			},
			"elements": elements,
		},
	}
}

// extractTableRow extracts key fields from a log entry and formats as table row
// Shows: response_code, @timestamp, request URL (without params), cf_ray
func (lc *LarkClient) extractTableRow(rowNum int, log map[string]interface{}) string {
	// 1. Response Code (匹配字段的错误码)
	responseCode := "-"
	if val, ok := log["response_code"]; ok && val != nil {
		responseCode = fmt.Sprintf("**%v**", val)
	} else if val, ok := log["status_code"]; ok && val != nil {
		responseCode = fmt.Sprintf("**%v**", val)
	} else if val, ok := log["status"]; ok && val != nil {
		responseCode = fmt.Sprintf("**%v**", val)
	}

	// 2. Timestamp
	timestamp := "-"
	if val, ok := log["@timestamp"]; ok && val != nil {
		timestampStr := fmt.Sprintf("%v", val)
		// If it's ISO format, try to format it nicely
		if strings.Contains(timestampStr, "T") {
			timestampStr = strings.Replace(timestampStr, "T", " ", 1)
			timestampStr = strings.Replace(timestampStr, "Z", "", 1)
			// Truncate milliseconds if present
			if idx := strings.Index(timestampStr, "."); idx > 0 {
				timestampStr = timestampStr[:idx]
			}
		}
		timestamp = timestampStr
	}

	// 3. Request URL (without query parameters)
	requestURL := "-"
	if val, ok := log["request"]; ok && val != nil && val != "" {
		requestStr := fmt.Sprintf("%v", val)
		// Remove query parameters (everything after ?)
		if idx := strings.Index(requestStr, "?"); idx > 0 {
			requestStr = requestStr[:idx]
		}
		// Truncate if still too long
		if len(requestStr) > 60 {
			requestStr = requestStr[:60] + "..."
		}
		requestURL = requestStr
	} else if val, ok := log["path"]; ok && val != nil && val != "" {
		pathStr := fmt.Sprintf("%v", val)
		if idx := strings.Index(pathStr, "?"); idx > 0 {
			pathStr = pathStr[:idx]
		}
		if len(pathStr) > 60 {
			pathStr = pathStr[:60] + "..."
		}
		requestURL = pathStr
	}

	// 4. CF Ray ID
	cfRay := "-"
	if val, ok := log["cf_ray"]; ok && val != nil && val != "" {
		cfRay = fmt.Sprintf("%v", val)
	}

	// Build table row: | # | 状态码 | 时间 | URL | CF Ray |
	return fmt.Sprintf("| %d | %s | %s | %s | %s |", rowNum, responseCode, timestamp, requestURL, cfRay)
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
