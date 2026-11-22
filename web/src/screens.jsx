import { useState } from 'react'
import { assetSeries, Donut, LineChart, portfolioSeries, Spark } from './charts'
import { Bull, Icon, Mark, NewsArt } from './icons'
import { FEE, TRADERS, allocOf, chgOf, dayPL, fmt0, fmt2, fmtMoney, fmtPct, investedVal, levelOf, myRet, totalVal, unrealPL } from './mock'
import { findAsset, useStudio } from './studio'

const tiny = { width: 12, height: 12 }

function moneySigned(value, compact) {
  return `${value >= 0 ? '+' : '−'}${fmt2(Math.abs(value)).slice(1)}`
}

export function Home() {
  const { model, actions } = useStudio()
  const account = model.account
  const d = dayPL(account)
  const base = totalVal(account) - d
  const pct = base ? d / base * 100 : 0
  const up = d >= 0
  const chart = portfolioSeries(account)
  const news = account.news
  const done = account.lessons.length
  const [drag, setDrag] = useState(null)
  const [over, setOver] = useState(null)
  return (
    <>
      <div className="pagehead enter"><h1>Главная</h1><div className="right">
        <button className="btn btn-ghost btn-sm" onClick={() => actions.tick()}><Icon name="trend" /> Обновить котировки</button>
        <button className="btn btn-brand btn-sm" onClick={() => actions.openTopup()}><Icon name="plus" /> Пополнить</button>
      </div>
      <span className="sub">Суббота, 22 ноября 2026 · демо-данные симулятора · котировки обновляются по кнопке и автоматически</span></div>
      <div className="balance enter" style={{ '--d': '.05s' }}>
        <svg className="arrowbg" viewBox="0 0 240 240"><path d="M30 210 L150 90 L120 88 L190 30 L182 108 L160 100 L60 220 Z" fill="#E8112D" /><path d="M10 190 C 80 150 120 190 230 60" stroke="#E8112D" strokeWidth="10" fill="none" opacity=".5" strokeLinecap="round" /></svg>
        <div className="bleft"><div className="lbl"><span className="live" />Ваш инвестиционный счёт</div>
          <div className="bnum js-total">{fmtMoney(totalVal(account), account.settings.compact)}</div>
          <div className={`bchange ${up ? '' : 'neg'}`}><Icon name={up ? 'up' : 'down'} /><span>{moneySigned(d)}</span><span style={{ opacity: .6 }}>•</span><span>{fmtPct(pct)} за день</span></div>
        </div>
        <div className="bright">
          <div className="brow"><Icon name="coins" /><div><div className="t">Свободные средства</div><div className="v">{fmtMoney(account.cash, account.settings.compact)}</div></div></div>
          <div className="brow"><Icon name="trend" /><div><div className="t">В инвестициях</div><div className="v">{fmtMoney(investedVal(account), account.settings.compact)}</div></div></div>
        </div>
        <div className="bcta"><button className="btn btn-brand" onClick={() => actions.openTopup()}><Icon name="plus" /> Пополнить</button></div>
      </div>
      <div className="sechead enter" style={{ '--d': '.1s' }}><h2>Избранные инструменты</h2><span className="hint">перетащите карточки для порядка</span>
        <button className="link" onClick={() => actions.market()}>Все активы <Icon name="arrR" /></button></div>
      <div className="favgrid" id="favGrid">
        {account.favs.length ? account.favs.map((ticker, i) => {
          const asset = findAsset(model, ticker)
          if (!asset) return null
          const change = chgOf(asset)
          return (
            <div key={ticker} className={`favcard enter${drag === ticker ? ' dragging' : ''}${over === ticker ? ' dragover' : ''}`} style={{ '--d': `${i * .06}s` }} draggable role="button" tabIndex={0}
              onClick={() => actions.openAsset(ticker)}
              onKeyDown={e => { if (e.key === 'Enter') actions.openAsset(ticker) }}
              onDragStart={() => setDrag(ticker)} onDragEnd={() => { setDrag(null); setOver(null) }}
              onDragOver={e => { e.preventDefault(); setOver(ticker) }}
              onDrop={e => { e.preventDefault(); actions.reorderFav(drag, ticker); setDrag(null); setOver(null) }}>
              <div className="fhead"><Mark asset={asset} /><div><div className="fname">{asset.n}</div><div className="ftick">{ticker} · {asset.sec === 'fin' && asset.live ? 'MOEX' : sectorName(model, asset)}</div></div>
                <div className="fctrl">
                  <button className="icobtn fmove" aria-label="Переместить влево" onClick={e => { e.stopPropagation(); actions.favMove(ticker, -1) }}><Icon name="arrL" /></button>
                  <button className="icobtn fmove" aria-label="Переместить вправо" onClick={e => { e.stopPropagation(); actions.favMove(ticker, 1) }}><Icon name="arrR" /></button>
                  <span className="icobtn grip" title="Перетащите для порядка"><Icon name="grip" /></span>
                  <button className="icobtn star on" aria-label={`Убрать ${ticker} из избранного`} onClick={e => { e.stopPropagation(); actions.favToggle(ticker) }}><Icon name="star" /></button>
                </div></div>
              <div className="fprice">{fmt2(asset.p)}</div>
              <div className={`fchange ${change >= 0 ? 'up' : 'down'}`}><Icon name={change >= 0 ? 'up' : 'down'} style={tiny} /><span>{fmtPct(change)}</span></div>
              <Spark asset={asset} tickSeed={account.tickSeed} />
            </div>
          )
        }) : <div className="favempty enter"><Icon name="star" /><b>Пока пусто</b><p>Добавьте инструменты — и следите за котировками прямо здесь, без лишних экранов.</p><button className="btn btn-brand btn-sm" onClick={() => actions.market()}>Открыть все активы</button></div>}
      </div>
      <div className="sechead enter" style={{ '--d': '.14s' }}><h2>Новости рынка</h2><button className="link" onClick={() => actions.openNewsList()}>Все новости <Icon name="arrR" /></button></div>
      {news[0] && (
        <div className="newsrow enter" style={{ '--d': '.16s' }}>
          <article className="nfeature" role="button" tabIndex={0} onClick={() => actions.openNews(news[0].id)}>
            {news[0].art === 'apple' ? <NewsArt art="apple" /> : <div className="nimg"><NewsArt art={news[0].art} /></div>}
            <div className="nbody"><h3>{news[0].t}</h3><div className="nmeta"><Icon name="clock" /><span>{news[0].m}</span></div></div>
          </article>
          <div className="nside">
            {news.slice(1, 3).map(item => (
              <article key={item.id} className="nitem" role="button" tabIndex={0} onClick={() => actions.openNews(item.id)}>
                <div className="nthumb">{item.art === 'desk' ? <NewsArt art="desk" /> : item.art === 'flag' ? <NewsArt art="flag" /> : <Bull />}</div>
                <div className="ntxt"><span className="nchip2">{item.cat}</span><h4>{item.t}</h4><div className="nmeta"><span>{item.m}</span></div></div>
              </article>
            ))}
          </div>
        </div>
      )}
      <div className="bottomrow enter" style={{ '--d': '.2s' }}>
        <div className="card pcard">
          <div className="sechead" style={{ marginBottom: 12 }}><h2 style={{ fontSize: 17 }}>Ваш портфель</h2><button className="link" onClick={() => actions.nav('portfolio')}>Все детали <Icon name="arrR" /></button></div>
          <div className="pbody">
            <Donut state={account} />
            <div className="chartblock">
              <PeriodTabs />
              <LineChart vals={chart.vals} labels={chart.labels} />
            </div>
          </div>
        </div>
        <div className="card">
          <div className="lhead"><Icon name="book" /><h2>Обучающие материалы</h2></div>
          <div className="lbody"><div className="lthumb art-bull"><Bull /></div>
            <div className="ltxt"><h4>Как анализировать рынок: базовые методы</h4><div className="lmeta">{account.lessons.includes(1) ? 'Урок 1 пройден ✓' : 'Урок 1 из 8'} · 7 мин чтения</div>
              <div className="lprog"><i style={{ width: `${done / 8 * 100}%` }} /></div>
              <div className="lmeta" style={{ marginTop: 6, fontSize: 11 }}>Курс пройден на {Math.round(done / 8 * 100)}% · +25 XP за урок</div></div></div>
          <button className="btn btn-brand lbtn" onClick={() => actions.openLesson(1)}>Читать статью <Icon name="arrR" /></button>
        </div>
      </div>
    </>
  )
}

