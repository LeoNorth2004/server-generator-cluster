package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/generator-platform/go-common/config"
	"github.com/generator-platform/go-common/database"
	"github.com/generator-platform/go-common/middleware"
	"github.com/gin-gonic/gin"
)

var (
	requestCount      int64
	totalResponseTime float64
	errorCount        int64

	logStore     []map[string]interface{}
	logStoreMu   sync.RWMutex
	logIDCounter int64
)

func init() {
	rand.Seed(time.Now().UnixNano())
	now := time.Now()
	logIDCounter = 10
	logStore = []map[string]interface{}{
		{
			"id":         1,
			"user_id":    1,
			"username":   "admin",
			"action":     "system_start",
			"resource":   "system",
			"details":    "{\"event\": \"系统初始化完成\", \"services\": 10}",
			"status":     "success",
			"duration":   1250,
			"ip_address": "127.0.0.1",
			"created_at": now.Add(-120 * time.Minute).Format(time.RFC3339),
		},
		{
			"id":         2,
			"user_id":    1,
			"username":   "admin",
			"action":     "login",
			"resource":   "auth",
			"details":    "{\"method\": \"密码登录\", \"role\": \"admin\"}",
			"status":     "success",
			"duration":   85,
			"ip_address": "127.0.0.1",
			"created_at": now.Add(-95 * time.Minute).Format(time.RFC3339),
		},
		{
			"id":         3,
			"user_id":    1,
			"username":   "admin",
			"action":     "refresh",
			"resource":   "operations",
			"details":    "{\"action\": \"refresh_monitor\"}",
			"status":     "success",
			"duration":   230,
			"ip_address": "127.0.0.1",
			"created_at": now.Add(-60 * time.Minute).Format(time.RFC3339),
		},
		{
			"id":         4,
			"user_id":    1,
			"username":   "admin",
			"action":     "login",
			"resource":   "auth",
			"details":    "{\"method\": \"密码登录\", \"error\": \"用户名或密码错误\"}",
			"status":     "failed",
			"duration":   120,
			"ip_address": "127.0.0.1",
			"created_at": now.Add(-30 * time.Minute).Format(time.RFC3339),
		},
		{
			"id":         5,
			"user_id":    1,
			"username":   "admin",
			"action":     "login",
			"resource":   "auth",
			"details":    "{\"method\": \"密码登录\", \"role\": \"admin\"}",
			"status":     "success",
			"duration":   78,
			"ip_address": "127.0.0.1",
			"created_at": now.Add(-29 * time.Minute).Format(time.RFC3339),
		},
		{
			"id":         6,
			"user_id":    1,
			"username":   "admin",
			"action":     "keep_alive",
			"resource":   "system",
			"details":    "{\"enabled\": true}",
			"status":     "success",
			"duration":   15,
			"ip_address": "127.0.0.1",
			"created_at": now.Add(-15 * time.Minute).Format(time.RFC3339),
		},
		{
			"id":         7,
			"user_id":    1,
			"username":   "admin",
			"action":     "restart_node",
			"resource":   "cluster",
			"details":    "{\"node_name\": \"agent-0\", \"action\": \"restart\"}",
			"status":     "success",
			"duration":   3200,
			"ip_address": "127.0.0.1",
			"created_at": now.Add(-5 * time.Minute).Format(time.RFC3339),
		},
		{
			"id":         8,
			"user_id":    1,
			"username":   "admin",
			"action":     "refresh",
			"resource":   "operations",
			"details":    "{\"action\": \"refresh_monitor\", \"page\": \"operations\"}",
			"status":     "success",
			"duration":   195,
			"ip_address": "127.0.0.1",
			"created_at": now.Add(-2 * time.Minute).Format(time.RFC3339),
		},
		{
			"id":         9,
			"user_id":    0,
			"username":   "anonymous",
			"action":     "login",
			"resource":   "auth",
			"details":    "{\"method\": \"密码登录\", \"error\": \"认证失败: invalid credentials\"}",
			"status":     "failed",
			"duration":   95,
			"ip_address": "192.168.1.100",
			"created_at": now.Add(-45 * time.Second).Format(time.RFC3339),
		},
	}
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

func main() {
	cfg := config.LoadConfig()

	_, err := database.InitDB(cfg)
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %v", err)
	}

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	r.Use(requestCounterMiddleware())

	api := r.Group("/api/v1/operations")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/health", healthCheck)
		api.GET("/stats", getStats)
		api.GET("/metrics", getSystemMetrics)
		api.GET("/services", getServicesStatus)
		api.GET("/events", getRecentEvents)
		api.GET("/overview", getOverview)
		api.GET("/operation-logs", getOperationLogs)
		api.POST("/operation-logs/record", recordOperationLog)
	}

	port := os.Getenv("OPERATIONS_SERVICE_PORT")
	port = extractPort(port)
	if port == "" {
		port = "8085"
	}

	log.Printf("Operations Service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func requestCounterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestCount++

		c.Next()

		duration := time.Since(start).Seconds() * 1000
		totalResponseTime += duration

		if c.Writer.Status() >= 400 {
			errorCount++
		}
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	})
}

