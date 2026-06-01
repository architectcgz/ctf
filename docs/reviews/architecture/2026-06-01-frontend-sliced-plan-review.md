# 2026-06-01 Frontend Sliced Plan Review

## Review Target

- Repository: `ctf`
- Branch: `main`
- Diff source: working tree plan/doc updates
- Focus: `TODO/frontend-sliced-architecture.md` 是否与当前前端事实源、既有架构 debt 和 guardrail 对齐
- Files reviewed:
  - `TODO/frontend-sliced-architecture.md`
  - `docs/architecture/frontend/01-architecture-overview.md`
  - `docs/architecture/frontend/07-pages-dataflow.md`
  - `code/frontend/src/router/routes/teacherRoutes.ts`
  - `code/frontend/src/router/__tests__/guards.test.ts`
  - `code/frontend/src/features/platform/contests/index.ts`
  - `code/frontend/src/api/admin/contests.ts`
  - `code/frontend/src/__tests__/architectureBoundaries.test.ts`
  - `code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts`

## Classification Check

- Classification: architecture review
- Decision: agree with non-trivial review classification

## Gate Verdict

- Verdict: pass with minor issues

## Findings

### P1-1 方案遗漏了当前最重的 contest domain 结构债，优先级会失真

- Location:
  - `TODO/frontend-sliced-architecture.md:66-139`
  - `code/frontend/src/features/platform/contests/index.ts:1-7`
  - `code/frontend/src/api/admin/contests.ts:1-40`
- Evidence:
  - 方案把后续重点放在 `/teacher` 兼容退场、route/widget/entity 边界和 entity 补强。
  - 但当前 runtime 主路径里，`features/platform/contests` 仍是活动兼容聚合层，route page 和 `contest-manage` 还在直接使用它。
  - `api/admin/contests.ts` 仍是 contest domain 的超级 API owner，`contest-manage`、`contest-operations`、`contest-awd-admin`、`contest-awd-config`、`contest-workbench` 等多条链路都继续依赖它。
- Risk:
  - 如果 plan 不把这组 debt 纳入当前阶段，后续会继续在 contest domain 上“目录拆开但 owner 没收口”。
  - 这样会让结构迁移看起来在推进，但真正最大的 blast radius 还留在主链路里。
- Recommendation:
  - 保留这份 plan 的整体方向，但把 `contest public API / API client owner 收口` 提升到和 `entity 补强` 同级，至少进入 P1。

### P1-2 文档把“现状事实”和“候选目标目录”写在了一起，容易让后续实现误以为已经定案

- Location:
  - `TODO/frontend-sliced-architecture.md:149-166`
- Evidence:
  - “本次记录依据”里写了：
    - `api/request` 和基础 API client -> `shared/api`
    - 主题、路由 meta、导航配置 -> `shared/config` 或 `app/config`
  - 但当前代码树里没有 `src/shared/api`、`src/shared/config`、`src/app/config` 这些活动目录。
- Risk:
  - 这会把“现在的事实”与“未来可能采用的落点”混在一起。
  - 后续执行者容易把它当成既定目标，提前做目录级迁移，反而偏离你前面强调的 minimal diff 和 reuse-first。
- Recommendation:
  - 这一段拆成两类：
    - 当前事实源：现有 `src/api`、`src/config`、`src/shared/lib/request` 等真实路径
    - 候选演进方向：如果未来要引入 `shared/api` / `app/config`，明确标成 proposal，不要混在“记录依据”里

### P2-1 新方案和当前 frontend 事实文档存在一处显式冲突，需要尽快统一

- Location:
  - `TODO/frontend-sliced-architecture.md:68-75`
  - `docs/architecture/frontend/01-architecture-overview.md:88-94`
- Evidence:
  - 新方案写的是：前端 runtime 已不再注册 `/teacher/* -> /academy/*` 页面 redirect，旧路径只会在登录 redirect 参数上回退到默认首页。
  - 代码和测试也支持这个说法，例如 `router/routes/teacherRoutes.ts` 只注册 `academy/*`，`guards.test.ts` 里对 `/teacher/dashboard` 的预期是回退到 `/academy/overview`。
  - 但 `01-architecture-overview.md` 仍写“旧 `/teacher/*` 当前只保留 redirect 兼容”。
- Risk:
  - 事实源冲突会让后续方案评审和 guardrail 设计基线不一致。
- Recommendation:
  - 以新方案和当前代码为准，更新 frontend overview 对教师端旧兼容路径的描述。

## Material Findings

- P1-1 需要把 contest domain 的 public API / API client 收口纳入方案优先级。
- P1-2 需要把“事实”与“候选目录目标”分开表述。

## Senior Implementation Assessment

- 这版方案的主判断是对的：当前前端已经不是“继续搬目录”的阶段，而是“收紧 owner、补 guardrail、保持 route page 薄壳”的阶段。
- 真正的问题不是方向错，而是收敛范围还差一块最重的 contest domain，外加末尾有少量“未来目录猜想”混入了事实文档。

## Required Re-validation

- 文档修正后，至少重新核对：
  - `code/frontend/src/router/routes/*.ts`
  - `code/frontend/src/features/platform/contests/index.ts`
  - `code/frontend/src/api/admin/contests.ts`
  - `docs/architecture/frontend/01-architecture-overview.md`
- 若据此更新 guardrail 或结构计划，建议补一次代码搜索：
  - `cd code/frontend && rg "@/features/platform/contests|@/api/admin/contests" src`

## Residual Risk

- 这次 review 只评估方案与当前代码结构的一致性，没有重新展开所有页面级运行时行为验证。
- `entity` 补强本身不是错误方向，只是当前优先级低于 contest domain 结构收口。

## Touched Known-Debt Status

- 本次方案已经吸收了 route page / widget / shared-layout 边界收口的多数历史结论。
- 但 contest domain 的兼容 barrel 和超级 API client 仍未进入方案主线，当前仍应视为已知未收口 debt。
