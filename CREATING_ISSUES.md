# How to Create GitHub Issues from Status.md

## Overview

Based on Status.md, 12 comprehensive GitHub issues need to be created for unimplemented features in Kotomi. This document explains how to create them.

## Current Situation

The repository currently has **8 open issues** (mostly older infrastructure tasks). The new issues based on Status.md will add **12 additional issues** covering all unimplemented features.

## Option 1: Automated Script (Recommended)

### Prerequisites
1. GitHub CLI (`gh`) installed and authenticated
2. Write access to the `saasuke-labs/kotomi` repository

### Steps
```bash
# 1. Authenticate with GitHub
gh auth login

# 2. Run the script
./scripts/create_issues.sh
```

The script will create all 12 issues automatically with:
- Detailed descriptions
- Step-by-step implementation guides
- Testing checklists
- Dependencies documented
- Status.md update reminders

See `scripts/README.md` for detailed documentation.

## Option 2: Manual Creation

If you cannot run the script, you can create issues manually using the detailed specifications in `ISSUES_SUMMARY.md`.

For each issue:
1. Go to https://github.com/saasuke-labs/kotomi/issues/new
2. Copy the title from ISSUES_SUMMARY.md
3. Copy the full description from `scripts/create_issues.sh`
4. Add appropriate labels (priority, type)
5. Create the issue

## Option 3: GitHub API

Use the GitHub REST API directly:

```bash
# Set your GitHub token
export GITHUB_TOKEN="your_token_here"

# Create an issue (example for Issue #1)
curl -X POST \
  -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/saasuke-labs/kotomi/issues \
  -d @issue_data.json
```

Where `issue_data.json` contains the issue data.

## What Will Be Created

### 12 New Issues:

1. **🚨 [BLOCKING] Implement CORS Configuration** (Critical)
2. **🚨 [BLOCKING] Implement Rate Limiting** (Critical, Security)
3. **⭐ Implement Reactions System** (High Priority)
4. **🔧 Implement API Versioning** (Medium Priority)
5. **🎨 Create Frontend Widget / JavaScript Embed** (High Priority)
6. **🔒 Conduct Security Audit** (Critical, Blocking)
7. **🤖 Implement Automatic/AI Moderation** (Medium Priority)
8. **👤 Implement User Authentication for Comments** (Medium Priority)
9. **📧 Implement Email Notifications** (Low Priority)
10. **📊 Implement Analytics & Reporting** (Low Priority)
11. **💾 Implement Export/Import Functionality** (Low Priority)
12. **🔍 Improve Error Handling & Logging** (Medium Priority)

See `ISSUES_SUMMARY.md` for complete details on each issue.

## Labels That Will Be Applied

- `priority:critical` - Issues #1, #2, #6
- `priority:high` - Issues #3, #5
- `priority:medium` - Issues #4, #7, #8, #12
- `priority:low` - Issues #9, #10, #11
- `blocking` - Issues #1, #2, #6
- `security` - Issues #2, #6
- `enhancement` - Most issues
- `feature` - Most issues
- `frontend` - Issue #5
- `ai` - Issue #7
- `observability` - Issue #12

These labels will be created automatically by GitHub if they don't exist.

## After Creating Issues

1. **Review**: Check all issues are created correctly
2. **Prioritize**: Start with blocking issues (#1, #2, #6)
3. **Assign**: Assign to team members or agents
4. **Track**: Use project boards to track progress
5. **Update Status.md**: As each issue is completed, update Status.md

## Verification

After running the script, verify issues were created:

```bash
# List all issues
gh issue list --repo saasuke-labs/kotomi

# Filter by label
gh issue list --repo saasuke-labs/kotomi --label "priority:critical"
```

## For Agent Implementation

Each issue is designed to be implemented by agents with:
- Clear requirements and acceptance criteria
- Step-by-step implementation guides
- Specific files to create/modify
- Testing checklists
- Dependency information

Agents should:
1. Read the full issue description
2. Follow the implementation steps
3. Run all tests
4. Update Status.md as instructed
5. Close the issue when complete

## Need Help?

- **Script Issues**: See `scripts/README.md`
- **Issue Details**: See `ISSUES_SUMMARY.md`
- **Status Reference**: See `Status.md`

## Notes

- The script is idempotent but will create duplicates if run multiple times
- Check for existing issues before running: `gh issue list`
- Close or delete duplicates if needed
- Labels can be customized after creation in the GitHub UI
