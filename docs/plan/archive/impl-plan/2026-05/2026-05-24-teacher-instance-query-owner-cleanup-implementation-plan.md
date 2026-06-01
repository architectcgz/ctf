> 状态：Current
> 事实源：2026-05-24 前端架构 review、当前 `/teacher/instances` 前后端实现、teacher/platform 实例页现状
> 替代：无

# Teacher Instance Query Owner Cleanup Implementation Plan

## 目标

- 把 teacher / platform 实例目录的筛选、分页和统计 owner 从前端本地数组计算收口到 `/teacher/instances` 查询契约。
- 为实例目录建立单一的 server-owned query contract：`class_name`、`keyword`、`student_no`、`status`、`page`、`page_size`，以及 summary 统计。

## 非目标

- 本轮不改实例管理页的视觉结构、表格字段或销毁交互。
- 本轮不新增 `/platform/instances` 独立后台 API。
- 本轮不重构实例销毁命令链路。

## 输入依据

- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/teacher-instances/model/useTeacherInstances.ts`
- `code/frontend/src/api/teacher/instances.ts`
- `code/backend/internal/module/instance/contracts/teacher_instance.go`
- `code/backend/internal/module/instance/application/queries/instance_service.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/runtime/api/http/teacher_instance_types.go`

## 当前结论

- 现在 `/teacher/instances` 只支持 `class_name`、`keyword`、`student_no`，返回整数组。
- teacher / platform 两个页面都在前端本地做分页；platform 还在本地做状态筛选和 hero 统计。
- 这让 query owner 分裂成“后端做部分过滤 + 页面再做剩余过滤/分页/统计”，扩展到大数据量时会继续恶化。

## 任务切片

### Slice 1：实例查询 contract 升级为分页 + summary

- 目标：
  - `TeacherInstanceListQuery`、`TeacherInstanceFilter` 支持 `status/page/page_size`
  - service 接入 `PaginationConfig`，统一默认页大小与上限
  - 仓储直接返回分页数据与 summary，而不是整表数组
- 预期改动：
  - `docs/plan/impl-plan/2026-05-24-teacher-instance-query-owner-cleanup-implementation-plan.md`
  - `.harness/reuse-decisions/frontend-architecture-review-prompt-and-platform-feature-split.md`
  - `code/backend/internal/module/instance/contracts/teacher_instance.go`
  - `code/backend/internal/module/instance/ports/ports.go`
  - `code/backend/internal/module/instance/application/queries/instance_service.go`
  - `code/backend/internal/app/composition/instance_module.go`
  - `code/backend/internal/module/runtime/infrastructure/repository.go`
  - `code/backend/internal/module/runtime/api/http/teacher_instance_types.go`
  - `code/backend/internal/module/runtime/api/http/teacher_instance_mapper.go`
  - `code/backend/internal/module/runtime/api/http/handler.go`
  - 受影响 Go 测试
- 依赖：
  - 继续复用现有 `visibleInstanceStatus()`、teacher class scope 检查和实例行映射。
  - 不借用其他模块的 `PageResult` 契约类型。
- Review focus：
  - status 过滤是否真正下推到服务端 owner，而不是只在 service 里拿全量结果再 slice。
  - summary 统计是否覆盖 platform / teacher 现有 hero 指标。
  - 默认分页配置是否来自统一 `PaginationConfig`。

### Slice 2：teacher / platform 两个页面切换到同一条 server-owned query

- 目标：
  - `getTeacherInstances` 返回分页结构与 summary
  - teacher / platform 实例页都改成基于后端分页和后端统计
  - 不保留“一个 endpoint，两套前端分页语义”的长期双轨
- 预期改动：
  - `code/frontend/src/api/contracts.ts`
  - `code/frontend/src/api/teacher/instances.ts`
  - `code/frontend/src/api/__tests__/teacher.test.ts`
  - `code/frontend/src/features/teacher-instances/model/useTeacherInstances.ts`
  - `code/frontend/src/features/teacher-instances/model/useTeacherInstanceManagementPage.ts`
  - `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
  - `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
  - `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
- 依赖：
  - 继续复用现有组件、路由跳转和销毁确认逻辑。
  - 不新增并行 `platform instance` API helper。
- Review focus：
  - 两个页面是否都只依赖同一条分页 query owner。
  - 前端是否不再对整表数据做本地 status filter / slice 分页。
  - 异步切页、筛选和销毁后的刷新是否仍有明确 owner。

## 风险

- `expired`、`inactive` 这类状态部分来自显示层派生，若 SQL 过滤和现有 `visibleInstanceStatus()` 规则不一致，会出现前后端状态统计漂移。
- teacher 页面当前依赖“自动搜索 + 本地分页”，切到服务端分页后，测试和页面行为都要一起调整，否则容易留下隐式兼容分支。

## 回退方式

- 如分页 contract 引出明显回归，可回退实例查询 service / repository / frontend helper 的分页改动，恢复整数组响应与本地分页。
- 计划和 reuse decision 文档保留，作为该结构收口尝试的证据。
