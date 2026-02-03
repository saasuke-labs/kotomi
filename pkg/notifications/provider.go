package notifications

import (
	"context"
	"fmt"
)

// EmailProvider defines the interface for sending emails
type EmailProvider interface {
	// SendEmail sends an email with the given parameters
	SendEmail(ctx context.Context, to, subject, htmlBody string) error
	// GetName returns the provider name
	GetName() string
}

// EmailSender manages email sending operations
type EmailSender struct {
	provider EmailProvider
}

// NewEmailSender creates a new email sender with the given provider
func NewEmailSender(provider EmailProvider) *EmailSender {
	return &EmailSender{
		provider: provider,
	}
}

// Send sends an email using the configured provider
func (s *EmailSender) Send(ctx context.Context, to, subject, htmlBody string) error {
	if s.provider == nil {
		return fmt.Errorf("no email provider configured")
	}
	return s.provider.SendEmail(ctx, to, subject, htmlBody)
}

// GetProviderName returns the name of the current provider
func (s *EmailSender) GetProviderName() string {
	if s.provider == nil {
		return "none"
	}
	return s.provider.GetName()
}
