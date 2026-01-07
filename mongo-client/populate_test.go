package mongoclient

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// mockAggregator is a mock implementation of Aggregator for testing
type mockAggregator struct {
	results      []bson.M
	err          error
	lastPipeline interface{}
}

func (m *mockAggregator) AggregateAll(ctx context.Context, pipeline interface{}, results interface{}, opts ...*options.AggregateOptions) error {
	m.lastPipeline = pipeline
	if m.err != nil {
		return m.err
	}
	// Use reflection to set results
	resultsVal := reflect.ValueOf(results).Elem()
	resultsVal.Set(reflect.ValueOf(m.results))
	return nil
}

func TestPopulateOptions_withDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    PopulateOptions
		expected PopulateOptions
	}{
		{
			name: "basic defaults",
			input: PopulateOptions{
				Path: "userId",
				From: "users",
			},
			expected: PopulateOptions{
				Path:         "userId",
				From:         "users",
				LocalField:   "userId",
				ForeignField: "_id",
				As:           "user",
			},
		},
		{
			name: "with ID suffix",
			input: PopulateOptions{
				Path: "categoryID",
				From: "categories",
			},
			expected: PopulateOptions{
				Path:         "categoryID",
				From:         "categories",
				LocalField:   "categoryID",
				ForeignField: "_id",
				As:           "category",
			},
		},
		{
			name: "with _id suffix",
			input: PopulateOptions{
				Path: "author_id",
				From: "users",
			},
			expected: PopulateOptions{
				Path:         "author_id",
				From:         "users",
				LocalField:   "author_id",
				ForeignField: "_id",
				As:           "author",
			},
		},
		{
			name: "with Ref suffix",
			input: PopulateOptions{
				Path: "productRef",
				From: "products",
			},
			expected: PopulateOptions{
				Path:         "productRef",
				From:         "products",
				LocalField:   "productRef",
				ForeignField: "_id",
				As:           "product",
			},
		},
		{
			name: "custom As field",
			input: PopulateOptions{
				Path: "userId",
				From: "users",
				As:   "owner",
			},
			expected: PopulateOptions{
				Path:         "userId",
				From:         "users",
				LocalField:   "userId",
				ForeignField: "_id",
				As:           "owner",
			},
		},
		{
			name: "custom foreign field",
			input: PopulateOptions{
				Path:         "email",
				From:         "users",
				ForeignField: "email",
			},
			expected: PopulateOptions{
				Path:         "email",
				From:         "users",
				LocalField:   "email",
				ForeignField: "email",
				As:           "email",
			},
		},
		{
			name: "no suffix to remove",
			input: PopulateOptions{
				Path: "author",
				From: "users",
			},
			expected: PopulateOptions{
				Path:         "author",
				From:         "users",
				LocalField:   "author",
				ForeignField: "_id",
				As:           "author",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.withDefaults()

			if result.Path != tt.expected.Path {
				t.Errorf("Path = %v, want %v", result.Path, tt.expected.Path)
			}
			if result.From != tt.expected.From {
				t.Errorf("From = %v, want %v", result.From, tt.expected.From)
			}
			if result.LocalField != tt.expected.LocalField {
				t.Errorf("LocalField = %v, want %v", result.LocalField, tt.expected.LocalField)
			}
			if result.ForeignField != tt.expected.ForeignField {
				t.Errorf("ForeignField = %v, want %v", result.ForeignField, tt.expected.ForeignField)
			}
			if result.As != tt.expected.As {
				t.Errorf("As = %v, want %v", result.As, tt.expected.As)
			}
			if result.Unwind == nil || !*result.Unwind {
				t.Error("Unwind should default to true")
			}
			if result.PreserveNull == nil || !*result.PreserveNull {
				t.Error("PreserveNull should default to true")
			}
		})
	}
}

