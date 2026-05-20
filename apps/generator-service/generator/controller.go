package main

import (
	"fmt"
	"strings"
)

func generateController(projectName string, tables []TableConfig) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`package controller

import (
	"net/http"
	"strconv"

	"%s/internal/service"
	"%s/pkg/utils"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	service *service.Service
}

func NewController(s *service.Service) *Controller {
	return &Controller{service: s}
}

`, projectName, projectName))

	for _, table := range tables {
		modelName := toCamelCase(table.Name)
		varName := toLowerCamelCase(table.Name)

		sb.WriteString(fmt.Sprintf(`

func (ctrl *Controller) Create%s(c *gin.Context) {
	var req service.Create%sRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	
	%sService := ctrl.service.New%sService()
	result, err := %sService.Create(c.Request.Context(), &req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	
	utils.Success(c, result)
}

func (ctrl *Controller) Get%s(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid id")
		return
	}
	
	%sService := ctrl.service.New%sService()
	result, err := %sService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		utils.Error(c, http.StatusNotFound, "not found")
		return
	}
	
	utils.Success(c, result)
}

func (ctrl *Controller) List%ss(c *gin.Context) {
	var req service.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 10
	}
	
	%sService := ctrl.service.New%sService()
	result, err := %sService.List(c.Request.Context(), &req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	
	utils.Success(c, result)
}

func (ctrl *Controller) Update%s(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid id")
		return
	}
	
	var req service.Update%sRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	
	%sService := ctrl.service.New%sService()
	result, err := %sService.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	
	utils.Success(c, result)
}

func (ctrl *Controller) Delete%s(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid id")
		return
	}
	
	%sService := ctrl.service.New%sService()
	if err := %sService.Delete(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	
	utils.Success(c, gin.H{"message": "deleted successfully"})
}
`,
			modelName, modelName,
			varName, modelName, varName,
			modelName,
			varName, modelName, varName,
			modelName,
			varName, modelName, varName,
			modelName, modelName,
			varName, modelName, varName,
			modelName,
			varName, modelName, varName,
		))
	}

	return sb.String()
}
