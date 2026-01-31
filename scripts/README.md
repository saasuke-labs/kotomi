# GitHub Issues Creation Script

This directory contains a script to automatically create GitHub issues based on the features documented in `Status.md`.

## Script: `create_issues.sh`

This script creates 12 comprehensive GitHub issues for all unimplemented features in Kotomi, based on the Status.md file.

### Prerequisites

1. **GitHub CLI (`gh`)** must be installed and authenticated:
   ```bash
   # Install gh CLI (if not already installed)
   # macOS
   brew install gh
   
   # Linux
   # See: https://github.com/cli/cli/blob/trunk/docs/install_linux.md
   
   # Authenticate
   gh auth login
   ```

2. **Repository Access**: You must have permission to create issues in the `saasuke-labs/kotomi` repository.

### Usage

Simply run the script from the repository root:

```bash
./scripts/create_issues.sh
```

Or from the scripts directory:

```bash
cd scripts
./create_issues.sh
```

### What Issues Are Created

The script creates the following 12 issues:

#### Blocking/Critical Priority Issues
1. **🚨 [BLOCKING] Implement CORS Configuration** - Critical for API to work with static sites
2. **🚨 [BLOCKING] Implement Rate Limiting** - Security issue, prevents spam/abuse
3. **🔒 Conduct Security Audit** - Required before production deployment

#### High Priority Issues
4. **⭐ Implement Reactions System** - Core feature from PRD
5. **🎨 Create Frontend Widget / JavaScript Embed** - Essential for user integration

#### Medium Priority Issues
6. **🔧 Implement API Versioning** - Important for API stability
7. **🤖 Implement Automatic/AI Moderation** - Reduces manual moderation burden
8. **👤 Implement User Authentication for Comments** - Allows users to edit/delete comments
9. **🔍 Improve Error Handling & Logging** - Better observability for production

#### Low Priority Issues
10. **📧 Implement Email Notifications** - Nice to have feature
11. **📊 Implement Analytics & Reporting** - Post-launch feature
12. **💾 Implement Export/Import Functionality** - Data portability feature

### Issue Features

Each issue includes:
- **Detailed description** of the feature and current state
- **Specific requirements** and implementation details
- **Step-by-step implementation guide**
- **Testing checklist** for validation
- **Files to create/modify** with exact paths
- **Priority level** and estimated effort
- **Dependencies** on other issues
- **⚠️ Reminder to update Status.md** when completed

### Dependencies Between Issues

The script creates issues with documented dependencies:
- Issue #3 (Reactions) should be done AFTER #1 (CORS) and #2 (Rate Limiting)
- Issue #4 (API Versioning) should be done BEFORE #5 (Frontend Widget)
- Issue #5 (Frontend Widget) depends on #1 (CORS)
- Issue #6 (Security Audit) should be done AFTER #1 and #2
- Issue #8 (User Auth) should be done AFTER #5 (Frontend Widget)
- Issue #9 (Email) can be done after #8 (User Auth)

### Recommended Implementation Order

1. **Phase 1 - Blocking Issues** (Do first):
   - Issue #1: CORS Configuration
   - Issue #2: Rate Limiting
   - Issue #6: Security Audit

2. **Phase 2 - High Priority** (Do next):
   - Issue #4: API Versioning
   - Issue #5: Frontend Widget
   - Issue #3: Reactions System

3. **Phase 3 - Medium Priority** (Post-launch):
   - Issue #12: Error Handling & Logging
   - Issue #7: AI Moderation
   - Issue #8: User Authentication

4. **Phase 4 - Low Priority** (Future enhancements):
   - Issue #9: Email Notifications
   - Issue #10: Analytics & Reporting
   - Issue #11: Export/Import

### Labels Used

The script applies the following labels to issues:
- `priority:critical` - Blocking issues for production
- `priority:high` - Important features
- `priority:medium` - Nice to have
- `priority:low` - Future enhancements
- `blocking` - Must fix before deployment
- `security` - Security-related issues
- `enhancement` - New features
- `feature` - Feature requests
- `frontend` - Frontend work
- `ai` - AI/ML related
- `observability` - Logging/monitoring

**Note**: These labels will be automatically created by GitHub if they don't exist.

### After Running the Script

1. Review the created issues on GitHub:
   ```bash
   gh issue list --repo saasuke-labs/kotomi
   ```

2. Prioritize and assign issues to team members or agents

3. Start with blocking issues (#1, #2, #6)

4. **Remember**: Update `Status.md` as each issue is completed

### Troubleshooting

#### "gh CLI is not authenticated"
Run `gh auth login` and follow the prompts to authenticate.

#### "Permission denied"
Make sure the script is executable:
```bash
chmod +x scripts/create_issues.sh
```

#### "Labels not found"
GitHub will automatically create labels that don't exist. You can customize label colors later in the GitHub UI.

### Re-running the Script

If you need to re-run the script, it will create duplicate issues. To avoid this:
1. Check existing issues first: `gh issue list --repo saasuke-labs/kotomi`
2. Close or delete duplicate issues manually
3. Or modify the script to check for existing issues before creating new ones

## Status.md Updates

**IMPORTANT**: Each issue includes a reminder to update `Status.md` when completed. This ensures the status document stays synchronized with the actual implementation progress.

When completing an issue:
1. Move the feature from "❌ Not Implemented" to "✅ Fully Implemented" in Status.md
2. Add implementation details and configuration requirements
3. Update the deployment readiness assessment
4. Update any relevant sections (testing, dependencies, etc.)
