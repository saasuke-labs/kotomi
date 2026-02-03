package notifications

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Store handles notification database operations
type Store struct {
	db *sql.DB
}

// NewStore creates a new notification store
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// SaveNotification saves a notification to the queue
func (s *Store) SaveNotification(n *Notification) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now()
	}
	if n.UpdatedAt.IsZero() {
		n.UpdatedAt = time.Now()
	}
	if n.Status == "" {
		n.Status = "pending"
	}

	// Marshal data to JSON
	dataJSON, err := json.Marshal(n.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	query := `
		INSERT INTO notification_queue (id, site_id, type, recipient, subject, body, data, status, attempts, error, created_at, sent_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var sentAt sql.NullTime
	if n.SentAt != nil {
		sentAt.Time = *n.SentAt
		sentAt.Valid = true
	}

	var errorStr sql.NullString
	if n.Error != "" {
		errorStr.String = n.Error
		errorStr.Valid = true
	}

	_, err = s.db.Exec(query, n.ID, n.SiteID, n.Type, n.To, n.Subject, n.Body, string(dataJSON), n.Status, n.Attempts, errorStr, n.CreatedAt, sentAt, n.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to save notification: %w", err)
	}

	return nil
}

// GetPendingNotifications retrieves pending notifications
func (s *Store) GetPendingNotifications(limit int) ([]*Notification, error) {
	query := `
		SELECT id, site_id, type, recipient, subject, body, data, status, attempts, error, created_at, sent_at, updated_at
		FROM notification_queue
		WHERE status = 'pending' AND attempts < 3
		ORDER BY created_at ASC
		LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		n := &Notification{}
		var dataJSON string
		var sentAt sql.NullTime
		var errorStr sql.NullString

		err := rows.Scan(&n.ID, &n.SiteID, &n.Type, &n.To, &n.Subject, &n.Body, &dataJSON, &n.Status, &n.Attempts, &errorStr, &n.CreatedAt, &sentAt, &n.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		if sentAt.Valid {
			n.SentAt = &sentAt.Time
		}
		if errorStr.Valid {
			n.Error = errorStr.String
		}

		// Unmarshal data
		if err := json.Unmarshal([]byte(dataJSON), &n.Data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal data: %w", err)
		}

		notifications = append(notifications, n)
	}

	return notifications, nil
}

// UpdateNotificationStatus updates the status of a notification
func (s *Store) UpdateNotificationStatus(id, status, errorMsg string) error {
	now := time.Now()
	var sentAt sql.NullTime
	if status == "sent" {
		sentAt.Time = now
		sentAt.Valid = true
	}

	var errorStr sql.NullString
	if errorMsg != "" {
		errorStr.String = errorMsg
		errorStr.Valid = true
	}

	query := `
		UPDATE notification_queue
		SET status = ?, error = ?, sent_at = ?, updated_at = ?, attempts = attempts + 1
		WHERE id = ?
	`

	_, err := s.db.Exec(query, status, errorStr, sentAt, now, id)
	if err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	return nil
}

