import React, { useState, useEffect } from 'react';
import { Zap, AlertCircle } from 'lucide-react';
import type { CompareResponse, FullReport, PlanNode, SamplePlan } from './types/plan';
import { analyzePlan, comparePlans, exportPlanHtml, fetchSamples } from './lib/api';
import { PlanInputEditor } from './components/PlanInputEditor';
import { SummaryHeader } from './components/SummaryHeader';
import { PlanDagCanvas } from './components/PlanDagCanvas';
import { NodeDetailDrawer } from './components/NodeDetailDrawer';
import { TimeWaterfall } from './components/TimeWaterfall';
import { BufferWaterfall } from './components/BufferWaterfall';
import { MissingIndexCard } from './components/MissingIndexCard';
import { PlanComparisonView } from './components/PlanComparisonView';

export const App: React.FC = () => {
  const [report, setReport] = useState<FullReport | null>(null);
  const [comparison, setComparison] = useState<CompareResponse | null>(null);
  const [selectedNode, setSelectedNode] = useState<PlanNode | null>(null);
  const [samples, setSamples] = useState<SamplePlan[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activeMode, setActiveMode] = useState<'single' | 'compare'>('single');

  // Load sample presets on mount
  useEffect(() => {
    fetchSamples()
      .then((data) => {
        setSamples(data);
        // Automatically analyze first preset on load
        if (data.length > 0) {
          handleAnalyze(data[0].raw_content, false);
        }
      })
      .catch(() => {
        // Fallback default sample if server is initializing
      });
  }, []);

  const handleAnalyze = async (planText: string, anonymize: boolean) => {
    setIsLoading(true);
    setError(null);
    setComparison(null);
    try {
      const res = await analyzePlan(planText, anonymize);
      setReport(res);
      setSelectedNode(res.plan.root);
    } catch (err: any) {
      setError(err.message || 'Failed to parse and analyze execution plan');
    } finally {
      setIsLoading(false);
    }
  };

  const handleCompare = async (baseline: string, optimized: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const res = await comparePlans(baseline, optimized);
      setComparison(res);
    } catch (err: any) {
      setError(err.message || 'Failed to compare execution plans');
    } finally {
      setIsLoading(false);
    }
  };

  const handleExport = async () => {
    if (!report) return;
    try {
      const blob = await exportPlanHtml(report.plan.raw_input || '', false);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `planforge-report-${Date.now()}.html`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(url);
    } catch (err: any) {
      alert('Export failed: ' + err.message);
    }
  };

  // Helper to extract all nodes flat for waterfall
  const flattenNodes = (node: PlanNode | undefined): PlanNode[] => {
    if (!node) return [];
    const list: PlanNode[] = [node];
    if (node.children) {
      for (const ch of node.children) {
        list.push(...flattenNodes(ch));
      }
    }
    return list;
  };

  const allNodes = report ? flattenNodes(report.plan.root) : [];

  return (
    <div className="min-h-screen bg-[#090d16] text-gray-100 flex flex-col">
      {/* Top Navbar */}
      <header className="bg-[#0f172a] border-b border-[#1f2937] px-6 py-3.5 sticky top-0 z-50 shadow-md">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-blue-600 flex items-center justify-center text-white shadow-md shadow-blue-500/30">
              <Zap className="w-5 h-5 fill-current" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-base font-extrabold text-white tracking-tight">PlanForge</h1>
                <span className="text-[10px] font-semibold px-2 py-0.5 rounded bg-emerald-950/60 text-emerald-400 border border-emerald-800/40">
                  Local-First (<span className="font-mono">&lt; 30MB RAM</span>)
                </span>
              </div>
              <p className="text-[11px] text-gray-400">
                Self-Hosted PostgreSQL & MySQL Execution Plan Visualizer & Index Advisor
              </p>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <a
              href="https://github.com/alexandrmotologa/planforge"
              target="_blank"
              rel="noreferrer"
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-[#1f2937] hover:bg-[#374151] border border-gray-700 text-xs font-semibold text-gray-200 transition"
            >
              <svg className="w-4 h-4 fill-current" viewBox="0 0 24 24">
                <path fillRule="evenodd" clipRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
              </svg>
              <span>GitHub</span>
            </a>
          </div>
        </div>
      </header>

      {/* Main Container */}
      <main className="max-w-7xl mx-auto px-6 py-6 flex-1 w-full">
        {/* Plan Input Panel */}
        <PlanInputEditor
          onAnalyze={handleAnalyze}
          onCompare={handleCompare}
          samples={samples}
          isLoading={isLoading}
          activeMode={activeMode}
          setActiveMode={setActiveMode}
        />

        {/* Error notification */}
        {error && (
          <div className="bg-red-950/30 border border-red-800/60 rounded-xl p-4 mb-6 flex items-start gap-3 text-red-300 text-xs">
            <AlertCircle className="w-5 h-5 shrink-0 text-red-400" />
            <div>
              <p className="font-bold text-red-200 text-sm">Failed to Analyze Plan</p>
              <p className="mt-1">{error}</p>
            </div>
          </div>
        )}

        {/* Comparison View */}
        {comparison && (
          <PlanComparisonView data={comparison} onClose={() => setComparison(null)} />
        )}

        {/* Single Plan Visual Studio */}
        {!comparison && report && (
          <div className="space-y-6">
            <SummaryHeader report={report} onExport={handleExport} />

            {/* Split layout: Tree + Waterfalls on Left, Deep Inspector on Right */}
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
              {/* Left Column (8 cols) */}
              <div className="lg:col-span-8 space-y-6">
                <PlanDagCanvas
                  root={report.plan.root}
                  selectedNode={selectedNode}
                  onSelectNode={setSelectedNode}
                />

                <TimeWaterfall
                  nodes={allNodes}
                  selectedNode={selectedNode}
                  onSelectNode={setSelectedNode}
                />

                <BufferWaterfall
                  efficiency={report.buffer_efficiency}
                  nodes={allNodes}
                />

                <MissingIndexCard
                  recommendations={report.index_recommendations}
                  engine={report.plan.engine}
                />
              </div>

              {/* Right Column (4 cols sticky inspector) */}
              <div className="lg:col-span-4 sticky top-20">
                <NodeDetailDrawer
                  node={selectedNode}
                  onClose={() => setSelectedNode(null)}
                />
              </div>
            </div>
          </div>
        )}
      </main>

      {/* Footer */}
      <footer className="border-t border-[#1f2937] py-6 px-6 text-center text-xs text-gray-500 bg-[#090d16]">
        <p>PlanForge — Private, Offline Database Query Optimization Engine & Index Advisor.</p>
        <p className="mt-1">Zero Cloud Leaks &bull; 100% Client & Local Execution &bull; Built with Go & React 19</p>
      </footer>
    </div>
  );
};

export default App;
