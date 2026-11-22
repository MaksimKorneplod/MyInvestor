import { createContext, useContext, useEffect, useMemo, useRef, useState } from 'react'
import { quoteBatch } from './api'
import { loadCatalog, loadCloses, loadNews } from './live'
import { FEE, INSTRUMENTS, LESSONS, SECTORS, lcg, seedAccount, uid } from './mock'

const KEY = 'myinvestor-redesign-v2'
const StudioContext = createContext(null)
export function useStudio() { return useContext(StudioContext) }

function loadAccount() {
  try {
    const raw = JSON.parse(localStorage.getItem(KEY) || 'null')
    if (raw?.v === 2 && Array.isArray(raw.instruments) && raw.holdings && raw.settings) {
      return { ...seedAccount(), ...raw, instruments: raw.instruments, news: raw.news?.length ? raw.news : seedAccount().news }
    }
  } catch { /* стартовые данные макета */ }
  return seedAccount()
}

function persist(account) {
  const { catalog, ...rest } = account
  localStorage.setItem(KEY, JSON.stringify({ ...rest, v: 2 }))
}

function clone(value) { return JSON.parse(JSON.stringify(value)) }

export function findAsset(model, ticker) {
  return (model.catalog || []).find(item => item.t === ticker) || model.account.instruments.find(item => item.t === ticker)
}

export function marketRows(model) {
  const source = model.catalog?.length ? model.catalog : model.account.instruments
  const m = model.account.mkt
  const q = (m.q || '').toLowerCase()
  const list = source.filter(item => (m.sec === 'all' || item.sec === m.sec) && (!q || item.n.toLowerCase().includes(q) || item.t.toLowerCase().includes(q)))
  const key = { name: item => item.n, price: item => item.p, chg: item => (item.p / item.pc - 1) * 100, cap: item => item.p }
  list.sort((a, b) => {
    const ka = key[m.sort](a)
    const kb = key[m.sort](b)
    return (typeof ka === 'string' ? ka.localeCompare(kb, 'ru') : ka - kb) * m.dir
  })
  return list
}

function beep(kind) {
  try {
    const ctx = new AudioContext()
    const o = ctx.createOscillator()
    const g = ctx.createGain()
    o.connect(g); g.connect(ctx.destination)
    o.frequency.value = kind === 'error' ? 180 : kind === 'coin' ? 880 : kind === 'tick' ? 520 : 660
    g.gain.value = .04
    o.start()
    o.stop(ctx.currentTime + .08)
    setTimeout(() => ctx.close(), 200)
  } catch { /* звук необязателен */ }
}

