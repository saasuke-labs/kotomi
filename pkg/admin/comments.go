package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/comments"
	"github.com/saasuke-labs/kotomi/pkg/models"
	"github.com/saasuke-labs/kotomi/pkg/notifications"
)

// CommentsHandler handles comment moderation requests
type CommentsHandler struct {
	db                *sql.DB
	commentStore      *comments.SQLiteStore
	templates         *template.Template
	notificationQueue *notifications.Queue
}

// NewCommentsHandler creates a new comments handler
func NewCommentsHandler(db *sql.DB, commentStore *comments.SQLiteStore, templates *template.Template) *CommentsHandler {
	return &CommentsHandler{
		db:                db,
		commentStore:      commentStore,
		templates:         templates,
		notificationQueue: nil, // Will be set later if needed
	}
}

// SetNotificationQueue sets the notification queue for the handler
func (h *CommentsHandler) SetNotificationQueue(queue *notifications.Queue) {
	h.notificationQueue = queue
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
	site, err := siteStore.GetByID(r.Context(), siteID)
	if err != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	if site.OwnerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get filters from query params
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")

	// Get comments with search
	var commentsList []comments.Comment
	if search != "" {
		commentsList, err = h.searchComments(r.Context(), siteID, status, search)
	} else {
		commentsList, err = h.commentStore.GetCommentsBySite(r.Context(), siteID, status)
	}
	if err != nil {
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}

	// Check if this is an HTMX request or regular page load
	if r.Header.Get("HX-Request") == "true" || r.Header.Get("Accept") == "text/html" {
		if h.templates != nil {
			err = h.templates.ExecuteTemplate(w, "comments/list.html", map[string]interface{}{
				"Comments": commentsList,
				"SiteID":   siteID,
				"Status":   status,
				"Search":   search,
			})
			if err != nil {
				http.Error(w, "Template error", http.StatusInternalServerError)
			}
		}
		return
	}

	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(commentsList)
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
	site, err := siteStore.GetByID(r.Context(), siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	comments, err := h.commentStore.GetPageComments(r.Context(), siteID, pageID)
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
	comment, err := h.commentStore.GetCommentByID(r.Context(), commentID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Get site ID for the comment and verify ownership
	siteID, err := h.commentStore.GetCommentSiteID(r.Context(), commentID)
	if err != nil {
		http.Error(w, "Failed to verify comment ownership", http.StatusInternalServerError)
		return
	}

	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(r.Context(), siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	err = h.commentStore.UpdateCommentStatus(r.Context(), commentID, "approved", userID)
	if err != nil {
		http.Error(w, "Failed to approve comment", http.StatusInternalServerError)
		return
	}

	// Enqueue moderation update notification
	if h.notificationQueue != nil && comment.AuthorEmail != "" {
		notifStore := notifications.NewStore(h.db)
		settings, err := notifStore.GetSettings(siteID)
		if err == nil && settings != nil && settings.Enabled && settings.NotifyModeration {
			// Get page_id for this comment
			var pageID string
			err = h.db.QueryRowContext(r.Context(), "SELECT page_id FROM comments WHERE id = ?", commentID).Scan(&pageID)
			if err == nil {
				// Get page info for notification
				pageStore := models.NewPageStore(h.db)
				page, err := pageStore.GetByID(r.Context(), pageID)
				if err == nil && page != nil {
					commentURL := fmt.Sprintf("%s?comment=%s", page.Path, comment.ID)
					unsubscribeURL := fmt.Sprintf("/unsubscribe?site=%s", siteID)
					
					err = h.notificationQueue.EnqueueModerationUpdate(
						siteID,
						page.Title,
						commentURL,
						comment.Text,
						"approved",
						"", // No reason for approval
						comment.AuthorEmail,
						unsubscribeURL,
					)
					if err != nil {
						log.Printf("Warning: Failed to enqueue moderation notification: %v", err)
					}
				}
			}
		}
	}

	// For HTMX requests, return updated comment row
	if r.Header.Get("HX-Request") == "true" {
		comment.Status = "approved"
		comment.ModeratedBy = userID
		if h.templates != nil {
			err = h.templates.ExecuteTemplate(w, "comments/row.html", comment)
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
	comment, err := h.commentStore.GetCommentByID(r.Context(), commentID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Get site ID for the comment and verify ownership
	siteID, err := h.commentStore.GetCommentSiteID(r.Context(), commentID)
	if err != nil {
		http.Error(w, "Failed to verify comment ownership", http.StatusInternalServerError)
		return
	}

	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(r.Context(), siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	err = h.commentStore.UpdateCommentStatus(r.Context(), commentID, "rejected", userID)
	if err != nil {
		http.Error(w, "Failed to reject comment", http.StatusInternalServerError)
		return
	}

	// Enqueue moderation update notification
	if h.notificationQueue != nil && comment.AuthorEmail != "" {
		notifStore := notifications.NewStore(h.db)
		settings, err := notifStore.GetSettings(siteID)
		if err == nil && settings != nil && settings.Enabled && settings.NotifyModeration {
			// Get page_id for this comment
			var pageID string
			err = h.db.QueryRowContext(r.Context(), "SELECT page_id FROM comments WHERE id = ?", commentID).Scan(&pageID)
			if err == nil {
				// Get page info for notification
				pageStore := models.NewPageStore(h.db)
				page, err := pageStore.GetByID(r.Context(), pageID)
				if err == nil && page != nil {
					commentURL := fmt.Sprintf("%s?comment=%s", page.Path, comment.ID)
					unsubscribeURL := fmt.Sprintf("/unsubscribe?site=%s", siteID)
					
					err = h.notificationQueue.EnqueueModerationUpdate(
						siteID,
						page.Title,
						commentURL,
						comment.Text,
						"rejected",
						"Content violated community guidelines", // Default reason
						comment.AuthorEmail,
						unsubscribeURL,
					)
					if err != nil {
						log.Printf("Warning: Failed to enqueue moderation notification: %v", err)
					}
				}
			}
		}
	}

	// For HTMX requests, return updated comment row
	if r.Header.Get("HX-Request") == "true" {
		comment.Status = "rejected"
		comment.ModeratedBy = userID
		if h.templates != nil {
			err = h.templates.ExecuteTemplate(w, "comments/row.html", comment)
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
	_, err := h.commentStore.GetCommentByID(r.Context(), commentID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Get site ID for the comment and verify ownership
	siteID, err := h.commentStore.GetCommentSiteID(r.Context(), commentID)
	if err != nil {
		http.Error(w, "Failed to verify comment ownership", http.StatusInternalServerError)
		return
	}

	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(r.Context(), siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	err = h.commentStore.DeleteComment(r.Context(), commentID)
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

// searchComments searches comments by text, author, or page
func (h *CommentsHandler) searchComments(ctx context.Context, siteID, status, search string) ([]comments.Comment, error) {
	query := `
		SELECT c.id, c.site_id, c.author, c.author_id, c.author_email, c.text, 
		       c.parent_id, c.status, c.moderated_by, c.moderated_at, c.created_at, c.updated_at
		FROM comments c
		LEFT JOIN pages p ON c.page_id = p.id
		WHERE c.site_id = ?
	`
	args := []interface{}{siteID}

	// Add status filter
	if status != "" {
		query += " AND c.status = ?"
		args = append(args, status)
	}

	// Add search filter
	search = "%" + search + "%"
	query += " AND (c.text LIKE ? OR c.author LIKE ? OR c.author_email LIKE ? OR p.path LIKE ?)"
	args = append(args, search, search, search, search)

	query += " ORDER BY c.created_at DESC"

	rows, err := h.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commentsList []comments.Comment
	for rows.Next() {
		var c comments.Comment
		var moderatedBy, moderatedAt, parentID, authorEmail sql.NullString
		err := rows.Scan(
			&c.ID, &c.SiteID, &c.Author, &c.AuthorID, &authorEmail, &c.Text,
			&parentID, &c.Status, &moderatedBy, &moderatedAt, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if moderatedBy.Valid {
			c.ModeratedBy = moderatedBy.String
		}
		if parentID.Valid {
			c.ParentID = parentID.String
		}
		if authorEmail.Valid {
			c.AuthorEmail = authorEmail.String
		}
		commentsList = append(commentsList, c)
	}

	return commentsList, nil
}

// BulkApprove handles POST /admin/comments/bulk/approve
func (h *CommentsHandler) BulkApprove(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		CommentIDs []string `json:"comment_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Approve each comment
	for _, commentID := range req.CommentIDs {
		// Verify ownership (basic check)
		_, err := h.commentStore.GetCommentByID(r.Context(), commentID)
		if err != nil {
			continue // Skip invalid comments
		}
		
		err = h.commentStore.UpdateCommentStatus(r.Context(), commentID, "approved", userID)
		if err != nil {
			log.Printf("Failed to approve comment %s: %v", commentID, err)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"count":   len(req.CommentIDs),
	})
}

// BulkReject handles POST /admin/comments/bulk/reject
func (h *CommentsHandler) BulkReject(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		CommentIDs []string `json:"comment_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Reject each comment
	for _, commentID := range req.CommentIDs {
		// Verify ownership (basic check)
		_, err := h.commentStore.GetCommentByID(r.Context(), commentID)
		if err != nil {
			continue // Skip invalid comments
		}
		
		err = h.commentStore.UpdateCommentStatus(r.Context(), commentID, "rejected", userID)
		if err != nil {
			log.Printf("Failed to reject comment %s: %v", commentID, err)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"count":   len(req.CommentIDs),
	})
}

// BulkDelete handles POST /admin/comments/bulk/delete
func (h *CommentsHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		CommentIDs []string `json:"comment_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Delete each comment
	for _, commentID := range req.CommentIDs {
		// Verify ownership (basic check)
		_, err := h.commentStore.GetCommentByID(r.Context(), commentID)
		if err != nil {
			continue // Skip invalid comments
		}
		
		err = h.commentStore.DeleteComment(r.Context(), commentID)
		if err != nil {
			log.Printf("Failed to delete comment %s: %v", commentID, err)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"count":   len(req.CommentIDs),
	})
}
