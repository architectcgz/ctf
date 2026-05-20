# teaching_readmodel 重命名为 teaching_query 实施计划

> 状态：Draft
> 输入：`docs/architecture/backend/01-system-architecture.md`、`docs/architecture/backend/07-modular-monolith-refactor.md`、`code/backend/internal/module/teaching_readmodel/`

## 目标

- 将当前教师侧跨 owner 查询聚合模块从 `teaching_readmodel` 重命名为 `teaching_query`。
- 同步收口代码、装配、测试和当前事实文档中的命名，避免继续把该模块表述成“物理读写分离”或“单独一层 readmodel”。

## 非目标

- 不调整该模块的查询职责边界，不把目录查询、班级洞察或学生复盘继续拆成更多子模块。
- 不重写教师侧查询 SQL、权限规则或 DTO 结构。
- 不批量改写历史 review / impl-plan / archive 文档里的历史叙述，除非当前事实文档或入口文档明确依赖这些名称。

## 影响范围

- 代码目录：
  - `code/backend/internal/module/teaching_readmodel/**`
  - `code/backend/internal/app/composition/teaching_readmodel_module.go`
  - `code/backend/internal/app/router_routes.go`
  - `code/backend/internal/app/*integration_test.go`
  - `code/backend/cmd/seed-teaching-review-data/main.go`
- 当前事实文档与入口：
  - `README.md`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/design/backend-module-boundary-target.md`
  - `harness/policies/project-patterns.yaml`

## 切片

### Slice 1：代码级重命名

- Goal：把模块目录、composition builder、import path、typed alias 和测试引用从 `teaching_readmodel` 收口到 `teaching_query`。
- Validation：
  - `go test ./internal/app/... ./internal/module/teaching_query/... ./cmd/seed-teaching-review-data -count=1`
- Review Focus：
  - import path 是否完整收口
  - builder / alias / router 测试是否仍引用旧名称
  - 模块装配顺序和路由归属是否保持不变

### Slice 2：当前事实文档与入口同步

- Goal：只更新当前事实文档和必要入口，把模块名称解释成“教师侧查询聚合模块”，不把它继续写成特殊“readmodel 层”。
- Validation：
  - `python3 scripts/check-docs-consistency.py`
- Review Focus：
  - 是否只改当前事实源和入口
  - 是否避免把历史 review / 计划文档改成失真叙事

## 风险

- 当前仓库中存在大量历史 review、计划和设计文档引用旧路径；若全量替换，容易放大 diff 并污染历史语境。
- 路由和 integration test 对 builder 名称、模块装配顺序和 handler 包路径有显式断言，容易因为漏改而失败。

## 回退

- 若改名后出现大面积路径引用失败，优先保留代码级改名，回滚非必要的历史文档批量替换，只收口当前事实文档和入口。
