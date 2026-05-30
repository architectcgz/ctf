> 状态：In Progress
> 事实源：`code/frontend/src/components/platform/challenge/*`、`code/frontend/src/features/platform/challenge-package-import/**`、`code/frontend/src/entities/challenge/**`
> 替代：无

# Platform Challenge Feature Owner Migration Batch 1

## 目标

- 清理 `components/platform/challenge/*` 这一组历史 owner。
- 将导入页相关壳迁入 `features/platform/challenge-package-import/ui`。
- 将 `ChallengeDescriptionPanel` 下沉为 `entities/challenge/ui` 的共享展示块。

## 非目标

- 不重做 challenge import / preview / format 的页面行为。
- 不改 platform challenge detail / writeup view 的业务流程。
- 不处理 `components/platform/contest/*`。

## 输入依据

- `code/frontend/src/components/platform/challenge/*`
- `code/frontend/src/pages/platform/challenges/ChallengeImportManageRoutePage.vue`
- `code/frontend/src/pages/platform/challenges/ChallengeImportPreviewRoutePage.vue`
- `code/frontend/src/pages/platform/challenges/ChallengePackageFormatRoutePage.vue`
- `code/frontend/src/features/platform/challenge-package-import/**`
- `code/frontend/src/features/platform/challenge-detail/ui/AdminChallengeProfilePanel.vue`
- `code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupViewPage.vue`
- `code/frontend/src/features/platform/awd-challenges/ui/AwdChallengeImportSection.vue`

## 当前结论

- `ChallengeImportHeroPanel`、`ChallengeImportPreviewWorkspacePanel`、`ChallengeImportQueuePanel`、`ChallengeImportUploadResultsPanel`、`ChallengePackageFormatGuidePanel`、`ChallengePackageImportEntry`、`ChallengePackageImportReview` 都是 challenge import capability 的私有 UI。
- `ChallengeDescriptionPanel` 由 import preview、platform challenge detail、writeup view 三处共享，更适合落在 `entities/challenge/ui`。
- 最小可审阅切片是：先补 owner，再切 route page / feature consumer / raw-source 测试，最后删旧文件。

## 设计边界

### `features/platform/challenge-package-import/ui` 本轮负责

- import manage hero
- import preview workspace
- import queue panel
- import upload results panel
- package format guide
- package import entry
- package import review

### `entities/challenge/ui` 本轮负责

- markdown challenge description 展示与 sanitize owner

### `features/platform/challenge-package-import/model` 本轮继续负责

- import upload / preview / format route owner
- redirect / route target builder

## 任务切片

- [ ] Slice 1：建立新 owner
  - 目标：
    - 新建 `features/platform/challenge-package-import/ui/*`
    - 新建 `features/platform/challenge-package-import/ui/index.ts`
    - 新建 `entities/challenge/ui/ChallengeDescriptionPanel.vue`
    - 更新 `features/platform/challenge-package-import/index.ts`、`entities/challenge/index.ts`
  - 验证：
    - `rg -n "@/components/platform/challenge" code/frontend/src/features/platform/challenge-package-import code/frontend/src/entities/challenge`

- [ ] Slice 2：切 consumer 与测试
  - 目标：
    - route page、feature consumer、raw-source 测试改到新 owner
    - `components.d.ts` 同步
  - 验证：
    - `pnpm vitest run src/pages/platform/challenges/__tests__/ChallengeImportManage.test.ts src/pages/platform/challenges/__tests__/ChallengePackageFormat.test.ts src/pages/__tests__/workspacePageHeaderStyles.test.ts src/pages/__tests__/workspaceShellStyles.test.ts`

- [ ] Slice 3：删除旧组件并更新 backlog
  - 目标：
    - 删除旧 `components/platform/challenge/*`
    - backlog 同步 `components/platform/challenge` 当前事实
  - 验证：
    - `bash scripts/check-frontend-architecture.sh --quick`

## 验证计划

- `python3 harness/checks/check-reuse-decision.py`
- `bash scripts/check-task-intake.sh --reuse-decision platform-challenge-feature-owner-migration-batch1`
- `cd code/frontend && pnpm vitest run src/pages/platform/challenges/__tests__/ChallengeImportManage.test.ts src/pages/platform/challenges/__tests__/ChallengePackageFormat.test.ts src/pages/__tests__/workspacePageHeaderStyles.test.ts src/pages/__tests__/workspaceShellStyles.test.ts src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `bash scripts/check-frontend-architecture.sh --quick`

## 残余风险

- 这一轮之后，`components/platform/challenge/*` 若还有残留，就应重新判断是不是 shared primitive，而不是再默认留在历史目录。
- `ChallengeDescriptionPanel` 下沉为实体展示块后，后续若学生侧题面也要复用，应继续沿实体 owner 方向，而不是回流到 `components/platform/*`。
