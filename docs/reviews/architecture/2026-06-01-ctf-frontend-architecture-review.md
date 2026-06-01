# 2026-06-01 CTF Frontend Architecture Review

## Review Target

- Repository: `ctf`
- Scope: `code/frontend/src`, `code/frontend/scripts`, `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Focus: current frontend slice ownership, API client boundaries, architecture guard coverage
- Files reviewed:
  - `code/frontend/src/__tests__/architectureBoundaries.test.ts`
  - `code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts`
  - `code/frontend/src/features/platform/contests/index.ts`
  - `code/frontend/src/features/platform/contest-manage/model/useContestEditPage.ts`
  - `code/frontend/src/pages/platform/contests/ContestManageRoutePage.vue`
  - `code/frontend/src/pages/platform/contests/ContestOperationsRoutePage.vue`
  - `code/frontend/src/api/admin/contests.ts`
  - `code/frontend/src/api/admin/index.ts`
  - `code/frontend/src/api/admin/teaching.ts`
  - `code/frontend/src/features/platform/student-management/model/usePlatformStudentDirectory.ts`
  - `code/frontend/scripts/frontend-growth-baseline.json`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Classification Check

- Classification: architecture review
- Decision: agree with non-trivial review classification

## Gate Verdict

- Verdict: pass with minor issues

## Findings

### P1-1 Contest API client boundary is still a super-module, so feature拆分还没有收口到 transport owner

- Location:
  - `code/frontend/src/api/admin/contests.ts:1-8`
  - `code/frontend/src/api/admin/contests.ts:274-303`
  - `code/frontend/src/api/admin/contests.ts:1084-1490`
- Evidence:
  - `contests.ts` 同时承接 AWD review archive、contest CRUD、announcement、round、readiness、service orchestration、traffic、scoreboard、attack log。
  - 该文件开头直接依赖 `../teaching/awd-reviews`，再以 `listPlatformAWDReviews()` / `getPlatformAWDReview()` / `exportPlatformAWDReview*()` 的形式重新暴露平台端接口。
  - 运行时 feature 仍大面积直接依赖这个超大模块，例如 `contest-manage`、`contest-operations`、`contest-awd-admin`、`contest-awd-config`、`contest-projector`。
- Risk:
  - 页面/feature 层已经开始按 `contest-manage`、`contest-operations`、`contest-awd-admin` 拆开，但 API client 仍把这些职责重新耦回一个超级入口。
  - 任意一条竞赛相关链路调整接口、DTO 归一、错误处理或缓存策略时，blast radius 仍然覆盖整条 contest domain。
  - `admin` 与 `teaching` 的 review/reporting 边界在 transport 层继续模糊，后续一旦两端权限、字段或 rollout 节奏分叉，前端会先被这层别名耦合卡住。
- Recommendation:
  - 以 feature family 为单位继续拆 `api/admin/contests.ts`，至少先拆成 `contest-manage`、`contest-announcements`、`contest-operations`、`contest-awd-admin`、`contest-reviews`。
  - 让 role-neutral 的 review API 有独立 owner，不再通过 `admin/contests.ts` 间接转发 `teaching/awd-reviews`。

### P1-2 `features/platform/contests` 仍是活动中的兼容聚合层，平台竞赛切片边界还不够硬

- Location:
  - `code/frontend/src/features/platform/contests/index.ts:1-7`
  - `code/frontend/src/pages/platform/contests/ContestManageRoutePage.vue:1-7`
  - `code/frontend/src/pages/platform/contests/ContestOperationsRoutePage.vue:1-11`
  - `code/frontend/src/features/platform/contest-manage/model/useContestEditPage.ts:5-18`
- Evidence:
  - `features/platform/contests/index.ts` 明确标注“向后兼容重导出”，但当前 route page 仍通过它导入 `PlatformContestManagePage`、`PlatformContestOperationsPage`、`PlatformContestAnnouncementsPage`、`PlatformContestEditPage`。
  - `useContestEditPage()` 也继续从这个兼容入口反向拿 `buildContestAnnouncementsRoute`、`buildContestAwdConfigRoute`、`buildContestUpdatePayload` 等能力，而不是直接依赖 owning slice。
- Risk:
  - `contest-manage`、`contest-announcements`、`contest-operations` 表面上已经分成独立目录，但对 consumer 来说仍像一个大 feature，owner 边界仍然是软约定。
  - 兼容桶一旦长期存在，就会继续吸收“顺手放这里”的新导出，最后把已完成的 feature split 再次拉回伞状命名空间。
  - 当前架构测试只约束 layer 和 route page，不约束 feature 同层之间必须走明确 public API，类似的回流不会被 guard 拦住。
- Recommendation:
  - 先把 route pages 和 `contest-manage` 内部 consumer 全部切到各自 owning feature 的直接 public API。
  - consumer 清零后删除 `features/platform/contests/index.ts`，再补一条 guard，禁止新增同类兼容聚合层。

### P2-1 当前 guard 能证明“路由层变薄”，但还不能证明“大 owner 正在持续收口”

- Location:
  - `code/frontend/src/__tests__/architectureBoundaries.test.ts:148-239`
  - `code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts:29-98`
  - `code/frontend/scripts/frontend-growth-baseline.json:1-27`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md:42-97`
