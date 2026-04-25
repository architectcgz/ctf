# CTF 后端代码 Review（竞赛提交与计分 第 1 轮）：竞赛提交与计分功能实现

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | contest-submission |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | 5 个 commit (9e0fa7d..ca8a103)，11 个文件，364 行新增 |
| 变更概述 | 实现竞赛提交 Flag、首杀记录、团队计分功能 |
| 审查基准 | `/home/azhi/workspace/projects/ctf/CLAUDE.md` |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | - |

## 问题清单

### 🔴 高优先级

#### [H1] 首杀竞争条件：并发提交可能导致多个首杀
- **文件**：`code/backend/internal/module/contest/submission_service.go:88-95`
- **问题描述**：
  ```go
  isFirstBlood := cc.FirstBloodBy == nil
  if isFirstBlood {
      bonus := int(float64(score) * s.cfg.Contest.FirstBloodBonus)
      score += bonus
  }

  if err := s.updateScoreAndFirstBlood(...); err != nil {
      return nil, err
  }
  ```
  首杀判断（`cc.FirstBloodBy == nil`）和更新（`updateScoreAndFirstBlood`）之间存在时间窗口。两个用户同时提交正确 Flag 时，都可能读到 `FirstBloodBy == nil`，导致双首杀。
- **影响范围/风险**：
  - 多个用户获得首杀奖励
  - 团队分数计算错误
  - 排行榜数据不准确
- **修正建议**：
  使用数据库乐观锁或悲观锁保证首杀的原子性：

  **方案 1：乐观锁（推荐）**
  ```go
  func (s *SubmissionService) updateScoreAndFirstBlood(...) error {
      return s.db.Transaction(func(tx *gorm.DB) error {
          // 使用 CAS 更新首杀
          result := tx.Model(&model.ContestChallenge).
              Where("contest_id = ? AND challenge_id = ? AND first_blood_by IS NULL",
                    contestID, challengeID).
              Update("first_blood_by", userID)

          // 检查是否真的抢到首杀
          actualFirstBlood := result.RowsAffected > 0

          // 根据实际首杀结果重新计算分数
          finalScore := baseScore
          if actualFirstBlood {
              finalScore += int(float64(baseScore) * s.cfg.Contest.FirstBloodBonus)
          }

          // 更新团队分数
          if teamID != nil {
              // ...
          }

          return nil
      })
  }
  ```

  **方案 2：悲观锁**
  ```go
  tx.Clauses(clause.Locking{Strength: "UPDATE"}).
      Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
      First(&cc)
  ```

#### [H2] 重复提交检查不完整：可能绕过已解决检查
- **文件**：`code/backend/internal/module/contest/submission_service.go:61-68`
- **问题描述**：
  ```go
  var existingSub model.Submission
  err := s.db.Where("user_id = ? AND challenge_id = ? AND contest_id = ? AND is_correct = ?",
      userID, challengeID, contestID, true).First(&existingSub).Error
  if err == nil {
      return nil, errcode.ErrAlreadySolved
  }
  ```
  检查在 Flag 验证之前，但在首杀判断和分数更新之后没有再次检查。并发场景下，两个请求都可能通过检查，导致同一用户多次获得分数。
- **影响范围/风险**：
  - 用户可能通过并发请求刷分
  - 团队总分虚高
  - 排行榜作弊
- **修正建议**：
  在事务内使用唯一索引或再次检查：

  **方案 1：数据库唯一约束（推荐）**
  ```sql
  -- 在迁移文件中添加
  CREATE UNIQUE INDEX uk_submission_contest_user_challenge
  ON submissions(contest_id, user_id, challenge_id)
  WHERE is_correct = TRUE AND contest_id IS NOT NULL;
  ```

  **方案 2：事务内二次检查**
  ```go
  return s.db.Transaction(func(tx *gorm.DB) error {
      // 在事务内加锁检查
      var count int64
      tx.Model(&model.Submission{}).
          Where("user_id = ? AND challenge_id = ? AND contest_id = ? AND is_correct = ?",
                userID, challengeID, contestID, true).
          Count(&count)

      if count > 0 {
          return errcode.ErrAlreadySolved
      }

      // 继续首杀和计分逻辑
      // ...
  })
  ```

