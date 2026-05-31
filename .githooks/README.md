# Git Hooks

本目录用于存放可版本化的 Git hooks。

安装（在 `ctf/` 仓库根目录执行）：

```bash
bash scripts/install-githooks.sh
```

当前 hooks：

- `pre-commit`：仅当修改 API 相关文件时自动运行 `scripts/sync_openapi_from_contract.py`，当前包含 API 合同文档、`docs/architecture/backend/04-api-design.md`、后端路由/handler/dto、统一响应与错误码；不会再因容器、数据库等非 API 改动触发。
- `commit-msg`：运行 `scripts/check-commit-message.sh`，要求普通提交采用“标题 + 正文”结构。标题继续使用英文类型前缀，如 `fix`、`refactor`、`docs`，可选 scope 放在英文括号里，冒号后的描述必须包含中文说明；正文至少两行有效内容，且需要有足够具体的变更说明。

<!-- BEGIN HARNESS ENGINEERING: hook-docs -->

## Harness 检查

- `pre-commit`：运行 `scripts/check-consistency.sh`，检查严格参考 harness 的顶层目录、导航和资料计数。
- `pre-commit`：运行 `scripts/check-reuse-first.sh --staged`，要求受保护页面、组件、hook、API wrapper、store、表单、表格和 schema 变更先完成复用决策。
- `pre-commit`：运行 `scripts/check-architecture.sh --quick`，检查后端模块依赖方向和前端分层边界。
- `pre-commit`：运行 `scripts/check-skill-sync-reminder.sh --staged`，当 `feedback/`、`harness/reuse/history.md`、`harness/reuse/index.yaml` 或 `harness/prompts|policies|templates/` 变更时，非阻塞提醒是否需要同步到全局 skill。
- `commit-msg`：运行 `scripts/check-commit-message.sh`，阻止中文类型前缀、纯英文描述标题，或只有简短标题而没有详细正文的普通提交进入历史。
- 原有 API 合同同步逻辑继续保留。

## 本地工作流优先

- 本仓库的 reuse-first 约束以本地 hook 和本地脚本为准，不依赖 GitHub Actions 才能生效。
- 安装 hook 后，提交前会先执行：
  1. `scripts/check-consistency.sh`
  2. `scripts/check-reuse-first.sh --staged`
  3. `scripts/check-architecture.sh --quick`
  4. `scripts/check-skill-sync-reminder.sh --staged`
  5. `scripts/check-commit-message.sh <commit-message-file>`（在 `commit-msg` 阶段执行）
- 若需要在提交前手工自检，可直接运行 `bash scripts/check-reuse-first.sh --staged`。
- 推荐提交流程示例：

```bash
git commit -m "refactor(platform): 拆分用户页迁移边界" \
  -m "把页面流程 owner 收回 feature 层，避免 platform 目录继续混放行为逻辑。" \
  -m "同步调整接线和测试，保证后续 instance 迁移可以沿同样模式推进。"
```
<!-- END HARNESS ENGINEERING: hook-docs -->
