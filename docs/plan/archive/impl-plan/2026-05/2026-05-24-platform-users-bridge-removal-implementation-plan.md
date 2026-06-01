> 状态：Current
> 事实源：2026-05-24 前端架构 review、`platform-user-management` 已落地 owner、`platform-users` 当前桥接壳状态
> 替代：无

# Platform Users Bridge Removal Implementation Plan

## 目标

- 删除 `platform-users` 与 `composables/usePlatformUsers.ts` 这批已无 runtime 引用的历史桥接壳文件。
- 同步更新边界测试和实施文档，避免代码结构与计划描述继续分叉。

## 非目标

- 本轮不调整 `platform-user-management` 的真实实现。
- 本轮不改用户治理页面的用户可见结构、接口契约或交互流程。
- 本轮不扩大到其他 legacy composable / feature 桥接层。

## 输入依据

- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `docs/plan/impl-plan/2026-05-24-platform-user-management-feature-split-implementation-plan.md`
- `code/frontend/src/features/platform-users/**`
- `code/frontend/src/composables/usePlatformUsers.ts`
- `code/frontend/src/features/__tests__/featureBoundaries.test.ts`

## 当前结论

- `platform-user-management` 已经成为平台用户治理的真实 owner。
- `platform-users` 与 `composables/usePlatformUsers.ts` 当前只剩转发壳，runtime 搜索未发现实际导入。
- 在继续保留这些空壳时，review 中的 “over-broad bucket” 债务仍然留在磁盘结构上，边界测试也会继续维护历史入口。

## 任务切片

### Slice 1：删除无引用桥接壳并同步边界清单

- 目标：
  - 删除 `platform-users` 目录和 `usePlatformUsers` composable 的历史桥接壳。
  - 更新 `featureBoundaries` 与实施计划，使结构事实与当前 owner 一致。
- 预期改动：
  - `docs/plan/impl-plan/2026-05-24-platform-users-bridge-removal-implementation-plan.md`
  - `docs/plan/impl-plan/2026-05-24-platform-user-management-feature-split-implementation-plan.md`
  - `.harness/reuse-decisions/platform-users-bridge-removal.md`
  - `code/frontend/src/features/__tests__/featureBoundaries.test.ts`
  - `code/frontend/src/features/platform-users/**`
  - `code/frontend/src/composables/usePlatformUsers.ts`
- 验证：
  - `rg -n "platform-users|usePlatformUsers" code/frontend/src`
  - `npm run test:run -- src/features/__tests__/featureBoundaries.test.ts src/views/platform/__tests__/UserManage.test.ts`
  - `npm run typecheck`
  - `bash scripts/check-consistency.sh`
  - `git diff --check -- <touched files>`
- Review focus：
  - 删除后是否仍存在 runtime 入口引用。
  - 边界测试是否开始约束新的真实 owner，而不是继续维护桥接路径。

## 风险

- 如果还有未被搜索到的历史引用，删除桥接壳后会直接暴露为 typecheck 或测试失败。
- 计划文档若不一起更新，后续结构性任务仍可能误以为这些桥接壳需要继续保留。

## 回退方式

- 如删除后发现仍有存量引用，可从 Git 历史恢复单个桥接文件，再按真实引用位置重新拆切。
