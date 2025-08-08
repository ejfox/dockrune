import { defineStore } from 'pinia'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.extend(relativeTime)

interface Deployment {
  ID: string
  Owner: string
  Repo: string
  Ref: string
  SHA: string
  Environment: string
  Status: 'queued' | 'in_progress' | 'success' | 'failed'
  StartedAt: string
  CompletedAt?: string
  URL?: string
  Port?: number
  ProjectType?: string
  Error?: string
}

interface DeploymentStats {
  total: number
  successful: number
  failed: number
  averageDuration: number
  successRate: number
}

export const useDeploymentsStore = defineStore('deployments', {
  state: () => ({
    deployments: [] as Deployment[],
    loading: false,
    error: null as string | null,
    selectedDeployment: null as Deployment | null,
    stats: {
      total: 0,
      successful: 0,
      failed: 0,
      averageDuration: 0,
      successRate: 0
    } as DeploymentStats,
    wsConnected: false
  }),

  getters: {
    activeDeployments: (state) => {
      return state.deployments.filter(d => 
        d.Status === 'in_progress' || d.Status === 'queued'
      )
    },

    recentDeployments: (state) => {
      return state.deployments
        .sort((a, b) => new Date(b.StartedAt).getTime() - new Date(a.StartedAt).getTime())
        .slice(0, 10)
    },

    deploymentsByEnvironment: (state) => {
      const grouped = {} as Record<string, Deployment[]>
      state.deployments.forEach(d => {
        if (!grouped[d.Environment]) {
          grouped[d.Environment] = []
        }
        grouped[d.Environment].push(d)
      })
      return grouped
    },

    deploymentTimeline: (state) => {
      const last7Days = Array.from({ length: 7 }, (_, i) => {
        const date = dayjs().subtract(i, 'day').format('YYYY-MM-DD')
        return {
          date,
          successful: 0,
          failed: 0
        }
      }).reverse()

      state.deployments.forEach(d => {
        const date = dayjs(d.StartedAt).format('YYYY-MM-DD')
        const dayData = last7Days.find(day => day.date === date)
        if (dayData) {
          if (d.Status === 'success') {
            dayData.successful++
          } else if (d.Status === 'failed') {
            dayData.failed++
          }
        }
      })

      return last7Days
    }
  },

  actions: {
    async fetchDeployments() {
      this.loading = true
      this.error = null
      
      try {
        const config = useRuntimeConfig()
        const response = await $fetch<Deployment[]>(`${config.public.apiBase}/api/deployments`, {
          headers: {
            'Authorization': `Bearer ${useAuthStore().token}`
          }
        })
        
        this.deployments = response
        this.calculateStats()
      } catch (error) {
        this.error = 'Failed to fetch deployments'
        console.error('Fetch deployments error:', error)
      } finally {
        this.loading = false
      }
    },

    async fetchDeployment(id: string) {
      try {
        const config = useRuntimeConfig()
        const response = await $fetch<Deployment>(`${config.public.apiBase}/api/deployments/${id}`, {
          headers: {
            'Authorization': `Bearer ${useAuthStore().token}`
          }
        })
        
        this.selectedDeployment = response
        
        // Update in list if exists
        const index = this.deployments.findIndex(d => d.ID === id)
        if (index >= 0) {
          this.deployments[index] = response
        }
      } catch (error) {
        console.error('Fetch deployment error:', error)
        throw error
      }
    },

    async redeployDeployment(id: string) {
      try {
        const config = useRuntimeConfig()
        const response = await $fetch(`${config.public.apiBase}/api/deployments/${id}/redeploy`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${useAuthStore().token}`
          }
        })
        
        await this.fetchDeployments()
        return response
      } catch (error) {
        console.error('Redeploy error:', error)
        throw error
      }
    },

    async stopDeployment(id: string) {
      try {
        const config = useRuntimeConfig()
        await $fetch(`${config.public.apiBase}/api/deployments/${id}/stop`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${useAuthStore().token}`
          }
        })
        
        await this.fetchDeployments()
      } catch (error) {
        console.error('Stop deployment error:', error)
        throw error
      }
    },

    async fetchLogs(id: string): Promise<string> {
      try {
        const config = useRuntimeConfig()
        const response = await $fetch<string>(`${config.public.apiBase}/api/deployments/${id}/logs`, {
          headers: {
            'Authorization': `Bearer ${useAuthStore().token}`
          },
          responseType: 'text'
        })
        
        return response
      } catch (error) {
        console.error('Fetch logs error:', error)
        throw error
      }
    },

    updateDeploymentFromWS(deployment: Deployment) {
      const index = this.deployments.findIndex(d => d.ID === deployment.ID)
      if (index >= 0) {
        this.deployments[index] = deployment
      } else {
        this.deployments.push(deployment)
      }
      
      // Update selected if it's the same
      if (this.selectedDeployment?.ID === deployment.ID) {
        this.selectedDeployment = deployment
      }
      
      this.calculateStats()
    },

    calculateStats() {
      const completed = this.deployments.filter(d => 
        d.Status === 'success' || d.Status === 'failed'
      )
      
      this.stats.total = this.deployments.length
      this.stats.successful = this.deployments.filter(d => d.Status === 'success').length
      this.stats.failed = this.deployments.filter(d => d.Status === 'failed').length
      this.stats.successRate = this.stats.total > 0 
        ? (this.stats.successful / this.stats.total) * 100 
        : 0
      
      // Calculate average duration
      const durations = completed
        .filter(d => d.CompletedAt)
        .map(d => {
          const start = new Date(d.StartedAt).getTime()
          const end = new Date(d.CompletedAt!).getTime()
          return end - start
        })
      
      this.stats.averageDuration = durations.length > 0
        ? durations.reduce((a, b) => a + b, 0) / durations.length / 1000 // in seconds
        : 0
    },

    setWsConnected(connected: boolean) {
      this.wsConnected = connected
    }
  }
})