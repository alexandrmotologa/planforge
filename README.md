<p align="center">
  <img src="docs/images/logo.png" alt="PlanForge Logo" width="130" style="border-radius: 24px;" />
</p>
<h1 align="center">PlanForge</h1>
<p align="center">
  <strong>Self-Hosted PostgreSQL &amp; MySQL Execution Plan Visualizer, Buffer Cost Waterfall &amp; Missing Index Advisor</strong>
</p>
<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License" /></a>
  <img src="https://img.shields.io/badge/memory-%3C%2030MB%20RAM-emerald.svg" alt="Memory Footprint" />
  <img src="https://img.shields.io/badge/privacy-100%25%20Offline-cyan.svg" alt="Local First" />
  <img src="https://img.shields.io/badge/engines-PostgreSQL%20%7C%20MySQL-blueviolet.svg" alt="Supported Engines" />
</p>

<p align="center">
  <img src="docs/images/planforge_demo.gif" alt="PlanForge Interactive Studio Demo" width="100%" style="border-radius: 12px; border: 1px solid #1f2937;" />
</p>

Self-hosted PostgreSQL and MySQL query execution plan visualizer, buffer cost waterfall, and missing index advisor operating in a single binary under 30MB of RAM.

PlanForge runs completely offline and local-first. It replaces third-party cloud paste sites like explain.depesz.com and explain.dalibo.com, eliminating the risk of transmitting confidential production schemas, table structures, row counts, and tenant data to external servers.

---

## Visual Tour

| Execution DAG & Node Inspector | FlameChart & Latency Waterfall |
| :---: | :---: |
| ![Desktop Overview](docs/images/planforge-1.png) | ![FlameChart & Time Waterfall](docs/images/planforge-2.png) |

| Missing Index Advisor & HypoPG Simulation | Performance Comparison (A/B Diff) |
| :---: | :---: |
| ![Missing Index Advisor](docs/images/planforge-3.png) | ![Performance Comparison A/B Diff](docs/images/planforge-4.png) |

<p align="center">
  <strong>Responsive Mobile &amp; Tablet Workspace</strong><br />
  <img src="docs/images/planforge-5.png" alt="Mobile Responsive View" width="380" style="border-radius: 12px; border: 1px solid #1f2937; margin-top: 8px;" />
</p>

---

## Capabilities

- **PostgreSQL and MySQL support**: Ingests `EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)`, raw `psql` text trees, MySQL `EXPLAIN FORMAT=JSON`, and MySQL 8.0 `EXPLAIN ANALYZE` outputs.
- **Latency bottleneck detection**: Flags operators taking more than 50% of overall query execution time as critical time hogs.
- **Planner estimation anomaly checks**: Compares actual row counts against planner row estimates. Discrepancies exceeding 10x signal stale catalog statistics and provide table-level `ANALYZE` commands.
- **Buffer cache and disk inspection**: Computes shared buffer hit ratios and converts physical disk block reads into total megabytes read from storage.
- **Automated missing index advisor**: Synthesizes production-ready `CREATE INDEX CONCURRENTLY` statements for PostgreSQL, standard index DDL for MySQL, and functional index expressions (`lower(email)`). Includes HypoPG simulation scripts for testing index impact before allocating physical storage.
- **Deep SQL anti-pattern diagnostics**: Detects correlated SubPlan loops executed in iterations, implicit typecasting that disables indexes, and lossy bitmap heap rechecks.
- **JIT compilation overhead detector**: Flags queries where Just-In-Time compilation duration exceeds actual execution time, providing tuning recommendations (`SET jit = off;`).
- **Live database query runner**: Executes queries directly against live PostgreSQL and MySQL databases, wrapping write queries in rolled-back transactions for safe inspection.
- **Execution FlameChart visualizer**: Displays hierarchical time proportions where widths match exact execution duration percentages.
- **Zero-server shareable URLs**: Encodes execution plans into URL hash fragments for sharing across teams without a central database.
- **Client-side schema anonymizer**: Replaces table names, column names, and literal filter parameters with generic tokens to enable safe team sharing without exposing business data.
- **Plan comparison (A/B diff)**: Compares baseline and optimized query execution plans side by side, reporting percentage reductions in execution time, planner cost, and disk block reads.
- **Standalone offline HTML export**: Exports self-contained interactive HTML reports that open in any browser without internet access or CDN dependencies.
- **Minimal resource consumption**: Runs as a single compiled Go binary with an embedded React 19 studio, consuming less than 30MB of RAM.

