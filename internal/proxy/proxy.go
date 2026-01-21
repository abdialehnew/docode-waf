package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/aleh/docode-waf/internal/config"
	"github.com/aleh/docode-waf/internal/models"
	"github.com/aleh/docode-waf/internal/services"
)

type ReverseProxy struct {
	config       *config.Config
	vhosts       map[string]*models.VHost
	proxies      map[string]*httputil.ReverseProxy
	transport    *http.Transport
	vhostService *services.VHostService
}

func NewReverseProxy(cfg *config.Config, vhostService *services.VHostService) *ReverseProxy {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   300 * time.Second, // 5 minutes to match nginx
			KeepAlive: 300 * time.Second, // 5 minutes keepalive
		}).DialContext,
		ForceAttemptHTTP2:     false, // Disable HTTP/2 to avoid connection issues
		MaxIdleConns:          200,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       300 * time.Second, // 5 minutes
		TLSHandshakeTimeout:   30 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
		ResponseHeaderTimeout: 300 * time.Second, // 5 minutes for slow APIs
		DisableKeepAlives:     false,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	return &ReverseProxy{
		config:       cfg,
		vhosts:       make(map[string]*models.VHost),
		proxies:      make(map[string]*httputil.ReverseProxy),
		transport:    transport,
		vhostService: vhostService,
	}
}

func (rp *ReverseProxy) LoadVHosts(vhosts []*models.VHost) {
	rp.vhosts = make(map[string]*models.VHost)
	rp.proxies = make(map[string]*httputil.ReverseProxy)

	for _, vhost := range vhosts {
		if !vhost.Enabled {
			continue
		}

		rp.vhosts[vhost.Domain] = vhost

		target, err := url.Parse(vhost.BackendURL)
		if err != nil {
			continue
		}

		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.Transport = rp.transport
		proxy.ErrorHandler = rp.errorHandler

		// Modify request before forwarding
		director := proxy.Director
		proxy.Director = func(req *http.Request) {
			director(req)
			req.Header.Set("X-Forwarded-Host", req.Host)
			req.Header.Set("X-Origin-Host", target.Host)
			req.Header.Set("X-Real-IP", getClientIP(req))
		}

		rp.proxies[vhost.Domain] = proxy
	}
}

// ReloadVHosts reloads all vhosts from database
func (rp *ReverseProxy) ReloadVHosts() error {
	vhosts, err := rp.vhostService.ListVHosts()
	if err != nil {
		return err
	}
	rp.LoadVHosts(vhosts)
	return nil
}

func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	// Remove port from host if present
	if colonPos := strings.Index(host, ":"); colonPos != -1 {
		host = host[:colonPos]
	}

	proxy, exists := rp.proxies[host]
	if !exists {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(getMaintenancePageHTML(host)))
		return
	}

	// Add context with start time for logging
	ctx := context.WithValue(r.Context(), "start_time", time.Now())
	r = r.WithContext(ctx)

	proxy.ServeHTTP(w, r)
}

