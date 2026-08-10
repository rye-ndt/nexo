package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hexago/internal/helpers/constances"
)

const shippedConfig = "../../../../config.yaml"

func writeConfig(t *testing.T, old, new string) string {
	t.Helper()

	raw, err := os.ReadFile(shippedConfig)
	if err != nil {
		t.Fatalf("cannot read the shipped config: %v", err)
	}

	body := string(raw)

	if old != "" {
		if !strings.Contains(body, old) {
			t.Fatalf("the shipped config no longer contains %q", old)
		}

		body = strings.Replace(body, old, new, 1)
	}

	path := filepath.Join(t.TempDir(), "config.yaml")

	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("cannot write the test config: %v", err)
	}

	return path
}

func TestNewAcceptsShippedConfig(t *testing.T) {
	cfg, err := New(writeConfig(t, "", ""))
	if err != nil {
		t.Fatalf("the shipped config must boot, got %v", err)
	}

	mcp := cfg.Read().MCPServers

	if mcp.VerifierBytes < constances.PKCEMinVerifierBytes {
		t.Fatalf("verifier_bytes %d is below the floor %d", mcp.VerifierBytes, constances.PKCEMinVerifierBytes)
	}

	if mcp.StateBytes < constances.PKCEMinStateBytes {
		t.Fatalf("state_bytes %d is below the floor %d", mcp.StateBytes, constances.PKCEMinStateBytes)
	}

	if mcp.ChallengeMethod != constances.PKCESupportedChallengeMethod {
		t.Fatalf("challenge_method %q, want %q", mcp.ChallengeMethod, constances.PKCESupportedChallengeMethod)
	}
}

func TestNewRejectsPKCEValuesBelowTheFloor(t *testing.T) {
	cases := []struct {
		name string
		old  string
		new  string
		want string
	}{
		{
			name: "verifier below the rfc 7636 floor",
			old:  "\n  verifier_bytes: 32",
			new:  "\n  verifier_bytes: 1",
			want: "verifier_bytes",
		},
		{
			name: "state below the floor",
			old:  "\n  state_bytes: 16",
			new:  "\n  state_bytes: 4",
			want: "state_bytes",
		},
		{
			name: "unsupported challenge method",
			old:  "\n  challenge_method: S256",
			new:  "\n  challenge_method: plain",
			want: "challenge_method",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := New(writeConfig(t, tc.old, tc.new))
			if err == nil {
				t.Fatalf("want a boot failure, got a config: %+v", cfg.Read().MCPServers)
			}

			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("the error must name %s, got %v", tc.want, err)
			}
		})
	}
}

func TestNewIgnoresRemovedFloorKeys(t *testing.T) {
	path := writeConfig(t, "\n  verifier_bytes: 32", "\n  min_verifier_bytes: 1\n  verifier_bytes: 32")

	if _, err := New(path); err != nil {
		t.Fatalf("a stale min_verifier_bytes key must not lower the floor nor break boot, got %v", err)
	}
}
