---
name: code-reviewer
description: Use when reviewing code changes, pull requests, patches, or implementation plans where correctness, regressions, architecture impact, security, test quality, and review communication all matter more than writing new code.
---

# Code Reviewer

本仓库里的这份 `code-reviewer` 是项目补充入口，不再作为通用 review workflow 的主体。

## 主体来源

- 通用主体：`~/.agents/skills/code-reviewer`
- 项目补充：
  - `development-pipeline`
  - `frontend-engineer`
  - `go-backend`

## 在 CTF 仓库中的使用方式

review 本仓库 diff、实现计划或评审文档时：

1. 先使用全局 `code-reviewer`
2. 再结合本仓库 references 和 `docs/reviews/` 的既有事实源判断风险
3. 如果 review 触达前后端边界或已知结构债，继续补读对应领域 skill 与架构文档

## 本地补充关注点

- touched surface 上已知的 owner 混杂、超大文件或结构债，在本仓库默认按 blocker 处理
- review 不只看功能对错，也要核对计划是否真的把结构收口到了唯一 owner
- 新的独立 review 记录应归档到 `docs/reviews/` 对应类别，而不是散落在临时笔记里

## 项目内附带参考

- `references/ctf-current-review-status-checks.md`
- `references/engineering-standards.md`
- `references/technical-risk-checks.md`
- `references/test-strategy-review.md`
- `references/review-communication.md`
- `docs/reviews/`
