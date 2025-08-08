package detector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ProjectType string

const (
	TypeDocker  ProjectType = "docker"
	TypeNuxt    ProjectType = "nuxt"
	TypeNode    ProjectType = "node"
	TypeStatic  ProjectType = "static"
	TypeGo      ProjectType = "go"
	TypeRust    ProjectType = "rust"
	TypePython  ProjectType = "python"
	TypeUnknown ProjectType = "unknown"
)

type Detection struct {
	Type       ProjectType
	Confidence float32
	BuildCmd   string
	StartCmd   string
	Port       int
	Metadata   map[string]interface{}
}

type Detector interface {
	Detect(projectPath string) (*Detection, error)
	Priority() int
}

type Manager struct {
	detectors []Detector
}

func NewManager() *Manager {
	return &Manager{
		detectors: []Detector{
			&DockerDetector{},
			&NuxtDetector{},
			&GoDetector{},
			&RustDetector{},
			&NodeDetector{},
			&PythonDetector{},
			&StaticDetector{},
		},
	}
}

func (m *Manager) DetectProject(projectPath string) (*Detection, error) {
	var bestDetection *Detection
	var highestConfidence float32

	for _, detector := range m.detectors {
		detection, err := detector.Detect(projectPath)
		if err != nil {
			continue
		}

		if detection != nil && detection.Confidence > highestConfidence {
			highestConfidence = detection.Confidence
			bestDetection = detection
		}
	}

	if bestDetection == nil {
		return &Detection{
			Type:       TypeUnknown,
			Confidence: 0,
		}, nil
	}

	return bestDetection, nil
}

// DockerDetector detects Docker-based projects
type DockerDetector struct{}

func (d *DockerDetector) Priority() int { return 100 }

func (d *DockerDetector) Detect(projectPath string) (*Detection, error) {
	// Check for docker-compose.yml
	composePath := filepath.Join(projectPath, "docker-compose.yml")
	if _, err := os.Stat(composePath); err == nil {
		return &Detection{
			Type:       TypeDocker,
			Confidence: 1.0,
			BuildCmd:   "docker-compose build",
			StartCmd:   "docker-compose up -d",
			Port:       0, // Will be detected from compose file
			Metadata: map[string]interface{}{
				"compose_file": "docker-compose.yml",
			},
		}, nil
	}

	// Check for Dockerfile
	dockerfilePath := filepath.Join(projectPath, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); err == nil {
		return &Detection{
			Type:       TypeDocker,
			Confidence: 0.9,
			BuildCmd:   "docker build -t app .",
			StartCmd:   "docker run -d --name app -p ${PORT}:${PORT} app",
			Port:       3000,
			Metadata: map[string]interface{}{
				"dockerfile": "Dockerfile",
			},
		}, nil
	}

	return nil, nil
}

// NuxtDetector detects Nuxt.js projects
type NuxtDetector struct{}

func (n *NuxtDetector) Priority() int { return 90 }

func (n *NuxtDetector) Detect(projectPath string) (*Detection, error) {
	packagePath := filepath.Join(projectPath, "package.json")
	data, err := os.ReadFile(packagePath)
	if err != nil {
		return nil, nil
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, nil
	}

	// Check dependencies for Nuxt
	deps, _ := pkg["dependencies"].(map[string]interface{})
	devDeps, _ := pkg["devDependencies"].(map[string]interface{})

	isNuxt3 := false
	if deps != nil {
		if _, ok := deps["nuxt"]; ok {
			isNuxt3 = true
		}
	}
	if devDeps != nil {
		if _, ok := devDeps["nuxt"]; ok {
			isNuxt3 = true
		}
	}

	if isNuxt3 {
		// Check for Nitro output
		nitroPath := filepath.Join(projectPath, ".output", "server", "index.mjs")
		if _, err := os.Stat(nitroPath); err == nil {
			return &Detection{
				Type:       TypeNuxt,
				Confidence: 1.0,
				BuildCmd:   "npm run build",
				StartCmd:   "node .output/server/index.mjs",
				Port:       3000,
				Metadata: map[string]interface{}{
					"version": "3",
					"nitro":   true,
				},
			}, nil
		}

		return &Detection{
			Type:       TypeNuxt,
			Confidence: 0.95,
			BuildCmd:   "npm install && npm run build",
			StartCmd:   "npm run start",
			Port:       3000,
			Metadata: map[string]interface{}{
				"version": "3",
			},
		}, nil
	}

	return nil, nil
}

// GoDetector detects Go projects
type GoDetector struct{}

func (g *GoDetector) Priority() int { return 85 }

func (g *GoDetector) Detect(projectPath string) (*Detection, error) {
	goModPath := filepath.Join(projectPath, "go.mod")
	if _, err := os.Stat(goModPath); err != nil {
		return nil, nil
	}

	// Check for main.go
	mainPath := filepath.Join(projectPath, "main.go")
	cmdMainPath := filepath.Join(projectPath, "cmd")

	startCmd := "./app"
	if _, err := os.Stat(mainPath); err == nil {
		startCmd = "./app"
	} else if _, err := os.Stat(cmdMainPath); err == nil {
		// Find the main package in cmd/
		entries, _ := os.ReadDir(cmdMainPath)
		for _, entry := range entries {
			if entry.IsDir() {
				startCmd = fmt.Sprintf("./app")
				break
			}
		}
	}

	return &Detection{
		Type:       TypeGo,
		Confidence: 1.0,
		BuildCmd:   "go build -o app",
		StartCmd:   startCmd,
		Port:       8080,
		Metadata: map[string]interface{}{
			"has_go_mod": true,
		},
	}, nil
}

