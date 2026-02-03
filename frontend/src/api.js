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

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
  module.exports = KotomiAPI;
}
