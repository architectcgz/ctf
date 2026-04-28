# CTF Platform 代码 Review（instance-lifecycle 第 1 轮）：容器生命周期管理与定时清理

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | instance-lifecycle |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | commit 99b3025，9 个文件，414 行新增 |
| 变更概述 | 实现容器实例生命周期管理（创建、销毁、延时）与定时清理任务 |
| 审查基准 | `docs/tasks/backend-task-breakdown.md` B12 任务定义、`CLAUDE.md` 项目规范 |
| 审查日期 | 2026-03-05 |
| 上轮问题数 | - |

## 问题清单

### 🔴 高优先级

#### [H1] 硬编码：实例并发数限制未配置化
- **文件**：`code/backend/internal/module/container/service.go:33`
- **问题描述**：用户并发实例数限制硬编码为 `3`，违反"禁止硬编码"规范
- **影响范围/风险**：无法根据环境或用户等级动态调整限制，运营灵活性差
- **修正建议**：
```go
// 在配置文件中定义
type InstanceConfig struct {
    MaxConcurrentPerUser int `mapstructure:"max_concurrent_per_user"`
    DefaultTTL           time.Duration `mapstructure:"default_ttl"`
    ExtendDuration       time.Duration `mapstructure:"extend_duration"`
    MaxExtends           int `mapstructure:"max_extends"`
    CleanupInterval      string `mapstructure:"cleanup_interval"`
}

// Service 中注入配置
func (s *Service) CreateInstance(userID, challengeID int64) (*dto.InstanceResp, error) {
    instances, err := s.repo.FindByUserID(userID)
    if err != nil {
        return nil, errcode.ErrInternal.WithCause(err)
    }
    if len(instances) >= s.config.MaxConcurrentPerUser {
        return nil, errcode.ErrInstanceLimitExceeded
    }
    // ...
}
```

#### [H2] 硬编码：实例过期时间未配置化
- **文件**：`code/backend/internal/module/container/service.go:43`
- **问题描述**：实例过期时间硬编码为 `2 * time.Hour`
- **影响范围/风险**：无法根据题目难度或比赛阶段调整时长
- **修正建议**：使用 H1 中的 `config.DefaultTTL`

#### [H3] 硬编码：延时时长未配置化
- **文件**：`code/backend/internal/module/container/service.go:87`
- **问题描述**：延时时长硬编码为 `1 * time.Hour`
- **影响范围/风险**：无法灵活调整延时策略
- **修正建议**：使用 H1 中的 `config.ExtendDuration`

#### [H4] 硬编码：最大延时次数未配置化
- **文件**：`code/backend/internal/module/container/service.go:44`
- **问题描述**：`MaxExtends` 硬编码为 `2`
- **影响范围/风险**：无法根据比赛规则动态调整
- **修正建议**：使用 H1 中的 `config.MaxExtends`

#### [H5] 硬编码：清理间隔未配置化
- **文件**：`code/backend/internal/module/container/cleaner.go:25`
- **问题描述**：定时清理间隔硬编码为 `"*/5 * * * *"`（5 分钟）
- **影响范围/风险**：无法根据系统负载调整清理频率
- **修正建议**：
```go
func (c *Cleaner) Start(interval string) error {
    _, err := c.cron.AddFunc(interval, func() {
        // ...
    })
    // ...
}
```

