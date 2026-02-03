# ADR 004: Beta Release Requirements

**Status:** Proposed  
**Date:** 2026-02-03  
**Authors:** Kotomi Development Team  
**Deciders:** Product Team, Engineering Team  

## Context and Problem Statement

Kotomi is currently at version 0.0.1 and in early development. The project has implemented core features including comments API, reactions system, JWT authentication, admin panel with Auth0, AI moderation, and email notifications. However, it is not yet ready for production use.

We need to define a clear path to release Kotomi for beta testing. This will allow early adopters to start using and testing the system in real-world scenarios, providing valuable feedback before a full production release. The first beta tester will be a member of the core team who will validate the deployment process and user experience.

**Current State:**
- ✅ Core features implemented (comments, reactions, JWT auth, admin panel, moderation)
- ✅ Comprehensive test coverage (>80%)
- ✅ Docker containerization ready
- ✅ CI/CD pipeline configured (GitHub Actions)
- ⚠️ Cloud Run deployment exists but is disabled (`if: false`)
- ⚠️ No formal release process documented
- ⚠️ Security audit not yet completed
- ⚠️ No production deployment runbook
- ⚠️ Missing beta tester onboarding guide

**Key Questions:**
1. What minimum requirements must be met before beta release?
2. What documentation is needed for beta testers?
3. What deployment process should be followed?
4. What support and feedback mechanisms are needed?
5. How do we version and track the beta releases?

## Decision Drivers

* **Beta Testing Goals**: Enable early adopters to test in real-world scenarios and provide feedback
* **Risk Management**: Ensure beta release is stable enough for testing without major security vulnerabilities
* **User Experience**: Provide clear documentation and setup instructions for beta testers
* **Feedback Loop**: Establish mechanisms to collect and act on beta tester feedback
* **Deployment Simplicity**: Make it easy for beta testers to deploy and configure Kotomi
* **Version Management**: Clear versioning strategy for beta releases
* **Rollback Capability**: Ability to quickly address critical issues
* **Minimal Scope**: Keep beta release focused on core functionality

## Decision

We will prepare Kotomi for beta release by completing the following requirements:

### Phase 1: Pre-Release Preparation (Week 1)

#### 1.1 Security & Code Quality
- [ ] **Security Audit** - Conduct internal security review focusing on:
  - JWT validation and token handling
  - SQL injection prevention (parameterized queries)
  - XSS prevention in comment rendering
  - CSRF protection for admin panel
  - Auth0 configuration security
  - API rate limiting effectiveness
  - Environment variable handling (secrets)
  - CORS configuration review
- [ ] **Dependency Audit** - Review and update dependencies for known vulnerabilities
- [ ] **Code Review** - Address critical code quality issues from ADR 002 that impact security/stability:
  - Review error handling in critical paths
  - Validate database connection handling
  - Check for potential resource leaks

#### 1.2 Documentation
- [ ] **Beta Tester Guide** - Create comprehensive guide including:
  - System requirements
  - Deployment options (Docker, Cloud Run, binary)
  - Configuration reference (environment variables)
  - Quick start tutorial (0 to first comment)
  - Authentication setup (Auth0 for admin, JWT for users)
  - Troubleshooting common issues
  - Known limitations and workarounds
- [ ] **API Documentation** - Ensure Swagger/OpenAPI docs are complete and accurate
- [ ] **Admin Panel Guide** - Document admin panel features and workflows
- [ ] **Integration Examples** - Provide sample code for:
  - Static site integration (HTML/JavaScript)
  - JWT token generation examples (Node.js, Python, Go)
  - Frontend widget setup
- [ ] **CHANGELOG** - Initialize changelog for tracking releases

#### 1.3 Deployment Infrastructure
- [ ] **Production Docker Image** - Optimize Dockerfile for production:
  - Multi-stage build validation
  - Minimal image size
  - Health check configuration
  - Proper signal handling
- [ ] **Cloud Run Configuration** - Prepare production deployment:
  - Enable deployment step in CI/CD (remove `if: false`)
  - Configure environment variables template
  - Set up secret management (GitHub Secrets)
  - Configure custom domain (optional for beta)
  - Set appropriate resource limits (CPU, memory)
  - Configure min/max instances
- [ ] **Database Backup** - Document backup and restore procedures for SQLite
- [ ] **Monitoring Setup** - Basic observability:
  - Logging configuration (structured logs)
  - Health check endpoints validated
  - Error tracking setup (optional: Sentry)
  - Basic metrics collection (Cloud Run native)

