# AWD Phase 13 学生复盘页结构化 UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将教师学生复盘页的主体区从双日志布局升级为“练习复盘 + AWD 复盘”的结构化案例档案卡视图，同时保持教师端既有 UI 风格一致。

**Architecture:** 保留现有统一总览层，在前端本地把现有 `timeline / evidence / writeups / manual_reviews` 整理成 `practice cases` 与 `awd cases` 两组视图模型，再由 `ReviewArchiveEvidencePanel` 渲染两个分区与 A1 案例卡。这样不改后端契约，只重构前端阅读结构。

**Tech Stack:** Vue 3, TypeScript, Vue Test Utils, Vitest, teacher-surface theme system

---

## Execution Notes

- 先写 RED 测试，再写最小实现。
- phase13 默认只改前端。
- 继续保持教师端 `teacher-surface`、`section-card`、`metric-panel` 的共享模式。
- 所有验证在当前 worktree `/home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design` 内串行执行。
- `docs/superpowers/*` 被 `.gitignore` 忽略，提交时仍需 `git add -f`。

## Planned File Map

### Docs

- Create: `docs/architecture/features/2026-04-13-awd-phase13-student-review-archive-structured-ui-design.md`
- Create: `docs/superpowers/plans/2026-04-13-awd-phase13-student-review-archive-structured-ui-implementation.md`

### Frontend

- Modify: `code/frontend/src/components/teacher/review-archive/ReviewArchiveEvidencePanel.vue`
- Create: `code/frontend/src/components/teacher/review-archive/reviewArchiveCases.ts`
- Create: `code/frontend/src/components/teacher/review-archive/__tests__/ReviewArchiveEvidencePanel.test.ts`
- Modify: `code/frontend/src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts`
- Optionally modify: `code/frontend/src/api/contracts.ts` only if the panel needs local helper types exported, otherwise保持最小 diff

## Task 1: 为结构化复盘面板补红测

**Files:**
- Create: `code/frontend/src/components/teacher/review-archive/__tests__/ReviewArchiveEvidencePanel.test.ts`

- [ ] **Step 1: 写练习复盘与 AWD 复盘分区红测**

新增至少这些用例：

- `应该将练习事件归入练习复盘区`
- `应该将AWD事件归入AWD复盘区`

至少断言：

- 页面出现 `练习复盘` 与 `AWD 复盘`
- 练习区不会渲染 AWD 目标队伍摘要
- AWD 区会渲染目标队伍与命中/未命中结果

- [ ] **Step 2: 写 A1 案例卡聚合红测**

新增至少这些用例：

- `应该按 challenge 聚合练习案例`
- `应该按 challenge 加 victim team 聚合AWD案例`

至少断言：

- 同一练习题的 timeline/evidence/writeup 会归进同一张卡
- 同一 AWD 题针对不同 victim team 会拆成两张卡

- [ ] **Step 3: 写折叠交互红测**

新增：

- `应该默认折叠案例卡并在展开后显示事件明细`

- [ ] **Step 4: 运行定向命令确认 RED**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/frontend
timeout 120s npm run test:run -- src/components/teacher/review-archive/__tests__/ReviewArchiveEvidencePanel.test.ts
```

## Task 2: 为页面级集成行为补红测

**Files:**
- Modify: `code/frontend/src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts`

- [ ] **Step 1: 扩展 mock 数据，加入 AWD 时间线与证据**

- [ ] **Step 2: 增加页面级断言**

至少断言：

- 页面能渲染两个复盘分区
- 练习与 AWD 两类案例都可见
- 原来的双日志布局关键文案不再作为主结构存在

- [ ] **Step 3: 运行定向命令确认 RED**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/frontend
timeout 120s npm run test:run -- src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts
```

## Task 3: 实现本地视图模型与 A1 案例卡

**Files:**
- Create: `code/frontend/src/components/teacher/review-archive/reviewArchiveCases.ts`
- Modify: `code/frontend/src/components/teacher/review-archive/ReviewArchiveEvidencePanel.vue`

- [ ] **Step 1: 创建案例视图模型整理函数**

输出至少包含：

- `practiceCases`
- `awdCases`

每个 case 至少包含：

- 标题
- 副线
- 状态摘要
- 事件数量
- 最近活动时间
- 阶段摘要
- 展开后的事件明细

- [ ] **Step 2: 用新的案例模型重写复盘面板模板**

至少实现：

- 练习复盘 section
- AWD 复盘 section
- A1 案例档案卡头部摘要
- 默认折叠 / 点击展开

- [ ] **Step 3: 补齐教师端 surface 样式**

要求：

- 使用现有 token
- 不引入亮色硬编码
- AWD 强调只做轻量边界和状态色

- [ ] **Step 4: 重跑组件测试至 GREEN**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/frontend
timeout 120s npm run test:run -- src/components/teacher/review-archive/__tests__/ReviewArchiveEvidencePanel.test.ts
timeout 120s npm run test:run -- src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts
```

## Task 4: 做最小充分回归

**Files:**
- Modify if needed: `code/frontend/src/views/teacher/__tests__/teacherSurface.test.ts`
- Modify if needed: `code/frontend/src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

- [ ] **Step 1: 跑教师复盘相关回归**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/frontend
timeout 120s npm run test:run -- src/components/teacher/review-archive/__tests__/ReviewArchiveEvidencePanel.test.ts
timeout 120s npm run test:run -- src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts
timeout 120s npm run test:run -- src/views/teacher/__tests__/teacherSurface.test.ts
timeout 120s npm run test:run -- src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts
```

- [ ] **Step 2: 视情况补一轮类型检查**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/frontend
timeout 120s npm run typecheck
```

- [ ] **Step 3: 提交 phase13**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design
git add code/frontend/src/components/teacher/review-archive/ReviewArchiveEvidencePanel.vue \
  code/frontend/src/components/teacher/review-archive/reviewArchiveCases.ts \
  code/frontend/src/components/teacher/review-archive/__tests__/ReviewArchiveEvidencePanel.test.ts \
  code/frontend/src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts
git add -f docs/architecture/features/2026-04-13-awd-phase13-student-review-archive-structured-ui-design.md \
  docs/superpowers/plans/2026-04-13-awd-phase13-student-review-archive-structured-ui-implementation.md
git commit -m "feat(teacher): 重构学生复盘结构化视图"
```
