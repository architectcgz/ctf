> 状态：Current
> 事实源：`AWDChallengeLibraryPage.vue` 当前 owner、`AWDChallengeLibrary.vue` / `AWDChallengeImport.vue` route view、`useAwdChallengeLibraryPage.ts` / `useAwdChallengeImportPage.ts` feature composable、现有 workspace page 分区拆分模式
> 替代：无

# AWD Challenge Library Page Decomposition Implementation Plan

## 目标

- 拆分 `code/frontend/src/components/platform/awd-service/AWDChallengeLibraryPage.vue`，把页头、题目目录工作区、导入工作区从单一 legacy component page 中拆成稳定展示分区。
- 保持 `code/frontend/src/views/platform/AWDChallengeLibrary.vue` 与 `code/frontend/src/views/platform/AWDChallengeImport.vue` 继续作为 route 组合面，保持 `features/platform-awd-challenges` 继续拥有路由、加载、导入、编辑与删除 workflow owner。
- 同步收口与该页面相关的源码护栏和页面测试，让后续新增 AWD 题目能力不再继续堆回单个 `components/*Page.vue`。

## 非目标

- 本轮不改变 `usePlatformAwdChallenges.ts` 的 API owner、分页策略、导入流程或对话框 workflow。
- 本轮不把 AWD 题目页整体迁移到新的 widget / entities 层级，也不处理 `AWDChallengeEditorDialog.vue` 的后续拆分。
- 本轮不变更 `AWDChallengeLibrary` 与 `AWDChallengeImport` 的用户可见交互、文案语义和测试断言主流程。

## 输入依据

- `code/frontend/src/components/platform/awd-service/AWDChallengeLibraryPage.vue`
- `code/frontend/src/views/platform/AWDChallengeLibrary.vue`
- `code/frontend/src/views/platform/AWDChallengeImport.vue`
- `code/frontend/src/features/platform-awd-challenges/model/useAwdChallengeLibraryPage.ts`
- `code/frontend/src/features/platform-awd-challenges/model/useAwdChallengeImportPage.ts`
- `code/frontend/src/features/platform-awd-challenges/model/usePlatformAwdChallenges.ts`
- `code/frontend/src/components/platform/dashboard/PlatformOverviewPage.vue`
- `code/frontend/src/components/platform/user/UserGovernancePage.vue`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/platform/__tests__/AWDChallengeLibrary.test.ts`
- `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `AWDChallengeLibraryPage.vue` 当前约 `896` 行，同时承担 library 与 import 两种 mode 的页头、指标卡、筛选栏、目录表格、分页、导入入口、上传结果、待确认队列和样式，已经明显超过“页面展示壳”应有职责密度。
- `AWDChallengeLibrary.vue` 和 `AWDChallengeImport.vue` 已经分别通过 `useAwdChallengeLibraryPage()` / `useAwdChallengeImportPage()` 持有 route / feature owner，所以最小正确收口点不是再移动 composable owner，而是把 page 组件内部按稳定展示分区拆开。
- 两个 route view 共用同一个 `AWDChallengeLibraryPage.vue`，说明拆分时需要保留单一 page surface 和 `mode` 驱动，而不是把 library 与 import 流程重新分叉成两套不一致壳体。

## 设计边界

### 父页继续负责

- `AWDChallengeLibraryPage.vue` 继续作为共享 page surface，保留 `mode` 选择、顶层 props / emits contract、展示 helper 的装配入口，以及 library / import 分区的切换。
- `AWDChallengeLibrary.vue` 继续拥有 library route 的 composable 选择与对话框装配。
- `AWDChallengeImport.vue` 继续拥有 import route 的 composable 选择。

### 子组件负责

- 共享页头：只负责展示标题、说明、模式说明文案和头部操作按钮，不直接依赖 router 或 feature composable。
- library 分区：负责指标卡、筛选工具栏、目录表格、空态、加载态和分页展示；通过 props / emits 接收筛选值和用户操作。
- import 分区：负责导入入口、上传结果和待确认队列展示；通过 props / emits 接收上传态、队列数据和确认导入动作。

### 本轮明确不负责

- 子组件不直接调用 `useAwdChallengeLibraryPage()`、`useAwdChallengeImportPage()` 或 `usePlatformAwdChallenges()`。
- 子组件不直接拥有 `useRouter`、`useRoute`、`onMounted`、API 请求或删除确认逻辑。
- 本轮不新增新的 feature 层状态，不把 `AWDChallengeEditorDialog` 吞并回 page 组件。

## 任务切片

### Slice 1：提取共享页头

