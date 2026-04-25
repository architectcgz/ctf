# CTF 后端代码 Review（skill-profile 第 2 轮）：能力画像功能修复验证

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | skill-profile |
| 轮次 | 第 2 轮（修复后复审） |
| 审查范围 | 5dbba98..09861a7（5 个提交，涉及 DTO、Service、Repository、Migration） |
| 变更概述 | 修复 round 1 发现的高优先级问题（H1-H5）和部分中低优先级问题 |
| 审查基准 | docs/reviews/backend/ctf-backend-code-review-skill-profile-round1-6b9dd04.md |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | 14（高5 + 中4 + 低5） |

## 问题清单

### 🔴 高优先级

#### [H6] 性能问题：全量重算未解决，H3 问题仍然存在
- **文件**：`code/backend/internal/module/assessment/service.go:37-96`
- **问题描述**：
  虽然添加了分布式锁和超时控制，但 `CalculateSkillProfileWithContext` 方法仍然是全量重算所有维度：
  ```go
  scores, err := s.repo.GetDimensionScores(userID)  // 每次都扫描用户所有提交记录
  ```
  问题：
  1. 用户提交 100 道题后，每次新提交都要重新聚合 100+ 条记录
  2. 即使只提交了 Web 类题目，也会重算 Pwn、Reverse 等所有维度
  3. 分布式锁虽然避免了并发重复计算，但单次计算的性能问题未解决

- **影响范围/风险**：
  - 数据库负载高，随着用户提交数增长性能持续下降
  - 5 秒超时可能不够（用户提交 500+ 道题时）

- **修正建议**：
  ```go
  // 方案 1：增量更新（推荐）
  func (s *Service) IncrementSkillProfile(userID int64, dimension string, points int) error {
      // 1. 查询该维度总分（可缓存）
      totalScore, err := s.getTotalScoreByDimension(dimension)
      if err != nil {
          return err
      }

      // 2. 原子更新用户该维度得分
      return s.repo.IncrementUserScore(userID, dimension, points, totalScore)
  }

  // Repository 层添加增量更新方法
  func (r *Repository) IncrementUserScore(userID int64, dimension string, points int, totalScore int) error {
      // 使用 SQL 原子更新
      return r.db.Exec(`
          INSERT INTO skill_profiles (user_id, dimension, score, updated_at)
          VALUES (?, ?, ? / ?, NOW())
          ON DUPLICATE KEY UPDATE
              score = (
                  SELECT COALESCE(SUM(c.points), 0) / ?
                  FROM challenges c
                  JOIN submissions s ON c.id = s.challenge_id
                  WHERE c.category = ? AND s.user_id = ? AND s.is_correct = 1
              ),
              updated_at = NOW()
      `, userID, dimension, points, totalScore, totalScore, dimension, userID).Error
  }
  ```

#### [H7] 并发安全：分布式锁失败时仍继续执行，未真正保护临界区
- **文件**：`code/backend/internal/module/assessment/service.go:43-52`
- **问题描述**：
  ```go
  locked, err := s.redis.SetNX(ctx, lockKey, 1, 10*time.Second).Result()
  if err != nil {
      s.logger.Warn("获取分布式锁失败", zap.Int64("userID", userID), zap.Error(err))
  } else if !locked {
      s.logger.Debug("画像正在计算中，跳过", zap.Int64("userID", userID))
      return s.getExistingProfile(userID)
  }
  defer s.redis.Del(ctx, lockKey)
  ```
  问题：
  1. 当 Redis 连接失败（`err != nil`）时，只记录日志但继续执行，**没有获取锁却进入了临界区**
  2. 这会导致多个 goroutine 同时计算画像，违背了加锁的初衷
  3. 应该在获取锁失败时直接返回错误或返回已有画像

- **影响范围/风险**：
  - Redis 故障时并发安全失效
  - 可能产生数据竞争和重复计算

- **修正建议**：
  ```go
  locked, err := s.redis.SetNX(ctx, lockKey, 1, 10*time.Second).Result()
  if err != nil {
      // Redis 故障时返回已有画像，避免进入临界区
      s.logger.Warn("获取分布式锁失败，返回已有画像", zap.Int64("userID", userID), zap.Error(err))
      return s.getExistingProfile(userID)
  }
  if !locked {
      s.logger.Debug("画像正在计算中，跳过", zap.Int64("userID", userID))
      return s.getExistingProfile(userID)
  }
  defer s.redis.Del(ctx, lockKey)
  ```

