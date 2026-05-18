<template>
  <div class="space-y-5">
    <h1 class="text-xl font-extrabold text-gray-950 dark:text-white">AZ Lists</h1>

    <!-- Letter filter bar -->
    <div class="flex flex-wrap gap-1.5">
      <button @click="setLetter('')"
        :class="['min-w-[2.25rem] h-9 px-3 text-xs font-bold rounded-lg transition-colors',
          activeLetter === '' ? 'bg-primary text-white shadow-sm' : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-card dark:text-gray-400 dark:hover:bg-dark-hover']">
        All
      </button>
      <button v-for="l in letters" :key="l" @click="setLetter(l)"
        :class="['min-w-[2.25rem] h-9 px-2 text-xs font-bold rounded-lg transition-colors',
          activeLetter === l ? 'bg-primary text-white shadow-sm' : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-card dark:text-gray-400 dark:hover:bg-dark-hover']">
        {{ l }}
      </button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="grid grid-cols-3 gap-2 sm:grid-cols-4 sm:gap-3 md:grid-cols-5">
      <div v-for="i in 24" :key="i" class="aspect-[2/3] bg-gray-200 rounded-xl animate-pulse dark:bg-dark-card"/>
    </div>

    <!-- Empty -->
    <div v-else-if="filtered.length === 0" class="py-20 text-center text-gray-400 dark:text-gray-600">
      <p class="text-4xl font-black mb-2 opacity-20">{{ activeLetter || 'All' }}</p>
      <p class="text-sm">No series found</p>
    </div>

    <template v-else>
      <!-- Count -->
      <p class="text-sm text-gray-500 dark:text-gray-500">
        <span class="font-semibold text-gray-900 dark:text-white">{{ filtered.length }}</span> series
        <span v-if="totalPages > 1"> · Page {{ currentPage }} of {{ totalPages }}</span>
      </p>

      <!-- Grid -->
      <div class="grid grid-cols-3 gap-2 sm:grid-cols-4 sm:gap-3 md:grid-cols-5">
        <RouterLink v-for="s in paginated" :key="s.id"
          :to="`/${route.meta.lang}/series/${s.slug}`" class="group block">
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
        <button @click="goPage(currentPage - 1)" :disabled="currentPage === 1"
          class="pagination-btn">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m15 19-7-7 7-7"/></svg>
        </button>
        <button v-for="p in pageNumbers" :key="p" @click="typeof p === 'number' && goPage(p)"
          :class="['pagination-btn', p === currentPage ? 'bg-primary text-white border-primary' : '', p === '…' ? 'cursor-default' : '']">
          {{ p }}
        </button>
        <button @click="goPage(currentPage + 1)" :disabled="currentPage === totalPages"
          class="pagination-btn">
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
import type { Series } from '@/services/api'

const PAGE_SIZE = 24

const route = useRoute()
const allSeries = ref<Series[]>([])
const loading = ref(true)
const activeLetter = ref('')
const currentPage = ref(1)

const letters = ['#', '0-9', 'A','B','C','D','E','F','G','H','I','J','K','L','M','N','O','P','Q','R','S','T','U','V','W','X','Y','Z']

const filtered = computed(() => {
  if (activeLetter.value === '') return allSeries.value
  return allSeries.value.filter(s => {
    const first = s.title?.[0]?.toUpperCase() ?? ''
    if (activeLetter.value === '#') return !/[A-Z0-9]/.test(first)
    if (activeLetter.value === '0-9') return /[0-9]/.test(first)
    return first === activeLetter.value
  })
})

const totalPages = computed(() => Math.ceil(filtered.value.length / PAGE_SIZE))

const paginated = computed(() => {
  const start = (currentPage.value - 1) * PAGE_SIZE
  return filtered.value.slice(start, start + PAGE_SIZE)
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

function setLetter(l: string) {
  activeLetter.value = l
  currentPage.value = 1
}

function goPage(p: number) {
  currentPage.value = Math.max(1, Math.min(p, totalPages.value))
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function imgError(e: Event) { (e.target as HTMLImageElement).style.display = 'none' }

async function load() {
  loading.value = true
  try {
    const res = await seriesApi.list({ sort: 'title', limit: 500, lang: route.meta.lang })
    allSeries.value = res.data.data
  } catch {
    allSeries.value = []
  } finally {
    loading.value = false
  }
}

watch(() => route.meta.lang, () => { activeLetter.value = ''; currentPage.value = 1; load() })
onMounted(load)
</script>

<style scoped>
.pagination-btn {
  @apply min-w-[2rem] h-8 px-2 flex items-center justify-center text-xs font-semibold rounded-lg border border-gray-200 dark:border-dark-border text-gray-600 dark:text-gray-400 hover:border-primary hover:text-primary disabled:opacity-30 disabled:cursor-not-allowed transition-colors;
}
</style>