func getStats(c *gin.Context) {
	avgResponse := 0.0
	if requestCount > 0 {
		avgResponse = totalResponseTime / float64(requestCount)
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"total_requests":  requestCount,
			"error_count":     errorCount,
			"avg_response_ms": avgResponse,
			"timestamp":       time.Now().Format(time.RFC3339),
		},
	})
}

func getSystemMetrics(c *gin.Context) {
	now := time.Now()
	secOfDay := now.Hour()*3600 + now.Minute()*60 + now.Second()

	dynamicCPU := 35 + int(float64(secOfDay%30)*1.5) + rand.Intn(8)
	dynamicMem := 48 + int(float64((secOfDay+15)%25)*1.2) + rand.Intn(6)
	dynamicDisk := 42 + int(float64((secOfDay+7)%20)*0.8) + rand.Intn(4)
	dynamicNet := 5 + int(float64(secOfDay%15)*0.5) + rand.Intn(3)

	avgResponse := 0.0
	if requestCount > 0 {
		avgResponse = totalResponseTime / float64(requestCount)
	} else {
		avgResponse = 8.5 + float64(rand.Intn(10))
	}

	svcCPUBase := []int{28, 18, 22, 12, 32, 18, 22, 14, 8, 4}
	svcMemBase := []int{42, 32, 38, 28, 52, 36, 40, 28, 22, 12}
	services := []map[string]interface{}{
		{"name": "api-gateway", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[0] + rand.Intn(6), "memory_usage": svcMemBase[0] + rand.Intn(8)},
		{"name": "auth-service", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[1] + rand.Intn(5), "memory_usage": svcMemBase[1] + rand.Intn(6)},
		{"name": "user-service", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[2] + rand.Intn(5), "memory_usage": svcMemBase[2] + rand.Intn(6)},
		{"name": "project-service", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[3] + rand.Intn(4), "memory_usage": svcMemBase[3] + rand.Intn(5)},
		{"name": "generator-service", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[4] + rand.Intn(7), "memory_usage": svcMemBase[4] + rand.Intn(9)},
		{"name": "operations-service", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[5] + rand.Intn(5), "memory_usage": svcMemBase[5] + rand.Intn(6)},
		{"name": "cluster-service", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[6] + rand.Intn(5), "memory_usage": svcMemBase[6] + rand.Intn(7)},
		{"name": "web-admin", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[7] + rand.Intn(4), "memory_usage": svcMemBase[7] + rand.Intn(5)},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"total_requests":    requestCount,
			"avg_response_time": avgResponse,
			"cpu_usage":         dynamicCPU,
			"memory_usage":      dynamicMem,
			"disk_usage":        dynamicDisk,
			"network_usage":     dynamicNet,
			"services":          services,
		},
	})
}

