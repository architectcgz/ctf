> 状态：Current
> 事实源：前端架构门禁脚本、AWD owner boundary tests、growth guard baseline
> 替代：无

# Frontend Architecture Growth Guards Plan

## 目标

- 为前端热点文件增加机械化 growth guard，防止文件逐步长大
- 把 `contest-awd-admin` 最近完成的 owner split 变成全局 architecture gate 的一部分
- 保持现有分层 / route view / overlay / theme checks 不变

## 非目标

- 本轮不为全仓库所有 feature 建立 baseline
- 本轮不把历史 extraction tests 全量并入 architecture gate
- 本轮不修改运行时业务逻辑，只增强 guardrail

## 输入依据

- `scripts/check-architecture.sh`
- `scripts/check-workflow-complete.sh`
- `docs/architecture/README.md`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- `code/frontend/src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts`
- `code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwdBoundary.test.ts`
- `code/frontend/src/features/contest-awd-admin/model/useAwdContestStateFlags.ts`
- `code/frontend/src/features/contest-awd-admin/model/useAwdReadinessDecision.ts`
- `code/frontend/src/features/contest-awd-admin/model/useAwdRoundOperations.ts`
- `code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwd.ts`

## 当前结论

- 现有全局前端架构脚本能守通用层级边界，但不能守热点文件逐步膨胀。
- AWD 线虽然已经有 extraction / boundary tests，但还停留在局部测试，未进入 `check-architecture.sh --full`。
- 当前最值得先建立 baseline 的是最近连续拆分、且后续最容易继续长回去的 `contest-awd-admin` owner surfaces。

## 设计边界

### Growth guard 本轮负责

- 热点文件当前行数基线
- 单次增长预算 `max_growth`
- 热点文件绝对上限 `max_lines`

### AWD owner boundary tests 本轮负责

- `useAwdContestStateFlags.ts` 继续持有 runtime policy owner
- `useAwdReadinessDecision.ts` 继续持有 readiness override workflow owner
- `useAwdRoundOperations.ts` 不重新内联 runtime/readiness gate
- `usePlatformContestAwd.ts` 继续作为 owner composition root，而不是重新内联子逻辑

### 本轮不负责

- 其它 feature 的 owner boundary
- 把 raw-source guard 升级成统一框架

## 任务切片

### Slice 1：新增 growth guard

- 新增：
  - `code/frontend/scripts/check-frontend-growth-guard.mjs`
  - `code/frontend/scripts/frontend-growth-baseline.json`
- 更新：
  - `code/frontend/package.json`
- 目标：
  - 守住 AWD 热点文件当前体量和小幅增长预算

### Slice 2：新增 AWD owner boundary tests 并接入 architecture gate

- 新增：
  - `code/frontend/src/features/contest-awd-admin/model/useAwdOwnerBoundaries.test.ts`
- 更新：
  - `scripts/check-architecture.sh`
  - `scripts/check-workflow-complete.sh`
- 目标：
  - 让 AWD owner split 回退能在 `--full` 被机械发现

### Slice 3：同步文档入口与 review

- 更新：
  - `docs/architecture/README.md`
  - `docs/reviews/frontend/2026-05-30-frontend-architecture-growth-guards-review.md`

## 验证

- `cd code/frontend && npm run test:run -- src/features/contest-awd-admin/model/useAwdOwnerBoundaries.test.ts src/features/contest-awd-admin/model/usePlatformContestAwdBoundary.test.ts`
- `cd code/frontend && npm run check:frontend-growth`
- `bash scripts/check-architecture.sh --full`
- `cd code/frontend && npm run typecheck`
- `git diff --check`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

## 残余风险

- growth baseline 是显式文件，后续在完成真实拆分后需要人工下调，不会自动漂移。
- 这轮只守 AWD 热点文件；如果其它 surface 重新成为热点，需要后续再追加 baseline。
