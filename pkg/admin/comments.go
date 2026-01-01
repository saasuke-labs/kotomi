package admin

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/comments"
	"github.com/saasuke-labs/kotomi/pkg/models"
)

// CommentsHandler handles comment moderation requests
type CommentsHandler struct {
	db           *sql.DB
	commentStore *comments.SQLiteStore
	templates    *template.Template
}

// NewCommentsHandler creates a new comments handler
func NewCommentsHandler(db *sql.DB, commentStore *comments.SQLiteStore, templates *template.Template) *CommentsHandler {
	return &CommentsHandler{
		db:           db,
		commentStore: commentStore,
		templates:    templates,
	}
}

// ListComments handles GET /admin/sites/{siteId}/comments
func (h *CommentsHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Verify ownership
	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	if site.OwnerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get status filter from query params
	status := r.URL.Query().Get("status")

	comments, err := h.commentStore.GetCommentsBySite(siteID, status)
	if err != nil {
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		if h.templates != nil {
			err = h.templates.ExecuteTemplate(w, "comments/list.html", map[string]interface{}{
				"Comments": comments,
				"SiteID":   siteID,
				"Status":   status,
			})
			if err != nil {
				http.Error(w, "Template error", http.StatusInternalServerError)
			}
		}
		return
	}

	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

// ListPageComments handles GET /admin/sites/{siteId}/pages/{pageId}/comments
func (h *CommentsHandler) ListPageComments(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]
	pageID := vars["pageId"]

	// Verify ownership
	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	comments, err := h.commentStore.GetPageComments(siteID, pageID)
	if err != nil {
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}

	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

// ApproveComment handles POST /admin/comments/{commentId}/approve
func (h *CommentsHandler) ApproveComment(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	commentID := vars["commentId"]

	// Get comment to verify ownership of site
	comment, err := h.commentStore.GetCommentByID(commentID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Get site ID for the comment and verify ownership
	siteID, err := h.commentStore.GetCommentSiteID(commentID)
	if err != nil {
		http.Error(w, "Failed to verify comment ownership", http.StatusInternalServerError)
		return
	}

	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	err = h.commentStore.UpdateCommentStatus(commentID, "approved", userID)
	if err != nil {
		http.Error(w, "Failed to approve comment", http.StatusInternalServerError)
		return
	}

	// For HTMX requests, return updated comment row
	if r.Header.Get("HX-Request") == "true" {
		comment.Status = "approved"
		comment.ModeratedBy = userID
		if h.templates != nil {
			err = h.templates.ExecuteTemplate(w, "comments/row.html", map[string]interface{}{
				"Comment": comment,
			})
			if err != nil {
				http.Error(w, "Template error", http.StatusInternalServerError)
			}
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RejectComment handles POST /admin/comments/{commentId}/reject
func (h *CommentsHandler) RejectComment(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	commentID := vars["commentId"]

	// Get comment to verify ownership
	comment, err := h.commentStore.GetCommentByID(commentID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Get site ID for the comment and verify ownership
	siteID, err := h.commentStore.GetCommentSiteID(commentID)
	if err != nil {
		http.Error(w, "Failed to verify comment ownership", http.StatusInternalServerError)
		return
	}

	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	err = h.commentStore.UpdateCommentStatus(commentID, "rejected", userID)
	if err != nil {
		http.Error(w, "Failed to reject comment", http.StatusInternalServerError)
		return
	}

	// For HTMX requests, return updated comment row
	if r.Header.Get("HX-Request") == "true" {
		comment.Status = "rejected"
		comment.ModeratedBy = userID
		if h.templates != nil {
			err = h.templates.ExecuteTemplate(w, "comments/row.html", map[string]interface{}{
				"Comment": comment,
			})
			if err != nil {
				http.Error(w, "Template error", http.StatusInternalServerError)
			}
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteComment handles DELETE /admin/comments/{commentId}
func (h *CommentsHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	commentID := vars["commentId"]

	// Get comment to verify ownership
	_, err := h.commentStore.GetCommentByID(commentID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Get site ID for the comment and verify ownership
	siteID, err := h.commentStore.GetCommentSiteID(commentID)
	if err != nil {
		http.Error(w, "Failed to verify comment ownership", http.StatusInternalServerError)
		return
	}

	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	err = h.commentStore.DeleteComment(commentID)
	if err != nil {
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	// For HTMX requests, return empty response (row will be removed)
	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
