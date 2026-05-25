# 2026-05-25 Student Insight Panel Owner Convergence Review

## Review Target

- Repository: `ctf`
- Branch / workspace: `main` on `/home/azhi/workspace/projects/ctf`
- Diff source: commit `65859722279c71a0cfde03d305aa44530f740d4c..3e6a4d462d1ab3f83ffb2d422d5c40fa7766f54e`
- Files reviewed:
  - `code/frontend/src/components/teacher/StudentInsightPanel.vue`
  - `code/frontend/src/components/teacher/student-insight/StudentInsightOverviewSection.vue`
  - `code/frontend/src/components/teacher/student-insight/StudentInsightRecommendationsSection.vue`
  - `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
  - `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
  - `code/frontend/src/views/teacher/__tests__/studentAnalysisPanelExtraction.test.ts`

## Classification Check

- 同意按非平凡前端重构做独立 gate review。
- 该提交同时触碰已知 `TD-1` 债面、页面 tab 挂载方式和测试护栏，必须做独立 review。

## Gate Verdict

- `pass with minor issues`

## Findings

### Non-blocking

1. `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
   - 已补“先点 tab 再断言内容”的正向路径，但当时还没有运行时断言去证明“未激活 tab 的内容不会提前留在 DOM”。
   - 如果后续有人把某个 section 从 `v-if` 回退成 `v-show`，仅靠源码匹配和切 tab 后内容出现，仍可能放过隐藏 DOM 回归。

## Material Findings

- 无

## Senior Implementation Assessment

- 当前实现方向是对的：`StudentInsightPanel` 退回组合层，只保留区块可见性和事件桥接；overview / recommendations 的模板与局部样式 owner 已下沉到新 section。
- `StudentAnalysisPage` 保留 tab owner，并改成真正的单个 `StudentInsightPanel` 挂载，避免了之前“源码只有一个挂载点，但运行时仍是多实例隐藏挂载”的伪单实例结构。
- 更低风险的替代方案不会比当前实现更简单；主要剩余问题只在测试护栏需要再补一层运行时负向断言。

## Required Re-validation

1. 补运行时断言，覆盖“切到 `recommendations` 前，`题解列表` / `POST /login` 等非当前 tab 内容不在 DOM；切换后才出现”。
2. 重新执行：

```bash
npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/teacher/__tests__/studentAnalysisPanelExtraction.test.ts src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts
```

## Residual Risk

- 独立 review 在 clean detached worktree 上完成；主工作树当前很脏，不能把主工作树状态直接当成这笔提交的 review 证据。
- review 本身没有发现 blocker，但如果没有补上运行时负向断言，未来仍可能对隐藏 DOM 回归保护不足。

## Touched Known-Debt Status

- 本次触碰的是 `TD-1` 下 `StudentInsightPanel.vue` 的 section owner 收口面。
- 该 touched surface 当前已完成本阶段收口，不是只把模板搬到别处：
  - `StudentInsightPanel.vue` 已退成组合层。
  - `StudentInsightOverviewSection.vue` / `StudentInsightRecommendationsSection.vue` 承接了原内联模板与局部样式 owner。
  - `StudentAnalysisPage.vue` 维持 tab owner，并改成真正的单实例挂载。

## Follow-up Absorption

- 上述 non-blocking finding 已在当前后续切片中转化为运行时负向断言，避免“隐藏 DOM 不得回归”只靠源码护栏保护。
