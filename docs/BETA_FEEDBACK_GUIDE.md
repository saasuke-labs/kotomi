# Beta Feedback Collection Guide

This guide explains how to collect, organize, and act on feedback from Kotomi beta testers.

## Feedback Channels

### 1. GitHub Issues (Bug Reports)
**Purpose**: Track and resolve bugs, errors, and technical issues  
**URL**: https://github.com/saasuke-labs/kotomi/issues

**When to use**:
- Application crashes or errors
- Features not working as documented
- Performance issues
- Security vulnerabilities
- Documentation inaccuracies

**Response Time**: 24-48 hours for initial acknowledgment

### 2. GitHub Discussions (Feature Requests & Questions)
**Purpose**: Discuss feature ideas, ask questions, share experiences  
**URL**: https://github.com/saasuke-labs/kotomi/discussions

**Categories**:
- 💡 **Ideas** - Feature requests and improvements
- 🙏 **Q&A** - Questions about usage, configuration, integration
- 📣 **Show and Tell** - Share your Kotomi implementation
- 💬 **General** - General discussion about Kotomi

**Response Time**: 2-3 days for questions, ongoing for discussions

### 3. Weekly Check-in (Direct Communication)
**Purpose**: Proactive feedback collection and relationship building  
**Format**: Email or video call

**Schedule**:
- First 2 weeks: Weekly 30-minute calls
- Weeks 3-4: Weekly email check-in
- Month 2+: Bi-weekly email check-in

## Issue Templates

### Bug Report Template

Copy this template to `.github/ISSUE_TEMPLATE/bug_report.md`:

```markdown
---
name: Bug Report
about: Report a bug or issue you've encountered
title: '[BUG] '
labels: bug, needs-triage
assignees: ''
---

**Beta Tester Information**
- Name: 
- Site URL: 
- Kotomi Version: (e.g., 0.1.0-beta.1)
- Deployment Method: (Docker / Cloud Run / Binary)

**Describe the Bug**
A clear and concise description of what the bug is.

**Steps to Reproduce**
1. Go to '...'
2. Click on '...'
3. See error

**Expected Behavior**
What you expected to happen.

**Actual Behavior**
What actually happened.

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Error Messages**
```
Paste any error messages or logs here
```

**Environment**
- OS: (e.g., Ubuntu 22.04, macOS 14, Windows 11)
- Browser: (if relevant, e.g., Chrome 120)
- Database Size: (approximate number of comments)

**Additional Context**
Add any other context about the problem here.

**Impact**
- [ ] Blocking - Cannot use Kotomi
- [ ] High - Major feature broken
- [ ] Medium - Workaround exists
- [ ] Low - Minor inconvenience
```

### Feature Request Template

Copy this template to `.github/ISSUE_TEMPLATE/feature_request.md`:

```markdown
---
name: Feature Request
about: Suggest a feature or improvement
title: '[FEATURE] '
labels: enhancement, needs-triage
assignees: ''
---

**Beta Tester Information**
- Name: 
- Use Case: (Blog / Documentation / Other)

**Problem Statement**
What problem are you trying to solve? What limitation are you encountering?

**Proposed Solution**
How would you like this to work? Describe your ideal solution.

**Alternative Solutions**
Have you considered any alternative solutions or workarounds?

**Use Case Details**
- Who would benefit from this feature?
- How often would you use this feature?
- Is this blocking you from using Kotomi?

**Priority**
- [ ] Must have - Critical for our use case
- [ ] Should have - Would significantly improve experience
- [ ] Nice to have - Would be helpful but not essential

**Additional Context**
Screenshots, mockups, examples from other tools, etc.
```

## Weekly Check-in Template

### Email Template

```
Subject: Kotomi Beta - Week [X] Check-in

Hi [Beta Tester Name],

Hope you're having a great week! This is your weekly Kotomi beta check-in.

Quick questions:

1. **Usage**: How's Kotomi performing this week?
   - Any issues or concerns?
   - Everything working as expected?

2. **Feedback**: Anything you'd like to see improved?
   - Missing features?
   - Confusing documentation?
   - Performance concerns?

3. **Blockers**: Anything preventing you from using Kotomi?

4. **Support**: Any questions or need help with anything?

Recent Updates:
[List any recent fixes or improvements relevant to this tester]

Reply to this email with your thoughts, or feel free to:
- Report bugs: https://github.com/saasuke-labs/kotomi/issues
- Request features: https://github.com/saasuke-labs/kotomi/discussions

Thanks for being a beta tester!

Best regards,
[Your Name]
Kotomi Team
```

### Video Call Agenda Template

```markdown
# Kotomi Beta Check-in Call - Week [X]

**Beta Tester**: [Name]
**Date**: [Date]
**Duration**: 30 minutes

## Agenda

1. **How's it Going?** (5 min)
   - Overall experience this week
   - Any wins or successes?
   - Any frustrations?

2. **Technical Review** (10 min)
   - Check deployment health
   - Review any errors or issues
   - Discuss performance
   - Review comment volume and activity

3. **Feature Discussion** (10 min)
   - What's working well?
   - What's missing or could be better?
   - Prioritize feature requests

4. **Next Steps** (5 min)
   - Action items for team
   - Action items for beta tester
   - Schedule next check-in

## Notes
[Take notes during call]

## Action Items
- [ ] [Action item 1]
- [ ] [Action item 2]

## Follow-up
[Send summary email within 24 hours]
```

