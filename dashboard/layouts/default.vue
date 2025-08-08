<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900">
    <!-- Navigation -->
    <nav class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex">
            <div class="flex-shrink-0 flex items-center">
              <NuxtLink to="/" class="text-2xl font-bold text-blue-600 dark:text-blue-400">
                dockrune
              </NuxtLink>
            </div>
            <div class="hidden sm:ml-6 sm:flex sm:space-x-8">
              <NuxtLink
                to="/deployments"
                class="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-gray-300 dark:hover:text-white inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium"
                :class="{ 'border-blue-500 text-gray-900 dark:text-white': $route.path.startsWith('/deployments') }"
              >
                Deployments
              </NuxtLink>
              <NuxtLink
                to="/projects"
                class="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-gray-300 dark:hover:text-white inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium"
                :class="{ 'border-blue-500 text-gray-900 dark:text-white': $route.path.startsWith('/projects') }"
              >
                Projects
              </NuxtLink>
              <NuxtLink
                to="/logs"
                class="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-gray-300 dark:hover:text-white inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium"
                :class="{ 'border-blue-500 text-gray-900 dark:text-white': $route.path.startsWith('/logs') }"
              >
                Logs
              </NuxtLink>
              <NuxtLink
                to="/settings"
                class="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-gray-300 dark:hover:text-white inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium"
                :class="{ 'border-blue-500 text-gray-900 dark:text-white': $route.path.startsWith('/settings') }"
              >
                Settings
              </NuxtLink>
            </div>
          </div>
          
          <div class="flex items-center space-x-4">
            <!-- WebSocket Status -->
            <div class="flex items-center space-x-2">
              <div
                class="h-2 w-2 rounded-full"
                :class="deploymentsStore.wsConnected ? 'bg-green-500' : 'bg-red-500'"
              ></div>
              <span class="text-sm text-gray-500 dark:text-gray-400">
                {{ deploymentsStore.wsConnected ? 'Connected' : 'Disconnected' }}
              </span>
            </div>

            <!-- User Menu -->
            <div v-if="authStore.isAuthenticated" class="relative">
              <button
                @click="showUserMenu = !showUserMenu"
                class="flex items-center text-sm rounded-full focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
              >
                <div class="h-8 w-8 rounded-full bg-blue-500 flex items-center justify-center text-white">
                  {{ authStore.user?.username?.charAt(0).toUpperCase() }}
                </div>
              </button>
              
              <div
                v-if="showUserMenu"
                class="origin-top-right absolute right-0 mt-2 w-48 rounded-md shadow-lg bg-white dark:bg-gray-800 ring-1 ring-black ring-opacity-5"
              >
                <div class="py-1">
                  <div class="px-4 py-2 text-sm text-gray-700 dark:text-gray-300">
                    {{ authStore.user?.username }}
                  </div>
                  <hr class="my-1 border-gray-200 dark:border-gray-700">
                  <button
                    @click="handleLogout"
                    class="block w-full text-left px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700"
                  >
                    Logout
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </nav>

    <!-- Main Content -->
    <main>
      <slot />
    </main>

    <!-- Footer -->
    <footer class="bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 mt-auto">
      <div class="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between items-center">
          <p class="text-sm text-gray-500 dark:text-gray-400">
            © 2024 dockrune. Self-hosted deployment daemon.
          </p>
          <a
            href="https://github.com/ejfox/dockrune"
            target="_blank"
            class="text-sm text-blue-600 dark:text-blue-400 hover:underline"
          >
            GitHub
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
const authStore = useAuthStore()
const deploymentsStore = useDeploymentsStore()
const showUserMenu = ref(false)

const handleLogout = () => {
  showUserMenu.value = false
  authStore.logout()
}

// Close user menu when clicking outside
onMounted(() => {
  document.addEventListener('click', (e) => {
    const target = e.target as HTMLElement
    if (!target.closest('.relative')) {
      showUserMenu.value = false
    }
  })
})
</script>