package services

import (
	"strings" // Added for strings.Contains
	"testing"

	"github.com/aleh/docode-waf/internal/models"
)

func TestGenerateVHostConfigCustomLocationRewrite(t *testing.T) {
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

func TestGenerateVHostConfigStripProtocol(t *testing.T) {
	service := NewNginxConfigService()

	vhost := &models.VHost{
		ID:                "vhost-strip",
		Name:              "Strip Test",
		Domain:            "strip.example.com",
		Type:              "proxy",
		BackendURL:        "http://backend1:8080/",
		Backends:          []string{"https://backend2:9090", "backend3:7070/"},
		LoadBalanceMethod: "round_robin",
	}

	// We'll use a hack to call the internal logic by manually constructing the data
	// since GenerateVHostConfig calls the DB and writes to disk.
	
	// Prepare vhost with locations (similar to logic in GenerateVHostConfig)
	upstreamName := sanitizeDomainForUpstream(vhost.Domain)
	allBackends := []string{}
	for _, b := range vhost.Backends {
		allBackends = append(allBackends, stripProtocol(b))
	}
	if len(allBackends) > 0 && vhost.BackendURL != "" {
		allBackends = append([]string{stripProtocol(vhost.BackendURL)}, allBackends...)
	}

	data := &VHostWithLocations{
		VHost:        vhost,
		UpstreamName: upstreamName,
		HasUpstream:  len(allBackends) > 0,
	}
	data.Backends = allBackends

	content, err := service.GenerateVHostConfigContent(data)
	if err != nil {
		t.Fatalf("Failed to generate config: %v", err)
	}

	// Check if protocols are stripped in the upstream block
	expectedServer1 := "server backend1:8080 weight=1"
	expectedServer2 := "server backend2:9090 weight=1"
	expectedServer3 := "server backend3:7070 weight=1"

	if !strings.Contains(content, expectedServer1) {
		t.Errorf("Expected upstream to contain %q, but it didn't. Content:\n%s", expectedServer1, content)
	}
	if !strings.Contains(content, expectedServer2) {
		t.Errorf("Expected upstream to contain %q, but it didn't.", expectedServer2)
	}
	if !strings.Contains(content, expectedServer3) {
		t.Errorf("Expected upstream to contain %q, but it didn't (trailing slash test).", expectedServer3)
	}

	// Ensure proxy_pass still has http://
	expectedProxyPass := "proxy_pass http://strip_example_com_backend;"
	if !strings.Contains(content, expectedProxyPass) {
		t.Errorf("Expected proxy_pass to contain %q, but it didn't.", expectedProxyPass)
	}
}
