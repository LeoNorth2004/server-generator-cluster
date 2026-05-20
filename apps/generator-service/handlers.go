package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/generator-platform/go-common/database"
	"github.com/generator-platform/go-common/models"
	"github.com/generator-platform/go-common/response"
	"github.com/gin-gonic/gin"
)

func recordOperationLog(c *gin.Context, action, resource string, resourceID uint, details interface{}, status, errorMsg string, duration int64) {
	if database.DB == nil {
		return
	}

	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	detailsJSON, _ := json.Marshal(details)

	logEntry := models.OperationLog{
		UserID:     userID.(uint),
		Username:   username.(string),
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    string(detailsJSON),
		Status:     status,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		Duration:   duration,
		Error:      errorMsg,
	}

	if err := database.DB.Create(&logEntry).Error; err != nil {
		log.Printf("[WARNING] Failed to record operation log: %v", err)
	} else {
		log.Printf("[OPERATION] User=%s Action=%s Resource=%s Status=%s Duration=%dms",
			username, action, resource, status, duration)
	}
}

func generateCode(c *gin.Context) {
	startTime := time.Now()
	userID, _ := c.Get("user_id")

	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	log.Printf("[DEBUG] Received request: %+v", req)
	log.Printf("[DEBUG] Number of tables: %d", len(req.Tables))
	for i, table := range req.Tables {
		log.Printf("[DEBUG] Table %d: %s, Number of fields: %d", i, table.Name, len(table.Fields))
		for j, field := range table.Fields {
			log.Printf("[DEBUG]   Field %d: %s, Type: %s", j, field.Name, field.Type)
		}
	}

	if req.ProjectName == "" {
		response.BadRequest(c, "Project name is required")
		return
	}
	if len(req.Tables) == 0 {
		response.BadRequest(c, "At least one table is required")
		return
	}
	for i, table := range req.Tables {
		if table.Name == "" {
			response.BadRequest(c, fmt.Sprintf("Table %d: Name is required", i+1))
			return
		}
		if len(table.Fields) == 0 {
			response.BadRequest(c, fmt.Sprintf("Table %s: At least one field is required", table.Name))
			return
		}
		for j, field := range table.Fields {
			if field.Name == "" {
				response.BadRequest(c, fmt.Sprintf("Table %s, Field %d: Name is required", table.Name, j+1))
				return
			}
			if field.Type == "" {
				response.BadRequest(c, fmt.Sprintf("Table %s, Field %d: Type is required", table.Name, j+1))
				return
			}
		}
	}

	generated, err := doGenerate(req)
	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "generate", "project", 0, gin.H{
			"project_name": req.ProjectName,
			"tables_count": len(req.Tables),
			"error":        err.Error(),
		}, "failed", err.Error(), duration)
		response.InternalServerError(c, err.Error())
		return
	}

	var projectID uint

	if database.DB != nil {
		dbConfigJSON, _ := json.Marshal(req.DBConfig)
		tableConfigJSON, _ := json.Marshal(req.Tables)

		project := models.Project{
			UserID:        userID.(uint),
			Name:          req.ProjectName,
			DBConfig:      string(dbConfigJSON),
			TableConfig:   string(tableConfigJSON),
			GeneratedCode: marshalGeneratedCode(generated),
			Status:        "generated",  // ✅ 首次生成也设置状态
		}

		if err := database.DB.Create(&project).Error; err != nil {
			log.Printf("Warning: Failed to save project: %v", err)
		} else {
			projectID = project.ID
		}
	}

	duration := time.Since(startTime).Milliseconds()
	recordOperationLog(c, "generate", "project", projectID, gin.H{
		"project_name":    req.ProjectName,
		"tables_count":    len(req.Tables),
		"files_generated": len(generated.Files),
	}, "success", "", duration)

	response.Success(c, gin.H{
		"code":      generated,
		"project_id": projectID,
	})
}

