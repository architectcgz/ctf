# 本地架构 Guardrail 补强实现方案

## Objective

在已有 `scripts/check-architecture.sh` 基础上继续补齐架构防腐约束，把后端 domain 纯度、跨模块依赖、context/time/transaction 边界，以及前端 API、router/store、旧 components 页面目录等风险点纳入本地可执行检查。

## Non-goals

- 不在本次重构业务代码，不清理已有架构债。
- 不引入 CI；本阶段仍以本地 hook 和手动脚本为主。
- 不把历史债直接判失败；采用 allowlist 锁住当前基线，禁止新增同类腐蚀。

## Inputs

- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/frontend/01-architecture-overview.md`
- `AGENTS.md` 中的 Backend Context Contract、Backend Time Contract、Frontend Guardrails
- 已有 `scripts/check-architecture.sh`
- 已有 `code/backend/internal/module/architecture_test.go`
- 已有 `code/frontend/src/__tests__/architectureBoundaries.test.ts`

## Task Slices

1. 后端模块架构测试补强
   - 扩展 `architecture_test.go`
   - 增加 domain 内部 model/dto/config 依赖基线
   - 增加模块依赖关系图基线
   - 增加 context.Background/TODO、time.Now、transaction 位置基线
   - 增加 runtime 文件大小与 `*gorm.DB` 层级污染检查

2. 前端架构测试补强
   - 扩展 `architectureBoundaries.test.ts`
   - 锁住 components/widgets 直接访问非 contracts API 的历史基线
   - 禁止 common/entities 依赖 router/store/API/feature 等上层能力
   - 锁住旧 `components/*Page.vue` 页面文件基线，禁止新增
   - 检查 features/ui 不直接访问非 contracts API

3. 验证与收口
   - 跑 `bash scripts/check-architecture.sh --full`
   - 跑 `bash scripts/check-consistency.sh`
   - 跑后端目标架构测试
   - 跑前端 lint / prettier / vitest
   - 只提交本次相关文件

## Compatibility

所有新增检查均使用当前仓库状态生成 allowlist。现有历史债不会阻断本次提交；后续新增同类依赖或新增超阈值文件会失败。历史债减少时，allowlist stale 检查会要求同步删除对应条目。

## Review Focus

- allowlist 是否只是锁当前基线，没有悄悄放宽规则。
- 检查项是否误伤测试文件、data/testsupport 等非运行时文件。
- 是否仍然可通过 `scripts/check-architecture.sh --full` 一键验证。

## Rollback

如果某条新增检查误报过高，可回滚对应测试块或将其从 `--quick` 中移除；本次不改变运行时代码，回滚只影响本地检查强度。
