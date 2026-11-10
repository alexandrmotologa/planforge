import React from 'react';
import { Clock, Flame } from 'lucide-react';
import type { PlanNode } from '../types/plan';

interface Props {
  nodes: PlanNode[];
  selectedNode: PlanNode | null;
  onSelectNode: (node: PlanNode) => void;
}

export const TimeWaterfall: React.FC<Props> = ({ nodes, selectedNode, onSelectNode }) => {
  // Sort nodes by exclusive duration descending
  const sorted = [...nodes].sort((a, b) => b.exclusive_duration_ms - a.exclusive_duration_ms);
  const maxDuration = sorted[0]?.exclusive_duration_ms || 1;

  return (
    <div className="bg-[#111827] border border-[#1f2937] rounded-xl p-5 shadow-xl">
      <div className="flex items-center justify-between border-b border-[#1f2937] pb-3 mb-4">
        <div className="flex items-center gap-2">
          <Clock className="w-4 h-4 text-amber-400" />
          <h3 className="text-xs font-semibold uppercase tracking-wider text-gray-300">
            Execution Time Waterfall (Exclusive Runtime)
          </h3>
        </div>
        <span className="text-[11px] text-gray-500">Sorted by latency impact</span>
      </div>

      <div className="space-y-3">
        {sorted.slice(0, 10).map((node) => {
          const isSelected = selectedNode?.id === node.id;
          const barWidth = Math.max((node.exclusive_duration_ms / maxDuration) * 100, 2);

          let barColor = 'bg-blue-600';
          if (node.is_time_hog) barColor = 'bg-red-500';
          else if (node.duration_percent >= 20) barColor = 'bg-amber-500';

          return (
            <div
              key={node.id}
              onClick={() => onSelectNode(node)}
              className={`p-2.5 rounded-lg border transition cursor-pointer select-none ${
                isSelected
                  ? 'bg-blue-950/30 border-blue-500'
                  : 'bg-[#090d16] border-[#1f2937] hover:border-gray-600'
              }`}
            >
              <div className="flex items-center justify-between text-xs mb-1.5">
                <div className="flex items-center gap-2 truncate">
                  <span className="font-bold text-gray-200">{node.node_type}</span>
                  {node.relation_name && (
                    <span className="text-[11px] font-mono text-cyan-400">
                      on {node.relation_name}
                    </span>
                  )}
                  {node.is_time_hog && (
                    <span className="flex items-center gap-0.5 text-[10px] font-bold text-red-400">
                      <Flame className="w-3 h-3 fill-current" /> Time Hog
                    </span>
                  )}
                </div>
                <div className="font-mono text-xs flex items-center gap-2 shrink-0">
                  <span className="text-white font-bold">
                    {node.exclusive_duration_ms.toFixed(2)} ms
                  </span>
                  <span className="text-gray-400">({node.duration_percent.toFixed(1)}%)</span>
                </div>
              </div>

              {/* Waterfall progress bar */}
              <div className="w-full bg-gray-800 h-2 rounded-full overflow-hidden">
                <div
                  className={`h-full rounded-full transition-all duration-300 ${barColor}`}
                  style={{ width: `${barWidth}%` }}
                />
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
