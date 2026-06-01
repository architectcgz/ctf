# 登录 redirect legacy 教师端路径归一化计划

## Objective

- 把登录 redirect 参数里的 `/teacher/*` 旧教师端页面路径归一到 `/academy/*` 正式路径。
- 让 router redirect allowlist 与 redirect sanitize 复用同一套 legacy route 映射事实。

## Non-goals

- 不删除现有 `/teacher/*` redirect allowlist。
- 不改教师端正式 `/academy/*` 路由注册。
- 不扩展到学生端 `/dashboard`、`/instances` 等其他 legacy redirect。

## Source Inputs

- `TODO/frontend-sliced-architecture.md`
- `docs/architecture/frontend/02-routing.md`
- `code/frontend/src/router/routes/teacherRoutes.ts`
- `code/frontend/src/utils/redirectPath.ts`
- `code/frontend/src/features/auth/model/useLoginPage.ts`
- `code/frontend/src/router/guards.ts`

## Task Slices

### Slice 1: 抽出 teacher legacy redirect canonical owner

- 目标：把 `/teacher/* -> /academy/*` 的 canonical path 计算抽成单点 owner，供 router 与 redirect sanitize 共用。
- 变更面：
  - `code/frontend/src/utils/teacherLegacyRedirect.ts`
  - `code/frontend/src/router/routes/teacherRoutes.ts`
- 风险：
  - dynamic path 映射如果遗漏，会让某些历史入口只在 route redirect 可用、但登录 redirect 归一化失效。

### Slice 2: redirect sanitize 接入 canonicalization

- 目标：在保留 open redirect 防御的前提下，把 legacy teacher redirect 一并归一到正式路径。
- 变更面：
  - `code/frontend/src/utils/redirectPath.ts`
  - `code/frontend/src/features/auth/model/useLoginPage.ts`
  - `code/frontend/src/router/guards.ts`
- 风险：
  - 如果把 canonicalization 写进 page/model，而不是 util owner，会继续留下重复 owner。

### Slice 3: 补测试与同步事实文档

- 目标：验证登录页、guard helper 和 legacy redirect mapping 已统一收口，并把当前事实记回台账/路由文档。
- 变更面：
  - `code/frontend/src/utils/__tests__/redirectPath.test.ts`
  - `code/frontend/src/router/__tests__/guards.test.ts`
  - `code/frontend/src/features/auth/model/useLoginPage.test.ts`
  - `code/frontend/src/pages/auth/__tests__/LoginRoutePage.test.ts`
  - `docs/architecture/frontend/02-routing.md`
  - `TODO/frontend-sliced-architecture.md`
- 风险：
  - 如果测试只测静态路径，不测 query/hash 或 dynamic path，legacy redirect 仍可能在细节上漂移。

## Validation Plan

- `npm run test:run -- src/utils/__tests__/redirectPath.test.ts src/router/__tests__/guards.test.ts src/features/auth/model/useLoginPage.test.ts src/pages/auth/__tests__/LoginRoutePage.test.ts src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

## Review Focus

- `/teacher/*` canonical path owner 是否只有一处。
- `sanitizeRedirectPath()` 是否同时守住 open redirect 防御和 legacy route 归一化。
- 登录页、已登录访问 `/login` 的 guard，以及源码级台账是否都反映“教师端正式前端入口只认 `/academy/*`”。

## Rollback / Recovery

- 如果共享 canonical owner 导致 legacy route 映射出错，可先回退到 `teacherRoutes.ts` 原有 route redirect 定义，同时保留 redirect sanitize 的安全清洗逻辑，再按更小切片补 canonical owner。
