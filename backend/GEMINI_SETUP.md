# Google Gemini File Search Setup Guide

This guide explains how to set up Google Gemini File Search for OrgMind's AI-powered chat feature.

## Overview

OrgMind uses Google Gemini's File Search capability to provide intelligent, context-aware responses based on your document content. The system uses a **shared store architecture** with metadata-based filtering:

1. Creates a **single shared File Search store** at application startup
2. Uploads documents with custom metadata (graph_id, domain, version)
3. Filters queries using metadata to ensure graph isolation
4. Uses Gemini API to generate AI responses with document context
5. Streams responses in real-time via Server-Sent Events (SSE)

### Shared Store Architecture

Instead of creating separate stores for each graph, OrgMind uses:
- **One shared File Search store** for all documents across all graphs
- **Custom metadata** attached to each document:
  - `graph_id`: Identifies which graph the document belongs to
  - `domain`: Tenant identifier (currently "topeic.com")
  - `version`: Schema version for future migrations (currently "1.1")
- **Metadata filtering** at query time to ensure users only see their graph's documents

**Benefits:**
- Reduced API overhead (one store vs. N stores)
- Simpler store lifecycle management
- Faster graph creation (no store creation needed)
- Better scalability for multi-tenant deployments

## Prerequisites

- Google Cloud account
- Google Cloud project with billing enabled
- Gemini API access

## Setup Steps

### 1. Get Your Gemini API Key