#### 1.4 Release Process
- [ ] **Version Bumping** - Update version from 0.0.1 to 0.1.0-beta.1:
  - Update VERSION file
  - Tag release in git
  - Update README badges
- [ ] **Release Checklist** - Create reusable checklist for future releases
- [ ] **Rollback Procedure** - Document how to rollback to previous version
- [ ] **CI/CD Validation** - Test full pipeline end-to-end:
  - Run all tests (unit, integration, E2E)
  - Build Docker image
  - Push to artifact registry
  - Deploy to Cloud Run (staging environment first)
  - Smoke test deployed version

### Phase 2: Beta Release Execution (Week 2)

#### 2.1 Initial Deployment
- [ ] **Deploy to Cloud Run** - Execute production deployment:
  - Deploy to Cloud Run with beta configuration
  - Verify health checks pass
  - Test API endpoints
  - Verify admin panel access
  - Validate JWT authentication
  - Test database persistence
- [ ] **Smoke Testing** - Validate critical paths:
  - Create site via admin panel
  - Add page to site
  - Post comment via API
  - React to page via API
  - Moderate comment in admin panel
  - Verify email notifications work
  - Test AI moderation (if enabled)

#### 2.2 Beta Tester Onboarding
- [ ] **First Beta Tester** - Onboard initial tester (core team member):
  - Provide deployment guide
  - Assist with Auth0 setup
  - Help configure JWT for their site
  - Walk through admin panel features
  - Set up their first site integration
  - Verify comment and reaction functionality
- [ ] **Feedback Collection** - Establish feedback mechanisms:
  - GitHub Issues for bug reports
  - GitHub Discussions for feature requests
  - Direct communication channel (email/Slack)
  - Weekly check-in schedule
- [ ] **Support Plan** - Define beta support approach:
  - Response time expectations (24-48 hours)
  - Office hours for live support
  - Documentation for self-service
  - Issue escalation process

#### 2.3 Monitoring & Iteration
- [ ] **Monitor Deployment** - Track beta release health:
  - Review error logs daily
  - Monitor API response times
  - Track database growth
  - Check resource utilization
  - Review security logs
- [ ] **Iterate Based on Feedback** - Rapid iteration cycle:
  - Prioritize critical bugs
  - Address documentation gaps
  - Fix deployment issues quickly
  - Release patches as needed (0.1.0-beta.2, etc.)
  - Update beta guide with lessons learned

### Phase 3: Beta Expansion (Week 3+)

#### 3.1 Additional Beta Testers
- [ ] **Gradual Expansion** - Add testers incrementally:
  - Onboard 2-3 additional beta testers
  - Different use cases (blog, docs site, app)
  - Collect diverse feedback
  - Validate scalability with multiple sites
- [ ] **Community Building** - Foster beta community:
  - Create beta tester communication channel
  - Share updates and progress
  - Encourage knowledge sharing
  - Recognize contributions

#### 3.2 Beta Stabilization
- [ ] **Address Beta Feedback** - Prioritized improvements:
  - Fix critical bugs reported by testers
  - Improve documentation based on confusion points
  - Enhance error messages and debugging
  - Add frequently requested features (if quick wins)
- [ ] **Performance Optimization** - Address performance issues:
  - Optimize slow API endpoints
  - Improve database query performance
  - Reduce memory usage if needed
  - Cache frequently accessed data
- [ ] **Production Readiness Review** - Assess path to v1.0:
  - Review ADR 002 implementation progress
  - Identify blocking issues for production
  - Plan next development phases
  - Set target date for v1.0 release

## Consequences

### Positive Consequences

* **Early Feedback**: Get real-world feedback before investing in v1.0 features
* **Risk Mitigation**: Identify and fix issues in controlled beta environment
* **Community Building**: Build relationships with early adopters who can become advocates
* **Validation**: Confirm product-market fit and use cases
* **Documentation**: Force clarity in deployment and usage documentation
* **Confidence**: Increase team confidence in production readiness
* **Incremental Rollout**: Gradual scaling allows for issue detection at small scale

### Negative Consequences

* **Support Overhead**: Beta testers will require support and guidance
* **Reputation Risk**: Poor beta experience could harm reputation (mitigated by clear "beta" labeling)
* **Scope Creep**: Feature requests from beta testers may distract from roadmap
* **Technical Debt**: Pressure to ship beta may defer some ADR 002 improvements
* **Breaking Changes**: Beta may require breaking changes based on feedback
* **Resource Commitment**: Team must dedicate time to support beta program

