/**
 * Kotomi - Comment and Reaction Widget
 * Version: 0.1.0
 * 
 * A simple, embeddable comment widget for static websites
 * 
 * Usage:
 *   const kotomi = new Kotomi({
 *     baseUrl: 'https://your-kotomi-server.com',
 *     siteId: 'your-site-id',
 *     pageId: 'page-slug',
 *     container: '#kotomi-comments',
 *     jwtToken: 'optional-jwt-token'
 *   });
 *   kotomi.render();
 */

(function(window) {
  'use strict';

/**
 * Kotomi API Client
 * Handles communication with the Kotomi comment server
 */

class KotomiAPI {
  constructor(config) {
    this.baseUrl = config.baseUrl || '';
    this.siteId = config.siteId;
    this.pageId = config.pageId;
    this.jwtToken = config.jwtToken || null;
    this.apiVersion = 'v1';
  }

  /**
   * Make an API request
   */
  async request(endpoint, options = {}) {
    const url = `${this.baseUrl}/api/${this.apiVersion}${endpoint}`;
    const headers = {
      'Content-Type': 'application/json',
      ...options.headers
    };

    // Add JWT token if available
    if (this.jwtToken) {
      headers['Authorization'] = `Bearer ${this.jwtToken}`;
    }

    const response = await fetch(url, {
      ...options,
      headers
    });

    if (!response.ok) {
      const error = new Error(`API request failed: ${response.statusText}`);
      error.status = response.status;
      try {
        error.data = await response.json();
      } catch (e) {
        // Response might not be JSON
      }
      throw error;
    }

    // Handle 204 No Content
    if (response.status === 204) {
      return null;
    }

    return response.json();
  }

  /**
   * Get all comments for the current page
   */
  async getComments() {
    return this.request(`/site/${this.siteId}/page/${this.pageId}/comments`);
  }

  /**
   * Post a new comment
   */
  async postComment(text, parentId = null) {
    const body = { text };
    if (parentId) {
      body.parent_id = parentId;
    }

    return this.request(`/site/${this.siteId}/page/${this.pageId}/comments`, {
      method: 'POST',
      body: JSON.stringify(body)
    });
  }

  /**
   * Update a comment
   */
  async updateComment(commentId, text) {
    return this.request(`/comments/${commentId}`, {
      method: 'PUT',
      body: JSON.stringify({ text })
    });
  }

  /**
   * Delete a comment
   */
  async deleteComment(commentId) {
    return this.request(`/comments/${commentId}`, {
      method: 'DELETE'
    });
  }

  /**
   * Get allowed reactions for the site
   */
  async getAllowedReactions(type = null) {
    const typeParam = type ? `?type=${type}` : '';
    return this.request(`/site/${this.siteId}/allowed-reactions${typeParam}`);
  }

  /**
   * Get reaction counts for a comment
   */
  async getCommentReactionCounts(commentId) {
    return this.request(`/comments/${commentId}/reactions/counts`);
  }

  /**
   * Get reaction counts for the current page
   */
  async getPageReactionCounts() {
    return this.request(`/pages/${this.pageId}/reactions/counts`);
  }

  /**
   * Add/remove a reaction to a comment (toggle)
   */
  async toggleCommentReaction(commentId, allowedReactionId) {
    return this.request(`/comments/${commentId}/reactions`, {
      method: 'POST',
      body: JSON.stringify({ allowed_reaction_id: allowedReactionId })
    });
  }

  /**
   * Add/remove a reaction to the page (toggle)
   */
  async togglePageReaction(allowedReactionId) {
    return this.request(`/pages/${this.pageId}/reactions`, {
      method: 'POST',
      body: JSON.stringify({ allowed_reaction_id: allowedReactionId })
    });
  }
}


/**
 * Kotomi UI Module
 * Handles rendering of comments and reactions
 */

class KotomiUI {
  constructor(api, container, options = {}) {
    this.api = api;
    this.container = typeof container === 'string' 
      ? document.querySelector(container) 
      : container;
    this.options = {
      theme: options.theme || 'light', // 'light' or 'dark'
      enableReactions: options.enableReactions !== false,
      enableReplies: options.enableReplies !== false,
      placeholder: options.placeholder || 'Write a comment...',
      submitText: options.submitText || 'Post Comment',
      ...options
    };
    
    this.comments = [];
    this.allowedReactions = [];
    this.reactionCounts = {};
    this.isAuthenticated = !!api.jwtToken;
  }

  /**
   * Initialize the widget
   */
  async init() {
    if (!this.container) {
      throw new Error('Kotomi: Container element not found');
    }

    // Add theme class
    this.container.classList.add('kotomi-widget');
    this.container.classList.add(`kotomi-theme-${this.options.theme}`);

    // Load data
    await this.loadData();

    // Render the widget
    this.render();
  }

  /**
   * Load comments and reactions from API
   */
  async loadData() {
    try {
      // Load comments
      this.comments = await this.api.getComments();

      // Load allowed reactions if enabled
      if (this.options.enableReactions) {
        this.allowedReactions = await this.api.getAllowedReactions('comment');
        
        // Load reaction counts for each comment
        for (const comment of this.comments) {
          try {
            this.reactionCounts[comment.id] = await this.api.getCommentReactionCounts(comment.id);
          } catch (e) {
            // Ignore if reactions not available for this comment
            this.reactionCounts[comment.id] = [];
          }
        }
      }
    } catch (error) {
      console.error('Failed to load Kotomi data:', error);
      this.showError('Failed to load comments. Please try refreshing the page.');
    }
  }

  /**
   * Render the entire widget
   */
  render() {
    this.container.innerHTML = `
      <div class="kotomi-container">
        ${this.renderCommentForm()}
        ${this.renderCommentsList()}
      </div>
    `;

    // Attach event listeners
    this.attachEventListeners();
  }

  /**
   * Render the comment submission form
   */
  renderCommentForm(parentId = null) {
    if (!this.isAuthenticated) {
      return `
        <div class="kotomi-auth-notice">
          <p>Please sign in to post a comment.</p>
        </div>
      `;
    }

    const formId = parentId ? `reply-form-${parentId}` : 'main-comment-form';
    return `
      <div class="kotomi-comment-form" id="${formId}">
        <textarea 
          class="kotomi-textarea" 
          placeholder="${this.options.placeholder}"
          data-parent-id="${parentId || ''}"
          rows="3"
        ></textarea>
        <div class="kotomi-form-actions">
          <button class="kotomi-btn kotomi-btn-primary" data-action="submit" data-parent-id="${parentId || ''}">
            ${this.options.submitText}
          </button>
          ${parentId ? '<button class="kotomi-btn kotomi-btn-secondary" data-action="cancel-reply">Cancel</button>' : ''}
        </div>
      </div>
    `;
  }

  /**
   * Render the comments list
   */
  renderCommentsList() {
    if (this.comments.length === 0) {
      return `
        <div class="kotomi-empty">
          <p>No comments yet. Be the first to comment!</p>
        </div>
      `;
    }

    // Organize comments by parent (for threading)
    const topLevelComments = this.comments.filter(c => !c.parent_id);
    
    return `
      <div class="kotomi-comments-list">
        <h3 class="kotomi-comments-title">Comments (${this.comments.length})</h3>
        ${topLevelComments.map(comment => this.renderComment(comment)).join('')}
      </div>
    `;
  }

  /**
   * Render a single comment with its replies
   */
  renderComment(comment, isReply = false) {
    const replies = this.comments.filter(c => c.parent_id === comment.id);
    const reactions = this.reactionCounts[comment.id] || [];
    const date = new Date(comment.created_at).toLocaleString();

    return `
      <div class="kotomi-comment ${isReply ? 'kotomi-reply' : ''}" data-comment-id="${comment.id}">
        <div class="kotomi-comment-header">
          <span class="kotomi-comment-author">${this.escapeHtml(comment.author || 'Anonymous')}</span>
          <span class="kotomi-comment-date">${date}</span>
        </div>
        <div class="kotomi-comment-body">
          ${this.escapeHtml(comment.text)}
        </div>
        <div class="kotomi-comment-actions">
          ${this.options.enableReactions ? this.renderReactions(comment.id, reactions) : ''}
          ${this.options.enableReplies && !isReply && this.isAuthenticated ? `
            <button class="kotomi-btn-link" data-action="reply" data-comment-id="${comment.id}">
              Reply
            </button>
          ` : ''}
        </div>
        <div class="kotomi-reply-form-container" id="reply-container-${comment.id}"></div>
        ${replies.length > 0 ? `
          <div class="kotomi-replies">
            ${replies.map(reply => this.renderComment(reply, true)).join('')}
          </div>
        ` : ''}
      </div>
    `;
  }

  /**
   * Render reactions for a comment
   */
  renderReactions(commentId, reactionCounts) {
    if (!this.allowedReactions || this.allowedReactions.length === 0) {
      return '';
    }

    return `
      <div class="kotomi-reactions" data-comment-id="${commentId}">
        ${this.allowedReactions.map(reaction => {
          const count = reactionCounts.find(r => r.name === reaction.name);
          const countValue = count ? count.count : 0;
          return `
            <button 
              class="kotomi-reaction-btn ${countValue > 0 ? 'kotomi-reaction-active' : ''}"
              data-action="react"
              data-comment-id="${commentId}"
              data-reaction-id="${reaction.id}"
              ${!this.isAuthenticated ? 'disabled' : ''}
            >
              <span class="kotomi-reaction-emoji">${reaction.emoji}</span>
              ${countValue > 0 ? `<span class="kotomi-reaction-count">${countValue}</span>` : ''}
            </button>
          `;
        }).join('')}
      </div>
    `;
  }

  /**
   * Attach event listeners
   */
  attachEventListeners() {
    // Submit comment
    this.container.querySelectorAll('[data-action="submit"]').forEach(btn => {
      btn.addEventListener('click', (e) => this.handleSubmitComment(e));
    });

    // Reply button
    this.container.querySelectorAll('[data-action="reply"]').forEach(btn => {
      btn.addEventListener('click', (e) => this.handleReplyClick(e));
    });

    // Cancel reply
    this.container.querySelectorAll('[data-action="cancel-reply"]').forEach(btn => {
      btn.addEventListener('click', (e) => this.handleCancelReply(e));
    });

    // React button
    this.container.querySelectorAll('[data-action="react"]').forEach(btn => {
      btn.addEventListener('click', (e) => this.handleReactionClick(e));
    });
  }

  /**
   * Handle comment submission
   */
  async handleSubmitComment(event) {
    const button = event.target;
    const parentId = button.dataset.parentId || null;
    const form = button.closest('.kotomi-comment-form');
    const textarea = form.querySelector('textarea');
    const text = textarea.value.trim();

    if (!text) {
      this.showError('Please enter a comment');
      return;
    }

    try {
      button.disabled = true;
      button.textContent = 'Posting...';

      await this.api.postComment(text, parentId);
      
      // Reload and re-render
      await this.loadData();
      this.render();
      
      this.showSuccess('Comment posted successfully!');
    } catch (error) {
      console.error('Failed to post comment:', error);
      this.showError('Failed to post comment. Please try again.');
      button.disabled = false;
      button.textContent = this.options.submitText;
    }
  }

  /**
   * Handle reply button click
   */
  handleReplyClick(event) {
    const button = event.target;
    const commentId = button.dataset.commentId;
    const container = document.getElementById(`reply-container-${commentId}`);
    
    if (container) {
      container.innerHTML = this.renderCommentForm(commentId);
      this.attachEventListeners();
    }
  }

  /**
   * Handle cancel reply
   */
  handleCancelReply(event) {
    const button = event.target;
    const form = button.closest('.kotomi-comment-form');
    form.remove();
  }

  /**
   * Handle reaction click
   */
  async handleReactionClick(event) {
    const button = event.target.closest('[data-action="react"]');
    const commentId = button.dataset.commentId;
    const reactionId = button.dataset.reactionId;

    try {
      button.disabled = true;
      
      await this.api.toggleCommentReaction(commentId, reactionId);
      
      // Reload reaction counts
      this.reactionCounts[commentId] = await this.api.getCommentReactionCounts(commentId);
      
      // Re-render just this comment's reactions
      const commentElement = this.container.querySelector(`[data-comment-id="${commentId}"]`);
      const reactionsContainer = commentElement.querySelector('.kotomi-reactions');
      if (reactionsContainer) {
        reactionsContainer.outerHTML = this.renderReactions(commentId, this.reactionCounts[commentId]);
        this.attachEventListeners();
      }
    } catch (error) {
      console.error('Failed to toggle reaction:', error);
      this.showError('Failed to update reaction. Please try again.');
    } finally {
      button.disabled = false;
    }
  }

  /**
   * Show error message
   */
  showError(message) {
    this.showMessage(message, 'error');
  }

  /**
   * Show success message
   */
  showSuccess(message) {
    this.showMessage(message, 'success');
  }

  /**
   * Show a message to the user
   */
  showMessage(message, type = 'info') {
    const existingMessage = this.container.querySelector('.kotomi-message');
    if (existingMessage) {
      existingMessage.remove();
    }

    const messageEl = document.createElement('div');
    messageEl.className = `kotomi-message kotomi-message-${type}`;
    messageEl.textContent = message;
    
    this.container.insertBefore(messageEl, this.container.firstChild);
    
    // Auto-remove after 5 seconds
    setTimeout(() => {
      messageEl.remove();
    }, 5000);
  }

  /**
   * Escape HTML to prevent XSS
   */
  escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  /**
   * Refresh the widget
   */
  async refresh() {
    await this.loadData();
    this.render();
  }
}



  /**
   * Main Kotomi class
   */
  class Kotomi {
    constructor(config) {
      // Validate required config
      if (!config.siteId) {
        throw new Error('Kotomi: siteId is required');
      }
      if (!config.pageId) {
        throw new Error('Kotomi: pageId is required');
      }

      this.config = {
        baseUrl: '',
        container: '#kotomi-comments',
        theme: 'light',
        enableReactions: true,
        enableReplies: true,
        jwtToken: null,
        ...config
      };

      // Initialize API client
      this.api = new KotomiAPI({
        baseUrl: this.config.baseUrl,
        siteId: this.config.siteId,
        pageId: this.config.pageId,
        jwtToken: this.config.jwtToken
      });

      // UI will be initialized when render is called
      this.ui = null;
    }

    /**
     * Render the widget
     */
    async render() {
      try {
        this.ui = new KotomiUI(this.api, this.config.container, {
          theme: this.config.theme,
          enableReactions: this.config.enableReactions,
          enableReplies: this.config.enableReplies,
          placeholder: this.config.placeholder,
          submitText: this.config.submitText
        });

        await this.ui.init();
      } catch (error) {
        console.error('Kotomi: Failed to render widget:', error);
        throw error;
      }
    }

    /**
     * Set JWT token for authentication
     */
    setToken(token) {
      this.config.jwtToken = token;
      this.api.jwtToken = token;
      if (this.ui) {
        this.ui.isAuthenticated = !!token;
        this.ui.render();
      }
    }

    /**
     * Refresh the widget (reload comments and reactions)
     */
    async refresh() {
      if (this.ui) {
        await this.ui.refresh();
      }
    }

    /**
     * Destroy the widget
     */
    destroy() {
      if (this.ui && this.ui.container) {
        this.ui.container.innerHTML = '';
        this.ui.container.classList.remove('kotomi-widget');
        this.ui.container.classList.remove(`kotomi-theme-${this.config.theme}`);
      }
      this.ui = null;
    }
  }

  // Export to window
  window.Kotomi = Kotomi;

  // Also export KotomiAPI for advanced usage
  window.KotomiAPI = KotomiAPI;

})(window);
