import { useState, useEffect } from 'react'
import { Settings as SettingsIcon, Upload, X, Mail, Eye, EyeOff } from 'lucide-react'
import api from '../services/api'
import logger from '../utils/logger'

const Settings = () => {
  const [appSettings, setAppSettings] = useState({
    app_name: 'Docode WAF',
    app_logo: null,
    signup_enabled: true,
    smtp_host: '',
    smtp_port: 587,
    smtp_username: '',
    smtp_password: '',
    smtp_from_email: '',
    smtp_from_name: 'Docode WAF',
    smtp_use_tls: true,
    turnstile_enabled: false,
    turnstile_login_enabled: false,
    turnstile_register_enabled: false
  })
  const [logoPreview, setLogoPreview] = useState(null)
  const [loading, setLoading] = useState(false)
  const [showPassword, setShowPassword] = useState(false)

  useEffect(() => {
    loadAppSettings()
  }, [])

  const loadAppSettings = async () => {
    try {
      const response = await api.get('/settings/app')
      if (response.data) {
        setAppSettings(response.data)
        if (response.data.app_logo) {
          setLogoPreview(response.data.app_logo)
        }
      }
    } catch (error) {
      logger.error('Failed to load app settings:', error)
    }
  }

  const handleLogoChange = (e) => {
    const file = e.target.files[0]
    if (file) {
      // Validate file type
      if (!file.type.startsWith('image/')) {
        alert('Please select an image file')
        return
      }

      // Validate file size (max 2MB)
      if (file.size > 2 * 1024 * 1024) {
        alert('Logo file size must be less than 2MB')
        return
      }

      // Create preview
      const reader = new FileReader()
      reader.onloadend = () => {
        setLogoPreview(reader.result)
        setAppSettings({ ...appSettings, app_logo: reader.result })
      }
      reader.readAsDataURL(file)
    }
  }

  const handleRemoveLogo = () => {
    setLogoPreview(null)
    setAppSettings({ ...appSettings, app_logo: null })
  }

  const handleSaveAppSettings = async () => {
    setLoading(true)
    try {
      await api.post('/settings/app', appSettings)
      alert('Application settings saved successfully!')
      // Reload page to update logo in sidebar
      window.location.reload()
    } catch (error) {
      logger.error('Failed to save app settings:', error)
      alert('Failed to save settings: ' + (error.response?.data?.error || error.message))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <h1 className="text-3xl font-bold mb-8">Settings</h1>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Application Settings */}
        <div className="card lg:col-span-2">
          <h2 className="text-xl font-semibold mb-4">Application Settings</h2>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-4">
              <div>
                <label className="label">Application Name</label>
                <input
                  type="text"
                  className="input"
                  placeholder="Enter application name"
                  value={appSettings.app_name}
                  onChange={(e) => setAppSettings({ ...appSettings, app_name: e.target.value })}
                />
                <p className="text-xs text-gray-500 mt-1">This name will appear in the sidebar and page title</p>
              </div>
            </div>

            <div className="space-y-4">
              <div>
                <label className="label">Application Logo</label>
                <div className="space-y-3">
                  {logoPreview ? (
                    <div className="relative inline-block">
                      <img
                        src={logoPreview}
                        alt="Logo preview"
                        className="h-20 w-20 object-contain border border-gray-300 rounded-lg p-2 bg-white"
                      />
                      <button
                        type="button"
                        onClick={handleRemoveLogo}
                        className="absolute -top-2 -right-2 bg-red-500 text-white rounded-full p-1 hover:bg-red-600"
                      >
                        <X className="w-4 h-4" />
                      </button>
                    </div>
                  ) : (
                    <div className="border-2 border-dashed border-gray-300 rounded-lg p-4 text-center">
                      <Upload className="w-8 h-8 mx-auto text-gray-400 mb-2" />
                      <p className="text-sm text-gray-600">No logo uploaded</p>
                    </div>
                  )}

                  <div>
                    <input
                      type="file"
                      id="logo-upload"
                      accept="image/*"
                      onChange={handleLogoChange}
                      className="hidden"
                    />
                    <label
                      htmlFor="logo-upload"
                      className="btn btn-secondary text-sm cursor-pointer inline-block"
                    >
                      {logoPreview ? 'Change Logo' : 'Upload Logo'}
                    </label>
                    <p className="text-xs text-gray-500 mt-1">
                      PNG, JPG, SVG up to 2MB. Recommended: 200x200px
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <button
            onClick={handleSaveAppSettings}
            disabled={loading}
            className="btn btn-primary w-full mt-6 disabled:opacity-50"
          >
            {loading ? 'Saving...' : 'Save Application Settings'}
          </button>
        </div>

        {/* Authentication Settings */}
        <div className="card lg:col-span-2">
          <h2 className="text-xl font-semibold mb-4">Authentication Settings</h2>

          <div className="space-y-4">
            <div>
              <label className="label">User Registration</label>
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="signup-enabled"
                  checked={appSettings.signup_enabled}
                  onChange={(e) => setAppSettings({ ...appSettings, signup_enabled: e.target.checked })}
                />
                <label htmlFor="signup-enabled" className="text-sm">
                  Enable user self-registration (show signup link on login page)
                </label>
              </div>
              <p className="text-xs text-gray-500 mt-1">
                When disabled, only existing users can login. New users can only be created by administrators.
              </p>
            </div>
          </div>
        </div>

        {/* Turnstile (CAPTCHA) Settings */}
        <div className="card lg:col-span-2">
          <h2 className="text-xl font-semibold mb-4">Cloudflare Turnstile (CAPTCHA)</h2>

          <p className="text-sm text-gray-600 mb-4">
            Configure Turnstile CAPTCHA verification on login and registration pages to protect against bots.
            Note: Turnstile Site Key and Secret Key must be configured via environment variables.
          </p>

          <div className="space-y-4">
            <div>
              <label className="label">Global Turnstile</label>
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="turnstile-enabled"
                  checked={appSettings.turnstile_enabled}
                  onChange={(e) => setAppSettings({ ...appSettings, turnstile_enabled: e.target.checked })}
                />
                <label htmlFor="turnstile-enabled" className="text-sm">
                  Enable Turnstile CAPTCHA verification
                </label>
              </div>
              <p className="text-xs text-gray-500 mt-1">
                Master switch to enable/disable Turnstile on all pages. Individual page settings below only apply if this is enabled.
              </p>
            </div>

            <div className={`pl-6 space-y-3 ${!appSettings.turnstile_enabled ? 'opacity-50' : ''}`}>
              <div>
                <div className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    id="turnstile-login-enabled"
                    checked={appSettings.turnstile_login_enabled}
                    onChange={(e) => setAppSettings({ ...appSettings, turnstile_login_enabled: e.target.checked })}
                    disabled={!appSettings.turnstile_enabled}
                  />
                  <label htmlFor="turnstile-login-enabled" className="text-sm">
                    Enable Turnstile on Login page
                  </label>
                </div>
              </div>

              <div>
                <div className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    id="turnstile-register-enabled"
                    checked={appSettings.turnstile_register_enabled}
                    onChange={(e) => setAppSettings({ ...appSettings, turnstile_register_enabled: e.target.checked })}
                    disabled={!appSettings.turnstile_enabled}
                  />
                  <label htmlFor="turnstile-register-enabled" className="text-sm">
                    Enable Turnstile on Registration page
                  </label>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mt-4">
            <h3 className="text-sm font-semibold text-yellow-900 mb-2">Environment Variables Required:</h3>
            <ul className="text-xs text-yellow-800 space-y-1">
              <li><code className="bg-yellow-100 px-1 rounded">TURNSTILE_SITE_KEY</code> - Your Cloudflare Turnstile site key</li>
              <li><code className="bg-yellow-100 px-1 rounded">TURNSTILE_SECRET_KEY</code> - Your Cloudflare Turnstile secret key</li>
            </ul>
            <p className="text-xs text-yellow-700 mt-2">
              Get your keys from <a href="https://dash.cloudflare.com/?to=/:account/turnstile" target="_blank" rel="noopener noreferrer" className="underline">Cloudflare Dashboard</a>
            </p>
          </div>
        </div>

        {/* SMTP Settings */}
        <div className="card lg:col-span-2">
          <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
            <Mail className="w-5 h-5" />
            SMTP Configuration (Email Sending)
          </h2>

          <p className="text-sm text-gray-600 mb-4">
            Configure SMTP settings to enable email notifications (password reset, alerts, etc.)
          </p>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="label">SMTP Host *</label>
              <input
                type="text"
                className="input"
                placeholder="smtp.gmail.com"
                value={appSettings.smtp_host || ''}
                onChange={(e) => setAppSettings({ ...appSettings, smtp_host: e.target.value })}
              />
            </div>

            <div>
              <label className="label">SMTP Port *</label>
              <input
                type="number"
                className="input"
                placeholder="587"
                value={appSettings.smtp_port || 587}
                onChange={(e) => setAppSettings({ ...appSettings, smtp_port: parseInt(e.target.value) })}
              />
            </div>

            <div>
              <label className="label">SMTP Username</label>
              <input
                type="text"
                className="input"
                placeholder="your-email@gmail.com"
                value={appSettings.smtp_username || ''}
                onChange={(e) => setAppSettings({ ...appSettings, smtp_username: e.target.value })}
              />
            </div>

            <div>
              <label className="label">SMTP Password</label>
              <div className="relative">
                <input
                  type={showPassword ? "text" : "password"}
                  className="input pr-10"
                  placeholder="••••••••"
                  value={appSettings.smtp_password || ''}
                  onChange={(e) => setAppSettings({ ...appSettings, smtp_password: e.target.value })}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600"
                >
                  {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                </button>
              </div>
            </div>

            <div>
              <label className="label">From Email *</label>
              <input
                type="email"
                className="input"
                placeholder="noreply@yourdomain.com"
                value={appSettings.smtp_from_email || ''}
                onChange={(e) => setAppSettings({ ...appSettings, smtp_from_email: e.target.value })}
              />
              <p className="text-xs text-gray-500 mt-1">Email address that appears as sender</p>
            </div>

            <div>
              <label className="label">From Name</label>
              <input
                type="text"
                className="input"
                placeholder="Docode WAF"
                value={appSettings.smtp_from_name || 'Docode WAF'}
                onChange={(e) => setAppSettings({ ...appSettings, smtp_from_name: e.target.value })}
              />
              <p className="text-xs text-gray-500 mt-1">Name that appears as sender</p>
            </div>

            <div className="md:col-span-2">
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="smtp-use-tls"
                  checked={appSettings.smtp_use_tls}
                  onChange={(e) => setAppSettings({ ...appSettings, smtp_use_tls: e.target.checked })}
                />
                <label htmlFor="smtp-use-tls" className="text-sm">
                  Use TLS/SSL (recommended for security)
                </label>
              </div>
            </div>
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mt-4">
            <h3 className="text-sm font-semibold text-blue-900 mb-2">Popular SMTP Providers:</h3>
            <ul className="text-xs text-blue-800 space-y-1">
              <li><strong>Gmail:</strong> smtp.gmail.com:587 (Use App Password, not regular password)</li>
              <li><strong>Outlook:</strong> smtp-mail.outlook.com:587</li>
              <li><strong>SendGrid:</strong> smtp.sendgrid.net:587</li>
              <li><strong>Mailgun:</strong> smtp.mailgun.org:587</li>
            </ul>
          </div>
        </div>

        {/* Keep existing WAF Configuration card */}
        <div className="card">
          <h2 className="text-xl font-semibold mb-4">WAF Configuration</h2>

          <div className="space-y-4">
            <div>
              <label className="label">Rate Limiting</label>
              <div className="flex items-center gap-2 mb-2">
                <input type="checkbox" id="rate-limit" defaultChecked />
                <label htmlFor="rate-limit" className="text-sm">Enable Rate Limiting</label>
              </div>
              <input
                type="number"
                className="input"
                placeholder="Requests per second"
                defaultValue="100"
              />
            </div>

            <div>
              <label className="label">HTTP Flood Protection</label>
              <div className="flex items-center gap-2 mb-2">
                <input type="checkbox" id="http-flood" defaultChecked />
                <label htmlFor="http-flood" className="text-sm">Enable HTTP Flood Protection</label>
              </div>
              <input
                type="number"
                className="input"
                placeholder="Max requests per minute"
                defaultValue="1000"
              />
            </div>

            <div>
              <label className="label">Anti-Bot Protection</label>
              <div className="flex items-center gap-2">
                <input type="checkbox" id="anti-bot" defaultChecked />
                <label htmlFor="anti-bot" className="text-sm">Enable Anti-Bot Protection</label>
              </div>
            </div>
          </div>

          <button className="btn btn-primary w-full mt-6">
            Save Configuration
          </button>
        </div>

        <div className="card">
          <h2 className="text-xl font-semibold mb-4">SSL/TLS Settings</h2>

          <div className="space-y-4">
            <div>
              <label className="label">Auto SSL</label>
              <div className="flex items-center gap-2">
                <input type="checkbox" id="auto-ssl" />
                <label htmlFor="auto-ssl" className="text-sm">Enable automatic SSL certificate management</label>
              </div>
            </div>

            <div>
              <label className="label">Certificate Directory</label>
              <input
                type="text"
                className="input"
                placeholder="/path/to/certs"
                defaultValue="./certs"
              />
            </div>
          </div>

          <button className="btn btn-primary w-full mt-6">
            Save SSL Settings
          </button>
        </div>

        <div className="card">
          <h2 className="text-xl font-semibold mb-4">Database Settings</h2>

          <div className="space-y-4">
            <div>
              <label className="label">Host</label>
              <input
                type="text"
                className="input"
                defaultValue="localhost"
              />
            </div>

            <div>
              <label className="label">Port</label>
              <input
                type="number"
                className="input"
                defaultValue="5432"
              />
            </div>

            <div>
              <label className="label">Database Name</label>
              <input
                type="text"
                className="input"
                defaultValue="docode_waf"
              />
            </div>
          </div>

          <button className="btn btn-primary w-full mt-6">
            Save Database Settings
          </button>
        </div>

        <div className="card">
          <h2 className="text-xl font-semibold mb-4">Logging Settings</h2>

          <div className="space-y-4">
            <div>
              <label className="label">Log Level</label>
              <select className="input">
                <option value="debug">Debug</option>
                <option value="info" selected>Info</option>
                <option value="warn">Warning</option>
                <option value="error">Error</option>
              </select>
            </div>

            <div>
              <label className="label">Log Format</label>
              <select className="input">
                <option value="json" selected>JSON</option>
                <option value="text">Text</option>
              </select>
            </div>

            <div>
              <label className="label">Log Output</label>
              <select className="input">
                <option value="stdout" selected>Standard Output</option>
                <option value="file">File</option>
              </select>
            </div>
          </div>

          <button className="btn btn-primary w-full mt-6">
            Save Logging Settings
          </button>
        </div>
      </div>
    </div>
  )
}

export default Settings