// RustDetector detects Rust projects
type RustDetector struct{}

func (r *RustDetector) Priority() int { return 85 }

func (r *RustDetector) Detect(projectPath string) (*Detection, error) {
	cargoPath := filepath.Join(projectPath, "Cargo.toml")
	if _, err := os.Stat(cargoPath); err != nil {
		return nil, nil
	}

	return &Detection{
		Type:       TypeRust,
		Confidence: 1.0,
		BuildCmd:   "cargo build --release",
		StartCmd:   "./target/release/app",
		Port:       8080,
		Metadata: map[string]interface{}{
			"has_cargo": true,
		},
	}, nil
}

// NodeDetector detects generic Node.js projects
type NodeDetector struct{}

func (n *NodeDetector) Priority() int { return 70 }

func (n *NodeDetector) Detect(projectPath string) (*Detection, error) {
	packagePath := filepath.Join(projectPath, "package.json")
	data, err := os.ReadFile(packagePath)
	if err != nil {
		return nil, nil
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, nil
	}

	// Check scripts
	scripts, _ := pkg["scripts"].(map[string]interface{})
	startCmd := "npm start"
	buildCmd := ""

	if scripts != nil {
		if _, ok := scripts["start"]; ok {
			startCmd = "npm start"
		}
		if _, ok := scripts["build"]; ok {
			buildCmd = "npm run build && "
		}
	}

	// Detect framework
	deps, _ := pkg["dependencies"].(map[string]interface{})
	framework := "generic"
	port := 3000

	if deps != nil {
		if _, ok := deps["express"]; ok {
			framework = "express"
			port = 3000
		} else if _, ok := deps["fastify"]; ok {
			framework = "fastify"
			port = 3000
		} else if _, ok := deps["@nestjs/core"]; ok {
			framework = "nestjs"
			port = 3000
		} else if _, ok := deps["next"]; ok {
			framework = "nextjs"
			port = 3000
			startCmd = "npm run start"
			buildCmd = "npm run build && "
		}
	}

	return &Detection{
		Type:       TypeNode,
		Confidence: 0.8,
		BuildCmd:   buildCmd + "npm install",
		StartCmd:   startCmd,
		Port:       port,
		Metadata: map[string]interface{}{
			"framework": framework,
		},
	}, nil
}

// PythonDetector detects Python projects
type PythonDetector struct{}

func (p *PythonDetector) Priority() int { return 70 }

func (p *PythonDetector) Detect(projectPath string) (*Detection, error) {
	// Check for requirements.txt
	reqPath := filepath.Join(projectPath, "requirements.txt")
	if _, err := os.Stat(reqPath); err != nil {
		// Check for pyproject.toml
		pyprojectPath := filepath.Join(projectPath, "pyproject.toml")
		if _, err := os.Stat(pyprojectPath); err != nil {
			return nil, nil
		}
	}

	// Detect framework
	framework := "generic"
	startCmd := "python app.py"
	port := 8000

	// Check for common Python web frameworks
	if _, err := os.Stat(filepath.Join(projectPath, "manage.py")); err == nil {
		framework = "django"
		startCmd = "python manage.py runserver 0.0.0.0:${PORT}"
		port = 8000
	} else if _, err := os.Stat(filepath.Join(projectPath, "app.py")); err == nil {
		// Could be Flask or FastAPI
		data, _ := os.ReadFile(filepath.Join(projectPath, "app.py"))
		content := string(data)
		if len(content) > 0 {
			if contains(content, "from flask") || contains(content, "import flask") {
				framework = "flask"
				startCmd = "python app.py"
				port = 5000
			} else if contains(content, "from fastapi") || contains(content, "import fastapi") {
				framework = "fastapi"
				startCmd = "uvicorn app:app --host 0.0.0.0 --port 8000"
				port = 8000
			}
		}
	}

	return &Detection{
		Type:       TypePython,
		Confidence: 0.8,
		BuildCmd:   "pip install -r requirements.txt",
		StartCmd:   startCmd,
		Port:       port,
		Metadata: map[string]interface{}{
			"framework": framework,
		},
	}, nil
}

// StaticDetector detects static sites
type StaticDetector struct{}

func (s *StaticDetector) Priority() int { return 50 }

func (s *StaticDetector) Detect(projectPath string) (*Detection, error) {
	// Check for index.html
	indexPath := filepath.Join(projectPath, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		return &Detection{
			Type:       TypeStatic,
			Confidence: 0.7,
			BuildCmd:   "",
			StartCmd:   "python -m http.server 8080",
			Port:       8080,
			Metadata: map[string]interface{}{
				"type": "html",
			},
		}, nil
	}

	// Check for common static site generators
	if _, err := os.Stat(filepath.Join(projectPath, "_config.yml")); err == nil {
		return &Detection{
			Type:       TypeStatic,
			Confidence: 0.8,
			BuildCmd:   "jekyll build",
			StartCmd:   "jekyll serve --host 0.0.0.0",
			Port:       4000,
			Metadata: map[string]interface{}{
				"generator": "jekyll",
			},
		}, nil
	}

	if _, err := os.Stat(filepath.Join(projectPath, "config.toml")); err == nil {
		return &Detection{
			Type:       TypeStatic,
			Confidence: 0.8,
			BuildCmd:   "hugo",
			StartCmd:   "hugo server --bind 0.0.0.0",
			Port:       1313,
			Metadata: map[string]interface{}{
				"generator": "hugo",
			},
		}, nil
	}

	return nil, nil
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
