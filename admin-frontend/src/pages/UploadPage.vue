<template>
  <div class="max-w-3xl space-y-8">

    <!-- Add New Series -->
    <div class="admin-card space-y-4">
      <h2 class="text-white font-bold text-lg">Add New Series</h2>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div>
          <label class="form-label">Title *</label>
          <input v-model="newSeries.title" type="text" class="form-input w-full" placeholder="My Sister's Friends"/>
        </div>
        <div>
          <label class="form-label">Language *</label>
          <select v-model="newSeries.language" class="form-input w-full">
            <option value="en">English (EN)</option>
            <option value="my">Myanmar (MY)</option>
          </select>
        </div>
        <div>
          <label class="form-label">Status</label>
          <select v-model="newSeries.status" class="form-input w-full">
            <option value="ongoing">Ongoing</option>
            <option value="completed">Completed</option>
          </select>
        </div>
        <div>
          <label class="form-label">Author</label>
          <input v-model="newSeries.author" type="text" class="form-input w-full" placeholder="Author name"/>
        </div>
        <div class="sm:col-span-2">
          <label class="form-label">Genres <span class="text-gray-600 text-xs">(comma-separated)</span></label>
          <input v-model="newSeries.genres" type="text" class="form-input w-full" placeholder="Action, Romance, Fantasy"/>
        </div>
        <div class="sm:col-span-2">
          <label class="form-label">Description</label>
          <textarea v-model="newSeries.description" rows="3" class="form-input w-full resize-none" placeholder="Series description..."/>
        </div>
        <div class="sm:col-span-2">
          <label class="form-label">Cover Image <span class="text-gray-600 text-xs">(optional, can upload later)</span></label>
          <input @change="handleSeriesFile" type="file" accept="image/*" class="form-input w-full"/>
        </div>
      </div>

      <div v-if="seriesError" class="text-red-400 text-sm">{{ seriesError }}</div>
      <div v-if="seriesSuccess" class="text-green-400 text-sm">{{ seriesSuccess }}</div>

      <button @click="createSeries" :disabled="creatingSeries || !newSeries.title" class="btn-primary disabled:opacity-50 disabled:cursor-not-allowed">
        <svg v-if="creatingSeries" class="w-4 h-4 animate-spin inline mr-1" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
        </svg>
        {{ creatingSeries ? 'Creating...' : 'Create Series' }}
      </button>
    </div>

    <!-- Upload Cover for Existing Series -->
    <div class="admin-card space-y-4">
      <h2 class="text-white font-bold text-lg">Upload Cover for Existing Series</h2>
      <div>
        <label class="form-label">Select Series *</label>
        <select v-model="coverSeriesId" class="form-input w-full">
          <option value="">— Select series —</option>
          <option v-for="s in seriesList" :key="s.id" :value="s.id">{{ s.title }} ({{ s.language?.toUpperCase() }})</option>
        </select>
      </div>
      <div>
        <label class="form-label">Cover Image *</label>
        <input @change="handleCoverFile" type="file" accept="image/*" class="form-input w-full"/>
      </div>
      <div v-if="coverError" class="text-red-400 text-sm">{{ coverError }}</div>
      <div v-if="coverSuccess" class="text-green-400 text-sm">{{ coverSuccess }}</div>
      <button @click="uploadCover" :disabled="uploadingCover || !coverSeriesId || !coverFile" class="btn-primary disabled:opacity-50 disabled:cursor-not-allowed">
        <svg v-if="uploadingCover" class="w-4 h-4 animate-spin inline mr-1" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
        </svg>
        {{ uploadingCover ? 'Uploading...' : 'Upload Cover' }}
      </button>
    </div>

    <!-- Add New Chapter -->
    <div class="admin-card space-y-4">
      <h2 class="text-white font-bold text-lg">Add New Chapter</h2>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div class="sm:col-span-2">
          <label class="form-label">Series *</label>
          <select v-model="newChapter.series_id" class="form-input w-full">
            <option value="">— Select series —</option>
            <option v-for="s in seriesList" :key="s.id" :value="s.id">{{ s.title }} ({{ s.language?.toUpperCase() }})</option>
          </select>
        </div>
        <div>
          <label class="form-label">Chapter Number *</label>
          <input v-model.number="newChapter.number" type="number" min="0" step="0.5" class="form-input w-full" placeholder="1"/>
        </div>
        <div>
          <label class="form-label">Chapter Title</label>
          <input v-model="newChapter.title" type="text" class="form-input w-full" placeholder="Chapter 1"/>
        </div>
        <div class="sm:col-span-2">
          <label class="form-label">Chapter Images *</label>
          <input @change="handleChapterFiles" type="file" accept="image/*" multiple class="form-input w-full"/>
          <p v-if="chapterFiles" class="text-gray-500 text-xs mt-1">{{ chapterFiles.length }} file(s) selected</p>
        </div>
      </div>

      <div v-if="chapterError" class="text-red-400 text-sm">{{ chapterError }}</div>
      <div v-if="chapterSuccess" class="text-green-400 text-sm">{{ chapterSuccess }}</div>

      <button @click="createChapter" :disabled="creatingChapter || !newChapter.series_id || !newChapter.number" class="btn-primary disabled:opacity-50 disabled:cursor-not-allowed">
        <svg v-if="creatingChapter" class="w-4 h-4 animate-spin inline mr-1" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
        </svg>
        {{ creatingChapter ? 'Creating & Uploading...' : 'Create Chapter & Upload Images' }}
      </button>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api, { adminApi } from '@/services/api'

