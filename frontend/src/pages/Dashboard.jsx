import { useEffect, useState } from 'react'
import { getDashboardStats } from '../services/api'
import { Activity, Shield, AlertTriangle, Globe } from 'lucide-react'
import { LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'
import { format } from 'date-fns'

const Dashboard = () => {
  const [stats, setStats] = useState(null)
  const [loading, setLoading] = useState(true)
  const [timeRange, setTimeRange] = useState('24h')

  useEffect(() => {
    loadStats()
  }, [timeRange])

  const loadStats = async () => {
    try {
      const response = await getDashboardStats(timeRange)
      setStats(response.data)
    } catch (error) {
      console.error('Failed to load stats:', error)
    } finally {
      setLoading(false)
    }
  }

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
    <div>
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">Dashboard</h1>
        <select
          value={timeRange}
          onChange={(e) => setTimeRange(e.target.value)}
          className="input w-auto"
        >
          <option value="1h">Last Hour</option>
          <option value="6h">Last 6 Hours</option>
          <option value="24h">Last 24 Hours</option>
          <option value="7d">Last 7 Days</option>
          <option value="30d">Last 30 Days</option>
        </select>
      </div>

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
                tickFormatter={(value) => format(new Date(value), 'HH:mm')}
              />
              <YAxis />
              <Tooltip labelFormatter={(value) => format(new Date(value), 'MMM dd, HH:mm')} />
              <Legend />
              <Line type="monotone" dataKey="count" stroke="#3b82f6" name="Requests" />
              <Line type="monotone" dataKey="blocked" stroke="#ef4444" name="Blocked" />
            </LineChart>
          </ResponsiveContainer>
        </div>

        <div className="card">
          <h2 className="text-xl font-semibold mb-4">Top Attack Types</h2>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={stats?.top_attack_types || []}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="type" />
              <YAxis />
              <Tooltip />
              <Bar dataKey="count" fill="#ef4444" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Recent Attacks */}
      <div className="card">
        <h2 className="text-xl font-semibold mb-4">Recent Attacks</h2>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b">
                <th className="text-left py-3 px-4">Time</th>
                <th className="text-left py-3 px-4">IP Address</th>
                <th className="text-left py-3 px-4">Attack Type</th>
                <th className="text-left py-3 px-4">Severity</th>
                <th className="text-left py-3 px-4">Description</th>
              </tr>
            </thead>
            <tbody>
              {stats?.recent_attacks?.map((attack, index) => (
                <tr key={index} className="border-b hover:bg-gray-50">
                  <td className="py-3 px-4">
                    {format(new Date(attack.timestamp), 'MMM dd, HH:mm:ss')}
                  </td>
                  <td className="py-3 px-4 font-mono">{attack.client_ip}</td>
                  <td className="py-3 px-4">
                    <span className="px-2 py-1 bg-red-100 text-red-800 rounded text-sm">
                      {attack.attack_type}
                    </span>
                  </td>
                  <td className="py-3 px-4">
                    <span className={`px-2 py-1 rounded text-sm ${
                      attack.severity === 'high' ? 'bg-red-100 text-red-800' :
                      attack.severity === 'medium' ? 'bg-orange-100 text-orange-800' :
                      'bg-yellow-100 text-yellow-800'
                    }`}>
                      {attack.severity}
                    </span>
                  </td>
                  <td className="py-3 px-4 text-sm text-gray-600">{attack.description}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

export default Dashboard
