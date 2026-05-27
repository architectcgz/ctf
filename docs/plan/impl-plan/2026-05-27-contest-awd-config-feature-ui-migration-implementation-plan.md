> 状态：Current
> 事实源：`ContestAwdConfigWorkspaceShell.vue` 到 `features/contest-awd-config/ui` 的迁移边界
> 替代：无

# Contest AWD Config Feature UI Migration Implementation Plan

## 目标

- 把 `ContestAwdConfigWorkspaceShell.vue` 从 `components/platform/contest/` 迁到 `features/contest-awd-config/ui/`。
- 让 `views/platform/ContestAwdConfig.vue` 通过 `features/contest-awd-config` public API 组合 page-sized workspace shell 与 `useContestAwdConfigPage()`。

## 非目标

- 本轮不改 `useContestAwdConfigPage.ts` 的 router、load、preview、save、draft owner。
- 本轮不继续拆 `ContestAwdCheckerConfigSection.vue` 或其它 AWD 配置子组件。
- 本轮不改用户可见的 AWD 配置交互、字段、按钮文案和错误提示。

## 输入依据

- `docs/architecture/frontend/06-components.md`
- `docs/architecture/frontend/07-pages-dataflow.md`
- `code/frontend/src/views/platform/ContestAwdConfig.vue`
- `code/frontend/src/components/platform/contest/ContestAwdConfigWorkspaceShell.vue`
- `code/frontend/src/features/contest-awd-config/index.ts`
- `code/frontend/src/features/contest-awd-config/model/useContestAwdConfigPage.ts`
- `code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `ContestAwdConfig.vue` 现在只承担 route composition，真正的 AWD 配置 owner 已经在 `useContestAwdConfigPage.ts`。
- `ContestAwdConfigWorkspaceShell.vue` 是标准 page-sized workspace shell，适合直接迁到 `features/contest-awd-config/ui/`。
- 这轮不需要额外新增 route-aware composable。

## 设计边界

### route view 继续负责

- 组合 `useContestAwdConfigPage()` 与 `ContestAwdConfigWorkspaceShell`
- 不直接持有 router / API / preview / save owner

### `features/contest-awd-config/model` 继续负责

- 路由参数与返回编辑页跳转
- AWD 服务加载、选择、draft、preview、save
- breadcrumb owner

### `features/contest-awd-config/ui` 本轮负责

- AWD 配置页 workspace shell
- 消费上层 props 与事件 handler
- 不直接持有 transport 或 router owner

## 任务切片

### Slice 1：迁移 AWD 配置 workspace shell 到 feature ui

- 目标：
  - 新增 `features/contest-awd-config/ui/ContestAwdConfigWorkspaceShell.vue`
  - `features/contest-awd-config/index.ts` 导出 `ui`
  - `views/platform/ContestAwdConfig.vue` 改从 feature public API 引用
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts`
- Review focus：
  - route view 是否继续只保留组合壳
  - page shell 是否没有重新吸回 model owner

### Slice 2：同步类型、测试与 backlog

- 目标：
  - 更新 `components.d.ts` 和 raw-source 测试路径
  - 记录 backlog 进展
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts`
- Review focus：
  - source-boundary 断言是否已切到 feature public API

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮与上一批未提交迁移共存，提交时仍要按任务边界筛文件。
