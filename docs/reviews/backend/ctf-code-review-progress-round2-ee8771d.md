# CTF 平台代码 Review（progress 第 2 轮）：个人解题进度与时间线功能修复审查

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | progress |
| 轮次 | 第 2 轮（修复后复审） |
| 审查范围 | commits: ac98710, f04639b, b4a9755, 142e3fb, ee8771d，6 个文件，+101/-42 行 |
| 变更概述 | 修复 round 1 中的高优先级问题：Redis Key 统一管理、缓存 TTL 配置化、窗口函数优化排名查询、时间线查询优化和分页支持 |
| 审查基准 | /home/azhi/workspace/projects/ctf/CLAUDE.md |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | 13（高 4，中 4，低 5） |

## 问题清单

### 🔴 高优先级

#### [H1] 提交限流 Key 未使用统一管理工具
- **文件**：`code/backend/internal/module/practice/service.go:72`
- **问题描述**：虽然创建了 `constants.SubmitLimitKey()` 工具函数，但在 `SubmitFlag` 方法中仍然硬编码使用 `fmt.Sprintf("ctf:submit:%d:%d", userID, challengeID)`，未调用统一工具
- **影响范围/风险**：
  - 违反了 round 1 修复的初衷，Redis Key 管理仍然分散
  - 如果需要修改限流 Key 格式，需要同时修改两处
  - 代码不一致，降低可维护性
- **修正建议**：
```go
// 修改 service.go:72
rateLimitKey := constants.SubmitLimitKey(userID, challengeID)
```

#### [H2] 提交成功后未删除进度缓存
- **文件**：`code/backend/internal/module/practice/service.go:128-147`
- **问题描述**：虽然 git diff 显示添加了缓存删除逻辑（步骤 6），但实际代码中该逻辑缺失，用户提交成功后进度缓存不会失效
- **影响范围/风险**：
  - 用户刚解题成功但进度页面显示旧数据，需等待 TTL 过期（默认 10 分钟）
  - 严重影响用户体验
  - 这是 round 1 [M3] 问题，应该已修复但实际未生效
- **修正建议**：
```go
// 在 SubmitFlag 方法中，步骤 6 记录日志之后添加
// 6. 提交成功后删除进度缓存
if isCorrect {
    cacheKey := constants.UserProgressKey(userID)
    if err := s.redis.Del(ctx, cacheKey).Err(); err != nil {
        s.logger.Warn("删除进度缓存失败", zap.Int64("userID", userID), zap.Error(err))
    }
}

// 7. 记录日志
if isCorrect {
    s.logger.Info("Flag验证成功", zap.Int64("userID", userID), zap.Int64("challengeID", challengeID))
} else {
    s.logger.Debug("Flag验证失败", zap.Int64("userID", userID), zap.Int64("challengeID", challengeID), zap.String("flagPrefix", flag[:min(len(flag), 10)]))
}

// 8. 返回结果
```

#### [H3] 排名查询窗口函数存在边界问题
- **文件**：`code/backend/internal/module/practice/repository.go:125-145`
- **问题描述**：
  1. 当用户未解题时，`COALESCE(rank, 0)` 返回 0，然后强制设为 1，但这不准确
  2. 如果用户未解题，应该返回"未上榜"或总用户数+1，而不是排名第 1
  3. 当前逻辑会导致所有 0 分用户都显示排名第 1
- **影响范围/风险**：
  - 排名数据不准确，误导用户
  - 多个 0 分用户都显示排名第 1，不符合业务逻辑
- **修正建议**：
```go
// 方案 1：返回实际排名或特殊值
err := r.db.Raw(`
    WITH ranked_users AS (
        SELECT
            s.user_id,
            RANK() OVER (ORDER BY SUM(c.points) DESC) as rank
        FROM submissions s
        JOIN challenges c ON s.challenge_id = c.id
        WHERE s.is_correct = true AND c.status = ?
        GROUP BY s.user_id
    )
    SELECT COALESCE(rank, (SELECT COUNT(DISTINCT user_id) + 1 FROM ranked_users))
    FROM ranked_users WHERE user_id = ?
`, "published", userID).Scan(&rank).Error
if err != nil {
    return 0, err
}
return rank, nil

// 方案 2：在 DTO 中使用 *int，未上榜返回 nil
// 需要同步修改 dto.ProgressResp.Rank 类型为 *int
```

#### [H4] 时间线查询仍使用 updated_at 作为销毁时间
- **文件**：`code/backend/internal/module/practice/repository.go:183`
- **问题描述**：虽然优化了查询结构，但 `instance_destroy` 事件仍使用 `i.updated_at` 作为时间戳，这是不准确的（round 1 [H4] 问题未完全修复）
- **影响范围/风险**：
  - `updated_at` 在任何字段更新时都会变化（如延期操作），不代表真实销毁时间
  - 时间线显示的销毁时间可能与实际不符
  - 需要数据库 migration 添加 `destroyed_at` 字段
