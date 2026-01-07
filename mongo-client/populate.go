package mongoclient

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Aggregator interface for testing - Collection implements this
type Aggregator interface {
	AggregateAll(ctx context.Context, pipeline interface{}, results interface{}, opts ...*options.AggregateOptions) error
}

// Ensure Collection implements Aggregator
var _ Aggregator = (*Collection)(nil)

// RefInfo contains reference information from schema
type RefInfo struct {
	Field      string // Field name (e.g., "userId")
	Collection string // Referenced collection (e.g., "users")
}

// GetRefFields returns all fields that have Ref defined in the schema
func (s *Schema) GetRefFields() []RefInfo {
	var refs []RefInfo
	for name, field := range s.Fields {
		if field.Ref != "" {
			refs = append(refs, RefInfo{
				Field:      name,
				Collection: field.Ref,
			})
		}
	}
	return refs
}

// GetRef returns the referenced collection for a field, or empty string if not a ref
func (s *Schema) GetRef(fieldName string) string {
	if field, ok := s.Fields[fieldName]; ok {
		return field.Ref
	}
	return ""
}

// HasRef checks if a field has a Ref defined
func (s *Schema) HasRef(fieldName string) bool {
	if field, ok := s.Fields[fieldName]; ok {
		return field.Ref != ""
	}
	return false
}

// PopulateOptions configures a populate (join) operation using $lookup
type PopulateOptions struct {
	// Path is the field in the current collection that holds the reference
	// Example: "userId" in a Product document
	Path string

	// From is the target collection to join with
	// Example: "users"
	From string

	// LocalField is the field in the current collection (defaults to Path)
	LocalField string

	// ForeignField is the field in the target collection (defaults to "_id")
	ForeignField string

	// As is the output field name for populated documents (defaults to Path without "Id" suffix)
	// Example: "userId" -> "user", "categoryId" -> "category"
	As string

	// Select specifies which fields to include from the populated document
	// Empty means all fields
	// Example: []string{"name", "email"}
	Select []string

	// Unwind converts the result array to a single object (for one-to-one relations)
	// Defaults to true
	Unwind *bool

	// PreserveNull keeps documents even if the join produces no match
	// When true and no match found, the field will be null instead of missing
	// Defaults to true
	PreserveNull *bool
}

// DefaultPopulateOptions returns populate options with defaults applied
func (p *PopulateOptions) withDefaults() *PopulateOptions {
	opts := *p

	if opts.LocalField == "" {
		opts.LocalField = opts.Path
	}

	if opts.ForeignField == "" {
		opts.ForeignField = "_id"
	}

	if opts.As == "" {
		// Remove common suffixes: userId -> user, category_id -> category
		opts.As = opts.Path
		suffixes := []string{"Id", "ID", "_id", "Ref", "_ref"}
		for _, suffix := range suffixes {
			if len(opts.As) > len(suffix) && opts.As[len(opts.As)-len(suffix):] == suffix {
				opts.As = opts.As[:len(opts.As)-len(suffix)]
				break
			}
		}
	}

	if opts.Unwind == nil {
		unwind := true
		opts.Unwind = &unwind
	}

	if opts.PreserveNull == nil {
		preserve := true
		opts.PreserveNull = &preserve
	}

	return &opts
}

// buildLookupStages creates $lookup and optional $unwind stages
func (p *PopulateOptions) buildLookupStages() []bson.M {
	opts := p.withDefaults()
	stages := []bson.M{}

	// Build $lookup stage
	lookupStage := bson.M{
		"$lookup": bson.M{
			"from":         opts.From,
			"localField":   opts.LocalField,
			"foreignField": opts.ForeignField,
			"as":           opts.As,
		},
	}

	// Add pipeline for projection if Select is specified
	if len(opts.Select) > 0 {
		projection := bson.M{}
		for _, field := range opts.Select {
			projection[field] = 1
		}
		lookupStage["$lookup"] = bson.M{
			"from":         opts.From,
			"localField":   opts.LocalField,
			"foreignField": opts.ForeignField,
			"as":           opts.As,
			"pipeline": bson.A{
				bson.M{"$project": projection},
			},
		}
	}

	stages = append(stages, lookupStage)

	// Add $unwind stage if needed (for one-to-one relations)
	if *opts.Unwind {
		unwindStage := bson.M{
			"$unwind": bson.M{
				"path": "$" + opts.As,
			},
		}
		if *opts.PreserveNull {
			unwindStage["$unwind"].(bson.M)["preserveNullAndEmptyArrays"] = true
		}
		stages = append(stages, unwindStage)
	}

	return stages
}

