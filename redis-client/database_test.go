package redisclient

import (
	"testing"
)

// TestDBManager_WithDatabaseNames tests the config-based database name registration
func TestDBManager_WithDatabaseNames(t *testing.T) {
	t.Run("manager with database names configured", func(t *testing.T) {
		manager, err := NewDBManagerWithOptions(
			WithAddr("localhost:6379"),
			WithDatabaseNames(map[string]int{
				"cache":   0,
				"session": 1,
				"queue":   2,
			}),
		)
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}
		defer manager.Close()

		// Should be able to access by name
		cache, err := manager.DB("cache")
		if err != nil {
			t.Errorf("DB(\"cache\") error = %v, want nil", err)
		}
		if cache == nil {
			t.Error("DB(\"cache\") returned nil client")
		}

		session, err := manager.DB("session")
		if err != nil {
			t.Errorf("DB(\"session\") error = %v, want nil", err)
		}
		if session == nil {
			t.Error("DB(\"session\") returned nil client")
		}
	})

	t.Run("access by number still works", func(t *testing.T) {
		manager, _ := NewDBManagerWithOptions(
			WithAddr("localhost:6379"),
			WithDatabaseNames(map[string]int{
				"cache": 0,
			}),
		)
		defer manager.Close()

		client, err := manager.DB(0)
		if err != nil {
			t.Errorf("DB(0) error = %v, want nil", err)
		}
		if client == nil {
			t.Error("DB(0) returned nil client")
		}
	})

	t.Run("same name returns same client (singleton)", func(t *testing.T) {
		manager, _ := NewDBManagerWithOptions(
			WithAddr("localhost:6379"),
			WithDatabaseNames(map[string]int{
				"cache": 0,
			}),
		)
		defer manager.Close()

		client1, _ := manager.DB("cache")
		client2, _ := manager.DB("cache")
		client3, _ := manager.DB(0)

		if client1 != client2 {
			t.Error("DB(\"cache\") should return same client")
		}
		if client1 != client3 {
			t.Error("DB(\"cache\") and DB(0) should return same client")
		}
	})

	t.Run("unregistered name returns error", func(t *testing.T) {
		manager, _ := NewDBManagerWithOptions(
			WithAddr("localhost:6379"),
			WithDatabaseNames(map[string]int{
				"cache": 0,
			}),
		)
		defer manager.Close()

		_, err := manager.DB("nonexistent")
		if err == nil {
			t.Error("DB(\"nonexistent\") should return error")
		}
	})

	t.Run("different names return different clients", func(t *testing.T) {
		manager, _ := NewDBManagerWithOptions(
			WithAddr("localhost:6379"),
			WithDatabaseNames(map[string]int{
				"cache":   0,
				"session": 1,
			}),
		)
		defer manager.Close()

		cache, _ := manager.DB("cache")
		session, _ := manager.DB("session")

		if cache == session {
			t.Error("Different database names should return different clients")
		}
	})
}

// TestDBManager_WithDatabaseName tests adding database names individually
func TestDBManager_WithDatabaseName(t *testing.T) {
	t.Run("add names one by one", func(t *testing.T) {
		manager, err := NewDBManagerWithOptions(
			WithAddr("localhost:6379"),
			WithDatabaseName("cache", 0),
			WithDatabaseName("session", 1),
			WithDatabaseName("queue", 2),
		)
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}
		defer manager.Close()

		cache, err := manager.DB("cache")
		if err != nil {
			t.Errorf("DB(\"cache\") error = %v, want nil", err)
		}
		if cache == nil {
			t.Error("DB(\"cache\") returned nil")
		}

		session, err := manager.DB("session")
		if err != nil {
			t.Errorf("DB(\"session\") error = %v, want nil", err)
		}
		if session == nil {
			t.Error("DB(\"session\") returned nil")
		}
	})
}

// TestDBManager_WithDBConfigBased tests WithDB() method with config-based names
func TestDBManager_WithDBConfigBased(t *testing.T) {
	manager, _ := NewDBManagerWithOptions(
		WithAddr("localhost:6379"),
		WithDatabaseNames(map[string]int{
			"cache":   0,
			"session": 1,
			"queue":   2,
		}),
	)
	defer manager.Close()

	t.Run("WithDB with name", func(t *testing.T) {
		dbClient, err := manager.WithDB("cache")
		if err != nil {
			t.Errorf("WithDB(\"cache\") error = %v, want nil", err)
		}
		if dbClient == nil {
			t.Error("WithDB(\"cache\") returned nil")
		}
		if dbClient.DBNum() != 0 {
			t.Errorf("DBNum() = %d, want 0", dbClient.DBNum())
		}
	})

	t.Run("WithDB with number", func(t *testing.T) {
		dbClient, err := manager.WithDB(1)
		if err != nil {
			t.Errorf("WithDB(1) error = %v, want nil", err)
		}
		if dbClient.DBNum() != 1 {
			t.Errorf("DBNum() = %d, want 1", dbClient.DBNum())
		}
	})

	t.Run("WithDB with invalid name", func(t *testing.T) {
		_, err := manager.WithDB("invalid")
		if err == nil {
			t.Error("WithDB(\"invalid\") should return error")
		}
	})
}

// TestDBManager_MustWithDBConfigBased tests MustWithDB() method
func TestDBManager_MustWithDBConfigBased(t *testing.T) {
	manager, _ := NewDBManagerWithOptions(
		WithAddr("localhost:6379"),
		WithDatabaseNames(map[string]int{
			"cache": 0,
		}),
	)
	defer manager.Close()

	t.Run("MustWithDB with name", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustWithDB(\"cache\") panicked: %v", r)
			}
		}()

		dbClient := manager.MustWithDB("cache")
		if dbClient == nil {
			t.Error("MustWithDB(\"cache\") returned nil")
		}
		if dbClient.DBNum() != 0 {
			t.Errorf("DBNum() = %d, want 0", dbClient.DBNum())
		}
	})

	t.Run("MustWithDB with invalid name panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustWithDB(\"invalid\") should panic")
			}
		}()

		manager.MustWithDB("invalid")
	})
}

// TestDBManager_MustDBConfigBased tests MustDB() method with config-based names
func TestDBManager_MustDBConfigBased(t *testing.T) {
	manager, _ := NewDBManagerWithOptions(
		WithAddr("localhost:6379"),
		WithDatabaseNames(map[string]int{
			"cache": 0,
		}),
	)
	defer manager.Close()

	t.Run("MustDB with name does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustDB(\"cache\") panicked: %v", r)
			}
		}()

		client := manager.MustDB("cache")
		if client == nil {
			t.Error("MustDB(\"cache\") returned nil")
		}
	})

	t.Run("MustDB with invalid name panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustDB(\"invalid\") should panic")
			}
		}()

		manager.MustDB("invalid")
	})
}

// TestDBManager_NoDatabaseNamesConfigured tests manager without database names
func TestDBManager_NoDatabaseNamesConfigured(t *testing.T) {
	manager, _ := NewDBManagerWithOptions(
		WithAddr("localhost:6379"),
	)
	defer manager.Close()

	t.Run("can access by number without names configured", func(t *testing.T) {
		client, err := manager.DB(0)
		if err != nil {
			t.Errorf("DB(0) error = %v, want nil", err)
		}
		if client == nil {
			t.Error("DB(0) returned nil")
		}
	})

	t.Run("cannot access by name when not configured", func(t *testing.T) {
		_, err := manager.DB("cache")
		if err == nil {
			t.Error("DB(\"cache\") should return error when names not configured")
		}
	})
}
