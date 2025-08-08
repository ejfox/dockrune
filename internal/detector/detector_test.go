package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDockerDetector(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		wantType ProjectType
		wantConf float32
	}{
		{
			name: "docker-compose project",
			files: map[string]string{
				"docker-compose.yml": "version: '3'\nservices:\n  app:\n    image: node:16",
			},
			wantType: TypeDocker,
			wantConf: 1.0,
		},
		{
			name: "dockerfile only",
			files: map[string]string{
				"Dockerfile": "FROM node:16\nWORKDIR /app",
			},
			wantType: TypeDocker,
			wantConf: 0.9,
		},
		{
			name:     "no docker files",
			files:    map[string]string{},
			wantType: TypeUnknown,
			wantConf: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir := t.TempDir()

			// Create test files
			for filename, content := range tt.files {
				path := filepath.Join(tmpDir, filename)
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			// Test detection
			detector := &DockerDetector{}
			detection, err := detector.Detect(tmpDir)

			if tt.wantType == TypeUnknown {
				if detection != nil {
					t.Errorf("Expected nil detection, got %+v", detection)
				}
				return
			}

			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			if detection == nil {
				t.Fatal("Expected detection, got nil")
			}

			if detection.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", detection.Type, tt.wantType)
			}

			if detection.Confidence != tt.wantConf {
				t.Errorf("Confidence = %v, want %v", detection.Confidence, tt.wantConf)
			}
		})
	}
}

func TestGoDetector(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		wantType ProjectType
	}{
		{
			name: "go module project",
			files: map[string]string{
				"go.mod":  "module example.com/test\n\ngo 1.21",
				"main.go": "package main\n\nfunc main() {}",
			},
			wantType: TypeGo,
		},
		{
			name: "no go files",
			files: map[string]string{
				"package.json": "{}",
			},
			wantType: TypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			for filename, content := range tt.files {
				path := filepath.Join(tmpDir, filename)
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			detector := &GoDetector{}
			detection, err := detector.Detect(tmpDir)

			if tt.wantType == TypeUnknown {
				if detection != nil {
					t.Errorf("Expected nil detection, got %+v", detection)
				}
				return
			}

			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			if detection == nil || detection.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", detection.Type, tt.wantType)
			}
		})
	}
}

func TestManager(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		wantType ProjectType
	}{
		{
			name: "docker project wins over node",
			files: map[string]string{
				"docker-compose.yml": "version: '3'",
				"package.json":       `{"dependencies": {"express": "^4.0.0"}}`,
			},
			wantType: TypeDocker,
		},
		{
			name: "nuxt project",
			files: map[string]string{
				"package.json": `{"dependencies": {"nuxt": "^3.0.0"}}`,
			},
			wantType: TypeNuxt,
		},
		{
			name: "static site",
			files: map[string]string{
				"index.html": "<html><body>Hello</body></html>",
			},
			wantType: TypeStatic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			for filename, content := range tt.files {
				path := filepath.Join(tmpDir, filename)
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			manager := NewManager()
			detection, err := manager.DetectProject(tmpDir)

			if err != nil {
				t.Fatalf("DetectProject() error = %v", err)
			}

			if detection == nil {
				t.Fatal("Expected detection, got nil")
			}

			if detection.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", detection.Type, tt.wantType)
			}
		})
	}
}
