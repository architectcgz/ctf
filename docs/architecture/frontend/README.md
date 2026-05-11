# 前端架构索引

> 状态：Current
> 事实源：`code/frontend/src/` 当前路由、页面、composable、共享原语与架构守卫
> 替代：无

## 定位

`docs/architecture/frontend/` 只保留当前前端架构事实、主题规则和长期页面组织约束。

- 这里回答“路由命名空间怎么分、状态放哪里、共享原语如何复用、主题和间距如何守住”。
- 过程中的页面草稿和方案比较不放这里，进入 `docs/design/`。
- 面向单个业务专题且已经稳定的页面边界，回到 `docs/architecture/features/` 的 owning 文档。

## 读取顺序

1. [01-architecture-overview.md](./01-architecture-overview.md)
2. [02-routing.md](./02-routing.md)
3. [03-state-management.md](./03-state-management.md)
4. [04-api-layer.md](./04-api-layer.md)
5. [05-websocket-composables.md](./05-websocket-composables.md)
6. [06-components.md](./06-components.md)
7. [07-pages-dataflow.md](./07-pages-dataflow.md)
8. [08-build-deploy.md](./08-build-deploy.md)
9. [09-spacing-system.md](./09-spacing-system.md)

## 当前活动文档

| 文档 | 说明 |
| --- | --- |
| [01-architecture-overview.md](./01-architecture-overview.md) | 前端分层、路由命名空间、共享原语与页面拆分基线 |
| [02-routing.md](./02-routing.md) | `/academy/*`、`/platform/*` 与学生端路由边界 |
| [03-state-management.md](./03-state-management.md) | Pinia、页面级 composable、只读派生与跨页状态归属 |
| [04-api-layer.md](./04-api-layer.md) | API 请求层、错误映射、认证与重试边界 |
| [05-websocket-composables.md](./05-websocket-composables.md) | WebSocket、ticket、通知与实时状态同步 |
| [06-components.md](./06-components.md) | 共享 UI 原语、工作台表格、弹窗模板与主题规则 |
| [07-pages-dataflow.md](./07-pages-dataflow.md) | 典型页面的数据流、owner 和交互拆分 |
| [08-build-deploy.md](./08-build-deploy.md) | Vite 构建、环境变量与部署约束 |
| [09-spacing-system.md](./09-spacing-system.md) | 间距 token、目录页节奏与工作台布局 guardrail |

## Guardrail

- 路由与页面边界：`code/frontend/src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- 前端分层约束：`code/frontend/src/__tests__/architectureBoundaries.test.ts`
- 导航命名空间：`code/frontend/src/config/__tests__/backofficeNavigation.test.ts`
- 共享弹窗模板：`code/frontend/src/components/common/__tests__/ModalTemplates.test.ts`
