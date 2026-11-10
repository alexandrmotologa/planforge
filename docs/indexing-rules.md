# Index Recommendation Heuristics

PlanForge synthesizes composite database index definitions when it detects unindexed table scans (`Seq Scan` in PostgreSQL or `Table scan (ALL)` in MySQL) with selective predicate filters.

## Predicate classification

When a table scan has a `Filter` clause:

1. **Equality operators (`=`, `IS NULL`, `IS TRUE`, `IS FALSE`)**:
   Columns tested with direct equality are placed first in the index column list. Equality predicates narrow the B-tree traversal to contiguous leaf page ranges.
   Example: `WHERE tenant_id = 42 AND status = 'active'`
   Column order: `(tenant_id, status)`

2. **Range and inequality operators (`>`, `<`, `>=`, `<=`, `BETWEEN`, `LIKE 'prefix%'`)**:
   Columns tested with range comparisons are appended after equality columns. B-tree indexes support range scans effectively once all leading equality columns are fixed.
   Example: `WHERE tenant_id = 42 AND created_at > '2026-01-01'`
   Column order: `(tenant_id, created_at)`

3. **Sort keys (`ORDER BY`)**:
   If a parent or sibling `Sort` operator exists for the same relation and orders rows by columns not already in the equality set, PlanForge appends those sort columns to the index definition. This allows the database engine to return rows in pre-sorted order, avoiding an explicit in-memory or disk sort step.
   Example: `WHERE tenant_id = 42 ORDER BY created_at DESC`
   Synthesized index:
   ```sql
   CREATE INDEX CONCURRENTLY idx_orders_tenant_id_created_at
   ON orders (tenant_id, created_at DESC);
   ```

4. **Functional expressions (`lower(email)`, `upper(name)`)**:
   When filter predicates apply functions to columns, standard B-tree column indexes are bypassed. PlanForge detects functional predicates and synthesizes functional expressions:
   ```sql
   CREATE INDEX CONCURRENTLY idx_users_lower_email
   ON users (lower(email));
   ```

## HypoPG simulation (Testing without disk allocation)

For multi-gigabyte or terabyte tables, building an index can take hours and consume gigabytes of storage. PlanForge generates simulation commands for the PostgreSQL `hypopg` extension:
```sql
SELECT * FROM hypopg_create_index('CREATE INDEX ON orders (tenant_id, status, created_at)');
EXPLAIN SELECT ...;
```
This enables developers and DBAs to verify whether the query planner will use the proposed index and how much the estimated cost drops before creating physical files.

## Concurrent creation in production

PlanForge generates PostgreSQL statements using `CREATE INDEX CONCURRENTLY`. In PostgreSQL, standard index creation locks the target table against concurrent writes (`INSERT`, `UPDATE`, `DELETE`) until the index build finishes. Adding the `CONCURRENTLY` keyword allows writes to proceed uninterrupted during index construction.

For MySQL, standard `CREATE INDEX` syntax is produced, which uses online DDL algorithms where supported by InnoDB storage engines.
