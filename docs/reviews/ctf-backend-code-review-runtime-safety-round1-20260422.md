# CTF 后端代码 Review（运行时安全性与模块边界专项 第 1 轮）

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | 后端 runtime 失败语义、模块边界、配置安全与端口契约 |
| 轮次 | 第 1 轮 |
| 审查范围 | `code/backend/internal/app`、`code/backend/internal/module/*`、`code/backend/configs/*` |
| 审查日期 | 2026-04-22 |
| 审查方式 | 静态代码审查 + 定向测试验证 |
| 审查状态 | 本轮问题已完成代码修复，并已合并到 `main` |

## 当前结论

- 2026-04-25 复核结论：本轮记录的 H1、H2、M1、M2、L1 均已有对应代码修复。
- 已确认问题主要集中在 4 类：
  - runtime 初始化失败或被关闭时，没有 `fail fast`，而是进入“伪创建成功”的降级路径。
  - CORS 在生产配置下存在“空白名单即允许任意 Origin”的危险语义。
  - ports 层 `context` 语义不一致，调用方无法稳定推导取消、超时和后台任务边界。
  - composition 层仍有反向跨模块依赖，目录分层与真实依赖边界不完全一致。
- 当前文档作为历史 review 归档保留，下面的问题描述仍记录首轮发现时的状态；每项的当前状态以“修复状态”为准。

## 已验证证据

- 已执行：

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
timeout 180s go test ./internal/module/runtime/... ./internal/app/...
```

- 结果：
  - `./internal/module/runtime/...` 通过。
  - `./internal/app/...` 失败。
  - 直接失败点集中在实例启动路径，返回 `code=12006`、`message="容器启动失败"`。

- 2026-04-23 补充验证：
  - `TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge` 在 isolated 运行时可稳定复现 `容器启动失败`。
  - 根因不是业务层新回归，而是 `practice_flow_integration_test` 手工装配了 `runtimeProvisioningService(..., nil, ...)`，绕开了 composition 中已修复的 test runtime engine。
  - 该测试在与其他 app 集成测试同跑时可能被共享端口上的残留监听“偶然救活”，属于测试装配缺陷。

- 失败测试：
  - `TestFullRouter_AuthorizedSmokeMatrix`
  - `TestFullRouter_ContestChallengeAndScoreboardStateMatrix`
  - `TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge`

- 2026-04-25 合并后验证：

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
timeout 300s go test ./... -count=1
```

- 结果：
  - 通过。
  - 本轮修复已通过 merge commit `1f526a31` 合并到 `main`。

## 修复状态

- [H1] 已修复。runtime engine 不可用时不再伪造成功结果；测试环境通过 composition 的 test runtime engine 装配稳定运行，清理链路也改为显式失败语义。
- [H2] 已修复。CORS 空白名单不再允许任意 Origin；`allow_credentials=true` 且 `allow_origins=[]` 会在配置校验阶段失败。
- [M1] 已修复。相关 ports、service、repo 边界默认显式传入 `ctx context.Context`，公开契约不再用 `FooWithContext` 作为规范命名，也不保留无 ctx 兼容包装。
- [M2] 已修复。runtime 代理流量事件记录迁移到 runtime infrastructure，composition 不再直接依赖 contest infrastructure。
- [L1] 已修复。开发基线配置中的 PostgreSQL、Redis 默认密码改为空值，并通过环境变量注入；生产配置中的 `change_me` 仍是部署占位值，部署时必须覆盖。

## 第 2 轮补充复核

- 2026-04-25 复核时发现首轮 review 仍遗漏了 4 个边界问题：
  - `runtime` 内部 provisioning、maintenance、repository 仍有部分 DB 查询和端口操作没有接收 `ctx`。
  - challenge self-check 和 AWD preview 通过 runtime provisioner 自动分配端口时，只读取已占用端口，没有原子写入 `port_allocations`，并发时存在端口碰撞风险。
  - `prod` 配置仍允许 PostgreSQL、Redis 密码保持 `change_me` 或空值，只依赖部署人员手工覆盖。
  - auth token service 与 auditlog context helper 仍有非入口层 `context.Background()` 兜底。
- 第 2 轮修复状态：
  - 已将 runtime provisioning、maintenance、repository 的相关操作改为显式 `ctx` 契约，并让 DB 查询走 `WithContext(ctx)`。
  - 已将未预留端口的 runtime 创建路径改为通过 `ReserveAvailablePort(ctx, ...)` 原子预留，并在创建失败或 cleanup 时释放端口。
  - 已在生产环境配置校验中拒绝空密码和 `change_me` 占位密码。
  - 已移除 auth/auditlog 的静默后台上下文兜底；nil ctx 现在会显式失败或保持 nil。
