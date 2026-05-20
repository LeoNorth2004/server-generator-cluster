package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sManager K8s管理器 - 管理当前项目所在的K8s集群
var (
	k8sManager     *K8sManager
	k8sManagerOnce sync.Once
)

// K8sManager K8s集群管理器
type K8sManager struct {
	clientset   *kubernetes.Clientset
	config      *rest.Config
	namespace   string
	isInCluster bool
	mu          sync.RWMutex
}

// K8sResource K8s资源基础信息
type K8sResource struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
	Age       string            `json:"age"`
	Status    string            `json:"status"`
}

// K8sPod K8s Pod信息
type K8sPod struct {
	K8sResource
	Phase      string         `json:"phase"`
	Node       string         `json:"node"`
	Restarts   int32          `json:"restarts"`
	IP         string         `json:"ip"`
	Containers []K8sContainer `json:"containers"`
}

// K8sContainer K8s容器信息
type K8sContainer struct {
	Name     string `json:"name"`
	Image    string `json:"image"`
	Ready    bool   `json:"ready"`
	State    string `json:"state"`
	Reason   string `json:"reason,omitempty"`
	Message  string `json:"message,omitempty"`
	CPUUsage string `json:"cpu_usage,omitempty"`
	MemUsage string `json:"mem_usage,omitempty"`
}

// K8sService K8s Service信息
type K8sService struct {
	K8sResource
	Type      string            `json:"type"`
	ClusterIP string            `json:"cluster_ip"`
	Ports     []K8sServicePort  `json:"ports"`
	Selector  map[string]string `json:"selector"`
}

// K8sServicePort K8s服务端口
type K8sServicePort struct {
	Name       string `json:"name"`
	Port       int32  `json:"port"`
	TargetPort int32  `json:"target_port"`
	NodePort   int32  `json:"node_port,omitempty"`
	Protocol   string `json:"protocol"`
}

// K8sDeployment K8s Deployment信息
type K8sDeployment struct {
	K8sResource
	Replicas          int32             `json:"replicas"`
	ReadyReplicas     int32             `json:"ready_replicas"`
	AvailableReplicas int32             `json:"available_replicas"`
	UpdatedReplicas   int32             `json:"updated_replicas"`
	Strategy          string            `json:"strategy"`
	Selector          map[string]string `json:"selector"`
}

// K8sNode K8s节点信息
type K8sNode struct {
	Name             string            `json:"name"`
	Status           string            `json:"status"`
	Roles            []string          `json:"roles"`
	Version          string            `json:"version"`
	OSImage          string            `json:"os_image"`
	KernelVersion    string            `json:"kernel_version"`
	ContainerRuntime string            `json:"container_runtime"`
	Labels           map[string]string `json:"labels"`
	Capacity         K8sNodeCapacity   `json:"capacity"`
	Allocatable      K8sNodeCapacity   `json:"allocatable"`
	Age              string            `json:"age"`
}

// K8sNodeCapacity K8s节点容量
type K8sNodeCapacity struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
	Pods   string `json:"pods"`
}

// K8sEvent K8s事件信息
type K8sEvent struct {
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	Object    string `json:"object"`
	Namespace string `json:"namespace"`
	Timestamp string `json:"timestamp"`
}

// K8sNamespace K8s命名空间信息
type K8sNamespace struct {
	Name   string            `json:"name"`
	Status string            `json:"status"`
	Labels map[string]string `json:"labels"`
	Age    string            `json:"age"`
}

// K8sClusterInfo K8s集群信息
type K8sClusterInfo struct {
	Version         string `json:"version"`
	Platform        string `json:"platform"`
	NodesCount      int    `json:"nodes_count"`
	NamespacesCount int    `json:"namespaces_count"`
	PodsCount       int    `json:"pods_count"`
	ServicesCount   int    `json:"services_count"`
}

// GetK8sManager 获取K8s管理器单例
func GetK8sManager() *K8sManager {
	k8sManagerOnce.Do(func() {
		k8sManager = &K8sManager{}
	})
	return k8sManager
}

// km() 为所有handler文件提供便捷访问K8sManager单例
func km() *K8sManager {
	return GetK8sManager()
}

