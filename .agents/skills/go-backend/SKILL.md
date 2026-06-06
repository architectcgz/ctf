---
name: go-backend
description: Use when implementing, refactoring, or reviewing Go backend code, especially context propagation, handlers, services, repositories, jobs, workers, database access, concurrency, idempotency, cache, queues, external integrations, runtime operations, and Go-specific tests.
---

# Go Backend

本仓库里的这份 `go-backend` 是项目补充入口，不再作为通用 Go 后端规则的主体。

## 主体来源

- 通用主体：`~/.agents/skills/go-backend`
- 项目补充：
  - `ctf-go-backend-architecture-reference`
  - `onion-clean-architecture`

## 在 CTF 仓库中的使用方式

处理 Go 后端实现、重构或 review 时：

1. 先使用全局 `go-backend`
2. 再根据任务形态叠加：
   - 边界/分层/端口设计：`onion-clean-architecture`
   - 当前仓库后端布局和本地文档入口：`ctf-go-backend-architecture-reference`

## 本地补充关注点

- 后端根目录：`code/backend`
- 当前主模块边界在 `internal/module/*`
- 实现前后继续遵守仓库自己的 reuse-first、review gate 和架构检查脚本

## 项目内附带参考

- `references/repository-interface-splitting.md`
