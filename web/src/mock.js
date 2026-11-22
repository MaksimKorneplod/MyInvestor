export const SECTORS = {
  tech: { n: 'Технологии', c: '#4F8DF7' },
  fin: { n: 'Финансы', c: '#F5A623' },
  cons: { n: 'Потребительский', c: '#EC4899' },
  ener: { n: 'Энергетика', c: '#2DD4BF' },
}

export const INSTRUMENTS = [
  { t: 'NVDA', n: 'NVIDIA', sec: 'tech', p: 182.45, pc: 178.28, seed: 11, bg: '#76B900', logo: 'nv', pe: 54.2, cap: '4,5 трлн', rng: [176.9, 184.1] },
  { t: 'AAPL', n: 'Apple', sec: 'tech', p: 267.12, pc: 269.38, seed: 22, bg: '#2A2A33', logo: 'apple', pe: 32.8, cap: '4,0 трлн', rng: [264.2, 270.9] },
  { t: 'TSLA', n: 'Tesla', sec: 'cons', p: 345.20, pc: 341.07, seed: 33, bg: '#E8112D', logo: 'tesla', pe: 71.4, cap: '1,1 трлн', rng: [338.4, 349.6] },
  { t: 'MSFT', n: 'Microsoft', sec: 'tech', p: 512.40, pc: 509.24, seed: 44, bg: '#4F8DF7', logo: 'letter', pe: 36.1, cap: '3,8 трлн', rng: [506.1, 515.8] },
  { t: 'GOOGL', n: 'Alphabet', sec: 'tech', p: 231.15, pc: 231.96, seed: 55, bg: '#F5A623', logo: 'letter', pe: 24.6, cap: '2,8 трлн', rng: [228.4, 233.7] },
  { t: 'AMZN', n: 'Amazon', sec: 'cons', p: 264.80, pc: 262.05, seed: 66, bg: '#EC4899', logo: 'letter', pe: 41.3, cap: '2,7 трлн', rng: [260.3, 267.2] },
  { t: 'JPM', n: 'JPMorgan', sec: 'fin', p: 302.15, pc: 300.71, seed: 77, bg: '#2DD4BF', logo: 'letter', pe: 13.9, cap: '830 млрд', rng: [298.2, 304.5] },
  { t: 'XOM', n: 'ExxonMobil', sec: 'ener', p: 141.62, pc: 142.93, seed: 88, bg: '#8E5CF7', logo: 'letter', pe: 14.7, cap: '590 млрд', rng: [140.1, 143.8] },
  { t: 'META', n: 'Meta Platforms', sec: 'tech', p: 718.30, pc: 710.42, seed: 99, bg: '#3FA7FF', logo: 'letter', pe: 27.9, cap: '1,8 трлн', rng: [706.5, 724.1] },
  { t: 'V', n: 'Visa', sec: 'fin', p: 358.90, pc: 357.12, seed: 111, bg: '#F5A623', logo: 'letter', pe: 31.2, cap: '700 млрд', rng: [354.8, 361.3] },
  { t: 'KO', n: 'Coca-Cola', sec: 'cons', p: 74.85, pc: 75.10, seed: 122, bg: '#E8465A', logo: 'letter', pe: 25.4, cap: '322 млрд', rng: [74.1, 75.6] },
  { t: 'NFLX', n: 'Netflix', sec: 'tech', p: 1204.60, pc: 1188.20, seed: 133, bg: '#C40E2E', logo: 'letter', pe: 48.7, cap: '515 млрд', rng: [1180.0, 1212.4] },
]

