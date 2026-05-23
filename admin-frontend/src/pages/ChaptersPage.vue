<template>
  <div class="flex gap-4 h-full" style="min-height: calc(100vh - 8rem)">

    <!-- Left panel: Series list -->
    <div class="w-72 flex-shrink-0 flex flex-col gap-3">
      <input v-model="seriesSearch" type="text" placeholder="Search series..." class="form-input"/>

      <div class="bg-[#1a1a24] rounded-xl border border-white/10 overflow-y-auto flex-1">
        <div v-if="seriesLoading" class="p-4 text-gray-500 text-sm">Loading...</div>
        <div v-else-if="filteredSeries.length === 0" class="p-4 text-gray-600 text-sm">No series found</div>
        <button
          v-for="s in paginatedSeries"
          :key="s.id"
          @click="selectSeries(s)"
          :class="[
            'w-full flex items-center gap-3 px-4 py-3 text-left border-b border-white/5 transition-colors last:border-b-0',
            selectedSeries?.id === s.id ? 'bg-indigo-600/20 border-l-2 border-l-indigo-500' : 'hover:bg-white/5'
          ]"
        >
          <img v-if="s.cover_url" :src="s.cover_url" :alt="s.title"
            class="w-8 h-11 object-cover rounded flex-shrink-0" @error="imgError"/>
          <div v-else class="w-8 h-11 bg-[#12121a] rounded flex-shrink-0"/>
          <div class="min-w-0 flex-1">
            <p class="text-sm text-white truncate font-medium">{{ s.title }}</p>
            <p class="text-xs text-gray-600">{{ s.chapter_count }} chapters</p>
          </div>
        </button>
      </div>

      <!-- Series pagination -->
      <div v-if="seriesTotalPages > 1" class="flex items-center justify-between px-1">
        <button @click="seriesPage--" :disabled="seriesPage === 1"
          class="px-2 py-1 text-xs text-gray-400 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed">‹ Prev</button>
        <span class="text-xs text-gray-500">{{ seriesPage }} / {{ seriesTotalPages }}</span>
        <button @click="seriesPage++" :disabled="seriesPage === seriesTotalPages"
          class="px-2 py-1 text-xs text-gray-400 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed">Next ›</button>
      </div>
    </div>

    <!-- Right panel: Chapters of selected series -->
    <div class="flex-1 flex flex-col gap-3 min-w-0">

      <!-- No series selected -->
      <div v-if="!selectedSeries" class="flex-1 flex items-center justify-center text-gray-600">
        <div class="text-center">
          <svg class="w-12 h-12 mx-auto mb-3 opacity-30" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path d="M15 19l-7-7 7-7"/>
          </svg>
          <p>Select a series to view its chapters</p>
        </div>
      </div>

      <template v-else>
        <!-- Header -->
        <div class="flex items-center justify-between gap-3 flex-wrap">
          <div>
            <h2 class="text-white font-bold text-lg">{{ selectedSeries.title }}</h2>
            <p class="text-gray-500 text-xs mt-0.5">{{ selectedSeries.chapter_count }} chapters · page {{ chapterPage }}</p>
          </div>
          <div class="flex items-center gap-2">
            <button @click="doDedup" :disabled="deduping"
              class="text-xs px-3 py-1.5 bg-red-600/20 text-red-400 hover:bg-red-600/30 rounded-lg transition-colors disabled:opacity-50 whitespace-nowrap">
              {{ deduping ? 'Removing...' : 'Remove Duplicates' }}
            </button>
            <button @click="doScrapeAllImages" :disabled="bulkScraping"
              class="text-xs px-3 py-1.5 bg-indigo-600/20 text-indigo-400 hover:bg-indigo-600/30 rounded-lg transition-colors disabled:opacity-50 whitespace-nowrap">
              {{ bulkScraping ? `Scraping ${bulkProgress}/${bulkTotal}...` : 'Get All Images' }}
            </button>
            <input v-model="chapterSearch" type="text" placeholder="Search chapters..." class="form-input w-48"/>
          </div>
        </div>
        <!-- Bulk scrape progress bar -->
        <div v-if="bulkScraping" class="w-full bg-white/5 rounded-full h-1.5">
          <div class="bg-indigo-500 h-1.5 rounded-full transition-all" :style="{ width: bulkTotal > 0 ? (bulkProgress / bulkTotal * 100) + '%' : '0%' }"/>
        </div>

        <!-- Loading chapters -->
        <div v-if="chaptersLoading" class="text-gray-500 text-sm">Loading chapters...</div>

        <!-- Chapter table -->
        <div v-else class="bg-[#1a1a24] rounded-xl border border-white/10 overflow-hidden overflow-y-auto flex-1">
          <table class="w-full">
            <thead>
              <tr class="border-b border-white/10">
                <th class="text-left text-xs font-medium text-gray-500 px-4 py-3">Chapter</th>
                <th class="text-left text-xs font-medium text-gray-500 px-4 py-3">Images</th>
                <th class="text-left text-xs font-medium text-gray-500 px-4 py-3">Views</th>
                <th class="text-right text-xs font-medium text-gray-500 px-4 py-3">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-white/5">
              <tr v-if="filteredChapters.length === 0">
                <td colspan="4" class="text-center text-gray-600 py-12 text-sm">No chapters found</td>
              </tr>
              <tr v-for="ch in filteredChapters" :key="ch.id" class="hover:bg-white/5 transition-colors">
                <td class="px-4 py-3">
                  <p class="text-white text-sm">{{ ch.title }}</p>
                  <p class="text-gray-600 text-xs font-mono">Ch.{{ ch.number }}</p>
                </td>
                <td class="px-4 py-3">
                  <span :class="['text-xs px-2 py-0.5 rounded-full', ch.image_count > 0 ? 'bg-green-600/20 text-green-400' : 'bg-yellow-600/20 text-yellow-400']">
                    {{ ch.image_count > 0 ? ch.image_count + ' imgs' : 'No images' }}
                  </span>
                </td>
                <td class="px-4 py-3 text-gray-400 text-sm">{{ ch.view_count }}</td>
                <td class="px-4 py-3">
                  <div class="flex items-center justify-end gap-2">
                    <button @click="doScrapeImages(ch)" :disabled="scrapingId === ch.id"
                      class="text-xs text-indigo-400 hover:text-indigo-300 transition-colors disabled:opacity-50">
                      {{ scrapingId === ch.id ? 'Scraping...' : 'Get Images' }}
                    </button>
                    <button @click="confirmDelete(ch)" class="btn-danger">Delete</button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div v-if="chapterPage > 1 || chapterHasMore" class="flex items-center justify-between">
          <span class="text-xs text-gray-500">Page {{ chapterPage }}</span>
          <div class="flex items-center gap-1">
            <button @click="chapterPagePrev" :disabled="chapterPage === 1 || chaptersLoading"
              class="px-3 py-1 text-xs text-gray-400 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed">‹ Prev</button>
            <span class="px-3 py-1 text-xs text-white bg-white/10 rounded">{{ chapterPage }}</span>
            <button @click="chapterPageNext" :disabled="!chapterHasMore || chaptersLoading"
              class="px-3 py-1 text-xs text-gray-400 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed">Next ›</button>
          </div>
        </div>
      </template>
    </div>

    <!-- Delete confirm modal -->
    <div v-if="deleteTarget" class="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4" @click.self="deleteTarget = null">
      <div class="bg-[#1a1a24] border border-white/10 rounded-2xl p-6 max-w-sm w-full">
        <h3 class="text-white font-bold mb-2">Delete Chapter?</h3>
        <p class="text-gray-400 text-sm mb-5">
          Delete <strong class="text-white">{{ deleteTarget.title }}</strong>?
          This cannot be undone.
        </p>
        <div class="flex gap-3">
          <button @click="deleteTarget = null" class="flex-1 py-2 border border-white/10 text-gray-400 rounded-lg hover:bg-white/5 text-sm">Cancel</button>
          <button @click="doDelete" :disabled="deleting"
            class="flex-1 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg text-sm font-medium disabled:opacity-50">
            {{ deleting ? 'Deleting...' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import api from '@/services/api'

// Series panel
const seriesList = ref<any[]>([])
const seriesLoading = ref(true)
const seriesSearch = ref('')
const selectedSeries = ref<any>(null)
const seriesPage = ref(1)
const seriesPerPage = 15
const filteredSeries = computed(() => {
  if (!seriesSearch.value) return seriesList.value
  const q = seriesSearch.value.toLowerCase()
  return seriesList.value.filter(s => s.title.toLowerCase().includes(q))
})
const seriesTotalPages = computed(() => Math.ceil(filteredSeries.value.length / seriesPerPage))
const paginatedSeries = computed(() => filteredSeries.value.slice((seriesPage.value - 1) * seriesPerPage, seriesPage.value * seriesPerPage))
watch(seriesSearch, () => { seriesPage.value = 1 })

// Chapters panel — server-side paginated
const chapters = ref<any[]>([])
const chaptersLoading = ref(false)
const chapterSearch = ref('')
const scrapingId = ref<string | null>(null)
const chapterPage = ref(1)
const chapterCursors = ref<string[]>(['']) // cursors[page-1] = cursor needed to fetch that page
const chapterHasMore = ref(false)

const filteredChapters = computed(() => {
  if (!chapterSearch.value) return chapters.value
  const q = chapterSearch.value.toLowerCase()
  return chapters.value.filter(c => c.title.toLowerCase().includes(q) || String(c.number).includes(q))
})

// Bulk scrape
const bulkScraping = ref(false)
const bulkProgress = ref(0)
const bulkTotal = ref(0)

// Dedup
const deduping = ref(false)

// Delete
const deleteTarget = ref<any>(null)
const deleting = ref(false)

function imgError(e: Event) { (e.target as HTMLImageElement).style.display = 'none' }

async function fetchChaptersPage(seriesId: string, page: number) {
  chaptersLoading.value = true
  chapters.value = []
  try {
    const cursor = chapterCursors.value[page - 1] ?? ''
    const res = await api.get('/admin/chapters', { params: { series_id: seriesId, cursor } })
    chapters.value = (res.data.data as any[]).map(ch => ({ ...ch, image_count: ch.image_count ?? 0 }))
    const next = res.data.next_cursor ?? ''
    chapterHasMore.value = !!next
    // Store cursor for the next page if we don't have it yet
    if (next && !chapterCursors.value[page]) {
      chapterCursors.value[page] = next
    }
  } catch {} finally {
    chaptersLoading.value = false
  }
}

async function selectSeries(s: any) {
  selectedSeries.value = s
  chapterSearch.value = ''
  chapterPage.value = 1
  chapterCursors.value = ['']
  chapterHasMore.value = false
  await fetchChaptersPage(s.id, 1)
}

async function chapterPageNext() {
  if (!chapterHasMore.value || !selectedSeries.value) return
  chapterPage.value++
  await fetchChaptersPage(selectedSeries.value.id, chapterPage.value)
}

async function chapterPagePrev() {
  if (chapterPage.value <= 1 || !selectedSeries.value) return
  chapterPage.value--
  await fetchChaptersPage(selectedSeries.value.id, chapterPage.value)
}

async function doScrapeImages(ch: any) {
  const srcURL = prompt(`Enter chapter URL for "${ch.title}":`, ch.source_url || '')
  if (!srcURL) return
  scrapingId.value = ch.id
  try {
    const res = await api.post('/admin/import/chapter-images', { chapter_id: ch.id, chapter_url: srcURL })
    ch.image_count = res.data.images_count
    alert(`Saved ${res.data.images_count} images!`)
  } catch (e: any) {
    alert('Failed: ' + (e.response?.data?.error || e.message))
  } finally {
    scrapingId.value = null
  }
}

async function doScrapeAllImages() {
  const targets = chapters.value.filter(ch => ch.source_url)
  if (targets.length === 0) { alert('No chapters have a source URL.'); return }
  if (!confirm(`Scrape images for all ${targets.length} chapters on this page?`)) return
  bulkScraping.value = true
  bulkProgress.value = 0
  bulkTotal.value = targets.length
  for (const ch of targets) {
    try {
      const res = await api.post('/admin/import/chapter-images', { chapter_id: ch.id, chapter_url: ch.source_url })
      ch.image_count = res.data.images_count
    } catch {}
    bulkProgress.value++
  }
  bulkScraping.value = false
  alert(`Done! Scraped images for ${targets.length} chapters.`)
}

async function doDedup() {
  if (!selectedSeries.value) return
  if (!confirm(`Remove duplicate chapters from "${selectedSeries.value.title}"?`)) return
  deduping.value = true
  try {
    await api.delete('/admin/chapters/duplicates', { params: { series_id: selectedSeries.value.id } })
    // Poll until background job finishes
    await new Promise<void>(resolve => {
      const poll = setInterval(async () => {
        try {
          const st = await api.get('/admin/chapters/duplicates/status')
          if (!st.data.running) {
            clearInterval(poll)
            const deleted = st.data.last_deleted ?? 0
            alert(`Removed ${deleted} duplicate chapter${deleted !== 1 ? 's' : ''}.`)
            if (deleted > 0 && selectedSeries.value)
              await fetchChaptersPage(selectedSeries.value.id, chapterPage.value)
            resolve()
          }
        } catch { clearInterval(poll); resolve() }
      }, 3000)
    })
  } catch (e: any) {
    alert('Failed: ' + (e.response?.data?.error || e.message))
  } finally {
    deduping.value = false
  }
}

function confirmDelete(ch: any) { deleteTarget.value = ch }

async function doDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await api.delete(`/admin/chapters/${deleteTarget.value.id}`)
    chapters.value = chapters.value.filter(c => c.id !== deleteTarget.value.id)
    if (selectedSeries.value) selectedSeries.value.chapter_count--
    deleteTarget.value = null
  } catch {} finally {
    deleting.value = false
  }
}

onMounted(async () => {
  seriesLoading.value = true
  try {
    const res = await api.get('/admin/series', { params: { limit: 500 } })
    seriesList.value = res.data.data
  } catch {} finally {
    seriesLoading.value = false
  }
})
</script>
