package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// GeoIPService handles geolocation lookups
type GeoIPService struct {
	httpClient *http.Client
}

// GeoIPResponse represents the response from ip-api.com
type GeoIPResponse struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Query       string  `json:"query"`
	Message     string  `json:"message,omitempty"`
}

// NewGeoIPService creates a new GeoIP service
func NewGeoIPService() *GeoIPService {
	return &GeoIPService{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// LookupIP performs a geolocation lookup for the given IP address
// Uses ip-api.com free API (no key required, 45 req/min limit)
func (s *GeoIPService) LookupIP(ip string) (*GeoIPResponse, error) {
	// Skip lookup for localhost/private IPs
	if ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
		return &GeoIPResponse{
			Status:      "success",
			Country:     "Localhost",
			CountryCode: "XX",
			Query:       ip,
		}, nil
	}

	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,country,countryCode,region,regionName,city,zip,lat,lon,timezone,isp,org,as,query", ip)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup IP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	var geoResp GeoIPResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if geoResp.Status == "fail" {
		return nil, fmt.Errorf("geolocation lookup failed: %s", geoResp.Message)
	}

	return &geoResp, nil
}

// GetCountryCode is a convenience method to get just the country code
func (s *GeoIPService) GetCountryCode(ip string) (string, error) {
	geoResp, err := s.LookupIP(ip)
	if err != nil {
		return "", err
	}
	return geoResp.CountryCode, nil
}
