package main

import (
	"context"
	"fmt"
	"log"
	"time"

	mongoclient "github.com/isimtekin/go-packages/mongo-client"
	"go.mongodb.org/mongo-driver/mongo"
)

// User model example using BaseModel (WITH auto-timestamps)
// BaseModel includes createdAt and updatedAt that are automatically managed
type User struct {
	mongoclient.BaseModel `bson:",inline"`
	Email                 string   `bson:"email" json:"email"`
	Name                  string   `bson:"name" json:"name"`
	Age                   int      `bson:"age" json:"age"`
	Tags                  []string `bson:"tags,omitempty" json:"tags,omitempty"`
	Active                bool     `bson:"active" json:"active"`
}

// Product model example using SimpleModel (WITHOUT auto-timestamps)
// SimpleModel only has ID field, no automatic timestamps
type Product struct {
	mongoclient.SimpleModel `bson:",inline"`
	Name                    string  `bson:"name" json:"name"`
	Price                   float64 `bson:"price" json:"price"`
	SKU                     string  `bson:"sku" json:"sku"`
}

func main() {
	// Create context
	ctx := context.Background()

	// Example 1: Create client with options
	fmt.Println("=== Example 1: Creating Client ===")

	// Option A: From environment variables (recommended for production)
	// Uncomment to use:
	// client, err := mongoclient.NewFromEnvWithDefaults(ctx)

	// Option B: Programmatic configuration (for this example)
	client, err := mongoclient.NewWithOptions(ctx,
		mongoclient.WithURI("mongodb://localhost:27017"),
		mongoclient.WithDatabase("exampledb"),
		mongoclient.WithMaxPoolSize(50),
		mongoclient.WithConnectTimeout(10*time.Second),
	)
	if err != nil {
		log.Printf("Failed to connect (this is normal if MongoDB isn't running): %v", err)
		log.Println("To run examples, start MongoDB with: docker run -d -p 27017:27017 mongo:latest")
		log.Println("\nOr set environment variables:")
		log.Println("  export MONGO_URI=mongodb://localhost:27017")
		log.Println("  export MONGO_DATABASE=exampledb")
		return
	}
	defer client.Close(ctx)

	fmt.Println("Connected to MongoDB successfully!")

	// Example 2: Health check
	fmt.Println("\n=== Example 2: Health Check ===")
	if err := client.Health(ctx); err != nil {
		log.Fatalf("Health check failed: %v", err)
	}
	fmt.Println("Health check passed!")

	// Example 3: Get collection
	fmt.Println("\n=== Example 3: Working with Collections ===")
	users := client.Collection("users")
	fmt.Printf("Collection name: %s\n", users.Name())

	// Example 4: Insert documents (AUTO-TIMESTAMPS)
	fmt.Println("\n=== Example 4: Inserting Documents with Auto-Timestamps ===")

	// Create a new user - timestamps will be added automatically!
	newUser := &User{
		Email:  "john@example.com",
		Name:   "John Doe",
		Age:    30,
		Tags:   []string{"developer", "golang"},
		Active: true,
	}
	// NO need to call BeforeInsert() - timestamps are automatic!

	result, err := users.InsertOne(ctx, newUser)
	if err != nil {
		log.Printf("Insert failed: %v", err)
	} else {
		fmt.Printf("Inserted user with ID: %v\n", result.InsertedID)
		fmt.Printf("CreatedAt: %v, UpdatedAt: %v\n", newUser.CreatedAt, newUser.UpdatedAt)
	}

	// Insert multiple users - timestamps added automatically for each
	manyUsers := []interface{}{
		&User{Email: "jane@example.com", Name: "Jane Smith", Age: 28, Active: true},
		&User{Email: "bob@example.com", Name: "Bob Johnson", Age: 35, Active: true},
	}

	manyResult, err := users.InsertMany(ctx, manyUsers)
	if err != nil {
		log.Printf("Bulk insert failed: %v", err)
	} else {
		fmt.Printf("Inserted %d users\n", len(manyResult.InsertedIDs))
	}

	// Example 5: Find documents
	fmt.Println("\n=== Example 5: Finding Documents ===")

	// Find one user
	var foundUser User
	err = users.FindOne(ctx, mongoclient.M{"email": "john@example.com"}).Decode(&foundUser)
	if err != nil {
		if mongoclient.IsNoDocuments(err) {
			fmt.Println("User not found")
		} else {
			log.Printf("Find failed: %v", err)
		}
	} else {
		fmt.Printf("Found user: %s (%s)\n", foundUser.Name, foundUser.Email)
	}

	// Find all users
	var allUsers []User
	filter := mongoclient.M{"active": true}
	err = users.FindAll(ctx, filter, &allUsers)
	if err != nil {
		log.Printf("Find all failed: %v", err)
	} else {
		fmt.Printf("Found %d active users\n", len(allUsers))
		for _, u := range allUsers {
			fmt.Printf("  - %s (%d years old)\n", u.Name, u.Age)
		}
	}

	// Example 6: Update documents (AUTO-TIMESTAMPS)
	fmt.Println("\n=== Example 6: Updating Documents with Auto-Timestamps ===")

	updateResult, err := users.UpdateOne(ctx,
		mongoclient.M{"email": "john@example.com"},
		mongoclient.Set(mongoclient.M{"age": 31}),
		// updatedAt will be automatically added to the $set operation!
	)
	if err != nil {
		log.Printf("Update failed: %v", err)
	} else {
		fmt.Printf("Matched: %d, Modified: %d\n", updateResult.MatchedCount, updateResult.ModifiedCount)
		fmt.Println("Note: updatedAt was automatically set to current time!")
	}

	// Example 6b: Insert without timestamps (using SimpleModel)
	fmt.Println("\n=== Example 6b: Model WITHOUT Auto-Timestamps ===")

	products := client.Collection("products")

	newProduct := &Product{
		Name:  "Laptop",
		Price: 999.99,
		SKU:   "LAP-001",
	}

	productResult, err := products.InsertOne(ctx, newProduct)
	if err != nil {
		log.Printf("Product insert failed: %v", err)
	} else {
		fmt.Printf("Inserted product with ID: %v\n", productResult.InsertedID)
		fmt.Println("No timestamps were added (SimpleModel doesn't have them)")
	}

	// Example 7: Using query operators
	fmt.Println("\n=== Example 7: Advanced Queries ===")

	// Find users older than 29
	var olderUsers []User
	err = users.FindAll(ctx,
		mongoclient.M{"age": mongoclient.Gt(29)},
		&olderUsers,
	)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		fmt.Printf("Found %d users older than 29\n", len(olderUsers))
	}

	// Example 8: Aggregation
	fmt.Println("\n=== Example 8: Aggregation Pipeline ===")

	pipeline := mongoclient.A{
		mongoclient.Match(mongoclient.M{"active": true}),
		mongoclient.Group("$active", mongoclient.M{
			"count":  mongoclient.M{"$sum": 1},
			"avgAge": mongoclient.M{"$avg": "$age"},
		}),
	}

	type AggResult struct {
		ID     bool    `bson:"_id"`
		Count  int     `bson:"count"`
		AvgAge float64 `bson:"avgAge"`
	}

	var aggResults []AggResult
	err = users.AggregateAll(ctx, pipeline, &aggResults)
	if err != nil {
		log.Printf("Aggregation failed: %v", err)
	} else {
		for _, r := range aggResults {
			fmt.Printf("Active=%v: Count=%d, Avg Age=%.1f\n", r.ID, r.Count, r.AvgAge)
		}
	}

	// Example 9: Count documents
	fmt.Println("\n=== Example 9: Counting Documents ===")

	count, err := users.CountDocuments(ctx, mongoclient.M{"active": true})
	if err != nil {
		log.Printf("Count failed: %v", err)
	} else {
		fmt.Printf("Total active users: %d\n", count)
	}

	// Example 10: Delete documents
	fmt.Println("\n=== Example 10: Deleting Documents ===")

	deleteResult, err := users.DeleteOne(ctx, mongoclient.M{"email": "bob@example.com"})
	if err != nil {
		log.Printf("Delete failed: %v", err)
	} else {
		fmt.Printf("Deleted %d document(s)\n", deleteResult.DeletedCount)
	}

	// Example 11: Transactions
	fmt.Println("\n=== Example 11: Transactions ===")

	err = client.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
		// Multiple operations in a transaction
		_, err := users.InsertOne(sessCtx, &User{
			Email: "tx@example.com",
			Name:  "Transaction User",
			Age:   25,
		})
		if err != nil {
			return err
		}

		_, err = users.UpdateOne(sessCtx,
			mongoclient.M{"email": "jane@example.com"},
			mongoclient.Set(mongoclient.M{"age": 29}),
		)
		return err
	})
	if err != nil {
		log.Printf("Transaction failed: %v", err)
	} else {
		fmt.Println("Transaction completed successfully")
	}

	// Example 12: Pagination
	fmt.Println("\n=== Example 12: Pagination ===")

	pagination := &mongoclient.PaginationOptions{
		Page:     1,
		PageSize: 2,
	}
	pagination.Validate()

	cursor, err := users.Find(ctx, mongoclient.M{})
	if err == nil {
		defer cursor.Close(ctx)
		fmt.Printf("Page %d with %d items per page (skip: %d)\n",
			pagination.Page, pagination.PageSize, pagination.GetSkip())
	}

	// Example 13: Using helper methods
	fmt.Println("\n=== Example 13: Helper Methods ===")

	// Generate new ObjectID
	newID := mongoclient.NewObjectID()
	fmt.Printf("Generated ObjectID: %s\n", newID.Hex())

	// Validate ObjectID
	validID := "507f1f77bcf86cd799439011"
	fmt.Printf("Is %s valid? %v\n", validID, mongoclient.IsValidObjectID(validID))

	// Example 14: Find by ID
	fmt.Println("\n=== Example 14: Find by ID ===")

	if result.InsertedID != nil {
		var userByID User
		err = users.FindOneByID(ctx, result.InsertedID).Decode(&userByID)
		if err == nil {
			fmt.Printf("Found user by ID: %s\n", userByID.Name)
		}
	}

	// Cleanup
	fmt.Println("\n=== Cleanup ===")
	err = users.Drop(ctx)
	if err != nil {
		log.Printf("Failed to drop collection: %v", err)
	} else {
		fmt.Println("Collection dropped successfully")
	}

	fmt.Println("\nExamples completed!")
}
