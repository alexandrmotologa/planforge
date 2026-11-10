import React from 'react';
import { Download, Database, ShieldAlert, Cpu, HardDrive, Zap, CheckCircle2 } from 'lucide-react';
import type { FullReport } from '../types/plan';

interface Props {
  report: FullReport;
  onExport: () => void;
}

export const SummaryHeader: React.FC<Props> = ({ report, onExport }) => {
  const { plan, health_score, health_grade, buffer_efficiency } = report;

  const getScoreColor = (score: number) => {
    if (score >= 90) return 'text-emerald-400 border-emerald-500/30 bg-emerald-500/10';
    if (score >= 75) return 'text-blue-400 border-blue-500/30 bg-blue-500/10';
    if (score >= 60) return 'text-amber-400 border-amber-500/30 bg-amber-500/10';
    return 'text-red-400 border-red-500/30 bg-red-500/10';
  };

  return (
    <div className="bg-[#111827] border border-[#1f2937] rounded-xl p-6 shadow-xl mb-8">
      {/* Top row */}
      <div className="flex flex-wrap items-center justify-between gap-4 border-b border-[#1f2937] pb-5 mb-5">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-lg bg-blue-600/10 border border-blue-500/20 flex items-center justify-center text-blue-400">
            <Database className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h2 className="text-lg font-bold text-white tracking-tight">Query Execution Diagnostics</h2>
              <span className="text-[11px] font-semibold uppercase tracking-wider px-2 py-0.5 rounded bg-gray-800 text-gray-300 border border-gray-700">
                {plan.engine}
              </span>
              <span className="text-[11px] font-semibold uppercase tracking-wider px-2 py-0.5 rounded bg-blue-900/40 text-blue-300 border border-blue-700/50">
                {plan.format}
              </span>
            </div>
            <p className="text-xs text-gray-400 mt-0.5">
              Parsed {plan.summary.total_nodes} operators with full buffer cache inspection and missing index advice
            </p>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={onExport}
            className="flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg bg-[#1f2937] hover:bg-[#374151] border border-gray-700 text-xs font-semibold text-gray-200 transition cursor-pointer"
          >
            <Download className="w-3.5 h-3.5 text-blue-400" />
            Export Standalone HTML
          </button>
        </div>
      </div>

      {/* Metric Tiles */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
        {/* Health Score */}
        <div className={`p-4 rounded-xl border flex flex-col justify-between ${getScoreColor(health_score)}`}>
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold uppercase tracking-wider">Health Score</span>
            <CheckCircle2 className="w-4 h-4 opacity-80" />
          </div>
          <div className="mt-2 flex items-baseline gap-2">
            <span className="text-3xl font-extrabold">{health_score}</span>
            <span className="text-sm font-semibold opacity-75">Grade {health_grade}</span>
          </div>
        </div>

        {/* Total Duration */}
        <div className="bg-[#090d16] border border-[#1f2937] p-4 rounded-xl flex flex-col justify-between">
          <div className="flex items-center justify-between text-gray-400">
            <span className="text-xs font-semibold uppercase tracking-wider">Execution Time</span>
            <Zap className="w-4 h-4 text-amber-400" />
          </div>
          <div className="mt-2">
            <span className="text-2xl font-bold text-white">
              {plan.total_duration_ms > 0 ? plan.total_duration_ms.toFixed(2) : plan.summary.total_cost.toFixed(1)}
            </span>
            <span className="text-xs text-gray-400 ml-1">
              {plan.total_duration_ms > 0 ? 'ms' : 'cost'}
            </span>
          </div>
        </div>

        {/* Buffer Hit Ratio */}
        <div className="bg-[#090d16] border border-[#1f2937] p-4 rounded-xl flex flex-col justify-between">
          <div className="flex items-center justify-between text-gray-400">
            <span className="text-xs font-semibold uppercase tracking-wider">Buffer Hit Ratio</span>
            <Cpu className="w-4 h-4 text-emerald-400" />
          </div>
          <div className="mt-2">
            <span className="text-2xl font-bold text-white">
              {buffer_efficiency.hit_ratio_percent.toFixed(1)}%
            </span>
            <span className="text-xs text-gray-400 ml-1">RAM Cache</span>
          </div>
        </div>

        {/* Disk Read Volume */}
        <div className="bg-[#090d16] border border-[#1f2937] p-4 rounded-xl flex flex-col justify-between">
          <div className="flex items-center justify-between text-gray-400">
            <span className="text-xs font-semibold uppercase tracking-wider">Physical Disk Reads</span>
            <HardDrive className="w-4 h-4 text-cyan-400" />
          </div>
          <div className="mt-2">
            <span className="text-2xl font-bold text-white">
              {buffer_efficiency.disk_read_mb.toFixed(2)}
            </span>
            <span className="text-xs text-gray-400 ml-1">MB read</span>
          </div>
        </div>

        {/* Dominant Bottleneck */}
        <div className="bg-[#090d16] border border-[#1f2937] p-4 rounded-xl flex flex-col justify-between col-span-2 md:col-span-1">
          <div className="flex items-center justify-between text-gray-400">
            <span className="text-xs font-semibold uppercase tracking-wider">Top Bottleneck</span>
            <ShieldAlert className="w-4 h-4 text-red-400" />
          </div>
          <div className="mt-2">
            <span className="text-sm font-bold text-red-400 block truncate" title={plan.summary.time_hog_node_type || 'None'}>
              {plan.summary.time_hog_node_type || 'None detected'}
            </span>
            <span className="text-xs text-gray-400">
              {plan.summary.time_hog_percent ? `${plan.summary.time_hog_percent.toFixed(1)}% runtime` : 'Balanced'}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
};
