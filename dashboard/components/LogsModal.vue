<template>
  <div class="fixed inset-0 z-50 overflow-y-auto">
    <div class="flex min-h-screen items-center justify-center p-4">
      <!-- Backdrop -->
      <div
        class="fixed inset-0 bg-black bg-opacity-50 transition-opacity"
        @click="$emit('close')"
      ></div>

      <!-- Modal -->
      <div class="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-4xl w-full max-h-[80vh] flex flex-col">
        <!-- Header -->
        <div class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            Deployment Logs
          </h3>
          <button
            @click="$emit('close')"
            class="text-gray-400 hover:text-gray-500 dark:hover:text-gray-300"
          >
            <svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Controls -->
        <div class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
          <div class="flex space-x-2">
            <button
              @click="toggleAutoScroll"
              class="btn btn-secondary text-sm"
              :class="{ 'bg-blue-600 text-white': autoScroll }"
            >
              {{ autoScroll ? '⏸ Pause' : '▶ Auto-scroll' }}
            </button>
            <button
              @click="refreshLogs"
              class="btn btn-secondary text-sm"
              :disabled="loading"
            >
              🔄 Refresh
            </button>
          </div>
          <div class="text-sm text-gray-500 dark:text-gray-400">
            {{ logs ? logs.split('\n').length : 0 }} lines
          </div>
        </div>

        <!-- Logs Content -->
        <div
          ref="logsContainer"
          class="flex-1 overflow-y-auto p-4 bg-gray-900 text-gray-100"
        >
          <div v-if="loading" class="text-center py-8">
            <div class="inline-flex items-center space-x-2">
              <svg class="animate-spin h-5 w-5 text-blue-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <span>Loading logs...</span>
            </div>
          </div>

          <div v-else-if="error" class="text-red-400">
            {{ error }}
          </div>

          <pre v-else class="font-mono text-xs leading-relaxed whitespace-pre-wrap">{{ logs || 'No logs available' }}</pre>
        </div>

        <!-- Footer -->
        <div class="p-4 border-t border-gray-200 dark:border-gray-700">
          <button
            @click="$emit('close')"
            class="btn btn-primary"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  deploymentId: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  close: []
}>()

const deploymentsStore = useDeploymentsStore()

const logs = ref<string>('')
const loading = ref(true)
const error = ref<string | null>(null)
const autoScroll = ref(true)
const logsContainer = ref<HTMLElement>()

const fetchLogs = async () => {
  loading.value = true
  error.value = null
  
  try {
    logs.value = await deploymentsStore.fetchLogs(props.deploymentId)
    
    // Auto scroll to bottom
    if (autoScroll.value && logsContainer.value) {
      nextTick(() => {
        if (logsContainer.value) {
          logsContainer.value.scrollTop = logsContainer.value.scrollHeight
        }
      })
    }
  } catch (err: any) {
    error.value = err.message || 'Failed to fetch logs'
  } finally {
    loading.value = false
  }
}

const refreshLogs = () => {
  fetchLogs()
}

const toggleAutoScroll = () => {
  autoScroll.value = !autoScroll.value
  if (autoScroll.value && logsContainer.value) {
    logsContainer.value.scrollTop = logsContainer.value.scrollHeight
  }
}

onMounted(() => {
  fetchLogs()
  
  // Auto-refresh logs every 5 seconds if deployment is in progress
  const interval = setInterval(() => {
    const deployment = deploymentsStore.deployments.find(d => d.ID === props.deploymentId)
    if (deployment?.Status === 'in_progress') {
      fetchLogs()
    }
  }, 5000)
  
  onUnmounted(() => {
    clearInterval(interval)
  })
})

// Close on Escape key
onMounted(() => {
  const handleEscape = (e: KeyboardEvent) => {
    if (e.key === 'Escape') {
      emit('close')
    }
  }
  document.addEventListener('keydown', handleEscape)
  onUnmounted(() => {
    document.removeEventListener('keydown', handleEscape)
  })
})
</script>