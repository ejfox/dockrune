<template>
  <span
    class="inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold"
    :class="[statusClass, large ? 'text-sm px-4 py-1.5' : 'text-xs px-3 py-1']"
  >
    <span
      v-if="status === 'in_progress'"
      class="mr-1.5 h-2 w-2 rounded-full bg-current animate-pulse"
    ></span>
    {{ statusText }}
  </span>
</template>

<script setup lang="ts">
interface Props {
  status: 'queued' | 'in_progress' | 'success' | 'failed'
  large?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  large: false
})

const statusClass = computed(() => {
  switch (props.status) {
    case 'queued':
      return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
    case 'in_progress':
      return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
    case 'success':
      return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
    case 'failed':
      return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
  }
})

const statusText = computed(() => {
  switch (props.status) {
    case 'queued':
      return 'Queued'
    case 'in_progress':
      return 'Deploying'
    case 'success':
      return 'Success'
    case 'failed':
      return 'Failed'
    default:
      return props.status
  }
})
</script>