export const LESSONS = [
  { id: 1, t: 'Как анализировать рынок: базовые методы', d: 'Фундаментал, техника и риск-менеджмент — три опоры первого портфеля.', min: 7, xp: 25, lvl: 1 },
  { id: 2, t: 'Что такое P/E и когда он врёт', d: 'Почему низкий P/E не всегда означает дешёвую акцию.', min: 6, xp: 25, lvl: 1 },
  { id: 3, t: 'Диверсификация: сколько секторов нужно', d: 'Как распределение по секторам снижает просадку сильнее угадывания точек входа.', min: 5, xp: 25, lvl: 1 },
  { id: 4, t: 'Отчёты компаний: что читать первым', d: 'Выручка, маржа, guidance — порядок чтения квартального отчёта.', min: 8, xp: 30, lvl: 2 },
  { id: 5, t: 'Стоп-лосс и тейк-профит без паники', d: 'Правила выхода, которые задаются до входа, а не во время просадки.', min: 6, xp: 30, lvl: 3 },
  { id: 6, t: 'Дивиденды: доходность и ловушки', d: 'Почему 12% дивдоходности иногда дороже, чем 3%.', min: 7, xp: 35, lvl: 4 },
  { id: 7, t: 'IPO и сплиты: что происходит с ценой', d: 'Механика первичных размещений и дробления акций.', min: 6, xp: 35, lvl: 5 },
  { id: 8, t: 'Психология трейдера: дисциплина > прогноз', d: 'Как журнал сделок и лимиты защищают от тильта.', min: 9, xp: 40, lvl: 6 },
]

export const LESSON_ART = {
  1: [['1. Фундаментальный анализ: что покупать', 'Смотрите на бизнес, а не на тикер. Выручка и её рост за 3–5 лет, маржинальность, долговая нагрузка. P/E сравнивайте внутри сектора.'], ['2. Технический анализ: когда смотреть', 'Тренд — ваш контекст: серия повышающихся максимумов и минимумов важнее любой свечи. Объём подтверждает движение.'], ['3. Риск-менеджмент: сколько держать', 'Одна идея — не больше 5–10% портфеля. Стоп-лосс — это план, а не паника.']],
  2: [['Что показывает P/E', 'Цена акции, делённая на прибыль на акцию. P/E 30 — грубо «окупаемость за 30 лет при текущей прибыли».'], ['Когда P/E врёт', 'Циклические компании на пике прибыли выглядят дешёвыми прямо перед спадом. Сравнивайте P/E внутри сектора.'], ['Практика в симуляторе', 'Откройте карточку NVDA и JPM: P/E 54 против 14. Рынок закладывает разный темп роста.']],
  3: [['Зачем сектора', 'Один сектор — один общий риск. Когда технологии падают на новостях о ставке, финансы могут расти.'], ['Сколько вешать в процентах', 'Ни один сектор не больше 45%, ни одна акция не больше 15%.'], ['Практика', 'Продайте часть крупнейшей позиции и купите актив из пустого сектора.']],
}

export const MOCK_NEWS = [
  { id: 1, cat: 'Технологии', t: 'Apple показала новые ИИ-функции на WWDC 2026', m: 'Сегодня · 22 ноября · 5 мин', art: 'apple', body: [['Что показали', 'Обновлённая Siri работает на устройстве, генеративные функции встроены в заметки, почту и Xcode.'], ['Реакция рынка', 'Акции AAPL прибавили на премаркете, затем откатились на фиксации прибыли.']] },
  { id: 2, cat: 'Рынок', t: 'Рынок США закрылся ростом на фоне отчётов', m: 'Вчера · 21 ноября · 3 мин', art: 'desk', body: [['Итоги дня', 'S&P 500 +0,7%, Nasdaq +1,1%. Лучше сектора — полупроводники.']] },
  { id: 3, cat: 'Экономика', t: 'Инфляция в США замедлилась до 2,8%', m: '20 ноября · 2 мин · BLS', art: 'flag', body: [['Данные', 'Годовой CPI ниже прогноза. Базовая инфляция тоже лучше ожиданий.']] },
  { id: 4, cat: 'Энергетика', t: 'Нефть Brent ниже $72 на данных по запасам', m: '19 ноября · 2 мин', art: 'desk', body: [['Запасы', 'Коммерческие запасы нефти США выросли четвёртую неделю подряд.']] },
  { id: 5, cat: 'Обучение', t: 'Новый урок курса: «Что такое P/E и когда он врёт»', m: '18 ноября · 6 мин', art: 'bull', body: [['В курсе', 'Второй урок разбирает, почему низкий P/E не всегда означает дешёвую акцию.']] },
]

