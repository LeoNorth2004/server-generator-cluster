-- Generator Platform Database Schema
-- PostgreSQL 15+
-- Generated for: generator-platform
-- Date: 2026-05-06

-- ============================================
-- 1. 创建数据库
-- ============================================
CREATE DATABASE IF NOT EXISTS generator_platform
    WITH OWNER = postgres
    ENCODING = 'UTF8'
    LC_COLLATE = 'en_US.UTF-8'
    LC_CTYPE = 'en_US.UTF-8'
    TABLESPACE = pg_default
    CONNECTION LIMIT = -1;

\c generator_platform

-- ============================================
-- 2. 启用必要的扩展
-- ============================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- 3. 用户表 (users)
-- ============================================
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE,
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('admin', 'user'))
);

-- 创建索引
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);

-- 添加注释
COMMENT ON TABLE users IS '用户表';
COMMENT ON COLUMN users.id IS '用户ID';
COMMENT ON COLUMN users.username IS '用户名（唯一）';
COMMENT ON COLUMN users.password IS '密码（加密存储）';
COMMENT ON COLUMN users.email IS '邮箱地址（唯一）';
COMMENT ON COLUMN users.role IS '角色：admin-管理员，user-普通用户';
COMMENT ON COLUMN users.deleted_at IS '软删除时间';

-- ============================================
-- 4. 项目表 (projects)
-- ============================================
CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    user_id INTEGER NOT NULL REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    db_config TEXT,
    table_config TEXT,
    generated_code TEXT,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'generating', 'generated', 'error'))
);

-- 创建索引
CREATE INDEX idx_projects_deleted_at ON projects(deleted_at);
CREATE INDEX idx_projects_user_id ON projects(user_id);
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_name ON projects(name);

-- 添加注释
COMMENT ON TABLE projects IS '项目表';
COMMENT ON COLUMN projects.id IS '项目ID';
COMMENT ON COLUMN projects.user_id IS '所属用户ID';
COMMENT ON COLUMN projects.name IS '项目名称';
COMMENT ON COLUMN projects.description IS '项目描述';
COMMENT ON COLUMN projects.db_config IS '数据库配置（JSON格式）';
COMMENT ON COLUMN projects.table_config IS '表结构配置（JSON格式）';
COMMENT ON COLUMN projects.generated_code IS '生成的代码（JSON格式）';
COMMENT ON COLUMN projects.status IS '状态：pending-待生成，generating-生成中，generated-已生成，error-错误';

-- ============================================
-- 5. 集群节点表 (clusters)
-- ============================================
CREATE TABLE IF NOT EXISTS clusters (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    user_id INTEGER NOT NULL REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    docker_host VARCHAR(255),
    api_server VARCHAR(255),
    version VARCHAR(50),
    node_count INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'inactive' CHECK (status IN ('active', 'inactive', 'error')),
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    kube_config TEXT,
    k8s_in_cluster BOOLEAN DEFAULT FALSE,
    cluster_type VARCHAR(20) DEFAULT 'docker' CHECK (cluster_type IN ('docker', 'k8s'))
);

-- 创建索引
CREATE INDEX idx_clusters_deleted_at ON clusters(deleted_at);
CREATE INDEX idx_clusters_user_id ON clusters(user_id);
CREATE INDEX idx_clusters_status ON clusters(status);
CREATE INDEX idx_clusters_cluster_type ON clusters(cluster_type);
CREATE INDEX idx_clusters_name ON clusters(name);

-- 添加注释
COMMENT ON TABLE clusters IS '集群节点表（Docker/K8s）';
COMMENT ON COLUMN clusters.id IS '集群ID';
COMMENT ON COLUMN clusters.user_id IS '所属用户ID';
COMMENT ON COLUMN clusters.name IS '集群名称';
COMMENT ON COLUMN clusters.description IS '集群描述';
COMMENT ON COLUMN clusters.docker_host IS 'Docker守护进程地址（如：tcp://192.168.1.100:2375）';
COMMENT ON COLUMN clusters.api_server IS 'Kubernetes API Server地址';
COMMENT ON COLUMN clusters.version IS 'Kubernetes版本';
COMMENT ON COLUMN clusters.node_count IS '节点数量';
COMMENT ON COLUMN clusters.status IS '状态：active-活跃，inactive-未激活，error-错误';
COMMENT ON COLUMN clusters.last_heartbeat IS '最后心跳时间';
COMMENT ON COLUMN clusters.kube_config IS 'KubeConfig文件内容';
COMMENT ON COLUMN clusters.k8s_in_cluster IS '是否在K8s集群内运行';
COMMENT ON COLUMN clusters.cluster_type IS '集群类型：docker或k8s';

-- ============================================
-- 6. 部署记录表 (deployments)
-- ============================================
CREATE TABLE IF NOT EXISTS deployments (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    user_id INTEGER NOT NULL REFERENCES users(id),
    project_id INTEGER NOT NULL REFERENCES projects(id),
    cluster_id INTEGER NOT NULL REFERENCES clusters(id),
    namespace VARCHAR(100),
    service_name VARCHAR(100),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'failed', 'stopped')),
    replicas INTEGER DEFAULT 0,
    cpu_usage DOUBLE PRECISION DEFAULT 0,
    memory_usage DOUBLE PRECISION DEFAULT 0,
    pod_status TEXT
);

