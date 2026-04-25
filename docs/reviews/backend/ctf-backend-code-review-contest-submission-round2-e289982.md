# CTF 后端代码 Review（竞赛提交与计分 第 2 轮）：修复首杀竞争和重复提交问题

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | contest-submission |
| 轮次 | 第 2 轮（复审） |
| 审查范围 | 2 个 commit (e0932fb, e289982)，4 个文件，39 行新增，75 行删除 |
| 变更概述 | 修复 round 1 中的高优先级并发安全问题和参数校验问题 |
| 审查基准 | `/home/azhi/workspace/projects/ctf/CLAUDE.md` |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | 17（高 4，中 6，低 7） |

## Round 1 问题修复情况

### 🔴 高优先级问题修复状态

#### [H1] 首杀竞争条件 ✅ 已修复
- **修复方式**：使用乐观锁（L108-113）
  ```go
  result := tx.Model(&model.ContestChallenge{}).
      Where("contest_id = ? AND challenge_id = ? AND first_blood_by IS NULL",
            contestID, challengeID).
      Update("first_blood_by", userID)
  isFirstBlood := result.RowsAffected > 0
  ```
- **验证结果**：✅ 正确实现了 CAS 原子更新，只有一个并发请求能成功更新首杀

#### [H2] 重复提交检查不完整 ✅ 已修复
- **修复方式**：
  1. 事务内二次检查（L85-94）
  2. 数据库唯一索引（migration L12-14）
- **验证结果**：✅ 双重保障，既有应用层检查又有数据库约束

#### [H3] 事务边界错误 ✅ 已修复
- **修复方式**：将 Submission 创建移入事务（L84-147）
- **验证结果**：✅ 所有数据库操作（首杀、计分、提交记录）在同一事务中

#### [H4] 缺少幂等性保障 ❌ 未修复
- **状态**：未实现幂等键机制
- **风险**：客户端重试仍可能导致重复提交记录（虽然不会重复计分）

### 🟡 中优先级问题修复状态

#### [M1] Model/DTO 分离不完整 ⚠️ 部分修复
- **状态**：代码中使用了 `dto.SubmissionResp`，但未在本次 commit 中看到定义
- **假设**：可能在其他 commit 中已定义

#### [M2] 硬编码错误消息 ✅ 已修复
- **修复方式**：提取到 `internal/constants/messages.go`
- **验证结果**：✅ 符合规范

#### [M3] 首杀奖励计算精度问题 ✅ 已修复
- **修复方式**：使用 `math.Round`（L118）
- **验证结果**：✅ 正确处理浮点精度

#### [M4] 缺少输入校验 ✅ 已修复
- **修复方式**：Handler 中添加参数校验（L22-32）
- **验证结果**：✅ 正确校验并返回错误

#### [M5] 竞赛状态校验不完整 ✅ 已修复
- **修复方式**：添加 `contest.Status` 检查（L41-43）
- **验证结果**：✅ 符合需求

#### [M6] 缺少注册状态校验 ✅ 已修复
- **修复方式**：添加 `reg.Status` 检查（L61-63）
- **验证结果**：✅ 符合需求

### 🟢 低优先级问题修复状态

#### [L1] 数据库索引可优化 ❌ 未修复
- **状态**：仍使用单列索引 `idx_submissions_contest_id`
- **影响**：查询性能未达到最优

#### [L2] 缺少日志记录 ❌ 未修复
- **状态**：代码中无日志记录
- **影响**：问题排查和审计困难

#### [L3-L7] 其他低优先级问题 ❌ 未修复
- **状态**：配置默认值、Model 注释、团队模式校验、性能优化、错误码语义均未处理
- **说明**：低优先级问题可在后续迭代中优化

## 新发现问题清单

### 🔴 高优先级

