package mongoclient

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"sync"
)

// Operation constants for hooks
const (
	OpInsert    = "insert"
	OpInsertMany = "insertMany"
	OpUpdate    = "update"
	OpUpdateMany = "updateMany"
	OpDelete    = "delete"
	OpDeleteMany = "deleteMany"
	OpFind      = "find"
	OpFindOne   = "findOne"
	OpWildcard  = "*"
)

// HookPhase represents when a hook runs
type HookPhase string

const (
	PhasePre  HookPhase = "pre"
	PhasePost HookPhase = "post"
)

// HookContext contains all information available to hooks during execution.
// Different fields are populated depending on the operation type.
type HookContext struct {
	// Operation is the type of operation being performed.
	// One of: "insert", "insertMany", "update", "updateMany", "delete", "deleteMany", "find", "findOne"
	Operation string

	// Document is the document being inserted (for insert operations).
	// For insertMany, use Documents instead.
	Document interface{}

	// Documents is the slice of documents being inserted (for insertMany operations).
	Documents []interface{}

	// Filter is the query filter (for update, delete, find operations).
	Filter M

	// Update is the update document (for update operations).
	// Contains operators like $set, $inc, etc.
	Update M

	// Result is the operation result (only available in post hooks).
	// Type depends on operation:
	//   - insert: *InsertOneResult
	//   - insertMany: *InsertManyResult
	//   - update/updateMany: *UpdateResult
	//   - delete/deleteMany: *DeleteResult
	//   - find/findOne: []interface{} or interface{}
	Result interface{}

	// Error is set if the operation failed (only available in post hooks).
	Error error

	// Schema is the schema associated with this operation.
	Schema *Schema

	// Model is the model executing this operation (may be nil for schema-only operations).
	Model *Model

	// Extra allows hooks to pass data to subsequent hooks.
	Extra map[string]interface{}
}

// Set stores a value in Extra that can be retrieved by subsequent hooks.
func (hc *HookContext) Set(key string, value interface{}) {
	if hc.Extra == nil {
		hc.Extra = make(map[string]interface{})
	}
	hc.Extra[key] = value
}

// Get retrieves a value from Extra set by a previous hook.
func (hc *HookContext) Get(key string) (interface{}, bool) {
	if hc.Extra == nil {
		return nil, false
	}
	v, ok := hc.Extra[key]
	return v, ok
}

