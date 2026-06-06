# Reuse-First Coding Agent Prompt（CTF 入口）

共享正文 owner：`/home/azhi/.agents/harness/prompts/coding-agent-system-prompt.md`

使用时先读取共享正文，再把下面这些 CTF 参数替换进去。

## CTF 参数

- `<project-search-roots>`
  - `code/frontend/src/views`
  - `code/frontend/src/components`
  - `code/frontend/src/features`
  - `code/frontend/src/widgets`
  - `code/frontend/src/composables`
  - `code/frontend/src/api`
  - `code/frontend/src/stores`
  - `code/backend/internal/module`
  - `code/backend/internal/app/composition`
  - `code/backend/internal/model`
  - `code/backend/migrations`
- `<project-pattern-policy>`：`harness/policies/project-patterns.yaml`
- `<reuse-decision-dir>`：`.harness/reuse-decisions/`
- `<reuse-index-root>`：`.harness/reuse-index/`
- `<task-intake-command>`：`bash scripts/start-implementation.sh <topic-or-slug>`

## CTF 使用补充

- 在本仓库里，新增 page、backend handler、repository、port、job、mapper、readmodel、runtime module 或 migration 前，先读 `harness/policies/project-patterns.yaml`。
- 进入非琐碎实现前，先运行 `bash scripts/check-task-intake.sh`，再用 `bash scripts/start-implementation.sh <topic-or-slug>` 建立 task worktree、implementation plan 和 startup gate。
- 复用与 owner 决策默认收口到 implementation plan 的“复用与 Owner 决策”一节；只有跨模块、高风险或 review 明确需要补充证据时，才额外写 `.harness/reuse-decisions/<task-slug>.md`。
- 本仓库的终态约束以 startup gate、implementation plan 和 review / doctor 审计为准，不再依赖旧的 `check-reuse-first` 提交前门禁链路。
