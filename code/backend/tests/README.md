# Backend Tests

这里放后端跨模块测试的目标目录说明，避免系统级测试继续堆回 `internal/app`。

## 当前约定

- `internal/app` 继续保留必须贴着 `package app` 的测试，以及旧系统测试的兼容壳。
- 共享测试环境、请求 helper、通用断言先收口到 `internal/testutil/`，当前 phase 1 落点是 `internal/testutil/systemapp`。
- 新增系统级测试时，优先复用 testutil，不再把大段 helper owner 放回 `internal/app/*_test.go`。

## 目标目录

### `tests/system/http`

- 放黑盒 HTTP / router 级系统测试。
- 关注角色权限、路由组合、端到端业务流和回归矩阵。
- 测试文件应尽量只表达场景，不再内嵌环境搭建和数据种子细节。

### `tests/runtime`

- 放依赖 runtime / practice / 容器环境的集成测试。
- 这里承接需要 PostgreSQL、runtime agent、容器端口或外部进程参与的测试。
- 与纯 HTTP 流程分开，避免 `internal/app` 同时承担路由和运行时环境两类 owner。

### `tests/testkit`

- 放可被多个测试目录复用的场景 builder、fixture、assert helper 和测试数据工厂。
- 优先放稳定复用能力，不放只服务单个测试文件的一次性封装。
- 如果 helper 必须访问未导出实现，继续留在 `internal/testutil/*` 或对应包内测试文件。

## 迁移原则

1. 先抽 helper owner，再迁测试场景，最后删除兼容壳。
2. 需要访问未导出符号的测试，默认继续贴近源码目录，不强行迁出。
3. 跨模块、跨角色、跨运行时的长流程测试，后续统一往 `tests/system/http` 或 `tests/runtime` 收。
