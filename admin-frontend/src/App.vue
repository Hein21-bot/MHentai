<template>
  <div :class="[theme, 'flex h-screen bg-[#0f0f13] overflow-hidden']">
    <!-- Sidebar -->
    <aside class="w-56 flex-shrink-0 bg-[#12121a] border-r border-white/10 flex flex-col">
      <div class="px-5 py-5 border-b border-white/10">
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 bg-indigo-600 rounded-lg flex items-center justify-center font-black text-white text-sm">M</div>
          <div>
            <p class="text-white font-bold text-sm">MHentai</p>
            <p class="text-gray-600 text-xs">Admin Panel</p>
          </div>
        </div>
      </div>

      <nav class="flex-1 p-3 space-y-1">
        <RouterLink to="/dashboard" custom v-slot="{ isActive, navigate }">
          <div @click="navigate" :class="['nav-item', isActive ? 'active' : '']">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M3 7a2 2 0 012-2h4a2 2 0 012 2v4a2 2 0 01-2 2H5a2 2 0 01-2-2V7zM13 7a2 2 0 012-2h4a2 2 0 012 2v4a2 2 0 01-2 2h-4a2 2 0 01-2-2V7zM3 17a2 2 0 012-2h4a2 2 0 012 2v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4zM13 17a2 2 0 012-2h4a2 2 0 012 2v4a2 2 0 01-2 2h-4a2 2 0 01-2-2v-4z"/></svg>
            <span>Dashboard</span>
          </div>
        </RouterLink>

        <RouterLink to="/import" custom v-slot="{ isActive, navigate }">
          <div @click="navigate" :class="['nav-item', isActive ? 'active' : '']">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"/></svg>
            <span>Import</span>
          </div>
        </RouterLink>

        <RouterLink to="/upload" custom v-slot="{ isActive, navigate }">
          <div @click="navigate" :class="['nav-item', isActive ? 'active' : '']">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"/></svg>
            <span>Upload</span>
          </div>
        </RouterLink>

        <RouterLink to="/series" custom v-slot="{ isActive, navigate }">
          <div @click="navigate" :class="['nav-item', isActive ? 'active' : '']">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/></svg>
            <span>Series</span>
          </div>
        </RouterLink>

        <RouterLink to="/chapters" custom v-slot="{ isActive, navigate }">
          <div @click="navigate" :class="['nav-item', isActive ? 'active' : '']">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M7 4V2a1 1 0 011-1h8a1 1 0 011 1v2M3 4h18M5 4v16a2 2 0 002 2h10a2 2 0 002-2V4"/></svg>
            <span>Chapters</span>
          </div>
        </RouterLink>
      </nav>

      <div class="p-3 border-t border-white/10">
        <p class="text-gray-700 text-xs text-center">v1.0.0</p>
      </div>
    </aside>

    <!-- Main -->
    <div class="flex-1 flex flex-col overflow-hidden">
      <header class="bg-[#12121a] border-b border-white/10 px-6 py-3 flex items-center gap-3">
        <h1 class="text-white font-semibold text-sm">{{ pageTitle }}</h1>
        <button @click="toggleTheme" class="ml-auto inline-flex items-center gap-2 rounded-lg border border-white/10 bg-white/10 px-3 py-2 text-sm font-medium text-white transition-colors hover:bg-white/20">
          <span>{{ theme === 'dark' ? 'Switch to Light' : 'Switch to Dark' }}</span>
          <svg v-if="theme === 'dark'" class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M12 3v2M12 19v2M5.6 5.6l1.4 1.4M16.99 16.99l1.4 1.4M3 12h2M19 12h2M5.6 18.4l1.4-1.4M16.99 7.01l1.4-1.4M12 7a5 5 0 100 10 5 5 0 000-10z"/></svg>
          <svg v-else class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z"/></svg>
        </button>
      </header>
      <main class="flex-1 overflow-y-auto p-6">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const theme = ref<'dark' | 'light'>(localStorage.getItem('adminTheme') === 'light' ? 'light' : 'dark')

const pageTitle = computed(() => {
  const name = route.name?.toString() || ''
  return name.charAt(0).toUpperCase() + name.slice(1)
})

const setBodyTheme = (value: 'dark' | 'light') => {
  document.body.classList.remove('dark', 'light')
  document.body.classList.add(value)
  localStorage.setItem('adminTheme', value)
}

const toggleTheme = () => {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
}

watch(theme, (value) => setBodyTheme(value), { immediate: true })

onMounted(() => {
  setBodyTheme(theme.value)
})
</script>
