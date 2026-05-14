# CTF 契约入口

## 当前契约入口

- `api-contract-v1.md`：面向人读的 v1 HTTP / WebSocket 契约说明。
- `openapi-v1.yaml`：面向工具和引用方的稳定 OpenAPI bundle 产物。
- `openapi-v1/`：OpenAPI v1 的拆分源目录，按 `paths/` 与 `components/schemas/` 分组维护。
- `challenge-pack-v1.md`：题包格式契约。

## OpenAPI 维护方式

- 修改 OpenAPI 时，先改 `openapi-v1/` 下的拆分源，不直接手工编辑 `openapi-v1.yaml`。
- 根入口是 `openapi-v1/index.yaml`，负责聚合 `paths/*.yaml` 和 `components/schemas/*.yaml`。
- 修改后运行 `python3 scripts/sync_openapi_from_contract.py` 生成 bundle。
- 提交前运行 `python3 scripts/sync_openapi_from_contract.py --check` 或 `bash scripts/check-consistency.sh`，确认 source 与 bundle 没有漂移。
