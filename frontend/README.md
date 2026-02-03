# Kotomi Widget

A simple, embeddable comment widget for static websites.

## Features

- 💬 **Comment Display** - Show comments from your Kotomi server
- ✍️ **Comment Submission** - Allow users to post new comments (requires JWT authentication)
- 👍 **Reactions** - Display and interact with reactions (emoji responses)
- 💬 **Threaded Replies** - Support for nested comment replies
- 🎨 **Themes** - Light and dark theme support
- 📱 **Responsive** - Mobile-friendly design
- 🔒 **Secure** - JWT-based authentication support
- 🚀 **Lightweight** - No dependencies, pure vanilla JavaScript

## Quick Start

### 1. Include the CSS and JavaScript

Add the following to your HTML page:

```html
<!-- Include Kotomi CSS -->
<link rel="stylesheet" href="https://your-cdn.com/kotomi.css">

<!-- Comment widget container -->
<div id="kotomi-comments"></div>

<!-- Include Kotomi JavaScript -->
<script src="https://your-cdn.com/kotomi.js"></script>
```

### 2. Initialize the Widget

```html
<script>
  const kotomi = new Kotomi({
    baseUrl: 'https://your-kotomi-server.com',
    siteId: 'your-site-id',
    pageId: 'page-slug',
    container: '#kotomi-comments'
  });
  
  kotomi.render();
</script>
```

## Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `baseUrl` | string | Yes | - | URL of your Kotomi server |
| `siteId` | string | Yes | - | Your site identifier |
| `pageId` | string | Yes | - | Unique identifier for the page |
| `container` | string | No | `#kotomi-comments` | CSS selector for the container element |
| `theme` | string | No | `light` | Theme: `light` or `dark` |
| `enableReactions` | boolean | No | `true` | Enable/disable reactions |
| `enableReplies` | boolean | No | `true` | Enable/disable threaded replies |
| `jwtToken` | string | No | `null` | JWT token for authenticated requests |
| `placeholder` | string | No | `Write a comment...` | Placeholder text for comment input |
| `submitText` | string | No | `Post Comment` | Submit button text |

## Examples

### Basic Usage

```html
<!DOCTYPE html>
<html>
<head>
  <link rel="stylesheet" href="kotomi.css">
</head>
<body>
  <h1>My Blog Post</h1>
  <p>Article content...</p>
  
  <div id="kotomi-comments"></div>
  
  <script src="kotomi.js"></script>
  <script>
    const kotomi = new Kotomi({
      baseUrl: 'https://comments.example.com',
      siteId: 'my-blog',
      pageId: 'article-1'
    });
    
    kotomi.render();
  </script>
</body>
</html>
```

### With Authentication

```html
<script>
  // Get JWT token from your authentication system
  const jwtToken = getYourJWTToken();
  
  const kotomi = new Kotomi({
    baseUrl: 'https://comments.example.com',
    siteId: 'my-blog',
    pageId: 'article-1',
    jwtToken: jwtToken
  });
  
  kotomi.render();
</script>
```

### Dark Theme

```html
<script>
  const kotomi = new Kotomi({
    baseUrl: 'https://comments.example.com',
    siteId: 'my-blog',
    pageId: 'article-1',
    theme: 'dark'
  });
  
  kotomi.render();
</script>
```

### Disable Reactions

```html
<script>
  const kotomi = new Kotomi({
    baseUrl: 'https://comments.example.com',
    siteId: 'my-blog',
    pageId: 'article-1',
    enableReactions: false
  });
  
  kotomi.render();
</script>
```

## API Methods

### `render()`

Renders the widget in the specified container.

```javascript
await kotomi.render();
```

### `refresh()`

Reloads comments and reactions from the server.

```javascript
await kotomi.refresh();
```

### `setToken(token)`

Sets or updates the JWT authentication token.

```javascript
kotomi.setToken('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...');
```

### `destroy()`

Removes the widget from the DOM.

```javascript
kotomi.destroy();
```

## Authentication

The Kotomi widget requires JWT authentication for write operations (posting, editing, deleting comments, adding reactions).

### JWT Token Format

Your JWT token must include a `kotomi_user` claim with user information:

```json
{
  "iss": "your-domain.com",
  "aud": "kotomi",
  "sub": "user123",
  "exp": 1234567890,
  "kotomi_user": {
    "id": "user123",
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

See the [Authentication API Documentation](../../docs/AUTHENTICATION_API.md) for more details.

## Building from Source

### Prerequisites

- Bash shell (Linux, macOS, or WSL on Windows)

### Build Steps

1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```

2. Run the build script:
   ```bash
   ./build.sh
   ```

3. The distributable files will be created in the `dist/` directory:
   - `kotomi.js` - Unminified JavaScript
   - `kotomi.min.js` - Minified JavaScript
   - `kotomi.css` - Unminified CSS
   - `kotomi.min.css` - Minified CSS

## Browser Support

The widget supports all modern browsers:

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+
- Opera 76+

## License

Same license as the Kotomi project.

## Contributing

Contributions are welcome! Please follow the project's contribution guidelines.
