package parser

import (
	"testing"
)

func TestParsePostgresJSON(t *testing.T) {
	rawJSON := `[
  {
    "Plan": {
      "Node Type": "Hash Join",
      "Parallel Aware": false,
      "Join Type": "Inner",
      "Startup Cost": 15.50,
      "Total Cost": 45.20,
      "Plan Rows": 25,
      "Plan Width": 64,
      "Actual Startup Time": 0.120,
      "Actual Total Time": 0.450,
      "Actual Rows": 30,
      "Actual Loops": 1,
      "Hash Cond": "(orders.user_id = users.id)",
      "Shared Hit Blocks": 42,
      "Shared Read Blocks": 5,
      "Plans": [
        {
          "Node Type": "Seq Scan",
          "Parent Relationship": "Outer",
          "Relation Name": "orders",
          "Alias": "orders",
          "Startup Cost": 0.00,
          "Total Cost": 20.00,
          "Plan Rows": 100,
          "Plan Width": 32,
          "Actual Startup Time": 0.010,
          "Actual Total Time": 0.150,
          "Actual Rows": 100,
          "Actual Loops": 1,
          "Filter": "(status = 'active'::text)",
          "Rows Removed by Filter": 20,
          "Shared Hit Blocks": 10,
          "Shared Read Blocks": 2
        },
        {
          "Node Type": "Hash",
          "Parent Relationship": "Inner",
          "Startup Cost": 10.00,
          "Total Cost": 10.00,
          "Plan Rows": 50,
          "Plan Width": 32,
          "Actual Startup Time": 0.080,
          "Actual Total Time": 0.080,
          "Actual Rows": 50,
          "Actual Loops": 1,
          "Shared Hit Blocks": 32,
          "Shared Read Blocks": 3,
          "Plans": [
            {
              "Node Type": "Seq Scan",
              "Parent Relationship": "Outer",
              "Relation Name": "users",
              "Alias": "users",
              "Startup Cost": 0.00,
              "Total Cost": 10.00,
              "Plan Rows": 50,
              "Plan Width": 32,
              "Actual Startup Time": 0.010,
              "Actual Total Time": 0.050,
              "Actual Rows": 50,
              "Actual Loops": 1,
              "Shared Hit Blocks": 32,
              "Shared Read Blocks": 3
            }
          ]
        }
      ]
    },
    "Planning Time": 0.180,
    "Execution Time": 0.520
  }
]`

	plan, err := ParsePostgresJSON([]byte(rawJSON))
	if err != nil {
		t.Fatalf("unexpected error parsing postgres JSON: %v", err)
	}

	if plan.Engine != EnginePostgres {
		t.Errorf("expected engine %s, got %s", EnginePostgres, plan.Engine)
	}
	if plan.Root.NodeType != "Hash Join" {
		t.Errorf("expected root node Hash Join, got %s", plan.Root.NodeType)
	}
	if len(plan.Root.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(plan.Root.Children))
	}
	if plan.ExecutionTimeMs != 0.520 {
		t.Errorf("expected execution time 0.520, got %f", plan.ExecutionTimeMs)
	}
	if plan.Summary.TotalNodes != 4 {
		t.Errorf("expected 4 total nodes, got %d", plan.Summary.TotalNodes)
	}
	if plan.Summary.BufferHitRatio <= 0 {
		t.Errorf("expected positive buffer hit ratio, got %f", plan.Summary.BufferHitRatio)
	}
}

func TestParsePostgresText(t *testing.T) {
	rawText := ` Hash Join  (cost=15.50..45.20 rows=25 width=64) (actual time=0.120..0.450 rows=30 loops=1)
   Hash Cond: (orders.user_id = users.id)
   Buffers: shared hit=42 read=5
   ->  Seq Scan on orders  (cost=0.00..20.00 rows=100 width=32) (actual time=0.010..0.150 rows=100 loops=1)
         Filter: (status = 'active'::text)
         Rows Removed by Filter: 20
         Buffers: shared hit=10 read=2
   ->  Hash  (cost=10.00..10.00 rows=50 width=32) (actual time=0.080..0.080 rows=50 loops=1)
         Buckets: 1024  Batches: 1  Memory Usage: 12kB
         Buffers: shared hit=32 read=3
         ->  Seq Scan on users  (cost=0.00..10.00 rows=50 width=32) (actual time=0.010..0.050 rows=50 loops=1)
               Buffers: shared hit=32 read=3
 Planning Time: 0.180 ms
 Execution Time: 0.520 ms`

	plan, err := ParsePostgresText(rawText)
	if err != nil {
		t.Fatalf("unexpected error parsing postgres text: %v", err)
	}

	if plan.Root.NodeType != "Hash Join" {
		t.Errorf("expected root Hash Join, got %s", plan.Root.NodeType)
	}
	if plan.PlanningTimeMs != 0.180 {
		t.Errorf("expected planning time 0.180, got %f", plan.PlanningTimeMs)
	}
	if plan.ExecutionTimeMs != 0.520 {
		t.Errorf("expected execution time 0.520, got %f", plan.ExecutionTimeMs)
	}
}

func TestParseMySQLJSON(t *testing.T) {
	rawMySQL := `{
  "query_block": {
    "select_id": 1,
    "cost_info": {
      "query_cost": "45.00"
    },
    "ordering_operation": {
      "using_filesort": true,
      "nested_loop": [
        {
          "table": {
            "table_name": "orders",
            "access_type": "ALL",
            "rows_examined_per_scan": 1000,
            "rows_produced_per_join": 100,
            "filtered": "10.00",
            "cost_info": {
              "prefix_cost": "25.00"
            },
            "attached_condition": "((orders.status = 'active') and (orders.created_at > '2026-01-01'))"
          }
        }
      ]
    }
  }
}`

	plan, err := ParseMySQLJSON([]byte(rawMySQL))
	if err != nil {
		t.Fatalf("unexpected error parsing MySQL JSON: %v", err)
	}

	if plan.Engine != EngineMySQL {
		t.Errorf("expected engine %s, got %s", EngineMySQL, plan.Engine)
	}
	if plan.Root.NodeType != "Sort (Filesort)" {
		t.Errorf("expected root Sort (Filesort), got %s", plan.Root.NodeType)
	}
}

func TestDetectAndParse(t *testing.T) {
	sample := `Seq Scan on products  (cost=0.00..15.50 rows=500 width=128) (actual time=0.050..1.200 rows=480 loops=1)
  Filter: (price > 100)
Planning Time: 0.05 ms
Execution Time: 1.25 ms`

	plan, err := DetectAndParse(sample)
	if err != nil {
		t.Fatalf("unexpected error in DetectAndParse: %v", err)
	}
	if plan.Root.RelationName != "products" {
		t.Errorf("expected products, got %s", plan.Root.RelationName)
	}
}
