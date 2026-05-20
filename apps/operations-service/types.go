package main

import (
	"time"
)

type SystemMetrics struct {
	TotalRequests     int64   `json:"total_requests"`
	AvgResponseTime   float64 `json:"avg_response_time"`
	ErrorRate         float64 `json:"error_rate"`
	ActiveConnections int64   `json:"active_connections"`
}

type ServiceStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Latency string `json:"latency"`
	Uptime  string `json:"uptime"`
}

type SystemEvent struct {
	Time    string                 `json:"time"`
	Event   string                 `json:"event"`
	Type    string                 `json:"type"`
	RawData map[string]interface{} `json:"raw_data,omitempty"`
}

type RecordOperationLogRequest struct {
	Action     string `json:"action"`
	Resource   string `json:"resource"`
	ResourceID uint   `json:"resource_id,omitempty"`
	Details    string `json:"details,omitempty"`
	Status     string `json:"status"`
	Duration   int64  `json:"duration,omitempty"`
	Error      string `json:"error,omitempty"`
}

var startTime = time.Now()

func getServicesStatusList() []ServiceStatus {
	return []ServiceStatus{
		{Name: "API Gateway", Status: "Running", Latency: "5ms", Uptime: "99.9%"},
		{Name: "Auth Service", Status: "Running", Latency: "3ms", Uptime: "99.8%"},
		{Name: "User Service", Status: "Running", Latency: "2ms", Uptime: "99.9%"},
		{Name: "Project Service", Status: "Running", Latency: "10ms", Uptime: "99.7%"},
		{Name: "Generator Service", Status: "Running", Latency: "30ms", Uptime: "99.5%"},
		{Name: "Cluster Service", Status: "Running", Latency: "15ms", Uptime: "99.6%"},
		{Name: "Operations Service", Status: "Running", Latency: "2ms", Uptime: "99.9%"},
		{Name: "PostgreSQL", Status: "Running", Latency: "1ms", Uptime: "99.9%"},
		{Name: "Redis", Status: "Running", Latency: "0ms", Uptime: "99.9%"},
	}
}
