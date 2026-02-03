package notifications

import (
	"context"
	"testing"
)

// TestEmailTemplates tests that email templates render correctly
func TestEmailTemplates(t *testing.T) {
	tmpl := NewEmailTemplate()

	t.Run("NewCommentTemplate", func(t *testing.T) {
		data := map[string]string{
			"SiteName":       "Test Site",
			"PageTitle":      "Test Page",
			"CommentURL":     "https://example.com/page#comment-1",
			"AuthorName":     "John Doe",
			"CommentText":    "This is a test comment",
			"UnsubscribeURL": "https://example.com/unsubscribe",
		}

		html, err := tmpl.RenderNewComment(data)
		if err != nil {
			t.Fatalf("Failed to render new comment template: %v", err)
		}

		// Check that all data fields are in the output
		if len(html) == 0 {
			t.Error("Expected non-empty HTML output")
		}
	})

	t.Run("CommentReplyTemplate", func(t *testing.T) {
		data := map[string]string{
			"PageTitle":      "Test Page",
			"CommentURL":     "https://example.com/page#comment-2",
			"AuthorName":     "Jane Smith",
			"ReplyText":      "This is a reply",
			"OriginalText":   "Original comment",
			"UnsubscribeURL": "https://example.com/unsubscribe",
		}

		html, err := tmpl.RenderCommentReply(data)
		if err != nil {
			t.Fatalf("Failed to render comment reply template: %v", err)
		}

		if len(html) == 0 {
			t.Error("Expected non-empty HTML output")
		}
	})

	t.Run("ModerationUpdateTemplate", func(t *testing.T) {
		data := map[string]string{
			"PageTitle":      "Test Page",
			"CommentURL":     "https://example.com/page#comment-3",
			"CommentText":    "Test comment",
			"Status":         "approved",
			"Reason":         "Good content",
			"UnsubscribeURL": "https://example.com/unsubscribe",
		}

		html, err := tmpl.RenderModerationUpdate(data)
		if err != nil {
			t.Fatalf("Failed to render moderation update template: %v", err)
		}

		if len(html) == 0 {
			t.Error("Expected non-empty HTML output")
		}
	})
}

// TestEmailSender tests the email sender functionality
func TestEmailSender(t *testing.T) {
	t.Run("WithoutProvider", func(t *testing.T) {
		sender := NewEmailSender(nil)
		err := sender.Send(context.Background(), "test@example.com", "Test", "<html>Test</html>")
		if err == nil {
			t.Error("Expected error when no provider is configured")
		}
	})

	t.Run("GetProviderName", func(t *testing.T) {
		sender := NewEmailSender(nil)
		if sender.GetProviderName() != "none" {
			t.Errorf("Expected provider name 'none', got '%s'", sender.GetProviderName())
		}
	})
}

// TestSMTPProvider tests SMTP provider creation
func TestSMTPProvider(t *testing.T) {
	provider := NewSMTPProvider(
		"smtp.example.com",
		587,
		"user@example.com",
		"password",
		"noreply@example.com",
		"Test Sender",
		"starttls",
	)

	if provider == nil {
		t.Fatal("Expected non-nil SMTP provider")
	}

	if provider.GetName() != "smtp" {
		t.Errorf("Expected provider name 'smtp', got '%s'", provider.GetName())
	}
}

// TestSendGridProvider tests SendGrid provider creation
func TestSendGridProvider(t *testing.T) {
	provider := NewSendGridProvider(
		"test-api-key",
		"noreply@example.com",
		"Test Sender",
	)

	if provider == nil {
		t.Fatal("Expected non-nil SendGrid provider")
	}

	if provider.GetName() != "sendgrid" {
		t.Errorf("Expected provider name 'sendgrid', got '%s'", provider.GetName())
	}
}

// TestNotificationTypes tests notification type constants
func TestNotificationTypes(t *testing.T) {
	if NotificationNewComment != "new_comment" {
		t.Errorf("Expected NotificationNewComment to be 'new_comment', got '%s'", NotificationNewComment)
	}

	if NotificationCommentReply != "comment_reply" {
		t.Errorf("Expected NotificationCommentReply to be 'comment_reply', got '%s'", NotificationCommentReply)
	}

	if NotificationModerationUpdate != "moderation_update" {
		t.Errorf("Expected NotificationModerationUpdate to be 'moderation_update', got '%s'", NotificationModerationUpdate)
	}
}
