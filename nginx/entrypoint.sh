#!/bin/sh
set -e

# Generate self-signed certificate at runtime if not exists
if [ ! -f /etc/nginx/ssl/default.key ] || [ ! -f /etc/nginx/ssl/default.crt ]; then
    echo "Generating self-signed SSL certificate..."
    openssl req -x509 -nodes -days 365 -newkey rsa:4096 \
        -keyout /etc/nginx/ssl/default.key \
        -out /etc/nginx/ssl/default.crt \
        -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost" 2>/dev/null
    chmod 600 /etc/nginx/ssl/default.key
    chmod 644 /etc/nginx/ssl/default.crt
    echo "SSL certificate generated successfully"
fi

# Start reload watcher in background
/usr/local/bin/watch-reload.sh &

# Start nginx
exec nginx -g 'daemon off;'
