# Generator Platform 项目文档

## 1. 项目概述

Generator Platform 是一个基于微服务架构的**代码生成器平台**，提供后端代码自动生成、用户管理、项目管理和集群管理等功能。

### 1.1 主要功能

- **用户认证与权限管理** - 基于JWT的认证系统，支持管理员和普通用户角色
- **项目管理** - 创建和管理代码生成项目，配置数据库和表结构
- **代码生成** - 基于 Gin + GORM 自动生成完整的后端项目代码
- **集群管理** - 支持 Docker 和 Kubernetes 集群的接入和管理
- **运维监控** - 集群状态监控、健康检查、指标统计

---

## 2. 技术架构

### 2.1 系统架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         Web Admin (React)                       │
│                    http://localhost:5173                        │
└─────────────────────────────┬───────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     API Gateway (Gin)                           │
│                      Port: 8080                                 │
│              路由分发 + JWT认证 + CORS                            │
└───────┬─────────┬─────────┬─────────┬─────────┬─────────┬────────┘
        │         │         │         │         │         │
        ▼         ▼         ▼         ▼         ▼         ▼
   ┌─────────┐┌─────────┐┌─────────┐┌─────────┐┌─────────┐┌─────────┐
   │  Auth   ││  User   ││Project  ││Generator││Operations││Cluster │
   │Service  ││Service  ││Service  ││Service  ││ Service  ││Service │
   │ 8082    ││ 8081    ││ 8083    ││ 8084    ││ 8085    ││ 8086   │
   └────┬────┘└────┬────┘└────┬────┘└────┬────┘└────┬────┘└────┬────┘
        │         │         │         │         │         │
        ▼         ▼         ▼         ▼         ▼         ▼
   ┌──────────────────────────────────────────────────────────────┐
   │                    PostgreSQL + Redis                          │
   └──────────────────────────────────────────────────────────────┘
```

### 2.2 技术栈

| 层级 | 技术 |
|------|------|
| 前端 | React 18 + Vite + Axios + React Router |
| 网关 | Go + Gin |
| 后端服务 | Go + Gin + GORM |
| 数据库 | PostgreSQL |
| 缓存 | Redis |
| 容器编排 | Docker + Kubernetes |

### 2.3 服务端口分配

| 服务 | 端口 | 说明 |
|------|------|------|
| API Gateway | 8080 | 统一入口，路由分发 |
| User Service | 8081 | 用户管理 |
| Auth Service | 8082 | 认证授权 |
| Project Service | 8083 | 项目管理 |
| Generator Service | 8084 | 代码生成 |
| Operations Service | 8085 | 运维监控 |
| Cluster Service | 8086 | 集群管理 |
| PostgreSQL | 5432 | 主数据库 |
| Redis | 6379 | 缓存 |

---

## 3. API 接口文档

### 3.1 统一响应格式

所有 API 响应都遵循以下格式：

```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

**响应码说明：**

| 响应码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未认证或认证失败 |
| 403 | 无权限访问 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

### 3.2 认证服务 (Authentication Service)

**基础路径：** `/api/v1/auth`

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/login` | 用户登录 | 否 |
| POST | `/register` | 用户注册 | 否 |
| GET | `/me` | 获取当前用户信息 | 是 |

#### 3.2.1 用户登录

```
POST /api/v1/auth/login
```

**请求参数：**

```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

#### 3.2.2 用户注册

```
POST /api/v1/auth/register
```

**请求参数：**

```json
{
  "username": "newuser",
  "password": "password123",
  "email": "user@example.com"
}
```

#### 3.2.3 获取当前用户

```
GET /api/v1/auth/me
```

**请求头：**

```
Authorization: Bearer <token>
```

---

### 3.3 用户服务 (User Service)

