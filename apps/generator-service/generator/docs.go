package main

import (
	"fmt"
	"strings"
)

func generateSwagger(projectName string, tables []TableConfig) string {
	return fmt.Sprintf(`package docs

import "github.com/swaggo/swag"

var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/api/v1",
	Schemes:          []string{"http", "https"},
	Title:            "%s API",
	Description:      "Auto-generated backend service API documentation",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

const docTemplate = `+"`"+`
{
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {}
}
`+"`"+`

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
`, projectName)
}

func generateConfigGuide() string {
	return "# 配置指南\n\n" +
		"## 环境变量\n\n" +
		"### 数据库配置\n\n" +
		"| 变量名 | 说明 | 默认值 |\n" +
		"|--------|------|--------|\n" +
		"| DB_HOST | 数据库主机地址 | localhost |\n" +
		"| DB_PORT | 数据库端口 | 5432 |\n" +
		"| DB_USER | 数据库用户名 | postgres |\n" +
		"| DB_PASSWORD | 数据库密码 | 123456 |\n" +
		"| DB_NAME | 数据库名称 | mydb |\n" +
		"| DB_SSLMODE | SSL模式 | disable |\n\n" +
		"### 服务器配置\n\n" +
		"| 变量名 | 说明 | 默认值 |\n" +
		"|--------|------|--------|\n" +
		"| SERVER_PORT | 服务器端口 | 8080 |\n" +
		"| SERVER_MODE | Gin运行模式 (debug/release) | debug |\n\n" +
		"### JWT配置\n\n" +
		"| 变量名 | 说明 | 默认值 |\n" +
		"|--------|------|--------|\n" +
		"| JWT_SECRET | JWT密钥 | your-secret-key |\n" +
		"| JWT_EXPIRE | JWT过期时间(小时) | 24 |\n\n" +
		"### 日志配置\n\n" +
		"| 变量名 | 说明 | 默认值 |\n" +
		"|--------|------|--------|\n" +
		"| LOG_LEVEL | 日志级别 (debug/info/warn/error) | info |\n\n" +
		"## 配置文件示例\n\n" +
		"创建 .env 文件：\n\n" +
		"```bash\n" +
		"# Database\n" +
		"DB_HOST=localhost\n" +
		"DB_PORT=5432\n" +
		"DB_USER=postgres\n" +
		"DB_PASSWORD=your_password\n" +
		"DB_NAME=mydb\n" +
		"DB_SSLMODE=disable\n\n" +
		"# Server\n" +
		"SERVER_PORT=8080\n" +
		"SERVER_MODE=debug\n\n" +
		"# JWT\n" +
		"JWT_SECRET=your-secret-key-here\n" +
		"JWT_EXPIRE=24\n\n" +
		"# Log\n" +
		"LOG_LEVEL=info\n" +
		"```\n\n" +
		"## 不同环境配置\n\n" +
		"### 开发环境\n\n" +
		"```bash\n" +
		"SERVER_MODE=debug\n" +
		"LOG_LEVEL=debug\n" +
		"```\n\n" +
		"### 生产环境\n\n" +
		"```bash\n" +
		"SERVER_MODE=release\n" +
		"LOG_LEVEL=error\n" +
		"DB_SSLMODE=require\n" +
		"```\n"
}

