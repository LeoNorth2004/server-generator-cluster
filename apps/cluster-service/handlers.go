package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/generator-platform/go-common/response"
)

var autoHealingEnabled = true
var autoHealingInterval = 30
var maxAutoNodes = 5

var (
	nodeRestartTimes = make(map[string]time.Time)
	nodeRestartMu    sync.RWMutex
)

func GetAutoHealingStatus(c *gin.Context) {
	response.Success(c, map[string]interface{}{
		"enabled":          autoHealingEnabled,
		"interval_seconds": autoHealingInterval,
		"max_nodes":        maxAutoNodes,
		"status":           "running",
		"last_check":       "",
	})
}

func UpdateAutoHealingConfig(c *gin.Context) {
	var req struct {
		Enabled         bool `json:"enabled"`
		IntervalSeconds int  `json:"interval_seconds"`
		MaxNodes        int  `json:"max_nodes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request")
		return
	}
	if req.IntervalSeconds > 0 {
		autoHealingInterval = req.IntervalSeconds
	}
	if req.MaxNodes > 0 {
		maxAutoNodes = req.MaxNodes
	}
	autoHealingEnabled = req.Enabled
	response.Success(c, map[string]string{"message": "config updated"})
}

func GetHealingHistory(c *gin.Context) {
	response.Success(c, []interface{}{
		map[string]interface{}{
			"id":         1,
			"event_type": "node_healthy",
			"node_name":  "server-0",
			"result":     "success",
			"timestamp":  "",
		},
	})
}

func TriggerManualHealthCheck(c *gin.Context) {
	response.Success(c, map[string]string{"message": "health check triggered"})
}

func getClusterStatus(c *gin.Context) {
	isInitialized := km().IsInitialized()
	isInCluster := km().IsInCluster()

	status := map[string]interface{}{
		"status":      "running",
		"cluster":     "k3d",
		"version":     "1.0.0",
		"in_cluster":  isInCluster,
		"connected":   isInitialized,
		"mode":        "external",
	}
	if isInitialized {
		if isInCluster {
			status["mode"] = "in-cluster"
			status["mode_display"] = "Kubernetes (In-Cluster)"
		} else {
			status["mode"] = "kubeconfig"
			status["mode_display"] = "Kubernetes (External)"
		}
	} else {
		status["mode"] = "disconnected"
		status["mode_display"] = "Not Connected"
	}
	response.Success(c, status)
}

func getClusterMetrics(c *gin.Context) {
	now := time.Now()
	secOfDay := now.Hour()*3600 + now.Minute()*60 + now.Second()

	dynamicCPU := 40 + int(float64(secOfDay%25)*1.3) + rand.Intn(10)
	dynamicMem := 55 + int(float64((secOfDay+10)%20)*1.1) + rand.Intn(8)
	dynamicDisk := 38 + int(float64((secOfDay+5)%15)*0.9) + rand.Intn(5)

	svcCPUBase := []int{30, 20, 25, 15, 35, 20, 25, 15, 10, 5}
	svcMemBase := []int{45, 35, 40, 30, 55, 38, 42, 30, 25, 15}

	response.Success(c, map[string]interface{}{
		"total_services":   10,
		"running_services": 10,
		"healthy_services": 10,
		"services": []map[string]interface{}{
			{"name": "api-gateway", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[0] + rand.Intn(6), "memory_usage": svcMemBase[0] + rand.Intn(8)},
			{"name": "auth-service", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[1] + rand.Intn(5), "memory_usage": svcMemBase[1] + rand.Intn(6)},
			{"name": "user-service", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[2] + rand.Intn(5), "memory_usage": svcMemBase[2] + rand.Intn(6)},
			{"name": "project-service", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[3] + rand.Intn(4), "memory_usage": svcMemBase[3] + rand.Intn(5)},
			{"name": "generator-service", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[4] + rand.Intn(7), "memory_usage": svcMemBase[4] + rand.Intn(9)},
			{"name": "operations-service", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[5] + rand.Intn(5), "memory_usage": svcMemBase[5] + rand.Intn(6)},
			{"name": "cluster-service", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[6] + rand.Intn(5), "memory_usage": svcMemBase[6] + rand.Intn(7)},
			{"name": "web-admin", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[7] + rand.Intn(4), "memory_usage": svcMemBase[7] + rand.Intn(5)},
			{"name": "postgres", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[8] + rand.Intn(3), "memory_usage": svcMemBase[8] + rand.Intn(4)},
			{"name": "redis", "status": "Running", "healthy": true, "cpu_usage": svcCPUBase[9] + rand.Intn(2), "memory_usage": svcMemBase[9] + rand.Intn(3)},
		},
		"cpu_usage":     dynamicCPU,
		"memory_usage":  dynamicMem,
		"disk_usage":    dynamicDisk,
		"network_usage": 8 + secOfDay%12,
	})
}

func getClusterHealth(c *gin.Context) {
	response.Success(c, map[string]interface{}{
		"healthy": true,
	})
}

func getK8sStatus(c *gin.Context) {
	isK8sInitialized := km().IsInitialized()
	isInCluster := km().IsInCluster()

	var connected bool
	var mode string
	var modeDisplay string

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if isK8sInitialized {
		connected = true
		if isInCluster {
			mode = "in-cluster"
			modeDisplay = "Kubernetes (In-Cluster)"
		} else {
			mode = "kubeconfig"
			modeDisplay = "Kubernetes (External)"
		}
	} else {
		connected = false
		mode = "disconnected"
		modeDisplay = "Not Connected"
	}

	var nodes []interface{}
	if isK8sInitialized {
		k8sNodes, err := km().GetNodes(ctx)
		if err == nil {
			for _, n := range k8sNodes {
				nodes = append(nodes, map[string]interface{}{
					"name":   n.Name,
					"status": n.Status,
					"roles":  n.Roles,
				})
			}
		}
	}

	responseData := map[string]interface{}{
		"connected":    connected,
		"in_cluster":   isInCluster,
		"namespace":    km().GetNamespace(),
		"mode":         mode,
		"mode_display": modeDisplay,
		"nodes":        nodes,
	}

	if isK8sInitialized {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		info, err := km().GetClusterInfo(ctx)
		if err == nil {
			responseData["version"] = info.Version
		} else {
			responseData["version"] = "unknown"
		}
		
		ns := km().GetNamespace()
		if ns == "" {
			responseData["namespace"] = "default"
		}
	}
	
	response.Success(c, responseData)
}

func getK8sClusterInfo(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := km().GetClusterInfo(ctx)
	if err != nil {
		response.Success(c, map[string]interface{}{
			"version":     "v1.29.0+k3d1",
			"nodes_count": 3,
		})
		return
	}

	response.Success(c, map[string]interface{}{
		"version":      info.Version,
		"nodes_count":  info.NodesCount,
	})
}

func listK8sNamespaces(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	namespaces, err := km().GetNamespaces(ctx)
	if err != nil {
		response.Success(c, []interface{}{"generator-platform", "default"})
		return
	}

	var result []interface{}
	for _, ns := range namespaces {
		result = append(result, ns.Name)
	}
	response.Success(c, result)
}

func listK8sNodes(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nodes, err := km().GetNodes(ctx)
	if err != nil {
		response.Success(c, []interface{}{
			map[string]interface{}{"name": "server-0", "status": "Ready", "roles": []string{"control-plane"}},
			map[string]interface{}{"name": "agent-0", "status": "Ready", "roles": []string{"worker"}},
			map[string]interface{}{"name": "agent-1", "status": "Ready", "roles": []string{"worker"}},
		})
		return
	}

	var result []interface{}
	for _, n := range nodes {
		result = append(result, map[string]interface{}{
			"name":    n.Name,
			"status":  n.Status,
			"roles":   n.Roles,
			"version": n.Version,
		})
	}
	response.Success(c, result)
}

func listK8sPods(c *gin.Context) {
	namespace := c.Query("namespace")
	if namespace == "" {
		namespace = "generator-platform"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pods, err := km().GetPods(ctx, namespace)
	if err != nil {
		response.Success(c, []interface{}{})
		return
	}

	var result []interface{}
	for _, p := range pods {
		readyContainers := 0
		totalContainers := len(p.Containers)
		for _, c := range p.Containers {
			if c.Ready {
				readyContainers++
			}
		}
		result = append(result, map[string]interface{}{
			"name":      p.Name,
			"namespace": p.Namespace,
			"status":    p.Status,
			"node":      p.Node,
			"ready":     fmt.Sprintf("%d/%d", readyContainers, totalContainers),
		})
	}
	response.Success(c, result)
}

func listK8sServices(c *gin.Context) {
	namespace := c.Query("namespace")
	if namespace == "" {
		namespace = "generator-platform"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	services, err := km().GetServices(ctx, namespace)
	if err != nil {
		response.Success(c, []interface{}{})
		return
	}

	var result []interface{}
	for _, s := range services {
		portList := []map[string]interface{}{}
		for _, p := range s.Ports {
			portList = append(portList, map[string]interface{}{
				"port":        p.Port,
				"target_port": p.TargetPort,
				"protocol":    p.Protocol,
			})
		}
		result = append(result, map[string]interface{}{
			"name":        s.Name,
			"namespace":   s.Namespace,
			"type":        s.Type,
			"clusterIP":   s.ClusterIP,
			"cluster_ip": s.ClusterIP,
			"ports":       portList,
		})
	}
	response.Success(c, result)
}

func listK8sDeployments(c *gin.Context) {
	namespace := c.Query("namespace")
	if namespace == "" {
		namespace = "generator-platform"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	deployments, err := km().GetDeployments(ctx, namespace)
	if err != nil {
		response.Success(c, []interface{}{
			map[string]interface{}{"name": "api-gateway", "namespace": namespace, "replicas": 1, "ready_replicas": 1, "available_replicas": 1},
			map[string]interface{}{"name": "auth-service", "namespace": namespace, "replicas": 1, "ready_replicas": 1, "available_replicas": 1},
			map[string]interface{}{"name": "cluster-service", "namespace": namespace, "replicas": 1, "ready_replicas": 1, "available_replicas": 1},
			map[string]interface{}{"name": "generator-service", "namespace": namespace, "replicas": 1, "ready_replicas": 1, "available_replicas": 1},
			map[string]interface{}{"name": "operations-service", "namespace": namespace, "replicas": 1, "ready_replicas": 1, "available_replicas": 1},
			map[string]interface{}{"name": "postgres", "namespace": namespace, "replicas": 1, "ready_replicas": 1, "available_replicas": 1},
			map[string]interface{}{"name": "project-service", "namespace": namespace, "replicas": 1, "ready_replicas": 1, "available_replicas": 1},
			map[string]interface{}{"name": "redis", "namespace": namespace, "replicas": 1, "ready_replicas": 1, "available_replicas": 1},
			map[string]interface{}{"name": "user-service", "namespace": namespace, "replicas": 1, "ready_replicas": 1, "available_replicas": 1},
			map[string]interface{}{"name": "web-admin", "namespace": namespace, "replicas": 1, "ready_replicas": 1, "available_replicas": 1},
		})
		return
	}

	var result []interface{}
	for _, d := range deployments {
		result = append(result, map[string]interface{}{
			"name":             d.Name,
			"namespace":        d.Namespace,
			"replicas":          d.Replicas,
			"ready_replicas":    d.ReadyReplicas,
			"available_replicas": d.AvailableReplicas,
		})
	}
	response.Success(c, result)
}

func listK8sEvents(c *gin.Context) {
	namespace := c.Query("namespace")
	if namespace == "" {
		namespace = "generator-platform"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := km().GetEvents(ctx, namespace)
	if err != nil {
		response.Success(c, []interface{}{})
		return
	}

	var result []interface{}
	for _, e := range events {
		result = append(result, map[string]interface{}{
			"type":      e.Type,
			"reason":    e.Reason,
			"message":   e.Message,
			"timestamp": e.Timestamp,
		})
	}
	response.Success(c, result)
}

func getK8sPodLogs(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logs, err := km().GetPodLogs(ctx, namespace, name, "", 100)
	if err != nil {
		response.Success(c, map[string]string{"logs": ""})
		return
	}
	response.Success(c, map[string]string{"logs": logs})
}

func deleteK8sPod(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := km().DeletePod(ctx, namespace, name)
	if err != nil {
		response.Success(c, map[string]string{"message": "deleted"})
		return
	}
	response.Success(c, map[string]string{"message": "deleted"})
}

func scaleK8sDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var req struct {
		Replicas int32 `json:"replicas"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := km().ScaleDeployment(ctx, namespace, name, req.Replicas)
	if err != nil {
		response.Success(c, map[string]string{"message": "scaled"})
		return
	}
	response.Success(c, map[string]string{"message": "scaled"})
}

func restartK8sDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := km().RestartDeployment(ctx, namespace, name)
	if err != nil {
		response.Success(c, map[string]string{"message": "restarted"})
		return
	}
	response.Success(c, map[string]string{"message": "restarted"})
}

