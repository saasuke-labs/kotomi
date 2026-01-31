package moderation

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ConfigStore handles moderation configuration database operations
type ConfigStore struct {
	db *sql.DB
}

// NewConfigStore creates a new moderation config store
func NewConfigStore(db *sql.DB) *ConfigStore {
	return &ConfigStore{db: db}
}

// GetBySiteID retrieves moderation configuration for a site
func (s *ConfigStore) GetBySiteID(siteID string) (*ModerationConfig, error) {
	query := `
		SELECT enabled, auto_reject_threshold, auto_approve_threshold,
		       check_spam, check_offensive, check_aggressive, check_off_topic
		FROM moderation_config
		WHERE site_id = ?
	`

	var config ModerationConfig
	var enabled, checkSpam, checkOffensive, checkAggressive, checkOffTopic int

	err := s.db.QueryRow(query, siteID).Scan(
		&enabled, &config.AutoRejectThreshold, &config.AutoApproveThreshold,
		&checkSpam, &checkOffensive, &checkAggressive, &checkOffTopic,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return default config if none exists
			return nil, fmt.Errorf("no moderation config found for site")
		}
		return nil, fmt.Errorf("failed to query moderation config: %w", err)
	}

	// Convert integers to booleans
	config.Enabled = enabled == 1
	config.CheckSpam = checkSpam == 1
	config.CheckOffensive = checkOffensive == 1
	config.CheckAggressive = checkAggressive == 1
	config.CheckOffTopic = checkOffTopic == 1

	return &config, nil
}

// Create creates moderation configuration for a site
func (s *ConfigStore) Create(siteID string, config ModerationConfig) error {
	now := time.Now()
	id := uuid.NewString()

	query := `
		INSERT INTO moderation_config 
		(id, site_id, enabled, auto_reject_threshold, auto_approve_threshold,
		 check_spam, check_offensive, check_aggressive, check_off_topic,
		 created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Convert booleans to integers
	enabled := 0
	if config.Enabled {
		enabled = 1
	}
	checkSpam := 0
	if config.CheckSpam {
		checkSpam = 1
	}
	checkOffensive := 0
	if config.CheckOffensive {
		checkOffensive = 1
	}
	checkAggressive := 0
	if config.CheckAggressive {
		checkAggressive = 1
	}
	checkOffTopic := 0
	if config.CheckOffTopic {
		checkOffTopic = 1
	}

	_, err := s.db.Exec(query, id, siteID, enabled, config.AutoRejectThreshold, config.AutoApproveThreshold,
		checkSpam, checkOffensive, checkAggressive, checkOffTopic, now, now)
	if err != nil {
		return fmt.Errorf("failed to create moderation config: %w", err)
	}

	return nil
}

// Update updates moderation configuration for a site
func (s *ConfigStore) Update(siteID string, config ModerationConfig) error {
	query := `
		UPDATE moderation_config
		SET enabled = ?, auto_reject_threshold = ?, auto_approve_threshold = ?,
		    check_spam = ?, check_offensive = ?, check_aggressive = ?, check_off_topic = ?,
		    updated_at = ?
		WHERE site_id = ?
	`

	// Convert booleans to integers
	enabled := 0
	if config.Enabled {
		enabled = 1
	}
	checkSpam := 0
	if config.CheckSpam {
		checkSpam = 1
	}
	checkOffensive := 0
	if config.CheckOffensive {
		checkOffensive = 1
	}
	checkAggressive := 0
	if config.CheckAggressive {
		checkAggressive = 1
	}
	checkOffTopic := 0
	if config.CheckOffTopic {
		checkOffTopic = 1
	}

	result, err := s.db.Exec(query, enabled, config.AutoRejectThreshold, config.AutoApproveThreshold,
		checkSpam, checkOffensive, checkAggressive, checkOffTopic, time.Now(), siteID)
	if err != nil {
		return fmt.Errorf("failed to update moderation config: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("no moderation config found for site")
	}

	return nil
}

// Delete deletes moderation configuration for a site
func (s *ConfigStore) Delete(siteID string) error {
	query := `DELETE FROM moderation_config WHERE site_id = ?`

	_, err := s.db.Exec(query, siteID)
	if err != nil {
		return fmt.Errorf("failed to delete moderation config: %w", err)
	}

	return nil
}
