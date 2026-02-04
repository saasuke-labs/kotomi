# Deployment Monitoring Guide

This guide covers how to monitor Kotomi deployments during the beta release phase, including what metrics to track, how to access logs, and what to watch for.

## Overview

Effective monitoring during beta release helps us:
- Detect issues before beta testers report them
- Understand usage patterns and performance
- Make data-driven decisions about scaling and optimization
- Provide better support to beta testers

## Monitoring Checklist

### Daily Checks (First 2 Weeks)
- [ ] Review error logs for new exceptions
- [ ] Check API response times
- [ ] Monitor database size and growth
- [ ] Review resource utilization (CPU, memory)
- [ ] Check for failed deployments
- [ ] Review security logs for suspicious activity

### Weekly Checks (Ongoing)
- [ ] Analyze usage trends (comments, reactions, page views)
- [ ] Review beta tester activity levels
- [ ] Check for degraded performance
- [ ] Assess database query performance
- [ ] Review rate limiting effectiveness
- [ ] Update monitoring dashboards

### Monthly Checks
- [ ] Comprehensive performance review
- [ ] Database optimization assessment
- [ ] Infrastructure cost review
- [ ] Capacity planning for growth
- [ ] Security audit
- [ ] Monitoring strategy refinement

## Cloud Run Monitoring

### Accessing Cloud Run Metrics

**Via Google Cloud Console**:
1. Navigate to Cloud Run: https://console.cloud.google.com/run
2. Select service: `kotomi-prod`
3. Click "METRICS" tab

**Key Metrics**:
- **Request count**: Total requests per time period
- **Request latency**: 50th, 95th, 99th percentile
- **Container instance count**: Min/max/average
- **CPU utilization**: Percentage of allocated CPU
- **Memory utilization**: MB used / MB allocated
- **Billable time**: Container-seconds (cost metric)

**Via gcloud CLI**:
```bash
# List services
gcloud run services list --platform managed --region us-central1

# Describe service
gcloud run services describe kotomi-prod --platform managed --region us-central1

# Get service URL
gcloud run services describe kotomi-prod --platform managed --region us-central1 --format='value(status.url)'
```

### Setting Up Alerts

**Create Alert Policy** (via Cloud Console):
1. Go to Monitoring > Alerting
2. Create Policy
3. Add conditions:

**Recommended Alerts**:

1. **High Error Rate**
   - Metric: Request count (status 5xx)
   - Condition: > 10 errors in 5 minutes
   - Notification: Email + Slack

2. **High Latency**
   - Metric: Request latency (95th percentile)
   - Condition: > 2 seconds for 5 minutes
   - Notification: Email

3. **High Memory Usage**
   - Metric: Memory utilization
   - Condition: > 80% for 10 minutes
   - Notification: Email

4. **Service Down**
   - Metric: Request count
   - Condition: 0 requests for 5 minutes (during business hours)
   - Notification: Email + SMS

### Cloud Run Logs

**Accessing Logs** (via Cloud Console):
1. Navigate to Cloud Run service
2. Click "LOGS" tab
3. Use filter to narrow down:
   ```
   resource.type="cloud_run_revision"
   resource.labels.service_name="kotomi-prod"
   ```

**Via gcloud CLI**:
```bash
# Stream logs
gcloud run services logs read kotomi-prod --platform managed --region us-central1 --tail

# Filter logs
gcloud run services logs read kotomi-prod --platform managed --region us-central1 --filter="severity>=ERROR" --limit=50

# Export logs to file
gcloud run services logs read kotomi-prod --platform managed --region us-central1 > logs.txt
```

**Log Levels to Monitor**:
- **ERROR**: Application errors requiring investigation
- **WARN**: Potential issues or degraded functionality
- **INFO**: Normal operations (comment posted, user login, etc.)
- **DEBUG**: Detailed debugging info (only enable when troubleshooting)

### Common Log Patterns

**Successful Operations**:
```
INFO: Comment created: id=abc-123 site_id=xyz-789
INFO: User authenticated: user_id=user-456
INFO: Health check: OK
```

**Errors to Watch For**:
```
ERROR: Database connection failed: unable to open database
ERROR: JWT validation failed: invalid signature
ERROR: OpenAI API error: rate limit exceeded
ERROR: Auth0 authentication failed: invalid token
```

**Performance Warnings**:
```
WARN: Slow query detected: 1.2s for SELECT comments
WARN: Database size exceeds 100MB
WARN: Rate limit triggered for IP: 1.2.3.4
```

## Application Metrics

### Health Check Monitoring

**Endpoint**: `GET /healthz`

**Expected Response**:
```json
{
  "status": "ok",
  "database": "connected",
  "version": "0.1.0-beta.1"
}
```

