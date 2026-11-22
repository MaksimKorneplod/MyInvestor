import { useApp, useLeave } from '../App'

export default function SettingsPage() {
  const { palette, setPalette, session } = useApp()
  const leave = useLeave()
  return (
    <>
      <div className="pagehead">
        <h1>Настройки</h1>
        <span className="sub">{session.role === 'school' ? 'Обучающая площадка' : 'Ученик'} · {session.email}</span>
      </div>
      <div className="card setcard">
        <div>
          <b>Палитра</b>
          <p>Классика: зелёный рост и красный спад. Второй режим различается при дальтонизме.</p>
        </div>
        <select value={palette} aria-label="Палитра" onChange={e => setPalette(e.target.value)}>
          <option value="classic">Классическая</option>
          <option value="deutan">Синий / оранжевый</option>
        </select>
      </div>
      <div className="card setcard">
        <div>
          <b>Сессия</b>
          <p>Выход закрывает сессию в этом браузере. Токен школы живёт 24 часа.</p>
        </div>
        <button className="btn btn-down btn-sm" type="button" onClick={leave}>Выйти</button>
      </div>
    </>
  )
}