## Feedback Organization

### Labels for Issues

**Priority**:
- `priority: critical` - Blocking bug, immediate attention needed
- `priority: high` - Important bug or highly requested feature
- `priority: medium` - Should address soon
- `priority: low` - Nice to have

**Type**:
- `bug` - Something is broken
- `enhancement` - New feature or improvement
- `documentation` - Documentation issues
- `question` - Question about usage

**Status**:
- `needs-triage` - Needs review and prioritization
- `confirmed` - Bug confirmed, reproducible
- `in-progress` - Currently being worked on
- `needs-info` - Waiting for more information
- `blocked` - Blocked by external factor

**Source**:
- `beta-tester` - Reported by beta tester
- `internal` - Reported by team

### Feedback Tracking Spreadsheet

Create a spreadsheet to track all feedback:

| Date | Tester | Type | Summary | Priority | Status | GitHub Link | Notes |
|------|--------|------|---------|----------|--------|-------------|-------|
| 2026-02-05 | Alice | Bug | Comments not saving | High | Fixed | #123 | Fixed in beta.2 |
| 2026-02-06 | Bob | Feature | Bulk moderation | Medium | Planned | #124 | For v1.0 |

## Feedback Analysis

### Weekly Review

Every Friday:
1. **Review all new feedback** from the week
2. **Triage and prioritize** issues
3. **Identify patterns** - Multiple testers reporting same issue?
4. **Plan fixes** for next release
5. **Update testers** on progress

### Monthly Review

End of each month:
1. **Aggregate feedback** across all testers
2. **Identify top issues**:
   - Most frequently reported bugs
   - Most requested features
   - Common confusion points
3. **Assess beta health**:
   - Tester satisfaction levels
   - Active vs. inactive testers
   - Deployment success rate
4. **Plan roadmap** adjustments
5. **Update documentation** based on common questions

## Acting on Feedback

### Bug Fixes

**High Priority Bugs** (Blocking, data loss, security):
1. Reproduce and confirm
2. Create hotfix branch
3. Fix and test
4. Release patch version (e.g., 0.1.0-beta.2)
5. Notify affected testers
6. Document in CHANGELOG

**Lower Priority Bugs**:
1. Add to backlog
2. Include in next scheduled release
3. Document workaround if available

### Feature Requests

**Evaluation Criteria**:
1. **Impact**: How many users would benefit?
2. **Effort**: How complex to implement?
3. **Alignment**: Fits Kotomi's vision and roadmap?
4. **Urgency**: Blocking beta testers?

**Response Options**:
1. **Accept - High Priority**: Add to next release
2. **Accept - Backlog**: Good idea, plan for later
3. **Accept - v1.0**: Wait until after beta
4. **Decline**: Doesn't fit vision, explain why
5. **Needs Discussion**: Need more info or community input

### Documentation Improvements

If testers ask the same questions repeatedly:
1. **Identify gaps** in documentation
2. **Update docs** to address confusion
3. **Add FAQ entries**
4. **Create tutorials** or examples
5. **Notify testers** of updates

## Response Templates

### Acknowledging Bug Report

```markdown
Thanks for reporting this, @username! 

I've confirmed this is a bug. We're prioritizing it as [priority] and 
will include a fix in the next release.

**Workaround**: [If available]

I'll keep you updated on progress here.
```

### Accepting Feature Request

```markdown
Great idea, @username! This aligns well with our roadmap.

We're planning to implement this [timeframe]. I'll add it to our backlog
and keep you updated on progress.

Would love to hear more about your use case in the meantime!
```

### Declining Feature Request

```markdown
Thanks for the suggestion, @username!

After discussion with the team, we've decided not to pursue this because
[reason]. However, we appreciate you sharing your perspective.

Have you considered [alternative approach]? That might address your 
use case.
```

## Success Metrics

Track these metrics to assess feedback process health:

**Response Metrics**:
- Average time to first response: Target < 48 hours
- Average time to resolution: Target < 1 week for bugs

**Satisfaction Metrics**:
- % of testers actively engaged: Target > 80%
- % of testers satisfied with support: Target > 90%
- Net Promoter Score: Target > 7/10

**Quality Metrics**:
- Number of critical bugs: Target = 0
- Number of repeat issues: Target decreasing
- Documentation improvement PRs: Target increasing

## Resources

- [Beta Onboarding Checklist](BETA_ONBOARDING_CHECKLIST.md)
- [Beta Tester Guide](BETA_TESTER_GUIDE.md)
- [Support Plan](BETA_SUPPORT_PLAN.md)
- [Release Process](RELEASE_PROCESS.md)

## Tips for Effective Feedback Collection

1. **Be Proactive**: Don't wait for feedback - ask for it
2. **Be Responsive**: Quick responses build trust
3. **Be Transparent**: Share roadmap and priorities
4. **Be Grateful**: Always thank testers for their time
5. **Be Honest**: If something won't be fixed soon, say so
6. **Close the Loop**: Always follow up with resolution

## Contact

For questions about feedback collection process:
- [Add team contact info]
