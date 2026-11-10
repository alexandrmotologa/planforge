package advisor

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alexandrmotologa/planforge/internal/analyzer"
	"github.com/alexandrmotologa/planforge/internal/parser"
)

var (
	// Matches column comparisons: (col = val), (col > val), (col >= val), (col <= val), (col < val), (col IS NULL/TRUE/FALSE)
	reEqPredicate    = regexp.MustCompile(`(?:(?:\w+\.)?([a-zA-Z_][a-zA-Z0-9_]*))\s*=\s*([^()ANDandORor]+)`)
	reIsPredicate    = regexp.MustCompile(`(?:(?:\w+\.)?([a-zA-Z_][a-zA-Z0-9_]*))\s+IS\s+(NOT\s+)?(NULL|TRUE|FALSE)`)
	reRangePredicate = regexp.MustCompile(`(?:(?:\w+\.)?([a-zA-Z_][a-zA-Z0-9_]*))\s*(?:>|<|>=|<=|BETWEEN)\s*([^()ANDandORor]+)`)
	reInPredicate    = regexp.MustCompile(`(?:(?:\w+\.)?([a-zA-Z_][a-zA-Z0-9_]*))\s+IN\s*\(([^)]+)\)`)
)

// PredicateAnalysis holds extracted equality and range columns.
type PredicateAnalysis struct {
	EqualityColumns []string
	RangeColumns    []string
	SortColumns     []string
}

// RecommendIndexes analyzes execution plan nodes and synthesizes composite index recommendations.
func RecommendIndexes(plan *parser.Plan) []analyzer.IndexRecommendation {
	if plan == nil || plan.Root == nil {
		return nil
	}

	var recommendations []analyzer.IndexRecommendation
	nodes := plan.FlattenNodes()

	// Build map of sort keys per relation to optimize compound sort+filter indexes
	sortKeysByRel := make(map[string][]string)
	for _, n := range nodes {
		if len(n.SortKey) > 0 {
			for _, sk := range n.SortKey {
				// sk may be "orders.created_at DESC" or "created_at DESC"
				clean := strings.TrimSpace(sk)
				tokens := strings.Fields(clean)
				if len(tokens) > 0 {
					colName := tokens[0]
					if strings.Contains(colName, ".") {
						parts := strings.Split(colName, ".")
						rel := parts[0]
						col := parts[1]
						sortKeysByRel[rel] = append(sortKeysByRel[rel], col+" "+strings.Join(tokens[1:], " "))
					} else if n.RelationName != "" {
						sortKeysByRel[n.RelationName] = append(sortKeysByRel[n.RelationName], clean)
					}
				}
			}
		}
	}

	seenIndexes := make(map[string]bool)

	for _, n := range nodes {
		nt := strings.ToLower(n.NodeType)
		isSeqScan := strings.Contains(nt, "seq scan") || strings.Contains(nt, "table scan")
		isBitmapRecheck := strings.Contains(nt, "bitmap heap scan") && n.RowsRemovedByIndexRecheck > 100

		if (!isSeqScan && !isBitmapRecheck) || n.RelationName == "" || n.Filter == "" {
			continue
		}

		pred := parsePredicates(n.Filter)

		// Check if sort keys exist for this relation
		if sorts, ok := sortKeysByRel[n.RelationName]; ok {
			for _, s := range sorts {
				colOnly := strings.Fields(s)[0]
				if !contains(pred.EqualityColumns, colOnly) && !contains(pred.RangeColumns, colOnly) {
					pred.SortColumns = append(pred.SortColumns, s)
				}
			}
		}

		// Combine columns: Equality first, then Range, then Sort
		var finalCols []string
		for _, col := range pred.EqualityColumns {
			if !contains(finalCols, col) {
				finalCols = append(finalCols, col)
			}
		}
		for _, col := range pred.RangeColumns {
			if !contains(finalCols, col) {
				finalCols = append(finalCols, col)
			}
		}
		for _, col := range pred.SortColumns {
			if !contains(finalCols, col) {
				finalCols = append(finalCols, col)
			}
		}

		if len(finalCols) == 0 {
			continue
		}

		cleanTable := strings.Trim(n.RelationName, "`\"")
		indexName := generateIndexName(cleanTable, finalCols)
		if seenIndexes[indexName] {
			continue
		}
		seenIndexes[indexName] = true

		ddlPG := fmt.Sprintf("CREATE INDEX CONCURRENTLY %s\nON %s (%s);", indexName, cleanTable, strings.Join(finalCols, ", "))
		ddlMY := fmt.Sprintf("CREATE INDEX %s\nON %s (%s);", indexName, cleanTable, strings.Join(finalCols, ", "))

		impact := fmt.Sprintf("Converts %s on %s into an index lookup.", n.NodeType, cleanTable)
		if n.RowsRemovedByFilter > 0 {
			impact += fmt.Sprintf(" Avoids evaluating %d rows discarded by filter.", n.RowsRemovedByFilter)
		}

		recommendations = append(recommendations, analyzer.IndexRecommendation{
			NodeID:          n.ID,
			Table:           cleanTable,
			IndexName:       indexName,
			DDLPostgres:     ddlPG,
			DDLMySQL:        ddlMY,
			Columns:         finalCols,
			FilterPredicate: n.Filter,
			EstimatedImpact: impact,
			Reason:          fmt.Sprintf("Table scan detected with selective filter expression: %s", n.Filter),
		})
	}

	return recommendations
}

