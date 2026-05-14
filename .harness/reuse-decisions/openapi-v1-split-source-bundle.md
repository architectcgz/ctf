# Reuse Decision

## Change type

- schema
- docs
- script

## Existing code searched

- `docs/contracts/openapi-v1.yaml`
- `docs/contracts/api-contract-v1.md`
- `docs/文档规范.md`
- `docs/README.md`
- `scripts/check-code-changes.sh`
- `scripts/check-consistency.sh`
- `scripts/install-githooks.sh`
- `AGENTS.md`

## Similar implementations found

- `docs/contracts/openapi-v1.yaml`
  - 现有 OpenAPI 已经是稳定事实源路径，被契约文档、架构文档和检查脚本广泛直接引用。
- `scripts/check-code-changes.sh`
  - 现有 API surface guardrail 把 `docs/contracts/openapi-v1.yaml` 当作契约更新凭据之一。
- `scripts/check-consistency.sh`
  - 现有 harness 一致性检查已经是 repo 内机械化 guardrail 的统一入口，适合接入 OpenAPI source/bundle 同步检查。
- `scripts/install-githooks.sh`
  - 已经预留 `scripts/sync_openapi_from_contract.py`，但脚本缺失，说明 repo 需要把 OpenAPI 同步链补成真实能力，而不是继续手工维护单文件。

## Decision

extend_existing

## Reason

这次不是把 `docs/contracts/openapi-v1.yaml` 从 repo 中移走，而是在保持稳定 bundle 路径不变的前提下，引入可维护的拆分源目录。

选定方向：

- 保留 `docs/contracts/openapi-v1.yaml` 作为对外稳定 bundle 产物。
- 新增 `docs/contracts/openapi-v1/` 作为拆分源目录，按 `paths/` 和 `components/schemas/` 分组维护。
- 新增 `scripts/sync_openapi_from_contract.py`，负责从拆分源生成 bundle，并提供 `--check` 防止 source/bundle 漂移。
- 复用现有 `scripts/check-consistency.sh`、`scripts/check-code-changes.sh` 和文档入口，不再引入第二套检查链。

不采用的方向：

- 不直接把根文件改成依赖外部 `$ref` 的“半成品入口”，避免现有消费链都被迫理解拆分结构。
- 不把 bundle 改名或迁走，否则会同时打断文档引用、review 引用和 API guardrail。
- 不继续手工编辑 5000+ 行单文件，避免契约维护成本继续增长。

## Files to modify

- `.harness/reuse-decisions/openapi-v1-split-source-bundle.md`
- `docs/plan/impl-plan/2026-05-14-openapi-v1-split-source-and-bundle-implementation-plan.md`
- `docs/contracts/README.md`
- `docs/contracts/api-contract-v1.md`
- `docs/contracts/openapi-v1.yaml`
- `docs/contracts/openapi-v1/`
- `docs/README.md`
- `docs/文档规范.md`
- `AGENTS.md`
- `scripts/check-code-changes.sh`
- `scripts/check-consistency.sh`
- `scripts/sync_openapi_from_contract.py`
- `harness/reuse/history.md`

## After implementation

- 后续编辑 OpenAPI 时，只需要改 `docs/contracts/openapi-v1/` 下的拆分源。
- `docs/contracts/openapi-v1.yaml` 继续作为稳定、可引用、可导出的 bundle 文件存在。
- 如果 source 与 bundle 不一致，`scripts/check-consistency.sh` 会直接失败，避免再次回到“手工改两份”的状态。
