import { useState } from 'react'
import { allocOf, fmt0, fmt2, fmtK, fmtMoney, investedVal, PERIODS, periodLabels, series, totalVal } from './mock'

export function Spark({ asset, tickSeed }) {
  const c = (asset.p / asset.pc - 1) * 100
  const up = c >= 0
  const s = series(asset.seed + tickSeed * 7, 18, c * 4, 3)
  const n = s.length
  const mn = Math.min(...s)
  const mx = Math.max(...s)
  const pad = (mx - mn) * .15 || 1
  const X = i => 4 + i * (232 / (n - 1))
  const Y = v => 60 - 6 - ((v - mn + pad) / (mx - mn + 2 * pad)) * (60 - 12)
  let d = `M ${X(0)} ${Y(s[0])}`
  for (let i = 1; i < n; i++) {
    const m = (X(i - 1) + X(i)) / 2
    d += ` C ${m} ${Y(s[i - 1])}, ${m} ${Y(s[i])}, ${X(i)} ${Y(s[i])}`
  }
  const col = up ? 'var(--up)' : 'var(--down)'
  return (
    <svg className="fspark" viewBox="0 0 240 60" preserveAspectRatio="none">
      <path d={`${d} L ${X(n - 1)} 60 L ${X(0)} 60 Z`} fill={`url(#${up ? 'sparkUp' : 'sparkDown'})`} />
      <path d={d} fill="none" stroke={col} strokeWidth="2.2" strokeLinecap="round" />
      <circle cx={X(n - 1)} cy={Y(s[n - 1])} r="3.4" fill={col} stroke="#211721" strokeWidth="1.6" />
    </svg>
  )
}

export function LineChart({ vals, labels, money = true, height }) {
  if (!vals?.length) return null
  const VW = 320
  const VH = 170
  const mn = Math.min(...vals)
  const mx = Math.max(...vals)
  const pad = (mx - mn) * .12 || 1
  const L = 46
  const R = VW - 8
  const T = 10
  const B = VH - 16
  const X = i => L + i * (R - L) / (vals.length - 1)
  const Y = v => T + (1 - (v - mn + pad) / (mx - mn + 2 * pad)) * (B - T)
  let d = `M ${X(0)} ${Y(vals[0])}`
  for (let i = 1; i < vals.length; i++) {
    const m = (X(i - 1) + X(i)) / 2
    d += ` C ${m} ${Y(vals[i - 1])}, ${m} ${Y(vals[i])}, ${X(i)} ${Y(vals[i])}`
  }
  const up = vals[vals.length - 1] >= vals[0]
  const col = up ? 'var(--up)' : 'var(--down)'
  const label = money ? fmtK : v => fmt2(v)
  const [tip, setTip] = useState(null)
  return (
    <div className="chartwrap" style={height ? { height } : undefined}>
      <svg viewBox={`0 0 ${VW} ${VH}`} preserveAspectRatio="none" onPointerMove={e => {
        const r = e.currentTarget.getBoundingClientRect()
        const x = (e.clientX - r.left) / r.width * VW
        const i = Math.max(0, Math.min(vals.length - 1, Math.round((x - L) / (R - L) * (vals.length - 1))))
        setTip({ i, left: X(i) / VW * 100, top: Y(vals[i]) / VH * 100 })
      }} onPointerLeave={() => setTip(null)}>
        {[0, .5, 1].map(f => {
          const y = T + f * (B - T)
          const v = mx + pad - f * (mx - mn + 2 * pad)
          return <g key={f}><line x1={L} y1={y} x2={R} y2={y} stroke="#3B2A36" strokeDasharray="2 4" /><text x={L - 6} y={y + 3} textAnchor="end" fontSize="8.5" fill="#BCA9B2" fontFamily="JetBrains Mono,monospace">{label(v)}</text></g>
        })}
        <path d={`${d} L ${X(vals.length - 1)} ${B} L ${X(0)} ${B} Z`} fill={`url(#${up ? 'sparkUp' : 'sparkDown'})`} />
        <path d={d} fill="none" stroke={col} strokeWidth="2.4" strokeLinecap="round" />
        <circle cx={X(vals.length - 1)} cy={Y(vals[vals.length - 1])} r="3.6" fill={col} stroke="#211721" strokeWidth="1.6" />
        {tip && <g><line x1={X(tip.i)} x2={X(tip.i)} y1={T} y2={B} stroke="#BCA9B2" strokeDasharray="3 3" /><circle cx={X(tip.i)} cy={Y(vals[tip.i])} r="4" fill={col} /></g>}
      </svg>
      <div className="ctip" style={{ opacity: tip ? 1 : 0, left: tip ? Math.min(86, Math.max(14, tip.left)) + '%' : '50%', top: tip ? tip.top + '%' : '40%' }}>{tip ? `${labels[tip.i]} · ${fmt2(vals[tip.i])}` : ''}</div>
    </div>
  )
}

export function portfolioSeries(state) {
  const spec = PERIODS[state.period]
  const total = totalVal(state)
  const start = total / (1 + spec.pct / 100)
  const vals = series(spec.seed + state.tickSeed * 3, spec.n, total - start, total * .008).map(v => start + v)
  return { vals, labels: periodLabels(state.period, spec.n), pct: spec.pct }
}

export function Donut({ state }) {
  const parts = allocOf(state)
  const C = 2 * Math.PI * 64
  const [hot, setHot] = useState(null)
  let cum = 0
  const segs = parts.map((s, i) => {
    const f = s.pct / 100
    const off = cum * C
    const dash = Math.max(f * C - 2.5, .5)
    const node = <circle key={s.k} className="dseg" cx="95" cy="95" r="64" fill="none" stroke={s.c} strokeWidth="26" strokeDasharray={`${dash} ${C - dash}`} strokeDashoffset={-off} transform="rotate(-90 95 95)" onMouseEnter={() => setHot(i)} />
    cum += f
    return node
  })
  const current = hot != null ? parts[hot] : null
  return (
    <>
      <div className="donutwrap" onMouseLeave={() => setHot(null)}>
        <svg className="js-donut" viewBox="0 0 190 190">{segs.length ? segs : <circle cx="95" cy="95" r="64" fill="none" stroke="#3B2A36" strokeWidth="26" />}</svg>
        <div className="dcenter">
          <b>{current ? fmt0(current.v) : fmtMoney(investedVal(state), state.settings.compact)}</b>
          <span style={{ color: current?.c }}>{current ? `${current.n} · ${Math.round(current.pct)}%` : 'инвестировано'}</span>
        </div>
      </div>
      <div className="dlegend">
        {parts.map((s, i) => (
          <div key={s.k} className={hot === i ? 'dleg hot' : 'dleg'} onMouseEnter={() => setHot(i)}>
            <i style={{ background: s.c }} />
            <div><div className="n">{s.n}</div><div className="p">{Math.round(s.pct)}%</div><div className="a">{fmt0(s.v)}</div></div>
          </div>
        ))}
      </div>
    </>
  )
}

export function assetSeries(asset, period, tickSeed, liveCloses) {
  if (liveCloses?.length) return { vals: liveCloses, labels: liveCloses.map((_, i) => String(i + 1)) }
  const spec = PERIODS[period]
  const start = asset.p / (1 + spec.pct / 100)
  const vals = series(asset.seed + tickSeed * 7, spec.n, asset.p - start, asset.p * .01).map(v => start + v)
  return { vals, labels: periodLabels(period, spec.n) }
}
