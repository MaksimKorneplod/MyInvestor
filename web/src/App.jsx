import { useEffect, useRef, useState } from 'react'
import { Logo, Icon, Mark } from './icons'
import { instruments } from './api'
import { LESSONS, TRADERS, chgOf, dayPL, fmt2, fmtMoney, fmtPct, levelOf, myRet, totalVal } from './mock'
import { Drawers, Home, Learn, Portfolio, Rating, Settings } from './screens'
import { MarketScreen, TickerPage } from './board'
import { StudioProvider, useStudio } from './studio'

const NAV = [['home', 'home', 'Главная'], ['portfolio', 'case', 'Портфель'], ['rating', 'trophy', 'Рейтинг'], ['learn', 'book', 'Обучение'], ['settings', 'gear', 'Настройки']]

function Defs() {
  return (
    <svg width="0" height="0" style={{ position: 'absolute' }} aria-hidden="true"><defs>
      <linearGradient id="flameGrad" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stopColor="#FF7A5C" /><stop offset=".5" stopColor="#FF4D5E" /><stop offset="1" stopColor="#C40E2E" /></linearGradient>
      <linearGradient id="sparkUp" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stopColor="#2ECC71" stopOpacity=".36" /><stop offset="1" stopColor="#2ECC71" stopOpacity="0" /></linearGradient>
      <linearGradient id="sparkDown" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stopColor="#FF6B6B" stopOpacity=".32" /><stop offset="1" stopColor="#FF6B6B" stopOpacity="0" /></linearGradient>
      <linearGradient id="sparkUpD" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stopColor="#3FA7FF" stopOpacity=".36" /><stop offset="1" stopColor="#3FA7FF" stopOpacity="0" /></linearGradient>
      <linearGradient id="sparkDownD" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stopColor="#FF9F1C" stopOpacity=".32" /><stop offset="1" stopColor="#FF9F1C" stopOpacity="0" /></linearGradient>
    </defs></svg>
  )
}

