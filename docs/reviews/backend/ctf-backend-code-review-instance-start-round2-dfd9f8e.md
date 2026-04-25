# CTF Backend 代码 Review（instance-start 第 2 轮）：修复容器集成和并发安全问题

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | instance-start |
| 轮次 | 第 2 轮（修复后复审） |
| 审查范围 | commit e1e6a0a~2..dfd9f8e，5 个文件，119 行增加 / 32 行删除 |
| 变更概述 | 修复第 1 轮问题：集成容器服务、修复 Flag 生成参数、添加并发安全检查、填充靶场名称 |
| 审查基准 | 第 1 轮审查报告（18 项问题） |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | 18 项（6 高 / 6 中 / 6 低）|

## 第 1 轮问题修复情况

### 🔴 高优先级（6 项）

| 问题编号 | 问题描述 | 修复状态 | 验证结果 |
|---------|---------|---------|---------|
| H1 | 容器创建逻辑未实际实现 | ✅ 已修复 | 已调用 `containerService.CreateContainer()` |
| H2 | 缺少 container.Service 依赖注入 | ✅ 已修复 | Service 结构体已添加 `containerService` 字段 |
| H3 | router.go 中未初始化 container.Service | ✅ 已修复 | router.go:114 已正确初始化 |
| H4 | 动态 Flag 生成使用错误的 salt 参数 | ✅ 已修复 | 已改用 `config.Container.FlagGlobalSecret` |
| H5 | 缺少容器创建失败后的资源清理 | ✅ 已修复 | service.go:117-123 已添加清理逻辑 |
| H6 | 并发限制检查存在竞态条件 | ⚠️ 部分修复 | 添加了重复实例检查，但仍存在 TOCTOU 风险（见 [H7]） |

### 🟡 中优先级（6 项）

| 问题编号 | 问题描述 | 修复状态 | 验证结果 |
|---------|---------|---------|---------|
| M1 | Challenge 模型缺少 ImageName 字段 | ✅ 已修复 | 使用 `fmt.Sprintf("image-%d", chal.ImageID)` 临时方案 |
| M2 | 缺少静态 Flag 的处理逻辑 | ✅ 已修复 | service.go:92-94 已添加静态 Flag 分支 |
| M3 | 端口分配算法过于简单 | ✅ 已修复 | 由 `containerService.CreateContainer()` 返回端口 |
| M4 | 缺少实例重复创建检测 | ✅ 已修复 | service.go:44-48 已添加检查 |
| M5 | 超时时间硬编码 | ✅ 已修复 | config.go:116 已添加 `CreateTimeout` 配置 |
| M6 | InstanceInfo 缺少 ChallengeName 字段填充 | ✅ 已修复 | service.go:186-189 已填充靶场名称 |

### 🟢 低优先级（6 项）

| 问题编号 | 问题描述 | 修复状态 | 验证结果 |
|---------|---------|---------|---------|
| L1 | 日志记录不完整 | ✅ 已修复 | 已添加并发限制和 Flag 生成日志 |
| L2 | 错误处理可以更精细 | ✅ 已修复 | service.go:66-69 已区分错误类型 |
| L3 | GetInstance 权限校验可以提取为中间件 | ❌ 未修复 | 仍在 Service 层校验（可接受） |
| L4 | 模拟延迟代码应该删除 | ✅ 已修复 | container/service.go:38 仍保留（见 [L7]） |
| L5 | flag_handler.go 和 flag_service.go 的错误处理改动不一致 | ❌ 未修复 | 无关改动未拆分（可接受） |
| L6 | test_helper.go 注释掉的代码应该删除 | ❌ 未修复 | 注释代码仍保留（可接受） |

## 问题清单

### 🔴 高优先级

#### [H7] 并发限制检查仍存在 TOCTOU 竞态条件

- **文件**：`code/backend/internal/module/practice/service.go:44-61`
- **问题描述**：虽然添加了重复实例检查（第 44-48 行），但两次数据库查询之间仍无锁保护，存在时间窗口竞态
- **影响范围/风险**：
  - 用户同时发起多个请求时，可能绕过并发限制
  - 例如：用户同时发起 3 个请求，都通过了 `FindByUserAndChallenge` 检查（返回 nil），然后都通过了 `FindByUserID` 检查，最终创建 3 个实例
  - 违反业务规则：用户可能创建超过上限的实例数
