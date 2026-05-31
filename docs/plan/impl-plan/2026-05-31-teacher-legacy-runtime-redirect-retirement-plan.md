# 教师端 legacy runtime redirect 退场计划

## Objective

- 从 router runtime 移除 `/teacher/*` legacy 前端页面 route。
- 保留登录 redirect 参数里的 legacy canonicalize 兼容，不让旧路径继续进入登录后导航。

## Non-goals

- 不删除 `teacherLegacyRedirect.ts` 的路径归一 helper。
- 不修改教师端后端 HTTP `/api/v1/teacher/*` transport path。
- 不改教师端 canonical `/academy/*` 页面结构或 feature owner。

## Source Inputs

- `code/frontend/src/router/routes/teacherRoutes.ts`
- `code/frontend/src/router/__tests__/sharedRoutes.test.ts`
- `code/frontend/src/utils/teacherLegacyRedirect.ts`
- `code/frontend/src/utils/redirectPath.ts`
- `docs/architecture/frontend/02-routing.md`
- `TODO/frontend-sliced-architecture.md`

## Task Slices

### Slice 1: 退掉 teacher runtime redirect owner

- 目标：停止把 `teacherLegacyRedirect.ts` 注入 router runtime route 列表。
- 变更面：
  - `code/frontend/src/router/routes/teacherRoutes.ts`
  - `code/frontend/src/router/__tests__/sharedRoutes.test.ts`
  - `code/frontend/src/utils/teacherLegacyRedirect.ts`
- 风险：
  - 如果还有测试或页面隐含依赖 `/teacher/*` runtime route，会在 shared route 测试里暴露出来。

### Slice 2: 同步当前事实源

- 目标：把“`/teacher/*` 已退出 router，legacy 兼容只剩 redirect 参数 canonicalize”写回事实文档和迁移台账。
- 变更面：
  - `docs/architecture/frontend/02-routing.md`
  - `TODO/frontend-sliced-architecture.md`
- 风险：
  - 不能把 helper 仍保留的兼容能力误写成“所有 `/teacher/*` 语义都完全删除”。

## Validation Plan

- `npm run test:run -- src/router/__tests__/sharedRoutes.test.ts src/__tests__/architectureBoundaries.test.ts src/router/__tests__/guards.test.ts src/features/auth/model/useLoginPage.test.ts`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-reuse-first.sh`

## Review Focus

- router runtime 是否已经完全移除 `/teacher/*` 页面 route，而不是只删掉部分 allowlist。
- 登录 redirect sanitize 是否仍然能把 legacy 教师端路径归一到 `/academy/*`。
- 文档是否准确表达“页面 route 已退场，但 redirect 参数 canonicalize 仍在”。

## Rollback / Recovery

- 如果发现仍有外部页面流程直接依赖 `/teacher/*` runtime route，可先恢复 `teacherRoutes.ts` 的 redirect 注册，并把依赖面补进新的退场前置计划。
