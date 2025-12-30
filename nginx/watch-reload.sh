#!/bin/sh

# Watch for nginx config changes and reload automatically
SIGNAL_FILE="/data/nginx/.reload"
LAST_MTIME=""

echo "Starting nginx reload watcher..."

while true; do
    if [ -f "$SIGNAL_FILE" ]; then
        CURRENT_MTIME=$(stat -c %Y "$SIGNAL_FILE" 2>/dev/null || stat -f %m "$SIGNAL_FILE" 2>/dev/null)
        
        if [ "$CURRENT_MTIME" != "$LAST_MTIME" ]; then
            echo "[$(date)] Config change detected, reloading nginx..."
            nginx -s reload
            LAST_MTIME="$CURRENT_MTIME"
            echo "[$(date)] Nginx reloaded successfully"
        fi
    fi
    
    sleep 2
done
