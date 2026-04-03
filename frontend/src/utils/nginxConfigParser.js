/**
 * Nginx Config Parser
 * 
 * Parses an Nginx vhost configuration string and extracts
 * form-compatible field values using regex patterns.
 */

/**
 * Parse an Nginx configuration string into form data fields.
 * Fields that cannot be determined remain undefined (caller should
 * merge with defaults).
 *
 * @param {string} config - Raw Nginx config text
 * @returns {object} Parsed form data fields
 */
export function parseNginxConfig(config) {
    if (!config || typeof config !== 'string') return {}

    const result = {}

    // --- server_name → domain ---
    // Match server_name directive, may have multiple domains
    const serverNameMatch = config.match(/server_name\s+([^;]+);/g)
    if (serverNameMatch) {
        // Take the last server_name (usually the SSL server block)
        const last = serverNameMatch[serverNameMatch.length - 1]
        const domains = last
            .replace(/server_name\s+/, '')
            .replace(';', '')
            .trim()
            .split(/\s+/)
            .filter(d => d && d !== '_')
        if (domains.length > 0) {
            result.domain = domains
        }
    }

    // --- Detect SSL ---
    if (/listen\s+443\s+ssl/i.test(config)) {
        result.ssl_enabled = true
    } else {
        result.ssl_enabled = false
    }

    // --- Detect VHost type from location / block ---
    // Extract the main location / block content
    const locationRootBlock = extractLocationBlock(config, '/')
    if (locationRootBlock) {
        if (/return\s+404\s*;/.test(locationRootBlock)) {
            result.type = 'dead'
        } else if (/return\s+301\s+(.+);/.test(locationRootBlock)) {
            result.type = 'redirect'
            const redirectMatch = locationRootBlock.match(/return\s+301\s+([^$;]+)/)
            if (redirectMatch) {
                result.backend_url = redirectMatch[1].trim().replace(/\$request_uri$/, '')
            }
        } else {
            result.type = 'proxy'
            // Extract proxy_pass from location /
            const proxyPassMatch = locationRootBlock.match(/proxy_pass\s+([^;]+);/)
            if (proxyPassMatch) {
                const proxyVal = proxyPassMatch[1].trim()
                // Skip internal waf proxy references
                if (!proxyVal.includes('waf:8080') && !proxyVal.includes('_backend')) {
                    result.backend_url = proxyVal
                }
            }
        }

        // --- WebSocket support ---
        if (/proxy_set_header\s+Upgrade\s+\$http_upgrade/i.test(locationRootBlock)) {
            result.websocket_enabled = true
        }
    }

    // --- Detect upstream backends ---
    // Parse all upstream blocks into a map keyed by name
    const upstreamMap = {}
    const upstreamRegex = /upstream\s+(\w+)_backend\s*\{([^}]+)\}/gs
    let upMatch
    while ((upMatch = upstreamRegex.exec(config)) !== null) {
        const name = upMatch[1]
        const body = upMatch[2]
        const servers = []
        const serverMatches = body.matchAll(/server\s+([^\s;]+)/g)
        for (const m of serverMatches) {
            servers.push(m[1])
        }
        let lbMethod = 'round_robin'
        if (/least_conn/i.test(body)) lbMethod = 'least_conn'
        else if (/ip_hash/i.test(body)) lbMethod = 'ip_hash'
        upstreamMap[name] = { servers, lbMethod }
    }

    // Find the main upstream (the one without _loc_ in the name)
    const mainUpstreamKey = Object.keys(upstreamMap).find(k => !k.includes('_loc_'))
    if (mainUpstreamKey && upstreamMap[mainUpstreamKey].servers.length > 0) {
        const main = upstreamMap[mainUpstreamKey]
        result.backend_url = main.servers[0]
        if (main.servers.length > 1) {
            result.backends = main.servers.slice(1)
        }
        result.load_balance_method = main.lbMethod
    }

    // --- client_max_body_size → max_upload_size ---
    const maxBodyMatch = config.match(/client_max_body_size\s+(\d+)\s*m?\s*;/i)
    if (maxBodyMatch) {
        result.max_upload_size = parseInt(maxBodyMatch[1], 10)
    }

    // --- proxy_read_timeout ---
    const readTimeoutMatch = config.match(/proxy_read_timeout\s+(\d+)\s*s?\s*;/i)
    if (readTimeoutMatch) {
        result.proxy_read_timeout = parseInt(readTimeoutMatch[1], 10)
    }

    // --- proxy_connect_timeout ---
    const connectTimeoutMatch = config.match(/proxy_connect_timeout\s+(\d+)\s*s?\s*;/i)
    if (connectTimeoutMatch) {
        result.proxy_connect_timeout = parseInt(connectTimeoutMatch[1], 10)
    }

    // --- proxy_http_version → http_version ---
    const httpVersionMatch = config.match(/proxy_http_version\s+([\d.]+)\s*;/i)
    if (httpVersionMatch) {
        result.http_version = `http/${httpVersionMatch[1]}`
    }

    // --- ssl_protocols → tls_version (pick minimum) ---
    const sslProtocolsMatch = config.match(/ssl_protocols\s+([^;]+);/i)
    if (sslProtocolsMatch) {
        const protocols = sslProtocolsMatch[1].trim()
        if (protocols.includes('TLSv1.3') && !protocols.includes('TLSv1.2')) {
            result.tls_version = 'TLSv1.3'
        } else if (protocols.includes('TLSv1.2')) {
            result.tls_version = 'TLSv1.2'
        }
    }

    // --- Custom locations (excluding / and known pattern locations) ---
    const customLocations = extractCustomLocations(config, upstreamMap)
    if (customLocations.length > 0) {
        result.custom_locations = customLocations
    }

    return result
}

