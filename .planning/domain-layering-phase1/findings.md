# Findings

- `challenge` 根目录 28 个文件，`contest` 根目录 32 个文件，已经明显超过“可单次可靠理解”的范围。
- composition 当前仍直接暴露了多个 concrete repository/service，说明模块边界还没有真正封住。
- `challenge/contracts` 已证明“先抽窄 contract 再迁移 owner”在这个仓库里可行。
- `practice` 的 read side 已迁到 `practice_readmodel`，所以后续应重点防止读逻辑再次回流。
- 参考 workspace 共享 Go 规范后，`ctf` 的模块内部分层建议优先对齐为：
  - `api/http`
  - `application/commands`
  - `application/queries`
  - `domain`
  - `ports`
  - `infrastructure/<adapter>`
  - 必要时单独保留 `runtime` 装配层
- 其中 `ports` 应由应用层消费方定义，`infrastructure` 只实现接口，`handler` 不直接依赖 repository 或 concrete service。
