package main

import (
	"log"
	"os"

	"github.com/generator-platform/go-common/config"
	"github.com/generator-platform/go-common/database"
	"github.com/generator-platform/go-common/middleware"
	"github.com/generator-platform/go-common/models"
	"github.com/gin-gonic/gin"
)

// ===== 使用 k8s.go 中的完整 K8sManager 实现 =====
// K8sManager struct, GetK8sManager(), km(), Init(), 所有方法都在 k8s.go 中定义

func getK8sManager() interface{} {
	return GetK8sManager()
}

func main() {
	cfg := config.LoadConfig()

	_, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := database.DB.AutoMigrate(&models.Cluster{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	k8sMgr := GetK8sManager()
	err = k8sMgr.Init("")
	if err != nil {
		log.Printf("⚠️ K8s initialization warning: %v (running in standalone mode)", err)
	} else {
		log.Println("✅ K8s Manager initialized successfully")
	}

	go startNodeHealthCheck()

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	api := r.Group("/api/v1/clusters")
	api.Use(middleware.AuthMiddleware())
	api.Use(middleware.AdminMiddleware())
	{
		api.GET("/status", getClusterStatus)
		api.GET("/metrics", getClusterMetrics)
		api.GET("/health", getClusterHealth)

		api.GET("/k8s/status", getK8sStatus)
		api.GET("/k8s/info", getK8sClusterInfo)
		api.GET("/k8s/namespaces", listK8sNamespaces)
		api.GET("/k8s/nodes", listK8sNodes)
		api.GET("/k8s/pods", listK8sPods)
		api.GET("/k8s/services", listK8sServices)
		api.GET("/k8s/deployments", listK8sDeployments)
		api.GET("/k8s/events", listK8sEvents)
		api.GET("/k8s/pods/:namespace/:name/logs", getK8sPodLogs)
		api.DELETE("/k8s/pods/:namespace/:name", deleteK8sPod)
		api.POST("/k8s/deployments/:namespace/:name/scale", scaleK8sDeployment)
		api.POST("/k8s/deployments/:namespace/:name/restart", restartK8sDeployment)
		api.GET("/k8s/nodes/join-command", generateK8sJoinCommand)

		// 节点扩缩容API
		api.GET("/nodes/scaling-info", getClusterEnvironmentInfo)
		api.GET("/nodes/list-detailed", getNodeListWithDetails)
		api.POST("/nodes/scale", scaleNode)
		api.GET("/nodes/k3d-check", checkK3dAvailable)

		// 自动保活API
		api.GET("/auto-healing/status", GetAutoHealingStatus)
		api.PUT("/auto-healing/config", UpdateAutoHealingConfig)
		api.GET("/auto-healing/history", GetHealingHistory)
		api.POST("/auto-healing/trigger", TriggerManualHealthCheck)

		api.GET("/docker/services", getDockerServices)
		api.GET("/docker/services/service_name/logs", getDockerServiceLogs)
		api.GET("/docker/services/service_name/stats", getDockerServiceStats)
		api.POST("/docker/services/service_name/restart", restartDockerService)

		api.POST("/config", createClusterConfig)
		api.GET("/config", listClusterConfigs)
		api.GET("/config/id", getClusterConfig)
		api.PUT("/config/id", updateClusterConfig)
		api.DELETE("/config/id", deleteClusterConfig)
		api.POST("/config/id/connect", connectToCluster)
	}

	clusterApi := r.Group("/api/v1/cluster-services")
	clusterApi.Use(middleware.AuthMiddleware())
	{
		clusterApi.POST("", createClusterService)
		clusterApi.GET("", listClusterServices)
		clusterApi.GET("/:id", getClusterService)
		clusterApi.PUT("/:id", updateClusterService)
		clusterApi.DELETE("/:id", deleteClusterService)
		clusterApi.POST("/:id/connect", testClusterConnection)
	}

	port := os.Getenv("CLUSTER_SERVICE_PORT")
	if port == "" {
		port = "8086"
	}

	log.Printf("Cluster Service starting on port %s", port)
	log.Printf("K8s Manager initialized: %v", k8sMgr.IsInitialized())

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
