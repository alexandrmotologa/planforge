package analyzer

import (
	"fmt"
	"math"

	"github.com/alexandrmotologa/planforge/internal/parser"
)

// AnalyzeBufferEfficiency calculates cache hit ratio and identifies disk I/O bottlenecks.
func AnalyzeBufferEfficiency(plan *parser.Plan) BufferEfficiencyFinding {
	if plan == nil {
		return BufferEfficiencyFinding{HitRatioPercent: 100.0, Severity: SeverityInfo}
	}

	hit := plan.Summary.TotalSharedHitBlocks
	read := plan.Summary.TotalSharedReadBlocks
	total := hit + read

	if total == 0 {
		return BufferEfficiencyFinding{
			HitRatioPercent: 100.0,
			Severity:        SeverityInfo,
			Message:         "No buffer statistics recorded in execution plan (BUFFERS option may not have been specified).",
			Recommendation:  "Run EXPLAIN (ANALYZE, BUFFERS) to capture buffer hit and disk read statistics.",
		}
	}

	ratio := float64(hit) / float64(total) * 100.0
	diskBytes := read * 8192
	diskMB := float64(diskBytes) / (1024 * 1024)

	severity := SeverityInfo
	msg := fmt.Sprintf("Buffer hit ratio is %.1f%% (%d hits, %d disk reads).", ratio, hit, read)
	rec := "Buffer cache efficiency is healthy. Data was served predominantly from RAM."

	if ratio < 50.0 {
		severity = SeverityCritical
		msg = fmt.Sprintf("Critical disk I/O load: %.1f%% buffer hit ratio. Query read %.2f MB directly from disk (%d blocks).", ratio, diskMB, read)
		rec = "Increase shared_buffers or add indexes to reduce full table pages read from physical storage."
	} else if ratio < 85.0 {
		severity = SeverityWarning
		msg = fmt.Sprintf("Moderate disk I/O: %.1f%% buffer hit ratio. Query read %.2f MB from disk (%d blocks).", ratio, diskMB, read)
		rec = "Consider index optimization or table vacuuming to reduce page misses."
	}

	return BufferEfficiencyFinding{
		HitRatioPercent: math.Round(ratio*100) / 100,
		SharedHitBlocks: hit,
		SharedReadBlocks: read,
		DiskReadBytes:   diskBytes,
		DiskReadMB:      math.Round(diskMB*100) / 100,
		Severity:        severity,
		Message:         msg,
		Recommendation:  rec,
	}
}