function sectorName(model, asset) {
  return model.sectors?.[asset.sec]?.n || ({ tech: 'Технологии', fin: 'Финансы', cons: 'Потребительский', ener: 'Энергетика' }[asset.sec] || asset.sec)
}

export function PeriodTabs() {
  const { model, actions } = useStudio()
  const chart = portfolioSeries(model.account)
  return (
    <div className="ctabs">
      {['1Д', '1Н', '1М', '3М', '1Г'].map(period => <button key={period} className={model.account.period === period ? 'on' : ''} onClick={() => actions.setPeriod(period)}>{period}</button>)}
      <span className={`cpchange ${chart.pct >= 0 ? 'up' : 'down'}`}>{fmtPct(chart.pct)}</span>
    </div>
  )
}

export function Portfolio() {
  const { model, actions } = useStudio()
  const account = model.account
  const tab = account.ptab
  const inv = investedVal(account)
  const upl = unrealPL(account)
  const d = dayPL(account)
  return (
    <>
      <div className="pagehead enter"><h1>Портфель</h1><div className="right">
        <button className="btn btn-ghost btn-sm" onClick={() => actions.exportCSV()}><Icon name="download" /> История CSV</button>
        <button className="btn btn-brand btn-sm" onClick={() => actions.openTopup()}><Icon name="plus" /> Пополнить</button>
      </div>
      <span className="sub">Позиции, структура и история операций · комиссия симулятора 0,1%</span></div>
      <div className="statgrid enter" style={{ '--d': '.05s' }}>
        <div className="stat"><div className="t">Стоимость счёта</div><b>{fmtMoney(totalVal(account), account.settings.compact)}</b><small className={myRet(account) >= 0 ? 'up' : 'down'}>{fmtPct(myRet(account))} с начала</small></div>
        <div className="stat"><div className="t">Инвестировано</div><b>{fmtMoney(inv, account.settings.compact)}</b><small style={{ color: 'var(--dim)' }}>{Object.keys(account.holdings).length} позиций</small></div>
        <div className="stat"><div className="t">Нереализованный P/L</div><b className={upl >= 0 ? 'up' : 'down'}>{moneySigned(upl)}</b><small style={{ color: 'var(--dim)' }}>к средней цене</small></div>
        <div className="stat"><div className="t">P/L за день</div><b className={d >= 0 ? 'up' : 'down'}>{moneySigned(d)}</b><small style={{ color: 'var(--dim)' }}>свободно {fmt0(account.cash)}</small></div>
      </div>
      <div className="tabs enter" style={{ '--d': '.08s' }}>
        <button className={tab === 'hold' ? 'on' : ''} onClick={() => actions.setPageTab('hold')}>Мои активы</button>
        <button onClick={() => actions.openMarket()}>Рынок</button>
        <button className={tab === 'hist' ? 'on' : ''} onClick={() => actions.setPageTab('hist')}>История · {account.tx.length}</button>
      </div>
      {tab === 'hold' && <Holdings />}
      {tab === 'hist' && <History />}
    </>
  )
}

