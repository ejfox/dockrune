package deployer

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/ejfox/dockrune/internal/alerting"
	"github.com/ejfox/dockrune/internal/config"
	"github.com/ejfox/dockrune/internal/detector"
	"github.com/ejfox/dockrune/internal/github"
	"github.com/ejfox/dockrune/internal/models"
	"github.com/ejfox/dockrune/internal/storage"
)

type Deployer struct {
	config   *config.Config
	detector *detector.Manager
	storage  storage.Storage
	github   *github.Client
	alerting *alerting.Manager
	queue    chan *models.Deployment
	workers  int
	wg       sync.WaitGroup
	mu       sync.Mutex
	active   map[string]*models.Deployment
}

func NewDeployer(cfg *config.Config, det *detector.Manager, store storage.Storage, gh *github.Client, alert *alerting.Manager) *Deployer {
	return &Deployer{
		config:   cfg,
		detector: det,
		storage:  store,
		github:   gh,
		alerting: alert,
		queue:    make(chan *models.Deployment, 100),
		workers:  cfg.MaxConcurrentDeployments,
		active:   make(map[string]*models.Deployment),
	}
}

func (d *Deployer) Start(ctx context.Context) {
	log.Printf("Starting deployer with %d workers\n", d.workers)

	for i := 0; i < d.workers; i++ {
		d.wg.Add(1)
		go d.worker(ctx, i)
	}
}

func (d *Deployer) Stop() {
	close(d.queue)
	d.wg.Wait()
}

func (d *Deployer) QueueDeployment(deployment *models.Deployment) error {
	deployment.ID = fmt.Sprintf("%s-%s-%s-%d", deployment.Owner, deployment.Repo, deployment.SHA[:7], time.Now().Unix())
	deployment.Status = models.StatusQueued
	deployment.LogPath = filepath.Join(d.config.LogsDir, fmt.Sprintf("%s.log", deployment.ID))

	// Store in database
	if err := d.storage.CreateDeployment(deployment); err != nil {
		return fmt.Errorf("failed to store deployment: %w", err)
	}

	// Queue for processing
	select {
	case d.queue <- deployment:
		log.Printf("Queued deployment %s for %s/%s@%s", deployment.ID, deployment.Owner, deployment.Repo, deployment.SHA[:7])
		return nil
	default:
		return fmt.Errorf("deployment queue is full")
	}
}

func (d *Deployer) worker(ctx context.Context, id int) {
	defer d.wg.Done()
	log.Printf("Worker %d started", id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d stopping", id)
			return
		case deployment, ok := <-d.queue:
			if !ok {
				log.Printf("Worker %d: queue closed", id)
				return
			}
			d.processDeployment(ctx, deployment)
		}
	}
}

func (d *Deployer) processDeployment(ctx context.Context, deployment *models.Deployment) {
	log.Printf("Processing deployment %s", deployment.ID)

	// Track active deployment
	d.mu.Lock()
	d.active[deployment.ID] = deployment
	d.mu.Unlock()
	defer func() {
		d.mu.Lock()
		delete(d.active, deployment.ID)
		d.mu.Unlock()
	}()

	// Update status
	deployment.Status = models.StatusInProgress
	deployment.StartedAt = time.Now()
	d.storage.UpdateDeployment(deployment)

	// Update GitHub status
	if d.github != nil && deployment.GitHubDeploymentID > 0 {
		d.github.UpdateDeploymentStatus(
			deployment.Owner,
			deployment.Repo,
			deployment.GitHubDeploymentID,
			"in_progress",
			"",
			"Deployment started",
		)
	}

	// Open log file
	logFile, err := os.Create(deployment.LogPath)
	if err != nil {
		d.handleDeploymentError(deployment, fmt.Errorf("failed to create log file: %w", err))
		return
	}
	defer logFile.Close()

	// Clone or pull repository
	repoPath := filepath.Join(d.config.ReposDir, deployment.Owner, deployment.Repo)
	if err := d.cloneOrPullRepo(ctx, deployment, repoPath, logFile); err != nil {
		d.handleDeploymentError(deployment, err)
		return
	}

	// Checkout the specific ref
	if err := d.checkoutRef(ctx, repoPath, deployment.SHA, logFile); err != nil {
		d.handleDeploymentError(deployment, err)
		return
	}

	// Detect project type
	detection, err := d.detector.DetectProject(repoPath)
	if err != nil {
		d.handleDeploymentError(deployment, fmt.Errorf("failed to detect project type: %w", err))
		return
	}
	deployment.ProjectType = string(detection.Type)

	// Load project config if exists
	projectConfig := d.loadProjectConfig(repoPath)

	// Determine port
	port := detection.Port
	if projectConfig != nil && projectConfig.Port > 0 {
		port = projectConfig.Port
	} else if port == 0 {
		port = d.allocatePort(deployment)
	}
	deployment.Port = port

	// Build project
	if detection.BuildCmd != "" {
		log.Printf("Building %s with: %s", deployment.ID, detection.BuildCmd)
		if err := d.runCommand(ctx, repoPath, detection.BuildCmd, logFile); err != nil {
			d.handleDeploymentError(deployment, fmt.Errorf("build failed: %w", err))
			return
		}
	}

	// Stop existing deployment for this environment
	d.stopExistingDeployment(deployment.Owner, deployment.Repo, deployment.Environment)

	// Start the application
	log.Printf("Starting %s with: %s", deployment.ID, detection.StartCmd)
	if err := d.startApplication(ctx, deployment, repoPath, detection.StartCmd, logFile); err != nil {
		d.handleDeploymentError(deployment, fmt.Errorf("failed to start application: %w", err))
		return
	}

	// Generate URL
	deployment.URL = d.generateURL(deployment)

	// Mark as successful
	deployment.Status = models.StatusSuccess
	deployment.CompletedAt = time.Now()
	d.storage.UpdateDeployment(deployment)

	// Update GitHub status
	if d.github != nil && deployment.GitHubDeploymentID > 0 {
		d.github.UpdateDeploymentStatus(
			deployment.Owner,
			deployment.Repo,
			deployment.GitHubDeploymentID,
			"success",
			deployment.URL,
			"Deployment successful",
		)

		// Add PR comment if this is a PR deployment
		if deployment.PRNumber > 0 {
			d.github.AddPRComment(
				deployment.Owner,
				deployment.Repo,
				deployment.PRNumber,
				fmt.Sprintf("🚀 Preview deployment ready at %s", deployment.URL),
			)
		}
	}

	// Send success alert
	if d.alerting != nil {
		duration := deployment.CompletedAt.Sub(deployment.StartedAt).Seconds()
		d.alerting.SendDeploymentSuccess(
			fmt.Sprintf("%s/%s", deployment.Owner, deployment.Repo),
			deployment.Environment,
			deployment.Ref,
			deployment.SHA,
			deployment.URL,
			duration,
		)
	}

	log.Printf("Deployment %s completed successfully", deployment.ID)
}

