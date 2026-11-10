// Utility to encode and decode plan text into URL hash for serverless sharing.

export function encodePlanToHash(planText: string): string {
  try {
    const encoded = btoa(encodeURIComponent(planText));
    return `#plan=${encoded}`;
  } catch (err) {
    console.error('Failed to encode plan to hash', err);
    return '';
  }
}

export function decodePlanFromHash(): string | null {
  try {
    const hash = window.location.hash;
    if (hash.startsWith('#plan=')) {
      const raw = hash.substring(6);
      return decodeURIComponent(atob(raw));
    }
  } catch (err) {
    console.error('Failed to decode plan from hash', err);
  }
  return null;
}