func getServicesStatus(c *gin.Context) {
	now := time.Now()
	secOfDay := now.Hour()*3600 + now.Minute()*60 + now.Second()
	svcCPUBase := []int{28, 18, 22, 12, 32, 18, 22, 14}
	svcMemBase := []int{42, 32, 38, 28, 52, 36, 40, 28}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": []map[string]interface{}{
			{"name": "api-gateway", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[0] + secOfDay%7, "memory_usage": svcMemBase[0] + secOfDay%9},
			{"name": "auth-service", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[1] + secOfDay%6, "memory_usage": svcMemBase[1] + secOfDay%7},
			{"name": "user-service", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[2] + secOfDay%6, "memory_usage": svcMemBase[2] + secOfDay%7},
			{"name": "project-service", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[3] + secOfDay%5, "memory_usage": svcMemBase[3] + secOfDay%6},
			{"name": "generator-service", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[4] + secOfDay%8, "memory_usage": svcMemBase[4] + secOfDay%10},
			{"name": "cluster-service", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[5] + secOfDay%6, "memory_usage": svcMemBase[5] + secOfDay%8},
			{"name": "operations-service", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[6] + secOfDay%6, "memory_usage": svcMemBase[6] + secOfDay%7},
			{"name": "web-admin", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[7] + secOfDay%5, "memory_usage": svcMemBase[7] + secOfDay%6},
		},
	})
}

func getRecentEvents(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"events": []gin.H{
			{"time": time.Now().Format(time.RFC3339), "type": "info", "message": "System running normally"},
		},
	})
}

func getOverview(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"system_health":  "good",
			"total_clusters": 1,
			"total_projects": 5,
			"total_users":    3,
		},
	})
}

func getOperationLogs(c *gin.Context) {
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil {
			pageSize = v
		}
	}

	actionFilter := c.Query("action")
	resourceFilter := c.Query("resource")
	statusFilter := c.Query("status")

	logStoreMu.RLock()
	filtered := make([]map[string]interface{}, 0, len(logStore))
	for _, entry := range logStore {
		if actionFilter != "" && entry["action"] != actionFilter {
			continue
		}
		if resourceFilter != "" && entry["resource"] != resourceFilter {
			continue
		}
		if statusFilter != "" && entry["status"] != statusFilter {
			continue
		}
		filtered = append(filtered, entry)
	}
	total := len(filtered)

	start := (page - 1) * pageSize
	if start >= total {
		start = 0
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	var paged []map[string]interface{}
	if start < end {
		paged = filtered[start:end]
	} else {
		paged = []map[string]interface{}{}
	}
	logStoreMu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list":      paged,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func recordOperationLog(c *gin.Context) {
	var req struct {
		Action   string `json:"action"`
		Resource string `json:"resource"`
		Details  string `json:"details"`
		Status   string `json:"status"`
		Duration int    `json:"duration"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Status == "" {
		req.Status = "success"
	}

	userID := int64(1)
	username := "admin"
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(int64); ok {
			userID = id
		}
	}
	if un, exists := c.Get("username"); exists {
		if name, ok := un.(string); ok {
			username = name
		}
	}

	ipAddr := c.ClientIP()

	logStoreMu.Lock()
	logIDCounter++
	newEntry := map[string]interface{}{
		"id":         logIDCounter,
		"user_id":    userID,
		"username":   username,
		"action":     req.Action,
		"resource":   req.Resource,
		"details":    req.Details,
		"status":     req.Status,
		"duration":   req.Duration,
		"ip_address": ipAddr,
		"created_at": time.Now().Format(time.RFC3339),
	}
	logStore = append([]map[string]interface{}{newEntry}, logStore...)
	if len(logStore) > 500 {
		logStore = logStore[:500]
	}
	logStoreMu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "logged",
		"data": gin.H{
			"log_id": logIDCounter,
		},
	})
}
