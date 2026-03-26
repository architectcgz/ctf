# Findings

- `teacher -> teaching_readmodel` 与 `practice_readmodel` 根壳清理已完成，说明“先删兼容层，再收紧架构规则”的迁移模式可复用。
- `identity`、`auth`、`ops` 已完成根包瘦身与物理分层，当前根目录均只保留架构守卫测试。
- `challenge`、`contest`、`assessment`、`practice` 已完成 Phase 1 收口，且后续已经继续推进到 `commands / queries / ports / domain` 分层。
- 当前 composition 收尾扫描未再发现跨模块读取 `IdentityModule` / `RuntimeModule` 私有嵌套字段的存量用法；跨模块装配已切到公开 contract。
- 现有 `.planning/practice-readmodel-phase2/` 仍可作为 readmodel 迁移样式参考，但原 roadmap 中的历史 pending 项已基本被后续 phase2 切片覆盖。
- workspace 共享 Go 规范已经沉淀出可复用的“模块内部最终分层模板”：
  - `runtime` 负责依赖装配
  - `api/http` 或 `transport/http` 只做协议映射
  - `application/commands` 与 `application/queries` 分离
  - `ports` 定义输入/输出依赖接口
  - `infrastructure/*` 负责具体实现
- `ctf` 不应整仓改成 `internal/services/<service>` 的服务优先结构，但每个 `internal/module/<module>` 都可以按这个共享模板做内部物理分层。
- 下一轮后端迁移不该继续沿用这份旧 roadmap 直接追加 pending，而应先基于现状重新盘点新的未完成目标。