export function StudioProvider({ children }) {
  const [model, setModel] = useState(() => ({
    account: loadAccount(),
    catalog: null,
    page: 'home',
    msel: [],
    hsel: [],
    hsort: 'value',
    hdir: -1,
    drawer: null,
    trade: { side: 'buy', qty: 1 },
    modal: null,
    dd: null,
    toasts: [],
    sideOpen: false,
    query: '',
    searchIdx: -1,
    ticker: '',
    tickerName: '',
    topAmt: '',
    viewName: '',
    closes: null,
  }))
  const undoRef = useRef(null)
  const toastFns = useRef({})
  const accountRef = useRef(model.account)
  accountRef.current = model.account

  const toast = (text, kind = 'success', action) => {
    const id = uid()
    if (action) toastFns.current[id] = action.fn
    setModel(m => ({ ...m, toasts: [...m.toasts, { id, text, kind, label: action?.label }] }))
    setTimeout(() => setModel(m => ({ ...m, toasts: m.toasts.filter(item => item.id !== id) })), 4200)
  }

  const snd = (kind) => { if (accountRef.current.settings.sound) beep(kind) }

  const commit = (recipe, { sound, note, undoLabel } = {}) => {
    setModel(m => {
      const account = clone(m.account)
      if (undoLabel) undoRef.current = clone(m.account)
      recipe(account)
      persist(account)
      return { ...m, account }
    })
    if (sound) snd(sound)
    if (note) toast(note.text, note.kind || 'success', undoLabel ? { label: 'Отменить', fn: undo } : note.action)
  }

  function undo() {
    const prev = undoRef.current
    if (!prev) return
    undoRef.current = null
    setModel(m => ({ ...m, account: prev }))
    persist(prev)
    toast('Действие отменено', 'info')
  }

  const api = useMemo(() => {
    const openAsset = (ticker, keep) => {
      const asset = findAsset({ account: accountRef.current, catalog: null }, ticker) || null
      setModel(m => {
        const known = findAsset(m, ticker)
        if (!known) return m
        return { ...m, drawer: { type: 'asset', t: ticker }, trade: keep ? m.trade : { side: 'buy', qty: 1 }, dd: null, query: m.dd === 'search' ? '' : m.query }
      })
      if (!asset && !keep) { /* catalog-only names still open */ }
    }
    return { openAsset }
  }, [])

  useEffect(() => {
    let stop = false
    loadNews().then(items => {
      if (stop || !items) return
      setModel(m => {
        const account = { ...m.account, news: items, newsLive: true }
        persist(account)
        return { ...m, account }
      })
    })
    loadCatalog('').then(list => { if (!stop && list) setModel(m => ({ ...m, catalog: list })) })
    return () => { stop = true }
  }, [])

  useEffect(() => {
    const id = setInterval(() => {
      if (document.hidden) return
      setModel(m => {
        if (!m.account.settings.autoTick || m.modal) return m
        const account = clone(m.account)
        account.tickSeed += 1
        account.instruments.forEach(item => {
          if (item.live) return
          const r = lcg(item.seed * 31 + account.tickSeed * 97)()
          item.p = Math.max(1, item.p * (1 + (r - .47) * .024))
          item.rng = [Math.min(item.rng[0], item.p * .995), Math.max(item.rng[1], item.p * 1.005)]
        })
        persist(account)
        return { ...m, account }
      })
      const current = accountRef.current
      if (current) {
        setModel(m => m)
      }
    }, 45000)
    return () => clearInterval(id)
  }, [])

  useEffect(() => {
    if (!model.catalog) return undefined
    let stop = false
    const refresh = () => {
      quoteBatch(model.catalog.map(item => item.t)).then(rows => {
        if (stop || !rows) return
        const by = Object.fromEntries(rows.map(row => [row.ticker, row]))
        setModel(m => ({
          ...m,
          catalog: (m.catalog || []).map(item => {
            const row = by[item.t]
            if (!row || row.last == null) return item
            const price = Number(row.last)
            const pct = row.changePercent != null ? Number(String(row.changePercent).replace('%', '').replace(',', '.')) : 0
            const prev = Number.isFinite(pct) && price ? price / (1 + pct / 100) : item.pc
            return { ...item, p: price, pc: prev }
          }),
        }))
      }).catch(() => {})
    }
    const id = setInterval(refresh, 45000)
    return () => { stop = true; clearInterval(id) }
  }, [model.catalog?.length])

  useEffect(() => {
    document.body.dataset.page = model.page
    document.body.dataset.palette = model.account.settings.palette
    document.body.dataset.motion = model.account.settings.motion
  }, [model.page, model.account.settings.palette, model.account.settings.motion])

  useEffect(() => {
    if (model.drawer?.type !== 'asset') return undefined
    const ticker = model.drawer.t
    const asset = findAsset(model, ticker)
    if (!asset?.live) { setModel(m => m.closes?.ticker === ticker ? { ...m, closes: null } : m); return undefined }
    let stop = false
    loadCloses(ticker, model.account.period).then(vals => {
      if (!stop) setModel(m => ({ ...m, closes: vals ? { ticker, period: model.account.period, vals } : null }))
    })
    return () => { stop = true }
  }, [model.drawer?.type, model.drawer?.t, model.account.period, model.catalog])

  const actions = {
    snd,
    toast,
    undo,
    nav(page) {
      snd('click')
      setModel(m => ({ ...m, page, msel: [], hsel: [], drawer: null, dd: null, sideOpen: false }))
      window.scrollTo({ top: 0, behavior: 'smooth' })
    },
    market() { this.openMarket() },
    openMarket() { snd('click'); setModel(m => ({ ...m, page: 'market', drawer: null, dd: null, sideOpen: false })); window.scrollTo({ top: 0, behavior: 'smooth' }) },
    setPageTab(tab) { snd('click'); setModel(m => ({ ...m, msel: [], hsel: [], account: { ...m.account, ptab: tab } })) },
    setPeriod(period) { snd('click'); setModel(m => { const account = { ...m.account, period }; persist(account); return { ...m, account } }) },
    toggleSide() { snd('click'); setModel(m => ({ ...m, sideOpen: !m.sideOpen })) },
    closeSide() { setModel(m => ({ ...m, sideOpen: false })) },
    setQuery(query) { setModel(m => ({ ...m, query, searchIdx: -1, dd: query.trim() ? 'search' : null })) },
    moveSearch(dir) { setModel(m => ({ ...m, searchIdx: m.searchIdx + dir })) },
    toggleDD(name) {
      snd('click')
      setModel(m => ({ ...m, dd: m.dd === name ? null : name }))
    },
    closeDD() { setModel(m => m.dd ? { ...m, dd: null } : m) },
    readAll() { setModel(m => { const account = { ...m.account, notif: m.account.notif.map(n => ({ ...n, read: true })) }; persist(account); return { ...m, account } }) },
    openNotif(id) {
      setModel(m => {
        const account = clone(m.account)
        const note = account.notif.find(item => item.id === id)
        if (note) note.read = true
        persist(account)
        const next = { ...m, account, dd: null }
        if (note?.act === 'asset') return { ...next, drawer: { type: 'asset', t: note.arg }, trade: { side: 'buy', qty: 1 } }
        if (note?.act === 'lesson') return { ...next, drawer: { type: 'lesson', id: Number(note.arg) } }
        if (note?.act === 'nav') return { ...next, page: note.arg }
        return next
      })
    },
    openAsset(ticker) { snd('click'); setModel(m => ({ ...m, drawer: { type: 'asset', t: ticker }, trade: { side: 'buy', qty: 1 }, dd: null, query: '' })) },
    openTicker(ticker, name = '') { snd('click'); setModel(m => ({ ...m, page: 'ticker', ticker, tickerName: name, drawer: null, dd: null, query: '' })); window.scrollTo({ top: 0, behavior: 'smooth' }) },
    closeTicker() { snd('click'); setModel(m => ({ ...m, page: 'market', drawer: null })); window.scrollTo({ top: 0, behavior: 'smooth' }) },
    openTrade(ticker, side) { snd('click'); setModel(m => ({ ...m, drawer: { type: 'asset', t: ticker }, trade: { side, qty: 1 } })) },
    openLesson(id) {
      const lesson = LESSONS.find(item => item.id === id)
      const level = Math.floor(accountRef.current.xp / 100) + 1
      if (lesson && lesson.lvl > level) { snd('error'); toast(`Урок ${lesson.id} откроется на уровне ${lesson.lvl}. Зарабатывайте XP сделками и уроками.`, 'warn'); return }
      snd('click')
      setModel(m => ({ ...m, drawer: { type: 'lesson', id }, dd: null, query: '' }))
    },
    openNews(id) { snd('click'); setModel(m => ({ ...m, drawer: { type: 'news', id: Number(id) } })) },
    openNewsList() { snd('click'); setModel(m => ({ ...m, drawer: { type: 'newslist' } })) },
    closeDrawer() { setModel(m => ({ ...m, drawer: null })) },
    setSide(side) { snd('click'); setModel(m => ({ ...m, trade: { ...m.trade, side } })) },
    setQty(qty) { setModel(m => ({ ...m, trade: { ...m.trade, qty } })) },
    bumpQty(delta) { setModel(m => ({ ...m, trade: { ...m.trade, qty: Math.max(1, (Math.floor(m.trade.qty) || 1) + delta) } })) },
    qtyPct(pct) {
      setModel(m => {
        const asset = findAsset(m, m.drawer?.t)
        if (!asset) return m
        const qty = m.trade.side === 'buy'
          ? Math.max(1, Math.floor(m.account.cash / (asset.p * (1 + FEE)) * pct / 100))
          : Math.max(1, Math.floor((m.account.holdings[asset.t]?.qty || 0) * pct / 100))
        return { ...m, trade: { ...m.trade, qty } }
      })
    },
    favToggle(ticker) {
      const had = accountRef.current.favs.includes(ticker)
      setModel(m => {
        const account = clone(m.account)
        if (had) {
          undoRef.current = clone(m.account)
          account.favs = account.favs.filter(item => item !== ticker)
        } else {
          const asset = findAsset(m, ticker)
          if (asset && !account.instruments.some(item => item.t === ticker)) account.instruments.push(clone(asset))
          account.favs.push(ticker)
          account.xp += 5
        }
        persist(account)
        return { ...m, account }
      })
      snd('click')
      toast(had ? `${ticker} убрана из избранного` : `${ticker} добавлена в избранное`, had ? 'info' : 'success', had ? { label: 'Отменить', fn: undo } : undefined)
    },
    favMove(ticker, dir) {
      setModel(m => {
        const account = clone(m.account)
        const i = account.favs.indexOf(ticker)
        const j = i + dir
        if (i < 0 || j < 0 || j >= account.favs.length) return m
        ;[account.favs[i], account.favs[j]] = [account.favs[j], account.favs[i]]
        persist(account)
        return { ...m, account }
      })
      snd('click')
    },
    reorderFav(from, to) {
      setModel(m => {
        const account = clone(m.account)
        const a = account.favs.indexOf(from)
        const b = account.favs.indexOf(to)
        if (a < 0 || b < 0 || a === b) return m
        const [item] = account.favs.splice(a, 1)
        account.favs.splice(b, 0, item)
        persist(account)
        return { ...m, account }
      })
      snd('click')
      toast('Порядок избранного обновлён', 'info')
    },
    tick() {
      snd('tick')
      setModel(m => {
        const account = clone(m.account)
        account.tickSeed += 1
        account.instruments.forEach(item => {
          if (item.live) return
          const r = lcg(item.seed * 31 + account.tickSeed * 97)()
          item.p = Math.max(1, item.p * (1 + (r - .47) * .024))
          item.rng = [Math.min(item.rng[0], item.p * .995), Math.max(item.rng[1], item.p * 1.005)]
        })
        persist(account)
        return { ...m, account }
      })
      if (model.catalog) {
        quoteBatch(model.catalog.map(item => item.t)).then(rows => {
          const by = Object.fromEntries((rows || []).map(row => [row.ticker, row]))
          setModel(m => ({
            ...m,
            catalog: (m.catalog || []).map(item => {
              const row = by[item.t]
              if (!row || row.last == null) return item
              const price = Number(row.last)
              const pct = row.changePercent != null ? Number(String(row.changePercent).replace('%', '').replace(',', '.')) : 0
              return { ...item, p: price, pc: price && Number.isFinite(pct) ? price / (1 + pct / 100) : item.pc }
            }),
          }))
        }).catch(() => {})
      }
      toast('Котировки обновлены · ' + new Date().toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' }), 'info')
    },
    openTopup() { snd('click'); setModel(m => ({ ...m, modal: 'topup', topAmt: '' })) },
    setTopAmt(topAmt) { setModel(m => ({ ...m, topAmt })) },
    addTop(amount) { snd('click'); setModel(m => ({ ...m, topAmt: String((parseFloat(String(m.topAmt).replace(',', '.')) || 0) + amount) })) },
    execTopup() {
      const raw = String(model.topAmt).trim().replace(/\s/g, '').replace(',', '.')
      const value = parseFloat(raw)
      if (!raw || Number.isNaN(value) || value <= 0) { snd('error'); toast('Введите сумму больше нуля — например, 5000.', 'error'); return }
      if (value > 1000000) { snd('error'); toast('Лимит демо-пополнения — $1 000 000 за операцию.', 'error'); return }
      setModel(m => {
        const account = clone(m.account)
        account.cash += value
        account.tx.push({ id: uid(), type: 'top', amt: value, d: new Date().toISOString() })
        account.notif.unshift({ id: uid(), icon: 'coins', c: 'info', t: `Пополнение +${value.toFixed(2)}`, s: 'только что · свободные средства', read: false, act: 'nav', arg: 'portfolio' })
        persist(account)
        return { ...m, account, modal: null }
      })
      snd('coin')
      toast(`Счёт пополнен на $${value.toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`)
    },
    closeModal() { setModel(m => ({ ...m, modal: null })) },
    ask(modal) { setModel(m => ({ ...m, modal })) },
    confirmYes() {
      const modal = model.modal
      setModel(m => ({ ...m, modal: null }))
      modal?.run?.()
    },
    trade() {
      const ticker = model.drawer?.t
      const asset = findAsset(model, ticker)
      if (!asset) return
      const qty = Math.floor(model.trade.qty) || 0
      const gross = qty * asset.p
      const fee = gross * FEE
      const total = model.trade.side === 'buy' ? gross + fee : gross - fee
      const pos = model.account.holdings[ticker]
      if (qty <= 0) { snd('error'); return }
      if (model.trade.side === 'buy' && total > model.account.cash) { snd('error'); toast('Недостаточно свободных средств', 'error'); return }
      if (model.trade.side === 'sell' && (!pos || qty > pos.qty)) { snd('error'); toast('Недостаточно акций в позиции', 'error'); return }
      setModel(m => {
        const account = clone(m.account)
        if (m.trade.side === 'sell') undoRef.current = clone(m.account)
        if (!account.instruments.some(item => item.t === ticker)) account.instruments.push(clone(asset))
        if (m.trade.side === 'buy') {
          account.cash -= total
          const held = account.holdings[ticker] || { qty: 0, avg: asset.p }
          held.avg = (held.avg * held.qty + asset.p * qty) / (held.qty + qty)
          held.qty += qty
          account.holdings[ticker] = held
        } else {
          account.cash += total
          account.holdings[ticker].qty -= qty
          if (account.holdings[ticker].qty <= 0) delete account.holdings[ticker]
        }
        account.tx.push({ id: uid(), type: m.trade.side, t: ticker, qty, price: asset.p, amt: total, d: new Date().toISOString() })
        account.xp += 20
        account.notif.unshift({ id: uid(), icon: m.trade.side === 'buy' ? 'up' : 'down', c: m.trade.side === 'buy' ? 'up' : 'down', t: `${m.trade.side === 'buy' ? 'Покупка' : 'Продажа'} ${ticker} × ${qty}`, s: 'только что', read: false, act: 'asset', arg: ticker })
        persist(account)
        return { ...m, account }
      })
      snd('success')
      toast(model.trade.side === 'buy' ? `Куплено: ${ticker} × ${qty}` : `Продано: ${ticker} × ${qty}`, 'success', model.trade.side === 'sell' ? { label: 'Отменить', fn: undo } : undefined)
    },
    finishLesson(id) {
      if (accountRef.current.lessons.includes(id)) { toast('Урок уже засчитан', 'info'); return }
      const lesson = LESSONS.find(item => item.id === id)
      setModel(m => {
        const account = clone(m.account)
        account.lessons.push(id)
        account.xp += lesson?.xp || 0
        account.notif.unshift({ id: uid(), icon: 'book', c: 'info', t: `Урок ${id} пройден · +${lesson?.xp || 0} XP`, s: 'только что · прогресс курса', read: false, act: 'lesson', arg: id })
        persist(account)
        return { ...m, account }
      })
      snd('success')
    },
    setMarket(patch) { setModel(m => { const account = { ...m.account, mkt: { ...m.account.mkt, ...patch } }; persist(account); return { ...m, account } }) },
    toggleDir() { snd('click'); setModel(m => { const account = { ...m.account, mkt: { ...m.account.mkt, dir: m.account.mkt.dir * -1 } }; persist(account); return { ...m, account } }) },
    resetMarket() { snd('click'); setModel(m => { const account = { ...m.account, mkt: { ...m.account.mkt, q: '', sec: 'all' } }; persist(account); return { ...m, account } }) },
    toggleMsel(ticker, on) { setModel(m => ({ ...m, msel: on ? [...new Set([...m.msel, ticker])] : m.msel.filter(item => item !== ticker) })) },
    toggleMall(on, rows) { setModel(m => ({ ...m, msel: on ? rows.map(item => item.t) : [] })) },
    clearMsel() { setModel(m => ({ ...m, msel: [] })) },
    toggleHsel(ticker, on) { setModel(m => ({ ...m, hsel: on ? [...new Set([...m.hsel, ticker])] : m.hsel.filter(item => item !== ticker) })) },
    toggleHall(on) { setModel(m => ({ ...m, hsel: on ? Object.keys(m.account.holdings) : [] })) },
    clearHsel() { setModel(m => ({ ...m, hsel: [] })) },
    hsort(key) { snd('click'); setModel(m => ({ ...m, hdir: m.hsort === key ? m.hdir * -1 : -1, hsort: key })) },
    bulkFav(mode) {
      const picked = model.msel
      setModel(m => {
        const account = clone(m.account)
        if (mode === 'del') undoRef.current = clone(m.account)
        if (mode === 'add') {
          picked.forEach(ticker => {
            if (!account.favs.includes(ticker)) account.favs.push(ticker)
            const asset = findAsset(m, ticker)
            if (asset && !account.instruments.some(item => item.t === ticker)) account.instruments.push(clone(asset))
          })
          account.xp += 5
        } else account.favs = account.favs.filter(ticker => !picked.includes(ticker))
        persist(account)
        return { ...m, account, msel: [] }
      })
      snd(mode === 'add' ? 'success' : 'click')
      toast(mode === 'add' ? `В избранное добавлено: ${picked.length}` : `Убрано из избранного: ${picked.length}`, mode === 'add' ? 'success' : 'info', mode === 'del' ? { label: 'Отменить', fn: undo } : undefined)
    },
    bulkSell() {
      const picked = model.hsel.filter(ticker => model.account.holdings[ticker])
      if (!picked.length) return
      const proceeds = picked.reduce((sum, ticker) => sum + model.account.holdings[ticker].qty * findAsset(model, ticker).p * (1 - FEE), 0)
      setModel(m => ({
        ...m,
        modal: {
          kind: 'confirm',
          title: 'Продать выбранное?',
          body: `Позиции: ${picked.join(', ')}. Будет зачислено ≈ ${proceeds.toFixed(2)} с учётом комиссии 0,1%.`,
          label: 'Продать всё',
          danger: true,
          run: () => {
            undoRef.current = clone(accountRef.current)
            setModel(cur => {
              const account = clone(cur.account)
              picked.forEach(ticker => {
                const held = account.holdings[ticker]
                const price = findAsset({ ...cur, account }, ticker)?.p || held.avg
                const amt = held.qty * price * (1 - FEE)
                account.cash += amt
                account.tx.push({ id: uid(), type: 'sell', t: ticker, qty: held.qty, price, amt, d: new Date().toISOString() })
                delete account.holdings[ticker]
              })
              account.xp += 20
              persist(account)
              return { ...cur, account, hsel: [] }
            })
            snd('success')
            toast(`Продано ${picked.length} позиций`, 'success', { label: 'Отменить', fn: undo })
          },
        },
      }))
    },
    setHist(f) { snd('click'); setModel(m => { const account = { ...m.account, hist: { f } }; persist(account); return { ...m, account } }) },
    openSaveView() { snd('click'); setModel(m => ({ ...m, modal: 'view', viewName: '' })) },
    setViewName(viewName) { setModel(m => ({ ...m, viewName })) },
    saveView() {
      const name = model.viewName.trim()
      if (!name) { snd('error'); toast('Введите название вида.', 'error'); return }
      setModel(m => {
        const account = clone(m.account)
        account.views.push({ id: uid(), name, q: account.mkt.q, sec: account.mkt.sec, sort: account.mkt.sort, dir: account.mkt.dir })
        persist(account)
        return { ...m, account, modal: null }
      })
      snd('success')
      toast(`Вид «${name}» сохранён`)
    },
    applyView(id) {
      const view = model.account.views.find(item => item.id === id)
      if (!view) return
      snd('click')
      setModel(m => { const account = { ...m.account, mkt: { q: view.q, sec: view.sec, sort: view.sort, dir: view.dir } }; persist(account); return { ...m, account } })
      toast(`Вид «${view.name}» применён`, 'info')
    },
    deleteView(id) {
      const view = model.account.views.find(item => item.id === id)
      undoRef.current = clone(model.account)
      setModel(m => { const account = clone(m.account); account.views = account.views.filter(item => item.id !== id); persist(account); return { ...m, account } })
      snd('click')
      toast(`Вид «${view?.name || ''}» удалён`, 'info', { label: 'Отменить', fn: undo })
    },
    setting(key) {
      snd('click')
      setModel(m => {
        const account = clone(m.account)
        if (key === 'compact') account.settings.compact = !account.settings.compact
        if (key === 'sound') account.settings.sound = !account.settings.sound
        if (key === 'autoTick') account.settings.autoTick = !account.settings.autoTick
        if (key === 'motion') account.settings.motion = account.settings.motion === 'on' ? 'off' : 'on'
        persist(account)
        return { ...m, account }
      })
    },
    palette(palette) {
      snd('click')
      setModel(m => { const account = clone(m.account); account.settings.palette = palette; persist(account); return { ...m, account } })
      toast(palette === 'deutan' ? 'Палитра применена: синий / оранжевый' : 'Палитра применена: классическая', 'info')
    },
    soundToggle() {
      setModel(m => { const account = clone(m.account); account.settings.sound = !account.settings.sound; persist(account); return { ...m, account } })
      toast(!accountRef.current.settings.sound ? 'Звук включён' : 'Звук выключен', 'info')
      if (!accountRef.current.settings.sound) snd('success')
    },
    exportJSON() {
      const blob = new Blob([JSON.stringify(model.account, null, 2)], { type: 'application/json' })
      const a = document.createElement('a')
      a.href = URL.createObjectURL(blob)
      a.download = 'myinvestor-state-2026-11-22.json'
      a.click()
      toast('JSON состояния сохранён')
    },
    exportCSV() {
      const rows = [['date', 'type', 'ticker', 'qty', 'price', 'amount']]
      model.account.tx.forEach(row => rows.push([row.d, row.type, row.t || '', row.qty || '', row.price || '', Number(row.amt).toFixed(2)]))
      const blob = new Blob(['\uFEFF' + rows.map(row => row.join(';')).join('\n')], { type: 'text/csv;charset=utf-8' })
      const a = document.createElement('a')
      a.href = URL.createObjectURL(blob)
      a.download = 'myinvestor-history.csv'
      a.click()
      toast('История операций сохранена в CSV')
    },
    reset() {
      setModel(m => ({
        ...m,
        modal: {
          kind: 'confirm',
          title: 'Сбросить демо-данные?',
          body: 'Счёт, позиции, избранное, история и прогресс вернутся к стартовому состоянию. Сразу после сброса появится кнопка «Отменить».',
          label: 'Сбросить',
          danger: true,
          run: () => {
            undoRef.current = clone(accountRef.current)
            const settings = { ...accountRef.current.settings }
            const account = seedAccount()
            account.settings = settings
            persist(account)
            setModel(cur => ({ ...cur, account, msel: [], hsel: [], modal: null }))
            toast('Демо-данные сброшены к старту', 'info', { label: 'Отменить', fn: undo })
          },
        },
      }))
    },
    runToast(id) { toastFns.current[id]?.() },
  }

  const value = { model, actions, sectors: SECTORS, lessons: LESSONS }
  return <StudioContext.Provider value={value}>{children}</StudioContext.Provider>
}