// GetString retrieves a string value from Extra.
func (hc *HookContext) GetString(key string) string {
	if v, ok := hc.Get(key); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// HookFunc is the signature for hook functions.
// Hooks receive context and HookContext, and return an error if something goes wrong.
// For pre hooks, returning an error will abort the operation.
// For post hooks, the error behavior depends on HookOptions.ContinueOnError.
type HookFunc func(ctx context.Context, hc *HookContext) error

// ConditionFunc determines whether a hook should run.
// Return true to run the hook, false to skip it.
type ConditionFunc func(ctx context.Context, hc *HookContext) bool

// HookOptions configures how a hook behaves.
type HookOptions struct {
	// ContinueOnError determines behavior when the hook returns an error.
	// For pre hooks: default is false (abort operation on error).
	// For post hooks: default is true (log error and continue).
	ContinueOnError *bool

	// Async determines if the hook runs asynchronously.
	// For pre hooks: default is false (must complete before operation).
	// For post hooks: default is true (runs in goroutine).
	// Async hooks that return errors will log them but not affect the operation.
	Async *bool

	// When is a condition function that determines if this hook should run.
	// If nil, the hook always runs.
	When ConditionFunc

	// Priority determines execution order. Lower values run first.
	// Default is 0. Hooks with same priority run in FIFO order.
	Priority int

	// Name is an optional identifier for debugging and logging.
	Name string
}

// hook represents a registered hook with its options
type hook struct {
	fn      HookFunc
	opts    HookOptions
	phase   HookPhase
	op      string
	order   int // registration order for FIFO
}

// hookStore manages hooks for a schema
type hookStore struct {
	mu       sync.RWMutex
	hooks    []hook
	counter  int
}

func newHookStore() *hookStore {
	return &hookStore{
		hooks: make([]hook, 0),
	}
}

// add registers a new hook
func (hs *hookStore) add(phase HookPhase, op string, fn HookFunc, opts HookOptions) {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	hs.counter++
	hs.hooks = append(hs.hooks, hook{
		fn:    fn,
		opts:  opts,
		phase: phase,
		op:    op,
		order: hs.counter,
	})
}

// get returns all hooks matching the phase and operation, sorted by priority then FIFO
func (hs *hookStore) get(phase HookPhase, op string) []hook {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	var result []hook
	for _, h := range hs.hooks {
		if h.phase == phase && (h.op == op || h.op == OpWildcard) {
			result = append(result, h)
		}
	}

	// Sort by priority (ascending), then by order (FIFO)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].opts.Priority < result[i].opts.Priority ||
				(result[j].opts.Priority == result[i].opts.Priority && result[j].order < result[i].order) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// clear removes all hooks
func (hs *hookStore) clear() {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.hooks = make([]hook, 0)
	hs.counter = 0
}

// --- Schema Hook Methods ---

// Pre registers a pre-operation hook that runs before the specified operation.
// Pre hooks can modify the document/filter/update before the operation executes.
// If a pre hook returns an error, the operation is aborted.
//
// Supported operations: "insert", "insertMany", "update", "updateMany", "delete", "deleteMany", "find", "findOne", "*"
//
// Example:
//
//	schema.Pre("insert", func(ctx context.Context, hc *HookContext) error {
//	    // Hash password before insert
//	    if doc, ok := hc.Document.(map[string]interface{}); ok {
//	        if pwd, ok := doc["password"].(string); ok {
//	            doc["password"] = hashPassword(pwd)
//	        }
//	    }
//	    return nil
//	})
func (s *Schema) Pre(operation string, fn HookFunc, opts ...HookOptions) *Schema {
	opt := HookOptions{}
	if len(opts) > 0 {
		opt = opts[0]
	}
	s.hooks.add(PhasePre, operation, fn, opt)
	return s
}

// Post registers a post-operation hook that runs after the specified operation.
// Post hooks can perform side effects like logging, notifications, or cache invalidation.
// By default, post hooks run asynchronously and continue on error.
//
// Example:
//
//	schema.Post("insert", func(ctx context.Context, hc *HookContext) error {
//	    // Send welcome email after user creation
//	    if user, ok := hc.Document.(*User); ok {
//	        return sendWelcomeEmail(user.Email)
//	    }
//	    return nil
//	}, HookOptions{Async: boolPtr(true)})
func (s *Schema) Post(operation string, fn HookFunc, opts ...HookOptions) *Schema {
	opt := HookOptions{}
	if len(opts) > 0 {
		opt = opts[0]
	}
	s.hooks.add(PhasePost, operation, fn, opt)
	return s
}

// PreWhen registers a conditional pre-operation hook.
// The hook only runs if the condition function returns true.
//
// Example:
//
//	schema.PreWhen("update", hashPasswordHook, func(ctx context.Context, hc *HookContext) bool {
//	    // Only run if password is being updated
//	    if set, ok := hc.Update["$set"].(M); ok {
//	        _, hasPassword := set["password"]
//	        return hasPassword
//	    }
//	    return false
//	})
func (s *Schema) PreWhen(operation string, fn HookFunc, cond ConditionFunc, opts ...HookOptions) *Schema {
	opt := HookOptions{}
	if len(opts) > 0 {
		opt = opts[0]
	}
	opt.When = cond
	s.hooks.add(PhasePre, operation, fn, opt)
	return s
}

// PostWhen registers a conditional post-operation hook.
func (s *Schema) PostWhen(operation string, fn HookFunc, cond ConditionFunc, opts ...HookOptions) *Schema {
	opt := HookOptions{}
	if len(opts) > 0 {
		opt = opts[0]
	}
	opt.When = cond
	s.hooks.add(PhasePost, operation, fn, opt)
	return s
}

// ClearHooks removes all registered hooks from the schema.
func (s *Schema) ClearHooks() *Schema {
	s.hooks.clear()
	return s
}

// --- Hook Execution ---

// executePreHooks runs all pre hooks for the given operation
func (s *Schema) executePreHooks(ctx context.Context, op string, hc *HookContext) error {
	hooks := s.hooks.get(PhasePre, op)

	for _, h := range hooks {
		// Check condition
		if h.opts.When != nil && !h.opts.When(ctx, hc) {
			continue
		}

		// Pre hooks are sync by default
		async := false
		if h.opts.Async != nil {
			async = *h.opts.Async
		}

		// ContinueOnError defaults to false for pre hooks
		continueOnError := false
		if h.opts.ContinueOnError != nil {
			continueOnError = *h.opts.ContinueOnError
		}

		if async {
			// Async pre hook - fire and forget
			go func(hook hook) {
				if err := hook.fn(ctx, hc); err != nil {
					// Log async pre hook error
					logHookError(hook.opts.Name, PhasePre, op, err)
				}
			}(h)
		} else {
			// Sync pre hook
			if err := h.fn(ctx, hc); err != nil {
				if !continueOnError {
					return fmt.Errorf("pre hook '%s' failed: %w", h.opts.Name, err)
				}
				// Log but continue
				logHookError(h.opts.Name, PhasePre, op, err)
			}
		}
	}

	return nil
}

// executePostHooks runs all post hooks for the given operation
func (s *Schema) executePostHooks(ctx context.Context, op string, hc *HookContext) error {
	hooks := s.hooks.get(PhasePost, op)

	var syncErrors []error

	for _, h := range hooks {
		// Check condition
		if h.opts.When != nil && !h.opts.When(ctx, hc) {
			continue
		}

		// Post hooks are async by default
		async := true
		if h.opts.Async != nil {
			async = *h.opts.Async
		}

		// ContinueOnError defaults to true for post hooks
		continueOnError := true
		if h.opts.ContinueOnError != nil {
			continueOnError = *h.opts.ContinueOnError
		}

		if async {
			// Async post hook - fire and forget
			go func(hook hook) {
				if err := hook.fn(ctx, hc); err != nil {
					logHookError(hook.opts.Name, PhasePost, op, err)
				}
			}(h)
		} else {
			// Sync post hook
			if err := h.fn(ctx, hc); err != nil {
				if !continueOnError {
					return fmt.Errorf("post hook '%s' failed: %w", h.opts.Name, err)
				}
				syncErrors = append(syncErrors, err)
				logHookError(h.opts.Name, PhasePost, op, err)
			}
		}
	}

	return nil
}

// logHookError logs hook execution errors (can be replaced with custom logger)
var logHookError = func(name string, phase HookPhase, op string, err error) {
	// Default: silent. Users can override this.
	// fmt.Printf("[HOOK ERROR] %s %s:%s - %v\n", phase, op, name, err)
}

// SetHookErrorLogger sets a custom error logger for hook errors
func SetHookErrorLogger(fn func(name string, phase HookPhase, op string, err error)) {
	logHookError = fn
}

// --- Hook Registry (for JSON Schema built-in hooks) ---

var (
	registryMu    sync.RWMutex
	hookRegistry  = make(map[string]HookFunc)
)

// RegisterHook registers a named hook function that can be referenced from JSON schemas.
// Hook names must be unique; registering a duplicate name will overwrite the previous hook.
//
// Example:
//
//	mongoclient.RegisterHook("hashPassword", func(ctx context.Context, hc *HookContext) error {
//	    // Hash password logic
//	    return nil
//	})
func RegisterHook(name string, fn HookFunc) {
	registryMu.Lock()
	defer registryMu.Unlock()
	hookRegistry[name] = fn
}

// GetRegisteredHook retrieves a registered hook by name.
func GetRegisteredHook(name string) (HookFunc, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	fn, ok := hookRegistry[name]
	return fn, ok
}

// UnregisterHook removes a hook from the registry.
func UnregisterHook(name string) {
	registryMu.Lock()
	defer registryMu.Unlock()
	delete(hookRegistry, name)
}

// ListRegisteredHooks returns all registered hook names.
func ListRegisteredHooks() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	names := make([]string, 0, len(hookRegistry))
	for name := range hookRegistry {
		names = append(names, name)
	}
	return names
}

// --- Condition Expressions (for JSON Schema) ---

// Condition represents a condition expression that can be evaluated against a HookContext
type Condition struct {
	// Field checks a field in the document (for insert operations)
	Field string `json:"field,omitempty"`

	// Path checks a path in the update document (e.g., "$set.password")
	Path string `json:"path,omitempty"`

	// Operators
	Exists   *bool       `json:"exists,omitempty"`
	Eq       interface{} `json:"eq,omitempty"`
	Ne       interface{} `json:"ne,omitempty"`
	Gt       interface{} `json:"gt,omitempty"`
	Gte      interface{} `json:"gte,omitempty"`
	Lt       interface{} `json:"lt,omitempty"`
	Lte      interface{} `json:"lte,omitempty"`
	In       []interface{} `json:"in,omitempty"`
	Nin      []interface{} `json:"nin,omitempty"`
	Matches  string      `json:"matches,omitempty"`
	NotEmpty *bool       `json:"notEmpty,omitempty"`

	// Logical operators
	And []Condition `json:"and,omitempty"`
	Or  []Condition `json:"or,omitempty"`
	Not *Condition  `json:"not,omitempty"`
}

// Evaluate evaluates the condition against the given HookContext
func (c *Condition) Evaluate(ctx context.Context, hc *HookContext) bool {
	// Handle logical operators first
	if len(c.And) > 0 {
		for _, cond := range c.And {
			if !cond.Evaluate(ctx, hc) {
				return false
			}
		}
		return true
	}

	if len(c.Or) > 0 {
		for _, cond := range c.Or {
			if cond.Evaluate(ctx, hc) {
				return true
			}
		}
		return false
	}

	if c.Not != nil {
		return !c.Not.Evaluate(ctx, hc)
	}

	// Get the value to check
	var value interface{}
	var exists bool

	if c.Field != "" {
		value, exists = c.getFieldValue(hc)
	} else if c.Path != "" {
		value, exists = c.getPathValue(hc)
	} else {
		return true // No field or path specified, condition passes
	}

	// Check exists operator
	if c.Exists != nil {
		if *c.Exists {
			return exists
		}
		return !exists
	}

	// If value doesn't exist, other operators fail
	if !exists {
		return false
	}

	// Check notEmpty
	if c.NotEmpty != nil && *c.NotEmpty {
		return !isEmptyValue(value)
	}

	// Check equality operators
	if c.Eq != nil {
		return reflect.DeepEqual(value, c.Eq)
	}

	if c.Ne != nil {
		return !reflect.DeepEqual(value, c.Ne)
	}

	// Check comparison operators (for numbers)
	if c.Gt != nil {
		return compareNumbers(value, c.Gt) > 0
	}

	if c.Gte != nil {
		return compareNumbers(value, c.Gte) >= 0
	}

	if c.Lt != nil {
		return compareNumbers(value, c.Lt) < 0
	}

	if c.Lte != nil {
		return compareNumbers(value, c.Lte) <= 0
	}

	// Check in/nin
	if len(c.In) > 0 {
		for _, v := range c.In {
			if reflect.DeepEqual(value, v) {
				return true
			}
		}
		return false
	}

	if len(c.Nin) > 0 {
		for _, v := range c.Nin {
			if reflect.DeepEqual(value, v) {
				return false
			}
		}
		return true
	}

	// Check regex match
	if c.Matches != "" {
		if str, ok := value.(string); ok {
			matched, _ := regexp.MatchString(c.Matches, str)
			return matched
		}
		return false
	}

	return true
}

// getFieldValue gets a field value from the document
func (c *Condition) getFieldValue(hc *HookContext) (interface{}, bool) {
	// Try document first
	if hc.Document != nil {
		if m, ok := toMapInterface(hc.Document); ok {
			v, exists := m[c.Field]
			return v, exists
		}
	}

	// Try filter
	if hc.Filter != nil {
		v, exists := hc.Filter[c.Field]
		return v, exists
	}

	return nil, false
}

// getPathValue gets a value from a path like "$set.password"
func (c *Condition) getPathValue(hc *HookContext) (interface{}, bool) {
	if hc.Update == nil {
		return nil, false
	}

	// Parse path like "$set.password" or "$inc.count"
	parts := splitPath(c.Path)
	if len(parts) == 0 {
		return nil, false
	}

	var current interface{} = map[string]interface{}(hc.Update)
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[part]
			if !ok {
				return nil, false
			}
			current = val
		case M:
			val, ok := v[part]
			if !ok {
				return nil, false
			}
			current = val
		default:
			return nil, false
		}
	}

	return current, true
}

