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
        hentai20.io
      </button>
      <button
        @click="activeTab = 'mangaboost'"
        :class="['px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors',
          activeTab === 'mangaboost'
            ? 'border-indigo-500 text-indigo-400'
            : 'border-transparent text-gray-500 hover:text-gray-300']">
        MangaBoost
      </button>
      <button
        @click="activeTab = 'manhwamyanmar'"
        :class="['px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors',
          activeTab === 'manhwamyanmar'
            ? 'border-indigo-500 text-indigo-400'
            : 'border-transparent text-gray-500 hover:text-gray-300']">
        ManhwaMyanmar
      </button>
      <button
        @click="activeTab = 'yotepya'"
        :class="['px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors',
          activeTab === 'yotepya'
            ? 'border-indigo-500 text-indigo-400'
            : 'border-transparent text-gray-500 hover:text-gray-300']">
        YotePya
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
          <div class="flex items-center justify-between mb-3">
            <h3 class="text-white font-semibold">Select Chapters to Import</h3>
            <div class="flex items-center gap-3">
              <button @click="h20SelectAll" class="text-xs text-indigo-400 hover:text-indigo-300">All</button>
              <button @click="h20SelectNone" class="text-xs text-gray-500 hover:text-gray-300">None</button>
              <span class="text-xs text-gray-600">{{ h20.selectedSlugs.size }} / {{ h20.preview.chapters.length }} selected</span>
            </div>
          </div>
          <div class="flex items-center gap-2 mb-3">
            <span class="text-xs text-gray-500">From</span>
            <input type="number" min="1" v-model.number="h20.rangeFrom" @change="h20SelectRange"
              class="form-input w-20 text-sm py-1" placeholder="1"/>
            <span class="text-xs text-gray-500">To</span>
            <input type="number" min="1" v-model.number="h20.rangeTo" @change="h20SelectRange"
              class="form-input w-20 text-sm py-1" placeholder="100"/>
            <button @click="h20SelectRange" class="text-xs px-3 py-1 bg-indigo-600/20 text-indigo-400 hover:bg-indigo-600/30 rounded-lg transition-colors">
              Apply
            </button>
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
          <div class="mt-3 flex items-center gap-2">
            <input id="h20-force" type="checkbox" v-model="h20.force" class="rounded accent-orange-500"/>
            <label for="h20-force" class="text-xs text-orange-400 cursor-pointer select-none">
              Force re-import (overwrite already-imported chapters)
            </label>
          </div>
          <div class="mt-3 flex gap-3">
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
              <p class="text-gray-500 text-xs mt-0.5">{{ h20.result.chapters_saved }} saved, {{ h20.result.chapters_skipped }} skipped</p>
              <p v-if="h20.result.chapters_failed?.length > 0" class="text-red-400 text-xs mt-1">
                Failed: {{ h20.result.chapters_failed.map((n: number) => `Ch.${n}`).join(', ') }}
              </p>
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
          <div class="flex items-center justify-between mb-3">
            <h3 class="text-white font-semibold">Select Chapters to Import</h3>
            <div class="flex items-center gap-3">
              <button @click="mbSelectAll" class="text-xs text-indigo-400 hover:text-indigo-300">All</button>
              <button @click="mbSelectNone" class="text-xs text-gray-500 hover:text-gray-300">None</button>
              <span class="text-xs text-gray-600">{{ mb.selectedSlugs.size }} / {{ mb.preview.chapters.length }} selected</span>
            </div>
          </div>
          <div class="flex items-center gap-2 mb-3">
            <span class="text-xs text-gray-500">From</span>
            <input type="number" min="1" v-model.number="mb.rangeFrom" @change="mbSelectRange"
              class="form-input w-20 text-sm py-1" placeholder="1"/>
            <span class="text-xs text-gray-500">To</span>
            <input type="number" min="1" v-model.number="mb.rangeTo" @change="mbSelectRange"
              class="form-input w-20 text-sm py-1" placeholder="100"/>
            <button @click="mbSelectRange" class="text-xs px-3 py-1 bg-indigo-600/20 text-indigo-400 hover:bg-indigo-600/30 rounded-lg transition-colors">
              Apply
            </button>
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
          <div class="mt-3 flex items-center gap-2">
            <input id="mb-force" type="checkbox" v-model="mb.force" class="rounded accent-orange-500"/>
            <label for="mb-force" class="text-xs text-orange-400 cursor-pointer select-none">
              Force re-import (overwrite already-imported chapters)
            </label>
          </div>
          <div class="mt-3 flex gap-3">
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
              <p class="text-gray-500 text-xs mt-0.5">{{ mb.result.chapters_saved }} saved, {{ mb.result.chapters_skipped }} skipped</p>
              <p v-if="mb.result.chapters_failed?.length > 0" class="text-red-400 text-xs mt-1">
                Failed: {{ mb.result.chapters_failed.map((n: number) => `Ch.${n}`).join(', ') }}
              </p>
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

    <!-- ───── ManhwaMyanmar Tab ───── -->
    <template v-if="activeTab === 'manhwamyanmar'">

      <div class="admin-card">
        <h2 class="text-white font-bold text-lg mb-1">Import from adult.manhwamyanmar.com</h2>
        <p class="text-gray-500 text-sm mb-5">Paste a series URL to preview and import chapters.</p>
        <div class="flex gap-3">
          <input v-model="mm.url" type="url" placeholder="https://adult.manhwamyanmar.com/series-name/"
            class="form-input flex-1" @keydown.enter="mmPreview" :disabled="mm.previewing || mm.importing"/>
          <button @click="mmPreview" :disabled="mm.previewing || !mm.url.trim() || mm.importing"
            class="btn-primary disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap">
            <svg v-if="mm.previewing" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
            </svg>
            {{ mm.previewing ? 'Fetching...' : 'Preview' }}
          </button>
        </div>
        <p v-if="mm.previewError" class="mt-3 text-red-400 text-sm">{{ mm.previewError }}</p>
      </div>

      <template v-if="mm.preview">
        <div class="admin-card">
          <div class="flex gap-4">
            <img v-if="mm.preview.cover_url" :src="mm.preview.cover_url" :alt="mm.preview.title"
              class="w-24 h-36 object-cover rounded-lg flex-shrink-0" @error="(e) => (e.target as HTMLImageElement).style.display='none'"/>
            <div v-else class="w-24 h-36 bg-[#12121a] rounded-lg flex-shrink-0"/>
            <div class="flex-1 min-w-0">
              <div class="flex items-start gap-2 flex-wrap">
                <h3 class="text-white font-bold text-lg">{{ mm.preview.title }}</h3>
                <span :class="['text-xs px-2 py-0.5 rounded-full font-medium mt-1',
                  mm.preview.status === 'ongoing' ? 'bg-green-600/20 text-green-400' : 'bg-blue-600/20 text-blue-400']">
                  {{ mm.preview.status }}
                </span>
              </div>
              <p v-if="mm.preview.author" class="text-gray-500 text-sm mt-1">{{ mm.preview.author }}</p>
              <p v-if="mm.preview.genres" class="text-gray-600 text-xs mt-1">{{ mm.preview.genres }}</p>
              <p v-if="mm.preview.description" class="text-gray-400 text-sm mt-2 line-clamp-3">{{ mm.preview.description }}</p>
              <p class="text-gray-600 text-xs mt-2">{{ mm.preview.chapters.length }} chapters found</p>
            </div>
          </div>
        </div>

        <div class="admin-card">
          <div class="flex items-center justify-between mb-3">
            <h3 class="text-white font-semibold">Select Chapters to Import</h3>
            <div class="flex items-center gap-3">
              <button @click="mmSelectAll" class="text-xs text-indigo-400 hover:text-indigo-300">All</button>
              <button @click="mmSelectNone" class="text-xs text-gray-500 hover:text-gray-300">None</button>
              <span class="text-xs text-gray-600">{{ mm.selectedSlugs.size }} / {{ mm.preview.chapters.length }} selected</span>
            </div>
          </div>
          <div class="flex items-center gap-2 mb-3">
            <span class="text-xs text-gray-500">From</span>
            <input type="number" min="1" v-model.number="mm.rangeFrom" @change="mmSelectRange"
              class="form-input w-20 text-sm py-1" placeholder="1"/>
            <span class="text-xs text-gray-500">To</span>
            <input type="number" min="1" v-model.number="mm.rangeTo" @change="mmSelectRange"
              class="form-input w-20 text-sm py-1" placeholder="100"/>
            <button @click="mmSelectRange" class="text-xs px-3 py-1 bg-indigo-600/20 text-indigo-400 hover:bg-indigo-600/30 rounded-lg transition-colors">
              Apply
            </button>
          </div>
          <div class="max-h-72 overflow-y-auto rounded-lg border border-white/10 divide-y divide-white/5">
            <label v-for="ch in mm.preview.chapters" :key="ch.slug"
              class="flex items-center gap-3 px-4 py-2.5 cursor-pointer hover:bg-white/5 transition-colors">
              <input type="checkbox"
                :checked="mm.selectedSlugs.has(ch.slug)"
                @change="mmToggleChapter(ch.slug)"
                class="rounded accent-indigo-500 flex-shrink-0"/>
              <span class="flex-1 text-sm text-gray-200 truncate">{{ ch.title || `Chapter ${ch.number}` }}</span>
              <span class="text-xs text-gray-600 flex-shrink-0">Ch.{{ ch.number }}</span>
            </label>
          </div>
          <div class="mt-3 flex items-center gap-2">
            <input id="mm-force" type="checkbox" v-model="mm.force" class="rounded accent-orange-500"/>
            <label for="mm-force" class="text-xs text-orange-400 cursor-pointer select-none">
              Force re-import (overwrite already-imported chapters)
            </label>
          </div>
          <div class="mt-3 flex gap-3">
            <button @click="mmReset" class="px-4 py-2 border border-white/10 text-gray-400 rounded-lg hover:bg-white/5 text-sm">
              ← New URL
            </button>
            <button @click="mmImport" :disabled="mm.importing || mm.selectedSlugs.size === 0"
              class="flex-1 btn-primary disabled:opacity-50 disabled:cursor-not-allowed">
              <svg v-if="mm.importing" class="w-4 h-4 animate-spin inline mr-1" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
              </svg>
              {{ mm.importing ? 'Importing...' : `Import ${mm.selectedSlugs.size} chapter${mm.selectedSlugs.size !== 1 ? 's' : ''}` }}
            </button>
          </div>
        </div>

        <div v-if="mm.logs.length > 0" class="admin-card">
          <h3 class="text-white font-semibold mb-3">Import Log</h3>
          <div class="space-y-1 font-mono text-xs max-h-48 overflow-y-auto">
            <div v-for="(log, i) in mm.logs" :key="i"
              :class="['py-0.5', log.type === 'error' ? 'text-red-400' : log.type === 'success' ? 'text-green-400' : 'text-gray-400']">
              {{ log.msg }}
            </div>
          </div>
        </div>

        <div v-if="mm.result" class="admin-card border-green-600/30">
          <div class="flex items-start gap-3">
            <svg class="w-5 h-5 text-green-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            <div>
              <p class="text-green-400 font-bold">Import Successful!</p>
              <p class="text-gray-300 text-sm mt-1"><strong class="text-white">{{ mm.result.series?.title }}</strong></p>
              <p class="text-gray-500 text-xs mt-0.5">{{ mm.result.chapters_saved }} saved, {{ mm.result.chapters_skipped }} skipped</p>
              <p v-if="mm.result.chapters_failed?.length > 0" class="text-red-400 text-xs mt-1">
                Failed: {{ mm.result.chapters_failed.map((n: number) => `Ch.${n}`).join(', ') }}
              </p>
              <div class="flex gap-2 mt-3">
                <RouterLink to="/series" class="btn-primary text-xs py-1.5">View in Series</RouterLink>
                <button @click="mmReset" class="px-3 py-1.5 border border-white/10 text-gray-400 rounded-lg hover:bg-white/5 text-xs">Import Another</button>
              </div>
            </div>
          </div>
        </div>

        <div v-if="mm.importError" class="admin-card border-red-600/30">
          <div class="flex items-start gap-3">
            <svg class="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            <div>
              <p class="text-red-400 font-bold">Import Failed</p>
              <p class="text-gray-400 text-sm mt-1">{{ mm.importError }}</p>
            </div>
          </div>
        </div>
      </template>
    </template>

    <!-- ───── YotePya Tab ───── -->
    <template v-if="activeTab === 'yotepya'">

      <div class="admin-card">
        <h2 class="text-white font-bold text-lg mb-1">Import from yotepya.com</h2>
        <p class="text-gray-500 text-sm mb-5">Paste a series URL to preview and import chapters.</p>
        <div class="flex gap-3">
          <input v-model="yt.url" type="url" placeholder="https://yotepya.com/contents/71/"
            class="form-input flex-1" @keydown.enter="ytPreview" :disabled="yt.previewing || yt.importing"/>
          <button @click="ytPreview" :disabled="yt.previewing || !yt.url.trim() || yt.importing"
            class="btn-primary disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap">
            <svg v-if="yt.previewing" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
            </svg>
            {{ yt.previewing ? 'Fetching...' : 'Preview' }}
          </button>
        </div>
        <p v-if="yt.previewError" class="mt-3 text-red-400 text-sm">{{ yt.previewError }}</p>
      </div>

      <template v-if="yt.preview">
        <div class="admin-card">
          <div class="flex gap-4">
            <img v-if="yt.preview.cover_url" :src="yt.preview.cover_url" :alt="yt.preview.title"
              class="w-24 h-36 object-cover rounded-lg flex-shrink-0" @error="(e) => (e.target as HTMLImageElement).style.display='none'"/>
            <div v-else class="w-24 h-36 bg-[#12121a] rounded-lg flex-shrink-0"/>
            <div class="flex-1 min-w-0">
              <div class="flex items-start gap-2 flex-wrap">
                <h3 class="text-white font-bold text-lg">{{ yt.preview.title }}</h3>
                <span :class="['text-xs px-2 py-0.5 rounded-full font-medium mt-1',
                  yt.preview.status === 'ongoing' ? 'bg-green-600/20 text-green-400' : 'bg-blue-600/20 text-blue-400']">
                  {{ yt.preview.status }}
                </span>
              </div>
              <p v-if="yt.preview.author" class="text-gray-500 text-sm mt-1">{{ yt.preview.author }}</p>
              <p v-if="yt.preview.genres" class="text-gray-600 text-xs mt-1">{{ yt.preview.genres }}</p>
              <p v-if="yt.preview.description" class="text-gray-400 text-sm mt-2 line-clamp-3">{{ yt.preview.description }}</p>
              <p class="text-gray-600 text-xs mt-2">{{ yt.preview.chapters.length }} chapters found</p>
            </div>
          </div>
        </div>

        <div class="admin-card">
          <div class="flex items-center justify-between mb-3">
            <h3 class="text-white font-semibold">Select Chapters to Import</h3>
            <div class="flex items-center gap-3">
              <button @click="ytSelectAll" class="text-xs text-indigo-400 hover:text-indigo-300">All</button>
              <button @click="ytSelectNone" class="text-xs text-gray-500 hover:text-gray-300">None</button>
              <span class="text-xs text-gray-600">{{ yt.selectedSlugs.size }} / {{ yt.preview.chapters.length }} selected</span>
            </div>
          </div>
          <div class="flex items-center gap-2 mb-3">
            <span class="text-xs text-gray-500">From</span>
            <input type="number" min="1" v-model.number="yt.rangeFrom" @change="ytSelectRange"
              class="form-input w-20 text-sm py-1" placeholder="1"/>
            <span class="text-xs text-gray-500">To</span>
            <input type="number" min="1" v-model.number="yt.rangeTo" @change="ytSelectRange"
              class="form-input w-20 text-sm py-1" placeholder="100"/>
            <button @click="ytSelectRange" class="text-xs px-3 py-1 bg-indigo-600/20 text-indigo-400 hover:bg-indigo-600/30 rounded-lg transition-colors">
              Apply
            </button>
          </div>
          <div class="max-h-72 overflow-y-auto rounded-lg border border-white/10 divide-y divide-white/5">
            <label v-for="ch in yt.preview.chapters" :key="ch.slug"
              class="flex items-center gap-3 px-4 py-2.5 cursor-pointer hover:bg-white/5 transition-colors">
              <input type="checkbox"
                :checked="yt.selectedSlugs.has(ch.slug)"
                @change="ytToggleChapter(ch.slug)"
                class="rounded accent-indigo-500 flex-shrink-0"/>
              <span class="flex-1 text-sm text-gray-200 truncate">{{ ch.title || `Chapter ${ch.number}` }}</span>
              <span class="text-xs text-gray-600 flex-shrink-0">Ch.{{ ch.number }}</span>
            </label>
          </div>
          <div class="mt-3 flex items-center gap-2">
            <input id="yt-force" type="checkbox" v-model="yt.force" class="rounded accent-orange-500"/>
            <label for="yt-force" class="text-xs text-orange-400 cursor-pointer select-none">
              Force re-import (overwrite already-imported chapters)
            </label>
          </div>
          <div class="mt-3 flex gap-3">
            <button @click="ytReset" class="px-4 py-2 border border-white/10 text-gray-400 rounded-lg hover:bg-white/5 text-sm">
              ← New URL
            </button>
            <button @click="ytImport" :disabled="yt.importing || yt.selectedSlugs.size === 0"
              class="flex-1 btn-primary disabled:opacity-50 disabled:cursor-not-allowed">
              <svg v-if="yt.importing" class="w-4 h-4 animate-spin inline mr-1" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
              </svg>
              {{ yt.importing ? 'Importing...' : `Import ${yt.selectedSlugs.size} chapter${yt.selectedSlugs.size !== 1 ? 's' : ''}` }}
            </button>
          </div>
        </div>

        <div v-if="yt.logs.length > 0" class="admin-card">
          <h3 class="text-white font-semibold mb-3">Import Log</h3>
          <div class="space-y-1 font-mono text-xs max-h-48 overflow-y-auto">
            <div v-for="(log, i) in yt.logs" :key="i"
              :class="['py-0.5', log.type === 'error' ? 'text-red-400' : log.type === 'success' ? 'text-green-400' : 'text-gray-400']">
              {{ log.msg }}
            </div>
          </div>
        </div>

        <div v-if="yt.result" class="admin-card border-green-600/30">
          <div class="flex items-start gap-3">
            <svg class="w-5 h-5 text-green-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            <div>
              <p class="text-green-400 font-bold">Import Successful!</p>
              <p class="text-gray-300 text-sm mt-1"><strong class="text-white">{{ yt.result.series?.title }}</strong></p>
              <p class="text-gray-500 text-xs mt-0.5">{{ yt.result.chapters_saved }} saved, {{ yt.result.chapters_skipped }} skipped</p>
              <p v-if="yt.result.chapters_failed?.length > 0" class="text-red-400 text-xs mt-1">
                Failed: {{ yt.result.chapters_failed.map((n: number) => `Ch.${n}`).join(', ') }}
              </p>
              <div class="flex gap-2 mt-3">
                <RouterLink to="/series" class="btn-primary text-xs py-1.5">View in Series</RouterLink>
                <button @click="ytReset" class="px-3 py-1.5 border border-white/10 text-gray-400 rounded-lg hover:bg-white/5 text-xs">Import Another</button>
              </div>
            </div>
          </div>
        </div>

        <div v-if="yt.importError" class="admin-card border-red-600/30">
          <div class="flex items-start gap-3">
            <svg class="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            <div>
              <p class="text-red-400 font-bold">Import Failed</p>
              <p class="text-gray-400 text-sm mt-1">{{ yt.importError }}</p>
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
  force: boolean
  rangeFrom: number | null
  rangeTo: number | null
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
    force: false,
    rangeFrom: null,
    rangeTo: null,
  }
}

