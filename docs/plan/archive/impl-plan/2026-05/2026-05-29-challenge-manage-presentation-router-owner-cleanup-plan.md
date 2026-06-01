> 状态：Current
> 事实源：平台题目管理 page / presentation owner、前端架构 allowlist、ChallengeManage 测试
> 替代：无

# Challenge Manage Presentation Router Owner Cleanup Plan

## 目标

- 把 `useChallengeManagePresentation.ts` 从 route-aware owner 收口回纯 presentation / action menu owner。
- 让 `useChallengeManagePage.ts` 保留唯一 router owner。
- 删除对应 `featureRouterImportAllowlist` 条目。

## 非目标

- 不重构平台题目管理页的数据流或 API owner。
- 不重写目录排序、筛选或发布检查流程。
- 不处理 `featureRouterImportAllowlist` 其它页面 owner。

## 输入依据

- `code/frontend/src/features/platform-challenges/model/useChallengeManagePresentation.ts`
- `code/frontend/src/features/platform-challenges/model/useChallengeManagePage.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/platform/__tests__/ChallengeManage.test.ts`

## 当前结论

- `useChallengeManagePresentation.ts` 当前只需要“打开详情 / 拓扑 / 题解 / 导入预览”这些动作，不需要直接知道 `Router`。
- `useChallengeManagePage.ts` 已经天然是 page owner，路由动作回到这里更合理。

## 设计边界

### `useChallengeManagePage.ts` 本轮负责

- `useRouter()` 获取
- 导航到导入列表、导入预览、题目详情、拓扑、题解

### `useChallengeManagePresentation.ts` 本轮负责

- 菜单开关
- 状态 / 颜色 / 时间格式化
- 在执行 action 前统一关闭菜单
- 调用外部注入的导航 / publish / remove 动作

## 任务切片

### Slice 1：presentation 改成 callback owner

- 目标：
  - 从 `useChallengeManagePresentation.ts` 移除 `Router` 依赖
  - 将导航动作改为 callback 注入
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeManage.test.ts`
- Review focus：
  - presentation 是否已经不再 import `vue-router`
  - 菜单关闭时机是否保持不变

### Slice 2：allowlist / 护栏 / backlog 收尾

- 目标：
  - 删除 allowlist 条目
  - 更新 raw-source 护栏与 backlog 记录
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/ChallengeManage.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - allowlist 是否真实下降
  - route owner 是否只回到 page，而没有漂到别的 presentation/helper

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/ChallengeManage.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/challenge-manage-presentation-router-owner-cleanup.md docs/plan/impl-plan/2026-05-29-challenge-manage-presentation-router-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-challenge-manage-presentation-router-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-challenges/model/useChallengeManagePresentation.ts code/frontend/src/features/platform-challenges/model/useChallengeManagePage.ts code/frontend/src/views/platform/__tests__/ChallengeManage.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只收口一条 `featureRouterImportAllowlist`，不代表剩余条目都不合理；仍需逐条判定。
- 这轮 review 默认仍是同上下文 self-review；独立 reviewer gate 仍需单独说明。
