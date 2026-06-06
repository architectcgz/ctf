# Git Hooks

本目录用于存放可版本化的 Git hooks。

安装（在 `ctf/` 仓库根目录执行）：

```bash
bash scripts/install-githooks.sh
```

当前 hooks：

- `pre-commit`：运行本地轻量门禁，包括前端过细样式测试守卫、startup gate、架构护栏和 skill sync reminder。
- `commit-msg`：运行 `scripts/check-commit-message.sh`。本地 wrapper 会调用全局 `~/.agents/harness/commit-message/check_commit_message.py`，再读取仓库内 `harness/policies/commit-message.json` 执行项目策略。普通提交仍要求采用“标题 + 正文”结构：标题继续使用英文类型前缀，如 `fix`、`refactor`、`docs`，可选 scope 放在英文括号里，冒号后的描述必须包含中文说明；正文至少两行有效内容，且需要有足够具体的变更说明。若当前 worktree 有激活中的非琐碎任务 gate，正文还必须显式写一行 `Task: <task-slug>`；这类 `Task:` 元数据不再计入正文说明行数和信息量统计。

<!-- BEGIN HARNESS ENGINEERING: hook-docs -->

## Harness 检查

- `pre-commit`：运行 `scripts/check-frontend-test-guard.sh --staged`，阻止新增只锁 `class`、局部 markup、`padding/gap` 文本的过细前端样式测试。
- `pre-commit`：运行 `scripts/check-startup-gate.sh --staged`，要求命中启动门禁的非琐碎改动先通过统一入口并拥有合法启动凭证。
- `pre-commit`：运行 `scripts/check-architecture.sh --quick`，检查后端模块依赖方向和前端分层边界。
- `pre-commit`：运行 `scripts/check-skill-sync-reminder.sh --staged`，当 `feedback/`、`harness/reuse/history.md`、`harness/reuse/index.yaml` 或 `harness/prompts|policies|templates/` 变更时，非阻塞提醒是否需要同步到全局 skill。
- `commit-msg`：运行 `scripts/check-commit-message.sh`，通过全局共享检查器 + 本地 policy 阻止中文类型前缀、纯英文描述标题、只有简短标题而没有详细正文的普通提交进入历史；若存在激活中的 task gate，还要求正文显式带上 `Task: <task-slug>`。
- 原有 API 合同同步逻辑继续保留。

## 本地工作流优先

- `scripts/check-review-governance.sh` 现在作为 review / doctor 级治理审计入口，不再作为所有提交前都无条件执行的本地门禁。
- `scripts/check-consistency.sh` 只保留为兼容别名，内部直接转发到 `scripts/check-review-governance.sh`。
- OpenAPI 拆分源与 bundle 的同步检查继续放在 `scripts/check-review-governance.sh`，不按提交路径临时下沉到 `pre-commit`。
- 安装 hook 后，提交前会先执行：
1. `scripts/check-frontend-test-guard.sh --staged`
2. `scripts/check-startup-gate.sh --staged`
3. `scripts/check-architecture.sh --quick`
4. `scripts/check-skill-sync-reminder.sh --staged`
6. `scripts/check-commit-message.sh <commit-message-file>`（在 `commit-msg` 阶段执行）
- 写完前端测试后，推荐先手工运行：

```bash
bash scripts/check-frontend-test-guard.sh
bash scripts/check-frontend-test-guard.sh --files code/frontend/src/path/to/example.test.ts
```

- `bash scripts/check-frontend-test-guard.sh` 默认检查当前工作区里变更过的前端测试文件。
- `bash scripts/check-frontend-test-guard.sh --staged` 保持给 `pre-commit` 使用，只检查暂存区新增断言。
- `bash scripts/check-frontend-test-guard.sh --files <path...>` 适合只针对当前正在改的测试文件做显式检查。
- 命中非琐碎启动门禁的任务，应先运行 `bash scripts/start-implementation.sh <topic-or-slug>` 创建独立 worktree、主 plan 和本地启动凭证。
- 当 task plan 已归档、但当前 worktree 还要完成最终提交或合并清理时，startup gate 会进入 `ready_to_merge` 状态；这仍然允许当前 task 继续提交，不需要重新起新 task。
- 需要做完整仓库治理审计或 review 前自检时，再手工运行 `bash scripts/check-review-governance.sh`；保留旧命令时，也可以继续用 `bash scripts/check-consistency.sh`。
- 推荐提交流程示例：

```bash
git commit -m "refactor(platform): 拆分用户页迁移边界" \
  -m "把页面流程 owner 收回 feature 层，避免 platform 目录继续混放行为逻辑。" \
  -m "同步调整接线和测试，保证后续 instance 迁移可以沿同样模式推进。" \
  -m "Task: 2026-06-05-platform-user-page-owner-split"
```
<!-- END HARNESS ENGINEERING: hook-docs -->
