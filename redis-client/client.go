package redisclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client represents the Redis client wrapper
type Client struct {
	config *Config
	client *redis.Client

	mu     sync.RWMutex
	closed bool
}

// New creates a new Redis client with the given configuration
func New(config *Config) (*Client, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	client := &Client{
		config: config,
		closed: false,
	}

	// Initialize Redis client
	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return client, nil
}

// NewWithOptions creates a new client with functional options
func NewWithOptions(opts ...Option) (*Client, error) {
	config := DefaultConfig()

	for _, opt := range opts {
		opt(config)
	}

	return New(config)
}

// connect establishes the Redis connection
func (c *Client) connect() error {
	opts := &redis.Options{
		Addr:            c.config.Addr,
		Password:        c.config.Password,
		DB:              c.config.DB,
		MaxRetries:      c.config.MaxRetries,
		MinIdleConns:    c.config.MinIdleConns,
		MaxIdleConns:    c.config.MaxIdleConns,
		PoolSize:        c.config.PoolSize,
		PoolTimeout:     c.config.PoolTimeout,
		ConnMaxIdleTime: c.config.ConnMaxIdleTime,
		ConnMaxLifetime: c.config.ConnMaxLifetime,
		DialTimeout:     c.config.DialTimeout,
		ReadTimeout:     c.config.ReadTimeout,
		WriteTimeout:    c.config.WriteTimeout,
	}

	// Configure TLS if enabled
	if c.config.TLSEnabled {
		tlsConfig, err := c.createTLSConfig()
		if err != nil {
			return fmt.Errorf("failed to create TLS config: %w", err)
		}
		opts.TLSConfig = tlsConfig
	}

	c.client = redis.NewClient(opts)
	return nil
}

// createTLSConfig creates TLS configuration
func (c *Client) createTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Load client cert if provided
	if c.config.TLSCertFile != "" && c.config.TLSKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(c.config.TLSCertFile, c.config.TLSKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Load CA cert if provided
	if c.config.TLSCAFile != "" {
		caCert, err := os.ReadFile(c.config.TLSCAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA cert")
		}
		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrAlreadyClosed
	}

	if c.client != nil {
		if err := c.client.Close(); err != nil {
			return fmt.Errorf("failed to close redis client: %w", err)
		}
	}

	c.closed = true
	return nil
}

// Ping checks if the connection is alive
func (c *Client) Ping(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	return c.client.Ping(ctx).Err()
}

// Client returns the underlying go-redis client for advanced operations
func (c *Client) Client() *redis.Client {
	return c.client
}

// ====================
// String Operations
// ====================

// Get retrieves the value of a key
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return "", ErrClientClosed
	}

	if key == "" {
		return "", ErrInvalidKey
	}

	return c.client.Get(ctx, key).Result()
}

// Set sets the value of a key with optional TTL
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	if key == "" {
		return ErrInvalidKey
	}

	return c.client.Set(ctx, key, value, ttl).Err()
}

// SetNX sets the value of a key only if it does not exist
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return false, ErrClientClosed
	}

	if key == "" {
		return false, ErrInvalidKey
	}

	return c.client.SetNX(ctx, key, value, ttl).Result()
}

// SetEX sets the value and expiration of a key
func (c *Client) SetEX(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	if key == "" {
		return ErrInvalidKey
	}

	if ttl <= 0 {
		return ErrInvalidTTL
	}

	return c.client.SetEx(ctx, key, value, ttl).Err()
}

// GetSet sets a new value and returns the old value
func (c *Client) GetSet(ctx context.Context, key string, value interface{}) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return "", ErrClientClosed
	}

	if key == "" {
		return "", ErrInvalidKey
	}

	return c.client.GetSet(ctx, key, value).Result()
}

// MGet retrieves values of multiple keys
func (c *Client) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	return c.client.MGet(ctx, keys...).Result()
}

// MSet sets multiple key-value pairs
func (c *Client) MSet(ctx context.Context, values ...interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	return c.client.MSet(ctx, values...).Err()
}

// Incr increments the integer value of a key by one
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.Incr(ctx, key).Result()
}

// IncrBy increments the integer value of a key by the given amount
func (c *Client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.IncrBy(ctx, key, value).Result()
}

// Decr decrements the integer value of a key by one
func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.Decr(ctx, key).Result()
}

// DecrBy decrements the integer value of a key by the given amount
func (c *Client) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.DecrBy(ctx, key, value).Result()
}

// ====================
// Key Operations
// ====================

// Del deletes one or more keys
func (c *Client) Del(ctx context.Context, keys ...string) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	return c.client.Del(ctx, keys...).Result()
}

// Exists checks if one or more keys exist
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	return c.client.Exists(ctx, keys...).Result()
}

// Expire sets a key's time to live in seconds
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return false, ErrClientClosed
	}

	if key == "" {
		return false, ErrInvalidKey
	}

	return c.client.Expire(ctx, key, ttl).Result()
}

// TTL returns the time to live for a key
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.TTL(ctx, key).Result()
}

// Persist removes the expiration from a key
func (c *Client) Persist(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return false, ErrClientClosed
	}

	if key == "" {
		return false, ErrInvalidKey
	}

	return c.client.Persist(ctx, key).Result()
}

