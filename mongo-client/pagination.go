package mongoclient

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PaginatedResult contains documents and pagination metadata
type PaginatedResult[T any] struct {
	// Documents for the current page
	Documents []T `json:"documents"`

	// Total number of documents matching the filter (0 if SkipCount was true)
	Total int64 `json:"total"`

	// Current page number (1-indexed)
	Page int64 `json:"page"`

	// Number of documents per page
	PageSize int64 `json:"pageSize"`

	// Total number of pages (0 if SkipCount was true)
	TotalPages int64 `json:"totalPages"`

	// Navigation helpers
	HasNextPage bool `json:"hasNextPage"`
	HasPrevPage bool `json:"hasPrevPage"`

	// Range indicators (1-indexed)
	From int64 `json:"from"` // First document index on this page
	To   int64 `json:"to"`   // Last document index on this page
}

// DefaultPaginationOptions returns default pagination options
func DefaultPaginationOptions() *PaginationOptions {
	return &PaginationOptions{
		Page:     1,
		PageSize: 10,
	}
}

// CursorResult contains documents and cursor-based pagination metadata
type CursorResult[T any] struct {
	// Documents for the current page
	Documents []T `json:"documents"`

	// Number of documents returned
	PageSize int64 `json:"pageSize"`

	// Cursor for the next page (empty if no more pages)
	NextCursor string `json:"nextCursor,omitempty"`

	// Cursor for the previous page (empty if on first page)
	PrevCursor string `json:"prevCursor,omitempty"`

	// Navigation helpers
	HasNextPage bool `json:"hasNextPage"`
	HasPrevPage bool `json:"hasPrevPage"`
}

// CursorOptions configures cursor-based pagination
type CursorOptions struct {
	// Number of documents per page (default: 10)
	PageSize int64

	// Cursor to start after (for forward pagination)
	AfterCursor string

	// Cursor to start before (for backward pagination)
	BeforeCursor string

	// Sort specification - REQUIRED for cursor pagination
	// The cursor field must be included in the sort
	Sort interface{}

	// Projection to include/exclude fields
	Projection interface{}

	// CursorField is the field used for cursor (default: "_id")
	CursorField string
}

// DefaultCursorOptions returns default cursor options
func DefaultCursorOptions() *CursorOptions {
	return &CursorOptions{
		PageSize:    10,
		CursorField: "_id",
	}
}

// cursorData holds the encoded cursor information
type cursorData struct {
	Field string      `json:"f"`
	Value interface{} `json:"v"`
	ID    string      `json:"id,omitempty"`
}

// FindWithPagination finds documents with full pagination metadata
func (c *Collection) FindWithPagination(ctx context.Context, filter interface{}, opts *PaginationOptions) (*PaginatedResult[bson.M], error) {
	if opts == nil {
		opts = DefaultPaginationOptions()
	}

	// Normalize page and pageSize
	page := opts.Page
	if page < 1 {
		page = 1
	}
	pageSize := opts.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 1000 {
		pageSize = 1000 // safety limit
	}

	var total int64
	var totalPages int64

	// Count total documents (unless SkipCount is true)
	if !opts.SkipCount {
		var err error
		total, err = c.CountDocuments(ctx, filter)
		if err != nil {
			return nil, fmt.Errorf("count failed: %w", err)
		}
		totalPages = (total + pageSize - 1) / pageSize
	}

	// Build find options
	skip := (page - 1) * pageSize
	findOpts := options.Find().SetSkip(skip).SetLimit(pageSize)

	if opts.Sort != nil {
		findOpts.SetSort(opts.Sort)
	}
	if opts.Projection != nil {
		findOpts.SetProjection(opts.Projection)
	}

	// Execute find
	var results []bson.M
	err := c.FindAll(ctx, filter, &results, findOpts)
	if err != nil {
		return nil, fmt.Errorf("find failed: %w", err)
	}

	if results == nil {
		results = []bson.M{}
	}

	// Calculate navigation metadata
	docCount := int64(len(results))
	from := int64(0)
	to := int64(0)
	hasNextPage := false
	hasPrevPage := page > 1

	if docCount > 0 {
		from = skip + 1
		to = skip + docCount
	}

	if opts.SkipCount {
		// When count is skipped, check if there are more documents
		hasNextPage = docCount == pageSize
	} else {
		hasNextPage = page < totalPages
	}

	return &PaginatedResult[bson.M]{
		Documents:   results,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		HasNextPage: hasNextPage,
		HasPrevPage: hasPrevPage,
		From:        from,
		To:          to,
	}, nil
}

