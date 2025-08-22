import React, { useEffect, useMemo, useRef, useState } from 'react'
import { Button } from '@ollamamax/ui'
import { AuthAPI } from '@ollamamax/api-client'
import { useAuthStore } from '../store/auth'
import { useNavigate, useSearchParams } from 'react-router-dom'

const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const passwordPolicy = {
  minLength: 8,
  upper: /[A-Z]/,
  lower: /[a-z]/,
  number: /[0-9]/,
  special: /[^A-Za-z0-9]/,
}

function PasswordStrength({ value }: { value: string }) {
  const score = useMemo(() => {
    let s = 0
    if (value.length >= passwordPolicy.minLength) s++
    if (passwordPolicy.upper.test(value)) s++
    if (passwordPolicy.lower.test(value)) s++
    if (passwordPolicy.number.test(value)) s++
    if (passwordPolicy.special.test(value)) s++
    return s
  }, [value])

  const labels = ['Very weak', 'Weak', 'Okay', 'Good', 'Strong']
  const colors = ['#ef4444', '#f97316', '#eab308', '#22c55e', '#16a34a']

  return (
    <div aria-live="polite" className="text-sm" style={{ color: colors[Math.max(0, score - 1)] }}>
      Password strength: {labels[Math.max(0, score - 1)]}
    </div>
  )
}

export function Login() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const setAuth = useAuthStore((s) => s.setAuth)
  const navigate = useNavigate()
  const emailRef = useRef<HTMLInputElement>(null)

  useEffect(() => emailRef.current?.focus(), [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      const result = await AuthAPI.login({ username: email, password })
      setAuth({ token: result.token, user: result.user })
      navigate('/v2')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="omx-v2 min-h-screen flex items-center justify-center p-6">
      <form onSubmit={handleSubmit} className="w-full max-w-sm bg-white shadow rounded p-6 space-y-4" aria-label="Login">
        <h1 className="text-xl font-semibold">Sign in</h1>

        {error && (
          <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm" role="alert">
            {error}
          </div>
        )}

        <label className="block">
          <span className="text-sm">Email</span>
          <input
            ref={emailRef}
            className="mt-1 w-full border rounded p-2"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            aria-required="true"
          />
        </label>

        <label className="block">
          <span className="text-sm">Password</span>
          <input
            className="mt-1 w-full border rounded p-2"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            aria-required="true"
          />
        </label>

        <Button type="submit" variant="primary" disabled={loading} style={{ width: '100%' }} aria-label="Sign in">
          {loading ? 'Signing in...' : 'Sign in'}
        </Button>
      </form>
    </div>
  )
}

export function Register() {
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [tos, setTos] = useState(false)
  const [privacy, setPrivacy] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  const emailValid = emailRegex.test(email)
  const passwordValid =
    password.length >= passwordPolicy.minLength &&
    passwordPolicy.upper.test(password) &&
    passwordPolicy.lower.test(password) &&
    passwordPolicy.number.test(password) &&
    passwordPolicy.special.test(password)
  const confirmValid = confirm === password

  const canSubmit = username && emailValid && passwordValid && confirmValid && tos && privacy

  const [availability, setAvailability] = useState<string | null>(null)
  useEffect(() => {
    const ctrl = new AbortController()
    setAvailability(null)
    if (!emailValid) return
    const t = setTimeout(async () => {
      try {
        // Try registration endpoint to check email availability if supported (HEAD or GET not provided; simulate via POST validate in future)
        setAvailability('ok')
      } catch (e) {
        setAvailability('taken')
      }
    }, 400)
    return () => {
      ctrl.abort()
      clearTimeout(t)
    }
  }, [emailValid, email])

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!canSubmit) return
    setLoading(true)
    setError('')
    setSuccess('')
    try {
      await AuthAPI.register({ username, email, password })
      setSuccess('Registration successful. Please check your email to verify your account.')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="omx-v2 min-h-screen flex items-center justify-center p-6">
      <form onSubmit={onSubmit} className="w-full max-w-md bg-white shadow rounded p-6 space-y-4" aria-label="Register">
        <h1 className="text-xl font-semibold">Create account</h1>
        {success && <div className="p-3 bg-green-50 border border-green-200 rounded text-green-700 text-sm" role="status">{success}</div>}
        {error && <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm" role="alert">{error}</div>}

        <label className="block">
          <span className="text-sm">Username</span>
          <input className="mt-1 w-full border rounded p-2" value={username} onChange={(e)=>setUsername(e.target.value)} required aria-required="true"/>
        </label>

        <label className="block">
          <span className="text-sm">Email</span>
          <input className="mt-1 w-full border rounded p-2" type="email" value={email} onChange={(e)=>setEmail(e.target.value)} required aria-required="true" aria-invalid={!emailValid && email ? 'true':'false'} />
          {email && !emailValid && <div className="text-xs text-red-600 mt-1" role="alert">Enter a valid email</div>}
          {availability === 'taken' && <div className="text-xs text-red-600 mt-1" role="alert">Email already in use</div>}
        </label>

        <label className="block">
          <span className="text-sm">Password</span>
          <input className="mt-1 w-full border rounded p-2" type="password" value={password} onChange={(e)=>setPassword(e.target.value)} required aria-required="true" />
          <PasswordStrength value={password} />
          <div className="text-xs text-slate-600 mt-1">Must be at least 8 characters and include uppercase, lowercase, number, and special character.</div>
        </label>

        <label className="block">
          <span className="text-sm">Confirm password</span>
          <input className="mt-1 w-full border rounded p-2" type="password" value={confirm} onChange={(e)=>setConfirm(e.target.value)} required aria-required="true" aria-invalid={!confirmValid && confirm ? 'true':'false'} />
          {confirm && !confirmValid && <div className="text-xs text-red-600 mt-1" role="alert">Passwords do not match</div>}
        </label>

        <div className="space-y-2">
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={tos} onChange={(e)=>setTos(e.target.checked)} aria-checked={tos} />
            I agree to the <a className="text-blue-600 underline" href="#" onClick={(e)=>{e.preventDefault(); alert('Terms of Service content...')}}>Terms of Service</a>
          </label>
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={privacy} onChange={(e)=>setPrivacy(e.target.checked)} aria-checked={privacy} />
            I agree to the <a className="text-blue-600 underline" href="#" onClick={(e)=>{e.preventDefault(); alert('Privacy Policy content...')}}>Privacy Policy</a>
          </label>
        </div>

        <Button type="submit" variant="primary" disabled={!canSubmit || loading} aria-label="Create account">{loading ? 'Creating...' : 'Create account'}</Button>
      </form>
    </div>
  )
}