func TestPopulateOptions_buildLookupStages(t *testing.T) {
	tests := []struct {
		name           string
		opts           PopulateOptions
		expectedStages int
		checkLookup    func(t *testing.T, stage bson.M)
	}{
		{
			name: "basic lookup with unwind",
			opts: PopulateOptions{
				Path: "userId",
				From: "users",
			},
			expectedStages: 2, // $lookup + $unwind
			checkLookup: func(t *testing.T, stage bson.M) {
				lookup := stage["$lookup"].(bson.M)
				if lookup["from"] != "users" {
					t.Errorf("from = %v, want users", lookup["from"])
				}
				if lookup["localField"] != "userId" {
					t.Errorf("localField = %v, want userId", lookup["localField"])
				}
				if lookup["foreignField"] != "_id" {
					t.Errorf("foreignField = %v, want _id", lookup["foreignField"])
				}
				if lookup["as"] != "user" {
					t.Errorf("as = %v, want user", lookup["as"])
				}
			},
		},
		{
			name: "lookup without unwind",
			opts: func() PopulateOptions {
				unwind := false
				return PopulateOptions{
					Path:   "tagIds",
					From:   "tags",
					Unwind: &unwind,
				}
			}(),
			expectedStages: 1, // only $lookup
		},
		{
			name: "lookup with select",
			opts: PopulateOptions{
				Path:   "userId",
				From:   "users",
				Select: []string{"name", "email"},
			},
			expectedStages: 2,
			checkLookup: func(t *testing.T, stage bson.M) {
				lookup := stage["$lookup"].(bson.M)
				pipeline, ok := lookup["pipeline"].(bson.A)
				if !ok {
					t.Fatal("expected pipeline in lookup")
				}
				if len(pipeline) != 1 {
					t.Errorf("pipeline length = %d, want 1", len(pipeline))
				}
				project := pipeline[0].(bson.M)["$project"].(bson.M)
				if project["name"] != 1 {
					t.Error("expected name in projection")
				}
				if project["email"] != 1 {
					t.Error("expected email in projection")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stages := tt.opts.buildLookupStages()

			if len(stages) != tt.expectedStages {
				t.Errorf("stages count = %d, want %d", len(stages), tt.expectedStages)
			}

			if tt.checkLookup != nil && len(stages) > 0 {
				tt.checkLookup(t, stages[0])
			}
		})
	}
}

func TestPopulate_Helper(t *testing.T) {
	// Test Populate helper function
	opts := Populate("userId", "users")
	if opts.Path != "userId" {
		t.Errorf("Path = %v, want userId", opts.Path)
	}
	if opts.From != "users" {
		t.Errorf("From = %v, want users", opts.From)
	}

	// With select fields
	opts = Populate("userId", "users", "name", "email")
	if len(opts.Select) != 2 {
		t.Errorf("Select length = %d, want 2", len(opts.Select))
	}
	if opts.Select[0] != "name" || opts.Select[1] != "email" {
		t.Errorf("Select = %v, want [name email]", opts.Select)
	}
}

func TestPopulateMany_Helper(t *testing.T) {
	opts := PopulateMany("tagIds", "tags")
	if opts.Path != "tagIds" {
		t.Errorf("Path = %v, want tagIds", opts.Path)
	}
	if opts.From != "tags" {
		t.Errorf("From = %v, want tags", opts.From)
	}
	if opts.Unwind == nil || *opts.Unwind != false {
		t.Error("Unwind should be false for PopulateMany")
	}
}

func TestPopulateQuery_Builder(t *testing.T) {
	// Create a mock collection (we can't fully test without DB, but we can test builder)
	query := &PopulateQuery{
		filter:    bson.M{"status": "active"},
		populates: []PopulateOptions{},
	}

	// Test chaining
	query.Join("userId", "users", "name").
		Join("categoryId", "categories").
		JoinMany("tagIds", "tags").
		Limit(10).
		Skip(5).
		Sort(bson.D{{"createdAt", -1}})

	if len(query.populates) != 3 {
		t.Errorf("populates count = %d, want 3", len(query.populates))
	}

	if query.limit != 10 {
		t.Errorf("limit = %d, want 10", query.limit)
	}

	if query.skip != 5 {
		t.Errorf("skip = %d, want 5", query.skip)
	}

	// Check first populate
	if query.populates[0].Path != "userId" {
		t.Errorf("first populate path = %v, want userId", query.populates[0].Path)
	}
	if len(query.populates[0].Select) != 1 {
		t.Errorf("first populate select length = %d, want 1", len(query.populates[0].Select))
	}

	// Check third populate (many)
	if query.populates[2].Unwind == nil || *query.populates[2].Unwind != false {
		t.Error("third populate (many) should have Unwind=false")
	}
}

func TestPopulateQuery_JoinWith(t *testing.T) {
	query := &PopulateQuery{
		filter:    bson.M{},
		populates: []PopulateOptions{},
	}

	customOpts := PopulateOptions{
		Path:         "authorEmail",
		From:         "users",
		LocalField:   "authorEmail",
		ForeignField: "email",
		As:           "author",
		Select:       []string{"name", "avatar"},
	}

	query.JoinWith(customOpts)

	if len(query.populates) != 1 {
		t.Fatalf("populates count = %d, want 1", len(query.populates))
	}

	p := query.populates[0]
	if p.LocalField != "authorEmail" {
		t.Errorf("LocalField = %v, want authorEmail", p.LocalField)
	}
	if p.ForeignField != "email" {
		t.Errorf("ForeignField = %v, want email", p.ForeignField)
	}
}

func TestPopulateOptions_PreserveNullInUnwind(t *testing.T) {
	opts := PopulateOptions{
		Path: "userId",
		From: "users",
	}

	stages := opts.buildLookupStages()
	if len(stages) < 2 {
		t.Fatal("expected at least 2 stages")
	}

	unwindStage := stages[1]["$unwind"].(bson.M)
	if unwindStage["preserveNullAndEmptyArrays"] != true {
		t.Error("expected preserveNullAndEmptyArrays to be true")
	}
}

func TestPopulateOptions_DisablePreserveNull(t *testing.T) {
	preserve := false
	opts := PopulateOptions{
		Path:         "userId",
		From:         "users",
		PreserveNull: &preserve,
	}

	stages := opts.buildLookupStages()
	if len(stages) < 2 {
		t.Fatal("expected at least 2 stages")
	}

	unwindStage := stages[1]["$unwind"].(bson.M)
	if _, ok := unwindStage["preserveNullAndEmptyArrays"]; ok {
		t.Error("preserveNullAndEmptyArrays should not be set when PreserveNull is false")
	}
}

// Schema-aware populate tests

func TestSchema_GetRefFields(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString).
		Field("userId", TypeObjectID, Ref("users")).
		Field("categoryId", TypeObjectID, Ref("categories")).
		Field("email", TypeString) // no ref

	refs := schema.GetRefFields()

	if len(refs) != 2 {
		t.Errorf("expected 2 refs, got %d", len(refs))
	}

	// Check refs are present (order may vary due to map)
	refMap := make(map[string]string)
	for _, r := range refs {
		refMap[r.Field] = r.Collection
	}

	if refMap["userId"] != "users" {
		t.Errorf("expected userId -> users, got %s", refMap["userId"])
	}
	if refMap["categoryId"] != "categories" {
		t.Errorf("expected categoryId -> categories, got %s", refMap["categoryId"])
	}
}

