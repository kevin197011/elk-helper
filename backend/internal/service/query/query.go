// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package query

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/kk/elk-helper/backend/internal/config"
	"github.com/kk/elk-helper/backend/internal/models"
)

const (
	maxScrollResults = 10000
	scrollTimeout    = 1 * time.Minute
)

// Service provides Elasticsearch query operations
type Service struct {
	client *elasticsearch.Client
}

// NewService creates a new query service using environment variables (backward compatibility)
func NewService() (*Service, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{config.AppConfig.ES.URL},
		// Connection pool settings for high concurrency (1000+ rules)
		MaxRetries: 3,
	}

	if config.AppConfig.ES.Username != "" && config.AppConfig.ES.Password != "" {
		cfg.Username = config.AppConfig.ES.Username
		cfg.Password = config.AppConfig.ES.Password
	}

	// Configure HTTP transport for high concurrency
	transport := &http.Transport{
		MaxIdleConns:        200, // Maximum idle connections
		MaxIdleConnsPerHost: 100, // Maximum idle connections per host
		MaxConnsPerHost:     200, // Maximum connections per host
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	}

	// Configure SSL/TLS for env-based ES connection
	if config.AppConfig.ES.UseSSL || strings.HasPrefix(config.AppConfig.ES.URL, "https://") {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: config.AppConfig.ES.SkipVerify,
		}

		// Add CA certificate if provided
		if config.AppConfig.ES.CACertificate != "" {
			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM([]byte(config.AppConfig.ES.CACertificate)) {
				return nil, fmt.Errorf("failed to parse ES_CA_CERTIFICATE")
			}
			tlsConfig.RootCAs = caCertPool
		}

		transport.TLSClientConfig = tlsConfig
	}
	cfg.Transport = transport

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create ES client: %w", err)
	}

	return &Service{client: client}, nil
}

// NewServiceFromConfig creates a new query service from ESConfig
func NewServiceFromConfig(esConfig *models.ESConfig) (*Service, error) {
	if esConfig == nil {
		return nil, fmt.Errorf("ES config is nil")
	}

	if !esConfig.Enabled {
		return nil, fmt.Errorf("ES config is disabled")
	}

	// Parse multiple URLs separated by semicolons
	// Example: "https://10.170.1.54:9200;https://10.170.1.55:9200;https://10.170.1.56:9200"
	addresses := parseESAddresses(esConfig.URL)
	if len(addresses) == 0 {
		return nil, fmt.Errorf("no valid ES addresses found in URL: %s", esConfig.URL)
	}

	cfg := elasticsearch.Config{
		Addresses: addresses,
		// Connection pool settings for high concurrency (1000+ rules)
		MaxRetries: 3,
	}

	// Configure authentication
	// Note: When ES security is enabled, username and password are required
	if esConfig.Username != "" && esConfig.Password != "" {
		cfg.Username = esConfig.Username
		cfg.Password = esConfig.Password
	} else {
		// Warn if authentication is likely required but not provided
		// This is a best-effort check; actual authentication requirement depends on ES configuration
	}

	// Configure HTTP transport for high concurrency
	transport := &http.Transport{
		MaxIdleConns:        200, // Maximum idle connections
		MaxIdleConnsPerHost: 100, // Maximum idle connections per host
		MaxConnsPerHost:     200, // Maximum connections per host
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	}

	// Configure SSL/TLS
	if esConfig.UseSSL {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: esConfig.SkipVerify,
		}

		// Add CA certificate if provided
		if esConfig.CACertificate != "" {
			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM([]byte(esConfig.CACertificate)) {
				return nil, fmt.Errorf("failed to parse CA certificate")
			}
			tlsConfig.RootCAs = caCertPool
		}

		transport.TLSClientConfig = tlsConfig
	}

	cfg.Transport = transport

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create ES client from config: %w", err)
	}

	return &Service{client: client}, nil
}

// parseESAddresses parses semicolon-separated ES addresses
// Example: "https://10.170.1.54:9200;https://10.170.1.55:9200" -> []string{"https://10.170.1.54:9200", "https://10.170.1.55:9200"}
func parseESAddresses(urlString string) []string {
	if urlString == "" {
		return nil
	}

	// Split by semicolon
	parts := strings.Split(urlString, ";")
	var addresses []string

	for _, part := range parts {
		// Trim whitespace
		addr := strings.TrimSpace(part)
		if addr != "" {
			addresses = append(addresses, addr)
		}
	}

	return addresses
}

