# 后端错误管理改进实施计划

> 状态：Active  
> 类型：质量改进 / 可观测性提升  
> 创建时间：2026-06-12  
> 预计工期：3-4 周（分 4 个阶段）

## 背景与目标

### 当前问题

经过全面审计，CTF 后端错误管理存在以下问题：

1. **错误日志覆盖不足**：1163 个 Go 文件中只有 46 个（4%）有日志记录
2. **内部错误分类粗糙**：`ErrInternal` 被使用 484 次，用户只能看到"服务器内部错误"
3. **缺少结构化上下文**：错误日志缺少 user_id、challenge_id 等关键业务维度
4. **panic 使用不当**：`contest/ports` 用 panic 做参数校验
5. **测试覆盖薄弱**：`apperror` 包只有 1 个测试
6. **无运维可观测性**：没有错误聚合、分类、告警和监控指标

### 改进目标

- **Phase 1**：补齐关键路径错误日志，覆盖率提升到 20%+
- **Phase 2**：细化内部错误分类，减少 `ErrInternal` 使用 60%+
- **Phase 3**：建立错误监控体系，支持 Prometheus metrics
- **Phase 4**：完善测试覆盖，达到 80%+ 错误场景覆盖

## Phase 1：关键路径错误日志补齐（Week 1）

### 目标

覆盖 5 条核心业务路径，确保关键错误有结构化日志记录。

### 实施范围

#### 1.1 容器生命周期路径

**涉及模块**：`container_runtime`、`instance`

**关键错误点**：
- 容器启动失败（`16001`）
- 镜像拉取失败（`12002`、`16004`）
- 资源分配失败（`16002`）
- 健康检查失败（`16006`）

**日志规范**：
```go
logger.ErrorCtx(ctx, "容器启动失败",
    zap.Int64("instance_id", instanceID),
    zap.Int64("challenge_id", challengeID),
    zap.String("image", imageName),
    zap.String("node_id", nodeID),
    zap.Duration("elapsed", time.Since(startTime)),
    zap.Error(err))
```

**交付物**：
- [ ] `container_runtime/infrastructure/docker_executor.go` 补齐日志
- [ ] `instance/application/instance_service.go` 启动失败路径补齐日志
- [ ] 提取常量到 `container_runtime/domain/log_fields.go`

#### 1.2 Flag 提交与校验路径

**涉及模块**：`practice`、`contest`

**关键错误点**：
- Flag 错误（`13003`）
- 提交频率限制（`13004`）
- 重复提交（`13007`）

**日志规范**：
```go
logger.WarnCtx(ctx, "Flag 校验失败",
    zap.Int64("user_id", userID),
    zap.Int64("challenge_id", challengeID),
    zap.String("submission_type", "practice"), // or "contest"
    zap.Int("remaining_attempts", remainingAttempts))
```

**交付物**：
- [ ] `practice/application/commands/flag_submission_service.go` 补齐日志
- [ ] `contest/application/commands/scoring_service.go` 补齐日志

#### 1.3 认证与会话路径

**涉及模块**：`auth`、`identity`

**关键错误点**：
- 登录失败（`11001`）
- 会话失效（`11002`、`11003`、`11005`）
- 账户锁定（`11006`、`11010`）

**日志规范**：
```go
logger.WarnCtx(ctx, "登录失败",
    zap.String("username", username),
    zap.String("ip", clientIP),
    zap.Int("failed_attempts", attempts))
```

**交付物**：
- [ ] `auth/application/commands/auth_service.go` 补齐日志
- [ ] `auth/infrastructure/session_store.go` 会话失效日志

#### 1.4 数据库与缓存层

**涉及模块**：所有 `infrastructure/*_repository.go`、`*_cache.go`

**关键错误点**：
- 数据库连接失败
- 查询超时
- Redis 连接失败
- 缓存穿透

**日志规范**：
```go
logger.ErrorCtx(ctx, "数据库查询失败",
    zap.String("operation", "FindChallengeByID"),
    zap.Int64("challenge_id", id),
    zap.Duration("elapsed", time.Since(start)),
    zap.Error(err))
```

**交付物**：
- [ ] 建立 `infrastructure/logging/db_logger.go` 统一包装器
- [ ] 建立 `infrastructure/logging/cache_logger.go` 统一包装器
- [ ] 各模块 repository 接入统一日志

