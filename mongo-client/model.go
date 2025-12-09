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

// Model represents a MongoDB model with schema validation (like Mongoose)
// All CRUD operations go through the Model, which enforces schema validation
type Model struct {
	collection *Collection
	schema     *Schema
	client     *Client
}

// NewModel creates a new model with schema validation
// This is the Mongoose-like way to interact with MongoDB
func NewModel(client *Client, collectionName string, schema *Schema) (*Model, error) {
	if client == nil {
		return nil, ErrClientNotConnected
	}
	if schema == nil {
		return nil, ErrSchemaRequired
	}

	schema.CollectionName = collectionName

	return &Model{
		collection: client.Collection(collectionName),
		schema:     schema,
		client:     client,
	}, nil
}

// Schema returns the model's schema
func (m *Model) Schema() *Schema {
	return m.schema
}

// Collection returns the underlying collection
func (m *Model) Collection() *Collection {
	return m.collection
}

// CollectionName returns the collection name
func (m *Model) CollectionName() string {
	return m.schema.CollectionName
}

// EnsureIndexes creates all indexes defined in the schema
func (m *Model) EnsureIndexes(ctx context.Context) error {
	for _, idx := range m.schema.Indexes {
		indexOpts := options.Index()
		if idx.Unique {
			indexOpts.SetUnique(true)
		}
		if idx.Name != "" {
			indexOpts.SetName(idx.Name)
		}

		_, err := m.collection.CreateIndex(ctx, idx.Keys, indexOpts)
		if err != nil {
			return fmt.Errorf("failed to create index %s: %w", idx.Name, err)
		}
	}
	return nil
}

// InsertOne inserts a single document with schema validation
// Automatically sets _id if not provided
// Automatically sets createdAt and updatedAt if schema has timestamps enabled
func (m *Model) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*InsertOneResult, error) {
	// Apply defaults from schema
	if err := m.schema.ApplyDefaults(document); err != nil {
		// Defaults might fail for non-struct types, that's okay
		_ = err
	}

	// Validate document against schema
	if err := m.schema.Validate(document); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Apply ObjectID if document implements IDSetter and ID is not set
	applyObjectID(document)

	// Apply timestamps if schema has timestamps enabled
	if m.schema.Timestamps {
		applyTimestamps(document, true)
	}

	// Execute pre-insert hooks
	hc := &HookContext{
		Operation: OpInsert,
		Document:  document,
		Schema:    m.schema,
		Model:     m,
	}
	if err := m.schema.executePreHooks(ctx, OpInsert, hc); err != nil {
		return nil, err
	}

	// Execute the operation
	result, err := m.collection.InsertOne(ctx, document, opts...)

	// Execute post-insert hooks
	hc.Result = result
	hc.Error = err
	_ = m.schema.executePostHooks(ctx, OpInsert, hc)

	return result, err
}

// InsertMany inserts multiple documents with schema validation
// Automatically sets _id for each document if not provided
// Automatically sets createdAt and updatedAt if schema has timestamps enabled
func (m *Model) InsertMany(ctx context.Context, documents []interface{}, opts ...*options.InsertManyOptions) (*InsertManyResult, error) {
	// Validate and prepare each document
	for i, doc := range documents {
		// Apply defaults from schema
		_ = m.schema.ApplyDefaults(doc)

		// Validate document against schema
		if err := m.schema.Validate(doc); err != nil {
			return nil, fmt.Errorf("validation failed for document %d: %w", i, err)
		}

		// Apply ObjectID
		applyObjectID(doc)

		// Apply timestamps
		if m.schema.Timestamps {
			applyTimestamps(doc, true)
		}
	}

	// Execute pre-insertMany hooks
	hc := &HookContext{
		Operation: OpInsertMany,
		Documents: documents,
		Schema:    m.schema,
		Model:     m,
	}
	if err := m.schema.executePreHooks(ctx, OpInsertMany, hc); err != nil {
		return nil, err
	}

	// Execute the operation
	result, err := m.collection.InsertMany(ctx, documents, opts...)

	// Execute post-insertMany hooks
	hc.Result = result
	hc.Error = err
	_ = m.schema.executePostHooks(ctx, OpInsertMany, hc)

	return result, err
}