#### [H3] 事务边界错误：Submission 创建在事务外
- **文件**：`code/backend/internal/module/contest/submission_service.go:95-111`
- **问题描述**：
  ```go
  if err := s.updateScoreAndFirstBlood(...); err != nil {
      return nil, err
  }

  submission := &model.Submission{...}
  if err := s.db.Create(submission).Error; err != nil {
      return nil, err
  }
  ```
  分数更新在事务内，但 Submission 记录创建在事务外。如果 `Create` 失败，团队已经加分但没有提交记录，导致数据不一致。
- **影响范围/风险**：
  - 分数增加但无提交记录
  - 审计日志缺失
  - 无法追溯分数来源
- **修正建议**：
  将 Submission 创建移入事务：
  ```go
  func (s *SubmissionService) SubmitFlagInContest(...) (*dto.SubmissionResp, error) {
      // ... 前置检查 ...

      isCorrect, err := s.flagService.ValidateFlag(...)
      if err != nil {
          return nil, err
      }

      var finalScore int
      var submission *model.Submission

      if isCorrect {
          err = s.db.Transaction(func(tx *gorm.DB) error {
              // 1. 计算分数和首杀
              // 2. 更新首杀和团队分数
              // 3. 创建 Submission 记录
              submission = &model.Submission{...}
              return tx.Create(submission).Error
          })
          if err != nil {
              return nil, err
          }
      } else {
          // 错误提交直接创建记录，不需要事务
          submission = &model.Submission{...}
          if err := s.db.Create(submission).Error; err != nil {
              return nil, err
          }
      }

      return &dto.SubmissionResp{...}, nil
  }
  ```

#### [H4] 缺少幂等性保障：网络重试可能导致重复计分
- **文件**：`code/backend/internal/module/contest/submission_service.go:28`
- **问题描述**：
  `SubmitFlagInContest` 方法没有幂等性设计。客户端网络超时重试时，可能导致同一次提交被处理多次（即使有 `ErrAlreadySolved` 检查，并发窗口仍存在）。
- **影响范围/风险**：
  - 用户体验差（不确定是否提交成功）
  - 可能触发并发竞争导致重复计分
  - 日志中出现大量重复提交记录
- **修正建议**：
  添加幂等键机制：
  ```go
  // DTO 添加幂等键
  type SubmitFlagReq struct {
      Flag        string `json:"flag" binding:"required"`
      IdempotencyKey string `json:"idempotency_key" binding:"required,uuid"`
  }

  // Service 检查幂等键
  func (s *SubmissionService) SubmitFlagInContest(..., idempotencyKey string) (*dto.SubmissionResp, error) {
      // 1. 检查幂等键是否已处理（使用 Redis 或数据库）
      cacheKey := fmt.Sprintf("submission:idempotency:%s", idempotencyKey)
      if cached, _ := s.redis.Get(ctx, cacheKey).Result(); cached != "" {
          // 返回缓存的结果
          var resp dto.SubmissionResp
          json.Unmarshal([]byte(cached), &resp)
          return &resp, nil
      }

      // 2. 正常处理提交
      resp, err := s.doSubmit(...)
      if err != nil {
          return nil, err
      }

      // 3. 缓存结果（TTL 5分钟）
      respJSON, _ := json.Marshal(resp)
      s.redis.Set(ctx, cacheKey, respJSON, 5*time.Minute)

      return resp, nil
  }
  ```

### 🟡 中优先级

#### [M1] Model/DTO 分离不完整：缺少 DTO 定义
- **文件**：`code/backend/internal/module/contest/submission_service.go:28`
- **问题描述**：
  Service 方法直接返回 `*dto.SubmissionResp`，但代码中未见 `dto.SubmissionResp` 的定义。根据项目规范，应该在 `internal/dto/` 中定义所有 API 响应结构。
- **影响范围/风险**：
  - 违反项目分层规范
  - 可能导致编译错误
  - 代码审查时无法确认响应字段是否合理
- **修正建议**：
  在 `internal/dto/contest.go` 中添加：
  ```go
  type SubmitFlagReq struct {
      Flag string `json:"flag" binding:"required,max=500"`
  }

  type SubmissionResp struct {
      IsCorrect   bool      `json:"is_correct"`
      Message     string    `json:"message"`
      Points      int       `json:"points"`
      SubmittedAt time.Time `json:"submitted_at"`
  }
  ```

