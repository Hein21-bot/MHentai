<template>
  <nav class="fixed bottom-0 left-0 right-0 z-50 bg-white border-t border-gray-200 flex sm:hidden dark:bg-dark-surface dark:border-dark-border safe-bottom">
    <RouterLink :to="homePath" class="bottom-tab" :class="{ active: isHome }">
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>
      </svg>
      <span>Home</span>
    </RouterLink>

    <RouterLink :to="{ path: homePath, query: { sort: 'views' } }" class="bottom-tab" :class="{ active: $route.query.sort === 'views' && !$route.query.status }">
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 18.657A8 8 0 016.343 7.343S7 9 9 10c0-2 .5-5 2.986-7C14 5 16.09 5.777 17.656 7.343A7.975 7.975 0 0120 13a7.975 7.975 0 01-2.343 5.657z"/>
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.879 16.121A3 3 0 1012.015 11L11 14H9c0 .768.293 1.536.879 2.121z"/>
      </svg>
      <span>Hot</span>
    </RouterLink>

    <RouterLink :to="{ path: homePath, query: { status: 'ongoing' } }" class="bottom-tab" :class="{ active: $route.query.status === 'ongoing' }">
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/>
      </svg>
      <span>Ongoing</span>
    </RouterLink>

    <RouterLink :to="{ path: homePath, query: { status: 'completed' } }" class="bottom-tab" :class="{ active: $route.query.status === 'completed' }">
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z"/>
      </svg>
      <span>Done</span>
    </RouterLink>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const homePath = computed(() => `/${route.meta.lang || 'en'}`)
const isHome = computed(() => !route.query.status && !route.query.sort && !route.query.q)
</script>

<style scoped>
.bottom-tab {
  @apply flex-1 flex flex-col items-center justify-center gap-0.5 py-2 text-gray-400 transition-colors dark:text-gray-600;
}
.bottom-tab span {
  @apply text-[10px] font-medium;
}
.bottom-tab.active {
  @apply text-primary;
}
.bottom-tab svg {
  @apply transition-colors;
}
.safe-bottom {
  padding-bottom: env(safe-area-inset-bottom);
}
</style>