// Init 初始化K8s客户端
func (km *K8sManager) Init(kubeConfigPath string) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	var config *rest.Config
	var err error

	// 首先尝试in-cluster配置（在K8s Pod中运行）
	config, err = rest.InClusterConfig()
	if err != nil {
		// 不在集群内，尝试使用kubeconfig文件
		km.isInCluster = false
		if kubeConfigPath != "" {
			config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		} else {
			config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				clientcmd.NewDefaultClientConfigLoadingRules(),
				&clientcmd.ConfigOverrides{},
			).ClientConfig()
		}
		if err != nil {
			return fmt.Errorf("failed to create k8s config: %v", err)
		}
	} else {
		km.isInCluster = true
		// 获取当前namespace
		km.namespace = getCurrentNamespace()
	}

	km.config = config

	// 创建clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %v", err)
	}

	// 测试连接
	_, err = clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to k8s cluster: %v", err)
	}

	km.clientset = clientset
	return nil
}

// IsInitialized 检查是否已初始化
func (km *K8sManager) IsInitialized() bool {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.clientset != nil
}

// GetClientset 获取K8s客户端
func (km *K8sManager) GetClientset() *kubernetes.Clientset {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.clientset
}

// GetNamespace 获取当前namespace
func (km *K8sManager) GetNamespace() string {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.namespace
}

// IsInCluster 是否在集群内运行
func (km *K8sManager) IsInCluster() bool {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.isInCluster
}

// GetClusterInfo 获取集群信息
func (km *K8sManager) GetClusterInfo(ctx context.Context) (*K8sClusterInfo, error) {
	clientset := km.GetClientset()
	if clientset == nil {
		return nil, fmt.Errorf("k8s client not initialized")
	}

	// 获取版本信息
	version, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, err
	}

	// 获取节点数
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 获取命名空间数
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 获取Pod数
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 获取Service数
	services, err := clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return &K8sClusterInfo{
		Version:         version.GitVersion,
		Platform:        version.Platform,
		NodesCount:      len(nodes.Items),
		NamespacesCount: len(namespaces.Items),
		PodsCount:       len(pods.Items),
		ServicesCount:   len(services.Items),
	}, nil
}

