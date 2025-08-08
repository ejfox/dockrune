# Dockrune Security Guide

*"Security is not a product, but a process." - Bruce Schneier*

This guide explains the security features implemented in dockrune and how to use them properly. Because the difference between secure and "secure" is usually about 6 months and a breach report.

## 🛡️ Security Features Overview

### 1. GitHub Webhook Signature Validation (HMAC-SHA256)

**What it does:** Validates that webhooks actually come from GitHub using HMAC-SHA256 signatures.

**Why it matters:** Without this, anyone who knows your webhook URL can trigger deployments. That's basically giving strangers the keys to your server.

```javascript
// Automatically validates X-Hub-Signature-256 header
// Uses timing-safe comparison to prevent timing attacks
// Rejects malformed or missing signatures with 401 status
```

**Configuration:**
- Set `GITHUB_WEBHOOK_SECRET` in your environment
- Use the same secret in your GitHub webhook configuration
- Secret should be at least 32 characters of random data

### 2. Rate Limiting

**What it does:** Prevents abuse by limiting request frequency per IP address.

**Different limits for different endpoints:**
- Webhooks: 10 requests per minute (deployments are expensive)
- Admin: 50 requests per 5 minutes (normal usage)
- Auth: 5 attempts per 15 minutes (brute force prevention)
- Global: 100 requests per minute (generous fallback)

**Key features:**
- Smart rate limiting based on endpoint type
- Uses IP + User-Agent + context for more granular control
- Includes retry-after headers for client backoff

### 3. Input Validation and Sanitization

**What it does:** Validates and cleans all input to prevent injection attacks.

**Validation includes:**
- JSON schema validation for webhook payloads
- String length limits (because 10MB commit messages are suspicious)
- Character filtering to remove control characters and potential XSS
- Array size limits to prevent memory exhaustion attacks

**Sanitization features:**
- Removes HTML tags and control characters
- Limits string lengths to prevent DoS
- Validates email formats and alphanumeric fields
- Recursive sanitization of nested objects

### 4. Encrypted Secret Management

**What it does:** Stores secrets encrypted at rest instead of plain environment variables.

**Why environment variables are bad:**
- Visible in process lists (`ps aux | grep node`)
- Logged by systemd and other process managers
- Accessible to any process running as the same user
- Often accidentally committed to version control

**How our system works:**
- Master password derives encryption key using PBKDF2
- Secrets encrypted with AES-256-GCM (authenticated encryption)
- Only decrypted in memory when needed
- Falls back to environment variables for development

### 5. Authentication System

**What it does:** Protects admin endpoints with JWT-based authentication.

**Security features:**
- Passwords hashed with bcrypt (12 rounds = ~250ms)
- Account lockout after 5 failed attempts (15 minute lockout)
- JWT tokens with 2-hour expiration
- Token revocation/blacklist support for logout
- Timing-safe password verification

**Login flow:**
1. User submits credentials
2. System checks for account lockout
3. Verifies username and password hash
4. Generates JWT with unique ID
5. Returns token for subsequent requests

### 6. Security Headers and CORS

**Security headers implemented:**
- `Content-Security-Policy`: Prevents XSS attacks
- `X-Frame-Options`: Prevents clickjacking
- `X-Content-Type-Options`: Prevents MIME sniffing
- `Strict-Transport-Security`: Forces HTTPS
- `Referrer-Policy`: Controls referrer information
- `Permissions-Policy`: Disables unnecessary browser features

**CORS configuration:**
- Restrictive origin validation in production
- Permissive localhost origins in development
- Credentials support for authenticated requests
- Limited allowed methods and headers

## 🔧 Setup and Configuration

### Initial Setup

1. **Run the setup script** (recommended):
   ```bash
   npm run setup
   ```
   This interactive script will guide you through secure configuration.

