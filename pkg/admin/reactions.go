package admin

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/models"
)

// ReactionsHandler handles admin operations for reactions
type ReactionsHandler struct {
	db        *sql.DB
	templates *template.Template
}

// NewReactionsHandler creates a new reactions handler
func NewReactionsHandler(db *sql.DB, templates *template.Template) *ReactionsHandler {
	return &ReactionsHandler{
		db:        db,
		templates: templates,
	}
}

// ListAllowedReactions shows all allowed reactions for a site
func (h *ReactionsHandler) ListAllowedReactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Verify user owns the site
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	// Get allowed reactions
	allowedReactionStore := models.NewAllowedReactionStore(h.db)
	reactions, err := allowedReactionStore.GetBySite(siteID)
	if err != nil {
		log.Printf("Error fetching allowed reactions: %v", err)
		http.Error(w, "Failed to fetch reactions", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Site":      site,
		"Reactions": reactions,
	}

	if err := h.templates.ExecuteTemplate(w, "admin/reactions/list.html", data); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// ShowReactionForm shows the form for creating/editing an allowed reaction
func (h *ReactionsHandler) ShowReactionForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]
	reactionID := vars["reactionId"]

	// Verify user owns the site
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	var reaction *models.AllowedReaction
	if reactionID != "" {
		// Edit mode
		allowedReactionStore := models.NewAllowedReactionStore(h.db)
		reaction, err = allowedReactionStore.GetByID(reactionID)
		if err != nil {
			http.Error(w, "Reaction not found", http.StatusNotFound)
			return
		}
		if reaction.SiteID != siteID {
			http.Error(w, "Reaction not found", http.StatusNotFound)
			return
		}
	}

	data := map[string]interface{}{
		"Site":     site,
		"Reaction": reaction,
		"IsEdit":   reactionID != "",
	}

	if err := h.templates.ExecuteTemplate(w, "admin/reactions/form.html", data); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// CreateAllowedReaction creates a new allowed reaction
func (h *ReactionsHandler) CreateAllowedReaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Verify user owns the site
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	// Parse form
	name := r.FormValue("name")
	emoji := r.FormValue("emoji")
	reactionType := r.FormValue("reaction_type")

	if name == "" || emoji == "" {
		http.Error(w, "Name and emoji are required", http.StatusBadRequest)
		return
	}

	// Default to comment if not specified
	if reactionType == "" {
		reactionType = "comment"
	}

	// Create reaction
	allowedReactionStore := models.NewAllowedReactionStore(h.db)
	_, err = allowedReactionStore.Create(siteID, name, emoji, reactionType)
	if err != nil {
		log.Printf("Error creating allowed reaction: %v", err)
		http.Error(w, "Failed to create reaction", http.StatusInternalServerError)
		return
	}

	// Redirect back to list
	http.Redirect(w, r, "/admin/sites/"+siteID+"/reactions", http.StatusSeeOther)
}

// UpdateAllowedReaction updates an allowed reaction
func (h *ReactionsHandler) UpdateAllowedReaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]
	reactionID := vars["reactionId"]

	// Verify user owns the site
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	// Verify reaction belongs to site
	allowedReactionStore := models.NewAllowedReactionStore(h.db)
	reaction, err := allowedReactionStore.GetByID(reactionID)
	if err != nil || reaction.SiteID != siteID {
		http.Error(w, "Reaction not found", http.StatusNotFound)
		return
	}

	// Parse form
	name := r.FormValue("name")
	emoji := r.FormValue("emoji")
	reactionType := r.FormValue("reaction_type")

	if name == "" || emoji == "" {
		http.Error(w, "Name and emoji are required", http.StatusBadRequest)
		return
	}

	// Default to comment if not specified
	if reactionType == "" {
		reactionType = "comment"
	}

	// Update reaction
	if err := allowedReactionStore.Update(reactionID, name, emoji, reactionType); err != nil {
		log.Printf("Error updating allowed reaction: %v", err)
		http.Error(w, "Failed to update reaction", http.StatusInternalServerError)
		return
	}

	// Redirect back to list
	http.Redirect(w, r, "/admin/sites/"+siteID+"/reactions", http.StatusSeeOther)
}

// DeleteAllowedReaction deletes an allowed reaction
func (h *ReactionsHandler) DeleteAllowedReaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]
	reactionID := vars["reactionId"]

	// Verify user owns the site
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	// Verify reaction belongs to site
	allowedReactionStore := models.NewAllowedReactionStore(h.db)
	reaction, err := allowedReactionStore.GetByID(reactionID)
	if err != nil || reaction.SiteID != siteID {
		http.Error(w, "Reaction not found", http.StatusNotFound)
		return
	}

	// Delete reaction
	if err := allowedReactionStore.Delete(reactionID); err != nil {
		log.Printf("Error deleting allowed reaction: %v", err)
		http.Error(w, "Failed to delete reaction", http.StatusInternalServerError)
		return
	}

	// Return success for HTMX
	w.WriteHeader(http.StatusOK)
}

// GetReactionStats returns statistics about reactions for a site (API endpoint for admin)
func (h *ReactionsHandler) GetReactionStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Verify user owns the site
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	// Get reaction statistics
	query := `
		SELECT ar.name, ar.emoji, COUNT(r.id) as count
		FROM allowed_reactions ar
		LEFT JOIN reactions r ON ar.id = r.allowed_reaction_id
		WHERE ar.site_id = ?
		GROUP BY ar.id, ar.name, ar.emoji
		ORDER BY count DESC, ar.name ASC
	`

	rows, err := h.db.QueryContext(r.Context(), query, siteID)
	if err != nil {
		log.Printf("Error querying reaction stats: %v", err)
		http.Error(w, "Failed to get reaction stats", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ReactionStat struct {
		Name  string `json:"name"`
		Emoji string `json:"emoji"`
		Count int    `json:"count"`
	}

	var stats []ReactionStat
	for rows.Next() {
		var stat ReactionStat
		if err := rows.Scan(&stat.Name, &stat.Emoji, &stat.Count); err != nil {
			log.Printf("Error scanning reaction stat: %v", err)
			continue
		}
		stats = append(stats, stat)
	}

	if stats == nil {
		stats = []ReactionStat{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
