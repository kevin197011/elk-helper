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
			return fmt.Errorf("Lark API error: %v", result)
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

	// Show summary of logs (only key fields, max 5 samples)
	if len(logs) > 0 {
		summaryContent := "**📝 日志摘要**\n"

		// Show up to 5 log samples with key fields only
		displayCount := len(logs)
		if displayCount > 5 {
			displayCount = 5
		}

		for i := 0; i < displayCount; i++ {
			log := logs[i]
			summary := lc.extractKeyFields(log)
			summaryContent += fmt.Sprintf("\n**#%d** %s", i+1, summary)
		}

		// If there are more logs, show count
		if len(logs) > 5 {
			summaryContent += fmt.Sprintf("\n\n**... 还有 %d 条日志 ...**", len(logs)-5)
		}

		elements = append(elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": summaryContent,
			},
		})
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

// extractKeyFields extracts key fields from a log entry for summary display
// Intelligently adapts to different log types (nginx, java, go, application logs, etc.)
func (lc *LarkClient) extractKeyFields(log map[string]interface{}) string {
	var parts []string

	// Detect log type
	logType, _ := log["log_type"].(string)

	// Define key fields based on log type
	var keyFields []string

	switch logType {
	case "nginx", "nginx-access":
		// Nginx access logs: focus on HTTP fields
		keyFields = []string{
			"response_code", "status_code", "status",
			"ip", "client_ip", "remote_addr",
			"request", "path", "url",
			"request_method", "method",
			"upstreamaddr", "upstream",
		}
	case "java", "go", "golang", "cpp", "c++", "python", "nodejs", "application":
		// Application logs: focus on message, level, module
		keyFields = []string{
			"level", "severity", "log_level",
			"module", "service", "app",
			"hostname", "host", "node_ip",
			"message", "msg",
		}
	default:
		// Generic logs: try common fields
		keyFields = []string{
			"level", "severity",
			"status", "status_code", "response_code",
			"message", "msg", "error", "error_message",
			"module", "service",
			"ip", "client_ip", "hostname", "host",
			"request", "path",
			"method", "request_method",
		}
	}

	// Extract fields with priority
	fieldCount := 0
	maxFields := 4

	for _, key := range keyFields {
		if fieldCount >= maxFields {
			break
		}
		if val, ok := log[key]; ok && val != nil && val != "" {
			// Special handling for message field (truncate if too long)
			if key == "message" || key == "msg" {
				valStr := fmt.Sprintf("%v", val)
				if len(valStr) > 100 {
					valStr = valStr[:100] + "..."
				}
				parts = append(parts, fmt.Sprintf("`%s`: %s", key, valStr))
			} else {
				parts = append(parts, fmt.Sprintf("`%s`: %v", key, val))
			}
			fieldCount++
		}
	}

	// If no key fields found, show timestamp and log_type
	if len(parts) == 0 {
		if lt, ok := log["log_type"]; ok {
			parts = append(parts, fmt.Sprintf("`log_type`: %v", lt))
		}
		if ts, ok := log["@timestamp"]; ok {
			parts = append(parts, fmt.Sprintf("`@timestamp`: %v", ts))
		}
		if len(parts) == 0 {
			parts = append(parts, "无关键字段")
		}
	}

	// Join with " | "
	summary := ""
	for i, part := range parts {
		if i > 0 {
			summary += " | "
		}
		summary += part
	}

	return summary
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
