package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/client"
)

type ServiceStatus struct {
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	Healthy     bool     `json:"healthy"`
	CPUUsage    float64  `json:"cpu_usage"`
	MemoryUsage float64  `json:"memory_usage"`
	Uptime      string   `json:"uptime"`
	Image       string   `json:"image"`
	Ports       []string `json:"ports"`
}

type ClusterMetrics struct {
	TotalServices    int             `json:"total_services"`
	RunningServices  int             `json:"running_services"`
	HealthyServices  int             `json:"healthy_services"`
	TotalCPUUsage    float64         `json:"total_cpu_usage"`
	TotalMemoryUsage float64        `json:"total_memory_usage"`
	Services         []ServiceStatus `json:"services"`
}

type ServiceLog struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

type K8sConnectionStatus struct {
	Connected bool   `json:"connected"`
	InCluster bool   `json:"in_cluster"`
	Namespace string `json:"namespace"`
	Version   string `json:"version,omitempty"`
	Error     string `json:"error,omitempty"`
}

type CreateClusterConfigRequest struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	DockerHost  string `json:"docker_host"`
	APIServer   string `json:"api_server"`
	KubeConfig  string `json:"kube_config"`
	Description string `json:"description"`
}

type UpdateClusterConfigRequest struct {
	Name        string `json:"name"`
	DockerHost  string `json:"docker_host"`
	APIServer   string `json:"api_server"`
	KubeConfig  string `json:"kube_config"`
	Description string `json:"description"`
}

var dockerCli *client.Client

func init() {
	var err error
	dockerCli, err = client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.44"),
	)
	if err != nil {
		log.Printf("Warning: Failed to create Docker client: %v", err)
	}
}

func startNodeHealthCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Println("Starting node health check service...")

	for {
		select {
		case <-ticker.C:
			checkNodeHealth()
		}
	}
}

func checkNodeHealth() {
	if !km().IsInitialized() {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nodes, err := km().GetNodes(ctx)
	if err != nil {
		log.Printf("Error getting nodes: %v", err)
		return
	}

	for _, node := range nodes {
		if node.Status != "Ready" {
			log.Printf("Node %s is not ready, status: %s", node.Name, node.Status)
		}
	}

	pods, err := km().GetPods(ctx, "")
	if err != nil {
		log.Printf("Error getting pods: %v", err)
		return
	}

	for _, pod := range pods {
		if pod.Status != "Running" {
			log.Printf("Pod %s in namespace %s is not running, status: %s", pod.Name, pod.Namespace, pod.Status)
			if pod.Status == "Failed" || strings.Contains(pod.Status, "CrashLoopBackOff") {
				log.Printf("Attempting to restart pod %s in namespace %s", pod.Name, pod.Namespace)
				err := km().DeletePod(ctx, pod.Namespace, pod.Name)
				if err != nil {
					log.Printf("Error deleting pod %s: %v", pod.Name, err)
				} else {
					log.Printf("Pod %s deleted successfully, will be restarted", pod.Name)
				}
			}
		}
	}
}

func initK8sConnection() {
	k8sMgr := GetK8sManager()

	kubeConfigPath := os.Getenv("KUBECONFIG")
	if kubeConfigPath == "" {
		kubeConfigPath = os.Getenv("HOME") + "/.kube/config"
	}

	err := k8sMgr.Init(kubeConfigPath)
	if err != nil {
		log.Printf("Warning: Failed to initialize K8s connection: %v", err)
		log.Printf("K8s features will be unavailable. Make sure running in K8s cluster or provide valid kubeconfig")
	} else {
		log.Printf("K8s connection initialized successfully")
		log.Printf("InCluster: %v, Namespace: %s", k8sMgr.IsInCluster(), k8sMgr.GetNamespace())
	}
}
