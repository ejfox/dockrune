package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ejfox/dockrune/internal/config"
	"github.com/ejfox/dockrune/internal/deployer"
	"github.com/ejfox/dockrune/internal/github"
	"github.com/ejfox/dockrune/internal/models"
	"github.com/gin-gonic/gin"
)

type Server struct {
	config   *config.Config
	deployer *deployer.Deployer
	github   *github.Client
}

func NewServer(cfg *config.Config, dep *deployer.Deployer, gh *github.Client) *Server {
	return &Server{
		config:   cfg,
		deployer: dep,
		github:   gh,
	}
}

func (s *Server) Start() error {
	r := gin.Default()

	// Middleware
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// Routes
	r.GET("/health", s.healthCheck)
	r.POST("/webhook/github", s.handleGitHubWebhook)

	addr := fmt.Sprintf(":%d", s.config.WebhookPort)
	fmt.Printf("Webhook server listening on %s\n", addr)
	return r.Run(addr)
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func (s *Server) handleGitHubWebhook(c *gin.Context) {
	// Read raw body for signature verification
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Verify signature
	signature := c.GetHeader("X-Hub-Signature-256")
	if !s.verifySignature(body, signature) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Parse event type
	eventType := c.GetHeader("X-GitHub-Event")
	if eventType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing X-GitHub-Event header"})
		return
	}

	// Handle different event types
	switch eventType {
	case "push":
		s.handlePushEvent(c, body)
	case "pull_request":
		s.handlePullRequestEvent(c, body)
	case "ping":
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	default:
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Event type %s not handled", eventType)})
	}
}

func (s *Server) verifySignature(payload []byte, signature string) bool {
	if signature == "" || s.config.WebhookSecret == "" {
		return false
	}

	// Remove "sha256=" prefix
	signature = strings.TrimPrefix(signature, "sha256=")

	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(s.config.WebhookSecret))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Constant time comparison
	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

func (s *Server) handlePushEvent(c *gin.Context, body []byte) {
	var event PushEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse push event"})
		return
	}

	// Extract deployment info
	deployment := &models.Deployment{
		Owner:       event.Repository.Owner.Login,
		Repo:        event.Repository.Name,
		Ref:         event.Ref,
		SHA:         event.After,
		CloneURL:    event.Repository.CloneURL,
		Environment: s.getEnvironmentFromRef(event.Ref),
	}

	// Create GitHub deployment
	if s.github != nil {
		deploymentID, err := s.github.CreateDeployment(
			deployment.Owner,
			deployment.Repo,
			deployment.SHA,
			deployment.Environment,
		)
		if err == nil {
			deployment.GitHubDeploymentID = deploymentID
		}
	}

	// Queue deployment
	if err := s.deployer.QueueDeployment(deployment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue deployment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Deployment queued",
		"sha":     event.After,
		"ref":     event.Ref,
	})
}

func (s *Server) handlePullRequestEvent(c *gin.Context, body []byte) {
	var event PullRequestEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse pull request event"})
		return
	}

	// Only deploy on opened/synchronize events
	if event.Action != "opened" && event.Action != "synchronize" {
		c.JSON(http.StatusOK, gin.H{"message": "No action taken"})
		return
	}

	// Create preview deployment
	deployment := &models.Deployment{
		Owner:       event.Repository.Owner.Login,
		Repo:        event.Repository.Name,
		Ref:         event.PullRequest.Head.Ref,
		SHA:         event.PullRequest.Head.SHA,
		CloneURL:    event.Repository.CloneURL,
		Environment: fmt.Sprintf("preview-pr-%d", event.PullRequest.Number),
		PRNumber:    event.PullRequest.Number,
	}

	// Create GitHub deployment
	if s.github != nil {
		deploymentID, err := s.github.CreateDeployment(
			deployment.Owner,
			deployment.Repo,
			deployment.SHA,
			deployment.Environment,
		)
		if err == nil {
			deployment.GitHubDeploymentID = deploymentID
		}
	}

	// Queue deployment
	if err := s.deployer.QueueDeployment(deployment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue deployment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Preview deployment queued",
		"pr":          event.PullRequest.Number,
		"environment": deployment.Environment,
	})
}

func (s *Server) getEnvironmentFromRef(ref string) string {
	switch {
	case strings.HasSuffix(ref, "/main") || strings.HasSuffix(ref, "/master"):
		return "production"
	case strings.Contains(ref, "/staging"):
		return "staging"
	case strings.Contains(ref, "/develop"):
		return "development"
	default:
		// Extract branch name
		parts := strings.Split(ref, "/")
		if len(parts) > 2 {
			return fmt.Sprintf("preview-%s", parts[len(parts)-1])
		}
		return "preview"
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Webhook event types
type PushEvent struct {
	Ref        string `json:"ref"`
	Before     string `json:"before"`
	After      string `json:"after"`
	Repository struct {
		Name     string `json:"name"`
		CloneURL string `json:"clone_url"`
		Owner    struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
}

type PullRequestEvent struct {
	Action      string `json:"action"`
	Number      int    `json:"number"`
	PullRequest struct {
		Number int    `json:"number"`
		State  string `json:"state"`
		Head   struct {
			Ref string `json:"ref"`
			SHA string `json:"sha"`
		} `json:"head"`
	} `json:"pull_request"`
	Repository struct {
		Name     string `json:"name"`
		CloneURL string `json:"clone_url"`
		Owner    struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
}
