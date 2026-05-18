<template>
  <div class="space-y-5">
    <h1 class="text-xl font-extrabold text-gray-950 dark:text-white">Manga Lists</h1>

    <!-- Filter panel -->
    <div class="bg-white dark:bg-dark-card border border-gray-200 dark:border-dark-border rounded-2xl p-4 space-y-3">
      <!-- Dropdowns row 1 -->
      <div class="grid grid-cols-2 gap-2">
        <select v-model="filterGenre" class="filter-select">
          <option value="">Genre All</option>
          <option v-for="g in genreOptions" :key="g" :value="g">{{ g }}</option>
        </select>
        <select v-model="filterStatus" class="filter-select">
          <option value="">Status All</option>
          <option value="ongoing">Ongoing</option>
          <option value="completed">Completed</option>
        </select>
      </div>
      <!-- Dropdowns row 2 -->
      <div class="grid grid-cols-2 gap-2">
        <select v-model="filterSort" class="filter-select">
          <option value="">Order by Default</option>
          <option value="views">Most Views</option>
          <option value="updated_at">Latest Update</option>
          <option value="title">A — Z</option>
        </select>
        <!-- placeholder for future type filter -->
        <select v-model="filterType" class="filter-select">
          <option value="">Type All</option>
          <option value="manhwa">Manhwa</option>
          <option value="manga">Manga</option>
          <option value="manhua">Manhua</option>
        </select>
      </div>

      <!-- Search button -->
      <button @click="doSearch"
        class="w-full flex items-center justify-center gap-2 py-2.5 bg-primary hover:bg-primary-600 text-white text-sm font-bold rounded-xl transition-colors">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
        </svg>
        Search
      </button>

      <!-- Text mode toggle + input -->
      <div>
        <button @click="textMode = !textMode"
          class="text-xs font-semibold text-gray-500 dark:text-gray-400 hover:text-primary transition-colors flex items-center gap-1.5">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536M9 13l6.586-6.586a2 2 0 112.828 2.828L11.828 15.828a2 2 0 01-1.414.586H9v-2a2 2 0 01.586-1.414z"/></svg>
          Text Mode
        </button>
        <div v-if="textMode" class="mt-2 flex gap-2">
          <input v-model="filterText" @keydown.enter="doSearch" type="text" placeholder="Search title..."
            class="flex-1 bg-gray-100 dark:bg-dark-surface border border-gray-200 dark:border-dark-border text-gray-900 dark:text-gray-100 text-sm rounded-xl px-3 py-2 focus:outline-none focus:border-primary/60 transition-colors placeholder-gray-400"/>
          <button @click="doSearch"
            class="px-4 py-2 bg-primary hover:bg-primary-600 text-white text-sm font-bold rounded-xl transition-colors">
            Search
          </button>
        </div>
      </div>
    </div>

    <!-- Results count -->
    <p v-if="!loading" class="text-sm text-gray-500 dark:text-gray-500">
      <span class="font-semibold text-gray-900 dark:text-white">{{ results.length }}</span> series found
      <span v-if="totalPages > 1"> · Page {{ currentPage }} of {{ totalPages }}</span>
    </p>

    <!-- Loading -->
    <div v-if="loading" class="grid grid-cols-3 gap-2 sm:grid-cols-4 sm:gap-3 md:grid-cols-5">
      <div v-for="i in 24" :key="i" class="aspect-[2/3] bg-gray-200 rounded-xl animate-pulse dark:bg-dark-card"/>
    </div>

    <!-- Empty -->
    <div v-else-if="results.length === 0"
      class="py-20 text-center text-gray-400 dark:text-gray-600 text-sm">
      No series found. Try different filters.
    </div>

    <template v-else>
      <!-- Grid -->
      <div class="grid grid-cols-3 gap-2 sm:grid-cols-4 sm:gap-3 md:grid-cols-5">
        <RouterLink v-for="s in paginated" :key="s.id"
          :to="`/${route.meta.lang}/series/${s.slug}`"
          class="group block">
          <div class="relative aspect-[2/3] rounded-xl overflow-hidden bg-gray-200 dark:bg-dark-card">
            <img v-if="s.cover_url" :src="s.cover_url" :alt="s.title"
              class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
              @error="imgError"/>
            <div class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/10 to-transparent"/>
            <span :class="['absolute top-1.5 left-1.5 text-[9px] font-bold px-1.5 py-0.5 rounded leading-none',
              s.status === 'ongoing' ? 'bg-green-500 text-white' : 'bg-blue-500 text-white']">
              {{ s.status === 'ongoing' ? 'Ongoing' : 'Completed' }}
            </span>
            <div class="absolute bottom-0 left-0 right-0 p-2">
              <p class="text-white text-[10px] font-semibold line-clamp-2 leading-snug">{{ s.title }}</p>
              <p class="text-white/50 text-[9px] mt-0.5">Chapter {{ s.chapter_count }}</p>
            </div>
          </div>
        </RouterLink>
      </div>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="flex items-center justify-center gap-1.5 pt-2">
        <button @click="goPage(currentPage - 1)" :disabled="currentPage === 1" class="pagination-btn">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m15 19-7-7 7-7"/></svg>
        </button>
        <button v-for="p in pageNumbers" :key="p" @click="typeof p === 'number' && goPage(p)"
          :class="['pagination-btn', p === currentPage ? 'bg-primary text-white border-primary' : '', p === '…' ? 'cursor-default' : '']">
          {{ p }}
        </button>
        <button @click="goPage(currentPage + 1)" :disabled="currentPage === totalPages" class="pagination-btn">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m9 5 7 7-7 7"/></svg>
        </button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { seriesApi } from '@/services/api'