export const TRADERS = [
  { n: 'Kai Rivera', ini: 'KR', c: '#4F8DF7', ret: 24.6, lvl: 7 },
  { n: 'Mira Chen', ini: 'MC', c: '#EC4899', ret: 19.2, lvl: 6 },
  { n: 'Олег Д.', ini: 'ОД', c: '#2DD4BF', ret: 15.8, lvl: 6 },
  { n: 'Sofia Park', ini: 'SP', c: '#F5A623', ret: 11.3, lvl: 5 },
  { n: 'Данил К.', ini: 'ДК', c: '#8E5CF7', ret: 7.9, lvl: 4 },
  { n: 'Leo Martins', ini: 'LM', c: '#3FA7FF', ret: 4.2, lvl: 3 },
  { n: 'Настя В.', ini: 'НВ', c: '#E8465A', ret: 1.6, lvl: 2 },
]

export const PERIODS = {
  '1Д': { pct: .42, seed: 101, n: 26, days: 0, api: 'day' },
  '1Н': { pct: 1.1, seed: 202, n: 28, days: 6, api: 'week' },
  '1М': { pct: 3.9, seed: 303, n: 30, days: 29, api: 'month' },
  '3М': { pct: 8.4, seed: 404, n: 32, days: 90, api: 'halfYear' },
  '1Г': { pct: 21.7, seed: 505, n: 34, days: 364, api: 'year' },
}

const MO = ['янв', 'фев', 'мар', 'апр', 'мая', 'июн', 'июл', 'авг', 'сен', 'окт', 'ноя', 'дек']
export function periodLabels(key, n) {
  const out = []
  const end = new Date(2026, 10, 22)
  for (let i = 0; i < n; i++) {
    const f = i / (n - 1)
    if (key === '1Д') {
      const m = 600 + Math.round(f * 480)
      out.push(String(Math.floor(m / 60)).padStart(2, '0') + ':' + String(m % 60).padStart(2, '0'))
    } else {
      const d = new Date(end)
      d.setDate(d.getDate() - Math.round((1 - f) * PERIODS[key].days))
      out.push(d.getDate() + ' ' + MO[d.getMonth()])
    }
  }
  return out
}

export function lcg(seed) { let s = seed >>> 0; return () => { s = (Math.imul(s, 1664525) + 1013904223) >>> 0; return s / 4294967296 } }
export function walk(seed, n) {
  const r = lcg(seed)
  const a = [0]
  for (let i = 1; i < n; i++) a.push(a[i - 1] + (r() - .5))
  const d = a[n - 1] - a[0]
  for (let i = 0; i < n; i++) a[i] -= d * i / (n - 1)
  return a
}
export function series(seed, n, endDelta, amp) {
  const w = walk(seed, n)
  const mx = Math.max(...w.map(v => Math.abs(v))) || 1
  return w.map((v, i) => endDelta * i / (n - 1) + v / mx * amp)
}

const nf2 = new Intl.NumberFormat('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
const nf0 = new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 0 })
export const fmt2 = v => '$' + nf2.format(v)
export const fmt0 = v => '$' + nf0.format(v)
export const fmtK = v => '$' + new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 1 }).format(v / 1000) + 'к'
export const fmtPct = v => (v >= 0 ? '+' : '−') + nf2.format(Math.abs(v)) + '%'
export const fmtMoney = (v, compact) => compact && Math.abs(v) >= 100000 ? fmtK(v) : fmt2(v)
export const chgOf = x => (x.p / x.pc - 1) * 100
export const levelOf = xp => Math.floor(xp / 100) + 1
export const uid = () => Math.random().toString(36).slice(2, 9)

