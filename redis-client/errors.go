package redisclient

import (
	"errors"

	"github.com/redis/go-redis/v9"
)

var (
	// ErrClientClosed is returned when operating on a closed client
	ErrClientClosed = errors.New("redis client is closed")

	// ErrAlreadyClosed is returned when closing an already closed client
	ErrAlreadyClosed = errors.New("redis client is already closed")

	// ErrNil is returned when key doesn't exist (alias for redis.Nil)
	ErrNil = redis.Nil

	// ErrInvalidTTL is returned when TTL value is invalid
	ErrInvalidTTL = errors.New("invalid TTL value")

	// ErrInvalidKey is returned when key is empty or invalid
	ErrInvalidKey = errors.New("invalid key")
)

// IsNil returns true if the error is redis.Nil (key doesn't exist)
func IsNil(err error) bool {
	return errors.Is(err, redis.Nil)
}

// IsConnectionError returns true if the error is connection related
func IsConnectionError(err error) bool {
	return errors.Is(err, ErrClientClosed)
}
