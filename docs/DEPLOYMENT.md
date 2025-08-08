# Production Deployment Guide

This guide walks you through deploying dockrune in a production environment on a VPS with SSL, monitoring, and best practices.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Server Setup](#server-setup)
- [Installation Methods](#installation-methods)
- [SSL Configuration](#ssl-configuration)
- [Monitoring Setup](#monitoring-setup)
- [Backup Strategy](#backup-strategy)
- [Security Hardening](#security-hardening)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Hardware Requirements

**Minimum:**
- 1 CPU core
- 1 GB RAM  
- 10 GB storage
- 1 Gbps network

**Recommended:**
- 2+ CPU cores
- 4+ GB RAM
- 50+ GB storage (for repositories and logs)
- Dedicated IP address

### Software Requirements

- Ubuntu 20.04+ or CentOS 8+ (other Linux distributions supported)
- Docker 24.0+ and Docker Compose 2.0+
- Git
- Port 80 and 443 open for HTTP/HTTPS
- Custom ports for webhook (8000) and dashboard (8001)

### DNS Setup

Before deployment, configure DNS:

```bash
# Primary domain for dashboard
dashboard.yourdomain.com → your-server-ip

# Webhook endpoint (can be same as dashboard)
hooks.yourdomain.com → your-server-ip

# Deployment subdomains (optional)
*.deploy.yourdomain.com → your-server-ip
```

## Server Setup

### 1. Initial Server Configuration

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install required packages
sudo apt install -y git curl wget unzip software-properties-common

# Create dockrune user
sudo useradd -m -s /bin/bash dockrune
sudo usermod -aG docker dockrune

# Create directory structure
sudo mkdir -p /opt/dockrune
sudo chown dockrune:dockrune /opt/dockrune
```

### 2. Install Docker

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Start Docker service
sudo systemctl enable docker
sudo systemctl start docker
```

### 3. Firewall Configuration

```bash
# Configure UFW
sudo ufw enable
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Allow SSH (adjust port if needed)
sudo ufw allow 22

# Allow HTTP/HTTPS
sudo ufw allow 80
sudo ufw allow 443

# Allow dockrune ports
sudo ufw allow 8000  # Webhook server
sudo ufw allow 8001  # Admin dashboard

# For deployed applications (adjust as needed)
sudo ufw allow 3000:9000/tcp
```

## Installation Methods

### Option 1: Docker Compose (Recommended)

Switch to the dockrune user and deploy:

```bash
sudo su - dockrune
cd /opt/dockrune

# Clone repository
git clone https://github.com/ejfox/dockrune.git .

# Create production environment file
cp .env.example .env
```

Edit `.env` with your production values:

```bash
# GitHub Configuration
GITHUB_TOKEN=ghp_your_github_token_here
GITHUB_WEBHOOK_SECRET=your_webhook_secret_here

# Domain Configuration  
DEPLOYMENT_DOMAIN=yourdomain.com
WEBHOOK_URL=https://hooks.yourdomain.com
DASHBOARD_URL=https://dashboard.yourdomain.com

# Admin Credentials
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your_secure_password_here
JWT_SECRET=your_jwt_secret_here

# Notification URLs (optional)
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/...
N8N_WEBHOOK_URL=https://your-n8n.com/webhook/...

# Server Configuration
WEBHOOK_PORT=8000
ADMIN_PORT=8001
LOG_LEVEL=info

# Database
DATABASE_PATH=/app/data/dockrune.db

# Storage Paths
REPOS_PATH=/app/repos
LOGS_PATH=/app/logs
```

Create production Docker Compose file:

```bash
cat > docker-compose.prod.yml << 'EOF'
version: '3.8'

services:
  dockrune:
    build: .
    restart: unless-stopped
    ports:
      - "8000:8000"  # Webhook server
      - "8001:8001"  # Admin dashboard
    volumes:
      - ./data:/app/data
      - ./logs:/app/logs  
      - ./repos:/app/repos
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - GITHUB_TOKEN=${GITHUB_TOKEN}
      - GITHUB_WEBHOOK_SECRET=${GITHUB_WEBHOOK_SECRET}
      - DEPLOYMENT_DOMAIN=${DEPLOYMENT_DOMAIN}
      - ADMIN_USERNAME=${ADMIN_USERNAME}
      - ADMIN_PASSWORD=${ADMIN_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
      - DISCORD_WEBHOOK_URL=${DISCORD_WEBHOOK_URL}
      - N8N_WEBHOOK_URL=${N8N_WEBHOOK_URL}
      - WEBHOOK_PORT=${WEBHOOK_PORT}
      - ADMIN_PORT=${ADMIN_PORT}
      - LOG_LEVEL=${LOG_LEVEL}
    networks:
      - dockrune-network

  traefik:
    image: traefik:v2.10
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./traefik:/etc/traefik
      - ./ssl:/ssl
    networks:
      - dockrune-network
    labels:
      - "traefik.enable=true"

networks:
  dockrune-network:
    external: false
EOF
```

Configure Traefik for SSL:

```bash
mkdir -p traefik ssl

cat > traefik/traefik.yml << 'EOF'
global:
  checkNewVersion: false
  sendAnonymousUsage: false

api:
  dashboard: true
  insecure: false

entryPoints:
  web:
    address: ":80"
    http:
      redirections:
        entrypoint:
          to: websecure
          scheme: https
  websecure:
    address: ":443"

certificatesResolvers:
  letsencrypt:
    acme:
      email: admin@yourdomain.com
      storage: /ssl/acme.json
      httpChallenge:
        entryPoint: web

providers:
  docker:
    endpoint: "unix:///var/run/docker.sock"
    exposedByDefault: false
EOF

cat > traefik/dynamic.yml << 'EOF'
http:
  routers:
    dashboard:
      rule: "Host(`dashboard.yourdomain.com`)"
      service: dockrune-dashboard
      tls:
        certResolver: letsencrypt
    webhook:
      rule: "Host(`hooks.yourdomain.com`)"
      service: dockrune-webhook
      tls:
        certResolver: letsencrypt

  services:
    dockrune-dashboard:
      loadBalancer:
        servers:
          - url: "http://dockrune:8001"
    dockrune-webhook:
      loadBalancer:
        servers:
          - url: "http://dockrune:8000"
EOF

chmod 600 ssl/acme.json
```

Deploy the stack:

```bash
# Build and start services
docker-compose -f docker-compose.prod.yml up -d

# Check status
docker-compose -f docker-compose.prod.yml ps
docker-compose -f docker-compose.prod.yml logs -f
```

### Option 2: Binary Installation

For direct installation without Docker:

```bash
# Download latest release
wget https://github.com/ejfox/dockrune/releases/latest/download/dockrune-linux-amd64.tar.gz
tar -xzf dockrune-linux-amd64.tar.gz
sudo mv dockrune /usr/local/bin/
sudo chmod +x /usr/local/bin/dockrune

# Create system user and directories
sudo useradd -r -s /bin/false dockrune
sudo mkdir -p /etc/dockrune /var/lib/dockrune /var/log/dockrune
sudo chown dockrune:dockrune /var/lib/dockrune /var/log/dockrune

# Create configuration
sudo tee /etc/dockrune/config.yml << 'EOF'
github:
  token: "your_github_token"
  webhook_secret: "your_webhook_secret"

server:
  webhook_port: 8000
  admin_port: 8001
  domain: "yourdomain.com"

admin:
  username: "admin"
  password: "your_secure_password"
  jwt_secret: "your_jwt_secret"

storage:
  database_path: "/var/lib/dockrune/dockrune.db"
  repos_path: "/var/lib/dockrune/repos"  
  logs_path: "/var/log/dockrune"
EOF

# Create systemd service
sudo tee /etc/systemd/system/dockrune.service << 'EOF'
[Unit]
Description=dockrune deployment daemon
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
User=dockrune
Group=dockrune
ExecStart=/usr/local/bin/dockrune serve --config /etc/dockrune/config.yml
Restart=always
RestartSec=10
Environment=PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

[Install]
WantedBy=multi-user.target
EOF

# Start service
sudo systemctl daemon-reload
sudo systemctl enable dockrune
sudo systemctl start dockrune
sudo systemctl status dockrune
```

## SSL Configuration

### Let's Encrypt with Certbot

If not using Traefik, set up SSL manually:

```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Get certificates
sudo certbot --nginx -d dashboard.yourdomain.com -d hooks.yourdomain.com

# Auto-renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

### Nginx Configuration

```bash
sudo tee /etc/nginx/sites-available/dockrune << 'EOF'
# Webhook server
server {
    listen 443 ssl http2;
    server_name hooks.yourdomain.com;
    
    ssl_certificate /etc/letsencrypt/live/hooks.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/hooks.yourdomain.com/privkey.pem;
    
    location / {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# Dashboard
server {
    listen 443 ssl http2;
    server_name dashboard.yourdomain.com;
    
    ssl_certificate /etc/letsencrypt/live/dashboard.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/dashboard.yourdomain.com/privkey.pem;
    
    location / {
        proxy_pass http://127.0.0.1:8001;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # WebSocket support
    location /api/ws {
        proxy_pass http://127.0.0.1:8001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
}

# HTTP redirect
server {
    listen 80;
    server_name hooks.yourdomain.com dashboard.yourdomain.com;
    return 301 https://$server_name$request_uri;
}
EOF

sudo ln -s /etc/nginx/sites-available/dockrune /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## Monitoring Setup

### 1. Log Management

Set up log rotation:

```bash
sudo tee /etc/logrotate.d/dockrune << 'EOF'
/var/log/dockrune/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 644 dockrune dockrune
    postrotate
        systemctl reload dockrune
    endscript
}
EOF
```

### 2. System Monitoring

Install basic monitoring tools:

```bash
# Install monitoring packages
sudo apt install htop iotop ncdu

# Create monitoring script
sudo tee /usr/local/bin/dockrune-health << 'EOF'
#!/bin/bash

# Check service status
if systemctl is-active --quiet dockrune; then
    echo "✓ dockrune service is running"
else
    echo "✗ dockrune service is NOT running"
    systemctl status dockrune --no-pager
fi

# Check disk space
df -h | grep -E '(/$|/var|/opt)'

# Check memory usage
free -h

# Check recent deployments
echo "Recent deployments:"
tail -20 /var/log/dockrune/deployments.log 2>/dev/null || echo "No deployment logs found"

# Check docker containers
docker ps --filter "name=dockrune"
EOF

chmod +x /usr/local/bin/dockrune-health
```

### 3. Alerting

Set up Discord/Slack alerts for system issues:

```bash
# Create alert script
sudo tee /usr/local/bin/dockrune-alert << 'EOF'
#!/bin/bash

SERVICE_NAME="dockrune"
DISCORD_WEBHOOK="your_discord_webhook_url_here"

if ! systemctl is-active --quiet $SERVICE_NAME; then
    MESSAGE="🚨 **ALERT**: $SERVICE_NAME is DOWN on $(hostname) at $(date)"
    
    curl -H "Content-Type: application/json" \
         -X POST \
         -d "{\"content\":\"$MESSAGE\"}" \
         "$DISCORD_WEBHOOK"
fi
EOF

chmod +x /usr/local/bin/dockrune-alert

# Add to crontab
sudo crontab -e
# Add: */5 * * * * /usr/local/bin/dockrune-alert
```

## Backup Strategy

### 1. Database Backup

```bash
# Create backup script
sudo tee /usr/local/bin/backup-dockrune << 'EOF'
#!/bin/bash

BACKUP_DIR="/opt/backups/dockrune"
DATE=$(date +%Y%m%d_%H%M%S)
DATABASE_PATH="/var/lib/dockrune/dockrune.db"

mkdir -p $BACKUP_DIR

# Stop service briefly for consistent backup
systemctl stop dockrune

# Backup database
cp $DATABASE_PATH $BACKUP_DIR/dockrune_${DATE}.db

# Backup configuration  
tar -czf $BACKUP_DIR/config_${DATE}.tar.gz /etc/dockrune/

# Start service
systemctl start dockrune

# Cleanup old backups (keep 30 days)
find $BACKUP_DIR -name "*.db" -mtime +30 -delete
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete

echo "Backup completed: $BACKUP_DIR/"
EOF

chmod +x /usr/local/bin/backup-dockrune

# Schedule daily backups
sudo crontab -e
# Add: 0 2 * * * /usr/local/bin/backup-dockrune
```

### 2. Remote Backup

For critical deployments, sync backups to remote storage:

```bash
# Install rclone for cloud storage
curl https://rclone.org/install.sh | sudo bash

# Configure rclone (follow prompts)
rclone config

# Add to backup script
echo "rclone sync /opt/backups/dockrune remote:dockrune-backups" >> /usr/local/bin/backup-dockrune
```

## Security Hardening

### 1. System Security

```bash
# Disable password authentication for SSH
sudo sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
sudo systemctl restart sshd

# Install fail2ban
sudo apt install fail2ban
sudo systemctl enable fail2ban

# Configure fail2ban for SSH
sudo tee /etc/fail2ban/jail.local << 'EOF'
[sshd]
enabled = true
port = 22
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 1800
EOF

sudo systemctl restart fail2ban
```

### 2. Application Security

```bash
# Set proper file permissions
sudo chmod 600 /etc/dockrune/config.yml
sudo chown root:root /etc/dockrune/config.yml

# Secure Docker socket (if needed)
sudo usermod -aG docker dockrune
sudo systemctl restart docker
```

### 3. Network Security

Consider using a VPN or private network for webhook communications:

```bash
# Install WireGuard (optional)
sudo apt install wireguard

# Configure as needed for your network topology
```

## Troubleshooting

### Common Issues

**Service won't start:**
```bash
# Check service status
sudo systemctl status dockrune

# Check logs
sudo journalctl -u dockrune -f

# Check configuration
dockrune --config /etc/dockrune/config.yml --check
```

**GitHub webhooks not received:**
```bash
# Check firewall
sudo ufw status

# Test webhook endpoint
curl -X POST https://hooks.yourdomain.com/health

# Check nginx logs
sudo tail -f /var/log/nginx/access.log
```

**Deployments failing:**
```bash
# Check deployment logs
tail -f /var/log/dockrune/deployments.log

# Check docker status
docker ps -a

# Check disk space
df -h
```

**SSL certificate issues:**
```bash
# Check certificate status
sudo certbot certificates

# Renew certificates manually
sudo certbot renew

# Test SSL configuration
openssl s_client -connect dashboard.yourdomain.com:443
```

### Performance Tuning

For high-traffic deployments:

```bash
# Increase file limits
echo "dockrune soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "dockrune hard nofile 65536" | sudo tee -a /etc/security/limits.conf

# Tune kernel parameters
echo "net.core.somaxconn = 1024" | sudo tee -a /etc/sysctl.conf
echo "net.ipv4.ip_local_port_range = 1024 65536" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

### Maintenance

Regular maintenance tasks:

```bash
# Update system packages
sudo apt update && sudo apt upgrade

# Update Docker images
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d

# Clean up old Docker images
docker system prune -a

# Rotate logs manually
sudo logrotate /etc/logrotate.d/dockrune

# Check backup integrity
ls -la /opt/backups/dockrune/
```

## Scaling Considerations

For larger deployments, consider:

1. **Load Balancing**: Use multiple dockrune instances behind a load balancer
2. **Database**: Move to PostgreSQL for better concurrency
3. **Storage**: Use dedicated storage for repositories and logs
4. **Monitoring**: Implement comprehensive monitoring with Prometheus/Grafana
5. **Security**: Use a service mesh or VPN for internal communications

This deployment guide should get you running securely in production. Adjust configurations based on your specific requirements and security policies.