# CTF Backend 代码 Review（student-query 第 1 轮）：B17 学员靶场查询功能实现

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | student-query |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | commit d7d88cd，12 个文件，855 行新增，28 行删除 |
| 变更概述 | 实现 B17 学员靶场查询功能，包括列表查询、详情查询、完成状态、缓存机制 |
| 审查基准 | internal/module/challenge/B17-README.md |
| 审查日期 | 2026-03-06 |

## 问题清单

### 🔴 高优先级

#### [H1] 缓存 TTL 硬编码违反配置外部化原则
- **文件**：`internal/module/challenge/service.go:262`
- **问题描述**：缓存过期时间 `5*time.Minute` 直接硬编码在代码中
- **影响范围/风险**：
  - 违反项目配置外部化规范（全局 CLAUDE.md 明确禁止硬编码）
  - 无法根据环境动态调整缓存策略
  - 生产环境可能需要不同的 TTL 配置
- **修正建议**：
```go
// 在配置文件中定义
type ChallengeConfig struct {
    SolvedCountCacheTTL time.Duration `mapstructure:"solved_count_cache_ttl"`
}

// Service 中注入配置
type Service struct {
    repo      *Repository
    imageRepo *ImageRepository
    redis     *redis.Client
    config    *ChallengeConfig
}

// 使用配置值
s.redis.Set(ctx, cacheKey, data, s.config.SolvedCountCacheTTL)
```

#### [H2] Redis Key 前缀硬编码，未使用统一工具类
- **文件**：`internal/module/challenge/service.go:246`
- **问题描述**：缓存键 `challenge:solved_count:%d` 直接在代码中拼接字符串
- **影响范围/风险**：
  - 违反全局 CLAUDE.md 规范（Redis Key 必须通过统一工具类管理）
  - 多环境部署时可能出现 key 冲突
  - 缺少命名空间隔离机制
- **修正建议**：
```go
// 创建 pkg/cache/keys.go
package cache

import "fmt"

const (
    KeyPrefixChallenge = "challenge"
)

func ChallengeSolvedCountKey(challengeID int64) string {
    return fmt.Sprintf("%s:solved_count:%d", KeyPrefixChallenge, challengeID)
}

// 在 service.go 中使用
cacheKey := cache.ChallengeSolvedCountKey(challengeID)
```

#### [H3] 错误消息硬编码，未提取为常量
- **文件**：`internal/module/challenge/service.go:36, 81, 95, 144, 221`
- **问题描述**：错误消息字符串散落在多处，如 "镜像不存在"、"存在运行中的实例，无法删除"、"challenge not published"
- **影响范围/风险**：
  - 违反全局 CLAUDE.md 规范（错误消息必须提取为常量）
  - 多处重复字符串，维护困难
  - 国际化支持困难
  - 错误消息不一致（中英文混用）
- **修正建议**：
```go
// 创建 internal/module/challenge/errors.go
package challenge

const (
    ErrMsgImageNotFound        = "镜像不存在"
    ErrMsgHasRunningInstances  = "存在运行中的实例，无法删除"
    ErrMsgImageNotConfigured   = "靶场未关联镜像，无法发布"
    ErrMsgChallengeNotPublished = "靶场未发布"
)

// 使用时
return nil, errors.New(ErrMsgImageNotFound)
```

#### [H4] 分页默认值硬编码，未通过配置注入
- **文件**：`internal/module/challenge/service.go:126, 201; repository.go:62, 111`
- **问题描述**：分页默认值 `page=1`, `size=20`, `max=100` 直接硬编码
- **影响范围/风险**：
  - 违反配置外部化原则
  - Repository 和 Service 层重复逻辑
  - 无法根据业务需求动态调整
- **修正建议**：
```go
// 在配置文件中定义
type PaginationConfig struct {
    DefaultPage     int `mapstructure:"default_page"`
    DefaultPageSize int `mapstructure:"default_page_size"`
    MaxPageSize     int `mapstructure:"max_page_size"`
}

// 注入到 Service 中使用
if page < 1 {
    page = s.config.Pagination.DefaultPage
}
```