function Holdings() {
  const { model, actions } = useStudio()
  const account = model.account
  const chart = portfolioSeries(account)
  const parts = allocOf(account)
  const sorted = Object.entries(account.holdings).sort((a, b) => {
    const left = findAsset(model, a[0])
    const right = findAsset(model, b[0])
    if (model.hsort === 'name') return left.n.localeCompare(right.n, 'ru') * model.hdir
    if (model.hsort === 'day') return (chgOf(left) - chgOf(right)) * model.hdir
    return (a[1].qty * left.p - b[1].qty * right.p) * model.hdir
  })
  const arrow = key => model.hsort === key ? <span className="arr">{model.hdir > 0 ? '▲' : '▼'}</span> : null
  return (
    <>
      <div className="allocgrid enter" style={{ '--d': '.1s' }}>
        <div className="card" style={{ display: 'flex', gap: 16, alignItems: 'center', flexWrap: 'wrap', justifyContent: 'center' }}><Donut state={account} /></div>
        <div className="card"><div className="sechead" style={{ marginBottom: 14 }}><h2 style={{ fontSize: 16 }}>Доли секторов</h2><span className="hint">лимит подсказки: сектор ≤ 45%</span></div>
          {parts.length ? parts.map(part => <div className="sectorbar" key={part.k}><div className="t"><b>{part.n}{part.pct > 45 ? <span className="down" style={{ fontSize: 11 }}> · перекос</span> : ''}</b><span>{fmt0(part.v)} · {Math.round(part.pct)}%</span></div><div className="bar"><i style={{ width: `${part.pct}%`, background: part.c }} /></div></div>) : <p style={{ color: 'var(--dim)', fontSize: 13 }}>Купите активы, чтобы увидеть структуру.</p>}
          <div className="chartblock" style={{ marginTop: 14 }}><PeriodTabs /><LineChart vals={chart.vals} labels={chart.labels} height={130} /></div>
        </div>
      </div>
      <div className="card tablecard enter" style={{ '--d': '.14s' }}>
        {model.hsel.length > 0 && <div className="bulkbar"><b>Выбрано: {model.hsel.length}</b><span className="sp" /><button className="btn btn-down btn-sm" onClick={() => actions.bulkSell()}><Icon name="down" /> Продать выбранное</button><button className="btn btn-ghost btn-sm" onClick={() => actions.clearHsel()}>Снять выделение</button></div>}
        <div className="tscroll"><table><thead><tr>
          <th style={{ width: 36 }}><input type="checkbox" className="ck" aria-label="Выбрать все позиции" checked={sorted.length > 0 && model.hsel.length === sorted.length} onChange={e => actions.toggleHall(e.target.checked)} /></th>
          <th className="sortable" onClick={() => actions.hsort('name')}>Актив{arrow('name')}</th>
          <th>Кол-во</th><th>Ср. цена</th><th>Цена</th>
          <th className="sortable" onClick={() => actions.hsort('value')}>Стоимость{arrow('value')}</th>
          <th>P/L</th><th className="sortable" onClick={() => actions.hsort('day')}>День{arrow('day')}</th><th />
        </tr></thead><tbody>
          {sorted.length ? sorted.map(([ticker, held]) => {
            const asset = findAsset(model, ticker)
            if (!asset) return null
            const value = held.qty * asset.p
            const pl = (asset.p - held.avg) * held.qty
            const change = chgOf(asset)
            return <tr key={ticker} className={model.hsel.includes(ticker) ? 'selected' : ''} onClick={() => actions.openAsset(ticker)}>
              <td onClick={e => e.stopPropagation()}><input type="checkbox" className="ck" checked={model.hsel.includes(ticker)} aria-label={`Выбрать позицию ${ticker}`} onChange={e => actions.toggleHsel(ticker, e.target.checked)} /></td>
              <td><div className="tasset"><Mark asset={asset} /><div><div className="n">{asset.n}</div><div className="t">{ticker}</div></div></div></td>
              <td className="num">{held.qty}</td><td className="num" style={{ color: 'var(--muted)' }}>{fmt2(held.avg)}</td><td className="num">{fmt2(asset.p)}</td>
              <td className="num">{fmt2(value)}</td><td className={`num ${pl >= 0 ? 'up' : 'down'}`}>{moneySigned(pl)}</td>
              <td className={`num ${change >= 0 ? 'up' : 'down'}`}>{fmtPct(change)}</td>
              <td onClick={e => e.stopPropagation()} style={{ whiteSpace: 'nowrap' }}><button className="btn btn-up btn-sm" onClick={() => actions.openTrade(ticker, 'buy')}>Купить</button> <button className="btn btn-down btn-sm" onClick={() => actions.openTrade(ticker, 'sell')}>Продать</button></td>
            </tr>
          }) : <tr><td colSpan={9}><div className="emptybox"><Icon name="case" /><b>Позиций пока нет</b><p>Купите первый актив на вкладке «Рынок» — портфель, донат и рейтинг обновятся сразу.</p><button className="btn btn-brand btn-sm" onClick={() => actions.openMarket()}>Перейти к рынку</button></div></td></tr>}
        </tbody></table></div>
      </div>
    </>
  )
}

