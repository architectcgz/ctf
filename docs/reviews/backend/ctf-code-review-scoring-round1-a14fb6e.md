# CTF 平台代码 Review（scoring 第 1 轮）：实时计分功能实现

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | scoring |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | 4 个 commit (f9389af, 56f2999, 1451899, a14fb6e)，4 个文件，219 行新增 |
| 变更概述 | 实现用户得分模型、计分服务、排行榜功能，并在 Flag 提交成功后异步更新得分 |
| 审查基准 | docs/architecture/*.md, CLAUDE.md |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | - |

## 问题清单

### 🔴 高优先级

#### [H1] 并发竞争：重复计分风险
- **文件**：`code/backend/internal/module/practice/score_service.go:56-97`
- **问题描述**：`UpdateUserScore` 方法存在严重的并发竞争问题。当用户短时间内提交多个正确 Flag 时，多个 goroutine 会同时调用此方法，导致：
  1. 多次查询 submissions 表，可能读到不一致的数据
  2. 多次计算总分，结果可能不准确
  3. 多次执行 UPSERT，最后一次写入可能覆盖前面的正确结果
- **影响范围/风险**：用户得分可能不准确，排行榜数据错误，影响比赛公平性
- **修正建议**：使用分布式锁保护更新逻辑
```go
func (s *ScoreService) UpdateUserScore(userID int64) error {
    ctx := context.Background()
    lockKey := fmt.Sprintf("ctf:lock:score:%d", userID)

    // 获取分布式锁，超时 5 秒
    lock, err := s.redis.SetNX(ctx, lockKey, 1, 5*time.Second).Result()
    if err != nil || !lock {
        return fmt.Errorf("获取锁失败")
    }
    defer s.redis.Del(ctx, lockKey)

    // 原有逻辑...
}
```

#### [H2] 幂等性缺失：重复提交同一题目会重复计分
- **文件**：`code/backend/internal/module/practice/score_service.go:60-64`
- **问题描述**：查询使用 `DISTINCT challenge_id`，但如果用户对同一题目提交多次正确答案（虽然业务上应该阻止，但数据层未保证），仍可能重复计分。更严重的是，`UpdateUserScore` 每次都全量重算，如果在异步更新期间用户又提交了新的 Flag，可能导致计分遗漏或重复。
- **影响范围/风险**：得分统计不准确，可能出现分数回退或跳跃
- **修正建议**：
  1. 在 submissions 表添加唯一索引 `UNIQUE(user_id, challenge_id, is_correct=true)`
  2. 或者改为增量更新而非全量重算：
```go
// 在 SubmitFlag 成功后直接增加分数
func (s *ScoreService) AddScore(userID, challengeID int64) error {
    score := s.CalculateScore(challengeID)

    // 原子性增加分数
    err := s.db.Exec(`
        INSERT INTO user_scores (user_id, total_score, solved_count, updated_at)
        VALUES (?, ?, 1, ?)
        ON CONFLICT (user_id) DO UPDATE SET
            total_score = user_scores.total_score + EXCLUDED.total_score,
            solved_count = user_scores.solved_count + 1,
            updated_at = EXCLUDED.updated_at
    `, userID, score, time.Now()).Error

    return err
}
```

#### [H3] 硬编码：缓存 TTL 和 Redis Key 前缀未外部化
- **文件**：`code/backend/internal/module/practice/score_service.go:89, 93, 132`
- **问题描述**：
  1. 缓存 TTL `5*time.Minute` 硬编码在代码中
  2. Redis Key 前缀 `"ctf:score:user:"`, `"ctf:ranking"` 直接拼接字符串
  3. 分布式锁超时时间未配置化
- **影响范围/风险**：无法根据环境调整缓存策略，Key 命名不统一，维护困难
- **修正建议**：
```go
// 1. 创建配置类
type ScoreConfig struct {
    CacheTTL      time.Duration `mapstructure:"cache_ttl"`
    LockTimeout   time.Duration `mapstructure:"lock_timeout"`
}

// 2. 创建 Redis Key 工具类
// internal/pkg/cache/keys.go
const (
    KeyPrefixUserScore = "ctf:score:user"
    KeyPrefixRanking   = "ctf:ranking"
    KeyPrefixScoreLock = "ctf:lock:score"
)

func UserScoreKey(userID int64) string {
    return fmt.Sprintf("%s:%d", KeyPrefixUserScore, userID)
}

func ScoreLockKey(userID int64) string {
    return fmt.Sprintf("%s:%d", KeyPrefixScoreLock, userID)
}

// 3. 使用配置和工具类
cacheKey := cache.UserScoreKey(userID)
s.redis.Set(ctx, cacheKey, totalScore, s.config.CacheTTL)
```

#### [H4] 数据一致性：Redis 和数据库更新非原子性
- **文件**：`code/backend/internal/module/practice/score_service.go:75-96`
- **问题描述**：先更新数据库，再更新 Redis。如果 Redis 更新失败，会导致缓存和数据库不一致。且没有错误处理，Redis 更新失败会被静默忽略。
- **影响范围/风险**：用户看到的得分可能是旧数据，排行榜不准确
- **修正建议**：
```go
// 1. 数据库更新成功后再更新 Redis
err = s.db.Exec(...).Error
if err != nil {
    return err
}

// 2. Redis 更新失败应记录日志
if err := s.redis.Set(ctx, cacheKey, totalScore, s.config.CacheTTL).Err(); err != nil {
    s.logger.Error("更新用户得分缓存失败", zap.Int64("userID", userID), zap.Error(err))
}

if err := s.redis.ZAdd(ctx, cache.RankingKey(), redis.Z{...}).Err(); err != nil {
    s.logger.Error("更新排行榜缓存失败", zap.Int64("userID", userID), zap.Error(err))
}

// 3. 考虑使用 Redis Pipeline 批量更新
pipe := s.redis.Pipeline()
pipe.Set(ctx, cacheKey, totalScore, s.config.CacheTTL)
pipe.ZAdd(ctx, cache.RankingKey(), redis.Z{...})
if _, err := pipe.Exec(ctx); err != nil {
    s.logger.Error("批量更新缓存失败", zap.Error(err))
}
```

### 🟡 中优先级

#### [M1] N+1 查询：排行榜查询用户信息效率低下
- **文件**：`code/backend/internal/module/practice/score_service.go:136-145, 158-161`
- **问题描述**：在循环中逐个查询用户信息，当排行榜有 100 条记录时会产生 100 次数据库查询
- **影响范围/风险**：排行榜接口响应慢，数据库压力大
- **修正建议**：
```go
// 1. 批量查询用户信息
userIDs := make([]int64, len(scores))
for i, score := range scores {
    userIDs[i] = score.UserID
}

var users []model.User
s.db.Select("id, username").Where("id IN ?", userIDs).Find(&users)

// 2. 构建 userID -> username 映射
userMap := make(map[int64]string, len(users))
for _, user := range users {
    userMap[user.ID] = user.Username
}

// 3. 使用映射填充结果
for i, score := range scores {
    result = append(result, &dto.RankingItem{
        Rank:        i + 1,
        UserID:      score.UserID,
        Username:    userMap[score.UserID],
        TotalScore:  score.TotalScore,
        SolvedCount: score.SolvedCount,
    })
}
```

#### [M2] 错误处理不完整：Redis 类型断言可能 panic
- **文件**：`code/backend/internal/module/practice/score_service.go:137`
- **问题描述**：`member.Member.(string)` 类型断言未检查，如果 Redis 数据损坏可能导致 panic
- **影响范围/风险**：服务崩溃
- **修正建议**：
```go
userIDStr, ok := member.Member.(string)
if !ok {
    s.logger.Error("排行榜数据类型错误", zap.Any("member", member.Member))
    continue
}

userID, err := strconv.ParseInt(userIDStr, 10, 64)
if err != nil {
    s.logger.Error("解析用户ID失败", zap.String("userIDStr", userIDStr), zap.Error(err))
    continue
}
```

#### [M3] 缓存穿透风险：用户不存在时未缓存空结果
- **文件**：`code/backend/internal/module/practice/score_service.go:101-125`
- **问题描述**：`GetUserScore` 方法未使用缓存，且当用户不存在时返回零值但不缓存，恶意请求可能击穿缓存
- **影响范围/风险**：数据库压力大，可能被攻击
- **修正建议**：
```go
func (s *ScoreService) GetUserScore(userID int64) (*dto.UserScoreInfo, error) {
    ctx := context.Background()
    cacheKey := cache.UserScoreKey(userID)

    // 1. 尝试从缓存获取
    cached, err := s.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        // 缓存命中，解析返回
        var info dto.UserScoreInfo
        json.Unmarshal([]byte(cached), &info)
        return &info, nil
    }

    // 2. 查询数据库
    var userScore model.UserScore
    err = s.db.Where("user_id = ?", userID).First(&userScore).Error

    var info *dto.UserScoreInfo
    if err == gorm.ErrRecordNotFound {
        // 用户不存在，返回零值
        info = &dto.UserScoreInfo{
            UserID:      userID,
            TotalScore:  0,
            SolvedCount: 0,
            Rank:        0,
        }
    } else if err != nil {
        return nil, err
    } else {
        // 查询用户名
        var user model.User
        s.db.Select("username").Where("id = ?", userID).First(&user)

        info = &dto.UserScoreInfo{
            UserID:      userScore.UserID,
            Username:    user.Username,
            TotalScore:  userScore.TotalScore,
            SolvedCount: userScore.SolvedCount,
            Rank:        userScore.Rank,
        }
    }

    // 3. 缓存结果（包括空结果，使用较短 TTL）
    data, _ := json.Marshal(info)
    s.redis.Set(ctx, cacheKey, data, s.config.CacheTTL)

    return info, nil
}
```

#### [M4] 异步更新无错误恢复机制
- **文件**：`code/backend/internal/module/practice/service.go:135-138`
- **问题描述**：异步更新得分失败只记录日志，不会重试或补偿。如果更新失败，用户得分将永久不准确
- **影响范围/风险**：得分统计可能遗漏，需要手动修复
- **修正建议**：
```go
// 方案 1：使用消息队列保证可靠性
type ScoreUpdateEvent struct {
    UserID      int64
    ChallengeID int64
    Timestamp   time.Time
}

// 发送到 MQ
s.mq.Publish("score.update", event)

// 方案 2：记录失败任务到数据库，定时补偿
if err := s.scoreService.UpdateUserScore(userID); err != nil {
    s.logger.Error("更新用户得分失败", zap.Int64("userID", userID), zap.Error(err))

    // 记录失败任务
    s.db.Create(&model.ScoreUpdateTask{
        UserID:    userID,
        Status:    "pending",
        RetryCount: 0,
        CreatedAt: time.Now(),
    })
}
```

#### [M5] 排行榜 Rank 字段未更新
- **文件**：`code/backend/internal/module/practice/score_service.go:75-86`
- **问题描述**：`UpdateUserScore` 方法将 rank 固定设置为 0，但 Model 和 DTO 中都有 Rank 字段，且数据库有 rank 索引。Rank 应该根据 total_score 排序后计算
- **影响范围/风险**：Rank 字段无意义，浪费存储和索引空间
- **修正建议**：
```go
// 方案 1：定时批量更新 Rank（推荐）
func (s *ScoreService) UpdateAllRanks() error {
    // 使用窗口函数更新排名
    return s.db.Exec(`
        UPDATE user_scores
        SET rank = ranked.new_rank
        FROM (
            SELECT user_id,
                   ROW_NUMBER() OVER (ORDER BY total_score DESC, updated_at ASC) as new_rank
            FROM user_scores
        ) AS ranked
        WHERE user_scores.user_id = ranked.user_id
    `).Error
}

// 方案 2：删除 Rank 字段，实时计算
// 如果 Rank 只用于展示，可以在查询时计算，不存储
```

### 🟢 低优先级

#### [L1] 代码重复：用户名查询逻辑重复
- **文件**：`code/backend/internal/module/practice/score_service.go:116, 138, 159`
- **问题描述**：三处都有 `s.db.Select("username").Where("id = ?", userID).First(&user)` 的逻辑
- **影响范围/风险**：代码维护性差
- **修正建议**：提取为私有方法
```go
func (s *ScoreService) getUsernames(userIDs []int64) (map[int64]string, error) {
    var users []model.User
    err := s.db.Select("id, username").Where("id IN ?", userIDs).Find(&users).Error
    if err != nil {
        return nil, err
    }

    result := make(map[int64]string, len(users))
    for _, user := range users {
        result[user.ID] = user.Username
    }
    return result, nil
}
```

#### [L2] 魔法数字：排行榜查询限制未配置化
- **文件**：`code/backend/internal/module/practice/score_service.go:132, 152`
- **问题描述**：排行榜查询使用传入的 limit 参数，但未设置上限。恶意请求可能传入超大值导致性能问题
- **影响范围/风险**：可能被滥用
- **修正建议**：
```go
const MaxRankingLimit = 100

func (s *ScoreService) GetRanking(limit int) ([]*dto.RankingItem, error) {
    if limit <= 0 || limit > MaxRankingLimit {
        limit = MaxRankingLimit
    }
    // ...
}
```

#### [L3] 日志不完整：缺少关键操作日志
- **文件**：`code/backend/internal/module/practice/score_service.go:56-97`
- **问题描述**：`UpdateUserScore` 方法没有记录成功日志，只在 `CalculateScore` 失败时记录错误。无法追踪得分更新历史
- **影响范围/风险**：问题排查困难
- **修正建议**：
```go
func (s *ScoreService) UpdateUserScore(userID int64) error {
    // ...

    s.logger.Info("更新用户得分",
        zap.Int64("userID", userID),
        zap.Int("totalScore", totalScore),
        zap.Int("solvedCount", len(submissions)),
    )

    return nil
}
```

#### [L4] SQL 方言兼容性：UPSERT 语法仅支持 PostgreSQL
- **文件**：`code/backend/internal/module/practice/score_service.go:75-82`
- **问题描述**：`ON CONFLICT ... DO UPDATE` 是 PostgreSQL 特有语法，MySQL 需要使用 `ON DUPLICATE KEY UPDATE`
- **影响范围/风险**：如果切换数据库会报错
- **修正建议**：使用 GORM 的 `Clauses` 实现跨数据库兼容
```go
err = s.db.Clauses(clause.OnConflict{
    Columns:   []clause.Column{{Name: "user_id"}},
    DoUpdates: clause.AssignmentColumns([]string{"total_score", "solved_count", "updated_at"}),
}).Create(&model.UserScore{
    UserID:      userID,
    TotalScore:  totalScore,
    SolvedCount: len(submissions),
    UpdatedAt:   time.Now(),
}).Error
```

#### [L5] 数据库索引冗余
- **文件**：`code/backend/migrations/000010_create_user_scores_table.up.sql:9-10`
- **问题描述**：同时创建了 `idx_user_scores_rank` 和 `idx_user_scores_total_score` 两个索引。如果 rank 字段会被更新为正确的排名，那么 total_score 索引是冗余的（因为 rank 已经是排序后的结果）
- **影响范围/风险**：写入性能略微下降，存储空间浪费
- **修正建议**：根据 Rank 字段的使用方式决定保留哪个索引
```sql
-- 如果 Rank 会被正确维护，只需要 rank 索引
CREATE INDEX idx_user_scores_rank ON user_scores(rank);

-- 如果 Rank 不维护，只需要 total_score 索引用于排序
CREATE INDEX idx_user_scores_total_score ON user_scores(total_score DESC, updated_at ASC);
```

#### [L6] DTO 字段冗余
- **文件**：`code/backend/internal/dto/score.go:4-19`
- **问题描述**：`UserScoreInfo` 和 `RankingItem` 结构几乎完全相同，只是字段顺序不同
- **影响范围/风险**：代码维护性差
- **修正建议**：合并为一个结构体，或让 `RankingItem` 嵌入 `UserScoreInfo`
```go
// 方案 1：合并
type UserScoreInfo struct {
    Rank        int    `json:"rank"`
    UserID      int64  `json:"user_id"`
    Username    string `json:"username"`
    TotalScore  int    `json:"total_score"`
    SolvedCount int    `json:"solved_count"`
}

// 方案 2：复用
type RankingItem = UserScoreInfo
```

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 4 |
| 🟡 中 | 5 |
| 🟢 低 | 6 |
| 合计 | 15 |

## 总体评价

本次实现的实时计分功能在架构设计上基本合理（Model/DTO 分离、分层清晰），但在并发安全、幂等性、配置外部化等关键方面存在严重问题：

**核心风险**：
1. 并发竞争和幂等性缺失可能导致得分统计错误，影响比赛公平性
2. 硬编码问题违反项目规范，维护困难
3. 异步更新无补偿机制，失败后需要手动修复

**必须修复**：
- 所有高优先级问题（H1-H4）必须在下一轮修复
- 中优先级问题（M1-M5）建议在本批次修复，避免技术债累积
- 低优先级问题可在后续迭代中优化

**架构建议**：
- 考虑引入消息队列实现可靠的异步计分
- 定时任务批量更新 Rank 字段，或删除该字段改为实时计算
- 统一 Redis Key 管理和配置外部化
