package middleware

import (
	"net/http"
	"time"
)
type LoggingMiddleware struct {
	logger Logger
}
type Logger interface {
	LogRequest(log RequestLog)
	LogAttack(log AttackLog)
type RequestLog struct {
	Timestamp    time.Time
	ClientIP     string
	Method       string
	URL          string
	StatusCode   int
	ResponseTime int
	BytesSent    int64
	UserAgent    string
	Blocked      bool
	BlockReason  string
type AttackLog struct {
	Timestamp   time.Time
	ClientIP    string
	AttackType  string
	Severity    string
	Description string
	Blocked     bool
func NewLoggingMiddleware(logger Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
func (lm *LoggingMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Wrap response writer to capture status and bytes
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		// Call next handler
		next.ServeHTTP(rw, r)
		// Calculate response time
		duration := time.Since(start)
		// Get context values
		blocked := false
		blockReason := ""
		if val := r.Context().Value("blocked"); val != nil {
			blocked = val.(bool)
		if val := r.Context().Value("block_reason"); val != nil {
			blockReason = val.(string)
		// Log request
		log := RequestLog{
			Timestamp:    start,
			ClientIP:     getClientIP(r),
			Method:       r.Method,
			URL:          r.URL.String(),
			StatusCode:   rw.statusCode,
			ResponseTime: int(duration.Milliseconds()),
			BytesSent:    rw.bytesWritten,
			UserAgent:    r.UserAgent(),
			Blocked:      blocked,
			BlockReason:  blockReason,
		lm.logger.LogRequest(log)
		// If blocked, log as attack
		if blocked {
			attackLog := AttackLog{
				Timestamp:   start,
				ClientIP:    getClientIP(r),
				AttackType:  blockReason,
				Severity:    getSeverity(blockReason),
				Description: getDescription(blockReason, r),
				Blocked:     true,
			}
			lm.logger.LogAttack(attackLog)
	})
func getSeverity(blockReason string) string {
	switch blockReason {
	case "http_flood", "sql_injection":
		return "high"
	case "bot_detected", "rate_limit":
		return "medium"
	case "ip_blocked":
		return "low"
	default:
func getDescription(blockReason string, r *http.Request) string {
	case "rate_limit":
		return "Rate limit exceeded for IP: " + getClientIP(r)
		return "IP address blocked: " + getClientIP(r)
	case "bot_detected":
		return "Bot detected with User-Agent: " + r.UserAgent()
	case "http_flood":
		return "HTTP flood attack detected from IP: " + getClientIP(r)
		return "Security rule triggered: " + blockReason
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
