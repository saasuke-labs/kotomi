# Beta Tester Onboarding Checklist

This checklist guides you through onboarding a new beta tester for Kotomi. Use this to ensure consistent and thorough onboarding for each beta participant.

## Pre-Onboarding Preparation

### Beta Tester Vetting
- [ ] Confirm beta tester meets selection criteria (see ADR 004 Appendix)
  - [ ] Has a static website (blog, docs, marketing site)
  - [ ] Technical enough to follow deployment guide
  - [ ] Willing to provide detailed feedback
  - [ ] Available for regular check-ins
  - [ ] Patient with beta quality software
  - [ ] Represents different use case from existing testers
- [ ] Record beta tester information:
  - Name: _________________
  - Email: _________________
  - Website: _________________
  - Use Case: _________________
  - Preferred Contact Method: _________________
  - Deployment Preference: _________________

### Resource Preparation
- [ ] Review [Beta Tester Guide](BETA_TESTER_GUIDE.md)
- [ ] Review [Admin Panel Guide](ADMIN_PANEL_GUIDE.md)
- [ ] Review [Authentication API Guide](AUTHENTICATION_API.md)
- [ ] Prepare any custom instructions specific to their use case
- [ ] Schedule initial onboarding call (1-2 hours)

## Day 1: Initial Setup

### Welcome & Orientation (30 min)
- [ ] Send welcome email with agenda
- [ ] Share overview of Kotomi features and capabilities
- [ ] Set expectations:
  - [ ] Beta software - bugs expected
  - [ ] Breaking changes possible
  - [ ] Response time: 24-48 hours
  - [ ] Weekly check-in schedule
- [ ] Share key documentation links
- [ ] Explain feedback mechanisms (GitHub Issues, Discussions, direct contact)

### Technical Setup (60-90 min)
- [ ] Review system requirements
- [ ] Choose deployment method:
  - [ ] Option 1: Docker (recommended)
  - [ ] Option 2: Cloud Run (if they prefer managed)
  - [ ] Option 3: Binary (for advanced users)
- [ ] Walk through deployment process
  - [ ] Share [Beta Tester Guide](BETA_TESTER_GUIDE.md) deployment section
  - [ ] Assist with environment variable configuration
  - [ ] Help troubleshoot any deployment issues
- [ ] Verify successful deployment
  - [ ] Health endpoint accessible: `/healthz`
  - [ ] Admin panel accessible: `/admin/dashboard`
  - [ ] API documentation accessible: `/docs/index.html`

### Authentication Configuration (30 min)
- [ ] Admin Panel Auth (Auth0):
  - [ ] Guide through Auth0 account creation (if needed)
  - [ ] Configure Auth0 application
  - [ ] Set up Auth0 credentials in Kotomi
  - [ ] Test admin panel login
  - [ ] Verify access to dashboard
- [ ] User Auth (JWT):
  - [ ] Explain JWT authentication model
  - [ ] Discuss their existing authentication system
  - [ ] Choose JWT validation method (HMAC, RSA, ECDSA)
  - [ ] Share JWT token generation examples

## Day 2-3: First Site Integration

### Create First Site (15 min)
- [ ] Access admin panel together
- [ ] Create first site:
  - Site name: _________________
  - Site URL: _________________
  - Description: _________________
- [ ] Configure site settings (CORS, rate limits)
- [ ] Add first page to site

### Configure JWT Authentication (30 min)
- [ ] Navigate to site authentication settings
- [ ] Configure JWT validation:
  - [ ] Choose validation type
  - [ ] Set issuer and audience
  - [ ] Configure secret or public key
  - [ ] Set expiration buffer (optional)
- [ ] Test JWT token generation with their auth system
- [ ] Verify JWT validation works

### Frontend Integration (60 min)
- [ ] Review integration options:
  - [ ] Direct API calls
  - [ ] Kotomi widget (if applicable)
  - [ ] Custom implementation
- [ ] Share integration examples:
  - [ ] HTML/JavaScript example
  - [ ] Framework-specific examples (if needed)
- [ ] Implement basic comment form on test page
- [ ] Implement comment display
- [ ] Test complete flow:
  - [ ] User logs in on their site
  - [ ] User posts comment via Kotomi
  - [ ] Comment appears in admin panel
  - [ ] Comment displays on frontend

### Test Core Features (30 min)
- [ ] Post first real comment
- [ ] Add reaction to page
- [ ] Moderate comment in admin panel:
  - [ ] Approve comment
  - [ ] Test reject (optional)
  - [ ] Test delete (optional)
- [ ] Verify email notifications (if configured)
- [ ] Test AI moderation (if enabled)

