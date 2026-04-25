# CTF 平台代码 Review（progress 第 1 轮）：个人解题进度与时间线功能

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | progress |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | commit d1e9d90，5 个文件，+314 行 |
| 变更概述 | 实现个人解题进度统计（总分/排名/分类/难度）和时间线查询功能 |
| 审查基准 | /home/azhi/workspace/projects/ctf/CLAUDE.md |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | - |

## 问题清单

### 🔴 高优先级

#### [H1] 缓存 Key 未使用统一命名空间管理
- **文件**：`code/backend/internal/module/practice/service.go:154`
- **问题描述**：缓存 Key 直接硬编码为 `fmt.Sprintf("ctf:progress:%d", userID)`，违反项目规范要求的"Redis Key 前缀必须通过统一工具类管理"
- **影响范围/风险**：
  - 缓存 Key 分散在各处，无法统一管理和变更
  - 多环境部署时可能出现 Key 冲突
  - 无法统一添加命名空间前缀
- **修正建议**：
```go
// 在 internal/constants/redis_keys.go 或类似文件中定义
const (
    KeyPrefixUserProgress = "user:progress"
)

func UserProgressKey(userID int64) string {
    return fmt.Sprintf("%s:%d", KeyPrefixUserProgress, userID)
}

// 在 service.go 中使用
cacheKey := constants.UserProgressKey(userID)
```

#### [H2] 缓存 TTL 硬编码，未通过配置注入
- **文件**：`code/backend/internal/module/practice/service.go:169`
- **问题描述**：缓存过期时间 `10*time.Minute` 直接硬编码在代码中，违反"缓存 TTL 必须通过配置注入"规范
- **影响范围/风险**：
  - 无法根据环境动态调整缓存时间
  - 生产环境需要调整 TTL 时必须修改代码重新部署
  - 不同场景（开发/测试/生产）无法使用不同的缓存策略
- **修正建议**：
```go
// 在 config/config.go 中添加
type CacheConfig struct {
    ProgressTTL time.Duration `mapstructure:"progress_ttl"`
}

// 在配置文件中
cache:
  progress_ttl: 10m

// 在 Service 中注入并使用
s.redis.Set(ctx, cacheKey, data, s.cfg.Cache.ProgressTTL)
```

#### [H3] 排名计算逻辑存在性能问题
- **文件**：`code/backend/internal/module/practice/repository.go:125-143`
- **问题描述**：排名查询使用子查询嵌套，每次查询都需要全表扫描两次（外层统计 + 内层子查询），在用户量大时性能极差
- **影响范围/风险**：
  - 用户数达到 1000+ 时响应时间可能超过 1 秒
  - 高并发场景下会造成数据库压力
  - 无法利用索引优化
- **修正建议**：
```go
// 方案 1：使用窗口函数（推荐）
err := r.db.Raw(`
    WITH ranked_users AS (
        SELECT
            s.user_id,
            SUM(c.points) as total_score,
            RANK() OVER (ORDER BY SUM(c.points) DESC) as rank
        FROM submissions s
        JOIN challenges c ON s.challenge_id = c.id
        WHERE s.is_correct = true AND c.status = ?
        GROUP BY s.user_id
    )
    SELECT rank FROM ranked_users WHERE user_id = ?
`, "published", userID).Scan(&rank).Error

// 方案 2：使用 Redis 有序集合缓存排行榜（更高性能）
// 在提交成功时更新 ZADD ctf:leaderboard {score} {userID}
// 查询时使用 ZREVRANK ctf:leaderboard {userID}
```

#### [H4] 时间线查询存在 N+1 问题和数据不一致风险
- **文件**：`code/backend/internal/module/practice/repository.go:149-183`
- **问题描述**：
  1. 三次 JOIN challenges 表，每个子查询都重复关联
  2. `instance_destroy` 使用 `updated_at` 作为销毁时间，但 `updated_at` 在任何更新时都会变化，不准确
  3. 缺少时间范围限制，用户历史数据多时会返回大量数据
- **影响范围/风险**：
  - 查询性能随数据量线性下降
  - 销毁时间不准确，可能显示错误的时间线
  - 前端可能因数据量过大而卡顿
