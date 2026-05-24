import { defineStore } from 'pinia'
import { api } from '@/api'

interface User { id: number; username: string }

export const useAuthStore = defineStore('auth', {
  state: () => ({ token: '' as string, user: null as User | null }),
  actions: {
    loadFromStorage() {
      this.token = localStorage.getItem('token') || ''
      const raw = localStorage.getItem('user')
      this.user = raw ? JSON.parse(raw) : null
    },
    async login(username: string, password: string) {
      const resp = await api.login(username, password)
      this.token = resp.token
      this.user = resp.user
      localStorage.setItem('token', resp.token)
      localStorage.setItem('user', JSON.stringify(resp.user))
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('token')
      localStorage.removeItem('user')
    }
  }
})