**基础路径：** `/api/v1/users`

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/` | 获取用户列表 | 是 |
| POST | `/` | 创建用户 | 是 |
| GET | `/:id` | 获取用户详情 | 是 |
| PUT | `/:id` | 更新用户 | 是 |
| DELETE | `/:id` | 删除用户 | 是 |

---

### 3.4 项目服务 (Project Service)

**基础路径：** `/api/v1/projects`

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/` | 获取项目列表 | 是 |
| POST | `/` | 创建项目 | 是 |
| GET | `/:id` | 获取项目详情 | 是 |
| PUT | `/:id` | 更新项目 | 是 |
| DELETE | `/:id` | 删除项目 | 是 |

**项目数据结构：**

```json
{
  "id": 1,
  "user_id": 1,
  "name": "my-service",
  "description": "我的微服务项目",
  "db_config": "{\"host\":\"localhost\",\"port\":\"5432\",...}",
  "table_config": "[{\"name\":\"users\",\"fields\":[...]}]",
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

### 3.5 代码生成服务 (Generator Service)

**基础路径：** `/api/v1/generator`

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/generate` | 根据配置生成代码 | 否 |
| POST | `/generate/:id` | 从已有项目重新生成 | 否 |
| GET | `/download/:project_id` | 下载生成代码ZIP | 否 |
| GET | `/preview/:project_id` | 预览生成代码 | 否 |
| POST | `/docs/generate` | 生成文档 | 否 |

#### 3.5.1 代码生成请求

```
POST /api/v1/generator/generate
```

**请求参数：**

```json
{
  "project_name": "my-service",
  "db_config": {
    "host": "localhost",
    "port": "5432",
    "user": "postgres",
    "password": "123456",
    "db_name": "mydb"
  },
  "tables": [
    {
      "name": "users",
      "comment": "用户表",
      "fields": [
        {"name": "id", "type": "bigint", "nullable": false, "primary": true, "comment": "主键"},
        {"name": "username", "type": "varchar", "nullable": false, "comment": "用户名"},
        {"name": "email", "type": "varchar", "nullable": false, "comment": "邮箱"},
        {"name": "created_at", "type": "timestamp", "comment": "创建时间"}
      ]
    },
    {
      "name": "orders",
      "comment": "订单表",
      "fields": [
        {"name": "id", "type": "bigint", "nullable": false, "primary": true, "comment": "主键"},
        {"name": "user_id", "type": "bigint", "nullable": false, "comment": "用户ID"},
        {"name": "total_amount", "type": "decimal", "nullable": false, "comment": "订单金额"},
        {"name": "created_at", "type": "timestamp", "comment": "创建时间"}
      ]
    }
  ]
}
```

#### 3.5.2 生成代码结构

生成的代码包含以下结构：

```
my-service/
├── go.mod                    # Go模块定义
├── main.go                   # 主程序入口
├── config/
│   └── config.go            # 配置加载
├── database/
│   └── database.go           # 数据库连接
├── internal/
│   ├── models/
│   │   ├── base.go          # 基础模型
│   │   ├── users.go         # 用户模型
│   │   └── orders.go        # 订单模型
│   ├── dao/
│   │   ├── dao.go           # DAO基类
│   │   └── gen.go           # DAO生成器
│   ├── service/
│   │   └── service.go       # 业务逻辑层
│   ├── controller/
│   │   └── controller.go    # 控制器层
│   ├── router/
│   │   └── router.go        # 路由配置
│   └── middleware/
│       ├── cors.go          # CORS中间件
│       └── logger.go        # 日志中间件
├── pkg/utils/
│   ├── response.go          # 统一响应
│   └── validator.go         # 参数验证
├── docs/
│   ├── swagger.go           # Swagger文档
│   ├── config_guide.md      # 配置指南
│   └── development_guide.md # 开发指南
├── migrations/
│   ├── schema.sql           # 数据库Schema
│   └── seeds.sql           # 种子数据
├── .env.example            # 环境变量示例
├── .gitignore              # Git忽略文件
└── README.md               # 项目说明
```

---

