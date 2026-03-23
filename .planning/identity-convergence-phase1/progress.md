# Progress

## 2026-03-23

- 创建 `identity` 收敛 Phase 1 计划
- 已确认这是“命名已定、所有权未收拢”的典型剩余迁移项
- 在 `identity` 下新增 `api/http + application + infrastructure`，把原 `adminuser` 物理并入
- 新增 `identity.ProfileService`，把 `auth/profile` 与 `auth/password` 的 owner 从 `auth` 下沉到 `identity`
- `auth.Service` 收缩为注册 / 登录 / 密码校验；CAS provider 改为依赖 `identity` 的用户 contract
- 新增 `composition.IdentityModule`，router 与通知装配改为依赖 `identity`
- 删除 `internal/module/adminuser/*` 与 `internal/module/auth/{errors,repository,repository_test}.go`
- 限核验证通过：`identity`、`auth`、`internal/app` 聚焦用例、`system` 通知集成测试
- 继续完成 `auth` 物理分层收口：
  - `auth/handler` 已迁到 `internal/module/auth/api/http`
  - `auth/service` 与 `auth/cas_provider` 已迁到 `internal/module/auth/application`
  - `auth/token_service` 已迁到 `internal/module/auth/infrastructure`
  - `auth` 根目录已清空，不再承载 concrete 实现或测试文件
  - `composition.IdentityModule`、`composition.AuthModule`、practice flow 集成测试与通知集成测试已切到新目录依赖
- 新一轮限核验证通过：
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/auth/... -count=1`
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestIdentityModuleContractsCompile|TestCompositionModulesExposeContracts|TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes|TestRouterBuildUsesCompositionModules|TestArchitectureRulesRejectConcreteCrossModuleImports|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/ops/api/http -run 'TestHTTP_NotificationsSupportTicketListReadAndWebSocketPush' -count=1`
