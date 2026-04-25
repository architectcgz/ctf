# CTF 架构与最终设计入口

## 定位

`docs/architecture/` 是当前项目的最终设计事实源。

- 架构边界、模块职责、接口协作、页面设计、专题设计，统一从这里读取。
- `docs/superpowers/specs/` 和 `docs/superpowers/plans/` 只保留过程资料，不再作为最终设计入口。
- `docs/reviews/` 是历史评审快照，不覆盖当前架构和产品设计。

## 读取顺序

1. `docs/architecture/README.md`
2. `docs/contracts/`
3. `docs/architecture/backend/`、`docs/architecture/frontend/`
4. `docs/architecture/backend/design/`
5. `docs/architecture/frontend/design-system/`
6. `docs/architecture/frontend/pages/`
7. `docs/architecture/features/`

## 目录说明

- `backend/`：后端总体架构、数据库、容器、API、关键流程等长期文档。
- `backend/design/`：已采用、但不适合并入总览文档的后端专题设计。
- `frontend/`：前端架构、路由、状态管理、API 层、组件体系等长期文档。
- `frontend/design-system/`：全局 UI 风格、主题、技术边界。
- `frontend/pages/`：页面级最终设计稿。
- `features/`：跨前后端的专题最终设计。

## 历史迁移说明

- 原 `design-system/ctf-platform/` 下的最终设计稿已迁入 `docs/architecture/frontend/`。
- 原 `docs/superpowers/specs/` 下仍有效的专题设计已同步收口到 `docs/architecture/features/`。
- 旧目录如果仍存在，只承担跳转和过程追溯，不再承载最终事实。
