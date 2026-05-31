# 2026-05-31 composables owner cleanup batch9 final retirement plan

> 状态：Draft
> 关联 reuse decision：`.harness/reuse-decisions/composables-owner-cleanup-batch9-final-retirement.md`

## 目标

把 `src/composables` 的最后一份历史测试迁到 `shared/model/common` 邻近，并同步更新前端脚本与当前架构文档，让 `src/composables` 完全退出活动层。

## 非目标

- 不改 `useProbeEasterEggs` 的阈值、存储 key、sessionStorage 降级逻辑或对外 API
- 不回填历史 review / archive / 旧 plan 文档里对 `src/composables` 的追溯描述
- 不触碰与本批无关的 `views / components / reports / awd-review` 历史线

## 输入事实源

- `code/frontend/src/composables/__tests__/useProbeEasterEggs.test.ts`
- `code/frontend/src/shared/model/common/useProbeEasterEggs.ts`
- `code/frontend/scripts/check-theme-tail.mjs`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/02-routing.md`
- `docs/architecture/frontend/03-state-management.md`
- `docs/architecture/frontend/09-spacing-system.md`

## 任务切片

### Slice 1

迁移最后一份历史测试：

- 新建 `shared/model/common/__tests__/useProbeEasterEggs.test.ts`
- 删除 `composables/__tests__/useProbeEasterEggs.test.ts`

验证：

- `cd code/frontend && timeout 180s npm run test:run -- src/shared/model/common/__tests__/useProbeEasterEggs.test.ts`

### Slice 2

同步活动脚本与当前事实源：

- 更新 `check-theme-tail.mjs` 的扫描根
- 更新前端架构文档里对 `src/composables` 活动层的描述

验证：

- `cd code/frontend && timeout 180s npm run check:theme-tail`
- `python3 scripts/check-docs-consistency.py`
- `git diff --check`

## 风险点

- `check-theme-tail.mjs` 改扫描根后，可能把之前未扫到的活动层硬编码 token 暴露出来
- 文档如果还把 `src/composables/use*.ts` 写成当前活动 owner，会和代码事实继续漂移

## Review focus

- `src/composables` 是否彻底退出活动层，而不是只把运行时代码清空
- `useProbeEasterEggs` 测试是否正确跟随 `shared/model/common` owner
- 文档是否只更新当前事实源，没有误改历史追溯文档
