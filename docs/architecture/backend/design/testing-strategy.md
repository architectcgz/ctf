# 后端测试架构与策略

> 状态：Current
> 事实源：`code/backend/tests/README.md`、`code/backend/internal/testutil/`、`code/backend/tests/`
> 替代：无

## 本文档范围

| 覆盖 | 不覆盖 |
|------|--------|
| 后端测试层级划分与职责边界 | 具体模块的业务测试用例 |
| Mock 策略与测试工具选型 | 前端测试策略（见前端文档） |
| 测试覆盖率目标与优先级 | CI/CD 配置细节 |
| 测试文件放置规则 | 生产环境监控策略 |

## 定位

本文档只说明后端测试的架构分层、测试类型、Mock 策略、覆盖率目标和测试文件的放置规则。

## 当前设计

- `code/backend/tests/README.md`
  - 负责：测试目录说明、放置判断、层级划分和迁移原则
  - 不负责：具体测试用例编写细节

- `code/backend/internal/testutil/`
  - 负责：需要访问包内未导出实现的测试工具和 helper
  - 不负责：业务测试用例本身

- `code/backend/tests/`
  - 负责：跨模块、系统级、运行时集成测试
  - 不负责：单个包内的单元测试

## 1. 测试层级划分

当前后端测试按职责分为四个层级：

### 1.1 单元测试（Unit Tests）

**位置**：`internal/<module>/*_test.go`

**职责**：

- 测试单个函数、方法、结构体的行为
- 验证业务逻辑、输入验证、边界条件
- 快速执行，无外部依赖

**示例**：

```go
// internal/module/challenges/domain/challenge_test.go
func TestChallenge_Validate(t *testing.T) {
    c := &Challenge{Title: ""}
    err := c.Validate()
    assert.Error(t, err, "空标题应该返回错误")
}
```

**特征**：

- 不需要数据库、HTTP 服务器、外部进程
- 执行时间 < 100ms
- 可以访问包私有符号

### 1.2 集成测试（Integration Tests）

**位置**：`internal/<module>/*_test.go` 或 `tests/system/http/`

**职责**：

- 测试多个模块协作
- 验证 Repository 与数据库交互
- 验证 HTTP Handler 与 Application Service 协作

**示例**：

```go
// tests/system/http/challenges/list_test.go
func TestListChallenges(t *testing.T) {
    app := setupTestApp(t)
    defer app.Cleanup()
    
    resp := app.GET("/api/v1/challenges")
    assert.Equal(t, 200, resp.StatusCode)
}
```

**特征**：

- 需要测试数据库（通常是内存数据库或 Docker 容器）
- 执行时间 100ms - 1s
- 验证跨模块契约

### 1.3 系统测试（System Tests）

**位置**：`tests/system/http/` 或 `tests/runtime/`

**职责**：

- 端到端黑盒测试
- 验证完整业务流程
- 测试权限、路由、序列化

**示例**：

```go
// tests/system/http/practiceflow/submit_test.go
func TestPracticeFlowE2E(t *testing.T) {
    // 1. 创建比赛
    // 2. 发布题目
    // 3. 学生提交 flag
    // 4. 验证得分
}
```

**特征**：

- 需要完整 HTTP Router
- 执行时间 1s - 10s
- 只通过 HTTP API 交互

### 1.4 架构测试（Architecture Tests）

**位置**：`tests/architecture/`

**职责**：

- 验证架构约束
- 检查依赖方向
- 验证目录结构

**示例**：

```go
// tests/architecture/module_boundaries_test.go
func TestModuleBoundaries(t *testing.T) {
    // 验证 domain 层不依赖 infrastructure
    // 验证 application 层不直接依赖 handler
}
```

**特征**：

- 不运行业务逻辑
- 通过静态分析或源码扫描
- 执行时间 < 1s

## 2. Mock 策略

### 2.1 Mock 使用原则

- **优先使用真实实现**：能用真实数据库就不用 Mock
- **在边界处 Mock**：Mock 外部依赖（第三方 API、消息队列），不 Mock 内部模块
- **使用接口而非具体实现**：`Repository` 定义为接口，测试时可以注入 `FakeRepository`

### 2.2 Mock 工具选型

| 场景 | 工具 | 示例 |
|------|------|------|
| 数据库 | 真实 PostgreSQL（Docker）或 `sqlmock` | `tests/runtime/` 使用真实 PostgreSQL |
| HTTP 客户端 | `httptest.Server` | 测试 runtime agent 调用 |
| 时间 | `TimeProvider` 接口 | 注入 `FakeTimeProvider` |
| Repository | 手写 Fake 实现 | `FakeUserRepository` |

### 2.3 不推荐的 Mock 方式

- **过度 Mock**：Mock 所有依赖，导致测试只验证 Mock 行为
- **Mock 包私有函数**：使用反射或 monkey patch 修改私有函数
- **Mock 标准库**：Mock `time.Now()`、`os.Open()` 等（应通过接口注入）

## 3. 测试覆盖率目标

### 3.1 覆盖率目标

