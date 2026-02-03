package notifications

import "time"

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationNewComment       NotificationType = "new_comment"
	NotificationCommentReply     NotificationType = "comment_reply"
	NotificationModerationUpdate NotificationType = "moderation_update"
)

// Notification represents a notification to be sent
type Notification struct {
	ID        string           `json:"id"`
	SiteID    string           `json:"site_id"`
	Type      NotificationType `json:"type"`
	To        string           `json:"to"` // Email address
	Subject   string           `json:"subject"`
	Body      string           `json:"body"` // HTML body
	Data      map[string]string `json:"data"` // Additional data for template
	Status    string           `json:"status"` // pending, sent, failed
	Attempts  int              `json:"attempts"`
	Error     string           `json:"error,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	SentAt    *time.Time       `json:"sent_at,omitempty"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// NotificationSettings represents site-level notification configuration
type NotificationSettings struct {
	ID                   string    `json:"id"`
	SiteID               string    `json:"site_id"`
	Enabled              bool      `json:"enabled"`
	Provider             string    `json:"provider"` // smtp, sendgrid, ses, mailgun
	FromEmail            string    `json:"from_email"`
	FromName             string    `json:"from_name"`
	ReplyTo              string    `json:"reply_to"`
	SMTPHost             string    `json:"smtp_host,omitempty"`
	SMTPPort             int       `json:"smtp_port,omitempty"`
	SMTPUser             string    `json:"smtp_user,omitempty"`
	SMTPPassword         string    `json:"smtp_password,omitempty"`
	SMTPEncryption       string    `json:"smtp_encryption,omitempty"` // tls, starttls, none
	SendGridAPIKey       string    `json:"sendgrid_api_key,omitempty"`
	NotifyNewComment     bool      `json:"notify_new_comment"`
	NotifyReply          bool      `json:"notify_reply"`
	NotifyModeration     bool      `json:"notify_moderation"`
	OwnerEmail           string    `json:"owner_email"` // Site owner email for notifications
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
