<template>
  <div class="min-h-screen overflow-x-hidden bg-slate-50 text-gray-950 transition-colors dark:bg-dark-bg dark:text-gray-100">
    <AppHeader />
    <main class="max-w-screen-xl mx-auto px-3 py-4 pb-16 sm:px-4 sm:py-6">
      <RouterView />
    </main>
    <AppFooter />
    <AdStickyBanner />
  </div>
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount } from 'vue'
import AppHeader from '@/components/AppHeader.vue'
import AppFooter from '@/components/AppFooter.vue'
import AdStickyBanner from '@/components/ads/AdStickyBanner.vue'

const healthUrl = `${import.meta.env.VITE_API_URL ?? 'http://localhost:8080/api'}/health`
let keepAliveTimer: number | undefined

const pingBackend = async () => {
  try {
    await fetch(healthUrl, { method: 'GET', cache: 'no-store' })
    console.log('Keep-alive ping sent:', healthUrl)
  } catch (err) {
    console.warn('Keep-alive ping failed:', err)
  }
}

onMounted(() => {
  pingBackend()
  keepAliveTimer = window.setInterval(pingBackend, 10 * 60 * 1000);
})

onBeforeUnmount(() => {
  if (keepAliveTimer) {
    window.clearInterval(keepAliveTimer)
  }
})
</script>