function Shell() {
  const { model, actions } = useStudio()
  const account = model.account
  const searchRef = useRef(null)
  const bellRef = useRef(null)
  const avatarRef = useRef(null)
  const level = levelOf(account.xp)
  const unread = account.notif.filter(item => !item.read).length
  const day = dayPL(account)
  const base = totalVal(account) - day
  const pct = base ? day / base * 100 : 0
  const q = model.query.trim()
  const [moexHits, setMoexHits] = useState([])
  const [searchState, setSearchState] = useState('ready')
  const lessons = q ? LESSONS.filter(lesson => lesson.t.toLowerCase().includes(q.toLowerCase())).slice(0, 3) : []
  const mockHits = q && searchState === 'error'
    ? account.instruments.filter(item => item.n.toLowerCase().includes(q.toLowerCase()) || item.t.toLowerCase().includes(q.toLowerCase())).slice(0, 6)
    : []
  const results = [
    ...moexHits.map(item => ({ kind: 'moex', item })),
    ...mockHits.map(asset => ({ kind: 'asset', asset })),
    ...lessons.map(lesson => ({ kind: 'lesson', lesson })),
  ]
  const rank = [...TRADERS.map(trader => trader.ret), myRet(account)].sort((a, b) => b - a).indexOf(myRet(account)) + 1

  useEffect(() => {
    if ([...q].length > 100 || !q) { setMoexHits([]); setSearchState('ready'); return undefined }
    let live = true
    setSearchState('loading')
    const id = setTimeout(() => {
      instruments({ q, limit: 8, offset: 0 })
        .then(data => { if (live) { setMoexHits(data?.items || []); setSearchState('ready') } })
        .catch(() => { if (live) { setMoexHits([]); setSearchState('error') } })
    }, 250)
    return () => { live = false; clearTimeout(id) }
  }, [q])

  useEffect(() => {
    const onKey = (event) => {
      if (event.key === 'Escape') {
        if (model.modal) actions.closeModal()
        else if (model.drawer) actions.closeDrawer()
        else if (model.dd) actions.closeDD()
        else actions.closeSide()
        return
      }
      if (event.key === '/' && !event.target.closest('input,select,textarea')) {
        event.preventDefault()
        searchRef.current?.focus()
      }
      if (event.target === searchRef.current && (event.key === 'ArrowDown' || event.key === 'ArrowUp')) {
        event.preventDefault()
        actions.moveSearch(event.key === 'ArrowDown' ? 1 : -1)
      }
      if (event.target === searchRef.current && event.key === 'Enter' && results.length) {
        const pick = results[Math.max(0, Math.min(results.length - 1, model.searchIdx))] || results[0]
        if (pick.kind === 'moex') actions.openTicker(pick.item.ticker, pick.item.name)
        else if (pick.kind === 'asset') actions.openAsset(pick.asset.t)
        else actions.openLesson(pick.lesson.id)
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [model.modal, model.drawer, model.dd, model.searchIdx, model.query, actions, results])

  const place = (node) => {
    if (!node) return {}
    const rect = node.getBoundingClientRect()
    return { top: Math.min(rect.bottom + 10, window.innerHeight - 140), left: Math.max(10, rect.right - 340) }
  }

  return (
    <>
      <Defs />
      <div className="aurora" aria-hidden="true"><i className="a1" /><i className="a2" /><i className="a3" /></div>
      <div className={`side-scrim${model.sideOpen ? ' open' : ''}`} onClick={() => actions.closeSide()} />
      <aside className={`sidebar${model.sideOpen ? ' open' : ''}`}>
        <div className="logo"><Logo /><div><b>MyInvestor</b><small>СИМУЛЯТОР · v2.0</small></div></div>
        <nav className="snav" aria-label="Основная навигация">
          {NAV.map(([page, icon, name]) => <button key={page} className={model.page === page || (page === 'portfolio' && model.page === 'ticker') ? 'on' : ''} onClick={() => actions.nav(page)}><Icon name={icon} /><span>{name}</span>{page === 'learn' ? <span className="nbadge">{account.lessons.length}/8</span> : null}</button>)}
        </nav>
        <div className="sxp"><div className="t"><span>Уровень <b>{level}</b></span><span>{account.xp % 100} / 100 XP</span></div><div className="xpbar"><i style={{ width: `${account.xp % 100}%` }} /></div></div>
        <div className="spromo">
          <svg width="120" height="40" viewBox="0 0 120 40">{[8, 14, 11, 20, 17, 26, 23, 32, 29, 38].map((h, i) => <rect key={h + i} x={i * 12} y={40 - h} width="7" height={h} rx="2" fill="#FF4D5E" opacity={.35 + i * .065} />)}</svg>
          <p>Учись. Инвестируй. Расти.</p>
          <button onClick={() => actions.nav('learn')}>К обучению <Icon name="arrR" /></button>
        </div>
      </aside>
      <div className="main">
        <header className="topbar">
          <button className="burger" aria-label="Открыть меню" onClick={() => actions.toggleSide()}><Icon name="menu" /></button>
          <div className="gsearch"><div className="box">
            <Icon name="search" />
            <input ref={searchRef} value={model.query} placeholder="Поиск по всем акциям: SBER, Газпром…" autoComplete="off" aria-label="Глобальный поиск" onChange={e => actions.setQuery(e.target.value)} />
            <kbd>/</kbd>
          </div></div>
          <div className="totalchip" title="Стоимость счёта"><span>{fmtMoney(totalVal(account), account.settings.compact)}</span><span className={`d ${pct >= 0 ? 'up' : 'down'}`}>{fmtPct(pct)}</span></div>
          <button className="tbtn" aria-label="Обновить котировки" title="Обновить котировки" onClick={() => actions.tick()}><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 12a9 9 0 1 1-2.6-6.3" /><path d="M21 3v6h-6" /></svg></button>
          <button className="tbtn" aria-pressed={account.settings.sound} aria-label={account.settings.sound ? 'Звук включён' : 'Звук выключен'} title="Звук" onClick={() => actions.soundToggle()}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">{account.settings.sound ? <><path d="M11 5L6 9H2v6h4l5 4V5z" /><path d="M15.5 8.5a5 5 0 0 1 0 7M18.5 5.5a9 9 0 0 1 0 13" /></> : <><path d="M11 5L6 9H2v6h4l5 4V5z" /><path d="M22 9l-6 6M16 9l6 6" /></>}</svg>
          </button>
          <button ref={bellRef} className="tbtn" aria-label="Уведомления" onClick={() => actions.toggleDD('notif')}><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M6 9a6 6 0 0 1 12 0c0 5 2 6 2 6H4s2-1 2-6" /><path d="M10 19a2 2 0 0 0 4 0" /></svg>{unread > 0 && <span className="badge">{unread}</span>}</button>
          <button ref={avatarRef} className="avatarbtn" aria-label="Профиль" onClick={() => actions.toggleDD('avatar')}><i>АК</i><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round"><path d="M6 9l6 6 6-6" /></svg></button>
        </header>
        <main className="page">
          {model.page === 'home' && <Home />}
          {model.page === 'market' && <MarketScreen />}
          {model.page === 'ticker' && <TickerPage />}
          {model.page === 'portfolio' && <Portfolio />}
          {model.page === 'rating' && <Rating />}
          {model.page === 'learn' && <Learn />}
          {model.page === 'settings' && <Settings />}
        </main>
      </div>
      <Drawers />
      {model.dd === 'search' && (
        <div className="dd open" role="listbox" aria-label="Результаты поиска" style={{ top: (searchRef.current?.getBoundingClientRect().bottom || 70) + 10, left: searchRef.current?.getBoundingClientRect().left || 260, width: 340 }}>
          {results.length ? results.map((item, index) => item.kind === 'moex' ? (
            <div key={item.item.ticker} className={`sr-item${index === model.searchIdx ? ' hot' : ''}`} role="option" onClick={() => actions.openTicker(item.item.ticker, item.item.name)}><span className="flogo" style={{ background: '#3B2A36' }}>{item.item.ticker.slice(0, 1)}</span><div><div className="n">{item.item.name}</div><div className="t">{item.item.ticker} · MOEX</div></div></div>
          ) : item.kind === 'asset' ? (
            <div key={item.asset.t} className={`sr-item${index === model.searchIdx ? ' hot' : ''}`} role="option" onClick={() => actions.openAsset(item.asset.t)}><span><Mark asset={item.asset} /></span><div><div className="n">{item.asset.n}</div><div className="t">{item.asset.t}</div></div><div className="p"><b>{fmt2(item.asset.p)}</b><small className={chgOf(item.asset) >= 0 ? 'up' : 'down'}>{fmtPct(chgOf(item.asset))}</small></div></div>
          ) : (
            <div key={item.lesson.id} className={`sr-item${index === model.searchIdx ? ' hot' : ''}`} role="option" onClick={() => actions.openLesson(item.lesson.id)}><span className="flogo" style={{ background: 'rgba(232,17,45,.18)', color: 'var(--brand2)' }}><Icon name="book" style={{ width: 16, height: 16 }} /></span><div><div className="n">{item.lesson.t}</div><div className="t">Урок {item.lesson.id} · {item.lesson.min} мин{item.lesson.lvl > level ? ` · ур. ${item.lesson.lvl}` : ''}</div></div></div>
          )) : <div className="sr-none">Ничего не найдено по «{model.query}». Попробуйте тикер: SBER, GAZP…</div>}
        </div>
      )}
      {model.dd === 'notif' && (
        <div className="dd open" role="dialog" aria-label="Уведомления" style={place(bellRef.current)}>
          <div className="dd-head">Уведомления <button onClick={() => actions.readAll()}>Прочитать все</button></div>
          {account.notif.length ? account.notif.map(item => <div key={item.id} className={`notif-item${item.read ? ' read' : ''}`} onClick={() => actions.openNotif(item.id)}><span className="ni" style={{ background: item.c === 'up' ? 'rgba(46,204,113,.14)' : item.c === 'down' ? 'rgba(255,107,107,.12)' : item.c === 'warn' ? 'rgba(245,166,35,.14)' : 'rgba(79,141,247,.14)', color: item.c === 'up' ? 'var(--up)' : item.c === 'down' ? 'var(--down)' : item.c === 'warn' ? 'var(--warn)' : 'var(--info)' }}><Icon name={item.icon} /></span><div><b>{item.t}</b><span>{item.s}</span></div></div>) : <div className="sr-none">Уведомлений нет</div>}
        </div>
      )}
      {model.dd === 'avatar' && (
        <div className="dd open" role="dialog" aria-label="Профиль" style={place(avatarRef.current)}>
          <div className="avmenu"><div className="avt"><i>АК</i><div><b>АК · демо-трейдер</b><span>Уровень {level} · {account.xp} XP · счёт {fmtMoney(totalVal(account), account.settings.compact)}</span></div></div>
            <div className="xpbar"><i style={{ width: `${account.xp % 100}%` }} /></div>
            <button className="btn btn-ghost btn-sm" onClick={() => actions.nav('rating')}><Icon name="trophy" /> Мой рейтинг · #{rank}</button>
            <button className="btn btn-ghost btn-sm" onClick={() => actions.nav('settings')}><Icon name="gear" /> Настройки</button>
            <button className="btn btn-ghost btn-sm" onClick={() => actions.exportJSON()}><Icon name="download" /> Экспорт данных</button>
          </div>
        </div>
      )}
      {model.modal && <Modal />}
      <div className="toasts" aria-live="polite">{model.toasts.map(item => <div key={item.id} className={`toast ${item.kind}`}><Icon name={item.kind === 'error' ? 'alert' : item.kind === 'info' ? 'info' : 'check'} /><span>{item.text}</span>{item.label && <button className="tact" onClick={() => actions.runToast(item.id)}>{item.label}</button>}</div>)}</div>
    </>
  )
}

function Modal() {
  const { model, actions, sectors } = useStudio()
  const modal = model.modal
  const account = model.account
  return (
    <div className="modal-scrim open" onClick={e => { if (e.target === e.currentTarget) actions.closeModal() }}>
      <div className="modal" role="dialog" aria-modal="true">
        {modal === 'topup' && <>
          <h3>Пополнение счёта</h3><p className="msub">Демо-операция симулятора · без реальных денег · сейчас свободно {fmt2(account.cash)}</p>
          <label htmlFor="topAmt">Сумма пополнения</label>
          <div className="minput"><span>$</span><input id="topAmt" inputMode="decimal" autoComplete="off" placeholder="0.00" value={model.topAmt} onChange={e => actions.setTopAmt(e.target.value)} onKeyDown={e => { if (e.key === 'Enter') actions.execTopup() }} /></div>
          <div className="mquick"><button onClick={() => actions.addTop(1000)}>+1 000</button><button onClick={() => actions.addTop(5000)}>+5 000</button><button onClick={() => actions.addTop(10000)}>+10 000</button></div>
          <div className="mactions"><button className="btn btn-ghost" onClick={() => actions.closeModal()}>Отмена</button><button className="btn btn-brand" onClick={() => actions.execTopup()}><Icon name="coins" /> Пополнить</button></div>
        </>}
        {modal === 'view' && <>
          <h3>Сохранить вид рынка</h3>
          <p className="msub">Текущие фильтры: {account.mkt.sec === 'all' ? 'все сектора' : sectors[account.mkt.sec]?.n} · сортировка {({ chg: 'по изменению', price: 'по цене', name: 'по имени' })[account.mkt.sort]} {account.mkt.dir > 0 ? '↑' : '↓'}{account.mkt.q ? ` · поиск «${account.mkt.q}»` : ''}</p>
          <label htmlFor="viewName">Название вида</label>
          <div className="minput"><input id="viewName" className="txt" maxLength={32} placeholder="Например: Тех · рост" autoComplete="off" value={model.viewName} onChange={e => actions.setViewName(e.target.value)} onKeyDown={e => { if (e.key === 'Enter') actions.saveView() }} /></div>
          <div className="mactions"><button className="btn btn-ghost" onClick={() => actions.closeModal()}>Отмена</button><button className="btn btn-brand" onClick={() => actions.saveView()}><Icon name="save" /> Сохранить</button></div>
        </>}
        {modal?.kind === 'confirm' && <>
          <h3>{modal.title}</h3><p className="msub">{modal.body}</p>
          <div className="mactions"><button className="btn btn-ghost" onClick={() => actions.closeModal()}>Отмена</button><button className={`btn ${modal.danger ? 'btn-down' : 'btn-brand'}`} onClick={() => actions.confirmYes()}>{modal.label}</button></div>
        </>}
      </div>
    </div>
  )
}

export default function App() {
  return <StudioProvider><Shell /></StudioProvider>
}
