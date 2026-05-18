<template>
  <RouterLink :to="`/${$route.meta.lang}/series/${series.slug}`" class="group block">
    <div class="relative aspect-[2/3] bg-gray-200 rounded-lg overflow-hidden dark:bg-dark-card">
      <img
        v-if="series.cover_url"
        :src="series.cover_url"
        :alt="series.title"
        loading="lazy"
        class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
        @error="imgError"
      />
      <div v-else class="w-full h-full flex items-center justify-center text-gray-300 dark:text-gray-700">
        <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/>
        </svg>
      </div>

      <!-- Status badge -->
      <span :class="['absolute top-1 left-1 text-[9px] font-bold px-1 py-0.5 rounded leading-none',
        series.status === 'ongoing' ? 'bg-green-500 text-white' : 'bg-blue-500 text-white']">
        {{ series.status === 'ongoing' ? 'ON' : 'END' }}
      </span>

      <!-- Bottom gradient overlay -->
      <div class="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/85 via-black/40 to-transparent pt-8 pb-1.5 px-1.5">
        <h3 class="text-white text-[10px] font-semibold line-clamp-2 leading-tight">{{ series.title }}</h3>
        <div class="flex items-center justify-between mt-0.5">
          <span class="text-yellow-400 text-[9px] tracking-tight">{{ starDisplay(series) }}</span>
          <span class="text-white/50 text-[9px]">Ch.{{ series.chapter_count }}</span>
        </div>
      </div>
    </div>
  </RouterLink>
</template>

<script setup lang="ts">
import type { Series } from '@/services/api'

const props = defineProps<{ series: Series }>()

function imgError(e: Event) { (e.target as HTMLImageElement).style.display = 'none' }

function getStars(series: Series): number {
  let base = 3.0
  if (series.view_count > 0) base = Math.min(4.5, 3.0 + Math.log10(series.view_count + 1) * 0.5)
  let hash = 0
  for (const c of series.id) hash = (hash * 31 + c.charCodeAt(0)) & 0xFF
  return Math.min(5.0, parseFloat((base + (hash % 6) / 10).toFixed(1)))
}

function starDisplay(series: Series): string {
  const rating = getStars(series)
  const full = Math.floor(rating)
  const half = rating - full >= 0.5 ? 1 : 0
  const empty = 5 - full - half
  return '★'.repeat(full) + (half ? '½' : '') + '☆'.repeat(empty) + ` ${rating.toFixed(1)}`
}
</script>