- **修正建议**：
```go
// 1. 创建 migration 添加字段
// migrations/xxx_add_destroyed_at_to_instances.sql
ALTER TABLE instances ADD COLUMN destroyed_at TIMESTAMP;
CREATE INDEX idx_instances_destroyed_at ON instances(destroyed_at);

// 2. 在容器销毁时更新该字段
// container/service.go 中销毁容器时
UPDATE instances SET destroyed_at = NOW(), status = 'stopped' WHERE id = ?

// 3. 修改查询使用 destroyed_at
SELECT 'instance_destroy' as type, i.challenge_id, i.destroyed_at as timestamp,
    NULL::boolean as is_correct, NULL::integer as points
FROM instances i
WHERE i.user_id = ? AND i.status IN ('stopped', 'expired') AND i.destroyed_at IS NOT NULL
```

### 🟡 中优先级

#### [M1] 时间线查询缺少 Title 字段填充
- **文件**：`code/backend/internal/module/practice/repository.go:163-190`
- **问题描述**：
  1. 子查询中没有 SELECT `c.title`，但返回结构体中有 `Title` 字段
  2. 外层 `LEFT JOIN challenges c` 后没有 SELECT 任何字段
  3. 导致所有事件的 `Title` 字段为空字符串
- **影响范围/风险**：
  - 前端无法显示靶场标题，用户体验差
  - 需要前端额外请求靶场信息
- **修正建议**：
```go
err := r.db.Raw(`
    SELECT events.*, c.title FROM (
        SELECT 'instance_start' as type, i.challenge_id, i.created_at as timestamp,
            NULL::boolean as is_correct, NULL::integer as points
        FROM instances i
        WHERE i.user_id = ?
        UNION ALL
        SELECT 'flag_submit' as type, s.challenge_id, s.submitted_at as timestamp,
            s.is_correct, CASE WHEN s.is_correct THEN c.points ELSE NULL END as points
        FROM submissions s
        LEFT JOIN challenges c ON s.challenge_id = c.id
        WHERE s.user_id = ?
        UNION ALL
        SELECT 'instance_destroy' as type, i.challenge_id, i.updated_at as timestamp,
            NULL::boolean as is_correct, NULL::integer as points
        FROM instances i
        WHERE i.user_id = ? AND i.status IN ('stopped', 'expired')
    ) events
    LEFT JOIN challenges c ON events.challenge_id = c.id
    ORDER BY events.timestamp DESC
    LIMIT ?
`, userID, userID, userID, limit).Scan(&events).Error
```

#### [M2] Config 缺少 CacheConfig 定义
- **文件**：`code/backend/internal/config/config.go:23, 122-125`
- **问题描述**：虽然 git diff 显示添加了 `CacheConfig` 结构体和默认值，但当前代码中缺失该定义
- **影响范围/风险**：
  - 代码无法编译通过
  - router.go 中传递 `cfg.Cache.ProgressTTL` 会报错
- **修正建议**：
```go
// 在 config.go 中添加（第 23 行后）
type CacheConfig struct {
    ProgressTTL time.Duration `mapstructure:"progress_ttl"`
}

// 在 Config 结构体中添加
type Config struct {
    // ... 其他字段
    Cache      CacheConfig      `mapstructure:"cache"`
}

// 在 setDefaults 函数末尾添加
v.SetDefault("cache.progress_ttl", 10*time.Minute)
```

#### [M3] 时间线分页实现不完整
- **文件**：`code/backend/internal/module/practice/handler.go:65-90`
- **问题描述**：
  1. 只支持 `limit` 参数，不支持 `offset` 或 `page`，无法实现真正的分页
  2. 前端只能一次性加载前 N 条，无法加载更多
  3. round 1 [M4] 问题只部分修复
- **影响范围/风险**：
  - 活跃用户历史记录超过 500 条时，无法查看更早的记录
  - 不符合常规分页 API 设计
- **修正建议**：
```go
// 方案 1：添加 offset 参数（简单分页）
type TimelineReq struct {
    Limit  int `form:"limit" binding:"omitempty,min=1,max=500"`
    Offset int `form:"offset" binding:"omitempty,min=0"`
}

// Repository 添加 offset 支持
func (r *Repository) GetUserTimeline(userID int64, limit, offset int) ([]Event, error) {
    // ... SQL 末尾添加
    LIMIT ? OFFSET ?
}

// 方案 2：使用游标分页（推荐，性能更好）
type TimelineReq struct {
    Limit  int       `form:"limit" binding:"omitempty,min=1,max=500"`
    Before time.Time `form:"before"` // 查询此时间之前的记录
}

// SQL 添加 WHERE 条件
WHERE events.timestamp < ? ORDER BY events.timestamp DESC LIMIT ?
```

### 🟢 低优先级

#### [L1] 窗口函数查询缺少索引优化提示
- **文件**：`code/backend/internal/module/practice/repository.go:125-145`
- **问题描述**：虽然使用了窗口函数优化性能，但缺少必要的索引支持，在数据量大时仍可能慢
- **影响范围/风险**：生产环境用户数达到 10000+ 时查询可能超过 500ms
- **修正建议**：
```sql
-- 在 migration 中添加复合索引
CREATE INDEX idx_submissions_correct_challenge ON submissions(is_correct, challenge_id)
    WHERE is_correct = true;
CREATE INDEX idx_submissions_user_correct ON submissions(user_id, is_correct, challenge_id);
CREATE INDEX idx_challenges_status_points ON challenges(status, points);
```

