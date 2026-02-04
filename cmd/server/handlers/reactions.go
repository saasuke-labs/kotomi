package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	apierrors "github.com/saasuke-labs/kotomi/pkg/errors"
	"github.com/saasuke-labs/kotomi/pkg/logging"
	"github.com/saasuke-labs/kotomi/pkg/middleware"
	"github.com/saasuke-labs/kotomi/pkg/models"
)

// GetAllowedReactions retrieves allowed reactions for a site
// @Summary Get allowed reactions
// @Description Retrieve all allowed reactions for a site, optionally filtered by type
// @Tags reactions
// @Produce json
// @Param siteId path string true "Site ID"
// @Param type query string false "Reaction type filter (page or comment)"
// @Success 200 {array} models.AllowedReaction
// @Failure 500 {string} string "Failed to retrieve allowed reactions"
// @Router /site/{siteId}/allowed-reactions [get]
func (s *ServerHandlers) GetAllowedReactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Enrich context with site_id for automatic logging
	ctx := r.Context()
	ctx = logging.WithSiteID(ctx, siteID)

	// Check if type filter is provided
	reactionType := r.URL.Query().Get("type")

	allowedReactionStore := models.NewAllowedReactionStore(s.DB)
	var reactions []models.AllowedReaction
	var err error

	if reactionType != "" && (reactionType == "page" || reactionType == "comment") {
		reactions, err = allowedReactionStore.GetBySiteAndType(ctx, siteID, reactionType)
	} else {
		reactions, err = allowedReactionStore.GetBySite(ctx, siteID)
	}

	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to retrieve allowed reactions", "error", err, "reaction_type", reactionType)
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to retrieve allowed reactions").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	s.WriteJsonResponse(w, reactions)
}

// AddReaction adds a reaction to a comment
func (s *ServerHandlers) AddReaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["commentId"]

	// Enrich context with comment_id for automatic logging
	ctx := r.Context()
	ctx = logging.WithCommentID(ctx, commentID)

	// Get authenticated user from context (set by JWT middleware)
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		apierrors.WriteError(w, apierrors.Unauthorized("Authentication required").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	var req struct {
		AllowedReactionID string `json:"allowed_reaction_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.WriteError(w, apierrors.InvalidJSON("Invalid request body").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	if req.AllowedReactionID == "" {
		apierrors.WriteError(w, apierrors.ValidationError("allowed_reaction_id is required").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	reactionStore := models.NewReactionStore(s.DB)
	reaction, err := reactionStore.AddReaction(ctx, commentID, req.AllowedReactionID, user.ID)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to add reaction", "error", err, "allowed_reaction_id", req.AllowedReactionID, "user_id", user.ID)
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to add reaction").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	// If reaction is nil, it means the user toggled off their reaction
	if reaction == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	s.WriteJsonResponse(w, reaction)
}

// GetReactionsByComment retrieves all reactions for a comment
func (s *ServerHandlers) GetReactionsByComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["commentId"]

	// Enrich context with comment_id for automatic logging
	ctx := r.Context()
	ctx = logging.WithCommentID(ctx, commentID)

	reactionStore := models.NewReactionStore(s.DB)
	reactions, err := reactionStore.GetReactionsByComment(ctx, commentID)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to retrieve reactions", "error", err)
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to retrieve reactions").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	s.WriteJsonResponse(w, reactions)
}

// GetReactionCounts retrieves reaction counts for a comment
func (s *ServerHandlers) GetReactionCounts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["commentId"]

	// Enrich context with comment_id for automatic logging
	ctx := r.Context()
	ctx = logging.WithCommentID(ctx, commentID)

	reactionStore := models.NewReactionStore(s.DB)
	counts, err := reactionStore.GetReactionCounts(ctx, commentID)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to retrieve reaction counts", "error", err)
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to retrieve reaction counts").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	s.WriteJsonResponse(w, counts)
}

// AddPageReaction adds a reaction to a page
func (s *ServerHandlers) AddPageReaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageID := vars["pageId"]

	// Enrich context with page_id for automatic logging
	ctx := r.Context()
	ctx = logging.WithPageID(ctx, pageID)

	// Get authenticated user from context (set by JWT middleware)
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		apierrors.WriteError(w, apierrors.Unauthorized("Authentication required").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	var req struct {
		AllowedReactionID string `json:"allowed_reaction_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.WriteError(w, apierrors.InvalidJSON("Invalid request body").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	if req.AllowedReactionID == "" {
		apierrors.WriteError(w, apierrors.ValidationError("allowed_reaction_id is required").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	reactionStore := models.NewReactionStore(s.DB)
	reaction, err := reactionStore.AddPageReaction(ctx, pageID, req.AllowedReactionID, user.ID)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to add page reaction", "error", err, "allowed_reaction_id", req.AllowedReactionID, "user_id", user.ID)
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to add reaction").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	// If reaction is nil, it means the user toggled off their reaction
	if reaction == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	s.WriteJsonResponse(w, reaction)
}

// GetReactionsByPage retrieves all reactions for a page
func (s *ServerHandlers) GetReactionsByPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageID := vars["pageId"]

	// Enrich context with page_id for automatic logging
	ctx := r.Context()
	ctx = logging.WithPageID(ctx, pageID)

	reactionStore := models.NewReactionStore(s.DB)
	reactions, err := reactionStore.GetReactionsByPage(ctx, pageID)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to retrieve page reactions", "error", err)
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to retrieve reactions").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	s.WriteJsonResponse(w, reactions)
}

// GetPageReactionCounts retrieves reaction counts for a page
func (s *ServerHandlers) GetPageReactionCounts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageID := vars["pageId"]

	// Enrich context with page_id for automatic logging
	ctx := r.Context()
	ctx = logging.WithPageID(ctx, pageID)

	reactionStore := models.NewReactionStore(s.DB)
	counts, err := reactionStore.GetPageReactionCounts(ctx, pageID)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to retrieve page reaction counts", "error", err)
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to retrieve reaction counts").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	s.WriteJsonResponse(w, counts)
}

// RemoveReaction removes a reaction
func (s *ServerHandlers) RemoveReaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reactionID := vars["reactionId"]

	// Enrich context for automatic logging
	ctx := r.Context()

	reactionStore := models.NewReactionStore(s.DB)
	if err := reactionStore.RemoveReaction(ctx, reactionID); err != nil {
		s.Logger.ErrorContext(ctx, "failed to remove reaction", "error", err, "reaction_id", reactionID)
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to remove reaction").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
