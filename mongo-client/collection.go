package mongoclient

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collection wraps mongo.Collection with convenient methods
type Collection struct {
	collection *mongo.Collection
	client     *Client
}

// FindOneResult wraps the result of a FindOne operation
type FindOneResult struct {
	result *mongo.SingleResult
}

// Decode decodes the result into the provided value
func (r *FindOneResult) Decode(v interface{}) error {
	return r.result.Decode(v)
}

// Err returns any error from the operation
func (r *FindOneResult) Err() error {
	return r.result.Err()
}

// InsertOneResult wraps the result of an InsertOne operation
type InsertOneResult struct {
	InsertedID interface{}
}

// InsertManyResult wraps the result of an InsertMany operation
type InsertManyResult struct {
	InsertedIDs []interface{}
}

// UpdateResult wraps the result of an Update operation
type UpdateResult struct {
	MatchedCount  int64
	ModifiedCount int64
	UpsertedCount int64
	UpsertedID    interface{}
}

// DeleteResult wraps the result of a Delete operation
type DeleteResult struct {
	DeletedCount int64
}

// FindOne finds a single document matching the filter
func (c *Collection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *FindOneResult {
	opCtx, cancel := c.createOperationContext(ctx)
	defer cancel()

	result := c.collection.FindOne(opCtx, filter, opts...)
	return &FindOneResult{result: result}
}

// FindOneByID finds a document by its ID
func (c *Collection) FindOneByID(ctx context.Context, id interface{}, opts ...*options.FindOneOptions) *FindOneResult {
	objID, err := c.toObjectID(id)
	if err != nil {
		// Return an error result
		return &FindOneResult{result: &mongo.SingleResult{}}
	}

	return c.FindOne(ctx, bson.M{"_id": objID}, opts...)
}

// Find finds all documents matching the filter
func (c *Collection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	opCtx, cancel := c.createOperationContext(ctx)
	defer cancel()

	return c.collection.Find(opCtx, filter, opts...)
}

// FindAll finds all documents and decodes them into the results slice
func (c *Collection) FindAll(ctx context.Context, filter interface{}, results interface{}, opts ...*options.FindOptions) error {
	cursor, err := c.Find(ctx, filter, opts...)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	return cursor.All(ctx, results)
}

// InsertOne inserts a single document
// Automatically sets createdAt and updatedAt if the document implements Timestamped interface
func (c *Collection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*InsertOneResult, error) {
	// Apply timestamps automatically if document implements Timestamped
	applyTimestamps(document, true)

	opCtx, cancel := c.createOperationContext(ctx)
	defer cancel()

	result, err := c.collection.InsertOne(opCtx, document, opts...)
	if err != nil {
		return nil, err
	}

	return &InsertOneResult{InsertedID: result.InsertedID}, nil
}

// InsertMany inserts multiple documents
// Automatically sets createdAt and updatedAt for each document that implements Timestamped interface
func (c *Collection) InsertMany(ctx context.Context, documents []interface{}, opts ...*options.InsertManyOptions) (*InsertManyResult, error) {
	// Apply timestamps automatically to all documents that implement Timestamped
	for _, doc := range documents {
		applyTimestamps(doc, true)
	}

	opCtx, cancel := c.createOperationContext(ctx)
	defer cancel()

	result, err := c.collection.InsertMany(opCtx, documents, opts...)
	if err != nil {
		return nil, err
	}

	return &InsertManyResult{InsertedIDs: result.InsertedIDs}, nil
}

// UpdateOne updates a single document matching the filter
// Automatically adds updatedAt to the update if using $set operator
func (c *Collection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*UpdateResult, error) {
	// Automatically add updatedAt to $set operations
	update = c.addUpdatedAtToUpdate(update)

	opCtx, cancel := c.createOperationContext(ctx)
	defer cancel()

	result, err := c.collection.UpdateOne(opCtx, filter, update, opts...)
	if err != nil {
		return nil, err
	}

	return &UpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: result.UpsertedCount,
		UpsertedID:    result.UpsertedID,
	}, nil
}

// UpdateOneByID updates a document by its ID
func (c *Collection) UpdateOneByID(ctx context.Context, id interface{}, update interface{}, opts ...*options.UpdateOptions) (*UpdateResult, error) {
	objID, err := c.toObjectID(id)
	if err != nil {
		return nil, err
	}

	return c.UpdateOne(ctx, bson.M{"_id": objID}, update, opts...)
}

// UpdateMany updates all documents matching the filter
// Automatically adds updatedAt to the update if using $set operator
func (c *Collection) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*UpdateResult, error) {
	// Automatically add updatedAt to $set operations
	update = c.addUpdatedAtToUpdate(update)

	opCtx, cancel := c.createOperationContext(ctx)
	defer cancel()

	result, err := c.collection.UpdateMany(opCtx, filter, update, opts...)
	if err != nil {
		return nil, err
	}

	return &UpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: result.UpsertedCount,
		UpsertedID:    result.UpsertedID,
	}, nil
}

// ReplaceOne replaces a single document matching the filter
func (c *Collection) ReplaceOne(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.ReplaceOptions) (*UpdateResult, error) {
	opCtx, cancel := c.createOperationContext(ctx)
	defer cancel()

	result, err := c.collection.ReplaceOne(opCtx, filter, replacement, opts...)
	if err != nil {
		return nil, err
	}

	return &UpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: result.UpsertedCount,
		UpsertedID:    result.UpsertedID,
	}, nil
}

// DeleteOne deletes a single document matching the filter
func (c *Collection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*DeleteResult, error) {
	opCtx, cancel := c.createOperationContext(ctx)
	defer cancel()

	result, err := c.collection.DeleteOne(opCtx, filter, opts...)
	if err != nil {
		return nil, err
	}

	return &DeleteResult{DeletedCount: result.DeletedCount}, nil
}

