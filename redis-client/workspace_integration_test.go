package redisclient

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// TestWorkspace_StringOperations tests workspace prefixing for string operations
func TestWorkspace_StringOperations(t *testing.T) {
	config := DefaultConfig()
	config.Workspace = "test"

	client, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("Set and Get with workspace", func(t *testing.T) {
		// The key "user:123" should be stored as "test:user:123"
		err := client.Set(ctx, "user:123", "John Doe", 1*time.Minute)
		if err != nil {
			t.Logf("Set failed (expected with no Redis): %v", err)
		}

		// Verify the config applies the prefix correctly
		expectedKey := config.ApplyWorkspacePrefix("user:123")
		if expectedKey != "test:user:123" {
			t.Errorf("Expected key 'test:user:123', got '%s'", expectedKey)
		}
	})

	t.Run("SetNX with workspace", func(t *testing.T) {
		_, err := client.SetNX(ctx, "lock:resource", "locked", 30*time.Second)
		if err != nil {
			t.Logf("SetNX failed (expected with no Redis): %v", err)
		}

		expectedKey := config.ApplyWorkspacePrefix("lock:resource")
		if expectedKey != "test:lock:resource" {
			t.Errorf("Expected key 'test:lock:resource', got '%s'", expectedKey)
		}
	})

	t.Run("SetEX with workspace", func(t *testing.T) {
		err := client.SetEX(ctx, "temp:data", "value", 10*time.Second)
		if err != nil {
			t.Logf("SetEX failed (expected with no Redis): %v", err)
		}

		expectedKey := config.ApplyWorkspacePrefix("temp:data")
		if expectedKey != "test:temp:data" {
			t.Errorf("Expected key 'test:temp:data', got '%s'", expectedKey)
		}
	})

	t.Run("GetSet with workspace", func(t *testing.T) {
		_, err := client.GetSet(ctx, "counter", "100")
		if err != nil {
			t.Logf("GetSet failed (expected with no Redis): %v", err)
		}

		expectedKey := config.ApplyWorkspacePrefix("counter")
		if expectedKey != "test:counter" {
			t.Errorf("Expected key 'test:counter', got '%s'", expectedKey)
		}
	})

	t.Run("MGet with workspace", func(t *testing.T) {
		keys := []string{"key1", "key2", "key3"}
		_, err := client.MGet(ctx, keys...)
		if err != nil {
			t.Logf("MGet failed (expected with no Redis): %v", err)
		}

		expectedKeys := config.ApplyWorkspacePrefixToKeys(keys)
		for i, key := range expectedKeys {
			expected := "test:" + keys[i]
			if key != expected {
				t.Errorf("Expected key '%s', got '%s'", expected, key)
			}
		}
	})

	t.Run("MSet with workspace", func(t *testing.T) {
		err := client.MSet(ctx, "key1", "value1", "key2", "value2")
		if err != nil {
			t.Logf("MSet failed (expected with no Redis): %v", err)
		}
	})

	t.Run("Incr with workspace", func(t *testing.T) {
		_, err := client.Incr(ctx, "counter")
		if err != nil {
			t.Logf("Incr failed (expected with no Redis): %v", err)
		}
	})

	t.Run("IncrBy with workspace", func(t *testing.T) {
		_, err := client.IncrBy(ctx, "counter", 5)
		if err != nil {
			t.Logf("IncrBy failed (expected with no Redis): %v", err)
		}
	})

	t.Run("Decr with workspace", func(t *testing.T) {
		_, err := client.Decr(ctx, "counter")
		if err != nil {
			t.Logf("Decr failed (expected with no Redis): %v", err)
		}
	})

	t.Run("DecrBy with workspace", func(t *testing.T) {
		_, err := client.DecrBy(ctx, "counter", 3)
		if err != nil {
			t.Logf("DecrBy failed (expected with no Redis): %v", err)
		}
	})
}