#### [H6] 资源泄漏：清理失败时容器未实际删除
- **文件**：`code/backend/internal/module/container/service.go:104-116`
- **问题描述**：`CleanExpiredInstances` 只更新数据库状态，未实际调用 Docker API 停止/删除容器和网络
- **影响范围/风险**：过期实例的容器和网络资源持续占用系统资源，导致资源耗尽
- **修正建议**：
```go
func (s *Service) CleanExpiredInstances(ctx context.Context) error {
    instances, err := s.repo.FindExpired()
    if err != nil {
        return err
    }

    for _, inst := range instances {
        s.logger.Info("清理过期实例", zap.Int64("instance_id", inst.ID))

        // 1. 停止并删除容器
        if err := s.dockerClient.ContainerStop(ctx, inst.ContainerID, nil); err != nil {
            s.logger.Error("停止容器失败", zap.Error(err))
        }
        if err := s.dockerClient.ContainerRemove(ctx, inst.ContainerID, types.ContainerRemoveOptions{Force: true}); err != nil {
            s.logger.Error("删除容器失败", zap.Error(err))
        }

        // 2. 删除网络
        if inst.NetworkID != "" {
            if err := s.dockerClient.NetworkRemove(ctx, inst.NetworkID); err != nil {
                s.logger.Error("删除网络失败", zap.Error(err))
            }
        }

        // 3. 更新状态
        s.repo.UpdateStatus(inst.ID, model.InstanceStatusExpired)
    }
    return nil
}
```

#### [H7] 并发安全：Repository 更新操作缺少事务保护
- **文件**：`code/backend/internal/module/container/service.go:72-88`
- **问题描述**：`ExtendInstance` 中先查询再更新，存在并发竞争窗口（TOCTOU 问题）
- **影响范围/风险**：多个请求同时延时可能绕过 `MaxExtends` 限制
- **修正建议**：
```go
// Repository 层添加原子更新方法
func (r *Repository) AtomicExtend(id int64, userID int64, maxExtends int, duration time.Duration) error {
    result := r.db.Model(&model.Instance{}).
        Where("id = ? AND user_id = ? AND status = ? AND extend_count < ?",
            id, userID, model.InstanceStatusRunning, maxExtends).
        Updates(map[string]interface{}{
            "expires_at":   gorm.Expr("expires_at + ?", duration),
            "extend_count": gorm.Expr("extend_count + 1"),
        })
    if result.RowsAffected == 0 {
        return errcode.ErrExtendLimitExceeded
    }
    return result.Error
}

// Service 层调用
func (s *Service) ExtendInstance(instanceID, userID int64) error {
    return s.repo.AtomicExtend(instanceID, userID, s.config.MaxExtends, s.config.ExtendDuration)
}
```

### 🟡 中优先级

#### [M1] 错误处理：DestroyInstance 未实际清理容器资源
- **文件**：`code/backend/internal/module/container/service.go:59-70`
- **问题描述**：只更新数据库状态为 `stopped`，未调用 Docker API 停止容器
- **影响范围/风险**：容器继续运行占用资源，与数据库状态不一致
- **修正建议**：参考 H6，添加实际的容器停止和删除逻辑

#### [M2] 数据一致性：CreateInstance 状态更新未检查错误
- **文件**：`code/backend/internal/module/container/service.go:54`
- **问题描述**：`UpdateStatus` 返回的错误未处理，可能导致容器已创建但状态仍为 `creating`
- **影响范围/风险**：状态不一致，用户无法访问已创建的实例
- **修正建议**：
```go
if err := s.repo.UpdateStatus(instance.ID, model.InstanceStatusRunning); err != nil {
    s.logger.Error("更新实例状态失败", zap.Error(err))
    // 考虑回滚容器创建
    return nil, errcode.ErrInternal.WithCause(err)
}
```

#### [M3] 可观测性：缺少关键操作的日志记录
- **文件**：`code/backend/internal/module/container/service.go`
- **问题描述**：创建、销毁、延时实例时缺少结构化日志
- **影响范围/风险**：问题排查困难，无法追踪实例生命周期
- **修正建议**：
```go
s.logger.Info("创建实例",
    zap.Int64("user_id", userID),
    zap.Int64("challenge_id", challengeID),
    zap.Int64("instance_id", instance.ID),
    zap.Time("expires_at", instance.ExpiresAt))

s.logger.Info("销毁实例",
    zap.Int64("instance_id", instanceID),
    zap.Int64("user_id", userID))

s.logger.Info("延时实例",
    zap.Int64("instance_id", instanceID),
    zap.Int("extend_count", instance.ExtendCount+1),
    zap.Time("new_expires_at", newExpiresAt))
```

