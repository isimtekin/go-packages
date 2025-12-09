package mongoclient

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestHookContext_SetGet(t *testing.T) {
	hc := &HookContext{}

	// Test Set and Get
	hc.Set("key1", "value1")
	hc.Set("key2", 42)

	v1, ok := hc.Get("key1")
	if !ok || v1 != "value1" {
		t.Errorf("expected 'value1', got %v", v1)
	}

	v2, ok := hc.Get("key2")
	if !ok || v2 != 42 {
		t.Errorf("expected 42, got %v", v2)
	}

	// Test GetString
	s := hc.GetString("key1")
	if s != "value1" {
		t.Errorf("expected 'value1', got %s", s)
	}

	// Test missing key
	_, ok = hc.Get("missing")
	if ok {
		t.Error("expected false for missing key")
	}
}

func TestSchema_PreHook(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString)

	var called bool
	schema.Pre("insert", func(ctx context.Context, hc *HookContext) error {
		called = true
		if hc.Operation != OpInsert {
			t.Errorf("expected operation 'insert', got '%s'", hc.Operation)
		}
		return nil
	})

	// Create hook context and execute
	hc := &HookContext{
		Operation: OpInsert,
		Document:  map[string]interface{}{"name": "test"},
		Schema:    schema,
	}

	err := schema.executePreHooks(context.Background(), OpInsert, hc)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !called {
		t.Error("pre hook was not called")
	}
}

func TestSchema_PostHook(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString)

	var called int32
	schema.Post("insert", func(ctx context.Context, hc *HookContext) error {
		atomic.AddInt32(&called, 1)
		return nil
	}, HookOptions{Async: BoolPtr(false)}) // Sync for testing

	hc := &HookContext{
		Operation: OpInsert,
		Document:  map[string]interface{}{"name": "test"},
		Schema:    schema,
	}

	err := schema.executePostHooks(context.Background(), OpInsert, hc)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if atomic.LoadInt32(&called) != 1 {
		t.Error("post hook was not called")
	}
}

func TestSchema_PreHook_Error(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString)

	expectedErr := errors.New("pre hook error")
	schema.Pre("insert", func(ctx context.Context, hc *HookContext) error {
		return expectedErr
	})

	hc := &HookContext{
		Operation: OpInsert,
		Document:  map[string]interface{}{"name": "test"},
		Schema:    schema,
	}

	err := schema.executePreHooks(context.Background(), OpInsert, hc)
	if err == nil {
		t.Error("expected error from pre hook")
	}
}

func TestSchema_PreHook_ContinueOnError(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString)

	var secondCalled bool
	schema.Pre("insert", func(ctx context.Context, hc *HookContext) error {
		return errors.New("first hook error")
	}, HookOptions{ContinueOnError: BoolPtr(true)})

	schema.Pre("insert", func(ctx context.Context, hc *HookContext) error {
		secondCalled = true
		return nil
	})

	hc := &HookContext{
		Operation: OpInsert,
		Schema:    schema,
	}

	err := schema.executePreHooks(context.Background(), OpInsert, hc)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !secondCalled {
		t.Error("second hook should have been called")
	}
}

func TestSchema_WildcardHook(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString)

	var operations []string
	var mu sync.Mutex

	schema.Pre("*", func(ctx context.Context, hc *HookContext) error {
		mu.Lock()
		operations = append(operations, hc.Operation)
		mu.Unlock()
		return nil
	})

	// Test with insert
	hc1 := &HookContext{Operation: OpInsert, Schema: schema}
	_ = schema.executePreHooks(context.Background(), OpInsert, hc1)

	// Test with update
	hc2 := &HookContext{Operation: OpUpdate, Schema: schema}
	_ = schema.executePreHooks(context.Background(), OpUpdate, hc2)

	// Test with delete
	hc3 := &HookContext{Operation: OpDelete, Schema: schema}
	_ = schema.executePreHooks(context.Background(), OpDelete, hc3)

	if len(operations) != 3 {
		t.Errorf("expected 3 operations, got %d", len(operations))
	}
}

func TestSchema_PreWhen_Condition(t *testing.T) {
	schema := NewSchema().
		Field("password", TypeString)

	var called bool
	schema.PreWhen("update", func(ctx context.Context, hc *HookContext) error {
		called = true
		return nil
	}, func(ctx context.Context, hc *HookContext) bool {
		// Only run if password is being updated
		if hc.Update == nil {
			return false
		}
		if set, ok := hc.Update["$set"].(M); ok {
			_, hasPassword := set["password"]
			return hasPassword
		}
		return false
	})

	// Test with password update - should trigger
	hc1 := &HookContext{
		Operation: OpUpdate,
		Update:    M{"$set": M{"password": "newpass"}},
		Schema:    schema,
	}
	_ = schema.executePreHooks(context.Background(), OpUpdate, hc1)

	if !called {
		t.Error("hook should have been called when password is updated")
	}

	// Reset and test without password - should not trigger
	called = false
	hc2 := &HookContext{
		Operation: OpUpdate,
		Update:    M{"$set": M{"name": "newname"}},
		Schema:    schema,
	}
	_ = schema.executePreHooks(context.Background(), OpUpdate, hc2)

	if called {
		t.Error("hook should not have been called when password is not updated")
	}
}

