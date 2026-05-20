package main

import (
	"fmt"
	"strings"
)

func generateService(projectName string, tables []TableConfig) string {
	var sb strings.Builder

	bt := string([]byte{96}) // backtick character
	qt := string([]byte{34}) // quote character

	header := "package service\n\nimport (\n\t\"context\"\n\t\"" + projectName + "/internal/dao\"\n\t\"" + projectName + "/internal/models\"\n)\n\ntype Service struct {\n\tdao *dao.DAO\n}\n\nfunc NewService(d *dao.DAO) *Service {\n\treturn &Service{dao: d}\n}\n\ntype ListRequest struct {\n\tPage     int    " + bt + "json" + qt + ":" + qt + "page" + qt + " form" + qt + ":" + qt + "page" + qt + bt + "\n\tPageSize int    " + bt + "json" + qt + ":" + qt + "page_size" + qt + " form" + qt + ":" + qt + "page_size" + qt + bt + "\n}\n\n"

	sb.WriteString(header)

	for _, table := range tables {
		modelName := toCamelCase(table.Name)
		varName := toLowerCamelCase(table.Name)

		sb.WriteString(fmt.Sprintf("\ntype %sService struct {\n\tservice *Service\n\tdao     *dao.%sDAO\n}\n\n", modelName, modelName))

		sb.WriteString(fmt.Sprintf("func (s *Service) New%sService() *%sService {\n\treturn &%sService{\n\t\tservice: s,\n\t\tdao:     s.dao.New%sDAO(),\n\t}\n}\n\n", modelName, modelName, modelName, modelName))

		sb.WriteString(fmt.Sprintf("type Create%sRequest struct {\n", modelName))
		for _, field := range table.Fields {
			if field.Name == "id" || field.Name == "created_at" || field.Name == "updated_at" || field.Name == "deleted_at" {
				continue
			}
			goType := goTypeFromSQL(field.Type)
			fieldName := toCamelCase(field.Name)
			sb.WriteString(fmt.Sprintf("\t%s %s "+bt+"json"+qt+":"+qt+"%s" + qt + " binding"+qt+":"+qt+"required"+qt+bt+"\n", fieldName, goType, field.Name))
		}
		sb.WriteString("}\n\n")

		sb.WriteString(fmt.Sprintf("func (s *%sService) Create(ctx context.Context, req *Create%sRequest) (*models.%s, error) {\n\t%s := &models.%s{\n", modelName, modelName, modelName, varName, modelName))
		for _, field := range table.Fields {
			if field.Name == "id" || field.Name == "created_at" || field.Name == "updated_at" || field.Name == "deleted_at" {
				continue
			}
			fieldName := toCamelCase(field.Name)
			sb.WriteString(fmt.Sprintf("\t\t%s: req.%s,\n", fieldName, fieldName))
		}
		sb.WriteString(fmt.Sprintf("\t}\n\n\tif err := s.dao.Create(ctx, %s); err != nil {\n\t\treturn nil, err\n\t}\n\treturn %s, nil\n}\n\n", varName, varName))

		sb.WriteString(fmt.Sprintf("func (s *%sService) GetByID(ctx context.Context, id uint) (*models.%s, error) {\n\treturn s.dao.GetByID(ctx, id)\n}\n\n", modelName, modelName))

		sb.WriteString(fmt.Sprintf("type %sListResponse struct {\n\tList  []*models.%s "+bt+"json"+qt+":"+qt+"list"+qt+bt+"\n\tTotal int64        "+bt+"json"+qt+":"+qt+"total"+qt+bt+"\n}\n\n", modelName, modelName))

		sb.WriteString(fmt.Sprintf("func (s *%sService) List(ctx context.Context, req *ListRequest) (*%sListResponse, error) {\n", modelName, modelName))
		sb.WriteString("\tif req.Page <= 0 { req.Page = 1 }\n")
		sb.WriteString("\tif req.PageSize <= 0 { req.PageSize = 10 }\n\n")
		sb.WriteString("\tlist, total, err := s.dao.List(ctx, req.Page, req.PageSize)\n")
		sb.WriteString("\tif err != nil { return nil, err }\n\n")
		sb.WriteString(fmt.Sprintf("\treturn &%sListResponse{ List: list, Total: total }, nil\n}\n\n", modelName))

		sb.WriteString(fmt.Sprintf("type Update%sRequest struct {\n", modelName))
		for _, field := range table.Fields {
			if field.Name == "id" || field.Name == "created_at" || field.Name == "updated_at" || field.Name == "deleted_at" {
				continue
			}
			goType := goTypeFromSQL(field.Type)
			fieldName := toCamelCase(field.Name)
			sb.WriteString(fmt.Sprintf("\t%s %s "+bt+"json"+qt+":"+qt+"%s,omitempty"+qt+bt+"\n", fieldName, goType, field.Name))
		}
		sb.WriteString("}\n\n")

		sb.WriteString(fmt.Sprintf("func (s *%sService) Update(ctx context.Context, id uint, req *Update%sRequest) (*models.%s, error) {\n\t%s, err := s.dao.GetByID(ctx, id)\n\tif err != nil { return nil, err }\n\n", modelName, modelName, modelName, varName))

		for _, field := range table.Fields {
			if field.Name == "id" || field.Name == "created_at" || field.Name == "updated_at" || field.Name == "deleted_at" {
				continue
			}
			fieldName := toCamelCase(field.Name)
			goType := goTypeFromSQL(field.Type)
			if goType == "string" {
				sb.WriteString(fmt.Sprintf("\tif req.%s != \"\" {\n\t\t%s.%s = req.%s\n\t}\n", fieldName, varName, fieldName, fieldName))
			}
		}

		sb.WriteString(fmt.Sprintf("\n\tif err := s.dao.Update(ctx, %s); err != nil {\n\t\treturn nil, err\n\t}\n\treturn %s, nil\n}\n\n", varName, varName))

		sb.WriteString(fmt.Sprintf("func (s *%sService) Delete(ctx context.Context, id uint) error {\n\treturn s.dao.Delete(ctx, id)\n}\n", modelName))
	}

	return sb.String()
}
