import { create } from 'zustand'

const TOKEN_KEY = 'ollama_auth_token'
const REFRESH_KEY = 'ollama_refresh_token'
const USER_KEY = 'ollama_user'

export type User = { id?: string; username?: string; email?: string; verified?: boolean } | null

interface AuthState {
  token: string | null
  refreshToken: string | null
  user: User
  loading: boolean
  setAuth: (data: { token?: string; refreshToken?: string; user?: User }) => void
  clear: () => void
  loadFromStorage: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: null,
  refreshToken: null,
  user: null,
  loading: false,
  setAuth: ({ token, refreshToken, user }) => {
    if (token) localStorage.setItem(TOKEN_KEY, token)
    if (refreshToken) localStorage.setItem(REFRESH_KEY, refreshToken)
    if (user) localStorage.setItem(USER_KEY, JSON.stringify(user))
    set((s) => ({
      token: token ?? s.token,
      refreshToken: refreshToken ?? s.refreshToken,
      user: user ?? s.user,
    }))
  },
  clear: () => {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(REFRESH_KEY)
    localStorage.removeItem(USER_KEY)
    set({ token: null, refreshToken: null, user: null })
  },
  loadFromStorage: () => {
    const token = localStorage.getItem(TOKEN_KEY)
    const refreshToken = localStorage.getItem(REFRESH_KEY)
    const userStr = localStorage.getItem(USER_KEY)
    const user = userStr ? (JSON.parse(userStr) as User) : null
    set({ token, refreshToken, user })
  },
}))

