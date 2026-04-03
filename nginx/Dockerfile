# Stage 1: Build Brotli modules
FROM nginx:1.27.3-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    gcc \
    libc-dev \
    make \
    openssl-dev \
    pcre-dev \
    zlib-dev \
    linux-headers \
    libtool \
    automake \
    autoconf \
    git \
    g++ \
    cmake \
    brotli-dev \
    libmaxminddb-dev

# Download Nginx source
# Note: Match the version of the base image
RUN wget http://nginx.org/download/nginx-1.27.3.tar.gz && \
    tar zxf nginx-1.27.3.tar.gz

# Download Brotli and GeoIP2 module source
RUN git clone --recursive https://github.com/google/ngx_brotli.git
RUN git clone --recursive https://github.com/leev/ngx_http_geoip2_module.git

# Compile the modules
WORKDIR /nginx-1.27.3
RUN ./configure --with-compat \
    --add-dynamic-module=../ngx_brotli \
    --add-dynamic-module=../ngx_http_geoip2_module && \
    make modules

# Stage 2: Final Image
FROM nginx:1.27.3-alpine

# Install runtime dependencies and security updates
RUN apk upgrade --no-cache && \
    apk add --no-cache \
    ca-certificates \
    tzdata \
    openssl \
    inotify-tools \
    brotli-libs \
    libmaxminddb && \
    # Create SSL and GeoIP directories
    mkdir -p /etc/nginx/ssl /etc/nginx/geoip && \
    chmod 755 /etc/nginx/ssl /etc/nginx/geoip && \
    # Clean up default config
    rm -f /etc/nginx/conf.d/default.conf && \
    mkdir -p /var/log/nginx && \
    chown -R nginx:nginx /var/log/nginx && \
    rm -rf /var/cache/apk/*

# Copy compiled modules from builder
# Alpine packages usually expect modules in /usr/lib/nginx/modules/
COPY --from=builder /nginx-1.27.3/objs/ngx_http_brotli_filter_module.so /usr/lib/nginx/modules/
COPY --from=builder /nginx-1.27.3/objs/ngx_http_brotli_static_module.so /usr/lib/nginx/modules/
COPY --from=builder /nginx-1.27.3/objs/ngx_http_geoip2_module.so /usr/lib/nginx/modules/

# Copy configuration and scripts
COPY html/ /usr/share/nginx/html/
COPY nginx.conf /etc/nginx/nginx.conf
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
COPY watch-reload.sh /usr/local/bin/watch-reload.sh

# Make scripts executable
RUN chmod +x /usr/local/bin/entrypoint.sh /usr/local/bin/watch-reload.sh

EXPOSE 80 443 443/udp

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider --server-response http://127.0.0.1/ 2>&1 | grep -q 'HTTP.*\(200\|404\|403\|502\|503\)' || exit 1

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
