# OpenAPI v1 拆分源与 Bundle 同步实施计划

> 状态：Draft
> 事实源：`docs/contracts/openapi-v1.yaml`、`docs/contracts/api-contract-v1.md`、`docs/文档规范.md`、`scripts/check-code-changes.sh`、`scripts/check-consistency.sh`
> 替代：无

## 1. 目标

- 把 `docs/contracts/openapi-v1.yaml` 从单文件手工维护改成“拆分源 + 稳定 bundle 产物”
- 保持 `docs/contracts/openapi-v1.yaml` 路径不变，避免现有文档和脚本引用失效
- 为 OpenAPI source/bundle 增加可重复执行的同步脚本与一致性检查

## 2. 非目标

- 不改变任何 API 路径、字段、schema 语义或 tag 归类
- 不把现有消费链改成直接读取外部 `$ref`
- 不在本轮引入第三方 OpenAPI bundler 或新的 Node/Python 依赖管理

## 3. Brainstorming 结论

- 候选方向 A：继续维护单文件
  - 不采用：`openapi-v1.yaml` 已经 5469 行，后续契约演进会越来越难 review 和 merge。
- 候选方向 B：直接把 `openapi-v1.yaml` 改成外部 `$ref` 入口
  - 不采用：现有检查脚本和大量文档把它当成稳定 bundle 文件使用，不应该把消费复杂度泄露给下游。
- 候选方向 C：新增拆分源目录，保留原 bundle 路径
  - 采用：对外路径稳定，内部维护成本下降，还能把 source/bundle drift 交给脚本检查。
- 选定方向：`docs/contracts/openapi-v1/` 维护拆分源，`scripts/sync_openapi_from_contract.py` 生成 `docs/contracts/openapi-v1.yaml`

## 4. owner 与边界

### 4.1 OpenAPI source owner

- `docs/contracts/openapi-v1/`
- 负责：
  - 维护 OpenAPI 拆分源
  - 按领域分组 `paths/` 与 `components/schemas/`
  - 作为人类编辑入口
- 不负责：
  - 对外消费兼容
  - 文档和脚本对 bundle 的历史引用迁移

### 4.2 OpenAPI bundle owner

- `scripts/sync_openapi_from_contract.py`
- 负责：
  - 读取拆分源
  - 合并 `$merge` 片段
  - 生成稳定 bundle 到 `docs/contracts/openapi-v1.yaml`
  - 提供 `--check` 校验 source/bundle 是否同步
- 不负责：
  - 修改 API 语义
  - 推断 schema 分组策略

### 4.3 契约入口与 guardrail owner

- `docs/contracts/README.md`、`docs/README.md`、`docs/文档规范.md`、`AGENTS.md`
  - 负责说明 source/bundle 的事实源关系和编辑入口
- `scripts/check-code-changes.sh`、`scripts/check-consistency.sh`
  - 负责在契约变更与日常检查里守住同步约束

## 5. 任务切片

### Slice 1：建立拆分源结构与同步脚本

目标：

- 建立 `docs/contracts/openapi-v1/` 源目录
- 提供 bundle 生成脚本

改动面：

- `docs/contracts/openapi-v1/`
- `scripts/sync_openapi_from_contract.py`
- `scripts/install-githooks.sh`

验证：

- `python3 scripts/sync_openapi_from_contract.py`
- `python3 scripts/sync_openapi_from_contract.py --check`

review focus：

- `$merge` 规则是否简单、稳定、可读
- bundle 入口路径是否保持不变
- 生成脚本是否会在重复 key 时直接失败

### Slice 2：迁移现有 OpenAPI 内容到拆分源

目标：

- 把现有 `paths` 和 `components.schemas` 按分组迁入 source
- 生成语义等价的 bundle

改动面：

- `docs/contracts/openapi-v1.yaml`
- `docs/contracts/openapi-v1/index.yaml`
- `docs/contracts/openapi-v1/paths/*.yaml`
- `docs/contracts/openapi-v1/components/schemas/*.yaml`

验证：

- `python3 scripts/sync_openapi_from_contract.py --check`

review focus：

- 是否存在遗漏 path / schema
- 内部 `#/components/schemas/*` 引用是否保持可解析
- bundle 是否完全由 source 生成，而不是继续混合手工编辑

### Slice 3：更新文档入口与机械检查

目标：

- 注册新路径
- 让 repo guardrail 识别新的 source/bundle 关系

改动面：

- `docs/contracts/README.md`
- `docs/contracts/api-contract-v1.md`
- `docs/README.md`
- `docs/文档规范.md`
- `AGENTS.md`
- `scripts/check-code-changes.sh`
- `scripts/check-consistency.sh`
- `harness/reuse/history.md`

验证：

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

review focus：

- 文档是否清楚说明“编辑 source，不直接编辑 bundle”
- 新路径是否按文档规范完成登记
- API contract guardrail 是否接受 source 目录作为契约更新依据

## 6. 风险与回退

- 如果 bundle 路径变化，现有文档和脚本引用会大面积失效；因此 bundle 路径必须固定不动。
- 如果 source 和 bundle 都允许手工维护，后续会再次产生双写漂移；因此必须用脚本单向生成 bundle。
- 如果把 schema 粒度拆得过细，会得到一堆难以导航的小文件；本轮按领域分组，而不是一 schema 一文件。
- 回退方式：
  - 若拆分方案不可用，可保留 `openapi-v1.yaml` 并移除新 source 目录与同步脚本，恢复单文件维护。
  - 因 bundle 路径不变，回退不会影响外部引用，只影响内部维护方式。

## 7. 最终验证

- `python3 scripts/sync_openapi_from_contract.py`
- `python3 scripts/sync_openapi_from_contract.py --check`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-code-changes.sh`
- `bash scripts/check-workflow-complete.sh`

## 8. Checklist

- [x] Slice 1：建立拆分源结构与同步脚本
- [x] Slice 2：迁移现有 OpenAPI 内容到拆分源
- [x] Slice 3：更新文档入口与机械检查
- [x] 最终验证完成并通过
