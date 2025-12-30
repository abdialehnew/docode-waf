#!/usr/bin/env bash

# Simple Attack Simulator - Compatible with macOS bash
# Generates various attack traffic from different countries

TARGET="https://aleh.lab"
USER_AGENT_NORMAL="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
USER_AGENT_BOT="python-requests/2.28.0"

# Countries and IPs
COUNTRIES=("China" "Russia" "Brazil" "India" "Vietnam" "Indonesia" "Iran" "Turkey" "Ukraine" "Mexico" "Thailand" "Nigeria" "Pakistan" "Philippines" "Egypt")
IPS=("220.181.38.148" "5.188.210.45" "177.72.80.12" "103.21.124.78" "115.79.24.89" "182.253.128.45" "5.160.247.32" "88.247.135.67" "91.200.12.74" "187.189.76.123" "171.100.200.45" "197.210.85.67" "39.42.98.156" "112.198.68.92" "156.192.105.89")

echo "🔥 Attack Simulator - Generating malicious traffic"
echo "🎯 Target: $TARGET"
echo "🌍 Simulating attacks from ${#COUNTRIES[@]} countries"
echo "================================================"
echo ""

# Helper function
get_random_ip() {
    local idx=$((RANDOM % ${#IPS[@]}))
    echo "${IPS[$idx]}"
}

get_random_country() {
    local idx=$((RANDOM % ${#COUNTRIES[@]}))
    echo "${COUNTRIES[$idx]}"
}

# 1. SQL Injection
echo "1️⃣  SQL Injection attacks (30 requests)..."
for i in $(seq 1 30); do
    ip=$(get_random_ip)
    country=$(get_random_country)
    
    curl -k -s -o /dev/null -w "$country ($ip): %{http_code}\n" \
        -H "User-Agent: $USER_AGENT_NORMAL" \
        -H "X-Forwarded-For: $ip" \
        "$TARGET/?id=1' OR '1'='1" &
    
    if [ $((i % 10)) -eq 0 ]; then
        wait
        sleep 0.3
    fi
done
wait

# 2. XSS Attacks  
echo ""
echo "2️⃣  XSS attacks (20 requests)..."
for i in $(seq 1 20); do
    ip=$(get_random_ip)
    country=$(get_random_country)
    
    curl -k -s -o /dev/null -w "$country ($ip): %{http_code}\n" \
        -H "User-Agent: $USER_AGENT_NORMAL" \
        -H "X-Forwarded-For: $ip" \
        "$TARGET/?search=<script>alert('xss')</script>" &
    
    if [ $((i % 10)) -eq 0 ]; then
        wait
        sleep 0.3
    fi
done
wait

# 3. Path Traversal
echo ""
echo "3️⃣  Path Traversal attacks (20 requests)..."
for i in $(seq 1 20); do
    ip=$(get_random_ip)
    country=$(get_random_country)
    
    curl -k -s -o /dev/null -w "$country ($ip): %{http_code}\n" \
        -H "User-Agent: $USER_AGENT_NORMAL" \
        -H "X-Forwarded-For: $ip" \
        "$TARGET/../../../../etc/passwd" &
    
    if [ $((i % 10)) -eq 0 ]; then
        wait
        sleep 0.3
    fi
done
wait

# 4. Bot Traffic
echo ""
echo "4️⃣  Bot traffic (20 requests)..."
for i in $(seq 1 20); do
    ip=$(get_random_ip)
    country=$(get_random_country)
    
    curl -k -s -o /dev/null -w "$country ($ip): %{http_code}\n" \
        -H "User-Agent: $USER_AGENT_BOT" \
        -H "X-Forwarded-For: $ip" \
        "$TARGET/" &
    
    if [ $((i % 10)) -eq 0 ]; then
        wait
        sleep 0.3
    fi
done
wait

# 5. HTTP Flood
echo ""
echo "5️⃣  HTTP Flood (50 rapid requests)..."
for i in $(seq 1 50); do
    ip=$(get_random_ip)
    
    curl -k -s -o /dev/null -w "%{http_code} " \
        -H "X-Forwarded-For: $ip" \
        "$TARGET/" &
    
    if [ $((i % 10)) -eq 0 ]; then
        echo ""
        wait
        sleep 0.1
    fi
done
wait
echo ""

# 6. Admin Panel Scanning
echo ""
echo "6️⃣  Admin panel scanning (15 requests)..."
ADMIN_PATHS=("admin" "wp-admin" "administrator" "cpanel" "phpmyadmin" "admin.php" "login.php" "admin/login.php")
for path in "${ADMIN_PATHS[@]}"; do
    ip=$(get_random_ip)
    country=$(get_random_country)
    
    curl -k -s -o /dev/null -w "$country ($ip) /$path: %{http_code}\n" \
        -H "User-Agent: $USER_AGENT_NORMAL" \
        -H "X-Forwarded-For: $ip" \
        "$TARGET/$path" &
done
wait

# 7. Suspicious Headers
echo ""
echo "7️⃣  Suspicious headers (20 requests)..."
for i in $(seq 1 20); do
    ip=$(get_random_ip)
    country=$(get_random_country)
    
    curl -k -s -o /dev/null -w "$country ($ip): %{http_code}\n" \
        -H "User-Agent: $USER_AGENT_NORMAL" \
        -H "X-Forwarded-For: $ip" \
        -H "X-Original-URL: /admin" \
        -H "X-Rewrite-URL: /../../../etc/passwd" \
        "$TARGET/" &
    
    if [ $((i % 10)) -eq 0 ]; then
        wait
        sleep 0.3
    fi
done
wait

# 8. Large POST requests
echo ""
echo "8️⃣  Large POST requests (10 requests)..."
LARGE_DATA=$(python3 -c "print('A' * 1000000)")
for i in $(seq 1 10); do
    ip=$(get_random_ip)
    country=$(get_random_country)
    
    curl -k -s -o /dev/null -w "$country ($ip): %{http_code}\n" \
        -X POST \
        -H "User-Agent: $USER_AGENT_NORMAL" \
        -H "X-Forwarded-For: $ip" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "$LARGE_DATA" \
        "$TARGET/upload" &
    
    if [ $((i % 5)) -eq 0 ]; then
        wait
        sleep 0.5
    fi
done
wait

# 9. Command Injection
echo ""
echo "9️⃣  Command injection attempts (20 requests)..."
for i in $(seq 1 20); do
    ip=$(get_random_ip)
    country=$(get_random_country)
    
    curl -k -s -o /dev/null -w "$country ($ip): %{http_code}\n" \
        -H "User-Agent: $USER_AGENT_NORMAL" \
        -H "X-Forwarded-For: $ip" \
        "$TARGET/?cmd=ls;cat%20/etc/passwd" &
    
    if [ $((i % 10)) -eq 0 ]; then
        wait
        sleep 0.3
    fi
done
wait

echo ""
echo "✅ Attack simulation completed!"
echo "📊 Check your WAF dashboard for results"