- Evidence:
  - 现有测试主要覆盖 layer 顺序、route page 入口、route page 体量、低层 forbidden imports。
  - growth baseline 只覆盖 `contest-awd-admin` 的 5 个文件。
  - backlog 自己已经把 `features/platform/contests`、`features/challenge-topology-studio`、`features/contest-awd-workspace`、`features/awd-inspector` 等列为当前最大的 owner 面。
- Risk:
  - 现在 `npm run check:frontend-growth` 通过，只能说明少数热点文件没有继续膨胀，不能说明当前最大的几组 owner 没在继续横向长大。
  - 团队容易把“guard 绿了”误读成“结构债已受控”，从而延后对真正热点面的持续切片。
- Recommendation:
  - 把 growth guard 从“单文件护栏”扩到“当前 backlog P1/P2 owner 面”。
  - 至少给 `platform/contest-*`、`challenge-topology-studio`、`contest-awd-workspace`、`awd-inspector` 增加目录级或关键 page-model 级预算。

## Material Findings

- P1-1 Contest API client 仍是超级 owner。
- P1-2 Platform contest compatibility barrel 仍在主路径使用。

## Senior Implementation Assessment

- 当前前端主骨架已经明显比 2026-05 的状态更健康：`pages` 很薄，route hooks 已经退到 feature/shared model，`components/*` 也基本退出活动结构。
- 剩下的问题不再是“有没有按 FSD 分目录”，而是“目录拆开后，transport/public API/guard 是否同步收口”。现在最明显的未收口点集中在 contest domain。

## Required Re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `cd code/frontend && npm run check:frontend-growth`
- `cd code/frontend && npm run typecheck`
- 对后续 contest API / feature barrel 收口，补一条代码搜索验证：
  - `cd code/frontend && rg "@/features/platform/contests|@/api/admin['\"]" src`

## Residual Risk

- 这次 review 主要看结构 owner、public API 和 guard，没有逐页重验所有交互行为。
- `src/api/contracts.ts` 体量也很大，但本次没有把它单独判成 finding，因为它更接近类型汇总事实源，是否拆分还取决于 OpenAPI/contract 生成策略。

## Touched Known-Debt Status

- 已知 debt 仍存在，但相较 2026-05 的”旧组件目录/route owner 混杂”阶段，当前 debt 已收敛到更具体的 contest domain 和 guard coverage 问题。
- 本次 review 结论：不阻塞当前总体架构方向，但后续只要继续碰 contest domain，应优先在 touched surface 内收掉上述两条 P1。

## Data Flow & Transport Boundary Analysis

### 竞赛域数据流现状

当前竞赛域的数据从 API 到 UI 经过三层 owner，每层都有不同程度的耦合残留：

```
route pages (pages/platform/contests/*)
  → compatibility barrel (features/platform/contests/index.ts)   ← P1-2
    → feature model (features/platform/contest-*/model/*)
      → API super-module (api/admin/contests.ts)                  ← P1-1
        → re-export bridge (api/teaching/awd-reviews)             ← 跨角色别名
```

关键耦合点：

| 耦合位置 | 当前状态 | 影响 |
| --- | --- | --- |
| route page → compatibility barrel | 5 个 route page 仍通过 `@/features/platform/contests` 导入 | contest-manage / contest-operations / contest-announcements 对 consumer 不是独立 slice |
| feature model → API super-module | 约 25 个 feature model 文件直接 import `@/api/admin` 或 `@/api/admin/contests` | 修改任意竞赛 API 签名都会波及多个 feature |
| `admin/contests.ts` → `teaching/awd-reviews` | 4 个 AWD review 函数通过 `import {...} from '../teaching/awd-reviews'` 桥接再重新暴露 | admin 和 teaching 的 review 边界在 transport 层未分离 |
| route page → API | `ContestManageRoutePage.vue` 等 route page 通过 feature model 间接依赖 API super-module | 链路长但未打破耦合；route page 切换 feature owner 时仍拖拽同一组 API 依赖 |

