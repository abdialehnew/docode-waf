import { useEffect, useState } from 'react'
import { getTrafficLogs } from '../services/api'
import { format } from 'date-fns'
import { RefreshCw, ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Search } from 'lucide-react'
import logger from '../utils/logger'

const TrafficLogs = () => {
  const [logs, setLogs] = useState([])
  const [filteredLogs, setFilteredLogs] = useState([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [rowsPerPage, setRowsPerPage] = useState(10)
  const [totalLogs, setTotalLogs] = useState(0)
  const [searchTerm, setSearchTerm] = useState('')
  const [sortField, setSortField] = useState('timestamp')
  const [sortOrder, setSortOrder] = useState('desc')

  useEffect(() => {
    loadLogs()
  }, [])

  useEffect(() => {
    filterAndSortLogs()
  }, [logs, searchTerm, sortField, sortOrder])

  const loadLogs = async () => {
    setLoading(true)
    try {
      // Load more data for client-side pagination
      const response = await getTrafficLogs(1000, 0)
      const logsData = response.data || []
      setLogs(logsData)
      setTotalLogs(logsData.length)
    } catch (error) {
      logger.error('Failed to load traffic logs:', error)
      setLogs([])
      setTotalLogs(0)
    } finally {
      setLoading(false)
    }
  }

  const filterAndSortLogs = () => {
    let filtered = [...logs]

    // Search filter
    if (searchTerm) {
      filtered = filtered.filter(log => 
        log.client_ip?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        log.url?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        log.method?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        log.host?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        log.status_code?.toString().includes(searchTerm)
      )
    }

    // Sorting
    filtered.sort((a, b) => {
      let aVal = a[sortField]
      let bVal = b[sortField]

      if (sortField === 'timestamp') {
        aVal = new Date(aVal).getTime()
        bVal = new Date(bVal).getTime()
      }

      if (sortOrder === 'asc') {
        return aVal > bVal ? 1 : -1
      } else {
        return aVal < bVal ? 1 : -1
      }
    })

    setFilteredLogs(filtered)
    setTotalLogs(filtered.length)
    // Reset to first page when filter changes
    if (page > Math.ceil(filtered.length / rowsPerPage)) {
      setPage(1)
    }
  }

  const handleSort = (field) => {
    if (sortField === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')
    } else {
      setSortField(field)
      setSortOrder('desc')
    }
  }

  const getSortIcon = (field) => {
    if (sortField !== field) return '↕'
    return sortOrder === 'asc' ? '↑' : '↓'
  }

  const getStatusColor = (status) => {
    if (status >= 200 && status < 300) return 'text-green-600'
    if (status >= 300 && status < 400) return 'text-blue-600'
    if (status >= 400 && status < 500) return 'text-orange-600'
    if (status >= 500) return 'text-red-600'
    return 'text-gray-600'
  }

  // Pagination calculations
  const totalPages = Math.ceil(totalLogs / rowsPerPage)
  const startIndex = (page - 1) * rowsPerPage
  const endIndex = startIndex + rowsPerPage
  const currentPageLogs = filteredLogs.slice(startIndex, endIndex)
  const showingFrom = totalLogs === 0 ? 0 : startIndex + 1
  const showingTo = Math.min(endIndex, totalLogs)

  const handleRowsPerPageChange = (newRowsPerPage) => {
    setRowsPerPage(newRowsPerPage)
    setPage(1)
  }

  if (loading) {
    return <div className="text-center py-12">Loading...</div>
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">Traffic Logs</h1>
        <button
          onClick={loadLogs}
          className="btn btn-secondary flex items-center gap-2"
        >
          <RefreshCw className="w-5 h-5" />
          Refresh
        </button>
      </div>

      <div className="card">
        {/* Search and Controls */}
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-4 pb-4 border-b">
          <div className="flex items-center gap-2">
            <label className="text-sm text-gray-600">Show</label>
            <select
              value={rowsPerPage}
              onChange={(e) => handleRowsPerPageChange(Number(e.target.value))}
              className="input py-1 px-2 text-sm"
            >
              <option value={5}>5</option>
              <option value={10}>10</option>
              <option value={50}>50</option>
              <option value={100}>100</option>
            </select>
            <label className="text-sm text-gray-600">entries</label>
          </div>

          <div className="relative w-full sm:w-64">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
            <input
              type="text"
              placeholder="Search logs..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="input pl-10 w-full"
            />
          </div>
        </div>

        {loading ? (
          <div className="text-center py-12">Loading...</div>
        ) : filteredLogs.length === 0 ? (
          <div className="text-center py-12 text-gray-500">
            <p className="text-lg mb-2">No traffic logs found</p>
            <p className="text-sm">
              {searchTerm 
                ? 'Try adjusting your search terms' 
                : 'Traffic logs will appear here once requests are made to your virtual hosts'}
            </p>
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b bg-gray-50">
                    <th 
                      className="text-left py-3 px-4 cursor-pointer hover:bg-gray-100"
                      onClick={() => handleSort('timestamp')}
                    >
                      <div className="flex items-center gap-1">
                        Time {getSortIcon('timestamp')}
                      </div>
                    </th>
                    <th 
                      className="text-left py-3 px-4 cursor-pointer hover:bg-gray-100"
                      onClick={() => handleSort('client_ip')}
                    >
                      <div className="flex items-center gap-1">
                        IP {getSortIcon('client_ip')}
                      </div>
                    </th>
                    <th 
                      className="text-left py-3 px-4 cursor-pointer hover:bg-gray-100"
                      onClick={() => handleSort('method')}
                    >
                      <div className="flex items-center gap-1">
                        Method {getSortIcon('method')}
                      </div>
                    </th>
                    <th 
                      className="text-left py-3 px-4 cursor-pointer hover:bg-gray-100"
                      onClick={() => handleSort('host')}
                    >
                      <div className="flex items-center gap-1">
                        Host {getSortIcon('host')}
                      </div>
                    </th>
                    <th className="text-left py-3 px-4">URL</th>
                    <th 
                      className="text-left py-3 px-4 cursor-pointer hover:bg-gray-100"
                      onClick={() => handleSort('status_code')}
                    >
                      <div className="flex items-center gap-1">
                        Status {getSortIcon('status_code')}
                      </div>
                    </th>
                    <th 
                      className="text-left py-3 px-4 cursor-pointer hover:bg-gray-100"
                      onClick={() => handleSort('response_time')}
                    >
                      <div className="flex items-center gap-1">
                        Response Time {getSortIcon('response_time')}
                      </div>
                    </th>
                    <th 
                      className="text-left py-3 px-4 cursor-pointer hover:bg-gray-100"
                      onClick={() => handleSort('blocked')}
                    >
                      <div className="flex items-center gap-1">
                        Status {getSortIcon('blocked')}
                      </div>
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {currentPageLogs.map((log, index) => (
                    <tr key={`${log.timestamp}-${log.client_ip}-${log.url}-${index}`} className="border-b hover:bg-gray-50">
                      <td className="py-3 px-4 text-sm">
                        {format(new Date(log.timestamp), 'MMM dd, HH:mm:ss')}
                      </td>
                      <td className="py-3 px-4 font-mono text-sm">{log.client_ip}</td>
                      <td className="py-3 px-4">
                        <span className="px-2 py-1 bg-gray-100 rounded text-xs font-medium">
                          {log.method}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-sm font-medium">
                        {log.host || '-'}
                      </td>
                      <td className="py-3 px-4 text-sm max-w-xs truncate" title={log.url}>
                        {log.url}
                      </td>
                      <td className="py-3 px-4">
                        <span className={`font-medium ${getStatusColor(log.status_code)}`}>
                          {log.status_code}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-sm">{log.response_time}ms</td>
                      <td className="py-3 px-4">
                        {log.blocked ? (
                          <span className="px-2 py-1 bg-red-100 text-red-800 rounded text-xs">
                            Blocked
                          </span>
                        ) : (
                          <span className="px-2 py-1 bg-green-100 text-green-800 rounded text-xs">
                            Allowed
                          </span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Pagination Controls */}
            <div className="flex flex-col sm:flex-row justify-between items-center gap-4 mt-4 pt-4 border-t">
              <div className="text-sm text-gray-600">
                Showing {showingFrom} to {showingTo} of {totalLogs} entries
              </div>

              <div className="flex items-center gap-2">
                <button
                  onClick={() => setPage(1)}
                  disabled={page === 1}
                  className="btn btn-secondary px-2 py-1 disabled:opacity-50 disabled:cursor-not-allowed"
                  title="First page"
                >
                  <ChevronsLeft className="w-4 h-4" />
                </button>
                <button
                  onClick={() => setPage(Math.max(1, page - 1))}
                  disabled={page === 1}
                  className="btn btn-secondary px-2 py-1 disabled:opacity-50 disabled:cursor-not-allowed"
                  title="Previous page"
                >
                  <ChevronLeft className="w-4 h-4" />
                </button>
                
                <div className="flex items-center gap-1">
                  {[...new Array(Math.min(5, totalPages))].map((_, idx) => {
                    let pageNum
                    if (totalPages <= 5) {
                      pageNum = idx + 1
                    } else if (page <= 3) {
                      pageNum = idx + 1
                    } else if (page >= totalPages - 2) {
                      pageNum = totalPages - 4 + idx
                    } else {
                      pageNum = page - 2 + idx
                    }

                    return (
                      <button
                        key={pageNum}
                        onClick={() => setPage(pageNum)}
                        className={`px-3 py-1 text-sm rounded ${
                          page === pageNum
                            ? 'bg-primary-600 text-white'
                            : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                        }`}
                      >
                        {pageNum}
                      </button>
                    )
                  })}
                </div>

                <button
                  onClick={() => setPage(Math.min(totalPages, page + 1))}
                  disabled={page === totalPages}
                  className="btn btn-secondary px-2 py-1 disabled:opacity-50 disabled:cursor-not-allowed"
                  title="Next page"
                >
                  <ChevronRight className="w-4 h-4" />
                </button>
                <button
                  onClick={() => setPage(totalPages)}
                  disabled={page === totalPages}
                  className="btn btn-secondary px-2 py-1 disabled:opacity-50 disabled:cursor-not-allowed"
                  title="Last page"
                >
                  <ChevronsRight className="w-4 h-4" />
                </button>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  )
}

export default TrafficLogs
