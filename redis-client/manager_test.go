package redisclient

import (
	"context"
	"sync"
	"testing"
)

// TestNewDBManager tests creating a new database manager
func TestNewDBManager(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		manager, err := NewDBManager(DefaultConfig())
		if err != nil {
			t.Errorf("NewDBManager() error = %v, want nil", err)
		}
		if manager == nil {
			t.Error("NewDBManager() returned nil manager")
		}
		if manager.clients == nil {
			t.Error("NewDBManager() clients map is nil")
		}
		if manager.closed {
			t.Error("NewDBManager() created closed manager")
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		invalidConfig := &Config{
			Addr: "", // invalid
		}
		_, err := NewDBManager(invalidConfig)
		if err == nil {
			t.Error("NewDBManager() with invalid config should return error")
		}
	})
}

// TestNewDBManagerWithOptions tests creating manager with options
func TestNewDBManagerWithOptions(t *testing.T) {
	manager, err := NewDBManagerWithOptions(
		WithAddr("redis:6379"),
		WithPassword("secret"),
		WithPoolSize(50),
	)

	if err != nil {
		t.Errorf("NewDBManagerWithOptions() error = %v, want nil", err)
	}

	if manager.config.Addr != "redis:6379" {
		t.Errorf("Addr = %v, want redis:6379", manager.config.Addr)
	}

	if manager.config.Password != "secret" {
		t.Errorf("Password = %v, want secret", manager.config.Password)
	}

	if manager.config.PoolSize != 50 {
		t.Errorf("PoolSize = %v, want 50", manager.config.PoolSize)
	}
}

// TestDBManager_DB tests getting database clients
func TestDBManager_DB(t *testing.T) {
	manager, err := NewDBManager(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	t.Run("get DB 0", func(t *testing.T) {
		client, err := manager.DB(0)
		if err != nil {
			t.Errorf("DB(0) error = %v, want nil", err)
		}
		if client == nil {
			t.Error("DB(0) returned nil client")
		}
	})

	t.Run("get DB 1", func(t *testing.T) {
		client, err := manager.DB(1)
		if err != nil {
			t.Errorf("DB(1) error = %v, want nil", err)
		}
		if client == nil {
			t.Error("DB(1) returned nil client")
		}
	})

	t.Run("get same DB twice returns same client", func(t *testing.T) {
		client1, err1 := manager.DB(2)
		if err1 != nil {
			t.Fatalf("First DB(2) error = %v", err1)
		}

		client2, err2 := manager.DB(2)
		if err2 != nil {
			t.Fatalf("Second DB(2) error = %v", err2)
		}

		if client1 != client2 {
			t.Error("DB(2) returned different clients on multiple calls")
		}
	})

	t.Run("multiple databases are independent", func(t *testing.T) {
		client0, _ := manager.DB(0)
		client1, _ := manager.DB(1)
		client2, _ := manager.DB(2)

		if client0 == client1 {
			t.Error("DB(0) and DB(1) returned same client")
		}

		if client0 == client2 {
			t.Error("DB(0) and DB(2) returned same client")
		}

		if client1 == client2 {
			t.Error("DB(1) and DB(2) returned same client")
		}
	})
}

// TestDBManager_MustDB tests MustDB panic behavior
func TestDBManager_MustDB(t *testing.T) {
	manager, err := NewDBManager(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	t.Run("successful call does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustDB() panicked: %v", r)
			}
		}()

		client := manager.MustDB(0)
		if client == nil {
			t.Error("MustDB() returned nil client")
		}
	})

	t.Run("closed manager panics", func(t *testing.T) {
		closedManager, _ := NewDBManager(DefaultConfig())
		closedManager.Close()

		defer func() {
			if r := recover(); r == nil {
				t.Error("MustDB() on closed manager should panic")
			}
		}()

		closedManager.MustDB(0)
	})
}

