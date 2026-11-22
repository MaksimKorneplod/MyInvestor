import { useEffect, useState } from 'react'
import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useApp, useLeave } from '../App'
import { initials } from '../format'
import { Icon, Logo } from '../icons'

const LINKS = [
  ['/market', 'home', 'Рынок'],
  ['/news', 'news', 'Новости'],
  ['/settings', 'gear', 'Настройки'],
]

export default function Shell() {
  const { session } = useApp()
  const leave = useLeave()
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)
  const [menu, setMenu] = useState(false)
  const [q, setQ] = useState('')
  const who = session.login || session.email

  useEffect(() => {
    document.body.dataset.page = 'app'
    return () => { document.body.dataset.page = '' }
  }, [])

  function search(e) {
    e.preventDefault()
    const value = q.trim()
    navigate(value ? `/market?q=${encodeURIComponent(value)}` : '/market')
    setOpen(false)
  }

  return (
    <>
      <div className="aurora" aria-hidden="true"><i className="a1" /><i className="a2" /><i className="a3" /></div>
      <button className={open ? 'side-scrim open' : 'side-scrim'} type="button" aria-label="Закрыть меню" onClick={() => setOpen(false)} />
      <aside className={open ? 'sidebar open' : 'sidebar'}>
        <div className="logo"><Logo /><div><b>MyInvestor</b><small>{session.role === 'school' ? 'ШКОЛА' : 'УЧЕНИК'}</small></div></div>
        <nav className="snav" aria-label="Основная навигация">
          {LINKS.map(([to, icon, label]) => (
            <NavLink key={to} to={to} className={({ isActive }) => isActive ? 'on' : ''} onClick={() => setOpen(false)}>
              <Icon name={icon} /><span>{label}</span>
            </NavLink>
          ))}
        </nav>
        <div className="spromo">
          <p>Акции MOEX. Цены последней сделки.</p>
          <button type="button" onClick={() => { navigate('/market'); setOpen(false) }}>К рынку <Icon name="arrR" /></button>
        </div>
      </aside>
      <div className="main">
        <header className="topbar">
          <button className="burger" type="button" aria-label="Открыть меню" onClick={() => setOpen(true)}><Icon name="menu" /></button>
          <form className="gsearch" onSubmit={search}>
            <div className="box">
              <Icon name="search" />
              <input value={q} placeholder="Тикер или название: SBER, Газпром…" aria-label="Поиск по тикеру или названию" onChange={e => setQ(e.target.value)} />
            </div>
          </form>
          <div className="avatarwrap">
            <button className="avatarbtn" type="button" aria-expanded={menu} aria-label="Профиль" onClick={() => setMenu(v => !v)}>
              <i>{initials(who)}</i>
            </button>
            {menu && (
              <div className="avmenu">
                <b>{session.role === 'school' ? 'Обучающая площадка' : 'Ученик'}</b>
                <span>{session.email}</span>
                <button type="button" onClick={leave}>Выйти</button>
              </div>
            )}
          </div>
        </header>
        <main className="page"><Outlet /></main>
      </div>
    </>
  )
}