### 拆分后目标数据流

```
route pages (pages/platform/contests/*)
  → owning feature public API (features/platform/contest-manage/index.ts)
    → feature model
      → owning transport module (api/admin/contest-manage.ts)
        → shared contract types only (api/contracts.ts)

teaching 侧 review → api/teaching/awd-reviews.ts
admin 侧 review  → api/admin/contest-reviews.ts（独立 owner，不复用 teaching bridge）
```

## Error Handling Consistency

### 竞赛域错误处理现状

竞赛域 feature model 的错误处理存在以下不一致：

- **模式 A（静默 fallback）**：部分 contest feature model 在 API 调用失败后只 `console.error`，不向上抛出也不设置 error state，UI 层对失败无感知。
- **模式 B（全局 toast）**：部分 feature 直接调用 `useToast()` 展示通用错误消息，不区分 4xx / 5xx / network error。
- **模式 C（structured error state）**：少数 feature（如 `useContestEditPage`）在 model 内维护 `error` ref 并交给 UI 做 inline fallback，但这一模式未在竞赛域统一推行。

- Risk：
  - 用户在不同竞赛页面遇到同一类失败（如 429、网络断开）时，看到的反馈形式和可恢复能力不一致。
  - 静默 fallback 模式会让操作看起来”没反应”，用户可能重复点击导致并发问题。
- Recommendation：
  - 竞赛域统一错误处理契约：feature model 必须对外暴露 `error: Ref<ApiError | null>` 或等价的错误状态。
  - 全局 toast 仅用于跨页面的系统级通知，页面内可恢复错误优先走 inline fallback。
  - 在 contest feature model 的代码审查中，把”API 错误未向上传播”视为 blocking。

## Observability Gaps

- 前端当前没有统一的性能/错误埋点工具。竞赛域的关键操作（AWD readiness 编排、round 操作、scoreboard 拉取）耗时尚无结构化统计。
- `globalErrorRuntime.ts` 的 `console.error` 是当前唯一的 runtime error 记录方式，不适合在生产环境追踪问题。
- `request.ts` 拦截器携带 `request_id`，但前端未将其写入结构化日志或埋点系统，排查跨层问题时需要人工关联。

- Recommendation：
  - 短期：在 `shared/lib/reporting/` 下增加一个轻量 `useErrorReporter()`，统一接管 `ApiError` → 结构化日志的链路，至少包含 `code`、`status`、`requestId`、`requestUrl`、页面路由。
  - 中期：对 contest domain 的关键长操作（>1s）增加 `performance.mark` / `performance.measure`，至少在 dev 模式下可视化。

## ADR：竞赛域 API 拆分策略权衡

### 背景

`api/admin/contests.ts` 当前 1501 行，同时承接 AWD review archive、contest CRUD、announcement、round、readiness、service orchestration、traffic、scoreboard、attack log。本次 review 建议将其拆分为 5 个独立 transport owner。

### 方案对比

| 维度 | 方案 A：按 feature family 拆（推荐） | 方案 B：按 HTTP 资源路径拆 | 方案 C：保持超级模块 + 内部 barrel |
| --- | --- | --- | --- |
| 拆分粒度 | `contest-manage` / `contest-announcements` / `contest-operations` / `contest-awd-admin` / `contest-reviews` | `contests-crud` / `contests-teams` / `contests-awd` / `contests-scoreboard` / `contests-reports` | 不动文件，仅在 `admin/contests/` 下拆子文件并 barrel 聚合 |
| 与 feature 对齐度 | 高：每个 transport owner 精确对应一组 feature | 中：资源分组不一定映射到页面 workflow | 低：consumer 仍依赖同一个入口 |
| 跨角色 review 解耦 | 好：`contest-reviews` 独立，不复用 teaching bridge | 中：review 可能仍被归入 `contests-reports` | 差：无法打破 admin/teaching alias |
| blast radius 收口 | 最好：修改只影响 owning feature | 好：修改只影响同资源组的 feature | 差：与现状相同 |
| 迁移成本 | 中：需要逐 feature 切换 import 并验证 | 中：与方案 A 类似 | 低：几乎无 consumer 改动 |
| 长期可维护性 | 最好：owner 边界清晰 | 好：但可能与 feature 拆分节奏脱节 | 差：只是物理拆分，逻辑耦合不变 |

