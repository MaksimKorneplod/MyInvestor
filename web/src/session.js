const KEY = 'myinvestor.session'

export function loadSession() {
  try {
    const raw = sessionStorage.getItem(KEY)
    if (!raw) return null
    const session = JSON.parse(raw)
    if (!session?.accessToken || (session.role !== 'school' && session.role !== 'student')) return null
    if (!session.expiresAt || Date.now() >= session.expiresAt) {
      sessionStorage.removeItem(KEY)
      return null
    }
    return session
  } catch {
    return null
  }
}

export function saveSession(session) {
  sessionStorage.setItem(KEY, JSON.stringify(session))
}

export function clearSession() {
  sessionStorage.removeItem(KEY)
}

export function sessionFromLogin(role, identity, payload) {
  const expiresIn = Number(payload?.expiresIn) || 86400
  return {
    role,
    email: identity.email,
    login: identity.login || '',
    accessToken: payload.accessToken,
    tokenType: payload.tokenType || 'Bearer',
    expiresAt: Date.now() + expiresIn * 1000,
  }
}
