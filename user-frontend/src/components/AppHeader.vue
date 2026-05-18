<template>
  <!-- Drawer overlay -->
  <Transition name="fade">
    <div v-if="menuOpen" class="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm" @click="menuOpen = false"/>
  </Transition>

  <!-- Drawer panel -->
  <Transition name="slide">
    <div v-if="menuOpen" class="fixed top-0 left-0 z-50 h-full w-64 bg-white dark:bg-[#1a1a24] shadow-2xl flex flex-col">
      <!-- Drawer header -->
      <div class="flex items-center justify-between px-4 py-3 border-b border-gray-200 dark:border-white/10">
        <RouterLink :to="homePath" @click="menuOpen = false" class="flex items-center gap-2">
          <div class="w-7 h-7 bg-primary rounded-lg flex items-center justify-center font-black text-white text-sm">M</div>
          <span class="font-extrabold text-gray-950 dark:text-white text-sm">MHentai</span>
        </RouterLink>
        <button @click="menuOpen = false" class="w-8 h-8 flex items-center justify-center rounded-lg text-gray-500 hover:bg-gray-100 dark:hover:bg-white/10 transition-colors">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
        </button>
      </div>

      <!-- Nav links -->
      <nav class="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
        <RouterLink :to="{ path: homePath }" @click="menuOpen = false"
          class="drawer-link" :class="{ 'drawer-link-active': !$route.query.status && !$route.query.q && $route.name?.toString().startsWith('home') }">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/></svg>
          Home
        </RouterLink>
        <RouterLink :to="`/${$route.meta.lang}/manga-list`" @click="menuOpen = false"
          class="drawer-link" :class="{ 'drawer-link-active': $route.name?.toString().startsWith('manga-list') }">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16"/></svg>
          Manga Lists
        </RouterLink>
        <RouterLink :to="{ path: homePath, query: { status: 'ongoing' } }" @click="menuOpen = false"
          class="drawer-link" :class="{ 'drawer-link-active': $route.query.status === 'ongoing' }">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
          Ongoing
        </RouterLink>
        <RouterLink :to="{ path: homePath, query: { status: 'completed' } }" @click="menuOpen = false"
          class="drawer-link" :class="{ 'drawer-link-active': $route.query.status === 'completed' }">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z"/></svg>
          Completed
        </RouterLink>
        <RouterLink :to="`/${$route.meta.lang}/az`" @click="menuOpen = false"
          class="drawer-link" :class="{ 'drawer-link-active': $route.name?.toString().startsWith('az') }">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h10M4 14h6M4 18h4"/></svg>
          A — Z List
        </RouterLink>
        <RouterLink :to="`/${$route.meta.lang}/genres`" @click="menuOpen = false"
          class="drawer-link" :class="{ 'drawer-link-active': $route.name?.toString().startsWith('genres') }">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A2 2 0 013 12V7a4 4 0 014-4z"/></svg>
          Genres
        </RouterLink>

        <div class="pt-3 mt-3 border-t border-gray-200 dark:border-white/10 space-y-1">
          <p class="px-3 text-[10px] font-semibold text-gray-400 uppercase tracking-wider mb-2">Type</p>
          <RouterLink to="/my" @click="menuOpen = false"
            class="drawer-link" :class="{ 'drawer-link-active': $route.meta.lang === 'my' }">
            🇲🇲 မြန်မာ
          </RouterLink>
          <RouterLink to="/en" @click="menuOpen = false"
            class="drawer-link" :class="{ 'drawer-link-active': $route.meta.lang === 'en' }">
            🇬🇧 English
          </RouterLink>
        </div>
      </nav>

      <!-- Theme toggle at bottom -->
      <div class="px-4 py-4 border-t border-gray-200 dark:border-white/10">
        <button @click="toggleTheme" class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-white/10 transition-colors text-sm font-medium">
          <svg v-if="isDark" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="4"/><path d="M12 2v2m0 16v2m10-10h-2M4 12H2m17.07-7.07-1.41 1.41M6.34 17.66l-1.41 1.41m14.14 0-1.41-1.41M6.34 6.34 4.93 4.93"/>
          </svg>
          <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path d="M20.35 15.35A8.5 8.5 0 0 1 8.65 3.65 8.5 8.5 0 1 0 20.35 15.35Z"/>
          </svg>
          {{ isDark ? 'Light Mode' : 'Dark Mode' }}
        </button>
      </div>
    </div>
  </Transition>

  <header class="sticky top-0 z-40 border-b border-gray-200 bg-white/95 backdrop-blur dark:border-dark-border dark:bg-dark-surface/95">
    <div class="max-w-screen-xl mx-auto px-3 sm:px-4">

      <!-- Main row -->
      <div class="flex items-center gap-2 h-12">

        <!-- Hamburger -->
        <button @click="menuOpen = true" class="w-8 h-8 flex items-center justify-center rounded-lg text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-dark-card transition-colors flex-shrink-0">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"/></svg>
        </button>

        <!-- Logo -->
        <RouterLink :to="homePath" class="flex items-center gap-1.5 flex-shrink-0">
          <div class="w-7 h-7 bg-primary rounded-lg flex items-center justify-center font-black text-white text-sm">M</div>
          <span class="font-extrabold text-gray-950 text-sm dark:text-white hidden sm:inline">MHentai</span>
        </RouterLink>

        <!-- Search -->
        <div class="flex-1 relative">
          <svg class="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400 dark:text-gray-500 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
          </svg>
          <input
            v-model="searchQuery"
            @keydown.enter="doSearch"
            type="text"
            placeholder="Search..."
            class="w-full bg-gray-100 border border-transparent text-gray-900 placeholder-gray-400 text-xs pl-8 pr-7 py-2 rounded-lg focus:outline-none focus:border-primary/50 focus:bg-white transition-colors dark:bg-dark-card dark:text-gray-100 dark:placeholder-gray-500 dark:focus:bg-dark-surface"
          />
          <button v-if="searchQuery" type="button"
            class="absolute right-2 top-1/2 -translate-y-1/2 h-4 w-4 inline-flex items-center justify-center text-gray-400 hover:text-gray-700 dark:hover:text-gray-200"
            @click="clearSearch">
            <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M6 6l12 12M18 6 6 18"/></svg>
          </button>
        </div>

        <!-- Theme -->
        <button type="button" class="theme-btn" @click="toggleTheme">
          <svg v-if="isDark" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="4"/><path d="M12 2v2m0 16v2m10-10h-2M4 12H2m17.07-7.07-1.41 1.41M6.34 17.66l-1.41 1.41m14.14 0-1.41-1.41M6.34 6.34 4.93 4.93"/>
          </svg>
          <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path d="M20.35 15.35A8.5 8.5 0 0 1 8.65 3.65 8.5 8.5 0 1 0 20.35 15.35Z"/>
          </svg>
        </button>

      </div>

    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const router = useRouter()