// FindOne finds a single document matching the filter
func (m *Model) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *FindOneResult {
	// Convert filter to M for hooks
	filterM := toM(filter)

	// Execute pre-findOne hooks
	hc := &HookContext{
		Operation: OpFindOne,
		Filter:    filterM,
		Schema:    m.schema,
		Model:     m,
	}
	if err := m.schema.executePreHooks(ctx, OpFindOne, hc); err != nil {
		// Return a result that will error on decode
		return &FindOneResult{err: err}
	}

	// Execute the operation
	result := m.collection.FindOne(ctx, filter, opts...)

	// Execute post-findOne hooks (async by default)
	hc.Result = result
	_ = m.schema.executePostHooks(ctx, OpFindOne, hc)

	return result
}

// FindOneByID finds a document by its ID
func (m *Model) FindOneByID(ctx context.Context, id interface{}, opts ...*options.FindOneOptions) *FindOneResult {
	objID, err := toObjectID(id)
	if err != nil {
		return &FindOneResult{err: err}
	}
	return m.FindOne(ctx, bson.M{"_id": objID}, opts...)
}

// Find finds all documents matching the filter
func (m *Model) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	// Convert filter to M for hooks
	filterM := toM(filter)

	// Execute pre-find hooks
	hc := &HookContext{
		Operation: OpFind,
		Filter:    filterM,
		Schema:    m.schema,
		Model:     m,
	}
	if err := m.schema.executePreHooks(ctx, OpFind, hc); err != nil {
		return nil, err
	}

	// Execute the operation
	cursor, err := m.collection.Find(ctx, filter, opts...)

	// Execute post-find hooks
	hc.Result = cursor
	hc.Error = err
	_ = m.schema.executePostHooks(ctx, OpFind, hc)

	return cursor, err
}

// FindAll finds all documents and decodes them into the results slice
func (m *Model) FindAll(ctx context.Context, filter interface{}, results interface{}, opts ...*options.FindOptions) error {
	// Convert filter to M for hooks
	filterM := toM(filter)

	// Execute pre-find hooks
	hc := &HookContext{
		Operation: OpFind,
		Filter:    filterM,
		Schema:    m.schema,
		Model:     m,
	}
	if err := m.schema.executePreHooks(ctx, OpFind, hc); err != nil {
		return err
	}

	// Execute the operation
	err := m.collection.FindAll(ctx, filter, results, opts...)

	// Execute post-find hooks
	hc.Result = results
	hc.Error = err
	_ = m.schema.executePostHooks(ctx, OpFind, hc)

	return err
}

// UpdateOne updates a single document with schema validation
// Automatically adds updatedAt if schema has timestamps enabled
func (m *Model) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*UpdateResult, error) {
	// Validate update data against schema
	if err := m.schema.ValidateUpdate(update); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Add updatedAt if timestamps are enabled
	if m.schema.Timestamps {
		update = m.addUpdatedAtToUpdate(update)
	}

	// Convert to M for hooks
	filterM := toM(filter)
	updateM := toM(update)

	// Execute pre-update hooks
	hc := &HookContext{
		Operation: OpUpdate,
		Filter:    filterM,
		Update:    updateM,
		Schema:    m.schema,
		Model:     m,
	}
	if err := m.schema.executePreHooks(ctx, OpUpdate, hc); err != nil {
		return nil, err
	}

	// Execute the operation
	result, err := m.collection.UpdateOne(ctx, filter, update, opts...)

	// Execute post-update hooks
	hc.Result = result
	hc.Error = err
	_ = m.schema.executePostHooks(ctx, OpUpdate, hc)

	return result, err
}

// UpdateOneByID updates a document by its ID with schema validation
func (m *Model) UpdateOneByID(ctx context.Context, id interface{}, update interface{}, opts ...*options.UpdateOptions) (*UpdateResult, error) {
	objID, err := toObjectID(id)
	if err != nil {
		return nil, err
	}

	return m.UpdateOne(ctx, bson.M{"_id": objID}, update, opts...)
}

// UpdateMany updates all documents matching the filter with schema validation
func (m *Model) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*UpdateResult, error) {
	// Validate update data against schema
	if err := m.schema.ValidateUpdate(update); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Add updatedAt if timestamps are enabled
	if m.schema.Timestamps {
		update = m.addUpdatedAtToUpdate(update)
	}

	// Convert to M for hooks
	filterM := toM(filter)
	updateM := toM(update)

	// Execute pre-updateMany hooks
	hc := &HookContext{
		Operation: OpUpdateMany,
		Filter:    filterM,
		Update:    updateM,
		Schema:    m.schema,
		Model:     m,
	}
	if err := m.schema.executePreHooks(ctx, OpUpdateMany, hc); err != nil {
		return nil, err
	}

	// Execute the operation
	result, err := m.collection.UpdateMany(ctx, filter, update, opts...)

	// Execute post-updateMany hooks
	hc.Result = result
	hc.Error = err
	_ = m.schema.executePostHooks(ctx, OpUpdateMany, hc)

	return result, err
}

