# Student Insight Review Workspace Section Plan

## 背景

当前教师侧学员详情页已经把题解区、人工审核区拆成了独立 section 组件，但“复盘工作台”仍直接在 `StudentInsightPanel.vue` 内引用 `TeacherStudentReviewWorkspace` widget。

这会带来两个问题：

- `StudentInsightPanel.vue` 继续承担过多结构装配职责。
- 三个详情区块里只有复盘区没有统一落到 `student-insight/` 目录，结构不对称。

## 目标

在不回退主线当前能力的前提下，把复盘工作台从 `StudentInsightPanel.vue` 直接依赖 widget，调整为依赖一个新的 section 组件。

## 约束

- 保留主线现有能力：
  - challenge / mode / result 筛选
  - 路由 query 同步
  - `TeacherStudentReviewWorkspace` 现有交互
  - 现有 `useTeacherReviewWorkspace` 数据流
- 不改后端 API、DTO、查询参数或页面路由行为。
- 本次只做前端结构收敛，不做 widget 内部再拆分。

## 实现步骤

1. 新增 `StudentInsightAttackSessionsSection.vue`
   - 负责复盘区标题、副标题和区块挂载。
   - 内部复用 `TeacherStudentReviewWorkspace`。
2. 调整 `StudentInsightPanel.vue`
   - 删除对 `TeacherStudentReviewWorkspace` 的直接引用。
   - 改为挂载 `StudentInsightAttackSessionsSection`。
3. 同步组件声明
   - 更新 `components.d.ts`。
4. 补测试
   - 在教师学员分析页测试中增加结构断言，确保 `StudentInsightPanel` 依赖 section 组件而不是直接依赖 widget。
   - 重跑现有教师分析页测试，确认筛选和交互未回退。

## 验证

最小验证集：

```bash
pnpm vitest run src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts
pnpm vitest run src/widgets/teacher-student-review-workspace/TeacherStudentReviewWorkspace.test.ts
```
