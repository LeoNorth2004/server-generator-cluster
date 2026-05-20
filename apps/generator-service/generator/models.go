package main

import (
	"fmt"
	"strings"
)

func generateBaseModel(needDatatypes bool) string {
	imports := `import (
	"time"
	"gorm.io/gorm"
)`
	if needDatatypes {
		imports = `import (
	"time"
	"gorm.io/gorm"
	"gorm.io/datatypes"
)`
	}
	return fmt.Sprintf(`package models

%s

type BaseModel struct {
	ID        uint           `+"`gorm:\"primarykey\" json:\"id\"`"+`
	CreatedAt time.Time      `+"`json:\"created_at\"`"+`
	UpdatedAt time.Time      `+"`json:\"updated_at\"`"+`
	DeletedAt gorm.DeletedAt `+"`gorm:\"index\" json:\"-\"`"+`
}
`, imports)
}

func generateSingleModel(table TableConfig, needDatatypes bool) string {
	modelName := toCamelCase(table.Name)
	var sb strings.Builder
	sb.WriteString("package models\n\n")

	sb.WriteString(fmt.Sprintf("// %s %s\n", modelName, table.Comment))
	sb.WriteString(fmt.Sprintf("type %s struct {\n", modelName))
	sb.WriteString("\tBaseModel\n")

	for _, field := range table.Fields {
		if field.Name == "id" || field.Name == "created_at" || field.Name == "updated_at" || field.Name == "deleted_at" {
			continue
		}
		goType := goTypeFromSQL(field.Type)
		fieldName := toCamelCase(field.Name)
		gormTag := gormTagFromField(field)
		jsonTag := fmt.Sprintf(`json:"%s"`, field.Name)

		if field.Comment != "" {
			sb.WriteString(fmt.Sprintf("\t// %s\n", field.Comment))
		}
		sb.WriteString(fmt.Sprintf("\t%s %s `%s %s`\n", fieldName, goType, gormTag, jsonTag))
	}

	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("// TableName 指定表名\n"))
	sb.WriteString(fmt.Sprintf("func (%s) TableName() string {\n", modelName))
	sb.WriteString(fmt.Sprintf("\treturn \"%s\"\n", table.Name))
	sb.WriteString("}\n")

	return sb.String()
}