#### [H5] 错误提交未防止并发重复插入
- **文件**：`code/backend/internal/module/contest/submission_service.go:152-167`
- **问题描述**：
  ```go
  } else {
      // 错误提交直接创建记录
      submission = &model.Submission{...}
      if err := s.db.Create(submission).Error; err != nil {
          return nil, err
      }
  }
  ```
  错误提交（`isCorrect = false`）在事务外创建，且没有唯一约束保护。用户可以通过并发提交错误 Flag 绕过频率限制，产生大量垃圾数据。
- **影响范围/风险**：
  - 恶意用户可以通过并发提交错误 Flag 刷数据库
  - 可能导致数据库性能下降
  - 影响统计数据准确性
- **修正建议**：
  添加提交频率限制（使用 Redis）：
  ```go
  // 在 ValidateFlag 之前检查频率
  rateLimitKey := fmt.Sprintf("submission:rate:%d:%d:%d", userID, contestID, challengeID)
  exists, _ := s.redis.Exists(ctx, rateLimitKey).Result()
  if exists > 0 {
      return nil, errcode.ErrSubmitTooFrequent
  }

  // 验证 Flag
  isCorrect, err := s.flagService.ValidateFlag(...)

  // 设置频率限制（例如 5 秒）
  if !isCorrect {
      s.redis.Set(ctx, rateLimitKey, "1", 5*time.Second)
  }
  ```

### 🟡 中优先级

#### [M7] 事务内查询 Challenge 可能导致死锁
- **文件**：`code/backend/internal/module/contest/submission_service.go:97-100`
- **问题描述**：
  ```go
  var chal model.Challenge
  if err := tx.First(&chal, challengeID).Error; err != nil {
      return err
  }
  ```
  在事务内查询 `challenges` 表，而该表可能被其他事务（如管理员修改题目分数）持有写锁，导致死锁风险。
- **影响范围/风险**：
  - 高并发时可能出现死锁
  - 事务回滚影响用户体验
- **修正建议**：
  将 Challenge 查询移到事务外，或使用 `cc.ContestScore` 作为唯一分数来源：
  ```go
  // 方案 1：事务外查询
  var chal model.Challenge
  if err := s.db.First(&chal, challengeID).Error; err != nil {
      return nil, err
  }

  baseScore := chal.Points
  if cc.ContestScore != nil {
      baseScore = *cc.ContestScore
  }

  if isCorrect {
      err = s.db.Transaction(func(tx *gorm.DB) error {
          // 使用事务外查询的 baseScore
          // ...
      })
  }

  // 方案 2：强制使用 ContestScore（推荐）
  if cc.ContestScore == nil {
      return nil, errcode.ErrContestScoreNotSet
  }
  baseScore := *cc.ContestScore
  ```

#### [M8] 唯一索引可能导致不友好的错误信息
- **文件**：`code/backend/migrations/000004_contest_submission.up.sql:12-14`
- **问题描述**：
  ```sql
  CREATE UNIQUE INDEX uk_submission_contest_user_challenge
  ON submissions(contest_id, user_id, challenge_id)
  WHERE is_correct = TRUE AND contest_id IS NOT NULL;
  ```
  当并发提交触发唯一索引冲突时，GORM 返回的是数据库原生错误（如 `duplicate key value violates unique constraint`），而不是业务错误码 `ErrContestChallengeAlreadySolved`。
- **影响范围/风险**：
  - 前端收到不友好的错误消息
  - 无法区分是并发冲突还是其他数据库错误
- **修正建议**：
  在 Service 层捕获唯一索引冲突并转换为业务错误：
  ```go
  if err := tx.Create(submission).Error; err != nil {
      // 检查是否是唯一索引冲突
      if strings.Contains(err.Error(), "uk_submission_contest_user_challenge") {
          return errcode.ErrContestChallengeAlreadySolved
      }
      return err
  }
  ```

### 🟢 低优先级

#### [L8] 变量作用域不清晰：finalScore 在事务外声明
- **文件**：`code/backend/internal/module/contest/submission_service.go:80`
- **问题描述**：
  ```go
  var submission *model.Submission
  var finalScore int

  if isCorrect {
      err = s.db.Transaction(func(tx *gorm.DB) error {
          // ...
          finalScore = baseScore  // 在闭包内赋值
          // ...
      })
  }
  ```
  `finalScore` 在事务外声明，但只在事务内赋值。如果事务失败，`finalScore` 为零值，可能导致返回错误的分数。
