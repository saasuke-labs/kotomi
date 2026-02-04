package moderation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Compile-time check to ensure OpenAIModerator implements Moderator interface
var _ Moderator = (*OpenAIModerator)(nil)

// OpenAIModerator implements the Moderator interface using OpenAI's API
type OpenAIModerator struct {
	APIKey     string
	Model      string // e.g., "gpt-3.5-turbo" or "gpt-4"
	HTTPClient *http.Client
}

// NewOpenAIModerator creates a new OpenAI-based moderator
func NewOpenAIModerator(apiKey string) *OpenAIModerator {
	return &OpenAIModerator{
		APIKey:     apiKey,
		Model:      "gpt-3.5-turbo", // Default to cost-effective model
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// OpenAI API structures
type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// AnalyzeComment analyzes a comment using OpenAI's GPT model
func (m *OpenAIModerator) AnalyzeComment(text string, config ModerationConfig) (*ModerationResult, error) {
	if m.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is not configured")
	}

	// Build the moderation prompt based on config
	prompt := m.buildPrompt(text, config)

	// Call OpenAI API
	requestBody := openAIRequest{
		Model: m.Model,
		Messages: []openAIMessage{
			{
				Role:    "system",
				Content: "You are a content moderation assistant. Analyze the provided comment and respond with a JSON object containing your analysis.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.APIKey)

	resp, err := m.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp openAIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse the AI response
	result, err := m.parseAIResponse(apiResp.Choices[0].Message.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	result.AnalyzedAt = time.Now()
	return result, nil
}

// buildPrompt constructs the moderation prompt based on configuration
func (m *OpenAIModerator) buildPrompt(text string, config ModerationConfig) string {
	var checks []string
	if config.CheckSpam {
		checks = append(checks, "spam")
	}
	if config.CheckOffensive {
		checks = append(checks, "offensive language")
	}
	if config.CheckAggressive {
		checks = append(checks, "aggressive or hostile tone")
	}
	if config.CheckOffTopic {
		checks = append(checks, "off-topic content")
	}

	checksStr := strings.Join(checks, ", ")

	return fmt.Sprintf(`Analyze the following comment for: %s.

Comment: "%s"

Respond with a JSON object in this exact format:
{
  "confidence": <number between 0 and 1, where 1 means definitely problematic>,
  "reason": "<brief explanation>",
  "categories": [<list of detected issues from: "spam", "offensive", "aggressive", "off_topic">]
}

Be strict but fair. Only flag content that clearly violates standards.`, checksStr, text)
}

// parseAIResponse parses the JSON response from the AI
func (m *OpenAIModerator) parseAIResponse(content string) (*ModerationResult, error) {
	// Extract JSON from the response (in case there's extra text)
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start == -1 || end == -1 {
		return nil, fmt.Errorf("no JSON found in AI response")
	}
	jsonStr := content[start : end+1]

	var parsed struct {
		Confidence float64  `json:"confidence"`
		Reason     string   `json:"reason"`
		Categories []string `json:"categories"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Determine decision based on confidence
	var decision string
	if parsed.Confidence >= 0.7 {
		decision = "reject"
	} else if parsed.Confidence <= 0.3 {
		decision = "approve"
	} else {
		decision = "flag"
	}

	return &ModerationResult{
		Decision:   decision,
		Confidence: parsed.Confidence,
		Reason:     parsed.Reason,
		Categories: parsed.Categories,
	}, nil
}
