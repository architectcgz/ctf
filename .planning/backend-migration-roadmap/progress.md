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
