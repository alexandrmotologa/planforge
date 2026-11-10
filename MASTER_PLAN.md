# PlanForge: Architecture and Engineering Master Plan

Self-hosted PostgreSQL and MySQL query execution plan visualizer, buffer cost waterfall, and missing index advisor operating under 30MB of RAM.

---

## 1. Problem space and objectives

### Problem definition
When database queries degrade application throughput, engineers execute `EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)`. Understanding multi-thousand-line JSON execution trees requires visualization tools.

Existing options present three primary drawbacks:
1. **Data privacy leaks**: Developers paste production execution plans into third-party cloud tools. These plans expose confidential schema structures, tenant identifiers, table names, row counts, and filter predicates, conflicting with compliance requirements (SOC 2, HIPAA, ISO 27001).
2. **Text readability**: Reading raw terminal execution trees obscures whether latency originates from sequential scans, nested loop cartesian explosions, or disk buffer misses.
3. **Absence of automated remediation**: Traditional visualizers display node hierarchies without computing actionable indexing statements or highlighting planner estimation miscalculations.

### Objectives
PlanForge provides a self-hosted, offline toolchain in a single Go binary with an embedded React 19 interface. It ingests execution plans from PostgreSQL and MySQL, identifies execution bottlenecks, inspects buffer cache efficiency, detects estimation errors caused by stale catalog statistics, and synthesizes database index creation statements (`CREATE INDEX CONCURRENTLY` for PostgreSQL and standard index DDL for MySQL).

---

## 2. System architecture

