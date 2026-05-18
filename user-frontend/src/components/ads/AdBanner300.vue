<template>
  <!-- Banner 300x250 — sandboxed in its own iframe -->
  <div class="flex justify-center">
    <iframe ref="adFrame" width="300" height="250" frameborder="0" scrolling="no" style="border:none;display:block;"></iframe>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const adFrame = ref<HTMLIFrameElement | null>(null)

const html = `<!DOCTYPE html><html><head>
<script>atOptions={'key':'2a0c2178a2b02c1fdb253d3c04bb9ad6','format':'iframe','height':250,'width':300,'params':{}}<\/script>
<script data-cfasync="false" src="https://www.highperformanceformat.com/2a0c2178a2b02c1fdb253d3c04bb9ad6/invoke.js"><\/script>
</head><body style="margin:0;padding:0;overflow:hidden;"></body></html>`

onMounted(() => {
  if (adFrame.value) adFrame.value.srcdoc = html
})
</script>
