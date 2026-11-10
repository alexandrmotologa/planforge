import React, { useState, useEffect } from 'react';
import { Zap, AlertCircle, Share2, Check, Keyboard } from 'lucide-react';
import type { CompareResponse, FullReport, PlanNode, SamplePlan } from './types/plan';
import { analyzePlan, comparePlans, exportPlanHtml, fetchSamples } from './lib/api';
import { encodePlanToHash, decodePlanFromHash } from './lib/urlSharing';
import { PlanInputEditor } from './components/PlanInputEditor';
import { SummaryHeader } from './components/SummaryHeader';
import { PlanDagCanvas } from './components/PlanDagCanvas';
import { NodeDetailDrawer } from './components/NodeDetailDrawer';
import { TimeWaterfall } from './components/TimeWaterfall';
import { BufferWaterfall } from './components/BufferWaterfall';
import { FlameGraph } from './components/FlameGraph';
import { DiagnosticsPanel } from './components/DiagnosticsPanel';
import { MissingIndexCard } from './components/MissingIndexCard';
import { PlanComparisonView } from './components/PlanComparisonView';
import { ShortcutsModal } from './components/ShortcutsModal';

export const App: React.FC = () => {
  const [report, setReport] = useState<FullReport | null>(null);
  const [comparison, setComparison] = useState<CompareResponse | null>(null);
  const [selectedNode, setSelectedNode] = useState<PlanNode | null>(null);
  const [samples, setSamples] = useState<SamplePlan[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activeMode, setActiveMode] = useState<'single' | 'compare'>('single');
  const [isCopiedShare, setIsCopiedShare] = useState(false);
  const [isShortcutsOpen, setIsShortcutsOpen] = useState(false);
  const [globalToast, setGlobalToast] = useState<string | null>(null);

  const showToast = (msg: string) => {
    setGlobalToast(msg);
    setTimeout(() => setGlobalToast(null), 3000);
  };

  // Load sample presets or URL hash on mount
  useEffect(() => {
    const fromHash = decodePlanFromHash();
    if (fromHash) {
      handleAnalyze(fromHash, false);
      showToast('Loaded execution plan from URL');
    }

    fetchSamples()
      .then((data) => {
        setSamples(data);
        if (!fromHash && data.length > 0) {
          handleAnalyze(data[0].raw_content, false);
        }
      })
      .catch(() => {
        // Fallback default sample if server is initializing
      });
  }, []);

  // Global Paste Listener (Ctrl+V anywhere in window)
  useEffect(() => {
    const handleGlobalPaste = (e: ClipboardEvent) => {
      // Don't intercept if user is typing in a textarea or input
      const target = e.target as HTMLElement;
      if (target && (target.tagName === 'TEXTAREA' || target.tagName === 'INPUT')) {
        return;
      }

      const pasted = e.clipboardData?.getData('text');
      if (pasted && pasted.trim()) {
        const trimmed = pasted.trim();
        if (
          trimmed.startsWith('{') ||
          trimmed.startsWith('[') ||
          trimmed.includes('Seq Scan') ||
          trimmed.includes('Index Scan') ||
          trimmed.includes('cost=') ||
          trimmed.includes('query_block')
        ) {
          e.preventDefault();
          showToast('Pasted execution plan detected');
          handleAnalyze(trimmed, false);
        }
      }
    };

    window.addEventListener('paste', handleGlobalPaste);
    return () => window.removeEventListener('paste', handleGlobalPaste);
  }, []);

  // Keyboard Shortcuts Listener
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      const isInput = target && (target.tagName === 'TEXTAREA' || target.tagName === 'INPUT');

      if (e.key === '?' && !isInput) {
        e.preventDefault();
        setIsShortcutsOpen((prev) => !prev);
      } else if (e.key === 'Escape') {
        setIsShortcutsOpen(false);
        setSelectedNode(null);
      } else if (e.key === 'e' && !isInput && !e.ctrlKey && !e.metaKey) {
        e.preventDefault();
        handleExport();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [report]);

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
      showToast('Exported standalone HTML report');
    } catch (err: any) {
      alert('Export failed: ' + err.message);
    }
  };

  const handleShareLink = () => {
    if (!report || !report.plan.raw_input) return;
    const hash = encodePlanToHash(report.plan.raw_input);
    const fullURL = `${window.location.origin}${window.location.pathname}${hash}`;
    navigator.clipboard.writeText(fullURL);
    setIsCopiedShare(true);
    showToast('Zero-server link copied to clipboard');
    setTimeout(() => setIsCopiedShare(false), 2500);
  };

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
      <header className="bg-[#0f172a] border-b border-[#1f2937] px-6 py-3.5 sticky top-0 z-40 shadow-md">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-blue-600 flex items-center justify-center text-white shadow-md shadow-blue-500/30 shrink-0">
              <Zap className="w-5 h-5 fill-current" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-base font-extrabold text-white tracking-tight">PlanForge</h1>
                <span className="text-[10px] font-semibold px-2 py-0.5 rounded bg-emerald-950/60 text-emerald-400 border border-emerald-800/40 hidden sm:inline-block">
                  Local-First (<span className="font-mono">&lt; 30MB RAM</span>)
                </span>
              </div>
              <p className="text-[11px] text-gray-400 hidden sm:block">
                Self-Hosted PostgreSQL &amp; MySQL Execution Plan Visualizer &amp; Index Advisor
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            {report && (
              <button
                type="button"
                onClick={handleShareLink}
                className="flex items-center gap-1.5 px-2.5 sm:px-3 py-1.5 rounded-lg bg-[#1f2937] hover:bg-[#374151] border border-gray-700 text-xs font-semibold text-gray-200 transition"
                title="Copy shareable link encoded in URL hash"
              >
                {isCopiedShare ? (
                  <>
                    <Check className="w-3.5 h-3.5 text-emerald-400" />
                    <span className="text-emerald-400 hidden sm:inline">Link Copied!</span>
                  </>
                ) : (
                  <>
                    <Share2 className="w-3.5 h-3.5 text-blue-400" />
                    <span className="hidden sm:inline">Share URL</span>
                  </>
                )}
              </button>
            )}

            <button
              type="button"
              onClick={() => setIsShortcutsOpen(true)}
              className="p-1.5 rounded-lg bg-[#1f2937] hover:bg-[#374151] border border-gray-700 text-gray-300 transition"
              title="Keyboard Shortcuts (?)"
            >
              <Keyboard className="w-4 h-4" />
            </button>

            <a
              href="https://github.com/alexandrmotologa/planforge"
              target="_blank"
              rel="noreferrer"
              className="flex items-center gap-1.5 px-2.5 sm:px-3 py-1.5 rounded-lg bg-[#1f2937] hover:bg-[#374151] border border-gray-700 text-xs font-semibold text-gray-200 transition"
            >
              <svg className="w-4 h-4 fill-current" viewBox="0 0 24 24">
                <path fillRule="evenodd" clipRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
              </svg>
              <span className="hidden sm:inline">GitHub</span>
            </a>
          </div>
        </div>
      </header>

      {/* Floating Global Toast */}
      {globalToast && (
        <div className="fixed bottom-6 right-6 z-50 bg-blue-600 text-white text-xs font-semibold px-4 py-2.5 rounded-lg shadow-xl animate-fade-in flex items-center gap-2">
          <Zap className="w-3.5 h-3.5 fill-current" />
          <span>{globalToast}</span>
        </div>
      )}

      {/* Main Container */}
      <main className="max-w-7xl mx-auto px-3 sm:px-6 py-4 sm:py-6 flex-1 w-full">
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

            {/* Split layout: Tree + FlameGraph + Waterfalls on Left, Deep Inspector on Right */}
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
              {/* Left Column (8 cols) */}
              <div className="lg:col-span-8 space-y-6">
                <PlanDagCanvas
                  root={report.plan.root}
                  selectedNode={selectedNode}
                  onSelectNode={setSelectedNode}
                />

                <FlameGraph
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

                <DiagnosticsPanel report={report} />

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

      {/* Shortcuts Modal */}
      <ShortcutsModal isOpen={isShortcutsOpen} onClose={() => setIsShortcutsOpen(false)} />

      {/* Footer */}
      <footer className="border-t border-[#1f2937] py-6 px-6 text-center text-xs text-gray-500 bg-[#090d16]">
        <p>PlanForge — Private, Offline Database Query Optimization Engine & Index Advisor.</p>
        <p className="mt-1">Zero Cloud Leaks &bull; 100% Client & Local Execution &bull; Built with Go & React 19</p>
      </footer>
    </div>
  );
};

export default App;