import { imgError } from '@/utils/ratings'
import type { Series } from '@/services/api'

const route = useRoute()

const filterGenre = ref('')
const filterStatus = ref('')
const filterSort = ref('')
const filterType = ref('')
const filterText = ref('')
const textMode = ref(false)
const results = ref<Series[]>([])
const genreOptions = ref<string[]>([])
const loading = ref(false)
const currentPage = ref(1)
const PAGE_SIZE = 24

const totalPages = computed(() => Math.ceil(results.value.length / PAGE_SIZE))
const paginated = computed(() => {
  const start = (currentPage.value - 1) * PAGE_SIZE
  return results.value.slice(start, start + PAGE_SIZE)
})
const pageNumbers = computed(() => {
  const total = totalPages.value
  const cur = currentPage.value
  const pages: (number | '…')[] = []
  if (total <= 7) {
    for (let i = 1; i <= total; i++) pages.push(i)
  } else {
    pages.push(1)
    if (cur > 3) pages.push('…')
    for (let i = Math.max(2, cur - 1); i <= Math.min(total - 1, cur + 1); i++) pages.push(i)
    if (cur < total - 2) pages.push('…')
    pages.push(total)
  }
  return pages
})

function goPage(p: number) {
  currentPage.value = Math.max(1, Math.min(p, totalPages.value))
  window.scrollTo({ top: 0, behavior: 'smooth' })
}



async function doSearch() {
  loading.value = true
  currentPage.value = 1
  try {
    const params: Record<string, unknown> = {
      limit: 200,
      lang: route.meta.lang,
    }
    if (filterStatus.value) params.status = filterStatus.value
    if (filterSort.value) params.sort = filterSort.value
    if (filterText.value.trim()) params.q = filterText.value.trim()

    const res = await seriesApi.list(params)
    let data = res.data.data

    // client-side genre filter
    if (filterGenre.value) {
      data = data.filter(s =>
        s.genres?.split(',').map(g => g.trim()).includes(filterGenre.value)
      )
    }
    results.value = data
  } catch {
    results.value = []
  } finally {
    loading.value = false
  }
}

async function loadGenres() {
  try {
    const res = await seriesApi.list({ sort: 'title', limit: 500, lang: route.meta.lang })
    const set = new Set<string>()
    for (const s of res.data.data) {
      if (s.genres) s.genres.split(',').forEach(g => { const t = g.trim(); if (t) set.add(t) })
    }
    genreOptions.value = Array.from(set).sort()
  } catch {}
}

async function loadAll() {
  await loadGenres()
  await doSearch()
}

watch(() => route.meta.lang, () => {
  filterGenre.value = ''
  filterStatus.value = ''
  filterSort.value = ''
  filterText.value = ''
  loadAll()
})

onMounted(loadAll)
</script>

<style scoped>
.filter-select {
  @apply w-full bg-gray-100 dark:bg-dark-surface border border-gray-200 dark:border-dark-border text-gray-700 dark:text-gray-300 text-sm rounded-xl px-3 py-2.5 focus:outline-none focus:border-primary/60 transition-colors cursor-pointer;
}
.pagination-btn {
  @apply min-w-[2rem] h-8 px-2 flex items-center justify-center text-xs font-semibold rounded-lg border border-gray-200 dark:border-dark-border text-gray-600 dark:text-gray-400 hover:border-primary hover:text-primary disabled:opacity-30 disabled:cursor-not-allowed transition-colors;
}
</style>
