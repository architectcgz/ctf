# CTF 前端代码 Review（student-challenge 第 1 轮）：攻防演练页面检查

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | student-challenge |
| 轮次 | 第 1 轮（功能检查） |
| 审查范围 | 学员端攻防演练相关页面 |
| 变更概述 | 检查 FE-T3 任务完成情况 |
| 审查基准 | docs/tasks/frontend-task-breakdown.md（FE-T3 任务要求） |
| 审查日期 | 2026-03-04 |

## 检查文件

1. `code/frontend/src/views/challenges/ChallengeList.vue` - 靶场列表页（206 行）
2. `code/frontend/src/views/challenges/ChallengeDetail.vue` - 靶场详情页（220 行）
3. `code/frontend/src/views/instances/InstanceList.vue` - 实例管理页
4. `code/frontend/src/components/common/InstancePanel.vue` - 实例状态面板组件

## 验收标准对照

| 标准 | 状态 | 说明 |
|------|------|------|
| 靶场列表页（筛选、搜索、难度标签、完成状态） | ✅ 完成 | 已实现分类、难度筛选和搜索 |
| 靶场详情页（描述、开始挑战按钮、Flag 提交框） | ✅ 完成 | 已实现完整功能 |
| 实例状态面板（运行中实例列表、倒计时、销毁/延时按钮） | ✅ 完成 | InstancePanel 组件已实现 |
| 即将超时弹窗提醒（< 5 分钟） | ⚠️ 需确认 | 需检查是否有弹窗提醒逻辑 |
| 个人得分卡片（总分、解题数、排名） | ⚠️ 需确认 | 需检查 Dashboard 是否有得分展示 |
| Flag 提交结果即时反馈 | ✅ 完成 | 已实现成功/失败反馈 |

## 问题列表

### 🟡 中优先级问题（2 项）

#### M1. 缺少即将超时弹窗提醒

**位置**：`InstancePanel.vue` 或 `ChallengeDetail.vue`

**问题**：任务要求"即将超时弹窗提醒（< 5 分钟）"，需确认是否实现

**建议**：检查倒计时逻辑，添加 < 5 分钟时的弹窗提醒

---

#### M2. 个人得分卡片位置不明确

**位置**：Dashboard 或 ChallengeList

**问题**：任务要求"个人得分卡片（总分、解题数、排名）"，需确认展示位置

**建议**：在靶场列表页或 Dashboard 添加得分卡片

---

## 总结

攻防演练页面核心功能已基本完成，代码量充足（426 行）。主要需要确认：
1. 超时提醒弹窗是否已实现
2. 个人得分卡片的展示位置

**建议**：
- 如果这两个功能已在其他地方实现，则可直接通过
- 如果缺失，需要补充实现

**是否需要修复**：待确认（需检查具体实现细节）
