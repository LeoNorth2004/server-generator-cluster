---
AIGC:
    ContentProducer: Minimax Agent AI
    ContentPropagator: Minimax Agent AI
    Label: AIGC
    ProduceID: "00000000000000000000000000000000"
    PropagateID: "00000000000000000000000000000000"
    ReservedCode1: 304502205451f4915995bf1dc59231e9238a9094a3272ec7684bf68f304c911bc0632ce902210097027dab2f629a664dfe15f934c73e280613d62fb0be58dea5b8b344538f1cdd
    ReservedCode2: 304402207a7527c6095a8306e3ae178d42d1d1aecfedc3e8554aa8de1db7fcd695582ce402207e882da466dd65013781d778cc6c502ad7ca55cfa891eb0accc49925b835cd8b
---


# 基于Gin+React的微服务容器化基本服务端构建平台

## 项目简介

这是一个完整的微服务架构代码生成平台，使用Gin、React、Docker和Kubernetes构建，可以根据用户输入的数据库和表信息快速生成可运行的Go后端代码。

## 技术栈

- **后端**: Go 1.21, Gin框架, GORM, PostgreSQL, Redis
- **前端**: React 19, Vite, React Router, Axios
- **容器化**: Docker, Docker Compose
- **编排**: Kubernetes
- **其他**: JWT认证, Swagger API文档

## 项目结构

```
project1/
├── apps/
│   ├── api-gateway/              # API网关
│   ├── authentication-service/   # 认证服务
│   ├── generator-service/       # 代码生成服务
│   ├── operations-service/      # 运维服务
│   ├── project-service/         # 项目管理服务
│   ├── user-service/            # 用户服务
│   └── web-admin/               # 前端管理界面
├── libs/
│   └── go-common/               # Go共享库
├── infra/
│   ├── docker-debug/            # Docker Compose配置
│   └── k8s/                     # Kubernetes部署文件
├── .env                          # 环境变量配置
├── Makefile                      # 构建脚本
└── README.md
```

## 快速开始

### 前置要求

- Docker 和 Docker Compose
- Go 1.21+ (本地开发用)
- Node.js 18+ (前端开发用)
- Kubernetes (可选，用于K8s部署)

### 使用Docker Compose启动

1. 首先创建数据库：
```bash
docker run --name generator-postgres -e POSTGRES_PASSWORD=123456 -p 5432:5432 -d postgres:15-alpine
docker exec -it generator-postgres psql -U postgres
# 在psql中执行:
CREATE DATABASE generator_platform;
\q
```

2. 启动所有服务：
```bash
make run
# 或者
docker-compose -f infra/docker-debug/docker-compose.yaml up -d
```

3. 访问应用：
- 前端: http://localhost:3000
- API网关: http://localhost:8080

### 本地开发

#### 后端服务

每个服务都可以独立运行：

```bash
cd apps/user-service
go mod download
go run main.go
```

#### 前端

```bash
cd apps/web-admin
npm install
npm run dev
```

## 功能说明

### 1. 用户认证
- 用户注册/登录
- JWT Token认证
- 安全的密码加密

### 2. 代码生成器
- 配置数据库连接信息
- 配置表结构和字段
- 一键生成完整的Go后端代码，包括：
  - go.mod (依赖管理)
  - main.go (入口文件)
  - config/ (配置文件)
  - models/ (数据模型)
  - handlers/ (HTTP处理器)
  - routes/ (路由配置)
  - database/ (数据库连接)
  - migrations/schema.sql (SQL建表语句)
  - README.md (项目说明)
- 支持Swagger API文档
- 生成的代码包含完整的CRUD操作

### 3. 项目管理
- 查看历史生成的项目
- 重新生成代码
- 删除项目

### 4. 微服务架构
- API网关统一入口
- 各个服务独立部署和扩展
- 服务间通过HTTP通信

## Kubernetes部署

### 1. 构建镜像
```bash
make build
```

### 2. 部署到K8s
```bash
make deploy
# 或者
kubectl apply -f infra/k8s/
```

### 3. 检查部署状态
```bash
kubectl get pods -n generator-platform
kubectl get services -n generator-platform
```

## 生成的代码特点

1. **完整的Gin框架应用**
2. **GORM ORM集成**
3. **Swagger API文档**
4. **标准CRUD接口**
5. **分层架构** (config, models, handlers, routes, database)
6. **PostgreSQL支持**
7. **易于二次开发和扩展**

## 环境变量

主要环境变量在`.env`文件中配置：

- 数据库配置: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
- Redis配置: REDIS_HOST, REDIS_PORT, REDIS_PASSWORD, REDIS_DB
- JWT配置: JWT_SECRET, JWT_EXPIRE
- 服务端口: API_GATEWAY_PORT, USER_SERVICE_PORT等

## 开发建议

1. 首次使用建议先通过Docker Compose启动
2. 生成的代码可以直接下载并运行
3. 每个微服务都可以独立进行单元测试
4. 生产环境请务必修改JWT_SECRET和数据库密码

## 许可证

MIT License