const activeTab = ref<'hentai20' | 'mangaboost' | 'manhwamyanmar' | 'yotepya'>('hentai20')
const h20 = reactive<ImportState>(makeState())
const mb = reactive<ImportState>(makeState())
const mm = reactive<ImportState>(makeState())
const yt = reactive<ImportState>(makeState())

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

function h20SelectRange() {
  if (!h20.preview) return
  const from = h20.rangeFrom ?? 1
  const to = h20.rangeTo ?? Infinity
  h20.selectedSlugs = new Set(
    h20.preview.chapters.filter(c => c.number >= from && c.number <= to).map(c => c.slug)
  )
}

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
      selected_chapters: h20.preview!.chapters.filter(c => h20.selectedSlugs.has(c.slug)),
      force: h20.force,
    })
    if (res.status === 202) {
      const { job_id, series, total } = res.data
      h20.logs.push({ msg: `✓ Series ready: ${series?.title} — running ${total} chapters in background...`, type: 'info' })
      await new Promise<void>(resolve => {
        const poll = setInterval(async () => {
          try {
            const st = await api.get(`/admin/import/status?job_id=${job_id}`)
            const d = st.data
            const progressMsg = `⏳ ${d.done}/${d.total} (${d.saved} saved, ${d.skipped} skipped${d.failed?.length > 0 ? `, ${d.failed.length} failed` : ''})`
            const last = h20.logs[h20.logs.length - 1]
            if (last?.msg.startsWith('⏳')) h20.logs[h20.logs.length - 1] = { msg: progressMsg, type: 'info' }
            else h20.logs.push({ msg: progressMsg, type: 'info' })
            if (!d.running) {
              clearInterval(poll)
              const failed: number[] = d.failed ?? []
              h20.result = { series, chapters_saved: d.saved, chapters_skipped: d.skipped, chapters_failed: failed }
              h20.logs.push({ msg: `✓ Done! ${d.saved} saved, ${d.skipped} skipped${failed.length > 0 ? `, ${failed.length} failed` : ''}`, type: failed.length > 0 ? 'error' : 'success' })
              if (failed.length > 0) h20.logs.push({ msg: `✗ Failed: ${failed.map((n: number) => `Ch.${n}`).join(', ')}`, type: 'error' })
              resolve()
            }
          } catch { clearInterval(poll); resolve() }
        }, 3000)
      })
    }
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

