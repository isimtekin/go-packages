package mongoclient

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Client wraps MongoDB client with convenient methods
type Client struct {
	client   *mongo.Client
	config   *Config
	database *mongo.Database
}

// New creates a new MongoDB client with the given configuration
func New(ctx context.Context, config *Config) (*Client, error) {
	if config == nil {
		return nil, ErrNilConfig
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Build client options
	clientOpts := options.Client().ApplyURI(config.URI)

	if config.MaxPoolSize > 0 {
		clientOpts.SetMaxPoolSize(config.MaxPoolSize)
	}

	if config.MinPoolSize > 0 {
		clientOpts.SetMinPoolSize(config.MinPoolSize)
	}

	if config.MaxConnIdleTime > 0 {
		clientOpts.SetMaxConnIdleTime(config.MaxConnIdleTime)
	}

	if config.ConnectTimeout > 0 {
		clientOpts.SetConnectTimeout(config.ConnectTimeout)
	}

	if config.SocketTimeout > 0 {
		clientOpts.SetSocketTimeout(config.SocketTimeout)
	}

	if config.ServerSelectionTimeout > 0 {
		clientOpts.SetServerSelectionTimeout(config.ServerSelectionTimeout)
	}

	// Create context with timeout
	connectCtx := ctx
	if config.ConnectTimeout > 0 {
		var cancel context.CancelFunc
		connectCtx, cancel = context.WithTimeout(ctx, config.ConnectTimeout)
		defer cancel()
	}

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(connectCtx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	pingCtx := ctx
	if config.ConnectTimeout > 0 {
		var cancel context.CancelFunc
		pingCtx, cancel = context.WithTimeout(ctx, config.ConnectTimeout)
		defer cancel()
	}

	if err := mongoClient.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = mongoClient.Disconnect(ctx)
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	client := &Client{
		client:   mongoClient,
		config:   config,
		database: mongoClient.Database(config.Database),
	}

	return client, nil
}

// NewWithOptions creates a new MongoDB client with functional options
func NewWithOptions(ctx context.Context, opts ...Option) (*Client, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}
	return New(ctx, config)
}

// Close disconnects the MongoDB client
func (c *Client) Close(ctx context.Context) error {
	if c.client == nil {
		return nil
	}
	return c.client.Disconnect(ctx)
}

// Ping verifies the connection to MongoDB
func (c *Client) Ping(ctx context.Context) error {
	if c.client == nil {
		return ErrClientNotConnected
	}

	pingCtx := ctx
	if c.config.ConnectTimeout > 0 {
		var cancel context.CancelFunc
		pingCtx, cancel = context.WithTimeout(ctx, c.config.ConnectTimeout)
		defer cancel()
	}

	return c.client.Ping(pingCtx, readpref.Primary())
}

// Database returns the configured database
func (c *Client) Database() *mongo.Database {
	return c.database
}

// Collection returns a collection from the configured database
func (c *Client) Collection(name string) *Collection {
	return &Collection{
		collection: c.database.Collection(name),
		client:     c,
	}
}

// UseDatabase switches to a different database
func (c *Client) UseDatabase(name string) *mongo.Database {
	return c.client.Database(name)
}

// StartSession starts a new MongoDB session for transactions
func (c *Client) StartSession(opts ...*options.SessionOptions) (mongo.Session, error) {
	if c.client == nil {
		return nil, ErrClientNotConnected
	}
	return c.client.StartSession(opts...)
}

// WithTransaction executes a function within a transaction
func (c *Client) WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error, opts ...*options.TransactionOptions) error {
	session, err := c.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	// Execute the callback within a transaction
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	}, opts...)

	return err
}

// Health checks the health of the MongoDB connection
func (c *Client) Health(ctx context.Context) error {
	// Check if client exists
	if c.client == nil {
		return ErrClientNotConnected
	}

	// Create timeout context
	healthCtx := ctx
	if c.config.ConnectTimeout > 0 {
		var cancel context.CancelFunc
		healthCtx, cancel = context.WithTimeout(ctx, c.config.ConnectTimeout)
		defer cancel()
	}

	// Ping the database
	if err := c.client.Ping(healthCtx, readpref.Primary()); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}

// Stats returns connection pool statistics
func (c *Client) Stats() interface{} {
	// Note: MongoDB Go driver doesn't expose detailed pool stats
	// This is a placeholder for future implementation
	return map[string]interface{}{
		"connected": c.client != nil,
		"database":  c.config.Database,
	}
}

// Client returns the underlying MongoDB client for advanced operations
func (c *Client) Client() *mongo.Client {
	return c.client
}

// GetTimeout returns the appropriate timeout for the operation
func (c *Client) GetTimeout() time.Duration {
	if c.config.OperationTimeout > 0 {
		return c.config.OperationTimeout
	}
	return 30 * time.Second // default timeout
}
