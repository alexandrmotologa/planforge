package parser

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// pgWrapper represents the top-level Postgres EXPLAIN JSON output.
type pgWrapper struct {
	Plan           pgNode            `json:"Plan"`
	PlanningTime   float64           `json:"Planning Time"`
	ExecutionTime  float64           `json:"Execution Time"`
	Triggers       []pgTrigger       `json:"Triggers"`
	Settings       map[string]string `json:"Settings"`
	JIT            map[string]any    `json:"JIT"`
}

type pgTrigger struct {
	TriggerName string  `json:"Trigger Name"`
	Relation    string  `json:"Relation"`
	Time        float64 `json:"Time"`
	Calls       string  `json:"Calls"`
}

type pgNode struct {
	NodeType                  string      `json:"Node Type"`
	RelationName              string      `json:"Relation Name"`
	Schema                    string      `json:"Schema"`
	Alias                     string      `json:"Alias"`
	IndexName                 string      `json:"Index Name"`
	IndexCond                 string      `json:"Index Cond"`
	Filter                    string      `json:"Filter"`
	RowsRemovedByFilter       int64       `json:"Rows Removed by Filter"`
	RowsRemovedByIndexRecheck int64       `json:"Rows Removed by Index Recheck"`
	HashCond                  string      `json:"Hash Cond"`
	MergeCond                 string      `json:"Merge Cond"`
	JoinType                  string      `json:"Join Type"`
	ScanDirection             string      `json:"Scan Direction"`
	SortKey                   any         `json:"Sort Key"`
	SortMethod                string      `json:"Sort Method"`
	SortSpaceUsed             int64       `json:"Sort Space Used"`
	SortSpaceType             string      `json:"Sort Space Type"`
	Strategy                  string      `json:"Strategy"`
	Operation                 string      `json:"Operation"`
	ParentRelationship        string      `json:"Parent Relationship"`
	SubplanName               string      `json:"Subplan Name"`
	CTEName                   string      `json:"CTE Name"`
	Output                    []string    `json:"Output"`

	StartupCost               float64     `json:"Startup Cost"`
	TotalCost                 float64     `json:"Total Cost"`
	PlanRows                  float64     `json:"Plan Rows"`
	PlanWidth                 int64       `json:"Plan Width"`

	ActualStartupTime         float64     `json:"Actual Startup Time"`
	ActualTotalTime           float64     `json:"Actual Total Time"`
	ActualRows                float64     `json:"Actual Rows"`
	ActualLoops               int64       `json:"Actual Loops"`

	SharedHitBlocks           int64       `json:"Shared Hit Blocks"`
	SharedReadBlocks          int64       `json:"Shared Read Blocks"`
	SharedDirtiedBlocks       int64       `json:"Shared Dirtied Blocks"`
	SharedWrittenBlocks       int64       `json:"Shared Written Blocks"`
	LocalHitBlocks            int64       `json:"Local Hit Blocks"`
	LocalReadBlocks           int64       `json:"Local Read Blocks"`
	LocalDirtiedBlocks        int64       `json:"Local Dirtied Blocks"`
	LocalWrittenBlocks        int64       `json:"Local Written Blocks"`
	TempReadBlocks            int64       `json:"Temp Read Blocks"`
	TempWrittenBlocks         int64       `json:"Temp Written Blocks"`
	IOReadTime                float64     `json:"I/O Read Time"`
	IOWriteTime               float64     `json:"I/O Write Time"`

	WorkersPlanned            int         `json:"Workers Planned"`
	WorkersLaunched           int         `json:"Workers Launched"`
	Workers                   []pgWorker  `json:"Workers"`

	Plans                     []pgNode    `json:"Plans"`
}

type pgWorker struct {
	WorkerNumber      int     `json:"Worker Number"`
	ActualStartupTime float64 `json:"Actual Startup Time"`
	ActualTotalTime   float64 `json:"Actual Total Time"`
	ActualRows        float64 `json:"Actual Rows"`
	ActualLoops       int64   `json:"Actual Loops"`
}

