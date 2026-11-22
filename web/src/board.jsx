import { useEffect, useState } from 'react'
import { ApiError, history, instruments, oneQuote, quoteBatch } from './api'
import { money, percent, when } from './format'
import { Icon, Mark } from './icons'
import { INSTRUMENTS, chgOf, fmt2, fmtPct, series } from './mock'
import { useStudio } from './studio'

const PAGE = 48
const PERIODS = [
  ['day', '1Д'],
  ['week', '1Н'],
  ['month', '1М'],
  ['halfYear', '6М'],
  ['year', '1Г'],
  ['fiveYears', '5Л'],
]

function rub(value) {
  return money(value)
}

export function MarketScreen() {
  return (
    <>
      <div className="pagehead enter">
        <h1>Рынок</h1>
        <span className="sub">Акции MOEX · поиск по тикеру и названию идёт по всему каталогу</span>
      </div>
      <MoexBoard />
    </>
  )
}

export function MoexBoard() {
  const { actions } = useStudio()
  const [draft, setDraft] = useState('')
  const [query, setQuery] = useState('')
  const [offset, setOffset] = useState(0)
  const [page, setPage] = useState(null)
  const [quotes, setQuotes] = useState({})
  const [error, setError] = useState('')
  const [tip, setTip] = useState(null)

  useEffect(() => {
    const id = setTimeout(() => {
      setQuery(draft.trim())
      setOffset(0)
    }, 300)
    return () => clearTimeout(id)
  }, [draft])

  useEffect(() => {
    let live = true
    setError('')
    if ([...query].length > 100) {
      setPage({ items: [], total: 0 })
      setError('Запрос длиннее 100 символов')
      return
    }
    instruments({ q: query, limit: PAGE, offset })
      .then(async data => {
        if (!live) return
        const items = (data?.items || []).map(item => ({
          t: item.ticker,
          n: item.name || item.ticker,
          live: true,
          bg: '#3B2A36',
          logo: 'letter',
        }))
        setPage({ items, total: data?.total || 0, fallback: false })
        if (!items.length) { setQuotes({}); return }
        try {
          const rows = await quoteBatch(items.map(item => item.t))
          if (!live) return
          const by = {}
          for (const row of rows || []) by[row.ticker] = row
          setQuotes(by)
        } catch {
          if (live) setQuotes({})
        }
      })
      .catch(() => {
        if (!live) return
        const q = query.toLowerCase()
        const items = INSTRUMENTS.filter(item => !q || item.n.toLowerCase().includes(q) || item.t.toLowerCase().includes(q))
        setPage({ items, total: items.length, fallback: true })
        setQuotes({})
        setError('Каталог MOEX сейчас недоступен, показан список макета')
      })
    return () => { live = false }
  }, [query, offset])

  const items = page?.items || []
  const total = page?.total || 0

  return (
    <>
      <div className="toolbar enter" style={{ '--d': '.1s' }}>
        <div className="tsearch" style={{ flex: 1, maxWidth: 'none' }}>
          <Icon name="search" />
          <input value={draft} placeholder="Поиск по всем акциям: SBER, Газпром…" aria-label="Поиск по всем акциям" onChange={e => setDraft(e.target.value)} />
        </div>
        {page && !page.fallback && <span className="dim">{total ? offset + 1 : 0}–{Math.min(offset + items.length, total)} из {total}</span>}
      </div>
      {page === null && !error && <p className="muted">Загружаем каталог MOEX…</p>}
      {error && <p className="aerr">{error}</p>}
      {page && !items.length && <div className="emptybox"><Icon name="search" /><b>Ничего не найдено</b><p>Нет акций с таким тикером или названием.</p></div>}
      <div className="stockgrid">
        {items.map(item => {
          const quote = quotes[item.t]
          const pct = item.live ? percent(quote?.changePercent) : { n: chgOf(item), text: fmtPct(chgOf(item)) }
          const price = item.live ? (quote?.last != null ? rub(quote.last) : '—') : fmt2(item.p)
          const open = tip === item.t
          return (
            <button key={item.t} type="button" className="favcard stockcard" onClick={() => actions.openTicker(item.t, item.n)} onMouseEnter={() => setTip(item.t)} onMouseLeave={() => setTip(null)} onFocus={() => setTip(item.t)} onBlur={() => setTip(null)}>
              <div className="fhead"><Mark asset={item} /><div><div className="fname">{item.n}</div><div className="ftick">{item.t} · {item.live ? 'MOEX' : 'макет'}</div></div></div>
              <div className="fprice">{price}</div>
              <div className={`fchange ${pct ? (pct.n < 0 ? 'down' : 'up') : ''}`}>{pct ? pct.text : '—'}</div>
              {open && (
                <div className="stocktip" role="tooltip">
                  <b>{item.n}</b>
                  <span>{item.t}</span>
                  <span>{price}{pct ? ` · ${pct.text}` : ''}</span>
                  {quote && <span>{when(quote.quoteTime || quote.receivedAt)}</span>}
                </div>
              )}
            </button>
          )
        })}
      </div>
      {page && !page.fallback && total > PAGE && (
        <div className="pager">
          <button className="btn btn-ghost btn-sm" type="button" disabled={offset === 0} onClick={() => setOffset(value => Math.max(0, value - PAGE))}>Назад</button>
          <button className="btn btn-ghost btn-sm" type="button" disabled={offset + PAGE >= total} onClick={() => setOffset(value => value + PAGE)}>Дальше</button>
        </div>
      )}
    </>
  )
}

