<template>
  <!-- Adsterra Native Banner -->
  <!-- Only one instance can render per page (Adsterra hardcodes the container ID) -->
  <div v-if="isOwner" id="container-3aedf555ae8e8eae2ec0937e7fde1f12" class="w-full"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

// Module-level singleton — tracks which instance currently owns the slot
let activeInstance: symbol | null = null

const SCRIPT_ID = 'adsterra-native'
const SCRIPT_SRC = 'https://pl29484625.effectivecpmnetwork.com/3aedf555ae8e8eae2ec0937e7fde1f12/invoke.js'

const id = Symbol()
const isOwner = ref(false)

onMounted(() => {
  if (activeInstance !== null) return // another instance already active, skip
  activeInstance = id
  isOwner.value = true

  document.getElementById(SCRIPT_ID)?.remove()
  const script = document.createElement('script')
  script.id = SCRIPT_ID
  script.async = true
  script.setAttribute('data-cfasync', 'false')
  script.src = SCRIPT_SRC
  document.head.appendChild(script)
})

onUnmounted(() => {
  if (activeInstance === id) {
    activeInstance = null
    isOwner.value = false
    document.getElementById(SCRIPT_ID)?.remove()
  }
})
</script>
