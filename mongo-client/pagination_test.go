package mongoclient

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestDefaultPaginationOptions(t *testing.T) {
	opts := DefaultPaginationOptions()

	if opts.Page != 1 {
		t.Errorf("expected Page=1, got %d", opts.Page)
	}
	if opts.PageSize != 10 {
		t.Errorf("expected PageSize=10, got %d", opts.PageSize)
	}
	if opts.Sort != nil {
		t.Errorf("expected Sort=nil, got %v", opts.Sort)
	}
	if opts.Projection != nil {
		t.Errorf("expected Projection=nil, got %v", opts.Projection)
	}
	if opts.SkipCount {
		t.Error("expected SkipCount=false")
	}
}

func TestDefaultCursorOptions(t *testing.T) {
	opts := DefaultCursorOptions()

	if opts.PageSize != 10 {
		t.Errorf("expected PageSize=10, got %d", opts.PageSize)
	}
	if opts.CursorField != "_id" {
		t.Errorf("expected CursorField='_id', got %s", opts.CursorField)
	}
	if opts.AfterCursor != "" {
		t.Errorf("expected AfterCursor='', got %s", opts.AfterCursor)
	}
	if opts.BeforeCursor != "" {
		t.Errorf("expected BeforeCursor='', got %s", opts.BeforeCursor)
	}
}

func TestEncodeCursor(t *testing.T) {
	tests := []struct {
		name  string
		field string
		value interface{}
	}{
		{"string value", "_id", "test123"},
		{"int value", "seq", 42},
		{"float value", "score", 3.14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := encodeCursor(tt.field, tt.value)
			if cursor == "" {
				t.Error("expected non-empty cursor")
			}

			// Verify it's valid base64
			decoded, err := base64.URLEncoding.DecodeString(cursor)
			if err != nil {
				t.Errorf("cursor is not valid base64: %v", err)
			}

			// Verify it's valid JSON
			var data map[string]interface{}
			if err := json.Unmarshal(decoded, &data); err != nil {
				t.Errorf("cursor is not valid JSON: %v", err)
			}
		})
	}
}

func TestEncodeCursorObjectID(t *testing.T) {
	oid := primitive.NewObjectID()
	cursor := encodeCursor("_id", oid)

	if cursor == "" {
		t.Error("expected non-empty cursor")
	}

	// Decode and verify ObjectID is stored as hex string
	decoded, _ := base64.URLEncoding.DecodeString(cursor)
	var data cursorData
	if err := json.Unmarshal(decoded, &data); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if data.ID != oid.Hex() {
		t.Errorf("expected ID=%s, got %s", oid.Hex(), data.ID)
	}
	if data.Value != nil {
		t.Errorf("expected Value=nil for ObjectID, got %v", data.Value)
	}
}

func TestDecodeCursor(t *testing.T) {
	tests := []struct {
		name  string
		field string
		value interface{}
	}{
		{"string value", "_id", "test123"},
		{"int value", "seq", float64(42)}, // JSON numbers decode as float64
		{"float value", "score", 3.14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode first
			cursor := encodeCursor(tt.field, tt.value)

			// Then decode
			data, err := decodeCursor(cursor)
			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if data.Field != tt.field {
				t.Errorf("expected Field=%s, got %s", tt.field, data.Field)
			}
			if data.Value != tt.value {
				t.Errorf("expected Value=%v, got %v", tt.value, data.Value)
			}
		})
	}
}

func TestDecodeCursorObjectID(t *testing.T) {
	oid := primitive.NewObjectID()
	cursor := encodeCursor("_id", oid)

	data, err := decodeCursor(cursor)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if data.Field != "_id" {
		t.Errorf("expected Field='_id', got %s", data.Field)
	}

	decodedOID, ok := data.Value.(primitive.ObjectID)
	if !ok {
		t.Fatalf("expected ObjectID, got %T", data.Value)
	}
	if decodedOID != oid {
		t.Errorf("expected %v, got %v", oid, decodedOID)
	}
}

func TestDecodeCursorInvalid(t *testing.T) {
	tests := []struct {
		name   string
		cursor string
	}{
		{"empty", ""},
		{"invalid base64", "not-valid-base64!!!"},
		{"invalid json", base64.URLEncoding.EncodeToString([]byte("not json"))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decodeCursor(tt.cursor)
			if err == nil {
				t.Error("expected error for invalid cursor")
			}
		})
	}
}

func TestBuildPaginationOptions(t *testing.T) {
	config := map[string]interface{}{
		"page":       2,
		"pageSize":   20,
		"sort":       bson.D{{Key: "createdAt", Value: -1}},
		"projection": bson.M{"password": 0},
		"skipCount":  true,
	}

	opts := BuildPaginationOptions(config)

	if opts.Page != 2 {
		t.Errorf("expected Page=2, got %d", opts.Page)
	}
	if opts.PageSize != 20 {
		t.Errorf("expected PageSize=20, got %d", opts.PageSize)
	}
	if opts.Sort == nil {
		t.Error("expected Sort to be set")
	}
	if opts.Projection == nil {
		t.Error("expected Projection to be set")
	}
	if !opts.SkipCount {
		t.Error("expected SkipCount=true")
	}
}