- 第 2 轮验证：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/backend-runtime-review-round2/code/backend
timeout 300s go test ./... -count=1
```

- 结果：
  - 通过。

## 第 3 轮机械防线

- 2026-04-25 继续补充防回归规则：
  - 新增后端架构测试，统一禁止 ports/contracts 与 infrastructure repository 的公开操作边界遗漏 `ctx context.Context`。
  - 新增后端架构测试，统一禁止新增公开 `FooWithContext(...)` 命名，后端默认使用 `Foo(ctx, ...)`。
  - 新增后端架构测试，统一限制 `context.Background()` / `context.TODO()` 只能出现在明确的 root、基础设施初始化入口、后台任务 root 或 testsupport 中。
- 第 3 轮修复状态：
  - 已将 contest team ports、application service、infrastructure repository 的 team 查询、成员变更、创建/解散链路全部改为显式 ctx。
  - 已同步 challenge infrastructure 旧 `Create` 方法与 practice infrastructure 旧 `CountRecentSubmissions` 方法的 ctx 签名，避免保留无 ctx 可误用边界。
- 第 3 轮验证：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/backend-context-architecture-guards/code/backend
timeout 300s go test ./... -count=1
```

- 结果：
  - 通过。

## 问题清单

### 🔴 高优先级

- [H1] runtime 在不可用时走“伪成功降级”，把基础设施故障扩散成业务 500
  - 文件：
    - `code/backend/internal/app/composition/runtime_module.go`
    - `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
    - `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
    - `code/backend/internal/module/practice/application/commands/service.go`
  - 问题描述：
    - `test` 环境会直接禁用 runtime engine。
    - `ProvisioningService` 在 engine 为 `nil` 时仍生成假的 `container_id`、`network_id`、`access_url` 并返回成功。
    - 上层再继续走访问地址探活，最终以“容器启动失败”结束。
  - 影响范围/风险：
    - 实例创建链路出现非确定性失败。
    - 运行时资源清理也会被降级成“记录日志但什么都不做”。
    - 真正的 runtime 初始化故障不会在启动期暴露，而会延后到业务路径上爆炸。
  - 当前结论：
    - 这是已验证故障，不只是架构味道。

- [H2] CORS 在生产配置下存在高风险默认行为
  - 文件：
    - `code/backend/internal/middleware/cors.go`
    - `code/backend/configs/config.yaml`
    - `code/backend/configs/config.prod.yaml`
    - `code/backend/internal/config/config.go`
  - 问题描述：
    - 中间件把 `allow_origins=[]` 解释成允许任意 Origin。
    - 基线配置同时打开了 `allow_credentials=true`。
    - 生产覆盖配置正好把 `allow_origins` 设为空数组。
  - 影响范围/风险：
    - 生产环境跨站访问面被放大。
    - 配置语义与“白名单为空应拒绝”这一常见预期相反，容易误配。
  - 当前结论：
    - 应收紧为显式白名单，或至少在 `allow_credentials=true` 时禁止空白名单。

### 🟡 中优先级

- [M1] ports 层 `context` 语义不一致，超时与取消只在部分路径生效
  - 文件：
    - `code/backend/internal/module/practice/ports/ports.go`
    - `code/backend/internal/module/challenge/ports/ports.go`
    - `code/backend/internal/module/runtime/ports/http.go`
  - 问题描述：
    - 同一仓储接口中混用带 `ctx` 和不带 `ctx` 的方法。
    - HTTP 层和后台任务层即使传入了取消上下文，也不能保证真正传到 DB/缓存边界。
  - 影响范围/风险：
    - 长查询、后台任务和异步关闭链路的可控性差。
    - 后续统一超时策略会越来越难。

- [M2] composition 层仍有反向跨模块依赖
  - 文件：
    - `code/backend/internal/app/composition/runtime_module.go`
  - 问题描述：
    - `runtime` 通过 `contestinfra.NewAWDRepository(root.DB())` 记录代理流量事件。
    - 这让 runtime 的 HTTP 代理能力直接依赖 contest 的基础设施实现。
  - 影响范围/风险：
    - 模块边界继续漂移。
    - 后续若拆分 contest 持久化模型，runtime 也会被迫联动。
  - 2026-04-23 修复进展：
    - 已将 runtime 代理流量事件记录迁移到 `runtime/infrastructure`，composition 不再直接依赖 `contestinfra`。

### 🟢 低优先级

- [L1] 仓库默认配置仍保留明文开发口令
  - 文件：
    - `code/backend/configs/config.yaml`
  - 问题描述：
    - PostgreSQL、Redis 默认密码直接提交在仓库里。
  - 影响范围/风险：
    - 即使只面向开发环境，也会弱化配置卫生与安全基线。

## 修复顺序建议

1. 先修 [H1]
   - 目标：让 `test` 环境具备可用的 runtime 测试实现，停止走“伪创建成功 + 探活失败”的坏路径。
   - 同时收敛 runtime 缺失时的失败语义，为后续 `fail fast` 做铺垫。
2. 再修 [H2]
   - 目标：明确 CORS 空白名单语义，避免生产误配。
3. 然后处理 [M1]
   - 目标：统一关键 write/read path 的 `context` 契约。
4. 最后处理 [M2]、[L1]
   - 目标：继续收窄跨模块耦合，并整理默认配置卫生。

## 本轮实施约定

- 本轮先开始处理 [H1]。
- 每次提交只覆盖一个可闭环问题，不混入顺手重构。
- 每次修复都补对应回归测试，再跑最小充分验证。
