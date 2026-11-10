package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseMySQLJSON parses MySQL EXPLAIN FORMAT=JSON output.
func ParseMySQLJSON(data []byte) (*Plan, error) {
	var rawMap map[string]any
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return nil, fmt.Errorf("failed to parse MySQL JSON: %w", err)
	}

	qb, ok := rawMap["query_block"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing query_block in MySQL JSON plan")
	}

	nodeCounter := 0
	root := parseMySQLBlock("Root", qb, &nodeCounter)
	if root == nil {
		return nil, fmt.Errorf("could not construct plan from MySQL query_block")
	}

	totalCost := root.TotalCost
	if costInfo, ok := qb["cost_info"].(map[string]any); ok {
		if qc, ok := costInfo["query_cost"].(string); ok {
			totalCost, _ = strconv.ParseFloat(qc, 64)
		}
	}

	plan := &Plan{
		ID:              fmt.Sprintf("mysql-plan-%d", nodeCounter),
		Engine:          EngineMySQL,
		Format:          FormatJSON,
		Root:            root,
		TotalDurationMs: totalCost, // In MySQL FORMAT=JSON, duration is cost unless EXPLAIN ANALYZE is used
		RawInput:        string(data),
	}

	computeMetrics(plan)
	return plan, nil
}

func parseMySQLBlock(name string, block map[string]any, counter *int) *PlanNode {
	*counter++
	id := fmt.Sprintf("node-%d", *counter)

	node := &PlanNode{
		ID:          id,
		NodeType:    name,
		ActualLoops: 1,
	}

	// Cost info
	if costInfo, ok := block["cost_info"].(map[string]any); ok {
		if qc, ok := costInfo["query_cost"].(string); ok {
			c, _ := strconv.ParseFloat(qc, 64)
			node.TotalCost = c
		} else if pc, ok := costInfo["prefix_cost"].(string); ok {
			c, _ := strconv.ParseFloat(pc, 64)
			node.TotalCost = c
		}
	}

	// Operations: ordering_operation, grouping_operation, duplicates_removal, windowing
	for _, opKey := range []string{"ordering_operation", "grouping_operation", "duplicates_removal", "windowing"} {
		if op, ok := block[opKey].(map[string]any); ok {
			node.NodeType = formatMySQLOpName(opKey)
			if filesort, ok := op["using_filesort"].(bool); filesort && ok {
				node.SortMethod = "filesort"
			}
			child := parseMySQLBlock("Operation Child", op, counter)
			if child != nil {
				node.Children = append(node.Children, child)
			}
			return node
		}
	}

	// Nested loop join
	if nl, ok := block["nested_loop"].([]any); ok {
		node.NodeType = "Nested Loop Join"
		for _, item := range nl {
			if itemMap, ok := item.(map[string]any); ok {
				if tbl, ok := itemMap["table"].(map[string]any); ok {
					tblNode := parseMySQLTable(tbl, counter)
					node.Children = append(node.Children, tblNode)
				}
			}
		}
		return node
	}

	// Single table
	if tbl, ok := block["table"].(map[string]any); ok {
		return parseMySQLTable(tbl, counter)
	}

	return node
}

func parseMySQLTable(tbl map[string]any, counter *int) *PlanNode {
	*counter++
	id := fmt.Sprintf("node-%d", *counter)

	tableName, _ := tbl["table_name"].(string)
	accessType, _ := tbl["access_type"].(string)
	key, _ := tbl["key"].(string)
	cond, _ := tbl["attached_condition"].(string)

	nodeType := "Table Scan"
	switch accessType {
	case "ALL":
		nodeType = "Table Scan (ALL)"
	case "index":
		nodeType = "Index Scan"
	case "range":
		nodeType = "Index Range Scan"
	case "ref":
		nodeType = "Index Lookup (ref)"
	case "eq_ref":
		nodeType = "Unique Index Lookup (eq_ref)"
	case "const":
		nodeType = "Constant Lookup"
	default:
		if accessType != "" {
			nodeType = fmt.Sprintf("Access (%s)", accessType)
		}
	}

	var rowsExamined float64
	if r, ok := tbl["rows_examined_per_scan"].(float64); ok {
		rowsExamined = r
	} else if r, ok := tbl["rows_examined_per_scan"].(string); ok {
		rowsExamined, _ = strconv.ParseFloat(r, 64)
	}

	var rowsProduced float64
	if r, ok := tbl["rows_produced_per_join"].(float64); ok {
		rowsProduced = r
	} else if r, ok := tbl["rows_produced_per_join"].(string); ok {
		rowsProduced, _ = strconv.ParseFloat(r, 64)
	}

	var totalCost float64
	if costInfo, ok := tbl["cost_info"].(map[string]any); ok {
		if pc, ok := costInfo["prefix_cost"].(string); ok {
			totalCost, _ = strconv.ParseFloat(pc, 64)
		} else if rc, ok := costInfo["read_cost"].(string); ok {
			totalCost, _ = strconv.ParseFloat(rc, 64)
		}
	}

	var outputCols []string
	if usedCols, ok := tbl["used_columns"].([]any); ok {
		for _, col := range usedCols {
			if s, ok := col.(string); ok {
				outputCols = append(outputCols, s)
			}
		}
	}

	node := &PlanNode{
		ID:                  id,
		NodeType:            nodeType,
		RelationName:        tableName,
		IndexName:           key,
		Filter:              cond,
		TotalCost:           totalCost,
		PlanRows:            rowsExamined,
		ActualRows:          rowsProduced,
		ActualLoops:         1,
		Output:              outputCols,
	}

	return node
}

