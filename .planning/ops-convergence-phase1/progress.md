# Progress

## 2026-03-23

- 创建 `ops` 收敛 Phase 1 计划
- 已确认 `system` 仍是高耦合 owner，需要分阶段拆解
- 完成 `audit / dashboard / risk -> ops`：
  - `ops` 新增审计、仪表盘、风控实现与对应测试
  - `composition.SystemModule` 已改为暴露 `ops` 的 audit/dashboard/risk contracts
  - admin 审计、仪表盘、风控路由继续保持原路径，但 handler owner 已切到 `ops`
  - `system` 中对应的 audit/dashboard/risk owner 逻辑已删除
- `notification` 继续保留在 `system`：
  - `/api/v1/notifications`
  - `/ws/notifications`
  - 相关 websocket / ticket / 事件消费链路待后续 phase 单独迁移
- runtime 指标桥接已改为 nil-safe 装配：
  - 无 container engine 时不再强行构造 stats service
  - `ops` 仍通过 stats provider bridge 获取容器指标
- 补充 `risk` 聚合回归测试：
  - 覆盖 submit burst 的阈值、排序与截断
  - 覆盖 shared IP 的去重、排序与用户名收敛
- 定向验证完成：
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/ops ./internal/module/system -count=1`
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestNewRouterRegistersStudentChallengeRoutes|TestRouterBuildUsesCompositionModules|TestFullRouter_AdminSystemAndNotificationStateMatrix' -count=1`
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestArchitectureRulesRejectConcreteCrossModuleImports' -count=1`
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/auth -count=1`
