package analyzer

import (
	"fmt"
	"math"

	"github.com/alexandrmotologa/planforge/internal/parser"
)

// AnalyzeEstimationAnomalies detects severe disparities between planned rows and actual rows.
func AnalyzeEstimationAnomalies(plan *parser.Plan) []EstimationFinding {
	if plan == nil || plan.Root == nil {
		return nil
	}

	var findings []EstimationFinding
	nodes := plan.FlattenNodes()

	for _, n := range nodes {
		// Only check if actual execution data was collected
		if n.ActualLoops == 0 && n.ActualRows == 0 {
			continue
		}

		totalActual := n.ActualRows * float64(n.ActualLoops)
		planned := n.PlanRows
		if planned <= 0 {
			planned = 1.0
		}

		ratio := totalActual / planned
		skew := ratio
		if skew < 1.0 && skew > 0 {
			skew = 1.0 / skew
		}

		// Flag anomalies where skew exceeds 5x (warning) or 10x (critical)
		if skew < 5.0 {
			continue
		}

		severity := SeverityWarning
		if skew >= 10.0 {
			severity = SeverityCritical
		}

		msg := ""
		rec := ""
		if totalActual > planned {
			msg = fmt.Sprintf("Under-estimated rows by %.1fx on %s (planned %.0f rows, actual %.0f rows).", skew, n.DisplayLabel(), planned, totalActual)
		} else {
			msg = fmt.Sprintf("Over-estimated rows by %.1fx on %s (planned %.0f rows, actual %.0f rows).", skew, n.DisplayLabel(), planned, totalActual)
		}

		if n.RelationName != "" {
			if plan.Engine == parser.EnginePostgres {
				rec = fmt.Sprintf("Run 'ANALYZE %s;' to refresh table statistics and update planner row estimates.", n.RelationName)
			} else {
				rec = fmt.Sprintf("Run 'ANALYZE TABLE %s;' to recalculate key distribution statistics.", n.RelationName)
			}
		} else {
			rec = "Table statistics are outdated or predicate correlation is causing planner miscalculations."
		}

		findings = append(findings, EstimationFinding{
			NodeID:         n.ID,
			NodeType:       n.NodeType,
			Relation:       n.RelationName,
			PlannedRows:    planned,
			ActualRows:     totalActual,
			Ratio:          math.Round(ratio*100) / 100,
			Severity:       severity,
			Message:        msg,
			Recommendation: rec,
		})
	}

	return findings
}
