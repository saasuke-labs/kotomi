package moderation

import (
	"testing"
)

func TestDefaultModerationConfig(t *testing.T) {
	config := DefaultModerationConfig()

	if config.Enabled {
		t.Error("Default config should be disabled")
	}
	if config.AutoRejectThreshold != 0.85 {
		t.Errorf("Expected AutoRejectThreshold 0.85, got %f", config.AutoRejectThreshold)
	}
	if config.AutoApproveThreshold != 0.30 {
		t.Errorf("Expected AutoApproveThreshold 0.30, got %f", config.AutoApproveThreshold)
	}
	if !config.CheckSpam {
		t.Error("CheckSpam should be enabled by default")
	}
	if !config.CheckOffensive {
		t.Error("CheckOffensive should be enabled by default")
	}
	if !config.CheckAggressive {
		t.Error("CheckAggressive should be enabled by default")
	}
	if config.CheckOffTopic {
		t.Error("CheckOffTopic should be disabled by default")
	}
}

func TestDetermineStatus(t *testing.T) {
	config := DefaultModerationConfig()

	tests := []struct {
		name       string
		confidence float64
		expected   string
	}{
		{
			name:       "High confidence - should reject",
			confidence: 0.9,
			expected:   "rejected",
		},
		{
			name:       "Medium confidence - should flag",
			confidence: 0.5,
			expected:   "pending",
		},
		{
			name:       "Low confidence - should approve",
			confidence: 0.2,
			expected:   "approved",
		},
		{
			name:       "Threshold boundary - reject",
			confidence: 0.85,
			expected:   "rejected",
		},
		{
			name:       "Threshold boundary - approve",
			confidence: 0.30,
			expected:   "approved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ModerationResult{
				Confidence: tt.confidence,
			}
			status := DetermineStatus(result, config)
			if status != tt.expected {
				t.Errorf("Expected status %s, got %s for confidence %f", tt.expected, status, tt.confidence)
			}
		})
	}
}

func TestMockModerator_CleanComment(t *testing.T) {
	moderator := NewMockModerator()
	config := DefaultModerationConfig()
	config.Enabled = true

	result, err := moderator.AnalyzeComment("This is a great article, thanks for sharing!", config)
	if err != nil {
		t.Fatalf("AnalyzeComment failed: %v", err)
	}

	if result.Decision != "approve" {
		t.Errorf("Expected decision 'approve', got '%s'", result.Decision)
	}
	if result.Confidence > 0.3 {
		t.Errorf("Expected low confidence for clean comment, got %f", result.Confidence)
	}
	if len(result.Categories) > 0 {
		t.Errorf("Expected no categories for clean comment, got %v", result.Categories)
	}
}

func TestMockModerator_SpamDetection(t *testing.T) {
	moderator := NewMockModerator()
	config := DefaultModerationConfig()
	config.Enabled = true

	tests := []struct {
		name    string
		comment string
	}{
		{
			name:    "Spam keywords",
			comment: "Buy now! Limited offer! Click here!",
		},
		{
			name:    "Excessive links",
			comment: "Check out http://spam.com and http://more-spam.com and http://even-more.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := moderator.AnalyzeComment(tt.comment, config)
			if err != nil {
				t.Fatalf("AnalyzeComment failed: %v", err)
			}

			if !contains(result.Categories, "spam") {
				t.Errorf("Expected 'spam' category, got %v", result.Categories)
			}
			if result.Confidence == 0 {
				t.Errorf("Expected non-zero confidence for spam, got %f", result.Confidence)
			}
		})
	}
}

func TestMockModerator_OffensiveDetection(t *testing.T) {
	moderator := NewMockModerator()
	config := DefaultModerationConfig()
	config.Enabled = true

	result, err := moderator.AnalyzeComment("This is fucking terrible", config)
	if err != nil {
		t.Fatalf("AnalyzeComment failed: %v", err)
	}

	if !contains(result.Categories, "offensive") {
		t.Errorf("Expected 'offensive' category, got %v", result.Categories)
	}
	if result.Confidence == 0 {
		t.Errorf("Expected non-zero confidence for offensive content, got %f", result.Confidence)
	}
}

func TestMockModerator_AggressiveDetection(t *testing.T) {
	moderator := NewMockModerator()
	config := DefaultModerationConfig()
	config.Enabled = true

	tests := []struct {
		name    string
		comment string
	}{
		{
			name:    "Aggressive language",
			comment: "You're stupid and you don't know what you're talking about",
		},
		{
			name:    "Excessive caps",
			comment: "THIS IS COMPLETELY WRONG AND YOU SHOULD KNOW BETTER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := moderator.AnalyzeComment(tt.comment, config)
			if err != nil {
				t.Fatalf("AnalyzeComment failed: %v", err)
			}

			if !contains(result.Categories, "aggressive") {
				t.Errorf("Expected 'aggressive' category, got %v", result.Categories)
			}
			if result.Confidence == 0 {
				t.Errorf("Expected non-zero confidence for aggressive content, got %f", result.Confidence)
			}
		})
	}
}

func TestMockModerator_DisabledChecks(t *testing.T) {
	moderator := NewMockModerator()
	config := DefaultModerationConfig()
	config.Enabled = true
	config.CheckSpam = false
	config.CheckOffensive = false
	config.CheckAggressive = false

	// This comment has spam, offensive, and aggressive content
	result, err := moderator.AnalyzeComment("Buy now you stupid fuck!", config)
	if err != nil {
		t.Fatalf("AnalyzeComment failed: %v", err)
	}

	// Should pass all checks since they're disabled
	if result.Decision != "approve" {
		t.Errorf("Expected decision 'approve' with all checks disabled, got '%s'", result.Decision)
	}
	if len(result.Categories) > 0 {
		t.Errorf("Expected no categories with all checks disabled, got %v", result.Categories)
	}
}

func TestMockModerator_HighConfidenceReject(t *testing.T) {
	moderator := NewMockModerator()
	config := DefaultModerationConfig()
	config.Enabled = true

	// Comment with multiple issues should have high confidence
	result, err := moderator.AnalyzeComment("Buy now you fucking idiot! Click here http://spam.com", config)
	if err != nil {
		t.Fatalf("AnalyzeComment failed: %v", err)
	}

	if result.Decision != "reject" && result.Decision != "flag" {
		t.Errorf("Expected decision 'reject' or 'flag', got '%s'", result.Decision)
	}
	if result.Confidence < 0.5 {
		t.Errorf("Expected high confidence for multiple issues, got %f", result.Confidence)
	}
	if len(result.Categories) < 2 {
		t.Errorf("Expected multiple categories for multiple issues, got %v", result.Categories)
	}
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
