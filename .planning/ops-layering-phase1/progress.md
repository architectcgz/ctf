# Progress

## 2026-03-23

- 创建 `ops-layering-phase1` 计划
- 完成 `ops` 物理分层：
  - `audit / dashboard / risk / notification` handler 已迁入 `api/http`
  - 对应用例已迁入 `application`
  - 持久化适配器已迁入 `infrastructure`
  - 根包仅保留 `contracts.go`、`module.go`
- 新增根包架构测试，防止后续再次回到大平铺
- 定向验证通过：
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/ops/... -count=1`
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestNewRouterRegistersStudentChallengeRoutes|TestRouterBuildUsesCompositionModules|TestFullRouter_AdminSystemAndNotificationStateMatrix' -count=1`
