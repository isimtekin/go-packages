package main

import (
	"context"
	"fmt"
	"log"

	mongoclient "github.com/isimtekin/go-packages/mongo-client"
)

// This example shows how to load MongoDB configuration from environment variables

func main() {
	ctx := context.Background()

	fmt.Println("=== MongoDB Configuration from Environment Variables ===\n")

	// Example 1: Load with default MONGO_ prefix
	fmt.Println("Example 1: Using default MONGO_ prefix")
	fmt.Println("Environment variables:")
	fmt.Println("  MONGO_URI=mongodb://localhost:27017")
	fmt.Println("  MONGO_DATABASE=myapp")
	fmt.Println()

	client1, err := mongoclient.NewFromEnvWithDefaults(ctx)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		log.Println("This is normal if MongoDB isn't running or env vars aren't set")
		fmt.Println()
	} else {
		defer client1.Close(ctx)
		fmt.Println("✓ Client created successfully from environment!")
		fmt.Println()
	}

	// Example 2: Load with custom prefix
	fmt.Println("Example 2: Using custom DB_ prefix")
	fmt.Println("Environment variables:")
	fmt.Println("  DB_URI=mongodb://localhost:27017")
	fmt.Println("  DB_DATABASE=myapp")
	fmt.Println()

	client2, err := mongoclient.NewFromEnv(ctx, "DB_")
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		fmt.Println()
	} else {
		defer client2.Close(ctx)
		fmt.Println("✓ Client created successfully from DB_ prefix!")
		fmt.Println()
	}

	// Example 3: Load config without creating client
	fmt.Println("Example 3: Loading config only (no client connection)")
	fmt.Println("Environment variables:")
	fmt.Println("  MYAPP_HOST=localhost")
	fmt.Println("  MYAPP_PORT=27017")
	fmt.Println("  MYAPP_USERNAME=admin")
	fmt.Println("  MYAPP_PASSWORD=secret")
	fmt.Println("  MYAPP_DATABASE=myapp")
	fmt.Println()

	config, err := mongoclient.LoadConfigFromEnv("MYAPP_")
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		fmt.Println()
	} else {
		fmt.Printf("✓ Config loaded:\n")
		fmt.Printf("  URI: %s\n", maskPassword(config.URI))
		fmt.Printf("  Database: %s\n", config.Database)
		fmt.Printf("  Max Pool Size: %d\n", config.MaxPoolSize)
		fmt.Printf("  Connect Timeout: %v\n", config.ConnectTimeout)
		fmt.Println()
	}

	// Example 4: Building URI from components
	fmt.Println("Example 4: URI built from individual components")
	fmt.Println("When MONGO_URI is not set, it's built from:")
	fmt.Println("  MONGO_HOST=db.example.com")
	fmt.Println("  MONGO_PORT=27017")
	fmt.Println("  MONGO_USERNAME=myuser")
	fmt.Println("  MONGO_PASSWORD=mypass")
	fmt.Println("  MONGO_AUTH_SOURCE=admin")
	fmt.Println("  MONGO_DATABASE=production")
	fmt.Println()
	fmt.Println("Result: mongodb://myuser:****@db.example.com:27017/?authSource=admin")
	fmt.Println()

	// Example 5: Using .env file
	fmt.Println("Example 5: Loading from .env file")
	fmt.Println()
	fmt.Println("1. Copy .env.example to .env")
	fmt.Println("2. Update values in .env file")
	fmt.Println("3. Load it before creating client:")
	fmt.Println()
	fmt.Println("   import envutil \"github.com/isimtekin/go-packages/env-util\"")
	fmt.Println("   envutil.LoadEnvFile(\".env\")")
	fmt.Println("   client, err := mongoclient.NewFromEnvWithDefaults(ctx)")
	fmt.Println()

	// Example 6: Common patterns
	fmt.Println("Example 6: Common configuration patterns")
	fmt.Println()
	fmt.Println("Development (local):")
	fmt.Println("  MONGO_URI=mongodb://localhost:27017")
	fmt.Println("  MONGO_DATABASE=dev_db")
	fmt.Println()
	fmt.Println("Production (with auth):")
	fmt.Println("  MONGO_HOST=prod-mongodb.example.com")
	fmt.Println("  MONGO_PORT=27017")
	fmt.Println("  MONGO_USERNAME=prod_user")
	fmt.Println("  MONGO_PASSWORD=<secret>")
	fmt.Println("  MONGO_DATABASE=prod_db")
	fmt.Println("  MONGO_AUTH_SOURCE=admin")
	fmt.Println()
	fmt.Println("Docker Compose:")
	fmt.Println("  MONGO_URI=mongodb://mongo:27017")
	fmt.Println("  MONGO_DATABASE=app_db")
	fmt.Println()
	fmt.Println("MongoDB Atlas:")
	fmt.Println("  MONGO_URI=mongodb+srv://user:pass@cluster.mongodb.net/?retryWrites=true&w=majority")
	fmt.Println("  MONGO_DATABASE=atlas_db")
	fmt.Println()

	fmt.Println("=== Complete ===")
}

func maskPassword(uri string) string {
	// Simple password masking for display
	if len(uri) > 20 {
		return uri[:20] + "..."
	}
	return uri
}