// ParsePostgresJSON converts raw PostgreSQL EXPLAIN JSON into a normalized Plan structure.
func ParsePostgresJSON(data []byte) (*Plan, error) {
	var wrappers []pgWrapper

	// Postgres JSON typically returns an array of objects `[ { "Plan": { ... } } ]`
	trimmed := strings.TrimSpace(string(data))
	if strings.HasPrefix(trimmed, "[") {
		if err := json.Unmarshal([]byte(trimmed), &wrappers); err != nil {
			return nil, fmt.Errorf("failed to parse postgres JSON array: %w", err)
		}
	} else if strings.HasPrefix(trimmed, "{") {
		// Single object might be `{ "Plan": { ... } }` or just `{ "Node Type": ... }`
		var single pgWrapper
		if err := json.Unmarshal([]byte(trimmed), &single); err == nil && single.Plan.NodeType != "" {
			wrappers = []pgWrapper{single}
		} else {
			// Direct plan node
			var direct pgNode
			if err := json.Unmarshal([]byte(trimmed), &direct); err == nil && direct.NodeType != "" {
				wrappers = []pgWrapper{{Plan: direct}}
			} else {
				return nil, fmt.Errorf("unrecognized postgres JSON object structure")
			}
		}
	} else {
		return nil, fmt.Errorf("input is not valid JSON")
	}

	if len(wrappers) == 0 || wrappers[0].Plan.NodeType == "" {
		return nil, fmt.Errorf("no execution plan found in JSON")
	}

	wrapper := wrappers[0]
	nodeCounter := 0
	root := convertPgNode(&wrapper.Plan, &nodeCounter)

	totalDuration := wrapper.ExecutionTime
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
		ID:              fmt.Sprintf("plan-%d", nodeCounter),
		Engine:          EnginePostgres,
		Format:          FormatJSON,
		Root:            root,
		PlanningTimeMs:  wrapper.PlanningTime,
		ExecutionTimeMs: wrapper.ExecutionTime,
		TotalDurationMs: totalDuration,
		Settings:        wrapper.Settings,
		RawInput:        trimmed,
	}

	for _, trig := range wrapper.Triggers {
		var calls int64
		fmt.Sscanf(trig.Calls, "%d", &calls)
		plan.TriggerDetails = append(plan.TriggerDetails, TriggerInfo{
			TriggerName: trig.TriggerName,
			Relation:    trig.Relation,
			TimeMs:      trig.Time,
			Calls:       calls,
		})
	}

	// Calculate timing metrics and summary
	computeMetrics(plan)

	return plan, nil
}

func convertPgNode(raw *pgNode, counter *int) *PlanNode {
	*counter++
	id := fmt.Sprintf("node-%d", *counter)

	loops := raw.ActualLoops
	if loops <= 0 {
		loops = 1
	}

	node := &PlanNode{
		ID:                        id,
		NodeType:                  raw.NodeType,
		RelationName:              raw.RelationName,
		Schema:                    raw.Schema,
		Alias:                     raw.Alias,
		IndexName:                 raw.IndexName,
		IndexCond:                 raw.IndexCond,
		Filter:                    raw.Filter,
		RowsRemovedByFilter:       raw.RowsRemovedByFilter,
		RowsRemovedByIndexRecheck: raw.RowsRemovedByIndexRecheck,
		HashCond:                  raw.HashCond,
		MergeCond:                 raw.MergeCond,
		JoinType:                  raw.JoinType,
		ScanDirection:             raw.ScanDirection,
		SortMethod:                raw.SortMethod,
		SortSpaceUsedKb:           raw.SortSpaceUsed,
		SortSpaceType:             raw.SortSpaceType,
		Strategy:                  raw.Strategy,
		Operation:                 raw.Operation,
		ParentRelationship:        raw.ParentRelationship,
		SubplanName:               raw.SubplanName,
		CTE:                       raw.CTEName,
		Output:                    raw.Output,
		StartupCost:               raw.StartupCost,
		TotalCost:                 raw.TotalCost,
		PlanRows:                  raw.PlanRows,
		PlanWidth:                 raw.PlanWidth,
		ActualStartupTimeMs:       raw.ActualStartupTime,
		ActualTotalTimeMs:         raw.ActualTotalTime,
		ActualRows:                raw.ActualRows,
		ActualLoops:               loops,
		Buffers: BufferStats{
			SharedHitBlocks:     raw.SharedHitBlocks,
			SharedReadBlocks:    raw.SharedReadBlocks,
			SharedDirtiedBlocks: raw.SharedDirtiedBlocks,
			SharedWrittenBlocks: raw.SharedWrittenBlocks,
			LocalHitBlocks:      raw.LocalHitBlocks,
			LocalReadBlocks:     raw.LocalReadBlocks,
			LocalDirtiedBlocks:  raw.LocalDirtiedBlocks,
			LocalWrittenBlocks:  raw.LocalWrittenBlocks,
			TempReadBlocks:      raw.TempReadBlocks,
			TempWrittenBlocks:   raw.TempWrittenBlocks,
			IOReadTimeMs:        raw.IOReadTime,
			IOWriteTimeMs:       raw.IOWriteTime,
		},
		Workers: WorkerStats{
			Planned:  raw.WorkersPlanned,
			Launched: raw.WorkersLaunched,
		},
	}

	// Parse SortKey (can be string or []string)
	if raw.SortKey != nil {
		switch v := raw.SortKey.(type) {
		case string:
			node.SortKey = []string{v}
		case []any:
			for _, item := range v {
				if s, ok := item.(string); ok {
					node.SortKey = append(node.SortKey, s)
				}
			}
		case []string:
			node.SortKey = v
		}
	}

	// Workers details
	for _, w := range raw.Workers {
		node.Workers.Workers = append(node.Workers.Workers, WorkerDetail{
			WorkerNumber:  w.WorkerNumber,
			ActualStartup: w.ActualStartupTime,
			ActualTotal:   w.ActualTotalTime,
			ActualRows:    w.ActualRows,
			ActualLoops:   w.ActualLoops,
		})
	}

	// Children nodes
	for i := range raw.Plans {
		child := convertPgNode(&raw.Plans[i], counter)
		node.Children = append(node.Children, child)
	}

	return node
}