func TestBuildPaginationOptionsDefaults(t *testing.T) {
	config := map[string]interface{}{}

	opts := BuildPaginationOptions(config)

	if opts.Page != 1 {
		t.Errorf("expected Page=1, got %d", opts.Page)
	}
	if opts.PageSize != 10 {
		t.Errorf("expected PageSize=10, got %d", opts.PageSize)
	}
}

func TestBuildPaginationOptionsFloat64(t *testing.T) {
	// JSON numbers are often decoded as float64
	config := map[string]interface{}{
		"page":     float64(3),
		"pageSize": float64(25),
	}

	opts := BuildPaginationOptions(config)

	if opts.Page != 3 {
		t.Errorf("expected Page=3, got %d", opts.Page)
	}
	if opts.PageSize != 25 {
		t.Errorf("expected PageSize=25, got %d", opts.PageSize)
	}
}

func TestBuildPaginationOptionsInt64(t *testing.T) {
	config := map[string]interface{}{
		"page":     int64(4),
		"pageSize": int64(30),
	}

	opts := BuildPaginationOptions(config)

	if opts.Page != 4 {
		t.Errorf("expected Page=4, got %d", opts.Page)
	}
	if opts.PageSize != 30 {
		t.Errorf("expected PageSize=30, got %d", opts.PageSize)
	}
}

func TestBuildCursorOptions(t *testing.T) {
	cursor := encodeCursor("_id", "test123")
	config := map[string]interface{}{
		"pageSize":    15,
		"afterCursor": cursor,
		"sort":        bson.D{{Key: "createdAt", Value: -1}},
		"projection":  bson.M{"password": 0},
		"cursorField": "createdAt",
	}

	opts := BuildCursorOptions(config)

	if opts.PageSize != 15 {
		t.Errorf("expected PageSize=15, got %d", opts.PageSize)
	}
	if opts.AfterCursor != cursor {
		t.Errorf("expected AfterCursor=%s, got %s", cursor, opts.AfterCursor)
	}
	if opts.Sort == nil {
		t.Error("expected Sort to be set")
	}
	if opts.Projection == nil {
		t.Error("expected Projection to be set")
	}
	if opts.CursorField != "createdAt" {
		t.Errorf("expected CursorField='createdAt', got %s", opts.CursorField)
	}
}

func TestBuildCursorOptionsBeforeCursor(t *testing.T) {
	cursor := encodeCursor("_id", "test123")
	config := map[string]interface{}{
		"beforeCursor": cursor,
	}

	opts := BuildCursorOptions(config)

	if opts.BeforeCursor != cursor {
		t.Errorf("expected BeforeCursor=%s, got %s", cursor, opts.BeforeCursor)
	}
}

func TestPaginatedResultMetadata(t *testing.T) {
	// Test page 1 of 5 (total 45, pageSize 10)
	result := &PaginatedResult[bson.M]{
		Documents:   make([]bson.M, 10),
		Total:       45,
		Page:        1,
		PageSize:    10,
		TotalPages:  5,
		HasNextPage: true,
		HasPrevPage: false,
		From:        1,
		To:          10,
	}

	if !result.HasNextPage {
		t.Error("expected HasNextPage=true for page 1")
	}
	if result.HasPrevPage {
		t.Error("expected HasPrevPage=false for page 1")
	}
	if result.From != 1 {
		t.Errorf("expected From=1, got %d", result.From)
	}
	if result.To != 10 {
		t.Errorf("expected To=10, got %d", result.To)
	}
}

func TestPaginatedResultLastPage(t *testing.T) {
	// Test page 5 of 5 (total 45, pageSize 10, last page has 5 docs)
	result := &PaginatedResult[bson.M]{
		Documents:   make([]bson.M, 5),
		Total:       45,
		Page:        5,
		PageSize:    10,
		TotalPages:  5,
		HasNextPage: false,
		HasPrevPage: true,
		From:        41,
		To:          45,
	}

	if result.HasNextPage {
		t.Error("expected HasNextPage=false for last page")
	}
	if !result.HasPrevPage {
		t.Error("expected HasPrevPage=true for last page")
	}
	if result.From != 41 {
		t.Errorf("expected From=41, got %d", result.From)
	}
	if result.To != 45 {
		t.Errorf("expected To=45, got %d", result.To)
	}
}

func TestCursorResultMetadata(t *testing.T) {
	result := &CursorResult[bson.M]{
		Documents:   make([]bson.M, 10),
		PageSize:    10,
		NextCursor:  "next123",
		PrevCursor:  "prev123",
		HasNextPage: true,
		HasPrevPage: true,
	}

	if !result.HasNextPage {
		t.Error("expected HasNextPage=true")
	}
	if !result.HasPrevPage {
		t.Error("expected HasPrevPage=true")
	}
	if result.NextCursor != "next123" {
		t.Errorf("expected NextCursor='next123', got %s", result.NextCursor)
	}
	if result.PrevCursor != "prev123" {
		t.Errorf("expected PrevCursor='prev123', got %s", result.PrevCursor)
	}
}

// Integration tests would require a running MongoDB instance
// These would be tested separately with a test container or mock
