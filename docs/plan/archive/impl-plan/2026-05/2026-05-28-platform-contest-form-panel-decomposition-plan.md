> 状态：Current
> 事实源：`PlatformContestFormPanel.vue` 当前 form owner、`features/platform-contests/ui` 既有 workspace panel 组合模式、前端技术债 backlog
> 替代：无

# Platform Contest Form Panel Decomposition Plan

## 目标

- 把 `PlatformContestFormPanel.vue` 从“表单 owner + 三块模板 + 全量样式”收口成明确的 form owner
- 在 `features/platform-contests/ui` 内补齐 contest form section cluster
- 保持对外 props / emits contract 不变，让 dialog / workspace / orchestration 三个消费面无需改行为

## 非目标

- 本轮不改 `ContestFormDraft`、`ContestFieldLocks`、`statusOptions` 的数据结构
- 本轮不把校验、draft 同步或 save workflow 上提到 `useContestManagePage()`、`useContestEditPage()` 等 page model
- 本轮不顺手重排 `ContestOrchestrationPage.vue`、`ContestEditWorkspacePanel.vue` 或 `PlatformContestFormDialog.vue`

## 输入依据

- `code/frontend/src/features/platform-contests/ui/PlatformContestFormPanel.vue`
- `code/frontend/src/features/platform-contests/ui/PlatformContestFormDialog.vue`
- `code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue`
- `code/frontend/src/features/platform-contests/ui/ContestOrchestrationPage.vue`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts`
- `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `PlatformContestFormPanel.vue` 的真正行为 owner 只有三类：`props.draft -> localDraft` 同步、字段校验、`save/update:draft/cancel` 发射
- 其余大部分代码都属于稳定的展示分区和局部布局样式，适合按 section cluster 下沉
- 当前三个消费方都只依赖 `PlatformContestFormPanel` 的现有 contract，本轮应保持这层 API 稳定

## 设计边界

### `PlatformContestFormPanel.vue` 本轮继续负责

- `localDraft` 本地草稿
- `fieldErrors` 校验状态
- props 同步和 `update:draft` 发射
- `validate()` / `handleSubmit()`

### 新增 `PlatformContest*Section` 子组件本轮负责

- 基础信息区字段展示
- 赛制与状态区展示
- 时间轴区展示
- 底部动作区展示

### `PlatformContestFormSectionShell.vue` 本轮负责

- section icon / 标题 / 描述 / 内容壳体
- 共享 section 布局样式

### 本轮不动

- contest create / edit workflow owner
- dialog 打开关闭策略
- 竞赛表单保存 API

## 任务切片

### Slice 1：提取 section shell 与三块字段分区

- 目标：
  - 新增 `PlatformContestFormSectionShell.vue`
  - 新增 `PlatformContestIdentitySection.vue`
  - 新增 `PlatformContestRulesSection.vue`
  - 新增 `PlatformContestTimelineSection.vue`
  - 父面板改为只组合这些 section，并把输入状态通过 props 传下去
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts`
- Review focus：
  - 父面板是否仍然是唯一 draft / validate / submit owner
  - 子组件是否只消费 props，不偷偷改写外层业务状态

### Slice 2：提取动作区并收口样式 owner

- 目标：
  - 新增 `PlatformContestFormActions.vue`
  - 把 section / action 的样式从父面板大块 scoped CSS 中下沉到更贴近 owner 的组件
  - 保持 shared ui primitive、theme token 和响应式规则不回退
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts`
- Review focus：
  - 新增样式是否仍使用现有 token，不回退到硬编码值
  - 父面板文件体量是否显著下降，且没有留下半套旧样式残片

### Slice 3：backlog 与 review 收口

- 目标：
  - 更新 backlog 对 `PlatformContestFormPanel.vue` 的状态
  - 补 frontend review 记录
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- Review focus：
  - touched surface 上的超大组件债是否真的收口
  - 没有把新的 feature 内部例外转移到其它文件

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 如果 `ContestChallengeOrchestrationPanel.vue` 后续仍继续膨胀，需要单独按 orchestration 内部分区再开一刀；这不属于本轮 `PlatformContestFormPanel` 的 owner 收口范围
- 样式拆分后若仍有明显重复，后续可再评估是否抽 feature-local shared stylesheet，但本轮先以 owner 清晰为主
