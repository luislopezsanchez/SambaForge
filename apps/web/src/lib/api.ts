import axios from 'axios'

const API = axios.create({
  baseURL: '/api',
  timeout: 30000,
})

// Attach JWT token to every request
API.interceptors.request.use((config) => {
  const token = localStorage.getItem('sambaforge_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Handle 401 globally
API.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('sambaforge_token')
      localStorage.removeItem('sambaforge_user')
      window.location.href = '/'
    }
    return Promise.reject(error)
  }
)

export default API