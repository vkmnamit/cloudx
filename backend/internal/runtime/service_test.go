package runtime

import (
	"strings"
	"testing"
)

func TestContainerRunArgsUsePrivateGatewayBindingAndLimits(t *testing.T) {
	args := containerRunArgs(
		"12345678-1234-4234-8234-123456789abc",
		"cloudx-deployment-12345678123442348234123456789abc",
		"cloudx-test",
		[]string{"-e", "APP_ENV=production"},
		3000,
		"172.18.0.1",
	)
	joined := strings.Join(args, " ")
	for _, expected := range []string{
		"172.18.0.1::3000",
		"--memory 512m",
		"--cpus 0.5",
		"--pids-limit 128",
		"--cap-drop ALL",
		"--security-opt no-new-privileges:true",
	} {
		if !strings.Contains(joined, expected) {
			t.Errorf("container arguments do not contain %q: %s", expected, joined)
		}
	}
	if strings.Contains(joined, "0.0.0.0:") {
		t.Fatalf("container port must not be published on all host interfaces: %s", joined)
	}
}

func TestParseHostPortOutput(t *testing.T) {
	port, err := parseHostPortOutput("172.18.0.1:32768\n")
	if err != nil {
		t.Fatal(err)
	}
	if port != 32768 {
		t.Fatalf("parseHostPortOutput() = %d, want 32768", port)
	}
}

func TestParseHostPortOutputRejectsInvalidPort(t *testing.T) {
	for _, output := range []string{"172.18.0.1:0\n", "invalid\n", ""} {
		if _, err := parseHostPortOutput(output); err == nil {
			t.Errorf("parseHostPortOutput(%q) succeeded, want error", output)
		}
	}
}
