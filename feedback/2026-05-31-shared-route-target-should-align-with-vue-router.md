# 共享导航 target 契约应直接对齐 vue-router

## 问题描述

这次把 `components/navigation/routeTarget.ts` 迁到 `shared/lib/navigation/routeTarget.ts` 后，`typecheck` 仍然卡在平台页的导航 props 上。

表面症状是 `AppRouteLink`、`AppRouteRedirect` 和若干页面把 `to` / `route` / `auditLogRoute` 一类 props 写成了自定义 `AppRouteTarget` 或 `Record<string, unknown>`，但运行时实际消费的是 `vue-router` 的 `RouterLink`、`router.push()` 和 `router.replace()`。

结果就是共享导航契约声明的能力比运行时更窄，合法的 `name / path / params / query / hash` 组合在类型层反而不一定能通过。

## 原因分析

这里的共享层没有真正发明新的导航语义，只是在给 `vue-router` 做一层薄包装。

如果这时继续手写一套 `AppNamedRouteTarget`、`AppPathRouteTarget`、`AppRouteParamValue`，就需要长期追着 `vue-router` 的 `RouteLocationRaw` 能力和边界保持同步。一旦漏掉某种合法形状，类型约束就会比实际运行能力更弱，最后逼着页面改成 `Record<string, unknown>` 或局部断言，契约 owner 反而变得更模糊。

## 解决方案

- 共享导航契约默认直接别名到 `vue-router` 的 `RouteLocationRaw`，例如 `export type AppRouteTarget = RouteLocationRaw`。
- `AppRouteLink`、`AppRouteRedirect` 继续只消费这个薄别名，不再自己维护一套平行 target 类型。
- 页面和 feature 里凡是表达“跳去哪里”的 props、computed 或 helper，统一直接使用 `AppRouteTarget`，不要退回 `Record<string, unknown>`。
- 只有当项目确实引入了比 `vue-router` 更窄、并且稳定可复用的产品级导航约束时，才单独定义更上层的 route contract；否则不要在 shared 层重复造型。

## 收获

对第三方运行时能力做共享薄封装时，类型 owner 应尽量贴近底层事实源。共享层负责命名和收口，不负责重新定义一套容易漂移的平行契约。

## 沉淀状态

- 状态：待同步 skill
- Owner：`frontend-engineer` skill / CTF 前端共享导航封装
- 链接：`code/frontend/src/shared/lib/navigation/routeTarget.ts`、`code/frontend/src/shared/ui/navigation/AppRouteLink.vue`、`code/frontend/src/shared/ui/navigation/AppRouteRedirect.vue`