function mbSelectRange() {
  if (!mb.preview) return
  const from = mb.rangeFrom ?? 1
  const to = mb.rangeTo ?? Infinity
  mb.selectedSlugs = new Set(
    mb.preview.chapters.filter(c => c.number >= from && c.number <= to).map(c => c.slug)
  )
}

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
      selected_chapters: mb.preview!.chapters.filter(c => mb.selectedSlugs.has(c.slug)),
      force: mb.force,
    })
    if (res.status === 202) {
      const { job_id, series, total } = res.data
      mb.logs.push({ msg: `✓ Series ready: ${series?.title} — running ${total} chapters in background...`, type: 'info' })
      await new Promise<void>(resolve => {
        const poll = setInterval(async () => {
          try {
            const st = await api.get(`/admin/import/status?job_id=${job_id}`)
            const d = st.data
            const progressMsg = `⏳ ${d.done}/${d.total} (${d.saved} saved, ${d.skipped} skipped${d.failed?.length > 0 ? `, ${d.failed.length} failed` : ''})`
            const last = mb.logs[mb.logs.length - 1]
            if (last?.msg.startsWith('⏳')) mb.logs[mb.logs.length - 1] = { msg: progressMsg, type: 'info' }
            else mb.logs.push({ msg: progressMsg, type: 'info' })
            if (!d.running) {
              clearInterval(poll)
              const failed: number[] = d.failed ?? []
              mb.result = { series, chapters_saved: d.saved, chapters_skipped: d.skipped, chapters_failed: failed }
              mb.logs.push({ msg: `✓ Done! ${d.saved} saved, ${d.skipped} skipped${failed.length > 0 ? `, ${failed.length} failed` : ''}`, type: failed.length > 0 ? 'error' : 'success' })
              if (failed.length > 0) mb.logs.push({ msg: `✗ Failed: ${failed.map((n: number) => `Ch.${n}`).join(', ')}`, type: 'error' })
              resolve()
            }
          } catch { clearInterval(poll); resolve() }
        }, 3000)
      })
    }
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

