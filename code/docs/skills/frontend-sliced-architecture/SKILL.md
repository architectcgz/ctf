---
name: frontend-sliced-architecture
description: 本仓库前端架构迁移与边界治理规范，约束 route view/feature/model 的职责边界。
---

# Frontend Sliced Architecture (CTF Repo)

## 目标
- `views/*` 保持组合层，不承载业务流程。
- 数据访问、副作用、流程编排下沉到 `features/*/model/*`。
- feature 之间通过 public API（`index.ts`）通信，禁止跨 slice 深导入内部实现。

## 分层约束
- `views` 运行时代码禁止导入非 `@/api/contracts` 的 API。
- `views` 运行时代码禁止直接使用 `useRoute/useRouter/router.push/router.replace/useRouteQueryTabs`。
- 运行时代码禁止跨 slice 深导入 `@/features/*/model/*`。
- `features/model` 运行时代码禁止反向依赖 `@/components/*`。

## 标准扫描
```bash
rg -n "useRoute|useRouter|router\\.push|router\\.replace|useRouteQueryTabs" code/frontend/src/views/**/*.vue
rg -nP "from ['\"]@/api/(?!contracts)" code/frontend/src/views/**/*.vue
rg -n "from ['\"]@/features/.+/model/" code/frontend/src --glob '*.{ts,vue}'
rg -n "from ['\"]@/components/" code/frontend/src/features code/frontend/src/entities --glob '*.{ts,vue}'
rg -nP "from ['\"]@/api/(?!contracts)" code/frontend/src/components code/frontend/src/widgets code/frontend/src/entities --glob '*.{ts,vue}'
```

## 最小验证
```bash
npm run test:run -- src/features/__tests__/featureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts
npm run typecheck
```

## 迁移记录
- 批次记录：`code/docs/tasks/2026-05-02-frontend-fsd-migration-plan-and-scan.md`
