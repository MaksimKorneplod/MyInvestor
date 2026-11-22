export class ApiError extends Error {
  constructor(status, message) {
    super(message)
    this.status = status
  }
}

function messageOf(body, status) {
  if (!body) return `Сервер ответил ${status}`
  if (typeof body.error === 'string') return body.message || body.error
  if (body.error && typeof body.error === 'object') return body.error.message || body.error.code || `Сервер ответил ${status}`
  return `Сервер ответил ${status}`
}

export async function api(path, { method = 'GET', body, token } = {}) {
  const headers = { Accept: 'application/json' }
  if (body !== undefined) headers['Content-Type'] = 'application/json'
  if (token) headers.Authorization = `Bearer ${token}`
  let response
  try {
    response = await fetch(path, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    })
  } catch {
    throw new ApiError(0, 'Нет соединения с сервером')
  }
  const text = await response.text()
  let parsed = null
  if (text) {
    try { parsed = JSON.parse(text) } catch { parsed = { error: text } }
  }
  if (!response.ok) throw new ApiError(response.status, messageOf(parsed, response.status))
  return parsed
}

export const SHARES = { engine: 'stock', market: 'shares', board: 'TQBR' }

export function registerSchool(login, email, password) {
  return api('/api/v1/schools/register', { method: 'POST', body: { login, email, password } })
}

export function loginSchool(email, password) {
  return api('/api/v1/schools/login', { method: 'POST', body: { email, password } })
}

export function loginStudent(email, password) {
  return api('/api/v1/student/login', { method: 'POST', body: { email, password } })
}

export function logout(token) {
  return api('/api/v1/logout', { method: 'POST', token })
}

export function instruments({ q = '', limit = 50, offset = 0 } = {}) {
  const params = new URLSearchParams({ limit: String(limit), offset: String(offset) })
  if (q) params.set('q', q)
  return api(`/api/v1/instruments?${params}`)
}

export function quoteBatch(tickers) {
  return api('/api/v1/quotes:batch', { method: 'POST', body: { ...SHARES, tickers } })
}

export function oneQuote(ticker) {
  const q = new URLSearchParams({ board: SHARES.board })
  return api(`/api/v1/quotes/${SHARES.engine}/${SHARES.market}/${encodeURIComponent(ticker)}?${q}`)
}

export function history(ticker, period) {
  const q = new URLSearchParams({ period })
  return api(`/api/v1/instruments/${encodeURIComponent(ticker)}/history?${q}`)
}

export function news(limit = 30) {
  return api(`/api/v1/news?limit=${limit}`)
}
