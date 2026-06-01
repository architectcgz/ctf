# frontend views 测试迁移第一阶段计划

> 状态：In Progress
> 事实输入：`code/frontend/src/views/**/__tests__`、`code/frontend/src/pages/**`、`code/frontend/src/features/**`、`code/frontend/src/widgets/**`、`code/frontend/scripts/frontend-architecture-policy.json`

## Plan Summary

- Objective
  - 把 `src/views` 中剩余的前端测试资产迁到更符合最终架构的位置，并收口守卫脚本与策略测试。
- Non-goals
  - 不修改历史 review / plan 文档中的旧验证命令。
  - 不在本轮改变页面、feature、widget 的运行时实现 owner。
  - 不把行为测试重新集中到新的“测试仓库目录”里。
- Source architecture or design docs
  - `docs/architecture/README.md`
  - `docs/architecture/frontend/README.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Dependency order
  - 先迁架构守卫与 policy，再迁页面行为测试，再压缩阶段性策略测试，最后清理 `views` 测试残留。
- Expected specialist skills
  - `frontend-sliced-architecture`
  - `frontend-engineer`
  - `development-pipeline`

## Task 1

- Goal
  - 让前端架构守卫退出 `src/views`，并让脚本、policy、文档入口与实际路径一致。
- Touched modules or boundaries
  - `scripts/check-frontend-architecture.sh`
  - `code/frontend/src/__tests__/*`
  - `code/frontend/scripts/frontend-architecture-policy.json`
- Dependencies
  - 无
- Validation
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- Review focus
  - policy 中是否还保留把 `views` 当作活动 layer 的历史残片。
- Risk notes
  - 若脚本或入口文档仍指向旧路径，会造成 guardrail 假通过或假失败。

## Task 2

- Goal
  - 把 route page 行为测试从 `views/**/__tests__` 迁回 `pages/**`、`features/**`、`widgets/**` 邻近。
- Touched modules or boundaries
  - `code/frontend/src/views/**/__tests__/*.test.ts`
  - `code/frontend/src/pages/**/__tests__/*.test.ts`
  - `code/frontend/src/features/**/__tests__/*.test.ts`
  - `code/frontend/src/widgets/**/__tests__/*.test.ts`
- Dependencies
  - Task 1
- Validation
  - 针对迁移分组执行最小充分 Vitest 回归。
- Review focus
  - 测试 owner 是否与当前运行时 owner 一致，而不是只改路径。
- Risk notes
  - 个别测试如果仍在读取 feature / widget 源码，最终落点应以真实 owner 为准，而不是 route page 所在目录。

## Task 3

- Goal
  - 合并 `Phase / Extraction / SurfaceAlignment` 这一批阶段性迁移测试，改成稳定策略测试。
- Touched modules or boundaries
  - `code/frontend/src/views/platform/__tests__/*.test.ts`
  - `code/frontend/src/views/teacher/__tests__/*.test.ts`
  - 新的策略测试 owner 目录
- Dependencies
  - Task 2
- Validation
  - 相关策略测试通过，并删除被其覆盖的阶段性测试文件。
- Review focus
  - 新守卫要描述稳定结构事实，不再依赖“第几阶段”“从哪个旧文件抽取出来”的历史语义。
- Risk notes
  - 如果合并过度，可能会把定位能力丢掉；需要保留足够清晰的失败信息。

## Integration Checks

- `src/views` 不再承载活动架构守卫
- 页面行为测试按 page / feature / widget owner 分布
- `scripts/check-frontend-architecture.sh` 与 `docs/architecture/README.md` 指向当前真实测试路径
- `frontend-architecture-policy.json` 不再把 `views` 当作活动 layer

## Rollback / Recovery Notes

- 测试迁移以 Git 历史可回退；文件删除若发生，优先通过 Git 恢复单个测试文件。
- 若某批迁移引起定位困难，可先恢复该批文件，再重新拆成更明确的稳定策略测试。

## Residual Risks

- 历史 plan / review 文档仍会保留旧 `src/views/**/__tests__` 路径，这是时间点证据，不在本轮统一改写范围内。
