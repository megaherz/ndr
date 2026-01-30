import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export interface User {
  id: string
  telegramId: number
  telegramUsername?: string
  telegramFirstName: string
  telegramLastName?: string
}

export interface AuthTokens {
  appToken: string
  centrifugoToken: string
  expiresAt: number
}

interface AuthState {
  // State
  user: User | null
  tokens: AuthTokens | null
  isAuthenticated: boolean
  isLoading: boolean
  error: string | null

  // Actions
  setUser: (user: User) => void
  setTokens: (tokens: AuthTokens) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  logout: () => void
  clearError: () => void
  
  // Computed
  isTokenExpired: () => boolean
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      // Initial state
      user: null,
      tokens: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      // Actions
      setUser: (user) => {
        set({ 
          user, 
          isAuthenticated: true,
          error: null 
        })
      },

      setTokens: (tokens) => {
        set({ 
          tokens,
          isAuthenticated: true,
          error: null 
        })
      },

      setLoading: (loading) => {
        set({ isLoading: loading })
      },

      setError: (error) => {
        set({ error })
      },

      logout: () => {
        set({
          user: null,
          tokens: null,
          isAuthenticated: false,
          error: null,
          isLoading: false,
        })
      },

      clearError: () => {
        set({ error: null })
      },

      // Computed
      isTokenExpired: () => {
        const { tokens } = get()
        if (!tokens) return true
        return Date.now() >= tokens.expiresAt
      },
    }),
    {
      name: 'ndr-auth-storage',
      partialize: (state) => ({
        user: state.user,
        tokens: state.tokens,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
)