```
[ Developer / DBA (Pastes JSON/Text plan or uploads file) ]
                          │
                          │  HTTP POST /api/v1/plans/analyze (or CLI stdin)
                          ▼
┌────────────────────────────────────────────────────────────────────────┐
│                          PlanForge Core Engine                         │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Execution Plan Parsers                                           │  │
│  │ - PostgreSQL Parser (JSON format and raw text tree format)       │  │
│  │ - MySQL Parser (JSON format and EXPLAIN ANALYZE tree format)     │  │
│  │ - Timing normalizer (Exclusive vs. Inclusive subtree durations)  │  │
│  └──────────────────────────────────┬───────────────────────────────┘  │
│                                     │                                  │
│                                     ▼                                  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Diagnostic Intelligence Engine                                   │  │
│  │ - Time Hog Analyzer: flags nodes consuming > 50% of duration     │  │
│  │ - Planner Estimation Verifier: Actual vs. Planned rows ratio     │  │
│  │ - Buffer I/O Inspector: Shared Hit (RAM) vs. Shared Read (Disk)  │  │
│  │ - Memory and Temp Spill Detector: work_mem disk sorts / hashes   │  │
│  └──────────────────────────────────┬───────────────────────────────┘  │
│                                     │                                  │
│                                     ▼                                  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Missing Index Advisor                                            │  │
│  │ - Predicate extractor (equality, inequality, sort keys)          │  │
│  │ - Index statement synthesizer (CREATE INDEX CONCURRENTLY)        │  │
│  └──────────────────────────────────┬───────────────────────────────┘  │
│                                     │                                  │
│         ┌───────────────────────────┼──────────────────────────┐       │
│         ▼                           ▼                          ▼       │
│  ┌───────────────┐        ┌───────────────────┐      ┌───────────────┐ │
│  │ Schema        │        │ Embedded React 19 │      │ Standalone    │ │
│  │ Anonymizer    │        │ Web Studio        │      │ HTML Exporter │ │
│  │ (PII Masking) │        │ (Interactive DAG) │      │ (Zero Network)│ │
│  └───────────────┘        └───────────────────┘      └───────────────┘ │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Technology stack

| Layer | Component | Specification |
| :--- | :--- | :--- |
| **Backend & CLI** | Go 1.23+ | Native cross-platform execution, memory footprint below 30MB, zero CGo |
| **Web Studio** | React 19 + TypeScript + Vite | Client-side DAG visualizer, stateful inspector, zero external CDN calls |
| **Styling** | Tailwind CSS + Lucide Icons | Focused developer UI, dark theme optimized for terminal and IDE users |
| **Graph Layout** | Custom Hierarchical Layout | Fast DAG layout engine supporting deep query trees without canvas lag |
| **Single Binary Delivery** | Go `embed` package | Frontend assets embedded directly into the Go executable |

---

## 4. Repository structure

```
B:\workgit\planforge\
├── cmd\
│   └── planforge\
│       └── main.go                 # Entrypoint: serve, analyze, export, anonymize
├── internal\
│   ├── config\
│   │   └── config.go               # Server options, network bindings, defaults
│   ├── parser\
│   │   ├── models.go               # Unified Node, Timing, Cost, and Buffer structures
│   │   ├── postgres_json.go        # PostgreSQL JSON execution plan parser
│   │   ├── postgres_text.go        # PostgreSQL text format execution plan parser
│   │   ├── mysql_json.go           # MySQL FORMAT=JSON plan parser
│   │   └── detector.go             # Engine and format auto-detector
│   ├── analyzer\
│   │   ├── bottlenecks.go          # Exclusive duration calculations and time hog flags
│   │   ├── estimation.go           # Planner estimation error ratios
│   │   ├── buffers.go              # Shared hit vs disk read metrics and byte volumes
│   │   └── memory_spills.go        # WorkMem overflows and temporary file usage
│   ├── advisor\
│   │   ├── indexer.go              # Filter predicate parsing and column extraction
│   │   ├── rules.go                # Composite index column ordering heuristics
│   │   └── ddl.go                  # DDL statement generator (Postgres and MySQL)
│   ├── anonymizer\
│   │   └── sanitizer.go            # Literal and identifier replacement engine
│   ├── exporter\
│   │   └── html.go                 # Standalone offline single-file HTML report generator
│   └── api\
│       ├── router.go               # HTTP routing and middleware
│       ├── handlers.go             # Ingestion, comparison, and analysis endpoints
│       └── samples.go              # Embedded benchmark and educational sample plans
├── web\
│   ├── src\
│   │   ├── components\
│   │   │   ├── PlanDagCanvas.tsx   # Interactive execution tree with node heatmaps
│   │   │   ├── NodeDetailDrawer.tsx# Operator inspector (costs, loops, filters, buffers)
│   │   │   ├── TimeWaterfall.tsx   # Exclusive time breakdown chart
│   │   │   ├── BufferWaterfall.tsx # RAM hit vs disk read breakdown
│   │   │   ├── MissingIndexCard.tsx# Copyable index recommendation cards
│   │   │   ├── PlanComparison.tsx  # Dual plan diff visualizer (before vs after)
│   │   │   └── PlanInputEditor.tsx # Input panel with file drop and presets
│   │   ├── types\
│   │   │   └── plan.ts             # TypeScript definitions matching Go models
│   │   ├── lib\
│   │   │   ├── api.ts              # Analysis API client
│   │   │   └── anonymize.ts        # Client-side sanitization utilities
│   │   ├── App.tsx                 # Application root
│   │   ├── main.tsx                # Entrypoint
│   │   └── index.css               # Tailwind styles
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── docs\
│   ├── architecture.md             # Ingestion pipeline and normalization design
│   ├── explain-cheatsheet.md       # Guide for extracting plans across PostgreSQL and MySQL
│   └── indexing-rules.md           # Heuristics for composite index synthesis
├── Makefile                        # Build, test, and package targets
├── Dockerfile                      # Minimal container image (< 25MB)
├── README.md                       # Documentation, usage guide, and quick start
└── MASTER_PLAN.md                  # Specification document
```

---

## 5. Diagnostic rules and index synthesis

### Diagnostic thresholds

1. **Duration Hog**:
   - Condition: Node exclusive duration exceeds 50% of overall query time.
   - Classification: Critical bottleneck.
   - Output: Targeted operator identification for tuning.

2. **Stale Catalog Statistics**:
   - Condition: $\frac{\text{Actual Rows}}{\text{Plan Rows}} > 10.0$ or $\frac{\text{Actual Rows}}{\text{Plan Rows}} < 0.1$.
   - Classification: High planner misestimate.
   - Output: Recommendation to run `ANALYZE <table>;` or increase statistics targets.

3. **High Disk Buffer I/O**:
   - Condition: $\frac{\text{Shared Read}}{\text{Shared Read} + \text{Shared Hit}} > 0.35$.
   - Classification: Cache miss bottleneck.
   - Output: Notification of physical disk reads, converted to total Megabytes read.

4. **WorkMem Disk Overflow**:
   - Condition: Sort or Hash operators reporting external disk spill (`Sort Method: external merge` or temporary file creation).
   - Classification: Memory allocation bottleneck.
   - Output: Recommendation to configure `SET work_mem = '...';`.

5. **JIT Compilation Overhead**:
   - Condition: JIT compilation duration exceeds 20% of query turnaround or exceeds execution duration.
   - Classification: JIT compiler overhead bottleneck.
   - Output: Recommendation to configure `SET jit = off;` or tune `jit_above_cost`.

6. **SQL Anti-Patterns**:
   - Condition: Correlated SubPlan loops, implicit typecasts on indexed columns, or lossy bitmap rechecks.
   - Classification: Query structure antipattern.
   - Output: Targeted SQL rewrite instructions (JOIN/CTE) or schema alignment.

### Index DDL synthesis rules

When an unindexed table scan (`Seq Scan` or MySQL `table scan`) with predicate filters is detected:
1. Extract equality columns (`col = value` or `col IS NULL`). Place these first in index definition.
2. Extract functional expressions (`lower(email)`, `upper(name)`). Formulate functional indexes to prevent table scans on expression lookups.
3. Extract inequality or range columns (`col > value`, `col BETWEEN a AND b`). Place after equality columns.
4. Inspect subsequent or parent `Sort` operations on the same table. If sort keys align with index order, append them to avoid explicit sort steps.
5. Generate HypoPG simulation SQL (`SELECT * FROM hypopg_create_index(...)`) for testing planner cost drop without disk allocation.
6. Output:
   ```sql
   -- PostgreSQL
   CREATE INDEX CONCURRENTLY idx_<table>_<cols> ON <table> (<col1>, <col2>);

   -- MySQL
   CREATE INDEX idx_<table>_<cols> ON <table> (<col1>, <col2>);

   -- HypoPG Simulation
   SELECT * FROM hypopg_create_index('CREATE INDEX ON <table> (<col1>, <col2>)');
   ```

---

## 6. Implementation steps

### Phase 1: Go backend, models, and execution plan parsers
- Initialize Go module `github.com/alexandrmotologa/planforge`.
- Define unified models in `internal/parser/models.go` representing operators, execution timings, cost envelopes, buffer counters, and children nodes.
- Implement PostgreSQL JSON parser handling all node variants (`Seq Scan`, `Index Scan`, `Index Only Scan`, `Bitmap Heap Scan`, `Hash Join`, `Merge Join`, `Nested Loop`, `Sort`, `Aggregate`, `CTE Scan`, `Materialize`, `Gather`).
- Implement PostgreSQL text format parser for handling standard `psql` outputs.
- Implement MySQL JSON and text format parsers.
- Write unit tests covering diverse query plan inputs.

### Phase 2: Diagnostic engine, advisor, and sanitization
- Implement bottleneck duration calculations and time hog detection in `internal/analyzer/bottlenecks.go`.
- Implement row estimation skew detector in `internal/analyzer/estimation.go`.
- Implement buffer hit ratio and disk I/O metrics in `internal/analyzer/buffers.go`.
- Implement work_mem disk spill detector in `internal/analyzer/memory_spills.go`.
- Implement missing index advisor in `internal/advisor/indexer.go` with composite column heuristics.
- Implement identifier and literal anonymizer in `internal/anonymizer/sanitizer.go`.
- Implement plan comparison engine (Plan A vs. Plan B metrics diff).
- Implement standalone offline HTML exporter in `internal/exporter/html.go`.

### Phase 3: Web studio frontend (React 19, TypeScript, Tailwind)
- Set up Vite + React 19 project in `web/`.
- Build input panel with file drop, format auto-detection, and bundled real-world sample plans.
- Implement hierarchical execution tree DAG with cost and duration heatmaps.
- Implement Node Detail Drawer inspecting exact operator metadata, filter expressions, loops, and buffer counters.
- Implement Time Waterfall and Buffer Waterfall charts.
- Implement Missing Index cards with syntax highlighting and one-click copy.
- Implement Plan Comparison view comparing two query plans side-by-side.
- Implement Client-side Anonymization toggle.
- Implement Export to Standalone HTML button.

### Phase 4: CLI tooling, embedded assets, and HTTP server
- Connect Go HTTP server with embedded frontend static files using `//go:embed`.
- Implement CLI subcommands:
  - `planforge serve`: Runs web studio locally, with optional browser auto-open.
  - `planforge analyze <plan.json>`: Terminal report highlighting time hogs, buffer stats, and index recommendations.
  - `planforge export <plan.json> -o report.html`: Generates an offline interactive HTML report.
  - `planforge anonymize <plan.json> -o scrubbed.json`: Outputs sanitized plan JSON.
- Provide built-in sample query plans for instant testing.

### Phase 5: Verification, documentation, and repository setup
- Add comprehensive test suites for parsing, analysis, advisor, and CLI.
- Write humanized `README.md` following technical writing principles (no AI clichés, no em/en dashes).
- Write `docs/architecture.md`, `docs/explain-cheatsheet.md`, and `docs/indexing-rules.md`.
- Provide Dockerfile and multi-platform Makefile.
- Verify end-to-end functionality across sample plans.
