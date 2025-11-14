package redisclient

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DBManager manages multiple Redis database connections as a singleton
type DBManager struct {
	config        *Config
	clients       map[int]*Client // map of DB number to Client
	databaseNames map[string]int  // map of name to DB number
	mu            sync.RWMutex
	closed        bool
}

var (
	managerInstance *DBManager
	managerOnce     sync.Once
)

// NewDBManager creates a new database manager with the given configuration
// The config's DB field will be ignored as we'll manage multiple DBs
func NewDBManager(config *Config) (*DBManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Copy database names from config
	databaseNames := make(map[string]int)
	if config.DatabaseNames != nil {
		for name, dbNum := range config.DatabaseNames {
			databaseNames[name] = dbNum
		}
	}

	manager := &DBManager{
		config:        config,
		clients:       make(map[int]*Client),
		databaseNames: databaseNames,
		closed:        false,
	}

	return manager, nil
}

// NewDBManagerWithOptions creates a new database manager with functional options
func NewDBManagerWithOptions(opts ...Option) (*DBManager, error) {
	config := DefaultConfig()

	for _, opt := range opts {
		opt(config)
	}

	return NewDBManager(config)
}

// GetGlobalManager returns the global singleton DBManager instance
// If not initialized, it creates one with default configuration
func GetGlobalManager() (*DBManager, error) {
	var err error
	managerOnce.Do(func() {
		managerInstance, err = NewDBManager(DefaultConfig())
	})
	return managerInstance, err
}

// InitGlobalManager initializes the global DBManager with custom configuration
// Must be called before GetGlobalManager if you want custom config
// Returns error if already initialized
func InitGlobalManager(config *Config) error {
	if managerInstance != nil {
		return fmt.Errorf("global manager already initialized")
	}

	var err error
	managerOnce.Do(func() {
		managerInstance, err = NewDBManager(config)
	})

	return err
}

// InitGlobalManagerWithOptions initializes the global DBManager with functional options
func InitGlobalManagerWithOptions(opts ...Option) error {
	if managerInstance != nil {
		return fmt.Errorf("global manager already initialized")
	}

	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	var err error
	managerOnce.Do(func() {
		managerInstance, err = NewDBManager(config)
	})

	return err
}

// DB returns a client for the specified database
// Accepts either int (database number) or string (database name from config)
// Creates a new connection if it doesn't exist
//
// Examples:
//   manager.DB(0)        // By number
//   manager.DB("cache")  // By name (if configured)
func (m *DBManager) DB(identifier DBIdentifier) (*Client, error) {
	dbNum, ok := m.resolveDBNumber(identifier)
	if !ok {
		return nil, fmt.Errorf("invalid database identifier: %v (not found in configuration)", identifier)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil, ErrClientClosed
	}

	// Check if client already exists
	if client, exists := m.clients[dbNum]; exists {
		return client, nil
	}

	// Create new client for this database
	config := m.cloneConfigWithDB(dbNum)
	client, err := New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for DB %v: %w", identifier, err)
	}

	m.clients[dbNum] = client
	return client, nil
}

// resolveDBNumber converts a DBIdentifier to a database number using this manager's config
func (m *DBManager) resolveDBNumber(id DBIdentifier) (int, bool) {
	switch v := id.(type) {
	case int:
		return v, true
	case string:
		if num, ok := m.databaseNames[v]; ok {
			return num, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// MustDB returns a client for the specified database or panics on error
// Accepts either int (database number) or string (database name from config)
// Useful for initialization code where you want to fail fast
//
// Examples:
//   manager.MustDB(0)        // By number
//   manager.MustDB("cache")  // By name (if configured)
func (m *DBManager) MustDB(identifier DBIdentifier) *Client {
	client, err := m.DB(identifier)
	if err != nil {
		panic(fmt.Sprintf("failed to get DB %v: %v", identifier, err))
	}
	return client
}

// Close closes all database connections
func (m *DBManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return ErrAlreadyClosed
	}

	var errs []error
	for dbNum, client := range m.clients {
		if err := client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("DB %d: %w", dbNum, err))
		}
	}

	m.clients = make(map[int]*Client)
	m.closed = true

	if len(errs) > 0 {
		return fmt.Errorf("errors closing clients: %v", errs)
	}

	return nil
}