// ── ManhwaMyanmar ──

async function mmPreview() {
  if (!mm.url.trim()) return
  mm.previewing = true
  mm.previewError = ''
  mm.preview = null
  try {
    const res = await api.post('/admin/import/manhwamyanmar/preview', { url: mm.url.trim() })
    mm.preview = res.data
    mm.selectedSlugs = new Set(res.data.chapters.map((c: PreviewChapter) => c.slug))
  } catch (e: any) {
    mm.previewError = e.response?.data?.error || e.message || 'Failed to fetch preview'
  } finally {
    mm.previewing = false
  }
}

function mmToggleChapter(slug: string) {
  const set = new Set(mm.selectedSlugs)
  if (set.has(slug)) set.delete(slug)
  else set.add(slug)
  mm.selectedSlugs = set
}

function mmSelectAll() {
  if (!mm.preview) return
  mm.selectedSlugs = new Set(mm.preview.chapters.map(c => c.slug))
}

function mmSelectNone() { mm.selectedSlugs = new Set() }

function mmSelectRange() {
  if (!mm.preview) return
  const from = mm.rangeFrom ?? 1
  const to = mm.rangeTo ?? Infinity
  mm.selectedSlugs = new Set(
    mm.preview.chapters.filter(c => c.number >= from && c.number <= to).map(c => c.slug)
  )
}

