# 教师兼容层移除与教学读模型边界设计

## 目标

把教师侧查询能力从旧的 `teacher` 兼容壳迁移到 `teaching_readmodel`，让教师端接口直接依赖教学读模型，而不是继续保留一个只按角色命名的中转模块。

这份文档承接 `docs/superpowers/plans/2026-03-22-remove-teacher-compat.md` 的已实现结果。当前以本文和代码为最终事实源，原 plan 只用于追溯实施步骤。

## 当前状态

- 已删除后端 `internal/module/teacher` 兼容层。
- `app/composition` 中教师侧装配单元已经改为 `TeachingReadmodelModule`。
- 教师侧路由仍保留 `/api/v1/teacher/*` 作为 HTTP API 命名空间。
- API 命名空间不代表后端模块边界；当前模块边界以 `teaching_readmodel` 为准。

## 设计决策

### 1. `teacher` 不再作为后端领域模块

`teacher` 是角色视角，不是独立业务域。继续把教师查询放在 `internal/module/teacher` 容易让模块承担过多跨域聚合职责，并让后续代码误以为可以继续向这个兼容壳里追加业务逻辑。

当前约定：

- 教师侧查询读模型归属 `internal/module/teaching_readmodel`。
- 教师侧 HTTP handler、query service、repository 都按教学读模型分层组织。
- 后续新增教师查询时，优先判断是否属于教学读模型、assessment、contest 或 challenge 等明确业务边界，不新增 `teacher` 兼容模块。

### 2. HTTP 路由保持兼容

后端模块删除不等于外部 API 改名。`/api/v1/teacher/*` 已经是前端和接口文档中的角色命名空间，保留它能避免无意义的客户端迁移。

当前边界：

- `/api/v1/teacher/*` 是接口命名空间。
- `internal/module/teaching_readmodel` 是后端实现边界。
- 前端页面路由仍遵守当前项目约定：教师端页面使用 `/academy/*`，不是 `/teacher/*`。

### 3. 装配层直接暴露教学读模型

`composition.BuildTeachingReadmodelModule` 直接构造教学读模型所需依赖，并向 router 提供 handler 与 query contract。

这样做的结果是：

- router 测试可以直接断言教师查询路由落到 `teaching_readmodel`。
- 架构规则不需要继续放行 `teacher -> teaching_readmodel` 这类兼容依赖。
- 后续模块依赖检查能更早发现角色壳回流。

## 文件边界

当前核心边界：

- `code/backend/internal/app/composition/teaching_readmodel_module.go`
- `code/backend/internal/app/router.go`
- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/module/teaching_readmodel/`

不应恢复：

- 后端 `internal/module/teacher` 兼容模块
- `TeacherModule`
- `BuildTeacherModule`

## 风险与约束

- 不要把前端组件目录里的 `teacher` 命名误认为后端模块边界；前端仍有教师端组件目录，这是 UI 组织问题。
- 不要把 `/api/v1/teacher/*` 的存在理解成后端还需要 `teacher` 模块。
- 新增教师查询时，避免在 `app/router_routes.go` 内直接拼装跨模块查询逻辑，应放回明确的 read model 或业务模块。

## 验收标准

- 生产代码中没有 `internal/module/teacher` 依赖。
- router 装配使用 `TeachingReadmodelModule`。
- 教师侧查询接口仍能按原 HTTP 路径访问。
- 架构规则不再需要 `teacher -> teaching_readmodel` 兼容例外。