export function TickerPage() {
  const { model, actions } = useStudio()
  const symbol = model.ticker
  const [name, setName] = useState(model.tickerName || '')
  const [period, setPeriod] = useState('month')
  const [quote, setQuote] = useState(undefined)
  const [quoteErr, setQuoteErr] = useState('')
  const [candles, setCandles] = useState(null)
  const [histErr, setHistErr] = useState('')

  useEffect(() => {
    let live = true
    setName(model.tickerName || '')
    instruments({ q: symbol, limit: 20, offset: 0 })
      .then(data => {
        if (!live) return
        const hit = (data?.items || []).find(item => String(item.ticker).toUpperCase() === symbol)
        if (hit?.name) setName(hit.name)
      })
      .catch(() => {})
    return () => { live = false }
  }, [symbol, model.tickerName])

  useEffect(() => {
    let live = true
    setQuote(undefined)
    setQuoteErr('')
    oneQuote(symbol)
      .then(data => { if (live) setQuote(data) })
      .catch(err => { if (live) { setQuote(null); setQuoteErr(err instanceof ApiError && err.status === 404 ? 'Бумага не найдена' : (err.message || 'Котировка недоступна')) } })
    return () => { live = false }
  }, [symbol])

  useEffect(() => {
    let live = true
    setCandles(null)
    setHistErr('')
    history(symbol, period)
      .then(data => {
        if (!live) return
        const rows = Array.isArray(data?.candles) ? data.candles : []
        setCandles(rows.length ? rows : fallbackCandles(symbol, period))
      })
      .catch(err => {
        if (!live) return
        const fallback = fallbackCandles(symbol, period)
        setCandles(fallback)
        if (!fallback.length) setHistErr(err.message || 'История недоступна')
      })
    return () => { live = false }
  }, [symbol, period])

  const pct = percent(quote?.changePercent)
  const mock = INSTRUMENTS.find(item => item.t === symbol)

  return (
    <>
      <div className="pagehead enter">
        <h1>{name ? `${name} · ${symbol}` : symbol}</h1>
        <div className="right">
          <button className="btn btn-ghost btn-sm" type="button" onClick={() => actions.closeTicker()}><Icon name="arrL" /> К рынку</button>
        </div>
        <span className="sub">Акция MOEX · {when(quote?.quoteTime || quote?.receivedAt)}</span>
      </div>
      <div className="card quotehead enter">
        {quoteErr && <p className="aerr">{quoteErr}</p>}
        {!quoteErr && quote === undefined && <p className="muted">Загружаем цену…</p>}
        {quote && (
          <>
            <div className="qprice">{rub(quote.last)}</div>
            <div className={pct ? (pct.n < 0 ? 'down' : 'up') : 'dim'}>{pct ? pct.text : 'Изменение не пришло'}</div>
          </>
        )}
        {!quote && quoteErr && mock && <div className="qprice">{fmt2(mock.p)}</div>}
      </div>
      <div className="card enter" style={{ marginTop: 16 }}>
        <div className="ctabs" style={{ marginBottom: 14 }}>
          {PERIODS.map(([id, label]) => (
            <button key={id} type="button" className={period === id ? 'on' : ''} onClick={() => setPeriod(id)}>{label}</button>
          ))}
        </div>
        {histErr && <p className="aerr">{histErr}</p>}
        {candles === null && !histErr && <p className="muted">Загружаем свечи…</p>}
        {candles && <CandleBoard candles={candles} period={period} />}
      </div>
    </>
  )
}

function fallbackCandles(symbol, period) {
  const asset = INSTRUMENTS.find(item => item.t === symbol)
  if (!asset) return []
  const spec = { day: [26, .4], week: [28, 1.1], month: [30, 3.9], halfYear: [32, 8.4], year: [34, 21], fiveYears: [34, 21] }[period] || [30, 3.9]
  const start = asset.p / (1 + spec[1] / 100)
  const vals = series(asset.seed, spec[0], asset.p - start, asset.p * .01).map(value => start + value)
  const end = new Date(2026, 10, 22).getTime()
  return vals.map((value, index) => ({
    begin: new Date(end - (vals.length - 1 - index) * 86400000).toISOString(),
    open: value,
    high: value * 1.008,
    low: value * 0.992,
    close: value,
  }))
}

