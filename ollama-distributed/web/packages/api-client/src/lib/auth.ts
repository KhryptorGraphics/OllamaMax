import { http } from './http'

export interface LoginReq { username: string; password: string }
export interface LoginRes { token: string; user: any; session_id?: string; expires_at?: string }
export interface RegisterReq { username: string; email: string; password: string }
export interface ForgotPasswordReq { email: string }
export interface ResetPasswordReq { token: string; password: string }
export interface VerifyEmailReq { token: string }

async function tryAuth<T>(primary: string, fallback: string | null, init: RequestInit) {
  try {
    return await http<T>(primary, init as any)
  } catch (e: any) {
    if (fallback) {
      return await http<T>(fallback, init as any)
    }
    throw e
  }
}

export const AuthAPI = {
  login: (payload: LoginReq) =>
    tryAuth<LoginRes>(
      '/v1/auth/login',
      '/v1/login',
      { method: 'POST', body: JSON.stringify(payload) },
    ),
  register: (payload: RegisterReq) =>
    tryAuth<any>(
      '/v1/auth/register',
      '/v1/register',
      { method: 'POST', body: JSON.stringify(payload) },
    ),
  logout: () =>
    tryAuth<{ message: string }>(
      '/v1/auth/logout',
      '/v1/user/logout',
      { method: 'POST' },
    ),
  profile: () =>
    tryAuth<{ user: any }>(
      '/v1/profile',
      '/v1/user/profile',
      { method: 'GET' },
    ),
  forgotPassword: (payload: ForgotPasswordReq) =>
    tryAuth<{ message: string }>(
      '/v1/auth/forgot-password',
      null,
      { method: 'POST', body: JSON.stringify(payload) },
    ),
  resetPassword: (payload: ResetPasswordReq) =>
    tryAuth<{ message: string }>(
      '/v1/auth/reset-password',
      null,
      { method: 'POST', body: JSON.stringify(payload) },
    ),
  verifyEmail: (payload: VerifyEmailReq) =>
    tryAuth<{ message: string }>(
      '/v1/auth/verify-email',
      null,
      { method: 'POST', body: JSON.stringify(payload) },
    ),
}

