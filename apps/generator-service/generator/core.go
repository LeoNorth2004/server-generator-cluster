package main

import (
	"log"
	"strings"
)

func doGenerate(req GenerateRequest) (*GeneratedCode, error) {
	files := make(map[string]string)

	files["go.mod"] = generateGoMod(req.ProjectName)
	files["main.go"] = generateMain(req.ProjectName, req.DBConfig, req.Tables)
	files["README.md"] = generateReadme(req.ProjectName, req.Tables)
	files[".env.example"] = generateEnvExample()
	files[".gitignore"] = generateGitignore()

	files["config/config.go"] = generateConfig()

	files["database/database.go"] = generateDatabase(req.ProjectName, req.DBConfig)

	needDatatypes := false
	for _, table := range req.Tables {
		for _, field := range table.Fields {
			goType := goTypeFromSQL(field.Type)
			if goType == "datatypes.JSON" {
				needDatatypes = true
				break
			}
		}
	}
	files["internal/models/base.go"] = generateBaseModel(needDatatypes)
	for _, table := range req.Tables {
		fileName := toLowerCamelCase(table.Name) + ".go"
		modelCode := generateSingleModel(table, needDatatypes)
		files["internal/models/"+fileName] = modelCode
		log.Printf("[DEBUG] Generated model for %s:\n%s", table.Name, modelCode)
	}

	files["internal/dao/dao.go"] = generateDAO(req.ProjectName, req.Tables)
	files["internal/dao/gen.go"] = generateDAOGen(req.Tables)

	files["internal/service/service.go"] = generateService(req.ProjectName, req.Tables)

	files["internal/controller/controller.go"] = generateController(req.ProjectName, req.Tables)

	files["internal/router/router.go"] = generateRouter(req.ProjectName, req.Tables)

	files["internal/middleware/cors.go"] = generateCorsMiddleware()
	files["internal/middleware/logger.go"] = generateLoggerMiddleware()

	files["pkg/utils/response.go"] = generateResponse()
	files["pkg/utils/validator.go"] = generateValidator()

	files["docs/swagger.go"] = generateSwagger(req.ProjectName, req.Tables)
	files["docs/config_guide.md"] = generateConfigGuide()
	files["docs/development_guide.md"] = generateDevelopmentGuide(req.ProjectName, req.Tables)

	files["migrations/schema.sql"] = generateSQL(req.Tables, req.DBConfig)
	files["migrations/seeds.sql"] = generateSeeds(req.Tables)

	log.Println("[INFO] 开始检查未使用的导入...")
	for fileName, content := range files {
		if strings.HasSuffix(fileName, ".go") {
			checkUnusedImports(fileName, content)
		}
	}
	log.Println("[INFO] 未使用导入检查完成")

	return &GeneratedCode{Files: files}, nil
}
