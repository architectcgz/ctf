> 状态：Current
> 事实源：`ContestAwdConfigWorkspaceShell.vue` 当前 owner、`useContestAwdConfigPage.ts` 的 page-level workflow owner、现有 contest / AWD 大组件拆分模式
> 替代：无

# Contest AWD Config Workspace Shell Decomposition Implementation Plan

## 目标

- 拆分 `code/frontend/src/components/platform/contest/ContestAwdConfigWorkspaceShell.vue` 中剩余的 `Checker Parameters` 超大画布，把不同 checker type 的稳定配置分区从父壳抽成独立组件。
- 保持 `code/frontend/src/views/platform/ContestAwdConfig.vue` 与 `features/contest-awd-config` 继续拥有路由、数据加载、预览、保存、服务选择和字段校验 owner。
- 同步更新 AWD 配置页相关的 raw-source 护栏和 backlog 进展，让后续 contest / AWD 配置需求不再继续堆回同一个 1000 行 workspace 壳。

## 非目标

- 本轮不改变 `useContestAwdConfigPage.ts`、`useAwdCheckerConfigDraft.ts`、`useAwdCheckerPreview.ts`、`useAwdCheckerSaveFlow.ts` 的业务 owner 或 API 调用策略。
- 本轮不迁移 `ContestAwdConfigWorkspaceShell.vue` 的目录位置，不把它改造成新的 feature-owned UI。
- 本轮不改变用户可见的 AWD checker 配置行为、字段命名、保存 payload 和预览流程。

## 输入依据

- `code/frontend/src/components/platform/contest/ContestAwdConfigWorkspaceShell.vue`
- `code/frontend/src/views/platform/ContestAwdConfig.vue`
- `code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts`
- `code/frontend/src/features/contest-awd-config/model/useContestAwdConfigPage.ts`
- `code/frontend/src/components/platform/contest/ContestAwdConfigTopbar.vue`
- `code/frontend/src/components/platform/contest/ContestAwdServiceDirectory.vue`
- `code/frontend/src/components/platform/contest/ContestAwdEditorHeader.vue`
- `code/frontend/src/components/platform/contest/ContestAwdScoreWeights.vue`
- `code/frontend/src/components/platform/contest/ContestAwdDebugStation.vue`
- `code/frontend/src/components/platform/contest/ContestAwdConfigFooter.vue`
- `code/frontend/src/components/platform/contest/ContestChallengeEditorDialog.vue`
- `code/frontend/src/components/platform/awd-service/AWDChallengeLibraryPage.vue`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `ContestAwdConfigWorkspaceShell.vue` 当前约 `1009` 行，虽然已经拆出 topbar / service directory / header / score weights / debug station / footer，但剩余 `Checker Parameters` 画布仍同时承担四种 checker type 的完整模板、局部交互和大量专用样式。
- `ContestAwdConfig.vue` 已经通过 `useContestAwdConfigPage()` 持有 route、加载、保存、预览和服务选择 owner，因此这轮不应该把 workflow owner 再迁出，而应继续按稳定展示分区拆 page shell。
- `Checker Parameters` 是天然的局部边界：父壳只需把 `selectedCheckerType`、draft、fieldErrors 和局部 handler 透传，类型分支内部则适合拆成稳定子分区。

## 设计边界

### 父壳继续负责

- `ContestAwdConfigWorkspaceShell.vue` 继续作为 AWD 配置页共享 workspace surface。
- 服务选择、`selectedCheckerType` 判定、`form` / `previewForm` draft、`fieldErrors`、保存、预览、刷新和返回动作 owner 保持在父壳 / page model。
- `ContestAwdDebugStation.vue` 与 `ContestAwdConfigFooter.vue` 的接线不变。

### 新增 checker 画布组件负责

- `ContestAwdCheckerConfigSection.vue`
  - 负责 `Checker Parameters` 标题区、checker 类型异常提醒和具体 checker type 子分区装配。
  - 只接收 props 与事件 handler，不直接持有 route、API、保存或预览逻辑。
- 类型分区子组件
  - `ContestAwdLegacyProbeFields.vue`
  - `ContestAwdHttpStandardFields.vue`
  - `ContestAwdTcpStandardFields.vue`
  - `ContestAwdScriptCheckerFields.vue`
  - 只承接对应 checker type 的字段模板与局部展示结构。

### 本轮明确不负责

- 新子组件不直接调用 `useContestAwdConfigPage()`、`useAwdCheckerConfigDraft()` 或 `useRouter()`。
- 新子组件不拥有 `fieldErrors` 的生成逻辑，不直接构造保存 payload 或预览请求。
- 本轮不顺手调整 `ContestAwdDebugStation.vue`、`ContestAwdConfigFooter.vue` 或其它已拆出的稳定区块。

