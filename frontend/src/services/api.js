import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

// Dashboard APIs
export const getDashboardStats = (range = '24h') => 
  api.get('/dashboard/stats', { params: { range } })

export const getTrafficLogs = (limit = 100, offset = 0) =>
  api.get('/dashboard/traffic', { params: { limit, offset } })

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
export const addIPToGroup = (groupId, data) => api.post(`/ip-groups/${groupId}/ips`, data)
export const getGroupIPs = (groupId) => api.get(`/ip-groups/${groupId}/ips`)
export const removeIPFromGroup = (groupId, ipId) => api.delete(`/ip-groups/${groupId}/ips/${ipId}`)

export default api
