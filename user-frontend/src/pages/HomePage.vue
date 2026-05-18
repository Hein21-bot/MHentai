<template>
  <div class="space-y-6">

    <!-- Popular Today -->
    <section>
      <div class="flex items-center justify-between mb-3">
        <h2 class="section-title">Popular Today</h2>
      </div>
      <div class="overflow-x-hidden -mx-3 sm:-mx-4">
      <div class="flex gap-3 overflow-x-auto pb-2 px-3 sm:px-4" style="scrollbar-width:none;-webkit-overflow-scrolling:touch">
        <template v-if="popularLoading">
          <div v-for="i in 8" :key="i" class="flex-shrink-0 w-28 sm:w-36 md:w-44">
            <div class="aspect-[2/3] bg-gray-200 rounded-lg animate-pulse dark:bg-dark-card"/>
          </div>
        </template>
        <RouterLink v-else v-for="s in popularSeries" :key="s.id"
          :to="`/${route.meta.lang}/series/${s.slug}`"
          class="flex-shrink-0 w-28 sm:w-36 md:w-44 group">
          <div class="relative aspect-[2/3] rounded-lg overflow-hidden bg-gray-200 dark:bg-dark-card">
            <img v-if="s.cover_url" :src="s.cover_url" :alt="s.title"
              class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-200" @error="imgError"/>
            <div class="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/85 via-black/40 to-transparent pt-10 pb-2 px-2">
              <p class="text-white text-xs font-semibold line-clamp-2 leading-tight sm:text-sm">{{ s.title }}</p>
              <div class="flex items-center justify-between mt-1">
                <span class="text-yellow-400 text-[10px] sm:text-xs">{{ starText(s) }}</span>
                <span class="text-white/50 text-[10px] sm:text-xs">Ch.{{ s.chapter_count }}</span>
              </div>
            </div>
          </div>
        </RouterLink>
      </div>
      </div>
    </section>

    <!-- Ad top -->
    <div class="flex justify-center">
      <!-- <AdNative /> -->
            <AdBanner300 />
    </div>

    <!-- Latest Updates -->

    <section>
      <div class="flex items-center justify-between mb-3">
        <h2 class="section-title">Latest Updates</h2>
      </div>
      <div v-if="latestLoading" class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <div v-for="i in 12" :key="i" class="h-32 bg-gray-200 rounded animate-pulse dark:bg-dark-card" />
      </div>
      <template v-else>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div v-for="g in latestGroups" :key="g.series.id" class="group bg-white border border-gray-200 rounded overflow-hidden hover:border-primary/50 transition-all dark:bg-dark-card dark:border-dark-border">
            <div class="flex gap-3 min-h-[8rem]">
              <RouterLink :to="`/${route.meta.lang}/series/${g.series.slug}`" class="flex-shrink-0 w-20 sm:w-24 self-stretch">
                <img v-if="g.series.cover_url" :src="g.series.cover_url" :alt="g.series.title"
                  class="w-full h-full object-cover" @error="imgError"/>
                <div v-else class="w-full h-full bg-gray-100 dark:bg-gray-700 flex items-center justify-center">
                  <svg class="w-6 h-6 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/>
                  </svg>
                </div>
              </RouterLink>
              <div class="flex-1 p-2.5 flex flex-col min-w-0">
                <RouterLink :to="`/${route.meta.lang}/series/${g.series.slug}`" class="block group-hover:text-primary transition-colors mb-2 flex-shrink-0">
                  <h3 class="text-sm font-semibold text-gray-900 line-clamp-1 dark:text-gray-100">{{ g.series.title }}</h3>
                </RouterLink>
                <div class="flex-1 space-y-1 min-w-0 overflow-hidden">
                  <RouterLink v-for="c in g.chapters" :key="c.id"
                    :to="`/${route.meta.lang}/read/${c.slug}`"
                    class="flex items-center justify-between gap-2 py-0.5 hover:text-primary transition-colors text-gray-600 dark:text-gray-400 text-2xs">
                    <span class="truncate">{{ c.title || `Ch. ${c.number}` }}</span>
                    <span class="flex-shrink-0 text-gray-400 dark:text-gray-600">{{ timeAgo(c.updated_at) }}</span>
                  </RouterLink>
                </div>
                <div class="flex items-center justify-between mt-1 flex-shrink-0">
                  <div class="flex items-center gap-1">
                    <span :class="['w-2 h-2 rounded-full', g.series.status === 'ongoing' ? 'bg-green-600' : 'bg-blue-600']"></span>
                    <span class="text-2xs font-medium" :class="g.series.status === 'ongoing' ? 'text-green-600 dark:text-green-400' : 'text-blue-600 dark:text-blue-400'">
                      {{ g.series.status === 'ongoing' ? 'Ongoing' : 'Completed' }}
                    </span>
                  </div>
                  <span class="text-yellow-400 text-[10px]">{{ starText(g.series) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div v-if="latestTotalPages > 1" class="flex items-center justify-center gap-1.5 pt-4">
          <button @click="goLatestPage(latestPage - 1)" :disabled="latestPage === 1" class="pagination-btn">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m15 19-7-7 7-7"/></svg>
          </button>
          <button v-for="p in latestPageNumbers" :key="p" @click="typeof p === 'number' && goLatestPage(p)"
            :class="['pagination-btn', p === latestPage ? 'bg-primary text-white border-primary' : '', p === '…' ? 'cursor-default' : '']">
            {{ p }}
          </button>
          <button @click="goLatestPage(latestPage + 1)" :disabled="latestPage === latestTotalPages" class="pagination-btn">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m9 5 7 7-7 7"/></svg>
          </button>
        </div>
      </template>
    </section>


    <!-- Ad slot after Recommendations -->
    <div class="flex justify-center mt-4">
            <AdNative />
    </div>

    <!-- Recommendations -->
    <section>
      <div class="flex items-center justify-between mb-3">
        <h2 class="section-title">Recommendations</h2>
      </div>

      <!-- Genre tabs -->
      <div class="flex gap-1.5 mb-3">
        <button v-for="(genre, i) in recoGenres" :key="genre"
          @click="activeGenre = genre"
          :class="['px-3 py-1.5 text-xs font-semibold rounded-lg whitespace-nowrap transition-colors flex-shrink-0',
            i >= 4 ? 'hidden sm:inline-flex' : '',
            activeGenre === genre
              ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900'
              : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-card dark:text-gray-400 dark:hover:bg-dark-hover']">
          {{ genre }}
        </button>
      </div>

      <!-- Grid -->
      <div v-if="recoLoading" class="grid grid-cols-3 gap-2 sm:grid-cols-5 sm:gap-3">
        <div v-for="i in 5" :key="i" class="aspect-[2/3] bg-gray-200 rounded-lg animate-pulse dark:bg-dark-card"
          :class="i > 3 ? 'hidden sm:block' : ''"/>
      </div>
      <div v-else class="grid grid-cols-3 gap-2 sm:grid-cols-5 sm:gap-3">
        <RouterLink v-for="(s, i) in randomRecos" :key="s.id"
          :to="`/${route.meta.lang}/series/${s.slug}`"
          :class="['group block', i >= 3 ? 'hidden sm:block' : '']">
          <div class="relative aspect-[2/3] rounded-lg overflow-hidden bg-gray-200 dark:bg-dark-card">
            <img v-if="s.cover_url" :src="s.cover_url" :alt="s.title"
              class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-200" @error="imgError"/>
            <span :class="['absolute top-1 left-1 text-[9px] font-bold px-1 py-0.5 rounded leading-none',
              s.status === 'ongoing' ? 'bg-green-500 text-white' : 'bg-blue-500 text-white']">
              {{ s.status === 'ongoing' ? 'ON' : 'END' }}
            </span>
          </div>
          <div class="mt-1.5 px-0.5">
            <h3 class="text-[11px] font-semibold text-gray-900 line-clamp-2 leading-snug dark:text-gray-100 group-hover:text-primary transition-colors">{{ s.title }}</h3>
            <p class="text-[10px] text-gray-500 mt-0.5 dark:text-gray-600">Ch. {{ s.chapter_count }}</p>
            <p class="text-yellow-500 text-[10px] mt-0.5">{{ starText(s) }}</p>
          </div>
        </RouterLink>
      </div>
    </section>


  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { seriesApi } from '@/services/api'
import AdNative from '@/components/ads/AdNative.vue'
import AdBanner300 from '@/components/ads/AdBanner300.vue'
import { imgError, starText } from '@/utils/ratings'
import type { Series, Chapter } from '@/services/api'

interface LatestGroup { series: Series; chapters: Chapter[] }

const route = useRoute()
const latestGroups = ref<LatestGroup[]>([])
const latestTotal = ref(0)
const latestPage = ref(1)
const LATEST_PAGE_SIZE = 12
const popularSeries = ref<Series[]>([])
const recommendations = ref<Series[]>([])
const latestLoading = ref(false)
const popularLoading = ref(false)
const recoLoading = ref(false)
const activeGenre = ref('')

const activeStatus = computed(() => route.query.status as string || '')
const activeSort = computed(() => route.query.sort as string || '')

const latestTotalPages = computed(() => Math.ceil(latestTotal.value / LATEST_PAGE_SIZE))

const latestPageNumbers = computed(() => {
  const total = latestTotalPages.value
  const cur = latestPage.value
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

function goLatestPage(p: number) {
  latestPage.value = Math.max(1, Math.min(p, latestTotalPages.value))
  loadLatest()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

const recoGenres = ref<string[]>([])
const randomRecos = ref<Series[]>([])

function shuffle<T>(arr: T[]): T[] {
  const a = [...arr]
  for (let i = a.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [a[i], a[j]] = [a[j], a[i]]
  }
  return a
}

function pickRecos() {
  const filtered = recommendations.value.filter(s =>
    s.genres?.split(',').map(g => g.trim()).includes(activeGenre.value)
  )
  randomRecos.value = shuffle(filtered).slice(0, 5)
}

watch(activeGenre, pickRecos)

async function loadPopular() {
  popularLoading.value = true
  try {
    const params: Record<string, unknown> = { sort: activeSort.value || 'views', limit: 12, lang: route.meta.lang }
    if (activeStatus.value) params.status = activeStatus.value
    const res = await seriesApi.list(params)
    popularSeries.value = res.data.data
  } catch { popularSeries.value = [] }
  finally { popularLoading.value = false }
}

async function loadRecommendations() {
  recoLoading.value = true
  try {
    const params: Record<string, unknown> = { sort: 'title', limit: 500, lang: route.meta.lang }
    if (activeStatus.value) params.status = activeStatus.value
    const res = await seriesApi.list(params)
    recommendations.value = res.data.data
    // Build shuffled genre list
    const seen = new Set<string>()
    for (const s of recommendations.value) {
      if (s.genres) s.genres.split(',').forEach(g => { const t = g.trim(); if (t) seen.add(t) })
    }
    recoGenres.value = shuffle(Array.from(seen)).slice(0, 6)
    if (!activeGenre.value) activeGenre.value = recoGenres.value[0] ?? ''
    pickRecos()
  } catch { recommendations.value = [] }
  finally { recoLoading.value = false }
}

function timeAgo(date: string): string {
  const s = Math.floor((Date.now() - new Date(date).getTime()) / 1000)
  if (s < 3600) return `${Math.floor(s / 60)}m ago`
  if (s < 86400) return `${Math.floor(s / 3600)}h ago`
  return `${Math.floor(s / 86400)}d ago`
}

async function loadLatest() {
  latestLoading.value = true
  try {
    const params: Record<string, unknown> = { sort: 'updated_at', limit: LATEST_PAGE_SIZE, page: latestPage.value, lang: route.meta.lang }
    if (activeStatus.value) params.status = activeStatus.value
    const res = await seriesApi.list(params)
    latestTotal.value = res.data.total
    // Fetch last 3 chapters for each series in parallel
    const groups = await Promise.all(
      res.data.data.map(async s => {
        try {
          const r = await seriesApi.latestChapters(s.id)
          return { series: s, chapters: r.data.data }
        } catch {
          return { series: s, chapters: [] }
        }
      })
    )
    latestGroups.value = groups
  } catch {
    latestGroups.value = []
    latestTotal.value = 0
  } finally {
    latestLoading.value = false
  }
}

watch(() => [route.meta.lang, route.query.status, route.query.sort], () => {
  activeGenre.value = ''
  latestPage.value = 1
  loadPopular()
  loadRecommendations()
  loadLatest()
})

onMounted(() => {
  loadPopular()
  loadRecommendations()
  loadLatest()
})
</script>

<style scoped>
.pagination-btn {
  @apply min-w-[2rem] h-8 px-2 flex items-center justify-center text-xs font-semibold rounded-lg border border-gray-200 dark:border-dark-border text-gray-600 dark:text-gray-400 hover:border-primary hover:text-primary disabled:opacity-30 disabled:cursor-not-allowed transition-colors;
}
</style>
