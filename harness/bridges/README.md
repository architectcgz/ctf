# Harness Bridges

`harness/bridges/` 存放 `ctf` 项目本地、可提交、长期维护的 harness adapter。

边界约定：

- `scripts/`：稳定入口。给用户、hook、README、AGENTS 和其他脚本直接调用。
- `harness/bridges/`：项目本地 bridge 实现。负责把共享 harness 能力和本项目的 policy、路径、目录契约接起来。
- `.harness/`：执行态和本地态目录，不放长期 bridge。

当前 bridge 范围：

- `check-commit-message.sh`
- `check-skill-sync-reminder.sh`
- `install-agent-symlinks.sh`
- `uninstall-agent-symlinks.sh`

不属于这里的内容：

- shared `code-workflow` 安装出来的 managed 入口，例如 `start-implementation.sh`、`check-task-intake.sh`、`check-startup-gate.sh`，以及 `harness/workflow-plugins/code-workflow/archive_task_artifacts.sh`
- 业务级 guardrail、workflow stage plugin 和项目 policy
