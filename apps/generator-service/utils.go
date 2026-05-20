package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
)

func safeSprintf(format string, args ...interface{}) string {
	placeholderCount := strings.Count(format, "%s") + strings.Count(format, "%d") + strings.Count(format, "%v") +
		strings.Count(format, "%+v") + strings.Count(format, "%#v") + strings.Count(format, "%T") +
		strings.Count(format, "%%")
	escapedPercentCount := strings.Count(format, "%%")
	placeholderCount -= escapedPercentCount

	if placeholderCount != len(args) {
		log.Printf("[WARNING] 占位符数量不匹配: 模板有 %d 个占位符, 但提供了 %d 个参数", placeholderCount, len(args))
		log.Printf("[WARNING] 模板前100字符: %s", format[:min(100, len(format))])
	}
	return fmt.Sprintf(format, args...)
}

func checkUnusedImports(fileName string, content string) {
	lines := strings.Split(content, "\n")
	inImportBlock := false
	var imports []string
	var codeLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "import (") {
			inImportBlock = true
			continue
		}
		if inImportBlock && trimmed == ")" {
			inImportBlock = false
			continue
		}
		if inImportBlock && trimmed != "" && !strings.HasPrefix(trimmed, "//") {
			importPath := strings.Trim(trimmed, `"`)
			parts := strings.Split(importPath, "/")
			pkgName := parts[len(parts)-1]
			if strings.Contains(trimmed, " ") {
				aliasParts := strings.Fields(trimmed)
				if len(aliasParts) >= 2 {
					pkgName = strings.Trim(aliasParts[0], `"`)
				}
			}
			imports = append(imports, pkgName)
		} else if !inImportBlock {
			codeLines = append(codeLines, line)
		}
	}

	code := strings.Join(codeLines, "\n")
	for _, pkg := range imports {
		usagePattern := pkg + "."
		if !strings.Contains(code, usagePattern) {
			log.Printf("[WARNING] 文件 %s 中有未使用的导入: %s", fileName, pkg)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

func marshalGeneratedCode(code *GeneratedCode) string {
	data, _ := json.Marshal(code)
	return string(data)
}

func toCamelCase(s string) string {
	if len(s) == 0 {
		return ""
	}
	parts := strings.Split(s, "_")
	for i := range parts {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}

func toLowerCamelCase(s string) string {
	if len(s) == 0 {
		return ""
	}
	camel := toCamelCase(s)
	if len(camel) == 0 {
		return ""
	}
	return strings.ToLower(camel[:1]) + camel[1:]
}

func goTypeFromSQL(sqlType string) string {
	switch strings.ToLower(sqlType) {
	case "int", "integer":
		return "int"
	case "bigint":
		return "int64"
	case "varchar", "text", "char":
		return "string"
	case "boolean", "bool":
		return "bool"
	case "float", "double", "decimal", "numeric":
		return "float64"
	case "timestamp", "datetime", "date":
		return "time.Time"
	case "json", "jsonb":
		return "datatypes.JSON"
	case "uuid":
		return "string"
	default:
		return "string"
	}
}

func gormTagFromField(field TableField) string {
	tags := []string{"column:" + field.Name}
	if field.Primary {
		tags = append(tags, "primaryKey")
	}
	if !field.Nullable && !field.Primary {
		tags = append(tags, "not null")
	}
	if field.Comment != "" {
		tags = append(tags, "comment:"+field.Comment)
	}
	return `gorm:"` + strings.Join(tags, ";") + `"`
}
