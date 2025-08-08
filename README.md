# dockrune

self-hosted deployment daemon. receives github webhooks, deploys your stuff.

## for who

you have a $5 vps. you want to deploy your side projects to it. you don't want to ssh in and pull manually. you don't want a paas, kubernetes, yaml files, or a web ui you'll never use.

you want: push to github → deployed to server. that's it.

## the problem

existing solutions assume you want complexity:
- enterprise platforms that need 4gb ram just to idle
- "simplified" tools that still require learning their special config format
- managed services that cost 10x your vps
- solutions that replace your entire workflow instead of enhancing it

## how dockrune is different

```bash
# installation
wget binary && ./dockrune init

# configuration  
none. it reads your code and figures it out.

# resource usage
~50mb ram. your apps use what they use.

# mental overhead
zero. push code, get deployment.
```

dockrune does exactly one thing: when github sends a webhook, it deploys your code. no more, no less.

## alternatives

| project | hosting | complexity | languages | config | vibe |
|---------|---------|------------|-----------|--------|------|
| **dockrune** | self-hosted | minimal | go | zero | unix philosophy |
| [dokku](https://dokku.com) | self-hosted | medium | bash/go | buildpacks | heroku-like paas |
| [coolify](https://coolify.io) | self-hosted | high | php/vue | gui-heavy | all-in-one platform |
| [caprover](https://caprover.com) | self-hosted | medium | node | gui-based | docker swarm |
| [waypoint](https://www.waypointproject.io) | self/cloud | high | go | hcl files | hashicorp ecosystem |
| [render](https://render.com) | managed | low | - | yaml/gui | vercel competitor |
| [railway](https://railway.app) | managed | low | - | gui | heroku successor |
| [fly.io](https://fly.io) | managed | medium | - | toml | edge computing |

**why dockrune?**
- single binary, no dependencies
- learns your project structure automatically
- minimal resource usage (~50mb ram)
- direct github integration, no git push deploy
- you already have a vps, just use it

## quick start

```bash
# get it
git clone https://github.com/ejfox/dockrune
cd dockrune

# build it
go build -o dockrune ./cmd/dockrune

# configure it
./dockrune init

# run it
./dockrune serve
```

## docker

```bash
docker run -d \
  -p 8000:8000 \
  -p 8001:8001 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd)/data:/app/data \
  -e GITHUB_WEBHOOK_SECRET=your-secret \
  ejfox/dockrune
```

## configure github

1. go to repo settings → webhooks
2. add webhook:
   - url: `https://your-server.com/webhook/github`
   - content type: `application/json`
   - secret: your webhook secret
   - events: push, pull request

## env vars

```bash
GITHUB_WEBHOOK_SECRET=    # required
ADMIN_USERNAME=admin      # default: admin
ADMIN_PASSWORD=           # required
JWT_SECRET=               # required
DEPLOYMENT_DOMAIN=        # your domain
GITHUB_TOKEN=             # for private repos
DISCORD_WEBHOOK_URL=      # optional alerts
```

## api

- webhook: `:8000/webhook/github`
- health: `:8000/health`
- admin: `:8001`
- openapi: `:8001/openapi.json`
- deployments: `:8001/api/deployments` (jwt required)

## project detection

dockrune automatically detects and handles:

- **docker**: `docker-compose.yml` or `Dockerfile`
- **go**: `go.mod`
- **rust**: `Cargo.toml`
- **node**: `package.json`
- **python**: `requirements.txt`
- **static**: `index.html`

## architecture

```
github → webhook → detector → queue → deployer → [docker|pm2|binary]
                                ↓
                            storage ← admin api ← dashboard
```

## security

- hmac webhook validation
- jwt auth for admin api
- command injection prevention
- path traversal protection
- non-root docker execution

## development

```bash
# run tests
go test ./...

# run dashboard dev
cd dashboard && npm run dev

# run smoke tests
bash smoke_test.sh
```

## license

mit