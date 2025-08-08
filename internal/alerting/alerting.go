package alerting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Manager struct {
	discordURL string
	n8nURL     string
}

func NewManager(discordURL, n8nURL string) *Manager {
	return &Manager{
		discordURL: discordURL,
		n8nURL:     n8nURL,
	}
}

func (m *Manager) SendDeploymentSuccess(project, environment, ref, sha, url string, duration float64) error {
	if m.discordURL != "" {
		embed := DiscordEmbed{
			Title:       "Deployment Successful ✅",
			Description: fmt.Sprintf("**%s** deployed to **%s**", project, environment),
			Color:       0x28a745, // Green
			Fields: []DiscordField{
				{Name: "Project", Value: project, Inline: true},
				{Name: "Environment", Value: environment, Inline: true},
				{Name: "Branch", Value: ref, Inline: true},
				{Name: "Commit", Value: fmt.Sprintf("`%s`", sha[:7]), Inline: true},
				{Name: "Duration", Value: fmt.Sprintf("%.1fs", duration), Inline: true},
				{Name: "URL", Value: fmt.Sprintf("[View Deployment](%s)", url), Inline: false},
			},
			Timestamp: time.Now().Format(time.RFC3339),
			Footer: DiscordFooter{
				Text: "dockrune deployment system",
			},
		}

		if err := m.sendDiscordWebhook(embed); err != nil {
			return fmt.Errorf("failed to send Discord alert: %w", err)
		}
	}

	if m.n8nURL != "" {
		payload := map[string]interface{}{
			"event":       "deployment_success",
			"project":     project,
			"environment": environment,
			"ref":         ref,
			"sha":         sha,
			"url":         url,
			"duration":    duration,
			"timestamp":   time.Now().Unix(),
		}

		if err := m.sendN8NWebhook(payload); err != nil {
			return fmt.Errorf("failed to send n8n alert: %w", err)
		}
	}

	return nil
}

func (m *Manager) SendDeploymentFailure(project, environment, ref, sha, error, logSnippet string, duration float64) error {
	if m.discordURL != "" {
		fields := []DiscordField{
			{Name: "Project", Value: project, Inline: true},
			{Name: "Environment", Value: environment, Inline: true},
			{Name: "Branch", Value: ref, Inline: true},
			{Name: "Commit", Value: fmt.Sprintf("`%s`", sha[:7]), Inline: true},
			{Name: "Error", Value: error, Inline: false},
		}

		if duration > 0 {
			fields = append(fields, DiscordField{
				Name: "Duration", Value: fmt.Sprintf("%.1fs", duration), Inline: true,
			})
		}

		if logSnippet != "" {
			snippet := logSnippet
			if len(snippet) > 500 {
				snippet = snippet[:500] + "..."
			}
			fields = append(fields, DiscordField{
				Name: "Logs", Value: fmt.Sprintf("```\n%s\n```", snippet), Inline: false,
			})
		}

		embed := DiscordEmbed{
			Title:       "Deployment Failed ❌",
			Description: fmt.Sprintf("**%s** failed to deploy to **%s**", project, environment),
			Color:       0xdc3545, // Red
			Fields:      fields,
			Timestamp:   time.Now().Format(time.RFC3339),
			Footer: DiscordFooter{
				Text: "dockrune deployment system",
			},
		}

		if err := m.sendDiscordWebhook(embed); err != nil {
			return fmt.Errorf("failed to send Discord alert: %w", err)
		}
	}

	if m.n8nURL != "" {
		payload := map[string]interface{}{
			"event":       "deployment_failure",
			"project":     project,
			"environment": environment,
			"ref":         ref,
			"sha":         sha,
			"error":       error,
			"duration":    duration,
			"timestamp":   time.Now().Unix(),
		}

		if err := m.sendN8NWebhook(payload); err != nil {
			return fmt.Errorf("failed to send n8n alert: %w", err)
		}
	}

	return nil
}

func (m *Manager) sendDiscordWebhook(embed DiscordEmbed) error {
	webhook := DiscordWebhook{
		Username: "dockrune",
		Embeds:   []DiscordEmbed{embed},
	}

	data, err := json.Marshal(webhook)
	if err != nil {
		return err
	}

	resp, err := http.Post(m.discordURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Discord webhook returned status %d", resp.StatusCode)
	}

	return nil
}

func (m *Manager) sendN8NWebhook(payload map[string]interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(m.n8nURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("n8n webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// Discord webhook structures
type DiscordWebhook struct {
	Username string         `json:"username"`
	Content  string         `json:"content,omitempty"`
	Embeds   []DiscordEmbed `json:"embeds"`
}

type DiscordEmbed struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Color       int            `json:"color"`
	Fields      []DiscordField `json:"fields"`
	Timestamp   string         `json:"timestamp"`
	Footer      DiscordFooter  `json:"footer"`
}

type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type DiscordFooter struct {
	Text string `json:"text"`
}