-- 创建索引
CREATE INDEX idx_deployments_deleted_at ON deployments(deleted_at);
CREATE INDEX idx_deployments_user_id ON deployments(user_id);
CREATE INDEX idx_deployments_project_id ON deployments(project_id);
CREATE INDEX idx_deployments_cluster_id ON deployments(cluster_id);
CREATE INDEX idx_deployments_status ON deployments(status);

-- 添加注释
COMMENT ON TABLE deployments IS '部署记录表';
COMMENT ON COLUMN deployments.id IS '部署ID';
COMMENT ON COLUMN deployments.user_id IS '所属用户ID';
COMMENT ON COLUMN deployments.project_id IS '关联项目ID';
COMMENT ON COLUMN deployments.cluster_id IS '关联集群ID';
COMMENT ON COLUMN deployments.namespace IS '命名空间';
COMMENT ON COLUMN deployments.service_name IS '服务名称';
COMMENT ON COLUMN deployments.status IS '状态：pending-待部署，running-运行中，failed-失败，stopped-已停止';
COMMENT ON COLUMN deployments.replicas IS '副本数量';
COMMENT ON COLUMN deployments.cpu_usage IS 'CPU使用率（百分比）';
COMMENT ON COLUMN deployments.memory_usage IS '内存使用量（MB）';
COMMENT ON COLUMN deployments.pod_status IS 'Pod状态（JSON格式）';

-- ============================================
-- 7. 操作日志表 (operation_logs)
-- ============================================
CREATE TABLE IF NOT EXISTS operation_logs (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    user_id INTEGER,
    username VARCHAR(50),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100),
    resource_id BIGINT,
    details TEXT,
    status VARCHAR(20) DEFAULT 'success' CHECK (status IN ('success', 'failed', 'error')),
    ip_address VARCHAR(45),
    user_agent VARCHAR(255),
    duration BIGINT DEFAULT 0,
    error TEXT
);

-- 创建索引
CREATE INDEX idx_operation_logs_user_id ON operation_logs(user_id);
CREATE INDEX idx_operation_logs_action ON operation_logs(action);
CREATE INDEX idx_operation_logs_resource ON operation_logs(resource);
CREATE INDEX idx_operation_logs_status ON operation_logs(status);
CREATE INDEX idx_operation_logs_created_at ON operation_logs(created_at);
CREATE INDEX idx_operation_logs_resource_id ON operation_logs(resource_id);

-- 添加注释
COMMENT ON TABLE operation_logs IS '操作日志表';
COMMENT ON COLUMN operation_logs.id IS '日志ID';
COMMENT ON COLUMN operation_logs.created_at IS '操作时间';
COMMENT ON COLUMN operation_logs.user_id IS '操作用户ID';
COMMENT ON COLUMN operation_logs.username IS '操作用户名';
COMMENT ON COLUMN operation_logs.action IS '操作类型：generate, regenerate, download, preview, login, register等';
COMMENT ON COLUMN operation_logs.resource IS '资源类型：project, user, cluster, code等';
COMMENT ON COLUMN operation_logs.resource_id IS '资源ID';
COMMENT ON COLUMN operation_logs.details IS '操作详情（JSON格式）';
COMMENT ON COLUMN operation_logs.status IS '状态：success-成功，failed-失败，error-错误';
COMMENT ON COLUMN operation_logs.ip_address IS '客户端IP地址';
COMMENT ON COLUMN operation_logs.user_agent IS '客户端User-Agent信息';
COMMENT ON COLUMN operation_logs.duration IS '操作耗时（毫秒）';
COMMENT ON COLUMN operation_logs.error IS '错误信息（如果有）';

-- ============================================
-- 8. 插入默认数据
-- ============================================

-- 插入默认管理员账户（密码：admin123，已使用bcrypt加密）
INSERT INTO users (username, password, email, role)
VALUES (
    'admin',
    '$2a$10$N9qo8uLOickg2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWyv',
    'admin@generator.platform',
    'admin'
) ON CONFLICT (username) DO NOTHING;

-- 插入示例项目数据
INSERT INTO projects (user_id, name, description, generated_code, status)
VALUES
    (1, '示例电商系统', '一个完整的电商平台，包含商品管理、订单处理、用户系统等模块', '{"files": {}}', 'pending'),
    (1, '博客管理系统', '支持多用户博客发布、评论、标签管理等功能的CMS系统', '{"files": {}}', 'pending'),
    (1, '任务管理系统', '企业级任务跟踪和项目管理工具', '', 'pending')
ON CONFLICT DO NOTHING;

-- ============================================
-- 9. 创建更新时间触发器函数
-- ============================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为所有需要更新时间戳的表创建触发器
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_clusters_updated_at ON clusters;
CREATE TRIGGER update_clusters_updated_at BEFORE UPDATE ON clusters FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_deployments_updated_at ON deployments;
CREATE TRIGGER update_deployments_updated_at BEFORE UPDATE ON deployments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- 10. 数据库权限设置（可选）
-- ============================================
-- GRANT ALL PRIVILEGES ON DATABASE generator_platform TO app_user;
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO app_user;
-- ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO app_user;

-- ============================================
-- 完成
-- ============================================
SELECT 'Database initialization completed successfully!' AS message;
SELECT COUNT(*) AS total_tables FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE';