| 模块类型 | 目标覆盖率 | 说明 |
|---------|-----------|------|
| Domain 层 | 80%+ | 核心业务逻辑必须充分测试 |
| Application 层 | 70%+ | 用例编排和状态转换 |
| Repository 层 | 60%+ | 数据访问逻辑 |
| Handler 层 | 50%+ | 通过系统测试覆盖 |
| Infrastructure 层 | 30%+ | 外部依赖集成，优先通过集成测试 |

### 3.2 覆盖率非唯一指标

高覆盖率不等于高质量测试：

- **关键路径优先**：核心业务流程必须有测试
- **边界条件优先**：空输入、非法输入、并发冲突
- **回归测试优先**：每个 bug 修复后必须有对应测试

### 3.3 不追求 100% 覆盖率

以下代码可以不测试：

- 简单的 getter/setter
- 纯粹的数据结构（无业务逻辑）
- 生成的代码（protobuf、swagger）
- 明显的样板代码

## 4. 测试文件放置规则

完整规则见 `code/backend/tests/README.md`，此处摘要关键判断：

### 4.1 包内测试（`*_test.go`）

**适用场景**：

- 需要访问包私有符号
- 模块内业务语义测试
- Application Service / Command / Query / Repository 的局部契约

**位置**：与源码同目录

### 4.2 跨目录测试（`tests/`）

**适用场景**：

- 跨模块协作测试
- 黑盒 HTTP 系统测试
- 需要 PostgreSQL / runtime agent / 容器环境的集成测试

**位置**：

- `tests/architecture/`：架构守卫
- `tests/system/http/`：HTTP 黑盒测试
- `tests/runtime/`：运行时集成测试
- `tests/testkit/`：跨测试复用的 helper / fixture

### 4.3 判断流程

```
是否需要访问包私有符号？
  ├─ 是 → 放在包内 `*_test.go`
  └─ 否 → 是否跨模块？
      ├─ 是 → 放在 `tests/system/http/` 或 `tests/runtime/`
      └─ 否 → 是否需要真实 PostgreSQL / runtime agent？
          ├─ 是 → 放在 `tests/runtime/`
          └─ 否 → 放在包内 `*_test.go`
```

## 5. 测试工具与 Fixture

### 5.1 测试工具位置

| 工具类型 | 位置 | 说明 |
|---------|------|------|
| 需要访问私有实现 | `internal/testutil/` | 例如 `systemapp` 测试环境 |
| 跨测试复用 | `tests/testkit/` | 场景 builder、fixture、assert helper |
| 只服务单个测试 | 测试文件内 | 不单独抽出 |

### 5.2 测试数据工厂

优先使用 **Builder 模式** 而非大量 fixture 文件：

```go
// tests/testkit/builders/challenge_builder.go
type ChallengeBuilder struct {
    title string
    score int
}

func NewChallenge() *ChallengeBuilder {
    return &ChallengeBuilder{
        title: "Test Challenge",
        score: 100,
    }
}

func (b *ChallengeBuilder) WithTitle(title string) *ChallengeBuilder {
    b.title = title
    return b
}

func (b *ChallengeBuilder) Build() *Challenge {
    return &Challenge{Title: b.title, Score: b.score}
}
```

### 5.3 断言 Helper

优先使用 `testify/assert` 和 `testify/require`：

```go
assert.Equal(t, expected, actual, "应该返回正确的值")
require.NoError(t, err, "不应该返回错误")
```

## 6. TDD 与测试生命周期

### 6.1 TDD 流程

本项目鼓励但不强制 TDD：

1. **Red**：先写失败的测试
2. **Green**：写最少代码让测试通过
3. **Refactor**：重构代码，保持测试通过

### 6.2 测试保留原则

TDD 产出的测试默认是交付物和回归护栏，不因为功能开发完成而删除。

**删除测试的条件**：

- 同一行为信号已由更清晰的测试覆盖
- 测试锁定了实现细节（测试 `how` 而非 `what`）
- 迁移 guard 到期且 owner / 移除条件满足
- 行为本身被明确废弃

### 6.3 测试重复处理

当测试文件过多时，优先按行为 owner 和测试层级治理：

- 模块语义留在对应模块
- 黑盒 HTTP 场景放 `tests/system/http`
- Runtime / PostgreSQL / 容器协作放 `tests/runtime`
- 重复 setup 再抽 builder、fixture、assert helper

## 7. 边界

### 7.1 本文档不覆盖

- **具体模块的测试用例**：见各模块文档
- **CI/CD 配置**：见 `.github/workflows/` 或运维文档
- **性能测试**：暂未系统化
- **前端测试**：见 `docs/architecture/frontend/testing-strategy.md`

### 7.2 与其他文档的关系

- 测试目录说明：见 `code/backend/tests/README.md`
- 模块设计：见 `docs/architecture/backend/modules/`
- 前端测试：见 `docs/architecture/frontend/testing-strategy.md`

## 8. Guardrail

- 新增功能必须有对应测试（单元或集成）
- 修复 bug 必须先写回归测试
- 测试不得访问生产数据库
- 系统测试只通过 HTTP API 交互，不直接调用内部模块
- 架构测试失败必须修复，不得修改测试让它通过

## 9. 已知限制

- 当前没有自动化的覆盖率报告
- 测试执行时间较长（完整测试套件 > 30s）
- 部分旧测试仍在 `internal/app` 未迁移
- 缺少性能测试和压力测试
