package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/comments"
	apierrors "github.com/saasuke-labs/kotomi/pkg/errors"
	"github.com/saasuke-labs/kotomi/pkg/logging"
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
	ctx := r.Context()
	
	// Fallback to manual parsing if vars is empty (e.g., in unit tests)
	if len(vars) == 0 {
		parsedVars, err := GetUrlParams(r)
		if err != nil {
			apierrors.WriteErrorWithRequestID(w, apierrors.BadRequest("Invalid URL"), logging.GetRequestID(ctx))
			return
		}
		vars = parsedVars
	}
	
	siteId := vars["siteId"]
	pageId := vars["pageId"]

	// Enrich context with site_id and page_id for automatic logging
	ctx = logging.WithSiteID(ctx, siteId)
	ctx = logging.WithPageID(ctx, pageId)
	r = r.WithContext(ctx)

	// Get authenticated user from context (set by JWT middleware)
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		apierrors.WriteErrorWithRequestID(w, apierrors.Unauthorized("Authentication required"), logging.GetRequestID(ctx))
		return
	}

	// Decode body as a Comment
	var comment comments.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		apierrors.WriteErrorWithRequestID(w, apierrors.InvalidJSON("Invalid JSON format").WithDetails(err.Error()), logging.GetRequestID(ctx))
		return
	}
	
	// Validate required fields
	if comment.Text == "" {
		apierrors.WriteErrorWithRequestID(w, apierrors.ValidationError("Text is required"), logging.GetRequestID(ctx))
		return
	}
	
	// Set user information from authenticated user
	comment.ID = uuid.NewString()
	comment.AuthorID = user.ID
	comment.Author = user.Name
	comment.AuthorEmail = user.Email
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	// Enrich context with comment_id for logging
	ctx = logging.WithCommentID(ctx, comment.ID)
	r = r.WithContext(ctx)

	// Apply AI moderation if enabled
	if s.Moderator != nil && s.ModerationConfigStore != nil {
		config, err := s.ModerationConfigStore.GetBySiteID(ctx, siteId)
		if err == nil && config != nil && config.Enabled {
			// Analyze comment with AI moderation
			result, err := s.Moderator.AnalyzeComment(comment.Text, *config)
			if err != nil {
				s.Logger.ErrorContext(ctx, "AI moderation failed", "error", err)
				// Continue with default status on error
			} else {
				// Determine status based on moderation result
				comment.Status = moderation.DetermineStatus(result, *config)
				s.Logger.InfoContext(ctx, "AI moderation completed",
					"decision", result.Decision,
					"confidence", result.Confidence,
					"reason", result.Reason)
			}
		}
	}

	// Set default status if not set by moderation
	if comment.Status == "" {
		comment.Status = "pending"
	}

	if err := s.CommentStore.AddPageComment(ctx, siteId, pageId, comment); err != nil {
		s.Logger.ErrorContext(ctx, "failed to add comment", "error", err)
		apierrors.WriteErrorWithRequestID(w, apierrors.DatabaseError("Failed to add comment").WithDetails(err.Error()), logging.GetRequestID(ctx))
		return
	}

	// Enqueue notification for new comment (if notifications are enabled)
	if s.NotificationQueue != nil {
		// Get site and page info for notification
		siteStore := models.NewSiteStore(s.DB)
		site, err := siteStore.GetByID(ctx, siteId)
		if err == nil && site != nil {
			pageStore := models.NewPageStore(s.DB)
			page, err := pageStore.GetByID(ctx, pageId)
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
						s.Logger.WarnContext(ctx, "failed to enqueue notification", "error", err)
					} else {
						s.Logger.InfoContext(ctx, "enqueued new comment notification")
					}
				}
			}
		}
	}

	s.WriteJsonResponse(w, comment)
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
	ctx := r.Context()
	
	// Fallback to manual parsing if vars is empty (e.g., in unit tests)
	if len(vars) == 0 {
		parsedVars, err := GetUrlParams(r)
		if err != nil {
			apierrors.WriteErrorWithRequestID(w, apierrors.BadRequest("Invalid URL"), logging.GetRequestID(ctx))
			return
		}
		vars = parsedVars
	}
	
	siteId := vars["siteId"]
	pageId := vars["pageId"]

	// Enrich context with site_id and page_id for automatic logging
	ctx = logging.WithSiteID(ctx, siteId)
	ctx = logging.WithPageID(ctx, pageId)
	
	commentsData, err := s.CommentStore.GetPageComments(ctx, siteId, pageId)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to retrieve comments", "error", err)
		apierrors.WriteErrorWithRequestID(w, apierrors.DatabaseError("Failed to retrieve comments").WithDetails(err.Error()), logging.GetRequestID(ctx))
		return
	}

	s.WriteJsonResponse(w, commentsData)
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

	// Enrich context with site_id and comment_id for automatic logging
	ctx := r.Context()
	ctx = logging.WithSiteID(ctx, siteID)
	ctx = logging.WithCommentID(ctx, commentID)

	// Get authenticated user from context
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		apierrors.WriteError(w, apierrors.Unauthorized("Authentication required").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	// Parse request body
	var updateReq struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		apierrors.WriteError(w, apierrors.InvalidJSON("Invalid request body").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	if updateReq.Text == "" {
		apierrors.WriteError(w, apierrors.ValidationError("Text is required").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	// Get the comment to verify ownership
	comment, err := s.CommentStore.GetCommentByID(ctx, commentID)
	if err != nil {
		apierrors.WriteError(w, apierrors.NotFound("Comment not found").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	// Verify the comment belongs to this site
	if comment.SiteID != siteID {
		apierrors.WriteError(w, apierrors.NotFound("Comment not found").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	// Verify ownership - user can only edit their own comments
	if comment.AuthorID != user.ID {
		apierrors.WriteError(w, apierrors.Forbidden("Forbidden - you can only edit your own comments").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	// Update the comment text
	if err := s.CommentStore.UpdateCommentText(ctx, commentID, updateReq.Text); err != nil {
		s.Logger.ErrorContext(ctx, "failed to update comment", "error", err)
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to update comment").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	// Retrieve and return the updated comment
	updatedComment, err := s.CommentStore.GetCommentByID(ctx, commentID)
	if err != nil {
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to retrieve updated comment").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	s.WriteJsonResponse(w, updatedComment)
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

	// Enrich context with site_id and comment_id for automatic logging
	ctx := r.Context()
	ctx = logging.WithSiteID(ctx, siteID)
	ctx = logging.WithCommentID(ctx, commentID)

	// Get authenticated user from context
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		apierrors.WriteError(w, apierrors.Unauthorized("Authentication required").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	// Get the comment to verify ownership
	comment, err := s.CommentStore.GetCommentByID(ctx, commentID)
	if err != nil {
		apierrors.WriteError(w, apierrors.NotFound("Comment not found").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	// Verify the comment belongs to this site
	if comment.SiteID != siteID {
		apierrors.WriteError(w, apierrors.NotFound("Comment not found").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	// Verify ownership - user can only delete their own comments
	if comment.AuthorID != user.ID {
		apierrors.WriteError(w, apierrors.Forbidden("Forbidden - you can only delete your own comments").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	// Delete the comment
	if err := s.CommentStore.DeleteComment(ctx, commentID); err != nil {
		s.Logger.ErrorContext(ctx, "failed to delete comment", "error", err)
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to delete comment").WithRequestID(logging.GetRequestID(ctx)))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
