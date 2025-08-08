package storage

import (
	"database/sql"
	"fmt"

	"github.com/ejfox/dockrune/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

type Storage interface {
	CreateDeployment(d *models.Deployment) error
	UpdateDeployment(d *models.Deployment) error
	GetDeployment(id string) (*models.Deployment, error)
	ListDeployments(limit int) ([]*models.Deployment, error)
	GetActiveDeployments() ([]*models.Deployment, error)
	Close() error
}

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(path string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	s := &SQLiteStorage{db: db}
	if err := s.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return s, nil
}

func (s *SQLiteStorage) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS deployments (
		id TEXT PRIMARY KEY,
		owner TEXT NOT NULL,
		repo TEXT NOT NULL,
		ref TEXT NOT NULL,
		sha TEXT NOT NULL,
		clone_url TEXT NOT NULL,
		environment TEXT NOT NULL,
		pr_number INTEGER,
		github_deployment_id INTEGER,
		status TEXT NOT NULL,
		started_at DATETIME,
		completed_at DATETIME,
		log_path TEXT,
		url TEXT,
		port INTEGER,
		project_type TEXT,
		error TEXT,
		metadata TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_deployments_status ON deployments(status);
	CREATE INDEX IF NOT EXISTS idx_deployments_repo ON deployments(owner, repo);
	CREATE INDEX IF NOT EXISTS idx_deployments_environment ON deployments(environment);
	`

	_, err := s.db.Exec(query)
	return err
}

func (s *SQLiteStorage) CreateDeployment(d *models.Deployment) error {
	query := `
	INSERT INTO deployments (
		id, owner, repo, ref, sha, clone_url, environment,
		pr_number, github_deployment_id, status, started_at,
		log_path, port, project_type
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		d.ID, d.Owner, d.Repo, d.Ref, d.SHA, d.CloneURL, d.Environment,
		d.PRNumber, d.GitHubDeploymentID, d.Status, d.StartedAt,
		d.LogPath, d.Port, d.ProjectType,
	)

	return err
}

func (s *SQLiteStorage) UpdateDeployment(d *models.Deployment) error {
	query := `
	UPDATE deployments SET
		status = ?,
		completed_at = ?,
		url = ?,
		port = ?,
		project_type = ?,
		error = ?,
		updated_at = CURRENT_TIMESTAMP
	WHERE id = ?
	`

	_, err := s.db.Exec(query,
		d.Status, d.CompletedAt, d.URL, d.Port, d.ProjectType, d.Error, d.ID,
	)

	return err
}

func (s *SQLiteStorage) GetDeployment(id string) (*models.Deployment, error) {
	query := `
	SELECT id, owner, repo, ref, sha, clone_url, environment,
		pr_number, github_deployment_id, status, started_at, completed_at,
		log_path, url, port, project_type, error
	FROM deployments
	WHERE id = ?
	`

	var d models.Deployment
	var completedAt sql.NullTime
	var url, projectType, errorMsg sql.NullString
	var prNumber, port sql.NullInt64

	err := s.db.QueryRow(query, id).Scan(
		&d.ID, &d.Owner, &d.Repo, &d.Ref, &d.SHA, &d.CloneURL, &d.Environment,
		&prNumber, &d.GitHubDeploymentID, &d.Status, &d.StartedAt, &completedAt,
		&d.LogPath, &url, &port, &projectType, &errorMsg,
	)

	if err != nil {
		return nil, err
	}

	if completedAt.Valid {
		d.CompletedAt = completedAt.Time
	}
	if url.Valid {
		d.URL = url.String
	}
	if prNumber.Valid {
		d.PRNumber = int(prNumber.Int64)
	}
	if port.Valid {
		d.Port = int(port.Int64)
	}
	if errorMsg.Valid {
		d.Error = errorMsg.String
	}

	return &d, nil
}

func (s *SQLiteStorage) ListDeployments(limit int) ([]*models.Deployment, error) {
	query := `
	SELECT id, owner, repo, ref, sha, environment, status, 
		started_at, completed_at, url, project_type
	FROM deployments
	ORDER BY created_at DESC
	LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*models.Deployment
	for rows.Next() {
		var d models.Deployment
		var completedAt sql.NullTime
		var url, projectType sql.NullString

		err := rows.Scan(
			&d.ID, &d.Owner, &d.Repo, &d.Ref, &d.SHA, &d.Environment,
			&d.Status, &d.StartedAt, &completedAt, &url, &projectType,
		)
		if err != nil {
			continue
		}

		if completedAt.Valid {
			d.CompletedAt = completedAt.Time
		}
		if url.Valid {
			d.URL = url.String
		}

		deployments = append(deployments, &d)
	}

	return deployments, nil
}

func (s *SQLiteStorage) GetActiveDeployments() ([]*models.Deployment, error) {
	query := `
	SELECT id, owner, repo, ref, sha, environment, status,
		started_at, url, port, project_type
	FROM deployments
	WHERE status IN ('in_progress', 'success')
	ORDER BY started_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*models.Deployment
	for rows.Next() {
		var d models.Deployment
		var url, projectType sql.NullString
		var port sql.NullInt64

		err := rows.Scan(
			&d.ID, &d.Owner, &d.Repo, &d.Ref, &d.SHA, &d.Environment,
			&d.Status, &d.StartedAt, &url, &port, &projectType,
		)
		if err != nil {
			continue
		}

		if url.Valid {
			d.URL = url.String
		}
		if port.Valid {
			d.Port = int(port.Int64)
		}

		deployments = append(deployments, &d)
	}

	return deployments, nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
