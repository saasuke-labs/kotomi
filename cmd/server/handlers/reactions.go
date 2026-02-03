package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	apierrors "github.com/saasuke-labs/kotomi/pkg/errors"
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

	// Check if type filter is provided
	reactionType := r.URL.Query().Get("type")

	allowedReactionStore := models.NewAllowedReactionStore(s.DB)
	var reactions []models.AllowedReaction
	var err error

	if reactionType != "" && (reactionType == "page" || reactionType == "comment") {
		reactions, err = allowedReactionStore.GetBySiteAndType(r.Context(), siteID, reactionType)
	} else {
		reactions, err = allowedReactionStore.GetBySite(r.Context(), siteID)
	}

	if err != nil {
		requestID := middleware.GetRequestID(r)
		s.Logger.Error("failed to retrieve allowed reactions",
			"error", err,
			"site_id", siteID,
			"reaction_type", reactionType,
			"request_id", requestID)
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to retrieve allowed reactions").WithRequestID(requestID))
		return
	}

	s.WriteJsonResponse(w, reactions)
}

// AddReaction adds a reaction to a comment
func (s *ServerHandlers) AddReaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["commentId"]

	// Get authenticated user from context (set by JWT middleware)
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		requestID := middleware.GetRequestID(r)
		apierrors.WriteError(w, apierrors.Unauthorized("Authentication required").WithRequestID(requestID))
		return
	}

	var req struct {
		AllowedReactionID string `json:"allowed_reaction_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		requestID := middleware.GetRequestID(r)
		apierrors.WriteError(w, apierrors.InvalidJSON("Invalid request body").WithRequestID(requestID))
		return
	}

	if req.AllowedReactionID == "" {
		requestID := middleware.GetRequestID(r)
		apierrors.WriteError(w, apierrors.ValidationError("allowed_reaction_id is required").WithRequestID(requestID))
		return
	}

	reactionStore := models.NewReactionStore(s.DB)
	reaction, err := reactionStore.AddReaction(r.Context(), commentID, req.AllowedReactionID, user.ID)
	if err != nil {
		requestID := middleware.GetRequestID(r)
		s.Logger.Error("failed to add reaction",
			"error", err,
			"comment_id", commentID,
			"allowed_reaction_id", req.AllowedReactionID,
			"user_id", user.ID,
			"request_id", requestID)
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to add reaction").WithRequestID(requestID))
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

	reactionStore := models.NewReactionStore(s.DB)
	reactions, err := reactionStore.GetReactionsByComment(r.Context(), commentID)
	if err != nil {
		requestID := middleware.GetRequestID(r)
		s.Logger.Error("failed to retrieve reactions",
			"error", err,
			"comment_id", commentID,
			"request_id", requestID)
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to retrieve reactions").WithRequestID(requestID))
		return
	}

	s.WriteJsonResponse(w, reactions)
}

// GetReactionCounts retrieves reaction counts for a comment
func (s *ServerHandlers) GetReactionCounts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["commentId"]

	reactionStore := models.NewReactionStore(s.DB)
	counts, err := reactionStore.GetReactionCounts(r.Context(), commentID)
	if err != nil {
		requestID := middleware.GetRequestID(r)
		s.Logger.Error("failed to retrieve reaction counts",
			"error", err,
			"comment_id", commentID,
			"request_id", requestID)
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to retrieve reaction counts").WithRequestID(requestID))
		return
	}

	s.WriteJsonResponse(w, counts)
}

// AddPageReaction adds a reaction to a page
func (s *ServerHandlers) AddPageReaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageID := vars["pageId"]

	// Get authenticated user from context (set by JWT middleware)
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		requestID := middleware.GetRequestID(r)
		apierrors.WriteError(w, apierrors.Unauthorized("Authentication required").WithRequestID(requestID))
		return
	}

	var req struct {
		AllowedReactionID string `json:"allowed_reaction_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		requestID := middleware.GetRequestID(r)
		apierrors.WriteError(w, apierrors.InvalidJSON("Invalid request body").WithRequestID(requestID))
		return
	}

	if req.AllowedReactionID == "" {
		requestID := middleware.GetRequestID(r)
		apierrors.WriteError(w, apierrors.ValidationError("allowed_reaction_id is required").WithRequestID(requestID))
		return
	}

	reactionStore := models.NewReactionStore(s.DB)
	reaction, err := reactionStore.AddPageReaction(r.Context(), pageID, req.AllowedReactionID, user.ID)
	if err != nil {
		requestID := middleware.GetRequestID(r)
		s.Logger.Error("failed to add page reaction",
			"error", err,
			"page_id", pageID,
			"allowed_reaction_id", req.AllowedReactionID,
			"user_id", user.ID,
			"request_id", requestID)
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to add reaction").WithRequestID(requestID))
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

	reactionStore := models.NewReactionStore(s.DB)
	reactions, err := reactionStore.GetReactionsByPage(r.Context(), pageID)
	if err != nil {
		requestID := middleware.GetRequestID(r)
		s.Logger.Error("failed to retrieve page reactions",
			"error", err,
			"page_id", pageID,
			"request_id", requestID)
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to retrieve reactions").WithRequestID(requestID))
		return
	}

	s.WriteJsonResponse(w, reactions)
}

// GetPageReactionCounts retrieves reaction counts for a page
func (s *ServerHandlers) GetPageReactionCounts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageID := vars["pageId"]

	reactionStore := models.NewReactionStore(s.DB)
	counts, err := reactionStore.GetPageReactionCounts(r.Context(), pageID)
	if err != nil {
		requestID := middleware.GetRequestID(r)
		s.Logger.Error("failed to retrieve page reaction counts",
			"error", err,
			"page_id", pageID,
			"request_id", requestID)
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to retrieve reaction counts").WithRequestID(requestID))
		return
	}

	s.WriteJsonResponse(w, counts)
}

// RemoveReaction removes a reaction
func (s *ServerHandlers) RemoveReaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reactionID := vars["reactionId"]

	reactionStore := models.NewReactionStore(s.DB)
	if err := reactionStore.RemoveReaction(r.Context(), reactionID); err != nil {
		requestID := middleware.GetRequestID(r)
		s.Logger.Error("failed to remove reaction",
			"error", err,
			"reaction_id", reactionID,
			"request_id", requestID)
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to remove reaction").WithRequestID(requestID))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
