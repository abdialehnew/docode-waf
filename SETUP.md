# Docode WAF - Setup Guide

## Prerequisites

- Go 1.21 or higher
- Node.js 18+ and npm
- PostgreSQL 13+
- Redis 6+
- Docker and Docker Compose (optional)

## Quick Start with Docker

The easiest way to get started is using Docker Compose:

```bash
# Copy environment file
cp .env.example .env

# Edit .env with your credentials
nano .env

# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f

# Stop services
docker-compose down
```

The services will be available at:
- WAF Proxy: http://localhost:8080
- Admin API: http://localhost:9090
- Dashboard: http://localhost:3000

## Manual Setup

### 1. Environment Configuration

```bash
# Copy environment file
cp .env.example .env

# Edit with your credentials
nano .env
```

### 2. Database Setup

```bash
# Create database
createdb docode_waf

# Run the backend (will load .env automatically)tup

```bash
# Install Go dependencies
go mod download

# Copy and configure
cp config.yaml config.local.yaml
# Edit config.local.yaml with your settings

# Run the backend
go run cmd/waf/main.go
```

### 3. Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev

# Or build for production
npm run build
```

## Configuration

Edit `config.yaml` or `config.local.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080          # WAF proxy port
  admin_port: 9090    # Admin API port

database:
  host: "localhost"
  port: 5432
  name: "docode_waf"
  user: "waf_user"
  password: "waf_password"

redis:
  host: "localhost"
  port: 6379

waf:
  rate_limit:
    enabled: true
    requests_per_second: 100
    burst: 200
    
  http_flood:
    enabled: true
    max_requests_per_minute: 1000
    block_duration: 300
    
  anti_bot:
    enabled: true
    challenge_mode: "js"
```

## Usage

### Adding a Virtual Host

1. Open the dashboard at http://localhost:3000
2. Go to "Virtual Hosts"
3. Click "Add Virtual Host"
4. Fill in the details:
   - Name: My App
   - Domain: myapp.example.com
   - Backend URL: http://localhost:3001
5. Click "Create"

### Managing IP Groups

1. Go to "IP Groups"
2. Click "Add IP Group"
3. Configure the group (blacklist/whitelist)
4. Add individual IPs or CIDR blocks

### Monitoring

The dashboard provides:
- Real-time traffic statistics
- Attack detection and blocking
- Traffic logs
- Performance metrics
- Geographic distribution

## API Endpoints

### Dashboard
- `GET /api/v1/dashboard/stats?range=24h` - Get statistics
- `GET /api/v1/dashboard/traffic?limit=100` - Get traffic logs

### Virtual Hosts
- `GET /api/v1/vhosts` - List all vhosts
- `POST /api/v1/vhosts` - Create vhost
- `PUT /api/v1/vhosts/:id` - Update vhost
- `DELETE /api/v1/vhosts/:id` - Delete vhost

### IP Groups
- `GET /api/v1/ip-groups` - List all IP groups
- `POST /api/v1/ip-groups` - Create IP group
- `POST /api/v1/ip-groups/:id/ips` - Add IP to group
- `DELETE /api/v1/ip-groups/:id/ips/:ipId` - Remove IP from group

## Features

### Reverse Proxy
The WAF acts as a reverse proxy, forwarding requests to backend servers based on virtual host configuration.

### Rate Limiting
Configurable rate limiting per IP address to prevent abuse.

### IP Blocking
- Blacklist: Block specific IPs or CIDR ranges
- Whitelist: Allow only specific IPs

### HTTP Flood Protection
Detects and blocks HTTP flood attacks automatically.

### Bot Detection
Identifies and blocks malicious bots based on:
- User-Agent patterns
- Missing HTTP headers
- Suspicious request patterns

### SSL/TLS Management
Support for SSL certificates per virtual host.

### Real-time Monitoring
- Traffic statistics
- Attack detection logs
- Performance metrics
- Geographic analysis

## Development

### Running Tests

```bash
# Backend tests
go test ./...

# Frontend tests
cd frontend
npm test
```

### Building for Production

```bash
# Build backend
make build

# Build frontend
cd frontend
npm run build
```

## Troubleshooting

### Database Connection Issues
- Ensure PostgreSQL is running
- Check database credentials in config.yaml
- Verify database exists and migrations are applied

### Frontend Not Loading
- Check if backend API is running on port 9090
- Verify CORS settings
- Check browser console for errors

### Rate Limiting Not Working
- Ensure Redis is running
- Check WAF configuration
- Verify rate_limit.enabled is true

## Security Considerations

1. **Change Default Passwords**: Update all default passwords in production
2. **Use SSL/TLS**: Enable SSL for all virtual hosts
3. **Regular Updates**: Keep dependencies updated
4. **Monitor Logs**: Regularly review attack logs
5. **Backup Database**: Schedule regular database backups

## Performance Tuning

1. **Database**: Add indexes for frequently queried columns
2. **Redis**: Increase pool size for high traffic
3. **Go Runtime**: Adjust GOMAXPROCS for multi-core systems
4. **Rate Limits**: Tune based on expected traffic

## License

MIT License
