package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/generator-platform/go-common/config"
	"github.com/generator-platform/go-common/database"
	"github.com/generator-platform/go-common/jwt"
	"github.com/generator-platform/go-common/middleware"
	"github.com/generator-platform/go-common/models"
	"github.com/generator-platform/go-common/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func recordAuthLog(c *gin.Context, action, username string, status, errorMsg string, duration int64) {
	if database.DB == nil {
		return
	}

	logEntry := models.OperationLog{
		UserID:     0,
		Username:   username,
		Action:     action,
		Resource:   "user",
		ResourceID: 0,
		Details:    fmt.Sprintf(`{"username": "%s", "action": "%s"}`, username, action),
		Status:     status,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		Duration:   duration,
		Error:      errorMsg,
	}

	if err := database.DB.Create(&logEntry).Error; err != nil {
		log.Printf("[WARNING] Failed to record auth log: %v", err)
	} else {
		log.Printf("[AUTH] User=%s Action=%s Status=%s Duration=%dms", username, action, status, duration)
	}
}

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

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func main() {
	cfg := config.LoadConfig()

	_, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := database.DB.AutoMigrate(&models.User{}, &models.OperationLog{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	initDefaultUser()

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	api := r.Group("/api/v1/auth")
	{
		api.POST("/login", login)
		api.POST("/register", register)
		api.GET("/me", middleware.AuthMiddleware(), getCurrentUser)
	}

	port := os.Getenv("AUTH_SERVICE_PORT")
	port = extractPort(port)
	if port == "" {
		port = "8082"
	}

	log.Printf("Auth Service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initDefaultUser() {
	var count int64
	database.DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		log.Println("Users already exist, skipping default user creation")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash default password: %v", err)
		return
	}

	defaultUser := models.User{
		Username: "admin",
		Password: string(hashedPassword),
		Email:    "admin@example.com",
		Role:     models.RoleAdmin,
	}

	if err := database.DB.Create(&defaultUser).Error; err != nil {
		log.Printf("Failed to create default user: %v", err)
		return
	}

	log.Println("Default admin user created successfully")
	log.Println("Username: admin")
	log.Println("Password: admin123")
}

func login(c *gin.Context) {
	startTime := time.Now()
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		recordAuthLog(c, "login", req.Username, "failed", err.Error(), time.Since(startTime).Milliseconds())
		response.BadRequest(c, err.Error())
		return
	}

	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		recordAuthLog(c, "login", req.Username, "failed", "Invalid username or password", time.Since(startTime).Milliseconds())
		response.Unauthorized(c, "Invalid username or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		recordAuthLog(c, "login", req.Username, "failed", "Invalid password", time.Since(startTime).Milliseconds())
		response.Unauthorized(c, "Invalid username or password")
		return
	}

	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		recordAuthLog(c, "login", user.Username, "failed", err.Error(), time.Since(startTime).Milliseconds())
		response.InternalServerError(c, "Failed to generate token")
		return
	}

	recordAuthLog(c, "login", user.Username, "success", "", time.Since(startTime).Milliseconds())

	response.Success(c, LoginResponse{
		Token: token,
		User:  user,
	})
}

func register(c *gin.Context) {
	startTime := time.Now()
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		recordAuthLog(c, "register", req.Username, "failed", err.Error(), time.Since(startTime).Milliseconds())
		response.BadRequest(c, err.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		recordAuthLog(c, "register", req.Username, "failed", err.Error(), time.Since(startTime).Milliseconds())
		response.InternalServerError(c, "Failed to hash password")
		return
	}

	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Role:     models.RoleUser, // 默认注册为普通用户
	}

	if err := database.DB.Create(&user).Error; err != nil {
		recordAuthLog(c, "register", req.Username, "failed", "Username already exists", time.Since(startTime).Milliseconds())
		response.BadRequest(c, "Username already exists")
		return
	}

	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		recordAuthLog(c, "register", user.Username, "failed", err.Error(), time.Since(startTime).Milliseconds())
		response.InternalServerError(c, "Failed to generate token")
		return
	}

	recordAuthLog(c, "register", user.Username, "success", fmt.Sprintf("User ID: %d, Email: %s", user.ID, user.Email), time.Since(startTime).Milliseconds())

	response.Success(c, LoginResponse{
		Token: token,
		User:  user,
	})
}

func getCurrentUser(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		response.NotFound(c, "User not found")
		return
	}

	response.Success(c, user)
}
