package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/generator-platform/go-common/config"
	"github.com/generator-platform/go-common/middleware"
	"github.com/gin-gonic/gin"
)

var serviceUrls map[string]string

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func initServiceUrls() {
	serviceUrls = map[string]string{
		"auth":       getEnv("AUTH_SERVICE_URL", "http://localhost:8082"),
		"users":      getEnv("USER_SERVICE_URL", "http://localhost:8081"),
		"projects":   getEnv("PROJECT_SERVICE_URL", "http://localhost:8083"),
		"generator":  getEnv("GENERATOR_SERVICE_URL", "http://localhost:8084"),
		"operations": getEnv("OPERATIONS_SERVICE_URL", "http://localhost:8085"),
		"clusters":   getEnv("CLUSTER_SERVICE_URL", "http://localhost:8086"),
	}

	// 打印服务URL以便调试
	log.Printf("Service URLs initialized:")
	for name, url := range serviceUrls {
		log.Printf("  %s: %s", name, url)
	}
}

func proxy(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 移除/api/v1前缀，因为后端服务已经有这个前缀
		path := c.Request.URL.RequestURI()
		targetUrl := serviceUrls[serviceName] + path

		// 调试日志
		log.Printf("Proxying request to %s: %s %s", serviceName, c.Request.Method, targetUrl)

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request"})
			return
		}
		c.Request.Body.Close()

		req, err := http.NewRequest(c.Request.Method, targetUrl, bytes.NewBuffer(body))
		if err != nil {
			log.Printf("Error creating request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		req.Header = c.Request.Header

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error forwarding request to %s: %v", serviceName, err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
			return
		}
		defer resp.Body.Close()

		log.Printf("Received response from %s: status=%d", serviceName, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
			return
		}

		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
	}
}

func main() {
	cfg := config.LoadConfig()

	// 调试日志：打印JWT配置
	if cfg.JWTSecret != "" {
		log.Printf("JWT Config - Secret: %s..., Expire: %s", cfg.JWTSecret[:10], cfg.JWTExpire)
	} else {
		log.Printf("WARNING: JWT_SECRET is empty!")
	}

	// 初始化服务URL（在读取环境变量之后）
	initServiceUrls()

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	api := r.Group("/api/v1")
	{
		// Auth routes - no auth required for login/register
		api.POST("/auth/login", proxy("auth"))
		api.POST("/auth/register", proxy("auth"))
		api.GET("/auth/me", middleware.AuthMiddleware(), proxy("auth"))

		// User routes - all require auth
		api.GET("/users", middleware.AuthMiddleware(), proxy("users"))
		api.POST("/users", middleware.AuthMiddleware(), proxy("users"))
		api.GET("/users/:id", middleware.AuthMiddleware(), proxy("users"))
		api.PUT("/users/:id", middleware.AuthMiddleware(), proxy("users"))
		api.DELETE("/users/:id", middleware.AuthMiddleware(), proxy("users"))

		// Project routes
		api.GET("/projects", middleware.AuthMiddleware(), proxy("projects"))
		api.POST("/projects", middleware.AuthMiddleware(), proxy("projects"))
		api.GET("/projects/:id", middleware.AuthMiddleware(), proxy("projects"))
		api.PUT("/projects/:id", middleware.AuthMiddleware(), proxy("projects"))
		api.DELETE("/projects/:id", middleware.AuthMiddleware(), proxy("projects"))

		// Generator routes
		api.POST("/generator/generate", middleware.AuthMiddleware(), proxy("generator"))
		api.POST("/generator/generate/:id", middleware.AuthMiddleware(), proxy("generator"))
		api.GET("/generator/download/:project_id", middleware.AuthMiddleware(), proxy("generator"))
		api.GET("/generator/preview/:project_id", middleware.AuthMiddleware(), proxy("generator"))
		// 文档生成路由
		api.POST("/generator/docs/generate", middleware.AuthMiddleware(), proxy("generator"))

		// Operations routes
		api.GET("/operations/health", middleware.AuthMiddleware(), proxy("operations"))
		api.GET("/operations/stats", middleware.AuthMiddleware(), proxy("operations"))
		api.GET("/operations/metrics", middleware.AuthMiddleware(), proxy("operations"))
		api.GET("/operations/services", middleware.AuthMiddleware(), proxy("operations"))
		api.GET("/operations/events", middleware.AuthMiddleware(), proxy("operations"))
		api.GET("/operations/overview", middleware.AuthMiddleware(), proxy("operations"))
		api.GET("/operations/operation-logs", middleware.AuthMiddleware(), proxy("operations"))
		api.POST("/operations/operation-logs/record", middleware.AuthMiddleware(), proxy("operations"))

		// Cluster routes - Self-managed cluster (no :id parameter needed)
		// These routes must be defined BEFORE /clusters/:id to avoid conflicts
		api.GET("/clusters/status", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/metrics", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/health", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/k8s/status", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/k8s/info", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/k8s/namespaces", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/k8s/nodes", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/k8s/nodes/join-command", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/k8s/pods", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/k8s/services", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/k8s/deployments", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/k8s/events", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/k8s/pods/:namespace/:name/logs", middleware.AuthMiddleware(), proxy("clusters"))
		api.DELETE("/clusters/k8s/pods/:namespace/:name", middleware.AuthMiddleware(), proxy("clusters"))
		api.POST("/clusters/k8s/deployments/:namespace/:name/scale", middleware.AuthMiddleware(), proxy("clusters"))
		api.POST("/clusters/k8s/deployments/:namespace/:name/restart", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/docker/services", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/docker/services/:service_name/logs", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/docker/services/:service_name/stats", middleware.AuthMiddleware(), proxy("clusters"))
		api.POST("/clusters/docker/services/:service_name/restart", middleware.AuthMiddleware(), proxy("clusters"))
		// Legacy cluster management routes (kept for backward compatibility)
		api.GET("/clusters", middleware.AuthMiddleware(), proxy("clusters"))
		api.POST("/clusters", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/:id", middleware.AuthMiddleware(), proxy("clusters"))
		api.PUT("/clusters/:id", middleware.AuthMiddleware(), proxy("clusters"))
		api.DELETE("/clusters/:id", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/:id/metrics", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/:id/services", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/:id/services/:service_name/logs", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/:id/services/:service_name/stats", middleware.AuthMiddleware(), proxy("clusters"))
		api.POST("/clusters/:id/services/:service_name/restart", middleware.AuthMiddleware(), proxy("clusters"))
		api.GET("/clusters/:id/health", middleware.AuthMiddleware(), proxy("clusters"))
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	port := os.Getenv("API_GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API Gateway starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
