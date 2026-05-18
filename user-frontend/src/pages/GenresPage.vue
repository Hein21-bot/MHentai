<template>
  <div class="space-y-5">
    <h1 class="text-xl font-extrabold text-gray-950 dark:text-white">Genres</h1>

    <!-- Genre tags -->
    <div class="flex flex-wrap gap-2">
      <button v-for="g in allGenres" :key="g" @click="activeGenre = activeGenre === g ? '' : g"
        :class="['px-3 py-1.5 text-xs font-semibold rounded-full border transition-colors',
          activeGenre === g
            ? 'bg-primary border-primary text-white'
            : 'border-gray-200 text-gray-600 hover:border-primary hover:text-primary dark:border-dark-border dark:text-gray-400 dark:hover:border-primary dark:hover:text-primary']">
        {{ g }}
      </button>
    </div>

    <!-- Loading genres -->
    <div v-if="loading" class="flex flex-wrap gap-2">
      <div v-for="i in 16" :key="i" class="h-7 w-20 bg-gray-200 rounded-full animate-pulse dark:bg-dark-card"/>
    </div>

    <!-- No genre selected -->
    <div v-else-if="!activeGenre" class="py-16 text-center text-gray-400 dark:text-gray-600 text-sm">
      Select a genre to browse series
    </div>

    <!-- Results -->
    <template v-else>
      <p class="text-sm text-gray-500 dark:text-gray-500">
        <span class="font-semibold text-gray-900 dark:text-white">{{ filtered.length }}</span> series in
        <span class="text-primary font-semibold">{{ activeGenre }}</span>
      </p>
      <div class="grid grid-cols-3 gap-2 sm:grid-cols-4 sm:gap-3 md:grid-cols-5">
        <RouterLink v-for="s in filtered" :key="s.id"
          :to="`/${route.meta.lang}/series/${s.slug}`" class="group block">
          <div class="relative aspect-[2/3] rounded-xl overflow-hidden bg-gray-200 dark:bg-dark-card">
            <img v-if="s.cover_url" :src="s.cover_url" :alt="s.title"
              class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-200"
              @error="imgError"/>
            <div class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/10 to-transparent"/>
            <span :class="['absolute top-1.5 left-1.5 text-[9px] font-bold px-1.5 py-0.5 rounded-full leading-none',
              s.status === 'ongoing' ? 'bg-green-500 text-white' : 'bg-blue-500 text-white']">
              {{ s.status === 'ongoing' ? 'ON' : 'END' }}
            </span>
            <div class="absolute bottom-0 left-0 right-0 p-2">
              <p class="text-white text-[10px] font-semibold line-clamp-2 leading-snug">{{ s.title }}</p>
              <p class="text-white/50 text-[9px] mt-0.5">Ch.{{ s.chapter_count }}</p>
            </div>
          </div>
        </RouterLink>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { seriesApi } from '@/services/api'
import type { Series } from '@/services/api'

const route = useRoute()
const allSeries = ref<Series[]>([])
const loading = ref(true)
const activeGenre = ref('')

const allGenres = computed(() => {
  const set = new Set<string>()
  for (const s of allSeries.value) {
    if (s.genres) s.genres.split(',').forEach(g => { const t = g.trim(); if (t) set.add(t) })
  }
  return Array.from(set).sort()
})

const filtered = computed(() => {
  if (!activeGenre.value) return []
  return allSeries.value.filter(s =>
    s.genres?.split(',').map(g => g.trim()).includes(activeGenre.value)
  )
})

function imgError(e: Event) { (e.target as HTMLImageElement).style.display = 'none' }

async function load() {
  loading.value = true
  activeGenre.value = ''
  try {
    const res = await seriesApi.list({ sort: 'title', limit: 500, lang: route.meta.lang })
    allSeries.value = res.data.data
  } catch {
    allSeries.value = []
  } finally {
    loading.value = false
  }
}

watch(() => route.meta.lang, load)
onMounted(load)
</script>
