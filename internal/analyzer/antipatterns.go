package analyzer

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/planforge/internal/parser"
)

// AnalyzeAntiPatterns detects common SQL anti-patterns visible in execution plans.
func AnalyzeAntiPatterns(plan *parser.Plan) []AntiPatternFinding {
	if plan == nil || plan.Root == nil {
		return nil
	}

	var findings []AntiPatternFinding
	nodes := plan.FlattenNodes()

	for _, n := range nodes {
		// 1. Correlated SubPlan execution loop
		if n.SubplanName != "" && n.ActualLoops > 5 {
			findings = append(findings, AntiPatternFinding{
				NodeID:         n.ID,
				PatternType:    "Correlated SubPlan Loop",
				Severity:       SeverityCritical,
				Message:        fmt.Sprintf("SubPlan '%s' executed %d times in a nested loop.", n.SubplanName, n.ActualLoops),
				Recommendation: "Rewrite correlated subquery into an explicit JOIN or CTE to enable set-based hash or merge join execution.",
			})
		}

		// 2. Implicit Type Casting in Filters
		if n.Filter != "" {
			if strings.Contains(n.Filter, "::numeric") || strings.Contains(n.Filter, "::text") {
				// Check if on a scan node
				if strings.Contains(strings.ToLower(n.NodeType), "seq scan") {
					findings = append(findings, AntiPatternFinding{
						NodeID:         n.ID,
						PatternType:    "Implicit Typecast in Filter",
						Severity:       SeverityWarning,
						Message:        fmt.Sprintf("Filter on %s uses explicit typecast (%s) which may disable index lookups.", n.DisplayLabel(), n.Filter),
						Recommendation: "Align client parameter data types with column schema types to eliminate casting overhead and enable index scans.",
					})
				}
			}
		}

		// 3. Lossy Bitmap Recheck
		if strings.Contains(strings.ToLower(n.NodeType), "bitmap heap scan") && n.RowsRemovedByIndexRecheck > 100 {
			findings = append(findings, AntiPatternFinding{
				NodeID:         n.ID,
				PatternType:    "Lossy Bitmap Recheck",
				Severity:       SeverityWarning,
				Message:        fmt.Sprintf("Bitmap Heap Scan had %d rows rechecked against index, indicating memory pressure during bitmap generation.", n.RowsRemovedByIndexRecheck),
				Recommendation: "Increase work_mem to allow exact bitmap storage in RAM without falling back to lossy page rechecks.",
			})
		}
	}

	return findings
}
