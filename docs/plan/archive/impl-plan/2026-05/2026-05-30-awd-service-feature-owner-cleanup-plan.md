> 状态：Current
> 事实源：AwdChallengeLibraryPage feature owner 收口
> 替代：无

# AWD Service Feature Owner Cleanup Plan

## 目标

- 把 `components/platform/awd-service/` 下仅服务 `platform/awd-challenges` 的两个内部子块收口到 `features/platform/awd-challenges/ui/`。
- 同步更新 raw-source 测试和自动组件声明，消除旧目录引用。

## 非目标

- 本轮不重做 AWD 题库页交互、筛选、分页或导入流程。
- 本轮不扩展到其他 AWD 页面或 `components/platform/cheat/*`。
- 本轮不处理通用 `components/common/*` 依赖归属。

## 输入依据

- `code/frontend/src/components/platform/awd-service/AwdChallengeLibrarySection.vue`
- `code/frontend/src/components/platform/awd-service/AwdChallengeWorkspaceHeader.vue`
- `code/frontend/src/features/platform/awd-challenges/ui/AWDChallengeLibraryPage.vue`
- `code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts`
- `code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts`
- `code/frontend/src/pages/__tests__/studentDirectoryTypographyBoundary.test.ts`

## 设计边界

### `features/platform/awd-challenges/ui` 本轮负责

- AWD 题库页 workspace header
- AWD 题库列表 section
- 页面壳内部的组件装配

### 本轮继续不动

- 路由页的数据 owner
- AWD 题目导入 section 的行为
- 通用表格、分页、导航与 modal 模板

## 任务切片

### Slice 1：组件 owner 迁移

- 目标：
  - 把 `AwdChallengeLibrarySection.vue`、`AwdChallengeWorkspaceHeader.vue` 迁到 `features/platform/awd-challenges/ui/`
  - 更新 `AWDChallengeLibraryPage.vue` 为 feature 内部相对依赖
- 验证：
  - `npm run test:run -- src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts`
- Review focus：
  - 迁移后这两个子块是否仍只由 `AWDChallengeLibraryPage.vue` 装配
  - 是否没有把 page owner 或导入流程错误下沉到子块

### Slice 2：测试与声明收口

- 目标：
  - 更新 raw-source 测试到新路径
  - 更新 `components.d.ts`
  - 确认源码里不再引用旧 `components/platform/awd-service/*`
- 验证：
  - `npm run test:run -- src/pages/__tests__/workspaceShellStyles.test.ts src/pages/__tests__/studentDirectoryTypographyBoundary.test.ts src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts`
  - `rg -n "components/platform/awd-service/(AwdChallengeLibrarySection|AwdChallengeWorkspaceHeader)" code/frontend/src`
- Review focus：
  - raw-source 护栏是否继续覆盖新的 feature owner 路径
  - 自动声明是否与实际路径一致

## 验证计划

- `bash scripts/check-task-intake.sh --reuse-decision awd-service-feature-owner-cleanup`
- `cd code/frontend && npm run test:run -- src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts src/pages/__tests__/workspaceShellStyles.test.ts src/pages/__tests__/studentDirectoryTypographyBoundary.test.ts`
- `cd /home/azhi/workspace/projects/.worktrees/ctf-awd-service-feature-owner-cleanup && git diff --check`

## 风险与回退

- 风险：
  - raw-source 测试仍引用旧路径，导致 owner 收口不完整。
  - `components.d.ts` 未同步会造成 IDE / 类型声明漂移。
- 回退：
  - 若迁移后出现引用错误，可直接回退这两个组件文件和 `AWDChallengeLibraryPage.vue` 的 import 调整，不影响其它 AWD feature。
