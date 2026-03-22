# Modular Monolith Refactor Checklist

本清单用于在后续 code review 中快速确认后端是否持续遵守第一阶段模块化单体约束。

## Event Rules

- `practice` 的弱一致副作用优先通过事件发布，不在主提交流程里直接串联画像刷新、推荐缓存更新、通知推送。
- 当前已固化的事件名：
  - `practice.flag_accepted`
  - `practice.hint_unlocked`
- 事件消费者失败不得让训练主流程返回失败；最多记录日志并由后续补偿处理。

## Lifecycle Rules

- 后台任务必须从 module builder 注册到 `composition.Root`，禁止继续在 `internal/app/http_server.go` 手写模块内部构造逻辑。
- `HTTPServer` 只能启动/停止 `composition.Root` 暴露的 background jobs，不应直接 new cleaner / updater。
- 需要优雅关闭的异步资源必须在 composition 或 router runtime 中显式暴露关闭入口。

## Dependency Rules

- `internal/module/*` 不应直接跨模块导入其他模块的 concrete subpackage。
- 允许的过渡性例外只有 `teacher -> teaching_readmodel` 兼容包装层，后续收敛完成后应删除。
- 新增跨模块能力时，优先新增 contract、query facade 或事件，而不是直接拿对方 infrastructure/application 实现。

## Review Questions

- 这个改动是否把弱一致副作用重新塞回了主业务路径？
- 这个后台任务是否仍然需要 `http_server.go` 了解模块内部构造？
- 这个跨模块依赖是否可以改成 contract、query 或 event？
- 这个模块是否开始依赖其他模块的 concrete subpackage？
