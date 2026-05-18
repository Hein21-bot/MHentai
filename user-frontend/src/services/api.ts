import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api',
  timeout: 15000,
})

export interface Series {
  id: string
  slug: string
  title: string
  cover_url: string
  description: string
  status: 'ongoing' | 'completed'
  author: string
  genres: string
  view_count: number
  chapter_count: number
  chapters?: Chapter[]
  updated_at: string
}

export interface Chapter {
  id: string
  series_id: string
  slug: string
  title: string
  number: number
  view_count: number
  images?: string[]
  series?: Series
  source_url?: string
  created_at: string
  updated_at: string
}

export const seriesApi = {
  list: (params?: Record<string, unknown>) => api.get<{ data: Series[]; total: number; page: number }>('/series', { params }),
  get: (slug: string) => api.get<Series>(`/series/${slug}`),
  latestChapters: (slug: string) => api.get<{ data: Chapter[] }>(`/series/${slug}/latest-chapters`),
  latest: (limit = 12, lang?: string, page = 1) => api.get<{ data: Chapter[]; total: number; page: number }>('/latest', { params: { limit, lang, page } }),
}

export const chapterApi = {
  get: (slug: string) => api.get<{ chapter: Chapter; prev_chapter?: Chapter; next_chapter?: Chapter }>(`/chapters/${slug}`),
}

export default api