1. Visit [Google AI Studio](https://aistudio.google.com/app/apikey)
2. Sign in with your Google account
3. Click "Create API Key"
4. Select your Google Cloud project (or create a new one)
5. Copy the generated API key

**Important:** Keep your API key secure and never commit it to version control.

### 2. Enable Gemini API in Your Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Select your project
3. Navigate to "APIs & Services" > "Library"
4. Search for "Generative Language API"
5. Click "Enable"

### 3. Get Your Project ID

1. In [Google Cloud Console](https://console.cloud.google.com/)
2. Your project ID is shown in the project selector dropdown
3. Format: lowercase letters, numbers, and hyphens (e.g., `my-project-123`)

### 4. Choose Your Region

Select a region close to your users for better latency:

- **US Regions:** `us-central1`, `us-east1`, `us-west1`, `us-west4`
- **Europe Regions:** `europe-west1`, `europe-west2`, `europe-west4`
- **Asia Regions:** `asia-northeast1`, `asia-southeast1`

**Default:** `us-central1`

### 5. Configure Environment Variables

Add the following to your `backend/.env` file:

```bash
# Google Gemini API key
GEMINI_API_KEY=your-actual-api-key-here

# Google Cloud project ID
GEMINI_PROJECT_ID=your-project-id

# Google Cloud region
GEMINI_LOCATION=us-central1

# Optional: Model selection (default: gemini-1.5-flash)
GEMINI_MODEL=gemini-1.5-flash

# Optional: Retry configuration (default: 3)
GEMINI_MAX_RETRIES=3

# Optional: Timeout in seconds (default: 60)
GEMINI_TIMEOUT_SECONDS=60

# Optional: Shared File Search store name (default: OrgMind Documents)
GEMINI_STORE_NAME=OrgMind Documents
```

## Environment Variables Reference

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `GEMINI_API_KEY` | Your Gemini API key from AI Studio | `AIzaSyD...` |
| `GEMINI_PROJECT_ID` | Your Google Cloud project ID | `my-project-123` |
| `GEMINI_LOCATION` | Google Cloud region for API calls | `us-central1` |

### Optional Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GEMINI_MODEL` | `gemini-1.5-flash` | Model to use (`gemini-1.5-flash` or `gemini-1.5-pro`) |
| `GEMINI_MAX_RETRIES` | `3` | Number of retry attempts for failed API calls |
| `GEMINI_TIMEOUT_SECONDS` | `60` | Request timeout in seconds |
| `GEMINI_STORE_NAME` | `OrgMind Documents` | Display name for the shared File Search store |

## Model Selection

### gemini-1.5-flash (Recommended)
- **Speed:** Fast response times
- **Cost:** Lower cost per request
- **Use Case:** Most chat interactions
- **File Search:** Fully supported

### gemini-1.5-pro
- **Speed:** Slower but more thorough
- **Cost:** Higher cost per request
- **Use Case:** Complex queries requiring deeper analysis
- **File Search:** Fully supported

## How It Works

### 1. Application Startup
When the backend starts:
- A single shared File Search store is created (or retrieved if it exists)
- The store ID is saved in application configuration
- Store name comes from `GEMINI_STORE_NAME` environment variable
- If store creation fails, the application exits with an error

### 2. Graph Creation
When a user creates a knowledge graph:
- **No File Search store is created** (uses the shared store)
- Graph is ready immediately for document uploads
- Much faster than the old per-graph store approach

### 3. Document Upload
When a document is uploaded to a graph:
- Document content is extracted as plain text
- Content is uploaded to the **shared File Search store** with metadata:
  - `graph_id`: The graph's unique identifier
  - `domain`: "topeic.com" (tenant identifier)
  - `version`: "1.1" (schema version)
  - `display_name`: The graph's name
- The Gemini file ID is saved in `documents.gemini_file_id`
- Upload happens asynchronously (doesn't block main flow)

### 4. Chat Interaction
When a user sends a chat message:
1. User message is saved to database
2. Gemini queries the **shared File Search store** with a metadata filter:
   ```
   (chunk.custom_metadata.graph_id = "graph_abc123" AND 
    chunk.custom_metadata.domain = "topeic.com" AND 
    chunk.custom_metadata.version = "1.1")
   ```
3. Only documents matching the filter are used for context
4. AI generates a response using the retrieved context
5. Response is streamed back via SSE
6. Assistant message is saved to database

### Metadata-Based Filtering

**Graph Isolation:**
- Each document is tagged with its `graph_id`
- Queries filter by `graph_id` to ensure users only see their graph's documents
- No cross-graph data leakage

**Domain Filtering:**
- Documents are tagged with `domain` (currently "topeic.com")
- Enables future multi-tenant support
- Can filter by tenant in shared deployments

**Version Filtering:**
- Documents are tagged with `version` (currently "1.1")
- Enables schema migrations without re-uploading all documents
- Can query specific versions or migrate incrementally

## Testing Your Setup

### 1. Verify Configuration

```bash
# Check if environment variables are set
cd backend
source .env
echo $GEMINI_API_KEY
echo $GEMINI_PROJECT_ID
echo $GEMINI_LOCATION
```

### 2. Test API Connection

Start the backend server and check logs for Gemini initialization:

```bash
cd backend
go run cmd/server/main.go
```

Look for log messages like:
```
INFO: Gemini service initialized successfully
INFO: Using model: gemini-1.5-flash
INFO: Region: us-central1
INFO: Initializing Gemini File Search store...
INFO: File Search store initialized: OrgMind Documents (store ID: projects/.../fileSearchStores/...)
```

### 3. Test File Search

1. Create a knowledge graph in the UI
2. Upload a document
3. Check logs for document upload with metadata:
   ```
   INFO: Uploading document to File Search with metadata: graph_id=<graph-id>, domain=topeic.com, version=1.1
   INFO: Uploaded document to File Search: <document-id>
   ```
4. Open the chat interface
5. Ask a question about your document
6. Check logs for metadata filtering:
   ```
   INFO: Querying File Search with filter: (chunk.custom_metadata.graph_id = "<graph-id>" AND ...)
   ```
7. Verify you receive a streaming response

## Troubleshooting

### Error: "API key not valid"
- Verify your API key is correct
- Check that the Generative Language API is enabled in your project
- Ensure there are no extra spaces or newlines in the `.env` file

### Error: "Project not found"
- Verify your project ID is correct (lowercase, no spaces)
- Ensure the project exists in Google Cloud Console
- Check that billing is enabled for the project

### Error: "Region not supported"
- Use one of the supported regions listed above
- Default to `us-central1` if unsure

### Error: "Quota exceeded"
- Check your API usage in Google Cloud Console
- Request a quota increase if needed
- Consider implementing rate limiting

### Documents not appearing in chat responses
- Check that `gemini_file_id` is populated in the database
- Verify File Search store was initialized at startup (check logs)
- Look for upload errors in backend logs
- Verify metadata is attached correctly (check upload logs)
- Ensure metadata filter matches the document's metadata
- Allow a few seconds for document indexing after upload

### Error: "Failed to initialize File Search store"
- Verify `GEMINI_API_KEY` is valid
- Check that `GEMINI_PROJECT_ID` and `GEMINI_LOCATION` are correct
- Ensure the Generative Language API is enabled
- Check backend logs for detailed error messages
- The application will exit if store initialization fails

## Cost Considerations

### Pricing (as of 2024)
- **File Search:** Free for up to 1,000 files per day
- **API Calls:** Charged per 1,000 characters
  - Input: ~$0.00015 per 1K characters (flash model)
  - Output: ~$0.0006 per 1K characters (flash model)

### Cost Optimization Tips
1. Use `gemini-1.5-flash` instead of `gemini-1.5-pro` for most queries
2. Implement rate limiting (already configured: 20 messages/minute per user)
3. Cache common responses when appropriate
4. Monitor usage in Google Cloud Console
5. Set up billing alerts

## Security Best Practices

1. **Never commit API keys to version control**
   - Use `.env` files (already in `.gitignore`)
   - Use secret management in production (e.g., Google Secret Manager)

2. **Rotate API keys regularly**
   - Generate new keys every 90 days
   - Revoke old keys after rotation

3. **Restrict API key usage**
   - In Google Cloud Console, restrict keys to specific APIs
   - Limit keys to specific IP addresses in production

4. **Monitor API usage**
   - Set up alerts for unusual activity
   - Review usage logs regularly

## Additional Resources

- **Official Documentation:** https://ai.google.dev/gemini-api/docs/file-search
- **Go SDK Repository:** https://github.com/googleapis/go-genai
- **Example Code:** https://github.com/googleapis/go-genai/blob/main/examples/filesearchstores/
- **API Reference:** https://ai.google.dev/gemini-api/docs/file-search#rest
- **Pricing:** https://ai.google.dev/pricing
- **Support:** https://cloud.google.com/support

## Support

If you encounter issues:
1. Check the troubleshooting section above
2. Review backend logs for detailed error messages
3. Consult the official Gemini documentation
4. Contact Google Cloud support for API-specific issues
