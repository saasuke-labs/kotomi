# Quick Start: Kotomi with Gengo (5 Minutes)

Get Kotomi comments running on your Gengo blog in 5 minutes.

```
┌─────────────────────────────────────────────────────────────┐
│                    Integration Flow                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. Start Kotomi Server (Docker/Go)                          │
│     └─→ Provides REST API & Admin Panel                     │
│                                                              │
│  2. Create Site in Kotomi Admin                              │
│     └─→ Get your Site ID                                    │
│                                                              │
│  3. Add Widget to Gengo Template                             │
│     └─→ Include CSS + JS + Initialize                       │
│                                                              │
│  4. Build & Deploy Gengo Site                                │
│     └─→ Static HTML with embedded comment widget            │
│                                                              │
│  User visits page → Widget loads → Connects to Kotomi API   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Step 1: Start Kotomi Server (1 minute)

```bash
docker run -d --name kotomi -p 8080:8080 -v kotomi-data:/app/data ghcr.io/saasuke-labs/kotomi:latest
```

Verify it's running:
```bash
curl http://localhost:8080/healthz
# Should return: {"message":"OK"}
```

## Step 2: Create Your Site (1 minute)

1. Open http://localhost:8080/admin in your browser
2. Click "Sites" → "Create New Site"
3. Fill in:
   - **Site ID**: `my-blog` (lowercase, no spaces)
   - **Name**: "My Blog"
   - **Domain**: `localhost` (or your domain)
4. Click "Create"

**Remember your Site ID - you'll need it in the next step!**

## Step 3: Add to Gengo Template (2 minutes)

Edit your Gengo post layout file (e.g., `themes/your-theme/layouts/post.html`):

```html
<!DOCTYPE html>
<html>
<head>
  <title>{{ .Title }}</title>
  <!-- Your existing CSS -->
  
  <!-- Add Kotomi CSS -->
  <link rel="stylesheet" href="http://localhost:8080/static/kotomi.css">
</head>
<body>
  <!-- Your blog post content -->
  <article>
    <h1>{{ .Title }}</h1>
    {{ .Content }}
  </article>

  <!-- Add this section for comments -->
  <section class="comments">
    <h2>Comments</h2>
    <div id="kotomi-comments"></div>
  </section>

  <!-- Add Kotomi JavaScript -->
  <script src="http://localhost:8080/static/kotomi.js"></script>
  <script>
    const kotomi = new Kotomi({
      baseUrl: 'http://localhost:8080',
      siteId: 'my-blog',              // ← Change this to your Site ID from Step 2
      pageId: '{{ .Slug }}'           // ← Gengo provides this
    });
    kotomi.render();
  </script>
</body>
</html>
```

## Step 4: Build and Test (1 minute)

```bash
# Build your Gengo site
gengo build

# Serve locally
gengo serve
```

Open your blog in a browser - you should see the comment widget below your posts!

## That's It! 🎉

You now have a working comment system on your Gengo blog.

## Next Steps

### Make it Production-Ready

1. **Deploy Kotomi to a server:**
   ```bash
   # On your server
   docker run -d \
     --name kotomi \
     -p 8080:8080 \
     -v kotomi-data:/app/data \
     -e ENV=production \
     -e CORS_ORIGINS=https://yourdomain.com \
     ghcr.io/saasuke-labs/kotomi:latest
   ```

2. **Update your template to use production URL:**
   ```javascript
   const kotomi = new Kotomi({
     baseUrl: 'https://comments.yourdomain.com',  // Production URL
     siteId: 'my-blog',
     pageId: '{{ .Slug }}'
   });
   ```

3. **Set up HTTPS with nginx** (recommended)

### Add User Authentication

To allow users to post comments, you need JWT authentication:

1. Generate a JWT secret:
   ```bash
   openssl rand -base64 32
   ```

2. Configure Kotomi:
   ```bash
   docker run -d \
     -e JWT_SECRET=your-generated-secret \
     ghcr.io/saasuke-labs/kotomi:latest
   ```

3. Generate JWT tokens in your app and pass to widget:
   ```javascript
   const kotomi = new Kotomi({
     baseUrl: 'http://localhost:8080',
     siteId: 'my-blog',
     pageId: '{{ .Slug }}',
     jwtToken: 'your-jwt-token-here'  // From your auth system
   });
   ```

See [Authentication Guide](AUTHENTICATION_API.md) for details.

### Customize the Look

Add custom CSS to match your theme:

```css
/* In your theme CSS */
.kotomi-comments {
  font-family: your-font;
}

.kotomi-comment {
  border-left: 3px solid #007bff;
}
```

### Enable AI Moderation

1. Get an OpenAI API key from https://platform.openai.com
2. Configure Kotomi:
   ```bash
   docker run -d \
     -e OPENAI_API_KEY=sk-your-key-here \
     ghcr.io/saasuke-labs/kotomi:latest
   ```
3. Configure moderation settings in admin panel

## Common Issues

**Comments not showing?**
- Check your `siteId` matches the one in Kotomi admin
- Check browser console for errors
- Verify Kotomi server is running: `curl http://localhost:8080/healthz`

**Can't post comments?**
- You need JWT authentication for write operations
- See [Authentication Guide](AUTHENTICATION_API.md)

**Widget not loading?**
- Check the script URLs are correct
- Verify CORS is configured for your domain (production only)

## Complete Documentation

For detailed instructions, configuration options, and advanced features:

📖 **[Complete Gengo Integration Guide](GENGO_INTEGRATION_GUIDE.md)**

## Need Help?

- 💬 [GitHub Discussions](https://github.com/saasuke-labs/kotomi/discussions)
- 🐛 [Report Issues](https://github.com/saasuke-labs/kotomi/issues)
- 📚 [Full Documentation](../README.md)
