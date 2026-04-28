# CTF 后端代码 Review（skill-profile 第 1 轮）：能力画像生成功能

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | skill-profile |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | e3c120a..6b9dd04（2 个提交，9 个文件，+284/-17 行） |
| 变更概述 | 实现用户能力画像生成功能，包括画像计算、存储、查询和自动更新 |
| 审查基准 | docs/architecture/backend/02-database-design.md, CLAUDE.md |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | - |

## 问题清单

### 🔴 高优先级

#### [H1] 并发安全：goroutine 异步更新画像缺乏错误恢复机制
- **文件**：`code/backend/internal/module/practice/service.go:137-143`
- **问题描述**：
  ```go
  go func() {
      if _, err := s.assessmentService.CalculateSkillProfile(userID); err != nil {
          s.logger.Error("更新能力画像失败", zap.Int64("userID", userID), zap.Error(err))
      }
  }()
  ```
  使用裸 goroutine 异步更新画像，存在以下风险：
  1. goroutine panic 会导致整个进程崩溃，没有 recover 保护
  2. 没有超时控制，如果数据库查询慢会导致 goroutine 泄漏
  3. 没有并发数控制，高并发提交 Flag 时会创建大量 goroutine
  4. 错误只记录日志，用户无法感知画像更新失败

- **影响范围/风险**：
  - 高并发场景下可能导致服务崩溃或资源耗尽
  - 画像更新失败时用户无感知，数据不一致

- **修正建议**：
  ```go
  // 方案 1：使用 worker pool 限制并发
  select {
  case s.profileUpdateChan <- userID:
      // 成功入队
  default:
      s.logger.Warn("画像更新队列已满", zap.Int64("userID", userID))
  }

  // 方案 2：添加 panic 恢复和超时控制
  go func() {
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
  }()
  ```

#### [H2] 数据一致性：画像计算使用非事务查询，可能读到脏数据
- **文件**：`code/backend/internal/module/assessment/service.go:33-41`
- **问题描述**：
  ```go
  err := s.db.Raw(`
      SELECT
          c.category as dimension,
          SUM(c.points) as total_score,
          COALESCE(SUM(CASE WHEN s.is_correct = 1 THEN c.points ELSE 0 END), 0) as user_score
      FROM challenges c
      LEFT JOIN submissions s ON c.id = s.challenge_id AND s.user_id = ?
      WHERE c.status = 'published'
      GROUP BY c.category
  `, userID).Scan(&scores).Error
  ```
  问题：
  1. 在 Flag 提交成功后立即异步计算画像，此时 submission 记录可能还未提交（如果外层有事务）
  2. 没有使用 `FOR UPDATE` 或事务隔离，并发提交时可能读到中间状态
  3. 如果用户同时提交多个 Flag，多个 goroutine 并发计算画像会产生竞态条件

- **影响范围/风险**：
  - 画像分数可能不准确（遗漏刚提交的 Flag）
  - 并发更新时可能出现数据覆盖

- **修正建议**：
  ```go
  // 方案 1：延迟触发（推荐）
  // 在 Flag 提交事务提交后再触发画像更新
  time.AfterFunc(100*time.Millisecond, func() {
      // 异步更新逻辑
  })

  // 方案 2：使用分布式锁
  lockKey := fmt.Sprintf("skill_profile:update:%d", userID)
  if !s.redis.SetNX(ctx, lockKey, 1, 10*time.Second).Val() {
      return // 已有更新任务在执行
  }
  defer s.redis.Del(ctx, lockKey)

  // 方案 3：使用数据库行锁
  // 在 skill_profiles 表上加 SELECT ... FOR UPDATE
  ```

#### [H3] 性能问题：每次 Flag 提交都全量重算所有维度画像
- **文件**：`code/backend/internal/module/assessment/service.go:23-76`
- **问题描述**：
  `CalculateSkillProfile` 方法每次都扫描用户所有提交记录，重新计算所有维度的分数。在以下场景下性能低下：
  1. 用户提交了 100 道题后，每次新提交都要重新聚合 100+ 条记录
  2. 即使只提交了 Web 类题目，也会重算 Pwn、Reverse 等所有维度
  3. 高并发时会产生大量重复的聚合查询

- **影响范围/风险**：
  - 数据库负载高，响应变慢
  - 用户提交 Flag 的响应时间受影响（虽然是异步，但占用数据库连接）

- **修正建议**：
  ```go
  // 方案 1：增量更新（推荐）
  func (s *Service) IncrementSkillProfile(userID int64, dimension string, points int) error {
      // 1. 查询该维度总分（可缓存）
      totalScore := s.getTotalScoreByDimension(dimension)

      // 2. 查询用户该维度已得分
      userScore := s.getUserScoreByDimension(userID, dimension)

      // 3. 增量更新
      newScore := float64(userScore+points) / float64(totalScore)

      return s.repo.Upsert(&model.SkillProfile{
          UserID:    userID,
          Dimension: dimension,
          Score:     newScore,
          UpdatedAt: time.Now(),
      })
  }

  // 方案 2：使用 Redis 缓存聚合结果
  // 缓存每个维度的总分，避免重复计算
  ```