## Day 4-7: Advanced Features & Optimization

### Multi-Page Setup (15 min)
- [ ] Add multiple pages to site
- [ ] Test comments on different pages
- [ ] Verify page isolation
- [ ] Review page management in admin panel

### Moderation Workflow (30 min)
- [ ] Review moderation queue
- [ ] Test bulk operations (if available)
- [ ] Configure moderation rules (if applicable)
- [ ] Set up moderation notifications

### Performance & Monitoring (30 min)
- [ ] Review server logs
- [ ] Check database size and growth
- [ ] Monitor API response times
- [ ] Review any error messages
- [ ] Discuss scaling considerations (if high traffic expected)

### Polish & Refinement (flexible)
- [ ] Refine frontend styling
- [ ] Adjust rate limits based on traffic
- [ ] Configure additional CORS origins (if needed)
- [ ] Set up backup schedule (if self-hosted)

## Week 2: Feedback Collection

### Initial Feedback Session (60 min)
Schedule dedicated feedback call:
- [ ] What worked well?
- [ ] What was confusing or difficult?
- [ ] What features are missing?
- [ ] What bugs or issues encountered?
- [ ] Documentation gaps?
- [ ] Performance concerns?
- [ ] Security concerns?
- [ ] Overall impression and satisfaction

### Document Feedback
- [ ] Create GitHub Issues for bugs
- [ ] Create GitHub Discussions for feature requests
- [ ] Update internal feedback tracking
- [ ] Prioritize issues for next release

### Continuous Support
- [ ] Add to beta tester communication channel
- [ ] Send weekly check-in email
- [ ] Monitor their GitHub activity
- [ ] Respond to questions within 24-48 hours

## Ongoing: Success Metrics

### Track Beta Tester Health
- [ ] Site still active? (Check weekly)
- [ ] Comments being posted? (Check weekly)
- [ ] Responding to check-ins? (Monitor weekly)
- [ ] Reporting issues? (Track continuously)
- [ ] Overall satisfaction? (Assess monthly)

### Milestone Checklist
After 2 weeks:
- [ ] Beta tester successfully deployed
- [ ] At least one site actively using Kotomi
- [ ] Comments being posted regularly
- [ ] Beta tester provided feedback
- [ ] No blocking issues preventing usage

After 1 month:
- [ ] Beta tester still active
- [ ] Multiple pages using comments
- [ ] Positive feedback on core functionality
- [ ] Feature requests documented
- [ ] Beta tester willing to continue

## Offboarding (If Needed)

If beta tester decides to leave program:
- [ ] Exit interview (understand why)
- [ ] Export their data (if requested)
- [ ] Document lessons learned
- [ ] Thank them for participation
- [ ] Offer to stay in touch for future releases

## Notes

### Common Issues & Solutions

**Issue**: Docker deployment fails
- Solution: Check Docker version, verify port availability, review logs

**Issue**: Auth0 login not working
- Solution: Verify callback URL, check credentials, review Auth0 logs

**Issue**: JWT validation failing
- Solution: Verify token format, check secret/key, confirm issuer/audience match

**Issue**: Comments not appearing
- Solution: Check JWT authentication, verify site/page configuration, review API logs

**Issue**: Admin panel not accessible
- Solution: Verify Auth0 setup, check session secret, review browser console

### Tips for Successful Onboarding

1. **Be Patient**: First-time setup can take 2-4 hours
2. **Screen Share**: Visual guidance is much more effective than written instructions
3. **Test Everything**: Don't assume anything works - verify each step
4. **Document Issues**: Note any confusion or difficulty for documentation improvements
5. **Follow Up**: Check in within 24 hours to ensure they're not blocked
6. **Celebrate Wins**: Acknowledge when milestones are reached

### Resources

- [Beta Tester Guide](BETA_TESTER_GUIDE.md) - Complete deployment and usage guide
- [Admin Panel Guide](ADMIN_PANEL_GUIDE.md) - Admin panel features and workflows
- [Authentication API Guide](AUTHENTICATION_API.md) - JWT setup and examples
- [Database Backup Guide](DATABASE_BACKUP_RESTORE.md) - Backup and restore procedures
- [Release Process](RELEASE_PROCESS.md) - How we handle updates and patches

### Contact Information

- **GitHub Issues**: https://github.com/saasuke-labs/kotomi/issues (Bug reports)
- **GitHub Discussions**: https://github.com/saasuke-labs/kotomi/discussions (Feature requests, questions)
- **Direct Contact**: [Add team contact info]
- **Office Hours**: [Add schedule if applicable]
