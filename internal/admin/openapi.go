package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// OpenAPISpec returns a minimal OpenAPI 3.0 spec
func (s *Server) OpenAPISpec(c *gin.Context) {
	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":   "dockrune API",
			"version": "1.0.0",
			"description": "Self-hosted deployment daemon API",
		},
		"servers": []map[string]interface{}{
			{
				"url": "http://localhost:8001",
				"description": "Admin API",
			},
			{
				"url": "http://localhost:8000", 
				"description": "Webhook Server",
			},
		},
		"paths": map[string]interface{}{
			"/webhook/github": map[string]interface{}{
				"post": map[string]interface{}{
					"summary": "GitHub webhook receiver",
					"tags": []string{"webhooks"},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Webhook processed",
						},
					},
				},
			},
			"/health": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Health check",
					"tags": []string{"monitoring"},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Service healthy",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"status": map[string]string{"type": "string"},
											"version": map[string]string{"type": "string"},
											"timestamp": map[string]string{"type": "string", "format": "date-time"},
										},
									},
								},
							},
						},
					},
				},
			},
			"/admin/login": map[string]interface{}{
				"post": map[string]interface{}{
					"summary": "Admin login",
					"tags": []string{"auth"},
					"requestBody": map[string]interface{}{
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"required": []string{"username", "password"},
									"properties": map[string]interface{}{
										"username": map[string]string{"type": "string"},
										"password": map[string]string{"type": "string"},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Login successful",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"token": map[string]string{"type": "string"},
										},
									},
								},
							},
						},
					},
				},
			},
			"/api/deployments": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "List deployments",
					"tags": []string{"deployments"},
					"security": []map[string][]string{
						{"bearerAuth": {}},
					},
					"parameters": []map[string]interface{}{
						{
							"name": "limit",
							"in": "query",
							"schema": map[string]interface{}{
								"type": "integer",
								"default": 50,
							},
						},
						{
							"name": "status",
							"in": "query",
							"schema": map[string]interface{}{
								"type": "string",
								"enum": []string{"queued", "in_progress", "success", "failed"},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "List of deployments",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "array",
										"items": map[string]interface{}{
											"$ref": "#/components/schemas/Deployment",
										},
									},
								},
							},
						},
					},
				},
			},
			"/api/deployments/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Get deployment",
					"tags": []string{"deployments"},
					"security": []map[string][]string{
						{"bearerAuth": {}},
					},
					"parameters": []map[string]interface{}{
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": map[string]string{
								"type": "string",
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Deployment details",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/Deployment",
									},
								},
							},
						},
					},
				},
			},
			"/api/deployments/{id}/logs": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Get deployment logs",
					"tags": []string{"deployments"},
					"security": []map[string][]string{
						{"bearerAuth": {}},
					},
					"parameters": []map[string]interface{}{
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": map[string]string{
								"type": "string",
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Log stream",
							"content": map[string]interface{}{
								"text/plain": map[string]interface{}{
									"schema": map[string]string{
										"type": "string",
									},
								},
							},
						},
					},
				},
			},
			"/api/deployments/{id}/redeploy": map[string]interface{}{
				"post": map[string]interface{}{
					"summary": "Trigger redeployment",
					"tags": []string{"deployments"},
					"security": []map[string][]string{
						{"bearerAuth": {}},
					},
					"parameters": []map[string]interface{}{
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": map[string]string{
								"type": "string",
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Redeployment queued",
						},
					},
				},
			},
		},
		"components": map[string]interface{}{
			"securitySchemes": map[string]interface{}{
				"bearerAuth": map[string]interface{}{
					"type": "http",
					"scheme": "bearer",
					"bearerFormat": "JWT",
				},
			},
			"schemas": map[string]interface{}{
				"Deployment": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":                    map[string]string{"type": "string"},
						"owner":                 map[string]string{"type": "string"},
						"repo":                  map[string]string{"type": "string"},
						"ref":                   map[string]string{"type": "string"},
						"sha":                   map[string]string{"type": "string"},
						"clone_url":             map[string]string{"type": "string"},
						"environment":           map[string]string{"type": "string"},
						"pr_number":             map[string]string{"type": "integer"},
						"github_deployment_id":  map[string]string{"type": "integer"},
						"status":                map[string]interface{}{
							"type": "string",
							"enum": []string{"queued", "in_progress", "success", "failed"},
						},
						"started_at":            map[string]string{"type": "string", "format": "date-time"},
						"completed_at":          map[string]string{"type": "string", "format": "date-time"},
						"log_path":              map[string]string{"type": "string"},
						"url":                   map[string]string{"type": "string"},
						"port":                  map[string]string{"type": "integer"},
						"project_type":          map[string]string{"type": "string"},
						"error":                 map[string]string{"type": "string"},
					},
				},
			},
		},
	}

	c.JSON(http.StatusOK, spec)
}