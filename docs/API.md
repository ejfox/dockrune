# API Documentation

dockrune exposes two main API interfaces: the webhook server for receiving GitHub events and the admin API for dashboard functionality.

## Base URLs

- **Webhook Server**: `http://localhost:8000` (configurable via `WEBHOOK_PORT`)
- **Admin Dashboard**: `http://localhost:8001` (configurable via `ADMIN_PORT`)

## Webhook API

### POST /webhook/github

Receives GitHub webhook events for deployment triggers.

**Headers:**
```
Content-Type: application/json
X-GitHub-Delivery: <delivery-id>
X-GitHub-Event: <event-type>
X-Hub-Signature-256: sha256=<hmac-signature>
```

**Request Body:**
GitHub webhook payload (varies by event type)

**Response:**
```json
{
  "status": "success",
  "deployment_id": "dep_123abc",
  "message": "Deployment queued"
}
```

**Status Codes:**
- `200` - Webhook processed successfully
- `400` - Invalid payload or signature
- `401` - Missing or invalid signature
- `500` - Internal server error

### GET /health

Health check endpoint for monitoring.

**Response:**
```json
{
  "status": "ok",
  "version": "1.0.0",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Admin API

All admin API endpoints require authentication via JWT token.

### Authentication

#### POST /admin/login

Authenticate and receive JWT token.

**Request:**
```json
{
  "username": "admin",
  "password": "your-password"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Status Codes:**
- `200` - Login successful
- `401` - Invalid credentials

### Using Authentication

Include the JWT token in the `Authorization` header:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

### Deployments

#### GET /api/deployments

Get list of deployments.

**Query Parameters:**
- `limit` (optional): Number of deployments to return (default: 50)
- `status` (optional): Filter by status (`queued`, `in_progress`, `success`, `failed`)
- `owner` (optional): Filter by repository owner
- `repo` (optional): Filter by repository name

**Response:**
```json
[
  {
    "id": "dep_123abc",
    "owner": "ejfox",
    "repo": "myapp",
    "ref": "refs/heads/main",
    "sha": "abc123...",
    "clone_url": "https://github.com/ejfox/myapp.git",
    "environment": "production",
    "pr_number": 0,
    "github_deployment_id": 12345,
    "status": "success",
    "started_at": "2024-01-01T12:00:00Z",
    "completed_at": "2024-01-01T12:02:30Z",
    "log_path": "/app/logs/dep_123abc.log",
    "url": "https://myapp.example.com",
    "port": 3000,
    "project_type": "nodejs",
    "error": ""
  }
]
```

#### GET /api/deployments/{id}

Get specific deployment details.

**Response:**
```json
{
  "id": "dep_123abc",
  "owner": "ejfox",
  "repo": "myapp",
  "ref": "refs/heads/main",
  "sha": "abc123...",
  "clone_url": "https://github.com/ejfox/myapp.git",
  "environment": "production",
  "pr_number": 0,
  "github_deployment_id": 12345,
  "status": "success",
  "started_at": "2024-01-01T12:00:00Z",
  "completed_at": "2024-01-01T12:02:30Z",
  "log_path": "/app/logs/dep_123abc.log",
  "url": "https://myapp.example.com",
  "port": 3000,
  "project_type": "nodejs",
  "error": ""
}
```

**Status Codes:**
- `200` - Deployment found
- `404` - Deployment not found

#### POST /api/deployments/{id}/redeploy

Trigger redeployment of an existing deployment.

**Response:**
```json
{
  "message": "Redeployment queued",
  "id": "dep_456def"
}
```

**Status Codes:**
- `200` - Redeployment queued
- `404` - Original deployment not found
- `500` - Failed to queue redeployment

#### POST /api/deployments/{id}/stop

Stop a running deployment.

**Response:**
```json
{
  "message": "Deployment stopped",
  "id": "dep_123abc"
}
```

**Status Codes:**
- `200` - Deployment stopped
- `404` - Deployment not found
- `400` - Deployment cannot be stopped (already completed/failed)

#### GET /api/deployments/{id}/logs

Stream deployment logs.

**Response:** Plain text log stream

**Status Codes:**
- `200` - Logs found and streamed
- `404` - Deployment or logs not found

### WebSocket API

#### GET /api/ws

Establish WebSocket connection for real-time updates.

**Protocol:** WebSocket
**Authentication:** JWT token via query parameter `?token=<jwt-token>`

**Messages:**
The server sends JSON messages with deployment updates:

```json
{
  "type": "deployment_update",
  "data": {
    "id": "dep_123abc",
    "status": "in_progress",
    "timestamp": "2024-01-01T12:01:00Z"
  }
}
```

**Message Types:**
- `deployment_update` - Status change for a deployment
- `deployment_logs` - New log lines for active deployments
- `error` - Error messages

## Data Models

### Deployment

```json
{
  "id": "string",
  "owner": "string",
  "repo": "string", 
  "ref": "string",
  "sha": "string",
  "clone_url": "string",
  "environment": "string",
  "pr_number": "number",
  "github_deployment_id": "number",
  "status": "queued|in_progress|success|failed",
  "started_at": "string (ISO 8601)",
  "completed_at": "string (ISO 8601)",
  "log_path": "string",
  "url": "string", 
  "port": "number",
  "project_type": "string",
  "error": "string"
}
```

### Deployment Status

- `queued` - Deployment is waiting to start
- `in_progress` - Deployment is currently running
- `success` - Deployment completed successfully
- `failed` - Deployment failed with errors

### Project Types

Automatically detected project types:

- `docker` - Docker Compose or Dockerfile projects
- `go` - Go applications with go.mod
- `rust` - Rust applications with Cargo.toml
- `nodejs` - Node.js applications with package.json
- `nuxt` - Nuxt.js applications
- `python` - Python applications with requirements.txt
- `static` - Static HTML sites

## Error Responses

All API endpoints return consistent error responses:

```json
{
  "error": "Error message description",
  "code": "ERROR_CODE",
  "details": {
    "field": "Additional error details"
  }
}
```

Common error codes:
- `INVALID_REQUEST` - Malformed request
- `UNAUTHORIZED` - Authentication required
- `FORBIDDEN` - Insufficient permissions
- `NOT_FOUND` - Resource not found
- `INTERNAL_ERROR` - Server error

## Rate Limiting

The admin API includes basic rate limiting:
- 100 requests per minute per IP address
- WebSocket connections limited to 10 per IP address

Webhook endpoints have no rate limiting as they're controlled by GitHub.

## CORS

The admin API includes CORS headers for dashboard access:
- Allows all origins in development
- Production deployments should configure allowed origins

## Example Usage

### JavaScript/Node.js

```javascript
// Authenticate
const response = await fetch('http://localhost:8001/admin/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    username: 'admin',
    password: 'password'
  })
});
const { token } = await response.json();

// Get deployments
const deployments = await fetch('http://localhost:8001/api/deployments', {
  headers: { 'Authorization': `Bearer ${token}` }
});
const data = await deployments.json();

// WebSocket connection
const ws = new WebSocket(`ws://localhost:8001/api/ws?token=${token}`);
ws.onmessage = (event) => {
  const update = JSON.parse(event.data);
  console.log('Deployment update:', update);
};
```

### curl

```bash
# Login
TOKEN=$(curl -s -X POST http://localhost:8001/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' \
  | jq -r .token)

# Get deployments
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8001/api/deployments

# Trigger redeployment
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:8001/api/deployments/dep_123abc/redeploy
```

## Monitoring

The webhook server exposes a `/health` endpoint for monitoring and load balancer health checks. The admin API does not have a separate health endpoint but returns appropriate HTTP status codes for monitoring.

For production monitoring, consider:
- Monitoring deployment success rates via API
- Setting up alerts for failed deployments
- Tracking deployment duration metrics
- Monitoring disk space for logs and repositories