export function seedAccount() {
  const instruments = INSTRUMENTS.map(x => ({ ...x, rng: [...x.rng] }))
  const find = t => instruments.find(x => x.t === t)
  const holdings = { NVDA: { qty: 110, avg: 164.2 }, AAPL: { qty: 55, avg: 248.9 }, MSFT: { qty: 8, avg: 470 }, JPM: { qty: 95, avg: 271.5 }, TSLA: { qty: 40, avg: 310.4 }, AMZN: { qty: 28, avg: 240.6 }, XOM: { qty: 38, avg: 132.8 } }
  const invested = Object.entries(holdings).reduce((a, [t, h]) => a + h.qty * find(t).p, 0)
  return {
    instruments,
    cash: 34200,
    holdings,
    favs: ['NVDA', 'AAPL', 'TSLA'],
    xp: 15,
    lessons: [],
    tx: [
      { id: 'top1', type: 'top', amt: 10000, d: '2026-11-18T10:12:00' },
      { id: 'b1', type: 'buy', t: 'NVDA', qty: 60, price: 166.1, amt: 60 * 166.1 * 1.001, d: '2026-11-18T12:40:00' },
      { id: 'b2', type: 'buy', t: 'JPM', qty: 40, price: 274.3, amt: 40 * 274.3 * 1.001, d: '2026-11-19T15:05:00' },
      { id: 's1', type: 'sell', t: 'TSLA', qty: 10, price: 332.8, amt: 10 * 332.8 * .999, d: '2026-11-21T11:22:00' },
    ],
    notif: [
      { id: 'n1', icon: 'trend', c: 'up', t: 'NVDA +2,34% — лучшая в избранном', s: '5 мин назад · алерт watchlist', read: false, act: 'asset', arg: 'NVDA' },
      { id: 'n2', icon: 'book', c: 'info', t: 'Новый урок: «Что такое P/E»', s: '2 ч назад · курс «Базовые методы»', read: false, act: 'lesson', arg: 2 },
    ],
    views: [{ id: 'v1', name: 'Технологии · рост', q: '', sec: 'tech', sort: 'chg', dir: -1 }],
    baseTotal: 34200 + invested,
    tickSeed: 0,
    settings: { palette: 'classic', compact: false, sound: true, autoTick: true, motion: 'on' },
    mkt: { q: '', sec: 'all', sort: 'chg', dir: -1 },
    hist: { f: 'all' },
    ptab: 'hold',
    period: '1М',
    news: MOCK_NEWS,
    newsLive: false,
    catalog: null,
  }
}

export function investedVal(s) {
  return Object.entries(s.holdings).reduce((a, [t, h]) => a + h.qty * (s.instruments.find(x => x.t === t)?.p || 0), 0)
}
export function totalVal(s) { return s.cash + investedVal(s) }
export function dayPL(s) {
  return Object.entries(s.holdings).reduce((a, [t, h]) => {
    const x = s.instruments.find(i => i.t === t)
    return a + (x ? h.qty * (x.p - x.pc) : 0)
  }, 0)
}
export function unrealPL(s) {
  return Object.entries(s.holdings).reduce((a, [t, h]) => {
    const x = s.instruments.find(i => i.t === t)
    return a + (x ? h.qty * (x.p - h.avg) : 0)
  }, 0)
}
export function allocOf(s) {
  const inv = investedVal(s)
  const m = {}
  Object.entries(s.holdings).forEach(([t, h]) => {
    const x = s.instruments.find(i => i.t === t)
    if (!x) return
    m[x.sec] = (m[x.sec] || 0) + h.qty * x.p
  })
  return Object.entries(m).map(([k, v]) => ({ k, n: SECTORS[k].n, c: SECTORS[k].c, v, pct: inv ? v / inv * 100 : 0 })).sort((a, b) => b.v - a.v)
}
export const myRet = s => (totalVal(s) / s.baseTotal - 1) * 100
export const FEE = 0.001
