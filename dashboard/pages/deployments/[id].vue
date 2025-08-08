<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <!-- Back button -->
    <NuxtLink
      to="/deployments"
      class="inline-flex items-center text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 mb-4"
    >
      ← Back to deployments
    </NuxtLink>

    <div v-if="loading" class="text-center py-12">
      <div class="inline-flex items-center space-x-2">
        <svg class="animate-spin h-8 w-8 text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <span class="text-gray-600 dark:text-gray-400">Loading deployment...</span>
      </div>
    </div>

    <div v-else-if="error" class="bg-red-50 dark:bg-red-900/20 p-6 rounded-lg">
      <p class="text-red-600 dark:text-red-400">{{ error }}</p>
    </div>

    <div v-else-if="deployment">
      <!-- Header -->
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6 mb-6">
        <div class="flex items-start justify-between">
          <div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">
              {{ deployment.Owner }}/{{ deployment.Repo }}
            </h1>
            <div class="flex items-center space-x-4 text-sm text-gray-600 dark:text-gray-400">
              <span>{{ deployment.Ref }}</span>
              <span>•</span>
              <span>{{ deployment.SHA.substring(0, 7) }}</span>
              <span>•</span>
              <span>{{ deployment.Environment }}</span>
            </div>
          </div>
          <DeploymentStatus :status="deployment.Status" :large="true" />
        </div>

        <div class="mt-6 grid grid-cols-1 md:grid-cols-4 gap-4">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">Started</p>
            <p class="text-sm font-medium text-gray-900 dark:text-white">
              {{ formatDateTime(deployment.StartedAt) }}
            </p>
          </div>
          <div v-if="deployment.CompletedAt">
            <p class="text-sm text-gray-500 dark:text-gray-400">Completed</p>
            <p class="text-sm font-medium text-gray-900 dark:text-white">
              {{ formatDateTime(deployment.CompletedAt) }}
            </p>
          </div>
          <div v-if="deployment.CompletedAt">
            <p class="text-sm text-gray-500 dark:text-gray-400">Duration</p>
            <p class="text-sm font-medium text-gray-900 dark:text-white">
              {{ calculateDuration(deployment.StartedAt, deployment.CompletedAt) }}
            </p>
          </div>
          <div v-if="deployment.URL">
            <p class="text-sm text-gray-500 dark:text-gray-400">URL</p>
            <a
              :href="deployment.URL"
              target="_blank"
              class="text-sm font-medium text-blue-600 dark:text-blue-400 hover:underline"
            >
              {{ deployment.URL }}
            </a>
          </div>
        </div>

        <!-- Actions -->
        <div class="mt-6 flex space-x-3">
          <button
            v-if="deployment.Status === 'success' || deployment.Status === 'failed'"
            @click="handleRedeploy"
            class="btn btn-primary"
          >
            🔄 Redeploy
          </button>
          <button
            v-if="deployment.Status === 'in_progress'"
            @click="handleStop"
            class="btn btn-destructive"
          >
            ⏹ Stop Deployment
          </button>
          <button
            @click="showLogs = true"
            class="btn btn-secondary"
          >
            📋 View Logs
          </button>
        </div>
      </div>

      <!-- Error Details -->
      <div v-if="deployment.Error" class="bg-red-50 dark:bg-red-900/20 rounded-lg p-6 mb-6">
        <h2 class="text-lg font-semibold text-red-900 dark:text-red-100 mb-2">
          Deployment Error
        </h2>
        <pre class="text-sm text-red-800 dark:text-red-200 whitespace-pre-wrap">{{ deployment.Error }}</pre>
      </div>

      <!-- Deployment Info -->
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
          Deployment Information
        </h2>
        <dl class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <dt class="text-sm text-gray-500 dark:text-gray-400">Deployment ID</dt>
            <dd class="text-sm font-mono text-gray-900 dark:text-white">{{ deployment.ID }}</dd>
          </div>
          <div>
            <dt class="text-sm text-gray-500 dark:text-gray-400">Project Type</dt>
            <dd class="text-sm text-gray-900 dark:text-white">{{ deployment.ProjectType || 'Unknown' }}</dd>
          </div>
          <div>
            <dt class="text-sm text-gray-500 dark:text-gray-400">Port</dt>
            <dd class="text-sm text-gray-900 dark:text-white">{{ deployment.Port || 'N/A' }}</dd>
          </div>
          <div>
            <dt class="text-sm text-gray-500 dark:text-gray-400">Clone URL</dt>
            <dd class="text-sm font-mono text-gray-900 dark:text-white truncate">{{ deployment.CloneURL }}</dd>
          </div>
        </dl>
      </div>
    </div>

    <!-- Logs Modal -->
    <LogsModal
      v-if="showLogs && deployment"
      :deployment-id="deployment.ID"
      @close="showLogs = false"
    />
  </div>
</template>

<script setup lang="ts">
import dayjs from 'dayjs'

const route = useRoute()
const deploymentsStore = useDeploymentsStore()

const deployment = computed(() => deploymentsStore.selectedDeployment)
const loading = ref(true)
const error = ref<string | null>(null)
const showLogs = ref(false)

onMounted(async () => {
  try {
    await deploymentsStore.fetchDeployment(route.params.id as string)
  } catch (err: any) {
    error.value = err.message || 'Failed to load deployment'
  } finally {
    loading.value = false
  }
})

const formatDateTime = (timestamp: string) => {
  return dayjs(timestamp).format('MMM D, YYYY h:mm A')
}

const calculateDuration = (start: string, end: string) => {
  const duration = dayjs(end).diff(dayjs(start), 'second')
  if (duration < 60) {
    return `${duration} seconds`
  } else if (duration < 3600) {
    return `${Math.floor(duration / 60)} minutes ${duration % 60} seconds`
  } else {
    return `${Math.floor(duration / 3600)} hours ${Math.floor((duration % 3600) / 60)} minutes`
  }
}

const handleRedeploy = async () => {
  if (!deployment.value) return
  
  try {
    await deploymentsStore.redeployDeployment(deployment.value.ID)
    navigateTo('/deployments')
  } catch (err: any) {
    error.value = err.message || 'Failed to redeploy'
  }
}

const handleStop = async () => {
  if (!deployment.value) return
  
  try {
    await deploymentsStore.stopDeployment(deployment.value.ID)
    await deploymentsStore.fetchDeployment(deployment.value.ID)
  } catch (err: any) {
    error.value = err.message || 'Failed to stop deployment'
  }
}
</script>