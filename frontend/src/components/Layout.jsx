import { Link, useLocation } from 'react-router-dom'
import { Shield, LayoutDashboard, Server, Users, Activity, Settings } from 'lucide-react'

const Layout = ({ children }) => {
  const location = useLocation()

  const navigation = [
    { name: 'Dashboard', path: '/', icon: LayoutDashboard },
    { name: 'Virtual Hosts', path: '/vhosts', icon: Server },
    { name: 'IP Groups', path: '/ip-groups', icon: Users },
    { name: 'Traffic Logs', path: '/traffic', icon: Activity },
    { name: 'Settings', path: '/settings', icon: Settings },
  ]

  return (
    <div className="min-h-screen flex">
      {/* Sidebar */}
      <aside className="w-64 bg-gray-900 text-white">
        <div className="p-6">
          <div className="flex items-center gap-2 mb-8">
            <Shield className="w-8 h-8 text-primary-400" />
            <h1 className="text-xl font-bold">Docode WAF</h1>
          </div>

          <nav className="space-y-2">
            {navigation.map((item) => {
              const Icon = item.icon
              const isActive = location.pathname === item.path
              
              return (
                <Link
                  key={item.path}
                  to={item.path}
                  className={`flex items-center gap-3 px-4 py-3 rounded-lg transition-colors ${
                    isActive
                      ? 'bg-primary-600 text-white'
                      : 'text-gray-300 hover:bg-gray-800'
                  }`}
                >
                  <Icon className="w-5 h-5" />
                  <span>{item.name}</span>
                </Link>
              )
            })}
          </nav>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-auto">
        <div className="p-8">
          {children}
        </div>
      </main>
    </div>
  )
}

export default Layout
