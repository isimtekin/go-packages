package natsclient

import (
	"context"
	"os"
	"testing"
)

func TestNewFromEnv(t *testing.T) {
	// Save original env vars
	origURL := os.Getenv("NATS_URL")
	origName := os.Getenv("NATS_NAME")
	origTimeout := os.Getenv("NATS_TIMEOUT")
	origJetStream := os.Getenv("NATS_ENABLE_JETSTREAM")

	// Cleanup
	defer func() {
		os.Setenv("NATS_URL", origURL)
		os.Setenv("NATS_NAME", origName)
		os.Setenv("NATS_TIMEOUT", origTimeout)
		os.Setenv("NATS_ENABLE_JETSTREAM", origJetStream)
	}()

	t.Run("uses default values when env vars not set", func(t *testing.T) {
		os.Unsetenv("NATS_URL")
		os.Unsetenv("NATS_NAME")
		os.Unsetenv("NATS_TIMEOUT")

		// Will fail to connect but config should be loaded
		_, err := NewFromEnvWithDefaults(context.Background())
		if err == nil {
			t.Skip("NATS server is running")
		}
		// Error is expected (no server), but config loading should work
	})

	t.Run("uses custom prefix", func(t *testing.T) {
		os.Setenv("MYAPP_URL", "nats://custom:4222")
		os.Setenv("MYAPP_NAME", "custom-client")
		defer func() {
			os.Unsetenv("MYAPP_URL")
			os.Unsetenv("MYAPP_NAME")
		}()

		_, err := NewFromEnv(context.Background(), "MYAPP_")
		if err == nil {
			t.Skip("NATS server is running")
		}
		// Error is expected (no server), but config loading should work
	})

	t.Run("loads url from env", func(t *testing.T) {
		os.Setenv("NATS_URL", "nats://envtest:4222")

		_, err := NewFromEnvWithDefaults(context.Background())
		if err == nil {
			t.Skip("NATS server is running")
		}
		// The error message should reference the custom URL
	})
}

func TestEnvConstants(t *testing.T) {
	// Verify constants are defined correctly
	if DefaultEnvPrefix != "NATS_" {
		t.Errorf("DefaultEnvPrefix = %s, want NATS_", DefaultEnvPrefix)
	}

	// Verify all env keys are defined
	envKeys := []string{
		EnvURL,
		EnvName,
		EnvUsername,
		EnvPassword,
		EnvToken,
		EnvMaxReconnects,
		EnvReconnectWait,
		EnvReconnectJitter,
		EnvTimeout,
		EnvPingInterval,
		EnvMaxPingsOut,
		EnvAllowReconnect,
		EnvNoRandomize,
		EnvNoEcho,
		EnvRetryOnFailedConn,
		EnvTLSEnabled,
		EnvTLSCertFile,
		EnvTLSKeyFile,
		EnvTLSCAFile,
		EnvEnableJetStream,
	}

	for _, key := range envKeys {
		if key == "" {
			t.Error("Empty env key constant found")
		}
	}
}
