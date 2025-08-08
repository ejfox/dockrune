<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800">
    <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
      <h1 class="text-6xl font-bold text-gray-900 dark:text-white mb-4">
        dockrune
      </h1>
      <p class="text-xl text-gray-600 dark:text-gray-400 mb-8">
        Self-hosted deployment daemon for GitHub webhooks
      </p>
      
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-12">
        <div class="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-lg">
          <div class="text-3xl mb-3">🚀</div>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
            Push to Deploy
          </h3>
          <p class="text-gray-600 dark:text-gray-400">
            Automatic deployments triggered by GitHub webhooks
          </p>
        </div>
        
        <div class="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-lg">
          <div class="text-3xl mb-3">🔍</div>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
            Smart Detection
          </h3>
          <p class="text-gray-600 dark:text-gray-400">
            Automatically detects project type and deployment strategy
          </p>
        </div>
        
        <div class="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-lg">
          <div class="text-3xl mb-3">📊</div>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
            Real-time Status
          </h3>
          <p class="text-gray-600 dark:text-gray-400">
            GitHub status updates and Discord notifications
          </p>
        </div>
      </div>
      
      <div class="flex justify-center space-x-4">
        <NuxtLink
          to="/deployments"
          class="btn btn-primary"
        >
          View Deployments
        </NuxtLink>
        <a
          href="https://github.com/ejfox/dockrune"
          target="_blank"
          class="btn btn-secondary"
        >
          Documentation
        </a>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const authStore = useAuthStore()

// Redirect to deployments if already logged in
onMounted(() => {
  if (authStore.isAuthenticated) {
    navigateTo('/deployments')
  }
})
</script>