---
name: ctf-go-backend-architecture-reference
description: >
  Use when applying onion-clean-architecture or go-backend inside this CTF
  repository and you need the current backend layout, local architecture
  documents, module ownership, and repo-specific reference paths. Activate when
  the task moves files across layers, adds or splits ports, decides where a
  service/repository/handler belongs, or needs the current modular-monolith
  boundaries — i.e. "这个该放哪一层 / 拆个 port / 后端目录结构是怎样的".
---

# CTF Go Backend Architecture Reference

CTF 仓库对通用后端架构 skill 的**薄补充**：只提供本仓当前布局、本地架构文档和 owner 事实，
方法论本身在通用 skill 里。命中后按需读下面的本地文档。

## Use When
- 在本仓应用 onion-clean-architecture / go-backend 的方法论时需要本地落点。
- 移动跨层文件、新增或拆分 port、判断 service/repository/handler 归属。
- 需要当前 modular-monolith 边界和 module ownership。

## Do Not Use
- 通用分层方法、依赖方向、port/adapter 原理 → `onion-clean-architecture`、`go-backend`。
- 后端 pattern / 并发幂等 / 测试分层 → `ctf-backend-patterns`。

## Read Alongside（方法论事实源）
- `onion-clean-architecture`
- `go-backend`

## Current Local References
- Backend root: `code/backend`
- Architecture decision: `docs/architecture/01-backend-architecture-style-decision.md`
- Modular-monolith refactor: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Implementation plans: `docs/plan/impl-plan/`
- Review evidence: `docs/reviews/backend/` 与 `docs/reviews/general/`

## CTF Reading Focus（每条带 ✓Check）
- 移动文件前先读当前 module ownership。
  ✓Check：这次改动是否跨越了某个 `internal/module/*` 的 bounded-context 边界？
- 优先本地 modular-monolith 边界，而非教科书式 Clean Architecture 目录摆设。
  ✓Check：新建目录是为真实边界，还是只为"看起来分层完整"？
- 把 `internal/module/*` 当作主要 bounded-context surface。
- 新增/拆分 port 时保持 `handler -> application -> ports -> infrastructure` 的 owner 显式，
  并对齐本仓 review 与 reuse-first gate。
  ✓Check：每个 port 的实现 owner 是否唯一、依赖方向是否只向内？
