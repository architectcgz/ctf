# CTF 前端 Feature-Sliced Architecture 迁移台账

## 结论

这份台账从 2026-05-31 当前代码状态重新整理，旧的 `views / composables` 迁移记录已退场，不再作为后续判断依据。

当前前端主骨架已经稳定在：

```text
src/
  pages/
  features/
  entities/
  widgets/
  shared/
```

后续重点不是继续做目录级搬迁，而是把现有边界收紧到一致可维护的状态。

## 当前状态

### 已经收口的部分

- 路由主入口已经以 `pages/*RoutePage.vue` 为主，学生端、教师端、平台端的大多数 runtime route 都走 `pages` 层。
- `src/views`、`src/composables` 已经退出当前运行时主链路，不再是现状事实。
- 复杂页面流程已经有稳定 feature owner，例如：
  - `features/challenge-detail/model/useChallengeDetailPage.ts`
  - `features/contest-detail/model/useContestDetailRoutePage.ts`
  - `features/contest-projector/model/useContestProjectorPage.ts`
  - `features/challenge-list/model/useChallengeListPage.ts`
- `entities/challenge` 已经开始承接稳定展示规则，平台题目管理和学生题目详情都在复用这层能力。
- `widgets/awd-review-workspace`、`widgets/platform-challenge-detail`、`widgets/review-archive-workspace` 已经承担页面级组合区块，而不是把所有编排继续压回 route page。

### 本轮新增收口

- `platform/challenges` 路由入口补回标准 `pages` 层：
  - `pages/platform/challenges/ChallengeManageRoutePage.vue`
- router guardrail 改成限制运行时组件入口只能指向 `pages`，唯一例外是 app shell 布局：
  - `shared/ui/layout/AppLayout.vue`
- `NotificationDetailRoutePage.vue` 已收口成 route entry + widget 组合：
  - `pages/notifications/NotificationDetailRoutePage.vue`
  - `widgets/notification-detail-workspace/NotificationDetailWorkspace.vue`
- `NotificationListRoutePage.vue` 已收口成 route entry + widget 组合：
  - `pages/notifications/NotificationListRoutePage.vue`
  - `widgets/notification-list-workspace/NotificationListWorkspace.vue`
- `ContestDetailRoutePage.vue` 已收口成 route entry + widget 组合：
  - `pages/contests/ContestDetailRoutePage.vue`
  - `widgets/contest-detail-workspace/ContestDetailWorkspace.vue`
- `ContestListRoutePage.vue` 已收口成 route entry + widget 组合：
  - `pages/contests/ContestListRoutePage.vue`
  - `widgets/contest-list-workspace/ContestListWorkspace.vue`
- `ScoreboardDetailRoutePage.vue` 已收口成 route entry + widget 组合：
  - `pages/scoreboard/ScoreboardDetailRoutePage.vue`
  - `widgets/scoreboard-detail-workspace/ScoreboardDetailWorkspace.vue`
- `ChallengeDetailRoutePage.vue` 已收口成 route entry + widget 组合：
  - `pages/challenges/ChallengeDetailRoutePage.vue`
  - `widgets/challenge-detail-workspace/ChallengeDetailWorkspace.vue`

## 当前做得好的地方

- `features/*/model` 已经成为主要业务流程 owner，不再把异步 workflow 默认压在 route page。
- `pages` 层已经有独立边界测试，限制 route page 不直接碰业务 API、`useRoute`、`useRouter` 和旧 query tab hook。
- `architectureBoundaries.test.ts`、`featureBoundaries.test.ts` 已经把一部分结构规则机械化，不再只靠约定。
- `shared/model/navigation/*Transport` 这类路由 transport 已经形成可复用模式，route-aware feature 不需要再直接 import `vue-router`。

## 当前还需要继续迁移的工作

### 1. `contest` domain 仍是当前最重的结构 debt

