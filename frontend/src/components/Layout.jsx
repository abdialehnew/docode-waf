import { Link, useLocation } from 'react-router-dom'
import { Shield, LayoutDashboard, Server, Users, Activity, Lock, Settings, LogOut, User, FileText, ChevronDown, ChevronRight } from 'lucide-react'
import { useAuth } from '../context/AuthContext'
import { useState, useEffect } from 'react'
import ConfirmModal from './ConfirmModal'
import api from '../services/api'

const Layout = ({ children }) => {
  const location = useLocation()
  const { user, logout } = useAuth()
  const [confirmModal, setConfirmModal] = useState({ isOpen: false, onConfirm: null })
  const [expandedMenus, setExpandedMenus] = useState({})
  const [appSettings, setAppSettings] = useState({
    app_name: 'Docode WAF',
    app_logo: null
  })

  useEffect(() => {
    loadAppSettings()
    // Auto-expand menu if on monitoring page
    if (location.pathname.startsWith('/monitoring')) {
      setExpandedMenus(prev => ({ ...prev, monitoring: true }))
    }
  }, [])

  const loadAppSettings = async () => {
    try {
      const response = await api.get('/settings/app')
      if (response.data) {
        setAppSettings(response.data)
        // Update document title
        document.title = response.data.app_name
      }
    } catch (error) {
      console.error('Failed to load app settings:', error)
    }
  }

  const toggleMenu = (menuKey) => {
    setExpandedMenus(prev => ({
      ...prev,
      [menuKey]: !prev[menuKey]
    }))
  }

  const navigation = [
    { name: 'Dashboard', path: '/', icon: LayoutDashboard },
    { name: 'Virtual Hosts', path: '/vhosts', icon: Server },
    { name: 'IP Groups', path: '/ip-groups', icon: Users },
    { name: 'Traffic Logs', path: '/traffic', icon: Activity },
    { name: 'SSL Certificates', path: '/certificates', icon: Lock },
    { 
      name: 'Monitoring', 
      icon: FileText,
      submenus: [
        { name: 'Logging', path: '/monitoring' }
      ]
    },
    { name: 'Settings', path: '/settings', icon: Settings },
  ]

  const handleLogout = () => {
    setConfirmModal({
      isOpen: true,
      title: 'Logout',
      message: 'Are you sure you want to logout?',
      type: 'warning',
      onConfirm: () => logout()
    })
  }

  return (
    <div className="min-h-screen flex bg-gray-50">
      {/* Sidebar */}
      <aside className="w-64 bg-gray-900 text-white flex flex-col fixed left-0 top-0 h-screen z-50">
        {/* Logo Section */}
        <div className="p-6 flex-shrink-0">
          <div className="flex items-center gap-2">
            {appSettings.app_logo ? (
              <img 
                src={appSettings.app_logo} 
                alt="Logo" 
                className="w-8 h-8 object-contain"
              />
            ) : (
              <Shield className="w-8 h-8 text-primary-400" />
            )}
            <h1 className="text-xl font-bold">{appSettings.app_name}</h1>
          </div>
        </div>

        {/* Navigation - Scrollable */}
        <div className="flex-1 overflow-y-auto px-6">
          <nav className="space-y-2">
            {navigation.map((item, index) => {
              const Icon = item.icon
              const menuKey = item.name.toLowerCase().replace(/\s+/g, '-')
              const isExpanded = expandedMenus[menuKey]
              
              // Item with submenus
              if (item.submenus) {
                const hasActiveSubmenu = item.submenus.some(sub => location.pathname === sub.path)
                
                return (
                  <div key={index}>
                    <button
                      onClick={() => toggleMenu(menuKey)}
                      className={`w-full flex items-center gap-3 px-4 py-3 rounded-lg transition-colors ${
                        hasActiveSubmenu
                          ? 'bg-gray-800 text-white'
                          : 'text-gray-300 hover:bg-gray-800'
                      }`}
                    >
                      <Icon className="w-5 h-5" />
                      <span className="flex-1 text-left">{item.name}</span>
                      {isExpanded ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
                    </button>
                    {isExpanded && (
                      <div className="ml-8 mt-1 space-y-1">
                        {item.submenus.map((submenu) => {
                          const isActive = location.pathname === submenu.path
                          return (
                            <Link
                              key={submenu.path}
                              to={submenu.path}
                              className={`block px-4 py-2 rounded-lg text-sm transition-colors ${
                                isActive
                                  ? 'bg-primary-600 text-white'
                                  : 'text-gray-400 hover:bg-gray-800 hover:text-gray-300'
                              }`}
                            >
                              {submenu.name}
                            </Link>
                          )
                        })}
                      </div>
                    )}
                  </div>
                )
              }
              
              // Regular menu item
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

        {/* User info and logout at bottom - Fixed */}
        <div className="p-6 border-t border-gray-800 flex-shrink-0">
          <div className="flex items-center gap-3 mb-3">
            <div className="w-10 h-10 bg-gray-700 rounded-full flex items-center justify-center">
              <User className="w-5 h-5 text-gray-300" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-white truncate">
                {user?.full_name || user?.username}
              </p>
              <p className="text-xs text-gray-400 truncate">{user?.email}</p>
            </div>
          </div>
          <button
            onClick={handleLogout}
            className="w-full flex items-center gap-2 px-4 py-2 text-sm text-gray-300 hover:bg-gray-800 rounded-lg transition-colors"
          >
            <LogOut className="w-4 h-4" />
            <span>Logout</span>
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 ml-64">
        <div className="p-8">
          {children}
        </div>
      </main>

      {/* Confirm Modal */}
      <ConfirmModal
        isOpen={confirmModal.isOpen}
        onClose={() => setConfirmModal({ ...confirmModal, isOpen: false })}
        onConfirm={confirmModal.onConfirm}
        title={confirmModal.title}
        message={confirmModal.message}
        type={confirmModal.type}
      />
    </div>
  )
}

export default Layout