async function mmImport() {
  if (!mm.preview || mm.selectedSlugs.size === 0) return
  mm.importing = true
  mm.importError = ''
  mm.result = null
  mm.logs = []
  mm.logs.push({ msg: `Importing "${mm.preview.title}"...`, type: 'info' })
  mm.logs.push({ msg: `Chapters selected: ${mm.selectedSlugs.size}`, type: 'info' })
  try {
    const res = await api.post('/admin/import/manhwamyanmar', {
      url: mm.url.trim(),
      selected_chapters: mm.preview!.chapters.filter(c => mm.selectedSlugs.has(c.slug)),
      force: mm.force,
    })
    if (res.status === 202) {
      const { job_id, series, total } = res.data
      mm.logs.push({ msg: `✓ Series ready: ${series?.title} — running ${total} chapters in background...`, type: 'info' })
      await new Promise<void>(resolve => {
        const poll = setInterval(async () => {
          try {
            const st = await api.get(`/admin/import/status?job_id=${job_id}`)
            const d = st.data
            const progressMsg = `⏳ ${d.done}/${d.total} (${d.saved} saved, ${d.skipped} skipped${d.failed?.length > 0 ? `, ${d.failed.length} failed` : ''})`
            const last = mm.logs[mm.logs.length - 1]
            if (last?.msg.startsWith('⏳')) mm.logs[mm.logs.length - 1] = { msg: progressMsg, type: 'info' }
            else mm.logs.push({ msg: progressMsg, type: 'info' })
            if (!d.running) {
              clearInterval(poll)
              const failed: number[] = d.failed ?? []
              mm.result = { series, chapters_saved: d.saved, chapters_skipped: d.skipped, chapters_failed: failed }
              mm.logs.push({ msg: `✓ Done! ${d.saved} saved, ${d.skipped} skipped${failed.length > 0 ? `, ${failed.length} failed` : ''}`, type: failed.length > 0 ? 'error' : 'success' })
              if (failed.length > 0) mm.logs.push({ msg: `✗ Failed: ${failed.map((n: number) => `Ch.${n}`).join(', ')}`, type: 'error' })
              resolve()
            }
          } catch { clearInterval(poll); resolve() }
        }, 3000)
      })
    }
  } catch (e: any) {
    const errMsg = e.response?.data?.error || e.message || 'Unknown error'
    mm.importError = errMsg
    mm.logs.push({ msg: `✗ ${errMsg}`, type: 'error' })
  } finally {
    mm.importing = false
  }
}