#### [M4] 性能：FindByUserID 查询未限制返回数量
- **文件**：`code/backend/internal/module/container/repository.go:32-39`
- **问题描述**：查询用户所有实例时未分页，可能返回大量历史记录
- **影响范围/风险**：用户实例数量增长后查询变慢，内存占用增加
- **修正建议**：
```go
func (r *Repository) FindByUserID(userID int64) ([]*model.Instance, error) {
    var instances []*model.Instance
    err := r.db.Where("user_id = ? AND status IN ?", userID,
        []string{model.InstanceStatusCreating, model.InstanceStatusRunning}).
        Order("created_at DESC").
        Limit(100). // 添加限制
        Find(&instances).Error
    return instances, err
}
```

#### [M5] 架构一致性：Service 直接依赖 Repository 而非接口
- **文件**：`code/backend/internal/module/container/service.go:15-17`
- **问题描述**：Service 依赖具体的 `*Repository` 类型，不利于测试和扩展
- **影响范围/风险**：单元测试需要真实数据库，无法使用 mock
- **修正建议**：
```go
type InstanceRepository interface {
    Create(instance *model.Instance) error
    FindByID(id int64) (*model.Instance, error)
    FindByUserID(userID int64) ([]*model.Instance, error)
    UpdateStatus(id int64, status string) error
    FindExpired() ([]*model.Instance, error)
    UpdateExtend(id int64, expiresAt time.Time, extendCount int) error
}

type Service struct {
    repo   InstanceRepository
    logger *zap.Logger
}
```

#### [M6] 数据库设计：缺少复合索引优化查询
- **文件**：`code/backend/migrations/000002_create_instances_table.up.sql:16-19`
- **问题描述**：`FindByUserID` 查询 `user_id + status`，但只有单列索引
- **影响范围/风险**：查询效率低，随着数据量增长性能下降
- **修正建议**：
```sql
CREATE INDEX IF NOT EXISTS idx_instances_user_status ON instances(user_id, status);
```

#### [M7] 错误处理：Repository 错误未区分 NotFound 和其他错误
- **文件**：`code/backend/internal/module/container/repository.go:23-30`
- **问题描述**：`FindByID` 返回的 `gorm.ErrRecordNotFound` 未转换为业务错误
- **影响范围/风险**：Service 层需要判断 GORM 错误类型，违反分层原则
- **修正建议**：
```go
func (r *Repository) FindByID(id int64) (*model.Instance, error) {
    var instance model.Instance
    err := r.db.Where("id = ?", id).First(&instance).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errcode.ErrInstanceNotFound
        }
        return nil, err
    }
    return &instance, nil
}
```

### 🟢 低优先级

#### [L1] 代码质量：模拟容器创建的临时代码未标注 TODO
- **文件**：`code/backend/internal/module/container/service.go:41-42`
- **问题描述**：`ContainerID` 使用临时生成逻辑，但注释不够明显
- **影响范围/风险**：后续集成时可能遗漏替换
- **修正建议**：
```go
// TODO(B9-B11): 替换为实际的 Docker 容器创建逻辑
instance := &model.Instance{
    UserID:      userID,
    ChallengeID: challengeID,
    ContainerID: fmt.Sprintf("temp-container-%d-%d", userID, time.Now().Unix()),
    Status:      model.InstanceStatusCreating,
    ExpiresAt:   time.Now().Add(s.config.DefaultTTL),
    MaxExtends:  s.config.MaxExtends,
}
```

#### [L2] 代码质量：AccessURL 生成逻辑过于简单
- **文件**：`code/backend/internal/module/container/service.go:53`
- **问题描述**：端口号计算 `3000 + instance.ID` 可能与实际端口分配冲突
- **影响范围/风险**：与 B11 端口管理模块集成时需要重构
- **修正建议**：标注 TODO 并说明需要对接端口管理模块