// FindWithCursor finds documents using cursor-based pagination
// This is more efficient for large datasets and real-time data
func (c *Collection) FindWithCursor(ctx context.Context, filter interface{}, opts *CursorOptions) (*CursorResult[bson.M], error) {
	if opts == nil {
		opts = DefaultCursorOptions()
	}

	pageSize := opts.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 1000 {
		pageSize = 1000
	}

	cursorField := opts.CursorField
	if cursorField == "" {
		cursorField = "_id"
	}

	// Build filter with cursor condition
	finalFilter := bson.M{}
	if filter != nil {
		if f, ok := filter.(bson.M); ok {
			for k, v := range f {
				finalFilter[k] = v
			}
		} else if f, ok := filter.(map[string]interface{}); ok {
			for k, v := range f {
				finalFilter[k] = v
			}
		}
	}

	// Decode cursor and add to filter
	hasPrevPage := false
	if opts.AfterCursor != "" {
		cursor, err := decodeCursor(opts.AfterCursor)
		if err != nil {
			return nil, fmt.Errorf("invalid afterCursor: %w", err)
		}
		finalFilter[cursorField] = bson.M{"$gt": cursor.Value}
		hasPrevPage = true
	} else if opts.BeforeCursor != "" {
		cursor, err := decodeCursor(opts.BeforeCursor)
		if err != nil {
			return nil, fmt.Errorf("invalid beforeCursor: %w", err)
		}
		finalFilter[cursorField] = bson.M{"$lt": cursor.Value}
	}

	// Build find options - fetch one extra to detect hasNextPage
	findOpts := options.Find().SetLimit(pageSize + 1)

	if opts.Sort != nil {
		findOpts.SetSort(opts.Sort)
	} else {
		// Default sort by cursor field ascending
		findOpts.SetSort(bson.D{{Key: cursorField, Value: 1}})
	}

	if opts.Projection != nil {
		findOpts.SetProjection(opts.Projection)
	}

	// Execute find
	var results []bson.M
	err := c.FindAll(ctx, finalFilter, &results, findOpts)
	if err != nil {
		return nil, fmt.Errorf("find failed: %w", err)
	}

	if results == nil {
		results = []bson.M{}
	}

	// Check if there are more pages
	hasNextPage := int64(len(results)) > pageSize
	if hasNextPage {
		results = results[:pageSize] // Remove the extra document
	}

	// Generate cursors
	var nextCursor, prevCursor string
	if len(results) > 0 {
		// Next cursor from the last document
		lastDoc := results[len(results)-1]
		nextCursor = encodeCursor(cursorField, lastDoc[cursorField])

		// Prev cursor from the first document
		firstDoc := results[0]
		prevCursor = encodeCursor(cursorField, firstDoc[cursorField])
	}

	return &CursorResult[bson.M]{
		Documents:   results,
		PageSize:    pageSize,
		NextCursor:  nextCursor,
		PrevCursor:  prevCursor,
		HasNextPage: hasNextPage,
		HasPrevPage: hasPrevPage,
	}, nil
}

// Model pagination methods

// FindWithPagination finds documents with full pagination metadata (Model method)
func (m *Model) FindWithPagination(ctx context.Context, filter interface{}, opts *PaginationOptions) (*PaginatedResult[bson.M], error) {
	return m.collection.FindWithPagination(ctx, filter, opts)
}

// FindWithCursor finds documents using cursor-based pagination (Model method)
func (m *Model) FindWithCursor(ctx context.Context, filter interface{}, opts *CursorOptions) (*CursorResult[bson.M], error) {
	return m.collection.FindWithCursor(ctx, filter, opts)
}

// Helper functions

// encodeCursor creates a base64-encoded cursor from field and value
func encodeCursor(field string, value interface{}) string {
	data := cursorData{
		Field: field,
		Value: value,
	}

	// Handle ObjectID specially
	if oid, ok := value.(primitive.ObjectID); ok {
		data.ID = oid.Hex()
		data.Value = nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	return base64.URLEncoding.EncodeToString(jsonData)
}

// decodeCursor decodes a base64-encoded cursor
func decodeCursor(cursor string) (*cursorData, error) {
	jsonData, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor encoding: %w", err)
	}

	var data cursorData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("invalid cursor format: %w", err)
	}

	// Restore ObjectID
	if data.ID != "" {
		oid, err := primitive.ObjectIDFromHex(data.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor ID: %w", err)
		}
		data.Value = oid
	}

	return &data, nil
}

// BuildPaginationOptions is a helper to build PaginationOptions from a map
func BuildPaginationOptions(config map[string]interface{}) *PaginationOptions {
	opts := DefaultPaginationOptions()

	if page, ok := config["page"].(int); ok {
		opts.Page = int64(page)
	} else if page, ok := config["page"].(int64); ok {
		opts.Page = page
	} else if page, ok := config["page"].(float64); ok {
		opts.Page = int64(page)
	}

	if pageSize, ok := config["pageSize"].(int); ok {
		opts.PageSize = int64(pageSize)
	} else if pageSize, ok := config["pageSize"].(int64); ok {
		opts.PageSize = pageSize
	} else if pageSize, ok := config["pageSize"].(float64); ok {
		opts.PageSize = int64(pageSize)
	}

	if sort, ok := config["sort"]; ok {
		opts.Sort = sort
	}

	if projection, ok := config["projection"]; ok {
		opts.Projection = projection
	}

	if skipCount, ok := config["skipCount"].(bool); ok {
		opts.SkipCount = skipCount
	}

	return opts
}

// BuildCursorOptions is a helper to build CursorOptions from a map
func BuildCursorOptions(config map[string]interface{}) *CursorOptions {
	opts := DefaultCursorOptions()

	if pageSize, ok := config["pageSize"].(int); ok {
		opts.PageSize = int64(pageSize)
	} else if pageSize, ok := config["pageSize"].(int64); ok {
		opts.PageSize = pageSize
	} else if pageSize, ok := config["pageSize"].(float64); ok {
		opts.PageSize = int64(pageSize)
	}

	if afterCursor, ok := config["afterCursor"].(string); ok {
		opts.AfterCursor = afterCursor
	}

	if beforeCursor, ok := config["beforeCursor"].(string); ok {
		opts.BeforeCursor = beforeCursor
	}

	if sort, ok := config["sort"]; ok {
		opts.Sort = sort
	}

	if projection, ok := config["projection"]; ok {
		opts.Projection = projection
	}

	if cursorField, ok := config["cursorField"].(string); ok {
		opts.CursorField = cursorField
	}

	return opts
}
