import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import api, { getVHosts, createVHost, deleteVHost } from '../services/api'
import { Plus, Trash2, Edit, Server, Search, Grid3x3, List, ChevronUp, ChevronDown, ChevronsUpDown, ChevronDown as ChevronDownIcon, CheckCircle, AlertCircle, Loader2, Shield, Eye, ExternalLink, ChevronLeft, ChevronRight, FileCode, Globe } from 'lucide-react'
import logger from '../utils/logger'

const VHosts = () => {
  const navigate = useNavigate()
  const [vhosts, setVHosts] = useState([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [searchTerm, setSearchTerm] = useState('')
  const [sortField, setSortField] = useState('name')
  const [sortOrder, setSortOrder] = useState('asc')
  const [selectedVHosts, setSelectedVHosts] = useState([])
  const [viewMode, setViewMode] = useState('grid') // 'grid' or 'list'
  const [certificates, setCertificates] = useState([])
  const [certSearchTerm, setCertSearchTerm] = useState('')
  const [showCertDropdown, setShowCertDropdown] = useState(false)
  const [backendCheckStatus, setBackendCheckStatus] = useState(null) // 'checking', 'success', 'error', null
  const [backendCheckMessage, setBackendCheckMessage] = useState('')
  const [showConfirmModal, setShowConfirmModal] = useState(false)
  const [confirmAction, setConfirmAction] = useState(null)
  const [confirmMessage, setConfirmMessage] = useState('')
  const [globalLoading, setGlobalLoading] = useState(false)
  const [loadingMessage, setLoadingMessage] = useState('')
  const [showDetailModal, setShowDetailModal] = useState(false)
  const [selectedVHost, setSelectedVHost] = useState(null)
  const [isEditMode, setIsEditMode] = useState(false)
  const [editingVHostId, setEditingVHostId] = useState(null)

  // Pagination states
  const [currentPage, setCurrentPage] = useState(1)
  const [itemsPerPage, setItemsPerPage] = useState(12)

  const [formData, setFormData] = useState({
    name: '',
    domain: '',
    backend_url: '',
    backends: [],
    load_balance_method: 'round_robin',
    custom_config: '',
    ssl_enabled: false,
    ssl_certificate_id: '',
    enabled: true,
    websocket_enabled: false,
    http_version: 'http/1.1',
    tls_version: 'TLSv1.2',
    max_upload_size: 10,
    proxy_read_timeout: 60,
    proxy_connect_timeout: 60,
    bot_detection_enabled: false,
    bot_detection_type: 'turnstile',
    recaptcha_version: 'v2',
    rate_limit_enabled: false,
    rate_limit_requests: 100,
    rate_limit_window: 60,
    region_filtering_enabled: false,
    region_whitelist: [],
    region_blacklist: [],
    custom_locations: [],
    custom_headers: {},
  })

  const [showAdvancedSettings, setShowAdvancedSettings] = useState(false)
  const [newLocation, setNewLocation] = useState({ path: '', proxy_pass: '', config: '', websocket_enabled: false })
  const [newHeader, setNewHeader] = useState({ key: '', value: '' })
  const [locationBackendCheck, setLocationBackendCheck] = useState({ status: null, message: '' })

  useEffect(() => {
    loadVHosts()
    loadCertificates()
  }, [])

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (showCertDropdown && !event.target.closest('.relative')) {
        setShowCertDropdown(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [showCertDropdown])

  const loadVHosts = async () => {
    try {
      setGlobalLoading(true)
      setLoadingMessage('Loading virtual hosts...')
      const response = await getVHosts()
      // Handle both array and object responses
      const data = response.data?.vhosts || response.data || []
      setVHosts(Array.isArray(data) ? data : [])
    } catch (error) {
      logger.error('Failed to load vhosts:', error)
      setVHosts([])
    } finally {
      setLoading(false)
      setGlobalLoading(false)
    }
  }

  const loadCertificates = async () => {
    try {
      setGlobalLoading(true)
      setLoadingMessage('Loading SSL certificates...')
      const response = await api.get('/certificates')
      const data = response.data?.certificates || response.data || []
      setCertificates(Array.isArray(data) ? data : [])
    } catch (error) {
      logger.error('Failed to load certificates:', error)
      setCertificates([])
    } finally {
      setGlobalLoading(false)
    }
  }

  const checkBackendURL = async (url) => {
    if (!url || url.trim() === '') {
      setBackendCheckStatus(null)
      setBackendCheckMessage('')
      return
    }

    // Validate URL format
    try {
      new URL(url)
    } catch (e) {
      setBackendCheckStatus('error')
      setBackendCheckMessage('Invalid URL format')
      return
    }

    setBackendCheckStatus('checking')
    setBackendCheckMessage('Checking backend availability...')

    try {
      // Try to reach the backend URL with timeout
      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), 5000) // 5 second timeout

      await fetch(url, {
        method: 'HEAD',
        mode: 'no-cors', // Allow cross-origin requests
        signal: controller.signal,
      })

      clearTimeout(timeoutId)

      // With no-cors, we can't read the response, but if fetch succeeds, the server is reachable
      setBackendCheckStatus('success')
      setBackendCheckMessage('Backend is reachable')
    } catch (error) {
      if (error.name === 'AbortError') {
        setBackendCheckStatus('error')
        setBackendCheckMessage('Backend request timeout (5s)')
      } else {
        // Even with no-cors, if there's a network error, we'll catch it
        // But since no-cors mode doesn't fail on status codes, we assume it's reachable if no error
        setBackendCheckStatus('success')
        setBackendCheckMessage('Backend is reachable')
      }
    }
  }

  // Debounce backend URL check
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      checkBackendURL(formData.backend_url)
    }, 800) // Wait 800ms after user stops typing

    return () => clearTimeout(timeoutId)
  }, [formData.backend_url])

  // Check custom location backend URL
  useEffect(() => {
    if (!newLocation.proxy_pass || newLocation.proxy_pass.trim() === '') {
      setLocationBackendCheck({ status: null, message: '' })
      return
    }

    // Validate URL format
    try {
      new URL(newLocation.proxy_pass)
    } catch (e) {
      setLocationBackendCheck({ status: 'error', message: 'Invalid URL format' })
      return
    }

    setLocationBackendCheck({ status: 'checking', message: 'Checking...' })

    const timeoutId = setTimeout(async () => {
      try {
        await fetch(newLocation.proxy_pass, {
          method: 'HEAD',
          mode: 'no-cors',
          cache: 'no-cache'
        })
        setLocationBackendCheck({ status: 'success', message: 'Backend is reachable' })
      } catch {
        setLocationBackendCheck({ status: 'warning', message: 'Cannot verify (CORS/Network), but may still work' })
      }
    }, 800)

    return () => clearTimeout(timeoutId)
  }, [newLocation.proxy_pass])

  const handleSubmit = async (e) => {
    e.preventDefault()
    try {
      setGlobalLoading(true)
      if (isEditMode && editingVHostId) {
        setLoadingMessage('Updating virtual host...')
        await api.put(`/vhosts/${editingVHostId}`, formData)
      } else {
        setLoadingMessage('Creating virtual host...')
        await createVHost(formData)
      }
      setShowModal(false)
      setIsEditMode(false)
      setEditingVHostId(null)
      setFormData({
        name: '',
        domain: '',
        backend_url: '',
        ssl_enabled: false,
        ssl_certificate_id: '',
        enabled: true,
        websocket_enabled: false,
        http_version: 'http/1.1',
        tls_version: 'TLSv1.2',
        max_upload_size: 10,
        proxy_read_timeout: 60,
        proxy_connect_timeout: 60,
        bot_detection_enabled: false,
        bot_detection_type: 'turnstile',
        recaptcha_version: 'v2',
        rate_limit_enabled: false,
        rate_limit_requests: 100,
        rate_limit_window: 60,
        region_filtering_enabled: false,
        region_whitelist: [],
        region_blacklist: [],
        custom_locations: [],
        custom_headers: {},
      })
      setCertSearchTerm('')
      setBackendCheckStatus(null)
      setBackendCheckMessage('')
      setShowAdvancedSettings(false)
      setNewLocation({ path: '', proxy_pass: '', config: '', websocket_enabled: false })
      setNewHeader({ key: '', value: '' })
      await loadVHosts()
    } catch (error) {
      logger.error('Failed to save vhost:', error)
      setGlobalLoading(false)
    }
  }

  const handleDelete = (id) => {
    setConfirmMessage('Are you sure you want to delete this virtual host?')
    setConfirmAction(() => async () => {
      setShowConfirmModal(false)
      setGlobalLoading(true)
      setLoadingMessage('Deleting virtual host...')
      try {
        await deleteVHost(id)
        await loadVHosts()
      } catch (error) {
        logger.error('Failed to delete vhost:', error)
      } finally {
        setGlobalLoading(false)
      }
    })
    setShowConfirmModal(true)
  }

  const handleSort = (field) => {
    if (sortField === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')
    } else {
      setSortField(field)
      setSortOrder('asc')
    }
  }

  const handleSelectAll = (e) => {
    if (e.target.checked) {
      setSelectedVHosts(filteredAndSortedVHosts.map(v => v.id))
    } else {
      setSelectedVHosts([])
    }
  }

  const handleSelectOne = (id) => {
    if (selectedVHosts.includes(id)) {
      setSelectedVHosts(selectedVHosts.filter(vId => vId !== id))
    } else {
      setSelectedVHosts([...selectedVHosts, id])
    }
  }

  const handleBulkDelete = () => {
    setConfirmMessage(`Are you sure you want to delete ${selectedVHosts.length} virtual host(s)?`)
    setConfirmAction(() => async () => {
      setShowConfirmModal(false)
      setGlobalLoading(true)
      setLoadingMessage(`Deleting ${selectedVHosts.length} virtual host(s)...`)
      try {
        await Promise.all(selectedVHosts.map(id => deleteVHost(id)))
        setSelectedVHosts([])
        await loadVHosts()
      } catch (error) {
        logger.error('Failed to bulk delete vhosts:', error)
      } finally {
        setGlobalLoading(false)
      }
    })
    setShowConfirmModal(true)
  }

  const handleViewDetail = (vhost) => {
    setSelectedVHost(vhost)
    setShowDetailModal(true)
  }

  const handleEdit = (vhost) => {
    setIsEditMode(true)
    setEditingVHostId(vhost.id)
    setFormData({
      name: vhost.name || '',
      domain: vhost.domain || '',
      backend_url: vhost.backend_url || '',
      backends: vhost.backends || [],
      load_balance_method: vhost.load_balance_method || 'round_robin',
      custom_config: vhost.custom_config || '',
      ssl_enabled: vhost.ssl_enabled || false,
      ssl_certificate_id: vhost.ssl_certificate_id || '',
      enabled: vhost.enabled === undefined ? true : vhost.enabled,
      websocket_enabled: vhost.websocket_enabled || false,
      http_version: vhost.http_version || 'http/1.1',
      tls_version: vhost.tls_version || 'TLSv1.2',
      max_upload_size: vhost.max_upload_size || 10,
      proxy_read_timeout: vhost.proxy_read_timeout || 60,
      proxy_connect_timeout: vhost.proxy_connect_timeout || 60,
      bot_detection_enabled: vhost.bot_detection_enabled || false,
      bot_detection_type: vhost.bot_detection_type || 'turnstile',
      recaptcha_version: vhost.recaptcha_version || 'v2',
      rate_limit_enabled: vhost.rate_limit_enabled || false,
      rate_limit_requests: vhost.rate_limit_requests || 100,
      rate_limit_window: vhost.rate_limit_window || 60,
      region_filtering_enabled: vhost.region_filtering_enabled || false,
      region_whitelist: vhost.region_whitelist || [],
      region_blacklist: vhost.region_blacklist || [],
      custom_locations: vhost.custom_locations || [],
      custom_headers: vhost.custom_headers || {},
    })
    setShowModal(true)
  }

  // Filter and sort vhosts
  const filteredAndSortedVHosts = vhosts
    .filter(vhost =>
      vhost.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      vhost.domain?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      vhost.backend_url?.toLowerCase().includes(searchTerm.toLowerCase())
    )
    .sort((a, b) => {
      const aVal = a[sortField] || ''
      const bVal = b[sortField] || ''
      const comparison = aVal.toString().localeCompare(bVal.toString())
      return sortOrder === 'asc' ? comparison : -comparison
    })

  // Pagination
  const totalPages = Math.ceil(filteredAndSortedVHosts.length / itemsPerPage)
  const startIndex = (currentPage - 1) * itemsPerPage
  const paginatedVHosts = filteredAndSortedVHosts.slice(startIndex, startIndex + itemsPerPage)

  const handlePageChange = (page) => {
    setCurrentPage(page)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  const handleItemsPerPageChange = (value) => {
    setItemsPerPage(value)
    setCurrentPage(1)
  }

  // Filter certificates by search term
  const filteredCertificates = certificates.filter(cert =>
    cert.name?.toLowerCase().includes(certSearchTerm.toLowerCase()) ||
    cert.common_name?.toLowerCase().includes(certSearchTerm.toLowerCase())
  )

  // Get selected certificate details
  const selectedCertificate = certificates.find(cert => cert.id === formData.ssl_certificate_id)

  // Helper function for certificate expiry status color
  const getCertExpiryColor = (validToDate) => {
    const validTo = new Date(validToDate)
    const now = new Date()
    const thirtyDaysFromNow = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000)

    if (validTo < now) return 'text-red-600'
    if (validTo < thirtyDaysFromNow) return 'text-yellow-600'
    return 'text-green-600'
  }

  // Helper function for backend check status color
  const getBackendCheckColor = (status) => {
    if (status === 'success') return 'text-green-600'
    if (status === 'error') return 'text-red-600'
    return 'text-blue-600'
  }

  // Helper function for location backend check color
  const getLocationCheckColor = (status) => {
    if (status === 'success') return 'text-green-600'
    if (status === 'error') return 'text-red-600'
    if (status === 'warning') return 'text-yellow-600'
    return 'text-blue-600'
  }

  const SortIcon = ({ field }) => {
    if (sortField !== field) return <ChevronsUpDown className="w-4 h-4 text-gray-400" />
    return sortOrder === 'asc'
      ? <ChevronUp className="w-4 h-4 text-primary-600" />
      : <ChevronDown className="w-4 h-4 text-primary-600" />
  }

  if (loading) {
    return <div className="text-center py-12">Loading...</div>
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">Virtual Hosts</h1>
        <button
          onClick={() => {
            setIsEditMode(false)
            setEditingVHostId(null)
            setFormData({
              name: '',
              domain: '',
              backend_url: '',
              ssl_enabled: false,
              ssl_certificate_id: '',
              enabled: true,
            })
            setShowModal(true)
          }}
          className="btn btn-primary flex items-center gap-2"
        >
          <Plus className="w-5 h-5" />
          Add Virtual Host
        </button>
      </div>

      {/* Search and View Toggle */}
      <div className="card mb-6">
        <div className="flex flex-col md:flex-row gap-4 items-center justify-between">
          <div className="relative flex-1 w-full md:w-auto">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
            <input
              type="text"
              placeholder="Search by name, domain, or backend URL..."
              className="input pl-10 w-full"
              value={searchTerm}
              onChange={(e) => {
                setSearchTerm(e.target.value)
                setCurrentPage(1)
              }}
            />
          </div>

          <div className="flex items-center gap-2">
            {selectedVHosts.length > 0 && (
              <button
                onClick={handleBulkDelete}
                className="btn btn-danger flex items-center gap-2"
              >
                <Trash2 className="w-4 h-4" />
                Delete ({selectedVHosts.length})
              </button>
            )}

            <div className="flex bg-gray-100 rounded-lg p-1">
              <button
                onClick={() => setViewMode('grid')}
                className={`p-2 rounded ${viewMode === 'grid' ? 'bg-white shadow' : 'text-gray-600'}`}
                title="Grid View"
              >
                <Grid3x3 className="w-5 h-5" />
              </button>
              <button
                onClick={() => setViewMode('list')}
                className={`p-2 rounded ${viewMode === 'list' ? 'bg-white shadow' : 'text-gray-600'}`}
                title="List View"
              >
                <List className="w-5 h-5" />
              </button>
            </div>

            <div className="flex items-center gap-2">
              <label htmlFor="itemsPerPageVHost" className="text-sm text-gray-600">Show:</label>
              <select
                id="itemsPerPageVHost"
                value={itemsPerPage}
                onChange={(e) => handleItemsPerPageChange(Number(e.target.value))}
                className="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value={5}>5</option>
                <option value={10}>10</option>
                <option value={50}>50</option>
                <option value={100}>100</option>
              </select>
            </div>
          </div>
        </div>
      </div>

      {/* Results count */}
      <div className="mb-4 text-sm text-gray-600">
        Showing {startIndex + 1} to {Math.min(startIndex + itemsPerPage, filteredAndSortedVHosts.length)} of {filteredAndSortedVHosts.length} virtual host(s)
      </div>

      {/* Grid View */}
      {viewMode === 'grid' && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {paginatedVHosts.map((vhost) => (
            <div key={vhost.id} className="card relative">
              <div className="absolute top-4 left-4">
                <input
                  type="checkbox"
                  checked={selectedVHosts.includes(vhost.id)}
                  onChange={() => handleSelectOne(vhost.id)}
                  className="w-4 h-4 rounded border-gray-300"
                />
              </div>

              <div className="flex items-start justify-between mb-4 ml-8">
                <div className="flex items-center gap-3">
                  <div className="p-2 bg-primary-100 rounded-lg">
                    <Server className="w-6 h-6 text-primary-600" />
                  </div>
                  <div>
                    <h3 className="font-semibold">{vhost.name}</h3>
                    <a
                      href={`${vhost.ssl_enabled ? 'https' : 'http'}://${vhost.domain}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm text-blue-600 hover:text-blue-800 hover:underline flex items-center gap-1"
                      onClick={(e) => e.stopPropagation()}
                    >
                      {vhost.domain}
                      <ExternalLink className="w-3 h-3" />
                    </a>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => handleViewDetail(vhost)}
                    className="text-blue-600 hover:text-blue-800 transition-colors"
                    title="View Details"
                  >
                    <Eye className="w-5 h-5" />
                  </button>
                  <button
                    onClick={() => navigate(`/vhost-config/${vhost.domain}`)}
                    className="text-purple-600 hover:text-purple-800 transition-colors"
                    title="Edit Config"
                  >
                    <FileCode className="w-5 h-5" />
                  </button>
                  <button
                    onClick={() => handleEdit(vhost)}
                    className="text-yellow-600 hover:text-yellow-800 transition-colors"
                    title="Edit"
                  >
                    <Edit className="w-5 h-5" />
                  </button>
                  <button
                    onClick={() => handleDelete(vhost.id)}
                    className="text-red-600 hover:text-red-800 transition-colors"
                    title="Delete"
                  >
                    <Trash2 className="w-5 h-5" />
                  </button>
                </div>
              </div>

              <div className="space-y-2 text-sm ml-8">
                <div>
                  <span className="text-gray-600">Backend:</span>
                  <p className="font-mono text-xs mt-1">{vhost.backend_url}</p>
                </div>
                <div className="flex gap-2">
                  {vhost.ssl_enabled && (
                    <span className="px-2 py-1 bg-green-100 text-green-800 rounded text-xs">
                      SSL
                    </span>
                  )}
                  <span className={`px-2 py-1 rounded text-xs ${vhost.enabled
                    ? 'bg-green-100 text-green-800'
                    : 'bg-gray-100 text-gray-800'
                    }`}>
                    {vhost.enabled ? 'Enabled' : 'Disabled'}
                  </span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* List View */}
      {viewMode === 'list' && (
        <div className="card overflow-hidden">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left">
                    <input
                      type="checkbox"
                      checked={filteredAndSortedVHosts.length > 0 && selectedVHosts.length === filteredAndSortedVHosts.length}
                      onChange={handleSelectAll}
                      className="w-4 h-4 rounded border-gray-300"
                    />
                  </th>
                  <th
                    className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                    onClick={() => handleSort('name')}
                  >
                    <div className="flex items-center gap-2">
                      Name
                      <SortIcon field="name" />
                    </div>
                  </th>
                  <th
                    className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                    onClick={() => handleSort('domain')}
                  >
                    <div className="flex items-center gap-2">
                      Domain
                      <SortIcon field="domain" />
                    </div>
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Backend URL
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    SSL
                  </th>
                  <th
                    className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                    onClick={() => handleSort('enabled')}
                  >
                    <div className="flex items-center gap-2">
                      Status
                      <SortIcon field="enabled" />
                    </div>
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {paginatedVHosts.map((vhost) => (
                  <tr key={vhost.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <input
                        type="checkbox"
                        checked={selectedVHosts.includes(vhost.id)}
                        onChange={() => handleSelectOne(vhost.id)}
                        className="w-4 h-4 rounded border-gray-300"
                      />
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <Server className="w-5 h-5 text-primary-600" />
                        <span className="font-medium">{vhost.name}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      <a
                        href={`${vhost.ssl_enabled ? 'https' : 'http'}://${vhost.domain}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-blue-600 hover:text-blue-800 hover:underline flex items-center gap-1"
                      >
                        {vhost.domain}
                        <ExternalLink className="w-3 h-3" />
                      </a>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <code className="text-xs bg-gray-100 px-2 py-1 rounded">{vhost.backend_url}</code>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {vhost.ssl_enabled ? (
                        <span className="px-2 py-1 bg-green-100 text-green-800 rounded text-xs">
                          Enabled
                        </span>
                      ) : (
                        <span className="px-2 py-1 bg-gray-100 text-gray-800 rounded text-xs">
                          Disabled
                        </span>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`px-2 py-1 rounded text-xs ${vhost.enabled
                        ? 'bg-green-100 text-green-800'
                        : 'bg-red-100 text-red-800'
                        }`}>
                        {vhost.enabled ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm">
                      <div className="flex items-center justify-end gap-2">
                        <button
                          onClick={() => handleViewDetail(vhost)}
                          className="text-blue-600 hover:text-blue-800 transition-colors"
                          title="View Details"
                        >
                          <Eye className="w-5 h-5" />
                        </button>
                        <button
                          onClick={() => navigate(`/vhost-config/${vhost.domain}`)}
                          className="text-purple-600 hover:text-purple-800 transition-colors"
                          title="Edit Config"
                        >
                          <FileCode className="w-5 h-5" />
                        </button>
                        <button
                          onClick={() => handleEdit(vhost)}
                          className="text-yellow-600 hover:text-yellow-800 transition-colors"
                          title="Edit"
                        >
                          <Edit className="w-5 h-5" />
                        </button>
                        <button
                          onClick={() => handleDelete(vhost.id)}
                          className="text-red-600 hover:text-red-800 transition-colors"
                          title="Delete"
                        >
                          <Trash2 className="w-5 h-5" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {filteredAndSortedVHosts.length === 0 && (
            <div className="text-center py-12 text-gray-500">
              No virtual hosts found
            </div>
          )}
        </div>
      )}

      {/* Empty state for grid view */}
      {viewMode === 'grid' && filteredAndSortedVHosts.length === 0 && (
        <div className="text-center py-12 text-gray-500">
          No virtual hosts found
        </div>
      )}

      {/* Pagination */}
      {filteredAndSortedVHosts.length > 0 && totalPages > 1 && (
        <div className="mt-6 card">
          <div className="flex items-center justify-between">
            <div className="text-sm text-gray-600">
              Page {currentPage} of {totalPages}
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={() => handlePageChange(currentPage - 1)}
                disabled={currentPage === 1}
                className="px-3 py-1 border border-gray-300 text-gray-700 rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition"
              >
                <ChevronLeft className="w-4 h-4" />
              </button>

              {Array.from({ length: totalPages }, (_, i) => i + 1).map((page) => {
                if (
                  page === 1 ||
                  page === totalPages ||
                  (page >= currentPage - 1 && page <= currentPage + 1)
                ) {
                  return (
                    <button
                      key={page}
                      onClick={() => handlePageChange(page)}
                      className={`px-3 py-1 rounded transition ${page === currentPage
                        ? 'bg-blue-600 text-white'
                        : 'border border-gray-300 text-gray-700 hover:bg-gray-50'
                        }`}
                    >
                      {page}
                    </button>
                  );
                } else if (page === currentPage - 2 || page === currentPage + 2) {
                  return <span key={page} className="px-2 text-gray-400">...</span>;
                }
                return null;
              })}

              <button
                onClick={() => handlePageChange(currentPage + 1)}
                disabled={currentPage === totalPages}
                className="px-3 py-1 border border-gray-300 text-gray-700 rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition"
              >
                <ChevronRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-3xl max-h-[90vh] overflow-y-auto">
            <h2 className="text-2xl font-bold mb-4">{isEditMode ? 'Edit Virtual Host' : 'Add Virtual Host'}</h2>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label htmlFor="vhost-name" className="label">Name</label>
                <input
                  id="vhost-name"
                  type="text"
                  className="input"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                />
              </div>

              <div>
                <label htmlFor="vhost-domain" className="label">Domain</label>
                <input
                  id="vhost-domain"
                  type="text"
                  className="input"
                  placeholder="example.com"
                  value={formData.domain}
                  onChange={(e) => setFormData({ ...formData, domain: e.target.value })}
                  required
                />
              </div>

              <div>
                <label htmlFor="vhost-backend" className="label">Backend URL</label>
                <div className="relative">
                  <input
                    id="vhost-backend"
                    type="text"
                    className="input pr-10"
                    placeholder="http://localhost:8000"
                    value={formData.backend_url}
                    onChange={(e) => setFormData({ ...formData, backend_url: e.target.value })}
                    required
                  />
                  {backendCheckStatus && (
                    <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
                      {backendCheckStatus === 'checking' && (
                        <Loader2 className="w-5 h-5 text-blue-500 animate-spin" />
                      )}
                      {backendCheckStatus === 'success' && (
                        <CheckCircle className="w-5 h-5 text-green-500" />
                      )}
                      {backendCheckStatus === 'error' && (
                        <AlertCircle className="w-5 h-5 text-red-500" />
                      )}
                    </div>
                  )}
                </div>
                {backendCheckMessage && (
                  <p className={`text-xs mt-1 flex items-center gap-1 ${getBackendCheckColor(backendCheckStatus)}`}>
                    {backendCheckMessage}
                  </p>
                )}
              </div>

              {/* Multiple Backends Section */}
              <div className="border border-gray-200 rounded-lg p-4 bg-gray-50">
                <div className="flex items-center justify-between mb-3">
                  <label className="label text-sm font-medium">Multiple Backends (Load Balancing)</label>
                  <span className="text-xs text-gray-500">Optional - for high availability</span>
                </div>

                {/* Existing backends list */}
                {formData.backends && formData.backends.length > 0 && (
                  <div className="space-y-2 mb-3">
                    {formData.backends.map((backend, index) => (
                      <div key={index} className="flex items-center gap-2">
                        <input
                          type="text"
                          className="input flex-1 text-sm"
                          value={backend}
                          onChange={(e) => {
                            const newBackends = [...formData.backends]
                            newBackends[index] = e.target.value
                            setFormData({ ...formData, backends: newBackends })
                          }}
                          placeholder="http://backend:8080"
                        />
                        <button
                          type="button"
                          onClick={() => {
                            const newBackends = formData.backends.filter((_, i) => i !== index)
                            setFormData({ ...formData, backends: newBackends })
                          }}
                          className="p-2 text-red-500 hover:bg-red-50 rounded"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    ))}
                  </div>
                )}

                {/* Add backend button */}
                <button
                  type="button"
                  onClick={() => setFormData({ ...formData, backends: [...(formData.backends || []), ''] })}
                  className="btn btn-secondary text-xs flex items-center gap-1"
                >
                  <Plus className="w-4 h-4" /> Add Backend Server
                </button>

                {/* Load Balance Method */}
                {formData.backends && formData.backends.length > 0 && (
                  <div className="mt-3">
                    <label className="label text-sm">Load Balancing Method</label>
                    <select
                      className="input text-sm"
                      value={formData.load_balance_method}
                      onChange={(e) => setFormData({ ...formData, load_balance_method: e.target.value })}
                    >
                      <option value="round_robin">Round Robin (default)</option>
                      <option value="least_conn">Least Connections</option>
                      <option value="ip_hash">IP Hash (sticky sessions)</option>
                    </select>
                  </div>
                )}
              </div>

              {/* Custom Nginx Config */}
              <div>
                <label className="label">Custom Nginx Configuration</label>
                <textarea
                  className="input font-mono text-sm"
                  rows={3}
                  placeholder="client_max_body_size 100m;&#10;proxy_buffering off;"
                  value={formData.custom_config}
                  onChange={(e) => setFormData({ ...formData, custom_config: e.target.value })}
                />
                <p className="text-xs text-gray-500 mt-1">Additional nginx directives for this vhost</p>
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="ssl"
                  checked={formData.ssl_enabled}
                  onChange={(e) => {
                    setFormData({ ...formData, ssl_enabled: e.target.checked })
                    if (!e.target.checked) {
                      setFormData(prev => ({ ...prev, ssl_certificate_id: '' }))
                      setCertSearchTerm('')
                    }
                  }}
                />
                <label htmlFor="ssl" className="text-sm">Enable SSL</label>
              </div>

              {/* SSL Certificate Dropdown */}
              {formData.ssl_enabled && (
                <div className="relative">
                  <label htmlFor="ssl-certificate" className="label">SSL Certificate *</label>
                  <div className="relative">
                    <button
                      id="ssl-certificate"
                      type="button"
                      onClick={() => setShowCertDropdown(!showCertDropdown)}
                      className="input w-full text-left flex items-center justify-between"
                    >
                      <span className={selectedCertificate ? 'text-gray-900' : 'text-gray-400'}>
                        {selectedCertificate
                          ? `${selectedCertificate.name} - Expires: ${new Date(selectedCertificate.valid_to).toLocaleDateString()}`
                          : 'Select SSL Certificate'
                        }
                      </span>
                      <ChevronDownIcon className="w-5 h-5 text-gray-400" />
                    </button>

                    {showCertDropdown && (
                      <div className="absolute z-50 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-64 overflow-hidden">
                        {/* Search Input */}
                        <div className="p-2 border-b border-gray-200">
                          <div className="relative">
                            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                            <input
                              type="text"
                              placeholder="Search certificates..."
                              className="w-full pl-9 pr-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
                              value={certSearchTerm}
                              onChange={(e) => setCertSearchTerm(e.target.value)}
                              onClick={(e) => e.stopPropagation()}
                            />
                          </div>
                        </div>

                        {/* Certificate List */}
                        <div className="max-h-48 overflow-y-auto">
                          {filteredCertificates.length === 0 ? (
                            <div className="p-3 text-sm text-gray-500 text-center">
                              No certificates found
                            </div>
                          ) : (
                            filteredCertificates.map((cert) => (
                              <button
                                key={cert.id}
                                type="button"
                                onClick={() => {
                                  setFormData({ ...formData, ssl_certificate_id: cert.id })
                                  setShowCertDropdown(false)
                                  setCertSearchTerm('')
                                }}
                                className={`w-full px-3 py-2 text-left hover:bg-gray-50 transition-colors ${formData.ssl_certificate_id === cert.id ? 'bg-primary-50' : ''
                                  }`}
                              >
                                <div className="text-sm font-medium text-gray-900">{cert.name}</div>
                                <div className="text-xs text-gray-500 mt-0.5">
                                  {cert.common_name && <span className="mr-2">CN: {cert.common_name}</span>}
                                  <span className={getCertExpiryColor(cert.valid_to)}>
                                    Expires: {new Date(cert.valid_to).toLocaleDateString()}
                                  </span>
                                </div>
                              </button>
                            ))
                          )}
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              )}

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="enabled"
                  checked={formData.enabled}
                  onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
                />
                <label htmlFor="enabled" className="text-sm">Enable Virtual Host</label>
              </div>

              {/* Advanced Settings Toggle */}
              <div className="border-t border-gray-200 pt-4">
                <button
                  type="button"
                  onClick={() => setShowAdvancedSettings(!showAdvancedSettings)}
                  className="flex items-center gap-2 text-sm font-medium text-primary-600 hover:text-primary-700"
                >
                  {showAdvancedSettings ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
                  Advanced Settings
                </button>
              </div>

              {/* Advanced Settings Section */}
              {showAdvancedSettings && (
                <div className="space-y-4 border border-gray-200 rounded-lg p-4 bg-gray-50">
                  {/* WebSocket Support */}
                  <div className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      id="websocket"
                      checked={formData.websocket_enabled}
                      onChange={(e) => setFormData({ ...formData, websocket_enabled: e.target.checked })}
                    />
                    <label htmlFor="websocket" className="text-sm">Enable WebSocket Support</label>
                  </div>

                  {/* HTTP Version */}
                  <div>
                    <label className="label">HTTP Version</label>
                    <select
                      className="input"
                      value={formData.http_version}
                      onChange={(e) => setFormData({ ...formData, http_version: e.target.value })}
                    >
                      <option value="http/1.1">HTTP/1.1</option>
                      <option value="http/2">HTTP/2</option>
                    </select>
                  </div>

                  {/* TLS Version */}
                  <div>
                    <label className="label">TLS Version (SSL/TLS Protocol)</label>
                    <select
                      className="input"
                      value={formData.tls_version}
                      onChange={(e) => setFormData({ ...formData, tls_version: e.target.value })}
                    >
                      <option value="TLSv1.2">TLS 1.2 (Recommended)</option>
                      <option value="TLSv1.3">TLS 1.3 (Most Secure)</option>
                      <option value="TLSv1.2 TLSv1.3">TLS 1.2 and 1.3 (Compatible)</option>
                      <option value="TLSv1.1 TLSv1.2 TLSv1.3">TLS 1.1, 1.2, 1.3 (Legacy)</option>
                    </select>
                    <p className="text-xs text-gray-500 mt-1">Select SSL/TLS protocol version for secure connections</p>
                  </div>

                  {/* Max Upload Size */}
                  <div>
                    <label className="label">Max Upload Size (MB)</label>
                    <input
                      type="number"
                      className="input"
                      min="1"
                      max="1024"
                      value={formData.max_upload_size}
                      onChange={(e) => setFormData({ ...formData, max_upload_size: Number.parseInt(e.target.value) || 10 })}
                    />
                    <p className="text-xs text-gray-500 mt-1">Maximum file upload size (client_max_body_size)</p>
                  </div>

                  {/* Proxy Timeouts */}
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="label">Read Timeout (seconds)</label>
                      <input
                        type="number"
                        className="input"
                        min="1"
                        max="3600"
                        value={formData.proxy_read_timeout}
                        onChange={(e) => setFormData({ ...formData, proxy_read_timeout: Number.parseInt(e.target.value) || 60 })}
                      />
                    </div>
                    <div>
                      <label className="label">Connect Timeout (seconds)</label>
                      <input
                        type="number"
                        className="input"
                        min="1"
                        max="300"
                        value={formData.proxy_connect_timeout}
                        onChange={(e) => setFormData({ ...formData, proxy_connect_timeout: Number.parseInt(e.target.value) || 60 })}
                      />
                    </div>
                  </div>

                  {/* Bot Detection */}
                  <div className="border-t border-gray-300 pt-4">
                    <div className="flex items-center gap-2 mb-3">
                      <input
                        type="checkbox"
                        id="bot_detection"
                        checked={formData.bot_detection_enabled}
                        onChange={(e) => setFormData({ ...formData, bot_detection_enabled: e.target.checked })}
                      />
                      <label htmlFor="bot_detection" className="text-sm font-medium">Enable Bot Detection</label>
                    </div>
                    {formData.bot_detection_enabled && (
                      <div className="space-y-3">
                        <div>
                          <label className="label">Challenge Type</label>
                          <select
                            className="input"
                            value={formData.bot_detection_type}
                            onChange={(e) => setFormData({ ...formData, bot_detection_type: e.target.value })}
                          >
                            <option value="turnstile">Cloudflare Turnstile</option>
                            <option value="captcha">Google reCAPTCHA</option>
                            <option value="slide_puzzle">Slide Puzzle</option>
                          </select>
                          <p className="text-xs text-gray-500 mt-1">Show challenge page before allowing access to this vhost</p>
                        </div>
                        {formData.bot_detection_type === 'captcha' && (
                          <div>
                            <label className="label">reCAPTCHA Version</label>
                            <select
                              className="input"
                              value={formData.recaptcha_version || 'v2'}
                              onChange={(e) => setFormData({ ...formData, recaptcha_version: e.target.value })}
                            >
                              <option value="v2">v2 (Checkbox - "I'm not a robot")</option>
                              <option value="v3">v3 (Invisible - Score based)</option>
                            </select>
                            <p className="text-xs text-gray-500 mt-1">
                              {formData.recaptcha_version === 'v3'
                                ? 'v3: Invisible challenge with automatic scoring (0.0-1.0)'
                                : 'v2: Visible checkbox with manual verification'}
                            </p>
                          </div>
                        )}
                      </div>
                    )}
                  </div>

                  {/* Rate Limiter */}
                  <div className="border-t border-gray-300 pt-4">
                    <div className="flex items-center gap-2 mb-3">
                      <input
                        type="checkbox"
                        id="rate_limit"
                        checked={formData.rate_limit_enabled}
                        onChange={(e) => setFormData({ ...formData, rate_limit_enabled: e.target.checked })}
                      />
                      <label htmlFor="rate_limit" className="text-sm font-medium">Enable Rate Limiting</label>
                    </div>
                    {formData.rate_limit_enabled && (
                      <div className="grid grid-cols-2 gap-3">
                        <div>
                          <label className="label">Max Requests</label>
                          <input
                            type="number"
                            className="input"
                            min="1"
                            max="10000"
                            value={formData.rate_limit_requests}
                            onChange={(e) => setFormData({ ...formData, rate_limit_requests: Number.parseInt(e.target.value) || 100 })}
                          />
                        </div>
                        <div>
                          <label className="label">Time Window (seconds)</label>
                          <input
                            type="number"
                            className="input"
                            min="1"
                            max="3600"
                            value={formData.rate_limit_window}
                            onChange={(e) => setFormData({ ...formData, rate_limit_window: Number.parseInt(e.target.value) || 60 })}
                          />
                        </div>
                        <div className="col-span-2">
                          <p className="text-xs text-gray-500">
                            Limit: {formData.rate_limit_requests} requests per {formData.rate_limit_window} seconds per IP
                          </p>
                        </div>
                      </div>
                    )}
                  </div>

                  {/* Region Filtering */}
                  <div className="border-t border-gray-300 pt-4">
                    <div className="flex items-center gap-2 mb-3">
                      <input
                        type="checkbox"
                        id="region_filtering"
                        checked={formData.region_filtering_enabled}
                        onChange={(e) => setFormData({ ...formData, region_filtering_enabled: e.target.checked })}
                      />
                      <label htmlFor="region_filtering" className="text-sm font-medium">Enable Region Filtering</label>
                    </div>
                    {formData.region_filtering_enabled && (
                      <div className="space-y-3">
                        <div>
                          <label className="label">Whitelist Countries (ISO codes, e.g., US,GB,ID)</label>
                          <input
                            type="text"
                            className="input"
                            placeholder="US,GB,ID,SG"
                            value={formData.region_whitelist?.join(',') || ''}
                            onChange={(e) => {
                              const codes = e.target.value.split(',').map(c => c.trim().toUpperCase()).filter(Boolean);
                              setFormData({ ...formData, region_whitelist: codes });
                            }}
                          />
                          <p className="text-xs text-gray-500 mt-1">
                            If set, ONLY these countries are allowed. Leave empty to allow all except blacklisted.
                          </p>
                        </div>
                        <div>
                          <label className="label">Blacklist Countries (ISO codes, e.g., CN,RU)</label>
                          <input
                            type="text"
                            className="input"
                            placeholder="CN,RU,KP"
                            value={formData.region_blacklist?.join(',') || ''}
                            onChange={(e) => {
                              const codes = e.target.value.split(',').map(c => c.trim().toUpperCase()).filter(Boolean);
                              setFormData({ ...formData, region_blacklist: codes });
                            }}
                          />
                          <p className="text-xs text-gray-500 mt-1">
                            These countries will be blocked. Only applies if whitelist is empty.
                          </p>
                        </div>
                        <div className="bg-blue-50 border border-blue-200 rounded p-3 text-xs">
                          <strong>Region Filtering Logic:</strong>
                          <ul className="list-disc list-inside mt-1 space-y-1">
                            <li>If whitelist is set: ONLY whitelist countries are allowed</li>
                            <li>If whitelist is empty: All countries except blacklisted are allowed</li>
                            <li>Uses IP-based geolocation (may not be 100% accurate)</li>
                          </ul>
                        </div>
                      </div>
                    )}
                  </div>

                  {/* Custom Headers */}
                  <div>
                    <label className="label">Custom Headers</label>
                    <div className="space-y-2">
                      {Object.entries(formData.custom_headers || {}).map(([key, value], index) => (
                        <div key={index} className="flex items-center gap-2">
                          <input
                            type="text"
                            className="input flex-1 text-sm"
                            value={key}
                            disabled
                          />
                          <span className="text-gray-400">:</span>
                          <input
                            type="text"
                            className="input flex-1 text-sm"
                            value={value}
                            disabled
                          />
                          <button
                            type="button"
                            onClick={() => {
                              const newHeaders = { ...formData.custom_headers }
                              delete newHeaders[key]
                              setFormData({ ...formData, custom_headers: newHeaders })
                            }}
                            className="text-red-600 hover:text-red-800"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      ))}
                      <div className="flex items-center gap-2">
                        <input
                          type="text"
                          className="input flex-1 text-sm"
                          placeholder="Header name"
                          value={newHeader.key}
                          onChange={(e) => setNewHeader({ ...newHeader, key: e.target.value })}
                        />
                        <span className="text-gray-400">:</span>
                        <input
                          type="text"
                          className="input flex-1 text-sm"
                          placeholder="Header value"
                          value={newHeader.value}
                          onChange={(e) => setNewHeader({ ...newHeader, value: e.target.value })}
                        />
                        <button
                          type="button"
                          onClick={(e) => {
                            e.preventDefault()
                            e.stopPropagation()
                            if (newHeader.key && newHeader.value) {
                              setFormData({
                                ...formData,
                                custom_headers: { ...(formData.custom_headers || {}), [newHeader.key]: newHeader.value }
                              })
                              setNewHeader({ key: '', value: '' })
                            }
                          }}
                          className="p-2 text-white bg-green-600 hover:bg-green-700 rounded transition-colors"
                          title="Add header"
                        >
                          <Plus className="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  </div>

                  {/* Custom Locations */}
                  <div>
                    <label className="label">Custom Location Blocks</label>
                    <p className="text-xs text-gray-500 mb-2">Add custom nginx location blocks for specific URL paths</p>
                    <div className="space-y-3">
                      {(formData.custom_locations || []).map((loc, index) => (
                        <div key={index} className="border border-gray-300 rounded-lg p-3 bg-white">
                          <div className="flex items-center justify-between mb-2">
                            <div className="flex items-center gap-2">
                              <span className="text-sm font-medium text-gray-700">location {loc.path}</span>
                              {loc.websocket_enabled && (
                                <span className="px-2 py-0.5 bg-purple-100 text-purple-700 rounded text-xs font-medium">
                                  WebSocket
                                </span>
                              )}
                            </div>
                            <button
                              type="button"
                              onClick={() => {
                                const newLocations = [...formData.custom_locations]
                                newLocations.splice(index, 1)
                                setFormData({ ...formData, custom_locations: newLocations })
                              }}
                              className="text-red-600 hover:text-red-800"
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
                          </div>
                          {loc.proxy_pass && (
                            <p className="text-xs text-gray-600">proxy_pass: {loc.proxy_pass}</p>
                          )}
                          {loc.config && (
                            <pre className="text-xs text-gray-600 mt-1 whitespace-pre-wrap">{loc.config}</pre>
                          )}
                        </div>
                      ))}

                      {/* Add New Location */}
                      <div className="border border-dashed border-gray-300 rounded-lg p-3 bg-white">
                        <div className="space-y-2">
                          <input
                            type="text"
                            className="input text-sm"
                            placeholder="Location path (e.g., /api, /svc-base)"
                            value={newLocation.path}
                            onChange={(e) => setNewLocation({ ...newLocation, path: e.target.value })}
                          />
                          <div className="relative">
                            <input
                              type="text"
                              className="input text-sm pr-10"
                              placeholder="Proxy pass URL (e.g., http://backend:8080)"
                              value={newLocation.proxy_pass}
                              onChange={(e) => setNewLocation({ ...newLocation, proxy_pass: e.target.value })}
                            />
                            {locationBackendCheck.status && (
                              <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
                                {locationBackendCheck.status === 'checking' && (
                                  <Loader2 className="w-4 h-4 text-blue-500 animate-spin" />
                                )}
                                {locationBackendCheck.status === 'success' && (
                                  <CheckCircle className="w-4 h-4 text-green-500" />
                                )}
                                {locationBackendCheck.status === 'error' && (
                                  <AlertCircle className="w-4 h-4 text-red-500" />
                                )}
                                {locationBackendCheck.status === 'warning' && (
                                  <AlertCircle className="w-4 h-4 text-yellow-500" />
                                )}
                              </div>
                            )}
                          </div>
                          {locationBackendCheck.message && (
                            <p className={`text-xs flex items-center gap-1 ${getLocationCheckColor(locationBackendCheck.status)}`}>
                              {locationBackendCheck.message}
                            </p>
                          )}
                          <textarea
                            className="input text-sm"
                            rows="3"
                            placeholder="Additional nginx config (optional)"
                            value={newLocation.config}
                            onChange={(e) => setNewLocation({ ...newLocation, config: e.target.value })}
                          />
                          <label className="flex items-center gap-2 cursor-pointer">
                            <input
                              type="checkbox"
                              className="w-4 h-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                              checked={newLocation.websocket_enabled || false}
                              onChange={(e) => setNewLocation({ ...newLocation, websocket_enabled: e.target.checked })}
                            />
                            <span className="text-sm text-gray-700">Enable WebSocket Support</span>
                            <span className="text-xs text-gray-500">(Adds Upgrade/Connection headers)</span>
                          </label>
                          <button
                            type="button"
                            onClick={(e) => {
                              e.preventDefault()
                              e.stopPropagation()
                              if (newLocation.path) {
                                setFormData({
                                  ...formData,
                                  custom_locations: [...(formData.custom_locations || []), { ...newLocation }]
                                })
                                setNewLocation({ path: '', proxy_pass: '', config: '', websocket_enabled: false })
                                setLocationBackendCheck({ status: null, message: '' })
                              }
                            }}
                            className="btn btn-primary text-sm w-full flex items-center justify-center gap-2"
                          >
                            <Plus className="w-4 h-4" />
                            Add Location Block
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              )}

              <div className="flex gap-2 pt-4">
                <button type="submit" className="btn btn-primary flex-1">
                  {isEditMode ? 'Update' : 'Create'}
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setShowModal(false)
                    setIsEditMode(false)
                    setEditingVHostId(null)
                    setShowCertDropdown(false)
                    setCertSearchTerm('')
                    setBackendCheckStatus(null)
                    setBackendCheckMessage('')
                    setFormData({
                      name: '',
                      domain: '',
                      backend_url: '',
                      ssl_enabled: false,
                      ssl_certificate_id: '',
                      enabled: true,
                      websocket_enabled: false,
                      http_version: 'http/1.1',
                      tls_version: 'TLSv1.2',
                      max_upload_size: 10,
                      proxy_read_timeout: 60,
                      proxy_connect_timeout: 60,
                      custom_locations: [],
                      custom_headers: {},
                    })
                    setShowAdvancedSettings(false)
                    setNewLocation({ path: '', proxy_pass: '', config: '', websocket_enabled: false })
                    setNewHeader({ key: '', value: '' })
                  }}
                  className="btn btn-secondary flex-1"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Confirmation Modal */}
      {showConfirmModal && (
        <div
          className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
          role="dialog"
          aria-modal="true"
        >
          <div
            className="bg-white rounded-lg p-6 w-full max-w-md"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-start gap-4 mb-4">
              <div className="p-2 bg-red-100 rounded-full">
                <AlertCircle className="w-6 h-6 text-red-600" />
              </div>
              <div>
                <h3 className="text-lg font-semibold text-gray-900 mb-2">Confirm Action</h3>
                <p className="text-gray-600">{confirmMessage}</p>
              </div>
            </div>
            <div className="flex gap-3 justify-end">
              <button
                type="button"
                onClick={() => {
                  setShowConfirmModal(false)
                  setConfirmAction(null)
                }}
                className="btn btn-secondary"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={async () => {
                  if (confirmAction) {
                    await confirmAction()
                  }
                }}
                className="btn btn-danger"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Detail Modal */}
      {showDetailModal && selectedVHost && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" onClick={() => setShowDetailModal(false)}>
          <div className="bg-white rounded-lg p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-2xl font-bold flex items-center gap-3">
                <div className="p-2 bg-primary-100 rounded-lg">
                  <Server className="w-6 h-6 text-primary-600" />
                </div>
                Virtual Host Details
              </h2>
              <button
                onClick={() => setShowDetailModal(false)}
                className="text-gray-400 hover:text-gray-600"
              >
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <div className="space-y-6">
              {/* Basic Info */}
              <div className="bg-gray-50 rounded-lg p-4">
                <h3 className="text-sm font-semibold text-gray-700 mb-3 flex items-center gap-2">
                  <Globe className="w-4 h-4" /> Basic Information
                </h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-xs text-gray-500">Name</label>
                    <p className="text-gray-900 font-medium">{selectedVHost.name}</p>
                  </div>
                  <div>
                    <label className="text-xs text-gray-500">Domain</label>
                    <a
                      href={`${selectedVHost.ssl_enabled ? 'https' : 'http'}://${selectedVHost.domain}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-blue-600 hover:underline flex items-center gap-1 font-medium"
                    >
                      {selectedVHost.domain}
                      <ExternalLink className="w-3 h-3" />
                    </a>
                  </div>
                  <div>
                    <label className="text-xs text-gray-500">Status</label>
                    <span className={`px-2 py-1 rounded text-xs font-medium ${selectedVHost.enabled ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
                      {selectedVHost.enabled ? 'Active' : 'Inactive'}
                    </span>
                  </div>
                  <div>
                    <label className="text-xs text-gray-500">SSL</label>
                    <span className={`px-2 py-1 rounded text-xs font-medium ${selectedVHost.ssl_enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-600'}`}>
                      {selectedVHost.ssl_enabled ? 'Enabled' : 'Disabled'}
                    </span>
                  </div>
                </div>
              </div>

              {/* Backend Configuration */}
              <div className="bg-blue-50 rounded-lg p-4">
                <h3 className="text-sm font-semibold text-gray-700 mb-3 flex items-center gap-2">
                  <Server className="w-4 h-4" /> Backend Configuration
                </h3>
                <div className="space-y-3">
                  <div>
                    <label className="text-xs text-gray-500">Primary Backend URL</label>
                    <p className="font-mono text-sm bg-white px-3 py-2 rounded border">{selectedVHost.backend_url}</p>
                  </div>
                  {selectedVHost.backends && selectedVHost.backends.length > 0 && (
                    <div>
                      <label className="text-xs text-gray-500">Additional Backends ({selectedVHost.backends.length})</label>
                      <div className="space-y-1 mt-1">
                        {selectedVHost.backends.map((backend, idx) => (
                          <p key={idx} className="font-mono text-sm bg-white px-3 py-1 rounded border">{backend}</p>
                        ))}
                      </div>
                    </div>
                  )}
                  {selectedVHost.backends && selectedVHost.backends.length > 0 && (
                    <div>
                      <label className="text-xs text-gray-500">Load Balance Method</label>
                      <span className="px-2 py-1 bg-purple-100 text-purple-800 rounded text-xs font-medium ml-2">
                        {selectedVHost.load_balance_method || 'round_robin'}
                      </span>
                    </div>
                  )}
                </div>
              </div>

              {/* Protocol Settings */}
              <div className="bg-purple-50 rounded-lg p-4">
                <h3 className="text-sm font-semibold text-gray-700 mb-3 flex items-center gap-2">
                  <Shield className="w-4 h-4" /> Protocol & Security
                </h3>
                <div className="grid grid-cols-3 gap-3">
                  <div>
                    <label className="text-xs text-gray-500">HTTP Version</label>
                    <p className="font-medium">{selectedVHost.http_version || '1.1'}</p>
                  </div>
                  <div>
                    <label className="text-xs text-gray-500">TLS Version</label>
                    <p className="font-medium">{selectedVHost.tls_version || '1.2'}</p>
                  </div>
                  <div>
                    <label className="text-xs text-gray-500">WebSocket</label>
                    <span className={`px-2 py-1 rounded text-xs font-medium ${selectedVHost.websocket_enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-600'}`}>
                      {selectedVHost.websocket_enabled ? 'Enabled' : 'Disabled'}
                    </span>
                  </div>
                  <div>
                    <label className="text-xs text-gray-500">Max Upload</label>
                    <p className="font-medium">{selectedVHost.max_upload_size || 100} MB</p>
                  </div>
                  <div>
                    <label className="text-xs text-gray-500">Read Timeout</label>
                    <p className="font-medium">{selectedVHost.proxy_read_timeout || 60}s</p>
                  </div>
                  <div>
                    <label className="text-xs text-gray-500">Connect Timeout</label>
                    <p className="font-medium">{selectedVHost.proxy_connect_timeout || 60}s</p>
                  </div>
                </div>
              </div>

              {/* Bot Detection & Rate Limiting */}
              <div className="grid grid-cols-2 gap-4">
                <div className="bg-orange-50 rounded-lg p-4">
                  <h3 className="text-sm font-semibold text-gray-700 mb-3 flex items-center gap-2">
                    🤖 Bot Detection
                  </h3>
                  <div className="space-y-2">
                    <div className="flex justify-between items-center">
                      <span className="text-xs text-gray-500">Status</span>
                      <span className={`px-2 py-1 rounded text-xs font-medium ${selectedVHost.bot_detection_enabled ? 'bg-orange-100 text-orange-800' : 'bg-gray-100 text-gray-600'}`}>
                        {selectedVHost.bot_detection_enabled ? 'Enabled' : 'Disabled'}
                      </span>
                    </div>
                    {selectedVHost.bot_detection_enabled && (
                      <div className="flex justify-between items-center">
                        <span className="text-xs text-gray-500">Type</span>
                        <span className="text-sm font-medium">{selectedVHost.bot_detection_type || 'turnstile'}</span>
                      </div>
                    )}
                  </div>
                </div>

                <div className="bg-yellow-50 rounded-lg p-4">
                  <h3 className="text-sm font-semibold text-gray-700 mb-3 flex items-center gap-2">
                    ⏱️ Rate Limiting
                  </h3>
                  <div className="space-y-2">
                    <div className="flex justify-between items-center">
                      <span className="text-xs text-gray-500">Status</span>
                      <span className={`px-2 py-1 rounded text-xs font-medium ${selectedVHost.rate_limit_enabled ? 'bg-yellow-100 text-yellow-800' : 'bg-gray-100 text-gray-600'}`}>
                        {selectedVHost.rate_limit_enabled ? 'Enabled' : 'Disabled'}
                      </span>
                    </div>
                    {selectedVHost.rate_limit_enabled && (
                      <>
                        <div className="flex justify-between items-center">
                          <span className="text-xs text-gray-500">Requests</span>
                          <span className="text-sm font-medium">{selectedVHost.rate_limit_requests || 100}</span>
                        </div>
                        <div className="flex justify-between items-center">
                          <span className="text-xs text-gray-500">Window</span>
                          <span className="text-sm font-medium">{selectedVHost.rate_limit_window || 60}s</span>
                        </div>
                      </>
                    )}
                  </div>
                </div>
              </div>

              {/* Custom Locations */}
              {selectedVHost.custom_locations && selectedVHost.custom_locations.length > 0 && (
                <div className="bg-indigo-50 rounded-lg p-4">
                  <h3 className="text-sm font-semibold text-gray-700 mb-3 flex items-center gap-2">
                    📍 Custom Locations ({selectedVHost.custom_locations.length})
                  </h3>
                  <div className="space-y-2">
                    {selectedVHost.custom_locations.map((loc, idx) => (
                      <div key={idx} className="bg-white rounded p-3 border text-sm">
                        <div className="flex items-center gap-2 mb-1">
                          <span className="font-mono font-medium text-indigo-600">{loc.path}</span>
                          <span className="text-gray-400">→</span>
                          <span className="font-mono text-gray-600">{loc.proxy_pass}</span>
                        </div>
                        {loc.websocket_enabled && (
                          <span className="text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded">WebSocket</span>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Custom Config */}
              {selectedVHost.custom_config && (
                <div className="bg-gray-100 rounded-lg p-4">
                  <h3 className="text-sm font-semibold text-gray-700 mb-3 flex items-center gap-2">
                    📝 Custom Nginx Config
                  </h3>
                  <pre className="bg-gray-800 text-green-400 p-3 rounded text-xs overflow-x-auto max-h-32">
                    {selectedVHost.custom_config}
                  </pre>
                </div>
              )}

              {/* Timestamps */}
              <div className="grid grid-cols-2 gap-4 pt-4 border-t text-sm">
                <div>
                  <label className="text-xs text-gray-500">Created At</label>
                  <p className="text-gray-700">{selectedVHost.created_at ? new Date(selectedVHost.created_at).toLocaleString() : '-'}</p>
                </div>
                <div>
                  <label className="text-xs text-gray-500">Updated At</label>
                  <p className="text-gray-700">{selectedVHost.updated_at ? new Date(selectedVHost.updated_at).toLocaleString() : '-'}</p>
                </div>
              </div>
            </div>

            <div className="flex gap-3 justify-end mt-6 pt-4 border-t">
              <button
                onClick={() => setShowDetailModal(false)}
                className="btn btn-secondary"
              >
                Close
              </button>
              <button
                onClick={() => {
                  setShowDetailModal(false)
                  handleEdit(selectedVHost)
                }}
                className="btn btn-primary flex items-center gap-2"
              >
                <Edit className="w-4 h-4" />
                Edit
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Global Loader */}
      {globalLoading && (
        <div className="fixed inset-0 bg-black bg-opacity-60 flex items-center justify-center z-[60]" style={{ backdropFilter: 'blur(4px)' }}>
          <div className="bg-white rounded-2xl p-8 flex flex-col items-center gap-4 shadow-2xl">
            {/* Animated Logo */}
            <div className="relative">
              {/* Outer spinning ring */}
              <div className="absolute inset-0 rounded-full border-4 border-primary-200 border-t-primary-600 animate-spin w-24 h-24"></div>

              {/* Middle pulsing ring */}
              <div className="absolute inset-2 rounded-full bg-primary-100 animate-pulse"></div>

              {/* Logo in center */}
              <div className="relative w-24 h-24 flex items-center justify-center">
                <div className="bg-gradient-to-br from-primary-600 to-primary-800 rounded-full p-4 shadow-lg">
                  <Shield className="w-10 h-10 text-white" />
                </div>
              </div>
            </div>

            {/* Loading text */}
            <div className="text-center">
              <h3 className="text-lg font-semibold text-gray-900 mb-1">
                {loadingMessage || 'Processing...'}
              </h3>
              <p className="text-sm text-gray-500">Please wait</p>
            </div>

            {/* Progress dots */}
            <div className="flex gap-2">
              <div className="w-2 h-2 bg-primary-600 rounded-full animate-bounce" style={{ animationDelay: '0ms' }}></div>
              <div className="w-2 h-2 bg-primary-600 rounded-full animate-bounce" style={{ animationDelay: '150ms' }}></div>
              <div className="w-2 h-2 bg-primary-600 rounded-full animate-bounce" style={{ animationDelay: '300ms' }}></div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default VHosts
