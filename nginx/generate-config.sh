#!/bin/sh

# Wait for database to be ready
echo "Waiting for database..."
sleep 5

# PostgreSQL connection details from environment
export PGHOST=${DATABASE_HOST:-postgres}
export PGPORT=${DATABASE_PORT:-5432}
export PGDATABASE=${DATABASE_NAME:-docode_waf}
export PGUSER=${DATABASE_USER:-waf_user}
export PGPASSWORD=${DATABASE_PASSWORD}

# Directory for vhost configurations
VHOSTS_DIR="/etc/nginx/vhosts.d"
mkdir -p $VHOSTS_DIR

# Function to generate nginx config from database
generate_configs() {
    echo "Generating nginx configurations from database..."
    
    # Clear existing configs
    rm -f $VHOSTS_DIR/*.conf
    
    # Query vhosts from database and generate configs
    psql -t -A -F"|" -c "SELECT id, name, domain, backend_url, ssl_enabled, ssl_cert_path, ssl_key_path, enabled FROM vhosts WHERE enabled = true" | while IFS='|' read -r id name domain backend_url ssl_enabled ssl_cert_path ssl_key_path enabled; do
        
        if [ -z "$domain" ]; then
            continue
        fi
        
        CONF_FILE="$VHOSTS_DIR/${domain}.conf"
        
        echo "# Virtual Host: $name" > $CONF_FILE
        echo "# Generated: $(date)" >> $CONF_FILE
        echo "" >> $CONF_FILE
        
        # HTTP Server Block
        cat >> $CONF_FILE <<EOF
server {
    listen 80;
    server_name $domain;

    # Logging
    access_log /var/log/nginx/${domain}_access.log;
    error_log /var/log/nginx/${domain}_error.log;

    # Client settings
    client_max_body_size 100M;
    client_body_timeout 300s;

    # Proxy to WAF backend
    location / {
        proxy_pass $backend_url;
        proxy_http_version 1.1;
        
        # Headers
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header X-Forwarded-Host \$host;
        proxy_set_header X-Forwarded-Port \$server_port;
        
        # WebSocket support
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # Timeouts
        proxy_connect_timeout 300s;
        proxy_send_timeout 300s;
        proxy_read_timeout 300s;
        
        # Buffering
        proxy_buffering off;
        proxy_request_buffering off;
    }
}
EOF

        # SSL Server Block (if enabled)
        if [ "$ssl_enabled" = "t" ] && [ -n "$ssl_cert_path" ] && [ -n "$ssl_key_path" ]; then
            cat >> $CONF_FILE <<EOF

# HTTPS Server Block
server {
    listen 443 ssl http2;
    server_name $domain;

    # SSL Configuration
    ssl_certificate $ssl_cert_path;
    ssl_certificate_key $ssl_key_path;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Logging
    access_log /var/log/nginx/${domain}_ssl_access.log;
    error_log /var/log/nginx/${domain}_ssl_error.log;

    # Client settings
    client_max_body_size 100M;
    client_body_timeout 300s;

    # Proxy to WAF backend
    location / {
        proxy_pass $backend_url;
        proxy_http_version 1.1;
        
        # Headers
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header X-Forwarded-Host \$host;
        proxy_set_header X-Forwarded-Port \$server_port;
        
        # WebSocket support
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # Timeouts
        proxy_connect_timeout 300s;
        proxy_send_timeout 300s;
        proxy_read_timeout 300s;
        
        # Buffering
        proxy_buffering off;
        proxy_request_buffering off;
    }
}
EOF
        fi
        
        echo "Generated config for: $domain"
    done
    
    echo "Configuration generation completed!"
}

# Generate configs on startup
generate_configs

# Test nginx configuration
nginx -t

if [ $? -eq 0 ]; then
    echo "Nginx configuration is valid"
    # Start nginx in foreground
    exec nginx -g 'daemon off;'
else
    echo "Nginx configuration test failed!"
    exit 1
fi
