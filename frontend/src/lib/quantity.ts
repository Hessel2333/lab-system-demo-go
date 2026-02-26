export function formatNumber(value: unknown, maxFractionDigits = 2): string {
  const n = Number(value)
  if (!Number.isFinite(n)) return '--'
  return new Intl.NumberFormat('zh-CN', {
    minimumFractionDigits: 0,
    maximumFractionDigits: Math.max(0, maxFractionDigits),
  }).format(n)
}

export function normalizeUnit(unit?: string | null, fallback = ''): string {
  const raw = String(unit ?? '').trim()
  if (!raw) return fallback
  const normalized = raw.replace(/[0-9.\s]/g, '')
  return normalized || fallback
}

export function formatAmount(value: unknown, unit?: string | null, fallbackUnit = ''): string {
  const unitText = normalizeUnit(unit, fallbackUnit)
  const num = formatNumber(value)
  if (num === '--') return '--'
  return unitText ? `${num} ${unitText}` : num
}

export function formatRatio(remaining: unknown, capacity: unknown, unit?: string | null, fallbackUnit = ''): string {
  const r = formatNumber(remaining)
  const c = formatNumber(capacity)
  const unitText = normalizeUnit(unit, fallbackUnit)
  if (r === '--' || c === '--') return '--'
  return unitText ? `${r}/${c} ${unitText}` : `${r}/${c}`
}
