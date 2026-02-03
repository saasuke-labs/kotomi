package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// Queue manages the notification processing queue
type Queue struct {
	store      *Store
	templates  *EmailTemplate
	db         *sql.DB
	stopChan   chan struct{}
	interval   time.Duration
	batchSize  int
}

// NewQueue creates a new notification queue processor
func NewQueue(db *sql.DB, interval time.Duration, batchSize int) *Queue {
	return &Queue{
		store:     NewStore(db),
		templates: NewEmailTemplate(),
		db:        db,
		stopChan:  make(chan struct{}),
		interval:  interval,
		batchSize: batchSize,
	}
}

// Start begins processing the notification queue
func (q *Queue) Start(ctx context.Context) {
	ticker := time.NewTicker(q.interval)
	defer ticker.Stop()

	log.Println("Notification queue processor started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Notification queue processor stopping...")
			return
		case <-q.stopChan:
			log.Println("Notification queue processor stopped")
			return
		case <-ticker.C:
			q.processBatch(ctx)
		}
	}
}

// Stop stops the queue processor
func (q *Queue) Stop() {
	close(q.stopChan)
}

// processBatch processes a batch of pending notifications
func (q *Queue) processBatch(ctx context.Context) {
	// Get pending notifications
	notifications, err := q.store.GetPendingNotifications(q.batchSize)
	if err != nil {
		log.Printf("Error fetching pending notifications: %v", err)
		return
	}

	if len(notifications) == 0 {
		return // No notifications to process
	}

	log.Printf("Processing %d pending notifications", len(notifications))

	for _, notification := range notifications {
		q.processNotification(ctx, notification)
	}

	// Clean up old processed notifications (older than 7 days)
	oldDate := time.Now().AddDate(0, 0, -7)
	if err := q.store.DeleteProcessedNotifications(oldDate); err != nil {
		log.Printf("Error cleaning up old notifications: %v", err)
	}
}

// processNotification processes a single notification
func (q *Queue) processNotification(ctx context.Context, n *Notification) {
	// Get notification settings for the site
	settings, err := q.store.GetSettings(n.SiteID)
	if err != nil {
		log.Printf("Error getting notification settings for site %s: %v", n.SiteID, err)
		q.store.UpdateNotificationStatus(n.ID, "failed", fmt.Sprintf("Failed to get settings: %v", err))
		return
	}

	// Check if notifications are enabled
	if settings == nil || !settings.Enabled {
		log.Printf("Notifications disabled for site %s, skipping", n.SiteID)
		q.store.UpdateNotificationStatus(n.ID, "failed", "Notifications not enabled for site")
		return
	}

	// Check notification type settings
	switch n.Type {
	case NotificationNewComment:
		if !settings.NotifyNewComment {
			q.store.UpdateNotificationStatus(n.ID, "failed", "New comment notifications disabled")
			return
		}
	case NotificationCommentReply:
		if !settings.NotifyReply {
			q.store.UpdateNotificationStatus(n.ID, "failed", "Reply notifications disabled")
			return
		}
	case NotificationModerationUpdate:
		if !settings.NotifyModeration {
			q.store.UpdateNotificationStatus(n.ID, "failed", "Moderation notifications disabled")
			return
		}
	}

	// Create email provider based on settings
	var provider EmailProvider
	switch settings.Provider {
	case "smtp":
		provider = NewSMTPProvider(
			settings.SMTPHost,
			settings.SMTPPort,
			settings.SMTPUser,
			settings.SMTPPassword,
			settings.FromEmail,
			settings.FromName,
			settings.SMTPEncryption,
		)
	case "sendgrid":
		provider = NewSendGridProvider(
			settings.SendGridAPIKey,
			settings.FromEmail,
			settings.FromName,
		)
	default:
		log.Printf("Unknown email provider: %s", settings.Provider)
		q.store.UpdateNotificationStatus(n.ID, "failed", fmt.Sprintf("Unknown provider: %s", settings.Provider))
		return
	}

	// Create email sender
	sender := NewEmailSender(provider)

	// Send the email
	err = sender.Send(ctx, n.To, n.Subject, n.Body)
	if err != nil {
		log.Printf("Error sending notification %s: %v", n.ID, err)
		q.store.UpdateNotificationStatus(n.ID, "failed", fmt.Sprintf("Send failed: %v", err))
		return
	}

	// Mark as sent
	if err := q.store.UpdateNotificationStatus(n.ID, "sent", ""); err != nil {
		log.Printf("Error updating notification status: %v", err)
		return
	}

	// Log to history
	now := time.Now()
	n.SentAt = &now
	n.Status = "sent"
	if err := q.store.LogNotification(n); err != nil {
		log.Printf("Error logging notification: %v", err)
	}

	log.Printf("Successfully sent notification %s to %s", n.ID, n.To)
}

// EnqueueNewComment enqueues a new comment notification
func (q *Queue) EnqueueNewComment(siteID, siteName, pageTitle, commentURL, authorName, commentText, ownerEmail, unsubscribeURL string) error {
	data := map[string]string{
		"SiteName":       siteName,
		"PageTitle":      pageTitle,
		"CommentURL":     commentURL,
		"AuthorName":     authorName,
		"CommentText":    commentText,
		"UnsubscribeURL": unsubscribeURL,
	}

	body, err := q.templates.RenderNewComment(data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	notification := &Notification{
		SiteID:  siteID,
		Type:    NotificationNewComment,
		To:      ownerEmail,
		Subject: fmt.Sprintf("New comment on %s", siteName),
		Body:    body,
		Data:    data,
		Status:  "pending",
	}

	return q.store.SaveNotification(notification)
}

// EnqueueCommentReply enqueues a comment reply notification
func (q *Queue) EnqueueCommentReply(siteID, pageTitle, commentURL, authorName, replyText, originalText, recipientEmail, unsubscribeURL string) error {
	data := map[string]string{
		"PageTitle":      pageTitle,
		"CommentURL":     commentURL,
		"AuthorName":     authorName,
		"ReplyText":      replyText,
		"OriginalText":   originalText,
		"UnsubscribeURL": unsubscribeURL,
	}

	body, err := q.templates.RenderCommentReply(data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	notification := &Notification{
		SiteID:  siteID,
		Type:    NotificationCommentReply,
		To:      recipientEmail,
		Subject: "Someone replied to your comment",
		Body:    body,
		Data:    data,
		Status:  "pending",
	}

	return q.store.SaveNotification(notification)
}

// EnqueueModerationUpdate enqueues a moderation update notification
func (q *Queue) EnqueueModerationUpdate(siteID, pageTitle, commentURL, commentText, status, reason, recipientEmail, unsubscribeURL string) error {
	data := map[string]string{
		"PageTitle":      pageTitle,
		"CommentURL":     commentURL,
		"CommentText":    commentText,
		"Status":         status,
		"Reason":         reason,
		"UnsubscribeURL": unsubscribeURL,
	}

	body, err := q.templates.RenderModerationUpdate(data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	notification := &Notification{
		SiteID:  siteID,
		Type:    NotificationModerationUpdate,
		To:      recipientEmail,
		Subject: fmt.Sprintf("Your comment was %s", status),
		Body:    body,
		Data:    data,
		Status:  "pending",
	}

	return q.store.SaveNotification(notification)
}
