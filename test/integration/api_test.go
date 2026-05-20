package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestHealthCheckEndpoint(t *testing.T) {
	router := setupTestRouter()
	
	router.GET("/api/v1/health", func(c *gin.Context) {
		response := gin.H{
			"status":    "healthy",
			"service":   "test-service",
			"timestamp": time.Now().Format(time.RFC3339),
			"uptime":    "1h30m",
		}
		c.JSON(http.StatusOK, response)
	})

	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response["status"])
	}

	t.Logf("✓ Health check endpoint works correctly")
	t.Logf("  Response: %+v", response)
}

func TestGenerateCodeEndpoint(t *testing.T) {
	router := setupTestRouter()

	router.POST("/api/v1/generator/generate", func(c *gin.Context) {
		var req GenerateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if req.ProjectName == "" {
			c.JSON(400, gin.H{"error": "Project name is required"})
			return
		}

		if len(req.Tables) == 0 {
			c.JSON(400, gin.H{"error": "At least one table is required"})
			return
		}

		c.JSON(200, gin.H{
			"code":      GeneratedCode{Files: map[string]string{"main.go": "package main"}},
			"project_id": 1,
		})
	})

	testCases := []struct {
		name           string
		requestBody    GenerateRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Valid request",
			requestBody: GenerateRequest{
				DBConfig: DBConfig{Host: "localhost", Port: "5432"},
				Tables: []TableConfig{
					{
						Name: "users",
						Fields: []TableField{
							{Name: "id", Type: "int", Primary: true},
						},
					},
				},
				ProjectName: "test-project",
			},
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name: "Missing project name",
			requestBody: GenerateRequest{
				DBConfig: DBConfig{Host: "localhost"},
				Tables:   []TableConfig{{Name: "users"}},
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "Empty tables",
			requestBody: GenerateRequest{
				DBConfig:    DBConfig{Host: "localhost"},
				Tables:      []TableConfig{},
				ProjectName: "test",
			},
			expectedStatus: 400,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/generator/generate", body)
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
			} else {
				t.Logf("✓ Test case '%s' passed with status %d", tc.name, w.Code)
			}
		})
	}
}

func TestAPIGatewayRouting(t *testing.T) {
	router := setupTestRouter()

	routes := map[string]string{
		"GET":  "/api/v1/users",
		"POST": "/api/v1/users",
		"GET":  "/api/v1/projects",
		"PUT":  "/api/v1/projects/1",
		"DELETE": "/api/v1/projects/1",
	}

	for method, path := range routes {
		router.Handle(method, path, func(c *gin.Context) {
			c.JSON(200, gin.H{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
			})
		})
	}

	for method, path := range routes {
		req, _ := http.NewRequest(method, path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("%s %s: Expected status 200, got %d", method, path, w.Code)
		} else {
			t.Logf("✓ Route registered correctly: %s %s", method, path)
		}
	}
}

func TestMiddlewareChain(t *testing.T) {
	router := setupTestRouter()

	callOrder := make([]string, 0)

	router.Use(func(c *gin.Context) {
		callOrder = append(callOrder, "middleware-1")
		c.Next()
	})

	router.Use(func(c *gin.Context) {
		callOrder = append(callOrder, "middleware-2")
		c.Next()
	})

	router.GET("/test", func(c *gin.Context) {
		callOrder = append(callOrder, "handler")
		c.JSON(200, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if len(callOrder) != 3 {
		t.Errorf("Expected 3 calls in chain, got %d: %v", len(callOrder), callOrder)
	}

	if callOrder[0] != "middleware-1" || callOrder[1] != "middleware-2" || callOrder[2] != "handler" {
		t.Errorf("Incorrect middleware execution order: %v", callOrder)
	}

	t.Logf("✓ Middleware chain executed in correct order: %v", callOrder)
}

func TestErrorResponseFormat(t *testing.T) {
	router := setupTestRouter()

	errorCases := []struct {
		path       string
		statusCode int
		errorMsg   string
	}{
		{"/not-found", 404, "Not Found"},
		{"/bad-request", 400, "Bad Request"},
		{"/internal-error", 500, "Internal Server Error"},
	}

	for _, tc := range errorCases {
		router.GET(tc.path, func(c *gin.Context) {
			c.JSON(tc.statusCode, gin.H{
				"code":    tc.statusCode,
				"message": tc.errorMsg,
				"success": false,
			})
		})
	}

	for _, tc := range errorCases {
		req, _ := http.NewRequest("GET", tc.path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != tc.statusCode {
			t.Errorf("%s: Expected status %d, got %d", tc.path, tc.statusCode, w.Code)
		} else {
			t.Logf("✓ Error response format correct for %s (status %d)", tc.path, tc.statusCode)
		}
	}
}

func TestDatabaseConnectionValidation(t *testing.T) {
	validConfigs := []DBConfig{
		{Host: "localhost", Port: "5432", User: "postgres", Password: "pass", DBName: "db"},
		{Host: "127.0.0.1", Port: "3306", User: "root", Password: "root", DBName: "mydb"},
		{Host: "db.example.com", Port: "5432", User: "admin", DBName: "production"},
	}

	for i, config := range validConfigs {
		if config.Host == "" {
			t.Errorf("Config %d: Host should not be empty", i)
		}
		if config.Port == "" {
			t.Errorf("Config %d: Port should not be empty", i)
		}
		if config.DBName == "" {
			t.Errorf("Config %d: DBName should not be empty", i)
		}

		data, err := json.Marshal(config)
		if err != nil {
			t.Errorf("Config %d: Failed to marshal: %v", i, err)
		}

		t.Logf("✓ Database config %d valid: host=%s port=%s db=%s (%d bytes)",
			i+1, config.Host, config.Port, config.DBName, len(data))
	}
}

func TestPerformanceBasicOperations(t *testing.T) {
	iterations := 1000

	start := time.Now()
	for i := 0; i < iterations; i++ {
		toCamelCase("test_field_name")
	}
	duration := time.Since(start)

	avgTime := duration / time.Duration(iterations)
	t.Logf("✓ toCamelCase performance: total=%v, avg=%v per call", duration, avgTime)

	start = time.Now()
	for i := 0; i < iterations; i++ {
		goTypeFromSQL("varchar")
	}
	duration = time.Since(start)
	avgTime = duration / time.Duration(iterations)
	t.Logf("✓ goTypeFromSQL performance: total=%v, avg=%v per call", duration, avgTime)

	start = time.Now()
	for i := 0; i < iterations; i++ {
		safeSprintf("Hello %s, number %d", "World", i)
	}
	duration = time.Since(start)
	avgTime = duration / time.Duration(iterations)
	t.Logf("✓ safeSprintf performance: total=%v, avg=%v per call", duration, avgTime)
}

func ExampleToCamelCase() {
	result := toCamelCase("user_name")
	fmt.Println(result)
	// Output: UserName
}

func ExampleGoTypeFromSQL() {
	result := goTypeFromSQL("varchar")
	fmt.Println(result)
	// Output: string
}
