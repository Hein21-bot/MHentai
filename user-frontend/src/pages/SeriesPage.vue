<template>
  <div>
    <!-- Loading -->
    <div v-if="loading" class="space-y-6">
      <div class="flex flex-row gap-4 sm:gap-5">
        <div class="w-28 sm:w-48 aspect-[2/3] bg-gray-200 rounded-xl animate-pulse flex-shrink-0 dark:bg-dark-card" />
        <div class="flex-1 space-y-3 pt-2">
          <div class="h-7 bg-gray-200 rounded w-3/4 animate-pulse dark:bg-dark-card" />
          <div class="h-4 bg-gray-200 rounded w-1/3 animate-pulse dark:bg-dark-card" />
          <div class="h-4 bg-gray-200 rounded w-full animate-pulse dark:bg-dark-card" />
          <div class="h-4 bg-gray-200 rounded w-2/3 animate-pulse dark:bg-dark-card" />
        </div>
      </div>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="text-center py-20 text-gray-500 dark:text-gray-500">
      <p>Series not found.</p>
      <RouterLink to="/" class="text-primary hover:underline mt-2 inline-block">← Back to home</RouterLink>
    </div>

    <!-- Content -->
    <div v-else-if="series" class="space-y-6">
      <!-- Header -->
      <div class="flex flex-row gap-4 items-start sm:gap-5">
        <div class="w-28 sm:w-48 flex-shrink-0 rounded-xl overflow-hidden border border-gray-200 aspect-[2/3] bg-white shadow-sm dark:border-dark-border dark:bg-dark-card dark:shadow-none">
          <img v-if="series.cover_url" :src="series.cover_url" :alt="series.title" class="w-full h-full object-cover" @error="imgError"/>
        </div>
        <div class="flex-1 min-w-0 space-y-2">
          <h1 class="text-base sm:text-2xl font-extrabold text-gray-950 leading-tight dark:text-white">{{ series.title }}</h1>
          <div class="flex flex-wrap gap-2">
            <span :class="['text-xs font-bold px-2.5 py-1 rounded-full', series.status === 'ongoing' ? 'bg-green-600/20 text-green-400' : 'bg-blue-600/20 text-blue-400']">
              {{ series.status === 'ongoing' ? 'Ongoing' : 'Completed' }}
            </span>
            <span v-if="series.author" class="text-xs text-gray-500 py-1 dark:text-gray-500">By {{ series.author }}</span>
          </div>
          <div v-if="series.genres" class="flex flex-wrap gap-1">
            <span v-for="g in series.genres.split(',')" :key="g"
              class="text-xs bg-white border border-gray-200 text-gray-600 px-2 py-0.5 rounded-full dark:bg-dark-card dark:border-dark-border dark:text-gray-400">{{ g.trim() }}</span>
          </div>
          <p v-if="series.description" class="text-sm text-gray-600 leading-relaxed line-clamp-4 dark:text-gray-400">{{ series.description }}</p>
          <!-- Star rating -->
          <div class="flex items-center gap-1.5">
            <div class="flex items-center gap-0.5">
              <svg v-for="i in 5" :key="i" class="w-4 h-4" :class="i <= Math.round(getStars(series)) ? 'text-yellow-400' : 'text-gray-300 dark:text-gray-600'" viewBox="0 0 20 20" fill="currentColor">
                <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"/>
              </svg>
            </div>
            <span class="text-sm font-semibold text-gray-700 dark:text-gray-300">{{ getStars(series).toFixed(1) }}</span>
            <span class="text-xs text-gray-400 dark:text-gray-600">/ 5.0</span>
          </div>
          <div class="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-gray-500 dark:text-gray-600">
            <span>{{ series.chapters?.length ?? series.chapter_count }} chapters</span>
          </div>
          <!-- Reading buttons -->
          <div v-if="series.chapters && series.chapters.length > 0" class="flex flex-wrap gap-2 mt-1">
            <button @click="goToChapterAd(series.chapters[0].slug)"
              class="inline-flex items-center gap-2 bg-primary hover:bg-primary-600 text-white text-sm font-bold px-4 py-2.5 rounded-xl transition-colors">
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
              First Chapter
            </button>
            <button v-if="series.chapters.length > 1" @click="goToChapterAd(series.chapters[series.chapters.length - 1].slug)"
              class="inline-flex items-center gap-2 border border-primary text-primary hover:bg-primary hover:text-white text-sm font-bold px-4 py-2.5 rounded-xl transition-colors">
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
              Last Chapter
            </button>
          </div>
        </div>
      </div>

      <!-- Ad top -->
      <div class="flex justify-center">
        <AdBanner300 />
      </div>

      <!-- Latest Reading -->
      <div v-if="seriesHistory.length" class="rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-dark-border dark:bg-dark-card overflow-hidden">
        <div class="px-4 py-4 border-b border-gray-200 dark:border-dark-border">
          <p class="text-sm font-semibold text-gray-900 dark:text-white">Latest Reading</p>
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">Your recent chapters in this series.</p>
        </div>
        <div class="divide-y divide-gray-100 dark:divide-dark-border">
          <div
            v-for="item in seriesHistory"
            :key="item.chapterId"
            class="flex items-center justify-between gap-3 px-4 py-3 bg-white dark:bg-dark-card"
          >
            <div class="min-w-0">
              <p class="text-sm font-medium text-gray-800 truncate dark:text-gray-100">{{ item.chapterTitle }}</p>
              <p class="text-[11px] text-gray-500 dark:text-gray-400">{{ formatRelative(item.readAt) }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Chapter list -->
      <div>
        <h2 class="section-title mb-3">Chapters <span class="text-gray-500 font-normal text-sm dark:text-gray-600">({{ series.chapters?.length ?? 0 }})</span></h2>
        <div class="divide-y divide-gray-200 border border-gray-200 rounded-xl overflow-hidden bg-white shadow-sm dark:divide-dark-border dark:border-dark-border dark:bg-transparent dark:shadow-none max-h-[32rem] overflow-y-auto">
          <button
            v-for="chapter in series.chapters"
            :key="chapter.id"
            @click="goToChapterAd(chapter.slug)"
            class="w-full flex items-center gap-3 px-3 py-3 bg-white hover:bg-gray-50 transition-colors group dark:bg-dark-surface dark:hover:bg-dark-hover sm:px-4"
          >
            <div class="w-8 h-8 rounded-lg bg-gray-100 flex items-center justify-center text-xs font-bold text-gray-500 flex-shrink-0 dark:bg-dark-card">
              {{ chapter.number || '?' }}
            </div>
            <div class="flex-1 min-w-0 text-left">
              <p class="text-sm font-medium text-gray-800 truncate group-hover:text-primary transition-colors dark:text-gray-200">{{ chapter.title }}</p>
              <p class="text-2xs text-gray-500 dark:text-gray-600">{{ formatDate(chapter.created_at) }}</p>
            </div>
            <svg class="w-4 h-4 text-gray-400 flex-shrink-0 dark:text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" d="m9 5 7 7-7 7"/>
            </svg>
          </button>
        </div>
      </div>

      <!-- Ad slot between chapters and related series -->
      <div class="flex flex-col items-center gap-3">
        <AdNative />
      </div>

      <!-- Related Series -->
      <div v-if="relatedSeries.length">
        <h2 class="section-title mb-3">Related Series</h2>
        <div class="grid grid-cols-3 gap-2 sm:grid-cols-4 sm:gap-3 md:grid-cols-6">
          <RouterLink v-for="s in relatedSeries.slice(0, 6)" :key="s.id"
            :to="`/${route.meta.lang}/series/${s.slug}`"
            class="group block">
            <div class="relative aspect-[2/3] rounded-xl overflow-hidden bg-gray-200 dark:bg-dark-card">
              <img v-if="s.cover_url" :src="s.cover_url" :alt="s.title"
                class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-200"
                @error="imgError"/>
              <div class="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/90 to-transparent pt-6 pb-1.5 px-1.5">
                <p class="text-white text-[10px] font-semibold line-clamp-2 leading-tight">{{ s.title }}</p>
                <p class="text-yellow-400 text-[9px] mt-0.5">{{ starText(s) }}</p>
              </div>
            </div>
          </RouterLink>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { seriesApi } from '@/services/api'
import { getReadHistoryForSeries, ReadHistoryItem } from '@/services/history'
import AdNative from '@/components/ads/AdNative.vue'
import AdBanner300 from '@/components/ads/AdBanner300.vue'
import { imgError, getStars, starText } from '@/utils/ratings'
import type { Series } from '@/services/api'

const route = useRoute()
const series = ref<Series | null>(null)
const loading = ref(true)
const error = ref(false)
const relatedSeries = ref<Series[]>([])
const seriesHistory = ref<ReadHistoryItem[]>([])

function formatDate(iso: string) {
  if (!iso) return ''
  const d = new Date(iso)
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

const SMART_LINK = 'https://www.effectivecpmnetwork.com/h8ucit5btv?key=504242cfab2278b1c171b094c62a04e3'

function goToChapterAd(slug: string) {
  const url = `${window.location.origin}/${route.meta.lang}/read/${slug}`
  window.open(url, '_blank', 'noopener')
  window.location.href = SMART_LINK
}

function formatRelative(timestamp: number) {
  const delta = Math.floor((Date.now() - timestamp) / 1000)
  if (delta < 60) return 'just now'
  if (delta < 3600) return `${Math.floor(delta / 60)} min ago`
  if (delta < 86400) return `${Math.floor(delta / 3600)} hr ago`
  return `${Math.floor(delta / 86400)} day${Math.floor(delta / 86400) > 1 ? 's' : ''} ago`
}

function loadHistory() {
  seriesHistory.value = series.value ? getReadHistoryForSeries(series.value.id, 3) : []
}

async function loadRelated(currentId?: string) {
  try {
    const res = await seriesApi.list({ sort: 'views', limit: 7, lang: route.meta.lang })
    relatedSeries.value = res.data.data
      .filter(s => s.id !== currentId)
      .slice(0, 6)
  } catch {
    relatedSeries.value = []
  }
}

async function load() {
  loading.value = true
  error.value = false
  series.value = null
  relatedSeries.value = []
  try {
    const res = await seriesApi.get(route.params.slug as string)
    series.value = res.data
    loadRelated(res.data.id)
    loadHistory()
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
}

watch(() => route.params.slug, load)
onMounted(load)
</script>
