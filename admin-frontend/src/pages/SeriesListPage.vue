<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-white font-bold text-xl">Series ({{ total }})</h1>
      <div class="flex items-center gap-3">
        <button @click="deleteOrphans" :disabled="deletingOrphans"
          class="text-xs px-3 py-1.5 bg-red-600/20 text-red-400 hover:bg-red-600/30 rounded-lg transition-colors disabled:opacity-50 whitespace-nowrap">
          {{ deletingOrphans ? 'Cleaning...' : 'Clean orphan chapters' }}
        </button>
        <button @click="deleteEmpty" :disabled="deletingEmpty"
          class="text-xs px-3 py-1.5 bg-red-600/20 text-red-400 hover:bg-red-600/30 rounded-lg transition-colors disabled:opacity-50 whitespace-nowrap">
          {{ deletingEmpty ? 'Deleting...' : 'Delete 0-chapter series' }}
        </button>
        <input v-model="search" type="text" placeholder="Search..." class="form-input max-w-xs"/>
      </div>
    </div>

    <div v-if="loading" class="text-gray-500 text-sm">Loading...</div>

    <div class="bg-[#1a1a24] rounded-xl border border-white/10 overflow-hidden">
      <table class="w-full">
        <thead>
          <tr class="border-b border-white/10">
            <th class="text-left text-xs font-medium text-gray-500 px-4 py-3">Cover</th>
            <th class="text-left text-xs font-medium text-gray-500 px-4 py-3">Title</th>
            <th class="text-left text-xs font-medium text-gray-500 px-4 py-3">Lang</th>
            <th class="text-left text-xs font-medium text-gray-500 px-4 py-3">Status</th>
            <th class="text-left text-xs font-medium text-gray-500 px-4 py-3">Chapters</th>
            <th class="text-left text-xs font-medium text-gray-500 px-4 py-3">Views</th>
            <th class="text-right text-xs font-medium text-gray-500 px-4 py-3">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-white/5">
          <tr v-if="series.length === 0">
            <td colspan="6" class="text-center text-gray-600 py-12 text-sm">No series found</td>
          </tr>
          <tr v-for="s in series" :key="s.id" class="hover:bg-white/5 transition-colors">
            <td class="px-4 py-3">
              <img v-if="s.cover_url" :src="s.cover_url" :alt="s.title" class="w-10 h-14 object-cover rounded" @error="imgError"/>
              <div v-else class="w-10 h-14 bg-[#12121a] rounded"/>
            </td>
            <td class="px-4 py-3">
              <p class="text-white text-sm font-medium">{{ s.title }}</p>
              <p class="text-gray-600 text-xs font-mono">/{{ s.slug }}</p>
            </td>
            <td class="px-4 py-3">
              <span :class="['text-xs px-2 py-0.5 rounded-full font-medium font-mono', s.language === 'my' ? 'bg-yellow-600/20 text-yellow-400' : 'bg-indigo-600/20 text-indigo-400']">
                {{ s.language?.toUpperCase() || 'EN' }}
              </span>
            </td>
            <td class="px-4 py-3">
              <span :class="['text-xs px-2 py-0.5 rounded-full font-medium', s.status === 'ongoing' ? 'bg-green-600/20 text-green-400' : 'bg-blue-600/20 text-blue-400']">
                {{ s.status }}
              </span>
            </td>
            <td class="px-4 py-3 text-gray-400 text-sm">{{ s.chapter_count }}</td>
            <td class="px-4 py-3 text-gray-400 text-sm">{{ (s.view_count ?? 0).toLocaleString() }}</td>
            <td class="px-4 py-3">
              <div class="flex items-center justify-end gap-2">
                <button @click="openSync(s)" class="text-xs text-indigo-400 hover:text-indigo-300 transition-colors">
                  Sync
                </button>
                <button @click="rescrapeImages(s)" :disabled="rescraping === s.id" class="text-xs text-orange-400 hover:text-orange-300 transition-colors disabled:opacity-40">
                  {{ rescraping === s.id ? 'Fixing...' : 'Fix Images' }}
                </button>
                <button @click="openEdit(s)" class="text-xs text-green-400 hover:text-green-300 transition-colors">
                  Edit
                </button>
                <button @click="toggleStatus(s)" class="text-xs text-yellow-400 hover:text-yellow-300 transition-colors">
                  {{ s.status === 'ongoing' ? 'Mark Done' : 'Mark Ongoing' }}
                </button>
                <button @click="confirmDelete(s)" class="btn-danger">Delete</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="flex items-center justify-between">
      <span class="text-xs text-gray-500">
        {{ (page - 1) * perPage + 1 }}–{{ Math.min(page * perPage, total) }} of {{ total }} series
      </span>
      <div class="flex items-center gap-1">
        <button @click="goPage(1)" :disabled="page === 1"
          class="px-2 py-1 text-xs text-gray-400 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed">«</button>
        <button @click="goPage(page - 1)" :disabled="page === 1"
          class="px-3 py-1 text-xs text-gray-400 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed">‹ Prev</button>
        <span class="px-3 py-1 text-xs text-white bg-white/10 rounded">{{ page }} / {{ totalPages }}</span>
        <button @click="goPage(page + 1)" :disabled="page === totalPages"
          class="px-3 py-1 text-xs text-gray-400 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed">Next ›</button>
        <button @click="goPage(totalPages)" :disabled="page === totalPages"
          class="px-2 py-1 text-xs text-gray-400 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed">»</button>
      </div>
    </div>

    <!-- Edit Modal -->
    <div v-if="editTarget" class="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4" @click.self="editTarget = null">
      <div class="bg-[#1a1a24] border border-white/10 rounded-2xl p-6 max-w-lg w-full space-y-4">
        <h3 class="text-white font-bold">Edit Series</h3>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="text-xs text-gray-500 mb-1 block">Title</label>
            <input v-model="editForm.title" type="text" class="form-input w-full"/>
          </div>
          <div>
            <label class="text-xs text-gray-500 mb-1 block">Language</label>
            <select v-model="editForm.language" class="form-input w-full">
              <option value="en">English (EN)</option>
              <option value="my">Myanmar (MY)</option>
            </select>
          </div>
        </div>
        <div>
          <label class="text-xs text-gray-500 mb-1 block">Author</label>
          <input v-model="editForm.author" type="text" class="form-input w-full"/>
        </div>
        <div>
          <label class="text-xs text-gray-500 mb-1 block">Genres <span class="text-gray-600">(comma-separated, e.g. Action, Romance)</span></label>
          <input v-model="editForm.genres" type="text" class="form-input w-full" placeholder="Action, Romance, Fantasy"/>
        </div>
        <div>
          <label class="text-xs text-gray-500 mb-1 block">Description</label>
          <textarea v-model="editForm.description" rows="3" class="form-input w-full resize-none"/>
        </div>
        <div>
          <label class="text-xs text-gray-500 mb-1 block">Source URL <span class="text-gray-600">(used for auto-import)</span></label>
          <input v-model="editForm.source_url" type="text" class="form-input w-full" placeholder="https://adult.manhwamyanmar.com/manga/..."/>
        </div>
        <div v-if="editError" class="text-red-400 text-xs">{{ editError }}</div>
        <div class="flex gap-3 pt-1">
          <button @click="editTarget = null" class="flex-1 py-2 border border-white/10 text-gray-400 rounded-lg hover:bg-white/5 text-sm">Cancel</button>
          <button @click="doEdit" :disabled="editSaving" class="flex-1 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50">
            {{ editSaving ? 'Saving...' : 'Save' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Delete confirm -->
    <div v-if="deleteTarget" class="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4" @click.self="deleteTarget = null">
      <div class="bg-[#1a1a24] border border-white/10 rounded-2xl p-6 max-w-sm w-full">
        <h3 class="text-white font-bold mb-2">Delete Series?</h3>
        <p class="text-gray-400 text-sm mb-5">Delete "<strong class="text-white">{{ deleteTarget.title }}</strong>" and all its chapters? This cannot be undone.</p>
        <div class="flex gap-3">
          <button @click="deleteTarget = null" class="flex-1 py-2 border border-white/10 text-gray-400 rounded-lg hover:bg-white/5 text-sm">Cancel</button>
          <button @click="doDelete" :disabled="deleting" class="flex-1 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50">
            {{ deleting ? 'Deleting...' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Sync Chapters Modal -->
    <div v-if="syncTarget" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4" @click.self="closeSync">
      <div class="bg-[#1a1a24] border border-white/10 rounded-2xl w-full max-w-lg max-h-[80vh] flex flex-col">
        <!-- Header -->
        <div class="flex items-center justify-between px-5 py-4 border-b border-white/10">
          <div>
            <h3 class="text-white font-bold">Sync Chapters</h3>
            <p class="text-gray-500 text-xs mt-0.5 truncate max-w-xs">{{ syncTarget.title }}</p>
          </div>
          <button @click="closeSync" class="text-gray-500 hover:text-white">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M6 18L18 6M6 6l12 12"/></svg>
          </button>
        </div>

        <!-- Loading preview -->
        <div v-if="syncLoading" class="flex items-center justify-center py-16 text-gray-500">
          <svg class="w-5 h-5 animate-spin mr-2" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
          Fetching chapter list...
        </div>

        <!-- Error -->
        <div v-else-if="syncError" class="p-5 text-red-400 text-sm">{{ syncError }}</div>

        <!-- Chapter list -->
        <div v-else-if="previewChapters.length > 0" class="flex flex-col flex-1 min-h-0">
          <!-- Select all bar -->
          <div class="flex items-center gap-3 px-5 py-3 border-b border-white/5 bg-white/5">
            <input type="checkbox" :checked="allNewSelected" @change="toggleAll" class="rounded accent-indigo-500"/>
            <span class="text-xs text-gray-400">
              {{ selectedSlugs.size }} selected
              <span class="text-gray-600 ml-1">({{ newChapters.length }} new, {{ existingCount }} already imported)</span>
            </span>
            <div class="ml-auto flex items-center gap-2">
              <label class="flex items-center gap-1.5 text-xs text-gray-400 cursor-pointer">
                <input type="checkbox" v-model="syncImages" class="rounded accent-indigo-500"/>
                Scrape images
              </label>
            </div>
          </div>

          <!-- Scrollable chapter rows -->
          <div class="overflow-y-auto flex-1 divide-y divide-white/5">
            <label v-for="ch in previewChapters" :key="ch.slug"
              :class="['flex items-center gap-3 px-5 py-2.5 cursor-pointer hover:bg-white/5 transition-colors', ch.exists ? 'opacity-40' : '']">
              <input type="checkbox"
                :checked="selectedSlugs.has(ch.slug)"
                :disabled="ch.exists"
                @change="toggleChapter(ch.slug)"
                class="rounded accent-indigo-500 flex-shrink-0"/>
              <span class="text-sm text-gray-200 flex-1 truncate">{{ ch.title || `Chapter ${ch.number}` }}</span>
              <span class="text-xs text-gray-600 flex-shrink-0">Ch.{{ ch.number }}</span>
              <span v-if="ch.exists" class="text-2xs text-green-600 flex-shrink-0">imported</span>
            </label>
          </div>

          <!-- Footer -->
          <div class="px-5 py-4 border-t border-white/10 flex items-center gap-3">
            <button @click="closeSync" class="flex-1 py-2 border border-white/10 text-gray-400 rounded-lg hover:bg-white/5 text-sm">Cancel</button>
            <button @click="doSync" :disabled="syncSaving || selectedSlugs.size === 0"
              class="flex-1 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50">
              {{ syncSaving ? 'Importing...' : `Import ${selectedSlugs.size} chapter${selectedSlugs.size !== 1 ? 's' : ''}` }}
            </button>
          </div>
        </div>

        <div v-else class="p-5 text-gray-500 text-sm">No chapters found at source URL.</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import api from '@/services/api'

// Series list — server-side paginated
const series = ref<any[]>([])
const loading = ref(true)
const search = ref('')
const page = ref(1)
const totalPages = ref(1)
const total = ref(0)
const perPage = 15

async function load() {
  loading.value = true
  try {
    const res = await api.get('/admin/series', { params: { page: page.value, limit: perPage, search: search.value } })
    series.value = res.data.data
    totalPages.value = res.data.total_pages ?? 1
    total.value = res.data.total ?? 0
  } catch {} finally {
    loading.value = false
  }
}

watch(search, () => { page.value = 1; load() })

async function goPage(p: number) {
  page.value = p
  await load()
}

// Delete
const deleteTarget = ref<any>(null)
const deleting = ref(false)
const deletingEmpty = ref(false)
const deletingOrphans = ref(false)

function pollOrphanStatus() {
  const timer = setInterval(async () => {
    try {
      const res = await api.get('/admin/chapters/orphaned/status')
      if (!res.data.running) { clearInterval(timer); deletingOrphans.value = false }
    } catch { clearInterval(timer); deletingOrphans.value = false }
  }, 5000)
}

async function deleteOrphans() {
  if (!confirm('Delete all orphaned chapters? This runs in the background and takes a few minutes.')) return
  deletingOrphans.value = true
  try {
    const res = await api.delete('/admin/chapters/orphaned')
    alert(res.data.message || 'Orphan cleanup started in background.')
    pollOrphanStatus()
  } catch (e: any) {
    const msg = e.response?.data?.error || e.message
    alert(msg)
    if (e.response?.status !== 409) deletingOrphans.value = false
    else pollOrphanStatus()
  }
}

async function deleteEmpty() {
  if (!confirm('Delete all series with 0 chapters?')) return
  deletingEmpty.value = true
  try {
    const res = await api.delete('/admin/series/empty')
    alert(`Deleted ${res.data.deleted} empty series.`)
    page.value = 1
    await load()
  } catch (e: any) {
    alert('Failed: ' + (e.response?.data?.error || e.message))
  } finally {
    deletingEmpty.value = false
  }
}

// Edit
const editTarget = ref<any>(null)
const editSaving = ref(false)
const editError = ref('')
const editForm = ref({ title: '', language: 'en', author: '', genres: '', description: '', source_url: '' })

function openEdit(s: any) {
  editTarget.value = s
  editForm.value = { title: s.title || '', language: s.language || 'en', author: s.author || '', genres: s.genres || '', description: s.description || '', source_url: s.source_url || '' }
  editError.value = ''
}

async function doEdit() {
  if (!editTarget.value) return
  editSaving.value = true
  editError.value = ''
  try {
    await api.put(`/admin/series/${editTarget.value.id}`, editForm.value)
    Object.assign(editTarget.value, editForm.value)
    editTarget.value = null
  } catch (e: any) {
    editError.value = e.response?.data?.error || e.message || 'Save failed'
  } finally {
    editSaving.value = false
  }
}

// Rescrape
const rescraping = ref<string | null>(null)
async function rescrapeImages(s: any) {
  if (!confirm(`Re-scrape all chapter images for "${s.title}"? This may take a while.`)) return
  rescraping.value = s.id
  try {
    const res = await api.post('/admin/import/rescrape', { series_id: s.id })
    alert(`✓ Done: ${res.data.updated} chapters updated, ${res.data.failed} failed.`)
  } catch (e: any) {
    alert('Failed: ' + (e.response?.data?.error || e.message))
  } finally {
    rescraping.value = null
  }
}

// Sync
const syncTarget = ref<any>(null)
const syncLoading = ref(false)
const syncSaving = ref(false)
const syncError = ref('')
const syncImages = ref(false)
const previewChapters = ref<any[]>([])
const selectedSlugs = ref(new Set<string>())
const newChapters = computed(() => previewChapters.value.filter(c => !c.exists))
const existingCount = computed(() => previewChapters.value.filter(c => c.exists).length)
const allNewSelected = computed(() => newChapters.value.length > 0 && newChapters.value.every(c => selectedSlugs.value.has(c.slug)))

function toggleChapter(slug: string) {
  const set = new Set(selectedSlugs.value)
  if (set.has(slug)) set.delete(slug)
  else set.add(slug)
  selectedSlugs.value = set
}
function toggleAll() {
  selectedSlugs.value = allNewSelected.value ? new Set() : new Set(newChapters.value.map(c => c.slug))
}
async function openSync(s: any) {
  syncTarget.value = s; syncLoading.value = true; syncError.value = ''
  previewChapters.value = []; selectedSlugs.value = new Set()
  try {
    const res = await api.post('/admin/import/preview', { series_id: s.id })
    previewChapters.value = res.data.chapters
    selectedSlugs.value = new Set(res.data.chapters.filter((c: any) => !c.exists).map((c: any) => c.slug))
  } catch (e: any) {
    syncError.value = e.response?.data?.error || e.message || 'Failed to fetch chapters'
  } finally { syncLoading.value = false }
}
function closeSync() {
  syncTarget.value = null; previewChapters.value = []; selectedSlugs.value = new Set(); syncError.value = ''
}
async function doSync() {
  if (!syncTarget.value || selectedSlugs.value.size === 0) return
  syncSaving.value = true
  try {
    const res = await api.post('/admin/import/selected', {
      series_id: syncTarget.value.id, slugs: Array.from(selectedSlugs.value), scrape_images: syncImages.value,
    })
    const saved = res.data.chapters_saved
    const s = series.value.find(x => x.id === syncTarget.value.id)
    if (s) s.chapter_count += saved
    closeSync()
    alert(`✓ ${saved} chapter${saved !== 1 ? 's' : ''} imported.`)
  } catch (e: any) {
    syncError.value = e.response?.data?.error || e.message || 'Import failed'
  } finally { syncSaving.value = false }
}

async function toggleStatus(s: any) {
  const newStatus = s.status === 'ongoing' ? 'completed' : 'ongoing'
  try { await api.put(`/admin/series/${s.id}`, { status: newStatus }); s.status = newStatus } catch {}
}

function confirmDelete(s: any) { deleteTarget.value = s }
async function doDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await api.delete(`/admin/series/${deleteTarget.value.id}`)
    series.value = series.value.filter(s => s.id !== deleteTarget.value.id)
    total.value--
    deleteTarget.value = null
  } catch {} finally { deleting.value = false }
}

function imgError(e: Event) { (e.target as HTMLImageElement).style.display = 'none' }

onMounted(load)
</script>
