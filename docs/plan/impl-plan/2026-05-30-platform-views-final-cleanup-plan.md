# platform views 最终清场计划

> 状态：In Progress
> 事实输入：`code/frontend/src/router/routes/platformRoutes.ts`、`code/frontend/src/pages/platform/**`、`code/frontend/src/views/platform/**`、`docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Plan Summary

- Objective
  - 删除 `code/frontend/src/views/platform/*.vue` 的运行时迁移残片，让 `src/views/platform` 只保留 `__tests__/`。
- Non-goals
  - 不迁移 `views/platform/__tests__` 目录本身。
  - 不修改历史 review / plan 文档中的旧路径记录。
  - 不在这轮继续拆新的 feature owner。
- Source architecture or design docs
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/router/routes/platformRoutes.ts`
- Dependency order
  - 先改唯一活引用测试，再删除旧页面，再更新 backlog 事实，最后验证。
- Expected specialist skills
  - `frontend-engineer`
  - `development-pipeline`

## Task 1

- Goal
  - 确认 `views/platform/*.vue` 是否已经退出运行时与测试直接引用。
- Touched modules or boundaries
  - `router/routes/platformRoutes.ts`
  - `views/platform/*.vue`
  - `views/__tests__`
- Dependencies
  - 无
- Validation
  - `rg` 检查 `@/views/platform/`、`src/views/platform/`、raw import 旧页面命中。
- Review focus
  - 是否还有隐藏运行时引用或守卫脚本依赖旧路径。
- Risk notes
  - 若遗漏 raw source 测试引用，删除后会直接打断 Vitest。

## Task 2

- Goal
  - 把唯一仍直接依赖旧平台 view 的测试改到新的 route page。
- Touched modules or boundaries
  - `code/frontend/src/views/__tests__/adminRootHeroLayout.test.ts`
  - `code/frontend/src/pages/platform/challenges/ChallengeDetailRoutePage.vue`
- Dependencies
  - Task 1
- Validation
  - 相关测试通过；旧 `ChallengeDetail.vue` 不再被源码引用。
- Review focus
  - 断言语义是否保持一致，只改引用源不改行为预期。
- Risk notes
  - 如果 route page 和旧 view 结构不完全一致，需要确认测试断言仍对应当前事实。

## Task 3

- Goal
  - 删除 `code/frontend/src/views/platform/*.vue` 旧页面文件，并同步 backlog 当前结论。
- Touched modules or boundaries
  - `code/frontend/src/views/platform/*.vue`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Dependencies
  - Task 2
- Validation
  - `rg` 确认源码无 `views/platform/*.vue` 活引用。
  - `find code/frontend/src/views/platform -maxdepth 1 -name '*.vue'` 结果为空。
- Review focus
  - 删除范围是否只覆盖已经被 `pages/**` 替代的页面，不误删测试目录。
- Risk notes
  - 删除操作不可逆，执行前要确认对应 page owner 已稳定存在。

## Integration Checks

- `platformRoutes.ts` 仍全部指向 `pages/**`
- `views/platform` 目录只剩 `__tests__/`
- 受影响的架构 / 回归测试仍通过

## Rollback / Recovery Notes

- 若发现误删，可从 Git 恢复单个旧 view 文件。
- 删除使用回收站方式执行，避免直接不可恢复移除。

## Residual Risks

- 历史 plan / review 文档仍会保留旧 `views/platform/*.vue` 路径，这是刻意保留的时间点记录，不在本轮收口范围内。