- **影响范围/风险**：
  - 代码可读性差
  - 潜在的零值 bug
- **修正建议**：
  在事务内声明或确保错误处理正确：
  ```go
  if isCorrect {
      var finalScore int
      err = s.db.Transaction(func(tx *gorm.DB) error {
          // ...
          finalScore = baseScore
          // ...
          return tx.Create(submission).Error
      })

      if err != nil {
          return nil, err  // 事务失败时直接返回，不会使用 finalScore
      }

      return &dto.SubmissionResp{
          Points: finalScore,
          // ...
      }, nil
  }
  ```

#### [L9] 缺少团队模式一致性校验
- **文件**：`code/backend/internal/module/contest/submission_service.go:123-132`
- **问题描述**：
  ```go
  if reg.TeamID != nil {
      if err := tx.Model(&model.Team{}).
          Where("id = ?", *reg.TeamID).
          Updates(...).Error; err != nil {
          return err
      }
  }
  ```
  未检查 `contest.TeamMode` 与 `reg.TeamID` 的一致性。个人赛中用户可能错误地关联了 TeamID，导致团队分数被更新。
- **影响范围/风险**：
  - 个人赛数据污染
  - 排行榜错误
- **修正建议**：
  ```go
  // 检查团队模式一致性
  if contest.TeamMode && reg.TeamID == nil {
      return errcode.ErrTeamRequired
  }
  if !contest.TeamMode && reg.TeamID != nil {
      return errcode.ErrTeamNotAllowed
  }

  // 只在团队模式下更新团队分数
  if contest.TeamMode && reg.TeamID != nil {
      if err := tx.Model(&model.Team{}).
          Where("id = ?", *reg.TeamID).
          Updates(...).Error; err != nil {
          return err
      }
  }
  ```

## 统计摘要

| 级别 | Round 1 | 已修复 | 未修复 | Round 2 新增 | Round 2 总计 |
|------|---------|--------|--------|--------------|--------------|
| 🔴 高 | 4 | 3 | 1 | 1 | 2 |
| 🟡 中 | 6 | 5 | 1 | 3 | 4 |
| 🟢 低 | 7 | 0 | 7 | 2 | 9 |
| 合计 | 17 | 8 | 9 | 6 | 15 |

## 总体评价

本轮修复成功解决了 round 1 中最严重的并发安全问题，代码质量显著提升：

**修复亮点**：
1. ✅ 使用乐观锁完美解决首杀竞争问题（H1）
2. ✅ 双重保障（事务内检查 + 唯一索引）防止重复计分（H2）
3. ✅ 事务边界正确，保证数据一致性（H3）
4. ✅ 补充了完整的状态校验（M5、M6）
5. ✅ 参数校验和错误消息规范化（M2、M4）

**仍需修复的关键问题**：
1. 🔴 **[H4] 幂等性保障**：虽然不会重复计分，但缺少幂等键会导致重复提交记录和用户体验问题
2. 🔴 **[H5] 错误提交并发防护**：需要添加频率限制防止恶意刷数据
3. 🟡 **[M7] 事务内查询死锁风险**：建议将 Challenge 查询移到事务外
4. 🟡 **[M8] 唯一索引错误转换**：需要将数据库错误转换为友好的业务错误

**架构建议**：
- 当前实现已基本满足生产环境要求，核心并发安全问题已解决
- 建议优先修复 H5（错误提交频率限制）和 M7（死锁风险）
- H4（幂等性）可在后续版本中通过添加 Redis 幂等键实现
- 低优先级问题可在性能测试和实际运行中逐步优化

**测试建议**：
- 编写并发测试用例，验证首杀和重复提交防护
- 压力测试验证事务性能和死锁情况
- 模拟网络超时重试场景，验证幂等性需求

修复 H5 和 M7 后，代码可达到生产环境标准。