// Keys returns all keys matching a pattern
func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	return c.client.Keys(ctx, pattern).Result()
}

// ====================
// Hash Operations
// ====================

// HSet sets field in the hash stored at key to value
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.HSet(ctx, key, values...).Result()
}

// HGet returns the value associated with field in the hash
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return "", ErrClientClosed
	}

	if key == "" {
		return "", ErrInvalidKey
	}

	return c.client.HGet(ctx, key, field).Result()
}

// HGetAll returns all fields and values in a hash
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	if key == "" {
		return nil, ErrInvalidKey
	}

	return c.client.HGetAll(ctx, key).Result()
}

// HDel deletes one or more hash fields
func (c *Client) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.HDel(ctx, key, fields...).Result()
}

// HExists determines if a hash field exists
func (c *Client) HExists(ctx context.Context, key, field string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return false, ErrClientClosed
	}

	if key == "" {
		return false, ErrInvalidKey
	}

	return c.client.HExists(ctx, key, field).Result()
}

// HIncrBy increments the integer value of a hash field
func (c *Client) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.HIncrBy(ctx, key, field, incr).Result()
}

// HKeys returns all field names in a hash
func (c *Client) HKeys(ctx context.Context, key string) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	if key == "" {
		return nil, ErrInvalidKey
	}

	return c.client.HKeys(ctx, key).Result()
}

// HLen returns the number of fields in a hash
func (c *Client) HLen(ctx context.Context, key string) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.HLen(ctx, key).Result()
}

// ====================
// List Operations
// ====================

// LPush inserts all the specified values at the head of the list
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.LPush(ctx, key, values...).Result()
}

// RPush inserts all the specified values at the tail of the list
func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.RPush(ctx, key, values...).Result()
}

// LPop removes and returns the first element of the list
func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return "", ErrClientClosed
	}

	if key == "" {
		return "", ErrInvalidKey
	}

	return c.client.LPop(ctx, key).Result()
}

// RPop removes and returns the last element of the list
func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return "", ErrClientClosed
	}

	if key == "" {
		return "", ErrInvalidKey
	}

	return c.client.RPop(ctx, key).Result()
}

// LRange returns the specified elements of the list
func (c *Client) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	if key == "" {
		return nil, ErrInvalidKey
	}

	return c.client.LRange(ctx, key, start, stop).Result()
}

// LLen returns the length of the list
func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.LLen(ctx, key).Result()
}

// ====================
// Set Operations
// ====================

// SAdd adds one or more members to a set
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.SAdd(ctx, key, members...).Result()
}

// SMembers returns all members of a set
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	if key == "" {
		return nil, ErrInvalidKey
	}

	return c.client.SMembers(ctx, key).Result()
}

// SIsMember checks if a member exists in a set
func (c *Client) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return false, ErrClientClosed
	}

	if key == "" {
		return false, ErrInvalidKey
	}

	return c.client.SIsMember(ctx, key, member).Result()
}

// SRem removes one or more members from a set
func (c *Client) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.SRem(ctx, key, members...).Result()
}

// SCard returns the number of members in a set
func (c *Client) SCard(ctx context.Context, key string) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.SCard(ctx, key).Result()
}

// ====================
// Sorted Set Operations
// ====================

// ZAdd adds one or more members to a sorted set
func (c *Client) ZAdd(ctx context.Context, key string, members ...redis.Z) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.ZAdd(ctx, key, members...).Result()
}

// ZRange returns the specified range of elements in a sorted set
func (c *Client) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	if key == "" {
		return nil, ErrInvalidKey
	}

	return c.client.ZRange(ctx, key, start, stop).Result()
}

// ZRangeWithScores returns the specified range with scores
func (c *Client) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	if key == "" {
		return nil, ErrInvalidKey
	}

	return c.client.ZRangeWithScores(ctx, key, start, stop).Result()
}

// ZRem removes one or more members from a sorted set
func (c *Client) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.ZRem(ctx, key, members...).Result()
}

// ZCard returns the number of members in a sorted set
func (c *Client) ZCard(ctx context.Context, key string) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.ZCard(ctx, key).Result()
}

// ZScore returns the score of a member in a sorted set
func (c *Client) ZScore(ctx context.Context, key, member string) (float64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.ZScore(ctx, key, member).Result()
}

// ====================
// Pipeline Operations
// ====================

// Pipeline creates a new pipeline
func (c *Client) Pipeline() redis.Pipeliner {
	return c.client.Pipeline()
}

// TxPipeline creates a new transaction pipeline
func (c *Client) TxPipeline() redis.Pipeliner {
	return c.client.TxPipeline()
}

// ====================
// Pub/Sub Operations
// ====================

// Publish publishes a message to a channel
func (c *Client) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	return c.client.Publish(ctx, channel, message).Result()
}

// Subscribe subscribes to the given channels
func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.client.Subscribe(ctx, channels...)
}

// PSubscribe subscribes to channels matching the given patterns
func (c *Client) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	return c.client.PSubscribe(ctx, patterns...)
}

// ====================
// Transaction Operations
// ====================

// Watch watches the given keys for changes
func (c *Client) Watch(ctx context.Context, fn func(*redis.Tx) error, keys ...string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	return c.client.Watch(ctx, fn, keys...)
}
