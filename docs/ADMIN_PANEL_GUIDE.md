# Admin Panel Guide

This guide covers all features and workflows in the Kotomi admin panel.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Dashboard Overview](#dashboard-overview)
3. [Sites Management](#sites-management)
4. [Pages Management](#pages-management)
5. [Comments Management](#comments-management)
6. [Moderation Workflow](#moderation-workflow)
7. [User Management](#user-management)
8. [Settings](#settings)

## Getting Started

### Accessing the Admin Panel

1. Navigate to your Kotomi instance: `https://your-kotomi-url.com/admin/dashboard`
2. Click "Login" - you'll be redirected to Auth0
3. Sign in with your Auth0 account
4. You'll be redirected back to the dashboard

### First Login

On your first login:
1. You'll see an empty dashboard
2. Start by creating your first site
3. Add pages to your site
4. Begin receiving comments

## Dashboard Overview

The dashboard provides a quick overview of your Kotomi instance.

### Key Metrics (Coming Soon)

- Total sites
- Total pages
- Total comments
- Pending moderation count
- Recent activity

### Quick Actions

- Create new site
- View pending comments
- Access user management

## Sites Management

Sites represent the websites where you'll embed Kotomi comments.

### Creating a Site

1. Click **Sites** in the navigation menu
2. Click **Add New Site**
3. Fill in the form:
   - **Name**: A friendly name for your site (e.g., "My Blog")
   - **Domain**: The domain where comments will be used (e.g., "myblog.com")
   - **Description**: Optional description of the site
4. Click **Create Site**

**Note**: Copy the Site ID after creation - you'll need it for API calls.

### Viewing Sites

The sites list shows:
- Site name
- Domain
- Creation date
- Number of pages
- Number of comments

Click on a site name to view its details and pages.

### Editing a Site

1. Navigate to the site you want to edit
2. Click **Edit Site**
3. Update the fields:
   - Name
   - Domain
   - Description
4. Click **Save Changes**

### Deleting a Site

⚠️ **Warning**: Deleting a site will also delete all associated pages and comments. This action cannot be undone.

1. Navigate to the site you want to delete
2. Click **Delete Site**
3. Confirm the deletion

## Pages Management

Pages represent individual URLs on your website where comments can be posted.

### Creating a Page

1. Navigate to the site
2. Click **Add Page**
3. Fill in the form:
   - **URL**: The page path (e.g., "/blog/my-first-post")
   - **Title**: The page title (e.g., "My First Post")
4. Click **Create Page**

**Best Practice**: Create pages automatically via API when the first comment is posted, rather than manually creating each page.

### Viewing Pages

The pages list shows:
- Page title
- URL
- Creation date
- Number of comments
- Last comment date

Click on a page to view its comments.

### Editing a Page

1. Navigate to the page
2. Click **Edit Page**
3. Update the fields:
   - URL
   - Title
4. Click **Save Changes**

### Deleting a Page

⚠️ **Warning**: Deleting a page will also delete all its comments. This action cannot be undone.

1. Navigate to the page
2. Click **Delete Page**
3. Confirm the deletion

## Comments Management

The comments section is where you review and moderate all comments.

### Viewing Comments

#### All Comments View

1. Click **Comments** in the navigation
2. See all comments across all sites
3. Filter by:
   - Site
   - Page
   - Status (pending, approved, rejected)
   - Date range

#### Site Comments View

1. Navigate to a specific site
2. Click **Comments** tab
3. See all comments for that site

#### Page Comments View

1. Navigate to a specific page
2. See all comments for that page

### Comment Details

Each comment shows:
- **Author**: Name provided by commenter
- **Email**: Email provided by commenter (if any)
- **Text**: The comment content
- **Created**: When the comment was posted
- **Status**: Pending, Approved, or Rejected
- **IP Address**: IP address of commenter (for abuse tracking)
- **Parent Comment**: If this is a reply, shows the parent

### Comment Actions

Available actions for each comment:
- **Approve**: Make the comment visible
- **Reject**: Hide the comment
- **Delete**: Permanently remove the comment
- **View Thread**: See the comment in context with replies

## Moderation Workflow

Kotomi supports a moderation workflow to ensure quality discussions.

### Moderation Status

Comments can be in one of three states:

1. **Pending**: Newly posted, awaiting moderation
2. **Approved**: Reviewed and made visible
3. **Rejected**: Reviewed and hidden from public view

### Auto-Moderation (If Enabled)

If AI moderation is configured:
- Comments are automatically analyzed
- Toxic or inappropriate comments are flagged
- Flagged comments are marked as rejected
- You can review and override AI decisions

### Manual Moderation

#### Approving Comments

1. Navigate to pending comments
2. Read the comment
3. Click **Approve** if appropriate
4. Comment becomes visible via API

#### Rejecting Comments

1. Navigate to pending comments
2. Read the comment
3. Click **Reject** if inappropriate
4. Comment is hidden from API results

**Note**: Rejected comments are still stored for your records and abuse tracking.

### Bulk Moderation (Coming Soon)

- Select multiple comments
- Approve or reject in bulk
- Filter by patterns (e.g., spam domains)

### Moderation Best Practices

1. **Set Clear Guidelines**: Publish community guidelines for commenters
2. **Be Consistent**: Apply the same standards to all comments
3. **Be Timely**: Review comments regularly (daily for active sites)
4. **Context Matters**: Consider the discussion context
5. **Document Decisions**: Note why comments were rejected (for your records)

## User Management

Manage who has access to the Kotomi admin panel.

### Viewing Users

1. Click **Users** in the navigation
2. See all users with admin access
3. View:
   - Email
   - Name
   - Auth0 ID
   - Last login
   - Sites they own

### User Roles (Current Implementation)

Currently, all users who successfully log in via Auth0 are administrators. Each user can only manage their own sites, pages, and comments.

**Future Enhancement**: Role-based access control (admin, moderator, viewer)

### Removing User Access

To remove a user's access:
1. Delete the user from your Auth0 tenant
2. Or configure Auth0 rules to restrict access

## Settings

### Account Settings

View and update your account information:
- Email
- Display name
- Profile picture (from Auth0)

### Logout

Click **Logout** in the navigation to end your session.

## Keyboard Shortcuts (Coming Soon)

Speed up your workflow with keyboard shortcuts:

- `n`: New site
- `c`: View comments
- `a`: Approve selected comment
- `r`: Reject selected comment
- `?`: Show help

## Tips and Tricks

### Efficient Moderation

1. **Use Filters**: Filter by pending to focus on new comments
2. **Sort by Date**: Review oldest comments first
3. **Check Context**: Click through to see the page context
4. **Track Patterns**: Note repeat offenders for blocking

### Managing Multiple Sites

1. **Consistent Naming**: Use clear site names
2. **Use Domains**: Set accurate domains for tracking
3. **Regular Reviews**: Check each site's comments regularly

### API Integration

Use the admin panel to:
1. Create sites and pages
2. Get IDs for API calls
3. Monitor comment activity
4. Moderate content

Then use the API for:
- Automated page creation
- Frontend comment display
- Programmatic moderation (future)

## Mobile Access

The admin panel works on mobile devices:
- Responsive design
- Touch-friendly buttons
- Mobile-optimized views

Best experience on tablets and desktop.

## Browser Support

Supported browsers:
- Chrome/Edge (latest)
- Firefox (latest)
- Safari (latest)
- Mobile Safari (iOS 14+)
- Mobile Chrome (Android)

## Troubleshooting

### Can't Log In

1. Check Auth0 configuration
2. Verify callback URL is correct
3. Clear browser cookies
4. Try incognito/private mode

### Comments Not Showing

1. Check the site and page IDs
2. Verify comments are approved
3. Check API CORS settings
4. Review browser console for errors

### Slow Performance

1. SQLite may slow with many comments
2. Consider archiving old comments
3. For large sites, contact support about PostgreSQL

### Lost Data

1. Check if you're looking at the right site
2. Verify you didn't accidentally delete
3. Check database backups
4. Contact support if data is critical

## Getting Help

- **Documentation**: https://github.com/saasuke-labs/kotomi/docs
- **Issues**: https://github.com/saasuke-labs/kotomi/issues
- **Discussions**: https://github.com/saasuke-labs/kotomi/discussions
- **Email**: support@saasuke-labs.com

## What's Next

Planned admin panel features:
- Analytics dashboard
- Bulk moderation
- Comment export
- Webhook configuration
- Custom moderation rules
- User roles and permissions
- Comment templates
- Automated responses

---

**Version**: 0.1.0-beta.1
**Last Updated**: February 4, 2026