#### [L3] 命名规范：DTO 字段 `ExtendCount` 语义不清晰
- **文件**：`code/backend/internal/dto/instance.go:11`
- **问题描述**：`ExtendCount` 可能被误解为"可延时次数"而非"已延时次数"
- **影响范围/风险**：前端开发者可能误用字段
- **修正建议**：
```go
type InstanceResp struct {
    // ...
    ExtendedCount int `json:"extended_count"` // 已延时次数
    MaxExtends    int `json:"max_extends"`    // 最大延时次数
}
```

#### [L4] 可维护性：状态常量未集中管理
- **文件**：`code/backend/internal/model/instance.go:22-28`
- **问题描述**：状态常量定义在 Model 包中，但 Service 层也需要使用
- **影响范围/风险**：状态管理分散，不利于状态机扩展
- **修正建议**：考虑将状态常量提取到独立的 `constants` 包或保持现状（影响较小）

#### [L5] 测试覆盖：缺少单元测试
- **文件**：整个 `container` 模块
- **问题描述**：未提供任何单元测试文件
- **影响范围/风险**：代码质量无法保证，重构风险高
- **修正建议**：至少为 Service 层核心方法添加测试（可在后续任务中补充）

#### [L6] 文档：缺少 API 接口文档
- **文件**：`code/backend/internal/module/container/handler.go`
- **问题描述**：Handler 方法缺少 Swagger 注释
- **影响范围/风险**：前端对接时需要查看代码
- **修正建议**：
```go
// CreateInstance godoc
// @Summary 创建容器实例
// @Tags 容器管理
// @Accept json
// @Produce json
// @Param id path int true "题目 ID"
// @Success 200 {object} response.Response{data=dto.InstanceResp}
// @Router /api/v1/challenges/{id}/instances [post]
func (h *Handler) CreateInstance(c *gin.Context) {
    // ...
}
```

#### [L7] 安全性：Handler 层未校验 instanceID 和 challengeID 范围
- **文件**：`code/backend/internal/module/container/handler.go:21-25, 38-42`
- **问题描述**：`ParseInt` 后未检查 ID 是否为正数
- **影响范围/风险**：负数或零可能导致异常查询
- **修正建议**：
```go
challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
if err != nil || challengeID <= 0 {
    response.ValidationError(c, errors.New("无效的题目 ID"))
    return
}
```

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 7 |
| 🟡 中 | 7 |
| 🟢 低 | 7 |
| 合计 | 21 |

## 总体评价

**架构一致性**：✅ 良好
- 严格遵循 Repository → Service → Handler 三层架构
- Model 和 DTO 正确分离，未出现敏感字段泄漏
- 错误处理使用统一的 `errcode` 包

**核心问题**：
1. **硬编码严重**（H1-H5）：所有运行时参数（并发数、TTL、延时时长、清理间隔）均未配置化，违反项目规范
2. **资源泄漏风险**（H6, M1）：清理和销毁逻辑只更新数据库，未实际释放 Docker 资源
3. **并发安全缺陷**（H7）：延时操作存在 TOCTOU 竞争条件

**优点**：
- 数据库迁移脚本完整，索引设计基本合理
- 状态流转逻辑清晰，状态常量定义规范
- 定时清理任务结构合理，使用成熟的 cron 库

**修复优先级**：
1. 第一轮必须修复所有高优先级问题（H1-H7）
2. 中优先级问题（M1-M7）建议在本轮一并修复，避免后续返工
3. 低优先级问题（L1-L7）可在后续迭代中优化

**风险提示**：
- H6 和 M1 如不修复，系统运行一段时间后将出现严重的资源泄漏
- H7 如不修复，高并发场景下可能绕过延时次数限制