func TestSchema_HookPriority(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString)

	var order []int
	var mu sync.Mutex

	// Add hooks in reverse priority order
	schema.Pre("insert", func(ctx context.Context, hc *HookContext) error {
		mu.Lock()
		order = append(order, 3)
		mu.Unlock()
		return nil
	}, HookOptions{Priority: 3})

	schema.Pre("insert", func(ctx context.Context, hc *HookContext) error {
		mu.Lock()
		order = append(order, 1)
		mu.Unlock()
		return nil
	}, HookOptions{Priority: 1})

	schema.Pre("insert", func(ctx context.Context, hc *HookContext) error {
		mu.Lock()
		order = append(order, 2)
		mu.Unlock()
		return nil
	}, HookOptions{Priority: 2})

	hc := &HookContext{Operation: OpInsert, Schema: schema}
	_ = schema.executePreHooks(context.Background(), OpInsert, hc)

	// Check order: should be 1, 2, 3
	if len(order) != 3 {
		t.Fatalf("expected 3 hooks, got %d", len(order))
	}
	if order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Errorf("expected order [1,2,3], got %v", order)
	}
}

func TestSchema_HookFIFO(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString)

	var order []int
	var mu sync.Mutex

	// Add hooks with same priority - should maintain FIFO order
	schema.Pre("insert", func(ctx context.Context, hc *HookContext) error {
		mu.Lock()
		order = append(order, 1)
		mu.Unlock()
		return nil
	})

	schema.Pre("insert", func(ctx context.Context, hc *HookContext) error {
		mu.Lock()
		order = append(order, 2)
		mu.Unlock()
		return nil
	})

	schema.Pre("insert", func(ctx context.Context, hc *HookContext) error {
		mu.Lock()
		order = append(order, 3)
		mu.Unlock()
		return nil
	})

	hc := &HookContext{Operation: OpInsert, Schema: schema}
	_ = schema.executePreHooks(context.Background(), OpInsert, hc)

	// Check order: should be 1, 2, 3 (FIFO)
	if len(order) != 3 {
		t.Fatalf("expected 3 hooks, got %d", len(order))
	}
	if order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Errorf("expected order [1,2,3], got %v", order)
	}
}

func TestSchema_AsyncPostHook(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString)

	var called int32
	var wg sync.WaitGroup
	wg.Add(1)

	schema.Post("insert", func(ctx context.Context, hc *HookContext) error {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond) // Simulate async work
		atomic.AddInt32(&called, 1)
		return nil
	}, HookOptions{Async: BoolPtr(true)})

	hc := &HookContext{Operation: OpInsert, Schema: schema}
	_ = schema.executePostHooks(context.Background(), OpInsert, hc)

	// Should return immediately
	if atomic.LoadInt32(&called) != 0 {
		// Hook might not have started yet, that's okay
	}

	// Wait for async hook to complete
	wg.Wait()

	if atomic.LoadInt32(&called) != 1 {
		t.Error("async hook was not called")
	}
}

func TestSchema_ClearHooks(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString)

	var called bool
	schema.Pre("insert", func(ctx context.Context, hc *HookContext) error {
		called = true
		return nil
	})

	schema.ClearHooks()

	hc := &HookContext{Operation: OpInsert, Schema: schema}
	_ = schema.executePreHooks(context.Background(), OpInsert, hc)

	if called {
		t.Error("hook should not have been called after ClearHooks")
	}
}

func TestHookRegistry(t *testing.T) {
	// Register a hook
	RegisterHook("testHook", func(ctx context.Context, hc *HookContext) error {
		return nil
	})
	defer UnregisterHook("testHook")

	// Get the hook
	hook, ok := GetRegisteredHook("testHook")
	if !ok {
		t.Error("hook should be found")
	}
	if hook == nil {
		t.Error("hook should not be nil")
	}

	// List hooks
	hooks := ListRegisteredHooks()
	found := false
	for _, name := range hooks {
		if name == "testHook" {
			found = true
			break
		}
	}
	if !found {
		t.Error("testHook should be in the list")
	}

	// Unregister
	UnregisterHook("testHook")
	_, ok = GetRegisteredHook("testHook")
	if ok {
		t.Error("hook should not be found after unregister")
	}
}

