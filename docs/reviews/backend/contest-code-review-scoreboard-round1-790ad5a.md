# Contest 代码 Review（scoreboard 第 1 轮）：竞赛排行榜功能实现

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | scoreboard |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | commit 790ad5a，6 个文件，312 行新增 |
| 变更概述 | 实现竞赛排行榜功能，包括分数更新、榜单查询、冻结/解冻操作 |
| 审查基准 | CTF 平台开发规范（CLAUDE.md） |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | - |

## 问题清单

### 🔴 高优先级

#### [H1] teamIDToMember/memberToTeamID 实现严重错误，导致数据损坏
- **文件**：`code/backend/internal/module/contest/scoreboard_service.go:121-127`
- **问题描述**：
  ```go
  func teamIDToMember(teamID int64) string {
      return string(rune(teamID))  // ❌ 错误：将 int64 转为 rune 再转 string
  }

  func memberToTeamID(member string) int64 {
      return int64([]rune(member)[0])  // ❌ 错误：只取第一个字符的 Unicode 码点
  }
  ```
  这会导致：
  - teamID=1000 转换为 Unicode 字符 U+03E8（希腊字母 Ψ），再转回来变成 1000
  - teamID=100000 超出 Unicode 范围（最大 U+10FFFF），会被截断或产生无效字符
  - 不同的 teamID 可能映射到相同的字符（数据冲突）

- **影响范围/风险**：
  - 数据完整性破坏：大 teamID 无法正确存储和检索
  - 排行榜数据错乱：多个队伍可能被识别为同一个队伍
  - 生产环境数据损坏：一旦写入 Redis，历史数据无法恢复

- **修正建议**：
  ```go
  func teamIDToMember(teamID int64) string {
      return strconv.FormatInt(teamID, 10)  // 转为十进制字符串
  }

  func memberToTeamID(member string) int64 {
      id, _ := strconv.ParseInt(member, 10, 64)
      return id
  }
  ```

#### [H2] 冻结榜单逻辑错误：注释与实现不一致
- **文件**：`code/backend/internal/pkg/redis/keys.go:226-229`
- **问题描述**：
  - 注释声明：`数据结构: STRING (JSON 快照)`
  - 实际使用：`scoreboard_service.go:52-56` 使用 `ZRevRangeWithScores` 读取，说明实际是 ZSET
  - 代码逻辑：冻结时检查 `RankContestFrozenKey` 是否存在，不存在则回退到实时榜单
  - **缺失**：没有任何代码在冻结时创建快照（应在 `FreezeScoreboard` 时执行 `COPY` 命令）

- **影响范围/风险**：
  - 冻结功能完全失效：冻结后仍显示实时分数
  - 违反竞赛公平性：封榜期间选手可看到实时排名变化

- **修正建议**：
  在 `Handler.FreezeScoreboard` 中添加快照逻辑：
  ```go
  func (h *Handler) FreezeScoreboard(c *gin.Context) {
      // ... 现有代码 ...

      // 创建快照
      srcKey := redis.RankContestTeamKey(contestID)
      dstKey := redis.RankContestFrozenKey(contestID)
      if err := h.scoreboardService.CreateSnapshot(c.Request.Context(), srcKey, dstKey); err != nil {
          response.FromError(c, err)
          return
      }

      // ... 更新数据库 ...
  }
  ```

  在 `ScoreboardService` 中添加方法：
  ```go
  func (s *ScoreboardService) CreateSnapshot(ctx context.Context, srcKey, dstKey string) error {
      // 使用 COPY 命令（Redis 6.2+）或 ZUNIONSTORE
      return s.redis.Copy(ctx, srcKey, dstKey, 0, false).Err()
  }
  ```

#### [H3] 并发安全问题：分数更新无原子性保障
- **文件**：`code/backend/internal/module/contest/scoreboard_service.go:33-36`
- **问题描述**：
  ```go
  func (s *ScoreboardService) UpdateScore(ctx context.Context, contestID, teamID int64, points float64) error {
      key := redis.RankContestTeamKey(contestID)
      return s.redis.ZIncrBy(ctx, key, points, teamIDToMember(teamID)).Err()
  }
  ```
  - 虽然 `ZIncrBy` 本身是原子的，但缺少冻结状态检查
  - 冻结期间仍可更新实时榜单，导致解冻后数据不一致

- **影响范围/风险**：
  - 封榜期间的提交会更新实时榜单
  - 解冻后显示的分数与封榜时刻不一致