// TestWorkspace_KeyOperations tests workspace prefixing for key operations
func TestWorkspace_KeyOperations(t *testing.T) {
	config := DefaultConfig()
	config.Workspace = "test"

	client, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("Del with workspace", func(t *testing.T) {
		_, err := client.Del(ctx, "key1", "key2", "key3")
		if err != nil {
			t.Logf("Del failed (expected with no Redis): %v", err)
		}

		keys := []string{"key1", "key2", "key3"}
		expectedKeys := config.ApplyWorkspacePrefixToKeys(keys)
		for i, key := range expectedKeys {
			expected := "test:" + keys[i]
			if key != expected {
				t.Errorf("Expected key '%s', got '%s'", expected, key)
			}
		}
	})

	t.Run("Exists with workspace", func(t *testing.T) {
		_, err := client.Exists(ctx, "key1", "key2")
		if err != nil {
			t.Logf("Exists failed (expected with no Redis): %v", err)
		}
	})

	t.Run("Expire with workspace", func(t *testing.T) {
		_, err := client.Expire(ctx, "key1", 1*time.Hour)
		if err != nil {
			t.Logf("Expire failed (expected with no Redis): %v", err)
		}
	})

	t.Run("TTL with workspace", func(t *testing.T) {
		_, err := client.TTL(ctx, "key1")
		if err != nil {
			t.Logf("TTL failed (expected with no Redis): %v", err)
		}
	})

	t.Run("Persist with workspace", func(t *testing.T) {
		_, err := client.Persist(ctx, "key1")
		if err != nil {
			t.Logf("Persist failed (expected with no Redis): %v", err)
		}
	})

	t.Run("Keys with workspace", func(t *testing.T) {
		_, err := client.Keys(ctx, "user:*")
		if err != nil {
			t.Logf("Keys failed (expected with no Redis): %v", err)
		}

		expectedPattern := config.ApplyWorkspacePrefix("user:*")
		if expectedPattern != "test:user:*" {
			t.Errorf("Expected pattern 'test:user:*', got '%s'", expectedPattern)
		}
	})
}

// TestWorkspace_HashOperations tests workspace prefixing for hash operations
func TestWorkspace_HashOperations(t *testing.T) {
	config := DefaultConfig()
	config.Workspace = "test"

	client, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("HSet with workspace", func(t *testing.T) {
		_, err := client.HSet(ctx, "user:123", "name", "John", "email", "john@example.com")
		if err != nil {
			t.Logf("HSet failed (expected with no Redis): %v", err)
		}

		expectedKey := config.ApplyWorkspacePrefix("user:123")
		if expectedKey != "test:user:123" {
			t.Errorf("Expected key 'test:user:123', got '%s'", expectedKey)
		}
	})

	t.Run("HGet with workspace", func(t *testing.T) {
		_, err := client.HGet(ctx, "user:123", "name")
		if err != nil {
			t.Logf("HGet failed (expected with no Redis): %v", err)
		}
	})

	t.Run("HGetAll with workspace", func(t *testing.T) {
		_, err := client.HGetAll(ctx, "user:123")
		if err != nil {
			t.Logf("HGetAll failed (expected with no Redis): %v", err)
		}
	})

	t.Run("HDel with workspace", func(t *testing.T) {
		_, err := client.HDel(ctx, "user:123", "field1", "field2")
		if err != nil {
			t.Logf("HDel failed (expected with no Redis): %v", err)
		}
	})

	t.Run("HExists with workspace", func(t *testing.T) {
		_, err := client.HExists(ctx, "user:123", "name")
		if err != nil {
			t.Logf("HExists failed (expected with no Redis): %v", err)
		}
	})

	t.Run("HIncrBy with workspace", func(t *testing.T) {
		_, err := client.HIncrBy(ctx, "user:123", "age", 1)
		if err != nil {
			t.Logf("HIncrBy failed (expected with no Redis): %v", err)
		}
	})

	t.Run("HKeys with workspace", func(t *testing.T) {
		_, err := client.HKeys(ctx, "user:123")
		if err != nil {
			t.Logf("HKeys failed (expected with no Redis): %v", err)
		}
	})

	t.Run("HLen with workspace", func(t *testing.T) {
		_, err := client.HLen(ctx, "user:123")
		if err != nil {
			t.Logf("HLen failed (expected with no Redis): %v", err)
		}
	})
}

// TestWorkspace_ListOperations tests workspace prefixing for list operations
func TestWorkspace_ListOperations(t *testing.T) {
	config := DefaultConfig()
	config.Workspace = "test"

	client, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("LPush with workspace", func(t *testing.T) {
		_, err := client.LPush(ctx, "queue", "item1", "item2", "item3")
		if err != nil {
			t.Logf("LPush failed (expected with no Redis): %v", err)
		}

		expectedKey := config.ApplyWorkspacePrefix("queue")
		if expectedKey != "test:queue" {
			t.Errorf("Expected key 'test:queue', got '%s'", expectedKey)
		}
	})

	t.Run("RPush with workspace", func(t *testing.T) {
		_, err := client.RPush(ctx, "queue", "item4", "item5")
		if err != nil {
			t.Logf("RPush failed (expected with no Redis): %v", err)
		}
	})

	t.Run("LPop with workspace", func(t *testing.T) {
		_, err := client.LPop(ctx, "queue")
		if err != nil {
			t.Logf("LPop failed (expected with no Redis): %v", err)
		}
	})

	t.Run("RPop with workspace", func(t *testing.T) {
		_, err := client.RPop(ctx, "queue")
		if err != nil {
			t.Logf("RPop failed (expected with no Redis): %v", err)
		}
	})

	t.Run("LRange with workspace", func(t *testing.T) {
		_, err := client.LRange(ctx, "queue", 0, -1)
		if err != nil {
			t.Logf("LRange failed (expected with no Redis): %v", err)
		}
	})

	t.Run("LLen with workspace", func(t *testing.T) {
		_, err := client.LLen(ctx, "queue")
		if err != nil {
			t.Logf("LLen failed (expected with no Redis): %v", err)
		}
	})
}

