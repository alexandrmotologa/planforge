package parser

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	reCost      = regexp.MustCompile(`cost=([\d\.]+)\.\.([\d\.]+)\s+rows=([\d\.]+)\s+width=(\d+)`)
	reActual    = regexp.MustCompile(`actual\s+(?:time=([\d\.]+)\.\.([\d\.]+)\s+)?rows=([\d\.]+)\s+loops=(\d+)`)
	reBuffers   = regexp.MustCompile(`Buffers:\s*(.+)`)
	reBufPart   = regexp.MustCompile(`(shared|temp|local)\s+([a-zA-Z\s]+)=([0-9]+)`)
	rePlanTime  = regexp.MustCompile(`Planning\s+[tT]ime:\s*([\d\.]+)\s*ms`)
	reExecTime  = regexp.MustCompile(`Execution\s+[tT]ime:\s*([\d\.]+)\s*ms`)
	reFilter    = regexp.MustCompile(`Filter:\s*(.+)`)
	reRowsRem   = regexp.MustCompile(`Rows Removed by Filter:\s*(\d+)`)
	reIndexCond = regexp.MustCompile(`Index Cond:\s*(.+)`)
	reHashCond  = regexp.MustCompile(`Hash Cond:\s*(.+)`)
	reSortKey   = regexp.MustCompile(`Sort Key:\s*(.+)`)
	reSortMethod= regexp.MustCompile(`Sort Method:\s*([^\s]+)\s+(Memory|Disk):\s*(\d+)kB`)
)

type textLine struct {
	indent int
	isArrow bool
	text string
}

// ParsePostgresText parses standard psql text output of EXPLAIN (ANALYZE, BUFFERS).
func ParsePostgresText(raw string) (*Plan, error) {
	scanner := bufio.NewScanner(strings.NewReader(raw))
	var lines []textLine
	var planningTime, executionTime float64

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if match := rePlanTime.FindStringSubmatch(trimmed); len(match) > 1 {
			planningTime, _ = strconv.ParseFloat(match[1], 64)
			continue
		}
		if match := reExecTime.FindStringSubmatch(trimmed); len(match) > 1 {
			executionTime, _ = strconv.ParseFloat(match[1], 64)
			continue
		}

		// Calculate indentation
		indent := 0
		for _, ch := range line {
			if ch == ' ' {
				indent++
			} else if ch == '\t' {
				indent += 4
			} else {
				break
			}
		}

		isArrow := strings.HasPrefix(trimmed, "->")
		cleaned := trimmed
		if isArrow {
			cleaned = strings.TrimSpace(strings.TrimPrefix(cleaned, "->"))
		}

		lines = append(lines, textLine{
			indent:  indent,
			isArrow: isArrow,
			text:    cleaned,
		})
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("empty plan text")
	}

	nodeCounter := 0
	root, err := buildTextTree(lines, &nodeCounter)
	if err != nil {
		return nil, err
	}

	totalDuration := executionTime
	if totalDuration <= 0 && root.ActualTotalTimeMs > 0 {
		loops := root.ActualLoops
		if loops <= 0 {
			loops = 1
		}
		totalDuration = root.ActualTotalTimeMs * float64(loops)
	}
	if totalDuration <= 0 {
		totalDuration = root.TotalCost
	}

	plan := &Plan{
		ID:              fmt.Sprintf("plan-text-%d", nodeCounter),
		Engine:          EnginePostgres,
		Format:          FormatText,
		Root:            root,
		PlanningTimeMs:  planningTime,
		ExecutionTimeMs: executionTime,
		TotalDurationMs: totalDuration,
		RawInput:        raw,
	}

	computeMetrics(plan)
	return plan, nil
}

type textNodeBuilder struct {
	node      *PlanNode
	indent    int
	sublines  []string
	children  []*textNodeBuilder
}

func buildTextTree(lines []textLine, counter *int) (*PlanNode, error) {
	// Group lines into nodes based on indentation and `->` or primary operator lines
	// A new node starts when a line contains cost= or `->`
	var nodeBuilders []*textNodeBuilder
	var current *textNodeBuilder

	for _, l := range lines {
		if l.isArrow || reCost.MatchString(l.text) || current == nil {
			*counter++
			newNode := &PlanNode{
				ID: fmt.Sprintf("node-%d", *counter),
			}
			b := &textNodeBuilder{
				node:   newNode,
				indent: l.indent,
			}
			b.sublines = append(b.sublines, l.text)
			nodeBuilders = append(nodeBuilders, b)
			current = b
		} else {
			if current != nil {
				current.sublines = append(current.sublines, l.text)
			}
		}
	}

	if len(nodeBuilders) == 0 {
		return nil, fmt.Errorf("could not identify any execution plan nodes")
	}

	// Build hierarchy by indentation
	rootBuilder := nodeBuilders[0]
	var stack []*textNodeBuilder
	stack = append(stack, rootBuilder)

	for i := 1; i < len(nodeBuilders); i++ {
		b := nodeBuilders[i]
		for len(stack) > 0 && stack[len(stack)-1].indent >= b.indent {
			stack = stack[:len(stack)-1]
		}
		if len(stack) > 0 {
			parent := stack[len(stack)-1]
			parent.children = append(parent.children, b)
		} else {
			// Sibling to root or deeper
			rootBuilder.children = append(rootBuilder.children, b)
		}
		stack = append(stack, b)
	}

	return assembleNode(rootBuilder), nil
}

