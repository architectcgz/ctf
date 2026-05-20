# Practice Progress Cache Eventization Phase 3 Slice 4 Implementation Plan

## Objective

继续收口 phase 3 剩余的缓存副作用，让 `practice.flag_accepted` 成为“练习正确提交 / 人工评审通过”后的单一事实事件入口：

- 去掉 `practice/application/commands` 在写路径里对用户进度缓存的同步删除
- 让 `practice/application/queries.ProgressTimelineService` 作为查询 owner 订阅 `practice.flag_accepted` 并失效自己的 progress cache
- 保持正确提交、人工评审通过、得分刷新和事件 payload 的外部行为不变

## Non-goals

- 不删除 `practice` command service 对 Redis 的其余依赖，例如限流和得分缓存
- 不引入新的 MQ、outbox 或新的 practice 事件类型
- 不修改 `assessment` / `ops` 已经存在的事件消费者 contract
- 不进入 phase 5，不处理 Redis concrete allowlist 的其他链路

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/practice/application/commands/{service.go,submission_service.go,manual_review_service.go}`
- `code/backend/internal/module/practice/application/queries/progress_timeline_service.go`
- `code/backend/internal/module/practice/infrastructure/progress_cache.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/module/practice/contracts/events.go`
- `code/backend/internal/module/practice/ports/ports.go`

## Current Baseline

- `practice.flag_accepted` 已经被 `assessment` 和 `ops` 作为消费者使用。
- `SubmitFlag` 与 `ReviewManualReviewSubmission` 仍在写路径里直接执行 `redis.Del(constants.UserProgressKey(...))`。
- `ProgressTimelineService` 持有 `PracticeUserProgressCache`，但只负责读/写缓存，还没有承接失效事件。

## Chosen Direction

1. 继续复用现有 `practice.flag_accepted` 事件，不新增并行事件。
2. 在 `PracticeUserProgressCache` 上补齐显式失效能力，由模块内 infrastructure Redis adapter 实现。
3. 由 `ProgressTimelineService` 注册 practice 事件消费者，并在收到 `FlagAcceptedEvent` 后删除对应用户 progress cache。
4. `practice/runtime.Module` 在 wiring 边缘注册这个查询侧消费者。
5. `submission_service.go` 与 `manual_review_service.go` 只保留事件发布，不再同步执行缓存删除。

## Ownership Boundary

- `practice/application/commands`
  - 负责：提交判定、人工评审通过、发布 `practice.flag_accepted`、触发得分刷新
  - 不负责：直接操作 progress cache key
- `practice/application/queries.ProgressTimelineService`
  - 负责：progress query 的缓存读写与事件驱动失效
  - 不负责：定义提交成功事实或补发业务事件
- `practice/infrastructure.ProgressCache`
  - 负责：封装 progress cache 的 Redis key 与删除实现
  - 不负责：决定何时触发失效

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-12-practice-progress-cache-eventization-phase3-slice4-implementation-plan.md`
- Add: `.harness/reuse-decisions/practice-progress-cache-eventization-phase3-slice4.md`
- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Modify: `code/backend/internal/module/practice/ports/progress_timeline_context_contract_test.go`
- Modify: `code/backend/internal/module/practice/infrastructure/progress_cache.go`
- Modify: `code/backend/internal/module/practice/application/queries/progress_timeline_service.go`
- Modify: `code/backend/internal/module/practice/application/queries/progress_timeline_context_test.go`
- Modify: `code/backend/internal/module/practice/runtime/module.go`
- Modify: `code/backend/internal/module/practice/application/commands/submission_service.go`
- Modify: `code/backend/internal/module/practice/application/commands/manual_review_service.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task Slices

### Slice 1: 给 progress cache 补显式失效 port

目标：

- `PracticeUserProgressCache` 增加删除能力
- `ProgressCache` 实现 Redis 删除
- contract test 覆盖新的 context-aware 方法

Validation:

- `cd code/backend && go test ./internal/module/practice/ports ./internal/module/practice/infrastructure -count=1 -timeout 120s`

Review focus:

- cache port 是否仍然保持最小，只表达 progress cache owner 真正需要的能力
- application 是否没有回退成自己拼 Redis key

### Slice 2: 把缓存失效移到 query owner 事件消费者

目标：

- `ProgressTimelineService` 注册 `practice.flag_accepted` 消费者
- runtime wiring 注册该消费者
- 写路径删除同步 `redis.Del(...)`

Validation:

- `cd code/backend && go test ./internal/module/practice/application/... ./internal/module/practice/runtime/... -count=1 -timeout 120s`

Review focus:

- 提交成功与人工评审通过后是否仍然发布同一条事实事件
- progress cache 删除是否只经查询 owner 执行，而不是换个地方继续散落

### Slice 3: 回收 phase 3 当前事实

目标：

- design / architecture 文档明确 phase 3 的缓存副作用已并入事件化链路
- 不把与 phase 5 相关的 Redis concrete debt 混写成 phase 3 范围

Validation:

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

Review focus:

- 文档是否准确区分“事件化副作用已收口”和“phase 5 仍有 Redis 下沉尾项”

## Risks

- 如果 runtime 忘记注册 progress cache 消费者，用户提交成功后会读到旧 progress cache
- 如果事件 payload 解析不严谨，非 `FlagAcceptedEvent` 可能导致错误失效
- 如果写路径删缓存和事件消费者并存，phase 3 的 owner 仍会继续模糊

## Verification Plan

1. `cd code/backend && go test ./internal/module/practice/application/... ./internal/module/practice/runtime/... ./internal/module/practice/ports ./internal/module/practice/infrastructure -count=1 -timeout 120s`
2. `python3 scripts/check-docs-consistency.py`
3. `bash scripts/check-consistency.sh`

## Architecture-Fit Evaluation

- owner 明确：写路径只发事实事件，查询 owner 处理自己的 cache 副作用
- reuse point 明确：复用现有 `practice.flag_accepted` 事件链与 progress cache adapter，不新增桥接抽象
- 这刀同时解决行为与结构：不是只“还能删掉缓存”，而是让 phase 3 剩余缓存副作用与画像、推荐、通知链保持一致表达
