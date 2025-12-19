# Docode WAF - Environment Configuration Guide

## Setup

1. **Copy the example environment file:**
   ```bash
   cp .env.example .env
   ```

2. **Edit `.env` file with your credentials:**
   ```bash
   nano .env
   # or
   vim .env
   ```

## Configuration Categories

### Server Configuration
Configure the WAF server and admin API ports:
```env
SERVER_HOST=0.0.0.0
SERVER_PORT=8080          # Port untuk WAF proxy
SERVER_ADMIN_PORT=9090    # Port untuk Admin API
```

### Database Configuration
PostgreSQL database credentials:
```env
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=docode_waf
DATABASE_USER=waf_user
DATABASE_PASSWORD=your_secure_password_here
DATABASE_SSLMODE=disable
```

**Security Note:** Ganti `DATABASE_PASSWORD` dengan password yang kuat!

### Redis Configuration
Redis untuk caching dan rate limiting:
```env
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password_here
REDIS_DB=0
```

### WAF Features Configuration

#### Rate Limiting
```env
WAF_RATE_LIMIT_ENABLED=true
WAF_RATE_LIMIT_RPS=100      # Requests per second
WAF_RATE_LIMIT_BURST=200    # Burst capacity
```

#### HTTP Flood Protection
```env
WAF_HTTP_FLOOD_ENABLED=true
WAF_HTTP_FLOOD_MAX_RPM=1000     # Max requests per minute
WAF_HTTP_FLOOD_BLOCK_DURATION=300  # Block duration in seconds
```

#### Anti-Bot Protection
```env
WAF_ANTI_BOT_ENABLED=true
WAF_ANTI_BOT_CHALLENGE_MODE=js  # js, captcha, cookie
```

#### GeoIP (Optional)
```env
WAF_GEOIP_ENABLED=false
WAF_GEOIP_DATABASE_PATH=./geoip/GeoLite2-Country.mmdb
```

### SSL Configuration
```env
SSL_AUTO_CERT=false
SSL_CERT_DIR=./certs
```

### Logging Configuration
```env
LOG_LEVEL=info          # debug, info, warn, error
LOG_FORMAT=json         # json, text
LOG_OUTPUT=stdout       # stdout, file
LOG_FILE_PATH=./logs/waf.log
```

## Running the Application

### Development Mode
```bash
# Load environment variables from .env
go run cmd/waf/main.go
```

### Production Mode
```bash
# Make sure .env is configured
docker-compose up -d
```

### Verify Configuration
```bash
# Check if services are running
docker-compose ps

# Check logs
docker-compose logs waf
```

## Security Best Practices

1. **Never commit `.env` file to version control**
   - `.env` is already in `.gitignore`
   - Only commit `.env.example`

2. **Use strong passwords**
   - Database password: minimum 16 characters
   - Redis password: minimum 16 characters

3. **Change default credentials**
   - Replace all default passwords in production
   - Use different credentials for each environment

4. **Restrict access**
   - Use firewall to restrict access to database ports
   - Only expose necessary ports (8080, 9090, 3000)

5. **Enable SSL in production**
   - Set `SSL_AUTO_CERT=true`
   - Configure SSL certificates for each vhost

## Environment-Specific Configuration

### Development
```env
SERVER_HOST=localhost
LOG_LEVEL=debug
LOG_FORMAT=text
```

### Staging
```env
SERVER_HOST=0.0.0.0
LOG_LEVEL=info
LOG_FORMAT=json
```

### Production
```env
SERVER_HOST=0.0.0.0
LOG_LEVEL=warn
LOG_FORMAT=json
LOG_OUTPUT=file
SSL_AUTO_CERT=true
```

## Troubleshooting

### Cannot connect to database
- Check `DATABASE_HOST` and `DATABASE_PORT`
- Verify database credentials
- Ensure PostgreSQL is running

### Redis connection failed
- Check `REDIS_HOST` and `REDIS_PORT`
- Verify Redis password if set
- Ensure Redis is running

### Port already in use
- Change `SERVER_PORT` or `SERVER_ADMIN_PORT`
- Check for conflicting services: `lsof -i :8080`

## Configuration Priority

The application loads configuration in this order:
1. Default values from `config.yaml`
2. Override with values from `config.local.yaml` (if exists)
3. Override with environment variables from `.env` file
4. Override with system environment variables

Environment variables have the highest priority!