#### [H4] 数据库设计：缺少唯一索引导致可能插入重复记录
- **文件**：`code/backend/migrations/000010_create_skill_profiles_table.up.sql:7`
- **问题描述**：
  ```sql
  UNIQUE KEY idx_user_dimension (user_id, dimension),
  ```
  虽然定义了 UNIQUE KEY，但 GORM 的 `OnConflict` 行为依赖数据库实现：
  1. MySQL 使用 `ON DUPLICATE KEY UPDATE`，正常工作
  2. 但如果 dimension 字段为空字符串，某些数据库可能不认为是冲突
  3. 缺少 NOT NULL 约束检查（虽然 Model 中有 `not null` tag）

- **影响范围/风险**：
  - 理论上可能插入 (user_id, '') 的脏数据
  - 数据完整性风险

- **修正建议**：
  ```sql
  CREATE TABLE IF NOT EXISTS skill_profiles (
      id BIGINT PRIMARY KEY AUTO_INCREMENT,
      user_id BIGINT NOT NULL,
      dimension VARCHAR(20) NOT NULL CHECK (dimension != ''),  -- 添加非空检查
      score DOUBLE NOT NULL DEFAULT 0 CHECK (score >= 0 AND score <= 1),  -- 添加范围检查
      updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
      UNIQUE KEY idx_user_dimension (user_id, dimension),
      KEY idx_user_id (user_id),
      KEY idx_updated_at (updated_at)  -- 方便按时间查询
  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
  ```

#### [H5] 架构违规：Service 层直接使用 db 执行原生 SQL
- **文件**：`code/backend/internal/module/assessment/service.go:33-41`
- **问题描述**：
  ```go
  func (s *Service) CalculateSkillProfile(userID int64) ([]*dto.SkillDimension, error) {
      var scores []DimensionScore
      err := s.db.Raw(`...`).Scan(&scores).Error  // Service 直接用 db
  ```
  违反分层规范：
  1. Service 应该调用 Repository，而不是直接操作 `s.db`
  2. 原生 SQL 应该封装在 Repository 层，便于测试和维护
  3. 当前 Repository 只有 CRUD 方法，缺少聚合查询方法

- **影响范围/风险**：
  - 分层混乱，难以单元测试
  - SQL 逻辑分散，不利于维护

- **修正建议**：
  ```go
  // Repository 层添加聚合方法
  func (r *Repository) GetDimensionScores(userID int64) ([]DimensionScore, error) {
      var scores []DimensionScore
      err := r.db.Raw(`
          SELECT
              c.category as dimension,
              SUM(c.points) as total_score,
              COALESCE(SUM(CASE WHEN s.is_correct = 1 THEN c.points ELSE 0 END), 0) as user_score
          FROM challenges c
          LEFT JOIN submissions s ON c.id = s.challenge_id AND s.user_id = ?
          WHERE c.status = 'published'
          GROUP BY c.category
      `, userID).Scan(&scores).Error
      return scores, err
  }

  // Service 层调用
  func (s *Service) CalculateSkillProfile(userID int64) ([]*dto.SkillDimension, error) {
      scores, err := s.repo.GetDimensionScores(userID)
      if err != nil {
          return nil, err
      }
      // 转换逻辑...
  }
  ```

### 🟡 中优先级

#### [M1] 接口设计：缺少画像不存在时的默认值处理
- **文件**：`code/backend/internal/module/assessment/service.go:79-103`
- **问题描述**：
  ```go
  func (s *Service) GetSkillProfile(userID int64) (*dto.SkillProfileResp, error) {
      profiles, err := s.repo.FindByUserID(userID)
      if err != nil {
          return nil, err
      }
      // 如果 profiles 为空，返回空数组
  ```
  问题：
  1. 新用户或未提交过 Flag 的用户，返回空的 dimensions 数组
  2. 前端需要额外判断空数组，用户体验不好
  3. 应该返回所有维度的初始值（score=0）

- **影响范围/风险**：
  - 前端需要额外处理空数据
  - 用户看不到完整的能力雷达图