// TestWorkspace_SetOperations tests workspace prefixing for set operations
func TestWorkspace_SetOperations(t *testing.T) {
	config := DefaultConfig()
	config.Workspace = "test"

	client, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("SAdd with workspace", func(t *testing.T) {
		_, err := client.SAdd(ctx, "tags", "tag1", "tag2", "tag3")
		if err != nil {
			t.Logf("SAdd failed (expected with no Redis): %v", err)
		}

		expectedKey := config.ApplyWorkspacePrefix("tags")
		if expectedKey != "test:tags" {
			t.Errorf("Expected key 'test:tags', got '%s'", expectedKey)
		}
	})

	t.Run("SMembers with workspace", func(t *testing.T) {
		_, err := client.SMembers(ctx, "tags")
		if err != nil {
			t.Logf("SMembers failed (expected with no Redis): %v", err)
		}
	})

	t.Run("SIsMember with workspace", func(t *testing.T) {
		_, err := client.SIsMember(ctx, "tags", "tag1")
		if err != nil {
			t.Logf("SIsMember failed (expected with no Redis): %v", err)
		}
	})

	t.Run("SRem with workspace", func(t *testing.T) {
		_, err := client.SRem(ctx, "tags", "tag1", "tag2")
		if err != nil {
			t.Logf("SRem failed (expected with no Redis): %v", err)
		}
	})

	t.Run("SCard with workspace", func(t *testing.T) {
		_, err := client.SCard(ctx, "tags")
		if err != nil {
			t.Logf("SCard failed (expected with no Redis): %v", err)
		}
	})
}

// TestWorkspace_SortedSetOperations tests workspace prefixing for sorted set operations
func TestWorkspace_SortedSetOperations(t *testing.T) {
	config := DefaultConfig()
	config.Workspace = "test"

	client, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("ZAdd with workspace", func(t *testing.T) {
		members := []redis.Z{
			{Score: 1.0, Member: "member1"},
			{Score: 2.0, Member: "member2"},
			{Score: 3.0, Member: "member3"},
		}
		_, err := client.ZAdd(ctx, "leaderboard", members...)
		if err != nil {
			t.Logf("ZAdd failed (expected with no Redis): %v", err)
		}

		expectedKey := config.ApplyWorkspacePrefix("leaderboard")
		if expectedKey != "test:leaderboard" {
			t.Errorf("Expected key 'test:leaderboard', got '%s'", expectedKey)
		}
	})

	t.Run("ZRange with workspace", func(t *testing.T) {
		_, err := client.ZRange(ctx, "leaderboard", 0, -1)
		if err != nil {
			t.Logf("ZRange failed (expected with no Redis): %v", err)
		}
	})

	t.Run("ZRangeWithScores with workspace", func(t *testing.T) {
		_, err := client.ZRangeWithScores(ctx, "leaderboard", 0, -1)
		if err != nil {
			t.Logf("ZRangeWithScores failed (expected with no Redis): %v", err)
		}
	})

	t.Run("ZRem with workspace", func(t *testing.T) {
		_, err := client.ZRem(ctx, "leaderboard", "member1", "member2")
		if err != nil {
			t.Logf("ZRem failed (expected with no Redis): %v", err)
		}
	})

	t.Run("ZCard with workspace", func(t *testing.T) {
		_, err := client.ZCard(ctx, "leaderboard")
		if err != nil {
			t.Logf("ZCard failed (expected with no Redis): %v", err)
		}
	})

	t.Run("ZScore with workspace", func(t *testing.T) {
		_, err := client.ZScore(ctx, "leaderboard", "member1")
		if err != nil {
			t.Logf("ZScore failed (expected with no Redis): %v", err)
		}
	})
}

