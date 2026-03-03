# Git Hooks

本目录用于存放可版本化的 Git hooks。

安装（在 `ctf/` 仓库根目录执行）：

```bash
bash scripts/install-githooks.sh
```

当前 hooks：

- `pre-commit`：当修改 `docs/contracts/api-contract-v1.md`（或相关文档/代码）时，自动运行 `scripts/sync_openapi_from_contract.py` 同步补全 `docs/contracts/openapi-v1.yaml`，并自动 `git add` 该文件。

