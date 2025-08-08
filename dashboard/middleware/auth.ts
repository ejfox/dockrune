export default defineNuxtRouteMiddleware((to, from) => {
  const authStore = useAuthStore()
  
  // Pages that don't require auth
  const publicPages = ['/login', '/']
  
  if (!publicPages.includes(to.path) && !authStore.isAuthenticated) {
    return navigateTo('/login')
  }
})