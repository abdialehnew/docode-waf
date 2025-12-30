# SSL Certificates Directory

This directory is for storing SSL certificates that will be used by nginx reverse proxy.

## Structure

```
ssl/
├── domain.com.crt       # SSL Certificate
├── domain.com.key       # Private Key
├── subdomain.domain.com.crt
└── subdomain.domain.com.key
```

## Generating Self-Signed Certificates (for testing)

```bash
# Generate self-signed certificate for domain
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout ssl/yourdomain.com.key \
  -out ssl/yourdomain.com.crt \
  -subj "/C=ID/ST=Jakarta/L=Jakarta/O=Company/CN=yourdomain.com"
```

## Using Let's Encrypt Certificates

```bash
# Install certbot
sudo apt-get install certbot

# Generate certificate
sudo certbot certonly --standalone -d yourdomain.com

# Copy certificates
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem ssl/yourdomain.com.crt
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem ssl/yourdomain.com.key
```

## Important Notes

1. Certificate files must be readable by the nginx container
2. Update vhost configuration in the admin panel with certificate paths:
   - SSL Cert Path: `/etc/nginx/ssl/yourdomain.com.crt`
   - SSL Key Path: `/etc/nginx/ssl/yourdomain.com.key`
3. After adding/updating certificates, reload nginx: `docker-compose restart nginx`