function History() {
  const { model, actions } = useStudio()
  const filter = model.account.hist.f
  const list = [...model.account.tx].reverse().filter(row => filter === 'all' || row.type === filter)
  return (
    <>
      <div className="toolbar enter" style={{ '--d': '.1s' }}>
        {[['all', 'Все'], ['buy', 'Покупки'], ['sell', 'Продажи'], ['top', 'Пополнения']].map(([key, name]) => <button key={key} className={`chip ${filter === key ? 'on' : ''}`} onClick={() => actions.setHist(key)}>{name}</button>)}
        <button className="chip" style={{ marginLeft: 'auto' }} onClick={() => actions.exportCSV()}><Icon name="download" /> Экспорт CSV</button>
      </div>
      <div className="card enter" style={{ '--d': '.12s', padding: 0, overflow: 'hidden' }}>
        {list.length ? list.map(row => {
          const meta = row.type === 'top'
            ? { icon: 'coins', color: 'var(--info)', bg: 'rgba(79,141,247,.14)', name: 'Пополнение счёта', value: '+' + fmt2(row.amt), note: 'зачислено' }
            : row.type === 'buy'
              ? { icon: 'up', color: 'var(--up)', bg: 'rgba(46,204,113,.13)', name: `Покупка ${row.t} × ${row.qty}`, value: '−' + fmt2(row.amt), note: 'списано' }
              : { icon: 'down', color: 'var(--down)', bg: 'rgba(255,107,107,.12)', name: `Продажа ${row.t} × ${row.qty}`, value: '+' + fmt2(row.amt), note: 'зачислено' }
          const when = new Date(row.d).toLocaleString('ru-RU', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })
          return <div className="txrow" key={row.id}><span className="ti" style={{ background: meta.bg, color: meta.color }}><Icon name={meta.icon} /></span>
            <div><div className="tn">{meta.name}</div><div className="td">{when}{row.price ? ` · цена ${fmt2(row.price)} · комиссия 0,1%` : ''}</div></div>
            <div className="tv"><b style={{ color: meta.color }}>{meta.value}</b><small>{meta.note}</small></div></div>
        }) : <div className="emptybox"><Icon name="clock" /><b>Операций пока нет</b><p>Пополните счёт или совершите сделку — история появится здесь.</p></div>}
      </div>
    </>
  )
}