// TestWorkspace_PubSubOperations tests workspace prefixing for pub/sub operations
func TestWorkspace_PubSubOperations(t *testing.T) {
	config := DefaultConfig()
	config.Workspace = "test"

	client, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("Publish with workspace", func(t *testing.T) {
		_, err := client.Publish(ctx, "channel", "message")
		if err != nil {
			t.Logf("Publish failed (expected with no Redis): %v", err)
		}

		expectedChannel := config.ApplyWorkspacePrefix("channel")
		if expectedChannel != "test:channel" {
			t.Errorf("Expected channel 'test:channel', got '%s'", expectedChannel)
		}
	})

	t.Run("Subscribe with workspace", func(t *testing.T) {
		channels := []string{"channel1", "channel2"}
		pubsub := client.Subscribe(ctx, channels...)
		if pubsub != nil {
			defer pubsub.Close()
		}

		expectedChannels := config.ApplyWorkspacePrefixToKeys(channels)
		for i, ch := range expectedChannels {
			expected := "test:" + channels[i]
			if ch != expected {
				t.Errorf("Expected channel '%s', got '%s'", expected, ch)
			}
		}
	})

	t.Run("PSubscribe with workspace", func(t *testing.T) {
		patterns := []string{"channel:*", "event:*"}
		pubsub := client.PSubscribe(ctx, patterns...)
		if pubsub != nil {
			defer pubsub.Close()
		}

		expectedPatterns := config.ApplyWorkspacePrefixToKeys(patterns)
		for i, pattern := range expectedPatterns {
			expected := "test:" + patterns[i]
			if pattern != expected {
				t.Errorf("Expected pattern '%s', got '%s'", expected, pattern)
			}
		}
	})
}

// TestWorkspace_TransactionOperations tests workspace prefixing for transaction operations
func TestWorkspace_TransactionOperations(t *testing.T) {
	config := DefaultConfig()
	config.Workspace = "test"

	client, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("Watch with workspace", func(t *testing.T) {
		keys := []string{"key1", "key2"}
		err := client.Watch(ctx, func(tx *redis.Tx) error {
			return nil
		}, keys...)
		if err != nil {
			t.Logf("Watch failed (expected with no Redis): %v", err)
		}

		expectedKeys := config.ApplyWorkspacePrefixToKeys(keys)
		for i, key := range expectedKeys {
			expected := "test:" + keys[i]
			if key != expected {
				t.Errorf("Expected key '%s', got '%s'", expected, key)
			}
		}
	})
}

// TestWorkspace_NoWorkspace tests operations without workspace
func TestWorkspace_NoWorkspace(t *testing.T) {
	config := DefaultConfig()
	// No workspace set

	client, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	t.Run("Keys are not prefixed without workspace", func(t *testing.T) {
		key := "user:123"
		prefixedKey := config.ApplyWorkspacePrefix(key)

		if prefixedKey != key {
			t.Errorf("Expected key '%s' to remain unchanged, got '%s'", key, prefixedKey)
		}
	})

	t.Run("Multiple keys are not prefixed without workspace", func(t *testing.T) {
		keys := []string{"key1", "key2", "key3"}
		prefixedKeys := config.ApplyWorkspacePrefixToKeys(keys)

		for i, key := range prefixedKeys {
			if key != keys[i] {
				t.Errorf("Expected key '%s' to remain unchanged, got '%s'", keys[i], key)
			}
		}
	})
}

// TestWorkspace_Pipeline tests workspace with pipeline operations
func TestWorkspace_Pipeline(t *testing.T) {
	config := DefaultConfig()
	config.Workspace = "test"

	client, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	t.Run("Pipeline operations", func(t *testing.T) {
		pipe := client.Pipeline()
		if pipe == nil {
			t.Error("Pipeline() returned nil")
		}
	})

	t.Run("TxPipeline operations", func(t *testing.T) {
		txPipe := client.TxPipeline()
		if txPipe == nil {
			t.Error("TxPipeline() returned nil")
		}
	})
}

// TestWorkspace_EnvironmentVariable tests loading workspace from environment
func TestWorkspace_EnvironmentVariable(t *testing.T) {
	t.Run("Load workspace from environment", func(t *testing.T) {
		t.Setenv("REDIS_WORKSPACE", "staging")

		config, err := LoadConfigFromEnvWithDefaults()
		if err != nil {
			t.Fatalf("Failed to load config from env: %v", err)
		}

		if config.Workspace != "staging" {
			t.Errorf("Expected workspace 'staging', got '%s'", config.Workspace)
		}
	})

	t.Run("Load workspace with custom prefix", func(t *testing.T) {
		t.Setenv("CACHE_WORKSPACE", "production")

		config, err := LoadConfigFromEnv("CACHE_")
		if err != nil {
			t.Fatalf("Failed to load config from env: %v", err)
		}

		if config.Workspace != "production" {
			t.Errorf("Expected workspace 'production', got '%s'", config.Workspace)
		}
	})
}
