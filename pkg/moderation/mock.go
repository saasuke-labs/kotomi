package moderation

import (
	"strings"
	"time"
)

// Compile-time check to ensure MockModerator implements Moderator interface
var _ Moderator = (*MockModerator)(nil)

// MockModerator is a simple rule-based moderator for testing or when AI is not available
type MockModerator struct{}

// NewMockModerator creates a new mock moderator
func NewMockModerator() *MockModerator {
	return &MockModerator{}
}

// AnalyzeComment performs simple rule-based moderation
func (m *MockModerator) AnalyzeComment(text string, config ModerationConfig) (*ModerationResult, error) {
	result := &ModerationResult{
		Decision:   "approve",
		Confidence: 0.0,
		Reason:     "No issues detected",
		Categories: []string{},
		AnalyzedAt: time.Now(),
	}

	textLower := strings.ToLower(text)

	// Check for spam patterns
	if config.CheckSpam {
		spamKeywords := []string{"buy now", "click here", "limited offer", "act now", "viagra", "casino", "lottery", "prize"}
		for _, keyword := range spamKeywords {
			if strings.Contains(textLower, keyword) {
				result.Categories = append(result.Categories, "spam")
				result.Confidence += 0.3
				break
			}
		}
		
		// Check for excessive links
		linkCount := strings.Count(text, "http://") + strings.Count(text, "https://")
		if linkCount > 2 {
			result.Categories = append(result.Categories, "spam")
			result.Confidence += 0.2
		}
	}

	// Check for offensive language
	if config.CheckOffensive {
		offensiveWords := []string{"fuck", "shit", "damn", "ass", "bitch", "bastard", "crap"}
		for _, word := range offensiveWords {
			if strings.Contains(textLower, word) {
				result.Categories = append(result.Categories, "offensive")
				result.Confidence += 0.4
				break
			}
		}
	}

	// Check for aggressive tone
	if config.CheckAggressive {
		aggressivePatterns := []string{"you're stupid", "you idiot", "shut up", "you're wrong", "you suck"}
		for _, pattern := range aggressivePatterns {
			if strings.Contains(textLower, pattern) {
				result.Categories = append(result.Categories, "aggressive")
				result.Confidence += 0.5
				break
			}
		}
		
		// Check for excessive caps
		capsCount := 0
		for _, ch := range text {
			if ch >= 'A' && ch <= 'Z' {
				capsCount++
			}
		}
		if len(text) > 10 && float64(capsCount)/float64(len(text)) > 0.7 {
			result.Categories = append(result.Categories, "aggressive")
			result.Confidence += 0.3
		}
	}

	// Cap confidence at 1.0
	if result.Confidence > 1.0 {
		result.Confidence = 1.0
	}

	// Determine decision
	if result.Confidence >= 0.7 {
		result.Decision = "reject"
		result.Reason = "Content appears to be problematic"
	} else if result.Confidence > 0.3 {
		result.Decision = "flag"
		result.Reason = "Content may need review"
	} else {
		result.Decision = "approve"
		result.Reason = "No issues detected"
	}

	return result, nil
}