func (rp *ReverseProxy) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, "Bad Gateway", http.StatusBadGateway)
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// ResponseWriter wrapper to capture status code and bytes written
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// getMaintenancePageHTML returns a styled maintenance page for unregistered virtual hosts
func getMaintenancePageHTML(domain string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Under Maintenance - %s</title>
    <style>
        :root {
            --bg-primary: #f8fafc;
            --bg-secondary: #ffffff;
            --text-primary: #1e293b;
            --text-secondary: #64748b;
            --accent: #6366f1;
            --accent-light: #818cf8;
            --border: #e2e8f0;
        }
        
        [data-theme="dark"] {
            --bg-primary: #0f172a;
            --bg-secondary: #1e293b;
            --text-primary: #f1f5f9;
            --text-secondary: #94a3b8;
            --accent: #818cf8;
            --accent-light: #a5b4fc;
            --border: #334155;
        }
        
        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: var(--bg-primary);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
            transition: all 0.3s ease;
        }
        
        .container {
            background: var(--bg-secondary);
            border-radius: 24px;
            box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.1);
            max-width: 600px;
            width: 100%%;
            padding: 50px 40px;
            text-align: center;
            border: 1px solid var(--border);
            position: relative;
            overflow: hidden;
        }
        
        .theme-toggle {
            position: absolute;
            top: 20px;
            right: 20px;
            background: var(--bg-primary);
            border: 1px solid var(--border);
            border-radius: 50%%;
            width: 45px;
            height: 45px;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: all 0.3s ease;
        }
        
        .theme-toggle:hover {
            transform: scale(1.1);
            box-shadow: 0 4px 15px rgba(99, 102, 241, 0.3);
        }
        
        .theme-toggle svg {
            width: 22px;
            height: 22px;
            fill: var(--accent);
            transition: all 0.3s ease;
        }
        
        .sun-icon { display: none; }
        .moon-icon { display: block; }
        [data-theme="dark"] .sun-icon { display: block; }
        [data-theme="dark"] .moon-icon { display: none; }
        
        .illustration {
            margin-bottom: 30px;
        }
        
        .illustration svg {
            width: 200px;
            height: 200px;
        }
        
        .gear {
            fill: var(--accent);
            transform-origin: center;
            animation: rotate 8s linear infinite;
        }
        
        .gear-small {
            fill: var(--accent-light);
            transform-origin: center;
            animation: rotate-reverse 6s linear infinite;
        }
        
        @keyframes rotate {
            from { transform: rotate(0deg); }
            to { transform: rotate(360deg); }
        }
        
        @keyframes rotate-reverse {
            from { transform: rotate(360deg); }
            to { transform: rotate(0deg); }
        }
        
        .pulse {
            animation: pulse 2s ease-in-out infinite;
        }
        
        @keyframes pulse {
            0%%, 100%% { opacity: 1; }
            50%% { opacity: 0.5; }
        }
        
        h1 {
            color: var(--text-primary);
            font-size: 28px;
            margin-bottom: 15px;
            font-weight: 700;
        }
        
        .domain {
            color: var(--accent);
            font-family: 'Courier New', monospace;
            font-size: 16px;
            background: var(--bg-primary);
            padding: 8px 16px;
            border-radius: 8px;
            display: inline-block;
            margin-bottom: 20px;
            border: 1px solid var(--border);
        }
        
        .subtitle {
            color: var(--text-secondary);
            font-size: 16px;
            line-height: 1.7;
            margin-bottom: 30px;
        }
        
        .status-box {
            background: linear-gradient(135deg, var(--accent) 0%%, var(--accent-light) 100%%);
            color: white;
            padding: 20px 30px;
            border-radius: 16px;
            margin-bottom: 25px;
        }
        
        .status-box h3 {
            font-size: 18px;
            margin-bottom: 8px;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 10px;
        }
        
        .status-box p {
            font-size: 14px;
            opacity: 0.9;
        }
        
        .info-cards {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 15px;
            margin-top: 25px;
        }
        
        .info-card {
            background: var(--bg-primary);
            padding: 20px;
            border-radius: 12px;
            border: 1px solid var(--border);
        }
        
        .info-card h4 {
            color: var(--accent);
            font-size: 14px;
            margin-bottom: 5px;
        }
        
        .info-card p {
            color: var(--text-secondary);
            font-size: 13px;
        }
        
        .footer {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid var(--border);
            color: var(--text-secondary);
            font-size: 12px;
        }
        
        @media (max-width: 500px) {
            .container { padding: 40px 25px; }
            h1 { font-size: 24px; }
            .info-cards { grid-template-columns: 1fr; }
            .illustration svg { width: 150px; height: 150px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <button class="theme-toggle" onclick="toggleTheme()" aria-label="Toggle theme">
            <svg class="sun-icon" viewBox="0 0 24 24"><path d="M12 7c-2.76 0-5 2.24-5 5s2.24 5 5 5 5-2.24 5-5-2.24-5-5-5zM2 13h2c.55 0 1-.45 1-1s-.45-1-1-1H2c-.55 0-1 .45-1 1s.45 1 1 1zm18 0h2c.55 0 1-.45 1-1s-.45-1-1-1h-2c-.55 0-1 .45-1 1s.45 1 1 1zM11 2v2c0 .55.45 1 1 1s1-.45 1-1V2c0-.55-.45-1-1-1s-1 .45-1 1zm0 18v2c0 .55.45 1 1 1s1-.45 1-1v-2c0-.55-.45-1-1-1s-1 .45-1 1zM5.99 4.58c-.39-.39-1.03-.39-1.41 0-.39.39-.39 1.03 0 1.41l1.06 1.06c.39.39 1.03.39 1.41 0s.39-1.03 0-1.41L5.99 4.58zm12.37 12.37c-.39-.39-1.03-.39-1.41 0-.39.39-.39 1.03 0 1.41l1.06 1.06c.39.39 1.03.39 1.41 0 .39-.39.39-1.03 0-1.41l-1.06-1.06zm1.06-10.96c.39-.39.39-1.03 0-1.41-.39-.39-1.03-.39-1.41 0l-1.06 1.06c-.39.39-.39 1.03 0 1.41s1.03.39 1.41 0l1.06-1.06zM7.05 18.36c.39-.39.39-1.03 0-1.41-.39-.39-1.03-.39-1.41 0l-1.06 1.06c-.39.39-.39 1.03 0 1.41s1.03.39 1.41 0l1.06-1.06z"/></svg>
            <svg class="moon-icon" viewBox="0 0 24 24"><path d="M12 3c-4.97 0-9 4.03-9 9s4.03 9 9 9 9-4.03 9-9c0-.46-.04-.92-.1-1.36-.98 1.37-2.58 2.26-4.4 2.26-2.98 0-5.4-2.42-5.4-5.4 0-1.81.89-3.42 2.26-4.4-.44-.06-.9-.1-1.36-.1z"/></svg>
        </button>
        
        <div class="illustration">
            <svg viewBox="0 0 200 200" xmlns="http://www.w3.org/2000/svg">
                <!-- Main Gear -->
                <g class="gear" style="transform-origin: 100px 100px;">
                    <path d="M100 60 L105 70 L115 68 L118 78 L128 81 L125 91 L133 98 L125 105 L128 115 L118 118 L115 128 L105 125 L100 135 L95 125 L85 128 L82 118 L72 115 L75 105 L67 98 L75 91 L72 81 L82 78 L85 68 L95 70 Z" fill="var(--accent)"/>
                    <circle cx="100" cy="98" r="18" fill="var(--bg-secondary)"/>
                </g>
                
                <!-- Small Gear -->
                <g class="gear-small" style="transform-origin: 145px 145px;">
                    <path d="M145 125 L148 132 L155 130 L157 137 L164 139 L162 146 L168 151 L162 156 L164 163 L157 165 L155 172 L148 170 L145 177 L142 170 L135 172 L133 165 L126 163 L128 156 L122 151 L128 146 L126 139 L133 137 L135 130 L142 132 Z" fill="var(--accent-light)"/>
                    <circle cx="145" cy="151" r="10" fill="var(--bg-secondary)"/>
                </g>
                
                <!-- Wrench -->
                <g transform="translate(45, 130) rotate(-45)">
                    <rect x="0" y="8" width="35" height="8" rx="2" fill="var(--text-secondary)"/>
                    <circle cx="40" cy="12" r="12" fill="none" stroke="var(--text-secondary)" stroke-width="5"/>
                </g>
                
                <!-- Decorative dots -->
                <circle cx="50" cy="50" r="4" fill="var(--accent)" class="pulse"/>
                <circle cx="160" cy="60" r="3" fill="var(--accent-light)" class="pulse" style="animation-delay: 0.5s"/>
                <circle cx="40" cy="100" r="2" fill="var(--accent)" class="pulse" style="animation-delay: 1s"/>
            </svg>
        </div>
        
        <h1>Under Maintenance</h1>
        <div class="domain">%s</div>
        <p class="subtitle">
            This application is currently in <strong>maintenance mode</strong> or under <strong>development</strong>. 
            Our team is working hard to bring you an improved experience.
        </p>
        
        <div class="status-box">
            <h3>
                <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
                </svg>
                System Status: Configuring
            </h3>
            <p>The virtual host is being set up by the administrator</p>
        </div>
        
        <div class="info-cards">
            <div class="info-card">
                <h4>🔧 Development</h4>
                <p>Application is being developed or configured</p>
            </div>
            <div class="info-card">
                <h4>⏰ Check Back Soon</h4>
                <p>Please try again later</p>
            </div>
        </div>
        
        <div class="footer">
            <p>Protected by DoCode WAF • Web Application Firewall</p>
        </div>
    </div>
    
    <script>
        function toggleTheme() {
            const html = document.documentElement;
            const currentTheme = html.getAttribute('data-theme');
            const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
            html.setAttribute('data-theme', newTheme);
            localStorage.setItem('theme', newTheme);
        }
        
        // Load saved theme
        const savedTheme = localStorage.getItem('theme') || 
            (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
        document.documentElement.setAttribute('data-theme', savedTheme);
    </script>
</body>
</html>`, domain, domain)
}
