package redisclient

import (
	"testing"
)

func TestConfig_ApplyWorkspacePrefix(t *testing.T) {
	tests := []struct {
		name      string
		workspace string
		key       string
		expected  string
	}{
		{
			name:      "with workspace",
			workspace: "production",
			key:       "user:123",
			expected:  "production:user:123",
		},
		{
			name:      "without workspace",
			workspace: "",
			key:       "user:123",
			expected:  "user:123",
		},
		{
			name:      "empty key",
			workspace: "production",
			key:       "",
			expected:  "",
		},
		{
			name:      "both empty",
			workspace: "",
			key:       "",
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Workspace: tt.workspace,
			}
			result := config.ApplyWorkspacePrefix(tt.key)
			if result != tt.expected {
				t.Errorf("ApplyWorkspacePrefix() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConfig_ApplyWorkspacePrefixToKeys(t *testing.T) {
	tests := []struct {
		name      string
		workspace string
		keys      []string
		expected  []string
	}{
		{
			name:      "with workspace",
			workspace: "production",
			keys:      []string{"user:123", "session:456"},
			expected:  []string{"production:user:123", "production:session:456"},
		},
		{
			name:      "without workspace",
			workspace: "",
			keys:      []string{"user:123", "session:456"},
			expected:  []string{"user:123", "session:456"},
		},
		{
			name:      "empty keys",
			workspace: "production",
			keys:      []string{},
			expected:  []string{},
		},
		{
			name:      "with empty key in slice",
			workspace: "production",
			keys:      []string{"user:123", "", "session:456"},
			expected:  []string{"production:user:123", "", "production:session:456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Workspace: tt.workspace,
			}
			result := config.ApplyWorkspacePrefixToKeys(tt.keys)
			if len(result) != len(tt.expected) {
				t.Errorf("ApplyWorkspacePrefixToKeys() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("ApplyWorkspacePrefixToKeys()[%d] = %v, want %v", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestWithWorkspace(t *testing.T) {
	config := DefaultConfig()

	// Test default (no workspace)
	if config.Workspace != "" {
		t.Errorf("Default workspace should be empty, got %v", config.Workspace)
	}

	// Test WithWorkspace option
	workspace := "test-env"
	option := WithWorkspace(workspace)
	option(config)

	if config.Workspace != workspace {
		t.Errorf("WithWorkspace() = %v, want %v", config.Workspace, workspace)
	}
}