// DeleteOneByID deletes a document by its ID
func (c *Collection) DeleteOneByID(ctx context.Context, id interface{}, opts ...*options.DeleteOptions) (*DeleteResult, error) {
	objID, err := c.toObjectID(id)
	if err != nil {
		return nil, err
	}

	return c.DeleteOne(ctx, bson.M{"_id": objID}, opts...)
}

// DeleteMany deletes all documents matching the filter
func (c *Collection) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*DeleteResult, error) {
	opCtx, cancel := c.createOperationContext(ctx)
	defer cancel()

	result, err := c.collection.DeleteMany(opCtx, filter, opts...)
	if err != nil {
		return nil, err
	}

	return &DeleteResult{DeletedCount: result.DeletedCount}, nil
}

// CountDocuments counts documents matching the filter
func (c *Collection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	opCtx, cancel := c.createOperationContext(ctx)
	defer cancel()

	return c.collection.CountDocuments(opCtx, filter, opts...)
}

// Aggregate executes an aggregation pipeline
func (c *Collection) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	opCtx, cancel := c.createOperationContext(ctx)
	defer cancel()

	return c.collection.Aggregate(opCtx, pipeline, opts...)
}

// AggregateOne executes an aggregation pipeline and returns a single result
func (c *Collection) AggregateOne(ctx context.Context, pipeline interface{}, result interface{}, opts ...*options.AggregateOptions) error {
	cursor, err := c.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return ErrNoDocuments
	}

	return cursor.Decode(result)
}

// AggregateAll executes an aggregation pipeline and decodes all results
func (c *Collection) AggregateAll(ctx context.Context, pipeline interface{}, results interface{}, opts ...*options.AggregateOptions) error {
	cursor, err := c.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	return cursor.All(ctx, results)
}

// Distinct gets distinct values for a field
func (c *Collection) Distinct(ctx context.Context, fieldName string, filter interface{}, opts ...*options.DistinctOptions) ([]interface{}, error) {
	opCtx, cancel := c.createOperationContext(ctx)
	defer cancel()

	return c.collection.Distinct(opCtx, fieldName, filter, opts...)
}

// CreateIndex creates a new index
func (c *Collection) CreateIndex(ctx context.Context, keys interface{}, opts ...*options.IndexOptions) (string, error) {
	opCtx, cancel := c.createOperationContext(ctx)
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: nil,
	}

	if len(opts) > 0 {
		indexModel.Options = opts[0]
	}

	return c.collection.Indexes().CreateOne(opCtx, indexModel)
}

// CreateIndexes creates multiple indexes
func (c *Collection) CreateIndexes(ctx context.Context, models []mongo.IndexModel, opts ...*options.CreateIndexesOptions) ([]string, error) {
	opCtx, cancel := c.createOperationContext(ctx)
	defer cancel()

	return c.collection.Indexes().CreateMany(opCtx, models, opts...)
}

// DropIndex drops an index by name
func (c *Collection) DropIndex(ctx context.Context, name string, opts ...*options.DropIndexesOptions) error {
	opCtx, cancel := c.createOperationContext(ctx)
	defer cancel()

	_, err := c.collection.Indexes().DropOne(opCtx, name, opts...)
	return err
}

// Drop drops the entire collection
func (c *Collection) Drop(ctx context.Context) error {
	opCtx, cancel := c.createOperationContext(ctx)
	defer cancel()

	return c.collection.Drop(opCtx)
}

// BulkWrite performs multiple write operations
func (c *Collection) BulkWrite(ctx context.Context, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	opCtx, cancel := c.createOperationContext(ctx)
	defer cancel()

	return c.collection.BulkWrite(opCtx, models, opts...)
}

// Name returns the collection name
func (c *Collection) Name() string {
	return c.collection.Name()
}

// Collection returns the underlying mongo.Collection for advanced operations
func (c *Collection) Collection() *mongo.Collection {
	return c.collection
}

// createOperationContext creates a context with timeout for operations
func (c *Collection) createOperationContext(ctx context.Context) (context.Context, context.CancelFunc) {
	timeout := c.client.GetTimeout()
	return context.WithTimeout(ctx, timeout)
}

// toObjectID converts various types to primitive.ObjectID
func (c *Collection) toObjectID(id interface{}) (primitive.ObjectID, error) {
	switch v := id.(type) {
	case primitive.ObjectID:
		return v, nil
	case string:
		return primitive.ObjectIDFromHex(v)
	default:
		return primitive.NilObjectID, fmt.Errorf("%w: %T", ErrInvalidID, id)
	}
}

// addUpdatedAtToUpdate automatically adds updatedAt field to $set operations
func (c *Collection) addUpdatedAtToUpdate(update interface{}) interface{} {
	if updateMap, ok := update.(bson.M); ok {
		// Check if $set exists
		if setOp, hasSet := updateMap["$set"]; hasSet {
			if setMap, ok := setOp.(bson.M); ok {
				// Only add updatedAt if it doesn't exist already
				if _, hasUpdatedAt := setMap["updatedAt"]; !hasUpdatedAt {
					setMap["updatedAt"] = primitive.NewDateTimeFromTime(time.Now())
				}
			}
		} else {
			// If no operators, this might be a direct update (not recommended but supported)
			// Add updatedAt directly
			if _, hasUpdatedAt := updateMap["updatedAt"]; !hasUpdatedAt {
				updateMap["updatedAt"] = primitive.NewDateTimeFromTime(time.Now())
			}
		}
	}
	return update
}
