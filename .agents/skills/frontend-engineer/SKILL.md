---
name: frontend-engineer
description: Use when implementing or refactoring frontend code, especially Vue components, composables, route views, forms, tables, dialogs, async flows, or interaction behavior where correctness, state ownership, lifecycle cleanup, and maintainability matter more than visual polish alone.
---

# Frontend Engineer

本仓库里的这份 `frontend-engineer` 是项目补充入口，不再作为通用前端工程规则的主体。

## 主体来源

- 通用主体：`~/.agents/skills/frontend-engineer`
- 项目补充：
  - `ctf-dark-surface-alignment`
  - `ctf-ui-theme-system`

## 在 CTF 仓库中的使用方式

处理 Vue 页面、组件、路由视图、异步交互或前端重构时：

1. 先使用全局 `frontend-engineer`
2. 再补读本仓库前端 references，并按需要运行局部检查脚本
3. 如果问题已经上升到 slice / route owner / 公共 API 边界，再叠加 `frontend-sliced-architecture`

## 本地补充关注点

- 前端页面路由只使用 `/academy/*` 与 `/platform/*`
- 页面壳、tab 面板、drawer 和共享内容组件要显式拆 owner，不要再靠 `embedded` 一类开关混切布局语义
- 前端测试优先锁行为、状态 owner 和边界；改测试后先跑 `bash scripts/check-frontend-test-guard.sh`

## 项目内附带参考

- `references/ctf-vue-async-theme-route-ownership.md`
- `references/`
- `scripts/inspect-frontend-boundaries.mjs`
- `scripts/check-alias-paths.mjs`
- `scripts/check-frontend-test-guard.sh`
