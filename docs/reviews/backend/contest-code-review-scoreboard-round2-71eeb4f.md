# Contest 代码 Review（scoreboard 第 2 轮）：排行榜功能 Round 1 修复验证

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | scoreboard |
| 轮次 | 第 2 轮（修复后复审） |
| 审查范围 | commits: f77267c, 6369f66, 52b0be0, 71eeb4f，4 个文件，169 行新增/46 行删除 |
| 变更概述 | 修复 round 1 审查中发现的 14 个问题（4 高 + 4 中 + 6 低） |
| 审查基准 | CTF 平台开发规范（CLAUDE.md）、round 1 review 报告 |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | 14（4 高 + 4 中 + 6 低） |

## 问题清单

### 🔴 高优先级

#### [H1] UpdateScore 冻结检查逻辑存在竞态条件风险
- **文件**：`code/backend/internal/module/contest/scoreboard_service.go:40-53`
- **问题描述**：
  ```go
  func (s *ScoreboardService) UpdateScore(ctx context.Context, contestID, teamID int64, points float64) error {
      contest, err := s.repo.FindByID(contestID)
      if err != nil {
          return err
      }

      isFrozen := contest.FreezeTime != nil && time.Now().After(*contest.FreezeTime)
      if isFrozen {
          // ... 拒绝更新
          return errors.New("排行榜已冻结，无法更新分数")
      }

      // ... 执行 ZIncrBy
  }
  ```
  - 检查冻结状态和执行更新之间存在时间窗口
  - 如果在检查通过后、更新前触发冻结，仍会更新实时榜单
  - 虽然实际场景中这个窗口很小，但理论上存在不一致风险

- **影响范围/风险**：
  - 极端情况下，冻结瞬间的提交可能绕过检查
  - 封榜时刻的数据一致性无法严格保证

- **修正建议**：
  方案 1（推荐）：在 FreezeScoreboard 中设置标记位，UpdateScore 检查标记位而非时间比较
  ```go
  // 在 Redis 中设置冻结标记
  func (s *ScoreboardService) FreezeScoreboard(ctx context.Context, contestID int64, minutesBeforeEnd int) error {
      // ... 现有逻辑 ...

      // 设置冻结标记（原子操作）
      freezeFlagKey := redis.ContestFreezeFlag(contestID)
      s.redis.Set(ctx, freezeFlagKey, "1", 0)

      // ... 创建快照 ...
  }

  func (s *ScoreboardService) UpdateScore(ctx context.Context, contestID, teamID int64, points float64) error {
      // 原子检查冻结标记
      freezeFlagKey := redis.ContestFreezeFlag(contestID)
      isFrozen := s.redis.Exists(ctx, freezeFlagKey).Val() > 0
      if isFrozen {
          return errors.New("排行榜已冻结，无法更新分数")
      }

      // ... 执行更新 ...
  }
  ```

  方案 2：使用 Lua 脚本保证原子性（更复杂但更严格）

#### [H2] UnfreezeScoreboard 缺少竞赛状态校验
- **文件**：`code/backend/internal/module/contest/scoreboard_service.go:219-233`
- **问题描述**：
  ```go
  func (s *ScoreboardService) UnfreezeScoreboard(ctx context.Context, contestID int64) error {
      contest, err := s.repo.FindByID(contestID)
      if err != nil {
          return err
      }

      contest.FreezeTime = nil
      // ❌ 缺少：未检查竞赛是否已结束
      // ❌ 缺少：未检查当前是否处于冻结状态
  }
  ```
  - FreezeScoreboard 有状态检查（189-195 行），但 UnfreezeScoreboard 没有
  - 已结束的竞赛仍可解冻
  - 未冻结的竞赛也可以执行解冻操作（虽然无害但不规范）

- **影响范围/风险**：
  - 业务逻辑不完整，可能导致非预期操作
  - 已结束竞赛的数据可能被意外修改

- **修正建议**：
  ```go
  func (s *ScoreboardService) UnfreezeScoreboard(ctx context.Context, contestID int64) error {
      contest, err := s.repo.FindByID(contestID)
      if err != nil {
          return err
      }

      // 检查竞赛状态
      if contest.Status == model.ContestStatusFinished {
          return errors.New("竞赛已结束，无法解冻")
      }

      // 检查是否已冻结
      if contest.FreezeTime == nil {
          return errors.New("排行榜未冻结")
      }

      contest.FreezeTime = nil

      dstKey := redis.RankContestFrozenKey(contestID)
      s.redis.Del(ctx, dstKey)

      s.logger.Info("scoreboard unfrozen", zap.Int64("contest_id", contestID))

      return s.repo.Update(contest)
  }
  ```

