# Code Workflow Stage Plugins

本目录用于把 `ctf` 项目自己的检查脚本挂到 shared `code-workflow` 的阶段模型上。

边界约定：

- shared `code-workflow` 只定义阶段，不定义 `ctf` 的具体守卫集合。
- `ctf` 通过 `harness/workflow-plugins/code-workflow/<stage>.d/*.sh` 注册本地检查。
- stage runner 只负责发现、排序、执行插件，不理解业务规则。

当前阶段：

- `pre-commit-quick.d/`：提交前轻量门禁，只跑当前改动强相关的快速检查。
- `completion-full.d/`：任务收尾时的完整代码级检查。
- `review-governance.d/`：review / doctor / 治理审计阶段。

不放在这里的内容：

- `scripts/check-skill-sync-reminder.sh` 这类 harness 级知识治理提醒不属于 `code-workflow` 语义本体，应由 hook 或独立 harness 入口调用，而不是注册成 task workflow plugin。

插件约定：

- 统一使用可执行 `*.sh` 脚本，按文件名字典序执行，推荐用 `10-`、`20-` 前缀排序。
- 成功返回 `0`，失败返回非 `0`。
- stage runner 会提供：
  - `WORKFLOW_STAGE`
  - `WORKFLOW_REPO_ROOT`
  - `WORKFLOW_CHANGED_FILES`
  - `WORKFLOW_TASK_SLUG`
