package analyzer

import (
	"github.com/alexandrmotologa/planforge/internal/parser"
)

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// TimeHogFinding records nodes that consume a disproportionate share of query runtime.
type TimeHogFinding struct {
	NodeID          string   `json:"node_id"`
	NodeType        string   `json:"node_type"`
	Relation        string   `json:"relation,omitempty"`
	DurationMs      float64  `json:"duration_ms"`
	DurationPercent float64  `json:"duration_percent"`
	Severity        Severity `json:"severity"`
	Message         string   `json:"message"`
	Recommendation  string   `json:"recommendation"`
}

// EstimationFinding records significant divergence between planner row estimates and actual rows.
type EstimationFinding struct {
	NodeID          string   `json:"node_id"`
	NodeType        string   `json:"node_type"`
	Relation        string   `json:"relation,omitempty"`
	PlannedRows     float64  `json:"planned_rows"`
	ActualRows      float64  `json:"actual_rows"`
	Ratio           float64  `json:"ratio"`
	Severity        Severity `json:"severity"`
	Message         string   `json:"message"`
	Recommendation  string   `json:"recommendation"`
}

// BufferEfficiencyFinding evaluates RAM buffer cache hits versus physical disk reads.
type BufferEfficiencyFinding struct {
	HitRatioPercent float64  `json:"hit_ratio_percent"`
	SharedHitBlocks int64    `json:"shared_hit_blocks"`
	SharedReadBlocks int64   `json:"shared_read_blocks"`
	DiskReadBytes   int64    `json:"disk_read_bytes"`
	DiskReadMB      float64  `json:"disk_read_mb"`
	Severity        Severity `json:"severity"`
	Message         string   `json:"message"`
	Recommendation  string   `json:"recommendation"`
}

// MemorySpillFinding records memory allocations overflowing to disk.
type MemorySpillFinding struct {
	NodeID          string   `json:"node_id"`
	NodeType        string   `json:"node_type"`
	SpillType       string   `json:"spill_type"` // e.g. "Sort external merge disk", "Hash batches"
	SpaceUsedKb     int64    `json:"space_used_kb"`
	Severity        Severity `json:"severity"`
	Message         string   `json:"message"`
	Recommendation  string   `json:"recommendation"`
}

// IndexRecommendation holds synthesized DDL and explanation.
type IndexRecommendation struct {
	NodeID          string   `json:"node_id"`
	Table           string   `json:"table"`
	IndexName       string   `json:"index_name"`
	DDLPostgres     string   `json:"ddl_postgres"`
	DDLMySQL        string   `json:"ddl_mysql"`
	Columns         []string `json:"columns"`
	FilterPredicate string   `json:"filter_predicate"`
	EstimatedImpact string   `json:"estimated_impact"`
	Reason          string   `json:"reason"`
}

// FullReport contains the complete diagnostic evaluation of a query plan.
type FullReport struct {
	Plan                   *parser.Plan              `json:"plan"`
	HealthScore            int                       `json:"health_score"` // 0 to 100
	HealthGrade            string                    `json:"health_grade"` // A, B, C, D, F
	TimeHogs               []TimeHogFinding          `json:"time_hogs"`
	EstimationAnomalies    []EstimationFinding       `json:"estimation_anomalies"`
	BufferEfficiency       BufferEfficiencyFinding   `json:"buffer_efficiency"`
	MemorySpills           []MemorySpillFinding      `json:"memory_spills"`
	IndexRecommendations   []IndexRecommendation     `json:"index_recommendations"`
	QuickInsights          []string                  `json:"quick_insights"`
}