### 🟡 中优先级

#### [M1] memberToTeamID 错误处理不当，可能返回零值
- **文件**：`code/backend/internal/module/contest/scoreboard_service.go:239-242`
- **问题描述**：
  ```go
  func memberToTeamID(member string) int64 {
      id, _ := strconv.ParseInt(member, 10, 64)  // ❌ 忽略错误
      return id
  }
  ```
  - 如果 Redis 中存储了非法数据（非数字字符串），ParseInt 失败会返回 0
  - 调用方无法区分"解析失败"和"teamID 真的是 0"
  - 虽然正常情况下不会出现，但缺少防御性编程

- **影响范围/风险**：
  - Redis 数据损坏时，排行榜查询会返回 teamID=0 的记录
  - 可能导致前端显示异常或查询不存在的队伍

- **修正建议**：
  方案 1：记录日志并返回特殊值
  ```go
  func memberToTeamID(member string) int64 {
      id, err := strconv.ParseInt(member, 10, 64)
      if err != nil {
          // 记录错误但不中断流程（防御性编程）
          // 注意：这里无法访问 logger，需要调整设计
          return -1  // 返回非法 ID，调用方可识别
      }
      return id
  }
  ```

  方案 2（推荐）：在调用方处理
  ```go
  // 保持当前实现，但在 GetScoreboard 中添加校验
  for i, z := range results {
      teamID := memberToTeamID(z.Member.(string))
      if teamID <= 0 {
          s.logger.Error("invalid team_id in redis",
              zap.String("member", z.Member.(string)),
              zap.Int64("contest_id", contestID))
          continue  // 跳过非法记录
      }
      teamIDs[i] = teamID
  }
  ```

#### [M2] CreateSnapshot 使用 Redis COPY 命令，需确认 Redis 版本兼容性
- **文件**：`code/backend/internal/module/contest/scoreboard_service.go:178-180`
- **问题描述**：
  ```go
  func (s *ScoreboardService) CreateSnapshot(ctx context.Context, srcKey, dstKey string) error {
      return s.redis.Copy(ctx, srcKey, dstKey, 0, false).Err()
  }
  ```
  - Redis COPY 命令在 6.2.0 版本引入（2021 年 2 月）
  - 如果项目使用 Redis < 6.2，此命令会失败
  - 代码中未说明最低 Redis 版本要求

- **影响范围/风险**：
  - 低版本 Redis 环境下冻结功能完全失效
  - 错误信息不明确，难以排查

- **修正建议**：
  方案 1：在文档中明确 Redis 版本要求（最简单）

  方案 2：使用兼容性更好的 ZUNIONSTORE（推荐）
  ```go
  func (s *ScoreboardService) CreateSnapshot(ctx context.Context, srcKey, dstKey string) error {
      // ZUNIONSTORE 在 Redis 2.0+ 就支持
      // 将 srcKey 与自身做并集，结果写入 dstKey
      return s.redis.ZUnionStore(ctx, dstKey, &redislib.ZStore{
          Keys:    []string{srcKey},
          Weights: []float64{1},
      }).Err()
  }
  ```

### 🟢 低优先级

#### [L1] CalculateDynamicScore 方法仍未被调用
- **文件**：`code/backend/internal/module/contest/scoreboard_service.go:172-175`
- **问题描述**：
  - Round 1 中标记为 [L5]，本轮未修复
  - 该方法及相关配置（baseScore/minScore/decay）已实现，但从未被调用
  - 不清楚是否应该在题目提交时使用动态计分

- **影响范围/风险**：
  - 如果应该使用但未使用，则计分逻辑不完整
  - 如果不需要，则配置和代码都是冗余的

- **修正建议**：
  - 方案 1：如果确认不需要动态计分，删除该方法和 ContestConfig 中的相关字段
  - 方案 2：如果需要，在题目提交逻辑中调用该方法（需要明确业务需求）

#### [L2] 日志记录不完整：GetScoreboard 缺少关键操作日志
- **文件**：`code/backend/internal/module/contest/scoreboard_service.go:75-133`
- **问题描述**：
  - UpdateScore、FreezeScoreboard、UnfreezeScoreboard 都有日志
  - GetScoreboard 作为高频查询接口，缺少访问日志
  - 无法统计排行榜查询频率和性能

- **影响范围/风险**：
  - 生产环境性能问题难以排查
  - 无法监控排行榜访问热度

