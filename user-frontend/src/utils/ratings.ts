export function imgError(e: Event) {
  (e.target as HTMLImageElement).style.display = 'none'
}

export function getStars(s: { id: string; view_count: number }): number {
  let base = 4.0
  if (s.view_count > 0) base = Math.min(4.8, 4.0 + Math.log10(s.view_count + 1) * 0.2)
  let hash = 0
  for (const c of s.id) hash = (hash * 31 + c.charCodeAt(0)) & 0xFF
  return Math.min(5.0, parseFloat((base + (hash % 3) / 10).toFixed(1)))
}

export function starText(s: { id: string; view_count: number }): string {
  const r = getStars(s)
  return '★'.repeat(Math.round(r)) + '☆'.repeat(5 - Math.round(r)) + ` ${r.toFixed(1)}`
}
