package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/alexandrmotologa/planforge"
	"github.com/alexandrmotologa/planforge/internal/advisor"
	"github.com/alexandrmotologa/planforge/internal/analyzer"
	"github.com/alexandrmotologa/planforge/internal/anonymizer"
	"github.com/alexandrmotologa/planforge/internal/api"
	"github.com/alexandrmotologa/planforge/internal/exporter"
	"github.com/alexandrmotologa/planforge/internal/parser"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "serve":
		cmdServe(os.Args[2:])
	case "analyze":
		cmdAnalyze(os.Args[2:])
	case "export":
		cmdExport(os.Args[2:])
	case "anonymize":
		cmdAnonymize(os.Args[2:])
	case "compare":
		cmdCompare(os.Args[2:])
	case "samples":
		cmdSamples()
	case "version", "--version", "-v":
		fmt.Printf("PlanForge v%s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
	case "help", "--help", "-h":
		printUsage()
	default:
		// If argument is a file path, treat as analyze
		if _, err := os.Stat(command); err == nil {
			cmdAnalyze(os.Args[1:])
			return
		}
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`PlanForge v%s — PostgreSQL & MySQL Query Plan Visualizer & Index Advisor

Usage:
  planforge serve [--host 127.0.0.1] [--port 8080] [--open]
  planforge analyze <plan-file> [--anonymize]
  planforge export <plan-file> -o <report.html> [--anonymize]
  planforge anonymize <plan-file> -o <scrubbed.json>
  planforge compare <baseline-file> <optimized-file>
  planforge samples
  planforge version

Commands:
  serve       Start local web studio with embedded interactive DAG visualizer
  analyze     Analyze query plan in terminal with time hog and index recommendations
  export      Generate standalone offline interactive HTML report (zero cloud dependencies)
  anonymize   Scrub table names, columns, and literals for secure sharing
  compare     Compare two execution plans side-by-side (before vs after)
  samples     List built-in sample execution plans
  version     Display version information
`, version)
}

func normalizeArgs(args []string) []string {
	var flags []string
	var pos []string
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			flags = append(flags, args[i])
			if (args[i] == "-o" || args[i] == "-port" || args[i] == "--port" || args[i] == "-host" || args[i] == "--host") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flags = append(flags, args[i])
			}
		} else {
			pos = append(pos, args[i])
		}
	}
	return append(flags, pos...)
}

func cmdServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	host := fs.String("host", "127.0.0.1", "Bind address")
	port := fs.String("port", "8080", "Port to listen on")
	openBrowser := fs.Bool("open", false, "Open default web browser automatically")
	fs.Parse(normalizeArgs(args))

	staticFS, err := planforge.GetStaticFS()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Static assets not embedded: %v\n", err)
	}

	router := api.NewRouter(staticFS)
	addr := fmt.Sprintf("%s:%s", *host, *port)
	url := fmt.Sprintf("http://%s:%s", *host, *port)

	fmt.Printf("PlanForge Studio running at %s\n", url)
	fmt.Println("Offline & Local-First (Memory usage: < 30MB RAM). Press Ctrl+C to stop.")

	if *openBrowser {
		openURL(url)
	}

	if err := http.ListenAndServe(addr, router); err != nil {
		fmt.Fprintf(os.Stderr, "Server failed: %v\n", err)
		os.Exit(1)
	}
}

func cmdAnalyze(args []string) {
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)
	anonymizeOpt := fs.Bool("anonymize", false, "Mask table and column names")
	jsonOpt := fs.Bool("json", false, "Output analysis report as JSON")
	fs.Parse(normalizeArgs(args))

	files := fs.Args()
	var rawData []byte
	var err error

	if len(files) == 0 || files[0] == "-" {
		rawData, err = io.ReadAll(os.Stdin)
	} else {
		rawData, err = os.ReadFile(files[0])
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read input plan: %v\n", err)
		os.Exit(1)
	}

	plan, err := parser.DetectAndParse(string(rawData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing plan: %v\n", err)
		os.Exit(1)
	}

	if *anonymizeOpt {
		s := anonymizer.NewSanitizer()
		plan = s.AnonymizePlan(plan)
	}

	indexRecs := advisor.RecommendIndexes(plan)
	report := analyzer.EvaluatePlan(plan, indexRecs)

	if *jsonOpt {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(report)
		return
	}

	printTerminalReport(report)
}

