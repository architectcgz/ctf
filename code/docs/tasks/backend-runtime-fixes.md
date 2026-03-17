# Backend Runtime Fixes

## Runtime Ownership

- `internal/module/container` 统一负责 Docker 运行时操作，包括镜像信息查询、镜像删除、受管容器资源采集和实例运行时清理。
- `internal/module/system/dashboard_service.go` 只消费 container 模块暴露的受管容器指标，不再直接访问 Docker client。
- Flag 全局密钥统一收敛到 `container.flag_global_secret`，推荐通过 `CTF_CONTAINER_FLAG_GLOBAL_SECRET` 注入。

## Instance Lifecycle

- 练习态和竞赛态实例创建会在事务内完成作用域检查、端口预留和实例记录落库。
- 受管端口由 `instances.host_port` 与 `port_allocations` 协同管理，运行时创建失败会回写失败状态并释放端口。
- 实例进入 `failed`、`stopped`、`expired` 终态时，会在同一数据库事务内更新状态并删除对应 `port_allocations`，避免端口先释放但实例仍停留在活动态。
- 过期实例如果运行时资源清理失败，不会提前改成 `expired` 或释放端口，而是保留活动态等待下一轮清理重试。
- 在线用户统计基于 refresh session 键 `ctf:token:{user_id}`，登录时写入，登出时按 refresh token JTI 清理。
- Refresh 接口会校验 refresh session 键中的最新 JTI，旧登录态的 refresh token 在重复登录或登出后不能继续换发 access token。

## Scheduler Coordination

- 容器过期清理任务使用 `container.cleanup_lock_ttl` 控制跨实例锁。
- 竞赛状态自动流转使用 `contest.status_update_lock_ttl` 控制跨实例锁。
- AWD 轮次调度使用 `contest.awd.scheduler_lock_ttl` 控制跨实例锁，轮次级别仍保留 `contest.awd.round_lock_ttl`。

## Router Layout

- `internal/app/router.go` 负责依赖装配和路由分组初始化。
- 具体管理员路由和用户侧路由注册已拆分到 `internal/app/router_routes.go`，降低单文件复杂度并保持原有行为。

## Auth And Test Stability

- JWT 解析增加 1 秒时钟偏移容忍，避免 access token 在签发后立即校验时被 `nbf/iat` 边界误判为无效。
- `TestFullRouter_AuthorizedSmokeMatrix` 改为逐路由子测试，避免在单个父测试里长期堆积 router、sqlite 和 miniredis 资源。
- smoke 测试对异步班级报告导出路由会等待报告进入 `ready`，避免后台任务在测试资源释放后继续访问已关闭数据库。

## Request Context Propagation

