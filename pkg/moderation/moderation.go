package moderation

import (
	"time"
)

// ModerationResult represents the result of AI moderation analysis
type ModerationResult struct {
	Decision   string  `json:"decision"`    // "approve", "flag", "reject"
	Confidence float64 `json:"confidence"`  // 0.0 to 1.0
	Reason     string  `json:"reason"`      // Explanation for the decision
	Categories []string `json:"categories"` // List of detected issues (spam, offensive, etc.)
	AnalyzedAt time.Time `json:"analyzed_at"`
}

// ModerationConfig represents moderation settings for a site
type ModerationConfig struct {
	Enabled            bool    `json:"enabled"`
	AutoRejectThreshold float64 `json:"auto_reject_threshold"` // e.g., 0.9
	AutoApproveThreshold float64 `json:"auto_approve_threshold"` // e.g., 0.3
	CheckSpam          bool    `json:"check_spam"`
	CheckOffensive     bool    `json:"check_offensive"`
	CheckAggressive    bool    `json:"check_aggressive"`
	CheckOffTopic      bool    `json:"check_off_topic"`
}

// Moderator is the interface for content moderation
type Moderator interface {
	AnalyzeComment(text string, config ModerationConfig) (*ModerationResult, error)
}

// DefaultModerationConfig returns the default moderation configuration
func DefaultModerationConfig() ModerationConfig {
	return ModerationConfig{
		Enabled:              false, // Disabled by default
		AutoRejectThreshold:  0.85,  // High confidence for auto-reject
		AutoApproveThreshold: 0.30,  // Low confidence threshold for auto-approve
		CheckSpam:            true,
		CheckOffensive:       true,
		CheckAggressive:      true,
		CheckOffTopic:        false, // Off by default as it's subjective
	}
}

// DetermineStatus determines the comment status based on moderation result
func DetermineStatus(result *ModerationResult, config ModerationConfig) string {
	if result.Confidence >= config.AutoRejectThreshold {
		return "rejected"
	} else if result.Confidence <= config.AutoApproveThreshold {
		return "approved"
	}
	return "pending" // Flag for manual review
}
