#!/usr/bin/env bash

# Attack Simulator untuk WAF Dashboard
# Generate berbagai jenis attack traffic dari berbagai negara

TARGET="https://aleh.lab"
USER_AGENT_NORMAL="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
USER_AGENT_BOT="python-requests/2.28.0"

# IP addresses and countries - using indexed arrays instead of associative
COUNTRIES=("China" "Russia" "Brazil" "India" "Vietnam" "Indonesia" "Iran" "Turkey" "Ukraine" "Mexico" "Thailand" "Nigeria" "Pakistan" "Philippines" "Egypt")
IPS=("220.181.38.148" "5.188.210.45" "177.72.80.12" "103.21.124.78" "115.79.24.89" "182.253.128.45" "5.160.247.32" "88.247.135.67" "91.200.12.74" "187.189.76.123" "171.100.200.45" "197.210.85.67" "39.42.98.156" "112.198.68.92" "156.192.105.89")
    ["Mexico"]="187.189.76.123"
    ["Thailand"]="171.100.200.45"
    ["Nigeria"]="197.210.85.67"
    ["Pakistan"]="39.42.98.156"
    ["Philippines"]="112.198.68.92"
    ["Egypt"]="156.192.105.89"
)

echo "🔥 Starting Attack Simulation to $TARGET"
echo "🌍 Simulating attacks from 15 different countries"
echo "========================================="

