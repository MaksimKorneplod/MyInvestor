export default function CandleChart({ candles }) {
  const bars = candles.map(c => {
    const o = Number(c.open)
    const cl = Number(c.close)
    if (!Number.isFinite(o) || !Number.isFinite(cl)) return null
    const hi = c.high == null ? Math.max(o, cl) : Number(c.high)
    const lo = c.low == null ? Math.min(o, cl) : Number(c.low)
    return { o, c: cl, h: Number.isFinite(hi) ? hi : Math.max(o, cl), l: Number.isFinite(lo) ? lo : Math.min(o, cl), t: c.begin }
  }).filter(Boolean)

  if (!bars.length) {
    return <div className="emptybox"><b>Свечей за этот период нет</b><p>Пустой ряд — нормальный ответ, а не сбой биржи.</p></div>
  }

  const VW = 900
  const VH = 280
  const pad = 16
  const lo = Math.min(...bars.map(b => b.l))
  const hi = Math.max(...bars.map(b => b.h))
  const span = (hi - lo) || 1
  const step = (VW - pad * 2) / bars.length
  const bw = Math.max(2, step * 0.62)
  const y = v => pad + (1 - (v - lo) / span) * (VH - pad * 2)
  const x = i => pad + step * (i + 0.5)

  return (
    <svg className="chart" viewBox={`0 0 ${VW} ${VH}`} role="img" aria-label="Свечной график">
      {bars.map((b, i) => {
        const up = b.c >= b.o
        const top = y(Math.max(b.o, b.c))
        const height = Math.max(1.2, Math.abs(y(b.o) - y(b.c)))
        return (
          <g key={b.t || i}>
            <line x1={x(i)} x2={x(i)} y1={y(b.h)} y2={y(b.l)} stroke={up ? 'var(--up)' : 'var(--down)'} strokeWidth="1.2" />
            <rect x={x(i) - bw / 2} y={top} width={bw} height={height} fill={up ? 'var(--up)' : 'var(--down)'} rx="1" />
          </g>
        )
      })}
    </svg>
  )
}