function CandleBoard({ candles, period }) {
  const bars = candles.map(candle => {
    const open = Number(candle.open)
    const close = Number(candle.close)
    if (!Number.isFinite(open) || !Number.isFinite(close)) return null
    const high = candle.high == null ? Math.max(open, close) : Number(candle.high)
    const low = candle.low == null ? Math.min(open, close) : Number(candle.low)
    return {
      o: open,
      c: close,
      h: Number.isFinite(high) ? high : Math.max(open, close),
      l: Number.isFinite(low) ? low : Math.min(open, close),
      t: candle.begin,
    }
  }).filter(Boolean)
  const [tip, setTip] = useState(null)

  if (!bars.length) {
    return <div className="emptybox"><b>Свечей за этот период нет</b><p>Пустой ряд — нормальный ответ, а не сбой биржи.</p></div>
  }

  const VW = 960
  const VH = 420
  const L = 72
  const R = VW - 16
  const T = 18
  const B = VH - 36
  const lo = Math.min(...bars.map(bar => bar.l))
  const hi = Math.max(...bars.map(bar => bar.h))
  const span = (hi - lo) || 1
  const y = value => T + (1 - (value - lo) / span) * (B - T)
  const step = (R - L) / bars.length
  const bw = Math.max(2, step * 0.62)
  const x = index => L + step * (index + 0.5)
  const grids = [0, .25, .5, .75, 1]
  const vstep = Math.max(1, Math.round(bars.length / 6))

  function move(event) {
    const rect = event.currentTarget.getBoundingClientRect()
    const px = (event.clientX - rect.left) / rect.width * VW
    const index = Math.max(0, Math.min(bars.length - 1, Math.floor((px - L) / step)))
    setTip({ index, left: x(index) / VW * 100, top: y(bars[index].h) / VH * 100 })
  }

  const hot = tip ? bars[tip.index] : null

  return (
    <div className="chartwrap tickerchart">
      <svg viewBox={`0 0 ${VW} ${VH}`} role="img" aria-label="Свечной график" onPointerMove={move} onPointerLeave={() => setTip(null)}>
        {grids.map(part => {
          const value = hi - part * span
          const yy = y(value)
          return (
            <g key={part}>
              <line x1={L} x2={R} y1={yy} y2={yy} stroke="#3B2A36" strokeDasharray="3 4" />
              <text x={L - 8} y={yy + 3} textAnchor="end" fontSize="11" fill="#BCA9B2" fontFamily="JetBrains Mono,monospace">{rub(value)}</text>
            </g>
          )
        })}
        {bars.map((bar, index) => index % vstep === 0 ? <line key={`v${bar.t || index}`} x1={x(index)} x2={x(index)} y1={T} y2={B} stroke="#3B2A36" strokeDasharray="2 5" /> : null)}
        <line x1={L} x2={R} y1={B} y2={B} stroke="#4E3644" />
        {bars.map((bar, index) => {
          const up = bar.c >= bar.o
          const top = y(Math.max(bar.o, bar.c))
          const height = Math.max(1.4, Math.abs(y(bar.o) - y(bar.c)))
          return (
            <g key={bar.t || index}>
              <line x1={x(index)} x2={x(index)} y1={y(bar.h)} y2={y(bar.l)} stroke={up ? 'var(--up)' : 'var(--down)'} strokeWidth="1.4" />
              <rect x={x(index) - bw / 2} y={top} width={bw} height={height} fill={up ? 'var(--up)' : 'var(--down)'} rx="1" />
            </g>
          )
        })}
        {[0, Math.floor(bars.length / 2), bars.length - 1].map(index => (
          <text key={index} x={x(index)} y={VH - 12} textAnchor="middle" fontSize="11" fill="#BCA9B2" fontFamily="JetBrains Mono,monospace">{axisLabel(bars[index].t, period)}</text>
        ))}
        {hot && <line x1={x(tip.index)} x2={x(tip.index)} y1={T} y2={B} stroke="#F6EEF1" strokeDasharray="3 3" />}
      </svg>
      <div className="ctip" style={{ opacity: hot ? 1 : 0, left: tip ? `${Math.min(82, Math.max(18, tip.left))}%` : '50%', top: tip ? `${Math.max(8, tip.top)}%` : '12%' }}>
        {hot ? `${axisLabel(hot.t, period)} · O ${rub(hot.o)} · H ${rub(hot.h)} · L ${rub(hot.l)} · C ${rub(hot.c)}` : ''}
      </div>
    </div>
  )
}

function axisLabel(iso, period) {
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''
  if (period === 'day') return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
  return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })
}
