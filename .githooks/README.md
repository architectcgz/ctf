# Git Hooks

本目录用于存放可版本化的 Git hooks。

安装（在 `ctf/` 仓库根目录执行）：

```bash
bash scripts/install-githooks.sh
```

当前 hooks：

- `pre-commit`：仅当修改 API 相关文件时自动运行 `scripts/sync_openapi_from_contract.py`，当前包含 API 合同文档、`docs/architecture/backend/04-api-design.md`、后端路由/handler/dto、统一响应与错误码；不会再因容器、数据库等非 API 改动触发。

<!-- BEGIN HARNESS ENGINEERING: hook-docs -->
## Harness 检查

- `pre-commit`：运行 `scripts/check-consistency.sh`，检查严格参考 harness 的顶层目录、导航和资料计数。
- 原有 API 合同同步逻辑继续保留。
<!-- END HARNESS ENGINEERING: hook-docs -->
