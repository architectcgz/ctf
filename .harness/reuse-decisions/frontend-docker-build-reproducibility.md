# Reuse Decision

## Change type
frontend build / package manager / docker reproducibility

## Existing code searched
- code/frontend/src/api
- code/frontend/src/features
- code/frontend/src/components
- code/frontend/Dockerfile
- code/frontend/package.json
- code/frontend/pnpm-lock.yaml
- code/frontend/README.md

## Similar implementations found
- 当前前端工程已经以 `pnpm-lock.yaml` 为唯一锁文件来源，说明镜像构建应直接围绕 lockfile 做严格安装，而不是在 Docker build 里动态插入交互式批准流程。
- 仓库内没有现成 `.npmrc` 或 `packageManager` 锚点，说明最小正确落点应放在 `package.json`，让 pnpm 版本和 build script allowlist 成为源码的一部分。

## Decision
extend_existing

## Reason
当前 Dockerfile 的 `pnpm install --frozen-lockfile || true -> pnpm approve-builds esbuild -> pnpm install` 会把真实安装错误吞掉，并依赖半成品状态继续构建，无法满足可复现构建要求。

最小正确修正是：

- 在 `package.json` 固定 `packageManager`
- 在 `package.json` 显式声明 `pnpm.onlyBuiltDependencies = [\"esbuild\"]`
- Dockerfile 只保留单次严格安装，不再插入 `approve-builds` 和吞错分支

这样既保留 pnpm 10 的 build script 审批模型，又让镜像构建行为完全由源码和 lockfile 决定。

## Files to modify
- .harness/reuse-decisions/frontend-docker-build-reproducibility.md
- code/frontend/Dockerfile
- code/frontend/package.json

## After implementation
- 前端镜像构建恢复为严格、可复现的安装流程。
- `esbuild` 的 build script 白名单进入源码配置，不再依赖 Docker build 里的交互式批准命令。
- pnpm 版本不再跟随 `latest` 漂移。