export function ForgotPassword() {
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const [count, setCount] = useState(0)

  const emailValid = emailRegex.test(email)
  const canSubmit = emailValid && count < 3

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!canSubmit) return
    setLoading(true)
    setError('')
    try {
      await AuthAPI.forgotPassword({ email })
      setMessage('If an account exists for this email, you will receive a password reset link.')
      setCount((c) => c + 1)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Request failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="omx-v2 min-h-screen flex items-center justify-center p-6">
      <form onSubmit={onSubmit} className="w-full max-w-sm bg-white shadow rounded p-6 space-y-4" aria-label="Forgot Password">
        <h1 className="text-xl font-semibold">Forgot password</h1>
        {message && <div className="p-3 bg-green-50 border border-green-200 rounded text-green-700 text-sm" role="status">{message}</div>}
        {error && <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm" role="alert">{error}</div>}
        <label className="block">
          <span className="text-sm">Email</span>
          <input className="mt-1 w-full border rounded p-2" type="email" value={email} onChange={(e)=>setEmail(e.target.value)} required aria-required="true" />
        </label>
        <Button type="submit" variant="primary" disabled={!canSubmit || loading} aria-label="Send reset link">{loading? 'Sending...' : 'Send reset link'}</Button>
        <div className="text-xs text-slate-600">You can request up to 3 reset emails per hour.</div>
      </form>
    </div>
  )
}

export function ResetPassword() {
  const [search] = useSearchParams()
  const token = search.get('token') || ''
  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  const passwordValid =
    password.length >= passwordPolicy.minLength &&
    passwordPolicy.upper.test(password) &&
    passwordPolicy.lower.test(password) &&
    passwordPolicy.number.test(password) &&
    passwordPolicy.special.test(password)
  const confirmValid = confirm === password

  const canSubmit = token && passwordValid && confirmValid

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!canSubmit) return
    setLoading(true)
    setError('')
    try {
      await AuthAPI.resetPassword({ token, password })
      setMessage('Your password has been reset. You can now log in.')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Reset failed (token may be invalid or expired).')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="omx-v2 min-h-screen flex items-center justify-center p-6">
      <form onSubmit={onSubmit} className="w-full max-w-sm bg-white shadow rounded p-6 space-y-4" aria-label="Reset Password">
        <h1 className="text-xl font-semibold">Reset password</h1>
        {message && <div className="p-3 bg-green-50 border border-green-200 rounded text-green-700 text-sm" role="status">{message}</div>}
        {error && <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm" role="alert">{error}</div>}
        <div className="text-xs text-slate-600">Token: {token ? token.slice(0,6)+'...' : 'Missing token'}</div>
        <label className="block">
          <span className="text-sm">New password</span>
          <input className="mt-1 w-full border rounded p-2" type="password" value={password} onChange={(e)=>setPassword(e.target.value)} required aria-required="true" />
          <PasswordStrength value={password} />
        </label>
        <label className="block">
          <span className="text-sm">Confirm password</span>
          <input className="mt-1 w-full border rounded p-2" type="password" value={confirm} onChange={(e)=>setConfirm(e.target.value)} required aria-required="true" aria-invalid={!confirmValid && confirm ? 'true':'false'} />
          {confirm && !confirmValid && <div className="text-xs text-red-600 mt-1" role="alert">Passwords do not match</div>}
        </label>
        <Button type="submit" variant="primary" disabled={!canSubmit || loading} aria-label="Reset password">{loading ? 'Resetting...' : 'Reset password'}</Button>
      </form>
    </div>
  )
}

export function VerifyEmail() {
  const [search] = useSearchParams()
  const token = search.get('token') || ''
  const [status, setStatus] = useState<'pending'|'success'|'error'>('pending')
  const [message, setMessage] = useState('Verifying your email...')

  useEffect(() => {
    (async () => {
      try {
        await AuthAPI.verifyEmail({ token })
        setStatus('success')
        setMessage('Email verified successfully! You can now use all features.')
      } catch (e) {
        setStatus('error')
        setMessage('Verification failed or token expired.')
      }
    })()
  }, [token])

  return (
    <div className="omx-v2 min-h-screen flex items-center justify-center p-6">
      <div className="w-full max-w-md bg-white shadow rounded p-6 space-y-4" role="status" aria-live="polite">
        <h1 className="text-xl font-semibold">Verify Email</h1>
        <div className={status==='success'? 'text-green-700' : status==='error'? 'text-red-700':'text-slate-700'}>{message}</div>
      </div>
    </div>
  )
}

