/**
 * Kotomi - Comment and Reaction Widget
 * Version: 0.1.0
 * 
 * A simple, embeddable comment widget for static websites
 */

(function(window) {
  'use strict';

  // Include API client
  // <<API_CLIENT>>

  // Include UI module
  // <<UI_MODULE>>

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
