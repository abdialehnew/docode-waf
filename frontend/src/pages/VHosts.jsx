import { useEffect, useState } from 'react'
import { getVHosts, createVHost, deleteVHost } from '../services/api'
import { Plus, Trash2, Edit, Server } from 'lucide-react'

const VHosts = () => {
  const [vhosts, setVHosts] = useState([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [formData, setFormData] = useState({
    name: '',
    domain: '',
    backend_url: '',
    ssl_enabled: false,
    enabled: true,
  })

  useEffect(() => {
    loadVHosts()
  }, [])

  const loadVHosts = async () => {
    try {
      const response = await getVHosts()
      setVHosts(response.data)
    } catch (error) {
      console.error('Failed to load vhosts:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    try {
      await createVHost(formData)
      setShowModal(false)
      setFormData({
        name: '',
        domain: '',
        backend_url: '',
        ssl_enabled: false,
        enabled: true,
      })
      loadVHosts()
    } catch (error) {
      console.error('Failed to create vhost:', error)
    }
  }

  const handleDelete = async (id) => {
    if (!confirm('Are you sure you want to delete this virtual host?')) return
    
    try {
      await deleteVHost(id)
      loadVHosts()
    } catch (error) {
      console.error('Failed to delete vhost:', error)
    }
  }

  if (loading) {
    return <div className="text-center py-12">Loading...</div>
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">Virtual Hosts</h1>
        <button
          onClick={() => setShowModal(true)}
          className="btn btn-primary flex items-center gap-2"
        >
          <Plus className="w-5 h-5" />
          Add Virtual Host
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {vhosts.map((vhost) => (
          <div key={vhost.id} className="card">
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-primary-100 rounded-lg">
                  <Server className="w-6 h-6 text-primary-600" />
                </div>
                <div>
                  <h3 className="font-semibold">{vhost.name}</h3>
                  <p className="text-sm text-gray-600">{vhost.domain}</p>
                </div>
              </div>
              <button
                onClick={() => handleDelete(vhost.id)}
                className="text-red-600 hover:text-red-800"
              >
                <Trash2 className="w-5 h-5" />
              </button>
            </div>

            <div className="space-y-2 text-sm">
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
                <span className={`px-2 py-1 rounded text-xs ${
                  vhost.enabled
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

      {/* Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h2 className="text-2xl font-bold mb-4">Add Virtual Host</h2>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="label">Name</label>
                <input
                  type="text"
                  className="input"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                />
              </div>

              <div>
                <label className="label">Domain</label>
                <input
                  type="text"
                  className="input"
                  placeholder="example.com"
                  value={formData.domain}
                  onChange={(e) => setFormData({ ...formData, domain: e.target.value })}
                  required
                />
              </div>

              <div>
                <label className="label">Backend URL</label>
                <input
                  type="text"
                  className="input"
                  placeholder="http://localhost:8000"
                  value={formData.backend_url}
                  onChange={(e) => setFormData({ ...formData, backend_url: e.target.value })}
                  required
                />
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="ssl"
                  checked={formData.ssl_enabled}
                  onChange={(e) => setFormData({ ...formData, ssl_enabled: e.target.checked })}
                />
                <label htmlFor="ssl" className="text-sm">Enable SSL</label>
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="enabled"
                  checked={formData.enabled}
                  onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
                />
                <label htmlFor="enabled" className="text-sm">Enable Virtual Host</label>
              </div>

              <div className="flex gap-2 pt-4">
                <button type="submit" className="btn btn-primary flex-1">
                  Create
                </button>
                <button
                  type="button"
                  onClick={() => setShowModal(false)}
                  className="btn btn-secondary flex-1"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}

export default VHosts
