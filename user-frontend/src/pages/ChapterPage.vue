<template>
  <div class="-mx-3 -mt-4 sm:-mx-4 sm:-mt-6">
    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-32 min-h-screen bg-slate-50 dark:bg-black">
      <p class="text-gray-500 animate-pulse">Loading chapter...</p>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="flex flex-col items-center justify-center py-32 text-gray-500">
      <p>Chapter not found.</p>
      <RouterLink :to="`/${route.meta.lang}`" class="text-primary hover:underline mt-2">← Home</RouterLink>
    </div>

    <!-- Reader -->
    <div v-else-if="data" class="min-h-screen bg-gray-50 dark:bg-[#111]">

      <!-- ── Header ── -->
      <div class="relative overflow-hidden bg-white dark:bg-[#0d0d0d] border-b border-gray-200 dark:border-transparent">
        <!-- Blurred cover backdrop (dark mode only) -->
        <div v-if="data.chapter.series?.cover_url"
          class="absolute inset-0 scale-110 opacity-20 blur-2xl pointer-events-none hidden dark:block">
          <img :src="data.chapter.series.cover_url" class="w-full h-full object-cover"/>
        </div>
        <div class="absolute inset-0 bg-gradient-to-b from-black/30 via-transparent to-[#0d0d0d] pointer-events-none hidden dark:block"/>

        <div class="relative max-w-3xl mx-auto px-4 pt-5 pb-7">
          <!-- Breadcrumb -->
          <nav class="flex items-center flex-wrap gap-1 text-[11px] text-gray-400 dark:text-white/30 mb-5">
            <RouterLink :to="`/${route.meta.lang}`" class="hover:text-primary dark:hover:text-white/60 transition-colors">MHentai</RouterLink>
            <span class="text-gray-300 dark:text-white/20">›</span>
            <RouterLink v-if="data.chapter.series" :to="`/${route.meta.lang}/series/${data.chapter.series.slug}`"
              class="hover:text-primary dark:hover:text-white/60 transition-colors truncate max-w-[120px] sm:max-w-[180px]">
              {{ data.chapter.series.title }}
            </RouterLink>
            <span class="text-gray-300 dark:text-white/20">›</span>
            <span class="text-gray-500 dark:text-white/50 truncate max-w-[100px] sm:max-w-[140px]">{{ data.chapter.title }}</span>
          </nav>

          <!-- Cover + info -->
          <div class="flex gap-4 sm:gap-6 items-start mb-6">
            <RouterLink v-if="data.chapter.series?.cover_url"
              :to="`/${route.meta.lang}/series/${data.chapter.series.slug}`"
              class="flex-shrink-0 w-[72px] sm:w-24 rounded-xl overflow-hidden shadow-lg ring-1 ring-gray-200 dark:shadow-2xl dark:ring-white/10">
              <img :src="data.chapter.series.cover_url" :alt="data.chapter.series.title"
                class="w-full aspect-[2/3] object-cover"/>
            </RouterLink>
            <div class="flex-1 min-w-0 pt-1">
              <RouterLink v-if="data.chapter.series"
                :to="`/${route.meta.lang}/series/${data.chapter.series.slug}`"
                class="text-xs text-primary/80 hover:text-primary transition-colors font-medium block mb-1 truncate">
                {{ data.chapter.series.title }}
              </RouterLink>
              <h1 class="text-gray-900 dark:text-white font-bold text-xl sm:text-2xl leading-tight mb-3">{{ data.chapter.title }}</h1>
              <p class="text-[11px] text-gray-400 dark:text-white/30 leading-relaxed hidden sm:block">
                Read <span class="text-gray-500 dark:text-white/50">{{ data.chapter.title }}</span> at MHentai.
                Manga <span class="text-gray-500 dark:text-white/50">{{ data.chapter.series?.title }}</span> is always updated at MHentai.
                Don't forget to read the other manga updates.
              </p>
            </div>
          </div>

          <!-- Description (mobile) -->
          <p class="text-[11px] text-gray-400 dark:text-white/30 leading-relaxed mb-5 sm:hidden">
            Read <span class="text-gray-500 dark:text-white/50">{{ data.chapter.title }}</span> at MHentai.
            Manga <span class="text-gray-500 dark:text-white/50">{{ data.chapter.series?.title }}</span> is always updated at MHentai.
            Don't forget to read the other manga updates. A list of manga collections MHentai is in the Manga List menu.
          </p>

          <!-- Controls row -->
          <div class="flex items-center gap-2 flex-wrap">
            <select v-if="seriesChapters.length" v-model="selectedChapterSlug" @change="goToChapter"
              class="flex-1 min-w-0 bg-gray-100 border border-gray-300 text-gray-900 dark:bg-white/10 dark:border-white/20 dark:text-white text-sm rounded-xl px-3 py-2.5 focus:outline-none focus:border-primary/60 cursor-pointer appearance-none">
              <option v-for="ch in seriesChapters" :key="ch.id" :value="ch.slug">
                {{ ch.title || `Chapter ${ch.number}` }}
              </option>
            </select>
            <button v-if="data.prev_chapter?.slug"
              @click="goToChapterAd(data.prev_chapter.slug)"
              class="flex items-center gap-1.5 px-4 py-2.5 rounded-xl bg-gray-100 hover:bg-gray-200 text-gray-700 hover:text-gray-900 dark:bg-white/10 dark:hover:bg-white/20 dark:text-white/80 dark:hover:text-white text-sm font-medium transition-all border border-gray-200 dark:border-white/10 flex-shrink-0">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m15 19-7-7 7-7"/></svg>
              Prev
            </button>
            <button v-if="data.next_chapter?.slug"
              @click="goToChapterAd(data.next_chapter.slug)"
              class="flex items-center gap-1.5 px-5 py-2.5 rounded-xl bg-primary hover:bg-primary-600 text-white text-sm font-bold transition-all shadow-lg shadow-primary/20 flex-shrink-0">
              Next
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m9 5 7 7-7 7"/></svg>
            </button>
          </div>
        </div>
      </div>

      <!-- ── Ad slot top ── -->
      <div class="max-w-3xl mx-auto px-4 py-3 flex justify-center">
        <AdBanner300 />
      </div>

      <!-- ── Images ── -->
      <div class="max-w-3xl mx-auto">
        <div v-for="(img, idx) in data.chapter.images" :key="idx" class="w-full relative">
          <div v-if="!imgLoaded[idx] && !imgFailed[idx]"
            class="w-full bg-gray-200 dark:bg-[#1a1a1a] animate-pulse flex flex-col items-center justify-center gap-3 min-h-[280px] sm:min-h-[480px]">
            <svg class="w-8 h-8 text-gray-400 dark:text-white/20 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
            </svg>
            <span class="text-xs text-gray-400 dark:text-white/20">Page {{ idx + 1 }}</span>
          </div>
          <img
            :src="resolveImg(img)"
            :alt="`Page ${idx + 1}`"
            loading="lazy"
            :class="['w-full h-auto block select-none transition-opacity duration-300', imgLoaded[idx] ? 'opacity-100' : 'opacity-0 absolute inset-0 h-0']"
            @load="imgLoaded[idx] = true"
            @error="onImgError($event, idx)"
          />
        </div>
        <div v-if="!data.chapter.images || data.chapter.images.length === 0"
          class="flex items-center justify-center py-32 text-gray-400 dark:text-white/30">
          <div class="text-center">
            <p class="text-lg mb-2">No images in this chapter</p>
            <p class="text-sm opacity-60">Images may not have been imported yet</p>
          </div>
        </div>
      </div>

      <!-- ── Ad slot bottom (between images and nav) ── -->
      <div class="max-w-3xl mx-auto px-4 pt-6 pb-2 flex justify-center">
        <AdNative />
      </div>

      <!-- ── Bottom nav ── -->
      <div class="max-w-3xl mx-auto px-4 py-8">
        <div class="grid grid-cols-3 gap-2">
          <button v-if="data.prev_chapter?.slug"
            @click="goToChapterAd(data.prev_chapter.slug)"
            class="flex items-center justify-center gap-1.5 px-2 py-2.5 rounded-xl bg-gray-100 hover:bg-gray-200 text-gray-600 hover:text-gray-900 dark:bg-white/10 dark:hover:bg-white/20 dark:text-white/70 dark:hover:text-white text-xs sm:text-sm font-medium transition-all border border-gray-200 dark:border-white/10">
            <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m15 19-7-7 7-7"/></svg>
            <span class="truncate">Prev</span>
          </button>
          <div v-else/>
          <RouterLink v-if="data.chapter.series"
            :to="`/${route.meta.lang}/series/${data.chapter.series.slug}`"
            class="flex items-center justify-center gap-1.5 px-2 py-2.5 rounded-xl bg-gray-100 hover:bg-gray-200 text-gray-600 hover:text-gray-900 dark:bg-white/10 dark:hover:bg-white/20 dark:text-white/70 dark:hover:text-white text-xs sm:text-sm font-medium transition-all border border-gray-200 dark:border-white/10">
            <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m4 6h16M4 10h16M4 14h16M4 18h16"/></svg>
            <span class="truncate">Chapters</span>
          </RouterLink>
          <div v-else/>
          <button v-if="data.next_chapter?.slug"
            @click="goToChapterAd(data.next_chapter.slug)"
            class="flex items-center justify-center gap-1.5 px-2 py-2.5 rounded-xl bg-primary hover:bg-primary-600 text-white text-xs sm:text-sm font-bold transition-all shadow-lg shadow-primary/20">
            <span class="truncate">Next</span>
            <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m9 5 7 7-7 7"/></svg>
          </button>
          <div v-else/>
        </div>
      </div>

      <!-- ── Related Series ── -->
      <div v-if="relatedSeries.length" class="max-w-3xl mx-auto px-4 pb-12">
        <h2 class="text-gray-500 dark:text-white/70 font-bold text-sm mb-4 pb-2 border-b border-gray-200 dark:border-white/10 uppercase tracking-wider">You May Also Like</h2>
        <div class="grid grid-cols-3 gap-2.5 sm:grid-cols-4 sm:gap-3">
          <RouterLink v-for="s in relatedSeries" :key="s.id"
            :to="`/${route.meta.lang}/series/${s.slug}`"
            class="group block">
            <div class="relative aspect-[2/3] rounded-xl overflow-hidden bg-gray-200 dark:bg-white/5 ring-1 ring-gray-200 dark:ring-white/5">
              <img v-if="s.cover_url" :src="s.cover_url" :alt="s.title"
                class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                @error="(e) => (e.target as HTMLImageElement).style.display='none'"/>
              <div class="absolute inset-0 bg-gradient-to-t from-black via-black/20 to-transparent opacity-80"/>
              <div class="absolute bottom-0 left-0 right-0 p-2">
                <p class="text-white text-[10px] font-semibold line-clamp-2 leading-snug">{{ s.title }}</p>
                <p class="text-yellow-400 text-[9px] mt-0.5">{{ starText(s) }}</p>
              </div>
            </div>
          </RouterLink>
        </div>
      </div>

    </div>

    <!-- Admin: Fix Images button (only visible when admin token is in localStorage) -->
    <div v-if="data && adminToken"
      class="fixed bottom-4 left-4 z-50">
      <button @click="fixImages"
        :disabled="fixing"
        class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium shadow-lg transition-colors"
        :class="fixResult === 'ok' ? 'bg-green-600 text-white' : fixResult === 'err' ? 'bg-red-600 text-white' : 'bg-black/80 text-white hover:bg-black'">
        <svg v-if="fixing" class="w-3 h-3 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
        </svg>
        <svg v-else class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
        </svg>
        {{ fixing ? 'Fixing...' : fixResult === 'ok' ? 'Fixed!' : fixResult === 'err' ? 'Failed' : 'Fix Images' }}
      </button>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { chapterApi, seriesApi } from '@/services/api'