### Neutral Consequences

* **Two-Track Development**: Must maintain beta release while working on next features
* **Version Proliferation**: Multiple beta versions may exist simultaneously
* **Documentation Burden**: Keeping docs updated for beta requires discipline

## Implementation Plan

### Timeline

**Week 1: Pre-Release Preparation**
- Days 1-2: Security audit and dependency review
- Days 3-4: Documentation (beta guide, examples, changelog)
- Days 5-6: Deployment infrastructure and CI/CD validation
- Day 7: Internal testing and release preparation

**Week 2: Beta Release Execution**
- Days 1-2: Deploy to Cloud Run and smoke testing
- Days 3-4: First beta tester onboarding and support
- Days 5-7: Monitoring and initial feedback iteration

**Week 3+: Beta Expansion**
- Ongoing: Add additional beta testers gradually
- Ongoing: Address feedback and iterate
- Ongoing: Monitor, support, and improve

### Success Criteria for Beta Release

**Minimum Requirements (Must Have):**
1. ✅ Security audit completed with no critical vulnerabilities
2. ✅ Deployment to Cloud Run successful and stable
3. ✅ Beta tester guide published and reviewed
4. ✅ First beta tester successfully onboarded
5. ✅ Critical user paths working (create site, post comment, moderate)
6. ✅ Monitoring and logging operational
7. ✅ Feedback mechanism established

**Quality Indicators (Should Have):**
1. ✅ All tests passing (unit, integration, E2E)
2. ✅ API response times < 500ms for 95th percentile
3. ✅ Documentation clarity validated by first tester
4. ✅ Zero critical bugs discovered in first 48 hours
5. ✅ Rollback procedure tested and validated

**Beta Success Metrics (Post-Launch):**
1. 📊 5+ beta testers onboarded within 1 month
2. 📊 >3 sites actively using Kotomi
3. 📊 >80% uptime during beta period
4. 📊 <10% of beta testers report deployment difficulties
5. 📊 >50% of reported issues resolved within 1 week
6. 📊 Positive feedback from majority of beta testers

### Resource Requirements

**Team Effort:**
- Engineering: ~40 hours (security, deployment, bug fixes)
- Documentation: ~16 hours (guides, examples, changelog)
- DevOps: ~8 hours (CI/CD, Cloud Run setup)
- Product/Support: ~16 hours (onboarding, feedback management)

**Infrastructure Costs:**
- Cloud Run: ~$10-30/month (estimated for beta scale)
- Artifact Registry: Negligible (within free tier)
- Domain (optional): ~$12/year
- Auth0: Free tier (up to 7,000 users)

**Tools/Services:**
- GitHub (existing)
- Google Cloud Platform (existing)
- Auth0 (existing)
- OpenAI API (optional, pay-per-use)

## Alternatives Considered

### Alternative 1: Wait for ADR 002 Completion

**Description:** Complete all refactoring from ADR 002 before any beta release.

**Pros:**
- Cleaner codebase before external usage
- Less technical debt carried into production
- Better developer experience for contributors

**Cons:**
- Delayed feedback (6+ weeks per ADR 002)
- Risk of building features users don't need
- No real-world validation of assumptions
- Lost opportunity for early adopters

**Decision:** Rejected. Beta testing provides valuable feedback that should inform ADR 002 implementation priorities.

### Alternative 2: Public Beta Immediately

**Description:** Open beta to public without controlled rollout.

**Pros:**
- Maximum feedback volume
- Faster community growth
- More diverse use cases

**Cons:**
- Overwhelming support burden
- Higher risk of reputation damage
- Difficult to manage quality of feedback
- Potential security exposure at scale

**Decision:** Rejected. Controlled beta with gradual expansion is safer and more manageable.

### Alternative 3: Alpha Release to Internal Team Only

**Description:** Release only to internal team members before external beta.

**Pros:**
- Complete control over feedback
- No reputation risk
- Easier support

**Cons:**
- Limited use case diversity
- Internal bias in feedback
- Doesn't validate real-world deployment scenarios
- Delayed time to market