const route = useRoute()
const searchQuery = ref('')
const isDark = ref(document.documentElement.classList.contains('dark'))
const menuOpen = ref(false)
const homePath = computed(() => `/${route.meta.lang || 'my'}`)

watch(() => route.path, () => { menuOpen.value = false })

function applyTheme(dark: boolean) {
  isDark.value = dark
  document.documentElement.classList.toggle('dark', dark)
  localStorage.setItem('theme', dark ? 'dark' : 'light')
}

function toggleTheme() {
  applyTheme(!isDark.value)
}

function doSearch() {
  const q = searchQuery.value.trim()
  if (!q) return
  searchQuery.value = q
  router.push({ path: homePath.value, query: { q } })
}

function clearSearch() {
  searchQuery.value = ''
  if (route.query.q) {
    router.push({ path: homePath.value, query: { ...route.query, q: undefined } })
  }
}

onMounted(() => {
  const saved = localStorage.getItem('theme')
  const prefersDark = window.matchMedia?.('(prefers-color-scheme: dark)').matches
  applyTheme(saved ? saved === 'dark' : prefersDark)
  searchQuery.value = (route.query.q as string) || ''
})

watch(() => route.query.q, (q) => {
  searchQuery.value = (q as string) || ''
})
</script>

<style scoped>
.theme-btn {
  @apply h-8 w-8 inline-flex items-center justify-center rounded-lg border border-gray-200 text-gray-600 transition-colors hover:border-primary/60 hover:text-primary dark:border-dark-border dark:text-gray-400 dark:hover:text-primary flex-shrink-0;
}
.drawer-link {
  @apply flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium text-gray-600 hover:text-gray-950 hover:bg-gray-100 transition-colors dark:text-gray-400 dark:hover:text-white dark:hover:bg-white/10;
}
.drawer-link-active {
  @apply text-primary bg-primary/10 dark:text-primary dark:bg-primary/10;
}

/* Drawer slide */
.slide-enter-active, .slide-leave-active { transition: transform 0.25s ease; }
.slide-enter-from, .slide-leave-to { transform: translateX(-100%); }

/* Overlay fade */
.fade-enter-active, .fade-leave-active { transition: opacity 0.25s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
