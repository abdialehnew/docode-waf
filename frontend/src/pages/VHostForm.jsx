import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import toast from 'react-hot-toast'
import api, { createVHost, getVHost } from '../services/api'
import {
    ArrowLeft, Plus, Trash2, Search, ChevronUp, ChevronDown,
    ChevronDown as ChevronDownIcon, CheckCircle, AlertCircle, Loader2,
    Save, FileCode, Settings2
} from 'lucide-react'
import CodeMirror from '@uiw/react-codemirror'
import { monokai } from '../theme/monokai'
import { nginx } from '../lang/nginx'
import logger from '../utils/logger'
import { parseNginxConfig } from '../utils/nginxConfigParser'

const initialFormData = {
    name: '',
    type: 'proxy',
    domain: [],
    backend_url: '',
    backends: [],
    load_balance_method: 'round_robin',
    custom_config: '',
    ssl_enabled: false,
    ssl_certificate_id: '',
    enabled: true,
    websocket_enabled: false,
    http_version: 'http/1.1',
    tls_version: 'TLSv1.2',
    max_upload_size: 10,
    proxy_read_timeout: 60,
    proxy_connect_timeout: 60,
    bot_detection_enabled: false,
    bot_detection_type: 'turnstile',
    recaptcha_version: 'v2',
    rate_limit_enabled: false,
    rate_limit_requests: 100,
    rate_limit_window: 60,
    region_filtering_enabled: false,
    region_whitelist: [],
    region_blacklist: [],
    defense_mode: 'defense',
    custom_locations: [],
    custom_headers: {},
    ip_group_ids: [],
    hsts_enabled: false,
    hsts_max_age: 31536000,
    hsts_include_subdomains: false,
    hsts_preload: false,
    brotli_enabled: true,
    http3_enabled: false,
    hide_server_tokens: true,
    security_headers_enabled: true,
    client_body_buffer_size: 128,
}

