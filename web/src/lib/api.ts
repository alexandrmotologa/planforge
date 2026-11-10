import type { CompareResponse, FullReport, SamplePlan } from '../types/plan';

export async function analyzePlan(plan: string, anonymize: boolean): Promise<FullReport> {
  const res = await fetch('/api/v1/plans/analyze', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ plan, anonymize }),
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Analysis failed' }));
    throw new Error(err.error || `Server responded with ${res.status}`);
  }

  return res.json();
}

export async function comparePlans(baseline: string, optimized: string): Promise<CompareResponse> {
  const res = await fetch('/api/v1/plans/compare', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ baseline, optimized }),
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Comparison failed' }));
    throw new Error(err.error || `Server responded with ${res.status}`);
  }

  return res.json();
}

export async function exportPlanHtml(plan: string, anonymize: boolean): Promise<Blob> {
  const res = await fetch('/api/v1/plans/export', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ plan, anonymize }),
  });

  if (!res.ok) {
    throw new Error('Failed to generate export');
  }

  return res.blob();
}

export async function fetchSamples(): Promise<SamplePlan[]> {
  const res = await fetch('/api/v1/samples');
  if (!res.ok) {
    throw new Error('Failed to fetch samples');
  }
  return res.json();
}