// Populate creates a simple PopulateOptions for common use cases
// Usage: Populate("userId", "users") or Populate("userId", "users", "name", "email")
func Populate(path, from string, selectFields ...string) PopulateOptions {
	return PopulateOptions{
		Path:   path,
		From:   from,
		Select: selectFields,
	}
}

// PopulateMany creates PopulateOptions for one-to-many relations (keeps array)
func PopulateMany(path, from string, selectFields ...string) PopulateOptions {
	unwind := false
	return PopulateOptions{
		Path:   path,
		From:   from,
		Select: selectFields,
		Unwind: &unwind,
	}
}

// FindOneWithPopulate finds a single document and populates references
func (c *Collection) FindOneWithPopulate(ctx context.Context, filter interface{}, populates ...PopulateOptions) (bson.M, error) {
	results, err := c.FindWithPopulate(ctx, filter, populates, 1)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, ErrNoDocuments
	}
	return results[0], nil
}

// FindWithPopulate finds documents and populates references using $lookup
func (c *Collection) FindWithPopulate(ctx context.Context, filter interface{}, populates []PopulateOptions, limit ...int64) ([]bson.M, error) {
	if filter == nil {
		filter = bson.M{}
	}

	// Build aggregation pipeline
	pipeline := bson.A{
		bson.M{"$match": filter},
	}

	// Add $lookup stages for each populate
	for _, p := range populates {
		stages := p.buildLookupStages()
		for _, stage := range stages {
			pipeline = append(pipeline, stage)
		}
	}

	// Add limit if specified
	if len(limit) > 0 && limit[0] > 0 {
		pipeline = append(pipeline, bson.M{"$limit": limit[0]})
	}

	var results []bson.M
	err := c.AggregateAll(ctx, pipeline, &results)
	if err != nil {
		return nil, fmt.Errorf("populate aggregation failed: %w", err)
	}

	if results == nil {
		results = []bson.M{}
	}

	return results, nil
}

// FindAllWithPopulate finds all matching documents with populated references
func (c *Collection) FindAllWithPopulate(ctx context.Context, filter interface{}, populates ...PopulateOptions) ([]bson.M, error) {
	return c.FindWithPopulate(ctx, filter, populates)
}

// Model methods

// FindOneWithPopulate finds a single document and populates references (Model method)
func (m *Model) FindOneWithPopulate(ctx context.Context, filter interface{}, populates ...PopulateOptions) (bson.M, error) {
	return m.collection.FindOneWithPopulate(ctx, filter, populates...)
}

// FindWithPopulate finds documents and populates references (Model method)
func (m *Model) FindWithPopulate(ctx context.Context, filter interface{}, populates []PopulateOptions, limit ...int64) ([]bson.M, error) {
	return m.collection.FindWithPopulate(ctx, filter, populates, limit...)
}

// FindAllWithPopulate finds all matching documents with populated references (Model method)
func (m *Model) FindAllWithPopulate(ctx context.Context, filter interface{}, populates ...PopulateOptions) ([]bson.M, error) {
	return m.collection.FindAllWithPopulate(ctx, filter, populates...)
}

// FindByIDWithPopulate finds a document by ID and populates references
func (m *Model) FindByIDWithPopulate(ctx context.Context, id interface{}, populates ...PopulateOptions) (bson.M, error) {
	oid, err := toObjectID(id)
	if err != nil {
		return nil, err
	}
	return m.collection.FindOneWithPopulate(ctx, bson.M{"_id": oid}, populates...)
}

// Schema-aware populate methods (safe - uses Ref from schema)

// PopulateRef creates PopulateOptions from schema's Ref definition
// This is the SAFE way - collection name comes from schema, not manual input
func (m *Model) PopulateRef(fieldName string, selectFields ...string) (PopulateOptions, error) {
	if m.schema == nil {
		return PopulateOptions{}, fmt.Errorf("schema is required for PopulateRef")
	}

	ref := m.schema.GetRef(fieldName)
	if ref == "" {
		return PopulateOptions{}, fmt.Errorf("field '%s' does not have a Ref defined in schema", fieldName)
	}

	return PopulateOptions{
		Path:   fieldName,
		From:   ref,
		Select: selectFields,
	}, nil
}

// PopulateRefs creates PopulateOptions for multiple fields from schema
func (m *Model) PopulateRefs(fieldNames ...string) ([]PopulateOptions, error) {
	if m.schema == nil {
		return nil, fmt.Errorf("schema is required for PopulateRefs")
	}

	var opts []PopulateOptions
	for _, fieldName := range fieldNames {
		opt, err := m.PopulateRef(fieldName)
		if err != nil {
			return nil, err
		}
		opts = append(opts, opt)
	}
	return opts, nil
}