### 🟡 中优先级

#### [M5] 数据一致性：延迟触发未实现，H2 问题未完全解决
- **文件**：`code/backend/internal/module/practice/service.go:138-154`
- **问题描述**：
  虽然添加了超时控制和 panic 恢复，但仍然是在 Flag 提交成功后**立即**异步计算画像：
  ```go
  if isCorrect {
      s.logger.Info("Flag验证成功", ...)
      if s.assessmentService != nil {
          go func() { ... }()  // 立即触发
      }
  }
  ```
  问题：
  1. 如果外层有事务（虽然当前代码没有），submission 记录可能还未提交
  2. 虽然分布式锁能避免并发重复计算，但无法保证读到最新数据
  3. Round 1 建议的延迟触发（`time.AfterFunc(100*time.Millisecond, ...)`）未采纳

- **影响范围/风险**：
  - 理论上可能读到脏数据（虽然当前代码风险较低）
  - 未来如果添加事务会出现问题

- **修正建议**：
  ```go
  if isCorrect {
      s.logger.Info("Flag验证成功", ...)
      if s.assessmentService != nil {
          // 延迟 100ms 触发，确保数据已提交
          time.AfterFunc(100*time.Millisecond, func() {
              defer func() {
                  if r := recover(); r != nil {
                      s.logger.Error("画像更新 panic", zap.Int64("userID", userID), zap.Any("panic", r))
                  }
              }()

              ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
              defer cancel()

              if _, err := s.assessmentService.CalculateSkillProfileWithContext(ctx, userID); err != nil {
                  s.logger.Error("更新能力画像失败", zap.Int64("userID", userID), zap.Error(err))
              }
          })
      }
  }
  ```

#### [M6] 性能问题：challenges 表缓存未实现，M4 问题未解决
- **文件**：`code/backend/internal/module/assessment/repository.go:52-65`
- **问题描述**：
  SQL 查询中 `FROM challenges c WHERE c.status = 'published'` 每次都扫描 challenges 表：
  ```go
  SELECT
      c.category as dimension,
      SUM(c.points) as total_score,
      ...
  FROM challenges c
  WHERE c.status = 'published'
  GROUP BY c.category
  ```
  问题：
  1. 已发布题目列表变化频率很低，但每次计算都查询
  2. 各维度总分是相对固定的，应该缓存
  3. Round 1 建议的 Redis 缓存未实现

- **影响范围/风险**：
  - 数据库负载高
  - 响应时间慢

- **修正建议**：
  ```go
  // Service 层添加缓存方法
  func (s *Service) getTotalScoreByDimension(ctx context.Context, dimension string) (int, error) {
      cacheKey := fmt.Sprintf("skill_profile:total_score:%s", dimension)

      // 尝试从缓存读取
      if val, err := s.redis.Get(ctx, cacheKey).Int(); err == nil {
          return val, nil
      }

      // 缓存未命中，查询数据库
      var total int
      err := s.db.Model(&model.Challenge{}).
          Where("status = ? AND category = ?", "published", dimension).
          Select("COALESCE(SUM(points), 0)").
          Scan(&total).Error

      if err != nil {
          return 0, err
      }

      // 写入缓存（TTL 1小时）
      s.redis.Set(ctx, cacheKey, total, time.Hour)
      return total, nil
  }
  ```

### 🟢 低优先级

#### [L6] 代码风格：DimensionScore 已提取但位置不当
- **文件**：`code/backend/internal/module/assessment/repository.go:44-49`
- **问题描述**：
  `DimensionScore` 结构体已从函数内部提取，但放在了 Repository 文件中：
  ```go
  // DimensionScore 维度得分统计
  type DimensionScore struct {
      Dimension  string
      TotalScore int
      UserScore  int
  }
  ```
  问题：
  1. 这是一个内部数据传输结构，不是持久化模型
  2. 更适合放在 Service 文件或单独的 `types.go` 中
  3. Repository 文件应该只包含数据访问逻辑

