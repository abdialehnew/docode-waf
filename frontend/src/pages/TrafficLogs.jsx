import { useEffect, useState, useMemo } from 'react'
import { getTrafficLogs, getTrafficCountries, getTrafficIPs } from '../services/api'
import { format } from 'date-fns'
import { RefreshCw, ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Search, X, ChevronDown } from 'lucide-react'
import logger from '../utils/logger'

// Country code to flag emoji helper
const getCountryFlag = (countryCode) => {
  if (!countryCode || countryCode.length !== 2) return '🌐'
  const codePoints = countryCode
    .toUpperCase()
    .split('')
    .map(char => 127397 + char.charCodeAt(0))
  return String.fromCodePoint(...codePoints)
}

// Country code to name mapping
const countryNames = {
  'ID': 'Indonesia', 'US': 'United States', 'CN': 'China', 'JP': 'Japan', 'KR': 'South Korea',
  'SG': 'Singapore', 'MY': 'Malaysia', 'TH': 'Thailand', 'VN': 'Vietnam', 'PH': 'Philippines',
  'AU': 'Australia', 'IN': 'India', 'DE': 'Germany', 'FR': 'France', 'GB': 'United Kingdom',
  'NL': 'Netherlands', 'RU': 'Russia', 'BR': 'Brazil', 'CA': 'Canada', 'HK': 'Hong Kong',
  'TW': 'Taiwan', 'IT': 'Italy', 'ES': 'Spain', 'PL': 'Poland', 'UA': 'Ukraine',
  'XX': 'Unknown', 'EU': 'Europe', 'AP': 'Asia Pacific'
}

const getCountryName = (code) => countryNames[code] || code

