package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
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
		log.Printf("Error retrieving allowed reactions: %v", err)
		http.Error(w, "Failed to retrieve allowed reactions", http.StatusInternalServerError)
		return
	}

	WriteJsonResponse(w, reactions)
}

// AddReaction adds a reaction to a comment
func (s *ServerHandlers) AddReaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["commentId"]

	// Get authenticated user from context (set by JWT middleware)
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req struct {
		AllowedReactionID string `json:"allowed_reaction_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.AllowedReactionID == "" {
		http.Error(w, "allowed_reaction_id is required", http.StatusBadRequest)
		return
	}

	reactionStore := models.NewReactionStore(s.DB)
	reaction, err := reactionStore.AddReaction(commentID, req.AllowedReactionID, user.ID)
	if err != nil {
		log.Printf("Error adding reaction: %v", err)
		http.Error(w, "Failed to add reaction", http.StatusInternalServerError)
		return
	}

	// If reaction is nil, it means the user toggled off their reaction
	if reaction == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	WriteJsonResponse(w, reaction)
}

// GetReactionsByComment retrieves all reactions for a comment
func (s *ServerHandlers) GetReactionsByComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["commentId"]

	reactionStore := models.NewReactionStore(s.DB)
	reactions, err := reactionStore.GetReactionsByComment(r.Context(), commentID)
	if err != nil {
		log.Printf("Error retrieving reactions: %v", err)
		http.Error(w, "Failed to retrieve reactions", http.StatusInternalServerError)
		return
	}

	WriteJsonResponse(w, reactions)
}

// GetReactionCounts retrieves reaction counts for a comment
func (s *ServerHandlers) GetReactionCounts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["commentId"]

	reactionStore := models.NewReactionStore(s.DB)
	counts, err := reactionStore.GetReactionCounts(r.Context(), commentID)
	if err != nil {
		log.Printf("Error retrieving reaction counts: %v", err)
		http.Error(w, "Failed to retrieve reaction counts", http.StatusInternalServerError)
		return
	}

	WriteJsonResponse(w, counts)
}

// AddPageReaction adds a reaction to a page
func (s *ServerHandlers) AddPageReaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageID := vars["pageId"]

	// Get authenticated user from context (set by JWT middleware)
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req struct {
		AllowedReactionID string `json:"allowed_reaction_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.AllowedReactionID == "" {
		http.Error(w, "allowed_reaction_id is required", http.StatusBadRequest)
		return
	}

	reactionStore := models.NewReactionStore(s.DB)
	reaction, err := reactionStore.AddPageReaction(pageID, req.AllowedReactionID, user.ID)
	if err != nil {
		log.Printf("Error adding page reaction: %v", err)
		http.Error(w, "Failed to add reaction", http.StatusInternalServerError)
		return
	}

	// If reaction is nil, it means the user toggled off their reaction
	if reaction == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	WriteJsonResponse(w, reaction)
}

// GetReactionsByPage retrieves all reactions for a page
func (s *ServerHandlers) GetReactionsByPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageID := vars["pageId"]

	reactionStore := models.NewReactionStore(s.DB)
	reactions, err := reactionStore.GetReactionsByPage(r.Context(), pageID)
	if err != nil {
		log.Printf("Error retrieving page reactions: %v", err)
		http.Error(w, "Failed to retrieve reactions", http.StatusInternalServerError)
		return
	}

	WriteJsonResponse(w, reactions)
}

// GetPageReactionCounts retrieves reaction counts for a page
func (s *ServerHandlers) GetPageReactionCounts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageID := vars["pageId"]

	reactionStore := models.NewReactionStore(s.DB)
	counts, err := reactionStore.GetPageReactionCounts(r.Context(), pageID)
	if err != nil {
		log.Printf("Error retrieving page reaction counts: %v", err)
		http.Error(w, "Failed to retrieve reaction counts", http.StatusInternalServerError)
		return
	}

	WriteJsonResponse(w, counts)
}

// RemoveReaction removes a reaction
func (s *ServerHandlers) RemoveReaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reactionID := vars["reactionId"]

	reactionStore := models.NewReactionStore(s.DB)
	if err := reactionStore.RemoveReaction(r.Context(), reactionID); err != nil {
		log.Printf("Error removing reaction: %v", err)
		http.Error(w, "Failed to remove reaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
