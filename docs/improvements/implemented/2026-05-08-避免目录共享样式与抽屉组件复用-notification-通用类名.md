# 避免目录共享样式与抽屉组件复用 notification 通用类名

## Status

implemented

## Context

通知中心右侧抽屉的头部区域使用了 notification-summary、notification-filter 等通用类名，而 journal-user-directory.css 已在全局声明 notification-summary 等目录页骨架样式。由于全局样式会命中 teleported 抽屉 DOM，局部 scoped style 没有显式重置的属性会被污染，造成抽屉间距调整体感不明显。此次已将抽屉头部类名切到 notification-drawer-*，并在 TopNav 装配层按 drawer 语义别名化，但这一类命名边界规则仍应沉淀到 durable frontend rule。

## Problem

- 目录页共享样式是全局引入的，命中范围不受组件边界限制。
- 一旦抽屉、弹窗、目录页复用同一批通用类名，teleported overlay 上的 DOM 也会被全局规则影响。
- scoped style 只能覆盖它显式声明的属性；没有重置的 `padding`、`margin`、`display`、`gap` 等会继续继承全局样式，导致“代码改了但页面几乎没变化”。

## Suggested Direction

- 对 teleported 抽屉、弹窗、popover 等共享覆盖层，默认使用明确的组件命名空间类名，例如 `notification-drawer-*`、`audit-dialog-*`。
- 目录页、工作区页面、全局共享样式文件不要占用过于宽泛的业务前缀作为基础骨架类名，尤其避免和 overlay 组件共用 `notification-*`、`profile-*` 这类通用块名。
- 当样式调整“体感无变化”时，排查顺序里应明确加入“检查是否有全局样式类名串扰”，而不是只看当前组件 scoped style。

## Target Owner

- skill: `frontend-engineer`
- agent: CTF 前端实现 / review 流程
- docs: frontend durable styling boundary rule
- code area:
  - `code/frontend/src/assets/styles/journal-user-directory.css`
  - `code/frontend/src/components/layout/NotificationDropdown.vue`
  - `code/frontend/src/components/layout/TopNav.vue`

## Evidence

- file:
  - `code/frontend/src/assets/styles/journal-user-directory.css`
  - `code/frontend/src/components/layout/NotificationDropdown.vue`
  - `code/frontend/src/components/layout/TopNav.vue`
- command:
  - `rg -n "notification-summary|notification-filter-tabs|notification-drawer-summary" code/frontend/src`
  - `npm run test:run -- src/components/layout/__tests__/NotificationDropdown.test.ts src/components/layout/__tests__/TopNav.test.ts`
  - `npm run typecheck`
- behavior:
  - 右侧通知抽屉头部原本使用 `notification-summary` 等类名，和目录页全局共享样式发生命中重叠。
  - 抽屉头部改成 `notification-drawer-*` 后，全局目录样式不再串到抽屉头部，间距修正和语义别名化一并落地。

## Decision Log

- 2026-05-08: Created.
