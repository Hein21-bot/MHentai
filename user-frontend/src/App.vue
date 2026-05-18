<template>
  <div class="min-h-screen overflow-x-hidden bg-slate-50 text-gray-950 transition-colors dark:bg-dark-bg dark:text-gray-100">
    <AppHeader />
    <main class="max-w-screen-xl mx-auto px-3 py-4 sm:px-4 sm:py-6">
      <RouterView />
    </main>
    <AppFooter />
  </div>
</template>

<script setup lang="ts">
import AppHeader from '@/components/AppHeader.vue'
import AppFooter from '@/components/AppFooter.vue'
</script>