func generateFromProject(c *gin.Context) {
	startTime := time.Now()
	projectID := c.Param("project_id")
	pid, _ := strconv.ParseUint(projectID, 10, 32)

	if database.DB == nil {
		response.InternalServerError(c, "Database not available")
		return
	}

	var project models.Project
	if err := database.DB.Where("id = ?", projectID).First(&project).Error; err != nil {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "regenerate", "project", uint(pid), gin.H{
			"project_id": projectID,
			"error":      "Project not found",
		}, "failed", "Project not found", duration)
		response.NotFound(c, "Project not found")
		return
	}

	var dbConfig DBConfig
	var tables []TableConfig

	if err := json.Unmarshal([]byte(project.DBConfig), &dbConfig); err != nil {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "regenerate", "project", uint(pid), gin.H{
			"project_id": projectID,
			"error":      "Invalid db config",
		}, "failed", "Invalid db config", duration)
		response.BadRequest(c, "Invalid db config")
		return
	}

	if err := json.Unmarshal([]byte(project.TableConfig), &tables); err != nil {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "regenerate", "project", uint(pid), gin.H{
			"project_id": projectID,
			"error":      "Invalid table config",
		}, "failed", "Invalid table config", duration)
		response.BadRequest(c, "Invalid table config")
		return
	}

	req := GenerateRequest{
		DBConfig:    dbConfig,
		Tables:      tables,
		ProjectName: project.Name,
	}

	generated, err := doGenerate(req)
	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "regenerate", "project", uint(pid), gin.H{
			"project_name": project.Name,
			"tables_count": len(tables),
			"error":        err.Error(),
		}, "failed", err.Error(), duration)
		response.InternalServerError(c, err.Error())
		return
	}

	project.GeneratedCode = marshalGeneratedCode(generated)
	project.Status = "generated"  // ✅ 更新项目状态为已生成
	database.DB.Save(&project)

	duration := time.Since(startTime).Milliseconds()
	recordOperationLog(c, "regenerate", "project", uint(pid), gin.H{
		"project_name":    project.Name,
		"tables_count":    len(tables),
		"files_generated": len(generated.Files),
	}, "success", "", duration)

	response.Success(c, gin.H{
		"project": project,
		"code":    generated,
	})
}

func downloadZip(c *gin.Context) {
	startTime := time.Now()
	projectID := c.Param("project_id")
	pid, _ := strconv.ParseUint(projectID, 10, 32)

	if database.DB == nil {
		response.InternalServerError(c, "Database not available")
		return
	}

	var project models.Project
	if err := database.DB.Where("id = ?", projectID).First(&project).Error; err != nil {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "download", "project", uint(pid), gin.H{
			"project_id": projectID,
			"error":      "Project not found",
		}, "failed", "Project not found", duration)
		response.NotFound(c, "Project not found")
		return
	}

	log.Printf("[DEBUG] Download request for project: %s (ID: %s, Status: %s)", project.Name, projectID, project.Status)

	if project.Status != "generated" {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "download", "project", uint(pid), gin.H{
			"project_id":   projectID,
			"project_name": project.Name,
			"status":       project.Status,
			"error":        "Project not ready for download",
		}, "failed", "Project not ready for download", duration)
		response.BadRequest(c, "Project is not ready for download. Please generate the code first.")
		return
	}

	var generated GeneratedCode
	if err := json.Unmarshal([]byte(project.GeneratedCode), &generated); err != nil {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "download", "project", uint(pid), gin.H{
			"project_name": project.Name,
			"error":        "Invalid generated code",
		}, "failed", "Invalid generated code", duration)
		response.BadRequest(c, "Invalid generated code data")
		return
	}

	log.Printf("[DEBUG] Generated files count: %d", len(generated.Files))

	if len(generated.Files) == 0 {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "download", "project", uint(pid), gin.H{
			"project_name": project.Name,
			"error":        "No files to download",
		}, "failed", "No files to download", duration)
		response.BadRequest(c, "No generated files available for download")
		return
	}

	zipBytes, err := generateZip(&generated)
	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "download", "project", uint(pid), gin.H{
			"project_name": project.Name,
			"error":        err.Error(),
		}, "failed", err.Error(), duration)
		response.InternalServerError(c, "Failed to generate zip file")
		return
	}

	log.Printf("[DEBUG] ZIP file size: %d bytes", len(zipBytes))

	if len(zipBytes) == 0 {
		response.InternalServerError(c, "Generated zip file is empty")
		return
	}

	duration := time.Since(startTime).Milliseconds()
	recordOperationLog(c, "download", "project", uint(pid), gin.H{
		"project_name": project.Name,
		"files_count":  len(generated.Files),
		"zip_size":     len(zipBytes),
	}, "success", "", duration)

	safeName := sanitizeFilename(project.Name)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", safeName))
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Length", fmt.Sprintf("%d", len(zipBytes)))
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(200, "application/zip", zipBytes)
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	re := regexp.MustCompile(`[^\w\-.]`)
	name = re.ReplaceAllString(name, "")
	if len(name) > 100 {
		name = name[:100]
	}
	if name == "" {
		name = "project"
	}
	return name
}