#### [M2] 硬编码错误消息：应提取为常量
- **文件**：`code/backend/internal/module/contest/submission_service.go:114-117`
- **问题描述**：
  ```go
  message := "Flag 错误"
  if isCorrect {
      message = "恭喜，Flag 正确！"
  }
  ```
  错误消息硬编码在业务逻辑中，不利于国际化和统一管理。
- **影响范围/风险**：
  - 多处重复相同消息时难以维护
  - 无法支持多语言
  - 消息文案修改需要改代码
- **修正建议**：
  提取到常量或配置：
  ```go
  // internal/constants/messages.go
  package constants

  const (
      MsgFlagIncorrect = "Flag 错误"
      MsgFlagCorrect   = "恭喜，Flag 正确！"
  )

  // 使用
  message := constants.MsgFlagIncorrect
  if isCorrect {
      message = constants.MsgFlagCorrect
  }
  ```

#### [M3] 首杀奖励计算精度问题：浮点数转整数可能丢失精度
- **文件**：`code/backend/internal/module/contest/submission_service.go:91`
- **问题描述**：
  ```go
  bonus := int(float64(score) * s.cfg.Contest.FirstBloodBonus)
  ```
  直接截断浮点数可能导致奖励分数不符合预期。例如 `100 * 0.15 = 15.0`，但 `99 * 0.15 = 14.85` 截断为 `14`。
- **影响范围/风险**：
  - 首杀奖励不一致
  - 用户体验差（预期 15 分实际 14 分）
- **修正建议**：
  使用四舍五入：
  ```go
  import "math"

  bonus := int(math.Round(float64(score) * s.cfg.Contest.FirstBloodBonus))
  ```

#### [M4] 缺少输入校验：Handler 未校验路径参数
- **文件**：`code/backend/internal/module/contest/submission_handler.go:20-22`
- **问题描述**：
  ```go
  contestID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
  challengeID, _ := strconv.ParseInt(c.Param("cid"), 10, 64)
  ```
  忽略 `ParseInt` 的错误返回值，如果路径参数不是数字，会传入 `0` 导致逻辑错误。
- **影响范围/风险**：
  - 非法请求可能绕过校验
  - 错误信息不明确
- **修正建议**：
  ```go
  contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
  if err != nil || contestID <= 0 {
      response.Error(c, errcode.ErrInvalidParam("竞赛ID"))
      return
  }

  challengeID, err := strconv.ParseInt(c.Param("cid"), 10, 64)
  if err != nil || challengeID <= 0 {
      response.Error(c, errcode.ErrInvalidParam("题目ID"))
      return
  }
  ```

#### [M5] 竞赛状态校验不完整：未检查 Contest.Status 字段
- **文件**：`code/backend/internal/module/contest/submission_service.go:36-42`
- **问题描述**：
  ```go
  now := time.Now()
  if now.Before(contest.StartAt) {
      return nil, errcode.ErrContestNotStarted
  }
  if now.After(contest.EndAt) {
      return nil, errcode.ErrContestEnded
  }
  ```
  只检查了时间范围，未检查 `contest.Status`。如果管理员手动暂停竞赛（`status = 'paused'`），用户仍可提交。
- **影响范围/风险**：
  - 无法通过状态控制竞赛暂停
  - 管理员操作失效
- **修正建议**：
  ```go
  if contest.Status != model.ContestStatusRunning {
      return nil, errcode.ErrContestNotRunning
  }

  now := time.Now()
  if now.Before(contest.StartAt) {
      return nil, errcode.ErrContestNotStarted
  }
  if now.After(contest.EndAt) {
      return nil, errcode.ErrContestEnded
  }
  ```

#### [M6] 缺少注册状态校验：未检查 ContestRegistration.Status
- **文件**：`code/backend/internal/module/contest/submission_service.go:45-50`
- **问题描述**：
  ```go
  var reg model.ContestRegistration
  if err := s.db.Where("contest_id = ? AND user_id = ?", contestID, userID).First(&reg).Error; err != nil {
      if err == gorm.ErrRecordNotFound {
          return nil, errcode.ErrNotRegistered
      }
      return nil, err
  }
  ```
  只检查注册记录是否存在，未检查 `reg.Status`。如果用户注册被拒绝（`status = 'rejected'`）或取消（`status = 'cancelled'`），仍可提交。
- **影响范围/风险**：
  - 被拒绝的用户可以参赛
  - 权限控制失效
