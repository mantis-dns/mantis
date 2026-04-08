import { create } from 'zustand'
import { apiClient } from '../api/client'

interface AuthState {
  isAuthenticated: boolean
  login: (password: string) => Promise<void>
  logout: () => Promise<void>
  checkAuth: () => Promise<void>
}

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,

  login: async (password: string) => {
    await apiClient.post('/auth/login', { password })
    set({ isAuthenticated: true })
  },

  logout: async () => {
    await apiClient.post('/auth/logout')
    set({ isAuthenticated: false })
  },

  checkAuth: async () => {
    try {
      await apiClient.get('/system/info')
      set({ isAuthenticated: true })
    } catch {
      set({ isAuthenticated: false })
    }
  },
}))
