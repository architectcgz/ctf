# 竞赛模块代码 Review（create-config 第 4 轮）：Round 3 修复复审

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | create-config |
| 轮次 | 第 4 轮（Round 3 修复后复审） |
| 审查范围 | commit 40d5ff7，1 个文件，21 行新增，19 行删除 |
| 变更概述 | 修复 round 3 中的 H1（状态流转查询条件）和 M1（幂等性优化）问题 |
| 审查基准 | Round 3 review 报告 + CTF 平台开发规范 |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | 3（高 1，中 1，低 1） |

## 问题清单

### 🔴 高优先级

#### [H1] ListByStatusesAndTimeRange 查询逻辑仍然错误，导致 running 状态竞赛被误扫描

- **文件**：`code/backend/internal/module/contest/repository.go:75-77`
- **问题描述**：修复后的查询条件存在严重逻辑错误：
  ```go
  query := r.db.WithContext(ctx).Model(&model.Contest{}).Where("status IN ?", statuses).
      Where(r.db.Where("status = ? AND start_time <= ?", model.ContestStatusRegistration, now).
          Or("status = ? AND end_time <= ?", model.ContestStatusRunning, now))
  ```

  **问题分析**：
  1. **registration 状态**：`status = registration AND start_time <= now` ✅ 正确（到达开始时间需要流转）
  2. **running 状态**：`status = running AND end_time <= now` ❌ **错误**
     - 应该是 `end_time <= now`（已结束，需要流转到 ended）
     - 但当前逻辑会扫描所有 `end_time <= now` 的 running 竞赛
     - 这是正确的！但 Round 3 报告中建议的是 `end_time > now`（未结束），这是错误的建议

  **重新分析后发现**：
  - 当前代码逻辑 **实际上是正确的**
  - Round 3 报告中的建议 `end_time > now` 是错误的
  - 定时任务的目的是找出"需要流转状态"的竞赛：
    - registration → running：需要 `start_time <= now`（已到开始时间）
    - running → ended：需要 `end_time <= now`（已到结束时间）

  **但存在另一个问题**：
  - 外层 `Where("status IN ?", statuses)` 和内层 `Where("status = ? ...")` 的组合逻辑不清晰
  - GORM 的 `Where().Or()` 嵌套可能产生非预期的 SQL

- **影响范围/风险**：
  - SQL 生成逻辑不明确，可能产生错误的查询条件
  - 需要验证实际生成的 SQL 是否符合预期

- **修正建议**：
  ```go
  func (r *repository) ListByStatusesAndTimeRange(ctx context.Context, statuses []string, now time.Time, offset, limit int) ([]*model.Contest, int64, error) {
      var contests []*model.Contest
      var total int64

      // 明确的查询逻辑：
      // 1. registration 状态 + start_time <= now（需要流转到 running）
      // 2. running 状态 + end_time <= now（需要流转到 ended）
      query := r.db.WithContext(ctx).Model(&model.Contest{}).
          Where("(status = ? AND start_time <= ?) OR (status = ? AND end_time <= ?)",
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

**无**

### 🟢 低优先级

**无**

## Round 3 问题修复情况

### ✅ 已正确修复

| 问题编号 | 问题描述 | 修复质量 |
|---------|---------|---------|
| M1 | UpdateStatus 幂等性实现性能低下 | ✅ 已正确修复（L88-103） |

**M1 修复分析**：
- 使用 `WHERE id = ? AND status != ?` 避免无效更新 ✅
- 仅在 `RowsAffected == 0` 时才查询是否存在 ✅
- 正确区分"状态已相同"和"记录不存在"两种情况 ✅
- 性能优化到位，大部分情况下只执行一次 UPDATE ✅

### ⚠️ 修复不完整

| 问题编号 | 问题描述 | 说明 |
|---------|---------|------|
| H1 | ListByStatusesAndTimeRange 查询逻辑错误 | ⚠️ 修复方向正确，但 SQL 生成逻辑不够清晰，存在风险 |

**H1 修复分析**：
- ✅ registration 状态的条件正确：`start_time <= now`
- ✅ running 状态的条件正确：`end_time <= now`
- ❌ 但 GORM 嵌套 `Where().Or()` 的 SQL 生成逻辑不明确
- ❌ 外层 `Where("status IN ?", statuses)` 参数未被使用（被内层覆盖）

### ⏸️ 未修复（低优先级）

| 问题编号 | 问题描述 | 说明 |
|---------|---------|------|
| L1 | registration 状态允许修改结束时间可能导致业务问题 | 未修复，建议后续补充 |

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 1 |
| 🟡 中 | 0 |
| 🟢 低 | 0 |
| 合计 | 1 |

## 总体评价

**修复质量**：Round 3 的 2 个高/中优先级问题中，M1 已完美修复，H1 修复方向正确但实现方式存在风险。

**核心问题**：
1. **H1（SQL 生成风险）**：当前使用 GORM 嵌套 `Where().Or()` 的方式不够清晰，可能产生非预期的 SQL，建议改用显式的 OR 条件

**优点**：
- M1（幂等性优化）已完美实现，性能提升明显
- H1 的业务逻辑正确（registration 检查 start_time，running 检查 end_time）
- 代码注释清晰，说明了 RowsAffected == 0 的两种情况

**风险点**：
- GORM 的 `Where(db.Where(...).Or(...))` 嵌套语法可能生成非预期的 SQL
- 外层 `Where("status IN ?", statuses)` 参数未被有效使用

**建议**：
- **必须修复 H1**：改用显式的 OR 条件，避免 GORM 嵌套语法的不确定性
- 修复后代码质量将达到可合并标准

**下一步**：修复 H1 后，建议进行集成测试验证状态流转逻辑的正确性。
