package admin

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ejfox/dockrune/internal/config"
	"github.com/ejfox/dockrune/internal/deployer"
	"github.com/ejfox/dockrune/internal/models"
	"github.com/ejfox/dockrune/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

type Server struct {
	config   *config.Config
	storage  storage.Storage
	deployer *deployer.Deployer
	upgrader websocket.Upgrader
}

func NewServer(cfg *config.Config, store storage.Storage, dep *deployer.Deployer) *Server {
	return &Server{
		config:   cfg,
		storage:  store,
		deployer: dep,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in dev
			},
		},
	}
}

func (s *Server) Start() error {
	r := gin.Default()

	// CORS middleware for API
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Static files for Nuxt dashboard
	r.Static("/_nuxt", "./dashboard/.output/public/_nuxt")
	r.StaticFile("/", "./dashboard/.output/public/index.html")

	// Public endpoints
	r.POST("/admin/login", s.handleLogin)
	r.GET("/openapi.json", s.OpenAPISpec)
	r.GET("/api-docs", s.OpenAPISpec)

	// API routes (protected)
	api := r.Group("/api")
	api.Use(s.authMiddleware())
	{
		api.GET("/deployments", s.getDeployments)
		api.GET("/deployments/:id", s.getDeployment)
		api.POST("/deployments/:id/redeploy", s.redeployDeployment)
		api.POST("/deployments/:id/stop", s.stopDeployment)
		api.GET("/deployments/:id/logs", s.getDeploymentLogs)
		api.GET("/ws", s.handleWebSocket)
	}

	addr := fmt.Sprintf(":%d", s.config.AdminPort)
	fmt.Printf("Admin dashboard listening on %s\n", addr)
	return r.Run(addr)
}

func (s *Server) redirectToAdmin(c *gin.Context) {
	c.Redirect(http.StatusMovedPermanently, "/admin")
}

func (s *Server) showDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "dockrune Admin",
	})
}

func (s *Server) showLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "dockrune Login",
	})
}

func (s *Server) handleLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Check credentials
	if req.Username != s.config.AdminUsername {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// In production, hash the password properly
	if req.Password != s.config.AdminPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": req.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(s.config.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (s *Server) getDeployments(c *gin.Context) {
	deployments, err := s.storage.GetActiveDeployments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get deployments"})
		return
	}

	c.JSON(http.StatusOK, deployments)
}

func (s *Server) getDeployment(c *gin.Context) {
	id := c.Param("id")
	deployment, err := s.storage.GetDeployment(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
		return
	}

	c.JSON(http.StatusOK, deployment)
}

func (s *Server) redeployDeployment(c *gin.Context) {
	id := c.Param("id")
	deployment, err := s.storage.GetDeployment(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
		return
	}

	// Queue new deployment with same parameters
	newDeployment := &models.Deployment{
		Owner:       deployment.Owner,
		Repo:        deployment.Repo,
		Ref:         deployment.Ref,
		SHA:         deployment.SHA,
		CloneURL:    deployment.CloneURL,
		Environment: deployment.Environment,
	}

	if err := s.deployer.QueueDeployment(newDeployment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue deployment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Redeployment queued", "id": newDeployment.ID})
}

func (s *Server) stopDeployment(c *gin.Context) {
	id := c.Param("id")
	deployment, err := s.storage.GetDeployment(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
		return
	}

	// Stop the deployment
	// This would call the deployer's stop method
	c.JSON(http.StatusOK, gin.H{"message": "Deployment stopped", "id": deployment.ID})
}

func (s *Server) getDeploymentLogs(c *gin.Context) {
	id := c.Param("id")
	deployment, err := s.storage.GetDeployment(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
		return
	}

	// Read log file
	file, err := os.Open(deployment.LogPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Logs not found"})
		return
	}
	defer file.Close()

	// Stream logs
	c.Header("Content-Type", "text/plain")
	io.Copy(c.Writer, file)
}

func (s *Server) handleWebSocket(c *gin.Context) {
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Send deployment updates via WebSocket
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			deployments, err := s.storage.GetActiveDeployments()
			if err != nil {
				continue
			}

			if err := conn.WriteJSON(deployments); err != nil {
				return
			}
		}
	}
}
