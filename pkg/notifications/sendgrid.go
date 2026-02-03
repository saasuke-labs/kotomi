package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Compile-time check to ensure SendGridProvider implements EmailProvider interface
var _ EmailProvider = (*SendGridProvider)(nil)

// SendGridProvider implements email sending via SendGrid API
type SendGridProvider struct {
	apiKey    string
	fromEmail string
	fromName  string
}

// NewSendGridProvider creates a new SendGrid email provider
func NewSendGridProvider(apiKey, fromEmail, fromName string) *SendGridProvider {
	return &SendGridProvider{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

// sendGridRequest represents the SendGrid API v3 request format
type sendGridRequest struct {
	Personalizations []sendGridPersonalization `json:"personalizations"`
	From             sendGridEmail             `json:"from"`
	Subject          string                    `json:"subject"`
	Content          []sendGridContent         `json:"content"`
}

type sendGridPersonalization struct {
	To []sendGridEmail `json:"to"`
}

type sendGridEmail struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type sendGridContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// SendEmail sends an email via SendGrid API
func (p *SendGridProvider) SendEmail(ctx context.Context, to, subject, htmlBody string) error {
	// Build request
	req := sendGridRequest{
		Personalizations: []sendGridPersonalization{
			{
				To: []sendGridEmail{
					{Email: to},
				},
			},
		},
		From: sendGridEmail{
			Email: p.fromEmail,
			Name:  p.fromName,
		},
		Subject: subject,
		Content: []sendGridContent{
			{
				Type:  "text/html",
				Value: htmlBody,
			},
		},
	}

	// Marshal request
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("sendgrid API error: status %d", resp.StatusCode)
	}

	return nil
}

// GetName returns the provider name
func (p *SendGridProvider) GetName() string {
	return "sendgrid"
}
