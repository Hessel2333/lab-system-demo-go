export interface CabinetLike {
  name?: string | null
  location?: string | null
}

const splitBySemanticSeparators = (raw: string) =>
  raw
    .split(/[-/·—]/)
    .map((part) => part.trim())
    .filter(Boolean)

export const extractCabinetCode = (name?: string | null) => {
  const raw = String(name || '').trim()
  if (!raw) return ''

  const parts = splitBySemanticSeparators(raw)
  if (parts.length === 0) return raw

  const withCabinetKeyword = parts.filter((part) => part.includes('柜'))
  if (withCabinetKeyword.length > 0) {
    return withCabinetKeyword.join('-')
  }

  return parts[parts.length - 1] || raw
}

export const formatCabinetDisplayName = (cabinet?: CabinetLike | null) => {
  if (!cabinet) return '—'
  const room = String(cabinet.location || '').trim()
  const code = extractCabinetCode(cabinet.name)

  if (room && code) {
    if (code.startsWith(`${room}-`) || code === room) return code
    return `${room}-${code}`
  }

  return code || room || '未命名柜'
}

export const formatCabinetDisplayLines = (cabinet?: CabinetLike | null) => {
  const display = formatCabinetDisplayName(cabinet)
  if (!display || display === '—') return ['—']

  const parts = display
    .split('-')
    .map((part) => part.trim())
    .filter(Boolean)

  if (parts.length <= 1) return [display]
  return [parts[0]!, parts.slice(1).join('-')]
}