- **修正建议**：
  ```go
  var reg model.ContestRegistration
  if err := s.db.Where("contest_id = ? AND user_id = ?", contestID, userID).First(&reg).Error; err != nil {
      if err == gorm.ErrRecordNotFound {
          return nil, errcode.ErrNotRegistered
      }
      return nil, err
  }

  if reg.Status != "approved" {
      return nil, errcode.ErrRegistrationNotApproved
  }
  ```

### 🟢 低优先级

#### [L1] 数据库索引可优化：submissions 表缺少复合索引
- **文件**：`code/backend/migrations/000004_contest_submission.up.sql:7`
- **问题描述**：
  ```sql
  CREATE INDEX idx_submissions_contest_id ON submissions(contest_id) WHERE contest_id IS NOT NULL;
  ```
  只创建了单列索引，但查询条件通常是 `contest_id + user_id + challenge_id`（见 L62 的 WHERE 条件）。单列索引效率较低。
- **影响范围/风险**：
  - 查询性能不佳
  - 数据量增长后响应变慢
- **修正建议**：
  ```sql
  -- 替换为复合索引
  CREATE INDEX idx_submissions_contest_user_challenge
  ON submissions(contest_id, user_id, challenge_id)
  WHERE contest_id IS NOT NULL;
  ```

#### [L2] 缺少日志记录：关键操作无审计日志
- **文件**：`code/backend/internal/module/contest/submission_service.go:28`
- **问题描述**：
  整个提交流程没有日志记录，无法追踪：
  - 谁在什么时间提交了什么 Flag
  - 首杀是谁抢到的
  - 分数变更历史
- **影响范围/风险**：
  - 问题排查困难
  - 无法审计作弊行为
  - 用户投诉时无法举证
- **修正建议**：
  ```go
  import "go.uber.org/zap"

  func (s *SubmissionService) SubmitFlagInContest(...) (*dto.SubmissionResp, error) {
      s.logger.Info("用户提交 Flag",
          zap.Int64("user_id", userID),
          zap.Int64("contest_id", contestID),
          zap.Int64("challenge_id", challengeID),
      )

      // ... 处理逻辑 ...

      if isCorrect {
          s.logger.Info("Flag 正确",
              zap.Int64("user_id", userID),
              zap.Int64("challenge_id", challengeID),
              zap.Int("score", score),
              zap.Bool("is_first_blood", isFirstBlood),
          )
      }

      return resp, nil
  }
  ```

#### [L3] 配置默认值不合理：首杀奖励 10% 可能过低
- **文件**：`code/backend/internal/config/config.go:235`
- **问题描述**：
  ```go
  v.SetDefault("contest.first_blood_bonus", 0.1)
  ```
  首杀奖励默认 10%，对于高分题（如 500 分）只奖励 50 分，激励不足。
- **影响范围/风险**：
  - 首杀竞争不激烈
  - 用户参与度低
- **修正建议**：
  参考 CTF 惯例，建议 20%-50%：
  ```go
  v.SetDefault("contest.first_blood_bonus", 0.2) // 20%
  ```
  或使用固定分数 + 百分比：
  ```go
  type ContestConfig struct {
      FirstBloodBonusPercent float64 `mapstructure:"first_blood_bonus_percent"`
      FirstBloodBonusFixed   int     `mapstructure:"first_blood_bonus_fixed"`
  }

  bonus := s.cfg.Contest.FirstBloodBonusFixed +
           int(math.Round(float64(score) * s.cfg.Contest.FirstBloodBonusPercent))
  ```

#### [L4] Model 字段注释缺失：业务含义不明确
- **文件**：`code/backend/internal/model/contest_challenge.go:9`
- **问题描述**：
  ```go
  ContestScore *int `gorm:"column:contest_score"`
  ```
  `ContestScore` 字段为指针类型且可为 nil，但注释未说明：
  - nil 时使用什么分数？
  - 为什么设计为可选？
- **影响范围/风险**：
  - 代码可读性差
  - 新开发者容易误用
- **修正建议**：
  ```go
  // ContestScore 竞赛中该题目的自定义分数
  // 如果为 nil，则使用 Challenge.Points 作为基础分数
  ContestScore *int `gorm:"column:contest_score"`
  ```