- `internal/module/practice` 为启动实例、提交 Flag、查询实例和读取进度新增带 `ctx` 的服务入口，HTTP handler 不再在入口层把请求上下文替换成 `context.Background()`。
- 这样请求超时、客户端取消和上游链路取消可以继续传递到实例创建、Redis 限流、缓存读写和实例可见性查询，避免后台继续执行已无效请求。
- `internal/module/container` 为创建实例、销毁实例、延长实例、读取访问地址和查询实例列表新增带 `ctx` 的 service / repository 入口，实例相关 HTTP 路径不再在服务入口丢失请求上下文。
- 这样实例访问校验、可见性查询、延时更新和实例记录写入可以遵守请求级取消与超时控制，降低慢请求堆积风险。
- `internal/module/challenge` 为公开靶场列表、靶场详情及其 solved-count 缓存补齐带 `ctx` 的 service / repository 入口，学员查询链路不再在缓存读写和统计查询时退回 `context.Background()`。
- 这样公开靶场查询在请求取消后可以及时停止 Redis/GORM 访问，避免已取消请求继续占用缓存和数据库连接。
- `internal/module/assessment/recommendation_service` 为学员推荐接口和教师侧推荐读取补齐带 `ctx` 的 service 入口，并把 Redis 推荐缓存、已解题查询和推荐题目筛选统一绑定到请求上下文。
- 这样学员推荐和教师查看学生推荐时，取消请求会及时停止缓存和数据库访问，不会在内部继续跑完整个推荐流程。
- `internal/module/auth` 为登录、注册和 CAS 登录发 token 链路补齐带 `ctx` 的 token service 入口，refresh session 写入不再在发 token 时退回 `context.Background()`。
- 这样登录请求一旦超时或被客户端取消，refresh session 键不会继续在后台写入，认证链路的缓存副作用和请求生命周期保持一致。
- `internal/module/practice/score_service` 为计分更新、个人得分读取和排行榜读取补齐带 `ctx` 的入口；提交正确 Flag 后的异步计分改成脱离请求但带超时的后台任务。
- 这样计分缓存和数据库访问既能响应显式取消，也不会在异步分数刷新时无界悬挂，降低慢 Redis 或慢 SQL 对提交流程的拖尾影响。
- `internal/module/assessment` 为个人画像查询、画像重建和增量更新补齐 repository 级 `ctx` 传递，取消/超时不再在拿锁失败时被误判成 Redis 故障降级为旧画像。
- 这样学员查看画像、教师查看学员画像以及后台增量更新在请求取消或超时后都能及时停止 SQL/Redis 访问，避免“已取消请求却返回旧数据”的假成功结果。
- `internal/module/challenge/image_service` 为创建镜像时的 Docker 镜像探测补齐带 `ctx` 的 service 入口，管理员发起镜像录入时不再在同步校验阶段退回 `context.Background()`。
- 这样镜像探测在请求取消或超时后会立即停止，不会继续占用 Docker client 调用和请求线程。
- `internal/module/container.Cleaner` 与 `internal/module/assessment.Cleaner` 现在都持有 service 级可取消 context，并在 `Stop(ctx)` 时取消正在执行的清理/重建任务，同时等待 cron 中已启动 job 退出。
- `internal/app/http_server.go` 在关服时会先取消 contest updater，再等待 cleaner 和 updater 退出；构造失败时也会清理已启动的 cleaner，避免初始化半失败后残留后台任务。
- `internal/bootstrap/run.go` 在 HTTP 服务优雅关闭后会显式关闭 Postgres `sql.DB` 和 Redis client；HTTP 服务初始化失败时也会先释放已打开的连接资源。
- 这样进程退出不再完全依赖操作系统回收连接，开发/测试环境下重复启动停止时连接生命周期也更清晰可控。
- `internal/module/container/service.go` 为实例运行时清理、容器删除、网络删除和 ACL 清理补齐带 `ctx` 的入口，请求销毁、教师销毁、过期实例清理和孤儿容器清理不再统一退回 `context.Background()`。
- 同时拓扑创建失败后的回滚清理改为走 service 级删除包装，统一复用 10 秒运行时超时控制，避免直接用无界 `Background` 调 Docker 清理。
- `internal/module/assessment/report_service.go` 的同步个人报告生成现在会把请求 `ctx` 继续传到能力画像查询，不再在报告聚合阶段丢掉请求级取消和超时控制。
- `internal/app/router.go` 现在会把 `ReportService` 保留在 app 生命周期里，`internal/app/http_server.go` 在关服时会显式调用 `reportService.Close(ctx)`，等待异步班级报告任务退出。
- 这样 `ReportService` 之前已经实现的任务取消和 wait group 不再只是测试内生效，应用真实关服时也会收敛后台报告任务生命周期。
- `internal/module/assessment/report_service.go` 里的个人报告同步导出现在真正绑定 `report.personal_timeout`，报告聚合和落库会遵守配置化超时，而不是仅仅存在配置字段但没有生效。

## Async Task Lifecycle

- `internal/module/assessment/report_service.go` 移除了 `CreateClassReport -> go runAsyncReport -> go worker` 的双层 goroutine 包装，避免 worker pool 前面积压无限 goroutine。
- `ReportService` 现在持有 service 级 base context、任务 wait group 和 `Close(ctx)`，异步班级报告任务会绑定到 service 生命周期，在关闭时主动取消并等待任务退出。
- 失败回写改为带 5 秒超时的独立 context，避免失败标记本身无界阻塞。