// QueryLogs queries logs based on rule configuration
func (s *Service) QueryLogs(ctx context.Context, rule *models.Rule, fromTime, toTime time.Time, batchSize int) ([]map[string]interface{}, error) {
	// Guardrail: enforce a maximum ES query duration to avoid occupying worker slots forever.
	// If caller already has a deadline, keep the earlier one.
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		timeout := 30 * time.Second
		if config.AppConfig != nil && config.AppConfig.ES.QueryTimeoutSeconds > 0 {
			timeout = time.Duration(config.AppConfig.ES.QueryTimeoutSeconds) * time.Second
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	query := s.buildQuery(rule.Queries, fromTime, toTime)

	queryJSON, _ := json.MarshalIndent(query, "", "  ")
	slog.Debug("Elasticsearch query", "query", string(queryJSON))

	var results []map[string]interface{}
	scrollID := ""

	// Initial search with scroll
	searchBody, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := esapi.SearchRequest{
		Index:  []string{rule.IndexPattern},
		Body:   bytes.NewReader(searchBody),
		Scroll: scrollTimeout,
		Size:   &batchSize,
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, fmt.Errorf("ES search request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, fmt.Errorf("error parsing error response: %w", err)
		}
		return nil, fmt.Errorf("ES search error: %v", e)
	}

	var searchResp map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	scrollID, _ = searchResp["_scroll_id"].(string)
	initialDocs := s.extractDocuments(searchResp)
	results = append(results, initialDocs...)
	slog.Info("Initial search completed", "index_pattern", rule.IndexPattern, "initial_docs", len(initialDocs), "total_docs", len(results))

	// Continue scrolling
	for scrollID != "" && len(results) < maxScrollResults {
		scrollReq := esapi.ScrollRequest{
			ScrollID: scrollID,
			Scroll:   scrollTimeout,
		}

		scrollRes, err := scrollReq.Do(ctx, s.client)
		if err != nil {
			break
		}

		if scrollRes.IsError() {
			break
		}

		var scrollResp map[string]interface{}
		if err := json.NewDecoder(scrollRes.Body).Decode(&scrollResp); err != nil {
			scrollRes.Body.Close()
			break
		}

		hits, _ := scrollResp["hits"].(map[string]interface{})
		hitsList, _ := hits["hits"].([]interface{})
		if len(hitsList) == 0 {
			scrollRes.Body.Close()
			break
		}

		scrollDocs := s.extractDocuments(scrollResp)
		results = append(results, scrollDocs...)
		scrollID, _ = scrollResp["_scroll_id"].(string)
		scrollRes.Body.Close()
		slog.Debug("Scroll batch completed", "scroll_docs", len(scrollDocs), "total_docs", len(results))
	}

	// Clear scroll context
	if scrollID != "" {
		clearReq := esapi.ClearScrollRequest{
			ScrollID: []string{scrollID},
		}
		_, _ = clearReq.Do(ctx, s.client)
	}

	slog.Info("Query completed", "index_pattern", rule.IndexPattern, "total_results", len(results))
	return results, nil
}

// TestConnection tests ES connection
func (s *Service) TestConnection(ctx context.Context) error {
	res, err := s.client.Ping(s.client.Ping.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ping returned error: %s", res.String())
	}

	return nil
}

// buildQuery builds ES query from rule queries
func (s *Service) buildQuery(queries models.QueryConditions, fromTime, toTime time.Time) map[string]interface{} {
	var mustClauses []map[string]interface{}

	// Time range
	mustClauses = append(mustClauses, map[string]interface{}{
		"range": map[string]interface{}{
			"@timestamp": map[string]interface{}{
				"gte":    fromTime.UTC().Format(time.RFC3339),
				"lt":     toTime.UTC().Format(time.RFC3339),
				"format": "strict_date_optional_time",
			},
		},
	})

	// Process queries
	queryClauses := s.buildFlexibleQueries(queries)
	mustClauses = append(mustClauses, queryClauses...)

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustClauses,
			},
		},
		"sort": []map[string]interface{}{
			{
				"@timestamp": map[string]interface{}{
					"order": "asc",
				},
			},
		},
	}
}

func (s *Service) buildFlexibleQueries(queries models.QueryConditions) []map[string]interface{} {
	var andClauses []map[string]interface{}
	var orClauses []map[string]interface{}

	for _, q := range queries {
		clause := s.buildSingleQuery(q)
		if clause == nil {
			continue
		}

		logic := q.Logic
		if logic == "" {
			logic = "or"
		}

		switch logic {
		case "and":
			andClauses = append(andClauses, clause)
		case "or":
			orClauses = append(orClauses, clause)
		default:
			orClauses = append(orClauses, clause)
		}
	}

	var result []map[string]interface{}
	result = append(result, andClauses...)

	if len(orClauses) > 0 {
		result = append(result, map[string]interface{}{
			"bool": map[string]interface{}{
				"should":               orClauses,
				"minimum_should_match": 1,
			},
		})
	}

	return result
}

