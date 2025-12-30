# Build stage
FROM node:22-alpine3.20 AS builder

# Install security updates
RUN apk upgrade --no-cache

WORKDIR /app

COPY package*.json ./
RUN npm install --no-audit --no-fund

COPY . .
RUN npm run build

# Production stage - Using nginx stable alpine
FROM nginx:1.27-alpine3.20

# Install security updates and remove unnecessary packages
RUN apk upgrade --no-cache && \
    apk add --no-cache ca-certificates tzdata && \
    rm -rf /var/cache/apk/*

# Remove default nginx configs
RUN rm -f /etc/nginx/conf.d/default.conf

# Copy built files and config
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Set proper permissions
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chmod -R 755 /usr/share/nginx/html

EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://127.0.0.1/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