#### [L5] 团队模式未完全实现：TeamMode 字段未使用
- **文件**：`code/backend/internal/model/contest.go:15`
- **问题描述**：
  ```go
  TeamMode bool `gorm:"column:team_mode;not null;default:false"`
  ```
  Contest 有 `TeamMode` 字段，但 `SubmitFlagInContest` 未根据此字段区分个人/团队模式。个人赛中 `reg.TeamID` 应为 nil，但代码未校验。
- **影响范围/风险**：
  - 个人赛可能错误地更新团队分数
  - 业务逻辑不完整
- **修正建议**：
  ```go
  // 检查团队模式一致性
  if contest.TeamMode && reg.TeamID == nil {
      return nil, errcode.ErrTeamRequired
  }
  if !contest.TeamMode && reg.TeamID != nil {
      return nil, errcode.ErrTeamNotAllowed
  }

  // 只在团队模式下更新团队分数
  if isCorrect && contest.TeamMode && reg.TeamID != nil {
      if err := s.updateScoreAndFirstBlood(...); err != nil {
          return nil, err
      }
  }
  ```

#### [L6] 缺少性能优化：多次数据库查询可合并
- **文件**：`code/backend/internal/module/contest/submission_service.go:29-59`
- **问题描述**：
  前置校验执行了 4 次独立查询：
  1. 查询 Contest
  2. 查询 ContestRegistration
  3. 查询 ContestChallenge
  4. 查询已存在的 Submission

  可以通过 JOIN 或预加载减少查询次数。
- **影响范围/风险**：
  - 响应时间较长
  - 数据库负载高
- **修正建议**：
  ```go
  // 使用子查询或 JOIN 一次性获取所有数据
  type ValidationResult struct {
      Contest            model.Contest
      Registration       model.ContestRegistration
      ContestChallenge   model.ContestChallenge
      ExistingSubmission *model.Submission
  }

  var result ValidationResult
  err := s.db.Raw(`
      SELECT
          c.*,
          cr.*,
          cc.*,
          s.id as existing_submission_id
      FROM contests c
      INNER JOIN contest_registrations cr ON cr.contest_id = c.id AND cr.user_id = ?
      INNER JOIN contest_challenges cc ON cc.contest_id = c.id AND cc.challenge_id = ?
      LEFT JOIN submissions s ON s.user_id = ? AND s.challenge_id = ? AND s.contest_id = ? AND s.is_correct = true
      WHERE c.id = ?
  `, userID, challengeID, userID, challengeID, contestID, contestID).Scan(&result).Error
  ```

  注：此优化需权衡代码复杂度，建议在性能测试后决定是否实施。

#### [L7] 错误码语义不准确：ErrAlreadySolved 应为 ErrAlreadySubmitted
- **文件**：`code/backend/internal/module/contest/submission_service.go:65`
- **问题描述**：
  ```go
  return nil, errcode.ErrAlreadySolved
  ```
  使用了 `ErrAlreadySolved`，但此错误码可能在其他模块（如普通靶场）中使用。竞赛场景应使用更具体的错误码。
- **影响范围/风险**：
  - 错误码语义混淆
  - 前端难以区分错误来源
- **修正建议**：
  ```go
  // pkg/errcode/errcode.go
  var (
      ErrContestChallengeAlreadySolved = New(14013, "该题目已在本场竞赛中解决", http.StatusConflict)
  )

  // 使用
  return nil, errcode.ErrContestChallengeAlreadySolved
  ```

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 4 |
| 🟡 中 | 6 |
| 🟢 低 | 7 |
| 合计 | 17 |

## 总体评价

本次变更实现了竞赛提交与计分的核心功能，代码结构清晰，基本遵循了项目分层规范。但存在以下严重问题需要立即修复：

**必须修复的问题**：
1. **并发安全**：首杀竞争条件（H1）和重复提交检查（H2）存在严重的并发安全问题，可能导致数据不一致和作弊
2. **事务完整性**：Submission 创建在事务外（H3），可能导致分数与提交记录不一致
3. **幂等性**：缺少幂等性保障（H4），网络重试可能导致重复计分

**建议优化的问题**：
- 补充完整的状态校验（M5、M6）
- 添加审计日志（L2）
- 优化数据库索引（L1）

**架构建议**：
- 考虑将首杀、计分、提交记录创建封装到一个原子事务中
- 建议引入分布式锁（Redis）或数据库行锁保证并发安全
- 补充单元测试，重点覆盖并发场景

修复高优先级问题后，代码质量可达到生产环境标准。