### 3.6 运维服务 (Operations Service)

**基础路径：** `/api/v1/operations`

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/health` | 健康检查 | 是 |
| GET | `/stats` | 获取统计数据 | 是 |

**统计响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "user_count": 10,
    "project_count": 25
  }
}
```

---

### 3.7 集群服务 (Cluster Service)

**基础路径：** `/api/v1/clusters`

#### 3.7.1 集群状态管理

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/status` | 获取集群状态 |
| GET | `/metrics` | 获取集群指标 |
| GET | `/health` | 获取集群健康状态 |

#### 3.7.2 Kubernetes 管理

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/k8s/status` | K8s连接状态 |
| GET | `/k8s/info` | K8s集群信息 |
| GET | `/k8s/namespaces` | 命名空间列表 |
| GET | `/k8s/nodes` | 节点列表 |
| GET | `/k8s/pods` | Pod列表 |
| GET | `/k8s/services` | 服务列表 |
| GET | `/k8s/deployments` | 部署列表 |
| GET | `/k8s/events` | 事件列表 |
| GET | `/k8s/pods/:namespace/:name/logs` | Pod日志 |
| DELETE | `/k8s/pods/:namespace/:name` | 删除Pod |
| POST | `/k8s/deployments/:namespace/:name/scale` | 扩缩容 |
| POST | `/k8s/deployments/:namespace/:name/restart` | 重启部署 |
| GET | `/k8s/nodes/join-command` | 获取节点加入命令 |

#### 3.7.3 Docker 管理

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/docker/services` | Docker服务列表 |
| GET | `/docker/services/:service_name/logs` | 服务日志 |
| GET | `/docker/services/:service_name/stats` | 服务统计 |
| POST | `/docker/services/:service_name/restart` | 重启服务 |

---

## 4. 数据模型

### 4.1 User (用户)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| username | string | 用户名 (唯一) |
| password | string | 密码 (加密存储) |
| email | string | 邮箱 (唯一) |
| role | string | 角色: "admin" 或 "user" |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

### 4.2 Project (项目)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| user_id | uint | 所属用户ID |
| name | string | 项目名称 |
| description | string | 项目描述 |
| db_config | text | 数据库配置 (JSON) |
| table_config | text | 表配置 (JSON) |
| generated_code | text | 生成的代码 (JSON) |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

### 4.3 Cluster (集群)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| user_id | uint | 所属用户ID |
| name | string | 集群名称 |
| cluster_type | string | 集群类型: "docker" 或 "k8s" |
| docker_host | string | Docker守护进程地址 |
| api_server | string | K8s API Server地址 |
| kube_config | text | KubeConfig内容 |
| status | string | 状态: "active", "inactive", "error" |
| last_heartbeat | timestamp | 最后心跳时间 |

---

## 5. 项目目录结构

```
generator-project/
├── apps/
│   ├── api-gateway/              # API网关 (端口8080)
│   │   ├── main.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   ├── authentication-service/  # 认证服务 (端口8082)
│   │   ├── main.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   ├── user-service/             # 用户服务 (端口8081)
│   │   ├── main.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   ├── project-service/          # 项目服务 (端口8083)
│   │   ├── main.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   ├── generator-service/        # 代码生成服务 (端口8084)
│   │   ├── main.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   ├── operations-service/       # 运维服务 (端口8085)
│   │   ├── main.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   ├── cluster-service/          # 集群服务 (端口8086)
│   │   ├── main.go
│   │   ├── k8s.go               # K8s管理
│   │   ├── go.mod
│   │   └── Dockerfile
│   └── web-admin/                # React前端管理后台
│       ├── src/
│       │   ├── api.jsx          # API调用封装
│       │   ├── App.jsx          # 主应用组件
│       │   ├── main.jsx         # 入口文件
│       │   ├── pages/           # 页面组件
│       │   │   ├── Home.jsx
│       │   │   ├── Login.jsx
│       │   │   ├── Projects.jsx
│       │   │   ├── Generator.jsx
│       │   │   ├── Clusters.jsx
│       │   │   ├── Users.jsx
│       │   │   └── Docs.jsx
│       │   ├── components/      # 公共组件
│       │   │   ├── Navbar.jsx
│       │   │   ├── Sidebar.jsx
│       │   │   ├── ThemeToggle.jsx
│       │   │   └── K8sDashboard.jsx
│       │   ├── contexts/        # React上下文
│       │   │   ├── AuthContext.jsx
│       │   │   └── ThemeContext.jsx
│       │   └── assets/          # 静态资源
│       ├── public/              # 公共静态资源
│       ├── dist/                # 构建输出
│       ├── package.json
│       ├── vite.config.js
│       ├── index.html
│       ├── Dockerfile
│       └── nginx.conf
├── libs/
│   └── go-common/               # Go公共库
│       ├── config/
│       │   └── config.go        # 配置加载
│       ├── database/
│       │   └── database.go      # 数据库连接
│       ├── jwt/
│       │   └── jwt.go          # JWT工具
│       ├── middleware/
│       │   └── auth.go         # 认证中间件
│       ├── models/
│       │   └── models.go       # 数据模型定义
│       ├── redis/
│       │   └── redis.go        # Redis连接
│       ├── response/
│       │   └── response.go     # 统一响应
│       ├── go.mod
│       └── go.sum
├── infra/
│   ├── docker-debug/            # Docker Compose配置
│   │   └── docker-compose.yaml
│   └── k8s/                     # Kubernetes部署配置
│       ├── namespace.yaml
│       ├── api-gateway.yaml
│       ├── auth-service.yaml
│       ├── user-service.yaml
│       ├── project-service.yaml
│       ├── generator-service.yaml
│       ├── operations-service.yaml
│       ├── cluster-service.yaml
│       ├── postgres.yaml
│       ├── redis.yaml
│       ├── web-admin.yaml
│       └── rbac.yaml
├── README.md                     # 项目说明
├── Makefile                      # 构建脚本
├── .env                          # 环境变量
└── .env.example                  # 环境变量示例
```

---

## 6. 前端 API 调用

前端统一使用 Axios 封装 API 调用，配置文件位于 `apps/web-admin/src/api.jsx`。

### 6.1 API 封装

```javascript
import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器 - 添加Token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 响应拦截器 - 处理401
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

