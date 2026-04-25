# CTF 平台代码 Review（progress 第 4 轮）：缓存删除错误处理修复审查

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | progress |
| 轮次 | 第 4 轮（round 3 修复后复审） |
| 审查范围 | commit 2c26ed0，1 个文件，+3/-1 行 |
| 变更概述 | 修复缓存删除未检查错误返回值问题（[H1]） |
| 审查基准 | /home/azhi/workspace/projects/ctf/CLAUDE.md |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | 5（高 1，中 1，低 3） |

## 问题清单

### 🟡 中优先级

#### [M1] 时间线查询仍使用 updated_at 作为销毁时间
- **文件**：`code/backend/internal/module/practice/repository.go:183`
- **问题描述**：round 3 [M1] 问题未修复，`instance_destroy` 事件仍使用 `i.updated_at` 作为时间戳，这是不准确的
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

### 🟢 低优先级

#### [L1] 窗口函数查询缺少索引优化
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

#### [L2] Repository 返回匿名结构体
- **文件**：`code/backend/internal/module/practice/repository.go:147-156`
- **问题描述**：`GetUserTimeline` 返回匿名结构体切片，不利于代码复用和类型安全
- **影响范围/风险**：
  - 其他地方需要使用相同结构时无法复用
  - IDE 无法提供良好的类型提示
  - 代码可读性差
- **修正建议**：
```go
// 在 repository.go 顶部定义
type TimelineEvent struct {
    Type        string
    ChallengeID int64
    Title       string
    Timestamp   time.Time
    IsCorrect   *bool
    Points      *int
}

// 修改方法签名
func (r *Repository) GetUserTimeline(userID int64, limit, offset int) ([]TimelineEvent, error) {
    var events []TimelineEvent
    // ... 查询逻辑
    return events, err
}
```

#### [L3] 时间线排序缺少文档说明
- **文件**：`code/backend/internal/module/practice/handler.go:66-71`
- **问题描述**：时间线按时间降序排列（最新在前），但未在 DTO 或 API 文档中说明排序规则
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

## Round 3 问题修复情况总结

### ✅ 已完全修复（1 项）

1. **[H1] 缓存删除未检查错误返回值**：已添加 `Err()` 检查和 Warn 日志 ✅

### ❌ 未修复（4 项）

2. **[M1] 时间线查询仍使用 updated_at**：未修复，需要 migration 添加 `destroyed_at` 字段 ❌
3. **[L1] 窗口函数查询缺少索引优化**：未修复 ❌
4. **[L2] Repository 返回匿名结构体**：未修复 ❌
5. **[L3] 时间线排序缺少文档说明**：未修复 ❌

### 🆕 新增问题

本轮修复未引入新问题 ✅

## 统计摘要

| 级别 | 数量 | 说明 |
|------|------|------|
| 🔴 高 | 0 | 无阻塞性问题 |
| 🟡 中 | 1 | 时间线销毁时间不准确（需要 migration） |
| 🟢 低 | 3 | 优化建议 |
| 合计 | 4 | - |

## 总体评价

**✅ 本轮修复质量优秀，[H1] 问题已完全解决，代码可以合并上线。**

**修复验证**：
- ✅ **错误处理完整**：使用 `s.redis.Del(ctx, cacheKey).Err()` 正确检查返回值
- ✅ **日志记录规范**：使用 `s.logger.Warn()` 记录删除失败，包含 `userID` 和 `error` 信息
- ✅ **不影响主流程**：缓存删除失败不会中断 Flag 提交流程，符合降级设计原则
- ✅ **代码风格一致**：与项目其他缓存操作的错误处理方式保持一致

**剩余问题分析**：
- **[M1] 时间线销毁时间不准确**：需要数据库 migration，属于架构层面改动，建议在后续迭代中单独处理
- **[L1-L3] 低优先级优化**：索引优化、类型定义、文档说明等，不影响功能正确性，可在代码重构时统一处理

**上线建议**：
- ✅ **可以立即合并上线**：所有高优先级问题已修复，功能完整且稳定
- 📋 **后续优化计划**：
  1. 创建 migration 添加 `instances.destroyed_at` 字段（[M1]）
  2. 添加数据库索引优化查询性能（[L1]）
  3. 重构 Repository 使用命名类型（[L2]）
  4. 补充 API 文档说明（[L3]）

**代码质量评分**：9.5/10（相比 round 3 的 8.5/10 进一步提升）

**合并状态**：✅ **通过审查，可以合并到主分支**
