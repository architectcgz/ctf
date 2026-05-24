# Admin Teaching Query Owner Decoupling Implementation Plan

## 目标

- 把 platform 班级、学生、实例管理页面使用的共享教学查询 helper 从 `@/api/teacher` 收口到中性的 `@/api/teaching` owner。
- 保持请求路径、DTO 形状和页面行为不变，只调整前端模块边界。

## 非目标

- 本轮不改后端接口路径，不新增 `/platform/*` API。
- 本轮不处理 AWD 复盘、学生分析、复盘归档等更深的共享 workflow。
- 本轮不调整页面视觉结构或交互。

## 输入依据

- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `code/frontend/src/api/teacher/classes.ts`
- `code/frontend/src/api/teacher/instances.ts`
- `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`

## 任务切片

### Slice 1：建立中性 teaching query owner

- 目标：
  - 新增 `@/api/teaching` 入口，承接 platform/teacher 共用的 class 与 instance 查询实现。
  - `@/api/teacher` 对应模块退成兼容 re-export。
- 预期改动：
  - `code/frontend/src/api/teaching.ts`
  - `code/frontend/src/api/teaching/index.ts`
  - `code/frontend/src/api/teaching/classes.ts`
  - `code/frontend/src/api/teaching/instances.ts`
  - `code/frontend/src/api/teacher/classes.ts`
  - `code/frontend/src/api/teacher/instances.ts`
- Review focus：
  - 是否只是 owner 迁移，没有引入新的并行 wrapper 语义。
  - teacher 现有调用方是否仍能通过兼容 re-export 正常工作。

### Slice 2：platform 页面切到中性 owner

- 目标：
  - platform 班级、学生、实例管理页面不再直接依赖 `@/api/teacher`。
  - 对应测试改为校验新 owner。
- 预期改动：
  - `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
  - `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
  - `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
  - `code/frontend/src/views/platform/__tests__/ClassManage.test.ts`
  - `code/frontend/src/views/platform/__tests__/StudentManage.test.ts`
  - `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
- Review focus：
  - 平台 feature 是否只改 import owner，没有行为回归。
  - 测试是否覆盖新 owner 调用和关键筛选/分页/销毁流程。

## 验证

- `npm run test:run -- src/views/platform/__tests__/ClassManage.test.ts src/views/platform/__tests__/StudentManage.test.ts src/views/platform/__tests__/InstanceManage.test.ts src/api/__tests__/teacher.test.ts`
- `npm run typecheck`
- `bash scripts/check-consistency.sh`
- `git diff --check -- <touched files>`

## 风险

- `@/api/teacher` 兼容 re-export 如果导出遗漏，会影响 teacher 侧已有调用。
- 平台测试里原先 mock 的模块路径需要同步更新，否则会出现假绿或假红。

## 回退方式

- 如中性 owner 迁移引出回归，可回退 `@/api/teaching` 新入口和 platform feature import 改动，恢复为 `@/api/teacher` 直接调用。