function mmReset() {
  Object.assign(mm, makeState())
}

// ── YotePya ──

async function ytPreview() {
  if (!yt.url.trim()) return
  yt.previewing = true
  yt.previewError = ''
  yt.preview = null
  try {
    const res = await api.post('/admin/import/yotepya/preview', { url: yt.url.trim() })
    yt.preview = res.data
    yt.selectedSlugs = new Set(res.data.chapters.map((c: PreviewChapter) => c.slug))
  } catch (e: any) {
    yt.previewError = e.response?.data?.error || e.message || 'Failed to fetch preview'
  } finally {
    yt.previewing = false
  }
}

function ytToggleChapter(slug: string) {
  const set = new Set(yt.selectedSlugs)
  if (set.has(slug)) set.delete(slug)
  else set.add(slug)
  yt.selectedSlugs = set
}

function ytSelectAll() {
  if (!yt.preview) return
  yt.selectedSlugs = new Set(yt.preview.chapters.map(c => c.slug))
}

function ytSelectNone() { yt.selectedSlugs = new Set() }

function ytSelectRange() {
  if (!yt.preview) return
  const from = yt.rangeFrom ?? 1
  const to = yt.rangeTo ?? Infinity
  yt.selectedSlugs = new Set(
    yt.preview.chapters.filter(c => c.number >= from && c.number <= to).map(c => c.slug)
  )
}

