> 状态：Current
> 事实源：platform AWD challenge feature-owned UI normalization
> 替代：无

# Platform AWD Challenge Feature UI Normalization Plan

## 目标

- 把 `AWDChallengeEditorDialog.vue` 与 `AwdChallengeImportSection.vue` 从 `components/platform/awd-service` 迁入 `features/platform-awd-challenges/ui`。
- 让 `AWDChallengeLibrary.vue` 和 `AWDChallengeLibraryPage.vue` 都只通过 `features/platform-awd-challenges` 读取这组 UI。

## 非目标

- 本轮不改 `useAwdChallengeLibraryPage.ts`、`useAwdChallengeImportFlow.ts` 的请求和状态 owner。
- 本轮不重写 AWD 题库页视觉结构，不调整导入按钮、确认导入或编辑保存交互。
- 本轮不顺手迁 `AwdChallengeLibrarySection.vue`、`AwdChallengeWorkspaceHeader.vue`，因为它们当前不在 allowlist 收口目标内。

## 输入依据

- `code/frontend/src/features/platform-awd-challenges/ui/AWDChallengeLibraryPage.vue`
- `code/frontend/src/views/platform/AWDChallengeLibrary.vue`
- `code/frontend/src/components/platform/awd-service/AWDChallengeEditorDialog.vue`
- `code/frontend/src/components/platform/awd-service/AwdChallengeImportSection.vue`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `docs/architecture/frontend/06-components.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `platform-awd-challenges` 已经拥有自己的 feature page shell，但导入区块和编辑对话框仍沿用旧组件目录。
- 这两个 UI 只服务 `platform-awd-challenges`，并直接使用该 feature contract，因此符合 `feature-owned UI` 判定规则。
- 最小收口面是迁 UI owner 和 import 路径，不改 model owner。

## 设计边界

### `features/platform-awd-challenges/ui/*` 本轮负责

- AWD 题库页 page shell
- AWD 题目导入区块
- AWD 题目编辑对话框

### `features/platform-awd-challenges/model/*` 本轮继续负责

- 列表加载、分页、筛选、删除
- 导入上传、导入确认
- 编辑草稿、保存、对话框开关状态

### route view 本轮继续负责

- 组合 feature public API
- 保持薄壳，不重新接管编辑或导入逻辑

## 任务切片

### Slice 1：feature UI 落位

- 目标：
  - 新增 `features/platform-awd-challenges/ui/AWDChallengeEditorDialog.vue`
  - 新增 `features/platform-awd-challenges/ui/AwdChallengeImportSection.vue`
  - `AWDChallengeLibraryPage.vue` 与 `AWDChallengeLibrary.vue` 改从 feature UI / public API 引用
- 验证：
  - `npm run test:run -- src/components/platform/awd-service/__tests__/AWDChallengeEditorDialog.test.ts src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts`
- Review focus：
  - 对话框本地 draft、校验、保存 guard 是否保持原 owner
  - 导入区块是否只迁路径，不改变确认导入交互

### Slice 2：allowlist 与 raw-source 护栏同步

- 目标：
  - 收掉 `architectureAllowlist.ts` 里 `platform-awd-challenges` 的两条 component->feature 例外
  - 切换相关 raw-source 测试与 `components.d.ts` 到新路径
- 验证：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/studentDirectoryTypographyBoundary.test.ts`
- Review focus：
  - route view 是否继续只组合 feature public API
  - raw-source 护栏是否仍覆盖到新的 feature UI owner

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/platform/awd-service/__tests__/AWDChallengeEditorDialog.test.ts src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/studentDirectoryTypographyBoundary.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`

## 残余风险

- `AwdChallengeLibrarySection.vue` 与 `AwdChallengeWorkspaceHeader.vue` 仍暂留在 `components/platform/awd-service`。它们当前没有直接触发 `componentFeatureImportAllowlist`，所以本轮不扩大到整组 UI 全搬迁。
