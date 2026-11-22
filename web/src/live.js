import { history, instruments, news, quoteBatch } from './api'
import { MOCK_NEWS, PERIODS } from './mock'

export async function loadNews() {
  try {
    const data = await news(5)
    if (!data?.items?.length) return null
    const arts = ['desk', 'flag', 'bull', 'desk', 'flag']
    return data.items.slice(0, 5).map((item, i) => ({
      id: i + 1,
      cat: i % 2 ? 'Бизнес' : 'Экономика',
      t: item.title,
      m: item.pubDate || 'Интерфакс',
      art: arts[i % arts.length],
      link: item.link,
      body: [['Материал', item.description || 'Новость Интерфакса.'], ['Источник', item.link]],
    }))
  } catch {
    return null
  }
}

export async function loadCatalog(q = '') {
  try {
    const page = await instruments({ q, limit: 50, offset: 0 })
    const items = page?.items || []
    if (!items.length && q) return []
    if (!items.length) return null
    let quotes = []
    try { quotes = await quoteBatch(items.map(item => item.ticker)) } catch { quotes = [] }
    const by = Object.fromEntries((quotes || []).map(row => [row.ticker, row]))
    return items.map((item, i) => {
      const qrow = by[item.ticker]
      const last = qrow?.last != null ? Number(qrow.last) : null
      const pct = qrow?.changePercent != null ? Number(String(qrow.changePercent).replace('%', '').replace(',', '.')) : 0
      const price = Number.isFinite(last) ? last : 0
      const prev = Number.isFinite(pct) && price ? price / (1 + pct / 100) : price
      return {
        t: item.ticker,
        n: item.name || item.ticker,
        sec: 'fin',
        p: price || 1,
        pc: prev || price || 1,
        seed: 200 + i,
        bg: '#3B2A36',
        logo: 'letter',
        pe: 0,
        cap: 'MOEX',
        rng: [price * .99 || 1, price * 1.01 || 1],
        live: true,
        quoteTime: qrow?.quoteTime || qrow?.receivedAt || '',
      }
    })
  } catch {
    return null
  }
}

export async function loadCloses(ticker, periodKey) {
  try {
    const data = await history(ticker, PERIODS[periodKey]?.api || 'month')
    const values = (data?.candles || []).map(c => Number(c.close)).filter(Number.isFinite)
    return values.length ? values : null
  } catch {
    return null
  }
}

export { MOCK_NEWS }
