FROM nginx:1.27-alpine3.20

# Install security updates and required tools
RUN apk upgrade --no-cache && \
    apk add --no-cache ca-certificates tzdata openssl inotify-tools && \
    rm -rf /var/cache/apk/*

# Create SSL directory with proper permissions
RUN mkdir -p /etc/nginx/ssl && \
    chmod 755 /etc/nginx/ssl

# Generate self-signed certificate for default HTTPS with stronger encryption
RUN openssl req -x509 -nodes -days 365 -newkey rsa:4096 \
    -keyout /etc/nginx/ssl/default.key \
    -out /etc/nginx/ssl/default.crt \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost" && \
    chmod 600 /etc/nginx/ssl/default.key && \
    chmod 644 /etc/nginx/ssl/default.crt

# Copy nginx config
COPY nginx.conf /etc/nginx/nginx.conf

# Copy reload watcher script
COPY watch-reload.sh /usr/local/bin/watch-reload.sh
RUN chmod +x /usr/local/bin/watch-reload.sh

# Remove default config and create log directory
RUN rm -f /etc/nginx/conf.d/default.conf && \
    mkdir -p /var/log/nginx && \
    chown -R nginx:nginx /var/log/nginx

EXPOSE 80 443

# Health check - Accept 200, 404, 403, or 502 (502 is when upstream blocks the bot health check)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider --server-response http://127.0.0.1/ 2>&1 | grep -q 'HTTP.*\(200\|404\|403\|502\)' || exit 1

# Start nginx and reload watcher in background
CMD /usr/local/bin/watch-reload.sh & nginx -g 'daemon off;'