func cmdExport(args []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	outputFile := fs.String("o", "planforge-report.html", "Output HTML file path")
	anonymizeOpt := fs.Bool("anonymize", false, "Mask table and column names")
	fs.Parse(normalizeArgs(args))

	files := fs.Args()
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Error: Specify plan file to export.")
		os.Exit(1)
	}

	rawData, err := os.ReadFile(files[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read plan file: %v\n", err)
		os.Exit(1)
	}

	plan, err := parser.DetectAndParse(string(rawData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	if *anonymizeOpt {
		s := anonymizer.NewSanitizer()
		plan = s.AnonymizePlan(plan)
	}

	indexRecs := advisor.RecommendIndexes(plan)
	report := analyzer.EvaluatePlan(plan, indexRecs)

	htmlContent, err := exporter.GenerateStandaloneHTML(report)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Export generation error: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*outputFile, []byte(htmlContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Standalone offline report written to %s (file size: %d bytes)\n", *outputFile, len(htmlContent))
}

func cmdAnonymize(args []string) {
	fs := flag.NewFlagSet("anonymize", flag.ExitOnError)
	outputFile := fs.String("o", "", "Output file path (prints to stdout if omitted)")
	fs.Parse(normalizeArgs(args))

	files := fs.Args()
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Error: Specify plan file to anonymize.")
		os.Exit(1)
	}

	rawData, err := os.ReadFile(files[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read plan file: %v\n", err)
		os.Exit(1)
	}

	plan, err := parser.DetectAndParse(string(rawData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	s := anonymizer.NewSanitizer()
	scrubbed := s.AnonymizePlan(plan)

	outBytes, err := json.MarshalIndent(scrubbed, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal JSON: %v\n", err)
		os.Exit(1)
	}

	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, outBytes, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Anonymized plan written to %s\n", *outputFile)
	} else {
		fmt.Println(string(outBytes))
	}
}

func cmdCompare(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: planforge compare <baseline.json> <optimized.json>")
		os.Exit(1)
	}

	baseBytes, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read baseline file: %v\n", err)
		os.Exit(1)
	}

	optBytes, err := os.ReadFile(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read optimized file: %v\n", err)
		os.Exit(1)
	}

	basePlan, err := parser.DetectAndParse(string(baseBytes))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing baseline plan: %v\n", err)
		os.Exit(1)
	}

	optPlan, err := parser.DetectAndParse(string(optBytes))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing optimized plan: %v\n", err)
		os.Exit(1)
	}

	comp := analyzer.Compare(basePlan, optPlan)

	fmt.Println("================================================================================")
	fmt.Println("                        PlanForge Performance Comparison                        ")
	fmt.Println("================================================================================")
	fmt.Printf("Summary: %s\n\n", comp.Summary)
	fmt.Printf("Duration: %8.2f ms  -> %8.2f ms   (Delta: %+6.1f%%, %+.2f ms)\n",
		basePlan.TotalDurationMs, optPlan.TotalDurationMs, comp.DurationDeltaPercent, comp.DurationDeltaMs)
	fmt.Printf("Cost:     %8.1f     -> %8.1f        (Delta: %+6.1f%%, %+.1f)\n",
		basePlan.Root.TotalCost, optPlan.Root.TotalCost, comp.CostDeltaPercent, comp.CostDelta)
	fmt.Printf("Disk Read:%8d blk  -> %8d blk     (Delta: %+d blocks)\n",
		basePlan.Summary.TotalSharedReadBlocks, optPlan.Summary.TotalSharedReadBlocks, comp.BufferReadDelta)
	fmt.Println("================================================================================")
}

