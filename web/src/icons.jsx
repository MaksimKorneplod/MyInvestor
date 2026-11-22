export function Icon({ name, className, style }) {
  return (
    <svg className={className} style={style} width="1em" height="1em" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      {glyph(name)}
    </svg>
  )
}

function glyph(name) {
const PATHS = {
  home: <><path d="M3 11l9-8 9 8" /><path d="M5 10v10h14V10" /></>,
  case: <><rect x="3" y="7" width="18" height="13" rx="2" /><path d="M8 7V5a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2M3 12h18" /></>,
  trophy: <><path d="M8 21h8M12 17v4M7 4h10v5a5 5 0 0 1-10 0V4z" /><path d="M7 6H4a3 3 0 0 0 3 5M17 6h3a3 3 0 0 1-3 5" /></>,
  book: <><path d="M12 6c-2-1.5-5-2-8-2v14c3 0 6 .5 8 2 2-1.5 5-2 8-2V4c-3 0-6 .5-8 2z" /><path d="M12 6v14" /></>,
  gear: <><circle cx="12" cy="12" r="3.2" /><path d="M12 2v3M12 19v3M2 12h3M19 12h3M4.9 4.9l2.1 2.1M17 17l2.1 2.1M19.1 4.9L17 7M7 17l-2.1 2.1" /></>,
  search: <><circle cx="11" cy="11" r="7" /><path d="M21 21l-4.3-4.3" /></>,
  star: <path d="M12 3l2.7 5.6 6.1.9-4.4 4.3 1 6.1-5.4-2.9-5.4 2.9 1-6.1L3.2 9.5l6.1-.9L12 3z" />,
  arrR: <path d="M5 12h14M13 6l6 6-6 6" />,
  arrL: <path d="M19 12H5M11 6l-6 6 6 6" />,
  up: <path d="M12 19V5M6 11l6-6 6 6" />,
  down: <path d="M12 5v14M6 13l6 6 6-6" />,
  trend: <><path d="M3 17l6-6 4 4 8-8" /><path d="M15 7h6v6" /></>,
  coins: <><ellipse cx="12" cy="6" rx="8" ry="3" /><path d="M4 6v6c0 1.7 3.6 3 8 3s8-1.3 8-3V6" /><path d="M4 12v6c0 1.7 3.6 3 8 3s8-1.3 8-3v-6" /></>,
  clock: <><circle cx="12" cy="12" r="9" /><path d="M12 7v5l3 2" /></>,
  check: <path d="M4 12l5 5L20 6" />,
  x: <path d="M6 6l12 12M18 6L6 18" />,
  alert: <><path d="M12 3l10 18H2L12 3z" /><path d="M12 10v4M12 17.5v.5" /></>,
  plus: <path d="M12 5v14M5 12h14" />,
  minus: <path d="M5 12h14" />,
  zap: <path d="M13 2L4 14h6l-1 8 9-12h-6l1-8" />,
  download: <><path d="M12 3v12M7 10l5 5 5-5" /><path d="M4 21h16" /></>,
  grip: <><circle cx="9" cy="6" r="1.4" /><circle cx="15" cy="6" r="1.4" /><circle cx="9" cy="12" r="1.4" /><circle cx="15" cy="12" r="1.4" /><circle cx="9" cy="18" r="1.4" /><circle cx="15" cy="18" r="1.4" /></>,
  lock: <><rect x="5" y="11" width="14" height="10" rx="2" /><path d="M8 11V7a4 4 0 0 1 8 0v4" /></>,
  medal: <><circle cx="12" cy="9" r="6" /><path d="M8.5 14L7 22l5-3 5 3-1.5-8" /></>,
  target: <><circle cx="12" cy="12" r="9" /><circle cx="12" cy="12" r="5" /><circle cx="12" cy="12" r="1.2" /></>,
  layers: <><path d="M12 2l10 6-10 6L2 8l10-6z" /><path d="M2 14l10 6 10-6" /></>,
  save: <><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z" /><path d="M17 21v-8H7v8M7 3v5h8" /></>,
  trash: <path d="M3 6h18M8 6V4a1 1 0 0 1 1-1h6a1 1 0 0 1 1 1v2M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6" />,
  eye: <><path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12z" /><circle cx="12" cy="12" r="3" /></>,
  palette: <><circle cx="12" cy="12" r="9" /><circle cx="8" cy="10" r="1.2" /><circle cx="12" cy="7.5" r="1.2" /><circle cx="16" cy="10" r="1.2" /></>,
  undo: <><path d="M3 7v6h6" /><path d="M3 13a9 9 0 1 0 3-7.7L3 7" /></>,
  info: <><circle cx="12" cy="12" r="9" /><path d="M12 11v5M12 7.5v.5" /></>,
  menu: <path d="M4 7h16M4 12h16M4 17h16" />,
}
return PATHS[name]
}

export function Logo() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path fill="url(#flameGrad)" d="M13 2c.4 3 3.9 5.3 4.8 8.5.3 1 .5 2 .5 3A6.3 6.3 0 0 1 12 20a6.3 6.3 0 0 1-6.3-6.5c0-2.3 1.1-4 2.5-5.8.5 1.5 1.4 2.4 1.9 2.9C10.3 7.4 11.2 4.7 13 2z" />
    </svg>
  )
}