// TestDBManager_ActiveDBs tests getting list of active databases
func TestDBManager_ActiveDBs(t *testing.T) {
	manager, err := NewDBManager(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	t.Run("initially empty", func(t *testing.T) {
		dbs := manager.ActiveDBs()
		if len(dbs) != 0 {
			t.Errorf("ActiveDBs() = %v, want empty slice", dbs)
		}
	})

	t.Run("after creating connections", func(t *testing.T) {
		manager.DB(0)
		manager.DB(2)
		manager.DB(5)

		dbs := manager.ActiveDBs()
		if len(dbs) != 3 {
			t.Errorf("ActiveDBs() length = %v, want 3", len(dbs))
		}

		// Check all expected DBs are present
		dbMap := make(map[int]bool)
		for _, db := range dbs {
			dbMap[db] = true
		}

		if !dbMap[0] || !dbMap[2] || !dbMap[5] {
			t.Errorf("ActiveDBs() = %v, want [0, 2, 5]", dbs)
		}
	})
}

// TestDBManager_Close tests closing the manager
func TestDBManager_Close(t *testing.T) {
	t.Run("close empty manager", func(t *testing.T) {
		manager, _ := NewDBManager(DefaultConfig())

		err := manager.Close()
		if err != nil {
			t.Errorf("Close() error = %v, want nil", err)
		}

		if !manager.closed {
			t.Error("Close() did not set closed flag")
		}
	})

	t.Run("close manager with connections", func(t *testing.T) {
		manager, _ := NewDBManager(DefaultConfig())
		manager.DB(0)
		manager.DB(1)
		manager.DB(2)

		err := manager.Close()
		if err != nil {
			t.Errorf("Close() error = %v, want nil", err)
		}

		if !manager.closed {
			t.Error("Close() did not set closed flag")
		}

		if len(manager.clients) != 0 {
			t.Error("Close() did not clear clients map")
		}
	})

	t.Run("close already closed manager", func(t *testing.T) {
		manager, _ := NewDBManager(DefaultConfig())
		manager.Close()

		err := manager.Close()
		if err != ErrAlreadyClosed {
			t.Errorf("Close() on closed manager error = %v, want ErrAlreadyClosed", err)
		}
	})

	t.Run("operations on closed manager fail", func(t *testing.T) {
		manager, _ := NewDBManager(DefaultConfig())
		manager.Close()

		_, err := manager.DB(0)
		if err != ErrClientClosed {
			t.Errorf("DB() on closed manager error = %v, want ErrClientClosed", err)
		}

		err = manager.Ping(context.Background())
		if err != ErrClientClosed {
			t.Errorf("Ping() on closed manager error = %v, want ErrClientClosed", err)
		}
	})
}

// TestDBManager_Ping tests pinging all databases
func TestDBManager_Ping(t *testing.T) {
	manager, err := NewDBManager(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()

	t.Run("ping with no connections", func(t *testing.T) {
		err := manager.Ping(ctx)
		if err != nil {
			t.Errorf("Ping() with no connections error = %v, want nil", err)
		}
	})

	t.Run("ping on closed manager", func(t *testing.T) {
		closedManager, _ := NewDBManager(DefaultConfig())
		closedManager.Close()

		err := closedManager.Ping(ctx)
		if err != ErrClientClosed {
			t.Errorf("Ping() on closed manager error = %v, want ErrClientClosed", err)
		}
	})
}

// TestDBManager_WithDB tests the DBClient wrapper
func TestDBManager_WithDB(t *testing.T) {
	manager, err := NewDBManager(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	t.Run("create DBClient", func(t *testing.T) {
		dbClient, err := manager.WithDB(0)
		if err != nil {
			t.Errorf("WithDB(0) error = %v, want nil", err)
		}

		if dbClient == nil {
			t.Error("WithDB(0) returned nil")
		}

		if dbClient.DBNum() != 0 {
			t.Errorf("DBNum() = %v, want 0", dbClient.DBNum())
		}

		if dbClient.Client() == nil {
			t.Error("Client() returned nil")
		}
	})

	t.Run("DBClient operations", func(t *testing.T) {
		dbClient, _ := manager.WithDB(1)
		ctx := context.Background()

		// Test that operations don't panic (won't actually work without Redis)
		// This is just to verify the API is correct
		_ = dbClient.Set(ctx, "test", "value")
		_, _ = dbClient.Get(ctx, "test")
		_, _ = dbClient.Del(ctx, "test")
		_, _ = dbClient.Exists(ctx, "test")
		_, _ = dbClient.Incr(ctx, "counter")
		_, _ = dbClient.Decr(ctx, "counter")
	})
}

// TestDBManager_MustWithDB tests MustWithDB panic behavior
func TestDBManager_MustWithDB(t *testing.T) {
	manager, err := NewDBManager(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	t.Run("successful call does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustWithDB() panicked: %v", r)
			}
		}()

		dbClient := manager.MustWithDB(0)
		if dbClient == nil {
			t.Error("MustWithDB() returned nil")
		}
	})

	t.Run("closed manager panics", func(t *testing.T) {
		closedManager, _ := NewDBManager(DefaultConfig())
		closedManager.Close()

		defer func() {
			if r := recover(); r == nil {
				t.Error("MustWithDB() on closed manager should panic")
			}
		}()

		closedManager.MustWithDB(0)
	})
}

