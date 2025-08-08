<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>

<script setup lang="ts">
// Check auth on app mount
onMounted(() => {
  const authStore = useAuthStore()
  authStore.checkAuth()
})
</script>