- 当前 route page / feature page shell 已经基本薄化，但 `contest` 这条线还保留两类活动兼容层：
  - `features/platform/contests/index.ts`
    - 当前只是兼容重导出层，但 `ContestManageRoutePage.vue`、`ContestEditRoutePage.vue`、`ContestOperationsRoutePage.vue`、`ContestOperationsHubRoutePage.vue`、`ContestAnnouncementsRoutePage.vue` 与 `useContestEditPage.ts` 仍直接依赖它。
  - `api/admin/contests.ts`
    - 当前同时承接 contest CRUD、announcement、AWD round、readiness、service orchestration、traffic、scoreboard、attack log、review archive export，是实际运行时里的超级 API owner。
- 这说明前端当前最大的剩余问题已经不是 `views/components` 目录迁移，而是：
  - feature public API 还没有在 `contest` 这条线上彻底收口
  - transport owner 还没有跟着 feature split 一起拆开
- 后续优先级应先处理这条线，再继续扩展 entity 补强。

### 2. 教师端兼容命名空间已基本退场

- 当前产品规则要求教师端只使用 `/academy/*`。
- router runtime 已不再注册 `/teacher/* -> /academy/*` 页面 redirect。
- 登录 redirect 参数已经不再接受 `/teacher/*` 旧页面路径。
- 前端活跃源码里已经不再保留 `/teacher/*` 前端页面路径兼容 owner。
- 登录 redirect 参数遇到旧教师端页面路径时，会直接回退到角色默认首页，而不是再 canonicalize 成 `/academy/*`。

### 3. route page 厚壳层主问题已基本收口

当前主链路里的高优先级 runtime route page 已基本收口成 route entry + widget / feature 组合。

后续重点更适合转向：

- `entities` 展示 owner 补强
- 继续把 guardrail 跟现状一起机械化

### 4. `entities` 仍然偏薄

当前 `challenge`、`notification`、`contest`、`team` 都已经形成实体层入口，但下面这些对象还没有形成同等稳定的展示规则 owner：

- `user`
- `instance`

其中 `user` 的显示名 / 用户名 handle / option label owner 已经收口到 `entities/user`，当前已覆盖 student dashboard、review archive、teacher dashboard、teacher student management、class students workspace、student analysis review、platform user governance、skill profile teacher select 和 admin notification publish drawer。当前运行时剩余的用户展示散点主要嵌在 `instance` 相关页面里，后续按 `instance` 对象边界一起处理更合适。

后续优先迁入：

- 状态文案和 badge 映射
- 业务对象 meta 展示
- 单对象稳定 view model

不在这层处理：

- 路由跳转
- toast / confirm
- API mutation
- 轮询和跨模块 workflow

### 5. 边界测试需要继续跟着现状更新

当前 guardrail 已经有效，但还需要继续补：

- 教师端旧兼容路由的退场约束
- 厚 route page 的行数和 owner 约束
- 新 entity / widget 抽取后的 import public API 约束

## 推荐优先级

### P0：保持已收口入口不回退

- 保持教师端页面 runtime 只认 `/academy/*`，不再回流 `/teacher/*` route。
- 继续用源码级 guardrail 限制新的 `/teacher/*` 前端页面路径 producer；当前前端已不再保留这类页面兼容 owner。
- 保持 router runtime 只从 `pages` 取页面组件，不再新增绕过 `pages` 的入口。

### P1：先收口 `contest` domain 的 public API 与 transport owner

优先顺序：

- 清退 `features/platform/contests/index.ts` 这类活动兼容 barrel，让 route page 和 feature consumer 回到真实 owning feature。
- 拆分 `api/admin/contests.ts`，至少把：
  - contest manage
  - contest announcements
  - contest operations / AWD admin
  - contest review archive
  分回更窄的 API client owner。
- 在这条线收口后，再补对应 guardrail，防止兼容 barrel 与超级 API owner 回流。

### P2：补 route/widget/entity 边界 guardrail