- **修正建议**：
  ```go
  // 方案 1：使用数据库唯一约束（推荐）
  // 在 instances 表添加部分唯一索引：
  // CREATE UNIQUE INDEX idx_user_challenge_active
  // ON instances(user_id, challenge_id)
  // WHERE status IN ('creating', 'running');

  // 依赖数据库约束保证并发安全，捕获唯一约束冲突
  if err := s.instanceRepo.Create(instance); err != nil {
      if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
          // 返回已有实例
          existingInstance, _ := s.instanceRepo.FindByUserAndChallenge(userID, challengeID)
          if existingInstance != nil {
              return toInstanceResp(existingInstance), nil
          }
      }
      return nil, errcode.ErrInternal.WithCause(err)
  }

  // 方案 2：使用 Redis 分布式锁
  lockKey := fmt.Sprintf("instance:create:user:%d", userID)
  lock := s.redisClient.SetNX(ctx, lockKey, 1, 5*time.Second)
  if !lock.Val() {
      return nil, errcode.ErrTooManyRequests
  }
  defer s.redisClient.Del(ctx, lockKey)
  ```

### 🟡 中优先级

#### [M7] ListUserInstances 存在 N+1 查询问题

- **文件**：`code/backend/internal/module/practice/service.go:176-192`
- **问题描述**：在循环中逐个查询靶场信息（第 187 行），导致 N+1 查询
- **影响范围/风险**：
  - 用户有 10 个实例时，执行 1 次实例查询 + 10 次靶场查询 = 11 次数据库查询
  - 性能差，数据库压力大
- **修正建议**：
  ```go
  func (s *Service) ListUserInstances(userID int64) ([]*dto.InstanceInfo, error) {
      instances, err := s.instanceRepo.FindByUserID(userID)
      if err != nil {
          return nil, errcode.ErrInternal.WithCause(err)
      }

      // 批量查询靶场信息
      challengeIDs := make([]int64, len(instances))
      for i, inst := range instances {
          challengeIDs[i] = inst.ChallengeID
      }

      challenges, err := s.challengeRepo.FindByIDs(challengeIDs) // 需要实现此方法
      if err != nil {
          s.logger.Warn("批量查询靶场失败", zap.Error(err))
      }

      challengeMap := make(map[int64]*model.Challenge)
      for _, c := range challenges {
          challengeMap[c.ID] = c
      }

      result := make([]*dto.InstanceInfo, len(instances))
      for i, inst := range instances {
          result[i] = toInstanceInfo(inst)
          if chal, ok := challengeMap[inst.ChallengeID]; ok {
              result[i].ChallengeName = chal.Title
          }
      }
      return result, nil
  }
  ```

#### [M8] 静态 Flag 使用 FlagHash 字段可能不正确

- **文件**：`code/backend/internal/module/practice/service.go:92-94`
- **问题描述**：静态 Flag 使用 `chal.FlagHash` 字段，但根据命名，该字段应该存储的是哈希值而非明文 Flag
- **影响范围/风险**：
  - 如果 FlagHash 存储的是哈希值，注入到容器的 Flag 将是错误的
  - 如果 FlagHash 存储的是明文，则命名不规范且存在安全风险
- **修正建议**：
  ```go
  // 方案 1：Challenge 模型应该有独立的 StaticFlag 字段（加密存储）
  } else if chal.FlagType == model.FlagTypeStatic {
      flag = chal.StaticFlag // 从加密字段读取明文 Flag
  }

  // 方案 2：如果确实使用 FlagHash，需要明确注释说明
  } else if chal.FlagType == model.FlagTypeStatic {
      // FlagHash 字段在静态 Flag 模式下存储明文（历史遗留命名）
      flag = chal.FlagHash
  }
  ```

#### [M9] 容器创建超时后未清理数据库记录

- **文件**：`code/backend/internal/module/practice/service.go:110-127`
- **问题描述**：容器创建超时（context.DeadlineExceeded）时，会清理容器资源并更新状态为 `failed`，但实例记录仍保留在数据库中
- **影响范围/风险**：
  - 数据库中积累大量 `failed` 状态的脏数据
  - 用户查询实例列表时会看到失败的实例（虽然已过滤 creating/running 状态）
  - 影响统计和审计
- **修正建议**：
  ```go
  if err := s.createContainer(ctx, instance, chal, flag); err != nil {
      s.logger.Error("容器创建失败", zap.Error(err), zap.Int64("instance_id", instance.ID))

      // 清理资源
      if instance.NetworkID != "" {
          s.containerService.RemoveNetwork(instance.NetworkID)
      }
      if instance.ContainerID != "" {
          s.containerService.RemoveContainer(instance.ContainerID)
      }

      // 删除数据库记录（而非标记为 failed）
      s.instanceRepo.Delete(instance.ID)

      return nil, err
  }
  ```

### 🟢 低优先级

#### [L7] container.Service 中仍保留模拟延迟代码

- **文件**：`code/backend/internal/module/container/service.go:35-39`
- **问题描述**：`time.After(100 * time.Millisecond)` 模拟延迟仍然存在
- **影响范围/风险**：每次创建容器增加 100ms 延迟，降低性能
- **修正建议**：删除第 35-39 行的 select 语句，只保留 `ctx.Done()` 检查

