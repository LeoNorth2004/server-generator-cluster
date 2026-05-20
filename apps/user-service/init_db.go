package main

import (
	"log"

	"github.com/generator-platform/go-common/config"
	"github.com/generator-platform/go-common/database"
	"github.com/generator-platform/go-common/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func initDatabase() {
	cfg := config.LoadConfig()

	db, err := database.InitDB(cfg)
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
		return
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Cluster{},
	); err != nil {
		log.Printf("Failed to migrate database: %v", err)
		return
	}

	log.Println("Database migration completed successfully")

	createDefaultAdmin(db)
	createSampleProjects(db)

	log.Println("Database initialization completed!")
}

func createDefaultAdmin(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Count(&count)

	if count > 0 {
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return
	}

	admin := models.User{
		Username: "admin",
		Password: string(hashedPassword),
		Email:    "admin@generator.platform",
		Role:     models.RoleAdmin,
	}

	db.Create(&admin)
	log.Println("Default admin account created: admin / admin123")
}

func createSampleProjects(db *gorm.DB) {
	var count int64
	db.Model(&models.Project{}).Count(&count)

	if count > 0 {
		return
	}

	sampleProjects := []models.Project{
		{
			UserID:   1,
			Name:     "示例电商系统",
			Description: "一个完整的电商平台，包含商品管理、订单处理、用户系统等模块",
			GeneratedCode: `{"files": {}}`,
		},
		{
			UserID:   1,
			Name:     "博客管理系统",
			Description: "支持多用户博客发布、评论、标签管理等功能的CMS系统",
			GeneratedCode: `{"files": {}}`,
		},
		{
			UserID:   1,
			Name:     "任务管理系统",
			Description: "企业级任务跟踪和项目管理工具",
			GeneratedCode: "",
		},
	}

	for _, project := range sampleProjects {
		db.Create(&project)
	}
}