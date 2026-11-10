package api

// SamplePlan represents a preset query execution plan for instant testing and demonstration.
type SamplePlan struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Engine      string `json:"engine"`
	Description string `json:"description"`
	RawContent  string `json:"raw_content"`
}

// GetSamplePlans returns built-in benchmark query plans.
func GetSamplePlans() []SamplePlan {
	return []SamplePlan{
		{
			ID:          "pg_slow_seq_scan",
			Title:       "PostgreSQL E-Commerce Unindexed Seq Scan",
			Engine:      "postgresql",
			Description: "Slow sequential scan on 100,000 orders filtering by tenant_id, status and created_at with high disk reads and stale statistics.",
			RawContent: `[
  {
    "Plan": {
      "Node Type": "Seq Scan",
      "Parallel Aware": false,
      "Relation Name": "orders",
      "Schema": "public",
      "Alias": "o",
      "Startup Cost": 0.00,
      "Total Cost": 3482.00,
      "Plan Rows": 12,
      "Plan Width": 142,
      "Actual Startup Time": 0.150,
      "Actual Total Time": 148.500,
      "Actual Rows": 14500,
      "Actual Loops": 1,
      "Filter": "((tenant_id = 42) AND (status = 'active'::text) AND (created_at > '2026-01-01 00:00:00'::timestamp))",
      "Rows Removed by Filter": 85500,
      "Shared Hit Blocks": 310,
      "Shared Read Blocks": 1940
    },
    "Planning Time": 0.245,
    "Execution Time": 149.200
  }
]`,
		},
		{
			ID:          "pg_nested_loop_join",
			Title:       "PostgreSQL Cartesian Nested Loop Join",
			Engine:      "postgresql",
			Description: "High latency Nested Loop Join scanning users and repeated lookups on payments, leading to severe duration hog.",
			RawContent: `[
  {
    "Plan": {
      "Node Type": "Nested Loop",
      "Parallel Aware": false,
      "Join Type": "Inner",
      "Startup Cost": 0.00,
      "Total Cost": 8420.50,
      "Plan Rows": 500,
      "Plan Width": 210,
      "Actual Startup Time": 0.220,
      "Actual Total Time": 235.400,
      "Actual Rows": 520,
      "Actual Loops": 1,
      "Shared Hit Blocks": 1200,
      "Shared Read Blocks": 3500,
      "Plans": [
        {
          "Node Type": "Seq Scan",
          "Parent Relationship": "Outer",
          "Relation Name": "users",
          "Alias": "u",
          "Startup Cost": 0.00,
          "Total Cost": 45.00,
          "Plan Rows": 200,
          "Plan Width": 68,
          "Actual Startup Time": 0.015,
          "Actual Total Time": 1.200,
          "Actual Rows": 200,
          "Actual Loops": 1,
          "Filter": "(is_active IS TRUE)",
          "Shared Hit Blocks": 15,
          "Shared Read Blocks": 0
        },
        {
          "Node Type": "Seq Scan",
          "Parent Relationship": "Inner",
          "Relation Name": "payments",
          "Alias": "p",
          "Startup Cost": 0.00,
          "Total Cost": 41.50,
          "Plan Rows": 3,
          "Plan Width": 142,
          "Actual Startup Time": 0.010,
          "Actual Total Time": 1.150,
          "Actual Rows": 2.6,
          "Actual Loops": 200,
          "Filter": "((user_id = u.id) AND (amount > 100.00))",
          "Rows Removed by Filter": 450,
          "Shared Hit Blocks": 1185,
          "Shared Read Blocks": 3500
        }
      ]
    },
    "Planning Time": 0.380,
    "Execution Time": 236.100
  }
]`,
		},
		{
			ID:          "pg_workmem_disk_spill",
			Title:       "PostgreSQL External Merge Sort (Disk Spill)",
			Engine:      "postgresql",
			Description: "Large dataset sorting exceeding work_mem, forcing execution into an external merge sort on temporary disk files.",
			RawContent: `[
  {
    "Plan": {
      "Node Type": "Sort",
      "Startup Cost": 12500.00,
      "Total Cost": 13800.00,
      "Plan Rows": 80000,
      "Plan Width": 128,
      "Actual Startup Time": 85.000,
      "Actual Total Time": 142.300,
      "Actual Rows": 80000,
      "Actual Loops": 1,
      "Sort Key": ["customer_id", "created_at DESC"],
      "Sort Method": "external merge",
      "Sort Space Used": 14500,
      "Sort Space Type": "Disk",
      "Temp Read Blocks": 1812,
      "Temp Written Blocks": 1812,
      "Plans": [
        {
          "Node Type": "Seq Scan",
          "Relation Name": "invoices",
          "Startup Cost": 0.00,
          "Total Cost": 4500.00,
          "Plan Rows": 80000,
          "Plan Width": 128,
          "Actual Startup Time": 0.050,
          "Actual Total Time": 35.000,
          "Actual Rows": 80000,
          "Actual Loops": 1,
          "Shared Hit Blocks": 1200,
          "Shared Read Blocks": 450
        }
      ]
    },
    "Planning Time": 0.150,
    "Execution Time": 144.500
  }
]`,
		},
		{
			ID:          "mysql_orders_filesort",
			Title:       "MySQL 8.0 E-Commerce Orders with Filesort",
			Engine:      "mysql",
			Description: "MySQL query_block execution plan with nested loop join, unindexed table scan on orders, and ordering operation with filesort.",
			RawContent: `{
  "query_block": {
    "select_id": 1,
    "cost_info": {
      "query_cost": "185.50"
    },
    "ordering_operation": {
      "using_filesort": true,
      "nested_loop": [
        {
          "table": {
            "table_name": "orders",
            "access_type": "ALL",
            "rows_examined_per_scan": 5000,
            "rows_produced_per_join": 250,
            "filtered": "5.00",
            "cost_info": {
              "read_cost": "120.00",
              "eval_cost": "25.00",
              "prefix_cost": "145.00",
              "data_read_per_join": "64K"
            },
            "used_columns": ["id", "customer_id", "status", "created_at"],
            "attached_condition": "((orders.status = 'shipped') and (orders.created_at >= '2026-01-01'))"
          }
        },
        {
          "table": {
            "table_name": "customers",
            "access_type": "eq_ref",
            "possible_keys": ["PRIMARY"],
            "key": "PRIMARY",
            "used_key_parts": ["id"],
            "key_length": "8",
            "ref": ["orders.customer_id"],
            "rows_examined_per_scan": 1,
            "rows_produced_per_join": 250,
            "filtered": "100.00",
            "cost_info": {
              "read_cost": "35.00",
              "eval_cost": "5.50",
              "prefix_cost": "185.50",
              "data_read_per_join": "32K"
            },
            "used_columns": ["id", "email", "full_name"]
          }
        }
      ]
    }
  }
}`,
		},
	}
}
