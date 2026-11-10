import React from 'react';
import { X, Keyboard } from 'lucide-react';

interface Props {
  isOpen: boolean;
  onClose: () => void;
}

export const ShortcutsModal: React.FC<Props> = ({ isOpen, onClose }) => {
  if (!isOpen) return null;

  const shortcuts = [
    { key: 'Ctrl + V', description: 'Paste execution plan from anywhere in the window' },
    { key: '/', description: 'Focus search input in the Execution DAG' },
    { key: 'e', description: 'Export standalone offline HTML report' },
    { key: 'a', description: 'Toggle client-side schema & literal anonymization' },
    { key: 'Esc', description: 'Close active drawer, modal, or selection' },
    { key: '?', description: 'Open this keyboard shortcuts cheat sheet' },
  ];

  return (
    <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-xs flex items-center justify-center p-4">
      <div className="bg-[#111827] border border-[#1f2937] rounded-xl max-w-md w-full p-5 shadow-2xl">
        <div className="flex items-center justify-between border-b border-[#1f2937] pb-3 mb-4">
          <div className="flex items-center gap-2">
            <Keyboard className="w-4 h-4 text-blue-400" />
            <h3 className="text-sm font-bold text-white tracking-wide">Keyboard Shortcuts</h3>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="p-1 rounded text-gray-400 hover:text-white hover:bg-gray-800 transition"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        <div className="space-y-2.5 text-xs">
          {shortcuts.map((s, idx) => (
            <div
              key={idx}
              className="flex items-center justify-between p-2 rounded bg-[#090d16] border border-[#1f2937]"
            >
              <span className="text-gray-300">{s.description}</span>
              <kbd className="px-2 py-1 rounded bg-[#1e293b] border border-gray-700 text-cyan-300 font-mono text-[11px] font-semibold">
                {s.key}
              </kbd>
            </div>
          ))}
        </div>

        <div className="mt-4 pt-3 border-t border-[#1f2937] text-right">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold transition"
          >
            Got it
          </button>
        </div>
      </div>
    </div>
  );
};
