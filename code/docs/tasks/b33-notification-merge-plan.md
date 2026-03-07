# B33 分支安全合入方案

> 状态更新（2026-03-07）：
> 该方案已执行完成。主线已补齐 `challenges` 建表脚本，整理 migration 为连续版本链，并将通知能力以追加式 migration 合入到 `000024_create_notifications_table.*`。

## 背景

分支 `feat/b33-websocket-notifications` 相对当前 `main` 有两类内容：

1. **值得保留的能力**
   - WebSocket ticket
   - 通知列表 / 已读 API
   - 通知 WebSocket 推送
   - 通知相关测试

2. **不能直接合入的内容**
   - 历史 migration 编号重排
   - 旧基线上的整分支直接 merge

当前判断结论：

- **不要直接 merge 分支**
- **要把分支拆成独立补丁，在当前 `main` 上重新落地**

## 关键风险

### R1. 整分支 merge 会回退当前主线功能

当前 B33 分支分出时间较早，若直接 merge，会覆盖当前主线已经合入的：

- 教师端查询接口
- 前端多条并行交付结果
- 新增 review / docs

因此只能按补丁级别移植，不能按分支级别合并。

### R2. B33 分支重写了 migration 历史

B33 分支中的 `7687213` 重排了大量 migration 编号，并新增 `000024_create_notifications_table.*`。

当前主线使用 `golang-migrate` 顺序执行 migration：
[Makefile](/home/azhi/workspace/projects/ctf/code/backend/Makefile#L32)

这意味着直接 merge 旧分支不可行，必须先在当前主线上重新整理成唯一版本链，再追加通知表 migration。

### R3. 当前主线本身存在 `challenges` 建表缺口

当前主线 migration 里没有 `000004_create_challenges_table.*`，但后续 migration 已经在直接修改 `challenges` 表：

- [000007_add_flag_prefix_to_challenges.up.sql](/home/azhi/workspace/projects/ctf/code/backend/migrations/000007_add_flag_prefix_to_challenges.up.sql)

这说明 `1fb88f5` 里的 `challenges` 建表迁移不是“通知功能的一部分”，而是一个**应该单独修复的基础 migration 问题**。

## 建议拆分

### Patch A: 修复主线 migration 缺口

来源提交：

- `1fb88f5 fix(迁移): 补齐challenges建表迁移`

目标：

- 在当前主线上补齐 `challenges` 建表 migration

建议做法：

- 先确认是否要直接采用 `000004_create_challenges_table.*`
- 如果已有环境已经跳过该版本，需要额外评估：
  - 是补回历史版本号
  - 还是改成新的“兼容性补表 migration”

建议优先级：

- **高**

原因：

- 这是当前主线 fresh migrate 的基础风险
- 和通知能力无耦合，应该独立处理

### Patch B: WebSocket ticket 能力

来源提交：

- `7687213 feat(后端): 完成B33通知推送与迁移整理`

建议移植文件：

- `code/backend/internal/config/config.go`
- `code/backend/configs/config.yaml`
- `code/backend/internal/dto/auth.go`
- `code/backend/internal/module/auth/token_service.go`
- `code/backend/internal/module/auth/handler.go`
- `code/backend/pkg/errcode/errcode.go`

目标：

- 为登录用户签发一次性 WebSocket ticket
- WebSocket 握手改为 ticket 消费，而不是直接复用 access token

接口结果：

- `POST /api/v1/auth/ws-ticket`

注意事项：

- 当前主线 `tokenService` 已有变更，需在当前版本上手工合并，不要生搬旧文件
- 保持 refresh / revoke 逻辑与当前主线一致

### Patch C: 通知模型、仓储、服务、Handler

来源提交：

- `7687213`

建议移植文件：

- `code/backend/internal/dto/notification.go`
- `code/backend/internal/model/notification.go`
- `code/backend/internal/module/system/notification_repository.go`
- `code/backend/internal/module/system/notification_service.go`
- `code/backend/internal/module/system/notification_handler.go`

目标：

- 提供通知列表查询
- 提供通知已读更新
- 提供通知推送服务能力

接口结果：

- `GET /api/v1/notifications`
- `PUT /api/v1/notifications/:id/read`
- `GET /ws/notifications`

### Patch D: WebSocket manager

来源提交：

- `7687213`

建议移植文件：

- `code/backend/pkg/websocket/manager.go`

目标：

- 提供用户维度的连接管理
- 支持定向推送
- 支持心跳和简单重连建议元数据

注意事项：

- 当前实现是内存态 manager，只适用于单实例
- 如果后续要多副本部署，需要另加跨实例广播方案

### Patch E: 通知表 migration

来源提交：

- `7687213`

原分支文件：

- `code/backend/migrations/000024_create_notifications_table.up.sql`
- `code/backend/migrations/000024_create_notifications_table.down.sql`

当前已执行做法：

- 先将主线重复编号的 migration 整理为唯一版本链，末尾落在 `000023_create_reports_table.*`
- 再追加：
  - `000024_create_notifications_table.up.sql`
  - `000024_create_notifications_table.down.sql`

表结构沿用了 B33 分支版本。

### Patch F: 测试移植

来源提交：

- `7687213`

建议移植文件：

- `code/backend/internal/module/system/notification_http_integration_test.go`
- `code/backend/internal/module/auth/http_integration_test.go`
- `code/backend/internal/module/auth/service_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`

注意事项：

- B33 的通知集成测试目前用的是 `AutoMigrate`：
  [notification_http_integration_test.go](/home/azhi/workspace/projects/ctf-b33-worktree/code/backend/internal/module/system/notification_http_integration_test.go#L213)
- 这能验证接口行为，但**不能验证 migration 链是否能从零跑通**

因此测试补充建议：

- 保留现有 HTTP / WS 集成测试
- 额外补一条 migration smoke test，至少验证通知表 migration 能应用和回滚

## 不建议直接拿的内容

以下内容不应从 B33 分支原样带回主线：

- 历史 migration 编号重排
- 任何会删除当前主线教师模块路由的 router 旧版本
- 任何基于旧前端 / 旧 docs 状态生成的回退变更

## 推荐执行顺序

1. 先处理 `Patch A`，确认 `challenges` 建表缺口怎么修
2. 再做 `Patch B + C + D`
3. 追加 `Patch E`
4. 最后移植 `Patch F`

## 推荐落地方式

不要 merge `feat/b33-websocket-notifications`。

推荐在当前 `main` 上新开修复分支，例如：

`feat/b33-notifications-rebase`

然后按以下方式处理：

1. 手工移植 Go 代码
2. 新建追加式 migration
3. 跑测试与类型检查
4. 单独 review

## 最终建议

这条分支里的**通知能力值得合并**，但应当按“拆补丁重落地”的方式合入。

最合理的执行方式不是：

- `merge feat/b33-websocket-notifications`

而是：

- **在当前主线上重做 B33 通知功能**
- **保留业务能力，放弃旧分支的 migration 重排历史**
