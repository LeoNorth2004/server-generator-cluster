# Generator Platform - Test Suite

## 📋 测试概览

本项目包含完整的测试套件，用于验证代码生成平台的核心功能。

### 测试类型

#### 1. 单元测试 (`test/unit/`)

**utils_test.go** - 工具函数测试
- `toCamelCase()` - 蛇形转驼峰命名转换
- `toLowerCamelCase()` - 小驼峰命名转换
- `goTypeFromSQL()` - SQL类型到Go类型映射
- `gormTagFromField()` - GORM标签生成
- `safeSprintf()` - 安全的字符串格式化
- `extractPort()` - 端口提取
- `min()` - 最小值函数

**types_test.go** - 数据结构测试
- DBConfig JSON序列化/反序列化
- TableConfig验证逻辑
- GenerateRequest验证
- GeneratedCode结构验证
- GenerateDocsRequest验证

#### 2. 集成测试 (`test/integration/`)

**api_test.go** - API端点测试
- 健康检查端点
- 代码生成端点（正常/异常情况）
- API网关路由注册
- 中间件链执行顺序
- 错误响应格式
- 数据库连接配置验证
- 性能基准测试

---

## 🚀 运行测试

### 前置条件

1. **Go 环境**: 确保安装了 Go 1.21+
2. **依赖安装**: 在 `apps/generator-service` 目录下运行 `go mod download`
3. **数据库**: PostgreSQL（可选，某些集成测试需要）

### 运行命令

```bash
# 进入项目根目录
cd project1

# 运行所有单元测试
cd apps/generator-service
go test -v ./...

# 运行特定测试文件
go test -v -run TestToCamelCase ./...
go test -v -run TestGoTypeFromSQL ./...

# 运行基准测试
go test -bench=. -benchmem

# 运行集成测试
go test -v -run TestAPI ./test/integration/

# 查看测试覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 📊 测试用例统计

| 类别 | 文件数 | 测试函数数 | 覆盖范围 |
|------|--------|------------|----------|
| 工具函数 | 1 | 7+ | 核心工具 |
| 数据结构 | 1 | 5+ | 类型系统 |
| API端点 | 1 | 8+ | HTTP接口 |
| 性能测试 | 2 | 3 | 关键路径 |

---

## ✅ 测试检查清单

- [ ] 所有单元测试通过
- [ ] 集成测试通过
- [ ] 基准测试性能达标
- [ ] 代码覆盖率 > 80%
- [ ] 无竞态条件警告
- [ ] 内存泄漏检测通过

---

## 🔧 添加新测试

### 添加单元测试示例

```go
func TestYourFunction(t *testing.T) {
    // Arrange
    input := "your_input"
    expected := "expected_output"

    // Act
    result := yourFunction(input)

    // Assert
    if result != expected {
        t.Errorf("yourFunction(%s) = %s, want %s", input, result, expected)
    }
}
```

### 添加集成测试示例

```go
func TestYourEndpoint(t *testing.T) {
    router := setupTestRouter()
    
    // Setup route
    router.GET("/api/test", yourHandler)
    
    // Create request
    req, _ := http.NewRequest("GET", "/api/test", nil)
    w := httptest.NewRecorder()
    
    // Execute
    router.ServeHTTP(w, req)
    
    // Verify
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
}
```

---

## 📈 持续集成建议

在 CI/CD pipeline 中添加：

```yaml
test:
  stage: test
  script:
    - cd apps/generator-service
    - go test -v -race -coverprofile=coverage.out ./...
    - go tool cover -func=coverage.out
  coverage: '/total:\s*(\d+\.\d+)/'
```

---

## 🎯 测试目标

- **单元测试覆盖率**: ≥ 85%
- **集成测试通过率**: 100%
- **关键路径性能**: < 10ms per operation
- **内存使用**: 无明显泄漏

---

**最后更新**: 2026-04-22  
**维护者**: Generator Platform Team