#### 1.5 后台任务与定时清理

**涉及模块**：`instance/infrastructure/cleaner.go`、`practice/application/commands/provisioning.go`

**关键错误点**：
- 清理任务失败
- 锁获取失败
- 对账失败

**日志规范**：
```go
logger.ErrorCtx(ctx, "实例清理任务失败",
    zap.Int("expired_count", len(expiredInstances)),
    zap.Int("cleaned", cleanedCount),
    zap.Int("failed", failedCount),
    zap.Duration("elapsed", elapsed),
    zap.Error(err))
```

**交付物**：
- [ ] 已有 cleaner 日志补充业务上下文
- [ ] provisioning loop 补齐错误日志

### 验收标准

- [ ] 5 条关键路径至少 80% 错误分支有日志
- [ ] 日志包含至少 3 个业务维度字段（user_id、challenge_id 等）
- [ ] 所有新增日志通过 `grep "logger.ErrorCtx\|logger.WarnCtx"` 可检索
- [ ] 编写 `docs/operations/error-log-guide.md` 错误日志查询手册

---

## Phase 2：内部错误细化分类（Week 2）

### 目标

将常见内部错误从 `ErrInternal` 中剥离，建立细粒度错误类型。

### 实施清单

#### 2.1 基础设施错误定义

**新增 `internal/apperror/infra_errors.go`**：

```go
var (
    // 数据库错误（10100-10199）
    ErrDatabaseConnectionFailed = Define(10100, "数据库连接失败", http.StatusServiceUnavailable)
    ErrDatabaseQueryTimeout     = Define(10101, "数据库查询超时", http.StatusGatewayTimeout)
    ErrDatabaseConstraintViolation = Define(10102, "数据约束冲突", http.StatusConflict)
    
    // 缓存错误（10200-10299）
    ErrCacheConnectionFailed = Define(10200, "缓存服务连接失败", http.StatusServiceUnavailable)
    ErrCacheTimeout          = Define(10201, "缓存操作超时", http.StatusGatewayTimeout)
    
    // 外部服务错误（10300-10399）
    ErrDockerAPIFailed       = Define(10300, "Docker API 调用失败", http.StatusBadGateway)
    ErrDockerTimeout         = Define(10301, "Docker 操作超时", http.StatusGatewayTimeout)
)
```

**交付物**：
- [ ] 创建 `internal/apperror/infra_errors.go`
- [ ] 更新 `docs/architecture/backend/01-system-architecture.md` §6.1 错误码表

#### 2.2 迁移现有 `ErrInternal` 使用

**迁移策略**：
1. 扫描所有 `ErrInternal.WithCause()` 调用
2. 按 `cause` 类型分类（`sql.ErrNoRows`、`redis.Nil`、`context.DeadlineExceeded` 等）
3. 替换为对应细分错误

**优先级队列**（按出现频率排序）：
- [ ] 数据库错误（预计 ~150 处）
- [ ] Redis 错误（预计 ~80 处）
- [ ] Docker API 错误（预计 ~60 处）
- [ ] 文件系统错误（预计 ~40 处）

**工具脚本**：
```bash
# 扫描 ErrInternal 使用
grep -rn "ErrInternal" code/backend/internal/module --include="*.go" > /tmp/errinternal_audit.txt

# 按模块分组
awk -F: '{print $1}' /tmp/errinternal_audit.txt | sort | uniq -c | sort -rn
```

**交付物**：
- [ ] 迁移完成后 `ErrInternal` 使用减少到 200 次以内
- [ ] 提交前跑 `bash scripts/check-error-classification.sh`（新建）

#### 2.3 移除 panic 用法

**目标文件**：
- `code/backend/internal/module/contest/ports/contest.go:93`
- `code/backend/internal/module/contest/ports/contest.go:119`
- `code/backend/internal/module/contest/ports/contest.go:127`

**改造前**：
```go
func (f *ContestListFilter) Validate() {
    if f.Sort == "" {
        panic("contest list sort is required")
    }
}
```

**改造后**：
```go
func (f *ContestListFilter) Validate() error {
    if f.Sort == "" {
        return apperror.ErrInvalidParams.WithMessage("contest list sort is required")
    }
    return nil
}
```