- **修正建议**：
  ```go
  func (s *ScoreboardService) GetScoreboard(ctx context.Context, contestID int64) (*dto.ScoreboardResp, error) {
      s.logger.Debug("fetching scoreboard",
          zap.Int64("contest_id", contestID))

      // ... 现有逻辑 ...

      s.logger.Info("scoreboard fetched",
          zap.Int64("contest_id", contestID),
          zap.Int("team_count", len(items)),
          zap.Bool("is_frozen", isFrozen))

      return &dto.ScoreboardResp{...}, nil
  }
  ```

#### [L3] FreezeScoreboard 快照失败时数据库已更新，存在不一致风险
- **文件**：`code/backend/internal/module/contest/scoreboard_service.go:197-215`
- **问题描述**：
  ```go
  func (s *ScoreboardService) FreezeScoreboard(...) error {
      // ... 计算 freezeTime ...
      contest.FreezeTime = &freezeTime  // ← 先修改内存对象

      // 创建快照
      if err := s.CreateSnapshot(ctx, srcKey, dstKey); err != nil {
          s.logger.Error("failed to create scoreboard snapshot", ...)
          return err  // ← 快照失败直接返回
      }

      return s.repo.Update(contest)  // ← 数据库更新在最后
  }
  ```
  - 如果快照创建失败，数据库不会更新（这是对的）
  - 但如果快照成功、数据库更新失败，Redis 中有快照但 DB 中 FreezeTime 为空
  - 虽然概率很低，但存在不一致风险

- **影响范围/风险**：
  - 数据库更新失败时，Redis 快照已创建但无法回滚
  - GetScoreboard 会认为未冻结，但 Redis 中存在快照 key

- **修正建议**：
  ```go
  func (s *ScoreboardService) FreezeScoreboard(...) error {
      // ... 现有逻辑 ...

      // 先更新数据库
      if err := s.repo.Update(contest); err != nil {
          s.logger.Error("failed to update contest freeze time", zap.Error(err))
          return err
      }

      // 再创建快照
      srcKey := redis.RankContestTeamKey(contestID)
      dstKey := redis.RankContestFrozenKey(contestID)
      if err := s.CreateSnapshot(ctx, srcKey, dstKey); err != nil {
          s.logger.Error("failed to create scoreboard snapshot", zap.Error(err))

          // 回滚数据库（可选，取决于业务需求）
          contest.FreezeTime = nil
          s.repo.Update(contest)

          return err
      }

      s.logger.Info("scoreboard frozen", ...)
      return nil
  }
  ```

#### [L4] GetScoreboard 中 teamMap 查询结果未校验完整性
- **文件**：`code/backend/internal/module/contest/scoreboard_service.go:106-125`
- **问题描述**：
  ```go
  teams, err := s.repo.FindTeamsByIDs(teamIDs)
  if err != nil {
      s.logger.Error("failed to fetch teams", zap.Error(err))
      return nil, err
  }

  teamMap := make(map[int64]*model.Team)
  for _, team := range teams {
      teamMap[team.ID] = team
  }

  // ... 使用 teamMap ...
  team := teamMap[teamID]
  if team == nil {
      s.logger.Warn("team not found", zap.Int64("team_id", teamID))
  }
  ```
  - 如果 Redis 中有 100 个 teamID，但数据库只返回 95 个（5 个队伍被删除）
  - 代码会记录 5 次 Warn 日志，但继续处理
  - 排行榜会显示 5 个 teamID=0, teamName="" 的记录

- **影响范围/风险**：
  - 数据不一致时，排行榜显示异常
  - 虽然有日志记录，但前端体验不佳

- **修正建议**：
  方案 1：跳过缺失的队伍（推荐）
  ```go
  items := make([]*dto.ScoreboardItem, 0, len(results))
  rank := 1
  for i, z := range results {
      teamID := teamIDs[i]
      team := teamMap[teamID]
      if team == nil {
          s.logger.Warn("team not found, skipping", zap.Int64("team_id", teamID))
          continue  // 跳过缺失的队伍
      }
      items = append(items, toScoreboardItem(team, z.Score, rank))
      rank++
  }
  ```

  方案 2：返回占位符（保持排名连续性）
  ```go
  // 当前实现已经是这种方式，但可以改进占位符内容
  func toScoreboardItem(team *model.Team, score float64, rank int) *dto.ScoreboardItem {
      item := &dto.ScoreboardItem{
          Score: score,
          Rank:  rank,
      }
      if team != nil {
          item.TeamID = team.ID
          item.TeamName = team.Name
      } else {
          item.TeamID = -1  // 明确标记为无效
          item.TeamName = "[已删除]"
      }
      return item
  }
  ```

## 统计摘要

