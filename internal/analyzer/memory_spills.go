package analyzer

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/planforge/internal/parser"
)

// AnalyzeMemorySpills identifies operators overflowing memory into temporary disk files.
func AnalyzeMemorySpills(plan *parser.Plan) []MemorySpillFinding {
	if plan == nil || plan.Root == nil {
		return nil
	}

	var findings []MemorySpillFinding
	nodes := plan.FlattenNodes()

	for _, n := range nodes {
		// Postgres Sort Disk Spill
		if strings.EqualFold(n.SortSpaceType, "Disk") || strings.Contains(strings.ToLower(n.SortMethod), "external") {
			findings = append(findings, MemorySpillFinding{
				NodeID:         n.ID,
				NodeType:       n.NodeType,
				SpillType:      "Sort External Merge (Disk Spill)",
				SpaceUsedKb:    n.SortSpaceUsedKb,
				Severity:       SeverityCritical,
				Message:        fmt.Sprintf("Sort operator overflowed memory to disk (%d kB written).", n.SortSpaceUsedKb),
				Recommendation: fmt.Sprintf("Increase work_mem for this session (e.g. SET work_mem = '%dMB';) to perform sorting in RAM.", (n.SortSpaceUsedKb/1024)*2+16),
			})
		}

		// Temp buffer reads/writes
		if n.Buffers.TempReadBlocks > 0 || n.Buffers.TempWrittenBlocks > 0 {
			tempMB := float64((n.Buffers.TempReadBlocks+n.Buffers.TempWrittenBlocks)*8192) / (1024 * 1024)
			findings = append(findings, MemorySpillFinding{
				NodeID:         n.ID,
				NodeType:       n.NodeType,
				SpillType:      "Temporary File I/O Spill",
				SpaceUsedKb:    (n.Buffers.TempReadBlocks + n.Buffers.TempWrittenBlocks) * 8,
				Severity:       SeverityCritical,
				Message:        fmt.Sprintf("Operator wrote %.2f MB to temporary files on disk.", tempMB),
				Recommendation: "Increase work_mem or maintenance_work_mem to keep hash joins or aggregates in memory.",
			})
		}
	}

	return findings
}
