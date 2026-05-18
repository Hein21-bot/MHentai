<template>
  <div class="max-w-3xl space-y-6">

    <!-- Tab switcher -->
    <div class="flex gap-2 border-b border-white/10">
      <button
        @click="activeTab = 'hentai20'"
        :class="['px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors',
          activeTab === 'hentai20'
            ? 'border-indigo-500 text-indigo-400'
            : 'border-transparent text-gray-500 hover:text-gray-300']">
        Hentai20.io
      </button>
      <button
        @click="activeTab = 'mangaboost'"
        :class="['px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors',
          activeTab === 'mangaboost'
            ? 'border-indigo-500 text-indigo-400'
            : 'border-transparent text-gray-500 hover:text-gray-300']">
        MangaBoost
      </button>
    </div>

    <!-- ───── Hentai20.io Tab ───── -->
    <template v-if="activeTab === 'hentai20'">

      <!-- Step 1: URL input -->
      <div class="admin-card">
        <h2 class="text-white font-bold text-lg mb-1">Import from hentai20.io</h2>
        <p class="text-gray-500 text-sm mb-5">Paste a manga URL to preview series info and chapters before importing.</p>
        <div class="flex gap-3">
          <input v-model="h20.url" type="url" placeholder="https://hentai20.io/manga/series-name/"
            class="form-input flex-1" @keydown.enter="h20Preview" :disabled="h20.previewing || h20.importing"/>
          <button @click="h20Preview" :disabled="h20.previewing || !h20.url.trim() || h20.importing"
            class="btn-primary disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap">
            <svg v-if="h20.previewing" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
            </svg>
            {{ h20.previewing ? 'Fetching...' : 'Preview' }}
          </button>
        </div>
        <p v-if="h20.previewError" class="mt-3 text-red-400 text-sm">{{ h20.previewError }}</p>
      </div>

      <template v-if="h20.preview">
        <!-- Series info card -->
        <div class="admin-card">
          <div class="flex gap-4">
            <img v-if="h20.preview.cover_url" :src="h20.preview.cover_url" :alt="h20.preview.title"
              class="w-24 h-36 object-cover rounded-lg flex-shrink-0" @error="(e) => (e.target as HTMLImageElement).style.display='none'"/>
            <div v-else class="w-24 h-36 bg-[#12121a] rounded-lg flex-shrink-0"/>
            <div class="flex-1 min-w-0">
              <div class="flex items-start gap-2 flex-wrap">
                <h3 class="text-white font-bold text-lg">{{ h20.preview.title }}</h3>
                <span :class="['text-xs px-2 py-0.5 rounded-full font-medium mt-1',
                  h20.preview.status === 'ongoing' ? 'bg-green-600/20 text-green-400' : 'bg-blue-600/20 text-blue-400']">
                  {{ h20.preview.status }}
                </span>
              </div>
              <p v-if="h20.preview.author" class="text-gray-500 text-sm mt-1">{{ h20.preview.author }}</p>
              <p v-if="h20.preview.genres" class="text-gray-600 text-xs mt-1">{{ h20.preview.genres }}</p>
              <p v-if="h20.preview.description" class="text-gray-400 text-sm mt-2 line-clamp-3">{{ h20.preview.description }}</p>
              <p class="text-gray-600 text-xs mt-2">{{ h20.preview.chapters.length }} chapters found</p>
            </div>
          </div>
        </div>

        <!-- Chapter selection -->
        <div class="admin-card">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-white font-semibold">Select Chapters to Import</h3>
            <div class="flex items-center gap-3">
              <button @click="h20SelectAll" class="text-xs text-indigo-400 hover:text-indigo-300">All</button>
              <button @click="h20SelectNone" class="text-xs text-gray-500 hover:text-gray-300">None</button>
              <span class="text-xs text-gray-600">{{ h20.selectedSlugs.size }} / {{ h20.preview.chapters.length }} selected</span>
            </div>
          </div>
          <div class="max-h-72 overflow-y-auto rounded-lg border border-white/10 divide-y divide-white/5">
            <label v-for="ch in h20.preview.chapters" :key="ch.slug"
              class="flex items-center gap-3 px-4 py-2.5 cursor-pointer hover:bg-white/5 transition-colors">
              <input type="checkbox"
                :checked="h20.selectedSlugs.has(ch.slug)"
                @change="h20ToggleChapter(ch.slug)"
                class="rounded accent-indigo-500 flex-shrink-0"/>
              <span class="flex-1 text-sm text-gray-200 truncate">{{ ch.title || `Chapter ${ch.number}` }}</span>
              <span class="text-xs text-gray-600 flex-shrink-0">Ch.{{ ch.number }}</span>
            </label>
          </div>
          <div class="mt-4 flex gap-3">
            <button @click="h20Reset" class="px-4 py-2 border border-white/10 text-gray-400 rounded-lg hover:bg-white/5 text-sm">
              ← New URL
            </button>
            <button @click="h20Import" :disabled="h20.importing || h20.selectedSlugs.size === 0"
              class="flex-1 btn-primary disabled:opacity-50 disabled:cursor-not-allowed">
              <svg v-if="h20.importing" class="w-4 h-4 animate-spin inline mr-1" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
              </svg>
              {{ h20.importing ? 'Importing...' : `Import ${h20.selectedSlugs.size} chapter${h20.selectedSlugs.size !== 1 ? 's' : ''}` }}
            </button>
          </div>
        </div>

        <!-- Import log -->
        <div v-if="h20.logs.length > 0" class="admin-card">
          <h3 class="text-white font-semibold mb-3">Import Log</h3>
          <div class="space-y-1 font-mono text-xs max-h-48 overflow-y-auto">
            <div v-for="(log, i) in h20.logs" :key="i"
              :class="['py-0.5', log.type === 'error' ? 'text-red-400' : log.type === 'success' ? 'text-green-400' : 'text-gray-400']">
              {{ log.msg }}
            </div>
          </div>
        </div>

        <!-- Success -->
        <div v-if="h20.result" class="admin-card border-green-600/30">
          <div class="flex items-start gap-3">
            <svg class="w-5 h-5 text-green-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            <div>
              <p class="text-green-400 font-bold">Import Successful!</p>
              <p class="text-gray-300 text-sm mt-1"><strong class="text-white">{{ h20.result.series?.title }}</strong></p>
              <p class="text-gray-500 text-xs mt-0.5">{{ h20.result.chapters_saved }} chapters saved, {{ h20.result.chapters_skipped }} skipped</p>
              <div class="flex gap-2 mt-3">
                <RouterLink to="/series" class="btn-primary text-xs py-1.5">View in Series</RouterLink>
                <button @click="h20Reset" class="px-3 py-1.5 border border-white/10 text-gray-400 rounded-lg hover:bg-white/5 text-xs">Import Another</button>
              </div>
            </div>
          </div>
        </div>

        <!-- Error -->
        <div v-if="h20.importError" class="admin-card border-red-600/30">
          <div class="flex items-start gap-3">
            <svg class="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            <div>
              <p class="text-red-400 font-bold">Import Failed</p>
              <p class="text-gray-400 text-sm mt-1">{{ h20.importError }}</p>
            </div>
          </div>
        </div>
      </template>
    </template>

    <!-- ───── MangaBoost Tab ───── -->
    <template v-if="activeTab === 'mangaboost'">

      <!-- Step 1: URL input -->
      <div class="admin-card">
        <h2 class="text-white font-bold text-lg mb-1">Import from MangaBoost</h2>
        <p class="text-gray-500 text-sm mb-5">Paste a mangaboost.com manga URL to preview and import.</p>
        <div class="flex gap-3">
          <input v-model="mb.url" type="url" placeholder="https://mangaboost.com/manga/series-name/"
            class="form-input flex-1" @keydown.enter="mbPreview" :disabled="mb.previewing || mb.importing"/>
          <button @click="mbPreview" :disabled="mb.previewing || !mb.url.trim() || mb.importing"
            class="btn-primary disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap">
            <svg v-if="mb.previewing" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
            </svg>
            {{ mb.previewing ? 'Fetching...' : 'Preview' }}
          </button>
        </div>
        <p v-if="mb.previewError" class="mt-3 text-red-400 text-sm">{{ mb.previewError }}</p>
      </div>

      <template v-if="mb.preview">
        <!-- Series info card -->
        <div class="admin-card">
          <div class="flex gap-4">
            <img v-if="mb.preview.cover_url" :src="mb.preview.cover_url" :alt="mb.preview.title"
              class="w-24 h-36 object-cover rounded-lg flex-shrink-0" @error="(e) => (e.target as HTMLImageElement).style.display='none'"/>
            <div v-else class="w-24 h-36 bg-[#12121a] rounded-lg flex-shrink-0"/>
            <div class="flex-1 min-w-0">
              <div class="flex items-start gap-2 flex-wrap">
                <h3 class="text-white font-bold text-lg">{{ mb.preview.title }}</h3>
                <span :class="['text-xs px-2 py-0.5 rounded-full font-medium mt-1',
                  mb.preview.status === 'ongoing' ? 'bg-green-600/20 text-green-400' : 'bg-blue-600/20 text-blue-400']">
                  {{ mb.preview.status }}
                </span>
              </div>
              <p v-if="mb.preview.author" class="text-gray-500 text-sm mt-1">{{ mb.preview.author }}</p>
              <p v-if="mb.preview.genres" class="text-gray-600 text-xs mt-1">{{ mb.preview.genres }}</p>
              <p v-if="mb.preview.description" class="text-gray-400 text-sm mt-2 line-clamp-3">{{ mb.preview.description }}</p>
              <p class="text-gray-600 text-xs mt-2">{{ mb.preview.chapters.length }} chapters found</p>
            </div>
          </div>
        </div>

        <!-- Chapter selection -->
        <div class="admin-card">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-white font-semibold">Select Chapters to Import</h3>
            <div class="flex items-center gap-3">
              <button @click="mbSelectAll" class="text-xs text-indigo-400 hover:text-indigo-300">All</button>
              <button @click="mbSelectNone" class="text-xs text-gray-500 hover:text-gray-300">None</button>
              <span class="text-xs text-gray-600">{{ mb.selectedSlugs.size }} / {{ mb.preview.chapters.length }} selected</span>
            </div>
          </div>
          <div class="max-h-72 overflow-y-auto rounded-lg border border-white/10 divide-y divide-white/5">
            <label v-for="ch in mb.preview.chapters" :key="ch.slug"
              class="flex items-center gap-3 px-4 py-2.5 cursor-pointer hover:bg-white/5 transition-colors">
              <input type="checkbox"
                :checked="mb.selectedSlugs.has(ch.slug)"
                @change="mbToggleChapter(ch.slug)"
                class="rounded accent-indigo-500 flex-shrink-0"/>
              <span class="flex-1 text-sm text-gray-200 truncate">{{ ch.title || `Chapter ${ch.number}` }}</span>
              <span class="text-xs text-gray-600 flex-shrink-0">Ch.{{ ch.number }}</span>
            </label>
          </div>
          <div class="mt-4 flex gap-3">
            <button @click="mbReset" class="px-4 py-2 border border-white/10 text-gray-400 rounded-lg hover:bg-white/5 text-sm">
              ← New URL
            </button>
            <button @click="mbImport" :disabled="mb.importing || mb.selectedSlugs.size === 0"
              class="flex-1 btn-primary disabled:opacity-50 disabled:cursor-not-allowed">
              <svg v-if="mb.importing" class="w-4 h-4 animate-spin inline mr-1" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
              </svg>
              {{ mb.importing ? 'Importing...' : `Import ${mb.selectedSlugs.size} chapter${mb.selectedSlugs.size !== 1 ? 's' : ''}` }}
            </button>
          </div>
        </div>

        <!-- Import log -->
        <div v-if="mb.logs.length > 0" class="admin-card">
          <h3 class="text-white font-semibold mb-3">Import Log</h3>
          <div class="space-y-1 font-mono text-xs max-h-48 overflow-y-auto">
            <div v-for="(log, i) in mb.logs" :key="i"
              :class="['py-0.5', log.type === 'error' ? 'text-red-400' : log.type === 'success' ? 'text-green-400' : 'text-gray-400']">
              {{ log.msg }}
            </div>
          </div>
        </div>

        <!-- Success -->
        <div v-if="mb.result" class="admin-card border-green-600/30">
          <div class="flex items-start gap-3">
            <svg class="w-5 h-5 text-green-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            <div>
              <p class="text-green-400 font-bold">Import Successful!</p>
              <p class="text-gray-300 text-sm mt-1"><strong class="text-white">{{ mb.result.series?.title }}</strong></p>
              <p class="text-gray-500 text-xs mt-0.5">{{ mb.result.chapters_saved }} chapters saved, {{ mb.result.chapters_skipped }} skipped</p>
              <div class="flex gap-2 mt-3">
                <RouterLink to="/series" class="btn-primary text-xs py-1.5">View in Series</RouterLink>
                <button @click="mbReset" class="px-3 py-1.5 border border-white/10 text-gray-400 rounded-lg hover:bg-white/5 text-xs">Import Another</button>
              </div>
            </div>
          </div>
        </div>

        <!-- Error -->
        <div v-if="mb.importError" class="admin-card border-red-600/30">
          <div class="flex items-start gap-3">
            <svg class="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            <div>
              <p class="text-red-400 font-bold">Import Failed</p>
              <p class="text-gray-400 text-sm mt-1">{{ mb.importError }}</p>
            </div>
          </div>
        </div>
      </template>
    </template>

  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import api from '@/services/api'

