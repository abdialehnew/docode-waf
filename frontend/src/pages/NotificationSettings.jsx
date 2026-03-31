import { useState, useEffect, useCallback } from 'react'
import {
    Bell,
    Plus,
    Trash2,
    Pencil,
    CheckCircle,
    XCircle,
    Send,
    MessageCircle,
    SendHorizonal
} from 'lucide-react'
import Swal from 'sweetalert2'

// Available event types
const EVENT_TYPES = [
    { id: 'attack_detected', label: 'Attack Detected (High/Critical)' },
    { id: 'ip_banned', label: 'IP Banned' },
]

const INITIAL_FORM_STATE = {
    name: '',
    type: 'slack',
    config: { 
        webhook_url: '', 
        email_address: '', 
        bot_token: '', 
        chat_id: '',
        api_url: '',
        api_token: '',
        phone_number: ''
    },
    events: ['attack_detected', 'ip_banned'],
    enabled: true
}

const NotificationSettings = () => {
    const [channels, setChannels] = useState([])
    const [loading, setLoading] = useState(true)
    const [showModal, setShowModal] = useState(false)
    const [editingChannel, setEditingChannel] = useState(null)

    // Form State
    const [formData, setFormData] = useState(INITIAL_FORM_STATE)

    const fetchChannels = useCallback(async () => {
        try {
            const token = localStorage.getItem('token')
            const response = await fetch('/api/v1/notifications/channels', {
                headers: { Authorization: `Bearer ${token}` }
            })
            if (response.ok) {
                const data = await response.json()
                setChannels(data)
            }
        } catch (error) {
            console.error('Error fetching channels:', error)
        } finally {
            setLoading(false)
        }
    }, [])

    useEffect(() => {
        fetchChannels()
    }, [fetchChannels])

    const handleDelete = async (id) => {
        const result = await Swal.fire({
            title: 'Are you sure?',
            text: "You won't be able to revert this!",
            icon: 'warning',
            showCancelButton: true,
            confirmButtonColor: '#3085d6',
            cancelButtonColor: '#d33',
            confirmButtonText: 'Yes, delete it!'
        })

        if (result.isConfirmed) {
            try {
                const token = localStorage.getItem('token')
                await fetch(`/api/v1/notifications/channels/${id}`, {
                    method: 'DELETE',
                    headers: { Authorization: `Bearer ${token}` }
                })
                fetchChannels()
                Swal.fire('Deleted!', 'Your channel has been deleted.', 'success')
            } catch (error) {
                console.error('Error deleting channel:', error)
                Swal.fire('Error!', 'Failed to delete channel.', 'error')
            }
        }
    }

    const handleTest = async (channel) => {
        try {
            const token = localStorage.getItem('token')

            // Handle config (it might be an object already or JSON string)
            let configData = channel.config
            if (typeof configData === 'string') {
                try {
                    configData = JSON.parse(configData)
                } catch (e) {
                    console.error("Failed to parse config string", e)
                }
            }

            const response = await fetch('/api/v1/notifications/test', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: `Bearer ${token}`
                },
                body: JSON.stringify({
                    type: channel.type,
                    config: configData
                })
            })

            if (response.ok) {
                Swal.fire({
                    title: 'Success!',
                    text: 'Test notification sent!',
                    icon: 'success',
                    timer: 2000,
                    showConfirmButton: false
                })
            } else {
                const err = await response.json()
                Swal.fire('Error!', 'Failed to send test: ' + (err.error || 'Unknown error'), 'error')
            }
        } catch (error) {
            Swal.fire('Error!', 'Error testing channel: ' + error.message, 'error')
        }
    }

    const handleSubmit = async (e) => {
        e.preventDefault()

        try {
            const token = localStorage.getItem('token')
            const method = editingChannel ? 'PUT' : 'POST'
            const url = editingChannel
                ? `/api/v1/notifications/channels/${editingChannel.id}`
                : '/api/v1/notifications/channels'

            // Prepare payload
            const payload = {
                name: formData.name,
                type: formData.type,
                config: formData.config,
                events: formData.events,
                enabled: formData.enabled
            }

            const response = await fetch(url, {
                method: method,
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: `Bearer ${token}`
                },
                body: JSON.stringify(payload)
            })

            if (response.ok) {
                setShowModal(false)
                fetchChannels()
                resetForm()
                Swal.fire({
                    title: 'Saved!',
                    text: 'Channel saved successfully.',
                    icon: 'success',
                    timer: 1500,
                    showConfirmButton: false
                })
            } else {
                const err = await response.json()
                Swal.fire('Error!', 'Failed to save channel: ' + (err.error || 'Unknown error'), 'error')
            }
        } catch (error) {
            console.error('Error saving channel:', error)
            Swal.fire('Error!', 'An unexpected error occurred.', 'error')
        }
    }

    const resetForm = () => {
        setEditingChannel(null)
        setFormData(INITIAL_FORM_STATE)
    }

    const openEditModal = (channel) => {
        // Parse config/events if they come as strings, though they should be objects from JSON response
        // Usually standard fetch response.json() parses them if they are sent as JSON
        // But our backend sends them as json.RawMessage which becomes actual JSON object in response body.

        // Check if we need to parse.
        let config = channel.config
        let events = channel.events
        // If for some reason they are strings (double encoded), parse them. 
        // Assuming they are objects.

        setEditingChannel(channel)
        setFormData({
            name: channel.name,
            type: channel.type,
            config: config || {},
            events: events || [],
            enabled: channel.enabled
        })
        setShowModal(true)
    }

    const toggleEvent = (eventId) => {
        setFormData(prev => {
            const exists = prev.events.includes(eventId)
            if (exists) {
                return { ...prev, events: prev.events.filter(e => e !== eventId) }
            } else {
                return { ...prev, events: [...prev.events, eventId] }
            }
        })
    }

    // --- HELPER FUNCTIONS FOR RENDERING ---
    const getChannelStyles = (type) => {
        const styles = {
            slack: 'bg-purple-100 text-purple-600',
            discord: 'bg-indigo-100 text-indigo-600',
            email: 'bg-blue-100 text-blue-600',
            telegram: 'bg-sky-100 text-sky-600',
            whatsapp: 'bg-green-100 text-green-600',
        }
        return styles[type] || 'bg-gray-100 text-gray-600'
    }

    const getChannelIcon = (type) => {
        switch (type) {
            case 'telegram': return <SendHorizonal className="h-6 w-6" />
            case 'whatsapp': return <MessageCircle className="h-6 w-6" />
            default: return <Bell className="h-6 w-6" />
        }
    }

    const getChannelDestination = (channel) => {
        if (!channel.config) return 'Not configured'
        
        switch (channel.type) {
            case 'email': return channel.config.email_address || 'No email'
            case 'telegram': return `Chat: ${channel.config.chat_id || 'Not set'}`
            case 'whatsapp': return `To: ${channel.config.phone_number || 'Not set'}`
            default: return channel.config.webhook_url || 'No URL'
        }
    }

    const renderConfigFields = () => {
        switch (formData.type) {
            case 'email':
                return (
                    <div>
                        <label htmlFor="emailAddress" className="block text-sm font-medium text-gray-700">Email Address</label>
                        <input
                            id="emailAddress"
                            type="email"
                            value={formData.config.email_address || ''}
                            onChange={e => setFormData({
                                ...formData,
                                config: { ...formData.config, email_address: e.target.value }
                            })}
                            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border"
                            required
                        />
                    </div>
                )
            case 'telegram':
                return (
                    <div className="space-y-3">
                        <div>
                            <label htmlFor="botToken" className="block text-sm font-medium text-gray-700">Bot Token</label>
                            <input
                                id="botToken"
                                type="password"
                                value={formData.config.bot_token || ''}
                                onChange={e => setFormData({
                                    ...formData,
                                    config: { ...formData.config, bot_token: e.target.value }
                                })}
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border"
                                placeholder="123456789:ABCDefgh..."
                                required
                            />
                        </div>
                        <div>
                            <label htmlFor="chatId" className="block text-sm font-medium text-gray-700">Chat ID</label>
                            <input
                                id="chatId"
                                type="text"
                                value={formData.config.chat_id || ''}
                                onChange={e => setFormData({
                                    ...formData,
                                    config: { ...formData.config, chat_id: e.target.value }
                                })}
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border"
                                placeholder="-100123456789"
                                required
                            />
                            <p className="mt-1 text-xs text-info flex items-center gap-1">
                                <Bell className="h-3 w-3" />
                                Tip: Send <code>/start</code> to your bot first to get your Chat ID and enable notifications.
                            </p>
                        </div>
                    </div>
                )
            case 'whatsapp':
                return (
                    <div className="space-y-3">
                        <div>
                            <label htmlFor="apiUrl" className="block text-sm font-medium text-gray-700">API Gateway URL</label>
                            <input
                                id="apiUrl"
                                type="url"
                                value={formData.config.api_url || ''}
                                onChange={e => setFormData({
                                    ...formData,
                                    config: { ...formData.config, api_url: e.target.value }
                                })}
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border"
                                placeholder="https://api.ultramsg.com/instanceXXX/messages/chat"
                                required
                            />
                        </div>
                        <div>
                            <label htmlFor="apiToken" className="block text-sm font-medium text-gray-700">API Token / Instance Key</label>
                            <input
                                id="apiToken"
                                type="password"
                                value={formData.config.api_token || ''}
                                onChange={e => setFormData({
                                    ...formData,
                                    config: { ...formData.config, api_token: e.target.value }
                                })}
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border"
                                required
                            />
                        </div>
                        <div>
                            <label htmlFor="phoneNumber" className="block text-sm font-medium text-gray-700">Target Phone Number</label>
                            <input
                                id="phoneNumber"
                                type="text"
                                value={formData.config.phone_number || ''}
                                onChange={e => setFormData({
                                    ...formData,
                                    config: { ...formData.config, phone_number: e.target.value }
                                })}
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border"
                                placeholder="62812345678"
                                required
                            />
                        </div>
                    </div>
                )
            default:
                return (
                    <div>
                        <label htmlFor="webhookUrl" className="block text-sm font-medium text-gray-700">Webhook URL</label>
                        <input
                            id="webhookUrl"
                            type="url"
                            value={formData.config.webhook_url || ''}
                            onChange={e => setFormData({
                                ...formData,
                                config: { ...formData.config, webhook_url: e.target.value }
                            })}
                            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border"
                            placeholder={formData.type === 'slack' ? 'https://hooks.slack.com/services/...' : 'https://discord.com/api/webhooks/...'}
                            required
                        />
                    </div>
                )
        }
    }

    const renderChannelContent = () => {
        if (loading) {
            return <div className="text-center py-12 text-gray-500">Loading channels...</div>
        }

        if (channels.length === 0) {
            return (
                <div className="text-center py-12 bg-white rounded-lg shadow border border-gray-200">
                    <Bell className="h-12 w-12 text-gray-300 mx-auto mb-3" />
                    <h3 className="text-lg font-medium text-gray-900">No channels configured</h3>
                    <p className="text-gray-500 mt-1 mb-4">Add a notification channel to receive security alerts.</p>
                    <button
                        onClick={() => { resetForm(); setShowModal(true) }}
                        className="text-blue-600 hover:text-blue-800 font-medium"
                    >
                        Create your first channel
                    </button>
                </div>
            )
        }

        return (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {channels.map(channel => (
                    <div key={channel.id} className="bg-white rounded-lg border border-gray-200 shadow-sm hover:shadow-md transition p-5">
                        <div className="flex justify-between items-start mb-4">
                            <div className="flex items-center gap-3">
                                <div className={`p-2 rounded-lg ${getChannelStyles(channel.type)}`}>
                                    {getChannelIcon(channel.type)}
                                </div>
                                <div>
                                    <h3 className="font-semibold text-gray-900">{channel.name}</h3>
                                    <p className="text-xs text-gray-500 uppercase font-medium">{channel.type}</p>
                                </div>
                            </div>
                            <div className={`flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${channel.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'
                                }`}>
                                {channel.enabled ? <CheckCircle className="h-3 w-3" /> : <XCircle className="h-3 w-3" />}
                                {channel.enabled ? 'Active' : 'Disabled'}
                            </div>
                        </div>

                        <div className="space-y-3 mb-4">
                            <div className="text-sm">
                                <span className="text-gray-500 block text-xs">Destination:</span>
                                <span className="font-mono text-xs bg-gray-50 px-2 py-1 rounded block truncate mt-1">
                                    {getChannelDestination(channel)}
                                </span>
                            </div>
                            <div>
                                <span className="text-gray-500 block text-xs mb-1">Subscribed Events:</span>
                                <div className="flex flex-wrap gap-1">
                                    {channel.events.map(ev => (
                                        <span key={ev} className="text-xs bg-blue-50 text-blue-700 px-2 py-0.5 rounded border border-blue-100">
                                            {ev.replace('_', ' ')}
                                        </span>
                                    ))}
                                </div>
                            </div>
                        </div>

                        <div className="flex justify-end gap-2 pt-3 border-t border-gray-100">
                            <button
                                onClick={() => handleTest(channel)}
                                className="p-1.5 text-gray-500 hover:text-blue-600 transition"
                                title="Send Test Notification"
                            >
                                <Send className="h-4 w-4" />
                            </button>
                            <button
                                onClick={() => openEditModal(channel)}
                                className="p-1.5 text-gray-500 hover:text-blue-600 transition"
                                title="Edit"
                            >
                                <Pencil className="h-4 w-4" />
                            </button>
                            <button
                                onClick={() => handleDelete(channel.id)}
                                className="p-1.5 text-gray-500 hover:text-red-600 transition"
                                title="Delete"
                            >
                                <Trash2 className="h-4 w-4" />
                            </button>
                        </div>
                    </div>
                ))}
            </div>
        )
    }

    return (
        <div className="p-6">
            <div className="flex justify-between items-center mb-6">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
                        <Bell className="h-8 w-8 text-blue-600" />
                        Notification Channels
                    </h1>
                    <p className="text-gray-500 mt-1">Configure alerts for security events</p>
                </div>
                <button
                    onClick={() => { resetForm(); setShowModal(true) }}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                >
                    <Plus className="h-5 w-5" />
                    Add Channel
                </button>
            </div>

            {renderChannelContent()}

            {/* Modal */}
            {showModal && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
                    <div className="bg-white rounded-xl shadow-xl max-w-lg w-full p-6">
                        <h2 className="text-xl font-bold mb-4">
                            {editingChannel ? 'Edit Channel' : 'Add Notification Channel'}
                        </h2>

                        <form onSubmit={handleSubmit} className="space-y-4">
                            <div>
                                <label htmlFor="channelName" className="block text-sm font-medium text-gray-700">Channel Name</label>
                                <input
                                    id="channelName"
                                    type="text"
                                    value={formData.name}
                                    onChange={e => setFormData({ ...formData, name: e.target.value })}
                                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border"
                                    required
                                />
                            </div>

                            <div>
                                <label htmlFor="channelType" className="block text-sm font-medium text-gray-700">Channel Type</label>
                                <select
                                    id="channelType"
                                    value={formData.type}
                                    onChange={e => setFormData({ ...formData, type: e.target.value })}
                                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border"
                                >
                                    <option value="slack">Slack</option>
                                    <option value="discord">Discord</option>
                                    <option value="telegram">Telegram Bot</option>
                                    <option value="whatsapp">WhatsApp API</option>
                                    <option value="email">Email</option>
                                    <option value="webhook">Generic Webhook</option>
                                </select>
                            </div>

                            {/* Dynamic Config Fields */}
                            {renderConfigFields()}

                            {/* Subscribed Events */}
                            <div>
                                <span className="block text-sm font-medium text-gray-700 mb-2">Trigger Events</span>
                                <div className="space-y-2">
                                    {EVENT_TYPES.map(event => (
                                        <label key={event.id} htmlFor={`event-${event.id}`} className="flex items-center gap-2">
                                            <input
                                                id={`event-${event.id}`}
                                                type="checkbox"
                                                checked={formData.events.includes(event.id)}
                                                onChange={() => toggleEvent(event.id)}
                                                className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                                            />
                                            <span className="text-sm text-gray-700">{event.label}</span>
                                        </label>
                                    ))}
                                </div>
                            </div>

                            <div className="flex items-center gap-2">
                                <input
                                    type="checkbox"
                                    id="enabled"
                                    checked={formData.enabled}
                                    onChange={e => setFormData({ ...formData, enabled: e.target.checked })}
                                    className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                                />
                                <label htmlFor="enabled" className="text-sm font-medium text-gray-700">Enable this channel</label>
                            </div>

                            <div className="flex justify-end gap-3 pt-4">
                                <button
                                    type="button"
                                    onClick={() => setShowModal(false)}
                                    className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    className="px-4 py-2 bg-blue-600 border border-transparent rounded-md text-sm font-medium text-white hover:bg-blue-700"
                                >
                                    {editingChannel ? 'Save Changes' : 'Create Channel'}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    )
}

export default NotificationSettings
