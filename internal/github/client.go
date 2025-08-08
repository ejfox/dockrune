package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
	ctx    context.Context
}

func NewClient(token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		client: github.NewClient(tc),
		ctx:    ctx,
	}
}

func (c *Client) CreateDeployment(owner, repo, ref, environment string) (int64, error) {
	req := &github.DeploymentRequest{
		Ref:              github.String(ref),
		Environment:      github.String(environment),
		AutoMerge:        github.Bool(false),
		RequiredContexts: &[]string{},
		Description:      github.String(fmt.Sprintf("Deploying %s via dockrune", ref[:7])),
	}

	deployment, _, err := c.client.Repositories.CreateDeployment(c.ctx, owner, repo, req)
	if err != nil {
		return 0, fmt.Errorf("failed to create deployment: %w", err)
	}

	return deployment.GetID(), nil
}

func (c *Client) UpdateDeploymentStatus(owner, repo string, deploymentID int64, state, targetURL, description string) error {
	req := &github.DeploymentStatusRequest{
		State:       github.String(state),
		Description: github.String(description),
	}

	if targetURL != "" {
		req.EnvironmentURL = github.String(targetURL)
		req.LogURL = github.String(targetURL + "/logs")
	}

	_, _, err := c.client.Repositories.CreateDeploymentStatus(c.ctx, owner, repo, deploymentID, req)
	if err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	return nil
}

func (c *Client) CreateCommitStatus(owner, repo, sha, state, targetURL, description, context string) error {
	status := &github.RepoStatus{
		State:       github.String(state),
		Description: github.String(description),
		Context:     github.String(context),
	}

	if targetURL != "" {
		status.TargetURL = github.String(targetURL)
	}

	_, _, err := c.client.Repositories.CreateStatus(c.ctx, owner, repo, sha, status)
	if err != nil {
		return fmt.Errorf("failed to create commit status: %w", err)
	}

	return nil
}

func (c *Client) AddPRComment(owner, repo string, prNumber int, body string) error {
	comment := &github.IssueComment{
		Body: github.String(body),
	}

	_, _, err := c.client.Issues.CreateComment(c.ctx, owner, repo, prNumber, comment)
	if err != nil {
		return fmt.Errorf("failed to add PR comment: %w", err)
	}

	return nil
}
