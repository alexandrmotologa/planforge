import React, { useState } from 'react';
import {
  ChevronRight,
  ChevronDown,
  AlertTriangle,
  Flame,
  Search,
  Maximize2,
  Minimize2,
  Table,
  Zap,
} from 'lucide-react';
import type { PlanNode } from '../types/plan';

interface Props {
  root: PlanNode;
  selectedNode: PlanNode | null;
  onSelectNode: (node: PlanNode) => void;
}

export const PlanDagCanvas: React.FC<Props> = ({ root, selectedNode, onSelectNode }) => {
  const [collapsedMap, setCollapsedMap] = useState<Record<string, boolean>>({});
  const [searchQuery, setSearchQuery] = useState('');

  const toggleCollapse = (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setCollapsedMap((prev) => ({ ...prev, [id]: !prev[id] }));
  };

  const collapseAll = () => {
    const map: Record<string, boolean> = {};
    const walk = (n: PlanNode) => {
      if (n.children && n.children.length > 0) {
        map[n.id] = true;
        n.children.forEach(walk);
      }
    };
    walk(root);
    setCollapsedMap(map);
  };

  const expandAll = () => {
    setCollapsedMap({});
  };

  const renderNode = (node: PlanNode, depth: number = 0): React.ReactNode => {
    const isCollapsed = collapsedMap[node.id];
    const isSelected = selectedNode?.id === node.id;
    const hasChildren = node.children && node.children.length > 0;

    // Search match highlight
    const matchesSearch = searchQuery.trim() !== '' && (
      node.node_type.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (node.relation_name && node.relation_name.toLowerCase().includes(searchQuery.toLowerCase())) ||
      (node.index_name && node.index_name.toLowerCase().includes(searchQuery.toLowerCase()))
    );

    // Styling based on diagnostics
    let borderClass = 'border-[#1f2937] hover:border-blue-500/50';
    let bgClass = 'bg-[#111827]';

    if (node.is_time_hog) {
      borderClass = 'border-red-500 shadow-md shadow-red-500/10';
      bgClass = 'bg-red-950/20';
    } else if (node.is_estimation_anomaly) {
      borderClass = 'border-purple-500/60 shadow-md shadow-purple-500/10';
      bgClass = 'bg-purple-950/20';
    } else if (node.duration_percent >= 20) {
      borderClass = 'border-amber-500/60';
      bgClass = 'bg-amber-950/20';
    }

    if (isSelected) {
      borderClass = 'border-blue-500 ring-2 ring-blue-500/30';
      bgClass = 'bg-blue-950/30';
    }

    return (
      <div key={node.id} className="relative flex flex-col">
        <div
          onClick={() => onSelectNode(node)}
          className={`flex items-start gap-3 p-3.5 my-1.5 rounded-xl border transition-all cursor-pointer select-none ${borderClass} ${bgClass} ${
            matchesSearch ? 'ring-2 ring-amber-400' : ''
          }`}
          style={{ marginLeft: `${depth * 28}px` }}
        >
          {/* Collapse toggle / Connector icon */}
          <div className="mt-0.5">
            {hasChildren ? (
              <button
                type="button"
                onClick={(e) => toggleCollapse(node.id, e)}
                className="w-5 h-5 rounded flex items-center justify-center bg-gray-800 hover:bg-gray-700 text-gray-300 transition"
              >
                {isCollapsed ? <ChevronRight className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}
              </button>
            ) : (
              <div className="w-5 h-5 flex items-center justify-center text-gray-600">
                <span className="w-1.5 h-1.5 rounded-full bg-gray-600" />
              </div>
            )}
          </div>

          {/* Node content */}
          <div className="flex-1 min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <span className="text-xs font-bold text-white tracking-wide">{node.node_type}</span>

              {node.relation_name && (
                <span className="inline-flex items-center gap-1 text-[11px] font-mono px-2 py-0.5 rounded bg-gray-800 text-cyan-300 border border-gray-700">
                  <Table className="w-3 h-3" />
                  {node.relation_name}
                  {node.alias && node.alias !== node.relation_name && (
                    <span className="text-gray-400">as {node.alias}</span>
                  )}
                </span>
              )}

              {node.index_name && (
                <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-blue-900/30 text-blue-300 border border-blue-800/40">
                  {node.index_name}
                </span>
              )}

              {/* Diagnostic badges */}
              {node.is_time_hog && (
                <span className="inline-flex items-center gap-1 text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded bg-red-900/50 text-red-300 border border-red-700/60 animate-pulse">
                  <Flame className="w-3 h-3 fill-current" /> Time Hog ({node.duration_percent.toFixed(1)}%)
                </span>
              )}

              {node.is_estimation_anomaly && (
                <span className="inline-flex items-center gap-1 text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded bg-purple-900/40 text-purple-300 border border-purple-700/50">
                  <AlertTriangle className="w-3 h-3" /> Stale Stats ({node.estimation_ratio.toFixed(1)}x)
                </span>
              )}
            </div>

            {/* Metrics Row */}
            <div className="flex flex-wrap items-center gap-4 mt-2 text-[11px] text-gray-400 font-mono">
              <div className="flex items-center gap-1 text-gray-200">
                <Zap className="w-3 h-3 text-amber-400" />
                <span>
                  {node.exclusive_duration_ms > 0
                    ? `${node.exclusive_duration_ms.toFixed(2)} ms (${node.duration_percent.toFixed(1)}%)`
                    : `Cost: ${node.total_cost.toFixed(1)}`}
                </span>
              </div>

              <div>
                <span>Rows: </span>
                <span className="text-gray-200 font-semibold">{node.actual_rows.toLocaleString()}</span>
                {node.actual_loops > 1 && (
                  <span className="text-gray-400"> &times; {node.actual_loops} loops</span>
                )}
                {node.plan_rows > 0 && (
                  <span className="text-gray-500 ml-1">(est: {node.plan_rows.toLocaleString()})</span>
                )}
              </div>

              {node.buffers && (node.buffers.shared_hit_blocks > 0 || node.buffers.shared_read_blocks > 0) && (
                <div className="flex items-center gap-2">
                  <span className="text-emerald-400">{node.buffers.shared_hit_blocks} hits</span>
                  {node.buffers.shared_read_blocks > 0 && (
                    <span className="text-red-400 font-semibold">{node.buffers.shared_read_blocks} disk reads</span>
                  )}
                </div>
              )}
            </div>

            {/* Execution duration micro-bar */}
            {node.duration_percent > 0 && (
              <div className="w-full bg-gray-800/80 h-1 rounded-full mt-2 overflow-hidden">
                <div
                  className={`h-full rounded-full transition-all duration-300 ${
                    node.is_time_hog ? 'bg-red-500' : node.duration_percent >= 20 ? 'bg-amber-400' : 'bg-blue-500'
                  }`}
                  style={{ width: `${Math.min(node.duration_percent, 100)}%` }}
                />
              </div>
            )}
          </div>
        </div>

        {/* Recursive Children rendering */}
        {hasChildren && !isCollapsed && (
          <div className="flex flex-col">
            {node.children!.map((child) => renderNode(child, depth + 1))}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="bg-[#111827] border border-[#1f2937] rounded-xl p-5 shadow-xl flex flex-col h-full">
      {/* Toolbar */}
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-[#1f2937] pb-3 mb-4">
        <div className="flex items-center gap-2">
          <span className="text-xs font-semibold uppercase tracking-wider text-gray-300">
            Execution Operator DAG
          </span>
          <span className="text-xs text-gray-500">
            (click operator for deep inspection)
          </span>
        </div>

        <div className="flex items-center gap-2">
          {/* Search box */}
          <div className="relative">
            <Search className="w-3.5 h-3.5 text-gray-400 absolute left-2.5 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search table or operator..."
              className="w-48 bg-[#090d16] border border-[#1f2937] rounded-md pl-8 pr-3 py-1 text-xs text-gray-200 placeholder-gray-500 focus:outline-none focus:border-blue-500"
            />
          </div>

          <button
            type="button"
            onClick={expandAll}
            className="p-1.5 rounded bg-gray-800 hover:bg-gray-700 text-gray-300 text-xs transition"
            title="Expand All Nodes"
          >
            <Maximize2 className="w-3.5 h-3.5" />
          </button>
          <button
            type="button"
            onClick={collapseAll}
            className="p-1.5 rounded bg-gray-800 hover:bg-gray-700 text-gray-300 text-xs transition"
            title="Collapse All Nodes"
          >
            <Minimize2 className="w-3.5 h-3.5" />
          </button>
        </div>
      </div>

      {/* DAG Viewport */}
      <div className="overflow-x-auto overflow-y-auto max-h-[750px] pr-2">
        {renderNode(root)}
      </div>
    </div>
  );
};
