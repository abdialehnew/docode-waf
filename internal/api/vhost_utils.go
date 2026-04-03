package api

import (
	"encoding/json"

	"github.com/aleh/docode-waf/internal/constants"
	"github.com/aleh/docode-waf/internal/models"
)

// sanitizeVHostInput sets default values for VHost input
func sanitizeVHostInput(input *models.VHostInput) {
	if input.HTTPVersion == "" {
		input.HTTPVersion = constants.DefaultHTTPVersion
	}
	if input.MaxUploadSize == 0 {
		input.MaxUploadSize = 10
	}
	if input.ProxyReadTimeout == 0 {
		input.ProxyReadTimeout = 60
	}
	if input.ProxyConnectTimeout == 0 {
		input.ProxyConnectTimeout = 60
	}
	if input.CustomHeaders == nil {
		input.CustomHeaders = make(map[string]interface{})
	}
	if input.Backends == nil {
		input.Backends = []string{}
	}

	// Sanitize SSL Certificate ID
	if input.SSLCertificateID != nil && *input.SSLCertificateID == "" {
		input.SSLCertificateID = nil
	}
	if input.LoadBalanceMethod == "" {
		input.LoadBalanceMethod = constants.DefaultLoadBalanceMethod
	}
	if input.Type == "" {
		input.Type = "proxy"
	}
	if input.DefenseMode == "" {
		input.DefenseMode = "defense"
	}
	if input.BotDetectionType == "" {
		input.BotDetectionType = "turnstile"
	}
	if input.RecaptchaVersion == "" {
		input.RecaptchaVersion = "v2"
	}
	if input.RateLimitRequests == 0 {
		input.RateLimitRequests = 100
	}
	if input.RateLimitWindow == 0 {
		input.RateLimitWindow = 60
	}
}

// parseJSONStringArray parses a JSON string array into a string slice
func parseJSONStringArray(jsonStr *string) []string {
	if jsonStr == nil || *jsonStr == "" || *jsonStr == "[]" {
		return []string{}
	}
	var arr []string
	if err := json.Unmarshal([]byte(*jsonStr), &arr); err != nil {
		return []string{}
	}
	return arr
}

// parseJSONRawMessage parses json.RawMessage into a string slice
func parseJSONRawMessage(raw json.RawMessage) []string {
	if raw == nil {
		return []string{}
	}
	var arr []string
	if err := json.Unmarshal(raw, &arr); err != nil {
		return []string{}
	}
	if arr == nil {
		return []string{}
	}
	return arr
}
