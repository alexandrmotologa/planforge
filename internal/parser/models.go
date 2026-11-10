package parser

import (
	"fmt"
	"strings"
)

// DatabaseEngine defines the origin of the execution plan.
type DatabaseEngine string

const (
	EnginePostgres DatabaseEngine = "postgresql"
	EngineMySQL    DatabaseEngine = "mysql"
	EngineUnknown  DatabaseEngine = "unknown"
)

// PlanFormat defines whether the plan was parsed from JSON or raw text.
type PlanFormat string

const (
	FormatJSON PlanFormat = "json"
	FormatText PlanFormat = "text"
)

// Plan represents a complete parsed query execution plan.
type Plan struct {
	ID               string            `json:"id"`
	Title            string            `json:"title,omitempty"`
	Engine           DatabaseEngine    `json:"engine"`
	Format           PlanFormat        `json:"format"`
	Root             *PlanNode         `json:"root"`
	PlanningTimeMs   float64           `json:"planning_time_ms"`
	ExecutionTimeMs  float64           `json:"execution_time_ms"`
	TotalDurationMs  float64           `json:"total_duration_ms"`
	Settings         map[string]string `json:"settings,omitempty"`
	TriggerDetails   []TriggerInfo     `json:"trigger_details,omitempty"`
	RawInput         string            `json:"raw_input,omitempty"`
	Summary          PlanSummary       `json:"summary"`
}

// TriggerInfo stores execution timing for database triggers invoked during execution.
type TriggerInfo struct {
	TriggerName string  `json:"trigger_name"`
	Relation    string  `json:"relation"`
	TimeMs      float64 `json:"time_ms"`
	Calls       int64   `json:"calls"`
}

// PlanNode represents a single operation in the execution plan tree.
type PlanNode struct {
	ID                          string         `json:"id"`
	NodeType                    string         `json:"node_type"`
	RelationName                string         `json:"relation_name,omitempty"`
	Schema                      string         `json:"schema,omitempty"`
	Alias                       string         `json:"alias,omitempty"`
	IndexName                   string         `json:"index_name,omitempty"`
	IndexCond                   string         `json:"index_cond,omitempty"`
	Filter                      string         `json:"filter,omitempty"`
	RowsRemovedByFilter         int64          `json:"rows_removed_by_filter,omitempty"`
	RowsRemovedByIndexRecheck   int64          `json:"rows_removed_by_index_recheck,omitempty"`
	HashCond                    string         `json:"hash_cond,omitempty"`
	MergeCond                   string         `json:"merge_cond,omitempty"`
	JoinType                    string         `json:"join_type,omitempty"`
	ScanDirection               string         `json:"scan_direction,omitempty"`
	SortKey                     []string       `json:"sort_key,omitempty"`
	SortMethod                  string         `json:"sort_method,omitempty"`
	SortSpaceUsedKb             int64          `json:"sort_space_used_kb,omitempty"`
	SortSpaceType               string         `json:"sort_space_type,omitempty"`
	Strategy                    string         `json:"strategy,omitempty"`
	Operation                   string         `json:"operation,omitempty"`
	ParentRelationship          string         `json:"parent_relationship,omitempty"`
	SubplanName                 string         `json:"subplan_name,omitempty"`
	CTE                         string         `json:"cte,omitempty"`
	Output                      []string       `json:"output,omitempty"`

	// Cost metrics
	StartupCost                 float64        `json:"startup_cost"`
	TotalCost                   float64        `json:"total_cost"`
	PlanRows                    float64        `json:"plan_rows"`
	PlanWidth                   int64          `json:"plan_width"`

	// Execution metrics (when ANALYZE was executed)
	ActualStartupTimeMs         float64        `json:"actual_startup_time_ms"`
	ActualTotalTimeMs           float64        `json:"actual_total_time_ms"`
	ActualRows                  float64        `json:"actual_rows"`
	ActualLoops                 int64          `json:"actual_loops"`

	// Derived metrics
	ExclusiveDurationMs         float64        `json:"exclusive_duration_ms"`
	DurationPercent             float64        `json:"duration_percent"`
	CostPercent                 float64        `json:"cost_percent"`
	EstimationRatio             float64        `json:"estimation_ratio"`
	IsEstimationAnomaly         bool           `json:"is_estimation_anomaly"`
	IsTimeHog                   bool           `json:"is_time_hog"`

	// Buffer, I/O and Worker metrics
	Buffers                     BufferStats    `json:"buffers"`
	Workers                     WorkerStats    `json:"workers"`

	Warnings                    []string       `json:"warnings,omitempty"`
	Children                    []*PlanNode    `json:"children,omitempty"`
}

