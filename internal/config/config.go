package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	// Server
	WebhookPort int
	AdminPort   int

	// GitHub
	GitHubToken   string
	WebhookSecret string

	// Deployment
	DeploymentDomain         string
	MaxConcurrentDeployments int

	// Storage
	DatabasePath string
	ReposDir     string
	LogsDir      string

	// Alerting
	DiscordWebhookURL string
	N8NWebhookURL     string

	// Admin
	AdminUsername string
	AdminPassword string
	JWTSecret     string
}

func Load() (*Config, error) {
	// Load .env file if exists
	godotenv.Load()

	// Set defaults
	viper.SetDefault("webhook_port", 8000)
	viper.SetDefault("admin_port", 8001)
	viper.SetDefault("max_concurrent_deployments", 5)
	viper.SetDefault("database_path", "./data/dockrune.db")
	viper.SetDefault("repos_dir", "./repos")
	viper.SetDefault("logs_dir", "./logs")
	viper.SetDefault("deployment_domain", "localhost")

	// Bind environment variables
	viper.BindEnv("webhook_port", "WEBHOOK_PORT")
	viper.BindEnv("admin_port", "ADMIN_PORT")
	viper.BindEnv("github_token", "GITHUB_TOKEN")
	viper.BindEnv("webhook_secret", "GITHUB_WEBHOOK_SECRET")
	viper.BindEnv("discord_webhook_url", "DISCORD_WEBHOOK_URL")
	viper.BindEnv("n8n_webhook_url", "N8N_WEBHOOK_URL")
	viper.BindEnv("admin_username", "ADMIN_USERNAME")
	viper.BindEnv("admin_password", "ADMIN_PASSWORD")
	viper.BindEnv("jwt_secret", "JWT_SECRET")
	viper.BindEnv("deployment_domain", "DEPLOYMENT_DOMAIN")

	// Load config file if exists
	viper.SetConfigName("dockrune")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/dockrune/")
	viper.ReadInConfig()

	cfg := &Config{
		WebhookPort:              viper.GetInt("webhook_port"),
		AdminPort:                viper.GetInt("admin_port"),
		GitHubToken:              viper.GetString("github_token"),
		WebhookSecret:            viper.GetString("webhook_secret"),
		DeploymentDomain:         viper.GetString("deployment_domain"),
		MaxConcurrentDeployments: viper.GetInt("max_concurrent_deployments"),
		DatabasePath:             viper.GetString("database_path"),
		ReposDir:                 viper.GetString("repos_dir"),
		LogsDir:                  viper.GetString("logs_dir"),
		DiscordWebhookURL:        viper.GetString("discord_webhook_url"),
		N8NWebhookURL:            viper.GetString("n8n_webhook_url"),
		AdminUsername:            viper.GetString("admin_username"),
		AdminPassword:            viper.GetString("admin_password"),
		JWTSecret:                viper.GetString("jwt_secret"),
	}

	// Validate required fields
	if cfg.WebhookSecret == "" {
		return nil, fmt.Errorf("GITHUB_WEBHOOK_SECRET is required")
	}

	// Create directories
	os.MkdirAll(cfg.ReposDir, 0755)
	os.MkdirAll(cfg.LogsDir, 0755)
	os.MkdirAll(getDir(cfg.DatabasePath), 0755)

	return cfg, nil
}

func getDir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}
