# Reuse Decision

## Change type
frontend guard / script / docs / lint

## Existing code searched
- code/frontend/package.json
- code/frontend/eslint.config.js
- code/frontend/scripts/check-frontend-growth-guard.mjs
- code/frontend/scripts/check-theme-tail.mjs
- code/frontend/scripts/frontend-architecture-policy.json
- scripts/check-frontend-architecture.sh
- scripts/check-architecture.sh
- scripts/check-consistency.sh
- docs/architecture/README.md
- docs/architecture/frontend/README.md
- docs/architecture/frontend/08-build-deploy.md
- code/frontend/src

## Similar implementations found
- `code/frontend/scripts/check-theme-tail.mjs` 已经在做“扫描真实产品路径 -> 发现禁用样式模式 -> 直接失败”的前端样式守卫。
- `scripts/check-frontend-architecture.sh` 已经是前端架构 guard 的统一入口，并且会被 `.githooks/pre-commit` 通过 `scripts/check-architecture.sh --quick` 复用。
- `scripts/check-consistency.sh` 已经负责校验架构文档与 guardrail 接线是否保持一致。

## Decision
extend_existing

## Reason
这次不是新增另一套 lint / stylelint 体系，而是在现有前端架构检查链路里补一条 `:deep` 存量守卫。

最小正确改动是：

- 复用 `code/frontend/scripts/*.mjs` 的扫描脚本模式
- 复用 `scripts/check-frontend-architecture.sh` 作为 guard 入口
- 用显式 allowlist 记录当前存量 `:deep`，后续只允许减少，不允许新增或变种
- 同步更新架构入口文档与一致性检查

## Files to modify
- .harness/reuse-decisions/frontend-vue-deep-guard.md
- code/frontend/package.json
- code/frontend/scripts/check-vue-deep-guard.mjs
- code/frontend/scripts/vue-deep-allowlist.json
- scripts/check-frontend-architecture.sh
- scripts/check-consistency.sh
- docs/architecture/README.md
- docs/architecture/frontend/README.md
- docs/architecture/frontend/08-build-deploy.md

## After implementation
- `:deep` 不再只是人工约定，而是前端架构 guard 的一部分。
- 后续页面如果继续依赖现有 `:deep`，只能在 allowlist 存量内收缩；新增文件、新增选择器或回流旧语法会直接失败。
