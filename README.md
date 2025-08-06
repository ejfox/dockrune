**dockrune: dev memo**

dockrune is a self-hosted deployment daemon designed to receive github webhooks, detect project types, and run appropriate deploy commands on a local vps. it supports status reporting via github’s deployments api, discord alerts through local services like coach artie, and custom automation via n8n.

this document outlines dockrune’s intended functionality, architectural decisions, and acceptance criteria for development.

---

**primary goals**

- enable push-to-deploy workflows from github to a docker-hosted vps
- deploy multiple project types with minimal config (zero-config where possible)
- support preview environments per feature branch or pr
- maintain a simple but effective web ui to view deploy state, logs, and manage apps
- report deploy status visibly in github (green/red on commits/prs)
- alert developers of deploy errors via discord and optionally trigger n8n workflows
- never require ssh from github into the server; always pull-based architecture

---

**core features**

- webhook server receives github events (push, pull_request, etc)
- deploy daemon inspects project contents and runs corresponding deploy process:
  - docker projects: run docker-compose
  - nuxt3 nitro apps: run output/server/index.mjs
  - future support for static sites, custom commands, etc

- per-project .env file defines domain, port, optional metadata
- status is posted back to github using the deployments api
- logs are captured per deploy and can be viewed in the web ui
- re-deploys can be triggered manually via web ui or api
- deploy failures trigger alerts through discord (via coach artie) and/or n8n
- metadata is tracked locally in a sqlite db or flat file store
- project health and uptime are visualized in a minimal admin dashboard

---

**web interface**

- minimal, functional admin view available at /admin
- displays:
  - active apps
  - last deploy sha and timestamp
  - port + domain
  - status (live, failing, stopped)
  - log view per app
  - buttons to redeploy or stop apps

- built with fastapi or express.js, templated html + optional rest api

---

**alerting system**

- configurable alert system on success/failure
- alerts can be sent to:
  - coach artie (discord bot with http endpoints)
  - n8n (webhooks)
  - github comments or status updates

- alert templates are customizable
- alerts may include preview url, deploy sha, and short log summary

---

**acceptance criteria**

- the daemon can receive a github webhook and deploy a docker project automatically
- deploy status appears on the corresponding github commit/pr
- a minimal admin dashboard lists all active apps and their deploy state
- logs are accessible via the dashboard or api
- failure alerts are sent to a test discord channel via coach artie
- deploys can be manually re-triggered from the ui
- the system can handle multiple simultaneous repos with separate ports/domains
- no ssh access to the vps is required from github
- setup can be performed with basic instructions and no third-party cloud services

---

**non-goals (for now)**

- multi-tenant support
- autoscaling
- cloud provider integration
- ui/ux polish beyond basic utility
- database migrations or advanced data persistence

---

**dev notes**

dockrune should feel closer to a well-crafted script than an over-engineered ci platform. the goal is total control, fast iteration, minimal surface area, and strong developer ergonomics. make it composable, inspectable, and a pleasure to use—even if it stays “ugly but honest.”
