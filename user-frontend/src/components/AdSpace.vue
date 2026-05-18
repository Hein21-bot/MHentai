<template>
  <div :class="adContainerClass">
    <!-- Placeholder for ad network integration (Google AdSense, PropellerAds, etc.) -->
    <div class="flex items-center justify-center w-full h-full">
      <span class="text-2xs text-gray-400 dark:text-gray-600 tracking-widest uppercase">Advertisement</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  type?: 'leaderboard' | 'banner' | 'sidebar' | 'native' | 'popunder' | 'smartlink' | 'socialbar'
  dark?: boolean
  bordered?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  type: 'banner',
  dark: false,
  bordered: true
})

const adContainerClass = computed(() => {
  const base = 'rounded-lg flex items-center justify-center bg-gray-100 dark:bg-dark-card'
  const border = props.bordered ? 'border border-dashed border-gray-300 dark:border-dark-border' : ''

  let dimensions = ''
  switch (props.type) {
    case 'leaderboard':
      // 728x90
      dimensions = 'w-full h-[90px] sm:h-[90px]'
      break
    case 'sidebar':
      // 300x600 or 160x600
      dimensions = 'w-full h-[300px] sm:w-[300px] sm:h-[600px]'
      break
    case 'native':
      // Native banner format
      dimensions = 'w-full h-20'
      break
    case 'popunder':
      // Popunder format (usually auto-sized)
      dimensions = 'w-full h-[400px]'
      break
    case 'smartlink':
      // Smart link format
      dimensions = 'w-full h-16'
      break
    case 'socialbar':
      // Social bar format
      dimensions = 'w-full h-12'
      break
    case 'banner':
    default:
      // Default banner (16:4 aspect ratio)
      dimensions = 'w-full h-16'
      break
  }

  return `${base} ${border} ${dimensions}`
})
</script>

<style scoped>
/* Placeholder styles for ad spaces */
/* These will be replaced with actual ad network code */
</style>