export function Rating() {
  const { model, actions } = useStudio()
  const account = model.account
  const mine = { n: 'Вы · АК', ini: 'АК', c: '#E8112D', ret: myRet(account), lvl: levelOf(account.xp), you: true }
  const all = [...TRADERS.map(trader => ({ ...trader, you: false })), mine].sort((a, b) => b.ret - a.ret)
  const rank = all.findIndex(item => item.you) + 1
  const ach = [
    { icon: 'zap', n: 'Первая сделка', d: 'Совершите любую покупку или продажу.', got: account.tx.some(row => row.type !== 'top') },
    { icon: 'layers', n: 'Диверсификатор', d: 'Держите позиции в 3+ секторах.', got: allocOf(account).length >= 3 },
    { icon: 'book', n: 'Ученик', d: 'Пройдите первый урок курса.', got: account.lessons.length >= 1 },
    { icon: 'target', n: 'Снайпер', d: 'Нереализованный P/L выше +$5 000.', got: unrealPL(account) > 5000 },
    { icon: 'medal', n: 'Уровень 3', d: 'Наберите 200 XP сделками и уроками.', got: levelOf(account.xp) >= 3 },
    { icon: 'coins', n: 'Крупный счёт', d: 'Стоимость счёта выше $150 000.', got: totalVal(account) > 150000 },
  ]
  return (
    <>
      <div className="pagehead enter"><h1>Рейтинг трейдеров</h1><div className="right"><button className="btn btn-ghost btn-sm" onClick={() => actions.tick()}><Icon name="trend" /> Обновить котировки</button></div>
        <span className="sub">Доходность считается от стартового капитала симулятора · ваша позиция: <b className="mono">#{rank}</b> из {all.length} · обновляется после каждой сделки и тика котировок</span></div>
      <div className="podium enter" style={{ '--d': '.06s' }}>{all.slice(0, 3).map((trader, i) => <div className={`pod ${i === 0 ? 'first' : ''}`} key={trader.n}><div className="medal">{['🥇', '🥈', '🥉'][i]}</div><div className="pav" style={{ background: trader.c }}>{trader.ini}</div><b>{trader.n}</b><div className={`ret ${trader.ret >= 0 ? 'up' : 'down'}`}>{fmtPct(trader.ret)}</div><small>Уровень {trader.lvl}{trader.you ? ' · это вы' : ''}</small></div>)}</div>
      <div className="card tablecard enter" style={{ '--d': '.1s' }}><div className="tscroll"><table style={{ minWidth: 560 }}><thead><tr><th>#</th><th>Трейдер</th><th>Уровень</th><th>Доходность</th><th>Статус</th></tr></thead><tbody>
        {all.map((trader, i) => <tr key={trader.n} className={trader.you ? 'you' : ''} style={{ cursor: 'default' }}><td className="num">{i + 1}</td><td><div className="tasset"><span className="flogo" style={{ background: trader.c, borderRadius: '50%' }}>{trader.ini}</span><div className="n">{trader.n}</div></div></td><td className="num">{trader.lvl}</td><td className={`num ${trader.ret >= 0 ? 'up' : 'down'}`}>{fmtPct(trader.ret)}</td><td>{trader.you ? <span className="tag buy">ваш аккаунт</span> : <span className="tag sec">{trader.ret > 15 ? 'агрессивный' : 'сбалансированный'}</span>}</td></tr>)}
      </tbody></table></div></div>
      <div className="sechead enter" style={{ '--d': '.14s', marginTop: 24 }}><h2>Достижения</h2><span className="hint">{ach.filter(item => item.got).length} из {ach.length} получено · XP: {account.xp}</span></div>
      <div className="achgrid enter" style={{ '--d': '.16s' }}>{ach.map(item => <div className={`ach ${item.got ? 'got' : ''}`} key={item.n}><span className="ai"><Icon name={item.icon} /></span><div><b>{item.n}</b><p>{item.d}</p><span className="st">{item.got ? '✓ получено' : 'в процессе'}</span></div></div>)}</div>
    </>
  )
}

export function Learn() {
  const { model, actions, lessons } = useStudio()
  const done = model.account.lessons.length
  const level = levelOf(model.account.xp)
  const next = lessons.find(lesson => !model.account.lessons.includes(lesson.id) && lesson.lvl <= level)?.id || 1
  return (
    <>
      <div className="pagehead enter"><h1>Обучение</h1><span className="sub">Курс «Базовые методы инвестора» · 8 уроков · XP за прохождение открывает новые уроки и уровни</span></div>
      <div className="card coursehead enter" style={{ '--d': '.05s' }}>
        <div className="lthumb art-bull" style={{ width: 86, height: 86 }}><Bull /></div>
        <div className="cp"><div className="t"><span>Прогресс курса</span><b>{done}/8 · {Math.round(done / 8 * 100)}%</b></div>
          <div className="lprog" style={{ height: 9 }}><i style={{ width: `${done / 8 * 100}%` }} /></div>
          <div className="t" style={{ marginTop: 10 }}><span>Ваш уровень: <b>{level}</b></span><span>{model.account.xp} XP всего</span></div></div>
        <button className="btn btn-brand" onClick={() => actions.openLesson(next)}><Icon name="book" /> Продолжить курс</button>
      </div>
      <div className="lessongrid enter" style={{ '--d': '.1s' }}>{lessons.map((lesson, i) => {
        const locked = lesson.lvl > level
        const passed = model.account.lessons.includes(lesson.id)
        return <div key={lesson.id} className={`lesson ${locked ? 'locked' : ''} ${passed ? 'done' : ''}`} style={{ '--d': `${i * .04}s` }} role="button" tabIndex={0} onClick={() => actions.openLesson(lesson.id)}>
          <div className="ln">УРОК {lesson.id} {passed ? <span className="ok"><Icon name="check" /> пройден</span> : locked ? <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}><Icon name="lock" /> уровень {lesson.lvl}</span> : ''}</div>
          <h3>{lesson.t}</h3><p>{lesson.d}</p>
          <div className="lf"><Icon name="clock" style={tiny} /> {lesson.min} мин <span className="xp">+{lesson.xp} XP</span></div>
        </div>
      })}</div>
    </>
  )
}

