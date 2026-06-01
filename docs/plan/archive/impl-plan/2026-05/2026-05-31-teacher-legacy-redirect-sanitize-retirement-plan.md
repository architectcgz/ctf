# 教师端 legacy redirect sanitize 退场计划

## Objective

- 删除 `teacherLegacyRedirect.ts`，结束旧教师端前端页面路径的最后一层 canonicalize 兼容。
- 保持登录 redirect 参数遇到 `/teacher/*` 历史路径时安全回退到角色默认首页。

## Non-goals

- 不修改教师端后端 `/api/v1/teacher/*` transport path。
- 不改 `/academy/*` canonical 页面路由和导航结构。
- 不引入新的 route allowlist 体系。

## Source Inputs

- `code/frontend/src/utils/redirectPath.ts`
- `code/frontend/src/features/auth/model/useLoginPage.ts`
- `code/frontend/src/router/guards.ts`
- `code/frontend/src/utils/roleRoutes.ts`
- `docs/architecture/frontend/02-routing.md`
- `TODO/frontend-sliced-architecture.md`

## Task Slices

### Slice 1: 删除最后的 legacy redirect owner

- 目标：去掉 `teacherLegacyRedirect.ts`，并把 `sanitizeRedirectPath()` 改成直接拒绝 legacy `/teacher/*` 页面路径。
- 变更面：
  - `code/frontend/src/utils/redirectPath.ts`
  - `code/frontend/src/utils/teacherLegacyRedirect.ts`
  - `code/frontend/src/utils/__tests__/redirectPath.test.ts`
  - `code/frontend/src/router/__tests__/guards.test.ts`
  - `code/frontend/src/features/auth/model/useLoginPage.test.ts`
- 风险：
  - 如果 redirect 清洗后不再落到 `/`，登录后可能跳向已退场页面。

### Slice 2: 收紧源码 guardrail 与事实文档

- 目标：不再给旧教师端页面路径保留任何前端源码 owner，并把事实同步回文档。
- 变更面：
  - `code/frontend/src/__tests__/architectureBoundaries.test.ts`
  - `docs/architecture/frontend/02-routing.md`
  - `TODO/frontend-sliced-architecture.md`
- 风险：
  - 如果 guardrail 仍然为已删除 owner 留豁免，会留下结构性死规则。

## Validation Plan

- `npm run test:run -- src/utils/__tests__/redirectPath.test.ts src/router/__tests__/guards.test.ts src/features/auth/model/useLoginPage.test.ts src/__tests__/architectureBoundaries.test.ts src/router/__tests__/sharedRoutes.test.ts`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-reuse-first.sh`

## Review Focus

- legacy `/teacher/*` redirect 参数是否稳定回退到角色默认首页，而不是跳去 404。
- 删除 `teacherLegacyRedirect.ts` 后，源码 guardrail 是否准确反映“前端页面路径兼容 owner 已清空”。
- 文档是否明确区分“前端页面兼容已删完”与“后端 `/api/v1/teacher/*` transport 仍保留”。

## Rollback / Recovery

- 如果发现外部历史入口仍必须保留，可恢复 `sanitizeRedirectPath()` 的 teacher path 特例，但不恢复 router runtime route；需要重新单独评估兼容窗口与 owner 落点。