import { addReadHistory } from '@/services/history'
import AdNative from '@/components/ads/AdNative.vue'
import AdBanner300 from '@/components/ads/AdBanner300.vue'
import { starText } from '@/utils/ratings'
import type { Chapter, Series } from '@/services/api'

interface ChapterResponse {
  chapter: Chapter
  prev_chapter?: Chapter
  next_chapter?: Chapter
}

const route = useRoute()
const router = useRouter()
const data = ref<ChapterResponse | null>(null)
const loading = ref(true)
const error = ref(false)
const relatedSeries = ref<Series[]>([])
const imgLoaded = ref<boolean[]>([])
const imgFailed = ref<boolean[]>([])
const seriesChapters = ref<Chapter[]>([])

const selectedChapterSlug = computed(() => route.params.slug as string)

const SMART_LINK = 'https://www.effectivecpmnetwork.com/h8ucit5btv?key=504242cfab2278b1c171b094c62a04e3'

function goToChapter(e: Event) {
  const slug = (e.target as HTMLSelectElement).value
  router.push(`/${route.meta.lang}/read/${slug}`)
}

function goToChapterAd(slug: string) {
  const url = `${window.location.origin}/${route.meta.lang}/read/${slug}`
  window.open(url, '_blank', 'noopener')
  window.location.href = SMART_LINK
}

