# Findings

- `teacher -> teaching_readmodel` 已完成，说明“先删兼容层，再收紧架构规则”的迁移模式可复用。
- `identity` 已完成 Phase 1：用户主数据、管理端用户服务、profile/password 已收拢；但认证流程代码仍主要位于 `auth` 包内，后续还需继续物理内聚。
- `ops` 仍处于过渡态。
- `challenge`、`contest`、`assessment`、`practice` 内部还基本是根目录大平铺，内部物理分层尚未真正开始。
- 现有 `.planning/practice-readmodel-phase2/` 可以作为后续 readmodel 迁移的参考样式。
- workspace 共享 Go 规范已经沉淀出可复用的“模块内部最终分层模板”：
  - `runtime` 负责依赖装配
  - `api/http` 或 `transport/http` 只做协议映射
  - `application/commands` 与 `application/queries` 分离
  - `ports` 定义输入/输出依赖接口
  - `infrastructure/*` 负责具体实现
- `ctf` 不应整仓改成 `internal/services/<service>` 的服务优先结构，但每个 `internal/module/<module>` 都可以按这个共享模板做内部物理分层。
