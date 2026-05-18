import { createRouter, createWebHistory } from 'vue-router'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/dashboard' },
    { path: '/dashboard', name: 'dashboard', component: () => import('@/pages/DashboardPage.vue') },
    { path: '/series', name: 'series', component: () => import('@/pages/SeriesListPage.vue') },
    { path: '/chapters', name: 'chapters', component: () => import('@/pages/ChaptersPage.vue') },
    { path: '/import', name: 'import', component: () => import('@/pages/ImportPage.vue') },
    { path: '/upload', name: 'upload', component: () => import('@/pages/UploadPage.vue') },
  ]
})
