package proxy

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/aleh/docode-waf/internal/config"
	"github.com/aleh/docode-waf/internal/models"
	"github.com/aleh/docode-waf/internal/services"
)

type ReverseProxy struct {
	config       *config.Config
	vhosts       map[string]*models.VHost
	proxies      map[string]*httputil.ReverseProxy
	transport    *http.Transport
	vhostService *services.VHostService
}

func NewReverseProxy(cfg *config.Config, vhostService *services.VHostService) *ReverseProxy {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	return &ReverseProxy{
		config:       cfg,
		vhosts:       make(map[string]*models.VHost),
		proxies:      make(map[string]*httputil.ReverseProxy),
		transport:    transport,
		vhostService: vhostService,
	}
}

func (rp *ReverseProxy) LoadVHosts(vhosts []*models.VHost) {
	rp.vhosts = make(map[string]*models.VHost)
	rp.proxies = make(map[string]*httputil.ReverseProxy)

	for _, vhost := range vhosts {
		if !vhost.Enabled {
			continue
		}

		rp.vhosts[vhost.Domain] = vhost

		target, err := url.Parse(vhost.BackendURL)
		if err != nil {
			continue
		}

		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.Transport = rp.transport
		proxy.ErrorHandler = rp.errorHandler

		// Modify request before forwarding
		director := proxy.Director
		proxy.Director = func(req *http.Request) {
			director(req)
			req.Header.Set("X-Forwarded-Host", req.Host)
			req.Header.Set("X-Origin-Host", target.Host)
			req.Header.Set("X-Real-IP", getClientIP(req))
		}

		rp.proxies[vhost.Domain] = proxy
	}
}

// ReloadVHosts reloads all vhosts from database
func (rp *ReverseProxy) ReloadVHosts() error {
	vhosts, err := rp.vhostService.ListVHosts()
	if err != nil {
		return err
	}
	rp.LoadVHosts(vhosts)
	return nil
}

func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	// Remove port from host if present
	if colonPos := strings.Index(host, ":"); colonPos != -1 {
		host = host[:colonPos]
	}

	proxy, exists := rp.proxies[host]
	if !exists {
		http.Error(w, "Virtual host not found", http.StatusNotFound)
		return
	}

	// Add context with start time for logging
	ctx := context.WithValue(r.Context(), "start_time", time.Now())
	r = r.WithContext(ctx)

	proxy.ServeHTTP(w, r)
}

func (rp *ReverseProxy) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, "Bad Gateway", http.StatusBadGateway)
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// ResponseWriter wrapper to capture status code and bytes written
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}
