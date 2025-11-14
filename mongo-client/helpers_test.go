package mongoclient

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNewObjectID(t *testing.T) {
	id := NewObjectID()
	if id.IsZero() {
		t.Error("Expected non-zero ObjectID")
	}
}

func TestObjectIDFromHex(t *testing.T) {
	validHex := "507f1f77bcf86cd799439011"
	id, err := ObjectIDFromHex(validHex)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if id.Hex() != validHex {
		t.Errorf("Expected %s, got %s", validHex, id.Hex())
	}

	invalidHex := "invalid"
	_, err = ObjectIDFromHex(invalidHex)
	if err == nil {
		t.Error("Expected error for invalid hex")
	}
}

func TestIsValidObjectID(t *testing.T) {
	tests := []struct {
		hex   string
		valid bool
	}{
		{"507f1f77bcf86cd799439011", true},
		{"invalid", false},
		{"", false},
		{"507f1f77bcf86cd79943901", false}, // too short
	}

	for _, tt := range tests {
		t.Run(tt.hex, func(t *testing.T) {
			result := IsValidObjectID(tt.hex)
			if result != tt.valid {
				t.Errorf("IsValidObjectID(%s) = %v, want %v", tt.hex, result, tt.valid)
			}
		})
	}
}

func TestQueryBuilders(t *testing.T) {
	// Test Set
	update := Set(M{"name": "John", "age": 30})
	if update["$set"] == nil {
		t.Error("Expected $set operator")
	}

	// Test Inc
	update = Inc(M{"count": 1})
	if update["$inc"] == nil {
		t.Error("Expected $inc operator")
	}

	// Test Push
	update = Push("tags", "new-tag")
	if update["$push"] == nil {
		t.Error("Expected $push operator")
	}

	// Test Unset
	update = Unset("field1", "field2")
	if update["$unset"] == nil {
		t.Error("Expected $unset operator")
	}
}

func TestAggregationBuilders(t *testing.T) {
	// Test Match
	stage := Match(M{"status": "active"})
	if stage["$match"] == nil {
		t.Error("Expected $match stage")
	}

	// Test Group
	stage = Group("$category", M{"total": M{"$sum": 1}})
	if stage["$group"] == nil {
		t.Error("Expected $group stage")
	}

	// Test Sort
	stage = Sort(M{"createdAt": -1})
	if stage["$sort"] == nil {
		t.Error("Expected $sort stage")
	}

	// Test Limit
	stage = Limit(10)
	if stage["$limit"] != int64(10) {
		t.Error("Expected $limit stage with value 10")
	}

	// Test Lookup
	stage = Lookup("users", "userId", "_id", "user")
	if stage["$lookup"] == nil {
		t.Error("Expected $lookup stage")
	}
}

func TestQueryOperators(t *testing.T) {
	// Test comparison operators
	cond := Gt(10)
	if cond["$gt"] != 10 {
		t.Error("Expected $gt operator with value 10")
	}

	cond = Gte(10)
	if cond["$gte"] != 10 {
		t.Error("Expected $gte operator with value 10")
	}

	cond = Lt(10)
	if cond["$lt"] != 10 {
		t.Error("Expected $lt operator with value 10")
	}

	cond = In("a", "b", "c")
	if cond["$in"] == nil {
		t.Error("Expected $in operator")
	}

	cond = Exists(true)
	if cond["$exists"] != true {
		t.Error("Expected $exists operator with value true")
	}
}

func TestPaginationOptions(t *testing.T) {
	// Test basic pagination
	p := &PaginationOptions{Page: 2, PageSize: 10}
	if p.GetSkip() != 10 {
		t.Errorf("Expected skip 10, got %d", p.GetSkip())
	}
	if p.GetLimit() != 10 {
		t.Errorf("Expected limit 10, got %d", p.GetLimit())
	}

	// Test with invalid page
	p = &PaginationOptions{Page: 0, PageSize: 10}
	if p.GetSkip() != 0 {
		t.Errorf("Expected skip 0 for invalid page, got %d", p.GetSkip())
	}

	// Test validation
	p = &PaginationOptions{Page: -1, PageSize: 200}
	p.Validate()
	if p.Page != 1 {
		t.Errorf("Expected page 1 after validation, got %d", p.Page)
	}
	if p.PageSize != 100 {
		t.Errorf("Expected page size 100 (max), got %d", p.PageSize)
	}
}

func TestTimeConversions(t *testing.T) {
	testTime := primitive.NewDateTimeFromTime(getTestTime())
	_ = testTime // Use the variable

	// Test timestamp conversions
	ts := TimeToTimestamp(getTestTime())
	if ts.T == 0 {
		t.Error("Expected non-zero timestamp")
	}

	converted := TimestampToTime(ts)
	if converted.Unix() != getTestTime().Unix() {
		t.Error("Expected timestamps to match")
	}
}

func getTestTime() time.Time {
	return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
}