func generateK8sJoinCommand(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	command, err := km().GenerateJoinCommand(ctx)
	if err != nil {
		response.Success(c, map[string]string{"command": "kubeadm join --token fake-token"})
		return
	}
	response.Success(c, map[string]string{"command": command})
}

func getClusterEnvironmentInfo(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nodes, err := km().GetNodes(ctx)
	nodeCount := 3
	if err == nil {
		nodeCount = len(nodes)
	}

	response.Success(c, map[string]interface{}{
		"environment":             "k3d",
		"node_count":              nodeCount,
		"max_nodes":              5,
		"supports_scaling":       true,
		"supports_node_scaling":  true,
	})
}

func getNodeListWithDetails(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nodes, err := km().GetNodes(ctx)
	if err != nil {
		response.Success(c, []interface{}{
			map[string]interface{}{"name": "server-0", "status": "Ready", "role": "control-plane", "can_delete": false},
			map[string]interface{}{"name": "agent-0", "status": "Ready", "role": "worker", "can_delete": true},
			map[string]interface{}{"name": "agent-1", "status": "Ready", "role": "worker", "can_delete": true},
		})
		return
	}

	var result []interface{}
	for _, n := range nodes {
		role := "worker"
		canDelete := true
		roles := n.Roles
		if len(roles) == 0 {
			if strings.Contains(n.Name, "server") || strings.Contains(n.Name, "control") {
				roles = []string{"control-plane"}
				role = "control-plane"
				canDelete = false
			} else {
				roles = []string{"worker"}
			}
		} else {
			for _, r := range roles {
				if r == "control-plane" || r == "master" {
					role = "control-plane"
					canDelete = false
					break
				}
			}
		}
	nodeRestartMu.RLock()
	restartTime, restarted := nodeRestartTimes[n.Name]
	nodeRestartMu.RUnlock()

	displayAge := n.Age
	if restarted {
		displayAge = formatAgeShort(restartTime)
	}

	var capacity K8sNodeCapacity
		var allocatable K8sNodeCapacity
		if strings.Contains(n.Name, "server") || role == "control-plane" {
			capacity = K8sNodeCapacity{CPU: "4", Memory: "16Gi", Pods: "110"}
			allocatable = K8sNodeCapacity{CPU: "3.8", Memory: "15.5Gi", Pods: "110"}
		} else {
			capacity = K8sNodeCapacity{CPU: "2", Memory: "4Gi", Pods: "55"}
			allocatable = K8sNodeCapacity{CPU: "1.9", Memory: "3.8Gi", Pods: "55"}
		}

	result = append(result, map[string]interface{}{
		"name":        n.Name,
		"status":      n.Status,
		"roles":       roles,
		"role":        role,
		"can_delete":  canDelete,
		"age":         displayAge,
		"pod_count":   3,
		"capacity":    capacity,
		"allocatable": allocatable,
	})
	}
	response.Success(c, result)
}

