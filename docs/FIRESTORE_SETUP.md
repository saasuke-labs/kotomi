# Firestore Database Setup

This document explains how to configure Kotomi to use Google Cloud Firestore instead of SQLite.

## Overview

Kotomi supports two database backends:
- **SQLite** (default) - For local development, CI, and small deployments
- **Firestore** - For production deployments with high scalability needs

## Why Firestore?

Firestore provides several advantages for production deployments:
- **Scalability**: Automatically scales to handle millions of documents
- **Real-time**: Built-in real-time synchronization capabilities
- **Serverless**: No database server to manage or maintain
- **Global distribution**: Automatically replicated across regions
- **High availability**: 99.999% uptime SLA

## Firestore Setup

### 1. Create a Firestore Database

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Select or create a project
3. Navigate to Firestore
4. Click "Create Database"
5. Choose "Firestore Native Mode"
6. Select a location (choose one close to your users)

### 2. Deploy Firestore Indexes

Firestore requires composite indexes for complex queries. Deploy the indexes:

```bash
# Install the Firebase CLI if you haven't already
npm install -g firebase-tools

# Login to Firebase
firebase login

# Initialize Firebase in your project directory (if not already done)
firebase init firestore

# Deploy the indexes
firebase deploy --only firestore:indexes
```

The `firestore.indexes.json` file in the repository root defines the required indexes:
- Comments by site + page + creation date
- Comments by site + status + creation date
- Comments by author + creation date

### 3. Configure Authentication

For applications running on Google Cloud (Cloud Run, GKE, etc.), authentication is automatic through Application Default Credentials.

For local development or other environments:

```bash
# Set up Application Default Credentials
gcloud auth application-default login

# Or use a service account key
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json
```

### 4. Configure Environment Variables

Set the following environment variables:

```bash
# Use Firestore as the database provider
export DB_PROVIDER=firestore

# Set your Google Cloud project ID
export FIRESTORE_PROJECT_ID=your-project-id
# OR
export GCP_PROJECT=your-project-id
```

For SQLite (default, no configuration needed):
```bash
export DB_PROVIDER=sqlite  # or omit, SQLite is default
export DB_PATH=./kotomi.db  # optional, defaults to ./kotomi.db
```

## Performance Optimizations

The Firestore implementation includes several optimizations:

### 1. Composite Indexes
Pre-configured indexes for common query patterns ensure fast retrieval:
- Listing comments by page
- Filtering comments by status
- Sorting by creation date

### 2. Document Structure
Comments are stored with flattened structure for efficient queries:
```
comments/{commentId}
  ├─ id: string
  ├─ site_id: string
  ├─ page_id: string
  ├─ author: string
  ├─ text: string
  ├─ status: string
  ├─ created_at: timestamp
  └─ ...
```

### 3. Batch Operations
Future updates will include batch writes for bulk operations to reduce latency.

### 4. Caching Strategy
Consider implementing a caching layer (Redis, Memcached) for frequently accessed data:
- Popular page comments
- User statistics
- Site configurations

## Limitations with Firestore

Some features require SQL database (not available with Firestore):
- **Moderation Configuration**: Stores moderation settings
- **Notification Queue**: Email notification queue processing
- **Analytics**: Complex analytics queries

These features will be gracefully disabled when using Firestore. Future updates may add Firestore-native implementations.

## Cost Considerations

Firestore pricing is based on:
- **Document reads**: $0.06 per 100,000 reads
- **Document writes**: $0.18 per 100,000 writes
- **Document deletes**: $0.02 per 100,000 deletes
- **Storage**: $0.18 per GB/month

Free tier includes:
- 50,000 document reads per day
- 20,000 document writes per day
- 20,000 document deletes per day
- 1 GB storage

For most small to medium deployments, Firestore will fall within the free tier.

## Migration from SQLite to Firestore

To migrate existing data:

1. Export data from SQLite:
```bash
# Use the export API endpoint
curl http://localhost:8080/api/admin/export > export.json
```

2. Import to Firestore:
```bash
# Use the import API endpoint with Firestore configured
export DB_PROVIDER=firestore
export FIRESTORE_PROJECT_ID=your-project-id
curl -X POST http://localhost:8080/api/admin/import -d @export.json
```

## Testing

To test Firestore locally, use the Firestore Emulator:

```bash
# Install the emulator
gcloud components install cloud-firestore-emulator

# Start the emulator
gcloud beta emulators firestore start

# In another terminal, set the emulator environment variable
export FIRESTORE_EMULATOR_HOST=localhost:8080

# Run your application
export DB_PROVIDER=firestore
export FIRESTORE_PROJECT_ID=test-project
go run cmd/main.go
```

## CI/CD Configuration

For CI/CD pipelines, keep using SQLite as it's faster and doesn't require external services:

```yaml
# GitHub Actions example
env:
  DB_PROVIDER: sqlite
  DB_PATH: ./test.db
```

For production deployments, use Firestore:

```yaml
# Cloud Run deployment example
env:
  DB_PROVIDER: firestore
  FIRESTORE_PROJECT_ID: your-project-id
```

## Monitoring

Monitor Firestore usage in the Google Cloud Console:
- Dashboard → Firestore → Usage tab
- View read/write operations
- Monitor query performance
- Set up billing alerts

## Troubleshooting

### "Missing index" errors
Deploy the indexes using `firebase deploy --only firestore:indexes`

### Authentication errors
Ensure Application Default Credentials are set up correctly:
```bash
gcloud auth application-default login
```

### Performance issues
1. Check if indexes are deployed
2. Review query patterns in Firestore Console
3. Consider adding caching layer
4. Monitor Firestore metrics in Cloud Console

## Support

For issues or questions:
- GitHub Issues: https://github.com/saasuke-labs/kotomi/issues
- Documentation: https://github.com/saasuke-labs/kotomi/docs