// BufferStats captures cache hit, read, and write activity.
type BufferStats struct {
	SharedHitBlocks     int64   `json:"shared_hit_blocks"`
	SharedReadBlocks    int64   `json:"shared_read_blocks"`
	SharedDirtiedBlocks int64   `json:"shared_dirtied_blocks"`
	SharedWrittenBlocks int64   `json:"shared_written_blocks"`
	LocalHitBlocks      int64   `json:"local_hit_blocks"`
	LocalReadBlocks     int64   `json:"local_read_blocks"`
	LocalDirtiedBlocks  int64   `json:"local_dirtied_blocks"`
	LocalWrittenBlocks  int64   `json:"local_written_blocks"`
	TempReadBlocks      int64   `json:"temp_read_blocks"`
	TempWrittenBlocks   int64   `json:"temp_written_blocks"`
	IOReadTimeMs        float64 `json:"io_read_time_ms"`
	IOWriteTimeMs       float64 `json:"io_write_time_ms"`
}

// TotalHitBlocks returns sum of shared and local cache hits.
func (b BufferStats) TotalHitBlocks() int64 {
	return b.SharedHitBlocks + b.LocalHitBlocks
}

// TotalReadBlocks returns sum of shared and local disk reads.
func (b BufferStats) TotalReadBlocks() int64 {
	return b.SharedReadBlocks + b.LocalReadBlocks
}

// TotalBlocks returns all blocks accessed.
func (b BufferStats) TotalBlocks() int64 {
	return b.TotalHitBlocks() + b.TotalReadBlocks()
}

// HitRatio calculates the buffer hit ratio as a percentage.
func (b BufferStats) HitRatio() float64 {
	total := b.TotalBlocks()
	if total == 0 {
		return 100.0
	}
	return float64(b.TotalHitBlocks()) / float64(total) * 100.0
}

// DiskBytesRead returns total bytes read from disk assuming 8KB Postgres blocks.
func (b BufferStats) DiskBytesRead() int64 {
	return b.TotalReadBlocks() * 8192
}

// WorkerStats records parallel query execution details.
type WorkerStats struct {
	Planned int            `json:"planned"`
	Launched int           `json:"launched"`
	Workers []WorkerDetail `json:"workers,omitempty"`
}

// WorkerDetail records individual thread or worker timings.
type WorkerDetail struct {
	WorkerNumber int     `json:"worker_number"`
	ActualStartup float64 `json:"actual_startup_ms"`
	ActualTotal   float64 `json:"actual_total_ms"`
	ActualRows    float64 `json:"actual_rows"`
	ActualLoops   int64   `json:"actual_loops"`
}

// PlanSummary aggregates query-wide metrics.
type PlanSummary struct {
	TotalCost            float64 `json:"total_cost"`
	TotalDurationMs      float64 `json:"total_duration_ms"`
	TotalActualRows      float64 `json:"total_actual_rows"`
	TotalNodes           int     `json:"total_nodes"`
	MaxEstimationRatio   float64 `json:"max_estimation_ratio"`
	TotalSharedHitBlocks int64   `json:"total_shared_hit_blocks"`
	TotalSharedReadBlocks int64  `json:"total_shared_read_blocks"`
	BufferHitRatio       float64 `json:"buffer_hit_ratio"`
	DiskReadBytes        int64   `json:"disk_read_bytes"`
	TimeHogNodeID        string  `json:"time_hog_node_id,omitempty"`
	TimeHogNodeType      string  `json:"time_hog_node_type,omitempty"`
	TimeHogDurationMs    float64 `json:"time_hog_duration_ms,omitempty"`
	TimeHogPercent       float64 `json:"time_hog_percent,omitempty"`
}

// FlattenNodes returns all nodes in depth-first traversal order.
func (p *Plan) FlattenNodes() []*PlanNode {
	if p.Root == nil {
		return nil
	}
	var nodes []*PlanNode
	var walk func(node *PlanNode)
	walk = func(node *PlanNode) {
		if node == nil {
			return
		}
		nodes = append(nodes, node)
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(p.Root)
	return nodes
}

// DisplayLabel provides a readable description for graph badges.
func (n *PlanNode) DisplayLabel() string {
	if n.RelationName != "" {
		if n.Alias != "" && n.Alias != n.RelationName {
			return fmt.Sprintf("%s (%s as %s)", n.NodeType, n.RelationName, n.Alias)
		}
		return fmt.Sprintf("%s on %s", n.NodeType, n.RelationName)
	}
	if n.IndexName != "" {
		return fmt.Sprintf("%s using %s", n.NodeType, n.IndexName)
	}
	return n.NodeType
}

// CleanFilter removes redundant enclosing parentheses and typecasts.
func CleanFilter(filter string) string {
	res := strings.TrimSpace(filter)
	// Remove outer parens if wrapped
	for strings.HasPrefix(res, "(") && strings.HasSuffix(res, ")") {
		inner := res[1 : len(res)-1]
		// check if balanced
		depth := 0
		balanced := true
		for _, ch := range inner {
			if ch == '(' {
				depth++
			} else if ch == ')' {
				depth--
				if depth < 0 {
					balanced = false
					break
				}
			}
		}
		if balanced && depth == 0 {
			res = strings.TrimSpace(inner)
		} else {
			break
		}
	}
	return res
}
