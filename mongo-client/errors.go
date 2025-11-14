package mongoclient

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

// Common errors
var (
	ErrNilConfig          = errors.New("config cannot be nil")
	ErrClientNotConnected = errors.New("client is not connected")
	ErrNoDocuments        = mongo.ErrNoDocuments
	ErrInvalidID          = errors.New("invalid ID format")
	ErrEmptyFilter        = errors.New("filter cannot be empty")
)

// IsNoDocuments checks if the error is a "no documents" error
func IsNoDocuments(err error) bool {
	return errors.Is(err, mongo.ErrNoDocuments)
}

// IsDuplicateKey checks if the error is a duplicate key error
func IsDuplicateKey(err error) bool {
	var e mongo.WriteException
	if errors.As(err, &e) {
		for _, we := range e.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	}
	return false
}