func TestSchema_GetRef(t *testing.T) {
	schema := NewSchema().
		Field("userId", TypeObjectID, Ref("users")).
		Field("name", TypeString)

	// Field with ref
	if ref := schema.GetRef("userId"); ref != "users" {
		t.Errorf("GetRef(userId) = %s, want users", ref)
	}

	// Field without ref
	if ref := schema.GetRef("name"); ref != "" {
		t.Errorf("GetRef(name) = %s, want empty", ref)
	}

	// Non-existent field
	if ref := schema.GetRef("notExists"); ref != "" {
		t.Errorf("GetRef(notExists) = %s, want empty", ref)
	}
}

func TestSchema_HasRef(t *testing.T) {
	schema := NewSchema().
		Field("userId", TypeObjectID, Ref("users")).
		Field("name", TypeString)

	if !schema.HasRef("userId") {
		t.Error("HasRef(userId) should be true")
	}

	if schema.HasRef("name") {
		t.Error("HasRef(name) should be false")
	}

	if schema.HasRef("notExists") {
		t.Error("HasRef(notExists) should be false")
	}
}

func TestPopulateQuery_JoinRef(t *testing.T) {
	schema := NewSchema().
		Field("userId", TypeObjectID, Ref("users")).
		Field("categoryId", TypeObjectID, Ref("categories"))

	query := &PopulateQuery{
		schema:    schema,
		filter:    bson.M{},
		populates: []PopulateOptions{},
	}

	// Test JoinRef
	query.JoinRef("userId", "name", "email")

	if query.err != nil {
		t.Fatalf("unexpected error: %v", query.err)
	}

	if len(query.populates) != 1 {
		t.Fatalf("expected 1 populate, got %d", len(query.populates))
	}

	p := query.populates[0]
	if p.Path != "userId" {
		t.Errorf("Path = %s, want userId", p.Path)
	}
	if p.From != "users" {
		t.Errorf("From = %s, want users (from schema Ref)", p.From)
	}
	if len(p.Select) != 2 {
		t.Errorf("Select length = %d, want 2", len(p.Select))
	}
}

