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
	"log/slog"
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
		slog.Warn("SendAlert called with empty logs", "rule_name", ruleName)
		return nil
	}

	slog.Info("Sending alert to Lark", "rule_name", ruleName, "index_name", indexName, "log_count", len(logs), "webhook_url", lc.webhookURL, "retry_times", retryTimes)
	message := lc.buildMessage(ruleName, indexName, logs, fromTime, toTime)

	for attempt := 1; attempt <= retryTimes; attempt++ {
		slog.Debug("Lark send attempt", "rule_name", ruleName, "attempt", attempt, "max_attempts", retryTimes)
		body, err := json.Marshal(message)
		if err != nil {
			slog.Error("Failed to marshal message", "rule_name", ruleName, "error", err)
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		req, err := http.NewRequest("POST", lc.webhookURL, bytes.NewReader(body))
		if err != nil {
			slog.Error("Failed to create request", "rule_name", ruleName, "error", err)
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := lc.httpClient.Do(req)
		if err != nil {
			slog.Warn("Lark request failed", "rule_name", ruleName, "attempt", attempt, "error", err)
			if attempt < retryTimes {
				waitTime := time.Duration(1<<uint(attempt)) * time.Second
				time.Sleep(waitTime)
				continue
			}
			slog.Error("Failed to send to Lark after all attempts", "rule_name", ruleName, "attempts", retryTimes, "error", err)
			return fmt.Errorf("failed to send to Lark after %d attempts: %w", retryTimes, err)
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var result map[string]interface{}
		if err := json.Unmarshal(respBody, &result); err != nil {
			slog.Warn("Failed to parse Lark response", "rule_name", ruleName, "attempt", attempt, "error", err, "response_body", string(respBody))
			if attempt < retryTimes {
				waitTime := time.Duration(1<<uint(attempt)) * time.Second
				time.Sleep(waitTime)
				continue
			}
			slog.Error("Failed to parse Lark response after all attempts", "rule_name", ruleName, "error", err)
			return fmt.Errorf("failed to parse Lark response: %w", err)
		}

		if resp.StatusCode == http.StatusOK {
			if code, ok := result["code"].(float64); ok && code == 0 {
				slog.Info("Alert sent successfully to Lark", "rule_name", ruleName, "attempt", attempt)
				return nil
			}
			slog.Warn("Lark API returned non-zero code", "rule_name", ruleName, "attempt", attempt, "code", result["code"], "response", result)
		} else {
			slog.Warn("Lark API returned non-200 status", "rule_name", ruleName, "attempt", attempt, "status_code", resp.StatusCode, "response", result)
		}

		if attempt < retryTimes {
			waitTime := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(waitTime)
		} else {
			slog.Error("Lark API error after all attempts", "rule_name", ruleName, "response", result)
			return fmt.Errorf("lark API error: %v", result)
		}
	}

	slog.Error("Failed to send to Lark after all attempts", "rule_name", ruleName, "attempts", retryTimes)
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

	// Show summary of logs in card format (max 3 samples)
	if len(logs) > 0 {
		elements = append(elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": "**📝 日志摘要**（共 " + fmt.Sprintf("%d", len(logs)) + " 条，展示前 3 条）",
			},
		})

		// Show up to 3 log samples with key fields only
		displayCount := len(logs)
		if displayCount > 3 {
			displayCount = 3
		}

		// Build each log entry as a separate card section
		for i := 0; i < displayCount; i++ {
			log := logs[i]
			logFields := lc.extractLogFields(i+1, log, ruleName)

			// Add a separator before each log entry (except the first one)
			if i > 0 {
				elements = append(elements, map[string]interface{}{
					"tag": "hr",
				})
			}

			// Add log entry with fields layout
			elements = append(elements, map[string]interface{}{
				"tag":    "div",
				"fields": logFields,
			})
		}

		// If there are more logs, show count
		if len(logs) > 3 {
			elements = append(elements, map[string]interface{}{
				"tag": "hr",
			})
			elements = append(elements, map[string]interface{}{
				"tag": "div",
				"text": map[string]interface{}{
					"tag":     "lark_md",
					"content": fmt.Sprintf("**➕ 还有 %d 条日志未显示**\n💡 查看完整日志请登录系统", len(logs)-3),
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

// extractLogFields extracts key fields from a log entry and formats as card fields
// Uses rule name to determine the log type and shows relevant fields:
// - Rule name contains "nginx": response_code, @timestamp, request, cf_ray, domain
// - Rule name contains "java", "go", "c++", "python", "nodejs", etc.: module, node_ip, message, @timestamp
func (lc *LarkClient) extractLogFields(rowNum int, log map[string]interface{}, ruleName string) []map[string]interface{} {
	// Detect log type from rule name (case insensitive)
	ruleNameLower := strings.ToLower(ruleName)

	// Check if rule name contains "nginx"
	if strings.Contains(ruleNameLower, "nginx") {
		return lc.extractNginxLogFields(rowNum, log)
	}

	// Check if rule name contains application log types (java, go, c++, python, nodejs, app, etc.)
	appLogTypes := []string{"java", "go", "c++", "cpp", "python", "nodejs", "node", "app", "application", "service", "api", "web"}
	for _, appType := range appLogTypes {
		if strings.Contains(ruleNameLower, appType) {
			return lc.extractAppLogFields(rowNum, log)
		}
	}

	// Fallback: try to detect from log fields
	if _, hasResponseCode := log["response_code"]; hasResponseCode {
		return lc.extractNginxLogFields(rowNum, log)
	}
	if _, hasModule := log["module"]; hasModule {
		if _, hasMessage := log["message"]; hasMessage {
			return lc.extractAppLogFields(rowNum, log)
		}
	}

	// Default fallback to app log format (more generic)
	return lc.extractAppLogFields(rowNum, log)
}

// extractNginxLogFields extracts fields for nginx/nginx-access logs
// Shows: response_code, @timestamp, request, cf_ray, domain
func (lc *LarkClient) extractNginxLogFields(rowNum int, log map[string]interface{}) []map[string]interface{} {
	fields := []map[string]interface{}{}

	// 1. Response Code - highlighted
	responseCode := "-"
	if val, ok := log["response_code"]; ok && val != nil {
		responseCode = fmt.Sprintf("%v", val)
	} else if val, ok := log["status_code"]; ok && val != nil {
		responseCode = fmt.Sprintf("%v", val)
	} else if val, ok := log["status"]; ok && val != nil {
		responseCode = fmt.Sprintf("%v", val)
	}

	// 2. Timestamp
	timestamp := lc.formatTimestamp(log)

	// 3. Request URL (without query parameters)
	requestURL := "-"
	if val, ok := log["request"]; ok && val != nil && val != "" {
		requestStr := fmt.Sprintf("%v", val)
		if idx := strings.Index(requestStr, "?"); idx > 0 {
			requestStr = requestStr[:idx]
		}
		if len(requestStr) > 50 {
			requestStr = requestStr[:50] + "..."
		}
		requestURL = requestStr
	} else if val, ok := log["path"]; ok && val != nil && val != "" {
		pathStr := fmt.Sprintf("%v", val)
		if idx := strings.Index(pathStr, "?"); idx > 0 {
			pathStr = pathStr[:idx]
		}
		if len(pathStr) > 50 {
			pathStr = pathStr[:50] + "..."
		}
		requestURL = pathStr
	}

	// 4. CF Ray ID
	cfRay := "-"
	if val, ok := log["cf_ray"]; ok && val != nil && val != "" {
		cfRay = fmt.Sprintf("%v", val)
	}

	// 5. Domain
	domain := "-"
	if val, ok := log["domain"]; ok && val != nil && val != "" {
		domain = fmt.Sprintf("%v", val)
	}

	// Build fields
	fields = append(fields, map[string]interface{}{
		"is_short": true,
		"text": map[string]interface{}{
			"tag":     "lark_md",
			"content": fmt.Sprintf("**#%d | 状态码:** <font color='red'>%s</font>", rowNum, responseCode),
		},
	})

	fields = append(fields, map[string]interface{}{
		"is_short": true,
		"text": map[string]interface{}{
			"tag":     "lark_md",
			"content": fmt.Sprintf("**⏰ 时间:** %s", timestamp),
		},
	})

	fields = append(fields, map[string]interface{}{
		"is_short": true,
		"text": map[string]interface{}{
			"tag":     "lark_md",
			"content": fmt.Sprintf("**🔗 URL:** `%s`", requestURL),
		},
	})

	fields = append(fields, map[string]interface{}{
		"is_short": true,
		"text": map[string]interface{}{
			"tag":     "lark_md",
			"content": fmt.Sprintf("**☁️ CF Ray:** `%s`", cfRay),
		},
	})

	fields = append(fields, map[string]interface{}{
		"is_short": true,
		"text": map[string]interface{}{
			"tag":     "lark_md",
			"content": fmt.Sprintf("**🌐 Domain:** `%s`", domain),
		},
	})

	return fields
}

// extractAppLogFields extracts fields for application logs (java, go, c++, python, nodejs, etc.)
// Shows: module, node_ip, message, @timestamp
func (lc *LarkClient) extractAppLogFields(rowNum int, log map[string]interface{}) []map[string]interface{} {
	fields := []map[string]interface{}{}

	// 1. Module
	module := "-"
	if val, ok := log["module"]; ok && val != nil && val != "" {
		module = fmt.Sprintf("%v", val)
	}

	// 2. Node IP
	nodeIP := "-"
	if val, ok := log["node_ip"]; ok && val != nil && val != "" {
		nodeIP = fmt.Sprintf("%v", val)
	}

	// 3. Message (truncate if too long)
	message := "-"
	if val, ok := log["message"]; ok && val != nil && val != "" {
		messageStr := fmt.Sprintf("%v", val)
		// Truncate to first 200 chars for display
		if len(messageStr) > 200 {
			messageStr = messageStr[:200] + "..."
		}
		// Replace newlines with spaces for compact display
		messageStr = strings.ReplaceAll(messageStr, "\n", " ")
		messageStr = strings.ReplaceAll(messageStr, "\r", "")
		message = messageStr
	}

	// 4. Timestamp
	timestamp := lc.formatTimestamp(log)

	// Build fields - Application log format
	fields = append(fields, map[string]interface{}{
		"is_short": true,
		"text": map[string]interface{}{
			"tag":     "lark_md",
			"content": fmt.Sprintf("**#%d | 📦 模块:** `%s`", rowNum, module),
		},
	})

	fields = append(fields, map[string]interface{}{
		"is_short": true,
		"text": map[string]interface{}{
			"tag":     "lark_md",
			"content": fmt.Sprintf("**🖥️ 节点:** `%s`", nodeIP),
		},
	})

	fields = append(fields, map[string]interface{}{
		"is_short": true,
		"text": map[string]interface{}{
			"tag":     "lark_md",
			"content": fmt.Sprintf("**⏰ 时间:** %s", timestamp),
		},
	})

	// Message as full-width field (not short)
	fields = append(fields, map[string]interface{}{
		"is_short": false,
		"text": map[string]interface{}{
			"tag":     "lark_md",
			"content": fmt.Sprintf("**💬 消息:**\n```\n%s\n```", message),
		},
	})

	return fields
}

// formatTimestamp formats the @timestamp field
func (lc *LarkClient) formatTimestamp(log map[string]interface{}) string {
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
	return timestamp
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
