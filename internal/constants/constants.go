package constants

// Path constants
const (
	NginxConfigDir = "/app/nginx/conf.d"
	SSLCertDir     = "/app/ssl/certificates"
)

// Route constants
const (
	RouteVHostID       = "/vhosts/:id"
	RouteCertificateID = "/certificates/:id"
)

// SQL query constants
const (
	SQLCountTrafficLogs = "SELECT COUNT(*) FROM traffic_logs"
	SQLCountVHosts      = "SELECT COUNT(*) FROM vhosts WHERE enabled = true"
	SQLCountDistinctIPs = "SELECT COUNT(DISTINCT client_ip) FROM traffic_logs"
)

// Error message constants
const (
	ErrVHostNotFound = "VHost not found"
)
