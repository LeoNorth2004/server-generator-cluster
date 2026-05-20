package main

import (
	"fmt"
	"strings"
)

func generateSQL(tables []TableConfig, dbConfig DBConfig) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("-- Database: %s\n", dbConfig.DBName))
	sb.WriteString("-- Generated schema\n\n")

	for _, table := range tables {
		sb.WriteString(fmt.Sprintf("-- Table: %s\n", table.Name))
		if table.Comment != "" {
			sb.WriteString(fmt.Sprintf("-- %s\n", table.Comment))
		}
		sb.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", table.Name))
		fields := []string{"    id SERIAL PRIMARY KEY"}

		for _, field := range table.Fields {
			if field.Name == "id" {
				continue
			}
			nullStr := ""
			if !field.Nullable {
				nullStr = " NOT NULL"
			}
			fieldDef := fmt.Sprintf("    %s %s%s", field.Name, field.Type, nullStr)
			if field.Comment != "" {
				fieldDef += fmt.Sprintf(" -- %s", field.Comment)
			}
			fields = append(fields, fieldDef)
		}

		fields = append(fields,
			"    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
			"    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
			"    deleted_at TIMESTAMP",
		)
		sb.WriteString(strings.Join(fields, ",\n"))
		sb.WriteString("\n);\n\n")

		if table.Comment != "" {
			sb.WriteString(fmt.Sprintf("COMMENT ON TABLE %s IS '%s';\n", table.Name, table.Comment))
		}

		for _, field := range table.Fields {
			if field.Comment != "" {
				sb.WriteString(fmt.Sprintf("COMMENT ON COLUMN %s.%s IS '%s';\n", table.Name, field.Name, field.Comment))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func generateSeeds(tables []TableConfig) string {
	var sb strings.Builder
	sb.WriteString("-- Seed data\n\n")

	for _, table := range tables {
		sb.WriteString(fmt.Sprintf("-- Sample data for %s\n", table.Name))
		sb.WriteString(fmt.Sprintf("-- INSERT INTO %s (", table.Name))

		fieldNames := []string{}
		for _, field := range table.Fields {
			if field.Name != "id" && field.Name != "created_at" && field.Name != "updated_at" && field.Name != "deleted_at" {
				fieldNames = append(fieldNames, field.Name)
			}
		}
		sb.WriteString(strings.Join(fieldNames, ", "))
		sb.WriteString(") VALUES (\n")
		sb.WriteString("--     'sample_value'\n")
		sb.WriteString("-- );\n\n")
	}

	return sb.String()
}

func generateAPIDocs(projectName string, includeExamples bool, includeComments bool) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s API 文档\n\n", projectName))
	sb.WriteString("## 概述\n\n")
	sb.WriteString("本文档描述了自动生成的后端服务API接口。\n\n")

	sb.WriteString("## 基础信息\n\n")
	sb.WriteString(fmt.Sprintf("- **服务名称**: %s\n", projectName))
	sb.WriteString("- **API 基础路径**: /api/v1\n")
	sb.WriteString("- **认证方式**: JWT Token\n")
	sb.WriteString("- **内容类型**: application/json\n\n")

	sb.WriteString("## 响应格式\n\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"code\": 0,\n")
	sb.WriteString("  \"message\": \"success\",\n")
	sb.WriteString("  \"data\": { ... }\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## 错误响应\n\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"code\": 400,\n")
	sb.WriteString("  \"message\": \"Bad request\",\n")
	sb.WriteString("  \"data\": null\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	if includeExamples {
		sb.WriteString("## 示例请求\n\n")
		sb.WriteString("### 获取资源列表\n\n")
		sb.WriteString("```bash\n")
		sb.WriteString("GET /api/v1/resources?page=1&page_size=10\n")
		sb.WriteString("Authorization: Bearer <token>\n")
		sb.WriteString("```\n\n")

		sb.WriteString("### 创建资源\n\n")
		sb.WriteString("```bash\n")
		sb.WriteString("POST /api/v1/resources\n")
		sb.WriteString("Authorization: Bearer <token>\n")
		sb.WriteString("Content-Type: application/json\n\n")
		sb.WriteString("{\n")
		sb.WriteString("  \"name\": \"Example\",\n")
		sb.WriteString("  \"description\": \"Example resource\"\n")
		sb.WriteString("}\n")
		sb.WriteString("```\n\n")
	}

	return sb.String()
}
