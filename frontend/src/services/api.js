import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

// Add token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Handle 401 errors (token expired/invalid)
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      if (globalThis.location.pathname !== '/login') {
        globalThis.location.href = '/login'
      }
    }
    return Promise.reject(error)
  }
)

// Dashboard APIs
export const getDashboardStats = (params = {}) => {
  // Support both old range format and new custom date format
  if (typeof params === 'string') {
    // Legacy support: range as string
    return api.get('/dashboard/stats', { params: { range: params } })
  }
  // New format: object with start and end dates
  return api.get('/dashboard/stats', { params })
}

export const getTrafficLogs = (limit = 100, offset = 0) =>
  api.get('/dashboard/traffic', { params: { limit, offset } })

export const getAttacksByCountry = (params = {}) =>
  api.get('/dashboard/attacks-by-country', { params })

// VHost APIs
export const getVHosts = () => api.get('/vhosts')
export const getVHost = (id) => api.get(`/vhosts/${id}`)
export const createVHost = (data) => api.post('/vhosts', data)
export const updateVHost = (id, data) => api.put(`/vhosts/${id}`, data)
export const deleteVHost = (id) => api.delete(`/vhosts/${id}`)

// IP Group APIs
export const getIPGroups = () => api.get('/ip-groups')
export const getIPGroup = (id) => api.get(`/ip-groups/${id}`)
export const createIPGroup = (data) => api.post('/ip-groups', data)
export const deleteIPGroup = (id) => api.delete(`/ip-groups/${id}`)
export const addIPToGroup = (groupId, data) => api.post(`/ip-groups/${groupId}/addresses`, data)
export const getGroupIPs = (groupId) => api.get(`/ip-groups/${groupId}/addresses`)
export const removeIPFromGroup = (groupId, ipId) => api.delete(`/ip-groups/${groupId}/addresses/${ipId}`)

// Settings APIs (public endpoint for login page)
export const getAppSettings = () => api.get('/settings/app')
export const getTurnstileSiteKey = () => api.get('/turnstile/sitekey')

export default api