type LogEntry = { msg: string; type: 'info' | 'success' | 'error' }

interface PreviewChapter {
  title: string
  number: number
  slug: string
  url: string
}

interface Preview {
  title: string
  cover_url: string
  description: string
  status: string
  author: string
  genres: string
  chapters: PreviewChapter[]
}

interface ImportState {
  url: string
  previewing: boolean
  previewError: string
  preview: Preview | null
  selectedSlugs: Set<string>
  importing: boolean
  importError: string
  result: any
  logs: LogEntry[]
}

function makeState(): ImportState {
  return {
    url: '',
    previewing: false,
    previewError: '',
    preview: null,
    selectedSlugs: new Set(),
    importing: false,
    importError: '',
    result: null,
    logs: [],
  }
}

const activeTab = ref<'hentai20' | 'mangaboost'>('hentai20')
const h20 = reactive<ImportState>(makeState())
const mb = reactive<ImportState>(makeState())

// ── Hentai20 ──

async function h20Preview() {
  if (!h20.url.trim()) return
  h20.previewing = true
  h20.previewError = ''
  h20.preview = null
  try {
    const res = await api.post('/admin/import/preview-url', { url: h20.url.trim() })
    h20.preview = res.data
    h20.selectedSlugs = new Set(res.data.chapters.map((c: PreviewChapter) => c.slug))
  } catch (e: any) {
    h20.previewError = e.response?.data?.error || e.message || 'Failed to fetch preview'
  } finally {
    h20.previewing = false
  }
}

