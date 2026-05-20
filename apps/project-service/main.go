package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/generator-platform/go-common/config"
	"github.com/generator-platform/go-common/database"
	"github.com/generator-platform/go-common/middleware"
	"github.com/generator-platform/go-common/models"
	"github.com/generator-platform/go-common/response"
	"github.com/gin-gonic/gin"
)

func extractPort(portStr string) string {
	if portStr == "" {
		return ""
	}
	if strings.Contains(portStr, ":") {
		parts := strings.Split(portStr, ":")
		lastPart := parts[len(parts)-1]
		if _, err := strconv.Atoi(lastPart); err == nil {
			return lastPart
		}
	}
	if _, err := strconv.Atoi(portStr); err == nil {
		return portStr
	}
	return ""
}

type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	DBConfig    string `json:"db_config"`
	TableConfig string `json:"table_config"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DBConfig    string `json:"db_config"`
	TableConfig string `json:"table_config"`
}

func main() {
	cfg := config.LoadConfig()

	_, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := database.DB.AutoMigrate(&models.Project{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	api := r.Group("/api/v1/projects")
	api.Use(middleware.AuthMiddleware())
	{
		api.POST("", createProject)
		api.GET("/:id", getProject)
		api.PUT("/:id", updateProject)
		api.DELETE("/:id", deleteProject)
		api.GET("", listProjects)
	}

	port := os.Getenv("PROJECT_SERVICE_PORT")
	port = extractPort(port)
	if port == "" {
		port = "8083"
	}

	log.Printf("Project Service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func createProject(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	project := models.Project{
		UserID:      userID.(uint),
		Name:        req.Name,
		Description: req.Description,
		DBConfig:    req.DBConfig,
		TableConfig: req.TableConfig,
	}

	if err := database.DB.Create(&project).Error; err != nil {
		response.InternalServerError(c, "Failed to create project")
		return
	}

	response.Success(c, project)
}

func getProject(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	id := c.Param("id")

	var project models.Project
	query := database.DB.Where("id = ?", id)

	// 管理员可以看到所有项目
	if role != models.RoleAdmin {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&project).Error; err != nil {
		response.NotFound(c, "Project not found")
		return
	}

	response.Success(c, project)
}

func updateProject(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	id := c.Param("id")

	var project models.Project
	query := database.DB.Where("id = ?", id)

	// 管理员可以更新所有项目
	if role != models.RoleAdmin {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&project).Error; err != nil {
		response.NotFound(c, "Project not found")
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.DBConfig != "" {
		project.DBConfig = req.DBConfig
	}
	if req.TableConfig != "" {
		project.TableConfig = req.TableConfig
	}

	if err := database.DB.Save(&project).Error; err != nil {
		response.InternalServerError(c, "Failed to update project")
		return
	}

	response.Success(c, project)
}

func deleteProject(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	id := c.Param("id")

	query := database.DB.Where("id = ?", id)

	// 管理员可以删除所有项目
	if role != models.RoleAdmin {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Delete(&models.Project{}).Error; err != nil {
		response.InternalServerError(c, "Failed to delete project")
		return
	}

	response.Success(c, nil)
}

func listProjects(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var projects []models.Project
	query := database.DB.Preload("User")

	// 管理员可以看到所有项目
	// 普通用户只能看到自己的项目
	if role != models.RoleAdmin {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&projects).Error; err != nil {
		response.InternalServerError(c, "Failed to fetch projects")
		return
	}

	response.Success(c, projects)
}
