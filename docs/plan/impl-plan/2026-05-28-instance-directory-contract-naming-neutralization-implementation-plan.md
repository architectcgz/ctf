> 状态：Current
> 事实源：`api/contracts.ts`、instance API client、teacher / platform instance workspace、frontend backlog
> 替代：无

# Instance Directory Contract Naming Neutralization Implementation Plan

## 目标

- 把共享实例目录 DTO `TeacherInstanceItem` 收口成中性命名 `InstanceDirectoryItem`。
- 同步把共享分页 / 汇总 / 状态筛选类型收口成 `InstanceDirectoryPageData`、`InstanceDirectorySummaryData`、`InstanceDirectoryStatusFilter`。
- 保持 teacher / platform 实例目录、分页、筛选和销毁行为不变，只收 contract 命名语义。

## 非目标

- 本轮不改 `getTeacherInstances()`、`destroyTeacherInstance()`、`getPlatformInstances()` 这类 public API function owner。
- 本轮不改 `/api/v1/teacher/instances` 的 HTTP path、请求参数、分页 contract 或响应字段。
- 本轮不改教师概览、学生目录、班级洞察或其它仍带 teacher 前缀的非实例 DTO。

## 输入依据

- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/teaching/instances.ts`
- `code/frontend/src/api/teacher/instances.ts`
- `code/frontend/src/api/admin/teaching.ts`
- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/teacher-instances/model/useInstances.ts`
- `code/frontend/src/features/teacher-instances/ui/TeacherInstanceManagementPage.vue`
- `code/frontend/src/components/teacher/instance-management/TeacherInstanceDirectorySection.vue`
- `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
- `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/api/__tests__/teacher.test.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `TeacherInstance*` 这组类型当前同时被 teacher / platform 目录能力消费，已经不是教师专属 contract。
- 这组类型只表达实例目录项、分页汇总和目录筛选语义，适合作为一刀独立的命名收口切片。
- 最小安全切片是只收 `TeacherInstance* -> InstanceDirectory*`，同步 API client、feature、组件、测试与事实文档；public wrapper 函数名继续保持现状。

## 任务切片

### Slice 1：收口共享 instance directory contract 名称

- 目标：
  - 在 `api/contracts.ts` 提供中性 `InstanceDirectory*` 类型，实例共享 API client 切到新命名。
- 预期改动：
  - `code/frontend/src/api/contracts.ts`
  - `code/frontend/src/api/teaching/instances.ts`
  - `code/frontend/src/api/teacher/instances.ts`
  - `code/frontend/src/api/admin/teaching.ts`
- review focus：
  - `getTeacherInstances()`、`destroyTeacherInstance()` 与 admin alias 的行为、path 和返回结构保持不变

### Slice 2：同步 teacher / platform 实例目录消费面

- 目标：
  - 让 teacher / platform feature、teacher page shell 和实例目录 section 全部切到 `InstanceDirectory*`。
- 预期改动：
  - `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
  - `code/frontend/src/features/teacher-instances/model/useInstances.ts`
  - `code/frontend/src/features/teacher-instances/ui/TeacherInstanceManagementPage.vue`
  - `code/frontend/src/components/teacher/instance-management/TeacherInstanceDirectorySection.vue`
- review focus：
  - 实例列表加载、筛选、防重销毁和分页 owner 不变

### Slice 3：同步测试、contract 文档与 backlog 证据

- 目标：
  - 用 API / page raw-source 测试锁住中性命名，更新 contract 文档与 backlog 当前进展。
- 预期改动：
  - `code/frontend/src/api/__tests__/teacher.test.ts`
  - `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
  - `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
  - `docs/contracts/api-contract-v1.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-28-instance-directory-contract-naming-neutralization-review.md`

## 验证

- `npm run test:run -- src/api/__tests__/teacher.test.ts src/views/platform/__tests__/InstanceManage.test.ts src/views/teacher/__tests__/InstanceManagement.test.ts`
- `npm run typecheck`
- `git diff --check`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

## 回退 / 恢复说明

- 本轮回退按命名切：整体撤销 `InstanceDirectory*` 及其消费面引用更新即可，不涉及运行时 API owner 或行为回退。

## 残余风险

- 这轮只处理实例目录 contract 的中性化命名，不会顺手覆盖其它仍带 teacher 前缀的共享 DTO。
