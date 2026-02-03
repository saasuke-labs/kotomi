# Kotomi Widget Examples

This directory contains example integrations for the Kotomi comment widget.

## Available Examples

### 1. Basic Integration Example (`index.html`)

A complete, interactive example showing all Kotomi widget features:
- Live configuration panel
- All widget options demonstrated
- Interactive testing interface
- Local development ready

**Use this to:**
- Test the widget locally
- Experiment with different configurations
- Understand all available options

**To run:**
```bash
# Start Kotomi server
cd ../..
go run cmd/main.go

# Open in browser
open frontend/examples/index.html
```

### 2. Gengo Static Site Generator Example (`gengo-example.html`)

A template specifically designed for integration with Gengo static site generator:
- Uses Gengo template variables
- Shows recommended placement
- Includes conditional comments
- Production-ready structure

**Use this to:**
- Integrate Kotomi into your Gengo blog
- Copy template code into your theme
- Understand Gengo-specific setup

**How to use:**
1. Copy this template to your Gengo theme's layout directory (e.g., `themes/your-theme/layouts/post.html`)
2. Replace `'my-gengo-blog'` with your actual site ID from Kotomi admin
3. Adjust Gengo template variables based on your version (check `{{ .Slug }}`, `{{ .Title }}`, etc.)
4. Update the `baseUrl` to point to your Kotomi server
5. Build your Gengo site: `gengo build`

**For complete instructions, see:**
[Gengo Integration Guide](../../docs/GENGO_INTEGRATION_GUIDE.md)

## Quick Start

### For Gengo Users

1. **Start Kotomi server:**
   ```bash
   docker run -p 8080:8080 -v kotomi-data:/app/data ghcr.io/saasuke-labs/kotomi:latest
   ```

2. **Create a site in Kotomi admin:**
   - Visit http://localhost:8080/admin
   - Create a new site with ID: `my-gengo-blog`

3. **Copy the template:**
   - Copy `gengo-example.html` to your Gengo theme
   - Update `siteId` to match your site ID
   - Replace template variables as needed

4. **Build and test:**
   ```bash
   gengo build
   gengo serve
   ```

### For Other Static Site Generators

While these examples are focused on Gengo, Kotomi works with any static site generator:

- **Hugo**: Use `{{ .Slug }}` or `{{ .File.UniqueID }}`
- **Jekyll**: Use `{{ page.slug }}` or `{{ page.id }}`
- **Eleventy**: Use `{{ page.fileSlug }}` or `{{ page.url }}`
- **Next.js**: Use `router.asPath` or page slug from props
- **Gatsby**: Use page context or slug from GraphQL

The key is to provide:
1. A unique `siteId` for your website
2. A unique `pageId` for each page/post

## Configuration Reference

### Minimal Configuration

```javascript
const kotomi = new Kotomi({
  baseUrl: 'http://localhost:8080',
  siteId: 'my-site',
  pageId: 'unique-page-id'
});
kotomi.render();
```

### Full Configuration

```javascript
const kotomi = new Kotomi({
  // Required
  baseUrl: 'https://comments.yourdomain.com',
  siteId: 'my-site-id',
  pageId: 'unique-page-identifier',
  
  // Optional
  container: '#kotomi-comments',
  theme: 'light',                    // or 'dark'
  enableReactions: true,
  enableReplies: true,
  jwtToken: getUserToken(),          // For authenticated users
  placeholder: 'Write a comment...',
  submitText: 'Post Comment'
});
kotomi.render();
```

## Documentation Links

- [Complete Gengo Integration Guide](../../docs/GENGO_INTEGRATION_GUIDE.md)
- [Widget API Reference](../README.md)
- [Authentication Setup](../../docs/AUTHENTICATION_API.md)
- [Main Kotomi Documentation](../../README.md)

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/saasuke-labs/kotomi/issues)
- **Discussions**: [GitHub Discussions](https://github.com/saasuke-labs/kotomi/discussions)
- **Documentation**: [Full Documentation](https://github.com/saasuke-labs/kotomi)

## Contributing

Have an example for another static site generator? Contributions are welcome!

1. Create a new example file (e.g., `hugo-example.html`)
2. Add it to this README
3. Submit a pull request

---

**Happy commenting! 💬**
