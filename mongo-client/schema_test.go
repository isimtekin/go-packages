package mongoclient

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Test models
type TestUserWithTimestamps struct {
	BaseModel
	Email string
	Name  string
}

type TestProductWithoutTimestamps struct {
	SimpleModel
	Name  string
	Price float64
}

func TestBaseModel_Timestamped(t *testing.T) {
	user := &TestUserWithTimestamps{
		Email: "test@example.com",
		Name:  "Test User",
	}

	// Verify it implements Timestamped interface
	var _ Timestamped = user
}

func TestBaseModel_SetCreatedAt(t *testing.T) {
	user := &TestUserWithTimestamps{}
	now := time.Now()

	user.SetCreatedAt(now)

	if user.CreatedAt != now {
		t.Errorf("Expected CreatedAt to be %v, got %v", now, user.CreatedAt)
	}

	retrieved := user.GetCreatedAt()
	if retrieved != now {
		t.Errorf("Expected GetCreatedAt to return %v, got %v", now, retrieved)
	}
}

func TestBaseModel_SetUpdatedAt(t *testing.T) {
	user := &TestUserWithTimestamps{}
	now := time.Now()

	user.SetUpdatedAt(now)

	if user.UpdatedAt != now {
		t.Errorf("Expected UpdatedAt to be %v, got %v", now, user.UpdatedAt)
	}

	retrieved := user.GetUpdatedAt()
	if retrieved != now {
		t.Errorf("Expected GetUpdatedAt to return %v, got %v", now, retrieved)
	}
}

func TestBaseModel_GetID(t *testing.T) {
	user := &TestUserWithTimestamps{}
	user.ID = primitive.NewObjectID()

	idStr := user.GetID()
	if idStr == "" {
		t.Error("Expected non-empty ID string")
	}

	if idStr != user.ID.Hex() {
		t.Errorf("Expected ID string to be %s, got %s", user.ID.Hex(), idStr)
	}
}

func TestBaseModel_SetID(t *testing.T) {
	user := &TestUserWithTimestamps{}
	validID := "507f1f77bcf86cd799439011"

	err := user.SetID(validID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if user.ID.Hex() != validID {
		t.Errorf("Expected ID to be %s, got %s", validID, user.ID.Hex())
	}

	// Test invalid ID
	err = user.SetID("invalid")
	if err == nil {
		t.Error("Expected error for invalid ID")
	}
}

func TestSimpleModel_NoTimestamps(t *testing.T) {
	product := &TestProductWithoutTimestamps{
		Name:  "Test Product",
		Price: 99.99,
	}

	// Verify it does NOT implement Timestamped interface
	_, ok := interface{}(product).(Timestamped)
	if ok {
		t.Error("SimpleModel should not implement Timestamped interface")
	}
}

func TestSimpleModel_GetID(t *testing.T) {
	product := &TestProductWithoutTimestamps{}
	product.ID = primitive.NewObjectID()

	idStr := product.GetID()
	if idStr == "" {
		t.Error("Expected non-empty ID string")
	}

	if idStr != product.ID.Hex() {
		t.Errorf("Expected ID string to be %s, got %s", product.ID.Hex(), idStr)
	}
}

func TestSimpleModel_SetID(t *testing.T) {
	product := &TestProductWithoutTimestamps{}
	validID := "507f1f77bcf86cd799439011"

	err := product.SetID(validID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if product.ID.Hex() != validID {
		t.Errorf("Expected ID to be %s, got %s", validID, product.ID.Hex())
	}
}

func TestApplyTimestamps_Insert_WithTimestamps(t *testing.T) {
	user := &TestUserWithTimestamps{
		Email: "test@example.com",
		Name:  "Test User",
	}

	// Verify timestamps are zero before
	if !user.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be zero before applyTimestamps")
	}
	if !user.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be zero before applyTimestamps")
	}

	// Apply timestamps for insert
	applyTimestamps(user, true)

	// Verify timestamps are set
	if user.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set after applyTimestamps")
	}
	if user.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set after applyTimestamps")
	}

	// Verify both timestamps are approximately the same
	diff := user.UpdatedAt.Sub(user.CreatedAt)
	if diff > time.Second {
		t.Errorf("Expected CreatedAt and UpdatedAt to be close, diff: %v", diff)
	}
}

func TestApplyTimestamps_Insert_PreservesExisting(t *testing.T) {
	pastTime := time.Now().Add(-24 * time.Hour)
	user := &TestUserWithTimestamps{
		Email: "test@example.com",
		Name:  "Test User",
	}
	user.CreatedAt = pastTime
	user.UpdatedAt = pastTime

	// Apply timestamps for insert
	applyTimestamps(user, true)

	// Verify existing timestamps are preserved
	if user.CreatedAt != pastTime {
		t.Error("Expected CreatedAt to be preserved")
	}
	if user.UpdatedAt != pastTime {
		t.Error("Expected UpdatedAt to be preserved")
	}
}

func TestApplyTimestamps_Update_OnlyUpdatesUpdatedAt(t *testing.T) {
	pastTime := time.Now().Add(-24 * time.Hour)
	user := &TestUserWithTimestamps{
		Email: "test@example.com",
		Name:  "Test User",
	}
	user.CreatedAt = pastTime
	user.UpdatedAt = pastTime

	// Sleep a tiny bit to ensure different timestamp
	time.Sleep(10 * time.Millisecond)

	// Apply timestamps for update
	applyTimestamps(user, false)

	// Verify CreatedAt is preserved
	if user.CreatedAt != pastTime {
		t.Errorf("Expected CreatedAt to be preserved as %v, got %v", pastTime, user.CreatedAt)
	}

	// Verify UpdatedAt is changed
	if user.UpdatedAt == pastTime {
		t.Error("Expected UpdatedAt to be updated")
	}

	// Verify UpdatedAt is recent
	if time.Since(user.UpdatedAt) > time.Second {
		t.Error("Expected UpdatedAt to be recent")
	}
}

func TestApplyTimestamps_WithoutTimestamps_NoEffect(t *testing.T) {
	product := &TestProductWithoutTimestamps{
		Name:  "Test Product",
		Price: 99.99,
	}

	// This should have no effect since Product doesn't implement Timestamped
	applyTimestamps(product, true)

	// Verify no panic and product is unchanged
	if product.Name != "Test Product" {
		t.Error("Product should remain unchanged")
	}
}

func TestApplyTimestamps_NilDocument(t *testing.T) {
	// Should not panic with nil
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("applyTimestamps should not panic with nil, got: %v", r)
		}
	}()

	applyTimestamps(nil, true)
}

func TestApplyTimestamps_NonPointer(t *testing.T) {
	// Should not panic with non-pointer
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("applyTimestamps should not panic with non-pointer, got: %v", r)
		}
	}()

	user := TestUserWithTimestamps{
		Email: "test@example.com",
		Name:  "Test User",
	}

	applyTimestamps(user, true)
}
