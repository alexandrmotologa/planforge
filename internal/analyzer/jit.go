package analyzer

import (
	"fmt"
	"math"

	"github.com/alexandrmotologa/planforge/internal/parser"
)

// AnalyzeJIT evaluates whether Just-In-Time compilation caused disproportionate latency.
func AnalyzeJIT(plan *parser.Plan) *JITFinding {
	if plan == nil || plan.JIT == nil || plan.JIT.Timing.TotalMs <= 0 {
		return nil
	}

	jitTime := plan.JIT.Timing.TotalMs
	execTime := plan.ExecutionTimeMs
	if execTime <= 0 {
		execTime = plan.TotalDurationMs
	}

	total := jitTime + execTime
	if total <= 0 {
		return nil
	}

	overheadPct := (jitTime / total) * 100.0

	// If JIT compilation takes more than 20% of total query turnaround or is longer than execution itself
	if overheadPct < 20.0 && jitTime < execTime {
		return nil
	}

	severity := SeverityWarning
	if overheadPct >= 50.0 || jitTime > execTime*2 {
		severity = SeverityCritical
	}

	msg := fmt.Sprintf("JIT compilation took %.2f ms, accounting for %.1f%% of overall query time (execution was %.2f ms).",
		jitTime, overheadPct, execTime)

	rec := "Disable JIT for this query using 'SET jit = off;' or tune 'jit_above_cost' to avoid compilation overhead on short queries."

	return &JITFinding{
		JITTimeMs:       math.Round(jitTime*100) / 100,
		ExecutionTimeMs: math.Round(execTime*100) / 100,
		OverheadPercent: math.Round(overheadPct*10) / 10,
		Severity:        severity,
		Message:         msg,
		Recommendation:  rec,
	}
}