export function Settings() {
  const { model, actions } = useStudio()
  const settings = model.account.settings
  const row = (icon, title, text, control) => <div className="setrow"><span className="si"><Icon name={icon} /></span><div className="st"><b>{title}</b><p>{text}</p></div><div className="sc">{control}</div></div>
  const sw = (key, on) => <button className="switch" role="switch" aria-checked={on} aria-label={key} onClick={() => actions.setting(key)} />
  return (
    <>
      <div className="pagehead enter"><h1>Настройки</h1><span className="sub">Все изменения применяются сразу и сохраняются в этом браузере</span></div>
      <div className="setgrid enter" style={{ '--d': '.06s' }}>
        <div className="card"><div className="sechead" style={{ marginBottom: 6 }}><h2 style={{ fontSize: 16 }}>Интерфейс</h2></div>
          {row('palette', 'Семантическая палитра', 'Классика: зелёный рост / красный спад. Дейтеран-режим: синий / оранжевый — различимо при дальтонизме.',
            <select className="sel" aria-label="Палитра" value={settings.palette} onChange={e => actions.palette(e.target.value)}><option value="classic">Классическая</option><option value="deutan">Синий / оранжевый</option></select>)}
          <div className="palpreview" style={{ margin: '-6px 0 10px 54px' }}><i style={{ background: 'var(--up)' }} /><i style={{ background: 'var(--down)' }} /><i style={{ background: '#4F8DF7' }} /><i style={{ background: '#F5A623' }} /><i style={{ background: '#EC4899' }} /><i style={{ background: '#2DD4BF' }} /></div>
          {row('coins', 'Компактные суммы', 'Показывать $128,4к вместо полных значений на карточках счёта.', sw('compact', settings.compact))}
          {row('eye', 'Анимации и фон', 'Живой фон, входные анимации, конфетти и плавные числа.', sw('motion', settings.motion === 'on'))}
          {row('zap', 'Звук интерфейса', 'Клики, подтверждения сделок и ошибки. Управляется и кнопкой в шапке.', sw('sound', settings.sound))}
        </div>
        <div className="card"><div className="sechead" style={{ marginBottom: 6 }}><h2 style={{ fontSize: 16 }}>Данные симулятора</h2></div>
          {row('trend', 'Автообновление котировок', 'Лёгкий тик цен каждые 45 секунд, пока вкладка активна.', sw('autoTick', settings.autoTick))}
          {row('download', 'Экспорт состояния', 'Полный JSON: счёт, позиции, избранное, история, прогресс.', <button className="btn btn-ghost btn-sm" onClick={() => actions.exportJSON()}><Icon name="download" /> JSON</button>)}
          {row('download', 'Экспорт истории', 'CSV всех операций для таблицы или отчёта.', <button className="btn btn-ghost btn-sm" onClick={() => actions.exportCSV()}><Icon name="download" /> CSV</button>)}
          {row('undo', 'Сброс демо-данных', 'Вернуть стартовый счёт, позиции и прогресс. Можно отменить сразу после сброса.', <button className="btn btn-down btn-sm" onClick={() => actions.reset()}><Icon name="trash" /> Сбросить</button>)}
        </div>
      </div>
      <div className="card enter" style={{ '--d': '.1s', marginTop: 16 }}><div className="sechead" style={{ marginBottom: 8 }}><h2 style={{ fontSize: 16 }}>Что исправлено в этом редизайне</h2></div>
        <p style={{ fontSize: 13, color: 'var(--muted)', maxWidth: 820 }}>Бренд-красный отделён от семантики данных; весь вторичный текст ≥ 4.5:1; графики получили оси, тултипы и маркеры последних точек; донат — четыре различимых hue с прямыми подписями; единый форматтер Intl ru-RU; пустые состояния избранного, портфеля, истории и поиска; тач-таргеты ≥ 44px; focus-состояния и клавиатура; актуальные даты новостей с источниками.</p></div>
    </>
  )
}

