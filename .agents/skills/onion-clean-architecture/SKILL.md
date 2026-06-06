---
name: onion-clean-architecture
description: Use when designing, scaffolding, reviewing, or refactoring backend projects toward Onion Architecture, Clean Architecture, modular monolith boundaries, ports/adapters, service ownership, or framework-independent domain/application layers.
---

# Onion Clean Architecture

本仓库里的这份 `onion-clean-architecture` 是项目补充入口，不再作为通用 Onion / Clean Architecture 规则的主体。

## 主体来源

- 通用主体：`~/.agents/skills/onion-clean-architecture`
- 项目补充：
  - `ctf-go-backend-architecture-reference`
  - `go-backend`

## 在 CTF 仓库中的使用方式

遇到后端边界设计、端口拆分、模块职责重排时：

1. 先使用全局 `onion-clean-architecture`
2. 再补读：
   - `ctf-go-backend-architecture-reference`
   - 当前仓库 `docs/architecture/` 与 `docs/plan/impl-plan/` 中对应事实源

## 本地补充关注点

- 当前后端主边界在 `code/backend/internal/module/*`
- 这里优先做“当前 modular monolith 的 owner 收口”，而不是为了像 Clean Architecture 去机械搬目录
- 任何 repo / service / handler 拆分都要对齐仓库自己的 review gate、reuse-first 和架构守卫

## 项目内附带参考

- `references/go-ref.md`
- `references/rust-ref.md`