const adminToken = localStorage.getItem('admin_token') || ''
const fixing = ref(false)
const fixResult = ref<'ok' | 'err' | null>(null)

async function fixImages() {
  if (!data.value || !adminToken) return
  fixing.value = true
  fixResult.value = null
  try {
    const sourceUrl = data.value.chapter.source_url || ''
    if (!sourceUrl) throw new Error('No source URL for this chapter')
    await fetch(`${import.meta.env.VITE_API_URL || 'http://localhost:8080/api'}/admin/import/chapter-images`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'X-Admin-Token': adminToken },
      body: JSON.stringify({ chapter_id: data.value.chapter.id, chapter_url: sourceUrl, proxy_to_r2: false }),
    }).then(r => { if (!r.ok) throw new Error('Request failed') })
    fixResult.value = 'ok'
    // Reload chapter data to show new images
    setTimeout(load, 800)
  } catch {
    fixResult.value = 'err'
  } finally {
    fixing.value = false
  }
}

const PROXIED_HOSTS = ['img.myanhwa.xyz', 'img.manhwamyanmar.com', 'img.hentai20.io', 'img.hentai1.io', 's1.manhwa18.net']
const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

function resolveImg(src: string): string {
  try {
    const host = new URL(src).hostname.toLowerCase()
    if (PROXIED_HOSTS.includes(host)) {
      return `${API_BASE}/proxy/img?url=${encodeURIComponent(src)}`
    }
  } catch {}
  return src
}

