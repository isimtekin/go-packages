package mongoclient

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// Built-in hooks that can be used from JSON schemas or directly in code.
// Register your own hooks using RegisterHook().

func init() {
	// Register all built-in hooks
	RegisterHook("setTimestamps", SetTimestampsHook)
	RegisterHook("generateSlug", GenerateSlugHook)
	RegisterHook("trimStrings", TrimStringsHook)
	RegisterHook("toLowerCase", ToLowerCaseHook)
	RegisterHook("logOperation", LogOperationHook)
}

// SetTimestampsHook sets createdAt and updatedAt timestamps on documents.
// Use this if you need custom timestamp handling beyond the built-in schema.Timestamps feature.
//
// Usage:
//
//	schema.Pre("insert", SetTimestampsHook)
var SetTimestampsHook HookFunc = func(ctx context.Context, hc *HookContext) error {
	doc, ok := hc.Document.(map[string]interface{})
	if !ok {
		return nil
	}

	now := time.Now()

	switch hc.Operation {
	case OpInsert, OpInsertMany:
		if _, exists := doc["createdAt"]; !exists {
			doc["createdAt"] = now
		}
		if _, exists := doc["updatedAt"]; !exists {
			doc["updatedAt"] = now
		}
	case OpUpdate, OpUpdateMany:
		// For updates, we'd need to modify the update document
		if hc.Update != nil {
			if set, ok := hc.Update["$set"].(M); ok {
				if _, exists := set["updatedAt"]; !exists {
					set["updatedAt"] = now
				}
			}
		}
	}

	return nil
}

// GenerateSlugHook generates a URL-friendly slug from a source field.
// By default, it uses the "name" or "title" field as source and sets "slug" field.
//
// Configure source and target fields via HookContext.Extra:
//
//	hc.Set("slugSource", "title")
//	hc.Set("slugTarget", "urlSlug")
//
// Usage in JSON schema:
//
//	{
//	  "hooks": {
//	    "pre": {
//	      "insert": [{ "name": "generateSlug" }]
//	    }
//	  }
//	}
var GenerateSlugHook HookFunc = func(ctx context.Context, hc *HookContext) error {
	doc, ok := hc.Document.(map[string]interface{})
	if !ok {
		return nil
	}

	// Get source field (default: "name" or "title")
	sourceField := "name"
	if src := hc.GetString("slugSource"); src != "" {
		sourceField = src
	} else if _, hasTitle := doc["title"]; hasTitle {
		sourceField = "title"
	}

	// Get target field (default: "slug")
	targetField := "slug"
	if tgt := hc.GetString("slugTarget"); tgt != "" {
		targetField = tgt
	}

	// Skip if slug already exists
	if _, exists := doc[targetField]; exists {
		return nil
	}

	// Get source value
	sourceValue, ok := doc[sourceField].(string)
	if !ok || sourceValue == "" {
		return nil
	}

	// Generate slug
	doc[targetField] = generateSlug(sourceValue)

	return nil
}

// generateSlug creates a URL-friendly slug from a string
func generateSlug(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces and underscores with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	// Remove non-alphanumeric characters (except hyphens)
	var result strings.Builder
	prevHyphen := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
			prevHyphen = false
		} else if r == '-' && !prevHyphen && result.Len() > 0 {
			result.WriteRune('-')
			prevHyphen = true
		}
	}

	// Trim trailing hyphen
	slug := strings.TrimSuffix(result.String(), "-")

	return slug
}

// TrimStringsHook trims whitespace from all string fields in the document.
//
// Usage:
//
//	schema.Pre("insert", TrimStringsHook)
var TrimStringsHook HookFunc = func(ctx context.Context, hc *HookContext) error {
	doc, ok := hc.Document.(map[string]interface{})
	if !ok {
		return nil
	}

	for key, value := range doc {
		if str, ok := value.(string); ok {
			doc[key] = strings.TrimSpace(str)
		}
	}

	return nil
}

// ToLowerCaseHook converts specific fields to lowercase.
// Configure fields to lowercase via HookContext.Extra:
//
//	hc.Set("lowercaseFields", []string{"email", "username"})
//
// Default: converts "email" field if it exists.
var ToLowerCaseHook HookFunc = func(ctx context.Context, hc *HookContext) error {
	doc, ok := hc.Document.(map[string]interface{})
	if !ok {
		return nil
	}

	// Get fields to lowercase
	fields := []string{"email"}
	if f, ok := hc.Get("lowercaseFields"); ok {
		if fieldList, ok := f.([]string); ok {
			fields = fieldList
		}
	}

	for _, field := range fields {
		if value, ok := doc[field].(string); ok {
			doc[field] = strings.ToLower(value)
		}
	}

	return nil
}

// LogOperationHook logs the operation being performed.
// This is useful for debugging and audit trails.
//
// Usage:
//
//	schema.Pre("*", LogOperationHook)
//	schema.Post("*", LogOperationHook)
var LogOperationHook HookFunc = func(ctx context.Context, hc *HookContext) error {
	// This hook uses the custom logger set by SetHookErrorLogger
	// For actual logging, users should set up their own logger
	fmt.Printf("[HOOK LOG] Operation: %s, Collection: %s\n",
		hc.Operation,
		func() string {
			if hc.Schema != nil {
				return hc.Schema.CollectionName
			}
			return "unknown"
		}())
	return nil
}

// HashPasswordHook hashes a password field using bcrypt.
// The field to hash is determined by HookContext.Extra["passwordField"] or defaults to "password".
//
// Usage:
//
//	schema.Pre("insert", HashPasswordHook)
//	schema.PreWhen("update", HashPasswordHook, func(ctx context.Context, hc *HookContext) bool {
//	    // Only hash if password is being updated
//	    if set, ok := hc.Update["$set"].(M); ok {
//	        _, has := set["password"]
//	        return has
//	    }
//	    return false
//	})
//
// Note: This hook requires golang.org/x/crypto/bcrypt
var HashPasswordHook HookFunc = func(ctx context.Context, hc *HookContext) error {
	// Get password field name
	fieldName := "password"
	if f := hc.GetString("passwordField"); f != "" {
		fieldName = f
	}

	switch hc.Operation {
	case OpInsert, OpInsertMany:
		doc, ok := hc.Document.(map[string]interface{})
		if !ok {
			return nil
		}

		password, ok := doc[fieldName].(string)
		if !ok || password == "" {
			return nil
		}

		// Don't re-hash if already hashed (bcrypt hashes start with $2)
		if strings.HasPrefix(password, "$2") {
			return nil
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		doc[fieldName] = string(hashed)

	case OpUpdate, OpUpdateMany:
		if hc.Update == nil {
			return nil
		}

		set, ok := hc.Update["$set"].(M)
		if !ok {
			return nil
		}

		password, ok := set[fieldName].(string)
		if !ok || password == "" {
			return nil
		}

		// Don't re-hash if already hashed
		if strings.HasPrefix(password, "$2") {
			return nil
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		set[fieldName] = string(hashed)
	}

	return nil
}

// ValidateUniqueHook placeholder - actual implementation would require database access
// This is an example of how a more complex hook could be structured
// In practice, this would need to be implemented with access to the database
var ValidateUniqueHook HookFunc = func(ctx context.Context, hc *HookContext) error {
	// This is a placeholder - actual uniqueness validation requires database queries
	// Users should implement their own uniqueness validation hooks with database access
	return nil
}