- **修正建议**：
  ```go
  func (s *ScoreboardService) UpdateScore(ctx context.Context, contestID, teamID int64, points float64) error {
      // 检查是否冻结
      contest, err := s.repo.FindByID(contestID)
      if err != nil {
          return err
      }

      isFrozen := contest.FreezeTime != nil && time.Now().After(*contest.FreezeTime)

      key := redis.RankContestTeamKey(contestID)
      if err := s.redis.ZIncrBy(ctx, key, points, teamIDToMember(teamID)).Err(); err != nil {
          return err
      }

      // 冻结期间不更新快照
      if !isFrozen {
          // 可选：同步更新冻结快照（如果存在）
      }

      return nil
  }
  ```

#### [H4] 动态计分公式硬编码，违反配置外部化原则
- **文件**：`code/backend/internal/module/contest/scoreboard_service.go:22-30`
- **问题描述**：
  ```go
  func NewScoreboardService(repo *Repository, redis *redislib.Client) *ScoreboardService {
      return &ScoreboardService{
          repo:      repo,
          redis:     redis,
          baseScore: 1000,   // ❌ 硬编码
          minScore:  100,    // ❌ 硬编码
          decay:     0.9,    // ❌ 硬编码
      }
  }
  ```

- **影响范围/风险**：
  - 不同竞赛无法使用不同计分规则
  - 调整参数需要重新编译部署
  - 违反项目规范（CLAUDE.md 禁止硬编码）

- **修正建议**：
  1. 在 `config.go` 中添加配置：
  ```go
  type ContestConfig struct {
      BaseScore float64 `mapstructure:"base_score"`
      MinScore  float64 `mapstructure:"min_score"`
      Decay     float64 `mapstructure:"decay"`
  }
  ```

  2. 修改构造函数：
  ```go
  func NewScoreboardService(repo *Repository, redis *redislib.Client, cfg *config.ContestConfig) *ScoreboardService {
      return &ScoreboardService{
          repo:      repo,
          redis:     redis,
          baseScore: cfg.BaseScore,
          minScore:  cfg.MinScore,
          decay:     cfg.Decay,
      }
  }
  ```

### 🟡 中优先级

#### [M1] Model/DTO 分离不彻底：缺少 Model → DTO 转换层
- **文件**：`code/backend/internal/module/contest/scoreboard_service.go:65-78`
- **问题描述**：
  ```go
  for i, z := range results {
      teamID := memberToTeamID(z.Member.(string))
      team, _ := s.repo.FindTeamByID(teamID)  // 返回 *model.Team

      teamName := ""
      if team != nil {
          teamName = team.Name  // 直接访问 Model 字段
      }

      items = append(items, &dto.ScoreboardItem{...})
  }
  ```
  - Service 层直接访问 Model 字段，应通过转换函数封装
  - 错误被静默忽略（`team, _ := ...`），应记录日志

- **影响范围/风险**：
  - 违反分层架构规范
  - 查询失败时无法追踪问题

- **修正建议**：
  ```go
  func (s *ScoreboardService) GetScoreboard(ctx context.Context, contestID int64) (*dto.ScoreboardResp, error) {
      // ... 现有代码 ...

      for i, z := range results {
          teamID := memberToTeamID(z.Member.(string))
          team, err := s.repo.FindTeamByID(teamID)
          if err != nil {
              s.logger.Warn("failed to fetch team", zap.Int64("team_id", teamID), zap.Error(err))
              // 继续处理，使用默认值
          }

          items = append(items, toScoreboardItem(team, z.Score, i+1))
      }

      return &dto.ScoreboardResp{...}, nil
  }

  func toScoreboardItem(team *model.Team, score float64, rank int) *dto.ScoreboardItem {
      item := &dto.ScoreboardItem{
          Score: score,
          Rank:  rank,
      }
      if team != nil {
          item.TeamID = team.ID
          item.TeamName = team.Name
      }
      return item
  }
  ```

#### [M2] 缺少输入校验：contestID 和 teamID 未验证有效性
- **文件**：`code/backend/internal/module/contest/handler.go:25-29, 42-46, 73-77`
- **问题描述**：
  ```go
  contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
  if err != nil {
      response.InvalidParams(c, "无效的竞赛ID")
      return
  }
  // ❌ 缺少：contestID <= 0 的检查
  ```

- **影响范围/风险**：
  - 负数或零 ID 可能导致数据库查询异常
  - Redis Key 生成异常

- **修正建议**：
  ```go
  contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
  if err != nil || contestID <= 0 {
      response.InvalidParams(c, "无效的竞赛ID")
      return
  }
  ```

#### [M3] 冻结时间计算逻辑应在 Service 层，而非 Handler 层
- **文件**：`code/backend/internal/module/contest/handler.go:60-62`
- **问题描述**：
  ```go
  freezeTime := contest.EndTime.Add(-time.Duration(req.MinutesBeforeEnd) * time.Minute)
  contest.FreezeTime = &freezeTime
  ```
  - Handler 层包含业务逻辑（时间计算）
  - 违反分层职责：Handler 应只做参数校验和响应封装