- **修正建议**：
  ```go
  func (s *Service) GetSkillProfile(userID int64) (*dto.SkillProfileResp, error) {
      profiles, err := s.repo.FindByUserID(userID)
      if err != nil {
          return nil, err
      }

      // 构建所有维度的 map
      allDimensions := []string{
          model.DimensionWeb,
          model.DimensionPwn,
          model.DimensionReverse,
          model.DimensionCrypto,
          model.DimensionMisc,
          model.DimensionForensics,
      }

      dimensionMap := make(map[string]float64)
      var latestUpdate time.Time

      for _, p := range profiles {
          dimensionMap[p.Dimension] = p.Score
          if p.UpdatedAt.After(latestUpdate) {
              latestUpdate = p.UpdatedAt
          }
      }

      // 填充缺失的维度（默认 0 分）
      dimensions := make([]*dto.SkillDimension, 0, len(allDimensions))
      for _, dim := range allDimensions {
          dimensions = append(dimensions, &dto.SkillDimension{
              Dimension: dim,
              Score:     dimensionMap[dim], // 不存在时为 0
          })
      }

      return &dto.SkillProfileResp{
          UserID:     userID,
          Dimensions: dimensions,
          UpdatedAt:  latestUpdate.Format(time.RFC3339),
      }, nil
  }
  ```

#### [M2] 错误处理：缺少对无效 dimension 的校验
- **文件**：`code/backend/internal/model/skill_profile.go:5-12`
- **问题描述**：
  定义了维度常量，但没有校验函数：
  ```go
  const (
      DimensionWeb       = "web"
      DimensionPwn       = "pwn"
      // ...
  )
  ```
  问题：
  1. 如果 challenges 表的 category 字段有脏数据（如 "WEB"、"web-security"），会生成无效的画像记录
  2. Repository 的 Upsert 方法不校验 dimension 是否合法
  3. 前端可能收到未知的维度名称

- **影响范围/风险**：
  - 数据不一致，前端无法正确渲染
  - 脏数据污染画像表

- **修正建议**：
  ```go
  // model/skill_profile.go
  var ValidDimensions = map[string]bool{
      DimensionWeb:       true,
      DimensionPwn:       true,
      DimensionReverse:   true,
      DimensionCrypto:    true,
      DimensionMisc:      true,
      DimensionForensics: true,
  }

  func IsValidDimension(dimension string) bool {
      return ValidDimensions[dimension]
  }

  // service.go 中添加校验
  for _, score := range scores {
      if !model.IsValidDimension(score.Dimension) {
          s.logger.Warn("跳过无效维度", zap.String("dimension", score.Dimension))
          continue
      }
      // ...
  }
  ```

#### [M3] 依赖注入：assessmentService 可能为 nil 但未在构造函数中强制要求
- **文件**：`code/backend/internal/module/practice/service.go:137`
- **问题描述**：
  ```go
  if s.assessmentService != nil {
      go func() { ... }()
  }
  ```
  问题：
  1. 使用 nil 检查来判断是否启用画像功能，这是隐式依赖
  2. 如果忘记注入 assessmentService，功能会静默失败（只是不更新画像）
  3. 不符合依赖注入的显式原则

- **影响范围/风险**：
  - 配置错误时难以发现问题
  - 测试时容易遗漏依赖

- **修正建议**：
  ```go
  // 方案 1：构造函数中强制要求（推荐）
  func NewService(
      repo *Repository,
      challengeRepo ChallengeRepository,
      instanceRepo InstanceRepository,
      assessmentService AssessmentService,  // 必填
      redis *redis.Client,
      logger *zap.Logger,
      globalSecret string,
      submitLimit int,
      submitWindow time.Duration,
  ) *Service {
      if assessmentService == nil {
          panic("assessmentService is required")
      }
      return &Service{...}
  }

  // 方案 2：使用 Option 模式
  type ServiceOption func(*Service)

  func WithAssessmentService(svc AssessmentService) ServiceOption {
      return func(s *Service) {
          s.assessmentService = svc
      }
  }
  ```

#### [M4] 性能：缺少对 challenges 表的缓存，每次计算都查询
- **文件**：`code/backend/internal/module/assessment/service.go:33-41`
- **问题描述**：
  SQL 查询中 `FROM challenges c WHERE c.status = 'published'` 每次都扫描 challenges 表，但：
  1. 已发布的题目列表变化频率很低（只有管理员发布新题时才变）
  2. 每个维度的总分是相对固定的
  3. 高并发时会产生大量重复查询

- **影响范围/风险**：
  - 数据库负载高
  - 响应时间慢

- **修正建议**：
  ```go
  // 使用 Redis 缓存各维度总分
  func (s *Service) getTotalScoreByDimension(dimension string) (int, error) {
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

  // 在发布新题目时清除缓存
  // challenge/service.go PublishChallenge 方法中：
  s.redis.Del(ctx, fmt.Sprintf("skill_profile:total_score:%s", challenge.Category))
  ```

### 🟢 低优先级

#### [L1] 代码风格：DimensionScore 结构体应提取为独立类型
- **文件**：`code/backend/internal/module/assessment/service.go:27-30`
- **问题描述**：
  ```go
  type DimensionScore struct {
      Dimension  string
      TotalScore int
      UserScore  int
  }
  ```
  定义在函数内部，不利于复用和测试