**交付物**：
- [ ] 移除所有 panic 用于参数校验的场景
- [ ] 在 `code/backend/internal/module/architecture_test.go` 中新增禁止 panic 的 guardrail

### 验收标准

- [ ] `ErrInternal` 使用减少 60%+（从 484 次降到 200 次以内）
- [ ] 新增基础设施错误码覆盖 80%+ 的常见故障
- [ ] 所有 panic 替换为 error 返回
- [ ] 错误码文档更新完整

---

## Phase 3：错误监控与可观测性（Week 3）

### 目标

建立 Prometheus metrics 体系，支持错误聚合、告警和可视化。

### 实施清单

#### 3.1 错误指标采集

**新增 `internal/observability/error_metrics.go`**：

```go
package observability

import "github.com/prometheus/client_golang/prometheus"

var (
    // 错误计数器（按错误码、模块、严重级别分组）
    ErrorCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "ctf_errors_total",
            Help: "Total number of errors by code and module",
        },
        []string{"code", "module", "severity"},
    )
    
    // 关键路径错误率（滑动窗口）
    CriticalPathErrorRate = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "ctf_critical_path_error_rate",
            Help: "Error rate for critical business paths",
        },
        []string{"path"}, // "container_start", "flag_submit", "login"
    )
)
```

**集成点**：
- `httpresponse.FromError()` 中自动记录错误
- 关键业务方法结束时记录成功/失败

**交付物**：
- [ ] 创建 `internal/observability/error_metrics.go`
- [ ] 在 `httpresponse.FromError()` 中集成 `ErrorCounter`
- [ ] 在 5 条关键路径中集成 `CriticalPathErrorRate`

#### 3.2 Grafana Dashboard

**Dashboard 结构**：
1. **总览面板**：总错误数、错误率趋势、TOP 10 错误码
2. **模块面板**：按模块分组的错误分布
3. **关键路径面板**：容器启动、Flag 提交、登录成功率
4. **告警面板**：近 5 分钟异常错误码

**交付物**：
- [ ] 创建 `deployments/monitoring/grafana-error-dashboard.json`
- [ ] 编写 `docs/operations/error-monitoring.md` 使用手册

#### 3.3 告警规则

**Prometheus 告警规则 `deployments/monitoring/error-alerts.yml`**：

```yaml
groups:
  - name: ctf_errors
    interval: 30s
    rules:
      # 容器启动失败率超过 5%
      - alert: ContainerStartFailureHigh
        expr: rate(ctf_errors_total{code="16001"}[5m]) > 0.05
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "容器启动失败率过高"
          description: "最近 5 分钟容器启动失败率 {{ $value | humanizePercentage }}"
      
      # 数据库连接失败
      - alert: DatabaseConnectionFailure
        expr: increase(ctf_errors_total{code="10100"}[1m]) > 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "数据库连接失败"
          
      # Flag 提交错误率异常
      - alert: FlagSubmissionErrorHigh
        expr: rate(ctf_errors_total{code="13003"}[5m]) > 0.5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Flag 提交错误率异常高"
```

**交付物**：
- [ ] 创建告警规则文件
- [ ] 配置 Alertmanager 通知渠道（邮件/企业微信）
- [ ] 编写 `docs/operations/error-alerting.md` 告警手册

### 验收标准

- [ ] Prometheus 能采集到至少 5 个错误指标
- [ ] Grafana 有完整的错误监控 Dashboard
- [ ] 至少配置 3 条关键告警规则
- [ ] 在测试环境触发告警并验证通知

---

## Phase 4：测试覆盖完善（Week 4）

### 目标

建立错误场景的全面测试覆盖，确保错误处理逻辑可靠。

### 实施清单

#### 4.1 `apperror` 包单元测试

**新增 `internal/apperror/error_test.go` 测试**：

```go
// 已有：TestAppErrorWithCauseSupportsErrorsIs

// 新增：
func TestAppErrorUnwrap(t *testing.T)
func TestAppErrorWithMessage(t *testing.T)
func TestAppErrorHTTPStatus(t *testing.T)
func TestAppErrorChaining(t *testing.T) // 多层 WithCause 嵌套
```

**交付物**：
- [ ] 补齐 `apperror` 包测试到 90%+ 覆盖率

#### 4.2 错误响应格式测试

