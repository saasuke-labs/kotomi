package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/comments"
	apierrors "github.com/saasuke-labs/kotomi/pkg/errors"
	"github.com/saasuke-labs/kotomi/pkg/middleware"
	"github.com/saasuke-labs/kotomi/pkg/models"
	"github.com/saasuke-labs/kotomi/pkg/moderation"
	"github.com/saasuke-labs/kotomi/pkg/notifications"
)

// PostComments creates a new comment for a page
// @Summary Create a comment
// @Description Add a new comment to a page (requires JWT authentication)
// @Tags comments
// @Accept json
// @Produce json
// @Param siteId path string true "Site ID"
// @Param pageId path string true "Page ID"
// @Param comment body comments.Comment true "Comment to create"
// @Success 200 {object} comments.Comment
// @Failure 400 {string} string "Invalid JSON or missing required fields"
// @Failure 401 {string} string "Authentication required"
// @Failure 500 {string} string "Failed to add comment"
// @Security BearerAuth
// @Router /site/{siteId}/page/{pageId}/comments [post]
func (s *ServerHandlers) PostComments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := middleware.GetRequestID(r)
	
	// Fallback to manual parsing if vars is empty (e.g., in unit tests)
	if len(vars) == 0 {
		parsedVars, err := GetUrlParams(r)
		if err != nil {
			apierrors.WriteErrorWithRequestID(w, apierrors.BadRequest("Invalid URL"), requestID)
			return
		}
		vars = parsedVars
	}
	
	siteId := vars["siteId"]
	pageId := vars["pageId"]

	// Get authenticated user from context (set by JWT middleware)
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		apierrors.WriteErrorWithRequestID(w, apierrors.Unauthorized("Authentication required"), requestID)
		return
	}

	// Decode body as a Comment
	var comment comments.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		apierrors.WriteErrorWithRequestID(w, apierrors.InvalidJSON("Invalid JSON format").WithDetails(err.Error()), requestID)
		return
	}
	
	// Validate required fields
	if comment.Text == "" {
		apierrors.WriteErrorWithRequestID(w, apierrors.ValidationError("Text is required"), requestID)
		return
	}
	
	// Set user information from authenticated user
	comment.ID = uuid.NewString()
	comment.AuthorID = user.ID
	comment.Author = user.Name
	comment.AuthorEmail = user.Email
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	// Apply AI moderation if enabled
	if s.Moderator != nil && s.ModerationConfigStore != nil {
		config, err := s.ModerationConfigStore.GetBySiteID(r.Context(), siteId)
		if err == nil && config != nil && config.Enabled {
			// Analyze comment with AI moderation
			result, err := s.Moderator.AnalyzeComment(comment.Text, *config)
			if err != nil {
				log.Printf("AI moderation failed: %v", err)
				// Continue with default status on error
			} else {
				// Determine status based on moderation result
				comment.Status = moderation.DetermineStatus(result, *config)
				log.Printf("AI moderation result for comment %s: decision=%s, confidence=%.2f, reason=%s",
					comment.ID, result.Decision, result.Confidence, result.Reason)
			}
		}
	}

	// Set default status if not set by moderation
	if comment.Status == "" {
		comment.Status = "pending"
	}

	if err := s.CommentStore.AddPageComment(r.Context(), siteId, pageId, comment); err != nil {
		log.Printf("Error adding comment: %v", err)
		apierrors.WriteErrorWithRequestID(w, apierrors.DatabaseError("Failed to add comment").WithDetails(err.Error()), requestID)
		return
	}

	// Enqueue notification for new comment (if notifications are enabled)
	if s.NotificationQueue != nil {
		// Get site and page info for notification
		siteStore := models.NewSiteStore(s.DB)
		site, err := siteStore.GetByID(r.Context(), siteId)
		if err == nil && site != nil {
			pageStore := models.NewPageStore(s.DB)
			page, err := pageStore.GetByID(r.Context(), pageId)
			if err == nil && page != nil {
				// Get notification settings
				notifStore := notifications.NewStore(s.DB)
				settings, err := notifStore.GetSettings(siteId)
				if err == nil && settings != nil && settings.Enabled && settings.NotifyNewComment {
					// Build comment URL (placeholder - should be configured per site)
					commentURL := fmt.Sprintf("%s?comment=%s", page.Path, comment.ID)
					unsubscribeURL := fmt.Sprintf("/unsubscribe?site=%s", siteId)
					
					// Enqueue notification
					err = s.NotificationQueue.EnqueueNewComment(
						siteId,
						site.Name,
						page.Title,
						commentURL,
						comment.Author,
						comment.Text,
						settings.OwnerEmail,
						unsubscribeURL,
					)
					if err != nil {
						log.Printf("Warning: Failed to enqueue notification: %v", err)
					} else {
						log.Printf("Enqueued new comment notification for site %s", siteId)
					}
				}
			}
		}
	}

	WriteJsonResponse(w, comment)
}