func previewCode(c *gin.Context) {
	startTime := time.Now()
	userID, _ := c.Get("user_id")
	projectID := c.Param("project_id")
	filePath := c.Query("file")
	pid, _ := strconv.ParseUint(projectID, 10, 32)

	if database.DB == nil {
		response.InternalServerError(c, "Database not available")
		return
	}

	var project models.Project
	if err := database.DB.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "preview", "project", uint(pid), gin.H{
			"project_id": projectID,
			"error":      "Project not found",
		}, "failed", "Project not found", duration)
		response.NotFound(c, "Project not found")
		return
	}

	var generated GeneratedCode
	if err := json.Unmarshal([]byte(project.GeneratedCode), &generated); err != nil {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "preview", "project", uint(pid), gin.H{
			"project_name": project.Name,
			"error":        "Invalid generated code",
		}, "failed", "Invalid generated code", duration)
		response.BadRequest(c, "Invalid generated code")
		return
	}

	if filePath == "" {
		files := make([]string, 0, len(generated.Files))
		for path := range generated.Files {
			files = append(files, path)
		}

		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "preview", "project", uint(pid), gin.H{
			"project_name": project.Name,
			"action":       "list_files",
			"files_count":  len(files),
		}, "success", "", duration)

		response.Success(c, gin.H{
			"files": files,
		})
		return
	}

	content, exists := generated.Files[filePath]
	if !exists {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "preview", "project", uint(pid), gin.H{
			"project_name": project.Name,
			"file_path":    filePath,
			"error":        "File not found",
		}, "failed", "File not found", duration)
		response.NotFound(c, "File not found")
		return
	}

	duration := time.Since(startTime).Milliseconds()
	recordOperationLog(c, "preview", "project", uint(pid), gin.H{
		"project_name":   project.Name,
		"file_path":      filePath,
		"content_length": len(content),
	}, "success", "", duration)

	response.Success(c, gin.H{
		"path":    filePath,
		"content": content,
	})
}

func generateDocumentation(c *gin.Context) {
	var req GenerateDocsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	docs, err := generateDocs(req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"docs":   docs,
		"format": req.Format,
	})
}

func generateDocs(req GenerateDocsRequest) (map[string]string, error) {
	docs := make(map[string]string)

	switch req.DocType {
	case "api":
		docs["api_documentation.md"] = generateAPIDocs(req.ProjectName, req.IncludeExamples, req.IncludeComments)
	case "config":
		docs["config_guide.md"] = generateConfigGuide()
	case "dev":
		docs["development_guide.md"] = generateDevelopmentGuide(req.ProjectName, req.Tables)
	default:
		docs["api_documentation.md"] = generateAPIDocs(req.ProjectName, req.IncludeExamples, req.IncludeComments)
		docs["config_guide.md"] = generateConfigGuide()
		docs["development_guide.md"] = generateDevelopmentGuide(req.ProjectName, req.Tables)
	}

	return docs, nil
}

func generateZip(code *GeneratedCode) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	for path, content := range code.Files {
		f, err := w.Create(path)
		if err != nil {
			return nil, err
		}
		_, err = f.Write([]byte(content))
		if err != nil {
			return nil, err
		}
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
