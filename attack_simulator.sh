#!/usr/bin/env bash

# Attack Simulator for WAF Dashboard
# Simulates various attack vectors to test WAF detection, logging, blocking, and notifications.

TARGET="${1:-http://localhost:8080}"
USER_AGENT_NORMAL="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
USER_AGENT_BOT="python-requests/2.28.0"

# Countries and their simulated IPs
COUNTRIES=("China" "Russia" "Brazil" "India" "Vietnam" "Indonesia" "Iran" "Turkey" "Ukraine" "Mexico" "Thailand" "Nigeria" "Pakistan" "Philippines" "Egypt")
IPS=("220.181.38.148" "5.188.210.45" "177.72.80.12" "103.21.124.78" "115.79.24.89" "182.253.128.45" "5.160.247.32" "88.247.135.67" "91.200.12.74" "187.189.76.123" "171.100.200.45" "197.210.85.67" "39.42.98.156" "112.198.68.92" "156.192.105.89")

echo "🔥 Starting Attack Simulation against: $TARGET"
echo "🌍 Simulating attacks from ${#COUNTRIES[@]} different countries"
echo "========================================="

# Helper: Execute curl with spoofed headers
# Usage: attack "Type" "Country" "IP" "URL" "User-Agent" [Extra Headers...]
attack() {
    local type=$1
    local country=$2
    local ip=$3
    local url=$4
    local ua=$5
    shift 5
    local extra_args=("$@")

    # Run in background to simulate concurrency
    curl -k -s -o /dev/null -w "$type from $country ($ip): %{http_code}\n" \
        -H "User-Agent: $ua" \
        -H "X-Forwarded-For: $ip" \
        -H "X-Real-IP: $ip" \
        "${extra_args[@]}" \
        "$url" &
}

# 1. SQL Injection (Should trigger Notification)
echo ""
echo "1️⃣  Launching SQL Injection attacks (High Severity)..."
for i in "${!COUNTRIES[@]}"; do
    country="${COUNTRIES[$i]}"
    ip="${IPS[$i]}"
    
    attack "SQL Injection" "$country" "$ip" "$TARGET/?id=1' OR '1'='1" "$USER_AGENT_NORMAL"
    
    if [ $((i % 5)) -eq 0 ]; then wait; fi
done
wait

# 2. XSS Attacks
echo ""
echo "2️⃣  Launching XSS attacks..."
for i in "${!COUNTRIES[@]}"; do
    country="${COUNTRIES[$i]}"
    ip="${IPS[$i]}"
    
    attack "XSS" "$country" "$ip" "$TARGET/?search=<script>alert('XSS')</script>" "$USER_AGENT_NORMAL"
    
    if [ $((i % 5)) -eq 0 ]; then wait; fi
done
wait

# 3. Path Traversal
echo ""
echo "3️⃣  Launching Path Traversal attacks..."
for i in "${!COUNTRIES[@]}"; do
    country="${COUNTRIES[$i]}"
    ip="${IPS[$i]}"
    
    attack "Path Traversal" "$country" "$ip" "$TARGET/../../../../etc/passwd" "$USER_AGENT_NORMAL"
done
wait

# 4. Bot Traffic (User-Agent Detection)
echo ""
echo "4️⃣  Launching Bad Bot traffic..."
for i in "${!COUNTRIES[@]}"; do
    country="${COUNTRIES[$i]}"
    ip="${IPS[$i]}"
    
    attack "Bad Bot" "$country" "$ip" "$TARGET/" "$USER_AGENT_BOT"
done
wait

# 5. Admin Panel Scanning (Should trigger Notification)
echo ""
echo "5️⃣  Scanning for Admin Panels..."
admin_paths=("/admin" "/phpmyadmin" "/login" "/wp-admin")
for i in "${!COUNTRIES[@]}"; do
    country="${COUNTRIES[$i]}"
    ip="${IPS[$i]}"
    path="${admin_paths[$((RANDOM % ${#admin_paths[@]}))]}"
    
    attack "Admin Scan" "$country" "$ip" "$TARGET$path" "$USER_AGENT_NORMAL"
done
wait

# 6. Command Injection
echo ""
echo "6️⃣  Launching Command Injection attacks..."
for i in "${!COUNTRIES[@]}"; do
    country="${COUNTRIES[$i]}"
    ip="${IPS[$i]}"
    
    attack "Command Injection" "$country" "$ip" "$TARGET/?cmd=cat+/etc/passwd" "$USER_AGENT_NORMAL"
done
wait

# 7. Dynamic Ban Test (Fail2Ban Simulation)
# We will use a unique IP and hit the server 15 times rapidly.
TEST_BAN_IP="10.10.10.10"
echo ""
echo "7️⃣  Testing Dynamic IP Ban (IP: $TEST_BAN_IP)..."
echo "   Sending 15 malicious requests from $TEST_BAN_IP. Expecting 403 Forbidden after ~10th request."

for i in {1..15}; do
    curl -k -s -o /dev/null -w "Request #$i ($TEST_BAN_IP): %{http_code}\n" \
        -H "User-Agent: $USER_AGENT_NORMAL" \
        -H "X-Forwarded-For: $TEST_BAN_IP" \
        -H "X-Real-IP: $TEST_BAN_IP" \
        "$TARGET/?id=1' OR '1'='1"
done

# Check if banned
echo "   Verifying ban status for $TEST_BAN_IP..."
HTTP_CODE=$(curl -k -s -o /dev/null -w "%{http_code}" \
    -H "User-Agent: $USER_AGENT_NORMAL" \
    -H "X-Forwarded-For: $TEST_BAN_IP" \
    -H "X-Real-IP: $TEST_BAN_IP" \
    "$TARGET/")

if [ "$HTTP_CODE" == "403" ]; then
    echo "   ✅ SUCCESS: IP $TEST_BAN_IP is BANNED (HTTP 403)."
else
    echo "   ❌ FAILURE: IP $TEST_BAN_IP is NOT BANNED (HTTP $HTTP_CODE)."
fi

echo ""
echo "========================================="
echo "✅ Attack Simulation Completed!"
echo "➡️ Check Dashboard at http://localhost:3000"
echo ""
