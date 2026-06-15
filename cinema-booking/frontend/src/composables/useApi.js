import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '',
  timeout: 10000,
})

// Inject token into every request
api.interceptors.request.use(config => {
  const token = localStorage.getItem('cinema_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Handle 401 — only redirect if it's a protected endpoint
api.interceptors.response.use(
  res => res,
  err => {
    const url = err.config?.url || ''
    const isPublic = url.includes('/api/showtimes') || url.includes('/api/auth')

    if (err.response?.status === 401 && !isPublic) {
      localStorage.removeItem('cinema_token')
      localStorage.removeItem('cinema_user')
      window.location.href = '/'
    }
    return Promise.reject(err)
  }
)

export default api
