package main

import (
	"github.com/generator-platform/go-common/config"
	"github.com/generator-platform/go-common/database"
	"github.com/generator-platform/go-common/middleware"
	"github.com/generator-platform/go-common/models"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

func main() {
	cfg := config.LoadConfig()

	_, err := database.InitDB(cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize database: %v", err)
		log.Printf("Continuing without database connection...")
	} else {
		if err := database.DB.AutoMigrate(&models.Project{}, &models.OperationLog{}); err != nil {
			log.Printf("Warning: Failed to migrate database (table may already exist): %v", err)
		}
	}

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	api := r.Group("/api/v1/generator")
	api.Use(middleware.AuthMiddleware())
	{
		api.POST("/generate", generateCode)
		api.POST("/generate/:project_id", generateFromProject)
		api.GET("/download/:project_id", downloadZip)
		api.GET("/preview/:project_id", previewCode)
		api.POST("/docs/generate", generateDocumentation)
	}

	port := os.Getenv("GENERATOR_SERVICE_PORT")
	port = extractPort(port)
	if port == "" {
		port = "8084"
	}

	log.Printf("Generator Service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
