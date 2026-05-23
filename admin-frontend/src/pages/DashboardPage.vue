<template>
  <div class="space-y-6">
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="admin-card">
        <p class="text-gray-500 text-xs mb-1">Total Series</p>
        <p class="text-3xl font-black text-white">{{ stats?.total_series ?? '—' }}</p>
      </div>
      <div class="admin-card">
        <p class="text-gray-500 text-xs mb-1">Total Chapters</p>
        <p class="text-3xl font-black text-white">{{ stats?.total_chapters ?? '—' }}</p>
      </div>
      <div class="admin-card">
        <p class="text-gray-500 text-xs mb-1">Total Images</p>
        <p class="text-3xl font-black text-white">{{ stats?.total_images != null ? stats.total_images.toLocaleString() : '—' }}</p>
      </div>
      <div class="admin-card">
        <p class="text-gray-500 text-xs mb-1">Total Views</p>
        <p class="text-3xl font-black text-white">{{ stats?.total_views != null ? stats.total_views.toLocaleString() : '—' }}</p>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <div class="admin-card">
        <h3 class="text-white font-semibold mb-4">Recently Added</h3>
        <div v-if="recent.length === 0" class="text-gray-600 text-sm">No series yet.</div>
        <div class="space-y-3">
          <div v-for="s in recent" :key="s.id" class="flex items-center gap-3">
            <img v-if="s.cover_url" :src="s.cover_url" :alt="s.title" class="w-10 h-14 object-cover rounded" @error="imgError"/>
            <div v-else class="w-10 h-14 bg-[#12121a] rounded flex-shrink-0"/>
            <div class="flex-1 min-w-0">
              <p class="text-sm text-white font-medium truncate">{{ s.title }}</p>
              <p class="text-xs text-gray-600">{{ s.chapter_count }} chapters · {{ s.status }}</p>
            </div>
          </div>
        </div>
      </div>

      <div class="admin-card">
        <h3 class="text-white font-semibold mb-4">Quick Actions</h3>
        <div class="space-y-2">
          <RouterLink to="/import" class="flex items-center gap-3 p-3 bg-indigo-600/10 hover:bg-indigo-600/20 border border-indigo-600/30 rounded-lg transition-colors group">
            <svg class="w-5 h-5 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"/></svg>
            <div>
              <p class="text-white text-sm font-medium">Import from manhwamyanmar.com</p>
              <p class="text-gray-500 text-xs">Scrape and save series + chapters</p>
            </div>
          </RouterLink>
          <RouterLink to="/series" class="flex items-center gap-3 p-3 bg-white/5 hover:bg-white/10 border border-white/10 rounded-lg transition-colors">
            <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/></svg>
            <div>
              <p class="text-white text-sm font-medium">Manage Series</p>
              <p class="text-gray-500 text-xs">Edit, delete, update status</p>
            </div>
          </RouterLink>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/services/api'

const stats = ref<any>(null)
const recent = ref<any[]>([])
function imgError(e: Event) { (e.target as HTMLImageElement).style.display = 'none' }

onMounted(async () => {
  try {
    const [s, r] = await Promise.all([
      api.get('/admin/stats'),
      api.get('/admin/recent'),
    ])
    stats.value = s.data
    recent.value = r.data.data
  } catch {}
})
</script>
