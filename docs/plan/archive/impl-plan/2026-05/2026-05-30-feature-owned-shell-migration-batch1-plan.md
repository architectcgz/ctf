# feature-owned shell 第一批迁移计划

> 状态：In Progress
> 事实输入：`code/frontend/src/pages/auth/*.vue`、`code/frontend/src/pages/profile/*.vue`、`code/frontend/src/pages/instances/*.vue`、`code/frontend/src/pages/notifications/*.vue`、`code/frontend/src/pages/scoreboard/*.vue`、`code/frontend/src/components/auth/AuthEntryShell.vue`、`code/frontend/src/components/profile/*.vue`、`code/frontend/src/components/instance/InstanceListWorkspaceShell.vue`、`code/frontend/src/components/notifications/NotificationCategoryFilter.vue`、`code/frontend/src/components/scoreboard/ScoreboardWorkspaceShell.vue`

## Plan Summary

- Objective
  - 把第一批已经明确只服务单一 capability 的历史 workspace shell / page shell 从 `components/*` 收回各自 `features/*/ui` owner。
- Non-goals
  - 不在这轮迁移 `components/contests/*`
  - 不在这轮迁移 `components/challenge/*`
  - 不调整 route page 仍应保留的页面层职责
- Source architecture or design docs
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/architecture/frontend/06-components.md`
  - `docs/architecture/frontend/07-pages-dataflow.md`
- Dependency order
  - 先确认 owner 与引用，再迁移组件物理落点，再切 route page / tests 引用，最后更新 backlog 与验证。
- Expected specialist skills
  - `frontend-engineer`
  - `frontend-sliced-architecture`
  - `development-pipeline`

## Task 1

- Goal
  - 确认这批组件都是单 capability owner，而不是跨 feature 共享块。
- Touched modules or boundaries
  - `pages/auth`
  - `pages/profile`
  - `pages/instances`
  - `pages/notifications`
  - `pages/scoreboard`
  - `components/auth|profile|instance|notifications|scoreboard`
- Dependencies
  - 无
- Validation
  - `rg` 检查引用路径与 consumer 数量
- Review focus
  - `SkillProfile` 的 owner 是否应落在 `features/skill-profile` 而不是 `features/profile`
- Risk notes
  - 若误把跨 capability 组件塞回单一 feature，会制造新的架构漂移

## Task 2

- Goal
  - 在 owning feature 下建立 `ui/` public API，并迁移对应 shell 文件。
- Touched modules or boundaries
  - `features/auth/ui`
  - `features/profile/ui`
  - `features/skill-profile/ui`
  - `features/instance-list/ui`
  - `features/notifications/ui`
  - `features/scoreboard/ui`
- Dependencies
  - Task 1
- Validation
  - `rg` 确认旧 `components/*` 路径不再被运行时代码引用
- Review focus
  - public API 是否仍清晰，只导出当前 capability 需要暴露的 UI
- Risk notes
  - 若只移动文件不更新 public API，route page 可能继续走旧深路径

## Task 3

- Goal
  - 切换 route page、邻近测试和稳定策略测试到新 feature owner。
- Touched modules or boundaries
  - `pages/auth`
  - `pages/profile`
  - `pages/instances`
  - `pages/notifications`
  - `pages/scoreboard`
  - `pages/**/__tests__`
- Dependencies
  - Task 2
- Validation
  - 运行相关 page / shell 测试
- Review focus
  - 只改 import owner，不改变页面行为与断言语义
- Risk notes
  - raw source 断言测试若漏改，会直接造成 Vitest 失败

## Task 4

- Goal
  - 更新 backlog 当前事实，并确认本轮收口后还剩哪些 `components/*` 迁移面。
- Touched modules or boundaries
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Dependencies
  - Task 3
- Validation
  - backlog 与代码目录现状一致
- Review focus
  - 不把已完成项继续保留成活动迁移项
- Risk notes
  - 若事实源不更新，后续 agent 会继续把已完成项当成待办

## Integration Checks

- route page 仍只负责页面组合和事件桥接
- `components/auth|profile|instance|notifications|scoreboard` 不再承载单 capability shell owner
- owning feature 的 `index.ts` 能提供稳定 public API

## Rollback / Recovery Notes

- 若某个 feature public API 设计不合适，可单独回滚该 capability 的 import 切换
- 物理删除旧组件前先确保新路径已被 route page 和测试完全接管

## Residual Risks

- `components/contests/*` 与 `components/challenge/*` 仍是下一批高收益迁移面，不在本轮收口
