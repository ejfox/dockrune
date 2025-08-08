package storage

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ejfox/dockrune/internal/models"
)

func TestSQLiteStorage(t *testing.T) {
	// Create temp database
	dbFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	dbPath := dbFile.Name()
	dbFile.Close()
	defer os.Remove(dbPath)

	// Initialize storage
	store, err := NewSQLiteStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Test deployment lifecycle
	t.Run("CreateDeployment", func(t *testing.T) {
		deployment := &models.Deployment{
			ID:          "test-123",
			Owner:       "testuser",
			Repo:        "testrepo",
			Ref:         "main",
			SHA:         "abc123def456",
			CloneURL:    "https://github.com/testuser/testrepo.git",
			Environment: "production",
			Status:      models.StatusQueued,
			StartedAt:   time.Now(),
			LogPath:     "/logs/test-123.log",
			Port:        3000,
			ProjectType: "docker",
		}

		err := store.CreateDeployment(deployment)
		if err != nil {
			t.Errorf("CreateDeployment() error = %v", err)
		}
	})

	t.Run("GetDeployment", func(t *testing.T) {
		deployment, err := store.GetDeployment("test-123")
		if err != nil {
			t.Fatalf("GetDeployment() error = %v", err)
		}

		if deployment.Owner != "testuser" {
			t.Errorf("Owner = %v, want %v", deployment.Owner, "testuser")
		}
		if deployment.Repo != "testrepo" {
			t.Errorf("Repo = %v, want %v", deployment.Repo, "testrepo")
		}
		if deployment.Status != models.StatusQueued {
			t.Errorf("Status = %v, want %v", deployment.Status, models.StatusQueued)
		}
	})

	t.Run("UpdateDeployment", func(t *testing.T) {
		deployment, _ := store.GetDeployment("test-123")
		deployment.Status = models.StatusSuccess
		deployment.CompletedAt = time.Now()
		deployment.URL = "https://test.example.com"

		err := store.UpdateDeployment(deployment)
		if err != nil {
			t.Errorf("UpdateDeployment() error = %v", err)
		}

		// Verify update
		updated, _ := store.GetDeployment("test-123")
		if updated.Status != models.StatusSuccess {
			t.Errorf("Status = %v, want %v", updated.Status, models.StatusSuccess)
		}
		if updated.URL != "https://test.example.com" {
			t.Errorf("URL = %v, want %v", updated.URL, "https://test.example.com")
		}
	})

	t.Run("ListDeployments", func(t *testing.T) {
		// Add more deployments
		for i := 0; i < 5; i++ {
			d := &models.Deployment{
				ID:          fmt.Sprintf("test-%d", i),
				Owner:       "testuser",
				Repo:        "testrepo",
				Ref:         "main",
				SHA:         fmt.Sprintf("sha%d", i),
				CloneURL:    "https://github.com/testuser/testrepo.git",
				Environment: "production",
				Status:      models.StatusSuccess,
				StartedAt:   time.Now().Add(-time.Duration(i) * time.Hour),
			}
			store.CreateDeployment(d)
		}

		deployments, err := store.ListDeployments(10)
		if err != nil {
			t.Errorf("ListDeployments() error = %v", err)
		}

		if len(deployments) < 5 {
			t.Errorf("Expected at least 5 deployments, got %d", len(deployments))
		}
	})

	t.Run("GetActiveDeployments", func(t *testing.T) {
		// Add an in-progress deployment
		active := &models.Deployment{
			ID:          "active-1",
			Owner:       "testuser",
			Repo:        "testrepo",
			Ref:         "feature",
			SHA:         "xyz789",
			CloneURL:    "https://github.com/testuser/testrepo.git",
			Environment: "staging",
			Status:      models.StatusInProgress,
			StartedAt:   time.Now(),
		}
		store.CreateDeployment(active)

		deployments, err := store.GetActiveDeployments()
		if err != nil {
			t.Errorf("GetActiveDeployments() error = %v", err)
		}

		hasActive := false
		for _, d := range deployments {
			if d.Status == models.StatusInProgress {
				hasActive = true
				break
			}
		}

		if !hasActive {
			t.Error("Expected to find in-progress deployment")
		}
	})
}

func TestSQLiteStorageErrors(t *testing.T) {
	t.Run("InvalidPath", func(t *testing.T) {
		_, err := NewSQLiteStorage("/invalid/path/to/database.db")
		if err == nil {
			t.Error("Expected error for invalid path")
		}
	})

	t.Run("GetNonExistent", func(t *testing.T) {
		dbFile, _ := os.CreateTemp("", "test-*.db")
		dbPath := dbFile.Name()
		dbFile.Close()
		defer os.Remove(dbPath)

		store, _ := NewSQLiteStorage(dbPath)
		defer store.Close()

		_, err := store.GetDeployment("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent deployment")
		}
	})
}
