# CTF 平台代码 Review（skill-recommendation 第 1 轮）：薄弱项识别与推荐功能实现

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | skill-recommendation |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | 5 个提交（d4ff3d9..296859c），10 个文件，349 行新增 |
| 变更概述 | 实现基于能力画像的薄弱项识别与靶场推荐功能 |
| 审查基准 | CTF 平台开发规范（CLAUDE.md） |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | - |

## 问题清单

### 🔴 高优先级

#### [H1] Model 和 DTO 职责混乱，违反分层架构
- **文件**：`code/backend/internal/module/assessment/recommendation_service.go:109-118`
- **问题描述**：`RecommendChallenges` 方法直接在 Service 层构造 DTO，且使用 `model.Challenge` 的字段填充 `dto.ChallengeRecommendation`，违反了"Repository 返回 Model，Service 返回 DTO"的分层原则
- **影响范围/风险**：
  - Service 层直接依赖 Model 结构，耦合度高
  - 如果 Model 字段变更（如 Challenge 增加敏感字段），可能泄漏到 API
  - 违反项目架构规范，降低可维护性
- **修正建议**：
  1. 在 Repository 层添加转换方法，或在 Service 层添加独立的转换函数
  2. Service 方法应返回 `[]*dto.ChallengeRecommendation`，内部调用转换函数

```go
// Service 层添加转换函数
func toChallengeRecommendation(c *model.Challenge, reason string) *dto.ChallengeRecommendation {
    return &dto.ChallengeRecommendation{
        ID:         c.ID,
        Title:      c.Title,
        Category:   c.Category,
        Difficulty: c.Difficulty,
        Points:     c.Points,
        Reason:     reason,
    }
}

// 在循环中使用
for _, c := range challenges {
    reason := fmt.Sprintf("针对薄弱维度：%s", c.Category)
    recommendations = append(recommendations, toChallengeRecommendation(c, reason))
}
```

#### [H2] 缓存键前缀硬编码，未使用统一工具类
- **文件**：`code/backend/internal/module/assessment/recommendation_service.go:63`
- **问题描述**：缓存键使用字符串拼接 `fmt.Sprintf("%s:user:%d", s.cacheKeyPrefix, userID)`，未通过统一的 RedisKeys 工具类管理
- **影响范围/风险**：
  - 缓存键格式分散在各处，难以统一管理
  - 可能与其他模块的缓存键冲突
  - 违反"禁止硬编码"规范
- **修正建议**：
  1. 创建 `internal/pkg/cache/keys.go` 或使用现有的 RedisKeys 工具类
  2. 定义统一的缓存键生成方法

```go
// internal/pkg/cache/keys.go
package cache

import "fmt"

const (
    RecommendationPrefix = "ctf:recommendation"
)

func RecommendationKey(userID int64) string {
    return fmt.Sprintf("%s:user:%d", RecommendationPrefix, userID)
}

// recommendation_service.go 中使用
cacheKey := cache.RecommendationKey(userID)
```

#### [H3] 配置项缺少验证，可能导致运行时错误
- **文件**：`code/backend/internal/config/config.go:237-239`
- **问题描述**：`weak_threshold` 配置项未验证范围（应在 0-1 之间），`cache_ttl` 未验证最小值
- **影响范围/风险**：
  - 如果配置 `weak_threshold = -1` 或 `weak_threshold = 2`，会导致薄弱项识别逻辑错误
  - 如果配置 `cache_ttl = 0`，缓存将立即失效，失去意义
- **修正建议**：
  在 `Load` 函数中添加配置验证逻辑

```go
func Load(env string) (*Config, error) {
    // ... 现有加载逻辑 ...

    // 验证配置
    if cfg.Recommendation.WeakThreshold < 0 || cfg.Recommendation.WeakThreshold > 1 {
        return nil, fmt.Errorf("recommendation.weak_threshold 必须在 0-1 之间，当前值: %.2f", cfg.Recommendation.WeakThreshold)
    }

    if cfg.Recommendation.CacheTTL < time.Minute {
        return nil, fmt.Errorf("recommendation.cache_ttl 不能小于 1 分钟，当前值: %v", cfg.Recommendation.CacheTTL)
    }

    return cfg, nil
}
```

### 🟡 中优先级

#### [M1] Handler 层缺少参数校验和错误日志
- **文件**：`code/backend/internal/module/assessment/handler.go:20-45`
- **问题描述**：
  1. `limit` 参数硬编码上限 50，未通过配置注入
  2. `user_id` 从 context 获取后未检查是否为 0（可能表示未认证）
  3. 错误处理只调用 `response.FromError`，未记录日志