const applePath = 'M318.7 268.7c-.2-36.7 16.4-64.4 50-84.8-18.8-26.9-47.2-41.7-84.7-44.6-35.5-2.8-74.3 20.7-88.5 20.7-15 0-49.4-19.7-76.4-19.7C63.3 141.2 4 184.8 4 273.5q0 39.3 14.4 81.2c12.8 36.7 59 126.7 107.2 125.2 25.2-.6 43-17.9 75.8-17.9 31.8 0 48.3 17.9 76.4 17.9 48.6-.7 90.4-82.5 102.6-119.3-65.2-30.7-61.7-90-61.7-91.9zm-56.6-164.2c27.3-32.4 24.8-61.9 24-72.5-24.1 1.4-52 16.4-67.9 34.9-17.5 19.8-27.8 44.3-25.6 71.9 26.1 2 49.9-11.4 69.5-34.3z'

export function AppleGlyph() {
  return <svg viewBox="0 0 384 512" fill="currentColor"><path d={applePath} /></svg>
}
export function TeslaGlyph() {
  return <svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 4.6c3.4 0 6.6-.6 9-1.8l-1.3 2.8c-1.7.8-3.6 1.2-5.6 1.3L12 22.5 9.9 6.9c-2-.1-3.9-.5-5.6-1.3L3 2.8c2.4 1.2 5.6 1.8 9 1.8z" /><path d="M12 2.4c3.9 0 7.5-.7 10.3-2l-.7 1.5C19 3 15.6 3.6 12 3.6S5 3 2.4 1.9L1.7.4C4.5 1.7 8.1 2.4 12 2.4z" /></svg>
}
export function NvGlyph() {
  return <svg viewBox="0 0 24 24"><rect width="24" height="24" rx="6" fill="#76B900" /><path d="M11.2 7.4c3.9 0 7 2.5 7 5.6 0 3.1-3.1 5.6-7 5.6-3.4 0-6.2-1.9-6.9-4.5l2.5-.6c.5 1.6 2.3 2.8 4.4 2.8 2.5 0 4.5-1.5 4.5-3.3s-2-3.3-4.5-3.3c-.5 0-1 .1-1.5.2l.5-2.4c.3-.1.7-.1 1-.1z" fill="#fff" /><circle cx="9.4" cy="13" r="1.7" fill="#fff" /></svg>
}

export function Mark({ asset }) {
  if (asset.logo === 'nv') return <span className="flogo" style={{ background: 'transparent' }}><NvGlyph /></span>
  if (asset.logo === 'apple') return <span className="flogo" style={{ background: '#2A2A33', color: '#F6EEF1' }}><AppleGlyph /></span>
  if (asset.logo === 'tesla') return <span className="flogo" style={{ background: asset.bg }}><TeslaGlyph /></span>
  return <span className="flogo" style={{ background: asset.bg }}>{asset.t.slice(0, 1)}</span>
}

export function NewsArt({ art }) {
  if (art === 'apple') return <div className="nimg art-apple"><AppleGlyph /><span className="nchip">Технологии</span></div>
  if (art === 'desk') return <DeskArt />
  if (art === 'flag') return <FlagArt />
  return <Bull />
}

export function Bull() {
  return <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M6.2 9.6C4.3 8.7 3.2 6.7 3.7 4.2c2.6.5 4.1 2 4.6 3.6h7.4c.5-1.6 2-3.1 4.6-3.6.5 2.5-.6 4.5-2.5 5.4 1 1.4 1.6 3 1.6 4.6 0 4.1-3.6 7.3-7.7 7.3s-7.7-3.2-7.7-7.3c0-1.6.6-3.2 1.6-4.6z" /><circle cx="9.3" cy="13.4" r="1.05" fill="#7A0C1E" /><circle cx="14.7" cy="13.4" r="1.05" fill="#7A0C1E" /></svg>
}

export function DeskArt() {
  return <svg viewBox="0 0 104 104" preserveAspectRatio="xMidYMid slice" style={{ position: 'absolute', inset: 0, width: '100%', height: '100%' }}><rect width="104" height="104" fill="#0E1626" /><rect x="6" y="10" width="92" height="46" rx="4" fill="#16233E" stroke="#2B4470" /><path d="M14 40l10-12 8 6 14-16" stroke="#2ECC71" strokeWidth="2" fill="none" /><rect y="94" width="104" height="10" fill="#0A101D" /></svg>
}

export function FlagArt() {
  return <svg viewBox="0 0 104 104" preserveAspectRatio="xMidYMid slice" style={{ position: 'absolute', inset: 0, width: '100%', height: '100%' }}><rect width="104" height="104" fill="#F5F0EB" /><rect width="104" height="8" y="0" fill="#D8232A" /><rect width="104" height="8" y="16" fill="#D8232A" /><rect width="104" height="8" y="32" fill="#D8232A" /><rect width="52" height="40" fill="#26355D" /></svg>
}