### 6.2 API 模块

```javascript
// 认证API
export const authAPI = {
  login: (data) => api.post('/auth/login', data),
  register: (data) => api.post('/auth/register', data),
  getMe: () => api.get('/auth/me'),
};

// 项目API
export const projectAPI = {
  list: () => api.get('/projects'),
  get: (id) => api.get(`/projects/${id}`),
  create: (data) => api.post('/projects', data),
  update: (id, data) => api.put(`/projects/${id}`, data),
  delete: (id) => api.delete(`/projects/${id}`),
};

// 代码生成API
export const generatorAPI = {
  generate: (data) => api.post('/generator/generate', data),
  generateFromProject: (id) => api.post(`/generator/generate/${id}`),
};

// 用户API
export const userAPI = {
  list: () => api.get('/users'),
  get: (id) => api.get(`/users/${id}`),
  create: (data) => api.post('/users', data),
  update: (id, data) => api.put(`/users/${id}`, data),
  delete: (id) => api.delete(`/users/${id}`),
};

// 集群API
export const clusterAPI = {
  getStatus: () => api.get('/clusters/status'),
  getMetrics: () => api.get('/clusters/metrics'),
  getHealth: () => api.get('/clusters/health'),
  getK8sStatus: () => api.get('/clusters/k8s/status'),
  getK8sPods: (namespace) => api.get('/clusters/k8s/pods', { params: { namespace } }),
  // ...更多方法
};
```

### 6.3 使用示例

