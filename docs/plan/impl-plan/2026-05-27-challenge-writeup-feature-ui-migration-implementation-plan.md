> 状态：Current
> 事实源：题解管理三件套当前 owner、`features/platform-challenge-detail/ui` 既有模式、前端架构 allowlist 与 route page 组合边界
> 替代：无

# Challenge Writeup Feature UI Migration Implementation Plan

## 目标

- 把 `ChallengeWriteupManagePanel.vue`、`ChallengeWriteupEditorPage.vue`、`ChallengeWriteupViewPage.vue` 从 `components/platform/writeup/` 迁到 `features/challenge-writeup-editor/ui/`。
- 在前端架构事实源中补齐 `feature-owned UI` 规则，明确哪些 UI 应留在 `components/**`，哪些应跟随 feature 落到 `features/*/ui`。
- 收掉题解这组三件套对应的 `componentFeatureImportAllowlist` 和 `legacyComponentPageAllowlist` 例外。

## 非目标

- 本轮不改 `useChallengeWriteupEditorPage()`、`useChallengeWriteupManagement()`、`useChallengeWriteupPage()`、`useChallengeWriteupViewPage()` 的行为 owner。
- 本轮不改题解 API、题目详情跳转语义、用户可见文案和表单交互。
- 本轮不顺手处理其它 `components/*Page.vue` 候选，避免把架构收口扩成全局搬家。

## 输入依据

- `docs/architecture/frontend/06-components.md`
- `docs/architecture/frontend/07-pages-dataflow.md`
- `code/frontend/src/components/platform/writeup/ChallengeWriteupManagePanel.vue`
- `code/frontend/src/components/platform/writeup/ChallengeWriteupEditorPage.vue`
- `code/frontend/src/components/platform/writeup/ChallengeWriteupViewPage.vue`
- `code/frontend/src/features/challenge-writeup-editor/model/*.ts`
- `code/frontend/src/features/platform-challenge-detail/ui/*`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- 题解三件套当前并不是复用型组件，而是单一 feature 的 UI 面：它们直接依赖 `features/challenge-writeup-editor`，并且只被平台题目管理相关 route / workspace 使用。
- 现在继续把它们留在 `components/platform/writeup/`，只会让 `components/* -> features/*` 例外和 `*Page.vue` 遗留页例外继续增长。
- 已有 `features/platform-challenge-detail/ui` 可以作为本轮的现成落点模式，因此这次不需要新发明 layer，只需要把规则写清楚并迁到一致位置。

## 设计边界

### `components/**` 继续负责

- 共享原语、布局壳、跨 feature 可复用的展示组件
- 不直接绑定单一 feature model 的中立业务组件

### `features/*/ui` 负责

- 只服务单一 feature 的 workspace / editor / panel / page-sized surface
- 直接消费同一 feature 的 model contract、emit 或只读展示数据
- 不直接调用非 contract API，不持有跨 feature / route 的 owner

### route view 继续负责

- 作为路由入口组合 feature page model 与 feature ui
- 不直接承担题解加载、保存、删除和推荐动作

## 任务切片

### Slice 1：补齐前端架构规则

- 目标：
  - 在 `06-components.md`、`07-pages-dataflow.md` 明确 `feature-owned UI` 的归属和使用条件。
- 预期改动：
  - `docs/architecture/frontend/06-components.md`
  - `docs/architecture/frontend/07-pages-dataflow.md`
- 验证：
  - 人工复核文档是否明确写出负责 / 不负责 和代表性例子
- Review focus：
  - 是否避免和现有 `components/common`、route view、feature model 规则冲突

### Slice 2：迁移题解 feature UI

- 目标：
  - 新增 `features/challenge-writeup-editor/ui/`，迁入三件套并通过 feature public API 暴露。
  - route view 与题目详情工作区改从 feature 入口引用。
- 预期改动：
  - `code/frontend/src/features/challenge-writeup-editor/index.ts`
  - `code/frontend/src/features/challenge-writeup-editor/ui/*`
  - `code/frontend/src/views/platform/ChallengeWriteup.vue`
  - `code/frontend/src/views/platform/ChallengeWriteupView.vue`
  - `code/frontend/src/components/platform/challenge/AdminChallengeWorkspaceTabs.vue`
  - `code/frontend/src/components.d.ts`
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/ChallengeWriteup.test.ts src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts`
- Review focus：
  - route view 是否仍然保持薄壳
  - feature ui 是否没有新吸入 API / router owner

### Slice 3：清理 guardrail 与 backlog 记录

- 目标：
  - 移除题解三件套对应的 allowlist 例外，更新 raw-source 测试路径和技术债记录。
- 预期改动：
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - 相关 raw-source 测试
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-challenge-writeup-feature-ui-migration-review.md`
- 验证：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts src/views/platform/__tests__/ChallengeWriteup.test.ts src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- Review focus：
  - allowlist 是否真实减少
  - 测试是否已经按新边界断言，而不是继续绑定旧组件路径

## 结构收口检查

- 题解 editor / view / manage UI 不再作为 `components/*Page.vue` 遗留页存在。
- route view 继续只依赖 feature public API，不重新回到 `components/platform/writeup/*`。
- 当前 touched surface 上的已知结构债必须真实下降：至少移除三条 component->feature allowlist 和两条 legacy component page allowlist。

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts src/views/platform/__tests__/ChallengeWriteup.test.ts src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- docs/architecture/frontend/06-components.md docs/architecture/frontend/07-pages-dataflow.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md .harness/reuse-decisions/challenge-writeup-feature-ui-migration.md docs/plan/impl-plan/2026-05-27-challenge-writeup-feature-ui-migration-implementation-plan.md code/frontend/src/features/challenge-writeup-editor code/frontend/src/views/platform/ChallengeWriteup.vue code/frontend/src/views/platform/ChallengeWriteupView.vue code/frontend/src/components/platform/challenge/AdminChallengeWorkspaceTabs.vue code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/components.d.ts code/frontend/src/views/platform/__tests__/ChallengeWriteup.test.ts code/frontend/src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review 关注点

- `features/challenge-writeup-editor/ui` 是否成为题解 feature 唯一 UI owner，而不是只把旧路径平移后继续保留双入口。
- route page 与 feature ui 的 import 方向是否符合当前架构守卫。
- 文档规则是否足够具体，能指导后续其它 “该进 feature ui 还是继续留在 components” 的判断。

## 回退 / 恢复说明

- 如迁移后出现问题，可按文件粒度把三个 UI 文件移回 `components/platform/writeup/` 并恢复原 import。
- 本轮不触碰 API / DTO / 路由命名，回退主要是前端目录和 import 关系回退。

## 残余风险

- 题解 feature 自身仍只拆了 `model + ui` 两层，本轮不引入更细 widget / entity 分层。
- 其它 legacy `components/*Page.vue` 仍然存在，本轮只拿题解这一组做边界收口样板。