export function Drawers() {
  const { model, actions, lessons } = useStudio()
  const drawer = model.drawer
  if (!drawer) return null
  if (drawer.type === 'asset') return <AssetDrawer />
  if (drawer.type === 'lesson') {
    const lesson = lessons.find(item => item.id === drawer.id)
    if (!lesson) return null
    const done = model.account.lessons.includes(lesson.id)
    const art = lesson.id <= 3 ? null : [['Скоро в курсе', `Этот урок откроется на уровне ${lesson.lvl}. Проходите доступные уроки и совершайте сделки, чтобы набрать XP.`]]
    const blocks = lesson.id === 1 || lesson.id === 2 || lesson.id === 3 ? requireArt(lesson.id) : art
    const chips = lesson.id === 1 ? ['P/E', 'маржинальность', 'тренд', 'объём', 'диверсификация', 'стоп-лосс'] : lesson.id === 2 ? ['P/E', 'прогнозный P/E', 'цикличность', 'рост прибыли'] : ['сектора', 'ребалансировка', 'лимит 45%', 'просадка']
    return <Shell title={`Обучение · урок ${lesson.id}`} sub={lesson.t}>
      <div className="lesson-art art-bull"><Bull /></div>
      <div className="lesson-body"><p className="mono" style={{ fontSize: 11.5, color: 'var(--dim)' }}>УРОК {lesson.id} · {lesson.min} МИН · +{lesson.xp} XP {done ? '· ПРОЙДЕН ✓' : ''}</p>
        <h3 style={{ fontSize: 19, marginBottom: 10 }}>{lesson.t}</h3>
        {blocks.map(([head, text]) => <div key={head}><h4>{head}</h4><p>{text}</p></div>)}
        <div className="term-chips">{chips.map(chip => <span key={chip}>{chip}</span>)}</div>
        <button className={`btn ${done ? 'btn-ghost' : 'btn-brand'}`} style={{ width: '100%' }} onClick={() => actions.finishLesson(lesson.id)}><Icon name="check" /> {done ? 'Урок пройден' : `Отметить пройденным · +${lesson.xp} XP`}</button>
      </div>
    </Shell>
  }
  if (drawer.type === 'news') {
    const item = model.account.news.find(row => row.id === drawer.id)
    if (!item) return null
    return <Shell title="Новости рынка" sub={`${item.cat} · ${item.m}`}>
      <div className={`lesson-art ${item.art === 'apple' ? 'art-apple' : item.art === 'bull' ? 'art-bull' : ''}`} style={item.art === 'desk' ? { background: '#0E1626' } : item.art === 'flag' ? { background: '#F5F0EB' } : undefined}><NewsArt art={item.art} /></div>
      <div className="newsreader"><span className="nchip2">{item.cat}</span><h3 style={{ fontSize: 19, margin: '8px 0 4px' }}>{item.t}</h3>
        <div className="nmeta" style={{ marginBottom: 12 }}><Icon name="clock" /><span>{item.m}</span></div>
        {item.body.map(([head, text]) => <div key={head}><h4>{head}</h4><p>{text}</p></div>)}
        {item.link && <p><a href={item.link} target="_blank" rel="noreferrer">Источник</a></p>}
        <button className="btn btn-ghost" style={{ width: '100%', marginTop: 8 }} onClick={() => actions.openNewsList()}><Icon name="arrL" /> Ко всем новостям</button>
      </div>
    </Shell>
  }
  return <Shell title="Все новости" sub="лента · 22 ноября 2026">
    {model.account.news.map(item => <article key={item.id} className="nitem" style={{ marginBottom: 12 }} role="button" tabIndex={0} onClick={() => actions.openNews(item.id)}>
      <div className={`nthumb ${item.art === 'apple' ? 'art-apple' : ''}`}><NewsArt art={item.art === 'apple' ? 'apple' : item.art} /></div>
      <div className="ntxt"><span className="nchip2">{item.cat}</span><h4>{item.t}</h4><div className="nmeta"><span>{item.m}</span></div></div>
    </article>)}
  </Shell>
}

function requireArt(id) {
  const packs = {
    1: [['1. Фундаментальный анализ: что покупать', 'Смотрите на бизнес, а не на тикер. Выручка и её рост за 3–5 лет, маржинальность, долговая нагрузка. P/E сравнивайте внутри сектора.'], ['2. Технический анализ: когда смотреть', 'Тренд — ваш контекст: серия повышающихся максимумов и минимумов важнее любой свечи. Объём подтверждает движение.'], ['3. Риск-менеджмент: сколько держать', 'Одна идея — не больше 5–10% портфеля. Стоп-лосс — это план, а не паника.']],
    2: [['Что показывает P/E', 'Цена акции, делённая на прибыль на акцию. P/E 30 — грубо «окупаемость за 30 лет при текущей прибыли».'], ['Когда P/E врёт', 'Циклические компании на пике прибыли выглядят дешёвыми прямо перед спадом. Сравнивайте P/E внутри сектора.'], ['Практика в симуляторе', 'Откройте карточку NVDA и JPM: P/E 54 против 14. Рынок закладывает разный темп роста.']],
    3: [['Зачем сектора', 'Один сектор — один общий риск. Когда технологии падают на новостях о ставке, финансы могут расти.'], ['Сколько вешать в процентах', 'Ни один сектор не больше 45%, ни одна акция не больше 15%.'], ['Практика', 'Продайте часть крупнейшей позиции и купите актив из пустого сектора.']],
  }
  return packs[id]
}

