package envutil_test

import (
	"os"
	"testing"
	"time"

	envutil "github.com/isimtekin/go-packages/env-util"
)

func TestGetEnv(t *testing.T) {
	// Test with set value
	os.Setenv("TEST_STRING", "hello")
	defer os.Unsetenv("TEST_STRING")
	
	val := envutil.GetEnv("TEST_STRING", "default")
	if val != "hello" {
		t.Errorf("Expected 'hello', got '%s'", val)
	}
	
	// Test with unset value
	val = envutil.GetEnv("UNSET_STRING", "default")
	if val != "default" {
		t.Errorf("Expected 'default', got '%s'", val)
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     bool
		wantDef  bool
	}{
		{"true", "true", true, false},
		{"false", "false", false, true},
		{"1", "1", true, false},
		{"0", "0", false, true},
		{"yes", "yes", true, false},
		{"no", "no", false, true},
		{"invalid", "invalid", false, false}, // should return default
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TEST_BOOL", tt.envValue)
			defer os.Unsetenv("TEST_BOOL")
			
			result := envutil.GetEnvBool("TEST_BOOL", tt.wantDef)
			
			// For invalid values, we expect the default
			expected := tt.want
			if tt.name == "invalid" {
				expected = tt.wantDef
			}
			
			if result != expected {
				t.Errorf("GetEnvBool() = %v, want %v", result, expected)
			}
		})
	}
	
	// Test unset value
	result := envutil.GetEnvBool("UNSET_BOOL", true)
	if result != true {
		t.Errorf("Expected true (default), got %v", result)
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     int
		default_ int
	}{
		{"valid_positive", "42", 42, 0},
		{"valid_negative", "-10", -10, 0},
		{"valid_zero", "0", 0, 100},
		{"invalid", "not_a_number", 99, 99}, // should return default
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TEST_INT", tt.envValue)
			defer os.Unsetenv("TEST_INT")
			
			result := envutil.GetEnvInt("TEST_INT", tt.default_)
			
			if result != tt.want {
				t.Errorf("GetEnvInt() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestGetEnvDuration(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		want     time.Duration
		default_ time.Duration
	}{
		{"with_suffix_seconds", "TEST_DUR", "10s", 10 * time.Second, 0},
		{"with_suffix_minutes", "TEST_DUR", "5m", 5 * time.Minute, 0},
		{"with_suffix_hours", "TEST_DUR", "2h", 2 * time.Hour, 0},
		{"integer_as_seconds", "TEST_DUR", "30", 30 * time.Second, 0},
		{"integer_with_ms_key", "TEST_MS", "500", 500 * time.Millisecond, 0},
		{"integer_with_min_key", "TEST_MIN", "15", 15 * time.Minute, 0},
		{"invalid", "TEST_DUR", "invalid", time.Minute, time.Minute},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envKey, tt.envValue)
			defer os.Unsetenv(tt.envKey)
			
			result := envutil.GetEnvDuration(tt.envKey, tt.default_)
			
			if result != tt.want {
				t.Errorf("GetEnvDuration() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestGetEnvStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     []string
		default_ []string
	}{
		{"comma_separated", "a,b,c", []string{"a", "b", "c"}, nil},
		{"with_spaces", "a, b , c", []string{"a", "b", "c"}, nil},
		{"single_value", "single", []string{"single"}, nil},
		{"empty_items", "a,,c", []string{"a", "c"}, nil},
		{"empty_string", "", []string{"default"}, []string{"default"}},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("TEST_SLICE", tt.envValue)
				defer os.Unsetenv("TEST_SLICE")
			}
			
			result := envutil.GetEnvStringSlice("TEST_SLICE", tt.default_)
			
			if len(result) != len(tt.want) {
				t.Errorf("GetEnvStringSlice() length = %d, want %d", len(result), len(tt.want))
				return
			}
			
			for i := range result {
				if result[i] != tt.want[i] {
					t.Errorf("GetEnvStringSlice()[%d] = %s, want %s", i, result[i], tt.want[i])
				}
			}
		})
	}
}