- **影响范围/风险**：
  - 硬编码限制难以调整
  - 如果中间件失效，可能传入无效 userID
  - 错误排查困难，缺少上下文信息
- **修正建议**：

```go
func (h *Handler) GetRecommendations(c *gin.Context) {
    userID := c.GetInt64("user_id")
    if userID == 0 {
        response.Error(c, errcode.ErrUnauthorized())
        return
    }

    // 从配置读取默认值和最大值
    limit := h.config.Recommendation.DefaultLimit
    if limitStr := c.Query("limit"); limitStr != "" {
        if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= h.config.Recommendation.MaxLimit {
            limit = l
        }
    }

    weakDimensions, err := h.recommendationService.GetWeakDimensions(userID)
    if err != nil {
        h.logger.Error("获取薄弱维度失败", zap.Int64("userID", userID), zap.Error(err))
        response.FromError(c, err)
        return
    }

    challenges, err := h.recommendationService.RecommendChallenges(userID, limit)
    if err != nil {
        h.logger.Error("获取推荐靶场失败", zap.Int64("userID", userID), zap.Int("limit", limit), zap.Error(err))
        response.FromError(c, err)
        return
    }

    response.Success(c, gin.H{
        "weak_dimensions": weakDimensions,
        "challenges":      challenges,
    })
}
```

#### [M2] 缓存反序列化失败时未记录日志
- **文件**：`code/backend/internal/module/assessment/recommendation_service.go:66-72`
- **问题描述**：缓存命中后，如果 `json.Unmarshal` 失败，静默忽略错误，未记录日志
- **影响范围/风险**：
  - 缓存数据损坏时无法感知
  - 可能导致频繁回源查询，影响性能
- **修正建议**：

```go
cached, err := s.redis.Get(ctx, cacheKey).Result()
if err == nil {
    var recommendations []*dto.ChallengeRecommendation
    if err := json.Unmarshal([]byte(cached), &recommendations); err == nil {
        return recommendations, nil
    } else {
        s.logger.Warn("缓存反序列化失败",
            zap.String("cacheKey", cacheKey),
            zap.Error(err))
    }
}
```

#### [M3] Repository 查询方法缺少索引提示
- **文件**：`code/backend/internal/module/challenge/repository.go:152-169`
- **问题描述**：`FindPublishedWithTags` 方法使用 JOIN 查询，但未说明需要的索引
- **影响范围/风险**：
  - 如果 `challenge_tags` 表缺少 `(tag_id, challenge_id)` 联合索引，查询性能差
  - 如果 `challenges` 表缺少 `(status, difficulty, points)` 联合索引，排序慢
- **修正建议**：
  1. 在方法注释中说明索引要求
  2. 在 migration 文件中确保索引存在

```go
// FindPublishedWithTags 查询匹配标签的已发布靶场（用于推荐）
// 索引要求：
// - challenge_tags: (tag_id, challenge_id)
// - challenges: (status, difficulty, points)
func (r *Repository) FindPublishedWithTags(limit int, tagIDs []int64, excludeSolved []int64) ([]*model.Challenge, error) {
    // ...
}
```

#### [M4] DTO 响应结构与 Handler 返回不一致
- **文件**：`code/backend/internal/dto/assessment.go:24-27` 和 `handler.go:42-45`
- **问题描述**：定义了 `RecommendationResp` 结构体，但 Handler 使用 `gin.H` 返回，未使用定义的 DTO
- **影响范围/风险**：
  - DTO 定义失去意义
  - 字段名可能不一致（DTO 用 `Challenges`，Handler 用 `challenges`）
  - 无法通过类型检查保证响应格式
- **修正建议**：

```go
// handler.go
func (h *Handler) GetRecommendations(c *gin.Context) {
    // ... 现有逻辑 ...

    resp := &dto.RecommendationResp{
        WeakDimensions: weakDimensions,
        Challenges:     challenges,
    }
    response.Success(c, resp)
}
```

### 🟢 低优先级

#### [L1] Model 常量定义不完整
- **文件**：`code/backend/internal/model/skill_profile.go:5-11`
- **问题描述**：定义了 6 个维度常量，但未提供验证函数或常量切片
- **影响范围/风险**：
  - 其他代码需要验证维度时，需要重复写 switch 或 if 判断
  - 容易出现拼写错误（如 `"Web"` vs `"web"`）
- **修正建议**：

```go
const (
    DimensionWeb       = "web"
    DimensionPwn       = "pwn"
    DimensionReverse   = "reverse"
    DimensionCrypto    = "crypto"
    DimensionMisc      = "misc"
    DimensionForensics = "forensics"
)

var AllDimensions = []string{
    DimensionWeb,
    DimensionPwn,
    DimensionReverse,
    DimensionCrypto,
    DimensionMisc,
    DimensionForensics,
}

func IsValidDimension(dimension string) bool {
    for _, d := range AllDimensions {
        if d == dimension {
            return true
        }
    }
    return false
}
```

