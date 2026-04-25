# 后端 Review 文档状态

本文用于说明 `docs/reviews/backend/` 下历史 review 文档的当前使用方式，避免把旧轮次正文中的“未修复”直接当作当前待办。

## 当前约定

- 旧 review 文档正文保留原始审查结论，用于追溯问题来源、轮次变化和当时的修复依据。
- 当前状态优先看文档顶部新增的“当前状态 / 补充修复状态 / 已验证证据”区块。
- 没有顶部状态区块的旧文档，只能视为历史记录；是否仍是当前缺陷，需要按当前代码重新复核。
- 不直接删除旧 review 文档。确需清理时，优先移动到 `docs/reviews/backend/archive/`，并保留索引说明。

## 已复核完成

### runtime safety / context 契约

- 文档：`../ctf-backend-code-review-runtime-safety-round1-20260422.md`
- 当前状态：已完成。
- 覆盖问题：runtime 失败语义、CORS 空白名单、ctx 契约、runtime 与 contest 反向依赖、开发/生产配置安全、端口原子预留、`context.Background()` 使用边界、ctx 架构防线。
- 验证证据：文档内记录了 2026-04-25 合并后 `timeout 300s go test ./... -count=1` 通过。
- 说明：该文档下方的问题清单是首轮审查原文，当前状态以顶部“修复状态”“第 2 轮补充复核”“第 3 轮机械防线”为准。

### image-management

- 文档：`ctf-platform-code-review-image-management-round2-ae03fb5.md`
- 当前状态：已完成。
- 覆盖问题：分页配置、description 清空、镜像名/标签校验、审计日志、软删除查询、`images(deleted_at)` 索引、无效 ID 文案常量、泛型分页返回、`size_formatted` API 契约。
- 验证证据：文档内记录了 2026-04-25 后端全量 `timeout 300s go test ./... -count=1` 通过。
- 说明：该文档的“问题清单”和“未修复”段落保留第 2 轮原始审查文本，已经被顶部“2026-04-25 补充修复状态”覆盖。

## 需要重新复核

以下旧文档仍可能包含有效问题，但也可能引用旧目录、旧模块边界或已经被后续重构覆盖。处理前需要按当前 `main` 代码重新确认，不应只根据 `rg "未修复"` 建立待办。

- `backend-architecture-review-round1-9120e85.md`
- `backend-code-review-auth-foundation-*`
- `contest-code-review-*`
- `ctf-backend-code-review-*`
- `ctf-code-review-*`
- `ctf-platform-code-review-*`
- `zhicore-ctf-code-review-*`

## 后续维护规则

- 修复 review 问题时，优先在对应文档顶部追加当前状态和验证证据，不改写历史问题正文。
- 如果某个旧 review 已确认完全过期，先移动到 `archive/`，再在本文件补一行归档说明。
- 如果某个旧 review 仍有当前有效缺陷，先开新的复核小节，列出当前代码位置、风险和验证证据，再进入修复。
