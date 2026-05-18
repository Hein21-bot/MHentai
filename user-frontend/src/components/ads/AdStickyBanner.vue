<template>
  <!-- Sticky bottom banner 320x50 — sandboxed in its own iframe -->
  <div class="fixed bottom-0 left-0 right-0 z-50 flex justify-center">
    <iframe ref="adFrame" width="320" height="50" frameborder="0" scrolling="no" style="border:none;display:block;"></iframe>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const adFrame = ref<HTMLIFrameElement | null>(null)

const html = `<!DOCTYPE html><html><head>
<script>atOptions={'key':'9d4fc4b02c8c2bfd42fa097dce2ce6f1','format':'iframe','height':50,'width':320,'params':{}}<\/script>
<script data-cfasync="false" src="https://www.highperformanceformat.com/9d4fc4b02c8c2bfd42fa097dce2ce6f1/invoke.js"><\/script>
</head><body style="margin:0;padding:0;overflow:hidden;"></body></html>`

onMounted(() => {
  if (adFrame.value) adFrame.value.srcdoc = html
})
</script>
