# 2026-06-01 Contest Domain Convergence Review

> 当前状态：Current
> 本次结论：Pass
> 说明：本文件覆盖同一 worktree 里更早一版 `blocked` 判断。更早阻塞点是“运行时 consumer 仍把 `@/api/admin/contests` 当真实 owner”，该判断已被本次当前代码复核推翻；文档与实现关于 `contest-reviews` owner 的表述偏差也已在本 worktree 修正。

## Review Target

- Repository: `ctf`
- Worktree: `ctf/.claude/worktrees/contest-domain-convergence`
- Branch: `feat/contest-domain-convergence`
- Diff source: worktree working tree changes against `main`
- Files reviewed:
  - `code/frontend/src/api/admin/contests.ts`
  - `code/frontend/src/api/admin/index.ts`
  - `code/frontend/src/api/admin/contest-manage.ts`
  - `code/frontend/src/api/admin/contest-announcements.ts`
  - `code/frontend/src/api/admin/contest-operations.ts`
  - `code/frontend/src/api/admin/contest-awd-admin.ts`
  - `code/frontend/src/api/admin/contest-reviews.ts`
  - `code/frontend/src/api/__tests__/admin.test.ts`
  - `code/frontend/src/features/__tests__/featureBoundaries.test.ts`
  - `code/frontend/src/features/platform/contest-manage/model/useContestEditPage.ts`
  - `code/frontend/src/features/platform/contest-announcements/model/useContestAnnouncementsData.ts`
  - `code/frontend/src/features/platform/contest-operations/model/useContestOperationsData.ts`
  - `code/frontend/src/pages/platform/contests/*.vue`
  - `TODO/frontend-sliced-architecture.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Classification Check

- Classification: architecture review
- Decision: agree with non-trivial review classification

## Gate Verdict

- Verdict: pass

## Findings

- 无。上一版唯一成立的 P2 文档失真已修正。

## Material Findings

- 无。当前没有阻塞或非阻塞的 correctness / regression / touched-surface debt finding。

## Senior Implementation Assessment

- 这版比上一轮明显更接近真正的 owner convergence：`features/platform/contests/` 兼容 barrel 已删除，运行时 feature consumer 也已切到新的 owning API modules。
- `api/admin/contests.ts` 现在只剩 test/mock 兼容层，作为短期过渡壳是可接受的；它不再是运行时主 owner。
- `contest-reviews` 这条线仍然复用 `teaching` 实现层，但文档现在已经如实描述成“platform public owner + teaching implementation reuse”，当前表达和代码一致。

## Required Re-validation

- 无。

## Validation Evidence

- 运行时 owner 复核：
  - `cd code/frontend && rg -n "from '@/api/admin/contests'|from \"@/api/admin/contests\"" src/features src/api src/pages`
  - 结果：运行时代码未再命中旧 barrel；仅测试源码仍保留兼容断言或兼容 import。
- Targeted tests:
  - `cd code/frontend && timeout 180s npm run test:run -- src/features/__tests__/featureBoundaries.test.ts src/pages/platform/contests/__tests__/ContestManage.test.ts src/api/__tests__/admin.test.ts`
  - 结果：3 个 test files，70 个 tests，全部通过。
- Typecheck:
  - `cd code/frontend && timeout 180s npm run typecheck`
  - 结果：通过。
- Growth guard:
  - `cd code/frontend && timeout 180s npm run check:frontend-growth`
  - 结果：`[frontend-growth] guard passed`。
- 文档事实核对：
  - `rg -n "platform public owner|teaching/awd-reviews.ts|contest-reviews" TODO docs/reviews/architecture code/frontend/src/api/admin`
  - 结果：`TODO/frontend-sliced-architecture.md` 已改成与 `code/frontend/src/api/admin/contest-reviews.ts` 一致的表述，review 归档也同步记录为已修正状态。

## Residual Risk

- `code/frontend/src/api/__tests__/admin.test.ts:81` 仍显式从 `@/api/admin/contests` 导入，这是有意保留的兼容 coverage，不是运行时回流；但也意味着删除该 deprecated barrel 前还需要同步迁移测试 mock。
- 本次收口只修正文档事实，没有重跑整组 contest 相关页面测试矩阵。

## Touched Known-Debt Status

- This diff touches a previously tracked oversized / owner-mixed surface (`api/admin/contests.ts`).
- Current status: runtime owner convergence 已在 touched surface 内收口；旧 barrel 降为 test/mock compatibility layer。
- Remaining debt: `contest-reviews` 仍复用 `teaching` implementation owner；当前这不再是文档失真问题，而是一个已知、被如实记录的实现事实。
