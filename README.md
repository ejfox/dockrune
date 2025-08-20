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
  -p 9876:9876 \
  -p 9877:9877 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd)/data:/app/data \
  -e GITHUB_WEBHOOK_SECRET=your-secret \
  -e WEBHOOK_PORT=9876 \
  -e ADMIN_PORT=9877 \
  ejfox/dockrune
```

## configure github webhooks

### step 1: set up your webhook secret
```bash
# generate a secure webhook secret
openssl rand -hex 32

# add it to your .env file
echo "GITHUB_WEBHOOK_SECRET=your-generated-secret-here" >> .env
```

### step 2: configure github webhook

#### option a: single repository (manual)
1. go to your repository on GitHub
2. click **Settings** → **Webhooks** → **Add webhook**
3. configure the webhook (see settings below)

#### option b: organization-wide (recommended)
1. go to your GitHub organization settings
2. click **Settings** → **Webhooks** → **Add webhook**
3. this webhook will apply to ALL repos in your org automatically

#### option c: automate with gh cli (bulk setup)
```bash
# set up webhook for all your repos at once
gh api user/repos --paginate | jq -r '.[].full_name' | while read repo; do
  gh api repos/$repo/hooks -X POST -f name=web \
    -f config[url]=https://your-server.com:9876/webhook/github \
    -f config[secret]=your-webhook-secret \
    -f config[content_type]=json \
    -F events[]=push
  echo "✅ Added webhook to $repo"
done
```

#### webhook settings (for any option above):
- **Payload URL**: `https://your-server.com:9876/webhook/github`
- **Content type**: `application/json`
- **Secret**: paste your webhook secret from step 1
- **Which events**: select "Just the push event"
- **Active**: ✅ checked

### step 3: test the webhook
```bash
# make any commit and push to main branch
git add . && git commit -m "test webhook" && git push origin main

# check dockrune logs to see deployment
tail -f logs/dockrune.log
```

**webhook endpoint**: your dockrune server will receive webhooks at `/webhook/github`  
**admin dashboard**: access at `https://your-server.com:9877` for deployment monitoring

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

## api endpoints

- **webhook**: `:9876/webhook/github` (receives GitHub webhooks)
- **health**: `:9876/health` (server health check)
- **admin dashboard**: `:9877/` (web interface)
- **openapi spec**: `:9877/openapi.json` (api documentation)
- **deployments api**: `:9877/api/deployments` (jwt auth required)

**note**: ports are configurable via `WEBHOOK_PORT` and `ADMIN_PORT` environment variables

## zero config: how it works

dockrune looks at your code and knows what to do. no config files needed.

### detection logic

```bash
# if it finds docker-compose.yml
→ runs: docker-compose up -d

# if it finds package.json
→ runs: npm install && npm start

# if it finds go.mod
→ runs: go build && ./app

# if it finds requirements.txt
→ runs: pip install -r requirements.txt && python app.py
```

### real examples

**node project:**
```
my-app/
├── package.json     # detected: node project
├── index.js         # entry point found
└── src/            
```
dockrune runs: `npm install && npm start`

**docker project:**
```
my-app/
├── docker-compose.yml  # detected: docker compose
├── Dockerfile         
└── src/
```
dockrune runs: `docker-compose up -d`

**static site:**
```
my-site/
├── index.html      # detected: static files
├── style.css      
└── script.js
```
dockrune runs: `python -m http.server 8080`

### override if needed

don't like the defaults? add `.dockrune.yml`:
```yaml
build: make build
start: ./bin/server --prod
port: 9000
```

but most projects just work without it.

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