func TestGetEnvIntSlice(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     []int
		default_ []int
	}{
		{"comma_separated", "1,2,3", []int{1, 2, 3}, nil},
		{"with_spaces", "10, 20 , 30", []int{10, 20, 30}, nil},
		{"negative_numbers", "-1,0,1", []int{-1, 0, 1}, nil},
		{"with_invalid", "1,invalid,3", []int{1, 3}, nil}, // invalid items skipped
		{"empty_string", "", []int{99}, []int{99}},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("TEST_INT_SLICE", tt.envValue)
				defer os.Unsetenv("TEST_INT_SLICE")
			}
			
			result := envutil.GetEnvIntSlice("TEST_INT_SLICE", tt.default_)
			
			if len(result) != len(tt.want) {
				t.Errorf("GetEnvIntSlice() length = %d, want %d", len(result), len(tt.want))
				return
			}
			
			for i := range result {
				if result[i] != tt.want[i] {
					t.Errorf("GetEnvIntSlice()[%d] = %d, want %d", i, result[i], tt.want[i])
				}
			}
		})
	}
}

func TestMustGetEnv(t *testing.T) {
	os.Setenv("TEST_REQUIRED", "value")
	defer os.Unsetenv("TEST_REQUIRED")
	
	val := envutil.MustGetEnv("TEST_REQUIRED")
	if val != "value" {
		t.Errorf("Expected 'value', got '%s'", val)
	}
	
	// Test panic for unset variable
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for unset required variable")
		}
	}()
	
	_ = envutil.MustGetEnv("UNSET_REQUIRED")
}

func TestIsEnvSet(t *testing.T) {
	os.Setenv("SET_VAR", "value")
	defer os.Unsetenv("SET_VAR")
	
	if !envutil.IsEnvSet("SET_VAR") {
		t.Error("Expected SET_VAR to be set")
	}
	
	if envutil.IsEnvSet("UNSET_VAR") {
		t.Error("Expected UNSET_VAR to not be set")
	}
	
	// Test with empty value
	os.Setenv("EMPTY_VAR", "")
	defer os.Unsetenv("EMPTY_VAR")
	
	if !envutil.IsEnvSet("EMPTY_VAR") {
		t.Error("Expected EMPTY_VAR to be set even with empty value")
	}
}

func TestGetAllEnvWithPrefix(t *testing.T) {
	// Set up test environment variables
	os.Setenv("APP_NAME", "test")
	os.Setenv("APP_VERSION", "1.0")
	os.Setenv("OTHER_VAR", "value")
	defer os.Unsetenv("APP_NAME")
	defer os.Unsetenv("APP_VERSION")
	defer os.Unsetenv("OTHER_VAR")
	
	result := envutil.GetAllEnvWithPrefix("APP_")
	
	if len(result) != 2 {
		t.Errorf("Expected 2 variables with APP_ prefix, got %d", len(result))
	}
	
	if result["APP_NAME"] != "test" {
		t.Error("Expected APP_NAME to be 'test'")
	}
	
	if result["APP_VERSION"] != "1.0" {
		t.Error("Expected APP_VERSION to be '1.0'")
	}
	
	if _, exists := result["OTHER_VAR"]; exists {
		t.Error("OTHER_VAR should not be in results")
	}
}

func TestClient(t *testing.T) {
	// Test with prefix
	client := envutil.NewWithOptions(
		envutil.WithPrefix("MYAPP_"),
		envutil.WithSilent(true),
	)
	
	os.Setenv("MYAPP_HOST", "example.com")
	defer os.Unsetenv("MYAPP_HOST")
	
	host := client.GetString("HOST", "localhost")
	if host != "example.com" {
		t.Errorf("Expected 'example.com', got '%s'", host)
	}
	
	// Test without prefix (key not found)
	os.Setenv("HOST", "other.com")
	defer os.Unsetenv("HOST")
	
	// Should still get example.com because client is looking for MYAPP_HOST
	host = client.GetString("HOST", "localhost")
	if host != "example.com" {
		t.Errorf("Expected 'example.com' with prefix, got '%s'", host)
	}
}

func TestGetEnvPort(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     int
		default_ int
	}{
		{"valid_port", "8080", 8080, 3000},
		{"min_port", "1", 1, 3000},
		{"max_port", "65535", 65535, 3000},
		{"invalid_zero", "0", 3000, 3000},
		{"invalid_negative", "-1", 3000, 3000},
		{"invalid_too_large", "70000", 3000, 3000},
		{"not_a_number", "abc", 3000, 3000},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TEST_PORT", tt.envValue)
			defer os.Unsetenv("TEST_PORT")
			
			result := envutil.GetEnvPort("TEST_PORT", tt.default_)
			if result != tt.want {
				t.Errorf("GetEnvPort() = %d, want %d", result, tt.want)
			}
		})
	}
}