func parsePredicates(filter string) PredicateAnalysis {
	var analysis PredicateAnalysis
	cleaned := parser.CleanFilter(filter)

	// Clean backticks and quotes
	cleaned = strings.ReplaceAll(cleaned, "`", "")
	cleaned = strings.ReplaceAll(cleaned, "\"", "")

	// Extract Equality
	for _, match := range reEqPredicate.FindAllStringSubmatch(cleaned, -1) {
		if len(match) > 1 {
			col := match[1]
			if !isKeyword(col) && !contains(analysis.EqualityColumns, col) {
				analysis.EqualityColumns = append(analysis.EqualityColumns, col)
			}
		}
	}

	// Extract IS predicates
	for _, match := range reIsPredicate.FindAllStringSubmatch(cleaned, -1) {
		if len(match) > 1 {
			col := match[1]
			if !isKeyword(col) && !contains(analysis.EqualityColumns, col) {
				analysis.EqualityColumns = append(analysis.EqualityColumns, col)
			}
		}
	}

	// Extract IN predicates
	for _, match := range reInPredicate.FindAllStringSubmatch(cleaned, -1) {
		if len(match) > 1 {
			col := match[1]
			if !isKeyword(col) && !contains(analysis.EqualityColumns, col) {
				analysis.EqualityColumns = append(analysis.EqualityColumns, col)
			}
		}
	}

	// Extract Range predicates
	for _, match := range reRangePredicate.FindAllStringSubmatch(cleaned, -1) {
		if len(match) > 1 {
			col := match[1]
			if !isKeyword(col) && !contains(analysis.EqualityColumns, col) && !contains(analysis.RangeColumns, col) {
				analysis.RangeColumns = append(analysis.RangeColumns, col)
			}
		}
	}

	return analysis
}

func generateIndexName(table string, columns []string) string {
	var cleanCols []string
	for _, c := range columns {
		colOnly := strings.Fields(c)[0]
		cleanCols = append(cleanCols, colOnly)
	}
	base := fmt.Sprintf("idx_%s_%s", table, strings.Join(cleanCols, "_"))
	if len(base) > 60 {
		return base[:60]
	}
	return base
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

func isKeyword(s string) bool {
	keywords := map[string]bool{
		"true": true, "false": true, "null": true, "and": true, "or": true, "not": true,
		"case": true, "when": true, "then": true, "else": true, "end": true,
	}
	return keywords[strings.ToLower(s)]
}
