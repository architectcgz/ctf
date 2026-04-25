# CTF 平台代码 Review（team-management 第 3 轮）：错误码使用问题修复审查

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | team-management |
| 轮次 | 第 3 轮（修复后复审） |
| 审查范围 | commit 2f67c1f，3 个文件，15 行新增，11 行删除 |
| 变更概述 | 修复 round 2 中的 H5/H6 高优先级问题：统一错误码使用和自定义业务错误 |
| 审查基准 | round 2 review 报告 + 项目 CLAUDE.md 规范 |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | 高 2 个，中 4 个，低 2 个，合计 8 个 |

## Round 2 问题修复情况

### 🔴 高优先级问题修复验证

#### [H5] CreateTeam 重试失败后返回通用错误消息 - ✅ 已修复
- **修复内容**：
  - 新增错误码 `ErrInviteCodeGenerationFailed = New(14009, "创建队伍失败，请重试", 500)`（team.go:12）
  - 替换 `errors.New("创建队伍失败")` 为 `errcode.ErrInviteCodeGenerationFailed`（team_service.go:74）
- **验证结果**：✅ 修复正确，符合统一错误码规范

#### [H6] AddMemberWithLock 使用 gorm.ErrInvalidData 表示业务错误 - ✅ 已修复
- **修复内容**：
  - 定义自定义错误 `var ErrTeamFull = errors.New("team is full")`（team_repository.go:12）
  - Repository 返回自定义错误 `return ErrTeamFull`（team_repository.go:82）
  - Service 层判断 `if errors.Is(err, ErrTeamFull)`（team_service.go:98）
- **验证结果**：✅ 修复正确，解耦了业务逻辑与 GORM 内部错误

### 🟡 中优先级问题
- **M7-M9**：未在本轮修复，符合预期（可后续迭代优化）

### 🟢 低优先级问题
- **L7-L8**：未在本轮修复，符合预期（可后续迭代优化）

## 新发现问题清单

### 🟡 中优先级

#### [M10] ErrTeamFull 定义位置不当
- **文件**：`team_repository.go:12`
- **问题描述**：
  - `ErrTeamFull` 定义在 Repository 层（contest 包）
  - Service 层需要导入 contest 包才能使用 `errors.Is(err, ErrTeamFull)`
  - 但 Service 本身也在 contest 包，所以当前可以访问
  - 然而从语义上看，这是一个业务错误，应该定义在更合适的位置
- **影响范围/风险**：
  - 当前代码可以正常工作（同包访问）
  - 但如果未来需要在其他包中判断此错误，会产生循环依赖
  - 错误定义位置不符合最佳实践
- **修正建议**：
```go
// 方案 1：定义在 errcode 包（推荐）
// pkg/errcode/team.go
var (
    ErrTeamFullInternal = errors.New("team is full") // Repository 内部使用
)

// 方案 2：保持现状但添加注释说明
// internal/module/contest/team_repository.go
// ErrTeamFull 是 Repository 层内部错误，用于 Service 层判断队伍已满的情况
var ErrTeamFull = errors.New("team is full")
```
- **建议**：当前实现可以接受，但建议添加注释说明此错误的用途和作用域

## 统计摘要

| 级别 | Round 2 | Round 3 新增 | 已修复 | 未修复 |
|------|---------|--------------|--------|--------|
| 🔴 高 | 2 | 0 | 2 | 0 |
| 🟡 中 | 4 | 1 | 0 | 5 |
| 🟢 低 | 2 | 0 | 0 | 2 |
| 合计 | 8 | 1 | 2 | 7 |

## 总体评价

**✅ Round 2 的所有高优先级问题（H5/H6）均已正确修复，代码质量显著提升。**

### 修复亮点

1. ✅ **错误码统一管理**：新增 `ErrInviteCodeGenerationFailed`，符合项目规范
2. ✅ **业务错误解耦**：使用自定义 `ErrTeamFull` 替代 `gorm.ErrInvalidData`，分层清晰
3. ✅ **错误判断规范**：使用 `errors.Is` 进行错误类型判断，代码健壮

### 剩余问题分析

- 🟡 **M7-M9**（Round 2 遗留）：邀请码生成健壮性、冲突检测可靠性、性能优化
- 🟡 **M10**（新发现）：错误定义位置的最佳实践问题
- 🟢 **L7-L8**（Round 2 遗留）：代码可读性改进

### 合并建议

**✅ 可以合并**

理由：
1. 所有高优先级问题已修复
2. 核心功能（邀请码生成、并发安全、事务一致性、错误处理）均已达到生产标准
3. 剩余问题均为优化性质，不影响功能正确性和系统稳定性
4. M10 问题当前实现可以正常工作，仅是最佳实践层面的建议

### 后续优化建议

可在后续迭代中处理：
- M7：增强邀请码生成的长度保证
- M8：改进唯一索引冲突检测的跨数据库兼容性
- M9：优化 LeaveTeam 的成员检查性能
- M10：为 ErrTeamFull 添加注释说明或考虑重构到 errcode 包
- L7/L8：提升代码可读性
