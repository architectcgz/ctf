> 状态：Current
> 事实源：3 个 `utils/*` 文件与其现有 consumer
> 替代：无

# Utility Owner Realignment Plan

## 目标

- 把 `contest.ts` 迁到明确的 `entities/contest` owner。
- 把 `skillProfile.ts` 迁到明确的 `entities/skill-profile` owner。
- 把 `platformContestAwdChallengeLinks.ts` 迁到明确的 `entities/contest-awd-challenge-link` owner。
- 删除旧 `utils/*` 文件，不保留 bridge。

## 非目标

- 本轮不再改这些模块的 contract 设计。
- 本轮不处理 `featureRouterImportAllowlist` 与 `composableMultiBoundaryAllowlist`。
- 本轮不进一步把 AWD challenge link entity 再拆成更细的 `presentation / normalization / dto` 多文件结构。

## 输入依据

- `code/frontend/src/utils/contest.ts`
- `code/frontend/src/utils/skillProfile.ts`
- `code/frontend/src/utils/platformContestAwdChallengeLinks.ts`
- 对应 API / feature / component / view consumer

## 当前结论

- `contest.ts` 和 `skillProfile.ts` 语义上都属于领域展示 / 归一化 owner，不应该继续停在 `utils/`。
- `platformContestAwdChallengeLinks.ts` 是跨 `contest-awd-admin` 与 `contest-workbench` 的共享 mapper，更适合作为 entity owner，而不是 feature 私有 helper 或 utils。

## 设计边界

### `entities/contest` 本轮负责

- contest mode / status 的展示标签、可见性规则、强调色

### `entities/skill-profile` 本轮负责

- skill profile / recommendation 的 raw normalize
- 弱项标签和雷达图展示辅助

### `entities/contest-awd-challenge-link` 本轮负责

- AWD service 到 challenge link 的映射与最小字段 contract

### consumer 本轮约定

- 统一改为 import 新 entity public API
- 不保留 `utils/*` re-export bridge

## 任务切片

### Slice 1：contest / skill-profile owner 迁移

- 目标：
  - 建立 `entities/contest`、`entities/skill-profile`
  - 改掉相关 consumer import
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/contests/__tests__/ContestDetail.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - 旧 util 路径是否已经退出主路径
  - entity public API 是否足够稳定

### Slice 2：AWD challenge link owner 迁移

- 目标：
  - 建立 `entities/contest-awd-challenge-link`
  - 改掉 AWD admin / contest workbench 的 mapper import
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - cross-feature mapper 是否放在合适 owner
  - 不引入新的 feature 到 feature 深链

### Slice 3：删除旧 util 文件并收尾

- 目标：
  - 删除 `utils/contest.ts`
  - 删除 `utils/skillProfile.ts`
  - 删除 `utils/platformContestAwdChallengeLinks.ts`
  - 更新 backlog 与 review
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/contests/__tests__/ContestDetail.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/utility-owner-realignment.md docs/plan/impl-plan/2026-05-29-utility-owner-realignment-plan.md docs/reviews/frontend/2026-05-29-utility-owner-realignment-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/entities/contest code/frontend/src/entities/skill-profile code/frontend/src/entities/contest-awd-challenge-link code/frontend/src/api/assessment.ts code/frontend/src/api/teacher/students.ts code/frontend/src/api/teaching/students.ts code/frontend/src/components/contests/ContestOverviewPanel.vue code/frontend/src/components/scoreboard/ScoreboardWorkspaceShell.vue code/frontend/src/components/teacher/student-insight/StudentInsightOverviewSection.vue code/frontend/src/features/contest-awd-admin/model/useAwdContestSnapshotLoader.ts code/frontend/src/features/contest-awd-admin/model/useAwdChallengeLinkOperations.ts code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts code/frontend/src/features/contest-detail/model/useContestListPage.ts code/frontend/src/features/contest-workbench/model/useContestChallengeOrchestration.ts code/frontend/src/features/contest-workbench/model/useContestEditAwdWorkspace.ts code/frontend/src/features/platform-contests/ui/ContestOperationsHubWorkspacePanel.vue code/frontend/src/features/platform-contests/ui/PlatformContestTable.vue code/frontend/src/features/platform-contests/ui/PlatformContestRulesSection.vue code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts code/frontend/src/features/scoreboard/model/useScoreboardContestDirectoryPage.ts code/frontend/src/features/skill-profile/model/useSkillProfilePage.ts code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisDataState.ts code/frontend/src/features/student-dashboard/model/useStudentDashboardData.ts code/frontend/src/features/teacher-workspace/model/useWorkspace.ts code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts code/frontend/src/views/contests/ContestList.vue code/frontend/src/__tests__/architectureBoundaries.test.ts code/frontend/src/utils/contest.ts code/frontend/src/utils/skillProfile.ts code/frontend/src/utils/platformContestAwdChallengeLinks.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 本轮删除旧 util 文件后，仓库内如果还有漏改的 import，会直接在 typecheck 暴露。
- 这轮 review 默认仍是同上下文 self-review，独立 reviewer gate 仍需单独说明。

## 实施记录

- [x] Slice 1：已建立 `entities/contest` 与 `entities/skill-profile`，并改掉现有 consumer import。
- [x] Slice 2：已建立 `entities/contest-awd-challenge-link`，AWD admin / contest workbench 的 mapper import 已统一改到新 owner。
- [x] Slice 3：`utils/contest.ts`、`utils/skillProfile.ts`、`utils/platformContestAwdChallengeLinks.ts` 已删除，backlog 与 review 文档已同步。