async function ytImport() {
  if (!yt.preview || yt.selectedSlugs.size === 0) return
  yt.importing = true
  yt.importError = ''
  yt.result = null
  yt.logs = []
  yt.logs.push({ msg: `Importing "${yt.preview.title}"...`, type: 'info' })
  yt.logs.push({ msg: `Chapters selected: ${yt.selectedSlugs.size}`, type: 'info' })
  try {
    const res = await api.post('/admin/import/yotepya', {
      url: yt.url.trim(),
      selected_chapters: yt.preview!.chapters.filter(c => yt.selectedSlugs.has(c.slug)),
      force: yt.force,
    })
    if (res.status === 202) {
      const { job_id, series, total } = res.data
      yt.logs.push({ msg: `✓ Series ready: ${series?.title} — running ${total} chapters in background...`, type: 'info' })
      await new Promise<void>(resolve => {
        const poll = setInterval(async () => {
          try {
            const st = await api.get(`/admin/import/status?job_id=${job_id}`)
            const d = st.data
            const progressMsg = `⏳ ${d.done}/${d.total} (${d.saved} saved, ${d.skipped} skipped${d.failed?.length > 0 ? `, ${d.failed.length} failed` : ''})`
            const last = yt.logs[yt.logs.length - 1]
            if (last?.msg.startsWith('⏳')) yt.logs[yt.logs.length - 1] = { msg: progressMsg, type: 'info' }
            else yt.logs.push({ msg: progressMsg, type: 'info' })
            if (!d.running) {
              clearInterval(poll)
              const failed: number[] = d.failed ?? []
              yt.result = { series, chapters_saved: d.saved, chapters_skipped: d.skipped, chapters_failed: failed }
              yt.logs.push({ msg: `✓ Done! ${d.saved} saved, ${d.skipped} skipped${failed.length > 0 ? `, ${failed.length} failed` : ''}`, type: failed.length > 0 ? 'error' : 'success' })
              if (failed.length > 0) yt.logs.push({ msg: `✗ Failed: ${failed.map((n: number) => `Ch.${n}`).join(', ')}`, type: 'error' })
              resolve()
            }
          } catch { clearInterval(poll); resolve() }
        }, 3000)
      })
    }
  } catch (e: any) {
    const errMsg = e.response?.data?.error || e.message || 'Unknown error'
    yt.importError = errMsg
    yt.logs.push({ msg: `✗ ${errMsg}`, type: 'error' })
  } finally {
    yt.importing = false
  }
}

function ytReset() {
  Object.assign(yt, makeState())
}
</script>
