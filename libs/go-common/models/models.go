package models

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Role 角色模型
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

// User 用户模型
type User struct {
	BaseModel
	Username string    `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Password string    `gorm:"not null" json:"-"`
	Email    string    `gorm:"uniqueIndex;size:100" json:"email"`
	Role     Role      `gorm:"default:'user'" json:"role"`
	Projects []Project `gorm:"foreignKey:UserID" json:"projects,omitempty"`
	Clusters []Cluster `gorm:"foreignKey:UserID" json:"clusters,omitempty"`
}

// Project 项目模型
type Project struct {
	BaseModel
	UserID        uint   `gorm:"not null;index" json:"user_id"`
	Name          string `gorm:"not null;size:100" json:"name"`
	Description   string `gorm:"type:text" json:"description"`
	DBConfig      string `gorm:"type:text" json:"db_config"`
	TableConfig   string `gorm:"type:text" json:"table_config"`
	GeneratedCode string `gorm:"type:text" json:"generated_code"`
	Status        string `gorm:"size:20;default:pending" json:"status"`  // ✅ 新增状态字段
	User          User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Cluster Docker/K8s集群节点模型
type Cluster struct {
	BaseModel
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	Name          string    `gorm:"not null;size:100" json:"name"`
	Description   string    `gorm:"type:text" json:"description"`
	DockerHost    string    `gorm:"size:255" json:"docker_host"` // Docker守护进程地址，例如：tcp://192.168.1.100:2375
	APIServer     string    `gorm:"size:255" json:"api_server"`
	Version       string    `gorm:"size:50" json:"version"`
	NodeCount     int       `json:"node_count"`
	Status        string    `gorm:"default:'inactive'" json:"status"` // active, inactive, error
	LastHeartbeat time.Time `json:"last_heartbeat"`                   // 最后心跳时间
	// K8s相关字段
	KubeConfig   string `gorm:"type:text" json:"kube_config,omitempty"` // KubeConfig文件内容
	K8sInCluster bool   `gorm:"default:false" json:"k8s_in_cluster"`    // 是否在K8s集群内运行
	ClusterType  string `gorm:"default:'docker'" json:"cluster_type"`   // docker, k8s
	User         User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Deployment 部署记录模型
type Deployment struct {
	BaseModel
	UserID      uint    `gorm:"not null;index" json:"user_id"`
	ProjectID   uint    `gorm:"not null;index" json:"project_id"`
	ClusterID   uint    `gorm:"not null;index" json:"cluster_id"`
	Namespace   string  `gorm:"size:100" json:"namespace"`
	ServiceName string  `gorm:"size:100" json:"service_name"`
	Status      string  `gorm:"default:'pending'" json:"status"` // pending, running, failed, stopped
	Replicas    int     `json:"replicas"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	PodStatus   string  `gorm:"type:text" json:"pod_status"` // JSON格式存储Pod状态
}

// OperationLog 操作日志模型
type OperationLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UserID    uint      `gorm:"index" json:"user_id"`           // 操作用户ID
	Username  string    `gorm:"size:50" json:"username"`        // 操作用户名
	Action    string    `gorm:"not null;size:100;index" json:"action"`    // 操作类型：generate, regenerate, download, preview, login, register, etc.
	Resource  string    `gorm:"size:100" json:"resource"`       // 资源类型：project, user, cluster, code
	ResourceID uint    `json:"resource_id"`                     // 资源ID
	Details   string    `gorm:"type:text" json:"details"`        // 操作详情（JSON格式）
	Status    string    `gorm:"default:'success';size:20" json:"status"` // success, failed, error
	IPAddress string    `gorm:"size:45" json:"ip_address"`      // 客户端IP
	UserAgent string    `gorm:"size:255" json:"user_agent"`     // 客户端信息
	Duration  int64     `json:"duration"`                       // 操作耗时（毫秒）
	Error     string    `gorm:"type:text" json:"error"`          // 错误信息（如果有）
}