// splitPath splits a path like "$set.password" into ["$set", "password"]
func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	var parts []string
	current := ""
	for _, ch := range path {
		if ch == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// toMapInterface converts a struct or map to map[string]interface{}
func toMapInterface(v interface{}) (map[string]interface{}, bool) {
	if v == nil {
		return nil, false
	}

	// Already a map
	if m, ok := v.(map[string]interface{}); ok {
		return m, true
	}

	if m, ok := v.(M); ok {
		return map[string]interface{}(m), true
	}

	// Try reflection for structs
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, false
	}

	result := make(map[string]interface{})
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" { // unexported
			continue
		}

		name := field.Name
		// Check bson tag
		if tag := field.Tag.Get("bson"); tag != "" && tag != "-" {
			if idx := findCommaIndex(tag); idx > 0 {
				name = tag[:idx]
			} else if tag != ",inline" {
				name = tag
			}
		}

		result[name] = val.Field(i).Interface()
	}

	return result, true
}

// findCommaIndex finds the index of comma in a string
func findCommaIndex(s string) int {
	for i, ch := range s {
		if ch == ',' {
			return i
		}
	}
	return -1
}

// isEmptyValue checks if a value is empty/zero
func isEmptyValue(v interface{}) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.String:
		return val.String() == ""
	case reflect.Slice, reflect.Map, reflect.Array:
		return val.Len() == 0
	case reflect.Bool:
		return !val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Ptr, reflect.Interface:
		return val.IsNil()
	}

	return false
}

// compareNumbers compares two numeric values
func compareNumbers(a, b interface{}) int {
	af := numToFloat64(a)
	bf := numToFloat64(b)

	if af < bf {
		return -1
	} else if af > bf {
		return 1
	}
	return 0
}

// numToFloat64 converts a numeric value to float64 (for hooks condition evaluation)
func numToFloat64(v interface{}) float64 {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int8:
		return float64(n)
	case int16:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case uint:
		return float64(n)
	case uint8:
		return float64(n)
	case uint16:
		return float64(n)
	case uint32:
		return float64(n)
	case uint64:
		return float64(n)
	case float32:
		return float64(n)
	case float64:
		return n
	default:
		return 0
	}
}

// ConditionFromJSON creates a ConditionFunc from a Condition struct
func ConditionFromJSON(cond *Condition) ConditionFunc {
	if cond == nil {
		return nil
	}
	return func(ctx context.Context, hc *HookContext) bool {
		return cond.Evaluate(ctx, hc)
	}
}

// --- Helper Functions ---

// BoolPtr returns a pointer to a bool value (helper for HookOptions)
func BoolPtr(b bool) *bool {
	return &b
}