- 目标：
  - 从 `AWDChallengeLibraryPage.vue` 提取 library / import 共用的 workspace header。
  - 让 `mode` 对应的标题、摘要、提示文案和头部按钮在单独组件里声明，父页只装配 props / emits。
- 预期改动：
  - `code/frontend/src/components/platform/awd-service/AWDChallengeLibraryPage.vue`
  - `code/frontend/src/components/platform/awd-service/AwdChallengeWorkspaceHeader.vue`
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/AWDChallengeLibrary.test.ts`
- Review focus：
  - header 子组件是否仍然只吃 props / emits
  - 头部按钮动作是否仍由 route view / feature owner 间接驱动

### Slice 2：提取 library 工作区分区

- 目标：
  - 把指标卡、筛选栏、目录表格、空/加载态、分页从父页拆到独立 library section。
  - 保持筛选值、分页值和目录行操作仍由父页透传。
- 预期改动：
  - `code/frontend/src/components/platform/awd-service/AWDChallengeLibraryPage.vue`
  - `code/frontend/src/components/platform/awd-service/AwdChallengeLibrarySection.vue`
  - `code/frontend/src/views/platform/__tests__/AWDChallengeLibrary.test.ts`
  - `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/AWDChallengeLibrary.test.ts src/views/__tests__/workspaceShellStyles.test.ts`
- Review focus：
  - library section 是否没有接管 feature / router owner
  - 表格操作链路 `edit / delete / page / filter` 是否仍与原来一致
  - `workspace-page-header` / `workspace-directory-section` 共享结构是否保持不变

### Slice 3：提取 import 工作区分区

- 目标：
  - 把导入入口、上传结果和待确认队列从父页拆到独立 import section。
  - 保持导入文件选择、队列刷新和确认导入动作仍由父页透传。
- 预期改动：
  - `code/frontend/src/components/platform/awd-service/AWDChallengeLibraryPage.vue`
  - `code/frontend/src/components/platform/awd-service/AwdChallengeImportSection.vue`
  - `code/frontend/src/views/platform/__tests__/AWDChallengeLibrary.test.ts`
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/AWDChallengeLibrary.test.ts`
- Review focus：
  - import section 是否仍是纯展示 owner
  - `ChallengePackageImportEntry`、上传结果和待确认队列行为是否未回归

## 结构收口检查

- `AWDChallengeLibrary.vue` 与 `AWDChallengeImport.vue` 不新增 `useRouter` / API 直连。
- `AWDChallengeLibraryPage.vue` 的职责应收敛为 page surface，而不是继续维护完整 library/import 展示细节。
- 新增子组件不应出现在 `componentFeatureImportAllowlist` 里，不新增新的 legacy `*Page.vue`。
- 本轮即使 `AWDChallengeLibraryPage.vue` 仍保留在 `components/`，也要先把它从“超大混合页壳”收口到“清晰 page surface + 子分区”的状态，为后续继续迁移到更中立的 page/widget owner 创造条件。

## 验证计划

- `npm run test:run -- src/views/platform/__tests__/AWDChallengeLibrary.test.ts src/views/__tests__/workspaceShellStyles.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`
- 必要时：`git diff --check -- code/frontend/src/components/platform/awd-service code/frontend/src/views/platform/AWDChallengeLibrary.vue code/frontend/src/views/platform/AWDChallengeImport.vue`

## Review 关注点

- `mode="library" / "import"` 两条路径的展示行为是否保持一致。
- 子组件是否只是视觉与展示结构 owner，没有把 route / async / destructive action owner 从 feature 层抢走。
- `workspaceShellStyles` 相关断言是否仍能证明 AWD 页面沿用共享 workspace 壳，而不是回流局部骨架样式。
- 本次切片是否真实收口了 oversized page debt，而不是只把长模板平移到另一个不清晰名字的文件里。

## 回退 / 恢复说明

- 所有提取出的 section / header 组件都应能按文件粒度独立回退，父页只需回退 imports 和模板接线。
- 本轮不涉及 API、路由、状态持久化或配置迁移，回退主要是前端组件结构回退。

## 残余风险

- 当前 view 测试对源码结构有 `?raw` 断言，拆分后如果不同时调整测试，很容易出现“结构改善但 guardrail 仍按旧文件断言”的假失败。
- library 与 import 共用一个 page surface，如果 props contract 切得不干净，容易让 import-only / library-only 的字段在子组件间互相泄漏。
- `AWDChallengeEditorDialog.vue` 仍然是同一 feature 家族里的另一块大组件，本轮不动它，但要避免为了减少父页代码量把新的复杂逻辑又堆到 dialog 里。
