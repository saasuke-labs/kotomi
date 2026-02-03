package server

import (
	"database/sql"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/comments"
	"github.com/saasuke-labs/kotomi/pkg/moderation"
	"github.com/saasuke-labs/kotomi/pkg/notifications"
)

// Server holds all dependencies for the application
type Server struct {
	CommentStore          *comments.SQLiteStore
	DB                    *sql.DB
	Templates             *template.Template
	Auth0Config           *auth.Auth0Config
	Moderator             moderation.Moderator
	ModerationConfigStore *moderation.ConfigStore
	NotificationQueue     *notifications.Queue
}

// New creates a new Server instance with the provided configuration
func New(cfg Config) (*Server, error) {
	server := &Server{
		CommentStore:          cfg.CommentStore,
		DB:                    cfg.DB,
		Templates:             cfg.Templates,
		Auth0Config:           cfg.Auth0Config,
		Moderator:             cfg.Moderator,
		ModerationConfigStore: cfg.ModerationConfigStore,
		NotificationQueue:     cfg.NotificationQueue,
	}
	
	return server, nil
}

// Handler creates and returns the HTTP handler for the server
func (s *Server) Handler() http.Handler {
	router := mux.NewRouter()
	s.RegisterRoutes(router)
	return router
}
