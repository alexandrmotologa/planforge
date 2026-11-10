import React, { useState } from 'react';
import {
  X,
  Clock,
  Database,
  Filter,
  HardDrive,
  Users,
  Code,
  Layers,
  ArrowUpDown,
  AlertTriangle,
  Flame,
} from 'lucide-react';
import type { PlanNode } from '../types/plan';

interface Props {
  node: PlanNode | null;
  onClose: () => void;
}

export const NodeDetailDrawer: React.FC<Props> = ({ node, onClose }) => {
  const [showRawJson, setShowRawJson] = useState(false);

  if (!node) {
    return (
      <div className="bg-[#111827] border border-[#1f2937] rounded-xl p-6 shadow-xl text-center text-gray-500 text-xs flex flex-col items-center justify-center min-h-[350px]">
        <Layers className="w-8 h-8 text-gray-600 mb-2 opacity-50" />
        <p className="font-semibold text-gray-400">No Operator Selected</p>
        <p className="mt-1">Click any operator node in the execution DAG or waterfall to inspect its detailed performance breakdown.</p>
      </div>
    );
  }

  return (
    <div className="bg-[#111827] border border-[#1f2937] rounded-xl p-5 shadow-xl flex flex-col h-full max-h-[850px] overflow-y-auto">
      {/* Header */}
      <div className="flex items-start justify-between border-b border-[#1f2937] pb-3 mb-4">
        <div>
          <div className="flex items-center gap-2">
            <h3 className="text-base font-bold text-white">{node.node_type}</h3>
            {node.is_time_hog && (
              <span className="flex items-center gap-1 text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded bg-red-900/50 text-red-300 border border-red-700/60">
                <Flame className="w-3 h-3 fill-current" /> Time Hog
              </span>
            )}
            {node.is_estimation_anomaly && (
              <span className="flex items-center gap-1 text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded bg-purple-900/40 text-purple-300 border border-purple-700/50">
                <AlertTriangle className="w-3 h-3" /> Stale Stats
              </span>
            )}
          </div>
          {node.relation_name && (
            <p className="text-xs text-cyan-400 font-mono mt-0.5">
              Relation: {node.relation_name} {node.alias && node.alias !== node.relation_name && `(${node.alias})`}
            </p>
          )}
          {node.index_name && (
            <p className="text-xs text-blue-400 font-mono mt-0.5">
              Index: {node.index_name}
            </p>
          )}
        </div>

        <button
          type="button"
          onClick={onClose}
          className="p-1 rounded-md text-gray-400 hover:text-white hover:bg-gray-800 transition"
        >
          <X className="w-4 h-4" />
        </button>
      </div>

      <div className="space-y-4 text-xs">
        {/* Timing Section */}
        <div className="bg-[#090d16] border border-[#1f2937] rounded-lg p-3">
          <div className="flex items-center gap-1.5 font-semibold text-gray-300 mb-2">
            <Clock className="w-3.5 h-3.5 text-amber-400" />
            Execution Timing
          </div>
          <div className="grid grid-cols-2 gap-2 text-gray-400">
            <div>
              <span className="block text-[11px] text-gray-500">Exclusive Duration:</span>
              <span className="text-white font-bold font-mono">
                {node.exclusive_duration_ms > 0 ? `${node.exclusive_duration_ms.toFixed(2)} ms` : 'N/A'}
              </span>
              {node.duration_percent > 0 && (
                <span className="text-gray-400 ml-1">({node.duration_percent.toFixed(1)}%)</span>
              )}
            </div>
            <div>
              <span className="block text-[11px] text-gray-500">Subtree Total Time:</span>
              <span className="text-white font-bold font-mono">
                {node.actual_total_time_ms > 0 ? `${(node.actual_total_time_ms * node.actual_loops).toFixed(2)} ms` : 'N/A'}
              </span>
            </div>
            <div>
              <span className="block text-[11px] text-gray-500">Per-Loop Startup:</span>
              <span className="text-gray-200 font-mono">
                {node.actual_startup_time_ms > 0 ? `${node.actual_startup_time_ms.toFixed(3)} ms` : '0.00 ms'}
              </span>
            </div>
            <div>
              <span className="block text-[11px] text-gray-500">Executed Loops:</span>
              <span className="text-gray-200 font-mono font-semibold">{node.actual_loops}</span>
            </div>
          </div>
        </div>

        {/* Rows & Planner Estimation Section */}
        <div className="bg-[#090d16] border border-[#1f2937] rounded-lg p-3">
          <div className="flex items-center gap-1.5 font-semibold text-gray-300 mb-2">
            <Database className="w-3.5 h-3.5 text-blue-400" />
            Rows & Planner Estimation
          </div>
          <div className="grid grid-cols-2 gap-2 text-gray-400">
            <div>
              <span className="block text-[11px] text-gray-500">Actual Rows:</span>
              <span className="text-white font-bold font-mono">
                {node.actual_rows.toLocaleString()}
                {node.actual_loops > 1 && ` (x${node.actual_loops} loops)`}
              </span>
            </div>
            <div>
              <span className="block text-[11px] text-gray-500">Planned Estimate:</span>
              <span className="text-gray-200 font-mono">
                {node.plan_rows.toLocaleString()} rows
              </span>
            </div>
            <div>
              <span className="block text-[11px] text-gray-500">Estimation Ratio:</span>
              <span
                className={`font-mono font-bold ${
                  node.is_estimation_anomaly ? 'text-purple-400' : 'text-gray-200'
                }`}
              >
                {node.estimation_ratio > 0 ? `${node.estimation_ratio.toFixed(2)}x` : '1.00x'}
              </span>
            </div>
            {node.rows_removed_by_filter != null && node.rows_removed_by_filter > 0 && (
              <div>
                <span className="block text-[11px] text-gray-500">Removed by Filter:</span>
                <span className="text-amber-400 font-mono font-semibold">
                  {node.rows_removed_by_filter.toLocaleString()}
                </span>
              </div>
            )}
          </div>
        </div>

        {/* Buffers & I/O Section */}
        <div className="bg-[#090d16] border border-[#1f2937] rounded-lg p-3">
          <div className="flex items-center gap-1.5 font-semibold text-gray-300 mb-2">
            <HardDrive className="w-3.5 h-3.5 text-emerald-400" />
            Buffer Cache & Disk I/O
          </div>
          <div className="grid grid-cols-2 gap-2 text-gray-400">
            <div>
              <span className="block text-[11px] text-gray-500">Shared Hit (RAM):</span>
              <span className="text-emerald-400 font-bold font-mono">
                {node.buffers.shared_hit_blocks.toLocaleString()} blocks
              </span>
            </div>
            <div>
              <span className="block text-[11px] text-gray-500">Shared Read (Disk):</span>
              <span
                className={`font-mono font-bold ${
                  node.buffers.shared_read_blocks > 0 ? 'text-red-400' : 'text-gray-200'
                }`}
              >
                {node.buffers.shared_read_blocks.toLocaleString()} blocks
              </span>
            </div>
            {node.buffers.temp_read_blocks > 0 || node.buffers.temp_written_blocks > 0 ? (
              <div className="col-span-2 text-red-400 font-mono bg-red-950/30 p-1.5 rounded border border-red-900/50">
                Temp Disk Spills: {node.buffers.temp_read_blocks} read, {node.buffers.temp_written_blocks} written
              </div>
            ) : null}
          </div>
        </div>

        {/* Predicates Section */}
        {(node.filter || node.index_cond || node.hash_cond || node.merge_cond) && (
          <div className="bg-[#090d16] border border-[#1f2937] rounded-lg p-3 space-y-2">
            <div className="flex items-center gap-1.5 font-semibold text-gray-300 mb-1">
              <Filter className="w-3.5 h-3.5 text-cyan-400" />
              Filter & Predicates
            </div>

            {node.filter && (
              <div>
                <span className="text-[11px] text-gray-500">Filter Expression:</span>
                <pre className="mt-1 p-2 bg-[#030712] border border-gray-800 rounded font-mono text-[11px] text-cyan-300 whitespace-pre-wrap">
                  {node.filter}
                </pre>
              </div>
            )}

            {node.index_cond && (
              <div>
                <span className="text-[11px] text-gray-500">Index Condition:</span>
                <pre className="mt-1 p-2 bg-[#030712] border border-gray-800 rounded font-mono text-[11px] text-blue-300 whitespace-pre-wrap">
                  {node.index_cond}
                </pre>
              </div>
            )}

            {node.hash_cond && (
              <div>
                <span className="text-[11px] text-gray-500">Hash Condition:</span>
                <pre className="mt-1 p-2 bg-[#030712] border border-gray-800 rounded font-mono text-[11px] text-purple-300 whitespace-pre-wrap">
                  {node.hash_cond}
                </pre>
              </div>
            )}
          </div>
        )}

        {/* Sort & Memory Spills */}
        {(node.sort_key || node.sort_method) && (
          <div className="bg-[#090d16] border border-[#1f2937] rounded-lg p-3">
            <div className="flex items-center gap-1.5 font-semibold text-gray-300 mb-2">
              <ArrowUpDown className="w-3.5 h-3.5 text-purple-400" />
              Sort Operation
            </div>
            {node.sort_key && (
              <div className="mb-2">
                <span className="text-[11px] text-gray-500">Sort Keys:</span>
                <p className="font-mono text-gray-200 mt-0.5">{node.sort_key.join(', ')}</p>
              </div>
            )}
            {node.sort_method && (
              <div className="flex items-center justify-between text-gray-400">
                <span>Method: {node.sort_method}</span>
                {node.sort_space_used_kb != null && node.sort_space_used_kb > 0 && (
                  <span className={node.sort_space_type === 'Disk' ? 'text-red-400 font-bold' : 'text-gray-200'}>
                    {node.sort_space_used_kb} kB ({node.sort_space_type})
                  </span>
                )}
              </div>
            )}
          </div>
        )}

        {/* Parallel Workers */}
        {node.workers && node.workers.planned > 0 && (
          <div className="bg-[#090d16] border border-[#1f2937] rounded-lg p-3">
            <div className="flex items-center gap-1.5 font-semibold text-gray-300 mb-2">
              <Users className="w-3.5 h-3.5 text-indigo-400" />
              Parallel Workers
            </div>
            <p className="text-gray-400">
              Planned: <span className="text-white font-mono">{node.workers.planned}</span> | Launched:{' '}
              <span className="text-white font-mono">{node.workers.launched}</span>
            </p>
          </div>
        )}

        {/* Toggle Raw JSON */}
        <div className="pt-2 border-t border-[#1f2937]">
          <button
            type="button"
            onClick={() => setShowRawJson(!showRawJson)}
            className="flex items-center gap-1 text-[11px] text-gray-400 hover:text-white transition"
          >
            <Code className="w-3 h-3" />
            {showRawJson ? 'Hide Raw Node JSON' : 'View Raw Node JSON'}
          </button>
          {showRawJson && (
            <pre className="mt-2 p-2.5 bg-[#030712] border border-gray-800 rounded font-mono text-[10px] text-gray-300 max-h-48 overflow-y-auto">
              {JSON.stringify(node, null, 2)}
            </pre>
          )}
        </div>
      </div>
    </div>
  );
};