- **影响范围/风险**：
  - 代码可读性略差
  - 无法在其他方法中复用

- **修正建议**：
  ```go
  // 提取到文件顶部或单独的 types.go
  type DimensionScore struct {
      Dimension  string
      TotalScore int
      UserScore  int
  }
  ```

#### [L2] 日志：缺少画像计算的关键日志
- **文件**：`code/backend/internal/module/assessment/service.go:23-76`
- **问题描述**：
  `CalculateSkillProfile` 方法没有记录计算开始、耗时、结果等日志，不利于排查问题

- **影响范围/风险**：
  - 生产环境问题难以排查
  - 无法监控画像计算性能

- **修正建议**：
  ```go
  func (s *Service) CalculateSkillProfile(userID int64) ([]*dto.SkillDimension, error) {
      start := time.Now()
      defer func() {
          s.logger.Info("画像计算完成",
              zap.Int64("userID", userID),
              zap.Duration("duration", time.Since(start)),
          )
      }()

      // 原有逻辑...

      s.logger.Debug("画像计算结果",
          zap.Int64("userID", userID),
          zap.Int("dimensionCount", len(dimensions)),
      )

      return dimensions, nil
  }
  ```

#### [L3] 命名：GetStudentSkillProfile 方法名不准确
- **文件**：`code/backend/internal/module/assessment/handler.go:32`
- **问题描述**：
  ```go
  func (h *Handler) GetStudentSkillProfile(c *gin.Context) {
  ```
  方法名暗示只能查询学员，但实际上可以查询任意用户（包括教师、管理员）

- **影响范围/风险**：
  - 命名误导，可读性差

- **修正建议**：
  ```go
  // 改为更通用的名称
  func (h *Handler) GetUserSkillProfile(c *gin.Context) {
      userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
      // ...
  }

  // 路由改为
  teacherOrAbove.GET("/users/:id/skill-profile", assessmentHandler.GetUserSkillProfile)
  ```

#### [L4] 文档：缺少 DTO 字段的注释说明
- **文件**：`code/backend/internal/dto/skill_profile.go:4-13`
- **问题描述**：
  ```go
  type SkillDimension struct {
      Dimension string  `json:"dimension"`
      Score     float64 `json:"score"`
  }
  ```
  缺少字段说明，不清楚 Score 的取值范围（0-1 还是 0-100）

- **影响范围/风险**：
  - 前端开发者不清楚字段含义
  - API 文档不完整

- **修正建议**：
  ```go
  // SkillDimension 能力维度
  type SkillDimension struct {
      Dimension string  `json:"dimension"` // 维度名称（web/pwn/reverse/crypto/misc/forensics）
      Score     float64 `json:"score"`     // 得分率（0.0-1.0，表示该维度的完成百分比）
  }

  // SkillProfileResp 能力画像响应
  type SkillProfileResp struct {
      UserID     int64             `json:"user_id"`     // 用户ID
      Dimensions []*SkillDimension `json:"dimensions"`  // 各维度得分
      UpdatedAt  string            `json:"updated_at"`  // 最后更新时间（RFC3339格式）
  }
  ```

#### [L5] 测试：缺少单元测试
- **文件**：整个 assessment 模块
- **问题描述**：
  新增的 Repository、Service、Handler 都没有对应的测试文件

- **影响范围/风险**：
  - 代码质量无法保证
  - 重构时容易引入 bug

- **修正建议**：
  至少添加以下测试：
  ```
  assessment/repository_test.go
  - TestUpsert
  - TestFindByUserID
  - TestBatchUpsert

  assessment/service_test.go
  - TestCalculateSkillProfile
  - TestGetSkillProfile
  - TestGetSkillProfile_EmptyData

  assessment/handler_test.go
  - TestGetMySkillProfile
  - TestGetStudentSkillProfile
  ```

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 5 |
| 🟡 中 | 4 |
| 🟢 低 | 5 |
| 合计 | 14 |

## 总体评价

本次实现的能力画像功能在架构设计上基本合理，Model/DTO 分离正确，分层结构清晰。但存在以下关键问题需要修复：

**必须修复的问题**：
1. **并发安全**：裸 goroutine 缺乏 panic 恢复和并发控制（H1）
2. **数据一致性**：异步更新时可能读到脏数据，需要延迟触发或加锁（H2）
3. **性能问题**：全量重算效率低，应改为增量更新（H3）
4. **架构违规**：Service 层直接使用 db，应封装到 Repository（H5）

**建议优化的问题**：
1. 新用户应返回所有维度的初始值（M1）
2. 添加维度合法性校验（M2）
3. 使用 Redis 缓存各维度总分（M4）

修复高优先级问题后，该功能可以上线，但建议在下一轮迭代中优化性能和用户体验。
