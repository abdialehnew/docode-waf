import { Settings as SettingsIcon } from 'lucide-react'

const Settings = () => {
  return (
    <div>
      <h1 className="text-3xl font-bold mb-8">Settings</h1>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
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
