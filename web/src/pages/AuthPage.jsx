import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ApiError, loginSchool, loginStudent, registerSchool } from '../api'
import { sessionFromLogin } from '../session'
import { useApp } from '../App'
import { Icon, Logo } from '../icons'

const EMAIL = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

function passwordBytes(value) {
  return new TextEncoder().encode(value).length
}

function authMessage(err, fallback) {
  if (!(err instanceof ApiError)) return fallback
  if (err.status === 0) return err.message
  if (err.status === 401) return 'Неверный email или пароль'
  if (err.status === 409) return 'Такой email или логин уже зарегистрирован'
  if (err.status === 503) return 'Слишком много попыток. Попробуйте чуть позже'
  if (err.status === 404) return 'Этот вход на сервере ещё не включён'
  if (err.status === 400) return 'Проверьте поля и попробуйте ещё раз'
  return fallback
}

export default function AuthPage() {
  const { enter } = useApp()
  const navigate = useNavigate()
  const [schoolMode, setSchoolMode] = useState('login')
  const [pending, setPending] = useState('')
  const [schoolErr, setSchoolErr] = useState('')
  const [studentErr, setStudentErr] = useState('')
  const [school, setSchool] = useState({ login: '', email: '', password: '' })
  const [student, setStudent] = useState({ email: '', password: '' })

  function open(session) {
    if (!session?.accessToken) throw new ApiError(500, 'Сервер не вернул сессию')
    enter(session)
    navigate('/market', { replace: true })
  }

  async function onSchool(e) {
    e.preventDefault()
    setSchoolErr('')
    const email = school.email.trim().toLowerCase()
    const login = school.login.trim()
    if (schoolMode === 'register') {
      if (!login || [...login].length > 50) { setSchoolErr('Логин обязателен, до 50 символов'); return }
      if (!EMAIL.test(email) || email.length > 254) { setSchoolErr('Проверьте email'); return }
      if (passwordBytes(school.password) < 8 || passwordBytes(school.password) > 1024) { setSchoolErr('Пароль — от 8 до 1024 байт'); return }
      setPending('school')
      try {
        await registerSchool(login, email, school.password)
      } catch (err) {
        setSchoolErr(authMessage(err, 'Не удалось создать площадку'))
        setPending('')
        return
      }
      try {
        const payload = await loginSchool(email, school.password)
        open(sessionFromLogin('school', { email, login }, payload))
      } catch {
        setSchoolMode('login')
        setSchool(s => ({ ...s, password: '' }))
        setSchoolErr('Площадка создана. Войдите с этим email')
      } finally { setPending('') }
      return
    }
    if (!EMAIL.test(email)) { setSchoolErr('Проверьте email'); return }
    if (!school.password) { setSchoolErr('Введите пароль'); return }
    setPending('school')
    try {
      const payload = await loginSchool(email, school.password)
      open(sessionFromLogin('school', { email }, payload))
    } catch (err) {
      setSchoolErr(authMessage(err, 'Не удалось войти'))
    } finally { setPending('') }
  }

  async function onStudent(e) {
    e.preventDefault()
    setStudentErr('')
    const email = student.email.trim().toLowerCase()
    if (!EMAIL.test(email) || !student.password) { setStudentErr('Неверный email или пароль'); return }
    setPending('student')
    try {
      const payload = await loginStudent(email, student.password)
      open(sessionFromLogin('student', { email }, payload))
    } catch (err) {
      setStudentErr(authMessage(err, 'Не удалось войти'))
    } finally { setPending('') }
  }

  const reg = schoolMode === 'register'
  return (
    <div className="authwrap">
      <div className="authcard gate">
        <div className="brandline"><Logo /><b>MyInvestor</b></div>
        <div className="gatesplit">
          <form className="gatecol" onSubmit={onSchool}>
            <div className="role">Обучающая площадка</div>
            <h1>Школа</h1>
            <p className="asub">Площадка регистрируется сама. Ученик этой формой не создаётся.</p>
            <div className="gateswitch">
              <button type="button" className={reg ? '' : 'on'} onClick={() => { setSchoolMode('login'); setSchoolErr('') }}>Вход</button>
              <button type="button" className={reg ? 'on' : ''} onClick={() => { setSchoolMode('register'); setSchoolErr('') }}>Регистрация</button>
            </div>
            {reg && (
              <label className="afield">Логин
                <input value={school.login} maxLength={50} autoComplete="username" onChange={e => setSchool({ ...school, login: e.target.value })} />
              </label>
            )}
            <label className="afield">Email
              <input type="email" value={school.email} autoComplete="email" onChange={e => setSchool({ ...school, email: e.target.value })} />
            </label>
            <label className="afield">Пароль
              <input type="password" value={school.password} autoComplete={reg ? 'new-password' : 'current-password'} placeholder={reg ? 'Минимум 8 символов' : ''} onChange={e => setSchool({ ...school, password: e.target.value })} />
            </label>
            {schoolErr && <p className="aerr">{schoolErr}</p>}
            <button className="btn btn-brand" type="submit" disabled={pending === 'school'}>
              {reg ? <><Icon name="check" /> Создать площадку</> : <><Icon name="arrR" /> Войти в кабинет</>}
            </button>
          </form>
          <form className="gatecol" onSubmit={onStudent}>
            <div className="role">Ученик</div>
            <h1>Вход школьника</h1>
            <p className="asub">Отдельной регистрации нет. Email и пароль выдаёт школа.</p>
            <label className="afield">Email
              <input type="email" value={student.email} autoComplete="email" onChange={e => setStudent({ ...student, email: e.target.value })} />
            </label>
            <label className="afield">Пароль
              <input type="password" value={student.password} autoComplete="current-password" onChange={e => setStudent({ ...student, password: e.target.value })} />
            </label>
            {studentErr && <p className="aerr">{studentErr}</p>}
            <button className="btn btn-brand" type="submit" disabled={pending === 'student'}><Icon name="arrR" /> Войти</button>
          </form>
        </div>
      </div>
    </div>
  )
}
