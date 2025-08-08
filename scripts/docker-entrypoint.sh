#!/bin/bash
set -e

echo "Starting dockrune deployment daemon..."

# Create required directories if they don't exist
mkdir -p /app/data /app/logs /app/repos

# Initialize database if needed
if [ ! -f "/app/data/dockrune.db" ]; then
    echo "Initializing database..."
    python -c "from dockrune.storage.sqlite import SQLiteStorage; SQLiteStorage('/app/data/dockrune.db')"
fi

# Check Docker socket access
if [ ! -S /var/run/docker.sock ]; then
    echo "ERROR: Docker socket not found at /var/run/docker.sock"
    echo "Please mount the Docker socket with: -v /var/run/docker.sock:/var/run/docker.sock"
    exit 1
fi

# Test Docker access
if ! docker ps > /dev/null 2>&1; then
    echo "ERROR: Cannot access Docker. Please ensure the container has Docker permissions."
    exit 1
fi

# Generate default secrets if not provided
if [ -z "$GITHUB_WEBHOOK_SECRET" ]; then
    echo "WARNING: GITHUB_WEBHOOK_SECRET not set, generating random secret..."
    export GITHUB_WEBHOOK_SECRET=$(openssl rand -hex 32)
    echo "Generated webhook secret: $GITHUB_WEBHOOK_SECRET"
    echo "Please save this and configure it in your GitHub webhook settings!"
fi

if [ -z "$JWT_SECRET" ]; then
    echo "WARNING: JWT_SECRET not set, generating random secret..."
    export JWT_SECRET=$(openssl rand -hex 32)
fi

# Create admin user if credentials provided
if [ -n "$ADMIN_USERNAME" ] && [ -n "$ADMIN_PASSWORD" ]; then
    echo "Setting up admin user..."
    python -c "
from dockrune.web.admin import create_admin_user
create_admin_user('$ADMIN_USERNAME', '$ADMIN_PASSWORD')
" || echo "Admin user already exists or creation failed"
fi

# Start services
exec "$@"