func (s *Service) buildSingleQuery(q models.QueryCondition) map[string]interface{} {
	// Support both "operator" and "op" fields for compatibility
	operator := q.Operator
	if operator == "" {
		operator = q.Op
	}

	if operator != "" {
		// Create a copy with unified operator field
		qCopy := q
		qCopy.Operator = operator
		return s.buildOperatorQuery(qCopy)
	}

	queryType := q.Type
	if queryType == "" {
		queryType = "match_phrase"
	}

	switch queryType {
	case "match", "match_phrase":
		return map[string]interface{}{
			queryType: map[string]interface{}{
				q.Field: q.Value,
			},
		}
	case "term":
		return map[string]interface{}{
			"term": map[string]interface{}{
				q.Field: q.Value,
			},
		}
	case "terms":
		return map[string]interface{}{
			"terms": map[string]interface{}{
				q.Field: q.Value,
			},
		}
	case "range":
		if rangeMap, ok := q.Value.(map[string]interface{}); ok {
			return map[string]interface{}{
				"range": map[string]interface{}{
					q.Field: rangeMap,
				},
			}
		}
	case "exists":
		return map[string]interface{}{
			"exists": map[string]interface{}{
				"field": q.Field,
			},
		}
	case "regexp", "wildcard":
		return map[string]interface{}{
			queryType: map[string]interface{}{
				q.Field: q.Value,
			},
		}
	}

	return nil
}

func (s *Service) buildOperatorQuery(q models.QueryCondition) map[string]interface{} {
	switch q.Operator {
	case "=", "==", "equals":
		return map[string]interface{}{
			"term": map[string]interface{}{
				q.Field: q.Value,
			},
		}
	case "!=", "not_equals":
		return map[string]interface{}{
			"bool": map[string]interface{}{
				"must_not": []map[string]interface{}{
					{"term": map[string]interface{}{q.Field: q.Value}},
				},
			},
		}
	case ">", "gt":
		return map[string]interface{}{
			"range": map[string]interface{}{
				q.Field: map[string]interface{}{"gt": q.Value},
			},
		}
	case ">=", "gte":
		return map[string]interface{}{
			"range": map[string]interface{}{
				q.Field: map[string]interface{}{"gte": q.Value},
			},
		}
	case "<", "lt":
		return map[string]interface{}{
			"range": map[string]interface{}{
				q.Field: map[string]interface{}{"lt": q.Value},
			},
		}
	case "<=", "lte":
		return map[string]interface{}{
			"range": map[string]interface{}{
				q.Field: map[string]interface{}{"lte": q.Value},
			},
		}
	case "contains":
		// Use wildcard for true substring match (works for keyword fields too).
		// Escape special wildcard characters so "contains" behaves like literal contains.
		if s, ok := q.Value.(string); ok {
			val := escapeWildcardLiteral(s)
			return map[string]interface{}{
				"wildcard": map[string]interface{}{
					q.Field: map[string]interface{}{
						"value":            fmt.Sprintf("*%s*", val),
						"case_insensitive": true,
					},
				},
			}
		}
		// Fallback for non-string values.
		return map[string]interface{}{
			"match": map[string]interface{}{
				q.Field: q.Value,
			},
		}
	case "not_contains":
		if s, ok := q.Value.(string); ok {
			val := escapeWildcardLiteral(s)
			return map[string]interface{}{
				"bool": map[string]interface{}{
					"must_not": []map[string]interface{}{
						{
							"wildcard": map[string]interface{}{
								q.Field: map[string]interface{}{
									"value":            fmt.Sprintf("*%s*", val),
									"case_insensitive": true,
								},
							},
						},
					},
				},
			}
		}
		return map[string]interface{}{
			"bool": map[string]interface{}{
				"must_not": []map[string]interface{}{
					{"match": map[string]interface{}{q.Field: q.Value}},
				},
			},
		}
	case "exists":
		return map[string]interface{}{
			"exists": map[string]interface{}{
				"field": q.Field,
			},
		}
	}
	return nil
}

// escapeWildcardLiteral escapes characters that have special meaning in ES wildcard queries.
// See: https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl-wildcard-query.html
func escapeWildcardLiteral(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `*`, `\*`)
	s = strings.ReplaceAll(s, `?`, `\?`)
	return s
}

func (s *Service) extractDocuments(response map[string]interface{}) []map[string]interface{} {
	hits, _ := response["hits"].(map[string]interface{})
	hitsList, _ := hits["hits"].([]interface{})

	var docs []map[string]interface{}
	for _, hit := range hitsList {
		hitMap, _ := hit.(map[string]interface{})
		source, _ := hitMap["_source"].(map[string]interface{})
		index, _ := hitMap["_index"].(string)
		id, _ := hitMap["_id"].(string)

		doc := make(map[string]interface{})
		for k, v := range source {
			doc[k] = v
		}
		doc["_index"] = index
		doc["_id"] = id

		docs = append(docs, doc)
	}

	return docs
}
