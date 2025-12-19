package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
)
}	})		next.ServeHTTP(w, r)		}			return			http.Error(w, "Access denied", http.StatusForbidden)			r = r.WithContext(ctx)			ctx = context.WithValue(ctx, "block_reason", "ip_blocked")			ctx := context.WithValue(r.Context(), "blocked", true)		if ib.IsBlocked(ip) {		ip := getClientIP(r)	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {func (ib *IPBlocker) Middleware(next http.Handler) http.Handler {}	return false	}		}			}				return true			if err == nil && ipNet.Contains(clientIP) {			_, ipNet, err := net.ParseCIDR(blockedIP)		if strings.Contains(blockedIP, "/") {		// Check if it's a CIDR	for blockedIP := range ib.blacklist {	}		return false	if clientIP == nil {	clientIP := net.ParseIP(ip)	// Check CIDR blocks	}		return true	if ib.blacklist[ip] {	// Check if IP is in blacklist	}		return false	if ib.whitelist[ip] {	// Whitelist takes precedencefunc (ib *IPBlocker) IsBlocked(ip string) bool {}	}		ib.whitelist[ip] = true	for _, ip := range ips {	ib.whitelist = make(map[string]bool)func (ib *IPBlocker) LoadWhitelist(ips []string) {}	}		ib.blacklist[ip] = true	for _, ip := range ips {	ib.blacklist = make(map[string]bool)func (ib *IPBlocker) LoadBlacklist(ips []string) {}	delete(ib.whitelist, ip)func (ib *IPBlocker) RemoveFromWhitelist(ip string) {}	delete(ib.blacklist, ip)func (ib *IPBlocker) RemoveFromBlacklist(ip string) {}	ib.whitelist[ip] = truefunc (ib *IPBlocker) AddToWhitelist(ip string) {}	ib.blacklist[ip] = truefunc (ib *IPBlocker) AddToBlacklist(ip string) {}	}		whitelist: make(map[string]bool),		blacklist: make(map[string]bool),	return &IPBlocker{func NewIPBlocker() *IPBlocker {}	whitelist map[string]bool	blacklist map[string]booltype IPBlocker struct {
