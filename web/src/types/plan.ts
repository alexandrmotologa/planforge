export type DatabaseEngine = 'postgresql' | 'mysql' | 'unknown';
export type PlanFormat = 'json' | 'text';
export type FindingSeverity = 'info' | 'warning' | 'critical';

export interface BufferStats {
  shared_hit_blocks: number;
  shared_read_blocks: number;
  shared_dirtied_blocks: number;
  shared_written_blocks: number;
  local_hit_blocks: number;
  local_read_blocks: number;
  local_dirtied_blocks: number;
  local_written_blocks: number;
  temp_read_blocks: number;
  temp_written_blocks: number;
  io_read_time_ms: number;
  io_write_time_ms: number;
}

export interface WorkerDetail {
  worker_number: number;
  actual_startup_ms: number;
  actual_total_ms: number;
  actual_rows: number;
  actual_loops: number;
}

export interface WorkerStats {
  planned: number;
  launched: number;
  workers?: WorkerDetail[];
}

export interface PlanNode {
  id: string;
  node_type: string;
  relation_name?: string;
  schema?: string;
  alias?: string;
  index_name?: string;
  index_cond?: string;
  filter?: string;
  rows_removed_by_filter?: number;
  rows_removed_by_index_recheck?: number;
  hash_cond?: string;
  merge_cond?: string;
  join_type?: string;
  scan_direction?: string;
  sort_key?: string[];
  sort_method?: string;
  sort_space_used_kb?: number;
  sort_space_type?: string;
  strategy?: string;
  operation?: string;
  parent_relationship?: string;
  subplan_name?: string;
  cte?: string;
  output?: string[];

  startup_cost: number;
  total_cost: number;
  plan_rows: number;
  plan_width: number;

  actual_startup_time_ms: number;
  actual_total_time_ms: number;
  actual_rows: number;
  actual_loops: number;

  exclusive_duration_ms: number;
  duration_percent: number;
  cost_percent: number;
  estimation_ratio: number;
  is_estimation_anomaly: boolean;
  is_time_hog: boolean;

  buffers: BufferStats;
  workers: WorkerStats;
  warnings?: string[];
  children?: PlanNode[];
}

export interface TriggerInfo {
  trigger_name: string;
  relation: string;
  time_ms: number;
  calls: number;
}

export interface PlanSummary {
  total_cost: number;
  total_duration_ms: number;
  total_actual_rows: number;
  total_nodes: number;
  max_estimation_ratio: number;
  total_shared_hit_blocks: number;
  total_shared_read_blocks: number;
  buffer_hit_ratio: number;
  disk_read_bytes: number;
  time_hog_node_id?: string;
  time_hog_node_type?: string;
  time_hog_duration_ms?: number;
  time_hog_percent?: number;
}

export interface Plan {
  id: string;
  title?: string;
  engine: DatabaseEngine;
  format: PlanFormat;
  root: PlanNode;
  planning_time_ms: number;
  execution_time_ms: number;
  total_duration_ms: number;
  settings?: Record<string, string>;
  trigger_details?: TriggerInfo[];
  raw_input?: string;
  summary: PlanSummary;
}

export interface TimeHogFinding {
  node_id: string;
  node_type: string;
  relation?: string;
  duration_ms: number;
  duration_percent: number;
  severity: FindingSeverity;
  message: string;
  recommendation: string;
}

export interface EstimationFinding {
  node_id: string;
  node_type: string;
  relation?: string;
  planned_rows: number;
  actual_rows: number;
  ratio: number;
  severity: FindingSeverity;
  message: string;
  recommendation: string;
}

export interface BufferEfficiencyFinding {
  hit_ratio_percent: number;
  shared_hit_blocks: number;
  shared_read_blocks: number;
  disk_read_bytes: number;
  disk_read_mb: number;
  severity: FindingSeverity;
  message: string;
  recommendation: string;
}

export interface MemorySpillFinding {
  node_id: string;
  node_type: string;
  spill_type: string;
  space_used_kb: number;
  severity: FindingSeverity;
  message: string;
  recommendation: string;
}

export interface IndexRecommendation {
  node_id: string;
  table: string;
  index_name: string;
  ddl_postgres: string;
  ddl_mysql: string;
  hypopg_sql?: string;
  columns: string[];
  filter_predicate: string;
  estimated_impact: string;
  reason: string;
  is_functional?: boolean;
}

export interface JITFinding {
  jit_time_ms: number;
  execution_time_ms: number;
  overhead_percent: number;
  severity: FindingSeverity;
  message: string;
  recommendation: string;
}

export interface AntiPatternFinding {
  node_id?: string;
  pattern_type: string;
  severity: FindingSeverity;
  message: string;
  recommendation: string;
}

export interface FullReport {
  plan: Plan;
  health_score: number;
  health_grade: string;
  time_hogs: TimeHogFinding[];
  estimation_anomalies: EstimationFinding[];
  buffer_efficiency: BufferEfficiencyFinding;
  memory_spills: MemorySpillFinding[];
  jit_finding?: JITFinding;
  anti_patterns?: AntiPatternFinding[];
  index_recommendations: IndexRecommendation[];
  quick_insights: string[];
}

export interface PlanComparison {
  baseline_id: string;
  optimized_id: string;
  duration_delta_ms: number;
  duration_delta_percent: number;
  cost_delta: number;
  cost_delta_percent: number;
  buffer_read_delta: number;
  buffer_hit_delta: number;
  disk_bytes_read_delta: number;
  summary: string;
}

export interface CompareResponse {
  comparison: PlanComparison;
  baseline_report: FullReport;
  optimized_report: FullReport;
}

export interface SamplePlan {
  id: string;
  title: string;
  engine: string;
  description: string;
  raw_content: string;
}
