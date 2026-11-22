import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { ApiError, instruments, quoteBatch } from '../api'
import { money, percent, when } from '../format'
import { Icon } from '../icons'

const LIMIT = 50

export default function MarketPage() {
  const [params, setParams] = useSearchParams()
  const q = params.get('q') || ''
  const offset = Math.max(0, Number(params.get('offset')) || 0)
  const [draft, setDraft] = useState(q)
  const [page, setPage] = useState(null)
  const [quotes, setQuotes] = useState({})
  const [error, setError] = useState('')
  const [priceError, setPriceError] = useState('')
  const [tick, setTick] = useState(0)

  useEffect(() => { setDraft(q) }, [q])

  useEffect(() => {
    const id = setTimeout(() => {
      const nextQuery = draft.trim()
      if (nextQuery === q) return
      const next = new URLSearchParams(params)
      if (nextQuery) next.set('q', nextQuery)
      else next.delete('q')
      next.delete('offset')
      setParams(next, { replace: true })
    }, 300)
    return () => clearTimeout(id)
  }, [draft]) // q and params are read when the debounce fires

  useEffect(() => {
    let live = true
    setError('')
    setPriceError('')
    if ([...q].length > 100) {
      setPage({ items: [], total: 0, limit: LIMIT, offset: 0 })
      setError('Запрос длиннее 100 символов')
      return
    }
    instruments({ q, limit: LIMIT, offset })
      .then(async data => {
        if (!live) return
        const items = Array.isArray(data?.items) ? data.items : []
        setPage({ items, total: data?.total || 0, limit: data?.limit || LIMIT, offset: data?.offset || 0 })
        if (!items.length) { setQuotes({}); return }
        try {
          const rows = await quoteBatch(items.map(item => item.ticker))
          if (!live) return
          const byTicker = {}
          for (const row of rows || []) byTicker[row.ticker] = row
          setQuotes(byTicker)
        } catch (err) {
          if (!live) return
          setQuotes({})
          setPriceError(err instanceof ApiError ? err.message : 'Цены сейчас недоступны')
        }
      })
      .catch(err => {
        if (!live) return
        setPage({ items: [], total: 0, limit: LIMIT, offset })
        setError(err instanceof ApiError ? err.message : 'Каталог недоступен')
      })
    return () => { live = false }
  }, [q, offset, tick])

  function move(nextOffset) {
    const next = new URLSearchParams(params)
    if (nextOffset > 0) next.set('offset', String(nextOffset))
    else next.delete('offset')
    setParams(next)
  }

  const items = page?.items || []
  const total = page?.total || 0
  const from = total ? offset + 1 : 0
  const to = Math.min(offset + items.length, total)

  return (
    <>
      <div className="pagehead">
        <h1>Рынок</h1>
        <span className="sub">Акции MOEX · поиск по тикеру и названию · цена последней сделки</span>
        <div className="right">
          <button className="btn btn-ghost btn-sm" type="button" onClick={() => setTick(n => n + 1)}><Icon name="refresh" /> Обновить</button>
        </div>
      </div>
      <div className="card">
        <div className="toolbar">
          <label className="inline-search">
            <Icon name="search" />
            <input value={draft} placeholder="Тикер или название: SBER, Газпром…" onChange={e => setDraft(e.target.value)} aria-label="Поиск по тикеру или названию" />
          </label>
          {page && <span className="dim">{from}–{to} из {total}</span>}
        </div>
        {page === null && !error && <p className="muted">Загружаем каталог…</p>}
        {error && <p className="aerr">{error}</p>}
        {priceError && <p className="aerr">{priceError}</p>}
        {page && !error && !items.length && <div className="emptybox"><b>Ничего не найдено</b><p>{q ? 'Нет акций с таким тикером или названием.' : 'Каталог сейчас пуст.'}</p></div>}
        {!!items.length && (
          <div className="tablewrap">
            <table className="qtable">
              <thead><tr><th>Тикер</th><th>Название</th><th>Цена</th><th>Изменение</th><th>Сделка</th></tr></thead>
              <tbody>
                {items.map(item => {
                  const quote = quotes[item.ticker]
                  const pct = percent(quote?.changePercent)
                  return (
                    <tr key={item.id || item.ticker}>
                      <td><Link className="ticker" to={`/instrument/${encodeURIComponent(item.ticker)}`}>{item.ticker}</Link></td>
                      <td>{item.name}</td>
                      <td className="mono">{quote ? money(quote.last) : '—'}</td>
                      <td className={pct ? (pct.n < 0 ? 'down' : 'up') : ''}>{pct ? pct.text : '—'}</td>
                      <td className="dim">{quote ? when(quote.quoteTime || quote.receivedAt) : '—'}</td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
        {total > LIMIT && (
          <div className="pager">
            <button className="btn btn-ghost btn-sm" type="button" disabled={offset === 0} onClick={() => move(Math.max(0, offset - LIMIT))}>Назад</button>
            <button className="btn btn-ghost btn-sm" type="button" disabled={offset + LIMIT >= total} onClick={() => move(offset + LIMIT)}>Дальше</button>
          </div>
        )}
      </div>
    </>
  )
}
