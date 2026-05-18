<template>
  <footer class="border-t border-gray-200 dark:border-dark-border mt-12 bg-white dark:bg-dark-surface">
    <!-- Links row -->
    <div class="max-w-screen-xl mx-auto px-4 py-4 flex items-center justify-center gap-4 flex-wrap">
      <a href="#" class="text-xs text-gray-500 hover:text-primary dark:text-gray-500 dark:hover:text-primary transition-colors">Privacy Policy</a>
      <span class="text-gray-300 dark:text-gray-700">·</span>
      <a href="#" class="text-xs text-gray-500 hover:text-primary dark:text-gray-500 dark:hover:text-primary transition-colors">Terms of Service</a>
      <span class="text-gray-300 dark:text-gray-700">·</span>
      <a href="#" class="text-xs text-gray-500 hover:text-primary dark:text-gray-500 dark:hover:text-primary transition-colors">DMCA</a>
      <span class="text-gray-300 dark:text-gray-700">·</span>
      <a href="#" class="text-xs text-gray-500 hover:text-primary dark:text-gray-500 dark:hover:text-primary transition-colors">Contact</a>
    </div>

    <!-- Disclaimer -->
    <div class="bg-gray-50 dark:bg-dark-card border-t border-gray-100 dark:border-dark-border px-4 py-4">
      <p class="max-w-2xl mx-auto text-center text-[11px] text-gray-400 dark:text-gray-600 leading-relaxed">
        All the comics on this website are only previews of the original comics, there may be many language errors,
        character names, and story lines. For the original version, please buy the comic if it's available in your city.
      </p>
    </div>

    <!-- Copyright -->
    <div class="max-w-screen-xl mx-auto px-4 py-3 text-center">
      <p class="text-xs text-gray-400 dark:text-gray-600">© {{ year }} <span class="font-semibold text-gray-600 dark:text-gray-400">MHentai</span>. All rights reserved.</p>
    </div>
  </footer>
</template>

<script setup lang="ts">
const year = new Date().getFullYear()
</script>
