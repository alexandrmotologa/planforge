import React from 'react';
import { ArrowRightLeft, TrendingDown, TrendingUp, Zap, HardDrive, DollarSign } from 'lucide-react';
import type { CompareResponse } from '../types/plan';

interface Props {
  data: CompareResponse;
  onClose: () => void;
}

export const PlanComparisonView: React.FC<Props> = ({ data, onClose }) => {
  const { comparison, baseline_report, optimized_report } = data;
  const isFaster = comparison.duration_delta_percent < 0;

  return (
    <div className="space-y-6">
      {/* Top Banner */}
      <div className="bg-[#111827] border border-[#1f2937] rounded-xl p-6 shadow-xl">
        <div className="flex flex-wrap items-center justify-between gap-4 border-b border-[#1f2937] pb-4 mb-4">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-blue-600/10 border border-blue-500/20 flex items-center justify-center text-blue-400">
              <ArrowRightLeft className="w-5 h-5" />
            </div>
            <div>
              <h2 className="text-lg font-bold text-white tracking-tight">Performance Comparison (A/B Diff)</h2>
              <p className="text-xs text-gray-400 mt-0.5">Evaluating query execution latency, planner cost, and buffer metrics</p>
            </div>
          </div>

          <button
            type="button"
            onClick={onClose}
            className="px-3 py-1.5 rounded-lg bg-gray-800 hover:bg-gray-700 text-xs font-semibold text-gray-300 transition"
          >
            Back to Single Plan View
          </button>
        </div>

        {/* Verdict Box */}
        <div
          className={`p-4 rounded-xl border flex items-center gap-3 mb-6 ${
            isFaster
              ? 'bg-emerald-950/20 border-emerald-800/40 text-emerald-300'
              : 'bg-red-950/20 border-red-800/40 text-red-300'
          }`}
        >
          {isFaster ? (
            <TrendingDown className="w-6 h-6 text-emerald-400 shrink-0" />
          ) : (
            <TrendingUp className="w-6 h-6 text-red-400 shrink-0" />
          )}
          <div>
            <h3 className="font-bold text-sm">{comparison.summary}</h3>
            <p className="text-xs opacity-80 mt-0.5">
              Duration delta: {comparison.duration_delta_ms > 0 ? '+' : ''}{comparison.duration_delta_ms.toFixed(2)} ms ({comparison.duration_delta_percent > 0 ? '+' : ''}{comparison.duration_delta_percent.toFixed(1)}%)
            </p>
          </div>
        </div>

        {/* Metrics Diff Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {/* Duration Tile */}
          <div className="bg-[#090d16] border border-[#1f2937] p-4 rounded-xl">
            <div className="flex items-center justify-between text-gray-400 text-xs font-semibold uppercase tracking-wider mb-2">
              <span>Execution Time</span>
              <Zap className="w-4 h-4 text-amber-400" />
            </div>
            <div className="flex items-baseline justify-between">
              <div>
                <span className="text-xs text-gray-500 block">Baseline</span>
                <span className="text-lg font-bold text-gray-300 font-mono">
                  {baseline_report.plan.total_duration_ms.toFixed(2)} ms
                </span>
              </div>
              <span className="text-gray-500 font-bold">&rarr;</span>
              <div>
                <span className="text-xs text-emerald-400 block font-semibold">Optimized</span>
                <span className="text-lg font-bold text-white font-mono">
                  {optimized_report.plan.total_duration_ms.toFixed(2)} ms
                </span>
              </div>
            </div>
          </div>

          {/* Planner Cost Tile */}
          <div className="bg-[#090d16] border border-[#1f2937] p-4 rounded-xl">
            <div className="flex items-center justify-between text-gray-400 text-xs font-semibold uppercase tracking-wider mb-2">
              <span>Planner Total Cost</span>
              <DollarSign className="w-4 h-4 text-blue-400" />
            </div>
            <div className="flex items-baseline justify-between">
              <div>
                <span className="text-xs text-gray-500 block">Baseline</span>
                <span className="text-lg font-bold text-gray-300 font-mono">
                  {baseline_report.plan.root.total_cost.toFixed(1)}
                </span>
              </div>
              <span className="text-gray-500 font-bold">&rarr;</span>
              <div>
                <span className="text-xs text-emerald-400 block font-semibold">Optimized</span>
                <span className="text-lg font-bold text-white font-mono">
                  {optimized_report.plan.root.total_cost.toFixed(1)}
                </span>
              </div>
            </div>
          </div>

          {/* Disk Reads Tile */}
          <div className="bg-[#090d16] border border-[#1f2937] p-4 rounded-xl">
            <div className="flex items-center justify-between text-gray-400 text-xs font-semibold uppercase tracking-wider mb-2">
              <span>Disk Read Blocks</span>
              <HardDrive className="w-4 h-4 text-cyan-400" />
            </div>
            <div className="flex items-baseline justify-between">
              <div>
                <span className="text-xs text-gray-500 block">Baseline</span>
                <span className="text-lg font-bold text-red-400 font-mono">
                  {baseline_report.plan.summary.total_shared_read_blocks.toLocaleString()}
                </span>
              </div>
              <span className="text-gray-500 font-bold">&rarr;</span>
              <div>
                <span className="text-xs text-emerald-400 block font-semibold">Optimized</span>
                <span className="text-lg font-bold text-white font-mono">
                  {optimized_report.plan.summary.total_shared_read_blocks.toLocaleString()}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Side-by-side plan summaries */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="bg-[#111827] border border-[#1f2937] rounded-xl p-5 shadow-xl">
          <div className="flex items-center justify-between border-b border-[#1f2937] pb-3 mb-3">
            <h4 className="text-xs font-bold uppercase tracking-wider text-red-400">
              Baseline Plan: {baseline_report.plan.root.node_type}
            </h4>
            <span className="text-xs text-gray-400">Score: {baseline_report.health_score} ({baseline_report.health_grade})</span>
          </div>
          <p className="text-xs text-gray-300 mb-3">
            Health findings: {baseline_report.time_hogs.length} time hog(s), {baseline_report.estimation_anomalies.length} estimation anomaly(s).
          </p>
          <div className="p-3 bg-[#090d16] border border-gray-800 rounded font-mono text-xs text-gray-300">
            <div>Root Operator: {baseline_report.plan.root.node_type}</div>
            <div>Relation: {baseline_report.plan.root.relation_name || 'N/A'}</div>
            <div>Duration: {baseline_report.plan.total_duration_ms.toFixed(2)} ms</div>
            <div>Buffer Hit Ratio: {baseline_report.buffer_efficiency.hit_ratio_percent.toFixed(1)}%</div>
          </div>
        </div>

        <div className="bg-[#111827] border border-[#1f2937] rounded-xl p-5 shadow-xl">
          <div className="flex items-center justify-between border-b border-[#1f2937] pb-3 mb-3">
            <h4 className="text-xs font-bold uppercase tracking-wider text-emerald-400">
              Optimized Plan: {optimized_report.plan.root.node_type}
            </h4>
            <span className="text-xs text-gray-400">Score: {optimized_report.health_score} ({optimized_report.health_grade})</span>
          </div>
          <p className="text-xs text-gray-300 mb-3">
            Health findings: {optimized_report.time_hogs.length} time hog(s), {optimized_report.estimation_anomalies.length} estimation anomaly(s).
          </p>
          <div className="p-3 bg-[#090d16] border border-gray-800 rounded font-mono text-xs text-gray-300">
            <div>Root Operator: {optimized_report.plan.root.node_type}</div>
            <div>Relation: {optimized_report.plan.root.relation_name || 'N/A'}</div>
            <div>Duration: {optimized_report.plan.total_duration_ms.toFixed(2)} ms</div>
            <div>Buffer Hit Ratio: {optimized_report.buffer_efficiency.hit_ratio_percent.toFixed(1)}%</div>
          </div>
        </div>
      </div>
    </div>
  );
};