// GetAllRefs returns PopulateOptions for ALL Ref fields in schema
func (m *Model) GetAllRefs() []PopulateOptions {
	if m.schema == nil {
		return nil
	}

	var opts []PopulateOptions
	for _, ref := range m.schema.GetRefFields() {
		opts = append(opts, PopulateOptions{
			Path: ref.Field,
			From: ref.Collection,
		})
	}
	return opts
}

// FindOnePopulateRef finds one document and populates specified ref fields (schema-aware)
// Usage: productModel.FindOnePopulateRef(ctx, filter, "userId", "categoryId")
func (m *Model) FindOnePopulateRef(ctx context.Context, filter interface{}, fieldNames ...string) (bson.M, error) {
	opts, err := m.PopulateRefs(fieldNames...)
	if err != nil {
		return nil, err
	}
	return m.collection.FindOneWithPopulate(ctx, filter, opts...)
}

// FindAllPopulateRef finds all documents and populates specified ref fields (schema-aware)
func (m *Model) FindAllPopulateRef(ctx context.Context, filter interface{}, fieldNames ...string) ([]bson.M, error) {
	opts, err := m.PopulateRefs(fieldNames...)
	if err != nil {
		return nil, err
	}
	return m.collection.FindAllWithPopulate(ctx, filter, opts...)
}

// FindByIDPopulateRef finds by ID and populates specified ref fields (schema-aware)
func (m *Model) FindByIDPopulateRef(ctx context.Context, id interface{}, fieldNames ...string) (bson.M, error) {
	oid, err := toObjectID(id)
	if err != nil {
		return nil, err
	}
	opts, err := m.PopulateRefs(fieldNames...)
	if err != nil {
		return nil, err
	}
	return m.collection.FindOneWithPopulate(ctx, bson.M{"_id": oid}, opts...)
}

// FindOnePopulateAll finds one document and populates ALL ref fields from schema
func (m *Model) FindOnePopulateAll(ctx context.Context, filter interface{}) (bson.M, error) {
	opts := m.GetAllRefs()
	if len(opts) == 0 {
		// No refs, just do normal find
		var result bson.M
		err := m.FindOne(ctx, filter).Decode(&result)
		return result, err
	}
	return m.collection.FindOneWithPopulate(ctx, filter, opts...)
}

// FindAllPopulateAll finds all documents and populates ALL ref fields from schema
func (m *Model) FindAllPopulateAll(ctx context.Context, filter interface{}) ([]bson.M, error) {
	opts := m.GetAllRefs()
	if len(opts) == 0 {
		var results []bson.M
		err := m.FindAll(ctx, filter, &results)
		return results, err
	}
	return m.collection.FindAllWithPopulate(ctx, filter, opts...)
}

// FindByIDPopulateAll finds by ID and populates ALL ref fields from schema
func (m *Model) FindByIDPopulateAll(ctx context.Context, id interface{}) (bson.M, error) {
	oid, err := toObjectID(id)
	if err != nil {
		return nil, err
	}
	opts := m.GetAllRefs()
	if len(opts) == 0 {
		var result bson.M
		err := m.FindOneByID(ctx, oid).Decode(&result)
		return result, err
	}
	return m.collection.FindOneWithPopulate(ctx, bson.M{"_id": oid}, opts...)
}

// PopulateQuery provides a fluent interface for building populate queries
type PopulateQuery struct {
	collection *Collection
	aggregator Aggregator // for testing - if nil, uses collection
	schema     *Schema
	filter     interface{}
	populates  []PopulateOptions
	limit      int64
	skip       int64
	sort       interface{}
	err        error // stores first error for deferred error handling
}

// NewPopulateQuery creates a new populate query builder
func (c *Collection) Populate(filter interface{}) *PopulateQuery {
	return &PopulateQuery{
		collection: c,
		filter:     filter,
		populates:  []PopulateOptions{},
	}
}

// NewPopulateQuery creates a new populate query builder (Model method)
// This version has access to schema for safe Ref-based populates
func (m *Model) Populate(filter interface{}) *PopulateQuery {
	return &PopulateQuery{
		collection: m.collection,
		schema:     m.schema,
		filter:     filter,
		populates:  []PopulateOptions{},
	}
}

// Join adds a populate option to the query
func (q *PopulateQuery) Join(path, from string, selectFields ...string) *PopulateQuery {
	q.populates = append(q.populates, Populate(path, from, selectFields...))
	return q
}

// JoinMany adds a populate option for one-to-many relations
func (q *PopulateQuery) JoinMany(path, from string, selectFields ...string) *PopulateQuery {
	q.populates = append(q.populates, PopulateMany(path, from, selectFields...))
	return q
}