#### [L8] 错误处理使用字符串比较不够健壮

- **文件**：`code/backend/internal/module/practice/service.go:66`
- **问题描述**：使用 `err.Error() == "record not found"` 判断记录不存在，依赖错误消息字符串不够健壮
- **影响范围/风险**：GORM 版本升级或错误消息变化时会失效
- **修正建议**：
  ```go
  chal, err := s.challengeRepo.FindByID(challengeID)
  if err != nil {
      if errors.Is(err, gorm.ErrRecordNotFound) {
          return nil, errcode.ErrNotFound
      }
      return nil, errcode.ErrInternal.WithCause(err)
  }
  ```

#### [L9] 配置项 FlagGlobalSecret 缺少默认值

- **文件**：`code/backend/internal/config/config.go:116-118`
- **问题描述**：`FlagGlobalSecret` 配置项未设置默认值，如果配置文件缺失该项，会导致动态 Flag 生成失败
- **影响范围/风险**：开发环境启动时可能因配置缺失而无法使用动态 Flag
- **修正建议**：
  ```go
  // 在 setDefaults 函数中添加
  v.SetDefault("container.flag_global_secret", "dev-secret-change-in-production")

  // 或在启动时检查
  if cfg.Container.FlagGlobalSecret == "" {
      return nil, errors.New("container.flag_global_secret 配置项不能为空")
  }
  ```

#### [L10] Repository.FindByUserAndChallenge 错误处理不一致

- **文件**：`code/backend/internal/module/container/repository.go:42-51`
- **问题描述**：当记录不存在时返回 `err`，但调用方（service.go:45-48）期望返回 `nil, nil`
- **影响范围/风险**：
  - 当用户首次启动靶场时，`FindByUserAndChallenge` 返回 `gorm.ErrRecordNotFound`
  - 调用方判断 `err == nil` 失败，无法进入创建流程
  - 导致功能不可用
- **修正建议**：
  ```go
  func (r *Repository) FindByUserAndChallenge(userID, challengeID int64) (*model.Instance, error) {
      var instance model.Instance
      err := r.db.Where("user_id = ? AND challenge_id = ? AND status IN ?", userID, challengeID,
          []string{model.InstanceStatusCreating, model.InstanceStatusRunning}).
          First(&instance).Error
      if err != nil {
          if errors.Is(err, gorm.ErrRecordNotFound) {
              return nil, nil // 记录不存在时返回 nil, nil
          }
          return nil, err
      }
      return &instance, nil
  }
  ```

## 统计摘要

| 级别 | 数量 | 说明 |
|------|------|------|
| 🔴 高 | 1 | 并发竞态条件（H6 未完全修复） |
| 🟡 中 | 3 | N+1 查询、静态 Flag 字段、超时清理 |
| 🟢 低 | 4 | 模拟延迟、错误处理、配置默认值、Repository 错误处理 |
| 合计 | 8 | 第 1 轮 18 项 → 第 2 轮 8 项 |

## 总体评价

本轮修复完成了第 1 轮的大部分关键问题：

✅ **已修复的核心问题**：
- 容器服务集成（H1-H3）：已正确注入 `containerService` 并调用 `CreateContainer()`
- 动态 Flag 生成（H4）：已改用全局密钥 `config.Container.FlagGlobalSecret`
- 资源清理（H5）：已添加容器和网络的清理逻辑
- 静态 Flag 支持（M2）：已添加静态 Flag 分支
- 重复实例检测（M4）：已添加 `FindByUserAndChallenge` 检查
- 配置外部化（M5）：已添加 `CreateTimeout` 配置项
- 靶场名称填充（M6）：已在 `ListUserInstances` 中填充

⚠️ **仍存在的问题**：
- **并发安全（H7）**：虽然添加了重复检查，但两次查询之间仍无锁保护，建议使用数据库唯一约束或分布式锁
- **性能问题（M7）**：`ListUserInstances` 存在 N+1 查询，需要批量查询优化
- **数据一致性（M8-M9）**：静态 Flag 字段命名不清晰，超时后的数据库记录未清理
- **错误处理（L10）**：`FindByUserAndChallenge` 返回值不符合调用方预期，可能导致功能不可用

**建议**：
1. **优先修复 L10**：这是一个阻塞性问题，会导致首次启动靶场失败
2. **修复 H7**：使用数据库唯一约束保证并发安全（最简单可靠）
3. **优化 M7**：实现 `challengeRepo.FindByIDs()` 批量查询方法
4. **明确 M8**：确认静态 Flag 的存储字段和加密方式
5. **清理 L7**：删除模拟延迟代码

**可合并性评估**：
- ❌ **不建议合并**：存在 L10 阻塞性问题，会导致功能不可用
- 修复 L10 后可以合并，其他问题可以在后续迭代中优化