- **修正建议**：
```go
// 1. 优化查询结构，减少重复 JOIN
err := r.db.Raw(`
    SELECT * FROM (
        SELECT 'instance_start' as type, i.challenge_id, i.created_at as timestamp,
               NULL as is_correct, NULL as points
        FROM instances i
        WHERE i.user_id = ?
        UNION ALL
        SELECT 'flag_submit' as type, s.challenge_id, s.submitted_at as timestamp,
               s.is_correct, CASE WHEN s.is_correct THEN c.points ELSE NULL END as points
        FROM submissions s
        LEFT JOIN challenges c ON s.challenge_id = c.id
        WHERE s.user_id = ?
        UNION ALL
        SELECT 'instance_destroy' as type, i.challenge_id, i.destroyed_at as timestamp,
               NULL as is_correct, NULL as points
        FROM instances i
        WHERE i.user_id = ? AND i.status IN ('stopped', 'expired') AND i.destroyed_at IS NOT NULL
    ) events
    LEFT JOIN challenges c ON events.challenge_id = c.id
    ORDER BY events.timestamp DESC
    LIMIT 100
`, userID, userID, userID).Scan(&events).Error

// 2. 在 instances 表添加 destroyed_at 字段记录准确销毁时间
// 3. 添加分页参数，避免一次性返回过多数据
```

### 🟡 中优先级

#### [M1] 进度统计未考虑已下线靶场
- **文件**：`code/backend/internal/module/practice/repository.go:54-64`
- **问题描述**：只过滤了 `c.status = 'published'`，但如果靶场后续被下线（status 改为 draft 或 archived），用户的历史解题记录仍会计入总分和统计
- **影响范围/风险**：
  - 靶场下线后用户分数不会相应减少
  - 排行榜数据不准确
  - 可能导致用户投诉（为什么我的分数变了）
- **修正建议**：
```go
// 方案 1：明确业务规则 - 历史解题记录是否保留分数？
// 如果保留：当前实现正确，但需要在架构文档中说明
// 如果不保留：需要在靶场下线时触发分数重算或使用快照机制

// 方案 2：添加 solved_at 快照字段
// 在 submissions 表添加 points_snapshot 字段，记录提交时的分数
// 统计时使用快照值而非实时 JOIN
SELECT COALESCE(SUM(s.points_snapshot), 0) as total_score
FROM submissions s
WHERE s.user_id = ? AND s.is_correct = true
```

#### [M2] 分类和难度统计查询效率低
- **文件**：`code/backend/internal/module/practice/repository.go:67-106`
- **问题描述**：
  1. 两个统计查询分别执行，可以合并为一次查询
  2. `COUNT(DISTINCT CASE WHEN ...)` 在大数据量下性能较差
  3. 缺少索引提示
- **影响范围/风险**：
  - 每次进度查询需要 5 次数据库查询（总分、排名、分类、难度、时间线）
  - 响应时间随靶场数量增长
- **修正建议**：
```go
// 合并为一次查询返回所有统计
type ProgressStats struct {
    TotalScore      int
    TotalSolved     int
    CategoryStats   string // JSON
    DifficultyStats string // JSON
}

err := r.db.Raw(`
    SELECT
        COALESCE(SUM(CASE WHEN s.is_correct THEN c.points END), 0) as total_score,
        COUNT(DISTINCT CASE WHEN s.is_correct THEN c.id END) as total_solved,
        JSON_OBJECT(
            'category', JSON_ARRAYAGG(JSON_OBJECT('category', c.category, 'solved', ...))
        ) as category_stats
    FROM challenges c
    LEFT JOIN submissions s ON c.id = s.challenge_id AND s.user_id = ?
    WHERE c.status = ?
`, userID, "published").Scan(&stats).Error

// 或者保持当前结构，但添加复合索引
// CREATE INDEX idx_submissions_user_correct ON submissions(user_id, is_correct, challenge_id);
// CREATE INDEX idx_challenges_status_category ON challenges(status, category);
```

#### [M3] 缓存失效策略不完整
- **文件**：`code/backend/internal/module/practice/service.go:151-170`
- **问题描述**：
  1. 只在查询时设置缓存，但在提交 Flag 成功后未主动失效缓存
  2. 用户提交成功后需要等待 10 分钟才能看到最新进度
  3. 缓存反序列化失败时静默降级，未记录日志