func TestPopulateQuery_JoinRef_NoSchema(t *testing.T) {
	query := &PopulateQuery{
		schema:    nil, // no schema
		filter:    bson.M{},
		populates: []PopulateOptions{},
	}

	query.JoinRef("userId")

	if query.err == nil {
		t.Error("expected error when schema is nil")
	}
}

func TestPopulateQuery_JoinRef_NoRef(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString) // no Ref

	query := &PopulateQuery{
		schema:    schema,
		filter:    bson.M{},
		populates: []PopulateOptions{},
	}

	query.JoinRef("name")

	if query.err == nil {
		t.Error("expected error when field has no Ref")
	}
}

func TestPopulateQuery_JoinAllRefs(t *testing.T) {
	schema := NewSchema().
		Field("userId", TypeObjectID, Ref("users")).
		Field("categoryId", TypeObjectID, Ref("categories")).
		Field("name", TypeString)

	query := &PopulateQuery{
		schema:    schema,
		filter:    bson.M{},
		populates: []PopulateOptions{},
	}

	query.JoinAllRefs()

	if query.err != nil {
		t.Fatalf("unexpected error: %v", query.err)
	}

	if len(query.populates) != 2 {
		t.Errorf("expected 2 populates, got %d", len(query.populates))
	}
}

func TestPopulateQuery_JoinRefMany(t *testing.T) {
	schema := NewSchema().
		Field("tagIds", TypeArray, Ref("tags"))

	query := &PopulateQuery{
		schema:    schema,
		filter:    bson.M{},
		populates: []PopulateOptions{},
	}

	query.JoinRefMany("tagIds", "name")

	if query.err != nil {
		t.Fatalf("unexpected error: %v", query.err)
	}

	if len(query.populates) != 1 {
		t.Fatalf("expected 1 populate, got %d", len(query.populates))
	}

	p := query.populates[0]
	if p.Unwind == nil || *p.Unwind != false {
		t.Error("Unwind should be false for JoinRefMany")
	}
}

func TestPopulateQuery_ChainedJoinRef(t *testing.T) {
	schema := NewSchema().
		Field("userId", TypeObjectID, Ref("users")).
		Field("categoryId", TypeObjectID, Ref("categories")).
		Field("tagIds", TypeArray, Ref("tags"))

	query := &PopulateQuery{
		schema:    schema,
		filter:    bson.M{},
		populates: []PopulateOptions{},
	}

	// Chain multiple JoinRef calls
	query.
		JoinRef("userId", "name").
		JoinRef("categoryId").
		JoinRefMany("tagIds")

	if query.err != nil {
		t.Fatalf("unexpected error: %v", query.err)
	}

	if len(query.populates) != 3 {
		t.Errorf("expected 3 populates, got %d", len(query.populates))
	}

	// Verify each populate
	if query.populates[0].From != "users" {
		t.Errorf("first populate From = %s, want users", query.populates[0].From)
	}
	if query.populates[1].From != "categories" {
		t.Errorf("second populate From = %s, want categories", query.populates[1].From)
	}
	if query.populates[2].From != "tags" {
		t.Errorf("third populate From = %s, want tags", query.populates[2].From)
	}
}