# Helper function to get random country and IP
get_random_country_ip() {
    local idx=$((RANDOM % ${#COUNTRIES[@]}))
    echo "${COUNTRIES[$idx]}|${IPS[$idx]}"
}

# 1. SQL Injection Attempts from Various Countries
echo ""
echo "1️⃣  Launching SQL Injection attacks from multiple countries..."
for i in {1..15}; do
    IFS='|' read -r country ip <<< "$(get_random_country_ip)"
    
    curl -k -s -o /dev/null -w "SQL Injection from $country ($ip) #$i: %{http_code}\n" \
        -H "User-Agent: $USER_AGENT_NORMAL" \
        -H "X-Forwarded-For: $ip" \
        -H "X-Real-IP: $ip" \
        "$TARGET/?id=1' OR '1'='1" &
    
    curl -k -s -o /dev/null -w "SQL Injection from $country ($ip) #$i: %{http_code}\n" \
        -H "User-Agent: $USER_AGENT_NORMAL" \
        -H "X-Forwarded-For: $ip" \
        -H "X-Real-IP: $ip" \
        "$TARGET/?user=admin'--" &
    
    if [ $((i % 5)) -eq 0 ]; then
        wait
        sleep 0.2
    fi
done
wait
sleep 1

# 2. XSS Attempts from Various Countries
echo ""
echo "2️⃣  Launching XSS attacks from multiple countries..."
counter=0
for country in "${!COUNTRY_IPS[@]}"; do
    ip="${COUNTRY_IPS[$country]}"
    for i in {1..2}; do
        curl -k -s -o /dev/null -w "XSS from $country ($ip) #$i: %{http_code}\n" \
            -H "User-Agent: $USER_AGENT_NORMAL" \
            -H "X-Forwarded-For: $ip" \
            -H "X-Real-IP: $ip" \
            "$TARGET/?search=<script>alert('XSS')</script>" &
        
        curl -k -s -o /dev/null -w "XSS from $country ($ip) #$i: %{http_code}\n" \
            -H "User-Agent: $USER_AGENT_NORMAL" \
            -H "X-Forwarded-For: $ip" \
            -H "X-Real-IP: $ip" \
            "$TARGET/?q=<img src=x onerror=alert(1)>" &
        
        ((counter++))
        if [ $((counter % 10)) -eq 0 ]; then
            wait
            sleep 0.2
        fi
    done
done
wait
sleep 1

# 3. Path Traversal from Various Countries
echo ""
echo "3️⃣  Launching Path Traversal attacks from multiple countries..."
counter=0
for country in "${!COUNTRY_IPS[@]}"; do
    ip="${COUNTRY_IPS[$country]}"
    curl -k -s -o /dev/null -w "Path Traversal from $country ($ip): %{http_code}\n" \
        -H "User-Agent: $USER_AGENT_NORMAL" \
        -H "X-Forwarded-For: $ip" \
        -H "X-Real-IP: $ip" \
        "$TARGET/../../../../etc/passwd" &
    
    curl -k -s -o /dev/null -w "Path Traversal from $country ($ip): %{http_code}\n" \
        -H "User-Agent: $USER_AGENT_NORMAL" \
        -H "X-Forwarded-For: $ip" \
        -H "X-Real-IP: $ip" \
        "$TARGET/../../windows/system32/config/sam" &
    
    ((counter++))
    if [ $((counter % 5)) -eq 0 ]; then
        wait
        sleep 0.2
    fi
done
wait
sleep 1

# 4. Bot Traffic from Various Countries (will be blocked)
echo ""
echo "4️⃣  Launching Bot traffic from multiple countries..."
counter=0
for country in "${!COUNTRY_IPS[@]}"; do
    ip="${COUNTRY_IPS[$country]}"
    for i in {1..2}; do
        curl -k -s -o /dev/null -w "Bot from $country ($ip) #$i: %{http_code}\n" \
            -H "User-Agent: $USER_AGENT_BOT" \
            -H "X-Forwarded-For: $ip" \
            -H "X-Real-IP: $ip" \
            "$TARGET/" &
        
        curl -k -s -o /dev/null -w "Bot from $country ($ip) #$i: %{http_code}\n" \
            -H "User-Agent: Googlebot/2.1" \
            -H "X-Forwarded-For: $ip" \
            -H "X-Real-IP: $ip" \
            "$TARGET/" &
        
        ((counter++))
        if [ $((counter % 10)) -eq 0 ]; then
            wait
            sleep 0.2
        fi
    done
done
wait
sleep 1

# 5. HTTP Flood from Various Countries - Rate Limit Trigger
echo ""
echo "5️⃣  Launching HTTP Flood from multiple countries (Rate Limit)..."
counter=0
for country in "${!COUNTRY_IPS[@]}"; do
    ip="${COUNTRY_IPS[$country]}"
    for i in {1..8}; do
        curl -k -s -o /dev/null -w "Flood from $country ($ip) #$i: %{http_code}\n" \
            -H "User-Agent: $USER_AGENT_NORMAL" \
            -H "X-Forwarded-For: $ip" \
            -H "X-Real-IP: $ip" \
            "$TARGET/" &
        
        ((counter++))
        # Small delay to not overwhelm system
        if [ $((counter % 15)) -eq 0 ]; then
            wait
            sleep 0.1
        fi
    done
done
wait
sleep 1

# 6. Admin Panel Scanning from Various Countries
echo ""
echo "6️⃣  Scanning for Admin Panels from multiple countries..."
admin_paths=(
    "/admin"
    "/administrator"
    "/wp-admin"
    "/phpmyadmin"
    "/cpanel"
    "/admin.php"
    "/login"
    "/admin/login"
    "/dashboard"
    "/backend"
)
counter=0
for country in "${!COUNTRY_IPS[@]}"; do
    ip="${COUNTRY_IPS[$country]}"
    path="${admin_paths[$((RANDOM % ${#admin_paths[@]}))]}"
    curl -k -s -o /dev/null -w "Admin Scan from $country ($ip) $path: %{http_code}\n" \
        -H "User-Agent: $USER_AGENT_NORMAL" \
        -H "X-Forwarded-For: $ip" \
        -H "X-Real-IP: $ip" \
        "$TARGET$path" &
    
    ((counter++))
    if [ $((counter % 5)) -eq 0 ]; then
        wait
        sleep 0.2
    fi
done
wait
sleep 1

# 7. Suspicious Headers from Various Countries
echo ""
echo "7️⃣  Sending requests with suspicious headers from multiple countries..."
counter=0
for country in "${!COUNTRY_IPS[@]}"; do
    ip="${COUNTRY_IPS[$country]}"
    curl -k -s -o /dev/null -w "Suspicious Header from $country ($ip): %{http_code}\n" \
        -H "User-Agent: $USER_AGENT_NORMAL" \
        -H "X-Forwarded-For: $ip" \
        -H "X-Real-IP: $ip" \
        -H "X-Original-URL: /admin" \
        "$TARGET/" &
    
    curl -k -s -o /dev/null -w "Suspicious Header from $country ($ip): %{http_code}\n" \
        -H "User-Agent: $USER_AGENT_NORMAL" \
        -H "X-Forwarded-For: $ip" \
        -H "X-Real-IP: $ip" \
        -H "Referer: http://malicious-site.com" \
        "$TARGET/" &
    
    ((counter++))
    if [ $((counter % 5)) -eq 0 ]; then
        wait
        sleep 0.2
    fi
done
wait
sleep 1

# 8. Large POST Requests from Various Countries
echo ""
echo "8️⃣  Sending large POST requests from multiple countries..."
counter=0
for country in "${!COUNTRY_IPS[@]}"; do
    ip="${COUNTRY_IPS[$country]}"
    dd if=/dev/urandom bs=1024 count=50 2>/dev/null | \
    curl -k -s -o /dev/null -w "Large POST from $country ($ip): %{http_code}\n" \
        -H "User-Agent: $USER_AGENT_NORMAL" \
        -H "X-Forwarded-For: $ip" \
        -H "X-Real-IP: $ip" \
        -X POST \
        --data-binary @- \
        "$TARGET/api/upload" &
    
    ((counter++))
    if [ $((counter % 3)) -eq 0 ]; then
        wait
        sleep 0.3
    fi
done
wait
sleep 1

# 9. Command Injection from Various Countries
echo ""
echo "9️⃣  Launching Command Injection attacks from multiple countries..."
counter=0
for country in "${!COUNTRY_IPS[@]}"; do
    ip="${COUNTRY_IPS[$country]}"
    curl -k -s -o /dev/null -w "Command Injection from $country ($ip): %{http_code}\n" \
        -H "User-Agent: $USER_AGENT_NORMAL" \
        -H "X-Forwarded-For: $ip" \
        -H "X-Real-IP: $ip" \
        "$TARGET/?cmd=ls;cat /etc/passwd" &
    
    curl -k -s -o /dev/null -w "Command Injection from $country ($ip): %{http_code}\n" \
        -H "User-Agent: $USER_AGENT_NORMAL" \
        -H "X-Forwarded-For: $ip" \
        -H "X-Real-IP: $ip" \
        "$TARGET/?exec=whoami" &
    
    ((counter++))
    if [ $((counter % 5)) -eq 0 ]; then
        wait
        sleep 0.2
    fi
done
wait
sleep 1

# 10. Normal Traffic (untuk comparison)
echo ""
echo "🟢 Sending normal traffic for comparison..."
for i in {1..30}; do
    curl -k -s -o /dev/null -w "Normal Request #$i: %{http_code}\n" \
        -H "User-Agent: $USER_AGENT_NORMAL" \
        "$TARGET/" &
    
    if [ $((i % 5)) -eq 0 ]; then
        sleep 0.5
    fi
done
wait

echo ""
echo "========================================="
echo "✅ Attack Simulation Completed!"
echo ""
echo "📊 Attack Summary:"
echo "  - 15 countries simulated: China, Russia, Brazil, India, Vietnam,"
echo "    Indonesia, Iran, Turkey, Ukraine, Mexico, Thailand, Nigeria,"
echo "    Pakistan, Philippines, Egypt"
echo "  - 9 attack types executed from each country"
echo "  - Total requests: ~450+ malicious requests"
echo ""
echo "Check dashboard at: http://localhost:3000"
echo "You should see:"
echo "  - Total requests increased significantly"
echo "  - Blocked requests from various countries"
echo "  - Attack patterns: SQL injection, XSS, Path Traversal, etc."
echo "  - Geographic distribution of attacks"
echo "  - Rate limit violations"
echo "  - Bot traffic blocked"
echo ""
