package models

import (
	"time"
)

type DeploymentStatus string

const (
	StatusQueued     DeploymentStatus = "queued"
	StatusInProgress DeploymentStatus = "in_progress"
	StatusSuccess    DeploymentStatus = "success"
	StatusFailed     DeploymentStatus = "failed"
)

type Deployment struct {
	ID                 string
	Owner              string
	Repo               string
	Ref                string
	SHA                string
	CloneURL           string
	Environment        string
	PRNumber           int
	GitHubDeploymentID int64
	Status             DeploymentStatus
	StartedAt          time.Time
	CompletedAt        time.Time
	LogPath            string
	URL                string
	Port               int
	ProjectType        string
	Error              string
}