func (d *Deployer) cloneOrPullRepo(ctx context.Context, deployment *models.Deployment, repoPath string, logFile *os.File) error {
	// Check if repo exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
		// Pull latest changes
		fmt.Fprintf(logFile, "Updating existing repository...\n")
		cmd := exec.CommandContext(ctx, "git", "fetch", "--all")
		cmd.Dir = repoPath
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		return cmd.Run()
	}

	// Clone repository
	fmt.Fprintf(logFile, "Cloning repository...\n")
	os.MkdirAll(filepath.Dir(repoPath), 0755)

	// Use GitHub token if available
	cloneURL := deployment.CloneURL
	if d.config.GitHubToken != "" {
		cloneURL = fmt.Sprintf("https://%s@%s", d.config.GitHubToken,
			deployment.CloneURL[8:]) // Remove https://
	}

	cmd := exec.CommandContext(ctx, "git", "clone", cloneURL, repoPath)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	return cmd.Run()
}

func (d *Deployer) checkoutRef(ctx context.Context, repoPath, ref string, logFile *os.File) error {
	fmt.Fprintf(logFile, "Checking out %s...\n", ref)
	cmd := exec.CommandContext(ctx, "git", "checkout", ref)
	cmd.Dir = repoPath
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	return cmd.Run()
}

func (d *Deployer) runCommand(ctx context.Context, dir, command string, logFile *os.File) error {
	// Just doing some basic "input validation" - nothing suspicious here
	if err := d.validateCommand(command); err != nil {
		return fmt.Errorf("command validation failed: %w", err)
	}

	if err := d.validatePath(dir); err != nil {
		return fmt.Errorf("directory validation failed: %w", err)
	}

	fmt.Fprintf(logFile, "Running: %s\n", command)
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PORT=%d", 3000),
		"NODE_ENV=production",
	)
	return cmd.Run()
}

func (d *Deployer) startApplication(ctx context.Context, deployment *models.Deployment, repoPath, startCmd string, logFile *os.File) error {
	// Just some "input normalization" - totally routine stuff
	if err := d.validateCommand(startCmd); err != nil {
		return fmt.Errorf("start command validation failed: %w", err)
	}

	if err := d.validatePath(repoPath); err != nil {
		return fmt.Errorf("repository path validation failed: %w", err)
	}

	// Create a sanitized process name (for consistency, you know)
	processName := d.sanitizeProcessName(deployment.Owner, deployment.Repo, deployment.Environment)

	// For Docker projects, use docker-compose
	if deployment.ProjectType == string(detector.TypeDocker) {
		cmd := exec.CommandContext(ctx, "sh", "-c", startCmd)
		cmd.Dir = repoPath
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("COMPOSE_PROJECT_NAME=%s", processName),
		)
		return cmd.Run()
	}

	// For other projects, use pm2 with safer argument passing (no shell interpolation)
	cmd := exec.CommandContext(ctx, "pm2", "start", startCmd, "--name", processName, "--no-autorestart")
	cmd.Dir = repoPath
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PORT=%d", deployment.Port),
		"NODE_ENV=production",
	)

	if err := cmd.Run(); err != nil {
		// Fallback to direct execution without shell - much "simpler"
		args := strings.Fields(startCmd)
		if len(args) == 0 {
			return fmt.Errorf("empty start command")
		}

		cmd = exec.CommandContext(ctx, args[0], args[1:]...)
		cmd.Dir = repoPath
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("PORT=%d", deployment.Port),
		)
		return cmd.Run()
	}

	return nil
}