// GetComments retrieves all comments for a page
// @Summary Get comments for a page
// @Description Retrieve all comments for a specific page
// @Tags comments
// @Produce json
// @Param siteId path string true "Site ID"
// @Param pageId path string true "Page ID"
// @Success 200 {array} comments.Comment
// @Failure 400 {string} string "Invalid URL"
// @Failure 500 {string} string "Failed to retrieve comments"
// @Router /site/{siteId}/page/{pageId}/comments [get]
func (s *ServerHandlers) GetComments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := middleware.GetRequestID(r)
	
	// Fallback to manual parsing if vars is empty (e.g., in unit tests)
	if len(vars) == 0 {
		parsedVars, err := GetUrlParams(r)
		if err != nil {
			apierrors.WriteErrorWithRequestID(w, apierrors.BadRequest("Invalid URL"), requestID)
			return
		}
		vars = parsedVars
	}
	
	siteId := vars["siteId"]
	pageId := vars["pageId"]
	
	commentsData, err := s.CommentStore.GetPageComments(r.Context(), siteId, pageId)
	if err != nil {
		log.Printf("Error retrieving comments: %v", err)
		apierrors.WriteErrorWithRequestID(w, apierrors.DatabaseError("Failed to retrieve comments").WithDetails(err.Error()), requestID)
		return
	}

	WriteJsonResponse(w, commentsData)
}

// UpdateComment updates a comment's text (owner only)
// @Summary Update a comment
// @Description Update the text of an existing comment (requires JWT authentication and ownership)
// @Tags comments
// @Accept json
// @Produce json
// @Param siteId path string true "Site ID"
// @Param commentId path string true "Comment ID"
// @Param update body object{text=string} true "Updated comment text"
// @Success 200 {object} comments.Comment
// @Failure 400 {string} string "Invalid JSON or missing required fields"
// @Failure 401 {string} string "Authentication required"
// @Failure 403 {string} string "Forbidden - not the comment owner"
// @Failure 404 {string} string "Comment not found"
// @Failure 500 {string} string "Failed to update comment"
// @Security BearerAuth
// @Router /site/{siteId}/comments/{commentId} [put]
func (s *ServerHandlers) UpdateComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["commentId"]
	siteID := vars["siteId"]

	// Get authenticated user from context
	user, ok := r.Context().Value(middleware.ContextKeyUser).(*models.KotomiUser)
	if !ok || user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var updateReq struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if updateReq.Text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}

	// Get the comment to verify ownership
	comment, err := s.CommentStore.GetCommentByID(r.Context(), commentID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Verify the comment belongs to this site
	if comment.SiteID != siteID {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Verify ownership - user can only edit their own comments
	if comment.AuthorID != user.ID {
		http.Error(w, "Forbidden - you can only edit your own comments", http.StatusForbidden)
		return
	}

	// Update the comment text
	if err := s.CommentStore.UpdateCommentText(r.Context(), commentID, updateReq.Text); err != nil {
		log.Printf("Error updating comment: %v", err)
		http.Error(w, "Failed to update comment", http.StatusInternalServerError)
		return
	}

	// Retrieve and return the updated comment
	updatedComment, err := s.CommentStore.GetCommentByID(r.Context(), commentID)
	if err != nil {
		http.Error(w, "Failed to retrieve updated comment", http.StatusInternalServerError)
		return
	}

	WriteJsonResponse(w, updatedComment)
}

// DeleteComment deletes a comment (owner only)
// @Summary Delete a comment
// @Description Delete an existing comment (requires JWT authentication and ownership)
// @Tags comments
// @Param siteId path string true "Site ID"
// @Param commentId path string true "Comment ID"
// @Success 204 "Comment deleted successfully"
// @Failure 401 {string} string "Authentication required"
// @Failure 403 {string} string "Forbidden - not the comment owner"
// @Failure 404 {string} string "Comment not found"
// @Failure 500 {string} string "Failed to delete comment"
// @Security BearerAuth
// @Router /site/{siteId}/comments/{commentId} [delete]
func (s *ServerHandlers) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["commentId"]
	siteID := vars["siteId"]

	// Get authenticated user from context
	user, ok := r.Context().Value(middleware.ContextKeyUser).(*models.KotomiUser)
	if !ok || user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Get the comment to verify ownership
	comment, err := s.CommentStore.GetCommentByID(r.Context(), commentID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Verify the comment belongs to this site
	if comment.SiteID != siteID {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Verify ownership - user can only delete their own comments
	if comment.AuthorID != user.ID {
		http.Error(w, "Forbidden - you can only delete your own comments", http.StatusForbidden)
		return
	}

	// Delete the comment
	if err := s.CommentStore.DeleteComment(r.Context(), commentID); err != nil {
		log.Printf("Error deleting comment: %v", err)
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