/**
 * Extract a specific location block's content from config.
 * Handles nested braces.
 */
function extractLocationBlock(config, path) {
    // Escape regex special chars in path
    const escapedPath = path.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    const pattern = new RegExp(`location\\s+${escapedPath}\\s*\\{`, 'g')
    let match
    let lastMatch = null

    // Find the last match (could appear in both http and https server blocks)
    while ((match = pattern.exec(config)) !== null) {
        lastMatch = match
    }

    if (!lastMatch) return null

    // Extract block content by counting braces
    let depth = 1
    let i = lastMatch.index + lastMatch[0].length
    const start = i
    while (i < config.length && depth > 0) {
        if (config[i] === '{') depth++
        if (config[i] === '}') depth--
        i++
    }

    return config.substring(start, i - 1)
}

/**
 * Extract custom location blocks (non-standard ones defined by users).
 * Skips well-known built-in locations.
 * Resolves upstream references back to actual server lists.
 */
function extractCustomLocations(config, upstreamMap = {}) {
    const builtinPatterns = [
        /^\/$/,                             // location /
        /^\^~\s+\/\.well-known/,           // ACME challenges  
        /^~\*?\s+\\\./,                    // hidden files
        /^~\*?\s+\\\.\(git/,              // sensitive files
        /^~\*?\s+\\\.\(jpg/,             // image caching
        /^~\*?\s+\\\.\(css/,             // css/js caching
        /^~\*?\s+\^\/?api\//,             // API rate limiting
    ]

    const locations = []
    const locationRegex = /# Custom Location:\s*([^\n]+)\n\s*location\s+([^\s{]+)\s*\{/g
    let match

    while ((match = locationRegex.exec(config)) !== null) {
        const path = match[2].trim()

        // Skip built-in locations
        let isBuiltin = false
        for (const bp of builtinPatterns) {
            if (bp.test(path)) { isBuiltin = true; break }
        }
        if (isBuiltin) continue

        // Extract block content
        const blockContent = extractLocationBlock(config, path)
        if (!blockContent) continue

        const loc = { path, backends: [], load_balance_method: 'round_robin' }

        // Extract proxy_pass
        const proxyMatch = blockContent.match(/proxy_pass\s+([^;]+);/)
        if (proxyMatch) {
            const proxyVal = proxyMatch[1].trim()
            // Check if proxy_pass references an upstream (e.g. http://xxx_backend)
            const upstreamRef = proxyVal.match(/^https?:\/\/(\w+)_backend$/)
            if (upstreamRef) {
                const upstreamName = upstreamRef[1]
                if (upstreamMap[upstreamName]) {
                    // Resolve upstream to actual servers
                    loc.backends = upstreamMap[upstreamName].servers
                    loc.load_balance_method = upstreamMap[upstreamName].lbMethod
                    loc.proxy_pass = '' // Clear since it's resolved to backends
                } else {
                    loc.proxy_pass = proxyVal
                }
            } else {
                loc.proxy_pass = proxyVal
            }
        }

        // Check WebSocket
        if (/proxy_set_header\s+Upgrade/i.test(blockContent)) {
            loc.websocket_enabled = true
        }

        locations.push(loc)
    }

    return locations
}

export default parseNginxConfig
