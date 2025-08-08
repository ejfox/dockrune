<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <!-- Header -->
    <div class="mb-8">
      <h1 class="text-3xl font-bold text-gray-900 dark:text-white">Deployments</h1>
      <p class="mt-2 text-gray-600 dark:text-gray-400">
        Monitor and manage your deployments
      </p>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
      <div class="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-600 dark:text-gray-400">Total Deployments</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ deploymentsStore.stats.total }}
            </p>
          </div>
          <div class="text-3xl">📦</div>
        </div>
      </div>

      <div class="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-600 dark:text-gray-400">Success Rate</p>
            <p class="text-2xl font-bold text-green-600 dark:text-green-400">
              {{ deploymentsStore.stats.successRate.toFixed(1) }}%
            </p>
          </div>
          <div class="text-3xl">✅</div>
        </div>
      </div>

      <div class="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-600 dark:text-gray-400">Active</p>
            <p class="text-2xl font-bold text-blue-600 dark:text-blue-400">
              {{ deploymentsStore.activeDeployments.length }}
            </p>
          </div>
          <div class="text-3xl">🔄</div>
        </div>
      </div>

      <div class="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-600 dark:text-gray-400">Avg Duration</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ deploymentsStore.stats.averageDuration.toFixed(0) }}s
            </p>
          </div>
          <div class="text-3xl">⏱️</div>
        </div>
      </div>
    </div>

    <!-- Timeline Chart -->
    <div class="bg-white dark:bg-gray-800 p-6 rounded-lg shadow mb-8">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
        Deployment Timeline
      </h2>
      <DeploymentChart :data="deploymentsStore.deploymentTimeline" />
    </div>

    <!-- Active Deployments -->
    <div v-if="deploymentsStore.activeDeployments.length > 0" class="mb-8">
      <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">
        Active Deployments
      </h2>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <DeploymentCard
          v-for="deployment in deploymentsStore.activeDeployments"
          :key="deployment.ID"
          :deployment="deployment"
          :show-actions="true"
        />
      </div>
    </div>

    <!-- Recent Deployments -->
    <div>
      <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">
        Recent Deployments
      </h2>
      
      <div v-if="deploymentsStore.loading" class="text-center py-8">
        <div class="inline-flex items-center space-x-2">
          <svg class="animate-spin h-5 w-5 text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <span class="text-gray-600 dark:text-gray-400">Loading deployments...</span>
        </div>
      </div>

      <div v-else-if="deploymentsStore.error" class="bg-red-50 dark:bg-red-900/20 p-4 rounded-lg">
        <p class="text-red-600 dark:text-red-400">{{ deploymentsStore.error }}</p>
      </div>

      <div v-else-if="deploymentsStore.recentDeployments.length === 0" class="text-center py-8">
        <p class="text-gray-500 dark:text-gray-400">No deployments yet</p>
      </div>

      <div v-else class="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <table class="min-w-full">
          <thead class="bg-gray-50 dark:bg-gray-700">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                Project
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                Environment
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                Status
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                Time
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="deployment in deploymentsStore.recentDeployments" :key="deployment.ID">
              <td class="px-6 py-4 whitespace-nowrap">
                <div>
                  <div class="text-sm font-medium text-gray-900 dark:text-white">
                    {{ deployment.Owner }}/{{ deployment.Repo }}
                  </div>
                  <div class="text-sm text-gray-500 dark:text-gray-400">
                    {{ deployment.Ref }} • {{ deployment.SHA.substring(0, 7) }}
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="badge badge-info">{{ deployment.Environment }}</span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <DeploymentStatus :status="deployment.Status" />
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                {{ formatTime(deployment.StartedAt) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm">
                <div class="flex space-x-2">
                  <NuxtLink
                    :to="`/deployments/${deployment.ID}`"
                    class="text-blue-600 dark:text-blue-400 hover:underline"
                  >
                    View
                  </NuxtLink>
                  <button
                    v-if="deployment.Status === 'success' || deployment.Status === 'failed'"
                    @click="handleRedeploy(deployment.ID)"
                    class="text-green-600 dark:text-green-400 hover:underline"
                  >
                    Redeploy
                  </button>
                  <button
                    v-if="deployment.Status === 'in_progress'"
                    @click="handleStop(deployment.ID)"
                    class="text-red-600 dark:text-red-400 hover:underline"
                  >
                    Stop
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.extend(relativeTime)

const deploymentsStore = useDeploymentsStore()

// Fetch deployments on mount
onMounted(() => {
  deploymentsStore.fetchDeployments()
  
  // Set up WebSocket connection
  setupWebSocket()
  
  // Refresh every 30 seconds
  const interval = setInterval(() => {
    deploymentsStore.fetchDeployments()
  }, 30000)
  
  onUnmounted(() => {
    clearInterval(interval)
  })
})

const formatTime = (timestamp: string) => {
  return dayjs(timestamp).fromNow()
}

const handleRedeploy = async (id: string) => {
  try {
    await deploymentsStore.redeployDeployment(id)
  } catch (error) {
    console.error('Redeploy failed:', error)
  }
}

const handleStop = async (id: string) => {
  try {
    await deploymentsStore.stopDeployment(id)
  } catch (error) {
    console.error('Stop failed:', error)
  }
}

const setupWebSocket = () => {
  const config = useRuntimeConfig()
  const authStore = useAuthStore()
  
  if (!authStore.token) return
  
  const ws = new WebSocket(`${config.public.wsBase}/api/ws`)
  
  ws.onopen = () => {
    console.log('WebSocket connected')
    deploymentsStore.setWsConnected(true)
    
    // Send auth token
    ws.send(JSON.stringify({
      type: 'auth',
      token: authStore.token
    }))
  }
  
  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      if (data.type === 'deployment_update') {
        deploymentsStore.updateDeploymentFromWS(data.deployment)
      }
    } catch (error) {
      console.error('WebSocket message error:', error)
    }
  }
  
  ws.onclose = () => {
    console.log('WebSocket disconnected')
    deploymentsStore.setWsConnected(false)
    
    // Reconnect after 5 seconds
    setTimeout(setupWebSocket, 5000)
  }
  
  ws.onerror = (error) => {
    console.error('WebSocket error:', error)
  }
}
</script>