// ReplaceOne replaces a document with schema validation
func (m *Model) ReplaceOne(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.ReplaceOptions) (*UpdateResult, error) {
	// Apply defaults from schema
	_ = m.schema.ApplyDefaults(replacement)

	// Validate replacement document against schema
	if err := m.schema.Validate(replacement); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Apply timestamps for replacement
	if m.schema.Timestamps {
		applyTimestamps(replacement, false)
	}

	return m.collection.ReplaceOne(ctx, filter, replacement, opts...)
}

// DeleteOne deletes a single document matching the filter
// Note: Delete operations don't require schema validation since they only use filters
func (m *Model) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*DeleteResult, error) {
	// Validate that filter is not empty for safety
	if filter == nil {
		return nil, ErrEmptyFilter
	}

	filterMap, err := toMap(filter)
	if err == nil && len(filterMap) == 0 {
		return nil, ErrEmptyFilter
	}

	// Convert to M for hooks
	filterM := toM(filter)

	// Execute pre-delete hooks
	hc := &HookContext{
		Operation: OpDelete,
		Filter:    filterM,
		Schema:    m.schema,
		Model:     m,
	}
	if err := m.schema.executePreHooks(ctx, OpDelete, hc); err != nil {
		return nil, err
	}

	// Execute the operation
	result, err := m.collection.DeleteOne(ctx, filter, opts...)

	// Execute post-delete hooks
	hc.Result = result
	hc.Error = err
	_ = m.schema.executePostHooks(ctx, OpDelete, hc)

	return result, err
}

// DeleteOneByID deletes a document by its ID
func (m *Model) DeleteOneByID(ctx context.Context, id interface{}, opts ...*options.DeleteOptions) (*DeleteResult, error) {
	objID, err := toObjectID(id)
	if err != nil {
		return nil, err
	}

	return m.DeleteOne(ctx, bson.M{"_id": objID}, opts...)
}

// DeleteMany deletes all documents matching the filter
func (m *Model) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*DeleteResult, error) {
	// Validate that filter is not empty for safety
	if filter == nil {
		return nil, ErrEmptyFilter
	}

	filterMap, err := toMap(filter)
	if err == nil && len(filterMap) == 0 {
		return nil, ErrEmptyFilter
	}

	// Convert to M for hooks
	filterM := toM(filter)

	// Execute pre-deleteMany hooks
	hc := &HookContext{
		Operation: OpDeleteMany,
		Filter:    filterM,
		Schema:    m.schema,
		Model:     m,
	}
	if err := m.schema.executePreHooks(ctx, OpDeleteMany, hc); err != nil {
		return nil, err
	}

	// Execute the operation
	result, err := m.collection.DeleteMany(ctx, filter, opts...)

	// Execute post-deleteMany hooks
	hc.Result = result
	hc.Error = err
	_ = m.schema.executePostHooks(ctx, OpDeleteMany, hc)

	return result, err
}

// CountDocuments counts documents matching the filter
func (m *Model) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return m.collection.CountDocuments(ctx, filter, opts...)
}

// Aggregate executes an aggregation pipeline
func (m *Model) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	return m.collection.Aggregate(ctx, pipeline, opts...)
}

// AggregateOne executes an aggregation pipeline and returns a single result
func (m *Model) AggregateOne(ctx context.Context, pipeline interface{}, result interface{}, opts ...*options.AggregateOptions) error {
	return m.collection.AggregateOne(ctx, pipeline, result, opts...)
}

// AggregateAll executes an aggregation pipeline and decodes all results
func (m *Model) AggregateAll(ctx context.Context, pipeline interface{}, results interface{}, opts ...*options.AggregateOptions) error {
	return m.collection.AggregateAll(ctx, pipeline, results, opts...)
}

// Distinct gets distinct values for a field
func (m *Model) Distinct(ctx context.Context, fieldName string, filter interface{}, opts ...*options.DistinctOptions) ([]interface{}, error) {
	return m.collection.Distinct(ctx, fieldName, filter, opts...)
}