// GetSettings retrieves notification settings for a site
func (s *Store) GetSettings(siteID string) (*NotificationSettings, error) {
	query := `
		SELECT id, site_id, enabled, provider, from_email, from_name, reply_to,
		       smtp_host, smtp_port, smtp_user, smtp_password, smtp_encryption,
		       sendgrid_api_key, notify_new_comment, notify_reply, notify_moderation,
		       owner_email, created_at, updated_at
		FROM notification_settings
		WHERE site_id = ?
	`

	settings := &NotificationSettings{}
	var smtpHost, smtpUser, smtpPassword, smtpEncryption, sendGridAPIKey, replyTo sql.NullString
	var smtpPort sql.NullInt64

	err := s.db.QueryRow(query, siteID).Scan(
		&settings.ID, &settings.SiteID, &settings.Enabled, &settings.Provider,
		&settings.FromEmail, &settings.FromName, &replyTo,
		&smtpHost, &smtpPort, &smtpUser, &smtpPassword, &smtpEncryption,
		&sendGridAPIKey, &settings.NotifyNewComment, &settings.NotifyReply,
		&settings.NotifyModeration, &settings.OwnerEmail,
		&settings.CreatedAt, &settings.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	// Set null fields
	if replyTo.Valid {
		settings.ReplyTo = replyTo.String
	}
	if smtpHost.Valid {
		settings.SMTPHost = smtpHost.String
	}
	if smtpPort.Valid {
		settings.SMTPPort = int(smtpPort.Int64)
	}
	if smtpUser.Valid {
		settings.SMTPUser = smtpUser.String
	}
	if smtpPassword.Valid {
		settings.SMTPPassword = smtpPassword.String
	}
	if smtpEncryption.Valid {
		settings.SMTPEncryption = smtpEncryption.String
	}
	if sendGridAPIKey.Valid {
		settings.SendGridAPIKey = sendGridAPIKey.String
	}

	return settings, nil
}

// SaveSettings saves notification settings for a site
func (s *Store) SaveSettings(settings *NotificationSettings) error {
	if settings.ID == "" {
		settings.ID = uuid.New().String()
	}
	if settings.CreatedAt.IsZero() {
		settings.CreatedAt = time.Now()
	}
	settings.UpdatedAt = time.Now()

	// Check if settings exist
	existing, err := s.GetSettings(settings.SiteID)
	if err != nil {
		return err
	}

	var smtpHost, smtpUser, smtpPassword, smtpEncryption, sendGridAPIKey, replyTo sql.NullString
	var smtpPort sql.NullInt64

	if settings.ReplyTo != "" {
		replyTo.String = settings.ReplyTo
		replyTo.Valid = true
	}
	if settings.SMTPHost != "" {
		smtpHost.String = settings.SMTPHost
		smtpHost.Valid = true
	}
	if settings.SMTPPort > 0 {
		smtpPort.Int64 = int64(settings.SMTPPort)
		smtpPort.Valid = true
	}
	if settings.SMTPUser != "" {
		smtpUser.String = settings.SMTPUser
		smtpUser.Valid = true
	}
	if settings.SMTPPassword != "" {
		smtpPassword.String = settings.SMTPPassword
		smtpPassword.Valid = true
	}
	if settings.SMTPEncryption != "" {
		smtpEncryption.String = settings.SMTPEncryption
		smtpEncryption.Valid = true
	}
	if settings.SendGridAPIKey != "" {
		sendGridAPIKey.String = settings.SendGridAPIKey
		sendGridAPIKey.Valid = true
	}

	if existing != nil {
		// Update
		query := `
			UPDATE notification_settings
			SET enabled = ?, provider = ?, from_email = ?, from_name = ?, reply_to = ?,
			    smtp_host = ?, smtp_port = ?, smtp_user = ?, smtp_password = ?, smtp_encryption = ?,
			    sendgrid_api_key = ?, notify_new_comment = ?, notify_reply = ?, notify_moderation = ?,
			    owner_email = ?, updated_at = ?
			WHERE site_id = ?
		`

		_, err = s.db.Exec(query,
			settings.Enabled, settings.Provider, settings.FromEmail, settings.FromName, replyTo,
			smtpHost, smtpPort, smtpUser, smtpPassword, smtpEncryption,
			sendGridAPIKey, settings.NotifyNewComment, settings.NotifyReply, settings.NotifyModeration,
			settings.OwnerEmail, settings.UpdatedAt, settings.SiteID,
		)
	} else {
		// Insert
		query := `
			INSERT INTO notification_settings (
				id, site_id, enabled, provider, from_email, from_name, reply_to,
				smtp_host, smtp_port, smtp_user, smtp_password, smtp_encryption,
				sendgrid_api_key, notify_new_comment, notify_reply, notify_moderation,
				owner_email, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		_, err = s.db.Exec(query,
			settings.ID, settings.SiteID, settings.Enabled, settings.Provider,
			settings.FromEmail, settings.FromName, replyTo,
			smtpHost, smtpPort, smtpUser, smtpPassword, smtpEncryption,
			sendGridAPIKey, settings.NotifyNewComment, settings.NotifyReply, settings.NotifyModeration,
			settings.OwnerEmail, settings.CreatedAt, settings.UpdatedAt,
		)
	}

	if err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	return nil
}

// LogNotification logs a sent notification to history
func (s *Store) LogNotification(n *Notification) error {
	query := `
		INSERT INTO notification_log (id, site_id, type, recipient, subject, status, error, created_at, sent_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var errorStr sql.NullString
	if n.Error != "" {
		errorStr.String = n.Error
		errorStr.Valid = true
	}

	var sentAt sql.NullTime
	if n.SentAt != nil {
		sentAt.Time = *n.SentAt
		sentAt.Valid = true
	}

	_, err := s.db.Exec(query, n.ID, n.SiteID, n.Type, n.To, n.Subject, n.Status, errorStr, n.CreatedAt, sentAt)
	if err != nil {
		return fmt.Errorf("failed to log notification: %w", err)
	}

	return nil
}

// DeleteProcessedNotifications removes old notifications from the queue
func (s *Store) DeleteProcessedNotifications(olderThan time.Time) error {
	query := `
		DELETE FROM notification_queue
		WHERE (status = 'sent' OR attempts >= 3) AND updated_at < ?
	`

	_, err := s.db.Exec(query, olderThan)
	if err != nil {
		return fmt.Errorf("failed to delete notifications: %w", err)
	}

	return nil
}
