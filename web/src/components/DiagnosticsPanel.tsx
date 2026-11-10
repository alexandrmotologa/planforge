import React from 'react';
import { Cpu, AlertTriangle, CheckCircle2, Terminal } from 'lucide-react';
import type { FullReport } from '../types/plan';

interface Props {
  report: FullReport;
}

export const DiagnosticsPanel: React.FC<Props> = ({ report }) => {
  const { jit_finding, anti_patterns } = report;
  const hasAntiPatterns = anti_patterns && anti_patterns.length > 0;
  const hasIssues = jit_finding || hasAntiPatterns;

  if (!hasIssues) {
    return (
      <div className="bg-[#111827] border border-[#1f2937] rounded-xl p-5 shadow-xl">
        <div className="flex items-center gap-2 border-b border-[#1f2937] pb-3 mb-4">
          <CheckCircle2 className="w-4 h-4 text-emerald-400" />
          <h3 className="text-xs font-semibold uppercase tracking-wider text-gray-300">
            JIT & SQL Anti-Pattern Diagnostics
          </h3>
        </div>
        <div className="flex items-center gap-3 p-3.5 rounded-lg bg-emerald-950/20 border border-emerald-800/40 text-emerald-300 text-xs">
          <CheckCircle2 className="w-5 h-5 shrink-0" />
          <div>
            <p className="font-semibold text-emerald-200">Zero SQL Anti-Patterns Detected</p>
            <p className="text-[11px] opacity-80 mt-0.5">
              No JIT compilation bottlenecks, correlated subquery loops, or type-casting issues were identified in this plan.
            </p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-[#111827] border border-[#1f2937] rounded-xl p-5 shadow-xl space-y-4">
      <div className="flex items-center justify-between border-b border-[#1f2937] pb-3">
        <div className="flex items-center gap-2">
          <AlertTriangle className="w-4 h-4 text-amber-400" />
          <h3 className="text-xs font-semibold uppercase tracking-wider text-gray-300">
            Deep SQL & Engine Diagnostics
          </h3>
        </div>
        <span className="text-[11px] text-gray-400">JIT compiler & pattern flags</span>
      </div>

      {/* JIT Card */}
      {jit_finding && (
        <div className="bg-[#090d16] border border-amber-800/50 rounded-lg p-4 text-xs">
          <div className="flex items-center gap-2 text-amber-400 font-bold mb-1.5">
            <Cpu className="w-4 h-4" />
            <span>JIT Compilation Overhead Detected ({jit_finding.overhead_percent}% of turnaround)</span>
          </div>
          <p className="text-gray-300 mb-2">{jit_finding.message}</p>
          <div className="bg-[#030712] p-2.5 rounded border border-gray-800 font-mono text-[11px] text-amber-300 flex items-center gap-2">
            <Terminal className="w-3.5 h-3.5 shrink-0 text-gray-500" />
            <span>SET jit = off;</span>
          </div>
          <p className="text-[11px] text-gray-400 mt-2">{jit_finding.recommendation}</p>
        </div>
      )}

      {/* Anti-Patterns */}
      {hasAntiPatterns && (
        <div className="space-y-2.5">
          {anti_patterns!.map((ap, idx) => (
            <div
              key={`ap-${idx}`}
              className="bg-[#090d16] border border-purple-800/40 rounded-lg p-3.5 text-xs"
            >
              <div className="flex items-center justify-between mb-1">
                <span className="font-bold text-purple-300 font-mono">{ap.pattern_type}</span>
                <span className="text-[10px] uppercase font-bold px-1.5 py-0.5 rounded bg-purple-900/50 text-purple-300 border border-purple-700/50">
                  {ap.severity}
                </span>
              </div>
              <p className="text-gray-300">{ap.message}</p>
              <p className="text-[11px] text-gray-400 mt-1.5 italic font-mono bg-[#030712] p-2 rounded border border-gray-800">
                Fix: {ap.recommendation}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
