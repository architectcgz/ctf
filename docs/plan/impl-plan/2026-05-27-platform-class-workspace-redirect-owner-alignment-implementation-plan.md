> 状态：Current
> 事实源：`class-workspace-redirect`、platform / teacher class workspace route view、shared route tests、fronted backlog
> 替代：无

# Platform Class Workspace Redirect Owner Alignment Implementation Plan

## 目标

- 收口 `PlatformClassWorkspaceSection` / `TeacherClassWorkspaceSection` 共用 redirect feature 里的 target route owner，让共享 feature 不再同时决定 teacher / platform 两侧 canonical target。
- 保持 legacy alias route 到 canonical class workspace 的功能和 `panel` query 约定不变。
- 用一个小切片关闭 backlog 里 `PlatformClassWorkspaceSection` 的 redirect 命名残留。

## 非目标

- 本轮不改 `PlatformClassStudents.vue` / `TeacherClassStudents.vue` 的共享 workspace 结构。
- 本轮不处理 `ChallengeWriteupManagePanel`、`@/api/teaching` 更深层依赖、或 `Teacher*` DTO 命名。
- 本轮不改 router path / route name，也不删除 legacy alias route。

## 输入依据

- `code/frontend/src/features/class-workspace-redirect/model/useClassWorkspaceSection.ts`
- `code/frontend/src/views/platform/PlatformClassWorkspaceSection.vue`
- `code/frontend/src/views/teacher/TeacherClassWorkspaceSection.vue`
- `code/frontend/src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts`
- `code/frontend/src/router/__tests__/sharedRoutes.test.ts`
- `code/frontend/src/utils/teachingWorkspaceRouting.ts`
- `docs/plan/impl-plan/2026-05-24-platform-route-owner-decoupling-implementation-plan.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- 当前共享 `useClassWorkspaceSection` 已经把 platform alias route 从 teacher route view 上摘下来，但它仍在同一张 map 里同时声明 teacher / platform 的 canonical target route name。
- 这个实现功能上没问题，但 canonical target owner 仍由共享 feature 内部隐式决定，不够贴合最近几轮“route view 显式声明 owner，shared feature 只做中立 workflow”的收口方向。
- 最小切片是把共享 feature 收口为“根据当前 legacy route 解析 panel，并接收调用方传入的 canonical route name”，让 platform / teacher route view 各自承担最终 target owner 声明。

## 任务切片

### Slice 1：收口共享 redirect feature 的 owner 边界

- 目标：
  - 让 `useClassWorkspaceSection` 只负责 alias route -> panel 的解析，并由调用方传入 canonical target route name。
- 预期改动：
  - `code/frontend/src/features/class-workspace-redirect/model/useClassWorkspaceSection.ts`
- review focus：
  - `panel` 映射不回归
  - redirect 仍只发生在 legacy alias route
  - feature 内部不再硬编码 teacher / platform target route 对

### Slice 2：让 platform / teacher route view 显式声明 canonical target

- 目标：
  - 平台和教师别名页各自只声明自己的 canonical workspace route。
- 预期改动：
  - `code/frontend/src/views/platform/PlatformClassWorkspaceSection.vue`
  - `code/frontend/src/views/teacher/TeacherClassWorkspaceSection.vue`
- review focus：
  - platform route view 不回退到 teacher 语义依赖
  - teacher / platform page template 结构保持轻量

### Slice 3：同步测试与 backlog 事实

- 目标：
  - 更新源码护栏测试，明确共享 feature 现在通过参数接收 canonical target route。
  - 在 backlog 记录这轮 `PlatformClassWorkspaceSection` 残余收口进展。
- 预期改动：
  - `code/frontend/src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts`
  - `code/frontend/src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-platform-class-workspace-redirect-owner-alignment-review.md`

## 验证

- `npm run test:run -- src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts src/router/__tests__/sharedRoutes.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`

## 回退 / 恢复说明

- 本轮只涉及 redirect owner 归位，回退时可以整体撤销 `class-workspace-redirect` 和两个 route view 的参数化改动，不影响 canonical page、本体业务数据或 API contract。

## 残余风险

- `ChallengeWriteupManagePanel`、更深层 `Teacher*` DTO / contract 命名和其他 `@/api/teaching` 依赖仍未覆盖。
- 这轮不改变路由名字本身，因此 teacher / platform 双命名空间仍然存在；收口的是 redirect owner 边界，不是全量命名迁移。