// TestDBManager_cloneConfigWithDB tests config cloning
func TestDBManager_cloneConfigWithDB(t *testing.T) {
	originalConfig := DefaultConfig()
	originalConfig.Addr = "test:6379"
	originalConfig.Password = "secret"
	originalConfig.DB = 0

	manager, _ := NewDBManager(originalConfig)
	defer manager.Close()

	cloned := manager.cloneConfigWithDB(5)

	if cloned.DB != 5 {
		t.Errorf("cloned config DB = %v, want 5", cloned.DB)
	}

	if cloned.Addr != originalConfig.Addr {
		t.Error("cloned config should preserve Addr")
	}

	if cloned.Password != originalConfig.Password {
		t.Error("cloned config should preserve Password")
	}

	// Modify cloned config should not affect original
	cloned.Addr = "modified:6379"
	if manager.config.Addr == "modified:6379" {
		t.Error("modifying cloned config affected original")
	}
}

// TestGlobalManager tests the singleton global manager
func TestGlobalManager(t *testing.T) {
	// Note: This test might interfere with other tests if they use the global manager
	// In production, you'd want to test this in isolation

	t.Run("GetGlobalManager creates instance", func(t *testing.T) {
		// Reset the singleton for testing (not thread-safe, for testing only)
		managerInstance = nil
		managerOnce = sync.Once{}

		manager, err := GetGlobalManager()
		if err != nil {
			t.Errorf("GetGlobalManager() error = %v, want nil", err)
		}

		if manager == nil {
			t.Error("GetGlobalManager() returned nil")
		}
	})

	t.Run("GetGlobalManager returns same instance", func(t *testing.T) {
		manager1, _ := GetGlobalManager()
		manager2, _ := GetGlobalManager()

		if manager1 != manager2 {
			t.Error("GetGlobalManager() returned different instances")
		}
	})
}

// TestInitGlobalManager tests initializing global manager with custom config
func TestInitGlobalManager(t *testing.T) {
	// Reset singleton for testing
	managerInstance = nil
	managerOnce = sync.Once{}

	customConfig := DefaultConfig()
	customConfig.Addr = "custom:6379"

	err := InitGlobalManager(customConfig)
	if err != nil {
		t.Errorf("InitGlobalManager() error = %v, want nil", err)
	}

	manager, _ := GetGlobalManager()
	if manager.config.Addr != "custom:6379" {
		t.Error("Global manager does not have custom config")
	}

	// Reset for other tests
	managerInstance = nil
	managerOnce = sync.Once{}
}

// TestInitGlobalManagerWithOptions tests initializing with options
func TestInitGlobalManagerWithOptions(t *testing.T) {
	// Reset singleton for testing
	managerInstance = nil
	managerOnce = sync.Once{}

	err := InitGlobalManagerWithOptions(
		WithAddr("options:6379"),
		WithPassword("test123"),
	)

	if err != nil {
		t.Errorf("InitGlobalManagerWithOptions() error = %v, want nil", err)
	}

	manager, _ := GetGlobalManager()
	if manager.config.Addr != "options:6379" {
		t.Error("Global manager does not have options config")
	}

	// Reset for other tests
	managerInstance = nil
	managerOnce = sync.Once{}
}