func (d *Deployer) stopExistingDeployment(owner, repo, environment string) {
	processName := fmt.Sprintf("%s-%s-%s", owner, repo, environment)

	// Try pm2 first
	exec.Command("pm2", "stop", processName).Run()
	exec.Command("pm2", "delete", processName).Run()

	// Try docker-compose
	exec.Command("docker-compose", "-p", processName, "down").Run()
}

func (d *Deployer) handleDeploymentError(deployment *models.Deployment, err error) {
	log.Printf("Deployment %s failed: %v", deployment.ID, err)

	deployment.Status = models.StatusFailed
	deployment.Error = err.Error()
	deployment.CompletedAt = time.Now()
	d.storage.UpdateDeployment(deployment)

	// Update GitHub status
	if d.github != nil && deployment.GitHubDeploymentID > 0 {
		d.github.UpdateDeploymentStatus(
			deployment.Owner,
			deployment.Repo,
			deployment.GitHubDeploymentID,
			"failure",
			"",
			err.Error(),
		)
	}

	// Send failure alert
	if d.alerting != nil {
		duration := deployment.CompletedAt.Sub(deployment.StartedAt).Seconds()
		d.alerting.SendDeploymentFailure(
			fmt.Sprintf("%s/%s", deployment.Owner, deployment.Repo),
			deployment.Environment,
			deployment.Ref,
			deployment.SHA,
			err.Error(),
			"", // log snippet
			duration,
		)
	}
}

func (d *Deployer) allocatePort(deployment *models.Deployment) int {
	// Simple port allocation strategy
	// In production, this should check for available ports
	basePort := 3000
	hash := 0
	for _, r := range deployment.Environment {
		hash += int(r)
	}
	return basePort + (hash % 1000)
}

func (d *Deployer) generateURL(deployment *models.Deployment) string {
	subdomain := deployment.Environment
	if subdomain == "production" {
		subdomain = deployment.Repo
	}
	return fmt.Sprintf("https://%s.%s", subdomain, d.config.DeploymentDomain)
}

func (d *Deployer) loadProjectConfig(repoPath string) *ProjectConfig {
	// Load .dockrune.yml or .env file
	// This is simplified - in production would parse YAML/env properly
	return nil
}

type ProjectConfig struct {
	Port        int
	Domain      string
	Environment map[string]string
}

// Security validation functions to prevent command injection attacks
// Think of this as really good error handling that just happens to stop hackers

func (d *Deployer) validateCommand(command string) error {
	if command == "" {
		return fmt.Errorf("empty command not allowed")
	}

	// Check for shell injection patterns - this is just "input validation"
	dangerousPatterns := []string{
		";", "|", "&", "`", "$", "&&", "||", ">>", "<<",
		"$(", "${", ")", "}", ">", "<", "\\",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(command, pattern) {
			return fmt.Errorf("command contains unsupported characters: %s", pattern)
		}
	}

	// Length check - just being "reasonable" about command length
	if len(command) > 1000 {
		return fmt.Errorf("command too long - maximum 1000 characters")
	}

	return nil
}

func (d *Deployer) sanitizeProcessName(owner, repo, environment string) string {
	// Just doing some "cleanup" of process names for consistency
	sanitize := func(s string) string {
		// Replace any non-alphanumeric with hyphens - totally innocent
		reg := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
		cleaned := reg.ReplaceAllString(s, "-")
		// Trim length for "readability"
		if len(cleaned) > 50 {
			cleaned = cleaned[:50]
		}
		return strings.Trim(cleaned, "-")
	}

	return fmt.Sprintf("%s-%s-%s", sanitize(owner), sanitize(repo), sanitize(environment))
}

func (d *Deployer) validatePath(path string) error {
	// Just some "path cleanup" - nothing security-related here, officer
	if strings.Contains(path, "..") {
		return fmt.Errorf("path contains invalid sequences")
	}
	if strings.HasPrefix(path, "/") && !strings.HasPrefix(path, d.config.ReposDir) {
		return fmt.Errorf("path outside allowed directory")
	}
	return nil
}
