import { useEffect, useState, useRef } from 'react'
import { getIPGroups, createIPGroup, updateIPGroup, deleteIPGroup, addIPToGroup, getGroupIPs, updateIPAddress, removeIPFromGroup, getVHosts } from '../services/api'
import { Plus, Trash2, Shield, Edit2, Eye, Globe } from 'lucide-react'
import ConfirmModal from '../components/ConfirmModal'

const IPGroups = () => {
  const [groups, setGroups] = useState([])
  const [vhosts, setVhosts] = useState([])
  const [loading, setLoading] = useState(true)
  const [showGroupModal, setShowGroupModal] = useState(false)
  const vhostDropdownRef = useRef(null)
  const [showIPModal, setShowIPModal] = useState(false)
  const [selectedGroup, setSelectedGroup] = useState(null)
  const [editingGroup, setEditingGroup] = useState(null)
  const [editingIP, setEditingIP] = useState(null)
  const [groupIPs, setGroupIPs] = useState([])
  const [loadingIPs, setLoadingIPs] = useState(false)
  const [confirmModal, setConfirmModal] = useState({ isOpen: false, onConfirm: null, title: '', message: '' })
  const [vhostSearchTerm, setVhostSearchTerm] = useState('')
  const [showVhostDropdown, setShowVhostDropdown] = useState(false)
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    type: 'blacklist',
    vhost_ids: [],
  })
  const [ipFormData, setIPFormData] = useState({
    ip_address: '',
    cidr_mask: null,
    description: '',
  })

  useEffect(() => {
    loadGroups()
    loadVhosts()
  }, [])

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (vhostDropdownRef.current && !vhostDropdownRef.current.contains(event.target)) {
        setShowVhostDropdown(false)
      }
    }

    if (showVhostDropdown) {
      document.addEventListener('mousedown', handleClickOutside)
      return () => {
        document.removeEventListener('mousedown', handleClickOutside)
      }
    }
  }, [showVhostDropdown])

  const loadVhosts = async () => {
    try {
      const response = await getVHosts()
      setVhosts(response.data || [])
    } catch (error) {
      console.error('Failed to load vhosts:', error)
      setVhosts([])
    }
  }

  const loadGroups = async () => {
    try {
      const response = await getIPGroups()
      setGroups(response.data || [])
    } catch (error) {
      console.error('Failed to load IP groups:', error)
      setGroups([])
    } finally {
      setLoading(false)
    }
  }

  const handleCreateGroup = async (e) => {
    e.preventDefault()
    try {
      if (editingGroup) {
        await updateIPGroup(editingGroup.id, formData)
      } else {
        await createIPGroup(formData)
      }
      setShowGroupModal(false)
      setEditingGroup(null)
      setFormData({ name: '', description: '', type: 'blacklist', vhost_ids: [] })
      loadGroups()
    } catch (error) {
      console.error('Failed to save group:', error)
    }
  }

  const handleEditGroup = (group) => {
    setEditingGroup(group)
    setFormData({
      name: group.name,
      description: group.description,
      type: group.type,
      vhost_ids: group.vhost_ids || []
    })
    setShowGroupModal(true)
  }

  const handleDeleteGroup = (id) => {
    setConfirmModal({
      isOpen: true,
      title: 'Delete IP Group',
      message: 'Are you sure you want to delete this IP group? This action cannot be undone.',
      type: 'danger',
      onConfirm: async () => {
        try {
          await deleteIPGroup(id)
          loadGroups()
        } catch (error) {
          console.error('Failed to delete group:', error)
        }
      }
    })
  }

  const handleViewIPs = async (group) => {
    setSelectedGroup(group)
    setLoadingIPs(true)
    try {
      const response = await getGroupIPs(group.id)
      setGroupIPs(response.data || [])
    } catch (error) {
      console.error('Failed to load IPs:', error)
      setGroupIPs([])
    } finally {
      setLoadingIPs(false)
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

  const handleIPEdit = (ip) => {
    setEditingIP({
      id: ip.id,
      ip_address: ip.ip_address,
      cidr_mask: ip.cidr_mask,
      description: ip.description
    })
  }

  const handleIPChange = (field, value) => {
    setEditingIP(prev => ({
      ...prev,
      [field]: value
    }))
  }

  const handleIPSave = async (ipId) => {
    try {
      await updateIPAddress(selectedGroup.id, ipId, {
        ip_address: editingIP.ip_address,
        cidr_mask: editingIP.cidr_mask || null,
        description: editingIP.description
      })
      setEditingIP(null)
      handleViewIPs(selectedGroup)
    } catch (error) {
      console.error('Failed to update IP:', error)
    }
  }

  const handleIPKeyPress = (e, ipId) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleIPSave(ipId)
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
          onClick={() => {
            setEditingGroup(null)
            setFormData({ name: '', description: '', type: 'blacklist', vhost_id: null })
            setShowGroupModal(true)
          }}
          className="btn btn-primary flex items-center gap-2"
        >
          <Plus className="w-5 h-5" />
          Add IP Group
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {(groups || []).map((group) => (
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
                <div className="flex-1">
                  <h3 className="font-semibold">{group.name}</h3>
                  <p className="text-sm text-gray-600">{group.description}</p>
                  {group.vhost_domains && group.vhost_domains.length > 0 ? (
                    <div className="text-xs text-blue-600 mt-1 flex items-start gap-1">
                      <Globe className="w-3 h-3 mt-0.5" />
                      <span>
                        {group.vhost_domains.join(', ')}
                        {group.vhost_domains.length > 2 && (
                          <span className="ml-1 text-gray-500">({group.vhost_domains.length} vhosts)</span>
                        )}
                      </span>
                    </div>
                  ) : (
                    <p className="text-xs text-gray-500 mt-1 flex items-center gap-1">
                      <Globe className="w-3 h-3" />
                      Global (All Vhosts)
                    </p>
                  )}
                </div>
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => handleEditGroup(group)}
                  className="text-blue-600 hover:text-blue-800"
                  title="Edit Group"
                >
                  <Edit2 className="w-5 h-5" />
                </button>
                <button
                  onClick={() => handleDeleteGroup(group.id)}
                  className="text-red-600 hover:text-red-800"
                  title="Delete Group"
                >
                  <Trash2 className="w-5 h-5" />
                </button>
              </div>
            </div>
            <div className="flex items-center justify-between pt-3 border-t">
              <span className={`inline-flex px-2 py-1 rounded text-sm font-medium ${
                group.type === 'blacklist'
                  ? 'bg-red-100 text-red-800'
                  : 'bg-green-100 text-green-800'
              }`}>
                {group.type.charAt(0).toUpperCase() + group.type.slice(1)}
              </span>
              <button
                onClick={() => handleViewIPs(group)}
                className="btn btn-sm btn-secondary flex items-center gap-2"
              >
                <Eye className="w-4 h-4" />
                View IPs
              </button>
            </div>
          </div>
        ))}
      </div>

      {/* Create/Edit Group Modal */}
      {showGroupModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h2 className="text-2xl font-bold mb-4">
              {editingGroup ? 'Edit IP Group' : 'Create IP Group'}
            </h2>
            <form onSubmit={handleCreateGroup} className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">Name</label>
                <input
                  type="text"
                  className="input w-full"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Description</label>
                <textarea
                  className="input w-full"
                  rows="3"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Type</label>
                <select
                  className="input w-full"
                  value={formData.type}
                  onChange={(e) => setFormData({ ...formData, type: e.target.value })}
                >
                  <option value="blacklist">Blacklist</option>
                  <option value="whitelist">Whitelist</option>
                </select>
              </div>
              <div ref={vhostDropdownRef}>
                <label className="block text-sm font-medium mb-1">Virtual Host (Optional)</label>
                
                {/* Selected items display with tags */}
                <div 
                  className="input w-full min-h-[42px] cursor-pointer flex flex-wrap gap-2 items-center"
                  onClick={() => setShowVhostDropdown(!showVhostDropdown)}
                >
                  {(formData.vhost_ids || []).length === 0 ? (
                    <span className="text-gray-400 text-sm">Select vhosts...</span>
                  ) : (
                    <>
                      {(formData.vhost_ids || []).map(vhostId => {
                        const vhost = (vhosts || []).find(v => v.id === vhostId)
                        if (!vhost) return null
                        return (
                          <span
                            key={vhostId}
                            className="inline-flex items-center gap-1 px-2 py-1 bg-blue-100 text-blue-800 rounded text-sm"
                            onClick={(e) => e.stopPropagation()}
                          >
                            <span>{vhost.domain}</span>
                            <button
                              type="button"
                              className="hover:bg-blue-200 rounded-full p-0.5 transition-colors"
                              onClick={(e) => {
                                e.stopPropagation()
                                const currentIds = formData.vhost_ids || []
                                setFormData({ ...formData, vhost_ids: currentIds.filter(id => id !== vhostId) })
                              }}
                            >
                              <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                                <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                              </svg>
                            </button>
                          </span>
                        )
                      })}
                    </>
                  )}
                </div>
                
                {/* Dropdown menu */}
                {showVhostDropdown && (
                  <div className="relative mt-1">
                    <div className="absolute z-50 w-full bg-white border border-gray-300 rounded-lg shadow-lg overflow-hidden">
                      {/* Search input inside dropdown */}
                      <div className="p-2 border-b border-gray-200 bg-gray-50">
                        <input
                          type="text"
                          className="w-full px-3 py-2 border border-gray-300 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                          placeholder="🔍 Search vhosts..."
                          value={vhostSearchTerm}
                          onChange={(e) => setVhostSearchTerm(e.target.value)}
                          onClick={(e) => e.stopPropagation()}
                        />
                      </div>
                      
                      {/* Action buttons */}
                      {(vhosts || []).length > 0 && (
                        <div className="flex gap-2 p-2 border-b border-gray-200 bg-gray-50">
                          <button
                            type="button"
                            className="text-xs px-2 py-1 bg-blue-50 text-blue-600 rounded hover:bg-blue-100 transition"
                            onClick={(e) => {
                              e.stopPropagation()
                              const filteredVhosts = (vhosts || []).filter(vhost => 
                                vhostSearchTerm === '' ||
                                vhost.domain.toLowerCase().includes(vhostSearchTerm.toLowerCase()) ||
                                vhost.name.toLowerCase().includes(vhostSearchTerm.toLowerCase())
                              )
                              const allIds = filteredVhosts.map(v => v.id)
                              setFormData({ ...formData, vhost_ids: allIds })
                            }}
                          >
                            ✓ Select All
                          </button>
                          <button
                            type="button"
                            className="text-xs px-2 py-1 bg-gray-50 text-gray-600 rounded hover:bg-gray-100 transition"
                            onClick={(e) => {
                              e.stopPropagation()
                              setFormData({ ...formData, vhost_ids: [] })
                            }}
                          >
                            ✕ Clear All
                          </button>
                        </div>
                      )}
                      
                      {/* Vhost list */}
                      <div className="max-h-80 overflow-y-auto" style={{ maxHeight: '320px' }}>
                        {(vhosts || [])
                          .filter(vhost => 
                            vhostSearchTerm === '' ||
                            vhost.domain.toLowerCase().includes(vhostSearchTerm.toLowerCase()) ||
                            vhost.name.toLowerCase().includes(vhostSearchTerm.toLowerCase())
                          )
                          .map((vhost) => {
                            const isSelected = (formData.vhost_ids || []).includes(vhost.id)
                            return (
                              <div
                                key={vhost.id}
                                className={`flex items-center px-3 py-2 hover:bg-blue-50 cursor-pointer border-b border-gray-100 last:border-b-0 transition-colors ${
                                  isSelected ? 'bg-blue-50' : ''
                                }`}
                                onClick={(e) => {
                                  e.stopPropagation()
                                  const currentIds = formData.vhost_ids || []
                                  if (isSelected) {
                                    setFormData({ ...formData, vhost_ids: currentIds.filter(id => id !== vhost.id) })
                                  } else {
                                    setFormData({ ...formData, vhost_ids: [...currentIds, vhost.id] })
                                  }
                                }}
                              >
                                <div className="flex-1">
                                  <div className="text-sm font-medium text-gray-900">{vhost.domain}</div>
                                  <div className="text-xs text-gray-500">{vhost.name}</div>
                                </div>
                                {isSelected && (
                                  <svg className="w-5 h-5 text-blue-600" fill="currentColor" viewBox="0 0 20 20">
                                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                                  </svg>
                                )}
                              </div>
                            )
                          })}
                        {(vhosts || []).filter(vhost => 
                          vhostSearchTerm === '' ||
                          vhost.domain.toLowerCase().includes(vhostSearchTerm.toLowerCase()) ||
                          vhost.name.toLowerCase().includes(vhostSearchTerm.toLowerCase())
                        ).length === 0 && (
                          <div className="px-3 py-8 text-center text-gray-500 text-sm">
                            <div className="text-2xl mb-2">🔍</div>
                            No vhosts found matching "{vhostSearchTerm}"
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                )}
                
                <p className="text-xs text-gray-500 mt-2">
                  Select one or more vhosts, or leave empty to apply to all vhosts globally.
                  {formData.vhost_ids && formData.vhost_ids.length > 0 && (
                    <span className="block mt-1 font-medium text-blue-600">
                      ✓ Selected: {formData.vhost_ids.length} vhost(s)
                    </span>
                  )}
                </p>
              </div>
              <div className="flex gap-2">
                <button type="submit" className="btn btn-primary flex-1">
                  {editingGroup ? 'Update' : 'Create'}
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setShowGroupModal(false)
                    setEditingGroup(null)
                    setFormData({ name: '', description: '', type: 'blacklist', vhost_ids: [] })
                  }}
                  className="btn btn-secondary flex-1"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* View IPs Modal */}
      {selectedGroup && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-2xl">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-2xl font-bold">
                {selectedGroup.name} - IP Addresses
              </h2>
              <button
                onClick={() => setShowIPModal(true)}
                className="btn btn-primary btn-sm flex items-center gap-2"
              >
                <Plus className="w-4 h-4" />
                Add IP
              </button>
            </div>

            <div className="space-y-2 max-h-96 overflow-y-auto">
              {loadingIPs ? (
                <div className="text-center py-4 text-gray-500">Loading IPs...</div>
              ) : (groupIPs || []).length === 0 ? (
                <div className="text-center py-4 text-gray-500">No IPs added yet</div>
              ) : (
                (groupIPs || []).map((ip) => (
                  <div key={ip.id} className="flex items-center justify-between p-3 bg-gray-50 rounded hover:bg-gray-100 transition-colors">
                    {editingIP && editingIP.id === ip.id ? (
                      <div className="flex-1 space-y-2">
                        <div className="flex gap-2">
                          <input
                            type="text"
                            className="input flex-1"
                            placeholder="IP Address"
                            value={editingIP.ip_address}
                            onChange={(e) => handleIPChange('ip_address', e.target.value)}
                            onKeyPress={(e) => handleIPKeyPress(e, ip.id)}
                            autoFocus
                          />
                          <input
                            type="number"
                            className="input w-20"
                            placeholder="CIDR"
                            min="0"
                            max="32"
                            value={editingIP.cidr_mask || ''}
                            onChange={(e) => handleIPChange('cidr_mask', e.target.value ? Number.parseInt(e.target.value, 10) : null)}
                            onKeyPress={(e) => handleIPKeyPress(e, ip.id)}
                          />
                        </div>
                        <input
                          type="text"
                          className="input w-full"
                          placeholder="Description"
                          value={editingIP.description || ''}
                          onChange={(e) => handleIPChange('description', e.target.value)}
                          onKeyPress={(e) => handleIPKeyPress(e, ip.id)}
                        />
                        <div className="flex gap-2">
                          <button
                            onClick={() => handleIPSave(ip.id)}
                            className="btn btn-primary btn-sm"
                          >
                            Save
                          </button>
                          <button
                            onClick={() => setEditingIP(null)}
                            className="btn btn-secondary btn-sm"
                          >
                            Cancel
                          </button>
                        </div>
                      </div>
                    ) : (
                      <>
                        <div className="flex-1">
                          <p className="font-mono font-semibold">
                            {ip.ip_address}{ip.cidr_mask ? `/${ip.cidr_mask}` : ''}
                          </p>
                          {ip.description && (
                            <p className="text-sm text-gray-600">{ip.description}</p>
                          )}
                        </div>
                        <div className="flex gap-2">
                          <button
                            onClick={() => handleIPEdit(ip)}
                            className="text-blue-600 hover:text-blue-800"
                            title="Edit IP"
                          >
                            <Edit2 className="w-4 h-4" />
                          </button>
                          <button
                            onClick={() => handleRemoveIP(ip.id)}
                            className="text-red-600 hover:text-red-800"
                            title="Delete IP"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </>
                    )}
                  </div>
                ))
              )}
            </div>

            <button
              onClick={() => {
                setSelectedGroup(null)
                setEditingIP(null)
              }}
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
                <label className="block text-sm font-medium mb-1">IP Address</label>
                <input
                  type="text"
                  className="input w-full"
                  placeholder="e.g., 192.168.1.100"
                  value={ipFormData.ip_address}
                  onChange={(e) => setIPFormData({ ...ipFormData, ip_address: e.target.value })}
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">CIDR Mask (optional)</label>
                <input
                  type="number"
                  className="input w-full"
                  placeholder="e.g., 24 for /24"
                  min="0"
                  max="32"
                  value={ipFormData.cidr_mask || ''}
                  onChange={(e) => setIPFormData({
                    ...ipFormData,
                    cidr_mask: e.target.value ? Number.parseInt(e.target.value, 10) : null
                  })}
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Description</label>
                <textarea
                  className="input w-full"
                  rows="2"
                  placeholder="Optional description"
                  value={ipFormData.description}
                  onChange={(e) => setIPFormData({ ...ipFormData, description: e.target.value })}
                />
              </div>
              <div className="flex gap-2">
                <button type="submit" className="btn btn-primary flex-1">
                  Add IP
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setShowIPModal(false)
                    setIPFormData({ ip_address: '', cidr_mask: null, description: '' })
                  }}
                  className="btn btn-secondary flex-1"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      <ConfirmModal
        isOpen={confirmModal.isOpen}
        onClose={() => setConfirmModal({ ...confirmModal, isOpen: false })}
        onConfirm={() => {
          confirmModal.onConfirm()
          setConfirmModal({ ...confirmModal, isOpen: false })
        }}
        title={confirmModal.title}
        message={confirmModal.message}
        type={confirmModal.type}
      />
    </div>
  )
}

export default IPGroups
