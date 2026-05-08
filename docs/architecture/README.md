# CTF 架构与最终设计入口

## 定位

`docs/architecture/` 是当前项目的最终架构与最终页面设计事实源。

- 架构边界、模块职责、接口协作、页面设计、跨模块专题设计，统一从这里读取。
- `practice/superpowers-plan-index.md` 只保留过程资料索引，不再作为最终设计入口。
- `docs/reviews/` 是历史评审快照，不覆盖当前架构和产品设计。
- 单题题包设计、题面设计和解法设计不作为 `docs/architecture/` 的常驻内容；已落地题目以题包目录自身为事实源，仍在推演的题目稿进入 `docs/design/`。

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
- `features/`：面向产品能力或业务专题的最终架构事实，也承接专题当前已经固定的内部边界结论。
  - 目录索引：`features/专题架构索引.md`

## AWD Checker 扩展

- `features/AWD-http_standard检查器架构.md`：当前已实现的 `http_standard` checker 事实源。
- `features/AWD检查器运行器扩展设计.md`：当前 `tcp_standard`、`script_checker` 与 sandbox runner 执行链路的专题事实源；具体迁移状态以 `features/专题架构索引.md` 为准。

## 历史迁移说明

- 原 `design-system/ctf-platform/` 下的最终设计稿已迁入 `docs/architecture/frontend/`。
- 原 superpowers specs 下仍有效的产品专题设计已迁入 `docs/architecture/features/`。
- superpowers 过程资料索引已迁入 `practice/superpowers-plan-index.md`，不再承载最终专题设计。
