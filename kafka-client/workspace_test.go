package kafkaclient

import (
	"testing"
)

func TestApplyWorkspacePrefix(t *testing.T) {
	tests := []struct {
		name      string
		workspace string
		topic     string
		expected  string
	}{
		{
			name:      "workspace with topic",
			workspace: "production",
			topic:     "orders",
			expected:  "production.orders",
		},
		{
			name:      "empty workspace",
			workspace: "",
			topic:     "orders",
			expected:  "orders",
		},
		{
			name:      "empty topic",
			workspace: "production",
			topic:     "",
			expected:  "",
		},
		{
			name:      "both empty",
			workspace: "",
			topic:     "",
			expected:  "",
		},
		{
			name:      "workspace with hyphen",
			workspace: "dev-team",
			topic:     "events",
			expected:  "dev-team.events",
		},
		{
			name:      "topic with hyphen",
			workspace: "staging",
			topic:     "user-events",
			expected:  "staging.user-events",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Workspace: tt.workspace,
			}

			result := config.ApplyWorkspacePrefix(tt.topic)
			if result != tt.expected {
				t.Errorf("ApplyWorkspacePrefix() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestApplyWorkspacePrefixToTopics(t *testing.T) {
	tests := []struct {
		name      string
		workspace string
		topics    []string
		expected  []string
	}{
		{
			name:      "multiple topics with workspace",
			workspace: "production",
			topics:    []string{"orders", "users", "events"},
			expected:  []string{"production.orders", "production.users", "production.events"},
		},
		{
			name:      "no workspace",
			workspace: "",
			topics:    []string{"orders", "users"},
			expected:  []string{"orders", "users"},
		},
		{
			name:      "empty topics array",
			workspace: "production",
			topics:    []string{},
			expected:  []string{},
		},
		{
			name:      "single topic",
			workspace: "dev",
			topics:    []string{"test-topic"},
			expected:  []string{"dev.test-topic"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Workspace: tt.workspace,
			}

			result := config.ApplyWorkspacePrefixToTopics(tt.topics)
			if len(result) != len(tt.expected) {
				t.Errorf("ApplyWorkspacePrefixToTopics() returned %d topics, want %d", len(result), len(tt.expected))
				return
			}

			for i, topic := range result {
				if topic != tt.expected[i] {
					t.Errorf("ApplyWorkspacePrefixToTopics()[%d] = %v, want %v", i, topic, tt.expected[i])
				}
			}
		})
	}
}

func TestWithWorkspaceOption(t *testing.T) {
	tests := []struct {
		name      string
		workspace string
	}{
		{
			name:      "set production workspace",
			workspace: "production",
		},
		{
			name:      "set dev workspace",
			workspace: "dev",
		},
		{
			name:      "set empty workspace",
			workspace: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			WithWorkspace(tt.workspace)(config)

			if config.Workspace != tt.workspace {
				t.Errorf("WithWorkspace() set workspace to %v, want %v", config.Workspace, tt.workspace)
			}
		})
	}
}

func TestWorkspaceIntegrationWithProducer(t *testing.T) {
	// Test that workspace prefix is applied in producer messages
	config := DefaultConfig()
	config.Workspace = "test-workspace"

	// Test single topic
	topic := "orders"
	expectedTopic := "test-workspace.orders"

	prefixedTopic := config.ApplyWorkspacePrefix(topic)
	if prefixedTopic != expectedTopic {
		t.Errorf("Producer should use topic %v, got %v", expectedTopic, prefixedTopic)
	}
}

func TestWorkspaceIntegrationWithConsumer(t *testing.T) {
	// Test that workspace prefix is applied in consumer topics
	config := DefaultConfig()
	config.Workspace = "test-workspace"
	config.Consumer.Topics = []string{"orders", "users", "events"}

	expectedTopics := []string{"test-workspace.orders", "test-workspace.users", "test-workspace.events"}

	prefixedTopics := config.ApplyWorkspacePrefixToTopics(config.Consumer.Topics)

	if len(prefixedTopics) != len(expectedTopics) {
		t.Errorf("Consumer should have %d topics, got %d", len(expectedTopics), len(prefixedTopics))
		return
	}

	for i, topic := range prefixedTopics {
		if topic != expectedTopics[i] {
			t.Errorf("Consumer topic[%d] = %v, want %v", i, topic, expectedTopics[i])
		}
	}
}

func TestWorkspaceWithDifferentEnvironments(t *testing.T) {
	// Simulating different environments
	environments := []struct {
		name      string
		workspace string
		topic     string
		expected  string
	}{
		{
			name:      "development environment",
			workspace: "dev",
			topic:     "orders",
			expected:  "dev.orders",
		},
		{
			name:      "staging environment",
			workspace: "staging",
			topic:     "orders",
			expected:  "staging.orders",
		},
		{
			name:      "production environment",
			workspace: "production",
			topic:     "orders",
			expected:  "production.orders",
		},
		{
			name:      "tenant-based workspace",
			workspace: "tenant-123",
			topic:     "events",
			expected:  "tenant-123.events",
		},
	}

	for _, env := range environments {
		t.Run(env.name, func(t *testing.T) {
			config := &Config{
				Workspace: env.workspace,
			}

			result := config.ApplyWorkspacePrefix(env.topic)
			if result != env.expected {
				t.Errorf("Workspace prefix for %s: got %v, want %v", env.name, result, env.expected)
			}
		})
	}
}