#### [L2] Service 构造函数参数过多
- **文件**：`code/backend/internal/module/assessment/recommendation_service.go:31-42`
- **问题描述**：`NewRecommendationService` 有 7 个参数，可读性差
- **影响范围/风险**：
  - 调用时容易传错参数顺序
  - 未来新增配置项时需要修改所有调用处
- **修正建议**：
  使用配置结构体封装参数

```go
type RecommendationServiceConfig struct {
    Repo           *Repository
    ChallengeRepo  ChallengeRepository
    Redis          *redis.Client
    Logger         *zap.Logger
    WeakThreshold  float64
    CacheTTL       time.Duration
    CacheKeyPrefix string
}

func NewRecommendationService(cfg *RecommendationServiceConfig) *RecommendationService {
    return &RecommendationService{
        repo:           cfg.Repo,
        challengeRepo:  cfg.ChallengeRepo,
        redis:          cfg.Redis,
        logger:         cfg.Logger,
        weakThreshold:  cfg.WeakThreshold,
        cacheTTL:       cfg.CacheTTL,
        cacheKeyPrefix: cfg.CacheKeyPrefix,
        db:             cfg.Repo.db,
    }
}
```

#### [L3] 推荐理由过于简单
- **文件**：`code/backend/internal/module/assessment/recommendation_service.go:117`
- **问题描述**：推荐理由固定为 `"针对薄弱维度：{category}"`，信息量少
- **影响范围/风险**：
  - 用户体验差，无法理解为什么推荐这道题
  - 未来可能需要更丰富的推荐理由（如难度匹配、通过率等）
- **修正建议**：

```go
func (s *RecommendationService) buildRecommendationReason(challenge *model.Challenge, weakDimensions []string) string {
    // 找到匹配的薄弱维度
    matchedDimension := challenge.Category
    for _, dim := range weakDimensions {
        if strings.EqualFold(dim, challenge.Category) {
            matchedDimension = dim
            break
        }
    }

    return fmt.Sprintf("针对薄弱维度「%s」，难度：%s，建议优先练习",
        matchedDimension, challenge.Difficulty)
}
```

#### [L4] 缓存写入失败未记录日志
- **文件**：`code/backend/internal/module/assessment/recommendation_service.go:121-123`
- **问题描述**：缓存写入失败时静默忽略，未记录日志
- **影响范围/风险**：
  - Redis 故障时无法感知
  - 可能导致频繁回源查询
- **修正建议**：

```go
if data, err := json.Marshal(recommendations); err == nil {
    if err := s.redis.Set(ctx, cacheKey, data, s.cacheTTL).Err(); err != nil {
        s.logger.Warn("缓存写入失败",
            zap.String("cacheKey", cacheKey),
            zap.Error(err))
    }
} else {
    s.logger.Error("推荐结果序列化失败", zap.Error(err))
}
```

#### [L5] Repository 方法命名不一致
- **文件**：`code/backend/internal/module/challenge/repository.go:172-177`
- **问题描述**：`FindTagsByDimensions` 返回 `[]int64`（tag IDs），但方法名暗示返回 `[]*model.Tag`
- **影响范围/风险**：
  - 命名误导，降低代码可读性
- **修正建议**：
  重命名为 `FindTagIDsByDimensions` 或修改返回类型为 `[]*model.Tag`

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 3 |
| 🟡 中 | 4 |
| 🟢 低 | 5 |
| 合计 | 12 |

## 总体评价

本次变更实现了薄弱项识别与推荐功能的核心逻辑，整体架构清晰，但存在以下主要问题：

1. **架构一致性问题**：Service 层直接操作 Model 构造 DTO，违反分层原则（H1）
2. **配置外部化不完整**：缓存键未使用统一工具类（H2），Handler 层存在硬编码限制（M1）
3. **配置验证缺失**：关键配置项未验证范围，可能导致运行时错误（H3）
4. **可观测性不足**：多处错误处理缺少日志记录（M2、L4）

**优点**：
- Model 和 DTO 正确分离，未出现敏感字段泄漏
- 使用了缓存优化性能
- Repository 层职责清晰
- 配置项通过 Config 注入，未硬编码在代码中

**建议优先修复**：
1. H1：重构 Service 层的 Model → DTO 转换逻辑
2. H2：使用统一的缓存键管理工具类
3. H3：添加配置验证逻辑
4. M1：补充 Handler 层的参数校验和日志

修复后可进入第 2 轮 review。
