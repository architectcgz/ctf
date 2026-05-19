# backend shared pkg 边界清理实施计划

## Objective

把当前仍停留在 `code/backend/pkg/` 与历史共享桶中的后端能力按 owner 收回正确边界，先完成已经落地的第一刀，再继续清理剩余 HTTP / websocket / crypto / errcode 残留：

- 单模块私有缓存键下沉到 owning module
- 共享基础设施能力迁回 `internal/infrastructure`
- `pkg/*` 不再继续承载只服务于 `internal` 的实现

## Non-goals

- 不改变任何现有 HTTP JSON 契约、路由或错误码语义
- 不在本刀里改数据库 schema、Redis 数据格式或限流策略
- 不把真正跨进程、跨边界的公共协议强行塞回某个业务模块

## Inputs

- `code/backend/internal/module/challenge/infrastructure/solved_count_cache.go`
- `code/backend/internal/app/router.go`
- `code/backend/internal/bootstrap/run.go`
- `code/backend/internal/middleware/ratelimit.go`
- `code/backend/internal/middleware/ratelimit_test.go`
- `code/backend/pkg/{cache,logger,ratelimit,response,websocket,crypto,errcode}`
- `docs/tasks/backend-task-breakdown.md`
- `.harness/reuse-decisions/backend-shared-pkg-boundary-cleanup-phase1.md`

## Ownership Evaluation

- `challenge` solved-count key 只被 `challenge` 基础设施使用，应下沉到 `challenge/infrastructure/cachekeys`
- `logger` 与 `ratelimit` 都只服务后端内部 wiring / middleware，不应继续暴露在仓库级 `pkg`
- `response` 属于 HTTP transport 层共享能力，后续应收回 `internal` 内部传输边界
- `websocket` 若仅供后端通知与连接管理使用，应迁到共享基础设施层
- `crypto`、`errcode` 需要继续区分“共享内核”与“传输层 / 领域层” owner，避免为了删目录把稳定契约放错层

## Task slices

1. 收口 `pkg/cache`、`pkg/logger`、`pkg/ratelimit`：
   - `challenge` 缓存键下沉到模块内
   - `logger`、`ratelimit` 迁到 `internal/infrastructure`
   - 清理调用点 import
2. 清理 `pkg/response`：
   - 识别统一 HTTP response helper 的稳定 owner
   - 收回 `internal` 内部 transport/shared HTTP 层
   - 更新架构守卫与调用点
3. 清理 `pkg/websocket`：
   - 确认是否为通知 / 连接管理共享基础设施
   - 迁到 `internal/infrastructure` 或更明确 owner
4. 处理 `pkg/crypto`、`pkg/errcode`：
   - 先做 owner 分析
   - 再按最小可审阅切片迁移，避免一次混做多类 contract
5. 收口 `pkg/errcode`：
   - 共享 `AppError` 类型与平台级公共错误收回 `internal/apperror`
   - challenge / contest / instance / ops / auth 的公开错误按 owner 收回模块 `contracts`
   - `httpresponse` 统一消费 HTTP status，application / domain 不再读取 `HTTPStatus`

## Validation

- `cd code/backend && go test ./internal/module/challenge/... -count=1`
- `cd code/backend && go test ./internal/bootstrap -count=1`
- `cd code/backend && go test ./internal/app ./internal/middleware -count=1`
- `cd code/backend && go test ./internal/module -count=1`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-reuse-first.sh`

## Review focus

- `pkg/*` 是否只是换路径，还是 owner 真的已经明确
- 单模块私有常量是否已离开共享桶
- `internal/infrastructure` 是否只接收真正跨模块但不跨仓库边界的实现
- 是否仍有调用点留在旧 import 路径

## Rollback

本计划不涉及 schema 变更。若某一刀出现回归，可按提交粒度回退对应包迁移，不影响数据库或外部 API 兼容性。