func TestCondition_Evaluate(t *testing.T) {
	tests := []struct {
		name     string
		cond     Condition
		hc       *HookContext
		expected bool
	}{
		{
			name: "field exists - true",
			cond: Condition{Field: "name", Exists: BoolPtr(true)},
			hc: &HookContext{
				Document: map[string]interface{}{"name": "test"},
			},
			expected: true,
		},
		{
			name: "field exists - false",
			cond: Condition{Field: "missing", Exists: BoolPtr(true)},
			hc: &HookContext{
				Document: map[string]interface{}{"name": "test"},
			},
			expected: false,
		},
		{
			name: "field eq",
			cond: Condition{Field: "role", Eq: "admin"},
			hc: &HookContext{
				Document: map[string]interface{}{"role": "admin"},
			},
			expected: true,
		},
		{
			name: "field ne",
			cond: Condition{Field: "role", Ne: "admin"},
			hc: &HookContext{
				Document: map[string]interface{}{"role": "user"},
			},
			expected: true,
		},
		{
			name: "field gt",
			cond: Condition{Field: "age", Gt: 18},
			hc: &HookContext{
				Document: map[string]interface{}{"age": 25},
			},
			expected: true,
		},
		{
			name: "field gte",
			cond: Condition{Field: "age", Gte: 18},
			hc: &HookContext{
				Document: map[string]interface{}{"age": 18},
			},
			expected: true,
		},
		{
			name: "field lt",
			cond: Condition{Field: "age", Lt: 18},
			hc: &HookContext{
				Document: map[string]interface{}{"age": 10},
			},
			expected: true,
		},
		{
			name: "field lte",
			cond: Condition{Field: "age", Lte: 18},
			hc: &HookContext{
				Document: map[string]interface{}{"age": 18},
			},
			expected: true,
		},
		{
			name: "field in",
			cond: Condition{Field: "role", In: []interface{}{"admin", "moderator"}},
			hc: &HookContext{
				Document: map[string]interface{}{"role": "admin"},
			},
			expected: true,
		},
		{
			name: "field nin",
			cond: Condition{Field: "role", Nin: []interface{}{"banned", "suspended"}},
			hc: &HookContext{
				Document: map[string]interface{}{"role": "admin"},
			},
			expected: true,
		},
		{
			name: "field matches regex",
			cond: Condition{Field: "email", Matches: ".*@example\\.com$"},
			hc: &HookContext{
				Document: map[string]interface{}{"email": "test@example.com"},
			},
			expected: true,
		},
		{
			name: "field notEmpty - true",
			cond: Condition{Field: "name", NotEmpty: BoolPtr(true)},
			hc: &HookContext{
				Document: map[string]interface{}{"name": "test"},
			},
			expected: true,
		},
		{
			name: "field notEmpty - empty string",
			cond: Condition{Field: "name", NotEmpty: BoolPtr(true)},
			hc: &HookContext{
				Document: map[string]interface{}{"name": ""},
			},
			expected: false,
		},
		{
			name: "path in update",
			cond: Condition{Path: "$set.password", Exists: BoolPtr(true)},
			hc: &HookContext{
				Update: M{"$set": M{"password": "newpass"}},
			},
			expected: true,
		},
		{
			name: "path not in update",
			cond: Condition{Path: "$set.password", Exists: BoolPtr(true)},
			hc: &HookContext{
				Update: M{"$set": M{"name": "test"}},
			},
			expected: false,
		},
		{
			name: "and condition - all true",
			cond: Condition{
				And: []Condition{
					{Field: "role", Eq: "admin"},
					{Field: "active", Eq: true},
				},
			},
			hc: &HookContext{
				Document: map[string]interface{}{"role": "admin", "active": true},
			},
			expected: true,
		},
		{
			name: "and condition - one false",
			cond: Condition{
				And: []Condition{
					{Field: "role", Eq: "admin"},
					{Field: "active", Eq: true},
				},
			},
			hc: &HookContext{
				Document: map[string]interface{}{"role": "user", "active": true},
			},
			expected: false,
		},
		{
			name: "or condition - one true",
			cond: Condition{
				Or: []Condition{
					{Field: "role", Eq: "admin"},
					{Field: "role", Eq: "moderator"},
				},
			},
			hc: &HookContext{
				Document: map[string]interface{}{"role": "moderator"},
			},
			expected: true,
		},
		{
			name: "or condition - all false",
			cond: Condition{
				Or: []Condition{
					{Field: "role", Eq: "admin"},
					{Field: "role", Eq: "moderator"},
				},
			},
			hc: &HookContext{
				Document: map[string]interface{}{"role": "user"},
			},
			expected: false,
		},
		{
			name: "not condition",
			cond: Condition{
				Not: &Condition{Field: "role", Eq: "banned"},
			},
			hc: &HookContext{
				Document: map[string]interface{}{"role": "user"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cond.Evaluate(context.Background(), tt.hc)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestConditionFromJSON(t *testing.T) {
	cond := &Condition{
		Field:  "role",
		Eq:     "admin",
	}

	fn := ConditionFromJSON(cond)
	if fn == nil {
		t.Fatal("ConditionFromJSON returned nil")
	}

	hc := &HookContext{
		Document: map[string]interface{}{"role": "admin"},
	}

	if !fn(context.Background(), hc) {
		t.Error("condition should return true")
	}
}

func TestJSONSchema_Hooks(t *testing.T) {
	// Register a test hook
	var called bool
	RegisterHook("testInsertHook", func(ctx context.Context, hc *HookContext) error {
		called = true
		return nil
	})
	defer UnregisterHook("testInsertHook")

	jsonStr := `{
		"name": "users",
		"fields": {
			"name": { "type": "string", "required": true }
		},
		"hooks": {
			"pre": {
				"insert": [
					{ "name": "testInsertHook" }
				]
			}
		}
	}`

	schema, err := SchemaFromJSONString(jsonStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Execute pre hooks
	hc := &HookContext{
		Operation: OpInsert,
		Document:  map[string]interface{}{"name": "test"},
		Schema:    schema,
	}

	err = schema.executePreHooks(context.Background(), OpInsert, hc)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !called {
		t.Error("hook from JSON schema was not called")
	}
}

func TestJSONSchema_HooksWithCondition(t *testing.T) {
	// Register a test hook
	var called bool
	RegisterHook("conditionalHook", func(ctx context.Context, hc *HookContext) error {
		called = true
		return nil
	})
	defer UnregisterHook("conditionalHook")

	jsonStr := `{
		"name": "users",
		"fields": {
			"name": { "type": "string" },
			"role": { "type": "string" }
		},
		"hooks": {
			"pre": {
				"insert": [
					{
						"name": "conditionalHook",
						"when": { "field": "role", "eq": "admin" }
					}
				]
			}
		}
	}`

	schema, err := SchemaFromJSONString(jsonStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with admin role - should trigger
	hc1 := &HookContext{
		Operation: OpInsert,
		Document:  map[string]interface{}{"name": "test", "role": "admin"},
		Schema:    schema,
	}
	_ = schema.executePreHooks(context.Background(), OpInsert, hc1)

	if !called {
		t.Error("hook should have been called for admin role")
	}

	// Reset and test with user role - should not trigger
	called = false
	hc2 := &HookContext{
		Operation: OpInsert,
		Document:  map[string]interface{}{"name": "test", "role": "user"},
		Schema:    schema,
	}
	_ = schema.executePreHooks(context.Background(), OpInsert, hc2)

	if called {
		t.Error("hook should not have been called for user role")
	}
}

func TestJSONSchema_HookNotFound(t *testing.T) {
	jsonStr := `{
		"name": "users",
		"fields": {
			"name": { "type": "string" }
		},
		"hooks": {
			"pre": {
				"insert": [
					{ "name": "nonExistentHook" }
				]
			}
		}
	}`

	_, err := SchemaFromJSONString(jsonStr)
	if err == nil {
		t.Error("expected error for non-existent hook")
	}
}

func TestBoolPtr(t *testing.T) {
	truePtr := BoolPtr(true)
	if truePtr == nil || *truePtr != true {
		t.Error("BoolPtr(true) should return pointer to true")
	}

	falsePtr := BoolPtr(false)
	if falsePtr == nil || *falsePtr != false {
		t.Error("BoolPtr(false) should return pointer to false")
	}
}

func TestSetHookErrorLogger(t *testing.T) {
	var loggedName string
	var loggedErr error

	// Set custom logger
	SetHookErrorLogger(func(name string, phase HookPhase, op string, err error) {
		loggedName = name
		loggedErr = err
	})

	// Restore default logger after test
	defer SetHookErrorLogger(func(name string, phase HookPhase, op string, err error) {})

	// Trigger an error in a hook with ContinueOnError
	schema := NewSchema().Field("name", TypeString)

	expectedErr := errors.New("test error")
	schema.Pre("insert", func(ctx context.Context, hc *HookContext) error {
		return expectedErr
	}, HookOptions{Name: "testHook", ContinueOnError: BoolPtr(true)})

	hc := &HookContext{Operation: OpInsert, Schema: schema}
	_ = schema.executePreHooks(context.Background(), OpInsert, hc)

	if loggedName != "testHook" {
		t.Errorf("expected hook name 'testHook', got '%s'", loggedName)
	}
	if loggedErr != expectedErr {
		t.Errorf("expected error '%v', got '%v'", expectedErr, loggedErr)
	}
}
