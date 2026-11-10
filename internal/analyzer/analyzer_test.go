package analyzer_test

import (
	"testing"

	"github.com/alexandrmotologa/planforge/internal/advisor"
	"github.com/alexandrmotologa/planforge/internal/analyzer"
	"github.com/alexandrmotologa/planforge/internal/anonymizer"
	"github.com/alexandrmotologa/planforge/internal/parser"
)

func TestAnalysisAndAdvisor(t *testing.T) {
	planJSON := `[
  {
    "Plan": {
      "Node Type": "Seq Scan",
      "Relation Name": "orders",
      "Startup Cost": 0.00,
      "Total Cost": 500.00,
      "Plan Rows": 10,
      "Plan Width": 64,
      "Actual Startup Time": 0.100,
      "Actual Total Time": 50.000,
      "Actual Rows": 5000,
      "Actual Loops": 1,
      "Filter": "((tenant_id = 42) AND (status = 'active'::text) AND (created_at > '2026-01-01'))",
      "Rows Removed by Filter": 45000,
      "Shared Hit Blocks": 200,
      "Shared Read Blocks": 800
    },
    "Execution Time": 50.000
  }
]`

	plan, err := parser.ParsePostgresJSON([]byte(planJSON))
	if err != nil {
		t.Fatalf("failed to parse plan: %v", err)
	}

	indexRecs := advisor.RecommendIndexes(plan)
	if len(indexRecs) == 0 {
		t.Fatalf("expected missing index recommendation for Seq Scan")
	}

	rec := indexRecs[0]
	if rec.Table != "orders" {
		t.Errorf("expected orders table, got %s", rec.Table)
	}
	// Verify equality columns come first
	if len(rec.Columns) < 3 {
		t.Errorf("expected 3 columns (tenant_id, status, created_at), got %v", rec.Columns)
	}

	report := analyzer.EvaluatePlan(plan, indexRecs)
	if report.HealthScore >= 90 {
		t.Errorf("expected low health score due to Seq Scan, 80%% disk reads and 500x estimation anomaly, got %d", report.HealthScore)
	}
	if len(report.TimeHogs) == 0 {
		t.Errorf("expected time hog finding")
	}
	if len(report.EstimationAnomalies) == 0 {
		t.Errorf("expected estimation anomaly finding")
	}
	if report.BufferEfficiency.SharedReadBlocks != 800 {
		t.Errorf("expected 800 shared read blocks, got %d", report.BufferEfficiency.SharedReadBlocks)
	}

	// Test Sanitizer
	sanitizer := anonymizer.NewSanitizer()
	scrubbed := sanitizer.AnonymizePlan(plan)
	if scrubbed.Root.RelationName == "orders" {
		t.Errorf("expected anonymized table name, got %s", scrubbed.Root.RelationName)
	}
}
