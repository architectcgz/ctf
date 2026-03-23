# Findings

- `identity` 现在只是一层 token parser 包装，不是真正的用户 owner。
- `adminuser` 仍是独立完整模块，composition 和 router 直接装配它。
- `auth` 仍同时拥有 repository、service、handler 和用户登录流程的主要实现。

## After Phase 1

- `identity` 已收拢用户主数据 owner：用户 repository、管理端用户服务、profile/password 服务都在 `identity` 下。
- `adminuser` 已物理删除，外部 `/api/v1/admin/users*` 路径保持不变。
- `auth` 仍保留登录、注册、CAS、refresh/logout、ws-ticket 等认证流程，但不再 owner 用户资料与管理能力。
- `composition/router` 已新增 `IdentityModule`，认证中间件和管理端用户路由都改为依赖 `identity`。
