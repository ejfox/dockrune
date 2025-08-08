# 🚀 dockrune

**Self-hosted deployment daemon for GitHub webhooks** - Push to deploy on your own VPS!

## ✨ Features

- **🔄 Push-to-Deploy** - Automatic deployments triggered by GitHub webhooks
- **🔍 Smart Detection** - Auto-detects Docker, Go, Rust, Node.js, Python, Nuxt, and static sites
- **📊 Beautiful Dashboard** - Modern Nuxt 3 dashboard with real-time updates
- **✅ GitHub Integration** - Green/red status checks on commits and PRs
- **🔔 Notifications** - Discord and n8n webhook alerts
- **🌐 Multi-Environment** - Production, staging, and preview deployments
- **🎯 Zero Config** - Works out of the box for most project types
- **⚡ Fast & Efficient** - Written in Go for maximum performance

## 🏗️ Architecture

```
dockrune/
├── cmd/dockrune/        # CLI entry point
├── internal/            # Go application code
│   ├── webhook/         # GitHub webhook server
│   ├── detector/        # Project type detection
│   ├── deployer/        # Deployment orchestration
│   ├── github/          # GitHub API integration
│   ├── alerting/        # Discord/n8n notifications
│   └── admin/           # Dashboard API server
├── dashboard/           # Nuxt 3 web dashboard
│   ├── pages/          # Dashboard pages
│   ├── components/     # Vue components
│   └── stores/         # Pinia state management
└── docker-compose.yml   # Production deployment
```

## 🚀 Quick Start

### Option 1: Binary Installation

```bash
# Clone the repository
git clone https://github.com/ejfox/dockrune.git
cd dockrune

# Build the project
make build

# Initialize configuration
./dockrune init

# Start the server
./dockrune serve
```

### Option 2: Docker Deployment

```bash
# Clone and configure
git clone https://github.com/ejfox/dockrune.git
cd dockrune
cp .env.example .env
# Edit .env with your configuration

# Start with Docker Compose
docker-compose up -d
```

## 🔧 Configuration

### 1. Initialize dockrune

Run the interactive setup:

```bash
./dockrune init
```

This will prompt you for:
- GitHub personal access token
- GitHub webhook secret
- Deployment domain
- Admin credentials
- Optional Discord/n8n webhooks

### 2. Configure GitHub Webhook

In your GitHub repository settings:

1. Go to **Settings → Webhooks → Add webhook**
2. **Payload URL**: `https://your-domain.com/webhook/github`
3. **Content type**: `application/json`
4. **Secret**: Use the secret from your `.env` file
5. **Events**: Select `Push` and `Pull request`

### 3. Access the Dashboard

Open your browser to `http://localhost:8001` and login with your admin credentials.

## 📦 Supported Project Types

dockrune automatically detects and deploys:

| Type | Detection | Build Command | Start Command |
|------|-----------|---------------|---------------|
| **Docker** | `docker-compose.yml` | `docker-compose build` | `docker-compose up -d` |
| **Go** | `go.mod` | `go build -o app` | `./app` |
| **Rust** | `Cargo.toml` | `cargo build --release` | `./target/release/app` |
| **Node.js** | `package.json` | `npm install && npm run build` | `npm start` |
| **Nuxt 3** | `nuxt` in package.json | `npm run build` | `node .output/server/index.mjs` |
| **Python** | `requirements.txt` | `pip install -r requirements.txt` | `python app.py` |
| **Static** | `index.html` | - | `python -m http.server` |

## 🎨 Dashboard Features

The Nuxt 3 dashboard provides:

- **Real-time Updates** - WebSocket connection for live deployment status
- **Deployment Timeline** - Visual chart of deployment history
- **Log Viewer** - Stream deployment logs with auto-scroll
- **Quick Actions** - Redeploy, stop, or view deployment details
- **Statistics** - Success rate, average duration, active deployments
- **Mobile Responsive** - Works great on all devices

## 🔔 Alerting

### Discord Integration

Set `DISCORD_WEBHOOK_URL` in your `.env` file. Alerts include:
- Deployment success/failure notifications
- Embedded deployment details
- Direct links to deployment URLs
- Error logs for failed deployments

### n8n Integration

Set `N8N_WEBHOOK_URL` for custom workflow automation:
- Trigger complex workflows on deployment events
- Integrate with other services
- Custom notification routing

## 🐳 Docker Deployment

The included `docker-compose.yml` provides:

```yaml
services:
  dockrune:
    # Main application
    ports:
      - "8000:8000"  # Webhook server
      - "8001:8001"  # Admin dashboard
    volumes:
      - ./data:/app/data        # SQLite database
      - ./logs:/app/logs        # Deployment logs
      - ./repos:/app/repos      # Git repositories
      - /var/run/docker.sock:/var/run/docker.sock

  traefik:
    # Optional: Reverse proxy with SSL
    # Automatic Let's Encrypt certificates
```

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Node.js 20+
- Docker (for containerized deployments)

### Local Development

```bash
# Install dependencies
make deps
cd dashboard && npm install

# Run in development mode
make dev

# In another terminal, run the dashboard
cd dashboard && npm run dev

# Run tests
make test

# Format code
make fmt

# Lint
make lint
```

### Building from Source

```bash
# Build everything
make build

# Build for multiple platforms
make build-all

# Build Docker image
make docker-build
```

## 📝 Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `GITHUB_TOKEN` | GitHub personal access token | No* |
| `GITHUB_WEBHOOK_SECRET` | Webhook signature verification | Yes |
| `DEPLOYMENT_DOMAIN` | Your deployment domain | Yes |
| `ADMIN_USERNAME` | Dashboard admin username | Yes |
| `ADMIN_PASSWORD` | Dashboard admin password | Yes |
| `JWT_SECRET` | JWT signing secret | Yes |
| `DISCORD_WEBHOOK_URL` | Discord notifications | No |
| `N8N_WEBHOOK_URL` | n8n workflow triggers | No |

*Required for private repositories

## 🔒 Security

- **HMAC webhook verification** - All GitHub webhooks are cryptographically verified
- **JWT authentication** - Secure token-based auth for the dashboard
- **Docker socket isolation** - Deployments run in isolated containers
- **No SSH required** - Pull-based architecture, GitHub never needs server access

## 📚 CLI Commands

```bash
# Start the server
dockrune serve

# Initialize configuration
dockrune init

# Manually trigger deployment
dockrune deploy --owner ejfox --repo myapp --ref main

# Check deployment status
dockrune status
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

MIT License - see LICENSE file for details

## 🙏 Acknowledgments

Built with:
- [Gin](https://gin-gonic.com/) - HTTP web framework
- [Nuxt 3](https://nuxt.com/) - Vue.js framework
- [Pinia](https://pinia.vuejs.org/) - State management
- [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS
- [go-github](https://github.com/google/go-github) - GitHub API client

---

**dockrune** - *Self-hosted deployments made simple* 🚀