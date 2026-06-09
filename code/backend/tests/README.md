# Backend Tests

这里放后端跨模块测试的目标目录说明，避免系统级测试继续堆回 `internal/app`。

## 当前约定

- `internal/app` 继续保留必须贴着 `package app` 的测试，以及旧系统测试的兼容壳。
- 共享测试环境、请求 helper、通用断言先收口到 `internal/testutil/`，当前 phase 1 落点是 `internal/testutil/systemapp`。
- 新增系统级测试时，优先复用 testutil，不再把大段 helper owner 放回 `internal/app/*_test.go`。
- TDD 写出的测试默认是行为规格和回归护栏，不因为对应功能已经实现就删除；只在行为信号重复、实现细节锁定、迁移 guard 到期，或目标行为明确废弃时合并或删除。
- 测试文件变多时先按 owner 和层级重组，再考虑抽 helper；不要只为了减少文件数量把跨角色、跨模块、跨 runtime 的场景重新堆进一个大测试文件。

## 放置判断

新增后端测试前先判断这条测试证明的是哪一层契约：

- 包内 `*_test.go`：模块内业务语义、未导出实现、application service / command / query / repository 的局部契约；需要访问包私有符号时默认留在源码旁边。
- `internal/testutil/*`：需要接近内部实现、会被多个包内或系统测试复用的测试工具；不要放只服务单个测试文件的一次性封装。
- `tests/architecture`：源码级架构 guardrail，只检查边界、目录和迁移约束，不跑业务语义回归。
- `tests/system/http/<scenario>`：黑盒 HTTP / router 级长场景，只表达请求、角色、状态和响应断言；环境搭建、seed 和通用断言优先复用 testutil / testkit。
- `tests/runtime`：需要 PostgreSQL、runtime agent、容器端口、外部进程或真实 migration 参与的集成测试。
- `tests/testkit`：跨 `tests/*` 复用的场景 builder、fixture、assert helper 和测试数据工厂；不访问未导出实现。

同一行为不要同时在 handler、application、repository、HTTP 系统测试里重复断言。需要多层覆盖时，每一层必须证明不同契约：模块内证明业务规则，HTTP 系统测试证明路由组合 / 权限 / 序列化，runtime 测试证明真实依赖协作。

## 目标目录

### `tests/architecture`

- 放测试架构 guardrail。
- 关注 `internal/app` 系统测试壳是否继续变薄，以及 `tests/system/http` 是否仍然只持有 scenario/assertion owner。
- 这里不跑业务语义回归，只做低噪音的源码级边界检查。

### `tests/system/http`

- 放黑盒 HTTP / router 级系统测试。
- 关注角色权限、路由组合、端到端业务流和回归矩阵。
- 测试文件应尽量只表达场景，不再内嵌环境搭建和数据种子细节。
- 当前首个落点是 `tests/system/http/practiceflow/`，这里承接 `practice_flow` 的场景断言 owner。
- `tests/system/http/fullrouteraccess/` 承接 `full_router_access` 的场景断言 owner，当前仍通过 glue code 复用 `internal/app` 里的 full router fixture。
- `tests/system/http/fullrouteradmin/` 承接 `full_router_admin` 的 HTTP 场景断言 owner，当前已覆盖 AWD control、publish request lifecycle、admin challenge management 和 admin ops/notification 场景；page size 小回归仍暂留 `internal/app`。
- `tests/system/http/fullrouterteacherauthoring/` 承接 `full_router_teacher_authoring` 的 HTTP 场景断言 owner，当前仍通过 glue code 复用 `internal/app` 的 full router fixture，并保留少量 DB / 文件系统 seed。
- `tests/system/http/fullrouterawdstate/` 承接 `full_router_awd_state_matrix` 的 HTTP 场景断言 owner，当前仍通过 glue code 复用 `internal/app` 的 AWD/contest seed 和实例状态更新 helper。
- `tests/system/http/fullrouterconteststate/` 承接 `full_router_contest_state_matrix` 的 HTTP 场景断言 owner，当前仍通过 glue code 复用 `internal/app` 的 contest/challenge/user seed、scoreboard seed 和 report wait helper。
- `tests/system/http/fullrouterteacherstate/` 承接 `full_router_teacher_state_matrix` 中非 AWD review 的 HTTP 场景断言 owner；`TeacherAWDReviewExportStateMatrix` 因既有失败暂留 `internal/app` 单独处理。
- `tests/system/http/fullrouterreportstate/` 承接 `ReportPreviewAndDownloadStateMatrix` 的 HTTP 场景断言 owner；module builder 测试和报表共享 helper 仍暂留 `internal/app`。

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
