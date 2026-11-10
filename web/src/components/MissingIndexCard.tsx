import React, { useState } from 'react';
import { Sparkles, Copy, Check, ShieldCheck } from 'lucide-react';
import type { IndexRecommendation } from '../types/plan';

interface Props {
  recommendations: IndexRecommendation[];
  engine: string;
}

export const MissingIndexCard: React.FC<Props> = ({ recommendations, engine }) => {
  const [copiedMap, setCopiedMap] = useState<Record<string, boolean>>({});
  const [activeEngineMap, setActiveEngineMap] = useState<Record<string, 'postgres' | 'mysql' | 'hypopg'>>({});

  const handleCopy = (id: string, text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedMap((prev) => ({ ...prev, [id]: true }));
    setTimeout(() => {
      setCopiedMap((prev) => ({ ...prev, [id]: false }));
    }, 2000);
  };

  if (!recommendations || recommendations.length === 0) {
    return (
      <div className="bg-[#111827] border border-[#1f2937] rounded-xl p-6 shadow-xl">
        <div className="flex items-center gap-2 border-b border-[#1f2937] pb-3 mb-4">
          <Sparkles className="w-4 h-4 text-emerald-400" />
          <h3 className="text-xs font-semibold uppercase tracking-wider text-gray-300">
            Missing Index Advisor
          </h3>
        </div>
        <div className="flex items-center gap-3 p-4 rounded-lg bg-emerald-950/20 border border-emerald-800/40 text-emerald-300 text-xs">
          <ShieldCheck className="w-5 h-5 shrink-0" />
          <div>
            <p className="font-semibold text-emerald-200">No Missing Indexes Identified</p>
            <p className="text-[11px] opacity-80 mt-0.5">
              The query plan does not contain unindexed sequential table scans with selective filter expressions.
            </p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-[#111827] border border-[#1f2937] rounded-xl p-5 shadow-xl">
      <div className="flex items-center justify-between border-b border-[#1f2937] pb-3 mb-4">
        <div className="flex items-center gap-2">
          <Sparkles className="w-4 h-4 text-blue-400" />
          <h3 className="text-xs font-semibold uppercase tracking-wider text-gray-300">
            Missing Index Advisor ({recommendations.length} Recommended)
          </h3>
        </div>
        <span className="text-[11px] text-gray-400">
          Synthesized from filter predicates and sort keys
        </span>
      </div>

      <div className="space-y-4">
        {recommendations.map((rec) => {
          const currentEngine = activeEngineMap[rec.node_id] || (engine === 'mysql' ? 'mysql' : 'postgres');
          let ddl = currentEngine === 'postgres' ? rec.ddl_postgres : currentEngine === 'mysql' ? rec.ddl_mysql : (rec.hypopg_sql || rec.ddl_postgres);
          const isCopied = copiedMap[rec.node_id];

          return (
            <div
              key={rec.node_id}
              className="bg-[#090d16] border border-[#1f2937] rounded-lg p-4 transition-all hover:border-blue-500/40"
            >
              <div className="flex flex-wrap items-center justify-between gap-2 mb-2">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-bold text-white font-mono">{rec.index_name}</span>
                  <span className="text-[11px] px-2 py-0.5 rounded bg-gray-800 text-cyan-400 border border-gray-700">
                    table: {rec.table}
                  </span>
                  {rec.is_functional && (
                    <span className="text-[10px] font-bold px-2 py-0.5 rounded bg-purple-900/50 text-purple-300 border border-purple-700/60">
                      Functional Index
                    </span>
                  )}
                </div>

                {/* Dialect Switcher */}
                <div className="flex items-center gap-1 bg-[#111827] p-0.5 rounded border border-[#1f2937] text-[10px]">
                  <button
                    type="button"
                    onClick={() => setActiveEngineMap((prev) => ({ ...prev, [rec.node_id]: 'postgres' }))}
                    className={`px-2 py-0.5 rounded transition ${
                      currentEngine === 'postgres' ? 'bg-blue-600 text-white font-semibold' : 'text-gray-400 hover:text-white'
                    }`}
                  >
                    PostgreSQL
                  </button>
                  <button
                    type="button"
                    onClick={() => setActiveEngineMap((prev) => ({ ...prev, [rec.node_id]: 'mysql' }))}
                    className={`px-2 py-0.5 rounded transition ${
                      currentEngine === 'mysql' ? 'bg-blue-600 text-white font-semibold' : 'text-gray-400 hover:text-white'
                    }`}
                  >
                    MySQL
                  </button>
                  <button
                    type="button"
                    onClick={() => setActiveEngineMap((prev) => ({ ...prev, [rec.node_id]: 'hypopg' }))}
                    className={`px-2 py-0.5 rounded transition ${
                      currentEngine === 'hypopg' ? 'bg-purple-600 text-white font-semibold' : 'text-gray-400 hover:text-white'
                    }`}
                    title="Simulate index in PostgreSQL without creating physical files"
                  >
                    HypoPG Simulation
                  </button>
                </div>
              </div>

              <p className="text-xs text-gray-300 mb-2">{rec.estimated_impact}</p>

              {/* DDL Box */}
              <div className="relative group">
                <pre className="p-3 bg-[#030712] border border-gray-800 rounded font-mono text-xs text-blue-300 overflow-x-auto select-all">
                  {ddl}
                </pre>
                <button
                  type="button"
                  onClick={() => handleCopy(rec.node_id, ddl)}
                  className="absolute right-2 top-2 px-2.5 py-1 rounded bg-[#1f2937] hover:bg-gray-700 text-gray-200 text-xs flex items-center gap-1 border border-gray-700 transition"
                  title="Copy DDL Statement"
                >
                  {isCopied ? (
                    <>
                      <Check className="w-3.5 h-3.5 text-emerald-400" />
                      <span className="text-emerald-400 font-semibold">Copied!</span>
                    </>
                  ) : (
                    <>
                      <Copy className="w-3.5 h-3.5" />
                      <span>Copy SQL</span>
                    </>
                  )}
                </button>
              </div>

              {/* Reason / Columns tag */}
              <div className="flex flex-wrap items-center gap-2 mt-3 text-[11px] text-gray-400">
                <span className="text-gray-500">Columns:</span>
                <span className="font-mono text-gray-300 font-semibold">({rec.columns.join(', ')})</span>
                <span className="text-gray-600">•</span>
                <span className="truncate max-w-md">{rec.reason}</span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