// Exists checks if a document matching the filter exists
func (m *Model) Exists(ctx context.Context, filter interface{}) (bool, error) {
	count, err := m.collection.CountDocuments(ctx, filter, options.Count().SetLimit(1))
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByID checks if a document with the given ID exists
func (m *Model) ExistsByID(ctx context.Context, id interface{}) (bool, error) {
	objID, err := toObjectID(id)
	if err != nil {
		return false, err
	}

	return m.Exists(ctx, bson.M{"_id": objID})
}

// FindOneAndUpdate finds a document and updates it atomically
func (m *Model) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {
	// Validate update data against schema
	if err := m.schema.ValidateUpdate(update); err != nil {
		// Return an error result - we can't return an error directly
		// The caller should check the result's Err() method
		return nil
	}

	// Add updatedAt if timestamps are enabled
	if m.schema.Timestamps {
		update = m.addUpdatedAtToUpdate(update)
	}

	opCtx, cancel := m.collection.createOperationContext(ctx)
	defer cancel()

	return m.collection.collection.FindOneAndUpdate(opCtx, filter, update, opts...)
}

// FindOneAndDelete finds a document and deletes it atomically
func (m *Model) FindOneAndDelete(ctx context.Context, filter interface{}, opts ...*options.FindOneAndDeleteOptions) *mongo.SingleResult {
	opCtx, cancel := m.collection.createOperationContext(ctx)
	defer cancel()

	return m.collection.collection.FindOneAndDelete(opCtx, filter, opts...)
}

// FindOneAndReplace finds a document and replaces it atomically
func (m *Model) FindOneAndReplace(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.FindOneAndReplaceOptions) *mongo.SingleResult {
	// Validate replacement document
	if err := m.schema.Validate(replacement); err != nil {
		return nil
	}

	// Apply timestamps for replacement
	if m.schema.Timestamps {
		applyTimestamps(replacement, false)
	}

	opCtx, cancel := m.collection.createOperationContext(ctx)
	defer cancel()

	return m.collection.collection.FindOneAndReplace(opCtx, filter, replacement, opts...)
}

// BulkWrite performs multiple write operations with schema validation
func (m *Model) BulkWrite(ctx context.Context, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	// Validate each write model
	for i, model := range models {
		switch wm := model.(type) {
		case *mongo.InsertOneModel:
			if err := m.schema.Validate(wm.Document); err != nil {
				return nil, fmt.Errorf("validation failed for insert at index %d: %w", i, err)
			}
			applyObjectID(wm.Document)
			if m.schema.Timestamps {
				applyTimestamps(wm.Document, true)
			}
		case *mongo.UpdateOneModel:
			if err := m.schema.ValidateUpdate(wm.Update); err != nil {
				return nil, fmt.Errorf("validation failed for update at index %d: %w", i, err)
			}
			if m.schema.Timestamps {
				wm.Update = m.addUpdatedAtToUpdate(wm.Update)
			}
		case *mongo.UpdateManyModel:
			if err := m.schema.ValidateUpdate(wm.Update); err != nil {
				return nil, fmt.Errorf("validation failed for update at index %d: %w", i, err)
			}
			if m.schema.Timestamps {
				wm.Update = m.addUpdatedAtToUpdate(wm.Update)
			}
		case *mongo.ReplaceOneModel:
			if err := m.schema.Validate(wm.Replacement); err != nil {
				return nil, fmt.Errorf("validation failed for replace at index %d: %w", i, err)
			}
			if m.schema.Timestamps {
				applyTimestamps(wm.Replacement, false)
			}
			// DeleteOneModel and DeleteManyModel don't need validation
		}
	}

	return m.collection.BulkWrite(ctx, models, opts...)
}

// Drop drops the entire collection
func (m *Model) Drop(ctx context.Context) error {
	return m.collection.Drop(ctx)
}

// addUpdatedAtToUpdate adds updatedAt field to update operations
func (m *Model) addUpdatedAtToUpdate(update interface{}) interface{} {
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
			// Create $set with updatedAt if no $set exists
			updateMap["$set"] = bson.M{"updatedAt": primitive.NewDateTimeFromTime(time.Now())}
		}
	}
	return update
}

// toObjectID converts various types to primitive.ObjectID
func toObjectID(id interface{}) (primitive.ObjectID, error) {
	switch v := id.(type) {
	case primitive.ObjectID:
		return v, nil
	case string:
		return primitive.ObjectIDFromHex(v)
	default:
		return primitive.NilObjectID, fmt.Errorf("%w: %T", ErrInvalidID, id)
	}
}

// toM converts various types to M (bson.M alias) for hook context
func toM(v interface{}) M {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case M:
		return val
	case map[string]interface{}:
		return M(val)
	default:
		// Try to convert struct to map
		if m, ok := toMapInterface(v); ok {
			return M(m)
		}
		return nil
	}
}