## 任务切片

### Slice 1：抽出 checker 画布外壳与共享类型定义

- 目标：
  - 从 `ContestAwdConfigWorkspaceShell.vue` 提取 `Checker Parameters` 区块外壳。
  - 把当前仅在父壳内声明的 AWD checker draft 类型收口到共享本地类型文件，供父壳和新子组件复用。
- 预期改动：
  - `code/frontend/src/components/platform/contest/ContestAwdConfigWorkspaceShell.vue`
  - `code/frontend/src/components/platform/contest/ContestAwdCheckerConfigSection.vue`
  - `code/frontend/src/components/platform/contest/contestAwdConfigTypes.ts`
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts`
- Review focus：
  - 父壳是否仍保留所有 workflow owner
  - 新 section 是否只是 checker 画布 surface

### Slice 2：按 checker type 提取稳定字段分区

- 目标：
  - 把 `legacy_probe`、`http_standard`、`tcp_standard`、`script_checker` 四个模板分支拆成独立字段组件。
  - 保留父壳对 draft、字段错误和 step/preset 操作 handler 的唯一 owner。
- 预期改动：
  - `code/frontend/src/components/platform/contest/ContestAwdCheckerConfigSection.vue`
  - `code/frontend/src/components/platform/contest/ContestAwdLegacyProbeFields.vue`
  - `code/frontend/src/components/platform/contest/ContestAwdHttpStandardFields.vue`
  - `code/frontend/src/components/platform/contest/ContestAwdTcpStandardFields.vue`
  - `code/frontend/src/components/platform/contest/ContestAwdScriptCheckerFields.vue`
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts`
- Review focus：
  - 类型分区是否没有反向吸入保存 / 预览 / route owner
  - TCP step 展开/收起和 HTTP preset 行为是否保持原样

### Slice 3：更新护栏与 backlog

- 目标：
  - 把 AWD 配置页的 raw-source 护栏切到“父壳 + checker 画布 + 类型分区”的组合断言。
  - 更新 backlog，记录这条超大 contest / AWD 壳体拆分的进展。
- 预期改动：
  - `code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-contest-awd-config-workspace-shell-decomposition-review.md`
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts`
- Review focus：
  - 护栏是否已经绑定新的 owner，而不是继续把全部模板细节锁回父壳

## 结构收口检查

- `ContestAwdConfigWorkspaceShell.vue` 应收口成 workspace surface，不再继续混放四种 checker type 的完整模板实现。
- 新增的 checker 画布和字段分区组件不得接管 route、API、保存、预览或字段校验 owner。
- touched surface 上的 `TD-1` 要真实下降，不能只是把长模板平移到另一个同样含糊的大组件文件里。

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/contest-awd-config-workspace-shell-decomposition.md docs/plan/impl-plan/2026-05-27-contest-awd-config-workspace-shell-decomposition-implementation-plan.md docs/reviews/frontend/2026-05-27-contest-awd-config-workspace-shell-decomposition-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/components/platform/contest/ContestAwdConfigWorkspaceShell.vue code/frontend/src/components/platform/contest/ContestAwdCheckerConfigSection.vue code/frontend/src/components/platform/contest/ContestAwdLegacyProbeFields.vue code/frontend/src/components/platform/contest/ContestAwdHttpStandardFields.vue code/frontend/src/components/platform/contest/ContestAwdTcpStandardFields.vue code/frontend/src/components/platform/contest/ContestAwdScriptCheckerFields.vue code/frontend/src/components/platform/contest/contestAwdConfigTypes.ts code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review 关注点

- 父壳是否仍是 draft、fieldErrors、保存、预览和服务切换的唯一 owner。
- checker 画布拆分后，`legacy_probe` / `http_standard` / `tcp_standard` / `script_checker` 四条路径的用户可见行为是否一致。
- raw-source 护栏是否已经允许后续继续细分 checker 画布，而不是要求把长模板重新塞回父壳。

## 回退 / 恢复说明

- 新增的 checker 画布和字段组件都应能按文件粒度回退，父壳只需回退 imports、模板接线和局部类型引用。
- 本轮不涉及 API、路由或持久化变更，回退主要是前端组件结构回退。

## 残余风险

- `ContestAwdConfigWorkspaceShell.vue` 即使收口后仍会保留较多 page surface 接线，因为 AWD 配置页交互密度本身较高；本轮重点是去掉最肥的 checker 模板混写。
- 后续如继续新增 checker type，需要优先在独立字段组件扩展，而不是重新把模板堆回父壳。
