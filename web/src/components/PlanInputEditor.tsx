import React, { useState } from 'react';
import { Upload, Play, Sparkles, FileText, Shield, ArrowRightLeft } from 'lucide-react';
import type { SamplePlan } from '../types/plan';

interface Props {
  onAnalyze: (plan: string, anonymize: boolean) => void;
  onCompare: (baseline: string, optimized: string) => void;
  samples: SamplePlan[];
  isLoading: boolean;
  activeMode: 'single' | 'compare';
  setActiveMode: (mode: 'single' | 'compare') => void;
}

export const PlanInputEditor: React.FC<Props> = ({
  onAnalyze,
  onCompare,
  samples,
  isLoading,
  activeMode,
  setActiveMode,
}) => {
  const [singleText, setSingleText] = useState('');
  const [baseText, setBaseText] = useState('');
  const [optText, setOptText] = useState('');
  const [anonymize, setAnonymize] = useState(false);

  const handleSelectSample = (sample: SamplePlan) => {
    setSingleText(sample.raw_content);
  };

  const handleLoadCompareSample = () => {
    if (samples.length > 0) {
      const base = samples[0].raw_content;
      try {
        const opt = JSON.parse(base);
        opt[0]["Execution Time"] = 1.35;
        opt[0]["Planning Time"] = 0.12;
        opt[0].Plan["Node Type"] = "Index Scan";
        opt[0].Plan["Index Name"] = "idx_orders_customer_status";
        opt[0].Plan["Actual Total Time"] = 1.25;
        opt[0].Plan["Total Cost"] = 12.4;
        opt[0].Plan["Actual Rows"] = 150;
        opt[0].Plan["Shared Hit Blocks"] = 18;
        opt[0].Plan["Shared Read Blocks"] = 0;
        setBaseText(base);
        setOptText(JSON.stringify(opt, null, 2));
      } catch {
        setBaseText(base);
        setOptText(base);
      }
    }
  };

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>, target: 'single' | 'base' | 'opt') => {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (event) => {
      const content = event.target?.result as string;
      if (target === 'single') setSingleText(content);
      else if (target === 'base') setBaseText(content);
      else setOptText(content);
    };
    reader.readAsText(file);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (activeMode === 'single') {
      if (singleText.trim()) onAnalyze(singleText, anonymize);
    } else {
      if (baseText.trim() && optText.trim()) onCompare(baseText, optText);
    }
  };

  return (
    <div className="bg-[#111827] border border-[#1f2937] rounded-xl p-5 shadow-2xl mb-8">
      {/* Mode Switch & Samples */}
      <div className="flex flex-wrap items-center justify-between gap-4 border-b border-[#1f2937] pb-4 mb-4">
        <div className="flex items-center gap-2 bg-[#090d16] p-1 rounded-lg border border-[#1f2937]">
          <button
            type="button"
            onClick={() => setActiveMode('single')}
            className={`px-3 py-1.5 rounded-md text-xs font-semibold transition-all ${
              activeMode === 'single'
                ? 'bg-blue-600 text-white shadow-sm'
                : 'text-gray-400 hover:text-white'
            }`}
          >
            Plan Visualizer & Advisor
          </button>
          <button
            type="button"
            onClick={() => setActiveMode('compare')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-semibold transition-all ${
              activeMode === 'compare'
                ? 'bg-blue-600 text-white shadow-sm'
                : 'text-gray-400 hover:text-white'
            }`}
          >
            <ArrowRightLeft className="w-3.5 h-3.5" />
            Compare Plans (A/B Diff)
          </button>
        </div>

        {activeMode === 'single' && samples.length > 0 && (
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-xs text-gray-400 flex items-center gap-1">
              <Sparkles className="w-3.5 h-3.5 text-amber-400" /> Presets:
            </span>
            <div className="flex flex-wrap gap-1.5">
              {samples.map((s) => (
                <button
                  key={s.id}
                  type="button"
                  onClick={() => handleSelectSample(s)}
                  className="px-2.5 py-1 text-xs bg-[#1f2937] hover:bg-[#374151] text-gray-200 rounded border border-gray-700 transition"
                  title={s.description}
                >
                  {s.title}
                </button>
              ))}
            </div>
          </div>
        )}

        {activeMode === 'compare' && samples.length > 0 && (
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={handleLoadCompareSample}
              className="flex items-center gap-1.5 px-3 py-1.5 text-xs bg-[#1f2937] hover:bg-[#374151] text-emerald-300 rounded border border-emerald-800/40 transition font-semibold"
            >
              <Sparkles className="w-3.5 h-3.5 text-emerald-400" />
              Load Sample A/B Comparison
            </button>
          </div>
        )}
      </div>

      <form onSubmit={handleSubmit}>
        {activeMode === 'single' ? (
          <div>
            <div className="flex justify-between items-center mb-2">
              <label className="text-xs font-semibold text-gray-300 uppercase tracking-wider flex items-center gap-1.5">
                <FileText className="w-3.5 h-3.5 text-blue-400" />
                Paste EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) or Text
              </label>
              <label className="cursor-pointer text-xs text-blue-400 hover:text-blue-300 flex items-center gap-1">
                <Upload className="w-3.5 h-3.5" />
                Upload file
                <input
                  type="file"
                  accept=".json,.txt"
                  className="hidden"
                  onChange={(e) => handleFileUpload(e, 'single')}
                />
              </label>
            </div>
            <textarea
              value={singleText}
              onChange={(e) => setSingleText(e.target.value)}
              placeholder="Paste PostgreSQL EXPLAIN JSON, psql text output, or MySQL EXPLAIN FORMAT=JSON here..."
              rows={7}
              className="w-full bg-[#090d16] border border-[#1f2937] rounded-lg p-3 text-xs font-mono text-gray-200 placeholder-gray-500 focus:outline-none focus:border-blue-500 transition"
            />
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <div className="flex justify-between items-center mb-2">
                <span className="text-xs font-semibold text-red-400 uppercase tracking-wider">
                  Baseline Plan (Before)
                </span>
                <label className="cursor-pointer text-xs text-gray-400 hover:text-white flex items-center gap-1">
                  <Upload className="w-3 h-3" /> File
                  <input
                    type="file"
                    className="hidden"
                    onChange={(e) => handleFileUpload(e, 'base')}
                  />
                </label>
              </div>
              <textarea
                value={baseText}
                onChange={(e) => setBaseText(e.target.value)}
                placeholder="Paste original unoptimized plan..."
                rows={7}
                className="w-full bg-[#090d16] border border-[#1f2937] rounded-lg p-3 text-xs font-mono text-gray-200 placeholder-gray-500 focus:outline-none focus:border-red-500"
              />
            </div>
            <div>
              <div className="flex justify-between items-center mb-2">
                <span className="text-xs font-semibold text-emerald-400 uppercase tracking-wider">
                  Optimized Plan (After)
                </span>
                <label className="cursor-pointer text-xs text-gray-400 hover:text-white flex items-center gap-1">
                  <Upload className="w-3 h-3" /> File
                  <input
                    type="file"
                    className="hidden"
                    onChange={(e) => handleFileUpload(e, 'opt')}
                  />
                </label>
              </div>
              <textarea
                value={optText}
                onChange={(e) => setOptText(e.target.value)}
                placeholder="Paste optimized plan with index/tuning..."
                rows={7}
                className="w-full bg-[#090d16] border border-[#1f2937] rounded-lg p-3 text-xs font-mono text-gray-200 placeholder-gray-500 focus:outline-none focus:border-emerald-500"
              />
            </div>
          </div>
        )}

        {/* Footer Bar */}
        <div className="flex flex-wrap items-center justify-between gap-4 mt-4 pt-3 border-t border-[#1f2937]">
          <label className="flex items-center gap-2 cursor-pointer select-none">
            <input
              type="checkbox"
              checked={anonymize}
              onChange={(e) => setAnonymize(e.target.checked)}
              className="w-4 h-4 rounded border-gray-700 bg-gray-900 text-blue-600 focus:ring-0 focus:ring-offset-0"
            />
            <span className="text-xs text-gray-300 flex items-center gap-1">
              <Shield className="w-3.5 h-3.5 text-emerald-400" />
              Client-side anonymization (mask table names, columns, and filter literals)
            </span>
          </label>

          <button
            type="submit"
            disabled={isLoading || (activeMode === 'single' ? !singleText.trim() : !baseText.trim() || !optText.trim())}
            className="flex items-center gap-2 px-5 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed text-white text-xs font-semibold shadow-lg shadow-blue-500/20 transition-all cursor-pointer"
          >
            {isLoading ? (
              <span className="animate-spin inline-block w-3.5 h-3.5 border-2 border-current border-t-transparent rounded-full" />
            ) : (
              <Play className="w-3.5 h-3.5 fill-current" />
            )}
            {activeMode === 'single' ? 'Analyze & Synthesize Indexes' : 'Run Performance Comparison'}
          </button>
        </div>
      </form>
    </div>
  );
};
