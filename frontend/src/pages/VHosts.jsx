import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import Swal from 'sweetalert2'
import { getVHosts, deleteVHost, regenerateAllConfigs } from '../services/api'
import { Plus, Trash2, Edit, Server, Search, Grid3x3, List, ChevronUp, ChevronDown, ChevronsUpDown, AlertCircle, Loader2, Shield, Eye, ExternalLink, ChevronLeft, ChevronRight, Globe, RefreshCw, Clock, MoreVertical, Play, Pause, FileCode } from 'lucide-react'
import logger from '../utils/logger'

const VHosts = () => {
  const navigate = useNavigate()
  const [vhosts, setVHosts] = useState([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [sortField, setSortField] = useState('name')
  const [sortOrder, setSortOrder] = useState('asc')
  const [selectedVHosts, setSelectedVHosts] = useState([])
  const [viewMode, setViewMode] = useState('grid') // 'grid' or 'list'
  const [showConfirmModal, setShowConfirmModal] = useState(false)
  const [confirmAction, setConfirmAction] = useState(null)
  const [confirmMessage, setConfirmMessage] = useState('')
  const [confirmButtonText, setConfirmButtonText] = useState('Confirm')
  const [confirmButtonStyle, setConfirmButtonStyle] = useState('btn-danger')
  const [globalLoading, setGlobalLoading] = useState(false)
  const [loadingMessage, setLoadingMessage] = useState('')
  const [showDetailModal, setShowDetailModal] = useState(false)
  const [selectedVHost, setSelectedVHost] = useState(null)
  const [regenerating, setRegenerating] = useState(false)

  // Pagination states
  const [currentPage, setCurrentPage] = useState(1)
  const [itemsPerPage, setItemsPerPage] = useState(12)

  useEffect(() => {
    loadVHosts()
  }, [])

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

  const handleDelete = (id) => {
    setConfirmMessage('Are you sure you want to delete this virtual host?')
    setConfirmButtonText('Delete')
    setConfirmButtonStyle('btn-danger')
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

  const handleRegenerateConfigs = () => {
    setConfirmMessage('This will regenerate nginx configuration files for all enabled virtual hosts using the optimized template. Do you want to continue?')
    setConfirmButtonText('Regenerate')
    setConfirmButtonStyle('btn-primary')
    setConfirmAction(() => async () => {
      setShowConfirmModal(false)
      setRegenerating(true)
      setGlobalLoading(true)
      setLoadingMessage('Regenerating nginx configs...')
      try {
        const response = await regenerateAllConfigs()
        const { count, errors } = response.data
        if (errors && errors.length > 0) {
          logger.warn('Some configs had errors:', errors)
          Swal.fire({
            icon: 'warning',
            title: 'Partial Success',
            text: `Regenerated ${count} configs with ${errors.length} errors. Check console for details.`
          })
        } else {
          Swal.fire({
            icon: 'success',
            title: 'Success!',
            text: `Successfully regenerated ${count} nginx configuration files!`,
            timer: 2000,
            showConfirmButton: false
          })
        }
      } catch (error) {
        logger.error('Failed to regenerate configs:', error)
        Swal.fire({
          icon: 'error',
          title: 'Failed',
          text: 'Failed to regenerate configs: ' + (error.response?.data?.error || error.message)
        })
      } finally {
        setRegenerating(false)
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
    navigate(`/vhosts/${vhost.id}/edit`)
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
          onClick={handleRegenerateConfigs}
          disabled={regenerating}
          className="btn btn-secondary flex items-center gap-2"
          title="Regenerate nginx configs for all vhosts using optimized template"
        >
          <RefreshCw className={`w-5 h-5 ${regenerating ? 'animate-spin' : ''}`} />
          {regenerating ? 'Regenerating...' : 'Regenerate Configs'}
        </button>
        <button
          onClick={() => navigate('/vhosts/new')}
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
                <div className="flex flex-wrap gap-2">
                  <span className={`px-2 py-1 rounded text-xs font-medium border ${
                    !vhost.defense_mode || vhost.defense_mode === 'defense' ? 'bg-indigo-50 text-indigo-700 border-indigo-200' :
                    vhost.defense_mode === 'audited' ? 'bg-orange-50 text-orange-700 border-orange-200' :
                    'bg-gray-50 text-gray-700 border-gray-200'
                  }`}>
                    {!vhost.defense_mode || vhost.defense_mode === 'defense' ? 'Defense Mode' : 
                     vhost.defense_mode === 'audited' ? 'Audited Mode' : 
                     'Offline Mode'}
                  </span>
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
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Def Mode
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
                      <span className={`px-2 py-1 rounded text-xs font-medium border ${
                        !vhost.defense_mode || vhost.defense_mode === 'defense' ? 'bg-indigo-50 text-indigo-700 border-indigo-200' :
                        vhost.defense_mode === 'audited' ? 'bg-orange-50 text-orange-700 border-orange-200' :
                        'bg-gray-50 text-gray-700 border-gray-200'
                      }`}>
                        {!vhost.defense_mode || vhost.defense_mode === 'defense' ? 'Defense' : 
                         vhost.defense_mode === 'audited' ? 'Audited' : 
                         'Offline'}
                      </span>
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
                className={`btn ${confirmButtonStyle}`}
              >
                {confirmButtonText}
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
                  <div>
                    <label className="text-xs text-gray-500">Defense Mode</label>
                    <span className={`px-2 py-1 rounded text-xs font-medium border ${
                      !selectedVHost.defense_mode || selectedVHost.defense_mode === 'defense' ? 'bg-indigo-50 text-indigo-700 border-indigo-200' :
                      selectedVHost.defense_mode === 'audited' ? 'bg-orange-50 text-orange-700 border-orange-200' :
                      'bg-gray-50 text-gray-700 border-gray-200'
                    }`}>
                      {!selectedVHost.defense_mode || selectedVHost.defense_mode === 'defense' ? 'Defense' : 
                       selectedVHost.defense_mode === 'audited' ? 'Audited' : 
                       'Offline'}
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
                  navigate(`/vhosts/${selectedVHost.id}/edit`)
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
