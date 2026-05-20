# Server Generator Cluster - Web Admin Panel

**服务器生成器集群管理后台** - 基于 React 19 + Vite 8 构建的现代化微服务管理系统

## 📋 项目简介

这是一个功能完整的服务器生成器集群管理后台系统，提供项目管理、代码生成、Kubernetes/Docker 集群监控、用户权限管理等核心功能。

## ✨ 核心功能

### 🎯 主要模块
- **🏠 仪表盘** - 系统概览与实时状态监控
- **📁 项目管理** - 项目的创建、编辑、删除和代码生成
- **⚙️ 代码生成器** - 基于模板的自动化代码生成与下载
- **☸️ 集群管理** - Kubernetes 集群监控与管理（节点、Pod、Service、Deployment）
- **🐳 Docker 服务** - Docker 容器服务状态查看
- **👥 用户管理** - 用户账户的 CRUD 操作
- **📊 运维中心** - 系统健康检查、性能指标、操作日志
- **📖 文档中心** - API 文档与使用指南

### 🔐 安全特性
- JWT Token 认证机制
- 路由级别的权限控制（Protected Routes）
- 自动 Token 过期处理与重定向
- 401/403 错误统一拦截

### 🌍 国际化支持
- 中文 / English 双语切换
- 基于 i18next 的完整国际化方案
- 动态语言切换无需刷新页面

### 🎨 UI/UX 特性
- 暗色/亮色主题自由切换
- 响应式布局设计
- Material UI (MUI) 组件库
- NextUI 高级组件
- Recharts 数据可视化图表

## 🛠️ 技术栈

| 类别 | 技术 | 版本 |
|------|------|------|
| **框架** | React | ^19.2.4 |
| **构建工具** | Vite | ^8.0.1 |
| **CSS 框架** | TailwindCSS | ^4.2.2 |
| **UI 组件库** | MUI (Material UI) | ^7.3.9 |
| **高级组件** | NextUI | ^2.6.11 |
| **路由** | React Router DOM | ^6.22.0 |
| **HTTP 客户端** | Axios | ^1.6.7 |
| **图表库** | Recharts | ^3.8.1 |
| **数据表格** | MUI X DataGrid | ^8.28.2 |
| **国际化** | i18next + react-i18next | ^26.0.3 |
| **状态管理** | React Context API | - |

## 📁 项目结构

```
apps/web-admin/
├── public/
│   ├── favicon.svg              # 网站图标
│   └── icons.svg                # 图标资源
├── src/
│   ├── api.jsx                  # Axios API 配置与接口定义
│   ├── App.jsx                  # 应用入口与路由配置
│   ├── main.jsx                 # React 渲染入口
│   ├── App.css                  # 全局样式
│   ├── index.css                # TailwindCSS 基础样式
│   ├── assets/                  # 静态资源
│   │   └── react.svg
│   ├── components/              # 公共组件
│   │   ├── Layout.jsx           # 主布局（侧边栏+顶栏+内容区）
│   │   ├── Navbar.jsx           # 顶部导航栏
│   │   ├── Sidebar.jsx          # 侧边导航栏
│   │   ├── Cards.jsx            # 统计卡片组件
│   │   ├── OperationLogs.jsx    # 操作日志组件
│   │   └── ThemeToggle.jsx      # 主题切换按钮
│   ├── contexts/                # Context 状态管理
│   │   ├── AuthContext.jsx      # 认证上下文
│   │   ├── ThemeContext.jsx     # 主题上下文
│   │   └── I18nContext.jsx      # 国际化上下文
│   ├── pages/                   # 页面组件
│   │   ├── Home.jsx             # 首页仪表盘
│   │   ├── Login.jsx            # 登录页
│   │   ├── Projects.jsx         # 项目管理
│   │   ├── Generator.jsx        # 代码生成器
│   │   ├── Clusters.jsx         # 集群管理
│   │   ├── Users.jsx            # 用户管理
│   │   ├── Operations.jsx       # 运维中心
│   │   └── Docs.jsx             # 文档页面
│   └── i18n/                    # 国际化配置
│       ├── index.js             # i18n 初始化
│       └── locales/
│           ├── en.json          # 英文语言包
│           └── zh.json          # 中文语言包
├── .env                         # 环境变量（本地开发）
├── .env.example                 # 环境变量示例
├── Dockerfile                   # Docker 构建文件
├── docker-entrypoint.sh         # Docker 入口脚本
├── nginx.conf                   # Nginx 配置（开发）
├── nginx.docker.conf            # Nginx 配置（Docker）
├── nginx.k8s.conf               # Nginx 配置（Kubernetes）
├── package.json                 # 项目依赖
├── vite.config.js               # Vite 配置
└── eslint.config.js             # ESLint 配置
```

## 🚀 快速开始

### 环境要求