// Searchable multi-select dropdown component
const MultiSelectDropdown = ({ options, selected, onChange, placeholder, disabled }) => {
  const [isOpen, setIsOpen] = useState(false)
  const [searchTerm, setSearchTerm] = useState('')

  const filteredOptions = useMemo(() => {
    if (!searchTerm) return options
    return options.filter(opt => 
      opt.label.toLowerCase().includes(searchTerm.toLowerCase()) ||
      opt.value.toLowerCase().includes(searchTerm.toLowerCase())
    )
  }, [options, searchTerm])

  const toggleOption = (value) => {
    if (selected.includes(value)) {
      onChange(selected.filter(v => v !== value))
    } else {
      onChange([...selected, value])
    }
  }

  const removeSelected = (value, e) => {
    e.stopPropagation()
    onChange(selected.filter(v => v !== value))
  }

  return (
    <div className="relative">
      <div 
        className={`input min-h-[42px] flex flex-wrap gap-1 items-center cursor-pointer ${disabled ? 'bg-gray-100 cursor-not-allowed' : ''}`}
        onClick={() => !disabled && setIsOpen(!isOpen)}
      >
        {selected.length > 0 ? (
          <>
            {selected.slice(0, 3).map(val => {
              const opt = options.find(o => o.value === val)
              return (
                <span key={val} className="bg-primary-100 text-primary-800 text-xs px-2 py-1 rounded flex items-center gap-1">
                  {opt?.flag} {opt?.label || val}
                  <X className="w-3 h-3 cursor-pointer hover:text-red-600" onClick={(e) => removeSelected(val, e)} />
                </span>
              )
            })}
            {selected.length > 3 && (
              <span className="text-xs text-gray-500">+{selected.length - 3} more</span>
            )}
          </>
        ) : (
          <span className="text-gray-400">{placeholder}</span>
        )}
        <ChevronDown className={`w-4 h-4 ml-auto text-gray-400 transition-transform ${isOpen ? 'rotate-180' : ''}`} />
      </div>
      
      {isOpen && !disabled && (
        <div className="absolute z-50 mt-1 w-full bg-white border rounded-lg shadow-lg max-h-64 overflow-hidden">
          <div className="p-2 border-b">
            <input
              type="text"
              placeholder="Search..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="input w-full py-1 text-sm"
              onClick={(e) => e.stopPropagation()}
            />
          </div>
          <div className="max-h-48 overflow-y-auto">
            {filteredOptions.length === 0 ? (
              <div className="p-3 text-center text-gray-500 text-sm">No options found</div>
            ) : (
              filteredOptions.map(opt => (
                <div
                  key={opt.value}
                  className={`px-3 py-2 flex items-center gap-2 cursor-pointer hover:bg-gray-100 ${selected.includes(opt.value) ? 'bg-primary-50' : ''}`}
                  onClick={() => toggleOption(opt.value)}
                >
                  <input
                    type="checkbox"
                    checked={selected.includes(opt.value)}
                    onChange={() => {}}
                    className="rounded text-primary-600"
                  />
                  {opt.flag && <span>{opt.flag}</span>}
                  <span className="flex-1 text-sm">{opt.label}</span>
                  <span className="text-xs text-gray-400">{opt.count}</span>
                </div>
              ))
            )}
          </div>
          {selected.length > 0 && (
            <div className="border-t p-2">
              <button
                className="text-xs text-red-600 hover:text-red-800"
                onClick={(e) => { e.stopPropagation(); onChange([]); }}
              >
                Clear all
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

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

  // Filter states
  const [countries, setCountries] = useState([])
  const [selectedCountries, setSelectedCountries] = useState([])
  const [ips, setIPs] = useState([])
  const [selectedIPs, setSelectedIPs] = useState([])

  useEffect(() => {
    loadLogs()
    loadCountries()
  }, [])

  useEffect(() => {
    if (selectedCountries.length > 0) {
      loadIPs()
    } else {
      setIPs([])
      setSelectedIPs([])
    }
  }, [selectedCountries])

  useEffect(() => {
    filterAndSortLogs()
  }, [logs, searchTerm, sortField, sortOrder, selectedCountries, selectedIPs])

  const loadLogs = async () => {
    setLoading(true)
    try {
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

  const loadCountries = async () => {
    try {
      const response = await getTrafficCountries()
      setCountries(response.data?.data || [])
    } catch (error) {
      logger.error('Failed to load countries:', error)
    }
  }

  const loadIPs = async () => {
    try {
      const response = await getTrafficIPs(selectedCountries)
      setIPs(response.data?.data || [])
    } catch (error) {
      logger.error('Failed to load IPs:', error)
    }
  }

  const filterAndSortLogs = () => {
    let filtered = [...logs]

    // Country filter
    if (selectedCountries.length > 0) {
      filtered = filtered.filter(log => 
        selectedCountries.includes(log.country_code)
      )
    }

    // IP filter
    if (selectedIPs.length > 0) {
      filtered = filtered.filter(log => 
        selectedIPs.includes(log.client_ip)
      )
    }

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

  const getStatusText = (status) => {
    const statusTexts = {
      200: 'OK', 201: 'Created', 204: 'No Content',
      301: 'Moved', 302: 'Found', 304: 'Not Modified',
      400: 'Bad Request', 401: 'Unauthorized', 403: 'Forbidden', 404: 'Not Found',
      500: 'Server Error', 502: 'Bad Gateway', 503: 'Service Unavailable', 504: 'Gateway Timeout'
    }
    return statusTexts[status] || ''
  }

  // Prepare country options for dropdown
  const countryOptions = useMemo(() => 
    countries.map(c => ({
      value: c.country_code,
      label: getCountryName(c.country_code),
      flag: getCountryFlag(c.country_code),
      count: c.count
    }))
  , [countries])

  // Prepare IP options for dropdown
  const ipOptions = useMemo(() => 
    ips.map(ip => ({
      value: ip.client_ip,
      label: ip.client_ip,
      flag: getCountryFlag(ip.country_code),
      count: ip.count
    }))
  , [ips])

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

  const clearFilters = () => {
    setSelectedCountries([])
    setSelectedIPs([])
    setSearchTerm('')
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
        {/* Filters Row */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4 pb-4 border-b">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Country</label>
            <MultiSelectDropdown
              options={countryOptions}
              selected={selectedCountries}
              onChange={setSelectedCountries}
              placeholder="Select countries..."
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">IP Source</label>
            <MultiSelectDropdown
              options={ipOptions}
              selected={selectedIPs}
              onChange={setSelectedIPs}
              placeholder={selectedCountries.length > 0 ? "Select IPs..." : "Select country first..."}
              disabled={selectedCountries.length === 0}
            />
          </div>
          <div className="flex items-end">
            {(selectedCountries.length > 0 || selectedIPs.length > 0 || searchTerm) && (
              <button onClick={clearFilters} className="btn btn-secondary text-sm">
                Clear Filters
              </button>
            )}
          </div>
        </div>

        {/* Search and Controls */}
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-4 pb-4 border-b">
          <div className="flex items-center gap-2">
            <label htmlFor="rows-per-page" className="text-sm text-gray-600">Show</label>
            <select
              id="rows-per-page"
              value={rowsPerPage}
              onChange={(e) => handleRowsPerPageChange(Number(e.target.value))}
              className="input py-1 px-2 text-sm"
            >
              <option value={5}>5</option>
              <option value={10}>10</option>
              <option value={50}>50</option>
              <option value={100}>100</option>
            </select>
            <span className="text-sm text-gray-600">entries</span>
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

        {filteredLogs.length === 0 ? (
          <div className="text-center py-12 text-gray-500">
            <p className="text-lg mb-2">No traffic logs found</p>
            <p className="text-sm">
              {(searchTerm || selectedCountries.length > 0 || selectedIPs.length > 0) && 'Try adjusting your filters'}
              {!searchTerm && selectedCountries.length === 0 && 'Traffic logs will appear here once requests are made to your virtual hosts'}
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
                        HTTP Code {getSortIcon('status_code')}
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
                      <td className="py-3 px-4">
                        <div className="flex items-center gap-2">
                          <span className="text-lg" title={getCountryName(log.country_code)}>
                            {getCountryFlag(log.country_code)}
                          </span>
                          <span className="font-mono text-sm">{log.client_ip}</span>
                        </div>
                      </td>
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
                        <div className="flex flex-col">
                          <span className={`font-medium ${getStatusColor(log.status_code)}`}>
                            {log.status_code}
                          </span>
                          <span className="text-xs text-gray-400">{getStatusText(log.status_code)}</span>
                        </div>
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
                  {Array.from({ length: Math.min(5, totalPages) }, (_, idx) => {
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
