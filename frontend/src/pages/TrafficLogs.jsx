import { useEffect, useState } from 'react'
import { getTrafficLogs } from '../services/api'
import { format } from 'date-fns'
import { RefreshCw } from 'lucide-react'

const TrafficLogs = () => {
  const [logs, setLogs] = useState([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const limit = 50

  useEffect(() => {
    loadLogs()
  }, [page])

  const loadLogs = async () => {
    setLoading(true)
    try {
      const offset = (page - 1) * limit
      const response = await getTrafficLogs(limit, offset)
      setLogs(response.data)
    } catch (error) {
      console.error('Failed to load traffic logs:', error)
    } finally {
      setLoading(false)
    }
  }

  const getStatusColor = (status) => {
    if (status >= 200 && status < 300) return 'text-green-600'
    if (status >= 300 && status < 400) return 'text-blue-600'
    if (status >= 400 && status < 500) return 'text-orange-600'
    if (status >= 500) return 'text-red-600'
    return 'text-gray-600'
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

      <div className="card overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b">
              <th className="text-left py-3 px-4">Time</th>
              <th className="text-left py-3 px-4">IP</th>
              <th className="text-left py-3 px-4">Method</th>
              <th className="text-left py-3 px-4">URL</th>
              <th className="text-left py-3 px-4">Status</th>
              <th className="text-left py-3 px-4">Response Time</th>
              <th className="text-left py-3 px-4">Blocked</th>
            </tr>
          </thead>
          <tbody>
            {logs.map((log, index) => (
              <tr key={index} className="border-b hover:bg-gray-50">
                <td className="py-3 px-4 text-sm">
                  {format(new Date(log.timestamp), 'MMM dd, HH:mm:ss')}
                </td>
                <td className="py-3 px-4 font-mono text-sm">{log.client_ip}</td>
                <td className="py-3 px-4">
                  <span className="px-2 py-1 bg-gray-100 rounded text-xs font-medium">
                    {log.method}
                  </span>
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

        <div className="flex justify-between items-center mt-4 pt-4 border-t">
          <button
            onClick={() => setPage(Math.max(1, page - 1))}
            disabled={page === 1}
            className="btn btn-secondary disabled:opacity-50"
          >
            Previous
          </button>
          <span className="text-sm text-gray-600">Page {page}</span>
          <button
            onClick={() => setPage(page + 1)}
            disabled={logs.length < limit}
            className="btn btn-secondary disabled:opacity-50"
          >
            Next
          </button>
        </div>
      </div>
    </div>
  )
}

export default TrafficLogs
