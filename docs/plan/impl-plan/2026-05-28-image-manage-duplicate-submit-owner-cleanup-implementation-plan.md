> 状态：Current
> 事实源：`useImageManageMutations.ts`、`ImageManage.vue`、`ImageDirectoryPanel.vue`、相关测试与 backlog
> 替代：无

# Image Manage Duplicate Submit Owner Cleanup Implementation Plan

## 目标

- 收口图片管理页仍残留的重复动作缺口
- 让删除镜像动作像创建镜像动作一样，在本地 owner 上显式持有 in-flight guard
- 保持当前页面结构与镜像目录交互不变，只修正重复删除与重复确认问题

## 非目标

- 本轮不重做图片管理页 UI 结构
- 本轮不迁移图片管理组件路径
- 本轮不调整镜像创建 / 删除 API 契约
- 本轮不触碰其他页面的重复动作 guard

## 输入依据

- `code/frontend/src/views/platform/ImageManage.vue`
- `code/frontend/src/features/image-management/model/useImageManagePage.ts`
- `code/frontend/src/features/image-management/model/useImageManageMutations.ts`
- `code/frontend/src/components/platform/images/ImageDirectoryPanel.vue`
- `code/frontend/src/views/platform/__tests__/ImageManage.test.ts`
- `code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- 图片创建动作已经在 `useImageManageMutations.ts` 中通过 `creating.value` 短路重复提交。
- 图片删除动作当前没有对应的本地 in-flight guard；快速连点时，可能重复进入确认流程或重复发删除请求。
- 正确 owner 应该仍然是 `useImageManageMutations.ts`，而不是 `ImageDirectoryPanel.vue` 或确认框 composable。

## 设计边界

### `useImageManageMutations.ts` 本轮负责

- 为删除动作维护 per-image in-flight guard
- 对外暴露删除中的查询能力
- 继续负责删除成功 / 失败后的 toast 与 refresh

### `useImageManagePage.ts` 本轮负责

- 继续作为页面 model 聚合 mutations 状态
- 向视图层传出删除中的状态查询

### `ImageDirectoryPanel.vue` 本轮负责

- 消费删除中的显式状态
- 在按钮层表达 disabled / loading 文案
- 不自行决定删除流程或重复动作策略

## 任务切片

### Slice 1：收口删除动作 owner

- 目标：
  - 给 `useImageManageMutations.ts` 增加 per-image delete guard
  - 让 guard 覆盖确认弹窗与实际删除请求两个阶段
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/__tests__/duplicateActionGuardAudit.test.ts src/views/platform/__tests__/ImageManage.test.ts`
- Review focus：
  - 删除动作是否仍由单一 owner 承担
  - 是否避免把 guard 分散到按钮态或确认框实现

### Slice 2：视图层消费删除中状态

- 目标：
  - `useImageManagePage.ts` 透出删除中的查询能力
  - `ImageManage.vue` / `ImageDirectoryPanel.vue` 消费该状态并禁用对应删除按钮
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ImageManage.test.ts`
- Review focus：
  - 列表按钮是否只表达状态，不重新持有删除 owner
  - 删除中的按钮态是否与当前页面交互一致

### Slice 3：backlog / review 收尾

- 目标：
  - 更新 backlog 中图片管理页重复提交条目
  - 归档 review 记录
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/__tests__/duplicateActionGuardAudit.test.ts src/views/platform/__tests__/ImageManage.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只收口图片管理页；其他页面如果也有删除类重复动作缺口，需要单独继续排查。
- 如果后续产品希望删除按钮显示更明确的 loading 图标或行级 skeleton，那是 UI 优化，不是这次 owner 修复的范围。
