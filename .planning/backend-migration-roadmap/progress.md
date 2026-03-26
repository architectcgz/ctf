# Progress

## 2026-03-23

- 基于当前 `main` 分支完成剩余后端迁移盘点
- 将剩余工作拆成 5 个独立迁移切片
- 为每个切片创建单独的 `.planning/<task>/` 目录，便于后续逐个执行
- 完成 `identity-convergence-phase1`：
  - `adminuser` 已物理并入 `identity`
  - `identity` 新增 `UserRepository / AdminService / ProfileService / Authenticator`
  - `composition/router` 已改为通过 `IdentityModule` 装配
  - `auth` 已收缩，不再 owner 用户资料与管理能力
  - `auth` 现已进一步完成 `api/http + application + infrastructure` 物理分层，根目录已清空 concrete 实现
- 推进 `ops-convergence-phase1`：
  - `audit / dashboard / risk` 已从 `system` 收敛到 `ops`
  - `composition.SystemModule` 已通过 `ops` contract 装配对应 admin handler
  - 对外 admin 路径保持不变，`notification` 与 websocket 仍留在 `system`
- 完成 `ops-layering-phase1`：
  - `ops` 已物理拆分为 `api/http`、`application`、`infrastructure`
  - 根包仅保留对外 contract 与模块级 wrapper
- 完成 `ops-convergence-phase2`：
  - `notification` 已从 `system` 迁入 `ops`
  - `/api/v1/notifications` 与 `/ws/notifications` 路径保持不变
  - 后端 `internal/module/system` 实现已删除
- 完成 `ops-composition-phase2`：
  - `ops` composition 已切到 `typed deps + 局部 builder`
  - `audit / dashboard / risk / notification` 装配不再 inline concrete repo/service
  - `runtime` 提供给 `ops` 的 dashboard 依赖已通过 ports contract 收口

## 2026-03-26

- 完成 `contest-layering-phase2`：
  - `contest` 已删除宽 `Repository`
  - composition deps 已切到 ports 窄接口
  - `BuildContestModule` 已拆成按子能力划分的局部 builder
- 完成 `contest-crossdeps-phase2`：
  - `contest` composition 不再直接保存 `ChallengeModule` / `RuntimeModule`
  - `contest` 对 `challenge/runtime` 的跨模块依赖已收口到 typed contracts
  - AWD 注入、挑战目录、flag 校验均通过 typed deps 装配
- 完成 `challenge-layering-phase2`：
  - `challenge` 已删除宽 `ChallengeRepository`
  - application 构造依赖已切到按用例划分的窄端口
  - `challenge` composition 已收口到 typed deps，并拆成 image/core/flag/topology/writeup 局部 builder
- 完成 `practice-layering-phase2`：
  - `practice` 已删除宽 `PracticeRepository`
  - application 构造依赖已切到 command / command-tx / score / ranking 窄端口
  - `practice` composition 已收口到 typed deps，不再直接把 concrete repo 扩散给多个服务
- 完成 `practice-crossdeps-phase2`：
  - `practice` composition 已拆分 persistence deps、external deps、handler builder
  - `challenge/runtime/assessment` 跨模块依赖已从主装配函数中收口
  - `practice` runtime bridge 迁移后的装配边界进一步标准化
- 完成 `assessment-layering-phase2`：
  - `assessment` composition 已切到 typed deps
  - `BuildAssessmentModule` 已拆为 profile / recommendation / report 局部 builder
  - composition 不再直接持有 concrete assessment repo 字段
- 完成 `runtime-bridge-phase2`：
  - `practice` runtime adapter 已下沉到 `runtime` composition
  - `runtime` 对外暴露的 practice/contest 依赖已切到外部 ports 接口
  - `runtime` composition 不再反向依赖 `contest/infrastructure` 类型
- 完成 `identity-readmodel-composition-phase2`：
  - `identity` composition 已切到 `UserRepository / Authenticator` typed deps
  - `practice_readmodel` 与 `teaching_readmodel` composition 已切到 query ports typed deps
  - 轻量模块装配不再 inline new concrete repository
- 完成 `auth-composition-phase2`：
  - `auth` composition 已引入 `authModuleDeps`
  - 登录、CAS、profile、audit 依赖已通过 typed contracts 装配
  - `auth` 不再直接读取 `identity` 组合模块的私有仓储字段以外扩装配逻辑