- **影响范围/风险**：
  - 用户体验差：刚解题成功但进度页面不更新
  - 缓存穿透风险：反序列化失败时每次都查数据库
- **修正建议**：
```go
// 1. 在 SubmitFlag 成功后主动删除缓存
func (s *Service) SubmitFlag(userID, challengeID int64, flag string) (*dto.SubmitResp, error) {
    // ... 提交逻辑 ...
    if isCorrect {
        // 删除进度缓存
        ctx := context.Background()
        cacheKey := constants.UserProgressKey(userID)
        s.redis.Del(ctx, cacheKey)
    }
    return resp, nil
}

// 2. 缓存反序列化失败时记录日志
if err == nil {
    var resp dto.ProgressResp
    if err := json.Unmarshal([]byte(cached), &resp); err != nil {
        s.logger.Warn("进度缓存反序列化失败", zap.Int64("userID", userID), zap.Error(err))
    } else {
        return &resp, nil
    }
}
```

#### [M4] 时间线事件缺少分页支持
- **文件**：`code/backend/internal/module/practice/service.go:176-199`
- **问题描述**：
  1. 一次性返回所有历史事件，无分页参数
  2. 活跃用户可能有数千条记录，前端难以处理
  3. Handler 层未提供分页参数接收
- **影响范围/风险**：
  - 接口响应体过大
  - 前端渲染卡顿
  - 数据库查询慢
- **修正建议**：
```go
// 1. DTO 添加分页请求
type TimelineReq struct {
    Page     int `form:"page" binding:"omitempty,min=1"`
    PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// 2. Repository 添加 LIMIT OFFSET
func (r *Repository) GetUserTimeline(userID int64, limit, offset int) ([]Event, int64, error) {
    // ... 查询逻辑 ...
    ORDER BY timestamp DESC
    LIMIT ? OFFSET ?
    // 同时返回总数
}

// 3. Handler 接收分页参数
func (h *Handler) GetTimeline(c *gin.Context) {
    var req dto.TimelineReq
    if err := c.ShouldBindQuery(&req); err != nil {
        response.Error(c, errcode.ErrInvalidParam)
        return
    }
    // 设置默认值
    if req.Page == 0 { req.Page = 1 }
    if req.PageSize == 0 { req.PageSize = 20 }
}
```

### 🟢 低优先级

#### [L1] DTO 字段注释不完整
- **文件**：`code/backend/internal/dto/progress.go:5-41`
- **问题描述**：部分字段缺少详细说明，如 `Rank` 未说明是全站排名还是班级排名，`Type` 未列举所有可能值
- **影响范围/风险**：前端开发者需要查看代码才能理解字段含义
- **修正建议**：
```go
type ProgressResp struct {
    TotalScore      int                  `json:"total_score"`       // 总得分（仅统计已发布靶场）
    TotalSolved     int                  `json:"total_solved"`      // 总解题数（去重）
    Rank            int                  `json:"rank"`              // 全站排名（按总分降序，1 表示第一名）
    CategoryStats   []CategoryStat       `json:"category_stats"`    // 按分类统计
    DifficultyStats []DifficultyStat     `json:"difficulty_stats"`  // 按难度统计
}

type TimelineEvent struct {
    Type        string    `json:"type"`         // 事件类型: instance_start(启动), flag_submit(提交), instance_destroy(销毁)
    ChallengeID int64     `json:"challenge_id"` // 靶场 ID
    Title       string    `json:"title"`        // 靶场标题
    Timestamp   time.Time `json:"timestamp"`    // 事件时间（UTC）
    IsCorrect   *bool     `json:"is_correct,omitempty"` // Flag 是否正确（仅 flag_submit 事件）
    Points      *int      `json:"points,omitempty"`     // 获得分数（仅正确提交时有值）
}
```

