import React from 'react';
import { Flame } from 'lucide-react';
import type { PlanNode } from '../types/plan';

interface Props {
  root: PlanNode;
  selectedNode: PlanNode | null;
  onSelectNode: (node: PlanNode) => void;
}

export const FlameGraph: React.FC<Props> = ({ root, selectedNode, onSelectNode }) => {
  // Collect levels for flamechart
  const levels: PlanNode[][] = [];

  const traverse = (node: PlanNode, depth: number) => {
    if (!levels[depth]) levels[depth] = [];
    levels[depth].push(node);
    if (node.children) {
      for (const child of node.children) {
        traverse(child, depth + 1);
      }
    }
  };

  traverse(root, 0);

  const getNodeColor = (node: PlanNode, isSelected: boolean) => {
    if (isSelected) return 'bg-blue-500 border-white text-white shadow-lg ring-2 ring-blue-400';
    if (node.is_time_hog) return 'bg-red-600/90 border-red-500 text-white hover:bg-red-500';
    if (node.duration_percent >= 20) return 'bg-amber-600/90 border-amber-500 text-white hover:bg-amber-500';
    if (node.is_estimation_anomaly) return 'bg-purple-600/90 border-purple-500 text-white hover:bg-purple-500';
    return 'bg-[#1e293b] border-[#334155] text-gray-200 hover:bg-[#334155]';
  };

  return (
    <div className="bg-[#111827] border border-[#1f2937] rounded-xl p-5 shadow-xl">
      <div className="flex items-center justify-between border-b border-[#1f2937] pb-3 mb-4">
        <div className="flex items-center gap-2">
          <Flame className="w-4 h-4 text-orange-400" />
          <h3 className="text-xs font-semibold uppercase tracking-wider text-gray-300">
            Execution FlameChart (Depth vs Duration Proportions)
          </h3>
        </div>
        <span className="text-[11px] text-gray-400">
          Proportional time visualizer &bull; Click block to inspect
        </span>
      </div>

      <div className="space-y-1.5 overflow-x-auto pb-2">
        {levels.map((levelNodes, depthIndex) => (
          <div key={`depth-${depthIndex}`} className="flex gap-1 min-w-[600px] h-8">
            {levelNodes.map((n) => {
              const isSelected = selectedNode?.id === n.id;
              const colorClass = getNodeColor(n, isSelected);
              const flexGrow = Math.max(n.exclusive_duration_ms > 0 ? n.exclusive_duration_ms : n.total_cost, 1);

              return (
                <div
                  key={n.id}
                  onClick={() => onSelectNode(n)}
                  style={{ flex: `${flexGrow} 1 0%` }}
                  className={`border rounded flex items-center justify-between px-2 cursor-pointer transition select-none truncate text-[11px] font-mono ${colorClass}`}
                  title={`${n.node_type} ${n.relation_name ? `on ${n.relation_name}` : ''}\nTime: ${n.exclusive_duration_ms.toFixed(2)} ms (${n.duration_percent.toFixed(1)}%)\nCost: ${n.total_cost}`}
                >
                  <span className="truncate font-semibold">
                    {n.node_type} {n.relation_name && `(${n.relation_name})`}
                  </span>
                  <span className="text-[10px] opacity-80 ml-1 shrink-0">
                    {n.exclusive_duration_ms > 0 ? `${n.exclusive_duration_ms.toFixed(1)}ms` : `c:${n.total_cost.toFixed(0)}`}
                  </span>
                </div>
              );
            })}
          </div>
        ))}
      </div>
    </div>
  );
};
