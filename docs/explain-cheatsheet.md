# Database EXPLAIN Command Reference

Instructions for extracting execution plans with buffer statistics and timing metrics across PostgreSQL and MySQL.

## PostgreSQL

To gather execution timings, row counts, and buffer cache hits, run `EXPLAIN` with the `ANALYZE` and `BUFFERS` parameters.

### Recommended JSON format
```sql
EXPLAIN (ANALYZE, BUFFERS, SETTINGS, WAL, FORMAT JSON)
SELECT o.id, o.total_amount, c.name
FROM orders o
JOIN customers c ON c.id = o.customer_id
WHERE o.status = 'active'
  AND o.created_at >= '2026-01-01'
ORDER BY o.created_at DESC;
```

### Saving JSON output from psql
```bash
psql -d mydatabase -A -t -c "EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) SELECT * FROM orders WHERE status = 'active';" > plan.json
```

### Standard text format
If you prefer text output in `psql`, PlanForge parses text trees directly:
```sql
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM orders WHERE tenant_id = 42 AND created_at > NOW() - INTERVAL '7 days';
```

## MySQL

### MySQL 8.0+ JSON format
MySQL provides execution plan details through `FORMAT=JSON`:
```sql
EXPLAIN FORMAT=JSON
SELECT o.id, c.name, o.total_amount
FROM orders o
JOIN customers c ON c.id = o.customer_id
WHERE o.status = 'shipped'
  AND o.created_at >= '2026-01-01'
ORDER BY o.created_at DESC;
```

### Saving MySQL JSON output
```bash
mysql -u root -p -e "EXPLAIN FORMAT=JSON SELECT * FROM orders WHERE status = 'shipped';" mydatabase > mysql_plan.json
```

### MySQL 8.0.18+ EXPLAIN ANALYZE tree
For queries running on MySQL 8.0.18 or newer, `EXPLAIN ANALYZE` records actual execution timings and row counts:
```sql
EXPLAIN ANALYZE
SELECT o.id, c.name
FROM orders o
JOIN customers c ON c.id = o.customer_id
WHERE o.status = 'active';
```

Paste either the resulting JSON document or text tree directly into the PlanForge interface or CLI.
