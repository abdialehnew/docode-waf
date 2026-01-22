FROM nginx:1.27-alpine3.20

# Install security updates and required tools
RUN apk upgrade --no-cache && \
    apk add --no-cache ca-certificates tzdata openssl inotify-tools && \
    rm -rf /var/cache/apk/*

# Create SSL directory with proper permissions
RUN mkdir -p /etc/nginx/ssl && \
    chmod 755 /etc/nginx/ssl

# Copy nginx config
COPY nginx.conf /etc/nginx/nginx.conf

# Copy entrypoint script
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# Copy reload watcher script
COPY watch-reload.sh /usr/local/bin/watch-reload.sh
RUN chmod +x /usr/local/bin/watch-reload.sh

# Remove default config and create log directory
RUN rm -f /etc/nginx/conf.d/default.conf && \
    mkdir -p /var/log/nginx && \
    chown -R nginx:nginx /var/log/nginx

EXPOSE 80 443

# Health check - Accept 200, 404, 403, 502, or 503 (upstream errors are okay for health)
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider --server-response http://127.0.0.1/ 2>&1 | grep -q 'HTTP.*\(200\|404\|403\|502\|503\)' || exit 1

# Start via entrypoint
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