- **影响范围/风险**：
  - 业务逻辑分散，难以测试和复用

- **修正建议**：
  将逻辑移到 Service 层：
  ```go
  // Handler
  func (h *Handler) FreezeScoreboard(c *gin.Context) {
      // ... 参数解析 ...

      if err := h.scoreboardService.FreezeScoreboard(c.Request.Context(), contestID, req.MinutesBeforeEnd); err != nil {
          response.FromError(c, err)
          return
      }

      response.Success(c, gin.H{"message": "排行榜已冻结"})
  }

  // Service
  func (s *ScoreboardService) FreezeScoreboard(ctx context.Context, contestID int64, minutesBeforeEnd int) error {
      contest, err := s.repo.FindByID(contestID)
      if err != nil {
          return err
      }

      freezeTime := contest.EndTime.Add(-time.Duration(minutesBeforeEnd) * time.Minute)
      contest.FreezeTime = &freezeTime

      // 创建快照
      // ...

      return s.repo.Update(contest)
  }
  ```

#### [M4] 缺少竞赛状态校验：已结束的竞赛不应允许冻结/解冻
- **文件**：`code/backend/internal/module/contest/handler.go:42-69, 72-92`
- **问题描述**：
  - `FreezeScoreboard` 和 `UnfreezeScoreboard` 未检查竞赛状态
  - 已结束的竞赛仍可执行冻结操作

- **影响范围/风险**：
  - 业务逻辑不完整
  - 可能导致数据不一致

- **修正建议**：
  ```go
  func (h *Handler) FreezeScoreboard(c *gin.Context) {
      // ... 现有代码 ...

      contest, err := h.repo.FindByID(contestID)
      if err != nil {
          response.FromError(c, err)
          return
      }

      if contest.Status == model.ContestStatusFinished {
          response.Error(c, errcode.ErrBadRequest("竞赛已结束，无法冻结"))
          return
      }

      if time.Now().After(contest.EndTime) {
          response.Error(c, errcode.ErrBadRequest("竞赛已结束"))
          return
      }

      // ... 继续处理 ...
  }
  ```

### 🟢 低优先级

#### [L1] SolveTime 字段未使用，应移除或实现
- **文件**：`code/backend/internal/dto/contest.go:9`
- **问题描述**：
  ```go
  type ScoreboardItem struct {
      TeamID    int64   `json:"team_id"`
      TeamName  string  `json:"team_name"`
      Score     float64 `json:"score"`
      Rank      int     `json:"rank"`
      SolveTime int64   `json:"solve_time"`  // ❌ 从未赋值
  }
  ```

- **影响范围/风险**：
  - API 返回字段始终为 0，可能误导前端
  - 如果后续需要该字段，需要重新设计数据结构

- **修正建议**：
  - 方案 1：移除该字段（如果不需要）
  - 方案 2：在 Redis ZSET 中存储额外信息（使用 HASH 结构）：
  ```go
  // 存储：ctf:rank:contest:1:team -> ZSET (score)
  //      ctf:rank:contest:1:team:meta:{teamID} -> HASH (solve_time, last_submit_time)
  ```

#### [L2] 错误消息硬编码，应提取为常量
- **文件**：`code/backend/internal/module/contest/handler.go:28, 45, 76`
- **问题描述**：
  ```go
  response.InvalidParams(c, "无效的竞赛ID")  // 重复出现 3 次
  ```

- **影响范围/风险**：
  - 违反 DRY 原则
  - 修改错误消息需要多处改动

- **修正建议**：
  ```go
  // internal/module/contest/errors.go
  package contest

  const (
      ErrMsgInvalidContestID = "无效的竞赛ID"
      ErrMsgContestNotFound  = "竞赛不存在"
  )
  ```

#### [L3] GetScoreboard 性能问题：N+1 查询
- **文件**：`code/backend/internal/module/contest/scoreboard_service.go:65-78`
- **问题描述**：
  ```go
  for i, z := range results {
      teamID := memberToTeamID(z.Member.(string))
      team, _ := s.repo.FindTeamByID(teamID)  // ❌ 循环内查询数据库
      // ...
  }
  ```

- **影响范围/风险**：
  - 100 个队伍 = 100 次数据库查询
  - 排行榜加载缓慢

