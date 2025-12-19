import { useEffect, useState } from 'react'
import { getIPGroups, createIPGroup, deleteIPGroup, addIPToGroup, getGroupIPs, removeIPFromGroup } from '../services/api'
import { Plus, Trash2, Shield } from 'lucide-react'

const IPGroups = () => {
  const [groups, setGroups] = useState([])
  const [loading, setLoading] = useState(true)
  const [showGroupModal, setShowGroupModal] = useState(false)
  const [showIPModal, setShowIPModal] = useState(false)
  const [selectedGroup, setSelectedGroup] = useState(null)
  const [groupIPs, setGroupIPs] = useState([])
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    type: 'blacklist',
  })
  const [ipFormData, setIPFormData] = useState({
    ip_address: '',
    cidr_mask: null,
    description: '',
  })

  useEffect(() => {
    loadGroups()
  }, [])

  const loadGroups = async () => {
    try {
      const response = await getIPGroups()
      setGroups(response.data)
    } catch (error) {
      console.error('Failed to load IP groups:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleCreateGroup = async (e) => {
    e.preventDefault()
    try {
      await createIPGroup(formData)
      setShowGroupModal(false)
      setFormData({ name: '', description: '', type: 'blacklist' })
      loadGroups()
    } catch (error) {
      console.error('Failed to create group:', error)
    }
  }

  const handleDeleteGroup = async (id) => {
    if (!confirm('Are you sure you want to delete this IP group?')) return
    
    try {
      await deleteIPGroup(id)
      loadGroups()
    } catch (error) {
      console.error('Failed to delete group:', error)
    }
  }

  const handleViewIPs = async (group) => {
    setSelectedGroup(group)
    try {
      const response = await getGroupIPs(group.id)
      setGroupIPs(response.data)
    } catch (error) {
      console.error('Failed to load IPs:', error)
    }
  }

  const handleAddIP = async (e) => {
    e.preventDefault()
    try {
      await addIPToGroup(selectedGroup.id, ipFormData)
      setShowIPModal(false)
      setIPFormData({ ip_address: '', cidr_mask: null, description: '' })
      handleViewIPs(selectedGroup)
    } catch (error) {
      console.error('Failed to add IP:', error)
    }
  }

  const handleRemoveIP = async (ipId) => {
    try {
      await removeIPFromGroup(selectedGroup.id, ipId)
      handleViewIPs(selectedGroup)
    } catch (error) {
      console.error('Failed to remove IP:', error)
    }
  }

  if (loading) {
    return <div className="text-center py-12">Loading...</div>
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">IP Groups</h1>
        <button
          onClick={() => setShowGroupModal(true)}
          className="btn btn-primary flex items-center gap-2"
        >
          <Plus className="w-5 h-5" />
          Add IP Group
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {groups.map((group) => (
          <div key={group.id} className="card">
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center gap-3">
                <div className={`p-2 rounded-lg ${
                  group.type === 'blacklist' ? 'bg-red-100' : 'bg-green-100'
                }`}>
                  <Shield className={`w-6 h-6 ${
                    group.type === 'blacklist' ? 'text-red-600' : 'text-green-600'
                  }`} />
                </div>
                <div>
                  <h3 className="font-semibold">{group.name}</h3>
                  <p className="text-sm text-gray-600">{group.description}</p>
                </div>
              </div>
              <button
                onClick={() => handleDeleteGroup(group.id)}
                className="text-red-600 hover:text-red-800"
              >
                <Trash2 className="w-5 h-5" />
              </button>
            </div>

            <div className="space-y-2">
              <span className={`inline-block px-3 py-1 rounded text-sm ${
                group.type === 'blacklist'
                  ? 'bg-red-100 text-red-800'
                  : 'bg-green-100 text-green-800'
              }`}>
                {group.type}
              </span>
              
              <button
                onClick={() => handleViewIPs(group)}
                className="block w-full mt-2 text-sm text-primary-600 hover:text-primary-800"
              >
                View IPs →
              </button>
            </div>
          </div>
        ))}
      </div>

      {/* Group Modal */}
      {showGroupModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h2 className="text-2xl font-bold mb-4">Add IP Group</h2>
            <form onSubmit={handleCreateGroup} className="space-y-4">
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
                <label className="label">Description</label>
                <textarea
                  className="input"
                  rows="3"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                />
              </div>

              <div>
                <label className="label">Type</label>
                <select
                  className="input"
                  value={formData.type}
                  onChange={(e) => setFormData({ ...formData, type: e.target.value })}
                >
                  <option value="blacklist">Blacklist</option>
                  <option value="whitelist">Whitelist</option>
                </select>
              </div>

              <div className="flex gap-2 pt-4">
                <button type="submit" className="btn btn-primary flex-1">
                  Create
                </button>
                <button
                  type="button"
                  onClick={() => setShowGroupModal(false)}
                  className="btn btn-secondary flex-1"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* IP Details Modal */}
      {selectedGroup && !showIPModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-2xl">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-2xl font-bold">{selectedGroup.name} - IPs</h2>
              <button
                onClick={() => setShowIPModal(true)}
                className="btn btn-primary flex items-center gap-2"
              >
                <Plus className="w-5 h-5" />
                Add IP
              </button>
            </div>

            <div className="space-y-2 max-h-96 overflow-y-auto">
              {groupIPs.map((ip) => (
                <div key={ip.id} className="flex items-center justify-between p-3 bg-gray-50 rounded">
                  <div>
                    <p className="font-mono">
                      {ip.ip_address}{ip.cidr_mask ? `/${ip.cidr_mask}` : ''}
                    </p>
                    {ip.description && (
                      <p className="text-sm text-gray-600">{ip.description}</p>
                    )}
                  </div>
                  <button
                    onClick={() => handleRemoveIP(ip.id)}
                    className="text-red-600 hover:text-red-800"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              ))}
            </div>

            <button
              onClick={() => setSelectedGroup(null)}
              className="btn btn-secondary w-full mt-4"
            >
              Close
            </button>
          </div>
        </div>
      )}

      {/* Add IP Modal */}
      {showIPModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h2 className="text-2xl font-bold mb-4">Add IP Address</h2>
            <form onSubmit={handleAddIP} className="space-y-4">
              <div>
                <label className="label">IP Address</label>
                <input
                  type="text"
                  className="input"
                  placeholder="192.168.1.1"
                  value={ipFormData.ip_address}
                  onChange={(e) => setIPFormData({ ...ipFormData, ip_address: e.target.value })}
                  required
                />
              </div>

              <div>
                <label className="label">CIDR Mask (optional)</label>
                <input
                  type="number"
                  className="input"
                  placeholder="24"
                  min="0"
                  max="32"
                  value={ipFormData.cidr_mask || ''}
                  onChange={(e) => setIPFormData({ 
                    ...ipFormData, 
                    cidr_mask: e.target.value ? parseInt(e.target.value) : null 
                  })}
                />
                <p className="text-sm text-gray-600 mt-1">For IP blocks (e.g., /24 for 192.168.1.0/24)</p>
              </div>

              <div>
                <label className="label">Description</label>
                <input
                  type="text"
                  className="input"
                  value={ipFormData.description}
                  onChange={(e) => setIPFormData({ ...ipFormData, description: e.target.value })}
                />
              </div>

              <div className="flex gap-2 pt-4">
                <button type="submit" className="btn btn-primary flex-1">
                  Add
                </button>
                <button
                  type="button"
                  onClick={() => setShowIPModal(false)}
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

export default IPGroups
