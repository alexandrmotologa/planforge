package exporter

import (
	"encoding/json"
	"fmt"
	"html"

	"github.com/alexandrmotologa/planforge/internal/analyzer"
)

// GenerateStandaloneHTML creates a self-contained offline HTML file with embedded viewer and report data.
func GenerateStandaloneHTML(report *analyzer.FullReport) (string, error) {
	reportJSON, err := json.Marshal(report)
	if err != nil {
		return "", fmt.Errorf("failed to marshal report: %w", err)
	}

	escapedJSON := html.EscapeString(string(reportJSON))

	htmlTemplate := `<!DOCTYPE html>
<html lang="en" class="dark">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>PlanForge Report - Query Execution Plan</title>
  <style>
    :root {
      --bg: #090d16;
      --card: #111827;
      --border: #1f2937;
      --text: #f3f4f6;
      --text-muted: #9ca3af;
      --accent: #3b82f6;
      --hog: #ef4444;
      --warning: #f59e0b;
      --success: #10b981;
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      background: var(--bg);
      color: var(--text);
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
      padding: 24px;
      line-height: 1.5;
    }
    header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      border-bottom: 1px solid var(--border);
      padding-bottom: 16px;
      margin-bottom: 24px;
    }
    .brand { font-size: 20px; font-weight: 700; color: #60a5fa; letter-spacing: -0.5px; }
    .badge {
      padding: 4px 10px;
      border-radius: 9999px;
      font-size: 12px;
      font-weight: 600;
      text-transform: uppercase;
    }
    .badge-grade { background: #1e3a8a; color: #93c5fd; }
    .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 16px; margin-bottom: 24px; }
    .stat-card {
      background: var(--card);
      border: 1px solid var(--border);
      border-radius: 8px;
      padding: 16px;
    }
    .stat-label { font-size: 12px; color: var(--text-muted); text-transform: uppercase; font-weight: 600; margin-bottom: 6px; }
    .stat-val { font-size: 24px; font-weight: 700; color: #ffffff; }
    .section-title { font-size: 16px; font-weight: 600; margin-bottom: 12px; color: #e5e7eb; }
    .layout-2col { display: grid; grid-template-columns: 1fr 360px; gap: 24px; }
    @media (max-width: 900px) { .layout-2col { grid-template-columns: 1fr; } }
    .tree-box {
      background: var(--card);
      border: 1px solid var(--border);
      border-radius: 8px;
      padding: 16px;
      overflow-x: auto;
    }
    .tree-node {
      margin-left: 20px;
      position: relative;
      padding: 6px 0;
    }
    .node-pill {
      display: inline-flex;
      align-items: center;
      gap: 8px;
      background: #1f2937;
      border: 1px solid #374151;
      padding: 6px 12px;
      border-radius: 6px;
      cursor: pointer;
      font-family: ui-monospace, monospace;
      font-size: 13px;
      transition: all 0.15s ease;
    }
    .node-pill:hover { border-color: var(--accent); background: #283548; }
    .node-pill.hog { border-color: var(--hog); background: #3b1818; }
    .node-pill.warning { border-color: var(--warning); background: #3b2f18; }
    .detail-panel {
      background: var(--card);
      border: 1px solid var(--border);
      border-radius: 8px;
      padding: 16px;
      height: fit-content;
      position: sticky;
      top: 24px;
    }
    .code-box {
      background: #030712;
      border: 1px solid #1f2937;
      padding: 12px;
      border-radius: 6px;
      font-family: ui-monospace, monospace;
      font-size: 12px;
      white-space: pre-wrap;
      overflow-x: auto;
      margin-top: 8px;
      color: #93c5fd;
    }
    .waterfall-bar-bg {
      background: #1f2937;
      height: 8px;
      border-radius: 4px;
      overflow: hidden;
      margin-top: 4px;
    }
    .waterfall-bar-fill {
      height: 100%;
      background: var(--accent);
      border-radius: 4px;
    }
    .btn-copy {
      background: #2563eb;
      color: #ffffff;
      border: none;
      padding: 6px 12px;
      border-radius: 4px;
      font-size: 12px;
      font-weight: 600;
      cursor: pointer;
      margin-top: 8px;
    }
    .btn-copy:hover { background: #1d4ed8; }
  </style>
</head>
<body>
  <header>
    <div>
      <div class="brand">PlanForge Offline Report</div>
      <div style="font-size: 13px; color: var(--text-muted); margin-top: 2px;">Database Execution Plan and Bottleneck Analysis</div>
    </div>
    <div style="display: flex; gap: 8px; align-items: center;">
      <span class="badge badge-grade" id="badge-grade">Score: 100 (A)</span>
      <span class="badge" style="background:#374151;" id="badge-engine">PostgreSQL</span>
    </div>
  </header>

  <div class="grid">
    <div class="stat-card">
      <div class="stat-label">Total Execution Time</div>
      <div class="stat-val" id="val-duration">0.00 ms</div>
    </div>
    <div class="stat-card">
      <div class="stat-label">Buffer Cache Hit Ratio</div>
      <div class="stat-val" id="val-hit-ratio">100%</div>
    </div>
    <div class="stat-card">
      <div class="stat-label">Disk Read Volume</div>
      <div class="stat-val" id="val-disk-mb">0.00 MB</div>
    </div>
    <div class="stat-card">
      <div class="stat-label">Total Plan Operators</div>
      <div class="stat-val" id="val-nodes">0</div>
    </div>
  </div>

  <div class="layout-2col">
    <div>
      <div class="section-title">Execution Operator Hierarchy</div>
      <div class="tree-box" id="tree-container"></div>

      <div style="margin-top: 24px;">
        <div class="section-title">Missing Index Recommendations</div>
        <div id="index-recommendations-container"></div>
      </div>
    </div>

    <div>
      <div class="detail-panel" id="detail-panel">
        <div class="section-title">Operator Inspector</div>
        <div id="node-detail-content" style="color: var(--text-muted); font-size: 13px;">
          Click any operator in the execution tree to inspect execution timings, buffer counts, and filters.
        </div>
      </div>
    </div>
  </div>

  <script id="planforge-data" type="application/json">` + escapedJSON + `</script>

  <script>
    (function() {
      const raw = document.getElementById('planforge-data').textContent;
      const report = JSON.parse(raw);
      const plan = report.plan;

      // Stats
      document.getElementById('badge-grade').textContent = 'Score: ' + report.health_score + ' (' + report.health_grade + ')';
      document.getElementById('badge-engine').textContent = (plan.engine || 'Postgres').toUpperCase();
      document.getElementById('val-duration').textContent = plan.total_duration_ms.toFixed(2) + ' ms';
      document.getElementById('val-hit-ratio').textContent = report.buffer_efficiency.hit_ratio_percent + '%';
      document.getElementById('val-disk-mb').textContent = report.buffer_efficiency.disk_read_mb + ' MB';
      document.getElementById('val-nodes').textContent = plan.summary.total_nodes;

      // Render Tree
      const treeContainer = document.getElementById('tree-container');
      const nodeMap = {};

      function renderNode(node, container) {
        nodeMap[node.id] = node;
        const wrap = document.createElement('div');
        wrap.className = 'tree-node';

        let extraClass = '';
        if (node.is_time_hog) extraClass = ' hog';
        else if (node.is_estimation_anomaly) extraClass = ' warning';

        let label = node.node_type;
        if (node.relation_name) label += ' on ' + node.relation_name;
        if (node.index_name) label += ' (' + node.index_name + ')';

        const timeStr = node.exclusive_duration_ms ? node.exclusive_duration_ms.toFixed(2) + ' ms (' + node.duration_percent + '%)' : 'Cost: ' + node.total_cost;

        wrap.innerHTML = '<div class="node-pill' + extraClass + '" onclick="selectNode(\'' + node.id + '\')">' +
          '<strong>' + label + '</strong>' +
          '<span style="color:#9ca3af;font-size:11px;">' + timeStr + '</span>' +
        '</div>';

        if (node.children && node.children.length > 0) {
          const childrenBox = document.createElement('div');
          for (const ch of node.children) {
            renderNode(ch, childrenBox);
          }
          wrap.appendChild(childrenBox);
        }
        container.appendChild(wrap);
      }

      if (plan.root) {
        renderNode(plan.root, treeContainer);
      }

      // Render Index recommendations
      const recContainer = document.getElementById('index-recommendations-container');
      if (!report.index_recommendations || report.index_recommendations.length === 0) {
        recContainer.innerHTML = '<div style="color:#9ca3af; font-size:13px;">No missing indexes identified. Query operations are supported by existing indexes.</div>';
      } else {
        report.index_recommendations.forEach(rec => {
          const card = document.createElement('div');
          card.style = 'background:var(--card); border:1px solid var(--border); border-radius:8px; padding:16px; margin-bottom:12px;';
          card.innerHTML = '<div style="font-weight:600; color:#60a5fa; font-size:14px;">' + rec.index_name + '</div>' +
            '<div style="font-size:13px; color:#d1d5db; margin-top:4px;">' + rec.estimated_impact + '</div>' +
            '<div class="code-box">' + rec.ddl_postgres + '</div>' +
            '<button class="btn-copy" onclick="navigator.clipboard.writeText(\'' + rec.ddl_postgres.replace(/\n/g, '\\n') + '\')">Copy PostgreSQL DDL</button>';
          recContainer.appendChild(card);
        });
      }

      // Select Node inspector
      window.selectNode = function(id) {
        const n = nodeMap[id];
        if (!n) return;
        const p = document.getElementById('node-detail-content');
        let html = '<div style="font-weight:700; font-size:15px; color:#ffffff; margin-bottom:8px;">' + n.node_type + '</div>';
        if (n.relation_name) html += '<div><strong>Relation:</strong> ' + n.relation_name + (n.alias ? ' (alias ' + n.alias + ')' : '') + '</div>';
        if (n.index_name) html += '<div><strong>Index:</strong> ' + n.index_name + '</div>';
        html += '<div style="margin-top:8px;"><strong>Exclusive Time:</strong> ' + n.exclusive_duration_ms + ' ms (' + n.duration_percent + '%)</div>';
        html += '<div><strong>Actual Rows:</strong> ' + n.actual_rows + ' &times; ' + n.actual_loops + ' loops</div>';
        html += '<div><strong>Planned Rows:</strong> ' + n.plan_rows + ' (Ratio: ' + n.estimation_ratio + 'x)</div>';
        html += '<div style="margin-top:8px;"><strong>Shared Buffers:</strong> ' + (n.buffers.shared_hit_blocks || 0) + ' hit, ' + (n.buffers.shared_read_blocks || 0) + ' read</div>';
        if (n.filter) html += '<div style="margin-top:8px;"><strong>Filter:</strong><div class="code-box">' + n.filter + '</div></div>';
        if (n.index_cond) html += '<div style="margin-top:8px;"><strong>Index Condition:</strong><div class="code-box">' + n.index_cond + '</div></div>';
        p.innerHTML = html;
      };
    })();
  </script>
</body>
</html>`

	return htmlTemplate, nil
}