// computeMetrics calculates exclusive durations, estimation ratios, and summary metrics.
func computeMetrics(plan *Plan) {
	if plan.Root == nil {
		return
	}

	var totalSharedHit int64
	var totalSharedRead int64
	var totalActualRows float64
	totalNodes := 0
	maxEstRatio := 1.0

	var calcNode func(n *PlanNode)
	calcNode = func(n *PlanNode) {
		totalNodes++
		totalSharedHit += n.Buffers.SharedHitBlocks
		totalSharedRead += n.Buffers.SharedReadBlocks
		totalActualRows += n.ActualRows * float64(n.ActualLoops)

		// Compute exclusive duration
		nodeTotalTime := n.ActualTotalTimeMs * float64(n.ActualLoops)
		var childrenTotalTime float64
		for _, child := range n.Children {
			calcNode(child)
			childrenTotalTime += child.ActualTotalTimeMs * float64(child.ActualLoops)
		}

		if nodeTotalTime > 0 {
			exclusive := nodeTotalTime - childrenTotalTime
			if exclusive < 0 {
				exclusive = 0
			}
			n.ExclusiveDurationMs = math.Round(exclusive*1000) / 1000
		} else {
			// If timing not present, compute from cost
			var childrenCost float64
			for _, child := range n.Children {
				childrenCost += child.TotalCost
			}
			exclusiveCost := n.TotalCost - childrenCost
			if exclusiveCost < 0 {
				exclusiveCost = 0
			}
			n.ExclusiveDurationMs = math.Round(exclusiveCost*100) / 100
		}

		// Duration percentage
		if plan.TotalDurationMs > 0 {
			n.DurationPercent = math.Round((n.ExclusiveDurationMs/plan.TotalDurationMs)*10000) / 100
		}
		if plan.Root.TotalCost > 0 {
			n.CostPercent = math.Round((n.TotalCost/plan.Root.TotalCost)*10000) / 100
		}

		// Estimation Ratio
		if n.PlanRows > 0 {
			n.EstimationRatio = math.Round((n.ActualRows/n.PlanRows)*100) / 100
			ratioDiff := n.EstimationRatio
			if ratioDiff < 1.0 && ratioDiff > 0 {
				ratioDiff = 1.0 / ratioDiff
			}
			if ratioDiff > maxEstRatio {
				maxEstRatio = ratioDiff
			}
			if ratioDiff >= 10.0 {
				n.IsEstimationAnomaly = true
			}
		}

		// Time hog flag (> 50% exclusive time)
		if n.DurationPercent >= 50.0 {
			n.IsTimeHog = true
		}
	}

	calcNode(plan.Root)

	// Summary calculation
	totalBlocks := totalSharedHit + totalSharedRead
	var hitRatio float64 = 100.0
	if totalBlocks > 0 {
		hitRatio = math.Round((float64(totalSharedHit)/float64(totalBlocks))*10000) / 100
	}

	var timeHogNode *PlanNode
	nodes := plan.FlattenNodes()
	for _, n := range nodes {
		if timeHogNode == nil || n.ExclusiveDurationMs > timeHogNode.ExclusiveDurationMs {
			timeHogNode = n
		}
	}

	summary := PlanSummary{
		TotalCost:             plan.Root.TotalCost,
		TotalDurationMs:       plan.TotalDurationMs,
		TotalActualRows:       totalActualRows,
		TotalNodes:            totalNodes,
		MaxEstimationRatio:    maxEstRatio,
		TotalSharedHitBlocks:  totalSharedHit,
		TotalSharedReadBlocks: totalSharedRead,
		BufferHitRatio:        hitRatio,
		DiskReadBytes:         totalSharedRead * 8192,
	}

	if timeHogNode != nil && timeHogNode.ExclusiveDurationMs > 0 {
		summary.TimeHogNodeID = timeHogNode.ID
		summary.TimeHogNodeType = timeHogNode.NodeType
		summary.TimeHogDurationMs = timeHogNode.ExclusiveDurationMs
		summary.TimeHogPercent = timeHogNode.DurationPercent
	}

	plan.Summary = summary
}
