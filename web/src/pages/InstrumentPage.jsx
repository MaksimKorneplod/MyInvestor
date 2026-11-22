import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ApiError, history, instruments, oneQuote } from '../api'
import { PERIODS, money, percent, when } from '../format'
import { Icon } from '../icons'
import CandleChart from '../components/CandleChart'

export default function InstrumentPage() {
  const { ticker = '' } = useParams()
  const symbol = decodeURIComponent(ticker).toUpperCase()
  const [name, setName] = useState('')
  const [period, setPeriod] = useState('day')
  const [quote, setQuote] = useState(undefined)
  const [quoteErr, setQuoteErr] = useState('')
  const [candles, setCandles] = useState(null)
  const [histErr, setHistErr] = useState('')

  useEffect(() => {
    let live = true
    setName('')
    instruments({ q: symbol, limit: 20, offset: 0 })
      .then(data => {
        if (!live) return
        const hit = (data?.items || []).find(item => String(item.ticker).toUpperCase() === symbol)
        if (hit?.name) setName(hit.name)
      })
      .catch(() => {})
    return () => { live = false }
  }, [symbol])

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
      .then(data => { if (live) setCandles(Array.isArray(data?.candles) ? data.candles : []) })
      .catch(err => { if (live) { setCandles([]); setHistErr(err.message || 'История недоступна') } })
    return () => { live = false }
  }, [symbol, period])

  const pct = percent(quote?.changePercent)

  return (
    <>
      <div className="pagehead">
        <h1>{name ? `${name} · ${symbol}` : symbol}</h1>
        <span className="sub">Акция MOEX · {when(quote?.quoteTime || quote?.receivedAt)}</span>
        <div className="right">
          <Link className="btn btn-ghost btn-sm" to="/market"><Icon name="arrL" /> К рынку</Link>
        </div>
      </div>
      <div className="quotehead card">
        {quoteErr && <p className="aerr">{quoteErr}</p>}
        {!quoteErr && quote === undefined && <p className="muted">Загружаем цену…</p>}
        {quote && (
          <>
            <div className="qprice">{money(quote.last)}</div>
            <div className={pct ? (pct.n < 0 ? 'down' : 'up') : 'dim'}>{pct ? pct.text : 'Изменение не пришло'}</div>
          </>
        )}
      </div>
      <div className="card" style={{ marginTop: 16 }}>
        <div className="chips" style={{ marginBottom: 14 }}>
          {PERIODS.map(([id, label]) => (
            <button key={id} type="button" className={period === id ? 'chip on' : 'chip'} onClick={() => setPeriod(id)}>{label}</button>
          ))}
        </div>
        {histErr && <p className="aerr">{histErr}</p>}
        {candles === null && !histErr && <p className="muted">Загружаем свечи…</p>}
        {candles && <CandleChart candles={candles} />}
      </div>
    </>
  )
}
