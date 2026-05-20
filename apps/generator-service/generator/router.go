package main

import (
	"fmt"
	"strings"
)

func generateRouter(projectName string, tables []TableConfig) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`package router

import (
	"%s/internal/controller"
	"%s/internal/dao"
	"%s/internal/middleware"
	"%s/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	d := dao.NewDAO(db)
	s := service.NewService(d)
	ctrl := controller.NewController(s)

	api := r.Group("/api/v1")
	{
`, projectName, projectName, projectName, projectName))

	for _, table := range tables {
		modelName := toCamelCase(table.Name)
		sb.WriteString(fmt.Sprintf(`
		%s := api.Group("/%s")
		{
			%s.POST("", ctrl.Create%s)
			%s.GET("", ctrl.List%ss)
			%s.GET("/:id", ctrl.Get%s)
			%s.PUT("/:id", ctrl.Update%s)
			%s.DELETE("/:id", ctrl.Delete%s)
		}
`,
			table.Name, table.Name,
			table.Name, modelName,
			table.Name, modelName,
			table.Name, modelName,
			table.Name, modelName,
			table.Name, modelName,
		))
	}

	sb.WriteString(`	}
}
`)

	return sb.String()
}