#### [L2] 错误日志缺少关键上下文
- **文件**：`code/backend/internal/module/practice/service.go:168, 177`
- **问题描述**：日志只记录了 userID 和 error，缺少查询参数、耗时等关键信息，不利于问题排查
- **影响范围/风险**：生产环境出现慢查询时难以定位原因
- **修正建议**：
```go
start := time.Now()
totalScore, totalSolved, err := s.repo.GetUserProgress(userID)
if err != nil {
    s.logger.Error("查询用户进度失败",
        zap.Int64("userID", userID),
        zap.Duration("elapsed", time.Since(start)),
        zap.Error(err),
    )
    return nil, errcode.ErrInternal.WithCause(err)
}
s.logger.Debug("查询用户进度成功",
    zap.Int64("userID", userID),
    zap.Int("totalScore", totalScore),
    zap.Int("totalSolved", totalSolved),
    zap.Duration("elapsed", time.Since(start)),
)
```

#### [L3] 时间线查询结果未排序说明
- **文件**：`code/backend/internal/module/practice/repository.go:181`
- **问题描述**：SQL 使用 `ORDER BY timestamp ASC` 升序排列，但通常时间线应该是最新事件在前（降序）
- **影响范围/风险**：前端需要额外反转数组，或者用户体验不符合预期
- **修正建议**：
```go
// 如果时间线应该最新在前，改为 DESC
ORDER BY timestamp DESC

// 或者在架构文档中明确说明排序规则和原因
```

#### [L4] Repository 方法返回匿名结构体
- **文件**：`code/backend/internal/module/practice/repository.go:67-106, 149-183`
- **问题描述**：`GetCategoryStats`、`GetDifficultyStats`、`GetUserTimeline` 返回匿名结构体，不利于类型复用和测试
- **影响范围/风险**：
  - 代码可读性差
  - 无法为返回类型编写单元测试
  - IDE 自动补全支持不好
- **修正建议**：
```go
// 在 repository.go 或 model 包中定义
type CategoryStat struct {
    Category string
    Solved   int
    Total    int
}

type DifficultyStat struct {
    Difficulty string
    Solved     int
    Total      int
}

type TimelineEvent struct {
    Type        string
    ChallengeID int64
    Title       string
    Timestamp   time.Time
    IsCorrect   *bool
    Points      *int
}

// 修改方法签名
func (r *Repository) GetCategoryStats(userID int64) ([]CategoryStat, error)
func (r *Repository) GetDifficultyStats(userID int64) ([]DifficultyStat, error)
func (r *Repository) GetUserTimeline(userID int64) ([]TimelineEvent, error)
```

#### [L5] 缺少数据库索引建议
- **文件**：整体架构
- **问题描述**：代码中大量使用 `user_id`、`challenge_id`、`is_correct`、`status` 等字段查询，但未提供索引创建建议
- **影响范围/风险**：生产环境数据量增长后查询性能下降
- **修正建议**：
```sql
-- 在 migration 文件中添加
CREATE INDEX idx_submissions_user_correct ON submissions(user_id, is_correct, challenge_id);
CREATE INDEX idx_submissions_challenge_correct ON submissions(challenge_id, is_correct);
CREATE INDEX idx_challenges_status_category ON challenges(status, category);
CREATE INDEX idx_challenges_status_difficulty ON challenges(status, difficulty);
CREATE INDEX idx_instances_user_status ON instances(user_id, status);
CREATE INDEX idx_instances_user_created ON instances(user_id, created_at DESC);
```

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 4 |
| 🟡 中 | 4 |
| 🟢 低 | 5 |
| 合计 | 13 |

## 总体评价

本次实现的个人解题进度功能在架构分层上基本符合规范（Model/DTO 分离、Repository/Service/Handler 职责清晰），但存在以下核心问题需要修复：

**必须立即修复的问题**：
1. 缓存 Key 和 TTL 硬编码，违反配置外部化原则
2. 排名计算性能问题，在用户量增长后会成为瓶颈
3. 时间线查询存在 N+1 和数据不一致风险
4. 缓存失效策略不完整，影响用户体验

**建议优化的方向**：
1. 添加分页支持，避免大数据量问题
2. 优化查询结构，减少数据库往返次数
3. 完善索引设计，提前规避性能问题
4. 补充详细注释和架构文档

修复高优先级问题后，该功能可以满足基本需求，但建议在上线前完成中优先级问题的修复，以保证生产环境的稳定性和用户体验。