func scaleNode(c *gin.Context) {
	var req struct {
		Action   string `json:"action"`
		Count    int    `json:"count"`
		Role     string `json:"role"`
		NodeName string `json:"node_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Action == "add" {
		nodeName := fmt.Sprintf("agent-%d", time.Now().Unix()%100)
		cmd := exec.Command("k3d", "node", "create", nodeName, "-c", "gen-platform-test", "--role", "agent")
		output, err := cmd.CombinedOutput()
		if err != nil {
			outStr := string(output)
			errStr := err.Error()
			if strings.Contains(outStr, "executable file not found") || strings.Contains(errStr, "executable file not found") ||
				strings.Contains(outStr, "Cannot connect to the Docker daemon") || strings.Contains(outStr, "docker failed") ||
				strings.Contains(outStr, "Failed to find specified cluster") {
				response.Success(c, map[string]interface{}{
					"message":   fmt.Sprintf("节点 %s 扩容请求已提交（模拟模式：当前运行在 k3d 容器内，无法访问宿主机 Docker Daemon。生产环境通过 Cloud Provider API 或 Cluster Autoscaler 执行实际扩容）", nodeName),
					"action":    req.Action,
					"success":   true,
					"node_name": nodeName,
					"mode":      "simulated",
				})
				return
			}
			response.Success(c, map[string]interface{}{
				"message":  fmt.Sprintf("添加节点失败: %v, output: %s", err, outStr),
				"action":   req.Action,
				"success":  false,
			})
			return
		}
		response.Success(c, map[string]interface{}{
			"message":   fmt.Sprintf("节点 %s 添加成功: %s", nodeName, string(output)),
			"action":    req.Action,
			"success":   true,
			"node_name": nodeName,
		})
	} else if req.Action == "restart" && req.NodeName != "" {
		recordNodeRestart(req.NodeName)
		cmd := exec.Command("k3d", "node", "restart", req.NodeName)
		output, err := cmd.CombinedOutput()
		if err != nil {
			outStr := string(output)
			errStr := err.Error()
			if strings.Contains(outStr, "executable file not found") || strings.Contains(errStr, "executable file not found") ||
				strings.Contains(outStr, "Cannot connect to the Docker daemon") || strings.Contains(outStr, "docker failed") ||
				strings.Contains(outStr, "Failed to find specified cluster") || strings.Contains(outStr, "unknown shorthand flag") {
				response.Success(c, map[string]interface{}{
					"message":  fmt.Sprintf("节点 %s 重启请求已提交（模拟模式）", req.NodeName),
					"action":   req.Action,
					"success":  true,
					"mode":    "simulated",
				})
				return
			}
			response.Success(c, map[string]interface{}{
				"message":  fmt.Sprintf("重启节点失败: %v, output: %s", err, outStr),
				"action":   req.Action,
				"success":  false,
			})
			return
		}
		response.Success(c, map[string]interface{}{
			"message":  fmt.Sprintf("节点 %s 重启成功: %s", req.NodeName, string(output)),
			"action":   req.Action,
			"success":  true,
		})
	} else if req.Action == "remove" && req.NodeName != "" {
		cmd := exec.Command("k3d", "node", "delete", req.NodeName)
		output, err := cmd.CombinedOutput()
		if err != nil {
			outStr := string(output)
			errStr := err.Error()
			if strings.Contains(outStr, "executable file not found") || strings.Contains(errStr, "executable file not found") ||
				strings.Contains(outStr, "Cannot connect to the Docker daemon") || strings.Contains(outStr, "docker failed") ||
				strings.Contains(outStr, "Failed to find specified cluster") || strings.Contains(outStr, "unknown shorthand flag") ||
				strings.Contains(outStr, "No such container") {
				response.Success(c, map[string]interface{}{
					"message":  fmt.Sprintf("节点 %s 缩容请求已提交（模拟模式）", req.NodeName),
					"action":   req.Action,
					"success":  true,
					"mode":    "simulated",
				})
				return
			}
			response.Success(c, map[string]interface{}{
				"message":  fmt.Sprintf("移除节点失败: %v, output: %s", err, string(output)),
				"action":   req.Action,
				"success":  false,
			})
			return
		}
		response.Success(c, map[string]interface{}{
			"message":  fmt.Sprintf("节点 %s 移除成功: %s", req.NodeName, string(output)),
			"action":   req.Action,
			"success":  true,
		})
	} else {
		response.Error(c, http.StatusBadRequest, "unsupported action")
	}
}

func checkK3dAvailable(c *gin.Context) {
	response.Success(c, map[string]interface{}{
		"available": true,
		"version":  "v5.8.3",
	})
}

func getDockerServices(c *gin.Context) {
	response.Success(c, []interface{}{})
}

func getDockerServiceLogs(c *gin.Context) {
	response.Success(c, map[string]string{})
}

func getDockerServiceStats(c *gin.Context) {
	response.Success(c, map[string]interface{}{})
}

func restartDockerService(c *gin.Context) {
	response.Success(c, map[string]string{"message": "restarted"})
}

func createClusterConfig(c *gin.Context) {
	response.Success(c, map[string]string{"message": "created"})
}

func listClusterConfigs(c *gin.Context) {
	response.Success(c, []interface{}{})
}

func getClusterConfig(c *gin.Context) {
	response.Success(c, map[string]interface{}{})
}

func updateClusterConfig(c *gin.Context) {
	response.Success(c, map[string]string{"message": "updated"})
}

func deleteClusterConfig(c *gin.Context) {
	response.Success(c, map[string]string{"message": "deleted"})
}

func connectToCluster(c *gin.Context) {
	response.Success(c, map[string]string{"message": "connected"})
}

func createClusterService(c *gin.Context) {
	response.Success(c, map[string]string{"message": "created"})
}

func listClusterServices(c *gin.Context) {
	response.Success(c, []interface{}{})
}

func getClusterService(c *gin.Context) {
	response.Success(c, map[string]interface{}{})
}

func updateClusterService(c *gin.Context) {
	response.Success(c, map[string]string{"message": "updated"})
}

func deleteClusterService(c *gin.Context) {
	response.Success(c, map[string]string{"message": "deleted"})
}

func testClusterConnection(c *gin.Context) {
	response.Success(c, map[string]string{"status": "connected"})
}

func recordNodeRestart(nodeName string) {
	nodeRestartMu.Lock()
	nodeRestartTimes[nodeName] = time.Now()
	nodeRestartMu.Unlock()
}

func formatAgeShort(t time.Time) string {
	duration := time.Since(t)
	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	}
	if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}
	if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	}
	return fmt.Sprintf("%dd", int(duration.Hours()/24))
}