#### [H5] N+1 查询问题导致性能风险
- **文件**：`internal/module/challenge/service.go:173-194`
- **问题描述**：在 `ListPublishedChallenges` 中对每个靶场都执行 3 次独立查询（`GetSolvedStatus`、`getSolvedCountCached`、`GetTotalAttempts`）
- **影响范围/风险**：
  - 假设返回 20 条靶场，将产生 1 + 20×3 = 61 次数据库查询
  - 严重的性能问题，随着分页数据增加线性恶化
  - 违反架构文档中的性能优化要求
- **修正建议**：
```go
// 在 Repository 中批量查询
func (r *Repository) BatchGetSolvedStatus(userID int64, challengeIDs []int64) (map[int64]bool, error) {
    var results []struct {
        ChallengeID int64
    }
    err := r.db.Table("submissions").
        Select("DISTINCT challenge_id").
        Where("user_id = ? AND challenge_id IN ? AND is_correct = ?", userID, challengeIDs, true).
        Find(&results).Error

    statusMap := make(map[int64]bool)
    for _, r := range results {
        statusMap[r.ChallengeID] = true
    }
    return statusMap, err
}

// Service 层批量调用
challengeIDs := make([]int64, len(challenges))
for i, c := range challenges {
    challengeIDs[i] = c.ID
}
solvedMap, _ := s.repo.BatchGetSolvedStatus(userID, challengeIDs)
```

#### [H6] 敏感字段泄漏风险
- **文件**：`internal/dto/challenge.go:70`
- **问题描述**：`ChallengeDetailResp` 包含 `FlagType` 字段，但未包含 `FlagHash`、`FlagSalt`、`FlagPrefix`
- **影响范围/风险**：
  - 虽然当前未泄漏敏感字段，但 DTO 设计不够明确
  - 需要在代码审查中确认 Model → DTO 转换是否安全
  - `FlagType` 暴露给学员可能泄漏 Flag 生成机制
- **修正建议**：
```go
// 评估是否需要向学员暴露 FlagType
// 如果不需要，从 ChallengeDetailResp 中移除
type ChallengeDetailResp struct {
    // ... 其他字段
    // FlagType string `json:"flag_type"` // 移除或评估必要性
}
```

### 🟡 中优先级

#### [M1] 错误处理不统一，未使用 errcode 包
- **文件**：`internal/module/challenge/service.go:36, 81, 95, 144, 221`
- **问题描述**：使用 `errors.New()` 返回错误，未使用项目统一的 `errcode` 包
- **影响范围/风险**：
  - 违反 CTF CLAUDE.md 错误处理规范
  - Handler 层无法正确识别错误类型并返回对应 HTTP 状态码
  - 前端无法获得结构化错误码
- **修正建议**：
```go
// 使用 errcode 包
if _, err := s.imageRepo.FindByID(req.ImageID); err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, errcode.ErrNotFound("镜像")
    }
    return nil, err
}

if challenge.Status != model.ChallengeStatusPublished {
    return nil, errcode.ErrForbidden()
}
```

#### [M2] 缓存错误被静默忽略
- **文件**：`internal/module/challenge/service.go:184, 187, 190, 224-226`
- **问题描述**：所有数据库查询错误都被 `_` 忽略，包括缓存查询和统计查询
- **影响范围/风险**：
  - 数据库错误被静默吞掉，无法监控和排查
  - 缓存失效时返回零值，用户看到错误数据
  - 缺少可观测性
- **修正建议**：
```go
// 记录错误日志
isSolved, err := s.repo.GetSolvedStatus(userID, c.ID)
if err != nil {
    log.Errorf("failed to get solved status: %v", err)
    isSolved = false
}

// 或者在缓存失败时降级
solvedCount, err := s.getSolvedCountCached(c.ID)
if err != nil {
    log.Warnf("failed to get solved count for challenge %d: %v", c.ID, err)
    solvedCount = 0
}
```

