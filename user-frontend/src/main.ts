import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import './style.css'

const savedTheme = localStorage.getItem('theme')
const prefersDark = window.matchMedia?.('(prefers-color-scheme: dark)').matches
document.documentElement.classList.toggle('dark', savedTheme ? savedTheme === 'dark' : prefersDark)

createApp(App).use(createPinia()).use(router).mount('#app')
