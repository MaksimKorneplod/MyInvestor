import { createContext, useContext, useEffect, useMemo, useState } from 'react'
import { Navigate, Route, Routes, useNavigate } from 'react-router-dom'
import { logout } from './api'
import { clearSession, loadSession, saveSession } from './session'
import AuthPage from './pages/AuthPage'
import Shell from './shell/Shell'
import MarketPage from './pages/MarketPage'
import InstrumentPage from './pages/InstrumentPage'
import NewsPage from './pages/NewsPage'
import SettingsPage from './pages/SettingsPage'

const AppContext = createContext(null)
export function useApp() { return useContext(AppContext) }

export default function App() {
  const [session, setSession] = useState(() => loadSession())
  const [palette, setPalette] = useState(() => localStorage.getItem('myinvestor.palette') === 'deutan' ? 'deutan' : 'classic')

  useEffect(() => {
    document.body.dataset.palette = palette
    localStorage.setItem('myinvestor.palette', palette)
  }, [palette])

  const api = useMemo(() => ({
    session,
    palette,
    setPalette,
    enter(next) {
      saveSession(next)
      setSession(next)
    },
    async leave() {
      const current = loadSession()
      if (current?.accessToken) {
        try { await logout(current.accessToken) } catch { /* маршрут выхода может быть ещё не подключён */ }
      }
      clearSession()
      setSession(null)
    },
  }), [session, palette])

  return (
    <AppContext.Provider value={api}>
      <Routes>
        <Route path="/login" element={session ? <Navigate to="/market" replace /> : <AuthPage />} />
        <Route element={session ? <Shell /> : <Navigate to="/login" replace />}>
          <Route path="/market" element={<MarketPage />} />
          <Route path="/instrument/:ticker" element={<InstrumentPage />} />
          <Route path="/news" element={<NewsPage />} />
          <Route path="/settings" element={<SettingsPage />} />
        </Route>
        <Route path="*" element={<Navigate to={session ? '/market' : '/login'} replace />} />
      </Routes>
    </AppContext.Provider>
  )
}

export function useLeave() {
  const { leave } = useApp()
  const navigate = useNavigate()
  return async () => {
    await leave()
    navigate('/login', { replace: true })
  }
}
