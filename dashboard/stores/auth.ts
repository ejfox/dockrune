import { defineStore } from 'pinia'

interface User {
  username: string
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: null as string | null,
    user: null as User | null,
    loading: false,
    error: null as string | null
  }),

  getters: {
    isAuthenticated: (state) => !!state.token
  },

  actions: {
    async login(username: string, password: string) {
      this.loading = true
      this.error = null

      try {
        const config = useRuntimeConfig()
        const response = await $fetch<{ token: string }>(`${config.public.apiBase}/admin/login`, {
          method: 'POST',
          body: {
            username,
            password
          }
        })

        this.token = response.token
        this.user = { username }

        // Store in localStorage
        if (process.client) {
          localStorage.setItem('dockrune_token', response.token)
          localStorage.setItem('dockrune_user', username)
        }

        // Navigate to dashboard
        await navigateTo('/deployments')
      } catch (error: any) {
        this.error = error.data?.error || 'Login failed'
        throw error
      } finally {
        this.loading = false
      }
    },

    logout() {
      this.token = null
      this.user = null
      
      if (process.client) {
        localStorage.removeItem('dockrune_token')
        localStorage.removeItem('dockrune_user')
      }

      navigateTo('/login')
    },

    async checkAuth() {
      if (process.client) {
        const token = localStorage.getItem('dockrune_token')
        const username = localStorage.getItem('dockrune_user')
        
        if (token && username) {
          this.token = token
          this.user = { username }
          return true
        }
      }
      return false
    }
  }
})