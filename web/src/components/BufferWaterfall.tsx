import React from 'react';
import { HardDrive, Cpu, AlertCircle } from 'lucide-react';
import type { BufferEfficiencyFinding, PlanNode } from '../types/plan';

interface Props {
  efficiency: BufferEfficiencyFinding;
  nodes: PlanNode[];
}

export const BufferWaterfall: React.FC<Props> = ({ efficiency, nodes }) => {
  const totalBlocks = efficiency.shared_hit_blocks + efficiency.shared_read_blocks;
  const hitPct = totalBlocks > 0 ? (efficiency.shared_hit_blocks / totalBlocks) * 100 : 100;
  const readPct = totalBlocks > 0 ? (efficiency.shared_read_blocks / totalBlocks) * 100 : 0;

  // Filter nodes with disk reads
  const diskReadNodes = nodes
    .filter((n) => n.buffers && n.buffers.shared_read_blocks > 0)
    .sort((a, b) => b.buffers.shared_read_blocks - a.buffers.shared_read_blocks);

  return (
    <div className="bg-[#111827] border border-[#1f2937] rounded-xl p-5 shadow-xl">
      <div className="flex items-center justify-between border-b border-[#1f2937] pb-3 mb-4">
        <div className="flex items-center gap-2">
          <HardDrive className="w-4 h-4 text-emerald-400" />
          <h3 className="text-xs font-semibold uppercase tracking-wider text-gray-300">
            Buffer Cost Waterfall & I/O Locality
          </h3>
        </div>
        <span className="text-[11px] text-gray-400">
          RAM Hit Ratio: <strong className="text-white">{efficiency.hit_ratio_percent.toFixed(1)}%</strong>
        </span>
      </div>

      {/* Main Ratio Bar */}
      <div className="mb-4">
        <div className="flex justify-between text-xs text-gray-400 mb-1.5 font-mono">
          <span className="flex items-center gap-1 text-emerald-400 font-semibold">
            <Cpu className="w-3.5 h-3.5" /> RAM Hits: {efficiency.shared_hit_blocks.toLocaleString()} blocks ({hitPct.toFixed(1)}%)
          </span>
          <span className="flex items-center gap-1 text-red-400 font-semibold">
            <HardDrive className="w-3.5 h-3.5" /> Disk Reads: {efficiency.shared_read_blocks.toLocaleString()} blocks ({readPct.toFixed(1)}%)
          </span>
        </div>

        <div className="w-full bg-gray-800 h-3 rounded-full flex overflow-hidden">
          <div
            className="bg-emerald-500 h-full transition-all duration-300"
            style={{ width: `${hitPct}%` }}
            title={`RAM Cache Hits: ${hitPct.toFixed(1)}%`}
          />
          <div
            className="bg-red-500 h-full transition-all duration-300"
            style={{ width: `${readPct}%` }}
            title={`Physical Disk Reads: ${readPct.toFixed(1)}%`}
          />
        </div>

        {efficiency.disk_read_mb > 0 && (
          <p className="text-[11px] text-gray-400 mt-2">
            Storage throughput: Query forced <strong className="text-white">{efficiency.disk_read_mb.toFixed(2)} MB</strong> to be transferred from physical disk into OS/database cache.
          </p>
        )}
      </div>

      {/* Diagnostic Message */}
      <div
        className={`p-3 rounded-lg border text-xs flex items-start gap-2.5 mb-4 ${
          efficiency.severity === 'critical'
            ? 'bg-red-950/30 border-red-800/60 text-red-300'
            : efficiency.severity === 'warning'
            ? 'bg-amber-950/30 border-amber-800/60 text-amber-300'
            : 'bg-emerald-950/30 border-emerald-800/60 text-emerald-300'
        }`}
      >
        <AlertCircle className="w-4 h-4 shrink-0 mt-0.5" />
        <div>
          <p className="font-semibold">{efficiency.message}</p>
          <p className="text-[11px] opacity-80 mt-0.5">{efficiency.recommendation}</p>
        </div>
      </div>

      {/* Top Disk I/O Operators */}
      {diskReadNodes.length > 0 && (
        <div>
          <h4 className="text-[11px] font-semibold text-gray-400 uppercase tracking-wider mb-2">
            Operators Generating Physical Disk Reads
          </h4>
          <div className="space-y-1.5">
            {diskReadNodes.slice(0, 4).map((n) => (
              <div
                key={n.id}
                className="flex items-center justify-between p-2 rounded bg-[#090d16] border border-[#1f2937] text-xs font-mono"
              >
                <span className="text-gray-300 font-bold truncate">
                  {n.node_type} {n.relation_name && <span className="text-cyan-400">on {n.relation_name}</span>}
                </span>
                <span className="text-red-400 font-bold">
                  {n.buffers.shared_read_blocks.toLocaleString()} blocks ({(n.buffers.shared_read_blocks * 8 / 1024).toFixed(2)} MB)
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