func assembleNode(b *textNodeBuilder) *PlanNode {
	n := b.node
	header := b.sublines[0]

	// Extract cost
	if m := reCost.FindStringSubmatch(header); len(m) > 4 {
		n.StartupCost, _ = strconv.ParseFloat(m[1], 64)
		n.TotalCost, _ = strconv.ParseFloat(m[2], 64)
		n.PlanRows, _ = strconv.ParseFloat(m[3], 64)
		w, _ := strconv.ParseInt(m[4], 10, 64)
		n.PlanWidth = w
	}

	// Extract actual
	if m := reActual.FindStringSubmatch(header); len(m) > 3 {
		if m[1] != "" {
			n.ActualStartupTimeMs, _ = strconv.ParseFloat(m[1], 64)
		}
		if m[2] != "" {
			n.ActualTotalTimeMs, _ = strconv.ParseFloat(m[2], 64)
		}
		n.ActualRows, _ = strconv.ParseFloat(m[3], 64)
		l, _ := strconv.ParseInt(m[4], 10, 64)
		n.ActualLoops = l
	}
	if n.ActualLoops <= 0 {
		n.ActualLoops = 1
	}

	// Extract node type and relation
	// Example: "Seq Scan on orders orders_alias" or "Index Scan using idx_users_id on users"
	cleanHeader := header
	if idx := strings.Index(cleanHeader, "("); idx != -1 {
		cleanHeader = strings.TrimSpace(cleanHeader[:idx])
	}

	parseHeaderTokens(cleanHeader, n)

	// Process remaining lines for filters, buffers, sort keys, etc.
	for _, sl := range b.sublines[1:] {
		slTrim := strings.TrimSpace(sl)

		if m := reFilter.FindStringSubmatch(slTrim); len(m) > 1 {
			n.Filter = m[1]
		}
		if m := reRowsRem.FindStringSubmatch(slTrim); len(m) > 1 {
			rem, _ := strconv.ParseInt(m[1], 10, 64)
			n.RowsRemovedByFilter = rem
		}
		if m := reIndexCond.FindStringSubmatch(slTrim); len(m) > 1 {
			n.IndexCond = m[1]
		}
		if m := reHashCond.FindStringSubmatch(slTrim); len(m) > 1 {
			n.HashCond = m[1]
		}
		if m := reSortKey.FindStringSubmatch(slTrim); len(m) > 1 {
			keys := strings.Split(m[1], ",")
			for _, k := range keys {
				n.SortKey = append(n.SortKey, strings.TrimSpace(k))
			}
		}
		if m := reSortMethod.FindStringSubmatch(slTrim); len(m) > 3 {
			n.SortMethod = m[1]
			n.SortSpaceType = m[2]
			kb, _ := strconv.ParseInt(m[3], 10, 64)
			n.SortSpaceUsedKb = kb
		}

		// Buffer stats
		if strings.HasPrefix(slTrim, "Buffers:") {
			parseBufferLine(slTrim, &n.Buffers)
		}
	}

	for _, ch := range b.children {
		n.Children = append(n.Children, assembleNode(ch))
	}

	return n
}

func parseHeaderTokens(header string, n *PlanNode) {
	// Common patterns:
	// "Seq Scan on table [alias]"
	// "Index Scan using index_name on table [alias]"
	// "Bitmap Heap Scan on table"
	// "Hash Join"
	if strings.Contains(header, " on ") {
		parts := strings.Split(header, " on ")
		n.NodeType = strings.TrimSpace(parts[0])
		relPart := strings.TrimSpace(parts[1])

		// check if alias exists: "table alias"
		relTokens := strings.Fields(relPart)
		if len(relTokens) >= 1 {
			n.RelationName = relTokens[0]
			if len(relTokens) >= 2 {
				n.Alias = relTokens[1]
			}
		}

		if strings.Contains(n.NodeType, " using ") {
			idxParts := strings.Split(n.NodeType, " using ")
			n.NodeType = strings.TrimSpace(idxParts[0])
			n.IndexName = strings.TrimSpace(idxParts[1])
		}
	} else {
		n.NodeType = strings.TrimSpace(header)
	}
}

func parseBufferLine(line string, buf *BufferStats) {
	// Buffers: shared hit=42 read=5 dirtied=1 written=2, temp read=10 written=12
	parts := strings.Split(line, ",")
	for _, part := range parts {
		tokens := strings.Fields(part)
		scope := ""
		for _, tok := range tokens {
			if tok == "shared" || tok == "temp" || tok == "local" {
				scope = tok
				continue
			}
			if strings.Contains(tok, "=") {
				kv := strings.Split(tok, "=")
				val, _ := strconv.ParseInt(kv[1], 10, 64)
				key := kv[0]
				switch scope + " " + key {
				case "shared hit":
					buf.SharedHitBlocks += val
				case "shared read":
					buf.SharedReadBlocks += val
				case "shared dirtied":
					buf.SharedDirtiedBlocks += val
				case "shared written":
					buf.SharedWrittenBlocks += val
				case "local hit":
					buf.LocalHitBlocks += val
				case "local read":
					buf.LocalReadBlocks += val
				case "temp read":
					buf.TempReadBlocks += val
				case "temp written":
					buf.TempWrittenBlocks += val
				}
			}
		}
	}
}
