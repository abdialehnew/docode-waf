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

# Setup GeoIP2 if database exists
mkdir -p /etc/nginx/conf.d/geoip2
if [ -f /etc/nginx/geoip/GeoLite2-Country.mmdb ]; then
    echo "GeoIP2 database found, enabling region filtering..."
    cat <<EOF > /etc/nginx/conf.d/geoip2/geoip2.conf
geoip2 /etc/nginx/geoip/GeoLite2-Country.mmdb {
    auto_reload 15m;
    \$geoip2_data_country_code default=XX source=\$remote_addr country iso_code;
}
EOF
else
    echo "GeoIP2 database not found, region filtering will be disabled for now"
    rm -f /etc/nginx/conf.d/geoip2/geoip2.conf
fi

# Start reload watcher in background
/usr/local/bin/watch-reload.sh &

# Start nginx
exec nginx -g 'daemon off;'
