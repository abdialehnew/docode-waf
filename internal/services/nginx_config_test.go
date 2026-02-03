package services

import (
	"strings" // Added for strings.Contains
	"testing"

	"github.com/aleh/docode-waf/internal/models"
)

func TestGenerateVHostConfig_CustomLocationRewrite(t *testing.T) {
	service := NewNginxConfigService()

	vhost := &models.VHost{
		ID:     "vhost-1",
		Name:   "Test VHost",
		Domain: "example.com",
		Type:   "proxy",
	}

	// Manually construct VHostWithLocations since we are testing GenerateVHostConfigContent directly
	// and we don't have a DB to fetch locations.
	vhostWithLocs := &VHostWithLocations{
		VHost:        vhost,
		UpstreamName: "example_com",
		CustomLocations: []CustomLocation{
			{
				Path:         "/custom",
				HasUpstream:  true,
				UpstreamName: "example_com_loc_custom",
				Backends:     []string{"http://backend:8080"},
			},
		},
	}

	content, err := service.GenerateVHostConfigContent(vhostWithLocs)
	if err != nil {
		t.Fatalf("Failed to generate config: %v", err)
	}

	expectedRewrite := "rewrite ^/custom/(.*) /$1 break;"
	if !strings.Contains(content, expectedRewrite) {
		t.Errorf("Expected config to contain rewrite rule %q, but got:\n%s", expectedRewrite, content)
	}
}
