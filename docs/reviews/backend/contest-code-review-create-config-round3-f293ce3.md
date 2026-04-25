# 竞赛模块代码 Review（create-config 第 3 轮）：Round 2 修复复审

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | create-config |
| 轮次 | 第 3 轮（Round 2 修复后复审） |
| 审查范围 | dbcd4fa, f293ce3（2 个 commit），3 个文件，33 行新增，15 行删除 |
| 变更概述 | 修复 round 2 中的 H1-H2 高优先级和 M1-M3 中优先级问题 |
| 审查基准 | Round 2 review 报告 + CTF 平台开发规范 |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | 9（高 2，中 3，低 4） |

## 问题清单

### 🔴 高优先级

#### [H1] ListByStatusesAndTimeRange 查询逻辑错误，导致 registration 状态竞赛被遗漏
- **文件**：`code/backend/internal/module/contest/repository.go:75-78`
- **问题描述**：查询条件 `start_time <= ? AND end_time > ?` 存在严重逻辑错误：
  - **registration 状态**：`start_time` 在未来，`start_time <= now` 为 false，导致这些竞赛被完全排除
  - **running 状态**：`start_time <= now AND end_time > now` 正确
  - 结果：所有 registration 状态的竞赛都不会被定时任务扫描，无法自动流转到 running 状态
- **影响范围/风险**：
  - **功能完全失效**：registration → running 的自动流转永远不会触发
  - 竞赛到达开始时间后仍停留在 registration 状态
  - 用户无法参赛，运营事故
- **修正建议**：
```go
func (r *repository) ListByStatusesAndTimeRange(ctx context.Context, statuses []string, now time.Time, offset, limit int) ([]*model.Contest, int64, error) {
    var contests []*model.Contest
    var total int64

    // registration: 需要检查 start_time <= now（到达开始时间）
    // running: 需要检查 end_time > now（未结束）
    query := r.db.WithContext(ctx).Model(&model.Contest{}).
        Where("(status = ? AND start_time <= ?) OR (status = ? AND end_time > ?)",
            model.ContestStatusRegistration, now,
            model.ContestStatusRunning, now)

    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    err := query.Offset(offset).Limit(limit).Find(&contests).Error
    return contests, total, err
}
```

### 🟡 中优先级

#### [M1] UpdateStatus 幂等性实现性能低下
- **文件**：`code/backend/internal/module/contest/repository.go:88-108`
- **问题描述**：每次调用都先执行 `First` 查询，即使状态相同也会产生一次数据库查询。定时任务每分钟扫描所有竞赛，大部分情况下状态相同，导致大量无效查询。
- **影响范围/风险**：
  - 数据库负载增加（每次定时任务执行都会产生 N 次查询，N 为竞赛数量）
  - 性能优化不彻底
- **修正建议**：
```go
func (r *repository) UpdateStatus(ctx context.Context, id int64, status string) error {
    // 直接更新，利用 WHERE 条件过滤
    result := r.db.WithContext(ctx).Model(&model.Contest{}).
        Where("id = ? AND status != ?", id, status).
        Update("status", status)

    if result.Error != nil {
        return result.Error
    }

    // RowsAffected = 0 可能是：1) 状态已相同（幂等成功） 2) 记录不存在
    // 由于定时任务只处理已存在的竞赛，这里可以直接返回成功
    // 如果需要严格区分，可以在 RowsAffected = 0 时再查询一次
    return nil
}
```

### 🟢 低优先级

#### [L1] registration 状态允许修改结束时间可能导致业务问题
- **文件**：`code/backend/internal/module/contest/service.go:93-105`
- **问题描述**：当前逻辑允许在 registration 状态修改结束时间，但这可能导致：
  - 竞赛已经开始（running），但结束时间被提前，导致正在进行的竞赛突然结束
  - 虽然代码禁止修改开始时间，但结束时间的修改同样影响用户体验
- **影响范围/风险**：
  - 边界情况下的用户体验问题
  - 运营风险（虽然概率较低）
- **修正建议**：
```go
// 更严格的方案：registration 状态只允许延长结束时间
if contest.Status == model.ContestStatusRegistration {
    if req.StartTime != nil {
        return nil, errcode.ErrContestAlreadyStarted
    }
    if req.EndTime != nil && req.EndTime.Before(contest.EndTime) {
        return nil, errcode.New(14006, "报名阶段只允许延长结束时间", http.StatusForbidden)
    }
}
```

## Round 2 问题修复情况

### ✅ 已正确修复

| 问题编号 | 问题描述 | 修复质量 |
|---------|---------|---------|
| H1 | UpdateStatus 幂等性检查逻辑有误 | ✅ 已修复（但性能可优化，见本轮 M1） |
| H2 | 状态流转校验时机错误 | ✅ 已修复（L86-91, L132-135） |
| M1 | registration 状态禁止修改开始时间 | ✅ 已修复（L93-98） |
| M2 | 错误码语义不准确 | ✅ 已修复（新增 ErrCannotModifyAfterDraft） |
| M3 | 定时任务查询条件不够精确 | ⚠️ 部分修复，但引入新问题（见本轮 H1） |

### ⏸️ 未修复（低优先级）

| 问题编号 | 问题描述 | 说明 |
|---------|---------|------|
| L2 | 配置项缺少合理性校验 | 未修复，建议后续补充 |
| L3 | StatusUpdater 初始化时缺少参数校验 | 未修复，建议后续补充 |
| L4 | calculateStatus 对 draft 状态的处理不够明确 | 未修复，但 M3 已排除 draft 状态 |

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 1 |
| 🟡 中 | 1 |
| 🟢 低 | 1 |
| 合计 | 3 |

## 总体评价

**修复质量**：Round 2 的 5 个高/中优先级问题中，4 个已正确修复，1 个（M3）修复时引入了严重的逻辑错误。

**核心问题**：
1. **H1（严重）**：`ListByStatusesAndTimeRange` 的查询条件错误，导致 registration 状态竞赛无法被定时任务扫描，功能完全失效
2. **M1（性能）**：幂等性实现虽然正确，但每次都查询数据库，性能不佳

**优点**：
- H1（幂等性）已正确实现，状态相同时返回成功
- H2（状态流转校验时机）已修复，所有校验通过后才设置状态
- M1（registration 禁止修改开始时间）已正确实现
- M2（错误码语义）已优化，新增 `ErrCannotModifyAfterDraft`
- M3（排除 draft 状态）已实现，定时任务不再扫描 draft 竞赛

**建议**：
- **必须立即修复 H1**：当前代码会导致 registration → running 流转完全失效，这是功能性 bug
- **建议优化 M1**：减少不必要的数据库查询
- L1 可选修复，L2-L4 可延后处理

**下一步**：修复本轮发现的 H1 问题后，代码质量将达到可合并标准。