2. **Manual configuration** (if you know what you're doing):
   ```bash
   cp .env.example .env
   # Edit .env with your values
   ```

### Environment Variables

```bash
# Server Configuration
NODE_ENV=production              # Enables production security mode
PORT=3000                       # Server port
TRUST_PROXY=true               # Set to true if behind reverse proxy

# Security (CRITICAL - change these!)
DOCKRUNE_MASTER_PASSWORD=...    # 16+ char password for secret encryption
GITHUB_WEBHOOK_SECRET=...       # From GitHub webhook settings
ADMIN_USERNAME=...              # Admin panel username
ADMIN_PASSWORD_HASH=...         # bcrypt hash (use setup script)

# Optional
JWT_SECRET=...                  # Auto-generated if not provided
ALLOWED_DOMAINS=...             # Comma-separated CORS domains
```

### First Run

```bash
npm install
npm test                        # Verify security features work
npm start                      # Start the server
```

The server will:
1. Load or generate encrypted secrets
2. Validate all security configurations
3. Start with all security middleware enabled

## 🧪 Testing Security Features

Run the comprehensive security test suite:

```bash
npm test                        # All tests
npm run test:security          # Security-specific tests
npm run test:watch             # Watch mode for development
```

**What the tests cover:**
- Webhook signature validation (including timing attack resistance)
- Rate limiting for all endpoint types
- Authentication flows and account lockout
- Input validation and sanitization
- Security header configuration
- Integration tests with realistic attack scenarios

## 🚨 Security Best Practices

### Deployment Security

1. **Use HTTPS in production** (required, not optional):
   ```nginx
   # Nginx example
   server {
       listen 443 ssl;
       ssl_certificate /path/to/cert.pem;
       ssl_private_key /path/to/key.pem;
       
       location / {
           proxy_pass http://localhost:3000;
           proxy_set_header X-Real-IP $remote_addr;
           proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
       }
   }
   ```

2. **Set TRUST_PROXY=true** when behind reverse proxy

3. **Restrict network access**:
   ```bash
   # Firewall rules example (ufw)
   ufw allow ssh
   ufw allow 80/tcp
   ufw allow 443/tcp
   ufw deny 3000/tcp  # Don't expose dockrune directly
   ufw enable
   ```

### Secret Management

1. **Never commit secrets to git**:
   ```bash
   # .gitignore should include:
   .env
   .env.*
   .secrets/
   ```

2. **Use strong master password**:
   - At least 16 characters
   - Mix of uppercase, lowercase, numbers, symbols
   - Consider using a password manager

3. **Rotate secrets periodically**:
   ```bash
   # Generate new webhook secret in GitHub
   # Update GITHUB_WEBHOOK_SECRET in .env
   # Restart dockrune
   ```

### Monitoring and Alerting

1. **Monitor logs for suspicious activity**:
   ```bash
   # Look for these patterns in logs:
   grep "Invalid webhook signature" logs/
   grep "Rate limit exceeded" logs/
   grep "Authentication failed" logs/
   grep "Suspicious request" logs/
   ```

2. **Set up log rotation**:
   ```bash
   # /etc/logrotate.d/dockrune
   /var/log/dockrune/*.log {
       daily
       missingok
       rotate 52
       compress
       notifempty
   }
   ```

3. **Consider external monitoring**:
   - Uptime monitoring (Pingdom, UptimeRobot)
   - Log aggregation (ELK stack, Grafana)
   - Security scanning (Nessus, OpenVAS)

## 🔍 Attack Vectors and Mitigations

### Common Attack Scenarios

1. **Brute Force Authentication**
   - **Attack:** Automated login attempts with common passwords
   - **Mitigation:** Account lockout, rate limiting, strong password requirements

2. **Webhook Spoofing**
   - **Attack:** Sending fake webhooks to trigger unauthorized deployments
   - **Mitigation:** HMAC signature validation, IP restrictions

3. **Payload Injection**
   - **Attack:** Malicious payloads in commit messages or branch names
   - **Mitigation:** Input validation, sanitization, parameterized commands

4. **Denial of Service**
   - **Attack:** Overwhelming server with requests
   - **Mitigation:** Rate limiting, input size limits, connection limits

5. **Information Disclosure**
   - **Attack:** Extracting sensitive information from error messages
   - **Mitigation:** Generic error messages, proper logging, security headers

### Advanced Threats

1. **Supply Chain Attacks**
   - Keep dependencies updated
   - Use `npm audit` regularly
   - Consider dependency scanning tools

2. **Insider Threats**
   - Principle of least privilege
   - Audit logs and access patterns
   - Secret rotation procedures

3. **Social Engineering**
   - Security awareness training
   - Multi-factor authentication where possible
   - Incident response procedures

## 📚 Further Reading

- [OWASP Web Security Testing Guide](https://owasp.org/www-project-web-security-testing-guide/)
- [GitHub Webhook Security](https://docs.github.com/en/developers/webhooks-and-events/webhooks/securing-your-webhooks)
- [Node.js Security Best Practices](https://nodejs.org/en/docs/guides/security/)
- [Express.js Security Best Practices](https://expressjs.com/en/advanced/best-practice-security.html)

---

*"The only secure computer is one that's unplugged, locked in a safe, and buried 20 feet underground. And even then I have my doubts." - Dennis Huges*

Remember: Security is an ongoing process, not a one-time setup. Stay vigilant, keep things updated, and assume that if something can go wrong, it will.