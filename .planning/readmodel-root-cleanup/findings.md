# Findings

- `practice_readmodel/module.go` 只是 `QueryService` 的转发层。
- `teaching_readmodel/module.go` 当前只暴露 `GetClassSummary`，与 `api/http` 直接依赖 `application.QueryService` 的方式不一致。
- 这两个模块都还留着根包壳，说明 readmodel 的最终边界尚未完全稳定。
- `practice_readmodel` 适合直接让 `application.QueryService` 实现 `PracticeQuery`，不需要额外 façade。
- `teaching_readmodel` 需要先把 `TeachingQuery` 扩成 handler/composition 真正消费的接口集合，再删除根壳。