// GetPods 获取Pod列表
func (km *K8sManager) GetPods(ctx context.Context, namespace string) ([]K8sPod, error) {
	clientset := km.GetClientset()
	if clientset == nil {
		return nil, fmt.Errorf("k8s client not initialized")
	}

	// 如果没有指定namespace，获取所有
	if namespace == "" {
		namespace = ""
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []K8sPod
	for _, pod := range pods.Items {
		// 只显示项目相关的Pod
		if !isProjectPod(pod.Labels) {
			continue
		}

		k8sPod := K8sPod{
			K8sResource: K8sResource{
				Name:      pod.Name,
				Namespace: pod.Namespace,
				Labels:    pod.Labels,
				Age:       formatAge(pod.CreationTimestamp.Time),
				Status:    getPodStatus(pod),
			},
			Phase: string(pod.Status.Phase),
			Node:  pod.Spec.NodeName,
			IP:    pod.Status.PodIP,
		}

		// 计算重启次数
		var restartCount int32
		for _, containerStatus := range pod.Status.ContainerStatuses {
			restartCount += containerStatus.RestartCount
		}
		k8sPod.Restarts = restartCount

		// 获取容器信息
		for _, container := range pod.Spec.Containers {
			containerStatus := getContainerStatus(pod.Status.ContainerStatuses, container.Name)
			k8sPod.Containers = append(k8sPod.Containers, K8sContainer{
				Name:    container.Name,
				Image:   container.Image,
				Ready:   containerStatus.Ready,
				State:   getContainerState(containerStatus.State),
				Reason:  getContainerReason(containerStatus),
				Message: getContainerMessage(containerStatus),
			})
		}

		result = append(result, k8sPod)
	}

	return result, nil
}

// GetServices 获取Service列表
func (km *K8sManager) GetServices(ctx context.Context, namespace string) ([]K8sService, error) {
	clientset := km.GetClientset()
	if clientset == nil {
		return nil, fmt.Errorf("k8s client not initialized")
	}

	services, err := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []K8sService
	for _, svc := range services.Items {
		// 只显示项目相关的Service
		if !isProjectService(svc.Labels, svc.Name) {
			continue
		}

		k8sSvc := K8sService{
			K8sResource: K8sResource{
				Name:      svc.Name,
				Namespace: svc.Namespace,
				Labels:    svc.Labels,
				Age:       formatAge(svc.CreationTimestamp.Time),
			},
			Type:      string(svc.Spec.Type),
			ClusterIP: svc.Spec.ClusterIP,
			Selector:  svc.Spec.Selector,
		}

		for _, port := range svc.Spec.Ports {
			k8sSvc.Ports = append(k8sSvc.Ports, K8sServicePort{
				Name:       port.Name,
				Port:       port.Port,
				TargetPort: port.TargetPort.IntVal,
				NodePort:   port.NodePort,
				Protocol:   string(port.Protocol),
			})
		}

		result = append(result, k8sSvc)
	}

	return result, nil
}

// GetDeployments 获取Deployment列表
func (km *K8sManager) GetDeployments(ctx context.Context, namespace string) ([]K8sDeployment, error) {
	clientset := km.GetClientset()
	if clientset == nil {
		return nil, fmt.Errorf("k8s client not initialized")
	}

	deployments, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []K8sDeployment
	for _, deploy := range deployments.Items {
		// 只显示项目相关的Deployment
		if !isProjectDeployment(deploy.Labels, deploy.Name) {
			continue
		}

		replicas := int32(1)
		if deploy.Spec.Replicas != nil {
			replicas = *deploy.Spec.Replicas
		}

		result = append(result, K8sDeployment{
			K8sResource: K8sResource{
				Name:      deploy.Name,
				Namespace: deploy.Namespace,
				Labels:    deploy.Labels,
				Age:       formatAge(deploy.CreationTimestamp.Time),
				Status:    getDeploymentStatus(deploy),
			},
			Replicas:          replicas,
			ReadyReplicas:     deploy.Status.ReadyReplicas,
			AvailableReplicas: deploy.Status.AvailableReplicas,
			UpdatedReplicas:   deploy.Status.UpdatedReplicas,
			Strategy:          string(deploy.Spec.Strategy.Type),
			Selector:          deploy.Spec.Selector.MatchLabels,
		})
	}

	return result, nil
}

// GetNodes 获取Node列表
func (km *K8sManager) GetNodes(ctx context.Context) ([]K8sNode, error) {
	clientset := km.GetClientset()
	if clientset == nil {
		return nil, fmt.Errorf("k8s client not initialized")
	}

	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []K8sNode
	for _, node := range nodes.Items {
		k8sNode := K8sNode{
			Name:   node.Name,
			Labels: node.Labels,
			Age:    formatAge(node.CreationTimestamp.Time),
			Capacity: K8sNodeCapacity{
				CPU:    node.Status.Capacity.Cpu().String(),
				Memory: node.Status.Capacity.Memory().String(),
				Pods:   node.Status.Capacity.Pods().String(),
			},
			Allocatable: K8sNodeCapacity{
				CPU:    node.Status.Allocatable.Cpu().String(),
				Memory: node.Status.Allocatable.Memory().String(),
				Pods:   node.Status.Allocatable.Pods().String(),
			},
		}

		// 获取状态
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				if condition.Status == corev1.ConditionTrue {
					k8sNode.Status = "Ready"
				} else {
					k8sNode.Status = "NotReady"
				}
				break
			}
		}

		// 获取角色
		for label := range node.Labels {
			if label == "node-role.kubernetes.io/control-plane" || label == "node-role.kubernetes.io/master" {
				k8sNode.Roles = append(k8sNode.Roles, "control-plane")
			}
			if label == "node-role.kubernetes.io/worker" || label == "node-role.kubernetes.io/node" {
				k8sNode.Roles = append(k8sNode.Roles, "worker")
			}
		}

		// 获取版本信息
		k8sNode.Version = node.Status.NodeInfo.KubeletVersion
		k8sNode.OSImage = node.Status.NodeInfo.OSImage
		k8sNode.KernelVersion = node.Status.NodeInfo.KernelVersion
		k8sNode.ContainerRuntime = node.Status.NodeInfo.ContainerRuntimeVersion

		result = append(result, k8sNode)
	}

	return result, nil
}

// GetNamespaces 获取Namespace列表
func (km *K8sManager) GetNamespaces(ctx context.Context) ([]K8sNamespace, error) {
	clientset := km.GetClientset()
	if clientset == nil {
		return nil, fmt.Errorf("k8s client not initialized")
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []K8sNamespace
	for _, ns := range namespaces.Items {
		result = append(result, K8sNamespace{
			Name:   ns.Name,
			Status: string(ns.Status.Phase),
			Labels: ns.Labels,
			Age:    formatAge(ns.CreationTimestamp.Time),
		})
	}

	return result, nil
}

// GetEvents 获取事件列表
func (km *K8sManager) GetEvents(ctx context.Context, namespace string) ([]K8sEvent, error) {
	clientset := km.GetClientset()
	if clientset == nil {
		return nil, fmt.Errorf("k8s client not initialized")
	}

	events, err := clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []K8sEvent
	for _, event := range events.Items {
		// 只显示项目相关的事件
		if !isProjectEvent(event.InvolvedObject.Name) {
			continue
		}

		timestamp := event.LastTimestamp.Format("2006-01-02 15:04:05")
		if event.LastTimestamp.IsZero() {
			timestamp = event.EventTime.Format("2006-01-02 15:04:05")
		}

		result = append(result, K8sEvent{
			Type:      event.Type,
			Reason:    event.Reason,
			Message:   event.Message,
			Object:    fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name),
			Namespace: event.Namespace,
			Timestamp: timestamp,
		})
	}

	return result, nil
}

