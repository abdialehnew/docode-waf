#!/usr/bin/env bash

# Historical Attack Data Generator
# Generates attack data for multiple dates (Dec 15-22, 2025)

DB_HOST="postgres"
DB_NAME="docode_waf"
DB_USER="waf_user"
DB_PASS="waf_password"

# Countries and IPs
COUNTRIES=("China" "Russia" "Brazil" "India" "Vietnam" "Indonesia" "Iran" "Turkey" "Ukraine" "Mexico" "Thailand" "Nigeria" "Pakistan" "Philippines" "Egypt")
IPS=("220.181.38.148" "5.188.210.45" "177.72.80.12" "103.21.124.78" "115.79.24.89" "182.253.128.45" "5.160.247.32" "88.247.135.67" "91.200.12.74" "187.189.76.123" "171.100.200.45" "197.210.85.67" "39.42.98.156" "112.198.68.92" "156.192.105.89")
COUNTRY_CODES=("CN" "RU" "BR" "IN" "VN" "ID" "IR" "TR" "UA" "MX" "TH" "NG" "PK" "PH" "EG")

# Attack types
ATTACK_TYPES=("SQL Injection" "XSS" "Path Traversal" "Bot" "HTTP Flood" "Admin Scan" "Suspicious Headers" "Large POST" "Command Injection")

# URLs for attacks
URLS=(
    "/?id=1' OR '1'='1"
    "/?user=admin'--"
    "/?search=<script>alert('xss')</script>"
    "/?name=<img src=x onerror=alert(1)>"
    "/../../../../etc/passwd"
    "/../../../etc/shadow"
    "/admin"
    "/wp-admin"
    "/administrator"
    "/cpanel"
    "/phpmyadmin"
    "/admin.php"
    "/login.php"
    "/?cmd=ls;cat%20/etc/passwd"
    "/?exec=whoami"
    "/upload"
    "/"
)

echo "🔥 Historical Attack Data Generator"
echo "📅 Generating data for December 15-22, 2025"
echo "🌍 Random countries and IPs"
echo "================================================"
echo ""

# Function to get random element from array
get_random() {
    local arr=("$@")
    local idx=$((RANDOM % ${#arr[@]}))
    echo "${arr[$idx]}"
}

# Function to get random country data
get_random_country_data() {
    local idx=$((RANDOM % ${#COUNTRIES[@]}))
    echo "${COUNTRIES[$idx]}|${IPS[$idx]}|${COUNTRY_CODES[$idx]}"
}

# Function to generate random timestamp for a specific date
get_random_timestamp() {
    local date=$1
    local hour=$((RANDOM % 24))
    local minute=$((RANDOM % 60))
    local second=$((RANDOM % 60))
    printf "%s %02d:%02d:%02d" "$date" "$hour" "$minute" "$second"
}

# Counter
total_records=0

# Loop through dates (Dec 15-22, 2025)
for day in {15..22}; do
    date="2025-12-${day}"
    echo "📆 Generating attacks for ${date}..."
    
    # Random number of attacks per day (50-150)
    num_attacks=$((50 + RANDOM % 101))
    
    for i in $(seq 1 $num_attacks); do
        # Get random country data
        IFS='|' read -r country ip country_code <<< "$(get_random_country_data)"
        
        # Get random attack type
        attack_type=$(get_random "${ATTACK_TYPES[@]}")
        
        # Get random URL
        url=$(get_random "${URLS[@]}")
        
        # Generate random timestamp for this date
        timestamp=$(get_random_timestamp "$date")
        
        # Random HTTP method
        methods=("GET" "POST" "PUT" "DELETE")
        method=$(get_random "${methods[@]}")
        
        # Random status code (mostly 403/404/502 for attacks)
        status_codes=(403 403 403 404 404 404 502 500 400)
        status_code=$(get_random "${status_codes[@]}")
        
        # Random response time (50-2000ms for attacks)
        response_time=$((50 + RANDOM % 1951))
        
        # Random bytes sent
        bytes_sent=$((100 + RANDOM % 5000))
        
        # Determine if blocked (80% chance for attacks)
        blocked="true"
        is_attack="true"
        block_reason="$attack_type detected"
        
        # 20% chance not blocked
        if [ $((RANDOM % 10)) -lt 2 ]; then
            blocked="false"
            block_reason=""
        fi
        
        # User agents
        user_agents=(
            "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
            "python-requests/2.28.0"
            "curl/7.68.0"
            "Nikto/2.1.6"
            "sqlmap/1.4.7"
            "Mozilla/5.0 (compatible; Baiduspider/2.0)"
        )
        user_agent=$(get_random "${user_agents[@]}")
        
        # Build SQL INSERT statement
        sql="INSERT INTO traffic_logs (
            timestamp, client_ip, method, url, status_code, response_time, 
            bytes_sent, user_agent, country_code, blocked, block_reason, 
            is_attack, attack_type
        ) VALUES (
            '$timestamp'::timestamp,
            '$ip',
            '$method',
            '$url',
            $status_code,
            $response_time,
            $bytes_sent,
            '$user_agent',
            '$country_code',
            $blocked,
            '$block_reason',
            $is_attack,
            '$attack_type'
        );"
        
        # Execute SQL via docker compose
        docker compose exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" -c "$sql" > /dev/null 2>&1
        
        ((total_records++))
        
        # Progress indicator
        if [ $((i % 20)) -eq 0 ]; then
            echo -n "."
        fi
    done
    
    echo " ✅ $num_attacks attacks generated for $date"
done

echo ""
echo "================================================"
echo "✅ Historical data generation completed!"
echo "📊 Total records generated: $total_records"
echo ""
echo "Run this to verify:"
echo "docker compose exec -T postgres psql -U waf_user -d docode_waf -c \"SELECT DATE(timestamp) as date, COUNT(*) as attacks FROM traffic_logs WHERE timestamp >= '2025-12-15' GROUP BY DATE(timestamp) ORDER BY date;\""
