# DCode WAF - Web Application Firewall

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.25.5-00ADD8?logo=go)
![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)
[![Support Me](https://img.shields.io/badge/Support%20Me-Saweria-orange?logo=buy-me-a-coffee&logoColor=white)](https://saweria.co/abdialeh)

A powerful and modern Web Application Firewall (WAF) built with **Golang** and **React** with real-time monitoring, SSL management, and advanced security features.

## ☕ Support This Project

If you find this project helpful, consider supporting me:

[![Buy Me a Coffee](https://img.shields.io/badge/Buy%20Me%20a%20Coffee-Saweria-FF813F?style=for-the-badge&logo=buy-me-a-coffee&logoColor=white)](https://saweria.co/abdialeh)

Your support helps maintain and improve this project!

## 🔒 Security & Quality

This project is committed to maintaining high security and code quality standards:

- **🔍 SonarQube Analysis** - Source code analyzed for code quality, security vulnerabilities, and code smells
- **🛡️ Trivy Security Scanning** - All Docker images scanned for vulnerabilities (CRITICAL & HIGH severity)
- **✅ Zero Critical Issues** - Custom built images (waf, frontend, nginx-proxy) are clean
- **📊 Continuous Monitoring** - Regular scans to ensure security compliance

### Security Scan Results:
- **Backend (Go 1.25.5)**: ✅ Clean - No vulnerabilities
- **Frontend (React + Nginx)**: ✅ Clean - No vulnerabilities  
- **Nginx Proxy**: ✅ Clean - No vulnerabilities
- **PostgreSQL 15**: ✅ Patched - Minimal non-critical issues
- **Redis 7.4**: ✅ Updated - Latest stable version

## 🚀 Features

### Core Security
- ✅ **Reverse Proxy** - High-performance reverse proxy with nginx integration
- ✅ **Rate Limiting** - Configurable rate limiting per IP/endpoint
- ✅ **IP Blocking** - Block by single IP or CIDR blocks
- ✅ **Geographic Blocking** - Block requests by country (GeoIP2)
- ✅ **URL Filtering** - Pattern-based URL blocking
- ✅ **SSL Certificate Management** - Upload and manage SSL certificates per domain
- ✅ **HTTP Flood Protection** - Protect against DDoS and HTTP flood attacks
- ✅ **Anti-Bot Detection** - Intelligent bot detection and mitigation
- ✅ **Cloudflare Turnstile** - CAPTCHA protection for login & registration pages
- ✅ **Application Branding** - Custom app name and logo configuration

### Attack Detection
- 🔍 **SQL Injection** - Pattern matching for SQL injection attempts
- 🔍 **XSS (Cross-Site Scripting)** - Detect XSS attack patterns
- 🔍 **Path Traversal** - Block directory traversal attempts
- 🔍 **Command Injection** - Detect OS command injection
- 🔍 **Admin Scanning** - Identify admin panel scanning
- 🔍 **Bot Traffic** - User-agent based bot detection

### Management
- 🌐 **Virtual Host Management** - Configure custom vhosts with SSL
- 📊 **Real-time Dashboard** - Monitor traffic, attacks, and statistics
- 🗺️ **GeoIP Integration** - Country detection with flag display
- 📈 **Custom Date Range** - Filter dashboard data by custom date ranges
- 🔐 **JWT Authentication** - Secure API with JWT tokens
- 👤 **User Registration** - Self-service account registration
- 🔒 **Role-Based Access** - Admin and superadmin roles
- ⚙️ **Advanced VHost Config** - WebSocket, HTTP/2, TLS version, custom headers
- 🎨 **Application Settings** - Configure app name, logo, and branding
- 📝 **Nginx Config Editor** - Edit vhost configs with syntax highlighting (Monokai theme)
- 🔄 **Auto-Backup** - Automatic backup before config changes
- 📍 **Custom Locations** - Define custom location blocks per vhost
- 🏷️ **Custom Headers** - Add custom HTTP headers per vhost
- 🗑️ **Auto Cleanup** - Delete log files when vhost is removed

### UI/UX
- ⚡ **Modern React UI** - Built with React 18 + Vite
- 🎨 **Tailwind CSS** - Beautiful, responsive design
- 📊 **Recharts** - Interactive charts and visualizations
- 🔄 **Real-time Updates** - Live data updates with loading indicators
- 🌍 **Country Flags** - Visual country identification
- 🔍 **Search & Filter** - Search attacks with pagination
- 💻 **Code Editor** - CodeMirror with Nginx syntax highlighting
- 🌙 **Monokai Theme** - Sublime Text style code editor theme
- 📋 **Copy to Clipboard** - Quick copy config content
- ↩️ **Reset Changes** - Revert to original config before save

---

## 📋 Architecture

```
docode-waf/
├── cmd/
│   └── waf/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/                     # API handlers
│   │   ├── auth.go             # Authentication endpoints
│   │   ├── dashboard.go        # Dashboard statistics API
│   │   ├── vhost.go            # Virtual host management
│   │   └── certificate.go      # SSL certificate management
│   ├── middleware/
│   │   ├── auth.go             # JWT authentication middleware
│   │   ├── logging.go          # Traffic logging + attack detection
│   │   ├── ratelimit.go        # Rate limiting middleware
│   │   └── bot_detector.go     # Bot detection middleware
│   ├── models/                  # Database models
│   └── services/
│       └── nginx_config.go     # Nginx configuration generator
├── frontend/
│   ├── src/
│   │   ├── components/         # Reusable React components
│   │   ├── pages/              # Page components
│   │   │   ├── Dashboard.jsx   # Main dashboard with charts
│   │   │   ├── VirtualHosts.jsx
│   │   │   └── Login.jsx
│   │   ├── services/
│   │   │   └── api.js          # API client
│   │   └── App.jsx
│   └── package.json
├── migrations/                  # Database migrations
├── docker-compose.yaml          # Docker orchestration
├── Dockerfile                   # WAF backend container
├── GeoLite2-Country.mmdb       # GeoIP database
└── config.yaml                  # Configuration file
```

---

## 🛠️ Technology Stack

### Backend
- **Language**: Go 1.25.5
- **Framework**: Gin (HTTP framework)
- **Database**: PostgreSQL 15
- **Cache**: Redis 7.4
- **Authentication**: JWT
- **Proxy**: Nginx
- **GeoIP**: MaxMind GeoLite2

### Frontend
- **Framework**: React 18
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **Charts**: Recharts
- **Icons**: Lucide React
- **HTTP Client**: Axios
- **Date Handling**: date-fns
- **CAPTCHA**: Cloudflare Turnstile
- **Code Editor**: CodeMirror (@uiw/react-codemirror)
- **Syntax Highlighting**: Custom Nginx language mode

### DevOps
- **Containerization**: Docker & Docker Compose
- **Database Migration**: sqlx
- **Volume Sharing**: Nginx configs & SSL certificates

---

## 📦 Prerequisites

- **Docker** & **Docker Compose**
- **Go 1.25+** (for local development)
- **Node.js 18+** (for frontend development)
- **PostgreSQL 15+**
- **Redis 7.4+**

---

## 🚀 Quick Start

### 1. Clone Repository

```bash
git clone <repository-url>
cd docode-waf
```

### 2. Configure Environment

Create `.env` file:

```bash
# Database Configuration
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_NAME=docode_waf
DATABASE_USER=waf_user
DATABASE_PASSWORD=waf_secure_password_123

# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# CORS Configuration
# Comma-separated list of allowed origins
# Use "*" for development, specific domains for production
CORS_ALLOW_ORIGIN=http://localhost:3000

# Cloudflare Turnstile (Optional - for CAPTCHA on login/register)
TURNSTILE_SITE_KEY=1x00000000000000000000AA
TURNSTILE_SECRET_KEY=1x0000000000000000000000000000000AA

# Application Configuration
APP_PORT=8080
ADMIN_API_PORT=9090
```

### 3. Start with Docker Compose

```bash
# Build and start all services
docker compose up -d

# Check logs
docker compose logs -f

# Stop services
docker compose down
```

**First Time Setup**: Database migrations run automatically on first startup, including:
- Schema creation (`init.sql`)
- VHost advanced fields
- App settings table
- Attack detection fields
- **Default admin account** (username: `admin`, password: `Admin123!`)

### 4. Access Application

- **Frontend**: http://localhost:3000
- **WAF API**: http://localhost:9090
- **Proxy**: http://localhost:80 / https://localhost:443

---

## 🔐 Default Credentials

### Default Admin Account

After first deployment with Docker Compose, you can login with:

- **URL**: http://localhost:3000/login
- **Username**: `admin`
- **Password**: `Admin123!`
- **Email**: admin@docode.local
- **Role**: superadmin

> ⚠️ **Security**: Change default credentials immediately after first login for production deployments!

---

## 📊 Dashboard Features

### Statistics Cards
- **Total Requests** - All incoming requests
- **Blocked Requests** - Requests blocked by WAF
- **Total Attacks** - Detected attack attempts
- **Unique IPs** - Number of unique client IPs

### Traffic Over Time Chart
- Line chart showing requests per hour
- Separate lines for total requests and blocked requests
- Custom date range support

### Top Attack Types
- Bar chart of attack types distribution
- Shows: SQL Injection, XSS, Path Traversal, Bot Traffic, Command Injection, Admin Scan

### Recent Attacks Table
- **Time** - Timestamp of attack
- **IP Address** - Client IP address
- **Country** - Country flag 🇺🇸 + country name (via GeoIP)
- **Attack Type** - Type of attack detected
- **URL** - Attacked endpoint
- **Status** - Allowed/Blocked status
- **Search** - Search by IP, attack type, or URL
- **Pagination** - 10 attacks per page with navigation

### Date Range Filter
- **Predefined Ranges**: 1H, 24H, 7D, 30D
- **Custom Range**: Pick start and end dates
- **Loading Indicator**: Shows "Fetching data..." during updates
- **Overlay Loader**: Full-screen loader for better UX

---

## 🌐 Virtual Host Configuration

### Create New Virtual Host

1. Navigate to **Virtual Hosts** page
2. Click **"Add Virtual Host"**
3. Fill in details:
   - **Name**: Friendly name (e.g., "My API Server")
   - **Domain**: example.com
   - **Backend URL**: http://backend:9101
   - **Enable SSL**: Toggle on
   - **SSL Certificate**: Select from dropdown
   - **Advanced Settings**:
     - Enable WebSocket support
     - HTTP Version (1.1 or 2)
     - TLS Version (1.2 or 1.3)
     - Max upload size (MB)
     - Proxy timeouts
     - Custom Headers (JSON object)
     - Custom Location Blocks (path, proxy_pass, custom config)
4. Click **"Save"**

### Edit Nginx Config

1. Navigate to **Virtual Hosts** page
2. Click the **File Code icon** (📄) on any vhost
3. Edit config with:
   - **Syntax Highlighting**: Nginx keywords, strings, comments
   - **Monokai Theme**: Sublime Text style colors
   - **Line Numbers**: Easy navigation
   - **Auto-Save Backup**: Creates `.backup` file before changes
4. Actions:
   - **Save**: Apply changes and reload nginx
   - **Reset**: Revert to original content
   - **Copy**: Copy entire config to clipboard
   - **Back**: Return to vhosts list

### Custom Location Blocks

Add custom nginx location blocks for advanced routing:

```nginx
# Example: Rate limit specific endpoint
location /api/ {
    limit_req zone=api burst=50 nodelay;
    proxy_pass http://backend:8080;
}

# Example: Static file serving
location /static/ {
    alias /var/www/static/;
    expires 30d;
}

# Example: WebSocket endpoint
location /ws/ {
    proxy_pass http://backend:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
}
```

**Note**: Default API rate limiting (`location ~* ^/api/`) is automatically skipped if you define a custom `/api/` location to avoid duplicates.

### How It Works

1. **Certificate Storage**: Certificates saved to PostgreSQL database
2. **File Export**: On startup, WAF exports certificates to `/ssl_certificates/`
3. **Nginx Config**: Auto-generates nginx config in `/nginx_configs/`
4. **Volume Sharing**: Nginx container mounts shared volumes
5. **Reload**: Nginx automatically uses new configs

### Nginx Config Structure

```nginx
server {
    listen 443 ssl;
    server_name example.com;
    
    ssl_certificate /etc/nginx/ssl/certificates/cert_<ID>.pem;
    ssl_certificate_key /etc/nginx/ssl/certificates/key_<ID>.pem;
    
    location / {
        proxy_pass http://waf:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

---

## 🔍 Attack Detection Patterns

### SQL Injection
```
' OR '1'='1
' OR 1=1--
UNION SELECT
'; DROP TABLE
admin'--
```

### XSS (Cross-Site Scripting)
```
<script>alert(1)</script>
javascript:alert(1)
<img src=x onerror=alert(1)>
```

### Path Traversal
```
../../../etc/passwd
..\..\windows\system32
%2e%2e%2f (URL encoded)
```

### Command Injection
```
; ls -la
| cat /etc/passwd
`whoami`
$(id)
```

### Admin Scanning
```
/admin
/administrator
/wp-admin
/phpmyadmin
```

### Bot Detection
```
User-Agent: curl/7.x
User-Agent: python-requests
User-Agent: Wget/1.x
User-Agent: *bot*
```

---

## 🛡️ Security Configuration

### Rate Limiting

Configure in middleware:

```go
// 100 requests per minute per IP
rateLimit := middleware.NewRateLimiter(redis, 100, time.Minute)
router.Use(rateLimit.Limit())
```

### IP Blocking

```go
// Block single IP
blocked := []string{"192.168.1.100"}

// Block CIDR range
blocked := []string{"10.0.0.0/8", "172.16.0.0/12"}
```

### Geographic Blocking

```go
// Block by country code
blockedCountries := []string{"CN", "RU", "KP"}
```

---

## 📊 Database Schema

### Main Tables

- **users** - User authentication and roles
- **vhosts** - Virtual host configurations
  - Supports WebSocket, HTTP/2, TLS 1.2/1.3
  - Custom headers (JSONB) and location blocks
  - Max upload size and proxy timeouts
  - SSL certificate reference
- **vhost_locations** - Custom location blocks per vhost
  - Path, proxy_pass, custom_config
  - Enabled flag for easy toggle
- **certificates** - SSL certificates
- **traffic_logs** - HTTP traffic logs with attack detection
  - `is_attack` - Boolean flag for detected attacks
  - `attack_type` - Type of attack (SQL Injection, XSS, etc.)
  - `country_code` - ISO country code from GeoIP
  - `host` - Domain/host of the request
  - `blocked` - Whether request was blocked
- **app_settings** - Application branding and configuration
  - `app_name` - Custom application name
  - `app_logo` - Base64 encoded logo image
- **ip_groups** - IP whitelist/blacklist management
- **rules** - Custom WAF rules

---

## 🔧 Development

### Backend Development

```bash
# Install dependencies
go mod download

# Run locally
go run cmd/waf/main.go

# Build binary
go build -o waf cmd/waf/main.go

# Run tests
go test ./...
```

### Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Start dev server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

### Database Migration

```bash
# Create new migration
migrate create -ext sql -dir migrations -seq migration_name

# Run migrations
migrate -path migrations -database "postgresql://user:pass@localhost:5432/dbname" up

# Rollback migration
migrate -path migrations -database "postgresql://user:pass@localhost:5432/dbname" down
```

---

## 📝 API Endpoints

### Authentication
```
POST   /api/v1/auth/login       # Login (with optional turnstile_token)
POST   /api/v1/auth/register    # Register new user (with optional turnstile_token)
GET    /api/v1/auth/profile     # Get user profile
GET    /api/v1/turnstile/sitekey # Get Turnstile site key (public endpoint)
```

### Dashboard
```
GET    /api/v1/dashboard/stats?start=2025-12-21&end=2025-12-23
GET    /api/v1/dashboard/stats?range=24h
GET    /api/v1/dashboard/traffic
```

### Virtual Hosts
```
GET    /api/v1/vhosts              # List all vhosts
POST   /api/v1/vhosts              # Create new vhost
GET    /api/v1/vhosts/:id          # Get vhost details
PUT    /api/v1/vhosts/:id          # Update vhost
DELETE /api/v1/vhosts/:id          # Delete vhost (+ delete log files)
GET    /api/v1/vhost-config/:domain # Get nginx config content
PUT    /api/v1/vhost-config/:domain # Update nginx config (auto-backup)
```

### Application Settings
```
GET    /api/v1/settings/app     # Get app settings (name & logo)
POST   /api/v1/settings/app     # Update app settings
```

### SSL Certificates
```
GET    /api/v1/certificates     # List certificates
POST   /api/v1/certificates     # Upload certificate
DELETE /api/v1/certificates/:id # Delete certificate
```

---

## 🐛 Troubleshooting

### WAF Container Fails to Start

```bash
# Check logs
docker compose logs waf

# Common issues:
# - Database not ready: Wait for PostgreSQL to fully start
# - Port conflict: Check if port 8080/9090 already in use
# - Config error: Verify config.yaml syntax
```

### Dashboard Shows No Data

```bash
# Check API connectivity
curl http://localhost:9090/api/v1/dashboard/stats

# Verify authentication
# Open browser DevTools → Network tab → Check for 401 errors

# Check database
docker exec -it docode-waf-postgres-1 psql -U waf_user -d docode_waf -c "SELECT COUNT(*) FROM traffic_logs;"
```

### SSL Certificate Not Working

```bash
# Check certificate export
docker exec docode-waf-waf-1 ls -la /root/ssl_certificates/

# Check nginx config generation
docker exec docode-waf-waf-1 ls -la /root/nginx_configs/

# Verify nginx can read certificates
docker exec docode-waf-nginx-proxy-1 ls -la /etc/nginx/ssl/certificates/

# Test SSL
curl -k https://yourdomain.com
```

### GeoIP Not Working

```bash
# Verify GeoIP database exists
docker exec docode-waf-waf-1 ls -lh /root/GeoLite2-Country.mmdb

# Check logs for GeoIP initialization
docker logs docode-waf-waf-1 | grep -i geoip

# Expected output: "GeoIP database loaded successfully"
```

### Cloudflare Turnstile Not Showing

```bash
# Check if environment variables are set
docker exec docode-waf-waf-1 env | grep TURNSTILE

# Verify site key is configured
curl http://localhost:9090/api/v1/turnstile/sitekey

# Check browser console for JavaScript errors
# Open DevTools → Console → Look for Turnstile script loading errors

# Test with Cloudflare test keys:
# TURNSTILE_SITE_KEY=1x00000000000000000000AA
# TURNSTILE_SECRET_KEY=1x0000000000000000000000000000000AA
```

---

## 📈 Performance Optimization

### Database Indexing

```sql
-- Add indexes for better query performance
CREATE INDEX idx_traffic_logs_timestamp ON traffic_logs(timestamp);
CREATE INDEX idx_traffic_logs_client_ip ON traffic_logs(client_ip);
CREATE INDEX idx_traffic_logs_is_attack ON traffic_logs(is_attack);
CREATE INDEX idx_traffic_logs_attack_type ON traffic_logs(attack_type);
```

### Redis Caching

- Rate limit counters cached in Redis
- JWT blacklist stored in Redis
- Session management via Redis

### Nginx Optimization

```nginx
# Enable gzip compression
gzip on;
gzip_types text/plain text/css application/json application/javascript;

# Enable HTTP/2
listen 443 ssl http2;

# Connection pooling
keepalive_timeout 65;
keepalive_requests 100;
```

---

## 🔒 Production Deployment

### Security Checklist

- [ ] Change default admin password
- [ ] Use strong JWT secret (min 32 characters)
- [ ] Enable HTTPS only (redirect HTTP → HTTPS)
- [ ] Use environment variables for secrets
- [ ] Enable PostgreSQL SSL connection
- [ ] Set up firewall rules
- [ ] Enable rate limiting
- [ ] Configure log rotation
- [ ] Set up monitoring and alerts
- [ ] Regular security updates

### Environment Variables

```bash
# Use strong secrets
JWT_SECRET=$(openssl rand -base64 32)
DATABASE_PASSWORD=$(openssl rand -base64 24)

# Production database
DATABASE_HOST=prod-db.example.com
DATABASE_SSL_MODE=require

# Redis with password
REDIS_PASSWORD=$(openssl rand -base64 16)

# CORS settings
ALLOWED_ORIGINS=https://waf.example.com

# Or multiple origins (comma-separated)
CORS_ALLOW_ORIGIN=https://waf.example.com,https://admin.example.com
```

### Nginx Config File Locations

- **Config Files**: `/data/nginx/config/{domain}.conf`
- **Backup Files**: `/data/nginx/config/{domain}.conf.backup`
- **Log Files**: `/data/nginx/logs/{domain}_access.log`, `/data/nginx/logs/{domain}_error.log`
- **SSL Certificates**: `/data/nginx/ssl/{certificate_id}/cert.pem` and `key.pem`

### Best Practices

- **Custom Locations**: Use custom location blocks for fine-grained control
- **Rate Limiting**: Default API location (`/api/`) has rate limiting (200 req/burst)
- **Auto Cleanup**: Log files are automatically deleted when vhost is removed
- **Backup Safety**: Config editor creates `.backup` file before every save
- **CORS Security**: Use specific origins in production, avoid `*` wildcard

---

## Link Referensi

### Documentation & Tools
- **Go (Golang)**: [https://go.dev/](https://go.dev/) - Official Go documentation
- **Gin Framework**: [https://gin-gonic.com/](https://gin-gonic.com/) - HTTP web framework
- **React**: [https://react.dev/](https://react.dev/) - Official React documentation
- **Vite**: [https://vitejs.dev/](https://vitejs.dev/) - Frontend build tool
- **Tailwind CSS**: [https://tailwindcss.com/](https://tailwindcss.com/) - Utility-first CSS framework
- **PostgreSQL**: [https://www.postgresql.org/docs/](https://www.postgresql.org/docs/) - Database documentation
- **Redis**: [https://redis.io/docs/](https://redis.io/docs/) - In-memory data store
- **Nginx**: [https://nginx.org/en/docs/](https://nginx.org/en/docs/) - Reverse proxy documentation
- **Docker**: [https://docs.docker.com/](https://docs.docker.com/) - Container platform

### Security Tools
- **SonarQube**: [https://www.sonarsource.com/products/sonarqube/](https://www.sonarsource.com/products/sonarqube/) - Code quality analysis
- **Trivy**: [https://trivy.dev/](https://trivy.dev/) - Container security scanner
- **MaxMind GeoIP**: [https://dev.maxmind.com/geoip](https://dev.maxmind.com/geoip) - Geolocation database
- **Cloudflare Turnstile**: [https://developers.cloudflare.com/turnstile/](https://developers.cloudflare.com/turnstile/) - CAPTCHA alternative
- **OWASP**: [https://owasp.org/](https://owasp.org/) - Web Application Security Project

### Libraries & Packages
- **Recharts**: [https://recharts.org/](https://recharts.org/) - React charting library
- **Lucide React**: [https://lucide.dev/](https://lucide.dev/) - Icon library
- **CodeMirror**: [https://codemirror.net/](https://codemirror.net/) - Code editor component
- **Axios**: [https://axios-http.com/](https://axios-http.com/) - HTTP client
- **JWT**: [https://jwt.io/](https://jwt.io/) - JSON Web Tokens

---

## 📄 License

MIT License - See LICENSE file for details

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

## 📧 Support

For issues and questions, please open an issue on GitHub.

---

**Built with ❤️ using Go and React**
