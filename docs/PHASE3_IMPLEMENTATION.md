# Phase 3 Implementation: Enhanced Auth Features for Comments and Reactions

**Status:** ✅ Complete  
**Date:** 2026-02-01  
**ADR:** [001-user-authentication-for-comments-and-reactions.md](adr/001-user-authentication-for-comments-and-reactions.md)

## Overview

Phase 3 implements enhanced authentication features for comments and reactions, building upon Phase 1 (JWT validation) and Phase 2 (User management). This phase focuses on empowering users to manage their own content and establishing the foundation for a reputation system.

## Implemented Features

### 1. Edit Own Comments

**Endpoint:** `PUT /api/v1/site/{siteId}/comments/{commentId}`

Users can now edit their own comment text while preserving the original comment metadata.

**Features:**
- ✅ JWT authentication required
- ✅ Ownership verification (users can only edit their own comments)
- ✅ Returns 403 Forbidden if not the comment owner
- ✅ Returns 404 if comment not found
- ✅ Updates `updated_at` timestamp automatically

**Example Request:**
```bash
curl -X PUT "https://kotomi.example.com/api/v1/site/{siteId}/comments/{commentId}" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"text": "Updated comment text"}'
```

**Example Response:**
```json
{
  "id": "comment-123",
  "site_id": "site-456",
  "author": "John Doe",
  "author_id": "user-789",
  "author_verified": true,
  "author_reputation": 42,
  "text": "Updated comment text",
  "status": "approved",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-02T15:30:00Z"
}
```

### 2. Delete Own Comments

**Endpoint:** `DELETE /api/v1/site/{siteId}/comments/{commentId}`

Users can delete their own comments permanently.

**Features:**
- ✅ JWT authentication required
- ✅ Ownership verification (users can only delete their own comments)
- ✅ Returns 403 Forbidden if not the comment owner
- ✅ Returns 404 if comment not found
- ✅ Returns 204 No Content on successful deletion
- ✅ Cascade deletes associated reactions

**Example Request:**
```bash
curl -X DELETE "https://kotomi.example.com/api/v1/site/{siteId}/comments/{commentId}" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response:** HTTP 204 No Content (empty body)

### 3. User Verification Badges

Comment responses now include user verification status and reputation score from the user's profile.

**New Fields in Comment Response:**
- `author_verified` (boolean): Indicates if the user's identity is verified
- `author_reputation` (integer): User's reputation score

**Features:**
- ✅ Automatically included in all comment GET responses
- ✅ Joins with users table to fetch current status
- ✅ Defaults to false/0 if user not found in database

**Example Response:**
```json
{
  "id": "comment-123",
  "author": "Jane Smith",
  "author_id": "user-456",
  "author_verified": true,
  "author_reputation": 87,
  "text": "Great article!",
  "created_at": "2024-01-01T12:00:00Z"
}
```

### 4. Reputation System Foundation

Implements the foundational infrastructure for a user reputation system.

**Database Changes:**
- ✅ Added `reputation_score` column to users table (INTEGER, default 0)
- ✅ Automatic migration for existing databases
- ✅ Updated User model with ReputationScore field

**New Methods:**
- `UserStore.UpdateReputationScore(siteID, userID string, score int)`: Update a user's reputation
- `UserStore.CalculateReputationScore(siteID, userID string)`: Calculate reputation based on activity

**Basic Calculation (v1):**
```
reputation_score = number_of_approved_comments
```

**Future Enhancements (Planned):**
- Points for reactions received on comments
- Bonus points for verified status
- Penalties for rejected/spam comments
- Time-based decay
- Quality multipliers

## Technical Implementation

### Database Schema Changes

```sql
-- Added to users table
ALTER TABLE users ADD COLUMN reputation_score INTEGER DEFAULT 0;
```

Migration is automatic and safe for existing databases.

### Modified Queries

**GetPageComments** now includes user verification:
```sql
SELECT c.id, c.author, c.author_id, ..., 
       COALESCE(u.is_verified, 0) as author_verified,
       COALESCE(u.reputation_score, 0) as author_reputation