// GetPodLogs 获取Pod日志
func (km *K8sManager) GetPodLogs(ctx context.Context, namespace, name, container string, tailLines int64) (string, error) {
	clientset := km.GetClientset()
	if clientset == nil {
		return "", fmt.Errorf("k8s client not initialized")
	}

	options := &corev1.PodLogOptions{
		TailLines: &tailLines,
	}
	if container != "" {
		options.Container = container
	}

	req := clientset.CoreV1().Pods(namespace).GetLogs(name, options)
	logs, err := req.Do(ctx).Raw()
	if err != nil {
		return "", err
	}

	return string(logs), nil
}

// DeletePod 删除Pod
func (km *K8sManager) DeletePod(ctx context.Context, namespace, name string) error {
	clientset := km.GetClientset()
	if clientset == nil {
		return fmt.Errorf("k8s client not initialized")
	}

	return clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// ScaleDeployment 扩缩容Deployment
func (km *K8sManager) ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) error {
	clientset := km.GetClientset()
	if clientset == nil {
		return fmt.Errorf("k8s client not initialized")
	}

	scale, err := clientset.AppsV1().Deployments(namespace).GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	scale.Spec.Replicas = replicas
	_, err = clientset.AppsV1().Deployments(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
	return err
}

// RestartDeployment 重启Deployment
func (km *K8sManager) RestartDeployment(ctx context.Context, namespace, name string) error {
	clientset := km.GetClientset()
	if clientset == nil {
		return fmt.Errorf("k8s client not initialized")
	}

	deploy, err := clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if deploy.Spec.Template.Annotations == nil {
		deploy.Spec.Template.Annotations = make(map[string]string)
	}
	deploy.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format("2006-01-02T15:04:05Z")

	_, err = clientset.AppsV1().Deployments(namespace).Update(ctx, deploy, metav1.UpdateOptions{})
	return err
}

// GenerateJoinCommand 生成节点加入命令
func (km *K8sManager) GenerateJoinCommand(ctx context.Context) (string, error) {
	clientset := km.GetClientset()
	if clientset == nil {
		return "", fmt.Errorf("k8s client not initialized")
	}

	// 获取集群信息
	_, err := km.GetClusterInfo(ctx)
	if err != nil {
		return "", err
	}

	// 这里简化处理，实际应该从集群中获取真实的join命令
	// 真实的join命令需要通过kubeadm token create --print-join-command获取
	// 为了演示，返回一个示例命令
	controlPlaneIP := "192.168.1.100"        // 示例IP
	token := "example-token-123456"          // 示例token
	hash := "sha256:example-hash-1234567890" // 示例hash

	joinCommand := fmt.Sprintf("kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash %s",
		controlPlaneIP, token, hash)

	return joinCommand, nil
}

// GetPodMetrics 获取Pod指标（需要metrics-server）
func (km *K8sManager) GetPodMetrics(ctx context.Context, namespace string) (map[string]interface{}, error) {
	clientset := km.GetClientset()
	if clientset == nil {
		return nil, fmt.Errorf("k8s client not initialized")
	}

	// 尝试获取metrics（需要metrics-server）
	// 这里简化处理，返回基本信息
	pods, err := km.GetPods(ctx, namespace)
	if err != nil {
		return nil, err
	}

	metrics := make(map[string]interface{})
	for _, pod := range pods {
		metrics[pod.Name] = map[string]interface{}{
			"restarts": pod.Restarts,
			"status":   pod.Status,
			"phase":    pod.Phase,
		}
	}

	return metrics, nil
}

// 辅助函数

func getCurrentNamespace() string {
	if ns, ok := os.LookupEnv("POD_NAMESPACE"); ok && ns != "" {
		return ns
	}

	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err == nil {
		return string(data)
	}

	return "default"
}

func isProjectPod(labels map[string]string) bool {
	// 如果没有标签，返回true以显示所有Pod（用于调试）
	if len(labels) == 0 {
		return true
	}

	// 检查是否是项目相关的Pod
	projectLabels := []string{
		"app=api-gateway",
		"app=auth-service",
		"app=user-service",
		"app=project-service",
		"app=generator-service",
		"app=operations-service",
		"app=cluster-service",
		"app=web-admin",
		"app=postgres",
		"app=redis",
		// 也兼容旧的标签格式
		"app=generator-api-gateway",
		"app=generator-auth-service",
		"app=generator-user-service",
		"app=generator-project-service",
		"app=generator-generator-service",
		"app=generator-operations-service",
		"app=generator-cluster-service",
		"app=generator-web-admin",
	}

	for k, v := range labels {
		label := k + "=" + v
		for _, pl := range projectLabels {
			if label == pl {
				return true
			}
		}
	}

	// 如果包含 "app" 标签且值包含 "generator"，也认为是项目Pod
	if app, ok := labels["app"]; ok {
		if strings.Contains(strings.ToLower(app), "generator") ||
			strings.Contains(app, "postgres") ||
			strings.Contains(app, "redis") ||
			strings.Contains(app, "web-admin") ||
			strings.Contains(app, "-service") {
			return true
		}
	}

	// 也检查名称
	return false
}

func isProjectService(labels map[string]string, name string) bool {
	projectNames := []string{
		"api-gateway",
		"auth-service",
		"user-service",
		"project-service",
		"generator-service",
		"operations-service",
		"cluster-service",
		"web-admin",
		"postgres",
		"redis",
		// 也兼容旧的名称格式
		"generator-api-gateway",
		"generator-auth-service",
		"generator-user-service",
		"generator-project-service",
		"generator-generator-service",
		"generator-operations-service",
		"generator-cluster-service",
		"generator-web-admin",
	}

	for _, pn := range projectNames {
		if name == pn {
			return true
		}
	}

	for k, v := range labels {
		if k == "app" {
			for _, pn := range projectNames {
				if v == pn {
					return true
				}
			}
		}
	}

	return false
}

func isProjectDeployment(labels map[string]string, name string) bool {
	return isProjectService(labels, name)
}

func isProjectEvent(objectName string) bool {
	projectNames := []string{
		"api-gateway",
		"auth-service",
		"user-service",
		"project-service",
		"generator-service",
		"operations-service",
		"cluster-service",
		"web-admin",
		"postgres",
		"redis",
		// 也兼容旧的名称格式
		"generator-api-gateway",
		"generator-auth-service",
		"generator-user-service",
		"generator-project-service",
		"generator-generator-service",
		"generator-operations-service",
		"generator-cluster-service",
		"generator-web-admin",
	}

	for _, pn := range projectNames {
		if objectName == pn || len(objectName) > len(pn) && objectName[:len(pn)] == pn {
			return true
		}
	}

	return false
}

func getPodStatus(pod corev1.Pod) string {
	if pod.DeletionTimestamp != nil {
		return "Terminating"
	}

	if pod.Status.Phase == corev1.PodSucceeded {
		return "Succeeded"
	}
	if pod.Status.Phase == corev1.PodFailed {
		return "Failed"
	}

	// 检查容器状态
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.State.Waiting != nil {
			return containerStatus.State.Waiting.Reason
		}
		if containerStatus.State.Terminated != nil {
			return containerStatus.State.Terminated.Reason
		}
	}

	return string(pod.Status.Phase)
}

func getContainerStatus(statuses []corev1.ContainerStatus, name string) corev1.ContainerStatus {
	for _, status := range statuses {
		if status.Name == name {
			return status
		}
	}
	return corev1.ContainerStatus{}
}

func getContainerState(state corev1.ContainerState) string {
	if state.Running != nil {
		return "Running"
	}
	if state.Waiting != nil {
		return "Waiting"
	}
	if state.Terminated != nil {
		return "Terminated"
	}
	return "Unknown"
}

func getContainerReason(status corev1.ContainerStatus) string {
	if status.State.Waiting != nil {
		return status.State.Waiting.Reason
	}
	if status.State.Terminated != nil {
		return status.State.Terminated.Reason
	}
	return ""
}

func getContainerMessage(status corev1.ContainerStatus) string {
	if status.State.Waiting != nil {
		return status.State.Waiting.Message
	}
	if status.State.Terminated != nil {
		return status.State.Terminated.Message
	}
	return ""
}

func getDeploymentStatus(deploy interface{}) string {
	// 简化处理，返回基本状态
	return "Active"
}

func formatAge(t time.Time) string {
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
