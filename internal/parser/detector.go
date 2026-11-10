package parser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DetectAndParse automatically identifies the engine and format, then parses the input.
func DetectAndParse(input string) (*Plan, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, fmt.Errorf("empty plan input")
	}

	// 1. Check if JSON
	if strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{") {
		// Try parsing as generic JSON to check keys
		var js any
		if err := json.Unmarshal([]byte(trimmed), &js); err == nil {
			// Check MySQL query_block
			if obj, ok := js.(map[string]any); ok {
				if _, hasQB := obj["query_block"]; hasQB {
					return ParseMySQLJSON([]byte(trimmed))
				}
				if _, hasPlan := obj["Plan"]; hasPlan {
					return ParsePostgresJSON([]byte(trimmed))
				}
				if _, hasNodeType := obj["Node Type"]; hasNodeType {
					return ParsePostgresJSON([]byte(trimmed))
				}
			} else if arr, ok := js.([]any); ok && len(arr) > 0 {
				if first, ok := arr[0].(map[string]any); ok {
					if _, hasPlan := first["Plan"]; hasPlan {
						return ParsePostgresJSON([]byte(trimmed))
					}
				}
			}
		}
		// If it's JSON but not clearly identified, try Postgres JSON first
		if p, err := ParsePostgresJSON([]byte(trimmed)); err == nil {
			return p, nil
		}
		if m, err := ParseMySQLJSON([]byte(trimmed)); err == nil {
			return m, nil
		}
	}

	// 2. Check if MySQL EXPLAIN ANALYZE tree text
	if strings.Contains(trimmed, "-> Nested loop") ||
		strings.Contains(trimmed, "-> Table scan") ||
		strings.Contains(trimmed, "-> Filter:") ||
		strings.Contains(trimmed, "-> Single-row index lookup") {
		if m, err := ParseMySQLAnalyzeText(trimmed); err == nil {
			return m, nil
		}
	}

	// 3. Fallback to PostgreSQL text format
	if strings.Contains(trimmed, "Seq Scan") ||
		strings.Contains(trimmed, "Index Scan") ||
		strings.Contains(trimmed, "Hash Join") ||
		strings.Contains(trimmed, "cost=") ||
		strings.Contains(trimmed, "Planning Time:") ||
		strings.Contains(trimmed, "Execution Time:") {
		return ParsePostgresText(trimmed)
	}

	// Default attempt PostgreSQL text parser
	return ParsePostgresText(trimmed)
}
