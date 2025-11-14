package main

import (
	"context"
	"fmt"
	"log"
	"time"

	redisclient "github.com/isimtekin/go-packages/redis-client"
)

// This example demonstrates multi-database support in redis-client

func main() {
	ctx := context.Background()

	fmt.Println("=== Redis Multi-Database Example ===")
	fmt.Println()

	// ====================
	// Option 1: Using DBManager directly
	// ====================
	fmt.Println("Option 1: Using DBManager")
	fmt.Println("---------------------------")

	// Create a database manager with named databases
	manager, err := redisclient.NewDBManagerWithOptions(
		redisclient.WithAddr("localhost:6379"),
		redisclient.WithPoolSize(50),
		redisclient.WithDatabaseNames(map[string]int{
			"cache":   0,
			"session": 1,
			"user":    3,
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer manager.Close()

	// Get clients for different databases using configured names
	sessionDB, _ := manager.DB("session") // DB 1
	cacheDB, _ := manager.DB("cache")     // DB 0
	userDB, _ := manager.DB("user")       // DB 3

	// Use different databases for different purposes
	fmt.Println("✓ Created connections using configured database names")

	// Set data in session DB
	sessionDB.Set(ctx, "session:user123", "active", 1*time.Hour)
	fmt.Println("✓ Set session data in 'session' database")

	// Set data in cache DB
	cacheDB.Set(ctx, "cache:homepage", "<html>...</html>", 5*time.Minute)
	fmt.Println("✓ Set cache data in 'cache' database")

	// Set data in user DB
	userDB.HSet(ctx, "user:123", "name", "John Doe", "email", "john@example.com")
	fmt.Println("✓ Set user data in 'user' database")

	// Check active connections
	activeDatabases := manager.ActiveDBs()
	fmt.Printf("✓ Active databases: %v\n", activeDatabases)
	fmt.Println()

	// ====================
	// Option 2: Using DBClient wrapper (recommended)
	// ====================
	fmt.Println("Option 2: Using DBClient Wrapper (Recommended)")
	fmt.Println("-----------------------------------------------")

	// Create DBClient wrappers using configured names
	sessions, _ := manager.WithDB("session")
	cache, _ := manager.WithDB("cache")
	users, _ := manager.WithDB("user")

	// Now you can use them with a cleaner API
	sessions.Set(ctx, "session:user456", "active", 30*time.Minute)
	fmt.Printf("✓ Set session in DB%d ('session')\n", sessions.DBNum())

	cache.Set(ctx, "cache:products", "[]", 1*time.Minute)
	fmt.Printf("✓ Set cache in DB%d ('cache')\n", cache.DBNum())

	users.HSet(ctx, "user:456", "name", "Jane Smith")
	fmt.Printf("✓ Set user in DB%d ('user')\n", users.DBNum())
	fmt.Println()

	// ====================
	// Option 3: Using Global Singleton Manager
	// ====================
	fmt.Println("Option 3: Using Global Singleton Manager")
	fmt.Println("-----------------------------------------")

	// Initialize global manager (do this once at app startup)
	err = redisclient.InitGlobalManagerWithOptions(
		redisclient.WithAddr("localhost:6379"),
		redisclient.WithPoolSize(100),
	)
	if err != nil {
		// Already initialized, that's okay
		fmt.Println("✓ Global manager already initialized (this is fine)")
	}

	// Get global manager from anywhere in your app
	globalManager, _ := redisclient.GetGlobalManager()

	// Use it just like regular manager (by number since no names configured yet)
	globalSessionDB, _ := globalManager.DB(0)
	globalSessionDB.Set(ctx, "global:session", "value", 0)
	fmt.Println("✓ Used global singleton manager")
	fmt.Println()

	// ====================
	// Option 4: MustDB for Initialization Code
	// ====================
	fmt.Println("Option 4: MustDB for Initialization (Fail Fast)")
	fmt.Println("-----------------------------------------------")

	// Use MustDB with configured names in init code where you want to panic on errors
	sessionStore := manager.MustDB("session")
	cacheStore := manager.MustDB("cache")
	userStore := manager.MustDB("user")

	fmt.Println("✓ Created DB connections with MustDB (panics on error)")
	fmt.Printf("  Session DB: %p\n", sessionStore)
	fmt.Printf("  Cache DB: %p\n", cacheStore)
	fmt.Printf("  User DB: %p\n", userStore)
	fmt.Println()

	// ====================
	// Real-World Usage Example
	// ====================
	fmt.Println("Real-World Example: E-commerce Application")
	fmt.Println("-------------------------------------------")

	// Create manager for an e-commerce app with configured database names
	ecommerce, _ := redisclient.NewDBManagerWithOptions(
		redisclient.WithAddr("localhost:6379"),
		redisclient.WithDatabaseNames(map[string]int{
			"session":   0,
			"cache":     1,
			"cart":      2,
			"ratelimit": 3,
		}),
	)
	defer ecommerce.Close()

	// Use configured names for different data types
	sessionsClient := ecommerce.MustWithDB("session")   // User sessions
	productCache := ecommerce.MustWithDB("cache")       // Product cache
	cartDB := ecommerce.MustWithDB("cart")              // Shopping cart
	rateLimitDB := ecommerce.MustWithDB("ratelimit")    // Rate limiting

	// Simulate user activity
	userID := "user_12345"
	productID := "product_999"

	// 1. Create session
	sessionsClient.Set(ctx, fmt.Sprintf("session:%s", userID), "logged_in", 24*time.Hour)
	fmt.Printf("✓ Created session for %s in DB%d\n", userID, sessionsClient.DBNum())

	// 2. Cache product details
	productCache.HSet(ctx, productID,
		"name", "Awesome Product",
		"price", "99.99",
		"stock", "50",
	)
	fmt.Printf("✓ Cached product %s in DB%d\n", productID, productCache.DBNum())

	// 3. Add to cart
	cartDB.LPush(ctx, fmt.Sprintf("cart:%s", userID), productID)
	fmt.Printf("✓ Added %s to cart in DB%d\n", productID, cartDB.DBNum())

	// 4. Check rate limit
	key := fmt.Sprintf("ratelimit:%s", userID)
	count, _ := rateLimitDB.Incr(ctx, key)
	if count == 1 {
		// Set expiry on first request
		rateLimitDB.Client().Expire(ctx, key, 1*time.Minute)
	}
	fmt.Printf("✓ Rate limit check: %d requests in DB%d\n", count, rateLimitDB.DBNum())
	fmt.Println()

	// ====================
	// Benefits Summary
	// ====================
	fmt.Println("Benefits of Multi-Database Support:")
	fmt.Println("-----------------------------------")
	fmt.Println("✓ Data Isolation: Different data types in different DBs")
	fmt.Println("✓ Performance: Separate databases for different workloads")
	fmt.Println("✓ Easy Management: Flush one DB without affecting others")
	fmt.Println("✓ Configurable Names: Define custom friendly names for databases")
	fmt.Println("✓ Singleton Pattern: Reuses connections efficiently")
	fmt.Println("✓ Thread-Safe: All operations are thread-safe")
	fmt.Println("✓ Clean API: Both low-level and high-level interfaces")
	fmt.Println()

	// ====================
	// Connection Reuse
	// ====================
	fmt.Println("Connection Reuse (Singleton Pattern):")
	fmt.Println("--------------------------------------")

	// Getting the same DB multiple times returns the same client
	// Works with number or configured name - both return same instance
	client1, _ := manager.DB(0)
	client2, _ := manager.DB("cache")

	if client1 == client2 {
		fmt.Println("✓ Same DB connection is reused (singleton per DB)")
		fmt.Println("  DB(0) and DB(\"cache\") return the same instance")
	}

	// Different DBs return different clients
	client3, _ := manager.DB("session")
	if client1 != client3 {
		fmt.Println("✓ Different DBs have independent connections")
	}
	fmt.Println()

	fmt.Println("=== Example Complete ===")
}
