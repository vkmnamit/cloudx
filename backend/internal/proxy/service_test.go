package proxy

import (
	"strings"
	"testing"
)

func TestServiceHostname(t *testing.T) {
	host, err := serviceHostname("Rocketry Final", "12345678-1234-4234-8234-123456789abc", "172.198.131.112.sslip.io")
	if err != nil {
		t.Fatal(err)
	}
	if want := "rocketry-final-123456781234.172.198.131.112.sslip.io"; host != want {
		t.Fatalf("serviceHostname() = %q, want %q", host, want)
	}
}

func TestServiceHostnameLimitsLabelLength(t *testing.T) {
	name := strings.Repeat("a", 80)
	host, err := serviceHostname(name, "12345678-1234-4234-8234-123456789abc", "apps.example.net")
	if err != nil {
		t.Fatal(err)
	}
	label := strings.SplitN(host, ".", 2)[0]
	if len(label) > 63 {
		t.Fatalf("hostname label length = %d, want at most 63", len(label))
	}
}

func TestValidateDomainRejectsInvalidLabels(t *testing.T) {
	for _, domain := range []string{"-invalid.example", "invalid_.example", "example..net", "a/" + "b"} {
		if err := validateDomain(domain); err == nil {
			t.Errorf("validateDomain(%q) succeeded, want error", domain)
		}
	}
}

func TestRouteConfigurationTargetsContainerNetworkAddress(t *testing.T) {
	config := routeConfiguration("cloudx-1234", "app.example.net", "cloudx-deployment-5678", 3000)
	for _, expected := range []string{
		`Host("app.example.net")`,
		"certResolver: le",
		"http://cloudx-deployment-5678:3000",
	} {
		if !strings.Contains(config, expected) {
			t.Errorf("route configuration does not contain %q:\n%s", expected, config)
		}
	}
}
