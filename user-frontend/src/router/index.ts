import { createRouter, createWebHistory } from 'vue-router'

export default createRouter({
  history: createWebHistory(),
  scrollBehavior: () => ({ top: 0 }),
  routes: [
    { path: '/', redirect: '/my' },
    { path: '/en', name: 'home-en', component: () => import('@/pages/HomePage.vue'), meta: { lang: 'en' } },
    { path: '/en/series/:slug', name: 'series-en', component: () => import('@/pages/SeriesPage.vue'), meta: { lang: 'en' } },
    { path: '/en/read/:slug', name: 'read-en', component: () => import('@/pages/ChapterPage.vue'), meta: { lang: 'en' } },
    { path: '/en/az', name: 'az-en', component: () => import('@/pages/AZListPage.vue'), meta: { lang: 'en' } },
    { path: '/en/genres', name: 'genres-en', component: () => import('@/pages/GenresPage.vue'), meta: { lang: 'en' } },
    { path: '/en/manga-list', name: 'manga-list-en', component: () => import('@/pages/MangaListPage.vue'), meta: { lang: 'en' } },
    { path: '/my', name: 'home-my', component: () => import('@/pages/HomePage.vue'), meta: { lang: 'my' } },
    { path: '/my/series/:slug', name: 'series-my', component: () => import('@/pages/SeriesPage.vue'), meta: { lang: 'my' } },
    { path: '/my/read/:slug', name: 'read-my', component: () => import('@/pages/ChapterPage.vue'), meta: { lang: 'my' } },
    { path: '/my/az', name: 'az-my', component: () => import('@/pages/AZListPage.vue'), meta: { lang: 'my' } },
    { path: '/my/genres', name: 'genres-my', component: () => import('@/pages/GenresPage.vue'), meta: { lang: 'my' } },
    { path: '/my/manga-list', name: 'manga-list-my', component: () => import('@/pages/MangaListPage.vue'), meta: { lang: 'my' } },
    { path: '/:pathMatch(.*)*', redirect: '/my' },
  ]
})