// Ping pings all connected databases
func (m *DBManager) Ping(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return ErrClientClosed
	}

	for dbNum, client := range m.clients {
		if err := client.Ping(ctx); err != nil {
			return fmt.Errorf("DB %d: %w", dbNum, err)
		}
	}

	return nil
}

// ActiveDBs returns a list of database numbers that have active connections
func (m *DBManager) ActiveDBs() []int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dbs := make([]int, 0, len(m.clients))
	for dbNum := range m.clients {
		dbs = append(dbs, dbNum)
	}

	return dbs
}

// cloneConfigWithDB creates a copy of the config with a specific DB number
func (m *DBManager) cloneConfigWithDB(dbNum int) *Config {
	config := *m.config // shallow copy
	config.DB = dbNum
	return &config
}

// ====================
// Convenience Methods with DB Selection
// ====================

// DBClient wraps a client with a specific database context
type DBClient struct {
	client *Client
	dbNum  int
}

// WithDB creates a DBClient wrapper for convenient operations on a specific database
// Accepts either int (database number) or string (database name from config)
//
// Examples:
//   manager.WithDB(0)        // By number
//   manager.WithDB("cache")  // By name (if configured)
func (m *DBManager) WithDB(identifier DBIdentifier) (*DBClient, error) {
	dbNum, ok := m.resolveDBNumber(identifier)
	if !ok {
		return nil, fmt.Errorf("invalid database identifier: %v (not found in configuration)", identifier)
	}

	client, err := m.DB(identifier)
	if err != nil {
		return nil, err
	}

	return &DBClient{
		client: client,
		dbNum:  dbNum,
	}, nil
}

// MustWithDB creates a DBClient wrapper or panics on error
// Accepts either int (database number) or string (database name from config)
//
// Examples:
//   manager.MustWithDB(0)        // By number
//   manager.MustWithDB("cache")  // By name (if configured)
func (m *DBManager) MustWithDB(identifier DBIdentifier) *DBClient {
	dc, err := m.WithDB(identifier)
	if err != nil {
		panic(fmt.Sprintf("failed to create DBClient for DB %v: %v", identifier, err))
	}
	return dc
}

// Client returns the underlying Redis client
func (dc *DBClient) Client() *Client {
	return dc.client
}

// DBNum returns the database number this client is connected to
func (dc *DBClient) DBNum() int {
	return dc.dbNum
}

// All Client methods are available through the DBClient wrapper
// This provides a convenient way to work with a specific database

// Get retrieves the value of a key
func (dc *DBClient) Get(ctx context.Context, key string) (string, error) {
	return dc.client.Get(ctx, key)
}

// Set sets the value of a key with optional TTL
func (dc *DBClient) Set(ctx context.Context, key string, value interface{}, ttl ...time.Duration) error {
	var duration time.Duration
	if len(ttl) > 0 {
		duration = ttl[0]
	}
	return dc.client.Set(ctx, key, value, duration)
}

// Del deletes one or more keys
func (dc *DBClient) Del(ctx context.Context, keys ...string) (int64, error) {
	return dc.client.Del(ctx, keys...)
}

// Exists checks if keys exist
func (dc *DBClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return dc.client.Exists(ctx, keys...)
}

// HSet sets hash field
func (dc *DBClient) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return dc.client.HSet(ctx, key, values...)
}

// HGet gets hash field value
func (dc *DBClient) HGet(ctx context.Context, key, field string) (string, error) {
	return dc.client.HGet(ctx, key, field)
}

// HGetAll gets all hash fields and values
func (dc *DBClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return dc.client.HGetAll(ctx, key)
}

// LPush pushes to list head
func (dc *DBClient) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return dc.client.LPush(ctx, key, values...)
}

// RPush pushes to list tail
func (dc *DBClient) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return dc.client.RPush(ctx, key, values...)
}

// LRange gets list range
func (dc *DBClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return dc.client.LRange(ctx, key, start, stop)
}

// SAdd adds members to set
func (dc *DBClient) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return dc.client.SAdd(ctx, key, members...)
}

// SMembers gets all set members
func (dc *DBClient) SMembers(ctx context.Context, key string) ([]string, error) {
	return dc.client.SMembers(ctx, key)
}

// Incr increments a key
func (dc *DBClient) Incr(ctx context.Context, key string) (int64, error) {
	return dc.client.Incr(ctx, key)
}

// Decr decrements a key
func (dc *DBClient) Decr(ctx context.Context, key string) (int64, error) {
	return dc.client.Decr(ctx, key)
}
