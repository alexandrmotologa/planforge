package analyzer

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/planforge/internal/parser"
)

// AnalyzeBottlenecks identifies operations that consume an excessive share of query execution time.
func AnalyzeBottlenecks(plan *parser.Plan) []TimeHogFinding {
	if plan == nil || plan.Root == nil {
		return nil
	}

	var findings []TimeHogFinding
	nodes := plan.FlattenNodes()

	for _, n := range nodes {
		pct := n.DurationPercent
		if pct < 20.0 {
			continue
		}

		severity := SeverityWarning
		if pct >= 50.0 {
			severity = SeverityCritical
		}

		finding := TimeHogFinding{
			NodeID:          n.ID,
			NodeType:        n.NodeType,
			Relation:        n.RelationName,
			DurationMs:      n.ExclusiveDurationMs,
			DurationPercent: pct,
			Severity:        severity,
			Message:         fmt.Sprintf("%s accounts for %.1f%% of total execution time (%.2f ms).", n.DisplayLabel(), pct, n.ExclusiveDurationMs),
			Recommendation:  recommendForNode(n),
		}

		findings = append(findings, finding)
	}

	return findings
}

func recommendForNode(n *parser.PlanNode) string {
	nt := strings.ToLower(n.NodeType)
	if strings.Contains(nt, "seq scan") || strings.Contains(nt, "table scan") {
		return fmt.Sprintf("Add an index on %s matching filter predicates to convert this sequential scan into an index scan.", n.RelationName)
	}
	if strings.Contains(nt, "nested loop") {
		return "Nested loop iterations are dominating query duration. Ensure join keys on the inner relation have index coverage."
	}
	if strings.Contains(nt, "sort") {
		return "Sort operation is consuming significant execution time. Use an index matching the ORDER BY clause to avoid explicit sorting."
	}
	if strings.Contains(nt, "hash join") {
		return "Hash join build/probe duration is high. Verify hash join keys are indexed and join selectivity is optimal."
	}
	if strings.Contains(nt, "aggregate") {
		return "Aggregation is taking majority of time. Consider composite indexes covering GROUP BY columns or partial pre-aggregations."
	}
	return "Inspect predicate selectivity and data volume processed by this operator."
}
