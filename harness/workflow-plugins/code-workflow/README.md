# Code Workflow Stage Plugins

本目录用于把 `ctf` 项目自己的检查脚本挂到 shared `code-workflow` 的阶段模型上。

边界约定：

- shared `code-workflow` 只定义阶段，不定义 `ctf` 的具体守卫集合。
- `ctf` 通过 `harness/workflow-plugins/code-workflow/<stage>.d/*.sh` 注册本地检查。
- `harness/workflow-plugins/code-workflow/run_workflow_stage.sh` 是项目内 stage runner 主体；`scripts/run-workflow-stage.sh` 只保留稳定入口 wrapper。
- `harness/workflow-plugins/code-workflow/archive_task_artifacts.sh` 是 shared `code-workflow` 安装到项目 harness 内的归档入口；不再保留 `scripts/archive-task-artifacts.sh`。
- stage runner 只负责发现、排序、执行插件，不理解业务规则。

当前阶段：

- `pre-commit-quick.d/`：提交前轻量门禁，只跑当前改动强相关的快速检查。
- `completion-full.d/`：任务收尾时的完整代码级检查。
- `workflow-governance.d/`：`code-workflow` 后置的 workflow 治理审计阶段。

阶段时机：

- `pre-commit-quick` 只放“每次提交都不该退化”的便宜硬约束。目标是尽早阻断错误边界，避免坏状态继续扩散到后续提交。
- `completion-full` 放明显更重、但适合在任务收尾时统一确认的完整检查；它不是用来替代 pre-commit，而是补齐 pre-commit 故意没有放进去的重检查。
- `workflow-governance` 放仓库导航、文档事实源、OpenAPI 同步、project workflow guardrail 接线这类治理审计；这些不要求每次提交都无条件阻断，但在 workflow 收尾 / doctor 前必须可验证。
- `AGENTS.md / CLAUDE.md` 入口一致性和 shared skill bridge 不再把 `workflow-governance` 当 owner；主检查放在 `scripts/check-agent-entrypoints.sh`、`scripts/check-shared-skills.sh`，并由 `scripts/doctor-local-harness.sh` 与 `scripts/install-agent-symlinks.sh` 提前执行。

当前 `ctf` 的架构检查分层：

- `pre-commit-quick`
  - 前端过细测试守卫
  - startup gate
  - quick 架构检查
- quick 架构检查当前只放便宜的硬边界：
  - backend：`TestModuleArchitectureBoundaries`
  - frontend：`check:vue-deep`、`architectureBoundaries.test.ts`、`routePageArchitectureBoundary.test.ts`
- `completion-full`
  - code change contract checks
  - backend full architecture
  - frontend full architecture
- full 架构检查才覆盖更重的内容：
  - backend：`internal/app` composition / context 约束、`tests/architecture`
  - frontend：growth guard、feature owner boundaries、overlay boundaries、theme tail guard

为什么 backend module boundary 放在 `pre-commit-quick`：

- 它足够快，适合每次提交都执行。
- 它是仓库不变量，而不是只在“任务完成时”才关心的质量项。
- 一旦跨模块边界在早期提交里被打破，后续实现很容易建立在错误 owner 上，返工成本会明显上升。

不放在这里的内容：

- `scripts/check-skill-sync-reminder.sh` 这类 harness 级知识治理提醒不属于 `code-workflow` 语义本体，应由 hook 或独立 harness 入口调用，而不是注册成 task workflow plugin。

兼容说明：

- 历史命令里的 `review-governance` / `check-review-governance.sh` 仍可继续调用，但它们现在只作为兼容别名转发到 `workflow-governance`。

插件约定：

- 统一使用可执行 `*.sh` 脚本，按文件名字典序执行，推荐用 `10-`、`20-` 前缀排序。
- 成功返回 `0`，失败返回非 `0`。
- stage runner 会提供：
  - `WORKFLOW_STAGE`
  - `WORKFLOW_REPO_ROOT`
  - `WORKFLOW_CHANGED_FILES`
  - `WORKFLOW_TASK_SLUG`