---

## Installation

### Pre-built binaries
Download the latest binary for your operating system from GitHub Releases:
- Linux (x86_64, aarch64)
- macOS (Apple Silicon, Intel)
- Windows (x64)

### Using Go install
```bash
go install github.com/alexandrmotologa/planforge/cmd/planforge@latest
```

### Using Docker
```bash
docker run -p 8080:8080 alexandrmotologa/planforge:latest
```

### Building from source
Requires Go 1.23+ and Node.js 20+:
```bash
git clone https://github.com/alexandrmotologa/planforge.git
cd planforge
make build
```

---

## Quick start

### 1. Launch the web studio
Start the local server and open the browser interface:
```bash
planforge serve --port 8080 --open
```
Open `http://127.0.0.1:8080` in your web browser. Paste your query plan or select one of the built-in sample presets.

### 2. Execute directly against a live database
Run `EXPLAIN (ANALYZE, BUFFERS)` against a database and view the terminal report or open the browser:
```bash
planforge run --db "postgres://user:pass@localhost:5432/mydb" "SELECT * FROM orders WHERE tenant_id = 42;" --open
```

### 3. Analyze plans from a file or stdin
Inspect query execution plans directly in the terminal:
```bash
planforge analyze plan.json
```
Output includes execution timings, time hog classifications, buffer cache efficiency, and recommended index statements.

### 4. Generate a standalone offline HTML report
```bash
planforge export plan.json -o report.html
```

### 5. Compare baseline and optimized plans
```bash
planforge compare baseline.json optimized.json
```

### 6. Anonymize sensitive plans
```bash
planforge anonymize sensitive_plan.json -o sanitized_plan.json
```

---

## Extracting execution plans

### PostgreSQL
Run `EXPLAIN` with `ANALYZE` and `BUFFERS`:
```sql
EXPLAIN (ANALYZE, BUFFERS, SETTINGS, WAL, FORMAT JSON)
SELECT o.id, o.total_amount, c.name
FROM orders o
JOIN customers c ON c.id = o.customer_id
WHERE o.status = 'active'
  AND o.created_at >= '2026-01-01'
ORDER BY o.created_at DESC;
```

To capture text output directly in `psql`:
```sql
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM orders WHERE tenant_id = 42;
```

### MySQL
```sql
EXPLAIN FORMAT=JSON
SELECT * FROM orders WHERE status = 'shipped' ORDER BY created_at DESC;
```

For MySQL 8.0.18+:
```sql
EXPLAIN ANALYZE
SELECT * FROM orders WHERE status = 'shipped';
```

See [docs/explain-cheatsheet.md](docs/explain-cheatsheet.md) for more details.

---

## HTTP API reference

The embedded server exposes a JSON API:

| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `/api/v1/plans/analyze` | POST | Analyzes execution plan and returns diagnostic findings and index recommendations |
| `/api/v1/plans/compare` | POST | Compares baseline and optimized plans, calculating latency and cost deltas |
| `/api/v1/plans/export` | POST | Generates a standalone offline HTML report file |
| `/api/v1/samples` | GET | Returns built-in demonstration query plans |
| `/health` | GET | Healthcheck returning application status and version |

---

## Documentation

- [System Architecture](docs/architecture.md): Ingestion pipeline, models, and calculation details.
- [EXPLAIN Command Reference](docs/explain-cheatsheet.md): Query plan extraction instructions.
- [Index Recommendation Heuristics](docs/indexing-rules.md): Predicate parsing and composite index ordering rules.

---

## License

MIT License. See [LICENSE](LICENSE) for details.