// Mock-based integration tests

func TestPopulateQuery_All_WithMock(t *testing.T) {
	ctx := context.Background()

	// Setup mock with sample data
	userID := primitive.NewObjectID()
	mock := &mockAggregator{
		results: []bson.M{
			{
				"_id":    primitive.NewObjectID(),
				"name":   "iPhone 15",
				"userId": userID,
				"user": bson.M{
					"_id":   userID,
					"name":  "John Doe",
					"email": "john@example.com",
				},
			},
		},
	}

	query := &PopulateQuery{
		aggregator: mock,
		filter:     bson.M{"status": "active"},
		populates:  []PopulateOptions{Populate("userId", "users", "name", "email")},
	}

	results, err := query.All(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0]["name"] != "iPhone 15" {
		t.Errorf("expected name 'iPhone 15', got %v", results[0]["name"])
	}

	// Verify user was populated
	user, ok := results[0]["user"].(bson.M)
	if !ok {
		t.Fatal("expected user to be populated")
	}
	if user["name"] != "John Doe" {
		t.Errorf("expected user name 'John Doe', got %v", user["name"])
	}
}

func TestPopulateQuery_One_WithMock(t *testing.T) {
	ctx := context.Background()

	mock := &mockAggregator{
		results: []bson.M{
			{"_id": primitive.NewObjectID(), "name": "Product 1"},
		},
	}

	query := &PopulateQuery{
		aggregator: mock,
		filter:     bson.M{},
		populates:  []PopulateOptions{},
	}

	result, err := query.One(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["name"] != "Product 1" {
		t.Errorf("expected name 'Product 1', got %v", result["name"])
	}
}

func TestPopulateQuery_One_NoResults(t *testing.T) {
	ctx := context.Background()

	mock := &mockAggregator{
		results: []bson.M{}, // Empty results
	}

	query := &PopulateQuery{
		aggregator: mock,
		filter:     bson.M{},
		populates:  []PopulateOptions{},
	}

	_, err := query.One(ctx)
	if err != ErrNoDocuments {
		t.Errorf("expected ErrNoDocuments, got %v", err)
	}
}

func TestPopulateQuery_All_Error(t *testing.T) {
	ctx := context.Background()

	mock := &mockAggregator{
		err: errors.New("database error"),
	}

	query := &PopulateQuery{
		aggregator: mock,
		filter:     bson.M{},
		populates:  []PopulateOptions{},
	}

	_, err := query.All(ctx)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestPopulateQuery_All_WithSortSkipLimit(t *testing.T) {
	ctx := context.Background()

	mock := &mockAggregator{
		results: []bson.M{
			{"name": "Product 1"},
			{"name": "Product 2"},
		},
	}

	query := &PopulateQuery{
		aggregator: mock,
		filter:     bson.M{"status": "active"},
		populates:  []PopulateOptions{Populate("userId", "users")},
		sort:       bson.D{{Key: "createdAt", Value: -1}},
		skip:       10,
		limit:      20,
	}

	results, err := query.All(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// Verify pipeline was built correctly
	pipeline, ok := mock.lastPipeline.(bson.A)
	if !ok {
		t.Fatal("expected pipeline to be bson.A")
	}

	// Should have: $match, $sort, $skip, $limit, $lookup, $unwind
	if len(pipeline) < 6 {
		t.Errorf("expected at least 6 pipeline stages, got %d", len(pipeline))
	}

	// Verify $match
	matchStage := pipeline[0].(bson.M)
	if _, ok := matchStage["$match"]; !ok {
		t.Error("expected first stage to be $match")
	}

	// Verify $sort
	sortStage := pipeline[1].(bson.M)
	if _, ok := sortStage["$sort"]; !ok {
		t.Error("expected second stage to be $sort")
	}

	// Verify $skip
	skipStage := pipeline[2].(bson.M)
	if _, ok := skipStage["$skip"]; !ok {
		t.Error("expected third stage to be $skip")
	}

	// Verify $limit
	limitStage := pipeline[3].(bson.M)
	if _, ok := limitStage["$limit"]; !ok {
		t.Error("expected fourth stage to be $limit")
	}
}

func TestPopulateQuery_All_NilFilter(t *testing.T) {
	ctx := context.Background()

	mock := &mockAggregator{
		results: []bson.M{{"name": "test"}},
	}

	query := &PopulateQuery{
		aggregator: mock,
		filter:     nil, // nil filter
		populates:  []PopulateOptions{},
	}

	_, err := query.All(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use empty filter
	pipeline := mock.lastPipeline.(bson.A)
	matchStage := pipeline[0].(bson.M)
	match := matchStage["$match"].(bson.M)
	if len(match) != 0 {
		t.Errorf("expected empty match filter, got %v", match)
	}
}

func TestPopulateQuery_All_DeferredError(t *testing.T) {
	ctx := context.Background()

	query := &PopulateQuery{
		aggregator: &mockAggregator{results: []bson.M{}},
		filter:     bson.M{},
		err:        errors.New("deferred error from JoinRef"),
	}

	_, err := query.All(ctx)
	if err == nil || err.Error() != "deferred error from JoinRef" {
		t.Errorf("expected deferred error, got %v", err)
	}
}

func TestPopulateQuery_WithAggregator(t *testing.T) {
	mock := &mockAggregator{results: []bson.M{}}

	query := &PopulateQuery{
		filter:    bson.M{},
		populates: []PopulateOptions{},
	}

	// Initially no aggregator
	if query.aggregator != nil {
		t.Error("expected nil aggregator initially")
	}

	// Set aggregator
	query.WithAggregator(mock)

	if query.aggregator != mock {
		t.Error("expected aggregator to be set")
	}
}

func TestPopulateQuery_MultiplePopulates(t *testing.T) {
	ctx := context.Background()

	categoryID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	mock := &mockAggregator{
		results: []bson.M{
			{
				"_id":        primitive.NewObjectID(),
				"name":       "iPhone 15",
				"userId":     userID,
				"categoryId": categoryID,
				"user":       bson.M{"_id": userID, "name": "John"},
				"category":   bson.M{"_id": categoryID, "name": "Electronics"},
			},
		},
	}

	query := &PopulateQuery{
		aggregator: mock,
		filter:     bson.M{},
		populates: []PopulateOptions{
			Populate("userId", "users", "name"),
			Populate("categoryId", "categories", "name"),
		},
	}

	results, err := query.All(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify pipeline has lookups for both
	pipeline := mock.lastPipeline.(bson.A)

	// Count lookup stages
	lookupCount := 0
	for _, stage := range pipeline {
		if stageMap, ok := stage.(bson.M); ok {
			if _, hasLookup := stageMap["$lookup"]; hasLookup {
				lookupCount++
			}
		}
	}

	if lookupCount != 2 {
		t.Errorf("expected 2 $lookup stages, got %d", lookupCount)
	}

	// Verify results have populated fields
	if results[0]["user"] == nil {
		t.Error("expected user to be populated")
	}
	if results[0]["category"] == nil {
		t.Error("expected category to be populated")
	}
}

func TestPopulateQuery_EmptyResults(t *testing.T) {
	ctx := context.Background()

	mock := &mockAggregator{
		results: nil, // nil results
	}

	query := &PopulateQuery{
		aggregator: mock,
		filter:     bson.M{},
		populates:  []PopulateOptions{},
	}

	results, err := query.All(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return empty slice, not nil
	if results == nil {
		t.Error("expected empty slice, got nil")
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
