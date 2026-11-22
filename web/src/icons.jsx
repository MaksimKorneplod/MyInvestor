export function Icon({ name, className }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      {PATHS[name]}
    </svg>
  )
}

const PATHS = {
  home: <><path d="M3 11l9-8 9 8" /><path d="M5 10v10h14V10" /></>,
  trend: <><path d="M3 17l6-6 4 4 8-8" /><path d="M15 7h6v6" /></>,
  news: <><path d="M4 5h16v14H4z" /><path d="M8 9h8M8 13h8M8 17h5" /></>,
  gear: <><circle cx="12" cy="12" r="3.2" /><path d="M12 2v3M12 19v3M2 12h3M19 12h3M4.9 4.9l2.1 2.1M17 17l2.1 2.1M19.1 4.9L17 7M7 17l-2.1 2.1" /></>,
  search: <><circle cx="11" cy="11" r="7" /><path d="M21 21l-4.3-4.3" /></>,
  arrR: <path d="M5 12h14M13 6l6 6-6 6" />,
  arrL: <path d="M19 12H5M11 6l-6 6 6 6" />,
  check: <path d="M4 12l5 5L20 6" />,
  lock: <><rect x="5" y="11" width="14" height="10" rx="2" /><path d="M8 11V7a4 4 0 0 1 8 0v4" /></>,
  refresh: <><path d="M21 12a9 9 0 1 1-2.6-6.3" /><path d="M21 3v6h-6" /></>,
  menu: <path d="M4 7h16M4 12h16M4 17h16" />,
}

export function Logo() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <defs>
        <linearGradient id="flameGrad" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0" stopColor="#FF7A5C" />
          <stop offset=".5" stopColor="#FF4D5E" />
          <stop offset="1" stopColor="#C40E2E" />
        </linearGradient>
      </defs>
      <path fill="url(#flameGrad)" d="M13 2c.4 3 3.9 5.3 4.8 8.5.3 1 .5 2 .5 3A6.3 6.3 0 0 1 12 20a6.3 6.3 0 0 1-6.3-6.5c0-2.3 1.1-4 2.5-5.8.5 1.5 1.4 2.4 1.9 2.9C10.3 7.4 11.2 4.7 13 2z" />
    </svg>
  )
}
