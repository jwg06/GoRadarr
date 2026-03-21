import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'

interface AuthState {
  token: string | null
  username: string | null
  isAuthenticated: boolean
  login: (token: string, username: string) => void
  logout: () => void
}

// Use a lazy storage getter so window.localStorage is accessed at call-time,
// not captured once at store-creation time (avoids issues in test environments).
const lazyStorage = createJSONStorage(() => ({
  getItem: (key: string) => window.localStorage.getItem(key),
  setItem: (key: string, value: string) => window.localStorage.setItem(key, value),
  removeItem: (key: string) => window.localStorage.removeItem(key),
}))

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      username: null,
      isAuthenticated: false,
      login: (token, username) => set({ token, username, isAuthenticated: true }),
      logout: () => set({ token: null, username: null, isAuthenticated: false }),
    }),
    {
      name: 'goradarr-auth',
      storage: lazyStorage,
    },
  ),
)