- **影响范围/风险**：
  - 代码组织略显混乱
  - 可读性略差

- **修正建议**：
  ```go
  // 创建 internal/module/assessment/types.go
  package assessment

  // DimensionScore 维度得分统计（用于画像计算的中间结果）
  type DimensionScore struct {
      Dimension  string // 维度名称
      TotalScore int    // 该维度总分
      UserScore  int    // 用户在该维度的得分
  }
  ```

## Round 1 问题修复情况

### ✅ 已修复（7 个）

| 问题编号 | 问题描述 | 修复情况 |
|---------|---------|---------|
| **H1** | goroutine 缺乏 panic 恢复和超时控制 | ✅ 已添加 `defer recover()` 和 `context.WithTimeout(5s)` |
| **H4** | 数据库缺少 CHECK 约束 | ✅ Migration 已添加 `CHECK (dimension != '')` 和 `CHECK (score >= 0 AND score <= 1)` |
| **H5** | Service 层直接使用 db | ✅ SQL 查询已封装到 `Repository.GetDimensionScores()` |
| **M1** | 缺少画像默认值处理 | ✅ `GetSkillProfile` 已填充所有维度，缺失的默认为 0 |
| **M2** | 缺少维度合法性校验 | ✅ 已添加 `model.IsValidDimension()` 和 `model.ValidDimensions` |
| **L2** | 缺少关键日志 | ✅ 已添加画像计算开始、耗时、结果日志 |
| **L4** | 缺少 DTO 字段注释 | ✅ 已添加详细的字段注释 |

### ⚠️ 部分修复（2 个）

| 问题编号 | 问题描述 | 修复情况 |
|---------|---------|---------|
| **H2** | 数据一致性问题 | ⚠️ 添加了分布式锁，但未实现延迟触发（见 M5） |
| **L1** | DimensionScore 应提取为独立类型 | ⚠️ 已提取但位置不当（见 L6） |

### ❌ 未修复（5 个）

| 问题编号 | 问题描述 | 状态 |
|---------|---------|------|
| **H3** | 全量重算性能问题 | ❌ 未实现增量更新（见 H6） |
| **M3** | assessmentService 可能为 nil | ❌ 仍使用 nil 检查，未强制要求 |
| **M4** | 缺少 challenges 表缓存 | ❌ 未实现 Redis 缓存（见 M6） |
| **L3** | GetStudentSkillProfile 命名不准确 | ❌ 未修改 |
| **L5** | 缺少单元测试 | ❌ 未添加测试 |

## 统计摘要

| 级别 | 本轮新增 | Round 1 遗留 | 合计 |
|------|---------|-------------|------|
| 🔴 高 | 2 | 1 | 3 |
| 🟡 中 | 2 | 1 | 3 |
| 🟢 低 | 1 | 2 | 3 |
| 合计 | 5 | 4 | 9 |

## 总体评价

本轮修复解决了 Round 1 中 7 个问题（50%），特别是：
- ✅ 并发安全的基础保障（panic 恢复、超时控制）
- ✅ 架构分层问题（SQL 封装到 Repository）
- ✅ 数据完整性约束（CHECK 约束、维度校验）
- ✅ 用户体验改进（默认值填充、字段注释）

**仍需修复的关键问题**：

1. **H6（高）**：全量重算性能问题未解决，这是最严重的遗留问题
   - 建议优先实现增量更新或缓存优化
   - 当前方案在用户提交数增长后会成为性能瓶颈

2. **H7（高）**：分布式锁失败时的处理逻辑有缺陷
   - Redis 故障时并发安全失效
   - 修复成本低，应立即修复

3. **M5（中）**：数据一致性保障不够完善
   - 建议添加延迟触发机制
   - 为未来可能的事务改造预留空间

4. **M6（中）**：缺少 challenges 表缓存
   - 可以显著降低数据库负载
   - 实现成本低，收益高

**建议**：
- 必须修复：H6、H7（阻塞上线）
- 强烈建议修复：M5、M6（性能和稳定性）
- 可延后：L3、L5、M3（不影响功能）

修复 H6 和 H7 后可以上线，但建议在下一轮迭代中优化性能（M6）和数据一致性（M5）。
