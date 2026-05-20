package main

import (
	"fmt"
)

func generateGoMod(projectName string) string {
	return fmt.Sprintf(`module %s

go 1.21

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/lib/pq v1.10.9
	github.com/swaggo/files v1.0.1
	github.com/swaggo/gin-swagger v1.6.0
	github.com/swaggo/swag v1.16.3
	gorm.io/driver/postgres v1.5.7
	gorm.io/gen v0.3.22
	gorm.io/gorm v1.25.10
	gorm.io/datatypes v1.2.2
	github.com/go-playground/validator/v10 v10.20.0
)
`, projectName)
}

func generateMain(projectName string, dbConfig DBConfig, tables []TableConfig) string {
	modelList := ""
	tableNames := ""
	for i, table := range tables {
		if i > 0 {
			modelList += ",\n\t\t"
			tableNames += ","
		}
		modelList += "&models." + toCamelCase(table.Name) + "{}"
		tableNames += "\"" + table.Name + "\""
	}

	return fmt.Sprintf(`package main

import (
	"log"

	"%s/config"
	"%s/database"
	"%s/internal/models"
	"%s/internal/router"
	"%s/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           %s API
// @version         1.0
// @description     Auto-generated backend service with Gin + GORM
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

func main() {
	cfg := config.LoadConfig()

	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %%v", err)
	}

	tablesToMigrate := []string{%s}
	for _, tableName := range tablesToMigrate {
		var tableExists bool
		db.Raw("SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = ? AND table_schema = current_schema())", tableName).Scan(&tableExists)
		
		if tableExists {
			log.Printf("Dropping existing table: %%s (to recreate with correct structure)", tableName)
			if err := db.Exec("DROP TABLE IF EXISTS " + tableName + " CASCADE").Error; err != nil {
				log.Printf("Warning: Failed to drop table %%s: %%v", tableName, err)
			}
		}
	}

	if err := db.AutoMigrate(
		%s,
	); err != nil {
		log.Fatalf("Failed to migrate database: %%v", err)
	}

	gin.SetMode(cfg.ServerMode)

	r := gin.Default()

	router.SetupRoutes(r, db)

	docs.SwaggerInfo.Title = "%s API"
	docs.SwaggerInfo.Version = "1.0"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := cfg.ServerPort
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %%v", err)
	}
}
`, projectName, projectName, projectName, projectName, projectName,
		projectName, tableNames, modelList, projectName)
}

func generateConfig() string {
	return `package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	ServerPort  string
	ServerMode  string

	JWTSecret   string
	JWTExpire   int

	LogLevel    string
}

func LoadConfig() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "123456"),
		DBName:     getEnv("DB_NAME", "mydb"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		ServerPort: getEnv("SERVER_PORT", "8080"),
		ServerMode: getEnv("SERVER_MODE", "debug"),

		JWTSecret:  getEnv("JWT_SECRET", "your-secret-key"),
		JWTExpire:  getEnvAsInt("JWT_EXPIRE", 24),

		LogLevel:   getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
`
}

func generateDatabase(projectName string, dbConfig DBConfig) string {
	return fmt.Sprintf(`package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"%s/config"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	defaultDSN := fmt.Sprintf(
		"host=%%s port=%%s user=%%s password=%%s dbname=postgres sslmode=%%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBSSLMode,
	)

	db, err := sql.Open("postgres", defaultDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to default database: %%w", err)
	}
	defer db.Close()

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err = db.QueryRow(query, cfg.DBName).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check if database exists: %%w", err)
	}

	if !exists {
		log.Printf("Database %%s does not exist, creating...", cfg.DBName)
		_, err = db.Exec("CREATE DATABASE " + cfg.DBName)
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %%w", err)
		}
		log.Printf("Database %%s created successfully", cfg.DBName)
	} else {
		log.Printf("Database %%s already exists", cfg.DBName)
	}

	targetDSN := fmt.Sprintf(
		"host=%%s port=%%s user=%%s password=%%s dbname=%%s sslmode=%%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	var gormDB *gorm.DB
	gormDB, err = gorm.Open(postgres.Open(targetDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to target database: %%w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connected successfully")
	DB = gormDB
	return gormDB, nil
}
`, projectName)
}
