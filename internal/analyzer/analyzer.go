package analyzer

import (
	"fmt"
	"math"

	"github.com/alexandrmotologa/planforge/internal/parser"
)

// EvaluatePlan orchestrates all diagnostic analyzers and produces a consolidated report.
func EvaluatePlan(plan *parser.Plan, indexRecs []IndexRecommendation) *FullReport {
	if plan == nil {
		return nil
	}

	timeHogs := AnalyzeBottlenecks(plan)
	estimationAnomalies := AnalyzeEstimationAnomalies(plan)
	bufferEfficiency := AnalyzeBufferEfficiency(plan)
	memorySpills := AnalyzeMemorySpills(plan)
	jitFinding := AnalyzeJIT(plan)
	antiPatterns := AnalyzeAntiPatterns(plan)

	score := 100

	// Deduct for time hogs
	for _, th := range timeHogs {
		if th.Severity == SeverityCritical {
			score -= 25
		} else {
			score -= 10
		}
	}

	// Deduct for estimation anomalies
	for _, ea := range estimationAnomalies {
		if ea.Severity == SeverityCritical {
			score -= 20
		} else {
			score -= 10
		}
	}

	// Deduct for buffer inefficiency
	if bufferEfficiency.Severity == SeverityCritical {
		score -= 25
	} else if bufferEfficiency.Severity == SeverityWarning {
		score -= 15
	}

	// Deduct for memory spills
	for range memorySpills {
		score -= 20
	}

	// Deduct for JIT overhead
	if jitFinding != nil {
		if jitFinding.Severity == SeverityCritical {
			score -= 20
		} else {
			score -= 10
		}
	}

	// Deduct for anti-patterns
	for _, ap := range antiPatterns {
		if ap.Severity == SeverityCritical {
			score -= 15
		} else {
			score -= 5
		}
	}

	if score < 0 {
		score = 0
	}

	grade := "A"
	switch {
	case score >= 90:
		grade = "A"
	case score >= 75:
		grade = "B"
	case score >= 60:
		grade = "C"
	case score >= 40:
		grade = "D"
	default:
		grade = "F"
	}

	var insights []string
	if len(timeHogs) > 0 {
		insights = append(insights, fmt.Sprintf("Query has %d dominant operation(s) taking >= 20%% runtime.", len(timeHogs)))
	}
	if len(indexRecs) > 0 {
		insights = append(insights, fmt.Sprintf("Missing %d recommended index(es) for sequential table scans.", len(indexRecs)))
	}
	if len(estimationAnomalies) > 0 {
		insights = append(insights, fmt.Sprintf("Found %d planner estimation anomalies. Catalog statistics may be stale.", len(estimationAnomalies)))
	}
	if bufferEfficiency.SharedReadBlocks > 0 {
		insights = append(insights, fmt.Sprintf("Read %.2f MB from physical disk (%.1f%% RAM hit ratio).", bufferEfficiency.DiskReadMB, bufferEfficiency.HitRatioPercent))
	}
	if len(memorySpills) > 0 {
		insights = append(insights, "Detected memory overflow to disk temporary storage during execution.")
	}
	if jitFinding != nil {
		insights = append(insights, fmt.Sprintf("JIT compilation took %.2f ms (%.1f%% of query duration).", jitFinding.JITTimeMs, jitFinding.OverheadPercent))
	}
	if len(antiPatterns) > 0 {
		insights = append(insights, fmt.Sprintf("Found %d SQL anti-pattern(s) impacting plan efficiency.", len(antiPatterns)))
	}
	if len(insights) == 0 {
		insights = append(insights, "Plan execution profile is well balanced with good buffer cache locality.")
	}

	return &FullReport{
		Plan:                 plan,
		HealthScore:          score,
		HealthGrade:          grade,
		TimeHogs:             timeHogs,
		EstimationAnomalies:  estimationAnomalies,
		BufferEfficiency:     bufferEfficiency,
		MemorySpills:         memorySpills,
		JITFinding:           jitFinding,
		AntiPatterns:         antiPatterns,
		IndexRecommendations: indexRecs,
		QuickInsights:        insights,
	}
}

// ComparePlans calculates deltas between Plan A (baseline) and Plan B (optimized).
type PlanComparison struct {
	BaselineID           string  `json:"baseline_id"`
	OptimizedID          string  `json:"optimized_id"`
	DurationDeltaMs      float64 `json:"duration_delta_ms"`
	DurationDeltaPercent float64 `json:"duration_delta_percent"`
	CostDelta            float64 `json:"cost_delta"`
	CostDeltaPercent     float64 `json:"cost_delta_percent"`
	BufferReadDelta      int64   `json:"buffer_read_delta"`
	BufferHitDelta       int64   `json:"buffer_hit_delta"`
	DiskBytesReadDelta   int64   `json:"disk_bytes_read_delta"`
	Summary              string  `json:"summary"`
}

// Compare evaluates performance improvements between two execution plans.
func Compare(baseline, optimized *parser.Plan) PlanComparison {
	comp := PlanComparison{
		BaselineID:  baseline.ID,
		OptimizedID: optimized.ID,
	}

	comp.DurationDeltaMs = math.Round((optimized.TotalDurationMs-baseline.TotalDurationMs)*1000) / 1000
	if baseline.TotalDurationMs > 0 {
		comp.DurationDeltaPercent = math.Round(((optimized.TotalDurationMs-baseline.TotalDurationMs)/baseline.TotalDurationMs)*10000) / 100
	}

	comp.CostDelta = math.Round((optimized.Root.TotalCost-baseline.Root.TotalCost)*100) / 100
	if baseline.Root.TotalCost > 0 {
		comp.CostDeltaPercent = math.Round(((optimized.Root.TotalCost-baseline.Root.TotalCost)/baseline.Root.TotalCost)*10000) / 100
	}

	comp.BufferReadDelta = optimized.Summary.TotalSharedReadBlocks - baseline.Summary.TotalSharedReadBlocks
	comp.BufferHitDelta = optimized.Summary.TotalSharedHitBlocks - baseline.Summary.TotalSharedHitBlocks
	comp.DiskBytesReadDelta = optimized.Summary.DiskReadBytes - baseline.Summary.DiskReadBytes

	if comp.DurationDeltaPercent < 0 {
		comp.Summary = fmt.Sprintf("Optimized plan is %.1f%% faster (reduced runtime by %.2f ms).", math.Abs(comp.DurationDeltaPercent), math.Abs(comp.DurationDeltaMs))
	} else if comp.DurationDeltaPercent > 0 {
		comp.Summary = fmt.Sprintf("Optimized plan is %.1f%% slower (increased runtime by %.2f ms).", comp.DurationDeltaPercent, comp.DurationDeltaMs)
	} else {
		comp.Summary = "Both plans exhibit identical total execution durations."
	}

	return comp
}