// JoinWith adds a custom PopulateOptions to the query
func (q *PopulateQuery) JoinWith(opts PopulateOptions) *PopulateQuery {
	q.populates = append(q.populates, opts)
	return q
}

// JoinRef adds a populate using schema's Ref definition (SAFE - schema-aware)
// Usage: productModel.Populate(filter).JoinRef("userId").JoinRef("categoryId").All(ctx)
func (q *PopulateQuery) JoinRef(fieldName string, selectFields ...string) *PopulateQuery {
	if q.err != nil {
		return q // already has error, skip
	}

	if q.schema == nil {
		q.err = fmt.Errorf("schema is required for JoinRef (use Join instead for collection-level queries)")
		return q
	}

	ref := q.schema.GetRef(fieldName)
	if ref == "" {
		q.err = fmt.Errorf("field '%s' does not have a Ref defined in schema", fieldName)
		return q
	}

	q.populates = append(q.populates, PopulateOptions{
		Path:   fieldName,
		From:   ref,
		Select: selectFields,
	})
	return q
}

// JoinRefMany adds a populate for one-to-many using schema's Ref (keeps array)
func (q *PopulateQuery) JoinRefMany(fieldName string, selectFields ...string) *PopulateQuery {
	if q.err != nil {
		return q
	}

	if q.schema == nil {
		q.err = fmt.Errorf("schema is required for JoinRefMany")
		return q
	}

	ref := q.schema.GetRef(fieldName)
	if ref == "" {
		q.err = fmt.Errorf("field '%s' does not have a Ref defined in schema", fieldName)
		return q
	}

	unwind := false
	q.populates = append(q.populates, PopulateOptions{
		Path:   fieldName,
		From:   ref,
		Select: selectFields,
		Unwind: &unwind,
	})
	return q
}

// JoinAllRefs populates ALL Ref fields defined in schema
func (q *PopulateQuery) JoinAllRefs() *PopulateQuery {
	if q.err != nil {
		return q
	}

	if q.schema == nil {
		q.err = fmt.Errorf("schema is required for JoinAllRefs")
		return q
	}

	for _, ref := range q.schema.GetRefFields() {
		q.populates = append(q.populates, PopulateOptions{
			Path: ref.Field,
			From: ref.Collection,
		})
	}
	return q
}

// WithAggregator sets a custom aggregator (for testing)
func (q *PopulateQuery) WithAggregator(agg Aggregator) *PopulateQuery {
	q.aggregator = agg
	return q
}

// Limit sets the maximum number of documents to return
func (q *PopulateQuery) Limit(n int64) *PopulateQuery {
	q.limit = n
	return q
}

// Skip sets the number of documents to skip
func (q *PopulateQuery) Skip(n int64) *PopulateQuery {
	q.skip = n
	return q
}

// Sort sets the sort order
func (q *PopulateQuery) Sort(sort interface{}) *PopulateQuery {
	q.sort = sort
	return q
}

// One executes the query and returns a single document
func (q *PopulateQuery) One(ctx context.Context) (bson.M, error) {
	results, err := q.All(ctx)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, ErrNoDocuments
	}
	return results[0], nil
}

// All executes the query and returns all matching documents
func (q *PopulateQuery) All(ctx context.Context) ([]bson.M, error) {
	// Check for deferred errors from JoinRef etc
	if q.err != nil {
		return nil, q.err
	}

	if q.filter == nil {
		q.filter = bson.M{}
	}

	// Build aggregation pipeline
	pipeline := bson.A{
		bson.M{"$match": q.filter},
	}

	// Add sort if specified
	if q.sort != nil {
		pipeline = append(pipeline, bson.M{"$sort": q.sort})
	}

	// Add skip if specified
	if q.skip > 0 {
		pipeline = append(pipeline, bson.M{"$skip": q.skip})
	}

	// Add limit if specified
	if q.limit > 0 {
		pipeline = append(pipeline, bson.M{"$limit": q.limit})
	}

	// Add $lookup stages for each populate
	for _, p := range q.populates {
		stages := p.buildLookupStages()
		for _, stage := range stages {
			pipeline = append(pipeline, stage)
		}
	}

	var results []bson.M
	// Use aggregator if set (for testing), otherwise use collection
	var agg Aggregator = q.collection
	if q.aggregator != nil {
		agg = q.aggregator
	}
	err := agg.AggregateAll(ctx, pipeline, &results)
	if err != nil {
		return nil, fmt.Errorf("populate query failed: %w", err)
	}

	if results == nil {
		results = []bson.M{}
	}

	return results, nil
}