- **修正建议**：
  ```go
  func (s *ScoreboardService) GetScoreboard(ctx context.Context, contestID int64) (*dto.ScoreboardResp, error) {
      // ... 获取 Redis 数据 ...

      // 批量查询队伍信息
      teamIDs := make([]int64, len(results))
      for i, z := range results {
          teamIDs[i] = memberToTeamID(z.Member.(string))
      }

      teams, err := s.repo.FindTeamsByIDs(teamIDs)  // 一次查询
      if err != nil {
          return nil, err
      }

      teamMap := make(map[int64]*model.Team)
      for _, team := range teams {
          teamMap[team.ID] = team
      }

      // 构建响应
      for i, z := range results {
          teamID := teamIDs[i]
          team := teamMap[teamID]
          items = append(items, toScoreboardItem(team, z.Score, i+1))
      }

      return &dto.ScoreboardResp{...}, nil
  }
  ```

  在 Repository 中添加：
  ```go
  func (r *Repository) FindTeamsByIDs(ids []int64) ([]*model.Team, error) {
      var teams []*model.Team
      err := r.db.Where("id IN ?", ids).Find(&teams).Error
      return teams, err
  }
  ```

#### [L4] 缺少日志记录
- **文件**：`code/backend/internal/module/contest/scoreboard_service.go`（整个文件）
- **问题描述**：
  - 关键操作（分数更新、冻结/解冻）无日志
  - 错误场景无法追踪

- **影响范围/风险**：
  - 生产环境问题难以排查
  - 无法审计操作记录

- **修正建议**：
  ```go
  func (s *ScoreboardService) UpdateScore(ctx context.Context, contestID, teamID int64, points float64) error {
      s.logger.Info("updating team score",
          zap.Int64("contest_id", contestID),
          zap.Int64("team_id", teamID),
          zap.Float64("points", points),
      )

      // ... 执行更新 ...

      if err != nil {
          s.logger.Error("failed to update score",
              zap.Int64("contest_id", contestID),
              zap.Int64("team_id", teamID),
              zap.Error(err),
          )
          return err
      }

      return nil
  }
  ```

#### [L5] CalculateDynamicScore 方法未被调用
- **文件**：`code/backend/internal/module/contest/scoreboard_service.go:116-119`
- **问题描述**：
  ```go
  func (s *ScoreboardService) CalculateDynamicScore(solveCount int) float64 {
      score := s.baseScore * math.Pow(s.decay, float64(solveCount))
      return math.Max(s.minScore, score)
  }
  ```
  - 该方法在整个代码库中未被调用
  - 不清楚何时应该使用动态计分

- **影响范围/风险**：
  - 死代码，增加维护负担
  - 如果应该使用但未使用，则计分逻辑不完整

- **修正建议**：
  - 方案 1：如果不需要，删除该方法和相关配置
  - 方案 2：在题目提交时调用该方法计算分数：
  ```go
  // 在提交处理逻辑中
  solveCount := s.getSolveCount(challengeID)  // 获取已解决人数
  points := s.CalculateDynamicScore(solveCount)
  s.UpdateScore(ctx, contestID, teamID, points)
  ```

#### [L6] Redis Key 注释与实际数据结构不一致
- **文件**：`code/backend/internal/pkg/redis/keys.go:227`
- **问题描述**：
  - 注释：`数据结构: STRING (JSON 快照)`
  - 实际：代码中使用 ZSET 操作（`ZRevRangeWithScores`）

- **影响范围/风险**：
  - 文档与实现不一致，误导开发者

- **修正建议**：
  修改注释为：
  ```go
  // RankContestFrozenKey 封榜后的排行榜快照
  // 数据结构: ZSET (score=total_score, member=team_id) | TTL: 至竞赛结束
  func RankContestFrozenKey(contestID int64) string {
      return withNS(fmt.Sprintf(keyRankContestFrozenPrefix, contestID))
  }
  ```

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 4 |
| 🟡 中 | 4 |
| 🟢 低 | 6 |
| 合计 | 14 |

## 总体评价

本次实现完成了排行榜的基本功能框架，但存在多个严重问题需要修复：

**必须立即修复的问题**：
1. teamID 转换逻辑完全错误，会导致数据损坏（H1）
2. 冻结功能未实现快照逻辑，功能失效（H2）
3. 动态计分参数硬编码，违反项目规范（H4）

**架构层面**：
- Model/DTO 分离不彻底，Service 层直接访问 Model 字段
- Handler 层包含业务逻辑（时间计算），应移到 Service 层
- 缺少统一的错误处理和日志记录

**性能问题**：
- N+1 查询问题会在队伍数量增加时严重影响性能

**建议修复顺序**：
1. 优先修复 H1（数据损坏风险）
2. 修复 H2（核心功能失效）
3. 修复 H4（配置外部化）
4. 解决 M1-M4（架构一致性）
5. 优化 L3（性能问题）
6. 完善 L1、L2、L4、L5、L6（代码质量）

修复后建议进行集成测试，重点验证：
- 大 teamID（> 100000）的正确性
- 冻结/解冻功能的完整性
- 并发提交时的分数一致性
- 100+ 队伍时的查询性能
