#!/bin/bash

# Kotomi Widget Build Script
# Combines JavaScript files and creates distributable bundle

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC_DIR="$SCRIPT_DIR/src"
DIST_DIR="$SCRIPT_DIR/dist"

echo "Building Kotomi Widget..."

# Create dist directory
mkdir -p "$DIST_DIR"

# Read source files
API_CLIENT=$(cat "$SRC_DIR/api.js")
UI_MODULE=$(cat "$SRC_DIR/ui.js")
KOTOMI_MAIN=$(cat "$SRC_DIR/kotomi.js")

# Create combined JavaScript file
cat > "$DIST_DIR/kotomi.js" << 'EOF'
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

EOF

# Append API client (without exports)
echo "$API_CLIENT" | sed '/^\/\/ Export for module systems/,$d' >> "$DIST_DIR/kotomi.js"

echo "" >> "$DIST_DIR/kotomi.js"

# Append UI module (without exports)
echo "$UI_MODULE" | sed '/^\/\/ Export for module systems/,$d' >> "$DIST_DIR/kotomi.js"

echo "" >> "$DIST_DIR/kotomi.js"

# Append main Kotomi class
cat >> "$DIST_DIR/kotomi.js" << 'EOF'

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
EOF

# Copy CSS
cp "$SRC_DIR/styles.css" "$DIST_DIR/kotomi.css"

# Create minified versions (simple approach without external tools)
echo "Creating minified versions..."

# For CSS, remove comments and excess whitespace
# Note: This is a simple minification. For production, consider using a proper CSS minifier.
cat "$DIST_DIR/kotomi.css" | \
  tr '\n' ' ' | \
  sed 's|/\*[^*]*\*\+\([^/*][^*]*\*\+\)*/||g' | \
  sed 's/[[:space:]]\{2,\}/ /g' | \
  sed 's/[[:space:]]*{[[:space:]]*/ {/g' | \
  sed 's/[[:space:]]*}[[:space:]]*/ }/g' | \
  sed 's/[[:space:]]*;[[:space:]]*/ ;/g' | \
  sed 's/[[:space:]]*,[[:space:]]*/ ,/g' | \
  sed 's/^[[:space:]]*//g' > "$DIST_DIR/kotomi.min.css"

# For JS, we'd normally use a minifier, but for now just copy
# Note: For production, consider using UglifyJS, Terser, or similar
cp "$DIST_DIR/kotomi.js" "$DIST_DIR/kotomi.min.js"

echo "Build complete!"
echo "Files created:"
echo "  - $DIST_DIR/kotomi.js"
echo "  - $DIST_DIR/kotomi.min.js"
echo "  - $DIST_DIR/kotomi.css"
echo "  - $DIST_DIR/kotomi.min.css"