func formatMySQLOpName(key string) string {
	switch key {
	case "ordering_operation":
		return "Sort (Filesort)"
	case "grouping_operation":
		return "Aggregate / Group By"
	case "duplicates_removal":
		return "Distinct (Duplicates Removal)"
	case "windowing":
		return "Window Function"
	default:
		return key
	}
}

// MySQL 8.0.18+ EXPLAIN ANALYZE tree parser
var (
	reMySQLAnalyzeTime = regexp.MustCompile(`\(actual\s+time=([\d\.]+)\.\.([\d\.]+)\s+rows=([\d\.]+)\s+loops=(\d+)\)`)
	reMySQLAnalyzeCost = regexp.MustCompile(`\(cost=([\d\.]+)\s+rows=([\d\.]+)\)`)
)

// ParseMySQLAnalyzeText parses MySQL 8.0.18+ EXPLAIN ANALYZE tree text.
func ParseMySQLAnalyzeText(raw string) (*Plan, error) {
	scanner := bufio.NewScanner(strings.NewReader(raw))
	type myLine struct {
		indent int
		text   string
	}
	var lines []myLine

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
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
		cleaned := strings.TrimSpace(strings.TrimPrefix(trimmed, "->"))
		lines = append(lines, myLine{indent: indent, text: cleaned})
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("empty MySQL analyze text")
	}

	counter := 0
	type nodeWrap struct {
		node     *PlanNode
		indent   int
		children []*nodeWrap
	}

	var wraps []*nodeWrap
	for _, l := range lines {
		counter++
		node := parseMySQLAnalyzeLine(l.text, counter)
		wraps = append(wraps, &nodeWrap{
			node:   node,
			indent: l.indent,
		})
	}

	rootWrap := wraps[0]
	var stack []*nodeWrap
	stack = append(stack, rootWrap)

	for i := 1; i < len(wraps); i++ {
		w := wraps[i]
		for len(stack) > 0 && stack[len(stack)-1].indent >= w.indent {
			stack = stack[:len(stack)-1]
		}
		if len(stack) > 0 {
			stack[len(stack)-1].children = append(stack[len(stack)-1].children, w)
		} else {
			rootWrap.children = append(rootWrap.children, w)
		}
		stack = append(stack, w)
	}

	var buildTree func(w *nodeWrap) *PlanNode
	buildTree = func(w *nodeWrap) *PlanNode {
		for _, ch := range w.children {
			w.node.Children = append(w.node.Children, buildTree(ch))
		}
		return w.node
	}

	root := buildTree(rootWrap)
	totalDuration := root.ActualTotalTimeMs * float64(root.ActualLoops)
	if totalDuration <= 0 {
		totalDuration = root.TotalCost
	}

	plan := &Plan{
		ID:              fmt.Sprintf("mysql-analyze-%d", counter),
		Engine:          EngineMySQL,
		Format:          FormatText,
		Root:            root,
		TotalDurationMs: totalDuration,
		RawInput:        raw,
	}

	computeMetrics(plan)
	return plan, nil
}

func parseMySQLAnalyzeLine(text string, counter int) *PlanNode {
	node := &PlanNode{
		ID:          fmt.Sprintf("node-%d", counter),
		ActualLoops: 1,
	}

	// Actual time
	if m := reMySQLAnalyzeTime.FindStringSubmatch(text); len(m) > 4 {
		node.ActualStartupTimeMs, _ = strconv.ParseFloat(m[1], 64)
		node.ActualTotalTimeMs, _ = strconv.ParseFloat(m[2], 64)
		node.ActualRows, _ = strconv.ParseFloat(m[3], 64)
		loops, _ := strconv.ParseInt(m[4], 10, 64)
		node.ActualLoops = loops
	}

	// Cost
	if m := reMySQLAnalyzeCost.FindStringSubmatch(text); len(m) > 2 {
		node.TotalCost, _ = strconv.ParseFloat(m[1], 64)
		node.PlanRows, _ = strconv.ParseFloat(m[2], 64)
	}

	// Strip out (actual ...) and (cost ...)
	clean := text
	clean = reMySQLAnalyzeTime.ReplaceAllString(clean, "")
	clean = reMySQLAnalyzeCost.ReplaceAllString(clean, "")
	clean = strings.TrimSpace(clean)

	// Extract node type and relation
	// Example: "Filter: ((orders.status = 'active'))"
	// "Table scan on orders"
	// "Single-row index lookup on customers using PRIMARY (id=orders.customer_id)"
	if strings.HasPrefix(clean, "Table scan on ") {
		node.NodeType = "Table Scan"
		node.RelationName = strings.TrimPrefix(clean, "Table scan on ")
	} else if strings.HasPrefix(clean, "Filter: ") {
		node.NodeType = "Filter"
		node.Filter = strings.TrimPrefix(clean, "Filter: ")
	} else if strings.Contains(clean, " using ") {
		parts := strings.Split(clean, " using ")
		node.NodeType = strings.TrimSpace(parts[0])
		node.IndexName = strings.TrimSpace(parts[1])
	} else {
		node.NodeType = clean
	}

	return node
}