| 级别 | 数量 | Round 1 数量 | 变化 |
|------|------|--------------|------|
| 🔴 高 | 2 | 4 | -2（H1/H2 已修复，新增 2 个） |
| 🟡 中 | 2 | 4 | -2（M1-M4 已修复，新增 2 个） |
| 🟢 低 | 4 | 6 | -2（L1/L2/L6 已修复，L5 未修复） |
| 合计 | 8 | 14 | -6 |

## Round 1 问题修复情况

### ✅ 已完全修复（8 个）

| 问题编号 | 问题描述 | 修复 commit | 验证结果 |
|---------|---------|------------|---------|
| [H1] | teamID 转换逻辑数据损坏 | f77267c | ✅ 使用 strconv.FormatInt/ParseInt，正确实现 |
| [H2] | 冻结榜单逻辑错误 | f77267c | ✅ 实现 CreateSnapshot，冻结时创建快照 |
| [H3] | 分数更新无冻结检查 | f77267c | ✅ UpdateScore 中添加冻结检查和日志 |
| [H4] | 计分参数硬编码 | 6369f66 | ✅ 添加 ContestConfig，通过配置注入 |
| [M1] | Model/DTO 分离不彻底 | 52b0be0 | ✅ 添加 toScoreboardItem 转换函数 |
| [M2] | 缺少输入校验 | 52b0be0 | ✅ Handler 层添加 contestID <= 0 校验 |
| [M3] | 冻结逻辑在 Handler 层 | 52b0be0 | ✅ 移至 Service 层，Handler 只做参数校验 |
| [M4] | 缺少竞赛状态校验 | 52b0be0 | ✅ FreezeScoreboard 中检查竞赛状态 |

### ✅ 已完全修复（续）

| 问题编号 | 问题描述 | 修复 commit | 验证结果 |
|---------|---------|------------|---------|
| [L1] | SolveTime 字段未使用 | 71eeb4f | ✅ 已移除该字段 |
| [L2] | 错误消息硬编码 | 71eeb4f | ✅ 提取为常量 ErrMsgInvalidContestID |
| [L3] | N+1 查询问题 | 52b0be0 | ✅ 使用 FindTeamsByIDs 批量查询 |
| [L4] | 缺少日志记录 | f77267c | ✅ 添加 logger 字段和关键操作日志 |
| [L6] | Redis Key 注释不一致 | 71eeb4f | ✅ 修正为 ZSET 数据结构 |

### ⚠️ 未修复（1 个）

| 问题编号 | 问题描述 | 原因 |
|---------|---------|------|
| [L5] | CalculateDynamicScore 未被调用 | 需要明确业务需求后再决定是否删除或使用 |

## 总体评价

Round 1 的 14 个问题中，13 个已修复，1 个（L5）待明确需求。修复质量整体良好，但引入了 2 个新的高优先级问题和 2 个中优先级问题。

**修复亮点**：
1. ✅ H1（teamID 转换）修复彻底，使用正确的字符串转换方式
2. ✅ H2（冻结快照）正确实现，使用 Redis COPY 命令
3. ✅ H4（配置外部化）完整实现，包括 Config 结构、默认值、构造函数注入
4. ✅ M3（分层职责）重构到位，Handler 层职责清晰
5. ✅ L3（N+1 查询）优化有效，使用批量查询 + Map 索引

**新引入的问题**：
1. 🔴 H1（新）：UpdateScore 冻结检查存在竞态条件（虽然概率极低）
2. 🔴 H2（新）：UnfreezeScoreboard 缺少状态校验（与 FreezeScoreboard 不一致）
3. 🟡 M1（新）：memberToTeamID 错误处理不当
4. 🟡 M2（新）：Redis COPY 命令版本兼容性问题

**代码质量改进**：
- 分层架构更清晰，Handler/Service 职责分离到位
- 日志记录完善，关键操作都有审计日志
- 配置外部化符合项目规范
- 性能优化有效（批量查询）

**建议修复优先级**：
1. 🔴 H2（UnfreezeScoreboard 状态校验）- 简单且重要
2. 🟡 M2（Redis 版本兼容性）- 使用 ZUNIONSTORE 替代 COPY
3. 🔴 H1（竞态条件）- 使用 Redis 标记位方案
4. 🟡 M1（错误处理）- 在调用方添加校验
5. 🟢 L3（事务顺序）- 调整数据库和快照的执行顺序
6. 🟢 L4（数据完整性）- 跳过缺失的队伍记录

**后续建议**：
- 明确 CalculateDynamicScore 的使用场景，决定保留或删除
- 添加集成测试，覆盖冻结/解冻场景
- 压测排行榜查询性能（100+ 队伍）
- 补充 Redis 版本要求文档