function h20ToggleChapter(slug: string) {
  const set = new Set(h20.selectedSlugs)
  if (set.has(slug)) set.delete(slug)
  else set.add(slug)
  h20.selectedSlugs = set
}

function h20SelectAll() {
  if (!h20.preview) return
  h20.selectedSlugs = new Set(h20.preview.chapters.map(c => c.slug))
}

function h20SelectNone() { h20.selectedSlugs = new Set() }

async function h20Import() {
  if (!h20.preview || h20.selectedSlugs.size === 0) return
  h20.importing = true
  h20.importError = ''
  h20.result = null
  h20.logs = []
  h20.logs.push({ msg: `Importing "${h20.preview.title}"...`, type: 'info' })
  h20.logs.push({ msg: `Chapters selected: ${h20.selectedSlugs.size}`, type: 'info' })
  try {
    const res = await api.post('/admin/import', {
      url: h20.url.trim(),
      selected_slugs: Array.from(h20.selectedSlugs),
      scrape_images: true,
      proxy_to_r2: true,
    })
    h20.result = res.data
    h20.logs.push({ msg: `✓ Series saved: ${res.data.series?.title}`, type: 'success' })
    h20.logs.push({ msg: `✓ ${res.data.chapters_saved} chapters saved, ${res.data.chapters_skipped} skipped`, type: 'success' })
  } catch (e: any) {
    const errMsg = e.response?.data?.error || e.message || 'Unknown error'
    h20.importError = errMsg
    h20.logs.push({ msg: `✗ ${errMsg}`, type: 'error' })
  } finally {
    h20.importing = false
  }
}