**Monitoring Strategy**:
```bash
# Simple uptime check (run via cron every 5 minutes)
#!/bin/bash
HEALTH_URL="https://kotomi-prod-xyz.run.app/healthz"
response=$(curl -s -o /dev/null -w "%{http_code}" "$HEALTH_URL")

if [ "$response" != "200" ]; then
    echo "ALERT: Health check failed with status $response"
    # Send notification
fi
```

**Using External Monitoring** (recommended):
- **UptimeRobot**: Free tier, 5-minute checks
- **Pingdom**: Comprehensive monitoring
- **Better Uptime**: Developer-friendly

### API Response Time Tracking

**Using curl**:
```bash
# Measure response time
curl -w "@curl-format.txt" -o /dev/null -s "https://kotomi-prod-xyz.run.app/api/sites"

# curl-format.txt content:
time_namelookup:  %{time_namelookup}s\n
time_connect:     %{time_connect}s\n
time_starttransfer: %{time_starttransfer}s\n
time_total:       %{time_total}s\n
```

**Response Time Targets**:
- **Health check**: < 100ms
- **List endpoints** (GET /api/sites): < 500ms
- **Detail endpoints** (GET /api/sites/{id}): < 300ms
- **Create operations** (POST comments): < 1000ms
- **Admin panel pages**: < 2000ms

**What to Do if Slow**:
1. Check database query performance
2. Review logs for slow queries
3. Consider adding database indexes
4. Evaluate caching opportunities
5. Check resource allocation (CPU/Memory)

## Database Monitoring

### Database Size Tracking

**Check SQLite database size**:
```bash
# If deployed with volume
docker exec -it kotomi ls -lh /app/data/kotomi.db

# On Cloud Run (via shell)
ls -lh /app/data/kotomi.db

# Get detailed info
sqlite3 /app/data/kotomi.db "
SELECT 
    name,
    (SELECT COUNT(*) FROM comments) as comment_count,
    (SELECT COUNT(*) FROM reactions) as reaction_count,
    (SELECT COUNT(*) FROM sites) as site_count,
    (SELECT COUNT(*) FROM pages) as page_count;
"
```

**Growth Monitoring**:
- Track daily size increase
- Estimate time to 1GB (SQLite recommended limit)
- Plan migration to PostgreSQL if needed

**Size Alerts**:
- **100 MB**: Monitor closely
- **500 MB**: Consider optimization
- **1 GB**: Plan migration

### Query Performance

**Enable SQLite profiling** (development only):
```sql
-- In SQLite shell
.timer on
.explain on
```

**Common slow queries to watch**:
1. Comments without indexes
2. Large LIKE queries
3. Unoptimized JOINs
4. Full table scans

**Query optimization checklist**:
- [ ] All foreign keys have indexes
- [ ] WHERE clauses use indexed columns
- [ ] LIMIT clauses prevent large result sets
- [ ] Proper use of prepared statements

### Database Backups

**Backup verification**:
```bash
# Check backup exists
ls -lh /path/to/backups/kotomi-backup-*.db

# Verify backup integrity
sqlite3 /path/to/backups/kotomi-backup-latest.db "PRAGMA integrity_check;"

# Expected output: "ok"
```

**Backup schedule** (recommended):
- **Automated**: Daily at 2 AM UTC
- **Manual**: Before major updates
- **Retention**: Keep 7 daily, 4 weekly, 3 monthly

See [Database Backup & Restore Guide](DATABASE_BACKUP_RESTORE.md) for details.

## Security Monitoring

### What to Monitor

**Failed Authentication Attempts**:
```bash
# Check logs for auth failures
gcloud run services logs read kotomi-prod --filter="severity>=WARN AND textPayload=~'authentication failed'"

# Watch for patterns:
# - Multiple failures from same IP
# - Failures from unusual locations
# - Spike in failures after deployment
```

**Suspicious Activity**:
- Unusual traffic patterns
- High rate of 4xx errors
- Large number of requests from single IP
- Requests to non-existent endpoints (scanning)

**Rate Limiting**:
```bash
# Check rate limit violations
gcloud run services logs read kotomi-prod --filter="textPayload=~'rate limit'"

# Assess if limits are too strict or too lenient
```

**Data Integrity**:
- Monitor for unusual comment content (potential injection attempts)
- Watch for mass deletions or modifications
- Check for privilege escalation attempts

### Security Alert Actions

**If suspicious activity detected**:
1. **Document**: Capture logs and evidence
2. **Assess**: Is it an attack or legitimate traffic?
3. **Block**: Add IP to blocklist if malicious
4. **Investigate**: Review recent code changes
5. **Patch**: Fix vulnerability if found
6. **Notify**: Inform affected beta testers
7. **Report**: File security incident report

