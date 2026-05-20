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
	"golang.org/x/crypto/bcrypt"
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

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type UpdateUserRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func main() {
	cfg := config.LoadConfig()

	_, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := database.DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	initDefaultData()

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	api := r.Group("/api/v1/users")
	{
		api.POST("", middleware.AuthMiddleware(), createUser)
		api.GET("/:id", middleware.AuthMiddleware(), getUser)
		api.PUT("/:id", middleware.AuthMiddleware(), updateUser)
		api.DELETE("/:id", middleware.AuthMiddleware(), deleteUser)
		api.GET("", middleware.AuthMiddleware(), listUsers)
	}

	port := os.Getenv("USER_SERVICE_PORT")
	port = extractPort(port)
	if port == "" {
		port = "8081"
	}

	log.Printf("User Service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func createUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.InternalServerError(c, "Failed to hash password")
		return
	}

	role := models.RoleUser
	if req.Role == "admin" {
		role = models.RoleAdmin
	}

	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Role:     role,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		response.BadRequest(c, "Username already exists")
		return
	}

	response.Success(c, user)
}

func getUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	if err := database.DB.First(&user, id).Error; err != nil {
		response.NotFound(c, "User not found")
		return
	}

	response.Success(c, user)
}

func updateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	if err := database.DB.First(&user, id).Error; err != nil {
		response.NotFound(c, "User not found")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Email != "" {
		user.Email = req.Email
	}

	if req.Role != "" {
		if req.Role == "admin" {
			user.Role = models.RoleAdmin
		} else {
			user.Role = models.RoleUser
		}
	}

	if err := database.DB.Save(&user).Error; err != nil {
		response.InternalServerError(c, "Failed to update user")
		return
	}

	response.Success(c, user)
}

func deleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := database.DB.Delete(&models.User{}, id).Error; err != nil {
		response.InternalServerError(c, "Failed to delete user")
		return
	}

	response.Success(c, nil)
}

func listUsers(c *gin.Context) {
	var users []models.User

	if err := database.DB.Find(&users).Error; err != nil {
		response.InternalServerError(c, "Failed to fetch users")
		return
	}

	response.Success(c, users)
}

func initDefaultData() {
	var userCount int64
	database.DB.Model(&models.User{}).Count(&userCount)

	if userCount == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Warning: Failed to hash default password: %v", err)
			return
		}

		admin := models.User{
			Username: "admin",
			Password: string(hashedPassword),
			Email:    "admin@generator.platform",
			Role:     models.RoleAdmin,
		}

		if err := database.DB.Create(&admin).Error; err != nil {
			log.Printf("Warning: Failed to create default admin user: %v", err)
		} else {
			log.Println("✓ Default admin account created: admin / admin123")
		}
	}
}
