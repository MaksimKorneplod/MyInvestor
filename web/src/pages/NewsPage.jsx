import { useEffect, useState } from 'react'
import { ApiError, news } from '../api'
import { when } from '../format'
import { Icon } from '../icons'

export default function NewsPage() {
  const [data, setData] = useState(null)
  const [error, setError] = useState('')
  const [tick, setTick] = useState(0)

  useEffect(() => {
    let live = true
    setError('')
    news()
      .then(payload => { if (live) setData(payload) })
      .catch(err => { if (live) { setData(null); setError(err instanceof ApiError ? err.message : 'Лента недоступна') } })
    return () => { live = false }
  }, [tick])

  const items = data?.items || []
  return (
    <>
      <div className="pagehead">
        <h1>Новости</h1>
        <span className="sub">Интерфакс · Экономика и Бизнес</span>
        <div className="right">
          <button className="btn btn-ghost btn-sm" type="button" onClick={() => setTick(n => n + 1)}><Icon name="refresh" /> Обновить</button>
        </div>
      </div>
      {data === null && !error && <p className="muted">Загружаем ленту…</p>}
      {error && <p className="aerr">{error}</p>}
      {data && !items.length && <div className="emptybox"><b>Новостей нет</b><p>В ленте нет материалов этих категорий.</p></div>}
      <div className="newslist">
        {items.map(item => (
          <a key={item.link} className="card newsitem" href={item.link} target="_blank" rel="noreferrer">
            <div className="dim">{when(item.pubDate)}</div>
            <h2>{item.title}</h2>
            {item.description && <p>{item.description}</p>}
          </a>
        ))}
      </div>
    </>
  )
}