func cmdSamples() {
	samples := api.GetSamplePlans()
	fmt.Println("Built-in Sample Execution Plans:")
	fmt.Println("--------------------------------------------------------------------------------")
	for _, s := range samples {
		fmt.Printf("[%s] %s (%s)\n  %s\n\n", s.ID, s.Title, strings.ToUpper(s.Engine), s.Description)
	}
}

func printTerminalReport(r *analyzer.FullReport) {
	p := r.Plan
	fmt.Println("================================================================================")
	fmt.Printf(" PlanForge Analysis Report  |  Engine: %-10s  |  Health Score: %d (%s)\n",
		strings.ToUpper(string(p.Engine)), r.HealthScore, r.HealthGrade)
	fmt.Println("================================================================================")
	fmt.Printf("Execution Duration: %8.2f ms  |  Total Plan Cost: %10.1f\n", p.TotalDurationMs, p.Summary.TotalCost)
	fmt.Printf("Buffer Hit Ratio:   %8.1f %%   |  Disk Read Volume: %9.2f MB (%d blocks)\n",
		r.BufferEfficiency.HitRatioPercent, r.BufferEfficiency.DiskReadMB, r.BufferEfficiency.SharedReadBlocks)
	fmt.Printf("Total Operators:    %8d     |  Actual Rows:      %10.0f\n", p.Summary.TotalNodes, p.Summary.TotalActualRows)
	fmt.Println("--------------------------------------------------------------------------------")

	// Time Hogs
	if len(r.TimeHogs) > 0 {
		fmt.Println("LATENCY BOTTLENECK OPERATORS (TIME HOGS):")
		for _, th := range r.TimeHogs {
			fmt.Printf("  * [%s] %s accounts for %.1f%% of query runtime (%.2f ms)\n",
				strings.ToUpper(string(th.Severity)), th.NodeType, th.DurationPercent, th.DurationMs)
			fmt.Printf("    Advice: %s\n", th.Recommendation)
		}
		fmt.Println()
	}

	// Estimation Skew
	if len(r.EstimationAnomalies) > 0 {
		fmt.Println("PLANNER ESTIMATION ANOMALIES (STALE CATALOG STATISTICS):")
		for _, ea := range r.EstimationAnomalies {
			fmt.Printf("  * [%s] %s\n", strings.ToUpper(string(ea.Severity)), ea.Message)
			fmt.Printf("    Advice: %s\n", ea.Recommendation)
		}
		fmt.Println()
	}

	// Buffer Warnings
	if r.BufferEfficiency.Severity != analyzer.SeverityInfo {
		fmt.Println("BUFFER CACHE & DISK I/O WARNINGS:")
		fmt.Printf("  * [%s] %s\n", strings.ToUpper(string(r.BufferEfficiency.Severity)), r.BufferEfficiency.Message)
		fmt.Printf("    Advice: %s\n\n", r.BufferEfficiency.Recommendation)
	}

	// Memory Spills
	if len(r.MemorySpills) > 0 {
		fmt.Println("TEMPORARY FILE / WORK_MEM SPILLS:")
		for _, ms := range r.MemorySpills {
			fmt.Printf("  * [%s] %s\n", strings.ToUpper(string(ms.Severity)), ms.Message)
			fmt.Printf("    Advice: %s\n", ms.Recommendation)
		}
		fmt.Println()
	}

	// Index Recommendations
	if len(r.IndexRecommendations) > 0 {
		fmt.Println("MISSING INDEX RECOMMENDATIONS:")
		for _, rec := range r.IndexRecommendations {
			fmt.Printf("  Index: %s ON %s (%s)\n", rec.IndexName, rec.Table, strings.Join(rec.Columns, ", "))
			fmt.Printf("  Impact: %s\n", rec.EstimatedImpact)
			fmt.Println("  SQL Statement:")
			for _, line := range strings.Split(rec.DDLPostgres, "\n") {
				fmt.Printf("    %s\n", line)
			}
			fmt.Println()
		}
	} else {
		fmt.Println("MISSING INDEX RECOMMENDATIONS:\n  None. Query operators are adequately supported by existing indexes.")
	}
	fmt.Println("================================================================================")
}

func openURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