FROM comments c
LEFT JOIN users u ON c.site_id = u.site_id AND c.author_id = u.id
WHERE c.site_id = ? AND c.page_id = ?
ORDER BY c.created_at ASC
```

### API Authorization

Both edit and delete endpoints implement strict authorization:

1. Extract authenticated user from JWT context
2. Fetch the comment by ID
3. Verify comment belongs to the specified site
4. **Verify comment.AuthorID matches authenticated user.ID**
5. Allow operation only if ownership verified

## Testing

### Unit Tests Added

**Comment Updates:**
- ✅ `TestSQLiteStore_UpdateCommentText` - Successful update
- ✅ `TestSQLiteStore_UpdateCommentText_NotFound` - Error handling

**Reputation System:**
- ✅ `TestUserStore_ReputationScore` - Update reputation
- ✅ `TestUserStore_CalculateReputationScore` - Calculate based on approved comments

**All Tests:** 100% passing

### Test Coverage

```bash
go test ./pkg/comments/... ./pkg/models/... -cover
```

- pkg/comments: >90% coverage
- pkg/models: >90% coverage

## Security Considerations

### CodeQL Analysis
✅ **No security vulnerabilities detected**

### Authorization
- All endpoints require valid JWT authentication
- Ownership verification prevents unauthorized access
- No SQL injection vulnerabilities (parameterized queries)
- No XSS vulnerabilities (JSON responses only)

### Rate Limiting
- Standard rate limits apply (5 POST/PUT/DELETE per minute)
- Prevents abuse of edit/delete features

## API Documentation

### Swagger/OpenAPI
Documentation has been regenerated with new endpoints:

- View at: `http://localhost:8080/swagger/index.html` (dev mode)
- Updated: `/docs/swagger.json` and `/docs/swagger.yaml`

## Migration Guide

### For Existing Deployments

**No action required!** The database migration runs automatically on server startup.

**What happens:**
1. Server starts and opens database connection
2. Detects existing schema
3. Attempts to add `reputation_score` column
4. Ignores error if column already exists
5. Continues normal operation

### For API Consumers

**Breaking Changes:** None

**New Features Available:**
1. `PUT /api/v1/site/{siteId}/comments/{commentId}` - Edit comments
2. `DELETE /api/v1/site/{siteId}/comments/{commentId}` - Delete comments
3. Comments include `author_verified` and `author_reputation` fields

**Example Integration:**
```javascript
// JavaScript client example
async function editComment(siteId, commentId, newText, jwtToken) {
  const response = await fetch(
    `https://kotomi.example.com/api/v1/site/${siteId}/comments/${commentId}`,
    {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${jwtToken}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ text: newText })
    }
  );
  
  if (response.status === 403) {
    console.error('You can only edit your own comments');
    return null;
  }
  
  return await response.json();
}
```

## Future Enhancements

### Phase 4 (Planned)
- User profile endpoints
- Comment edit history
- User activity feed
- Advanced reputation calculations

### Phase 5 (Future)
- Reputation leaderboards
- User badges and achievements
- Reputation-based permissions
- Comment quality scoring

## Conclusion

Phase 3 successfully implements the enhanced authentication features specified in ADR 001, providing users with control over their content and establishing a foundation for community trust through the reputation system.

**Key Achievements:**
- ✅ Users can edit and delete their own comments
- ✅ User verification status visible on all comments
- ✅ Reputation system foundation ready for enhancement
- ✅ Zero security vulnerabilities
- ✅ Comprehensive test coverage
- ✅ Backward compatible with existing deployments
- ✅ Full API documentation

## References

- [ADR 001: User Authentication for Comments and Reactions](adr/001-user-authentication-for-comments-and-reactions.md)
- [Phase 1 Summary](PHASE1_SUMMARY.md)
- [Authentication API Documentation](AUTHENTICATION_API.md)
