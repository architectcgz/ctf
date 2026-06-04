# Student Analysis Surface Contract Convergence Plan

> 状态：Current
> 事实源：student-analysis shared surface primitive、workspace/review section consumer、teacher page tests

## 目标

- 把 `student-analysis` 这组 section 的 surface owner 收口成局部共享 contract。
- 统一 recommendations / writeups / evidence / manual-review 的 `loading / empty / content` 三态承接方式。
- 把 overview / timeline 的 loading glass 迁到同一共享 primitive，避免继续在各 section 内复制玻璃壳与 shimmer 动画。

## 非目标

- 不扩展到 topology、review archive、dashboard 或其他教师端页面。
- 不改 `StudentAnalysisPage.vue`、route sync、review workflow、接口参数或业务文案。
- 不重写 `TrainingTimelinePanel.vue` 的 content surface owner；本轮只统一它的 loading owner。

## 输入与现状

- `StudentInsightRecommendationsSection.vue` 已经接近正确模式：同一容器承接 loading / empty / list。
- `StudentInsightOverviewSection.vue`、`StudentInsightPrimarySections.vue`、`StudentInsightWriteupsSection.vue`、`StudentInsightAttackSessionsSection.vue`、`StudentInsightManualReviewSection.vue` 仍各自维护局部 glass shell 和 shimmer。
- `StudentInsightManualReviewSection.vue` 既有 section surface，又有右侧 detail surface，两层 owner 目前都不统一，是最容易继续漂移的区域。

## 架构收口判断

- 共享 owner 应落在 `features/teaching/student-analysis-*` 的共同语义边界，而不是继续散落在 workspace / review 两组 feature 内。
- 本轮采用 `student-analysis-shared` 局部 primitive，避免把本次视觉状态契约过早抬进全局 `shared/ui`。
- touched surface 上的已知结构债是 “三态 owner 分散 + glass shell 重复实现”；本轮必须在 student-analysis touched surface 内收口，不能只替换色值。

## 阶段切片

### Slice 1: 建立 student-analysis 共享 surface primitive

**Files**
- Create: `code/frontend/src/features/teaching/student-analysis-shared/ui/StudentInsightLoadingSurface.vue`
- Create: `code/frontend/src/features/teaching/student-analysis-shared/ui/StudentInsightStateSurface.vue`
- Create: `code/frontend/src/features/teaching/student-analysis-shared/ui/studentInsightSurface.css`
- Create: `code/frontend/src/features/teaching/student-analysis-shared/ui/index.ts`

- [ ] Step 1: 抽出统一 glass shell、highlight overlay、skeleton shimmer、reduced-motion contract。
- [ ] Step 2: 提供可复用的 loading-only surface 与 loading/empty/content state surface。
- [ ] Step 3: 保持 contract 只服务 student-analysis 语义，不扩到全局 shared。

**Validation**
- 读取新 primitive，确认 shell / shimmer / reduced-motion 是单一 owner。

### Slice 2: 迁移 workspace primary sections

**Files**
- Modify: `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightOverviewSection.vue`
- Modify: `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue`
- Modify: `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPrimarySections.vue`

- [ ] Step 1: 把 overview 两个 loading glass box 改成共享 loading surface。
- [ ] Step 2: 把 recommendations 迁成共享 state surface，同时保留当前列表容器作为真实 shape owner。
- [ ] Step 3: 把 timeline loading 从 `StudentInsightPrimarySections.vue` 的本地 `insight-timeline-glass` 收口到共享 primitive。

**Validation**
- `TeacherStudentAnalysis.test.ts` 覆盖 recommendations / timeline 的激活与渲染路径。

### Slice 3: 迁移 review sections

**Files**
- Modify: `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue`
- Modify: `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue`
- Modify: `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue`

- [ ] Step 1: 把 writeups 改成统一 state surface，不再自带独立 `writeup-glass` owner。
- [ ] Step 2: 把 evidence loading shell 改成共享 state surface，保留 `StudentReviewWorkspace` 作为 content owner。
- [ ] Step 3: 把 manual-review 的外层 section 和右侧 detail 两层状态壳体都改成共享 contract。

**Validation**
- `TeacherStudentAnalysis.test.ts` 覆盖 writeups / evidence / manual-review 切换与内容显示。

### Slice 4: 测试与收尾

**Files**
- Modify: `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- Modify: `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

- [ ] Step 1: 更新 raw-source 断言，确认 student-analysis 改为共享 surface contract，而不是继续散落多套 `*-glass` CSS。
- [ ] Step 2: 补齐 recommendations / writeups / evidence / manual-review 的三态 contract 断言。

**Validation**
- `cd code/frontend && pnpm exec vitest run src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `cd code/frontend && pnpm typecheck`

## 兼容性与回退

- 如果某个 section 的内容 owner 与共享 state surface 冲突，优先保留原有 content owner，只把 loading / empty owner 迁到共享 primitive。
- 如果 timeline 的 content shell 无法无副作用迁移，本轮不强行二次包裹 `TrainingTimelinePanel.vue`，避免出现双层 surface。

## Review 重点

- 共享 primitive 是否真的减少了 glass shell / shimmer 重复，而不是换个目录继续复制。
- recommendations / writeups / evidence / manual-review 是否都明确了同一个 state owner。
- manual-review 右侧 detail 是否不再是本地散装 loading/empty/detail 模式。
- 测试是否锁住了 “共享 contract 已建立” 这一层，而不是只锁某个 class 字面量。

## 完成判定

- student-analysis 相关 section 不再各自维护独立玻璃壳背景和 skeleton sweep 动画。
- 至少 recommendations / writeups / evidence / manual-review 四块走共享 state surface。
- overview / timeline 的 loading surface 改走共享 primitive。
- 相关教师端测试与 typecheck 通过。
