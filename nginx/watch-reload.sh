#!/bin/sh

# Watch for nginx config changes and reload automatically
SIGNAL_FILE="/data/nginx/.reload"
LAST_MTIME=""

echo "Starting nginx reload watcher..."

# Wait for nginx to start and create pid file
sleep 5

while true; do
    if [ -f "$SIGNAL_FILE" ]; then
        CURRENT_MTIME=$(stat -c %Y "$SIGNAL_FILE" 2>/dev/null || stat -f %m "$SIGNAL_FILE" 2>/dev/null)
        
        if [ "$CURRENT_MTIME" != "$LAST_MTIME" ]; then
            echo "[$(date)] Config change detected, reloading nginx..."
            
            # Check if nginx is running before reload
            if [ -f /var/run/nginx.pid ] && kill -0 $(cat /var/run/nginx.pid) 2>/dev/null; then
                nginx -s reload
                echo "[$(date)] Nginx reloaded successfully"
            else
                echo "[$(date)] Nginx not running yet, skipping reload"
            fi
            
            LAST_MTIME="$CURRENT_MTIME"
        fi
    fi
    
    sleep 2
done