## Usage Analytics

### Key Metrics to Track

**User Engagement**:
- Comments per day/week
- Unique commenters
- Pages with comments
- Average comments per page
- Reactions per day/week

**Beta Tester Activity**:
- Active sites
- Comments per site
- Last activity date per tester
- Feature adoption (AI moderation, email notifications)

**API Usage**:
- Requests per endpoint
- Most used endpoints
- Error rates by endpoint
- Authentication success rate

### Extracting Analytics

**Using SQLite queries**:
```sql
-- Comments per day (last 7 days)
SELECT 
    DATE(created_at) as date,
    COUNT(*) as comment_count
FROM comments
WHERE created_at >= datetime('now', '-7 days')
GROUP BY DATE(created_at)
ORDER BY date DESC;

-- Active sites
SELECT 
    s.name,
    s.url,
    COUNT(c.id) as comment_count,
    MAX(c.created_at) as last_comment
FROM sites s
LEFT JOIN pages p ON p.site_id = s.id
LEFT JOIN comments c ON c.page_id = p.id
GROUP BY s.id
ORDER BY comment_count DESC;

-- Top pages by comments
SELECT 
    s.name as site_name,
    p.slug,
    p.url,
    COUNT(c.id) as comment_count
FROM pages p
JOIN sites s ON s.id = p.site_id
LEFT JOIN comments c ON c.page_id = p.id
GROUP BY p.id
ORDER BY comment_count DESC
LIMIT 10;
```

**Recommended Dashboard** (manual tracking):
- Spreadsheet or simple dashboard
- Update weekly with key metrics
- Track trends over time
- Share with team and stakeholders

## Performance Benchmarks

### Beta Release Targets

**Response Times** (95th percentile):
- Health check: < 100ms ✓ Target
- List APIs: < 500ms ✓ Target
- Create APIs: < 1000ms ✓ Target
- Admin pages: < 2000ms ⚠️ Acceptable

**Resource Usage**:
- Memory: < 512 MB average ✓ Efficient
- CPU: < 50% average ✓ Efficient
- Database: < 100 MB ✓ Small scale

**Availability**:
- Uptime: > 99% (7.2 hours downtime/month)
- Health check success: > 99.9%
- Error rate: < 1% of requests

**Capacity**:
- Concurrent users: 10-50 (beta scale)
- Comments per day: 100-1000 (beta scale)
- Request throughput: 10-100 req/sec (beta scale)

## Troubleshooting Common Issues

### High Error Rate

**Symptoms**: Spike in 5xx errors

**Investigation**:
1. Check logs for error messages
2. Verify database connection
3. Check resource exhaustion
4. Review recent deployments
5. Test health endpoint

**Common Causes**:
- Database locked (too many concurrent writes)
- Memory exhausted (OOM)
- Bad deployment (code bug)
- External service down (OpenAI, Auth0)

### High Latency

**Symptoms**: Slow API responses

**Investigation**:
1. Check database query times
2. Review Cloud Run CPU/memory
3. Check external API calls (OpenAI)
4. Look for slow endpoints
5. Review concurrent request count

**Common Causes**:
- Unoptimized database queries
- Undersized containers
- External API latency
- High traffic spike

### Memory Issues

**Symptoms**: Container restarts, OOM errors

**Investigation**:
1. Check memory usage trends
2. Review database connection pooling
3. Look for memory leaks
4. Check query result size

**Common Causes**:
- Database connection leak
- Large result sets loaded into memory
- Inefficient query pagination

## Tools & Resources

### Recommended Tools

**Monitoring**:
- Google Cloud Console (built-in)
- Cloud Monitoring & Logging (Google)
- UptimeRobot (external uptime monitoring)

**Log Analysis**:
- gcloud CLI (command-line)
- Cloud Logging (web interface)
- Logz.io (if advanced analysis needed)

**Performance**:
- curl (response time testing)
- Apache Bench (load testing)
- Google Cloud Trace (distributed tracing)

**Database**:
- SQLite CLI (query analysis)
- DB Browser for SQLite (GUI)

### Internal Resources

- [Beta Tester Guide](BETA_TESTER_GUIDE.md)
- [Beta Support Plan](BETA_SUPPORT_PLAN.md)
- [Database Backup & Restore](DATABASE_BACKUP_RESTORE.md)
- [Release Process](RELEASE_PROCESS.md)

## Contact & Escalation

**Monitoring Issues**:
- If alert fires: Acknowledge and investigate
- If critical: Escalate to on-call engineer
- If security: Contact security team immediately

**Questions**: [Add team contact]

---

**Last Updated**: 2026-02-04  
**Version**: 1.0  
**Review**: Weekly during beta, monthly after