function h20Reset() {
  Object.assign(h20, makeState())
}

// ── MangaBoost ──

async function mbPreview() {
  if (!mb.url.trim()) return
  mb.previewing = true
  mb.previewError = ''
  mb.preview = null
  try {
    const res = await api.post('/admin/import/mangaboost/preview', { url: mb.url.trim() })
    mb.preview = res.data
    mb.selectedSlugs = new Set(res.data.chapters.map((c: PreviewChapter) => c.slug))
  } catch (e: any) {
    mb.previewError = e.response?.data?.error || e.message || 'Failed to fetch preview'
  } finally {
    mb.previewing = false
  }
}

function mbToggleChapter(slug: string) {
  const set = new Set(mb.selectedSlugs)
  if (set.has(slug)) set.delete(slug)
  else set.add(slug)
  mb.selectedSlugs = set
}

function mbSelectAll() {
  if (!mb.preview) return
  mb.selectedSlugs = new Set(mb.preview.chapters.map(c => c.slug))
}

function mbSelectNone() { mb.selectedSlugs = new Set() }

async function mbImport() {
  if (!mb.preview || mb.selectedSlugs.size === 0) return
  mb.importing = true
  mb.importError = ''
  mb.result = null
  mb.logs = []
  mb.logs.push({ msg: `Importing "${mb.preview.title}"...`, type: 'info' })
  mb.logs.push({ msg: `Chapters selected: ${mb.selectedSlugs.size}`, type: 'info' })
  try {
    const res = await api.post('/admin/import/mangaboost', {
      url: mb.url.trim(),
      selected_slugs: Array.from(mb.selectedSlugs),
    })
    mb.result = res.data
    mb.logs.push({ msg: `✓ Series saved: ${res.data.series?.title}`, type: 'success' })
    mb.logs.push({ msg: `✓ ${res.data.chapters_saved} chapters saved, ${res.data.chapters_skipped} skipped`, type: 'success' })
  } catch (e: any) {
    const errMsg = e.response?.data?.error || e.message || 'Unknown error'
    mb.importError = errMsg
    mb.logs.push({ msg: `✗ ${errMsg}`, type: 'error' })
  } finally {
    mb.importing = false
  }
}

function mbReset() {
  Object.assign(mb, makeState())
}
</script>