const VHostForm = () => {
    const navigate = useNavigate()
    const { id } = useParams()
    const isEditMode = Boolean(id)

    const [activeTab, setActiveTab] = useState('form') // 'form' or 'editor'
    const [formData, setFormData] = useState({ ...initialFormData })
    const [loading, setLoading] = useState(isEditMode)
    const [saving, setSaving] = useState(false)

    // For config editor tab
    const [configContent, setConfigContent] = useState('')
    const [originalConfig, setOriginalConfig] = useState('')
    const [configLoading, setConfigLoading] = useState(false)
    const [configDomain, setConfigDomain] = useState('')

    // Form helpers
    const [certificates, setCertificates] = useState([])
    const [certSearchTerm, setCertSearchTerm] = useState('')
    const [showCertDropdown, setShowCertDropdown] = useState(false)
    const [backendCheckStatus, setBackendCheckStatus] = useState(null)
    const [backendCheckMessage, setBackendCheckMessage] = useState('')
    const [showAdvancedSettings, setShowAdvancedSettings] = useState(false)
    const [newLocation, setNewLocation] = useState({ path: '', proxy_pass: '', config: '', websocket_enabled: false, backends: [], load_balance_method: 'round_robin' })
    const [newHeader, setNewHeader] = useState({ key: '', value: '' })
    const [locationBackendCheck, setLocationBackendCheck] = useState({ status: null, message: '' })
    const [ipGroups, setIpGroups] = useState([])

    // Preview


    useEffect(() => {
        loadCertificates()
        loadIpGroups()
        if (isEditMode) {
            loadVHost()
        }
    }, [id])

    // Debounce backend URL check
    useEffect(() => {
        const timeoutId = setTimeout(() => {
            checkBackendURL(formData.backend_url)
        }, 800)
        return () => clearTimeout(timeoutId)
    }, [formData.backend_url])

    // Check custom location backend URL
    useEffect(() => {
        if (!newLocation.proxy_pass || newLocation.proxy_pass.trim() === '') {
            setLocationBackendCheck({ status: null, message: '' })
            return
        }
        try { new URL(newLocation.proxy_pass) } catch { setLocationBackendCheck({ status: 'error', message: 'Invalid URL format' }); return }
        setLocationBackendCheck({ status: 'checking', message: 'Checking...' })
        const timeoutId = setTimeout(async () => {
            try {
                await fetch(newLocation.proxy_pass, { method: 'HEAD', mode: 'no-cors', cache: 'no-cache' })
                setLocationBackendCheck({ status: 'success', message: 'Backend is reachable' })
            } catch {
                setLocationBackendCheck({ status: 'warning', message: 'Cannot verify (CORS/Network), but may still work' })
            }
        }, 800)
        return () => clearTimeout(timeoutId)
    }, [newLocation.proxy_pass])

    useEffect(() => {
        const handleClickOutside = (event) => {
            if (showCertDropdown && !event.target.closest('.cert-dropdown-container')) {
                setShowCertDropdown(false)
            }
        }
        document.addEventListener('mousedown', handleClickOutside)
        return () => document.removeEventListener('mousedown', handleClickOutside)
    }, [showCertDropdown])

    const loadVHost = async () => {
        try {
            setLoading(true)
            const response = await getVHost(id)
            const vhost = response.data
            setFormData({
                name: vhost.name || '',
                type: vhost.type || 'proxy',
                domain: vhost.domain ? vhost.domain.split(' ') : [],
                backend_url: vhost.backend_url || '',
                backends: vhost.backends || [],
                load_balance_method: vhost.load_balance_method || 'round_robin',
                custom_config: vhost.custom_config || '',
                ssl_enabled: vhost.ssl_enabled || false,
                ssl_certificate_id: vhost.ssl_certificate_id || '',
                enabled: vhost.enabled === undefined ? true : vhost.enabled,
                websocket_enabled: vhost.websocket_enabled || false,
                http_version: vhost.http_version || 'http/1.1',
                tls_version: vhost.tls_version || 'TLSv1.2',
                max_upload_size: vhost.max_upload_size || 10,
                proxy_read_timeout: vhost.proxy_read_timeout || 60,
                proxy_connect_timeout: vhost.proxy_connect_timeout || 60,
                bot_detection_enabled: vhost.bot_detection_enabled || false,
                bot_detection_type: vhost.bot_detection_type || 'turnstile',
                recaptcha_version: vhost.recaptcha_version || 'v2',
                rate_limit_enabled: vhost.rate_limit_enabled || false,
                rate_limit_requests: vhost.rate_limit_requests || 100,
                rate_limit_window: vhost.rate_limit_window || 60,
                region_filtering_enabled: vhost.region_filtering_enabled || false,
                region_whitelist: vhost.region_whitelist || [],
                region_blacklist: vhost.region_blacklist || [],
                defense_mode: vhost.defense_mode || 'defense',
                custom_locations: vhost.custom_locations || [],
                custom_headers: vhost.custom_headers || {},
                ip_group_ids: vhost.ip_group_ids || [],
                hsts_enabled: vhost.hsts_enabled || false,
                hsts_max_age: vhost.hsts_max_age || 31536000,
                hsts_include_subdomains: vhost.hsts_include_subdomains || false,
                hsts_preload: vhost.hsts_preload || false,
                brotli_enabled: vhost.brotli_enabled === undefined ? true : vhost.brotli_enabled,
                http3_enabled: vhost.http3_enabled || false,
                hide_server_tokens: vhost.hide_server_tokens === undefined ? true : vhost.hide_server_tokens,
                security_headers_enabled: vhost.security_headers_enabled === undefined ? true : vhost.security_headers_enabled,
                client_body_buffer_size: vhost.client_body_buffer_size || 128,
            })
            // Save domain for config editor
            if (vhost.domain) {
                setConfigDomain(vhost.domain.split(' ')[0])
            }
            setLoading(false)
        } catch (error) {
            logger.error('Failed to load vhost:', error)
            toast.error('Failed to load virtual host')
            setLoading(false)
        }
    }

    const loadIpGroups = async () => {
        try {
            const response = await api.get('/ip-groups')
            setIpGroups(response.data || [])
        } catch (error) {
            logger.error('Failed to load IP groups:', error)
        }
    }

    const loadCertificates = async () => {
        try {
            const response = await api.get('/certificates')
            const data = response.data?.certificates || response.data || []
            setCertificates(Array.isArray(data) ? data : [])
        } catch (error) {
            logger.error('Failed to load certificates:', error)
            setCertificates([])
        }
    }

    const loadConfig = async () => {
        if (!configDomain) return
        try {
            setConfigLoading(true)
            const token = localStorage.getItem('token')
            const response = await fetch(`/api/v1/vhost-config/${configDomain}`, {
                headers: { Authorization: `Bearer ${token}` },
            })
            if (!response.ok) throw new Error('Failed to fetch config')
            const data = await response.json()
            setConfigContent(data.content)
            setOriginalConfig(data.content)
        } catch (err) {
            logger.error('Failed to load config:', err)
            setConfigContent('')
            setOriginalConfig('')
        } finally {
            setConfigLoading(false)
        }
    }

    const checkBackendURL = async (url) => {
        if (!url || url.trim() === '') { setBackendCheckStatus(null); setBackendCheckMessage(''); return }
        try { new URL(url) } catch { setBackendCheckStatus('error'); setBackendCheckMessage('Invalid URL format'); return }
        setBackendCheckStatus('checking')
        setBackendCheckMessage('Checking backend availability...')
        try {
            const controller = new AbortController()
            const timeoutId = setTimeout(() => controller.abort(), 5000)
            await fetch(url, { method: 'HEAD', mode: 'no-cors', signal: controller.signal })
            clearTimeout(timeoutId)
            setBackendCheckStatus('success')
            setBackendCheckMessage('Backend is reachable')
        } catch (error) {
            if (error.name === 'AbortError') { setBackendCheckStatus('error'); setBackendCheckMessage('Backend request timeout (5s)') }
            else { setBackendCheckStatus('success'); setBackendCheckMessage('Backend is reachable') }
        }
    }

    const handleSubmit = async (e) => {
        e.preventDefault()
        try {
            setSaving(true)
            const domainString = Array.isArray(formData.domain) ? formData.domain.join(' ') : formData.domain
            // Ensure array fields are always arrays, not objects
            const sanitizedData = {
                ...formData,
                domain: domainString,
                custom_config: configContent || formData.custom_config,
                region_whitelist: Array.isArray(formData.region_whitelist) ? formData.region_whitelist : [],
                region_blacklist: Array.isArray(formData.region_blacklist) ? formData.region_blacklist : [],
                backends: Array.isArray(formData.backends) ? formData.backends : [],
                custom_locations: Array.isArray(formData.custom_locations) ? formData.custom_locations : [],
            }
            if (isEditMode) {
                await api.put(`/vhosts/${id}`, sanitizedData)
                toast.success('Virtual host updated successfully')
            } else {
                await createVHost(sanitizedData)
                toast.success('Virtual host created successfully')
            }
            navigate('/vhosts')
        } catch (error) {
            logger.error('Failed to save vhost:', error)
            const errorMessage = error.response?.data?.error || 'Failed to save virtual host'
            toast.error(errorMessage)
        } finally {
            setSaving(false)
        }
    }

    const handleSaveConfig = async () => {
        if (isEditMode && configDomain) {
            // Edit mode: save directly to the nginx config file
            try {
                setSaving(true)
                const token = localStorage.getItem('token')
                const response = await fetch(`/api/v1/vhost-config/${configDomain}`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
                    body: JSON.stringify({ content: configContent }),
                })
                if (!response.ok) throw new Error('Failed to save config')
                setOriginalConfig(configContent)
                toast.success('Config saved successfully')
            } catch (err) {
                logger.error('Failed to save config:', err)
                toast.error('Failed to save config')
            } finally {
                setSaving(false)
            }
        } else {
            // New VHost: store config in formData, will be sent on form submit
            setFormData(prev => ({ ...prev, custom_config: configContent }))
            setOriginalConfig(configContent)
            toast.success('Config saved to form. It will be applied when you save the VHost.')
        }
    }



    // Bidirectional sync on tab change
    const handleTabChange = async (tab) => {
        setActiveTab(tab)

        if (tab === 'editor') {
            // Form → Editor: generate config from form data via preview API
            if (isEditMode && configDomain && !configContent) {
                // Edit mode: load existing config file from server
                loadConfig()
            } else {
                // Generate preview config from form fields
                try {
                    setConfigLoading(true)
                    const domainString = Array.isArray(formData.domain) ? formData.domain.join(' ') : formData.domain
                    const configData = { ...formData, domain: domainString }
                    const response = await api.post('/vhosts/preview', configData)
                    if (response.data?.config) {
                        setConfigContent(response.data.config)
                        setOriginalConfig(response.data.config)
                        toast.success('Config generated from form fields')
                    }
                } catch (error) {
                    logger.error('Failed to generate config preview:', error)
                    // Fallback: keep current content or use custom_config from formData
                    if (!configContent && formData.custom_config) {
                        setConfigContent(formData.custom_config)
                        setOriginalConfig(formData.custom_config)
                    }
                } finally {
                    setConfigLoading(false)
                }
            }
        }

        if (tab === 'form' && configContent) {
            // Editor → Form: parse config and fill form fields
            const parsed = parseNginxConfig(configContent)
            if (Object.keys(parsed).length > 0) {
                setFormData(prev => {
                    const updated = { ...prev }
                    // Only overwrite fields that were actually parsed
                    for (const [key, value] of Object.entries(parsed)) {
                        if (value !== undefined && value !== null) {
                            updated[key] = value
                        }
                    }
                    return updated
                })
                const fieldCount = Object.keys(parsed).length
                toast.success(`${fieldCount} form field(s) updated from config`)
            }
        }
    }

    // Helpers
    const filteredCertificates = certificates.filter(cert =>
        cert.name?.toLowerCase().includes(certSearchTerm.toLowerCase()) ||
        cert.common_name?.toLowerCase().includes(certSearchTerm.toLowerCase())
    )
    const selectedCertificate = certificates.find(cert => cert.id === formData.ssl_certificate_id)
    const getCertExpiryColor = (validToDate) => {
        const validTo = new Date(validToDate)
        const now = new Date()
        const thirtyDays = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000)
        if (validTo < now) return 'text-red-600'
        if (validTo < thirtyDays) return 'text-yellow-600'
        return 'text-green-600'
    }
    const getBackendCheckColor = (status) => {
        if (status === 'success') return 'text-green-600'
        if (status === 'error') return 'text-red-600'
        return 'text-blue-600'
    }
    const getLocationCheckColor = (status) => {
        if (status === 'success') return 'text-green-600'
        if (status === 'error') return 'text-red-600'
        if (status === 'warning') return 'text-yellow-600'
        return 'text-blue-600'
    }

    const hasConfigChanges = configContent !== originalConfig

    if (loading) {
        return (
            <div className="container mx-auto px-4 py-8">
                <div className="flex justify-center items-center h-96">
                    <Loader2 className="w-8 h-8 animate-spin text-primary-600" />
                </div>
            </div>
        )
    }

    return (
        <div className="container mx-auto px-4 py-8">
            {/* Header */}
            <div className="flex items-center gap-3 mb-6">
                <button onClick={() => navigate('/vhosts')} className="btn btn-secondary p-2">
                    <ArrowLeft className="w-5 h-5" />
                </button>
                <div>
                    <h1 className="text-3xl font-bold text-gray-900">
                        {isEditMode ? 'Edit Virtual Host' : 'Add Virtual Host'}
                    </h1>
                    {isEditMode && formData.name && (
                        <p className="text-sm text-gray-500 mt-1">{formData.name}</p>
                    )}
                </div>
            </div>

            {/* Tab Switcher */}
            <div className="flex border-b border-gray-200 mb-6">
                <button
                    onClick={() => handleTabChange('form')}
                    className={`flex items-center gap-2 px-6 py-3 text-sm font-medium border-b-2 transition-colors ${activeTab === 'form'
                        ? 'border-primary-600 text-primary-600'
                        : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                        }`}
                >
                    <Settings2 className="w-4 h-4" />
                    Form
                </button>
                <button
                    onClick={() => handleTabChange('editor')}
                    className={`flex items-center gap-2 px-6 py-3 text-sm font-medium border-b-2 transition-colors ${activeTab === 'editor'
                        ? 'border-primary-600 text-primary-600'
                        : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                        }`}
                >
                    <FileCode className="w-4 h-4" />
                    Config Editor
                </button>
            </div>

            {/* Form Tab */}
            {activeTab === 'form' && (
                <div className="card">
                    <form onSubmit={handleSubmit} className="space-y-4">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div>
                                <label htmlFor="vhost-name" className="label">Name</label>
                                <input id="vhost-name" type="text" className="input" value={formData.name}
                                    onChange={(e) => setFormData({ ...formData, name: e.target.value })} required />
                            </div>
                            <div>
                                <label htmlFor="vhost-type" className="label">Host Type</label>
                                <select id="vhost-type" className="input" value={formData.type}
                                    onChange={(e) => setFormData({ ...formData, type: e.target.value })}>
                                    <option value="proxy">Proxy Host</option>
                                    <option value="redirect">Redirect Host</option>
                                    <option value="dead">Dead Host (404)</option>
                                </select>
                            </div>
                        </div>

                        {/* Domain Names */}
                        <div>
                            <label htmlFor="domain-input" className="label">Domain Names</label>
                            <div className="flex flex-wrap items-center gap-2 p-2 border border-gray-300 rounded-lg focus-within:ring-2 focus-within:ring-blue-500 focus-within:border-transparent bg-white">
                                {formData.domain.map((domain, index) => (
                                    <span key={index} className="bg-blue-100 text-blue-800 text-sm px-2 py-1 rounded-md flex items-center gap-1">
                                        {domain}
                                        <button type="button" onClick={() => {
                                            const n = [...formData.domain]; n.splice(index, 1); setFormData({ ...formData, domain: n })
                                        }} className="hover:text-red-500">&times;</button>
                                    </span>
                                ))}
                                <input id="domain-input" type="text" className="flex-1 min-w-[120px] outline-none text-sm py-1"
                                    placeholder={formData.domain.length === 0 ? "example.com (Press Enter)" : ""}
                                    onKeyDown={(e) => {
                                        if (e.key === 'Enter' || e.key === ' ' || e.key === ',') {
                                            e.preventDefault()
                                            const val = e.target.value.trim()
                                            if (val && !formData.domain.includes(val)) {
                                                setFormData({ ...formData, domain: [...formData.domain, val] })
                                                e.target.value = ''
                                            }
                                        } else if (e.key === 'Backspace' && e.target.value === '' && formData.domain.length > 0) {
                                            const n = [...formData.domain]; n.pop(); setFormData({ ...formData, domain: n })
                                        }
                                    }}
                                    onBlur={(e) => {
                                        const val = e.target.value.trim()
                                        if (val && !formData.domain.includes(val)) {
                                            setFormData({ ...formData, domain: [...formData.domain, val] })
                                            e.target.value = ''
                                        }
                                    }}
                                />
                            </div>
                            <p className="text-xs text-gray-500 mt-1">Press Enter, Space, or Comma to add multiple domains</p>
                        </div>

                        {/* Backend Servers (unified) */}
                        {formData.type !== 'dead' && (
                            <div className="border border-gray-200 rounded-lg p-4 bg-gray-50">
                                <div className="flex items-center justify-between mb-3">
                                    <label className="label text-sm font-medium mb-0">
                                        {formData.type === 'redirect' ? 'Forward URL (Redirect Target)' : 'Backend Servers'}
                                    </label>
                                    {formData.type === 'proxy' && (
                                        <span className="text-xs text-gray-500">
                                            {(formData.backends?.length || 0) + 1} server(s)
                                        </span>
                                    )}
                                </div>

                                {/* Primary Backend URL */}
                                <div className="relative">
                                    <input id="vhost-backend" type="text" className="input pr-10"
                                        placeholder={formData.type === 'redirect' ? "https://google.com" : "http://localhost:8000"}
                                        value={formData.backend_url}
                                        onChange={(e) => setFormData({ ...formData, backend_url: e.target.value })} required />
                                    {backendCheckStatus && (
                                        <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
                                            {backendCheckStatus === 'checking' && <Loader2 className="w-5 h-5 text-blue-500 animate-spin" />}
                                            {backendCheckStatus === 'success' && <CheckCircle className="w-5 h-5 text-green-500" />}
                                            {backendCheckStatus === 'error' && <AlertCircle className="w-5 h-5 text-red-500" />}
                                        </div>
                                    )}
                                </div>
                                {backendCheckMessage && (
                                    <p className={`text-xs mt-1 flex items-center gap-1 ${getBackendCheckColor(backendCheckStatus)}`}>
                                        {backendCheckMessage}
                                    </p>
                                )}

                                {/* Additional Backends */}
                                {formData.type === 'proxy' && (
                                    <>
                                        {formData.backends && formData.backends.length > 0 && (
                                            <div className="space-y-2 mt-3">
                                                {formData.backends.map((backend, index) => (
                                                    <div key={index} className="flex items-center gap-2">
                                                        <input type="text" className="input flex-1 text-sm" value={backend}
                                                            onChange={(e) => { const n = [...formData.backends]; n[index] = e.target.value; setFormData({ ...formData, backends: n }) }}
                                                            placeholder="http://backend:8080" />
                                                        <button type="button" onClick={() => { const n = formData.backends.filter((_, i) => i !== index); setFormData({ ...formData, backends: n }) }}
                                                            className="p-2 text-red-500 hover:bg-red-50 rounded"><Trash2 className="w-4 h-4" /></button>
                                                    </div>
                                                ))}
                                            </div>
                                        )}
                                        <button type="button" onClick={() => setFormData({ ...formData, backends: [...(formData.backends || []), ''] })}
                                            className="btn btn-secondary text-xs flex items-center gap-1 mt-3"><Plus className="w-4 h-4" /> Add Backend Server</button>

                                        {/* Load Balancing - only when >1 total backends */}
                                        {formData.backends && formData.backends.length > 0 && (
                                            <div className="mt-3 pt-3 border-t border-gray-200">
                                                <label className="label text-sm">Load Balancing Method</label>
                                                <select className="input text-sm" value={formData.load_balance_method}
                                                    onChange={(e) => setFormData({ ...formData, load_balance_method: e.target.value })}>
                                                    <option value="round_robin">Round Robin (default)</option>
                                                    <option value="least_conn">Least Connections</option>
                                                    <option value="ip_hash">IP Hash (sticky sessions)</option>
                                                </select>
                                            </div>
                                        )}
                                    </>
                                )}
                            </div>
                        )}

                        {/* SSL */}
                        <div className="flex items-center gap-2">
                            <input type="checkbox" id="ssl" checked={formData.ssl_enabled}
                                onChange={(e) => { setFormData({ ...formData, ssl_enabled: e.target.checked }); if (!e.target.checked) { setFormData(prev => ({ ...prev, ssl_certificate_id: '' })); setCertSearchTerm('') } }} />
                            <label htmlFor="ssl" className="text-sm">Enable SSL</label>
                        </div>

                        {/* SSL Certificate Dropdown */}
                        {formData.ssl_enabled && (
                            <div className="cert-dropdown-container relative">
                                <label htmlFor="ssl-certificate" className="label">SSL Certificate *</label>
                                <div className="relative">
                                    <button id="ssl-certificate" type="button" onClick={() => setShowCertDropdown(!showCertDropdown)}
                                        className="input w-full text-left flex items-center justify-between">
                                        <span className={selectedCertificate ? 'text-gray-900' : 'text-gray-400'}>
                                            {selectedCertificate
                                                ? `${selectedCertificate.name} - Expires: ${new Date(selectedCertificate.valid_to).toLocaleDateString()}`
                                                : 'Select SSL Certificate'}
                                        </span>
                                        <ChevronDownIcon className="w-5 h-5 text-gray-400" />
                                    </button>
                                    {showCertDropdown && (
                                        <div className="absolute z-50 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-64 overflow-hidden">
                                            <div className="p-2 border-b border-gray-200">
                                                <div className="relative">
                                                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                                                    <input type="text" placeholder="Search certificates..." className="w-full pl-9 pr-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
                                                        value={certSearchTerm} onChange={(e) => setCertSearchTerm(e.target.value)} onClick={(e) => e.stopPropagation()} />
                                                </div>
                                            </div>
                                            <div className="max-h-48 overflow-y-auto">
                                                {filteredCertificates.length === 0 ? (
                                                    <div className="p-3 text-sm text-gray-500 text-center">No certificates found</div>
                                                ) : (
                                                    filteredCertificates.map((cert) => (
                                                        <button key={cert.id} type="button"
                                                            onClick={() => { setFormData({ ...formData, ssl_certificate_id: cert.id }); setShowCertDropdown(false); setCertSearchTerm('') }}
                                                            className={`w-full px-3 py-2 text-left hover:bg-gray-50 transition-colors ${formData.ssl_certificate_id === cert.id ? 'bg-primary-50' : ''}`}>
                                                            <div className="text-sm font-medium text-gray-900">{cert.name}</div>
                                                            <div className="text-xs text-gray-500 mt-0.5">
                                                                {cert.common_name && <span className="mr-2">CN: {cert.common_name}</span>}
                                                                <span className={getCertExpiryColor(cert.valid_to)}>
                                                                    Expires: {new Date(cert.valid_to).toLocaleDateString()}
                                                                </span>
                                                            </div>
                                                        </button>
                                                    ))
                                                )}
                                            </div>
                                        </div>
                                    )}
                                </div>
                            </div>
                        )}

                        {/* Enabled */}
                        <div className="flex items-center gap-2">
                            <input type="checkbox" id="enabled" checked={formData.enabled}
                                onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })} />
                            <label htmlFor="enabled" className="text-sm">Enable Virtual Host</label>
                        </div>

                        {/* Defense Mode */}
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div>
                                <label htmlFor="defense_mode" className="label">Defense Mode</label>
                                <select id="defense_mode" className="input" value={formData.defense_mode}
                                    onChange={(e) => {
                                        const newMode = e.target.value;
                                        const updates = { defense_mode: newMode };
                                        if (newMode !== 'defense') {
                                            updates.bot_detection_enabled = false;
                                            updates.rate_limit_enabled = false;
                                            updates.region_filtering_enabled = false;
                                        }
                                        setFormData({ ...formData, ...updates });
                                    }}>
                                    <option value="defense">Defense (Active Mitigation)</option>
                                    <option value="audited">Audited (Log Only)</option>
                                    <option value="offline">Offline (WAF Disabled)</option>
                                </select>
                                <p className="text-xs text-gray-500 mt-1">Defense: Block attacks. Audited: Log only. Offline: WAF disabled.</p>
                            </div>
                        </div>

                        {/* Advanced Settings Toggle */}
                        <div className="border-t border-gray-200 pt-4">
                            <button type="button" onClick={() => setShowAdvancedSettings(!showAdvancedSettings)}
                                className="flex items-center gap-2 text-sm font-medium text-primary-600 hover:text-primary-700">
                                {showAdvancedSettings ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
                                Advanced Settings
                            </button>
                        </div>

                        {/* Advanced Settings Section */}
                        {showAdvancedSettings && (
                            <div className="space-y-4 border border-gray-200 rounded-lg p-4 bg-gray-50">
                                {/* WebSocket */}
                                <div className="flex items-center gap-2">
                                    <input type="checkbox" id="websocket" checked={formData.websocket_enabled}
                                        onChange={(e) => setFormData({ ...formData, websocket_enabled: e.target.checked })} />
                                    <label htmlFor="websocket" className="text-sm">Enable WebSocket Support</label>
                                </div>

                                {/* HTTP Version */}
                                <div>
                                    <label htmlFor="http_version" className="label">HTTP Version</label>
                                    <select id="http_version" className="input" value={formData.http_version}
                                        onChange={(e) => setFormData({ ...formData, http_version: e.target.value })}>
                                        <option value="http/1.1">HTTP/1.1</option>
                                        <option value="http/2">HTTP/2</option>
                                    </select>
                                </div>

                                {/* TLS Version */}
                                <div>
                                    <label htmlFor="tls_version" className="label">TLS Version (SSL/TLS Protocol)</label>
                                    <select id="tls_version" className="input" value={formData.tls_version}
                                        onChange={(e) => setFormData({ ...formData, tls_version: e.target.value })}>
                                        <option value="TLSv1.2">TLS 1.2 (Recommended)</option>
                                        <option value="TLSv1.3">TLS 1.3 (Most Secure)</option>
                                        <option value="TLSv1.2 TLSv1.3">TLS 1.2 and 1.3 (Compatible)</option>
                                        <option value="TLSv1.1 TLSv1.2 TLSv1.3">TLS 1.1, 1.2, 1.3 (Legacy)</option>
                                    </select>
                                    <p className="text-xs text-gray-500 mt-1">Select SSL/TLS protocol version for secure connections</p>
                                </div>

                                {/* Max Upload Size */}
                                <div>
                                    <label htmlFor="max_upload_size" className="label">Max Upload Size (MB)</label>
                                    <input id="max_upload_size" type="number" className="input" min="1" max="1024" value={formData.max_upload_size}
                                        onChange={(e) => setFormData({ ...formData, max_upload_size: Number.parseInt(e.target.value) || 10 })} />
                                    <p className="text-xs text-gray-500 mt-1">Maximum file upload size (client_max_body_size)</p>
                                </div>
                                
                                {/* Client Body Buffer Size */}
                                <div>
                                    <label htmlFor="client_body_buffer_size" className="label">Client Body Buffer Size (KB)</label>
                                    <input id="client_body_buffer_size" type="number" className="input" min="8" max="102400" value={formData.client_body_buffer_size}
                                        onChange={(e) => setFormData({ ...formData, client_body_buffer_size: Number.parseInt(e.target.value) || 1024 })} />
                                    <p className="text-xs text-gray-500 mt-1">Buffer size for request body. Set higher (e.g., 20480 KB = 20MB) for large multi-file uploads to avoid disk buffering.</p>
                                </div>

                                {/* Proxy Timeouts */}
                                <div className="grid grid-cols-2 gap-3">
                                    <div>
                                        <label htmlFor="proxy_read_timeout" className="label">Read Timeout (seconds)</label>
                                        <input id="proxy_read_timeout" type="number" className="input" min="1" max="3600" value={formData.proxy_read_timeout}
                                            onChange={(e) => setFormData({ ...formData, proxy_read_timeout: Number.parseInt(e.target.value) || 60 })} />
                                    </div>
                                    <div>
                                        <label htmlFor="proxy_connect_timeout" className="label">Connect Timeout (seconds)</label>
                                        <input id="proxy_connect_timeout" type="number" className="input" min="1" max="300" value={formData.proxy_connect_timeout}
                                            onChange={(e) => setFormData({ ...formData, proxy_connect_timeout: Number.parseInt(e.target.value) || 60 })} />
                                    </div>
                                </div>

                                {/* Bot Detection */}
                                <div className="border-t border-gray-300 pt-4">
                                    <div className="flex items-center gap-2 mb-3">
                                        <input type="checkbox" id="bot_detection" checked={formData.bot_detection_enabled}
                                            disabled={formData.defense_mode !== 'defense'}
                                            onChange={(e) => setFormData({ ...formData, bot_detection_enabled: e.target.checked })} />
                                        <label htmlFor="bot_detection" className={`text-sm font-medium ${formData.defense_mode !== 'defense' ? 'text-gray-400' : ''}`}>Enable Bot Detection</label>
                                        {formData.defense_mode !== 'defense' && <span className="text-xs text-orange-500">(Requires Defense Mode)</span>}
                                    </div>
                                    {formData.bot_detection_enabled && (
                                        <div className="space-y-3">
                                            <div>
                                                <label htmlFor="bot_detection_type" className="label">Challenge Type</label>
                                                <select id="bot_detection_type" className="input" value={formData.bot_detection_type}
                                                    onChange={(e) => setFormData({ ...formData, bot_detection_type: e.target.value })}>
                                                    <option value="turnstile">Cloudflare Turnstile</option>
                                                    <option value="captcha">Google reCAPTCHA</option>
                                                    <option value="slide_puzzle">Slide Puzzle</option>
                                                </select>
                                                <p className="text-xs text-gray-500 mt-1">Show challenge page before allowing access to this vhost</p>
                                            </div>
                                            {formData.bot_detection_type === 'captcha' && (
                                                <div>
                                                    <label htmlFor="recaptcha_version" className="label">reCAPTCHA Version</label>
                                                    <select id="recaptcha_version" className="input" value={formData.recaptcha_version || 'v2'}
                                                        onChange={(e) => setFormData({ ...formData, recaptcha_version: e.target.value })}>
                                                        <option value="v2">v2 (Checkbox - &quot;I&apos;m not a robot&quot;)</option>
                                                        <option value="v3">v3 (Invisible - Score based)</option>
                                                    </select>
                                                    <p className="text-xs text-gray-500 mt-1">
                                                        {formData.recaptcha_version === 'v3'
                                                            ? 'v3: Invisible challenge with automatic scoring (0.0-1.0)'
                                                            : 'v2: Visible checkbox with manual verification'}
                                                    </p>
                                                </div>
                                            )}
                                        </div>
                                    )}
                                </div>

                                {/* Rate Limiter */}
                                <div className="border-t border-gray-300 pt-4">
                                    <div className="flex items-center gap-2 mb-3">
                                        <input type="checkbox" id="rate_limit" checked={formData.rate_limit_enabled}
                                            disabled={formData.defense_mode !== 'defense'}
                                            onChange={(e) => setFormData({ ...formData, rate_limit_enabled: e.target.checked })} />
                                        <label htmlFor="rate_limit" className={`text-sm font-medium ${formData.defense_mode !== 'defense' ? 'text-gray-400' : ''}`}>Enable Rate Limiting</label>
                                        {formData.defense_mode !== 'defense' && <span className="text-xs text-orange-500">(Requires Defense Mode)</span>}
                                    </div>
                                    {formData.rate_limit_enabled && (
                                        <div className="grid grid-cols-2 gap-3">
                                            <div>
                                                <label htmlFor="rate_limit_requests" className="label">Max Requests</label>
                                                <input id="rate_limit_requests" type="number" className="input" min="1" max="10000" value={formData.rate_limit_requests}
                                                    onChange={(e) => setFormData({ ...formData, rate_limit_requests: Number.parseInt(e.target.value) || 100 })} />
                                            </div>
                                            <div>
                                                <label htmlFor="rate_limit_window" className="label">Time Window (seconds)</label>
                                                <input id="rate_limit_window" type="number" className="input" min="1" max="3600" value={formData.rate_limit_window}
                                                    onChange={(e) => setFormData({ ...formData, rate_limit_window: Number.parseInt(e.target.value) || 60 })} />
                                            </div>
                                            <div className="col-span-2">
                                                <p className="text-xs text-gray-500">
                                                    Limit: {formData.rate_limit_requests} requests per {formData.rate_limit_window} seconds per IP
                                                </p>
                                            </div>
                                        </div>
                                    )}
                                </div>

                                {/* Region Filtering */}
                                <div className="border-t border-gray-300 pt-4">
                                    <div className="flex items-center gap-2 mb-3">
                                        <input type="checkbox" id="region_filtering" checked={formData.region_filtering_enabled}
                                            disabled={formData.defense_mode !== 'defense'}
                                            onChange={(e) => setFormData({ ...formData, region_filtering_enabled: e.target.checked })} />
                                        <label htmlFor="region_filtering" className={`text-sm font-medium ${formData.defense_mode !== 'defense' ? 'text-gray-400' : ''}`}>Enable Region Filtering</label>
                                        {formData.defense_mode !== 'defense' && <span className="text-xs text-orange-500">(Requires Defense Mode)</span>}
                                    </div>
                                    {formData.region_filtering_enabled && (
                                        <div className="space-y-3">
                                            <div>
                                                <label htmlFor="region_whitelist" className="label">Whitelist Countries (ISO codes, e.g., US,GB,ID)</label>
                                                <input id="region_whitelist" type="text" className="input" placeholder="US,GB,ID,SG"
                                                    value={formData.region_whitelist?.join(',') || ''}
                                                    onChange={(e) => { const codes = e.target.value.split(',').map(c => c.trim().toUpperCase()).filter(Boolean); setFormData({ ...formData, region_whitelist: codes }) }} />
                                                <p className="text-xs text-gray-500 mt-1">If set, ONLY these countries are allowed. Leave empty to allow all except blacklisted.</p>
                                            </div>
                                            <div>
                                                <label htmlFor="region_blacklist" className="label">Blacklist Countries (ISO codes, e.g., CN,RU)</label>
                                                <input id="region_blacklist" type="text" className="input" placeholder="CN,RU,KP"
                                                    value={formData.region_blacklist?.join(',') || ''}
                                                    onChange={(e) => { const codes = e.target.value.split(',').map(c => c.trim().toUpperCase()).filter(Boolean); setFormData({ ...formData, region_blacklist: codes }) }} />
                                                <p className="text-xs text-gray-500 mt-1">These countries will be blocked. Only applies if whitelist is empty.</p>
                                            </div>
                                            <div className="bg-blue-50 border border-blue-200 rounded p-3 text-xs">
                                                <strong>Region Filtering Logic:</strong>
                                                <ul className="list-disc list-inside mt-1 space-y-1">
                                                    <li>If whitelist is set: ONLY whitelist countries are allowed</li>
                                                    <li>If whitelist is empty: All countries except blacklisted are allowed</li>
                                                    <li>Uses IP-based geolocation (may not be 100% accurate)</li>
                                                </ul>
                                            </div>
                                        </div>
                                    )}
                                </div>

                                {/* IP Access Control */}
                                <div className="border-t border-gray-300 pt-4">
                                    <label className="label flex items-center gap-2">
                                        <Settings2 className="w-4 h-4 text-primary-600" />
                                        IP Access Control (IP Groups)
                                    </label>
                                    <p className="text-xs text-gray-500 mb-3">
                                        Restrict access to this Virtual Host using pre-defined IP Groups. Enforcement is handled by Nginx.
                                    </p>
                                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 max-h-60 overflow-y-auto p-1">
                                        {ipGroups.map((group) => (
                                            <div key={group.id} className={`flex items-start gap-3 p-3 rounded-lg border transition-all cursor-pointer ${formData.ip_group_ids?.includes(group.id) ? 'bg-primary-50 border-primary-200' : 'bg-gray-50 border-gray-200 hover:border-gray-300'}`}
                                                onClick={() => {
                                                    const current = formData.ip_group_ids || [];
                                                    const updated = current.includes(group.id)
                                                        ? current.filter(id => id !== group.id)
                                                        : [...current, group.id];
                                                    setFormData({ ...formData, ip_group_ids: updated });
                                                }}>
                                                <input type="checkbox" className="mt-1" checked={formData.ip_group_ids?.includes(group.id)} readOnly />
                                                <div className="flex-1 min-w-0">
                                                    <div className="flex items-center justify-between gap-2">
                                                        <span className="text-sm font-semibold truncate">{group.name}</span>
                                                        <span className={`px-1.5 py-0.5 rounded text-[10px] font-bold uppercase ${group.type === 'whitelist' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
                                                            {group.type}
                                                        </span>
                                                    </div>
                                                    <p className="text-[11px] text-gray-500 truncate">{group.description || 'No description'}</p>
                                                </div>
                                            </div>
                                        ))}
                                        {ipGroups.length === 0 && (
                                            <div className="col-span-full py-4 text-center text-sm text-gray-400 border border-dashed border-gray-300 rounded-lg">
                                                No IP Groups found. Create them in the "IP Groups" menu.
                                            </div>
                                        )}
                                    </div>
                                </div>

                                {/* Security Settings Section */}
                                <div className="border-t border-gray-300 pt-4">
                                    <h3 className="text-sm font-semibold text-gray-700 flex items-center gap-2 mb-3">
                                        <Settings2 className="w-4 h-4 text-primary-600" />
                                        Advanced Security
                                    </h3>

                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                        <div className="space-y-3">
                                            <div className="flex items-center gap-2">
                                                <input type="checkbox" id="security_headers" checked={formData.security_headers_enabled}
                                                    onChange={(e) => setFormData({ ...formData, security_headers_enabled: e.target.checked })} />
                                                <label htmlFor="security_headers" className="text-sm">Enable Security Headers</label>
                                            </div>
                                            <p className="text-xs text-gray-500 ml-6">Adds X-Frame-Options, X-Content-Type-Options, etc.</p>

                                            <div className="flex items-center gap-2">
                                                <input type="checkbox" id="hide_server" checked={formData.hide_server_tokens}
                                                    onChange={(e) => setFormData({ ...formData, hide_server_tokens: e.target.checked })} />
                                                <label htmlFor="hide_server" className="text-sm">Hide Server Tokens</label>
                                            </div>
                                            <p className="text-xs text-gray-500 ml-6">Removes Nginx version from Server header and error pages.</p>
                                        </div>

                                        <div className="space-y-3 border-l border-gray-200 pl-4">
                                            <div className="flex items-center gap-2">
                                                <input type="checkbox" id="hsts" checked={formData.hsts_enabled}
                                                    onChange={(e) => setFormData({ ...formData, hsts_enabled: e.target.checked })} />
                                                <label htmlFor="hsts" className="text-sm font-medium">Enable HSTS</label>
                                            </div>
                                            {formData.hsts_enabled && (
                                                <div className="space-y-3 ml-6">
                                                    <div>
                                                        <label htmlFor="hsts_max_age" className="label text-xs">Max Age (seconds)</label>
                                                        <input id="hsts_max_age" type="number" className="input text-sm py-1" value={formData.hsts_max_age}
                                                            onChange={(e) => setFormData({ ...formData, hsts_max_age: Number.parseInt(e.target.value) || 31536000 })} />
                                                    </div>
                                                    <div className="flex items-center gap-2">
                                                        <input type="checkbox" id="hsts_subdomains" checked={formData.hsts_include_subdomains}
                                                            onChange={(e) => setFormData({ ...formData, hsts_include_subdomains: e.target.checked })} />
                                                        <label htmlFor="hsts_subdomains" className="text-sm">Include Subdomains</label>
                                                    </div>
                                                    <div className="flex items-center gap-2">
                                                        <input type="checkbox" id="hsts_preload" checked={formData.hsts_preload}
                                                            onChange={(e) => setFormData({ ...formData, hsts_preload: e.target.checked })} />
                                                        <label htmlFor="hsts_preload" className="text-sm">Preload</label>
                                                    </div>
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                </div>

                                {/* Performance & Acceleration Section */}
                                <div className="border-t border-gray-300 pt-4">
                                    <h3 className="text-sm font-semibold text-gray-700 flex items-center gap-2 mb-3">
                                        <Settings2 className="w-4 h-4 text-primary-600" />
                                        Performance & Speed
                                    </h3>

                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                        <div className="space-y-2">
                                            <div className="flex items-center gap-2">
                                                <input type="checkbox" id="brotli" checked={formData.brotli_enabled}
                                                    onChange={(e) => setFormData({ ...formData, brotli_enabled: e.target.checked })} />
                                                <label htmlFor="brotli" className="text-sm font-medium">Enable Brotli Compression</label>
                                            </div>
                                            <p className="text-xs text-gray-500 ml-6">Modern compression for faster page loads (better than Gzip).</p>
                                        </div>

                                        <div className="space-y-2">
                                            <div className="flex items-center gap-2">
                                                <input type="checkbox" id="http3" checked={formData.http3_enabled}
                                                    onChange={(e) => setFormData({ ...formData, http3_enabled: e.target.checked })} />
                                                <label htmlFor="http3" className="text-sm font-medium">Enable HTTP/3 (QUIC)</label>
                                            </div>
                                            <p className="text-xs text-gray-500 ml-6">Next-gen protocol over UDP. Requires UDP port 443 to be open.</p>
                                        </div>
                                    </div>
                                </div>

                                {/* Custom Headers */}
                                <div>
                                    <label className="label">Custom Headers</label>
                                    <div className="space-y-2">
                                        {Object.entries(formData.custom_headers || {}).map(([key, value], index) => (
                                            <div key={index} className="flex items-center gap-2">
                                                <input type="text" className="input flex-1 text-sm" value={key} disabled />
                                                <span className="text-gray-400">:</span>
                                                <input type="text" className="input flex-1 text-sm" value={value} disabled />
                                                <button type="button" onClick={() => {
                                                    const h = { ...formData.custom_headers }; delete h[key]; setFormData({ ...formData, custom_headers: h })
                                                }} className="text-red-600 hover:text-red-800"><Trash2 className="w-4 h-4" /></button>
                                            </div>
                                        ))}
                                        <div className="flex items-center gap-2">
                                            <input type="text" className="input flex-1 text-sm" placeholder="Header name"
                                                value={newHeader.key} onChange={(e) => setNewHeader({ ...newHeader, key: e.target.value })} />
                                            <span className="text-gray-400">:</span>
                                            <input type="text" className="input flex-1 text-sm" placeholder="Header value"
                                                value={newHeader.value} onChange={(e) => setNewHeader({ ...newHeader, value: e.target.value })} />
                                            <button type="button" onClick={(e) => {
                                                e.preventDefault(); e.stopPropagation()
                                                if (newHeader.key && newHeader.value) {
                                                    setFormData({ ...formData, custom_headers: { ...(formData.custom_headers || {}), [newHeader.key]: newHeader.value } })
                                                    setNewHeader({ key: '', value: '' })
                                                }
                                            }} className="p-2 text-white bg-green-600 hover:bg-green-700 rounded transition-colors" title="Add header">
                                                <Plus className="w-4 h-4" />
                                            </button>
                                        </div>
                                    </div>
                                </div>

                                {/* Custom Locations */}
                                <div>
                                    <label className="label">Custom Location Blocks</label>
                                    <p className="text-xs text-gray-500 mb-2">Add custom nginx location blocks for specific URL paths</p>
                                    <div className="space-y-3">
                                        {(formData.custom_locations || []).map((loc, index) => (
                                            <div key={index} className="border border-gray-300 rounded-lg p-3 bg-white">
                                                <div className="flex items-center justify-between mb-2">
                                                    <div className="flex items-center gap-2">
                                                        <span className="text-sm font-medium text-gray-700">location {loc.path}</span>
                                                        {loc.websocket_enabled && (
                                                            <span className="px-2 py-0.5 bg-purple-100 text-purple-700 rounded text-xs font-medium">WebSocket</span>
                                                        )}
                                                    </div>
                                                    <button type="button" onClick={() => {
                                                        const n = [...formData.custom_locations]; n.splice(index, 1); setFormData({ ...formData, custom_locations: n })
                                                    }} className="text-red-600 hover:text-red-800"><Trash2 className="w-4 h-4" /></button>
                                                </div>
                                                {loc.proxy_pass && !loc.backends?.length && (
                                                    <p className="text-xs text-gray-600">proxy_pass: {loc.proxy_pass}</p>
                                                )}
                                                {loc.backends && loc.backends.length > 0 && (
                                                    <div className="space-y-1 mb-1">
                                                        <p className="text-xs text-gray-600 font-medium">Load Balanced Backends ({loc.load_balance_method || 'round_robin'}):</p>
                                                        {loc.backends.map((b, i) => (
                                                            <p key={i} className="text-xs text-gray-600 font-mono pl-2 border-l-2 border-gray-300">{b}</p>
                                                        ))}
                                                    </div>
                                                )}
                                                {loc.config && <pre className="text-xs text-gray-600 mt-1 whitespace-pre-wrap">{loc.config}</pre>}
                                            </div>
                                        ))}

                                        {/* Add New Location */}
                                        <div className="border border-dashed border-gray-300 rounded-lg p-3 bg-white">
                                            <div className="space-y-2">
                                                <input type="text" className="input text-sm" placeholder="Location path (e.g., /api, /svc-base)"
                                                    value={newLocation.path} onChange={(e) => setNewLocation({ ...newLocation, path: e.target.value })} />
                                                <div className="relative">
                                                    <input type="text" className="input text-sm pr-10" placeholder="Proxy pass URL (e.g., http://backend:8080)"
                                                        value={newLocation.proxy_pass} onChange={(e) => setNewLocation({ ...newLocation, proxy_pass: e.target.value })} />
                                                    {locationBackendCheck.status && !newLocation.backends?.length && (
                                                        <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
                                                            {locationBackendCheck.status === 'checking' && <Loader2 className="w-4 h-4 text-blue-500 animate-spin" />}
                                                            {locationBackendCheck.status === 'success' && <CheckCircle className="w-4 h-4 text-green-500" />}
                                                            {locationBackendCheck.status === 'error' && <AlertCircle className="w-4 h-4 text-red-500" />}
                                                            {locationBackendCheck.status === 'warning' && <AlertCircle className="w-4 h-4 text-yellow-500" />}
                                                        </div>
                                                    )}
                                                </div>
                                                {locationBackendCheck.message && !newLocation.backends?.length && (
                                                    <p className={`text-xs flex items-center gap-1 ${getLocationCheckColor(locationBackendCheck.status)}`}>
                                                        {locationBackendCheck.message}
                                                    </p>
                                                )}

                                                {/* Additional Backends for Location */}
                                                <div className="border-t border-gray-200 pt-2 mt-2">
                                                    {newLocation.backends && newLocation.backends.length > 0 && (
                                                        <div className="space-y-2 mb-2">
                                                            {newLocation.backends.map((backend, index) => (
                                                                <div key={index} className="flex items-center gap-2">
                                                                    <input type="text" className="input text-sm py-1" value={backend}
                                                                        onChange={(e) => { const n = [...newLocation.backends]; n[index] = e.target.value; setNewLocation({ ...newLocation, backends: n }) }}
                                                                        placeholder="http://backend:8080" />
                                                                    <button type="button" onClick={() => { const n = newLocation.backends.filter((_, i) => i !== index); setNewLocation({ ...newLocation, backends: n }) }}
                                                                        className="p-1 text-red-500 hover:bg-red-50 rounded"><Trash2 className="w-3 h-3" /></button>
                                                                </div>
                                                            ))}
                                                        </div>
                                                    )}
                                                    <button type="button" onClick={() => setNewLocation({ ...newLocation, backends: [...(newLocation.backends || []), ''] })}
                                                        className="text-xs text-blue-600 hover:text-blue-800 flex items-center gap-1">
                                                        <Plus className="w-3 h-3" /> Add Backend Server
                                                    </button>
                                                    {newLocation.backends && newLocation.backends.length > 0 && (
                                                        <div className="mt-2">
                                                            <label className="text-xs text-gray-600 block mb-1">Load Balance Method</label>
                                                            <select className="input text-xs py-1" value={newLocation.load_balance_method}
                                                                onChange={(e) => setNewLocation({ ...newLocation, load_balance_method: e.target.value })}>
                                                                <option value="round_robin">Round Robin</option>
                                                                <option value="least_conn">Least Connections</option>
                                                                <option value="ip_hash">IP Hash</option>
                                                            </select>
                                                        </div>
                                                    )}
                                                </div>
                                                <textarea className="input text-sm" rows="3" placeholder="Additional nginx config (optional)"
                                                    value={newLocation.config} onChange={(e) => setNewLocation({ ...newLocation, config: e.target.value })} />
                                                <label className="flex items-center gap-2 cursor-pointer">
                                                    <input type="checkbox" className="w-4 h-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                                                        checked={newLocation.websocket_enabled || false}
                                                        onChange={(e) => setNewLocation({ ...newLocation, websocket_enabled: e.target.checked })} />
                                                    <span className="text-sm text-gray-700">Enable WebSocket Support</span>
                                                    <span className="text-xs text-gray-500">(Adds Upgrade/Connection headers)</span>
                                                </label>
                                                <button type="button" onClick={(e) => {
                                                    e.preventDefault(); e.stopPropagation()
                                                    if (newLocation.path) {
                                                        // If proxy_pass is set and there are additional backends,
                                                        // merge proxy_pass as the first backend for unified load balancing
                                                        const loc = { ...newLocation }
                                                        if (loc.proxy_pass && loc.backends && loc.backends.length > 0) {
                                                            loc.backends = [loc.proxy_pass, ...loc.backends]
                                                            loc.proxy_pass = ''
                                                        }
                                                        setFormData({ ...formData, custom_locations: [...(formData.custom_locations || []), loc] })
                                                        setNewLocation({ path: '', proxy_pass: '', config: '', websocket_enabled: false, backends: [], load_balance_method: 'round_robin' })
                                                        setLocationBackendCheck({ status: null, message: '' })
                                                    }
                                                }} className="btn btn-primary text-sm w-full flex items-center justify-center gap-2">
                                                    <Plus className="w-4 h-4" /> Add Location Block
                                                </button>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        )}

                        {/* Action Buttons */}
                        <div className="flex gap-2 pt-4 border-t border-gray-200">
                            <button type="submit" disabled={saving} className="btn btn-primary flex-1 flex items-center justify-center gap-2">
                                {saving ? <><Loader2 className="w-4 h-4 animate-spin" /> Saving...</> : <><Save className="w-4 h-4" /> {isEditMode ? 'Update' : 'Create'}</>}
                            </button>
                            <button type="button" onClick={() => navigate('/vhosts')} className="btn btn-secondary flex-1">Cancel</button>
                        </div>
                    </form>
                </div >
            )}

            {/* Config Editor Tab */}
            {
                activeTab === 'editor' && (
                    <div className="card">
                        {configLoading ? (
                            <div className="flex justify-center items-center h-96">
                                <Loader2 className="w-8 h-8 animate-spin text-primary-600" />
                            </div>
                        ) : (
                            <>
                                <div className="flex justify-between items-center mb-4 pb-4 border-b border-gray-200">
                                    <div>
                                        <h2 className="text-xl font-semibold text-gray-900">{isEditMode ? configDomain : 'Custom Nginx Configuration'}</h2>
                                        <p className="text-sm text-gray-500 mt-1">{isEditMode ? 'Directly edit the generated Nginx config file' : 'Write custom Nginx config to include when creating this VHost'}</p>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <button onClick={() => setConfigContent(originalConfig)} disabled={!hasConfigChanges}
                                            className="btn btn-secondary flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed">
                                            Reset
                                        </button>
                                        <button onClick={handleSaveConfig} disabled={!hasConfigChanges || saving}
                                            className="btn btn-primary flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed">
                                            {saving ? <><Loader2 className="w-4 h-4 animate-spin" /> Saving...</> : <><Save className="w-4 h-4" /> Save Config</>}
                                        </button>
                                    </div>
                                </div>
                                <div className="rounded-lg overflow-hidden border border-gray-700">
                                    <CodeMirror value={configContent} height="600px" theme={monokai} extensions={[nginx]}
                                        onChange={(value) => setConfigContent(value)}
                                        basicSetup={{
                                            lineNumbers: true, highlightActiveLineGutter: true, highlightSpecialChars: true,
                                            foldGutter: true, drawSelection: true, dropCursor: true, allowMultipleSelections: true,
                                            indentOnInput: true, bracketMatching: true, closeBrackets: true, autocompletion: false,
                                            rectangularSelection: true, crosshairCursor: true, highlightActiveLine: true,
                                            highlightSelectionMatches: true, closeBracketsKeymap: true, searchKeymap: true,
                                            foldKeymap: true, completionKeymap: false, lintKeymap: true,
                                        }}
                                        style={{ fontSize: '14px', fontFamily: '"Fira Code", "Consolas", "Monaco", monospace' }}
                                    />
                                </div>
                                <div className="flex justify-between items-center mt-4 pt-4 border-t border-gray-200">
                                    <p className="text-sm text-gray-600">
                                        {hasConfigChanges
                                            ? <span className="text-yellow-600 font-medium">● Unsaved changes</span>
                                            : <span className="text-green-600">✓ No changes</span>}
                                    </p>
                                    <p className="text-xs text-gray-500">Note: Config will be backed up automatically before saving</p>
                                </div>
                            </>
                        )}
                    </div>
                )
            }

            {/* Preview Modal */}

        </div >
    )
}

export default VHostForm