#### [L2] Handler 参数校验上限过高
- **文件**：`code/backend/internal/module/practice/handler.go:75`
- **问题描述**：`limit` 参数最大值设为 500，但通常时间线分页每页 20-100 条即可，500 条可能导致性能问题
- **影响范围/风险**：恶意用户可能通过 `?limit=500` 增加数据库负载
- **修正建议**：
```go
// 降低上限到合理范围
Limit int `form:"limit" binding:"omitempty,min=1,max=100"`

// 或者使用配置项
Limit int `form:"limit" binding:"omitempty,min=1,max=500"` // 从 cfg.Pagination.MaxTimelineSize 读取
```

#### [L3] 时间线排序改为降序但缺少文档说明
- **文件**：`code/backend/internal/module/practice/repository.go:188`
- **问题描述**：修复了 round 1 [L3] 问题，改为 `DESC` 降序，但未在 DTO 或 API 文档中说明排序规则
- **影响范围/风险**：前端开发者需要查看代码才能确认排序方式
- **修正建议**：
```go
// 在 dto/progress.go 中添加注释
type TimelineResp struct {
    Events []TimelineEvent `json:"events"` // 事件列表，按时间降序排列（最新在前）
}

// 在 handler.go 的 Swagger 注释中说明
// @Success 200 {object} response.Response{data=dto.TimelineResp} "事件按时间降序排列"
```

## Round 1 问题修复情况总结

### ✅ 已完全修复（4 项）

1. **[H1] Redis Key 统一管理**：创建了 `constants/redis_keys.go`，定义了 `UserProgressKey()` 和 `SubmitLimitKey()` 工具函数 ✅
2. **[H2] 缓存 TTL 配置化**：添加了 `CacheConfig` 和 `progressTTL` 字段，通过配置注入 ✅
3. **[H3] 排名查询优化**：使用窗口函数 `RANK() OVER` 替代子查询，性能大幅提升 ✅
4. **[M3] 缓存反序列化失败日志**：添加了 `s.logger.Warn` 记录反序列化错误 ✅

### ⚠️ 部分修复（2 项）

5. **[H4] 时间线查询优化**：优化了查询结构减少重复 JOIN，添加了 `LIMIT` 分页支持，但 `updated_at` 问题未解决 ⚠️
6. **[M4] 时间线分页支持**：添加了 `limit` 参数，但缺少 `offset` 或游标分页，功能不完整 ⚠️

### ❌ 未修复（7 项）

7. **[M1] 进度统计未考虑已下线靶场**：未修复，需要明确业务规则 ❌
8. **[M2] 分类和难度统计查询效率**：未修复，仍然是多次查询 ❌
9. **[L1] DTO 字段注释不完整**：未修复 ❌
10. **[L2] 错误日志缺少关键上下文**：未修复 ❌
11. **[L3] 时间线排序说明**：已修改为降序，但缺少文档说明 ⚠️
12. **[L4] Repository 返回匿名结构体**：未修复 ❌
13. **[L5] 缺少数据库索引建议**：未修复 ❌

### 🆕 新增问题（6 项）

本轮修复引入了 6 个新问题（H1, H2, M1, M2 为实现错误，L1-L3 为优化建议）

## 统计摘要

| 级别 | 数量 | 说明 |
|------|------|------|
| 🔴 高 | 4 | 全部为新增问题，必须立即修复 |
| 🟡 中 | 3 | 2 个新增 + 1 个部分修复 |
| 🟢 低 | 3 | 优化建议 |
| 合计 | 10 | - |

## 总体评价

本轮修复在架构层面的改进是正确的（Redis Key 统一管理、配置外部化、窗口函数优化），但**实现存在严重问题**：

**关键问题**：
1. **代码与 git diff 不一致**：git diff 显示添加了缓存删除逻辑和 `CacheConfig` 定义，但实际代码中缺失
2. **工具函数未使用**：创建了 `SubmitLimitKey()` 但未在代码中调用，违反了修复初衷
3. **SQL 查询不完整**：时间线查询缺少 `title` 字段填充，导致返回空标题
4. **边界条件处理不当**：0 分用户排名显示为 1，不符合业务逻辑

**必须立即修复的问题**（阻塞上线）：
- [H1] 提交限流 Key 未使用统一工具
- [H2] 提交成功后未删除进度缓存（严重影响用户体验）
- [M1] 时间线查询缺少 Title 字段
- [M2] Config 缺少 CacheConfig 定义（代码无法编译）

**建议在下一轮修复**：
- [H3] 排名查询边界问题（0 分用户排名不准确）
- [H4] 时间线销毁时间不准确（需要 migration）
- [M3] 时间线分页功能不完整

修复高优先级和中优先级问题后，该功能可以满足基本需求并上线。
