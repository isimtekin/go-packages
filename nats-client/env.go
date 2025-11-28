package natsclient

import (
	"context"

	envutil "github.com/isimtekin/go-packages/env-util"
)

// Environment variable keys (without prefix)
const (
	EnvURL               = "URL"
	EnvName              = "NAME"
	EnvUsername          = "USERNAME"
	EnvPassword          = "PASSWORD"
	EnvToken             = "TOKEN"
	EnvMaxReconnects     = "MAX_RECONNECTS"
	EnvReconnectWait     = "RECONNECT_WAIT"
	EnvReconnectJitter   = "RECONNECT_JITTER"
	EnvTimeout           = "TIMEOUT"
	EnvPingInterval      = "PING_INTERVAL"
	EnvMaxPingsOut       = "MAX_PINGS_OUT"
	EnvAllowReconnect    = "ALLOW_RECONNECT"
	EnvNoRandomize       = "NO_RANDOMIZE"
	EnvNoEcho            = "NO_ECHO"
	EnvRetryOnFailedConn = "RETRY_ON_FAILED_CONN"
	EnvTLSEnabled        = "TLS_ENABLED"
	EnvTLSCertFile       = "TLS_CERT_FILE"
	EnvTLSKeyFile        = "TLS_KEY_FILE"
	EnvTLSCAFile         = "TLS_CA_FILE"
	EnvEnableJetStream   = "ENABLE_JETSTREAM"
)

// DefaultEnvPrefix is the default prefix for environment variables
const DefaultEnvPrefix = "NATS_"

// NewFromEnvWithDefaults creates a new client from environment variables using default prefix "NATS_"
// Example environment variables:
//   - NATS_URL: NATS server URL (default: nats://localhost:4222)
//   - NATS_NAME: Client name
//   - NATS_USERNAME: Username for authentication
//   - NATS_PASSWORD: Password for authentication
//   - NATS_TOKEN: Token for authentication
//   - NATS_TIMEOUT: Connection timeout (e.g., "5s")
//   - NATS_ENABLE_JETSTREAM: Enable JetStream (true/false)
func NewFromEnvWithDefaults(_ context.Context) (*Client, error) {
	return NewFromEnv(context.Background(), DefaultEnvPrefix)
}

// NewFromEnv creates a new client from environment variables with custom prefix
// Example with prefix "MYAPP_NATS_":
//   - MYAPP_NATS_URL
//   - MYAPP_NATS_USERNAME
//   - etc.
func NewFromEnv(_ context.Context, prefix string) (*Client, error) {
	config := DefaultConfig()

	env := envutil.NewWithOptions(
		envutil.WithPrefix(prefix),
		envutil.WithSilent(true),
	)

	// Connection settings
	config.URL = env.GetString(EnvURL, config.URL)
	config.Name = env.GetString(EnvName, config.Name)
	config.Username = env.GetString(EnvUsername, config.Username)
	config.Password = env.GetString(EnvPassword, config.Password)
	config.Token = env.GetString(EnvToken, config.Token)

	// Pool settings
	config.MaxReconnects = env.GetInt(EnvMaxReconnects, config.MaxReconnects)
	config.ReconnectWait = env.GetDuration(EnvReconnectWait, config.ReconnectWait)
	config.ReconnectJitter = env.GetDuration(EnvReconnectJitter, config.ReconnectJitter)
	config.Timeout = env.GetDuration(EnvTimeout, config.Timeout)
	config.PingInterval = env.GetDuration(EnvPingInterval, config.PingInterval)
	config.MaxPingsOut = env.GetInt(EnvMaxPingsOut, config.MaxPingsOut)
	config.AllowReconnect = env.GetBool(EnvAllowReconnect, config.AllowReconnect)
	config.NoRandomize = env.GetBool(EnvNoRandomize, config.NoRandomize)
	config.NoEcho = env.GetBool(EnvNoEcho, config.NoEcho)
	config.RetryOnFailedConn = env.GetBool(EnvRetryOnFailedConn, config.RetryOnFailedConn)

	// TLS settings
	config.TLSEnabled = env.GetBool(EnvTLSEnabled, config.TLSEnabled)
	config.TLSCertFile = env.GetString(EnvTLSCertFile, config.TLSCertFile)
	config.TLSKeyFile = env.GetString(EnvTLSKeyFile, config.TLSKeyFile)
	config.TLSCAFile = env.GetString(EnvTLSCAFile, config.TLSCAFile)

	// JetStream settings
	config.EnableJetStream = env.GetBool(EnvEnableJetStream, config.EnableJetStream)

	return New(config)
}