function AssetDrawer() {
  const { model, actions, sectors } = useStudio()
  const ticker = model.drawer.t
  const asset = findAsset(model, ticker)
  if (!asset) return null
  const account = model.account
  const change = chgOf(asset)
  const pos = account.holdings[ticker]
  const live = model.closes?.ticker === ticker && model.closes.period === account.period ? model.closes.vals : null
  const series = assetSeries(asset, account.period, account.tickSeed, live)
  const qty = Math.floor(model.trade.qty) || 0
  const gross = qty * asset.p
  const fee = gross * FEE
  const total = model.trade.side === 'buy' ? gross + fee : gross - fee
  let error = ''
  if (qty <= 0) error = 'Введите количество больше нуля.'
  else if (model.trade.side === 'buy' && total > account.cash) error = `Недостаточно свободных средств: не хватает ${fmt2(total - account.cash)}. Пополните счёт или уменьшите объём.`
  else if (model.trade.side === 'sell' && (!pos || qty > pos.qty)) error = `Нет позиции для продажи: доступно ${pos ? pos.qty : 0} шт.`
  const venue = asset.live ? 'MOEX TQBR' : 'NASDAQ/NYSE'
  return <Shell title={`${asset.n} · ${ticker}`} sub={`карточка актива · ${asset.live ? 'котировка площадки' : 'демо-котировки'}`}>
    <div className="ad-top"><Mark asset={asset} /><div><h4>{asset.n}</h4><span className="t">{ticker} · {sectors[asset.sec]?.n || 'Финансы'} · {venue}</span></div>
      <button className={`icobtn star ${account.favs.includes(ticker) ? 'on' : ''}`} style={{ marginLeft: 'auto' }} aria-label={`Избранное ${ticker}`} onClick={() => actions.favToggle(ticker)}><Icon name="star" /></button></div>
    <div className="ad-price"><b>{fmt2(asset.p)}</b><span className={`chg ${change >= 0 ? 'up' : 'down'}`}><Icon name={change >= 0 ? 'up' : 'down'} />{fmtPct(change)} сегодня</span></div>
    <div className="ctabs">{['1Д', '1Н', '1М', '3М', '1Г'].map(period => <button key={period} className={account.period === period ? 'on' : ''} onClick={() => actions.setPeriod(period)}>{period}</button>)}</div>
    <LineChart vals={series.vals} labels={series.labels} height={150} />
    <div className="ad-stats">
      <div className="ad-stat"><div className="t">Диапазон дня</div><b>{fmt2(asset.rng[0])} – {fmt2(asset.rng[1])}</b></div>
      <div className="ad-stat"><div className="t">P/E</div><b>{asset.live && !asset.pe ? '—' : Number(asset.pe).toFixed(1)}</b></div>
      <div className="ad-stat"><div className="t">Капитализация</div><b>{asset.cap}</b></div>
      <div className="ad-stat"><div className="t">Комиссия</div><b>0,1%</b></div>
    </div>
    {pos ? <div className="ad-pos"><div className="t">Ваша позиция</div>
      <div className="r"><span>Количество</span><b>{pos.qty} шт.</b></div>
      <div className="r"><span>Средняя цена</span><b>{fmt2(pos.avg)}</b></div>
      <div className="r"><span>Стоимость</span><b>{fmt2(pos.qty * asset.p)}</b></div>
      <div className="r"><span>Нереализованный P/L</span><b className={asset.p - pos.avg >= 0 ? 'up' : 'down'}>{moneySigned((asset.p - pos.avg) * pos.qty)}</b></div>
    </div> : <div className="ad-pos" style={{ background: 'var(--surface)', borderColor: 'var(--line)' }}><div className="t" style={{ color: 'var(--dim)' }}>Позиции нет</div><div className="r"><span>Свободно для покупки</span><b>{fmt2(account.cash)}</b></div></div>}
    <div className="tradetabs"><button className={model.trade.side === 'buy' ? 'on-buy' : ''} onClick={() => actions.setSide('buy')}>Купить</button><button className={model.trade.side === 'sell' ? 'on-sell' : ''} onClick={() => actions.setSide('sell')}>Продать</button></div>
    <label htmlFor="tradeQty">Количество акций</label>
    <div className="qtyrow"><div className="minput"><input id="tradeQty" inputMode="numeric" value={model.trade.qty} aria-label="Количество акций" onChange={e => actions.setQty(Math.max(0, Math.floor(parseFloat(e.target.value) || 0)))} onKeyDown={e => { if (e.key === 'Enter') actions.trade() }} /><span style={{ fontSize: 13 }}>шт.</span></div>
      <div className="stepper"><button aria-label="Плюс одна" onClick={() => actions.bumpQty(1)}><Icon name="plus" /></button><button aria-label="Минус одна" onClick={() => actions.bumpQty(-1)}><Icon name="minus" /></button></div></div>
    <div className="mquick">{[25, 50, 75, 100].map(pct => <button key={pct} onClick={() => actions.qtyPct(pct)}>{pct}%</button>)}</div>
    <div className="est">
      <div className="er"><span>Цена</span><b>{fmt2(asset.p)}</b></div>
      <div className="er"><span>Количество</span><b>{qty} шт.</b></div>
      <div className="er"><span>Сумма</span><b>{fmt2(gross)}</b></div>
      <div className="er"><span>Комиссия 0,1%</span><b>{fmt2(fee)}</b></div>
      <div className="er total"><span>{model.trade.side === 'buy' ? 'Списание со счёта' : 'Зачисление на счёт'}</span><b>{fmt2(total)}</b></div>
      <div className="er"><span>{model.trade.side === 'buy' ? 'Свободно сейчас' : 'В позиции сейчас'}</span><b>{model.trade.side === 'buy' ? fmt2(account.cash) : (pos ? `${pos.qty} шт.` : '0 шт.')}</b></div>
    </div>
    {error && <p className="merr"><Icon name="alert" />{error}</p>}
    <button className={`btn ${model.trade.side === 'buy' ? 'btn-up' : 'btn-down'}`} style={{ width: '100%' }} disabled={!!error} onClick={() => actions.trade()}>{model.trade.side === 'buy' ? <><Icon name="up" /> Купить</> : <><Icon name="down" /> Продать</>} {ticker}</button>
  </Shell>
}

function Shell({ title, sub, children }) {
  const { actions } = useStudio()
  return (
    <>
      <div className="scrim open" onClick={() => actions.closeDrawer()} />
      <aside className="drawer open" role="dialog" aria-modal="true" aria-label="Панель деталей">
        <div className="dr-head"><div><h3>{title}</h3><span className="dr-sub">{sub}</span></div>
          <button className="dr-close" aria-label="Закрыть панель" onClick={() => actions.closeDrawer()}><Icon name="x" /></button></div>
        <div className="dr-body">{children}</div>
      </aside>
    </>
  )
}

export { moneySigned }