**新增 `internal/httpresponse/response_test.go` 测试**：

```go
func TestFromError_AppError(t *testing.T)
func TestFromError_ValidationError(t *testing.T)
func TestFromError_UnknownError(t *testing.T)
func TestErrorEnvelopeStructure(t *testing.T)
func TestRequestIDInErrorResponse(t *testing.T)
```

**交付物**：
- [ ] 补齐 `httpresponse` 包测试到 85%+ 覆盖率

#### 4.3 模块错误集成测试

**每个模块新增 `*_error_integration_test.go`**：

测试内容：
- 业务错误正确映射到错误码
- 错误日志正确记录
- 错误响应格式符合契约

**示例（`challenge` 模块）**：
```go
func TestChallengeNotFound_ReturnsCorrectError(t *testing.T) {
    // 请求不存在的 challenge
    // 断言：返回 13004、HTTP 404、日志有记录
}

func TestImagePullFailure_LogsWithContext(t *testing.T) {
    // 模拟镜像拉取失败
    // 断言：日志包含 challenge_id、image_name、elapsed
}
```

**交付物**：
- [ ] 为 5 个核心模块各补充 5+ 错误集成测试
- [ ] 总计新增 25+ 错误场景测试

#### 4.4 错误码完整性测试

**新增 `internal/apperror/code_completeness_test.go`**：

```go
func TestAllErrorCodesDocumented(t *testing.T) {
    // 扫描所有 Define() 调用
    // 对比 docs/architecture/backend/04-api-design.md §3
    // 断言：所有错误码都有文档
}

func TestNoErrorCodeDuplicate(t *testing.T) {
    // 断言：没有重复的错误码
}

func TestErrorCodeRangeValid(t *testing.T) {
    // 断言：错误码符合模块区间规范
}
```

**交付物**：
- [ ] 创建错误码完整性测试
- [ ] 集成到 `bash scripts/check-workflow-governance.sh`

### 验收标准

- [ ] `apperror` 包测试覆盖率 ≥ 90%
- [ ] `httpresponse` 包测试覆盖率 ≥ 85%
- [ ] 新增 25+ 模块级错误集成测试
- [ ] 错误码完整性测试通过

---

## 跨阶段交付物

### 文档更新

- [ ] `docs/architecture/backend/01-system-architecture.md` §6.1 更新错误码表
- [ ] `docs/operations/error-log-guide.md` 错误日志查询手册（Phase 1）
- [ ] `docs/operations/error-monitoring.md` 错误监控使用手册（Phase 3）
- [ ] `docs/operations/error-alerting.md` 告警配置与响应手册（Phase 3）

### 工具脚本

- [ ] `bash scripts/check-error-classification.sh` 错误分类检查（Phase 2）
- [ ] `bash scripts/check-error-logs.sh` 日志覆盖率检查（Phase 1）
- [ ] `bash scripts/check-error-metrics.sh` metrics 导出检查（Phase 3）

### 护栏测试

- [ ] `internal/module/architecture_test.go` 禁止 panic 做参数校验（Phase 2）
- [ ] `internal/apperror/code_completeness_test.go` 错误码完整性（Phase 4）

---

## 风险与依赖

### 风险

1. **日志性能影响**：大量新增日志可能影响性能
   - 缓解：使用异步日志、生产环境控制日志级别
   
2. **错误码区间冲突**：新增错误码可能与未来模块冲突
   - 缓解：预留区间、在文档中明确标注

3. **告警噪音**：告警规则不当可能产生过多误报
   - 缓解：先在测试环境验证 2 周，再上生产

### 依赖

- Prometheus + Grafana 环境就绪
- 日志存储容量充足（预计增长 30%）
- 团队对 Zap 日志库和 Prometheus 有基本了解

---

## 验收总结

完成 4 个 Phase 后，系统应达到：

- ✅ **日志覆盖率**：从 4% 提升到 20%+
- ✅ **错误分类**：`ErrInternal` 使用减少 60%+
- ✅ **可观测性**：有完整的错误监控 Dashboard 和告警
- ✅ **测试覆盖**：80%+ 错误场景有测试
- ✅ **无 panic**：所有参数校验改用 error 返回
- ✅ **文档完整**：4 份运维文档 + 更新架构文档

最终产出可作为"错误管理最佳实践"在其他项目复用。