**Decision:** Rejected. Having at least one external beta tester (even if it's a friendly user) provides more valuable feedback than internal-only testing.

### Alternative 4: Soft Launch Without Beta Label

**Description:** Deploy to production without "beta" designation.

**Pros:**
- No expectation setting around bugs
- Simpler marketing message

**Cons:**
- Higher expectations from users
- Greater reputation risk if issues occur
- Less tolerance for breaking changes
- Inappropriate given current maturity level

**Decision:** Rejected. Clear "beta" labeling sets appropriate expectations and provides flexibility.

## Related Decisions

- **ADR 001**: User Authentication for Comments and Reactions - Implemented, part of beta
- **ADR 002**: Code Structure and Go 1.25 Improvements - Some items deferred until post-beta based on feedback
- **Future ADR**: Production Hardening Requirements - Will address full production release (v1.0) requirements

## References

- [Status.md](../../Status.md) - Current feature implementation status
- [README.md](../../README.md) - Project overview and quick start
- [CI/CD Pipeline](.github/workflows/deploy_kotomi.yaml) - Deployment automation
- [Dockerfile](../../Dockerfile) - Container configuration
- [ADR 001](001-user-authentication-for-comments-and-reactions.md) - Authentication architecture
- [ADR 002](002-code-structure-and-go-1.25-improvements.md) - Code quality improvements

## Appendix: Beta Release Checklist Template

This checklist can be used for each beta version release:

```markdown
## Beta Release Checklist: v0.X.0-beta.Y

### Pre-Release
- [ ] All tests passing (unit, integration, E2E)
- [ ] Security scan completed
- [ ] Documentation updated
- [ ] CHANGELOG updated
- [ ] Version bumped in VERSION file
- [ ] Git tag created
- [ ] Release notes drafted

### Deployment
- [ ] Docker image built and pushed
- [ ] Cloud Run deployment successful
- [ ] Health checks passing
- [ ] Database migrations applied (if any)
- [ ] Environment variables configured
- [ ] Secrets configured

### Validation
- [ ] Smoke tests passed
- [ ] API endpoints responsive
- [ ] Admin panel accessible
- [ ] Authentication working
- [ ] Email notifications working (if enabled)
- [ ] Logs showing no errors

### Post-Release
- [ ] Beta testers notified
- [ ] Monitoring dashboards reviewed
- [ ] Feedback channels monitored
- [ ] Documentation issues addressed
- [ ] Bug reports triaged

### Rollback Plan
- [ ] Previous version image ID recorded
- [ ] Rollback procedure documented
- [ ] Team briefed on rollback triggers
```

## Appendix: Beta Tester Selection Criteria

When selecting additional beta testers beyond the first (core team member):

**Ideal Beta Tester Profile:**
- ✅ Has a static website (blog, docs, marketing site)
- ✅ Active website with regular content updates
- ✅ Technical enough to follow deployment guide
- ✅ Willing to provide detailed feedback
- ✅ Available for regular check-ins
- ✅ Patient with beta quality software
- ✅ Understands "beta" expectations
- ✅ Represents different use case from existing testers

**Diversity Goals:**
- Different site types (blog, documentation, portfolio, etc.)
- Different authentication setups (Auth0, Firebase, custom, etc.)
- Different traffic levels (low, medium, high)
- Different deployment preferences (Docker, Cloud Run, VPS, etc.)
- Different technical skill levels

**Red Flags:**
- ❌ Needs production-grade reliability immediately
- ❌ Cannot tolerate breaking changes
- ❌ Requires extensive hand-holding
- ❌ Has extremely high traffic site (wait for v1.0)
- ❌ Wants custom features before providing feedback
- ❌ Not responsive to communication

## Appendix: Known Limitations for Beta

These limitations are acceptable for beta but should be addressed before v1.0:

**Functional Limitations:**
- No email verification for users (relies on external auth)
- No user profile management UI (admin panel only)
- No analytics dashboard (must query database directly)
- No export/import functionality for comments
- No bulk moderation operations
- No webhook support for external integrations
- No real-time updates (must refresh)

**Scalability Limitations:**
- SQLite may not scale to millions of comments (acceptable for beta)
- No horizontal scaling (single instance only)
- No CDN integration for static assets
- No caching layer (Redis, etc.)

**Deployment Limitations:**
- Cloud Run only (no Kubernetes, ECS, etc. guides)
- No automated backup solution
- No disaster recovery plan
- No multi-region deployment

**Documentation Limitations:**
- No video tutorials
- Limited troubleshooting guides
- Few integration examples
- No architecture diagrams

These limitations will be prioritized based on beta tester feedback and requirements for v1.0 production release.
