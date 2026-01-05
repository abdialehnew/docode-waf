import { useEffect, useState } from 'react'
import { getDashboardStats, getAttacksByCountry } from '../services/api'
import { Activity, Shield, AlertTriangle, Globe, Search, Calendar, Loader2 } from 'lucide-react'
import { LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'
import { format, subDays, differenceInDays } from 'date-fns'
import logger from '../utils/logger'

// Country code to flag emoji mapping
const getCountryFlag = (countryCode) => {
  if (!countryCode || countryCode === 'XX' || countryCode === '--') return '🌐'
  const codePoints = countryCode
    .toUpperCase()
    .split('')
    .map(char => 127397 + char.codePointAt(0))
  return String.fromCodePoint(...codePoints)
}

// Country code to name mapping (common countries)
const countryNames = {
  'US': 'United States', 'ID': 'Indonesia', 'CN': 'China', 'RU': 'Russia',
  'BR': 'Brazil', 'IN': 'India', 'GB': 'United Kingdom', 'DE': 'Germany',
  'FR': 'France', 'JP': 'Japan', 'KR': 'South Korea', 'SG': 'Singapore',
  'MY': 'Malaysia', 'TH': 'Thailand', 'VN': 'Vietnam', 'PH': 'Philippines',
  'IR': 'Iran', 'TR': 'Turkey', 'UA': 'Ukraine', 'MX': 'Mexico',
  'NG': 'Nigeria', 'PK': 'Pakistan', 'EG': 'Egypt',
  'XX': 'Unknown', '--': 'Unknown'
}

const Dashboard = () => {
  const [stats, setStats] = useState(null)
  const [countryStats, setCountryStats] = useState([])
  const [loading, setLoading] = useState(true)
  const [fetching, setFetching] = useState(false)
  const [timeRange, setTimeRange] = useState('24h')
  const [customRange, setCustomRange] = useState(false)
  const [startDate, setStartDate] = useState(format(subDays(new Date(), 7), 'yyyy-MM-dd'))
  const [endDate, setEndDate] = useState(format(new Date(), 'yyyy-MM-dd'))
  const [searchTerm, setSearchTerm] = useState('')
  const [currentPage, setCurrentPage] = useState(1)
  const itemsPerPage = 10

  useEffect(() => {
    loadStats()
  }, [timeRange])

  const loadStats = async () => {
    setFetching(true)
    try {
      let params = {}
      if (timeRange === 'custom' && startDate && endDate) {
        params = { start: startDate, end: endDate }
      } else {
        params = { range: timeRange }
      }
      const response = await getDashboardStats(params)
      setStats(response.data)
      
      // Load country stats with same params
      const countryResponse = await getAttacksByCountry(params)
      setCountryStats(countryResponse.data || [])
    } catch (error) {
      logger.error('Failed to load stats:', error)
    } finally {
      setLoading(false)
      setFetching(false)
    }
  }

  const handleRangeChange = (range) => {
    setTimeRange(range)
    setCustomRange(range === 'custom')
    if (range !== 'custom') {
      loadStats()
    }
  }

  const handleCustomRangeApply = () => {
    if (startDate && endDate) {
      loadStats()
    }
  }

  // Determine if date range is more than 1 day
  const isMultipleDays = () => {
    if (timeRange === 'custom' && startDate && endDate) {
      return differenceInDays(new Date(endDate), new Date(startDate)) > 0
    }
    // For predefined ranges
    return timeRange !== '1h' && timeRange !== '24h'
  }

  // Format X-axis based on date range
  const formatXAxis = (value) => {
    if (isMultipleDays()) {
      return format(new Date(value), 'MMM dd') // Show date only for multiple days
    }
    return format(new Date(value), 'HH:mm') // Show time for single day
  }

  // Format tooltip label
  const formatTooltipLabel = (value) => {
    if (isMultipleDays()) {
      return format(new Date(value), 'MMM dd, yyyy')
    }
    return format(new Date(value), 'MMM dd, HH:mm')
  }


  // Filter and paginate attacks
  const filteredAttacks = stats?.recent_attacks?.filter(attack => 
    attack.client_ip?.toLowerCase().includes(searchTerm.toLowerCase()) ||
    attack.attack_type?.toLowerCase().includes(searchTerm.toLowerCase()) ||
    attack.url?.toLowerCase().includes(searchTerm.toLowerCase())
  ) || []

  const totalPages = Math.ceil(filteredAttacks.length / itemsPerPage)
  const paginatedAttacks = filteredAttacks.slice(
    (currentPage - 1) * itemsPerPage,
    currentPage * itemsPerPage
  )
  if (loading) {
    return <div className="text-center py-12">Loading...</div>
  }

  const statCards = [
    {
      title: 'Total Requests',
      value: stats?.total_requests?.toLocaleString() || '0',
      icon: Activity,
      color: 'blue',
    },
    {
      title: 'Blocked Requests',
      value: stats?.blocked_requests?.toLocaleString() || '0',
      icon: Shield,
      color: 'red',
    },
    {
      title: 'Total Attacks',
      value: stats?.total_attacks?.toLocaleString() || '0',
      icon: AlertTriangle,
      color: 'orange',
    },
    {
      title: 'Unique IPs',
      value: stats?.unique_ips?.toLocaleString() || '0',
      icon: Globe,
      color: 'green',
    },
  ]

  return (
    <div className="w-full">
      {/* Header with Date Range Selector */}
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Dashboard</h1>
        
        <div className="flex items-center gap-4">
          {/* Time Range Selector */}
          <div className="flex items-center gap-2 bg-white border rounded-lg p-1">
            {['1h', '24h', '7d', '30d', 'custom'].map((range) => (
              <button
                key={range}
                onClick={() => handleRangeChange(range)}
                className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                  timeRange === range
                    ? 'bg-blue-500 text-white'
                    : 'text-gray-600 hover:bg-gray-100'
                }`}
              >
                {range === 'custom' ? <Calendar className="w-4 h-4" /> : range.toUpperCase()}
              </button>
            ))}
          </div>

          {/* Loading Indicator */}
          {fetching && (
            <div className="flex items-center gap-2 text-blue-600 bg-blue-50 px-4 py-2 rounded-lg">
              <Loader2 className="w-4 h-4 animate-spin" />
              <span className="text-sm font-medium">Loading...</span>
            </div>
          )}
        </div>
      </div>

      {/* Custom Date Range Picker */}
      {customRange && (
        <div className="mb-6 bg-white border rounded-lg p-4 shadow-sm">
          <div className="flex items-end gap-4">
            <div className="flex-1">
              <label htmlFor="start-date" className="block text-sm font-medium text-gray-700 mb-2">
                Start Date
              </label>
              <input
                id="start-date"
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                max={endDate}
                className="w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div className="flex-1">
              <label htmlFor="end-date" className="block text-sm font-medium text-gray-700 mb-2">
                End Date
              </label>
              <input
                id="end-date"
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
                min={startDate}
                max={format(new Date(), 'yyyy-MM-dd')}
                className="w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <button
              onClick={handleCustomRangeApply}
              disabled={!startDate || !endDate || fetching}
              className="px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              Apply
            </button>
          </div>
        </div>
      )}

      {/* Overlay Loader */}
      {fetching && (
        <div className="fixed inset-0 bg-black bg-opacity-10 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl p-8 flex flex-col items-center gap-4">
            <Loader2 className="w-12 h-12 text-blue-500 animate-spin" />
            <p className="text-lg font-medium text-gray-700">Fetching data...</p>
          </div>
        </div>
      )}

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {statCards.map((stat) => {
          const Icon = stat.icon
          return (
            <div key={stat.title} className="card">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600 mb-1">{stat.title}</p>
                  <p className="text-3xl font-bold">{stat.value}</p>
                </div>
                <div className={`p-3 rounded-lg bg-${stat.color}-100`}>
                  <Icon className={`w-6 h-6 text-${stat.color}-600`} />
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        <div className="card">
          <h2 className="text-xl font-semibold mb-4">Traffic Over Time</h2>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={stats?.requests_by_hour || []}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis 
                dataKey="hour" 
                tickFormatter={formatXAxis}
              />
              <YAxis />
              <Tooltip labelFormatter={formatTooltipLabel} />
              <Legend />
              <Line type="monotone" dataKey="count" stroke="#3b82f6" name="Requests" />
              <Line type="monotone" dataKey="blocked" stroke="#ef4444" name="Blocked" />
            </LineChart>
          </ResponsiveContainer>
        </div>

        <div className="card">
          <h2 className="text-xl font-semibold mb-4">Top Attack Types</h2>
          <ResponsiveContainer width="100%" height={400}>
            <BarChart data={stats?.top_attack_types || []} margin={{ top: 20, right: 30, left: 20, bottom: 80 }}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis 
                dataKey="name" 
                angle={-45} 
                textAnchor="end" 
                height={100}
                interval={0}
                style={{ fontSize: '12px' }}
              />
              <YAxis />
              <Tooltip />
              <Legend />
              <Bar dataKey="value" fill="#ef4444" name="Attacks" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Attacks by Country */}
      <div className="card">
        <h2 className="text-xl font-semibold mb-4">Attacks by Country</h2>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Country
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Total Attacks
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Blocked
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Unique IPs
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Attack Types
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Block Rate
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {countryStats.length === 0 ? (
                <tr>
                  <td colSpan="6" className="px-4 py-8 text-center text-gray-500">
                    No attack data available for selected time range
                  </td>
                </tr>
              ) : (
                countryStats.map((country, idx) => {
                  const totalAttacks = Number.parseInt(country.total_attacks) || 0
                  const blockedAttacks = Number.parseInt(country.blocked_attacks) || 0
                  const blockRate = totalAttacks > 0 ? ((blockedAttacks / totalAttacks) * 100).toFixed(1) : 0
                  const countryCode = country.country_code || 'XX'
                  const countryName = countryNames[countryCode] || countryCode

                  return (
                    <tr key={idx} className="hover:bg-gray-50">
                      <td className="px-4 py-3 whitespace-nowrap">
                        <div className="flex items-center gap-2">
                          <span className="text-2xl">{getCountryFlag(countryCode)}</span>
                          <div>
                            <div className="text-sm font-medium text-gray-900">{countryName}</div>
                            <div className="text-xs text-gray-500">{countryCode}</div>
                          </div>
                        </div>
                      </td>
                      <td className="px-4 py-3 whitespace-nowrap">
                        <span className="text-sm font-semibold text-red-600">
                          {totalAttacks.toLocaleString()}
                        </span>
                      </td>
                      <td className="px-4 py-3 whitespace-nowrap">
                        <span className="text-sm text-gray-900">
                          {blockedAttacks.toLocaleString()}
                        </span>
                      </td>
                      <td className="px-4 py-3 whitespace-nowrap">
                        <span className="text-sm text-gray-600">
                          {country.unique_ips}
                        </span>
                      </td>
                      <td className="px-4 py-3 whitespace-nowrap">
                        <span className="text-sm text-gray-600">
                          {country.attack_types}
                        </span>
                      </td>
                      <td className="px-4 py-3 whitespace-nowrap">
                        <div className="flex items-center gap-2">
                          <div className="w-full bg-gray-200 rounded-full h-2 max-w-[80px]">
                            <div
                              className={`h-2 rounded-full ${
                                blockRate >= 80 ? 'bg-green-500' :
                                blockRate >= 50 ? 'bg-yellow-500' :
                                'bg-red-500'
                              }`}
                              style={{ width: `${blockRate}%` }}
                            />
                          </div>
                          <span className="text-sm text-gray-600 min-w-[45px]">{blockRate}%</span>
                        </div>
                      </td>
                    </tr>
                  )
                })
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Recent Attacks */}
      <div className="card">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-semibold">Recent Attacks</h2>
          <div className="relative w-64">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
            <input
              type="text"
              placeholder="Search attacks..."
              value={searchTerm}
              onChange={(e) => {
                setSearchTerm(e.target.value)
                setCurrentPage(1)
              }}
              className="pl-10 pr-4 py-2 border rounded-lg w-full focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr className="border-b">
                <th className="text-left py-3 px-4 font-medium text-gray-700">Time</th>
                <th className="text-left py-3 px-4 font-medium text-gray-700">IP Address</th>
                <th className="text-left py-3 px-4 font-medium text-gray-700">Country</th>
                <th className="text-left py-3 px-4 font-medium text-gray-700">Host</th>
                <th className="text-left py-3 px-4 font-medium text-gray-700">Attack Type</th>
                <th className="text-left py-3 px-4 font-medium text-gray-700">URL</th>
                <th className="text-left py-3 px-4 font-medium text-gray-700">Status</th>
              </tr>
            </thead>
            <tbody>
              {paginatedAttacks.length > 0 ? (
                paginatedAttacks.map((attack) => (
                  <tr key={`${attack.timestamp}-${attack.client_ip}-${attack.attack_type}`} className="border-b hover:bg-gray-50 transition-colors">
                    <td className="py-3 px-4 text-sm">
                      {format(new Date(attack.timestamp), 'MMM dd, HH:mm:ss')}
                    </td>
                    <td className="py-3 px-4 font-mono text-sm">{attack.client_ip}</td>
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-2">
                        <span className="text-2xl">{getCountryFlag(attack.country_code || 'XX')}</span>
                        <span className="text-sm text-gray-600">
                          {countryNames[attack.country_code] || attack.country_code || 'Unknown'}
                        </span>
                      </div>
                    </td>
                    <td className="py-3 px-4 text-sm text-gray-900 font-medium">
                      {attack.host || '-'}
                    </td>
                    <td className="py-3 px-4">
                      <span className="px-3 py-1 bg-red-100 text-red-800 rounded-full text-xs font-medium">
                        {attack.attack_type}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-sm text-gray-600 max-w-xs truncate" title={attack.url}>
                      {attack.url}
                    </td>
                    <td className="py-3 px-4">
                      <span className={`px-3 py-1 rounded-full text-xs font-medium ${
                        attack.blocked ? 'bg-red-100 text-red-800' : 'bg-green-100 text-green-800'
                      }`}>
                        {attack.blocked ? 'Blocked' : 'Allowed'}
                      </span>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan="7" className="py-8 text-center text-gray-500">
                    No attacks found
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        
        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex justify-between items-center mt-4 pt-4 border-t">
            <div className="text-sm text-gray-600">
              Showing {(currentPage - 1) * itemsPerPage + 1} to {Math.min(currentPage * itemsPerPage, filteredAttacks.length)} of {filteredAttacks.length} attacks
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                disabled={currentPage === 1}
                className="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Previous
              </button>
              {Array.from({ length: totalPages }, (_, i) => (
                <button
                  key={i + 1}
                  onClick={() => setCurrentPage(i + 1)}
                  className={`px-3 py-1 border rounded ${
                    currentPage === i + 1 ? 'bg-blue-500 text-white' : 'hover:bg-gray-50'
                  }`}
                >
                  {i + 1}
                </button>
              ))}
              <button
                onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                disabled={currentPage === totalPages}
                className="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Next
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default Dashboard
