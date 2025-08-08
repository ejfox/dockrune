<template>
  <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6 hover:shadow-lg transition-shadow">
    <div class="flex items-start justify-between mb-4">
      <div>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ deployment.Owner }}/{{ deployment.Repo }}
        </h3>
        <p class="text-sm text-gray-600 dark:text-gray-400">
          {{ deployment.Ref }} • {{ deployment.SHA.substring(0, 7) }}
        </p>
      </div>
      <DeploymentStatus :status="deployment.Status" />
    </div>

    <div class="space-y-2 text-sm">
      <div class="flex justify-between">
        <span class="text-gray-500 dark:text-gray-400">Environment:</span>
        <span class="text-gray-900 dark:text-white font-medium">{{ deployment.Environment }}</span>
      </div>
      <div class="flex justify-between">
        <span class="text-gray-500 dark:text-gray-400">Started:</span>
        <span class="text-gray-900 dark:text-white">{{ formatTime(deployment.StartedAt) }}</span>
      </div>
      <div v-if="deployment.URL" class="flex justify-between">
        <span class="text-gray-500 dark:text-gray-400">URL:</span>
        <a
          :href="deployment.URL"
          target="_blank"
          class="text-blue-600 dark:text-blue-400 hover:underline truncate max-w-[200px]"
        >
          {{ deployment.URL }}
        </a>
      </div>
    </div>

    <div v-if="showActions" class="mt-4 flex space-x-2">
      <NuxtLink
        :to="`/deployments/${deployment.ID}`"
        class="flex-1 text-center btn btn-secondary text-sm"
      >
        View Details
      </NuxtLink>
      <button
        v-if="deployment.Status === 'in_progress'"
        @click="$emit('stop', deployment.ID)"
        class="flex-1 btn btn-destructive text-sm"
      >
        Stop
      </button>
    </div>

    <!-- Progress bar for in-progress deployments -->
    <div v-if="deployment.Status === 'in_progress'" class="mt-4">
      <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
        <div
          class="bg-blue-600 h-2 rounded-full animate-pulse"
          :style="{ width: '60%' }"
        ></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.extend(relativeTime)

interface Props {
  deployment: any
  showActions?: boolean
}

withDefaults(defineProps<Props>(), {
  showActions: false
})

defineEmits<{
  stop: [id: string]
}>()

const formatTime = (timestamp: string) => {
  return dayjs(timestamp).fromNow()
}
</script>