#### [M3] Repository 层分页逻辑重复
- **文件**：`internal/module/challenge/repository.go:56-66, 106-118`
- **问题描述**：`List` 和 `ListPublished` 方法中分页逻辑完全重复
- **影响范围/风险**：
  - 代码重复，维护成本高
  - 修改分页逻辑需要同步多处
- **修正建议**：
```go
// 提取分页逻辑为私有方法
func (r *Repository) applyPagination(db *gorm.DB, page, size int) *gorm.DB {
    if page < 1 {
        page = 1
    }
    if size < 1 {
        size = 20
    }
    if size > 100 {
        size = 100
    }
    offset := (page - 1) * size
    return db.Offset(offset).Limit(size)
}

// 使用
db = r.applyPagination(db, query.Page, query.Size)
```

#### [M4] 数据库索引设计不完整
- **文件**：`migrations/000008_create_submissions_table.up.sql:11-12`
- **问题描述**：缺少针对查询场景的复合索引
- **影响范围/风险**：
  - `GetSolvedCount` 查询 `WHERE challenge_id = ? AND is_correct = ?` 缺少复合索引
  - `GetSolvedStatus` 查询 `WHERE user_id = ? AND challenge_id = ? AND is_correct = ?` 现有索引不够优化
- **修正建议**：
```sql
-- 优化完成人数查询
CREATE INDEX idx_submissions_challenge_correct ON submissions(challenge_id, is_correct);

-- 优化用户完成状态查询（覆盖索引）
CREATE INDEX idx_submissions_user_challenge_correct ON submissions(user_id, challenge_id, is_correct);
```

#### [M5] Service 层分页逻辑与 Repository 层重复
- **文件**：`internal/module/challenge/service.go:120-127, 196-203`
- **问题描述**：Service 层再次处理分页默认值，与 Repository 层逻辑重复
- **影响范围/风险**：
  - 职责不清，分页逻辑应该在哪一层处理不明确
  - 两层都处理导致逻辑冗余
- **修正建议**：
```go
// 方案 1：分页逻辑统一放在 Repository 层，Service 层直接使用
// 方案 2：Service 层规范化参数后传给 Repository，Repository 不再处理默认值

// 推荐方案 1：Repository 层统一处理
func (s *Service) ListPublishedChallenges(userID int64, query *dto.ChallengeQuery) (*dto.PageResult, error) {
    challenges, total, err := s.repo.ListPublished(query)
    // ... 直接使用 query.Page 和 query.Size，不再处理默认值
}
```

#### [M6] 缺少对 userID = 0 的明确处理
- **文件**：`internal/module/challenge/handler.go:126-129, 148-151`
- **问题描述**：未登录用户 `userID` 设为 0，但未在 Service 层明确处理匿名访问逻辑
- **影响范围/风险**：
  - 匿名用户查询 `GetSolvedStatus(0, challengeID)` 可能返回错误结果
  - 业务逻辑不清晰，是否允许匿名访问未明确
- **修正建议**：
```go
// Handler 层明确处理
userID, exists := c.Get("user_id")
var uid int64
if exists {
    uid = userID.(int64)
} else {
    uid = 0 // 匿名用户
}

// Service 层明确处理匿名逻辑
func (s *Service) ListPublishedChallenges(userID int64, query *dto.ChallengeQuery) (*dto.PageResult, error) {
    // ...
    if userID > 0 {
        isSolved, _ = s.repo.GetSolvedStatus(userID, c.ID)
    } else {
        isSolved = false // 匿名用户未完成
    }
}
```

### 🟢 低优先级

#### [L1] 缺少对 Keyword 搜索的 SQL 注入防护说明
- **文件**：`internal/module/challenge/repository.go:92`
- **问题描述**：虽然使用了 GORM 参数化查询，但 `LIKE` 查询中 `%` 拼接可能引起误解
- **影响范围/风险**：
  - 当前实现安全（GORM 会转义），但代码可读性不佳
  - 未来维护者可能误认为存在注入风险