function onImgError(e: Event, idx: number) {
  const img = e.target as HTMLImageElement
  img.style.opacity = '0.3'
  img.alt = `Page ${idx + 1} (failed to load)`
  imgFailed.value[idx] = true
  imgLoaded.value[idx] = true
}

async function loadRelated(currentSeriesId?: string, genres?: string) {
  try {
    const primaryGenre = genres?.split(',')[0]?.trim() || ''
    const params: Record<string, unknown> = { sort: 'views', limit: 7, lang: route.meta.lang }
    if (primaryGenre) params.genre = primaryGenre
    const res = await seriesApi.list(params)
    relatedSeries.value = res.data.data.filter(s => s.id !== currentSeriesId).slice(0, 6)
  } catch {
    relatedSeries.value = []
  }
}

async function load() {
  loading.value = true
  error.value = false
  data.value = null
  relatedSeries.value = []
  seriesChapters.value = []
  imgLoaded.value = []
  imgFailed.value = []
  try {
    const res = await chapterApi.get(route.params.slug as string)
    data.value = res.data
    imgLoaded.value = new Array(res.data.chapter.images?.length ?? 0).fill(false)
    imgFailed.value = new Array(res.data.chapter.images?.length ?? 0).fill(false)
    if (data.value?.chapter?.series?.title) {
      document.title = `${data.value.chapter.title} - ${data.value.chapter.series.title} | MHentai`
    }
    if (data.value?.chapter?.series) {
      addReadHistory({
        seriesId: data.value.chapter.series.id,
        seriesSlug: data.value.chapter.series.slug,
        seriesTitle: data.value.chapter.series.title,
        chapterId: data.value.chapter.id,
        chapterSlug: data.value.chapter.slug,
        chapterTitle: data.value.chapter.title || `Chapter ${data.value.chapter.number ?? ''}`,
        readAt: Date.now(),
      })
    }
    loadRelated(data.value?.chapter?.series_id, data.value?.chapter?.series?.genres)
    if (data.value?.chapter?.series?.slug) {
      seriesApi.get(data.value.chapter.series.slug).then(r => {
        seriesChapters.value = r.data.chapters ?? []
      }).catch(() => {})
    }
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
}

watch(() => route.params.slug, load)
onMounted(load)
</script>

<style scoped>
.reader-btn {
  @apply flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-white/10 bg-white/10 text-white/70 text-xs font-medium hover:bg-white/20 hover:text-white transition-all;
}
</style>
