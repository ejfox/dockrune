package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ejfox/dockrune/internal/admin"
	"github.com/ejfox/dockrune/internal/alerting"
	"github.com/ejfox/dockrune/internal/config"
	"github.com/ejfox/dockrune/internal/deployer"
	"github.com/ejfox/dockrune/internal/detector"
	"github.com/ejfox/dockrune/internal/github"
	"github.com/ejfox/dockrune/internal/storage"
	"github.com/ejfox/dockrune/internal/webhook"
	"github.com/spf13/cobra"
)

func ServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the dockrune server",
		Long:  `Start the webhook server and admin dashboard`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer()
		},
	}
}

func runServer() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize storage
	store, err := storage.NewSQLiteStorage(cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	// Initialize components
	detectorManager := detector.NewManager()

	var githubClient *github.Client
	if cfg.GitHubToken != "" {
		githubClient = github.NewClient(cfg.GitHubToken)
	}

	alertManager := alerting.NewManager(cfg.DiscordWebhookURL, cfg.N8NWebhookURL)

	// Initialize deployer
	deployerInstance := deployer.NewDeployer(cfg, detectorManager, store, githubClient, alertManager)

	// Start deployer workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deployerInstance.Start(ctx)
	defer deployerInstance.Stop()

	// Start webhook server
	webhookServer := webhook.NewServer(cfg, deployerInstance, githubClient)
	go func() {
		if err := webhookServer.Start(); err != nil {
			log.Printf("Webhook server error: %v", err)
		}
	}()

	// Start admin dashboard
	adminServer := admin.NewServer(cfg, store, deployerInstance)
	go func() {
		if err := adminServer.Start(); err != nil {
			log.Printf("Admin server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Printf("dockrune server started")
	log.Printf("Webhook server: http://localhost:%d", cfg.WebhookPort)
	log.Printf("Admin dashboard: http://localhost:%d", cfg.AdminPort)

	<-sigChan
	log.Println("Shutting down...")

	return nil
}