- **Node.js**: >= 18.x (推荐 LTS 版本)
- **npm**: >= 9.x 或 **yarn** / **pnpm**
- **后端 API**: 需要启动对应的后端微服务（默认端口 8080）

### 安装依赖

```bash
# 使用 npm
npm install

# 或使用 yarn
yarn install

# 或使用 pnpm
pnpm install
```

### 环境变量配置

复制环境变量示例文件并配置：

```bash
cp .env.example .env
```

编辑 `.env` 文件：

```env
# API 基础路径（Vite 开发代理会自动转发到后端）
VITE_API_URL=/api/v1

# 如果需要直接访问后端（非代理模式）
# VITE_API_URL=http://localhost:8080/api/v1
```

### 启动开发服务器

```bash
# 启动开发服务器（默认运行在 http://localhost:3000）
npm run dev

# 或使用其他包管理器
yarn dev
pnpm dev
```

开发服务器会自动将 `/api` 请求代理到 `http://localhost:8080`（后端服务）。

### 构建生产版本

```bash
# 构建生产版本
npm run build

# 预览构建结果
npm run preview
```

## 🔧 开发指南

### API 接口说明

所有 API 请求通过 [src/api.jsx](src/api.jsx) 统一管理，主要模块：

```javascript
import { authAPI, projectAPI, generatorAPI, userAPI, clusterAPI, operationsAPI } from './api';

// 示例：获取项目列表
const projects = await projectAPI.list();

// 示例：用户登录
const { data } = await authAPI.login({ username: 'admin', password: 'password' });
```

#### API 模块清单

| 模块 | 功能 | 主要接口 |
|------|------|----------|
| `authAPI` | 用户认证 | login, register, getMe |
| `projectAPI` | 项目管理 | list, get, create, update, delete, regenerate |
| `generatorAPI` | 代码生成 | generate, download, preview |
| `userAPI` | 用户管理 | list, get, create, update, delete |
| `clusterAPI` | 集群管理 | K8s/Docker 监控、扩缩容、自动保活 |
| `operationsAPI` | 运维操作 | health, stats, metrics, events, logs |

### 路由保护

所有需要认证的页面都通过 `ProtectedRoute` 组件保护：

```jsx
<Route path="/projects" element={
  <ProtectedRoute>
    <Layout><Projects /></Layout>
  </ProtectedRoute>
} />
```

未登录用户会被自动重定向到 `/login` 页面。

### 主题切换

应用支持暗色/亮色主题切换，通过 `ThemeContext` 管理：

```jsx
import { useTheme } from './contexts/ThemeContext';

const { isDarkMode, toggleTheme } = useTheme();
```

### 国际化添加新语言

1. 在 `src/i18n/locales/` 下新建语言文件（如 `ja.json`）
2. 在 `src/i18n/index.js` 中添加语言配置
3. 在 `src/i18n/locales/zh.json` 和 `en.json` 中同步翻译

## 🐳 Docker 部署

### 构建镜像

```bash
docker build -t web-admin .
```

### 使用 Docker Compose

项目已集成到主项目的 Docker Compose 中，可通过以下命令启动：

```bash
# 在根目录执行
docker-compose up -d web-admin
```

### Nginx 配置

提供三套 Nginx 配置：
- `nginx.conf` - 本地开发环境
- `nginx.docker.conf` - Docker 容器部署
- `nginx.k8s.conf` - Kubernetes 集群部署

## ☸️ Kubernetes 部署

Kubernetes 部署配置位于 `../../infra/k8s/web-admin.yaml`。

## 📊 可用脚本

| 命令 | 说明 |
|------|------|
| `npm run dev` | 启动开发服务器 (http://localhost:3000) |
| `npm run build` | 构建生产版本到 `dist/` 目录 |
| `npm run preview` | 预览生产构建结果 |
| `npm run lint` | 运行 ESLint 代码检查 |

## 🔒 安全注意事项

⚠️ **重要提醒**：

1. `.env` 文件包含敏感信息，**切勿提交到 Git 仓库**
2. 生产环境请使用 HTTPS
3. JWT Token 应设置合理的过期时间
4. 后端 API 应实施 Rate Limiting
5. 敏感操作应记录审计日志

## 🐛 常见问题

### Q: 开发时 API 请求出现 CORS 错误？
A: Vite 已配置代理（见 `vite.config.js`），确保后端服务运行在 8080 端口。

### Q: 如何修改后端 API 地址？
A: 编辑 `.env` 文件中的 `VITE_API_URL`，或修改 `vite.config.js` 的 proxy 配置。

### Q: 构建产物过大？
A: Vite 会自动进行 Tree Shaking 和 Code Splitting，如需进一步优化可考虑动态导入。

## 📄 License

本项目采用 MIT 许可证 - 查看 [LICENSE](../../LICENSE) 文件了解详情。

## 👥 贡献

欢迎提交 Issue 和 Pull Request！

---

**技术支持**: 如有问题，请查看项目文档或提交 Issue。