func generateDevelopmentGuide(projectName string, tables []TableConfig) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("# %s - 二次开发指南\n\n", projectName))
	sb.WriteString("本文档基于项目实际生成的代码结构编写，包含针对当前数据模型的开发指导。\n\n")
	
	sb.WriteString("## 项目概览\n\n")
	sb.WriteString(fmt.Sprintf("- **项目名称**: %s\n", projectName))
	sb.WriteString(fmt.Sprintf("- **数据表数量**: %d\n", len(tables)))
	sb.WriteString("- **技术栈**: Go + Gin + GORM + PostgreSQL\n\n")

	sb.WriteString("## 数据模型\n\n")
	sb.WriteString("本项目包含以下数据表：\n\n")
	for _, table := range tables {
		sb.WriteString(fmt.Sprintf("### %s (%s)\n\n", table.Name, table.Comment))
		if len(table.Fields) > 0 {
			sb.WriteString("| 字段名 | 类型 | 可空 | 主键 | 说明 |\n")
			sb.WriteString("|--------|------|------|------|------|\n")
			for _, field := range table.Fields {
				primary := "✓" 
				if !field.Primary { primary = "" }
				nullable := "是"
				if !field.Nullable { nullable = "否" }
				comment := field.Comment
				if comment == "" { comment = "-" }
				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n", field.Name, field.Type, nullable, primary, comment))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("## 项目结构\n\n")
	sb.WriteString("```\n")
	sb.WriteString(fmt.Sprintf("%s/\n", projectName))
	sb.WriteString("├── config/                 # 配置文件\n")
	sb.WriteString("├── database/               # 数据库连接\n")
	sb.WriteString("├── internal/               # 内部代码\n")
	for i, table := range tables {
		modelName := toLowerCamelCase(table.Name)
		if i == 0 {
			sb.WriteString("│  ├── models/             # 数据模型\n")
		}
		sb.WriteString(fmt.Sprintf("│  │  └── %s.go           # %s 模型\n", modelName, toCamelCase(table.Name)))
	}
	sb.WriteString("│  ├── controller/         # 控制器层 (HTTP Handler)\n")
	sb.WriteString("│  ├── dao/                # 数据访问层(DAO)\n")
	sb.WriteString("│  ├── middleware/         # 中间件\n")
	sb.WriteString("│  ├── router/             # 路由配置\n")
	sb.WriteString("│  └── service/            # 业务逻辑层\n")
	sb.WriteString("├── pkg/                    # 公共包\n")
	sb.WriteString("│  └── utils/              # 工具函数\n")
	sb.WriteString("├── docs/                   # 文档\n")
	sb.WriteString("├── migrations/             # 数据库迁移脚本\n")
	sb.WriteString("├── go.mod                  # Go模块定义\n")
	sb.WriteString("├── .env.example            # 环境变量示例\n")
	sb.WriteString("└── README.md               # 项目说明\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## 各层职责\n\n")
	sb.WriteString("### 1. Model 层(internal/models)\n\n")
	sb.WriteString("定义数据结构和数据库表映射关系。每个实体一个文件。\n\n")
	sb.WriteString("**已生成的模型**:\n\n")
	for _, table := range tables {
		modelName := toCamelCase(table.Name)
		fileName := toLowerCamelCase(table.Name) + ".go"
		sb.WriteString(fmt.Sprintf("- `%s`: 定义 **%s** 结构体 (`internal/models/%s`)\n", fileName, modelName, fileName))
	}
	sb.WriteString("\n- 使用 GORM Tag 定义字段属性\n")
	sb.WriteString("- 实现 TableName 方法指定表名\n\n")

	sb.WriteString("### 2. DAO 层(internal/dao)\n\n")
	sb.WriteString("数据访问对象，负责数据库操作。\n\n")
	sb.WriteString("- 封装 CRUD 操作\n")
	sb.WriteString("- 支持事务处理\n")
	sb.WriteString("- 使用 GORM Gen 生成类型安全的查询代码\n\n")

	sb.WriteString("### 3. Service 层(internal/service)\n\n")
	sb.WriteString("业务逻辑层，处理业务规则。\n\n")
	sb.WriteString("**已生成的 Service 方法**:\n\n")
	for _, table := range tables {
		entityName := toCamelCase(table.Name)
		entityPlural := entityName + "s"
		sb.WriteString(fmt.Sprintf("- `New%sService()`: 创建 %s 服务实例\n", entityName, entityPlural))
		sb.WriteString(fmt.Sprintf("- `Create()`: 创建新 %s\n", entityName))
		sb.WriteString(fmt.Sprintf("- `GetByID()`: 根据 ID 获取 %s\n", entityName))
		sb.WriteString(fmt.Sprintf("- `List()`: 获取 %s 列表（支持分页）\n", entityPlural))
		sb.WriteString(fmt.Sprintf("- `Update()`: 更新 %s\n", entityName))
		sb.WriteString(fmt.Sprintf("- `Delete()`: 删除 %s\n", entityName))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	sb.WriteString("### 4. Controller 层(internal/controller)\n\n")
	sb.WriteString("控制器层，处理HTTP 请求。\n\n")
	sb.WriteString("**已生成的 API 端点**:\n\n")
	for _, table := range tables {
		entityName := toCamelCase(table.Name)
		entityLower := toLowerCamelCase(table.Name)
		sb.WriteString(fmt.Sprintf("- `POST   /api/v1/%s`      → Create%s()\n", entityLower, entityName))
		sb.WriteString(fmt.Sprintf("- `GET    /api/v1/%s/:id`   → Get%s()\n", entityLower, entityName))
		sb.WriteString(fmt.Sprintf("- `GET    /api/v1/%s`       → List%s()\n", entityLower, entityName+"s"))
		sb.WriteString(fmt.Sprintf("- `PUT    /api/v1/%s/:id`   → Update%s()\n", entityLower, entityName))
		sb.WriteString(fmt.Sprintf("- `DELETE /api/v1/%s/:id`   → Delete%s()\n\n", entityLower, entityName))
	}

	sb.WriteString("### 5. Router 层(internal/router)\n\n")
	sb.WriteString("路由配置。\n\n")
	sb.WriteString("- 定义 API 路由\n")
	sb.WriteString("- 应用中间件\n")
	sb.WriteString("- 绑定控制器方法\n\n")

	sb.WriteString("## 如何添加新功能\n\n")
	sb.WriteString("### 示例：为现有表添加新的查询接口\n\n")
	if len(tables) > 0 {
		table := tables[0]
		entityName := toCamelCase(table.Name)
		entityLower := toLowerCamelCase(table.Name)
		sb.WriteString(fmt.Sprintf("以 **%s** 表为例：\n\n", table.Name))
		sb.WriteString("```go\n")
		sb.WriteString(fmt.Sprintf("// 1. 在 internal/dao/%s_dao.go 中添加查询方法\n", entityLower))
		sb.WriteString(fmt.Sprintf("func (d *%sDAO) FindByName(ctx context.Context, name string) (*models.%s, error) {\n", entityName, entityName))
		sb.WriteString(fmt.Sprintf("    var %s models.%s\n", entityLower, entityName))
		sb.WriteString(fmt.Sprintf("    err := d.DB(ctx).Where(\"name = ?\", name).First(&%s).Error\n", entityLower))
		sb.WriteString(fmt.Sprintf("    return &%s, err\n", entityLower))
		sb.WriteString("}\n\n")
		sb.WriteString("// 2. 在 internal/service 中调用\n")
		sb.WriteString(fmt.Sprintf("func (s *%sService) GetByName(ctx context.Context, name string) (*models.%s, error) {\n", entityName, entityName))
		sb.WriteString("    return s.dao.FindByName(ctx, name)\n")
		sb.WriteString("}\n\n")
		sb.WriteString("// 3. 在 internal/controller 中添加 HTTP handler\n")
		sb.WriteString(fmt.Sprintf("func (ctrl *Controller) Get%sByName(c *gin.Context) {\n", entityName))
		sb.WriteString("    name := c.Query(\"name\")\n")
		sb.WriteString(fmt.Sprintf("    result, err := ctrl.service.New%sService().GetByName(c.Request.Context(), name)\n", entityName))
		sb.WriteString("    if err != nil {\n")
		sb.WriteString("        utils.Error(c, http.StatusNotFound, err.Error())\n")
		sb.WriteString("        return\n")
		sb.WriteString("    }\n")
		sb.WriteString("    utils.Success(c, result)\n")
		sb.WriteString("}\n\n")
		sb.WriteString("// 4. 在 internal/router/router.go 中注册路由\n")
		sb.WriteString(fmt.Sprintf("%s.GET(\"/%s/search\", ctrl.Get%sByName)\n", entityLower, entityLower, entityName))
		sb.WriteString("```\n\n")
	}

	sb.WriteString("## 引入第三方依赖\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("go get github.com/some/package\n")
	sb.WriteString("go mod tidy\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## 生成 Swagger 文档\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("swag init\n")
	sb.WriteString("```\n\n")
	sb.WriteString(fmt.Sprintf("访问 http://localhost:8080/swagger/index.html 查看 API 文档。\n\n"))

	sb.WriteString("## 测试\n\n")
	sb.WriteString("运行测试：\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("go test ./...\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## 构建\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("go build -o server main.go\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## 运行\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("./server\n")
	sb.WriteString("```\n")

	return sb.String()
}

func generateReadme(projectName string, tables []TableConfig) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", projectName))
	sb.WriteString("Auto-generated backend service with Gin + GORM + PostgreSQL.\n\n")

	sb.WriteString("## Features\n\n")
	sb.WriteString("- RESTful API design\n")
	sb.WriteString("- Layered architecture (Controller/Service/DAO/Model)\n")
	sb.WriteString("- GORM Gen for type-safe queries\n")
	sb.WriteString("- Swagger API documentation\n")
	sb.WriteString("- Middleware support (CORS, Logger)\n")
	sb.WriteString("- Environment-based configuration\n\n")

	sb.WriteString("## Generated Models\n\n")
	for _, table := range tables {
		sb.WriteString(fmt.Sprintf("- **%s**: %s\n", toCamelCase(table.Name), table.Comment))
	}

	sb.WriteString("\n## Quick Start\n\n")
	sb.WriteString("### 1. Install Dependencies\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("go mod download\n")
	sb.WriteString("```\n\n")
	sb.WriteString("### 2. Configure Environment\n\n")
	sb.WriteString("Copy .env.example to .env and update the values:\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("cp .env.example .env\n")
	sb.WriteString("```\n\n")
	sb.WriteString("### 3. Run Database Migrations\n\n")
	sb.WriteString("The migrations will run automatically when you start the server.\n\n")
	sb.WriteString("### 4. Run the Server\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("go run main.go\n")
	sb.WriteString("```\n\n")
	sb.WriteString("The server will start on http://localhost:8080\n\n")
	sb.WriteString("## API Documentation\n\n")
	sb.WriteString("Access Swagger UI at: http://localhost:8080/swagger/index.html\n\n")
	sb.WriteString("## Project Structure\n\n")
	sb.WriteString("```\n")
	sb.WriteString(".\n")
	sb.WriteString("├── config/          # Configuration\n")
	sb.WriteString("├── database/        # Database connection\n")
	sb.WriteString("├── internal/\n")
	sb.WriteString("│  ├── controller/  # HTTP handlers\n")
	sb.WriteString("│  ├── dao/         # Data access objects\n")
	sb.WriteString("│  ├── middleware/  # HTTP middleware\n")
	sb.WriteString("│  ├── models/      # Data models (one file per entity)\n")
	sb.WriteString("│  ├── router/      # Route definitions\n")
	sb.WriteString("│  └── service/     # Business logic\n")
	sb.WriteString("├── pkg/utils/       # Utilities\n")
	sb.WriteString("├── docs/            # Documentation\n")
	sb.WriteString("└── migrations/      # SQL migrations\n")
	sb.WriteString("```\n\n")
	sb.WriteString("## Development\n\n")
	sb.WriteString("See docs/development_guide.md for detailed development guide.\n\n")
	sb.WriteString("See docs/config_guide.md for configuration reference.\n")

	return sb.String()
}

func generateEnvExample() string {
	return `# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=123456
DB_NAME=mydb
DB_SSLMODE=disable

# Server Configuration
SERVER_PORT=8080
SERVER_MODE=debug

# JWT Configuration
JWT_SECRET=your-secret-key
JWT_EXPIRE=24

# Log Configuration
LOG_LEVEL=info
`
}

func generateGitignore() string {
	return `# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

*.test

*.out

vendor/

go.work

.env
.env.local

.idea/
.vscode/
*.swp
*.swo
*~

.DS_Store
Thumbs.db

/server
/build/
/dist/

*.log

*.db
*.sqlite
`
}
