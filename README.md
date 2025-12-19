# Docode WAF - Web Application Firewall

A powerful and flexible Web Application Firewall built with Golang and React.

## Features

- **Reverse Proxy**: High-performance reverse proxy with custom routing
- **Rate Limiting**: Configurable rate limiting per IP/endpoint
- **IP/Region/URL Blocking**: Block requests based on IP addresses, geographic regions, or URLs
- **SSL Certificate Management**: Automatic SSL certificate management
- **HTTP Flood Protection**: Protect against HTTP flood attacks
- **Anti-Bot Protection**: Intelligent bot detection and mitigation
- **IP Group Management**: Manage single IPs and IP blocks (CIDR)
- **Custom Virtual Hosts**: Configure custom vhosts and locations
- **Real-time Dashboard**: Monitor traffic and attacks in real-time
- **Modern UI**: React + Tailwind CSS interface

## Architecture

```
├── backend/          # Go backend
│   ├── api/         # API handlers
│   ├── core/        # WAF core logic
│   ├── models/      # Data models
│   └── services/    # Business logic
├── frontend/        # React frontend
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   └── services/
└── config.yaml      # Configuration file
```

## Prerequisites

- Go 1.21 or higher
- Node.js 18+ and npm
- PostgreSQL 13+
- Redis 6+

## Quick Start

### Backend Setup

1. Setup environment variables:
```bash
cp .env.example .env
# Edit .env with your database and Redis credentials
```

2. Install dependencies:
```bash
go mod download
```

3. Set up database:
```bash
psql -U postgres -c "CREATE DATABASE docode_waf;"
psql -U postgres -d docode_waf -f migrations/init.sql
```

4. Run the backend:
```bash
go run cmd/waf/main.go
```

### Frontend Setup

1. Navigate to frontend directory:
```bash
cd frontend
```

2. Install dependencies:
```bash
npm install
```

3. Start development server:
```bash
npm run dev
```

## API Documentation

The API documentation is available at `http://localhost:9090/api/docs` when the server is running.

## Configuration

The application uses environment variables for configuration. Copy `.env.example` to `.env` and customize:

```bash
cp .env.example .env
```

### Key Configuration Options:

**Server Configuration:**
- `SERVER_PORT` - WAF proxy port (default: 8080)
- `SERVER_ADMIN_PORT` - Admin API port (default: 9090)

**Database Configuration:**
- `DATABASE_HOST` - PostgreSQL host
- `DATABASE_PORT` - PostgreSQL port
- `DATABASE_NAME` - Database name
- `DATABASE_USER` - Database username
- `DATABASE_PASSWORD` - Database password

**WAF Features:**
- `WAF_RATE_LIMIT_ENABLED` - Enable/disable rate limiting
- `WAF_RATE_LIMIT_RPS` - Requests per second limit
- `WAF_HTTP_FLOOD_ENABLED` - Enable/disable HTTP flood protection
- `WAF_ANTI_BOT_ENABLED` - Enable/disable anti-bot protection

See [ENV_GUIDE.md](ENV_GUIDE.md) for complete configuration documentation.

## License

MIT License