- **修正建议**：
```go
// 添加注释说明
// GORM 会自动转义参数，防止 SQL 注入
db = db.Where("title LIKE ? OR description LIKE ?", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
```

#### [L2] 排序字段未校验，存在潜在风险
- **文件**：`internal/module/challenge/repository.go:99-104`
- **问题描述**：`query.SortBy` 未校验，虽然当前只有 `difficulty` 一个分支，但未来扩展时可能引入风险
- **影响范围/风险**：
  - 如果未来添加更多排序字段，可能引入 SQL 注入（如果直接拼接字符串）
  - 当前实现安全，但缺少防御性编程
- **修正建议**：
```go
// 在 DTO 中添加校验
type ChallengeQuery struct {
    SortBy string `form:"sort_by" binding:"omitempty,oneof=created_at difficulty points"`
}
```

#### [L3] 缺少对 Redis 连接失败的降级处理
- **文件**：`internal/module/challenge/service.go:243-265`
- **问题描述**：Redis 不可用时，缓存查询失败会直接查询数据库，但未记录降级事件
- **影响范围/风险**：
  - Redis 故障时无法及时发现
  - 缺少监控指标
- **修正建议**：
```go
cached, err := s.redis.Get(ctx, cacheKey).Result()
if err == nil {
    // 缓存命中
} else if err == redis.Nil {
    // 缓存未命中，正常
} else {
    // Redis 连接失败，记录告警
    log.Errorf("redis get failed, fallback to db: %v", err)
}
```

#### [L4] 缺少对分页参数越界的友好提示
- **文件**：`internal/module/challenge/repository.go:106-118`
- **问题描述**：当 `page` 超出总页数时，返回空列表，但未给出提示
- **影响范围/风险**：
  - 用户体验不佳，不知道是没有数据还是页码错误
  - 前端需要自行计算总页数
- **修正建议**：
```go
// 在 Service 层添加检查
totalPages := (total + int64(size) - 1) / int64(size)
if page > int(totalPages) && total > 0 {
    return nil, errcode.ErrInvalidParams("页码超出范围")
}
```

#### [L5] 缺少对 submissions 表的外键约束
- **文件**：`migrations/000008_create_submissions_table.up.sql`
- **问题描述**：`user_id` 和 `challenge_id` 未设置外键约束
- **影响范围/风险**：
  - 数据一致性依赖应用层保证
  - 删除用户或靶场时可能产生孤儿记录
- **修正建议**：
```sql
ALTER TABLE submissions
ADD CONSTRAINT fk_submissions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
ADD CONSTRAINT fk_submissions_challenge FOREIGN KEY (challenge_id) REFERENCES challenges(id) ON DELETE CASCADE;
```

#### [L6] 缺少对 flag 字段长度的业务校验
- **文件**：`migrations/000008_create_submissions_table.up.sql:6`
- **问题描述**：`flag VARCHAR(500)` 长度限制只在数据库层，应用层未校验
- **影响范围/风险**：
  - 超长 flag 提交会在数据库层被截断或报错
  - 缺少友好的错误提示
- **修正建议**：
```go
// 在提交 Flag 的 DTO 中添加校验
type SubmitFlagReq struct {
    Flag string `json:"flag" binding:"required,max=500"`
}
```

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 6 |
| 🟡 中 | 6 |
| 🟢 低 | 6 |
| 合计 | 18 |

## 总体评价

代码整体遵循了分层架构和 Model/DTO 分离原则，功能实现完整，但存在以下主要问题：

1. **配置外部化不彻底**：缓存 TTL、Redis Key 前缀、分页默认值、错误消息等多处硬编码，严重违反全局 CLAUDE.md 规范
2. **性能问题严重**：N+1 查询问题会导致列表接口性能极差，必须优化为批量查询
3. **错误处理不规范**：未使用统一的 errcode 包，错误被静默忽略，缺少可观测性
4. **数据库设计可优化**：索引不完整，缺少外键约束

建议优先修复所有高优先级问题，特别是 N+1 查询和配置外部化问题，然后再处理中低优先级问题。