```javascript
import { authAPI, projectAPI, generatorAPI } from '../api';

// 登录
const login = async () => {
  try {
    const response = await authAPI.login({
      username: 'admin',
      password: 'admin123'
    });
    localStorage.setItem('token', response.data.data.token);
  } catch (error) {
    console.error('Login failed:', error);
  }
};

// 获取项目列表
const loadProjects = async () => {
  try {
    const response = await projectAPI.list();
    setProjects(response.data.data);
  } catch (error) {
    console.error('Failed to load projects:', error);
  }
};

// 生成代码
const generateCode = async () => {
  try {
    const response = await generatorAPI.generate({
      project_name: 'my-service',
      db_config: { host: 'localhost', port: '5432', ... },
      tables: [{ name: 'users', fields: [...] }]
    });
    downloadZip(response.data.data);
  } catch (error) {
    console.error('Generation failed:', error);
  }
};
```

---

## 7. 快速开始

### 7.1 环境要求

- Go 1.21+
- Node.js 18+
- PostgreSQL 14+
- Redis 7+
- Docker (可选)
- Kubernetes (可选)

### 7.2 启动服务

```bash
# 1. 克隆项目
git clone <repository-url>
cd generator-project

# 2. 配置环境变量
cp .env.example .env
# 编辑 .env 文件配置数据库等信息

# 3. 启动数据库和Redis (Docker)
docker-compose -f infra/docker-debug/docker-compose.yaml up -d

# 4. 启动后端服务
cd apps/api-gateway && go run main.go &
cd apps/auth-service && go run main.go &
cd apps/user-service && go run main.go &
cd apps/project-service && go run main.go &
cd apps/generator-service && go run main.go &
cd apps/operations-service && go run main.go &
cd apps/cluster-service && go run main.go &

# 5. 启动前端
cd apps/web-admin
npm install
npm run dev
```

### 7.3 访问服务

- 前端管理界面: http://localhost:5173
- API网关: http://localhost:8080
- 默认管理员账号: admin / admin123

---

## 8. 部署架构

### 8.1 Docker Compose 部署

```yaml
services:
  api-gateway:
    build: ./apps/api-gateway
    ports:
      - "8080:8080"
    environment:
      - AUTH_SERVICE_URL=http://auth-service:8082
      - USER_SERVICE_URL=http://user-service:8081
      # ...

  auth-service:
    build: ./apps/authentication-service
    ports:
      - "8082:8082"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      # ...

  postgres:
    image: postgres:14
    environment:
      - POSTGRES_PASSWORD=123456
      - POSTGRES_DB=generator
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

### 8.2 Kubernetes 部署

项目提供了完整的 Kubernetes 部署配置，位于 `infra/k8s/` 目录：

```bash
# 部署所有服务
kubectl apply -f infra/k8s/namespace.yaml
kubectl apply -f infra/k8s/postgres.yaml
kubectl apply -f infra/k8s/redis.yaml
kubectl apply -f infra/k8s/api-gateway.yaml
kubectl apply -f infra/k8s/auth-service.yaml
# ...
```

---

## 9. 安全说明

### 9.1 认证机制

- 使用 JWT (JSON Web Token) 进行身份认证
- Token 默认有效期为 24 小时
- Token 在请求头中传递: `Authorization: Bearer <token>`

### 9.2 密码安全

- 密码使用 bcrypt 进行加密存储
- 不支持明文密码传输

### 9.3 API 安全

- 所有敏感 API 需要认证
- 支持 CORS 跨域配置
- 建议生产环境使用 HTTPS

---

## 10. 维护说明

### 10.1 日志

各服务日志默认输出到控制台，生产环境建议配置日志收集。

### 10.2 监控

- 集群服务提供 `/health` 和 `/metrics` 接口
- 可接入 Prometheus + Grafana 进行监控

### 10.3 备份

- PostgreSQL 数据建议每日备份
- 可使用 `pg_dump` 进行数据库导出
