package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/alexandrmotologa/planforge/internal/advisor"
	"github.com/alexandrmotologa/planforge/internal/analyzer"
	"github.com/alexandrmotologa/planforge/internal/anonymizer"
	"github.com/alexandrmotologa/planforge/internal/exporter"
	"github.com/alexandrmotologa/planforge/internal/parser"
)

type AnalyzeRequest struct {
	Plan      string `json:"plan"`
	Anonymize bool   `json:"anonymize"`
}

type CompareRequest struct {
	Baseline  string `json:"baseline"`
	Optimized string `json:"optimized"`
}

type CompareResponse struct {
	Comparison      analyzer.PlanComparison `json:"comparison"`
	BaselineReport  *analyzer.FullReport    `json:"baseline_report"`
	OptimizedReport *analyzer.FullReport    `json:"optimized_report"`
}

// HandleAnalyze handles plan ingestion, diagnostic analysis, and index synthesis.
func HandleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var planText string
	var anonymizeReq bool

	// Check if JSON wrapper
	var req AnalyzeRequest
	if err := json.Unmarshal(body, &req); err == nil && req.Plan != "" {
		planText = req.Plan
		anonymizeReq = req.Anonymize
	} else {
		// Raw string sent directly
		planText = string(body)
	}

	plan, err := parser.DetectAndParse(planText)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if anonymizeReq {
		s := anonymizer.NewSanitizer()
		plan = s.AnonymizePlan(plan)
	}

	indexRecs := advisor.RecommendIndexes(plan)
	report := analyzer.EvaluatePlan(plan, indexRecs)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// HandleCompare compares two plans and returns delta metrics.
func HandleCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CompareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	basePlan, err := parser.DetectAndParse(req.Baseline)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Baseline plan error: " + err.Error()})
		return
	}

	optPlan, err := parser.DetectAndParse(req.Optimized)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Optimized plan error: " + err.Error()})
		return
	}

	baseRecs := advisor.RecommendIndexes(basePlan)
	baseReport := analyzer.EvaluatePlan(basePlan, baseRecs)

	optRecs := advisor.RecommendIndexes(optPlan)
	optReport := analyzer.EvaluatePlan(optPlan, optRecs)

	comparison := analyzer.Compare(basePlan, optPlan)

	resp := CompareResponse{
		Comparison:      comparison,
		BaselineReport:  baseReport,
		OptimizedReport: optReport,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleExport exports a standalone self-contained offline HTML file.
func HandleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	plan, err := parser.DetectAndParse(req.Plan)
	if err != nil {
		http.Error(w, "Parse error: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Anonymize {
		s := anonymizer.NewSanitizer()
		plan = s.AnonymizePlan(plan)
	}

	indexRecs := advisor.RecommendIndexes(plan)
	report := analyzer.EvaluatePlan(plan, indexRecs)

	htmlContent, err := exporter.GenerateStandaloneHTML(report)
	if err != nil {
		http.Error(w, "Export error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"planforge-report.html\"")
	w.Write([]byte(htmlContent))
}

// HandleSamples returns list of sample query plans.
func HandleSamples(w http.ResponseWriter, r *http.Request) {
	samples := GetSamplePlans()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(samples)
}
