const priceFmt = new Intl.NumberFormat('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
const pctFmt = new Intl.NumberFormat('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })

export function money(value) {
  if (value == null || value === '') return '—'
  const n = Number(value)
  if (!Number.isFinite(n)) return '—'
  return `${priceFmt.format(n)} ₽`
}

export function percent(value) {
  if (value == null || value === '') return null
  const n = Number(String(value).replace('%', '').replace(',', '.').trim())
  if (!Number.isFinite(n)) return null
  const sign = n > 0 ? '+' : n < 0 ? '−' : ''
  return { n, text: `${sign}${pctFmt.format(Math.abs(n))}%` }
}

export function when(iso) {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '—'
  return d.toLocaleString('ru-RU', { timeZone: 'Europe/Moscow', day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })
}

export function initials(value) {
  const s = String(value || '').replace(/\s/g, '')
  return (s.slice(0, 2) || '—').toUpperCase()
}

export const PERIODS = [
  ['day', '1Д'],
  ['week', '1Н'],
  ['month', '1М'],
  ['halfYear', '6М'],
  ['year', '1Г'],
  ['fiveYears', '5Л'],
]