const seriesList = ref<any[]>([])

// New series form
const newSeries = ref({ title: '', language: 'en', status: 'ongoing', author: '', genres: '', description: '' })
const seriesFile = ref<File | null>(null)
const creatingSeries = ref(false)
const seriesError = ref('')
const seriesSuccess = ref('')

// Cover upload for existing series
const coverSeriesId = ref('')
const coverFile = ref<File | null>(null)
const uploadingCover = ref(false)
const coverError = ref('')
const coverSuccess = ref('')

// New chapter form
const newChapter = ref({ series_id: '', title: '', number: null as number | null })
const chapterFiles = ref<FileList | null>(null)
const creatingChapter = ref(false)
const chapterError = ref('')
const chapterSuccess = ref('')

function handleSeriesFile(e: Event) {
  seriesFile.value = (e.target as HTMLInputElement).files?.[0] || null
}
function handleCoverFile(e: Event) {
  coverFile.value = (e.target as HTMLInputElement).files?.[0] || null
}
function handleChapterFiles(e: Event) {
  chapterFiles.value = (e.target as HTMLInputElement).files
}

async function loadSeries() {
  try {
    const res = await api.get('/admin/series?limit=200')
    seriesList.value = res.data.data || []
  } catch {}
}

async function createSeries() {
  if (!newSeries.value.title) return
  creatingSeries.value = true
  seriesError.value = ''
  seriesSuccess.value = ''
  try {
    const res = await api.post('/admin/series', newSeries.value)
    const created = res.data
    seriesList.value.unshift(created)

    // Upload cover if provided
    if (seriesFile.value) {
      const formData = new FormData()
      formData.append('cover', seriesFile.value)
      await adminApi.uploadSeriesCover(created.id, formData)
    }

    seriesSuccess.value = `✓ Series "${created.title}" created!`
    newSeries.value = { title: '', language: 'en', status: 'ongoing', author: '', genres: '', description: '' }
    seriesFile.value = null
  } catch (e: any) {
    seriesError.value = e.response?.data?.error || e.message || 'Failed to create series'
  } finally {
    creatingSeries.value = false
  }
}

async function uploadCover() {
  if (!coverFile.value || !coverSeriesId.value) return
  uploadingCover.value = true
  coverError.value = ''
  coverSuccess.value = ''
  try {
    const formData = new FormData()
    formData.append('cover', coverFile.value)
    await adminApi.uploadSeriesCover(coverSeriesId.value, formData)
    coverSuccess.value = '✓ Cover uploaded!'
    coverSeriesId.value = ''
    coverFile.value = null
  } catch (e: any) {
    coverError.value = e.response?.data?.error || e.message || 'Upload failed'
  } finally {
    uploadingCover.value = false
  }
}

async function createChapter() {
  if (!newChapter.value.series_id || !newChapter.value.number) return
  creatingChapter.value = true
  chapterError.value = ''
  chapterSuccess.value = ''
  try {
    // Create chapter
    const res = await api.post('/admin/chapters', {
      series_id: newChapter.value.series_id,
      title: newChapter.value.title || `Chapter ${newChapter.value.number}`,
      number: newChapter.value.number,
    })
    const created = res.data

    // Upload images if provided
    if (chapterFiles.value && chapterFiles.value.length > 0) {
      const formData = new FormData()
      for (let i = 0; i < chapterFiles.value.length; i++) {
        formData.append('images', chapterFiles.value[i])
      }
      await adminApi.uploadChapterImages(created.id, formData)
    }

    chapterSuccess.value = `✓ Chapter ${created.number} created with ${chapterFiles.value?.length ?? 0} images!`
    newChapter.value = { series_id: '', title: '', number: null }
    chapterFiles.value = null
  } catch (e: any) {
    chapterError.value = e.response?.data?.error || e.message || 'Failed to create chapter'
  } finally {
    creatingChapter.value = false
  }
}

onMounted(loadSeries)
</script>

<style scoped>
.form-label {
  @apply block text-sm font-medium text-gray-300 mb-1;
}
</style>
