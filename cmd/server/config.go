package server

import (
	"database/sql"
	"html/template"

	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/comments"
	"github.com/saasuke-labs/kotomi/pkg/moderation"
	"github.com/saasuke-labs/kotomi/pkg/notifications"
)

// Config holds the configuration for creating a Server
type Config struct {
	CommentStore          *comments.SQLiteStore
	DB                    *sql.DB
	Templates             *template.Template
	Auth0Config           *auth.Auth0Config
	Moderator             moderation.Moderator
	ModerationConfigStore *moderation.ConfigStore
	NotificationQueue     *notifications.Queue
}
