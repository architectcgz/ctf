# Backend Test Architecture Rewrite Blueprint

这份蓝图不是记录某一轮迁移步骤，而是回答一个更稳定的问题：如果按现在的认知重做这个项目，后端测试应该怎么布置，哪些测试该留在模块内，哪些测试不该再堆到 `internal/app`。

## 1. 目标

- 让测试目录和断言内容都按 owner 分层，而不是只把大文件拆短。
- 优先在模块内验证业务语义，把 full-router / runtime 测试收回到真正需要它们的地方。
- 让 `internal/app` 只保留 app composition 和少量兼容壳，不再承担系统测试主 owner。

## 2. 推荐分层

### A. 模块内测试：默认主战场

推荐继续放在各模块源码旁边：

- `internal/module/*/application/queries/*_test.go`
  - 测 query service 的筛选、分页、聚合、归档组装、只读 contract
- `internal/module/*/application/commands/*_test.go`
  - 测 command service 的状态流、校验、异步任务、builder、renderer、导出产物
- `internal/module/*/infrastructure/*_test.go`
  - 测 repository / mapper / persistence 细节

这一层应该覆盖绝大多数业务语义。像 AWD review：

- archive / selected round / filter / snapshot type 属于 `application/queries`
- export lifecycle / ZIP 字段保真 / PDF 渲染内容 属于 `application/commands`

### B. 共享测试基建：只放稳定复用能力

推荐两个落点：

- `internal/testutil/`
  - 必须贴近未导出实现、会被多个测试文件复用的 helper
- `code/backend/tests/testkit/`
  - 跨目录复用、但不依赖未导出符号的场景 builder / fixture / assert helper

判断规则：

- 需要碰未导出符号，就留 `internal/testutil`
- 不需要碰包私有实现、而且会跨专题复用，就进 `tests/testkit`

### C. 黑盒 HTTP 系统测试：放 `code/backend/tests/system/http`

这一层只测：

- 路由可达
- 权限矩阵
- handler 到 module wiring
- 端到端业务流
- 下载链路、状态流、响应 envelope 和最小结构信号

这一层不再负责：

- 复杂 archive 组装细节
- renderer 文案细节
- 某个 module builder 的完整分支语义

`full_router_*` 的长场景断言都应迁到这里；`internal/app` 只保留 glue code 和极少数 app-specific fixture。

### D. 运行时 / 外部依赖集成测试：放 `code/backend/tests/runtime`

这一层专门承接：

- PostgreSQL 真库行为
- runtime / practice / container / port / 外部进程参与的流程
- 宿主环境、容器网络、runtime agent 协同

不要把这类测试和纯 HTTP 场景继续混在 `internal/app`。

## 3. `internal/app` 应该剩下什么

如果重做，`internal/app` 最终只保留三类测试：

1. router/composition 结构测试
2. 必须访问 `package app` 私有 glue 的极薄兼容壳
3. 少量 app-level smoke，证明 app composition 没断

不应继续让 `internal/app` 持有：

- 长 HTTP 场景断言 owner
- 各业务专题自己的 seed / assert helper owner
- 大块 sqlite schema 清单
- 二进制导出内容的细节断言

## 4. 各层应该断什么

### 模块内

- 断真实业务语义
- 断导出字段、archive 内容、PDF/ZIP 结构
- 断 builder / mapper / renderer 的稳定 contract

### full-router / system http

- 断 HTTP 状态码、角色权限、响应 envelope、状态流
- 对二进制产物只断最小稳定结构信号
  - `%PDF`
  - `Content-Type`
  - `Content-Disposition`
  - 稳定可提取 token，例如字段名、payload token、section heading

不要在这一层直接绑定脆弱文案，例如历史标题、语言版本或排版措辞。

### runtime

- 断真实运行时行为
- 断 PostgreSQL / Redis / container 协作
- 断端口、恢复、调度、外部进程和宿主环境约束

## 5. 数据库策略

- 模块内语义测试默认可以继续用 SQLite，只要目标是业务语义，不依赖方言差异。
- 只要测试目标涉及 PostgreSQL 方言、锁、时区、JSONB、真实 migration 或运行时协作，就直接进 `tests/runtime`，不要强塞回 SQLite harness。
- 不要要求所有测试都切到 `ctf-postgres`；应该按风险面分层，而不是全量一刀切。

## 6. 二进制导出断言规则

- 模块内 renderer 测试负责更细的内容回归。
- 系统测试只断可稳定提取的结构信号。
- 同类 PDF helper 必须能力对齐；如果模块层支持 UTF-16 BE，集成层也必须支持。

这条是从 AWD review report 失败里直接抽出来的规则，后面所有 PDF/ZIP 测试都应该照这个执行。

## 7. 重做顺序

如果从头再来，建议顺序是：

1. 先把模块内 `queries` / `commands` 测全
2. 再抽共享 testkit
3. 再把长场景迁到 `tests/system/http`
4. 最后才清理 `internal/app` 里的兼容壳

不要反过来先搬目录。先搬目录通常只会得到“文件位置变了，但 owner 没变”。

## 8. 当前项目的最终目标形态

- 模块语义测试以 `internal/module/*` 为主
- `internal/testutil/` 和 `tests/testkit/` 只承接稳定复用能力
- `tests/system/http` 承接黑盒 HTTP 场景
- `tests/runtime` 承接 PostgreSQL / runtime / container 集成
- `internal/app` 退回 app composition 与最小兼容壳

## 交叉链接

- `code/backend/tests/README.md`
- `docs/plan/impl-plan/2026-06-03-backend-test-architecture-phase1-plan.md`
- `docs/plan/impl-plan/2026-06-03-backend-test-architecture-phase10-plan.md`
- `feedback/2026-06-03-backend-test-layering-and-pdf-assertion-ownership.md`
