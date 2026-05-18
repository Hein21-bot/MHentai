import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api',
  headers: { 'X-Admin-Token': 'admin123' },
  timeout: 600000, // 10 min — imports with images can be slow
})

// Allow updating token dynamically
api.interceptors.request.use(config => {
  const token = localStorage.getItem('admin_token') || 'admin123'
  config.headers['X-Admin-Token'] = token
  return config
})

export default api

export const adminApi = {
  uploadSeriesCover: (seriesId: string, formData: FormData) => {
    formData.append('series_id', seriesId)
    return api.post('/admin/upload/series-cover', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  uploadChapterImages: (chapterId: string, formData: FormData) => {
    formData.append('chapter_id', chapterId)
    return api.post('/admin/upload/chapter-images', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
}
