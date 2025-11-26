package postgresclient

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Client wraps PostgreSQL connection pool with convenient methods
type Client struct {
	pool   *pgxpool.Pool
	config *Config
}

// New creates a new PostgreSQL client with the given configuration
func New(ctx context.Context, config *Config) (*Client, error) {
	if config == nil {
		return nil, ErrNilConfig
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Build pool configuration
	poolConfig, err := pgxpool.ParseConfig(config.ConnectionURL())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection URL: %w", err)
	}

	// Apply pool settings
	if config.MaxConns > 0 {
		poolConfig.MaxConns = config.MaxConns
	}

	if config.MinConns > 0 {
		poolConfig.MinConns = config.MinConns
	}

	if config.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = config.MaxConnLifetime
	}

	if config.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = config.MaxConnIdleTime
	}

	if config.HealthCheckPeriod > 0 {
		poolConfig.HealthCheckPeriod = config.HealthCheckPeriod
	}

	// Create connection context with timeout
	connectCtx := ctx
	if config.ConnectTimeout > 0 {
		var cancel context.CancelFunc
		connectCtx, cancel = context.WithTimeout(ctx, config.ConnectTimeout)
		defer cancel()
	}

	// Connect to PostgreSQL
	pool, err := pgxpool.NewWithConfig(connectCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Ping to verify connection
	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	client := &Client{
		pool:   pool,
		config: config,
	}

	return client, nil
}

// NewWithOptions creates a new PostgreSQL client with functional options
func NewWithOptions(ctx context.Context, opts ...Option) (*Client, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}
	return New(ctx, config)
}

// Close closes the PostgreSQL connection pool
func (c *Client) Close() {
	if c.pool != nil {
		c.pool.Close()
	}
}

// Ping verifies the connection to PostgreSQL
func (c *Client) Ping(ctx context.Context) error {
	if c.pool == nil {
		return ErrClientNotConnected
	}
	return c.pool.Ping(ctx)
}

// Exec executes a query without returning any rows
func (c *Client) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if c.pool == nil {
		return pgconn.CommandTag{}, ErrClientNotConnected
	}

	ctx = c.withTimeout(ctx)
	return c.pool.Exec(ctx, sql, args...)
}

// Query executes a query that returns rows
func (c *Client) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if c.pool == nil {
		return nil, ErrClientNotConnected
	}

	ctx = c.withTimeout(ctx)
	return c.pool.Query(ctx, sql, args...)
}

// QueryRow executes a query that returns at most one row
func (c *Client) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	ctx = c.withTimeout(ctx)
	return c.pool.QueryRow(ctx, sql, args...)
}

// BeginTx starts a new transaction with the given options
func (c *Client) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	if c.pool == nil {
		return nil, ErrClientNotConnected
	}
	return c.pool.BeginTx(ctx, txOptions)
}

// Begin starts a new transaction with default options
func (c *Client) Begin(ctx context.Context) (pgx.Tx, error) {
	return c.BeginTx(ctx, pgx.TxOptions{})
}

// WithTransaction executes a function within a transaction
// The transaction is automatically committed if the function returns nil,
// or rolled back if the function returns an error
func (c *Client) WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return c.WithTransactionOptions(ctx, pgx.TxOptions{}, fn)
}

// WithTransactionOptions executes a function within a transaction with custom options
func (c *Client) WithTransactionOptions(ctx context.Context, txOptions pgx.TxOptions, fn func(tx pgx.Tx) error) error {
	tx, err := c.BeginTx(ctx, txOptions)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CopyFrom performs a bulk insert using PostgreSQL's COPY protocol
func (c *Client) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	if c.pool == nil {
		return 0, ErrClientNotConnected
	}
	return c.pool.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

// SendBatch sends a batch of queries to the server
func (c *Client) SendBatch(ctx context.Context, batch *pgx.Batch) pgx.BatchResults {
	ctx = c.withTimeout(ctx)
	return c.pool.SendBatch(ctx, batch)
}

// Health checks the health of the PostgreSQL connection
func (c *Client) Health(ctx context.Context) error {
	if c.pool == nil {
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
	if err := c.pool.Ping(healthCtx); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}

// Stats returns connection pool statistics
func (c *Client) Stats() *pgxpool.Stat {
	if c.pool == nil {
		return nil
	}
	return c.pool.Stat()
}

// Pool returns the underlying connection pool for advanced operations
func (c *Client) Pool() *pgxpool.Pool {
	return c.pool
}

// Config returns the client configuration
func (c *Client) Config() *Config {
	return c.config
}

// GetTimeout returns the appropriate timeout for operations
func (c *Client) GetTimeout() time.Duration {
	if c.config.QueryTimeout > 0 {
		return c.config.QueryTimeout
	}
	return 30 * time.Second // default timeout
}

// withTimeout adds timeout to context if not already set
func (c *Client) withTimeout(ctx context.Context) context.Context {
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx
	}

	if c.config.QueryTimeout > 0 {
		ctx, _ = context.WithTimeout(ctx, c.config.QueryTimeout)
	}

	return ctx
}

// AcquireConn acquires a connection from the pool
func (c *Client) AcquireConn(ctx context.Context) (*pgxpool.Conn, error) {
	if c.pool == nil {
		return nil, ErrClientNotConnected
	}
	return c.pool.Acquire(ctx)
}