重点不是继续找新的厚 route page，而是防止已收口页面回退：

- route page 继续只保留入口、装配和少量局部桥接
- workspace 区块继续留在 widget
- 展示规则继续下沉到 entity
- 业务流程继续留在 feature model

### P3：补强实体层

`notification` 的类型文案、accent 和已读状态展示已收口到 `entities/notification`；`contest` 的状态 / 模式 / CTA / accent / status badge class 已进一步收口到 `entities/contest`；`team` 的成员数、队长关系、邀请码文案和成员展示项也已经收口到 `entities/team`；`user` 的主消费面已经完成收口，下一步转向 `instance` 实体层，不做一口气大搬迁。

### P4：同步文档与 guardrail

- 每完成一个结构切片，就更新这份台账，不再保留失效目录名和历史迁移说法。
- 结构收口完成后，再决定是否需要把这份台账继续拆成事实文档和 backlog。

## 验收口径

- route runtime 入口统一走 `pages`
- route page 不直接 import 业务 API
- route-aware feature 不直接 import `vue-router`
- 页面壳层变薄时，不把业务动作重新散回 widget / entity
- 每次结构迁移至少补一条源码级边界测试

## 本次记录依据

- `src/router/routes/*.ts`
- `src/pages/**/*RoutePage.vue`
- `src/features/**/model/*.ts`
- `src/entities/challenge/*`
- `src/widgets/*`
- `src/__tests__/routePageArchitectureBoundary.test.ts`
- `src/__tests__/architectureBoundaries.test.ts`
- `src/features/__tests__/featureBoundaries.test.ts`
- `src/api/admin/contests.ts`
- `src/features/platform/contests/index.ts`

规则：

- `shared` 不允许 import `features`、`widgets`、`pages`。（`2026-06-01` 已收口：原 `shared/model/layout/*Bridge.ts` 迁移至 `features/layout/`，AppShellRoutePage 组装层接线，shared 对 features 零依赖。）
- `shared` 不写业务 API 流程。
- `shared` 内只放跨业务通用能力。

## 候选演进方向（非当前事实）

下面这些只作为后续可能采用的落点，不代表当前目录事实：

- 如果未来需要把基础 transport helper 继续中性化，可以再评估是否引入更明确的 `shared/api` owner。
- 如果主题、导航 meta、路由配置继续增长，可以再评估是否把当前 `src/config` 分拆到更清晰的 `shared/config` 或 `app/config`。
- 在没有明确 consumer、guardrail 和迁移收益前，不先做这类目录级重排。

## 每次迁移的验收标准

- 路由页没有直接 import 本次迁移相关的业务 API。
- 业务异步流程集中在 `features/<slice>/model`。
- feature 通过 `index.ts` 暴露公共入口，外部不深链到内部文件。
- 原有用户行为不变。
- 聚焦测试通过。
- `npm run typecheck` 通过。
- 如改动页面结构，补或更新对应源码级架构测试。

## 不采用的方向

- 不一次性把 `views` 全部搬到 `pages`。
- 不一次性把 `components` 全部搬到 `widgets/entities/shared`。
- 不给每个 feature 强行拆 presentation/application/domain/infrastructure。
- 不把 toast、router、confirm 这类 UI workflow 当成 domain。
- 不为了目录漂亮改动大量 import。

## 后续记录方式

每完成一个迁移单元，在这里追加一条：

- 已完成：shared→features 边界收口。三个 `*Bridge.ts` 从 `shared/model/layout/` 迁至 `features/layout/model/`；`AppShellRoutePage.vue` 接管组装接线；`AppLayout`/`TopNav`/`NotificationDrawer` 改为 prop 驱动；policy JSON 清退残留 legacy 目录并收紧 shared 层 `forbidden_import_layers`。

```text
- 已完成：`views/...` 的页面状态下沉到 `features/...`，页面保留组件编排，并补充源码级边界测试。
```