### 决策

选择方案 A（按 feature family 拆分），理由：

1. 当前 feature 目录已经按 `contest-manage` / `contest-operations` / `contest-announcements` / `contest-awd-admin` 拆开，transport 层对齐 feature 边界是最小认知负担的选择。
2. `contest-reviews` 独立后可消除 `admin/contests.ts → teaching/awd-reviews` 的桥接别名，role-neutral review 有独立 owner。
3. 拆分后每个 transport module 预计 150-400 行，体量可控，且可直接被对应的 feature model 消费。

### 不选方案 C 的理由

方案 C（只拆子文件保留 barrel）是”看起来改动小”的陷阱——它不能解决 consumer blast radius 问题，而且会延续”顺手往这个 barrel 加新接口”的习惯，最终回到当前状态。

## Concrete Convergence Plan

### 第 1 步：切断 contest compatibility barrel（P1-2）

- 目标：5 个 route page 全部直接 import owning feature。
- 改动面：
  - `ContestManageRoutePage.vue` → `@/features/platform/contest-manage`
  - `ContestOperationsRoutePage.vue` / `ContestOperationsHubRoutePage.vue` → `@/features/platform/contest-operations`
  - `ContestAnnouncementsRoutePage.vue` → `@/features/platform/contest-announcements`
  - `ContestEditRoutePage.vue` → `@/features/platform/contest-manage`
- 验证：`rg “@/features/platform/contests['\”]” src/` 返回空
- 后续：删除 `features/platform/contests/index.ts`，新增 guard 禁止同类兼容聚合层

### 第 2 步：拆分 contest API super-module（P1-1）

- 目标：`api/admin/contests.ts` 拆为 5 个 transport owner。
- 拆分映射：
  - `api/admin/contest-manage.ts`：contest CRUD、team、mode、status 相关接口
  - `api/admin/contest-announcements.ts`：announcement CRUD、publish
  - `api/admin/contest-operations.ts`：round、readiness、service orchestration
  - `api/admin/contest-awd-admin.ts`：AWD instance、traffic、scoreboard、attack log、checker
  - `api/admin/contest-reviews.ts`：AWD review archive 接口（不复用 teaching bridge，独立定义）
- 共享类型（DTO、enum、page result）保留在 `api/contracts.ts`。
- 逐 feature 切换 import：每切换一个 feature family 就跑对应测试和 typecheck。
- 验证：
  - `cd code/frontend && npm run typecheck`
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
  - `cd code/frontend && rg “from.*@/api/admin/contests” src/` 返回空（或仅剩未迁移的已知例外）

### 第 3 步：扩展 growth guard 覆盖（P2-1）

- 目标：growth guard 从”5 个文件的单点护栏”扩展到”当前 backlog P1/P2 owner 面”。
- 改动面：
  - 在 `frontend-growth-baseline.json` 中至少增加：
    - `features/platform/contest-manage/model/*` 关键 page model
    - `features/platform/contest-operations/model/*` 关键 page model
    - `features/contest-awd-admin/model/*` 已识别的大文件
    - `features/challenge-topology-studio/model/*`、`ui/*` 关键 surface
    - `features/awd-inspector/model/*` 关键 surface
  - 每个 entry 记录当前 baseline_lines，设定合理的 max_growth。
- 验证：`cd code/frontend && npm run check:frontend-growth`

### 第 4 步：竞赛域错误处理统一

- 目标：竞赛域 feature model 对外统一暴露 error state。
- 改动面：逐 feature model 检查 API 调用处的 catch 分支，确保：
  - 不为 null/undefined 时设置 `error` ref
  - 不为 `ApiError` 且 `status === 401` 时（由 global runtime 处理）不重复 toast
  - 可恢复错误（429 / network error）提供 retry 入口
- 验证：code review + contest domain 页面的手动回归

### 建议 commit 拆分

```
1. refactor(contest): 切断 route page 对 compatibility barrel 的依赖
2. refactor(contest): 拆分 contest-manage transport owner
3. refactor(contest): 拆分 contest-announcements transport owner
4. refactor(contest): 拆分 contest-operations transport owner
5. refactor(contest): 拆分 contest-awd-admin transport owner
6. refactor(contest): 提取 contest-reviews 独立 transport owner，解除 teaching bridge
7. refactor(contest): 删除 features/platform/contests compatibility barrel
8. chore(guard): 扩展 frontend-growth-baseline 覆盖 P1/P2 owner 面
9. chore(guard): 新增 feature barrel 禁止规则
```
