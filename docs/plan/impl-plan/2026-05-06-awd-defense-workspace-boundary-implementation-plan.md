# AWD Defense Workspace Boundary Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 AWD 防守从“学生直接进 service 容器 + 浏览器文件工作台 + `editable_paths` 提示”迁移到“独立 defense workspace + 目录级挂载边界 + 学生端只保留连接入口”的正式实现。

**Architecture:** 题包先切到 `runtime / workspace / check` 三层结构，并在 `challenge.yml` 中用 `defense_workspace` 描述目录级工作区边界；运行时再为每个 `contest + team + service` 维护独立 workspace 状态、共享业务 roots 和 companion workspace 容器。学生侧 battle 页只消费工作区入口状态，不再消费 `defense_scope` 文件提示，也不再暴露浏览器文件树、在线编辑或命令执行接口。

**Tech Stack:** Go、Gin、GORM、PostgreSQL migrations、Docker Engine、Vue 3、Vite、Vitest。

---

## 输入文档

- `docs/architecture/features/awd-defense-workspace-boundary-design.md`
- `docs/architecture/features/awd-student-battle-workspace-design.md`
- `challenges/awd/challenge-package-contract.md`
- `docs/architecture/features/awd-final-design.md`

## 目标边界

- 标准 AWD 竞赛模式只保留战场态势、攻击入口、服务重启和 defense workspace 连接入口。
- 学生侧不再暴露 `defense_scope`、`editable_paths`、文件树、浏览器编辑器或浏览器命令执行入口。
- `restart` 只重建 runtime，不清空 workspace；`reseed / recreate` 才会重置 workspace 并轮换 workspace revision。
- 运行中赛事不做题包 revision 热覆盖；新 snapshot 只在显式 reprovision 或新赛事发放时切换。

## 当前实现事实

- SSH 网关当前直接连到实例 `container_id`，并默认执行 `/bin/sh`：
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
- 学生侧 runtime API 当前仍暴露浏览器文件和命令接口：
  - `code/backend/internal/app/router_routes.go`
  - `code/backend/internal/module/runtime/api/http/handler.go`
  - `code/backend/internal/module/runtime/runtime/adapters.go`
- 学生态 AWD workspace DTO 当前仍包含 `defense_scope`：
  - `code/backend/internal/module/contest/application/queries/awd_workspace_result.go`
  - `code/backend/internal/dto/contest_awd_workspace.go`
  - `code/backend/internal/module/contest/application/queries/awd_workspace_query.go`
- 题包 parser 当前仍强依赖文件级 `defense_scope.editable_paths`：
  - `code/backend/internal/module/challenge/domain/awd_package_parser.go`
- 前端当前仍存在浏览器防守工作台和单独路由：
  - `code/frontend/src/views/contests/ContestAWDDefenseWorkbench.vue`
  - `code/frontend/src/features/contest-awd-workspace/model/useContestAwdDefenseWorkbenchPage.ts`
  - `code/frontend/src/router/routes/studentRoutes.ts`

## 文件结构

### 题包契约与样例

- Modify: `code/backend/internal/module/challenge/domain/awd_package_parser.go`
- Modify: `code/backend/internal/module/challenge/domain/awd_package_parser_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service_test.go`
- Modify sample packages:
  - `challenges/awd/ctf-1/awd-campus-drive/challenge.yml`
  - `challenges/awd/ctf-1/awd-iot-hub/challenge.yml`
  - `challenges/awd/ctf-1/awd-supply-ticket/challenge.yml`
  - `challenges/awd/ctf-1/awd-tcp-length-gate/challenge.yml`
  - each package’s `docker/` tree should be migrated to `runtime/`, `workspace/`, `check/`

### Workspace 状态与仓储

- Create: `code/backend/internal/model/awd_defense_workspace.go`
- Create: `code/backend/migrations/000006_create_awd_defense_workspaces.up.sql`
- Create: `code/backend/migrations/000006_create_awd_defense_workspaces.down.sql`
- Modify: `code/backend/internal/module/runtime/infrastructure/repository.go`
- Modify: `code/backend/internal/module/runtime/ports/http.go`
- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Modify: `code/backend/internal/module/contest/testsupport/db.go`

### 运行时装配与重启语义

- Modify: `code/backend/internal/model/container.go`
- Modify: `code/backend/internal/module/runtime/ports/topology.go`
- Modify: `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- Modify: `code/backend/internal/module/runtime/infrastructure/engine.go`
- Modify: `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
- Modify: `code/backend/internal/module/practice/application/commands/instance_provisioning.go`
- Modify: `code/backend/internal/module/practice/application/commands/instance_start_service.go`
- Modify relevant tests:
  - `code/backend/internal/module/runtime/service_test.go`
  - `code/backend/internal/module/practice/application/commands/runtime_container_create_test.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
  - `code/backend/internal/module/practice/application/commands/instance_start_service_test.go`

### SSH 网关与学生态 runtime API

- Modify: `code/backend/internal/module/runtime/application/queries/proxy_ticket_service.go`
- Modify: `code/backend/internal/module/runtime/application/proxy_ticket_service_test.go`
- Modify: `code/backend/internal/module/runtime/infrastructure/awd_target_proxy_repository.go`
- Modify: `code/backend/internal/module/runtime/infrastructure/awd_target_proxy_repository_test.go`
- Modify: `code/backend/internal/module/runtime/runtime/adapters.go`
- Modify: `code/backend/internal/module/runtime/runtime/module.go`
- Modify: `code/backend/internal/module/runtime/api/http/handler.go`
- Modify: `code/backend/internal/module/runtime/api/http/handler_test.go`
- Modify: `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
- Modify: `code/backend/internal/app/router_routes.go`
- Modify: `code/backend/internal/app/router_test.go`

### Contest workspace read model 与 DTO

- Modify: `code/backend/internal/module/contest/ports/awd.go`
- Modify: `code/backend/internal/module/contest/infrastructure/awd_contest_relation_repository.go`
- Modify: `code/backend/internal/module/contest/application/queries/awd_workspace_query.go`
- Modify: `code/backend/internal/module/contest/application/queries/awd_workspace_result.go`
- Modify: `code/backend/internal/module/contest/application/queries/awd_service_test.go`
- Modify: `code/backend/internal/dto/contest_awd_workspace.go`
- Modify: `code/backend/internal/module/contest/api/http/request_mapper.go`
- Modify generated mapping output if the project keeps it checked in:
  - `code/backend/internal/module/contest/api/http/request_mapper_gen.go`

### 前端 battle 页清理

- Modify: `code/frontend/src/api/contracts.ts`
- Modify: `code/frontend/src/api/contest.ts`
- Modify: `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- Modify: `code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.ts`
- Modify: `code/frontend/src/router/routes/studentRoutes.ts`
- Delete: `code/frontend/src/views/contests/ContestAWDDefenseWorkbench.vue`
- Delete: `code/frontend/src/features/contest-awd-workspace/model/useContestAwdDefenseWorkbenchPage.ts`
- Modify/remove affected tests:
  - `code/frontend/src/api/__tests__/contest.test.ts`
  - `code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`
  - `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
  - `code/frontend/src/views/contests/__tests__/ContestAWDDefenseWorkbench.test.ts`
  - `code/frontend/src/features/contest-awd-workspace/model/useContestAwdDefenseWorkbenchPage.test.ts`

## 回退与恢复约定

- 数据库迁移先只新增 `awd_defense_workspaces` 表，不立刻删老字段，保证 parser / DTO / API 可以分阶段切换。
- 后端先完成 workspace 记录、runtime 挂载和 SSH scope 切换，再删除学生态浏览器文件接口，避免前后端半切换导致页面不可用。
- 样例题包迁移必须和 parser 切换同一批完成，避免仓库内 sample package 立即失效。

### Task 1: 目录级题包契约与样例迁移

**Files:**

- Modify: `code/backend/internal/module/challenge/domain/awd_package_parser.go`
- Modify: `code/backend/internal/module/challenge/domain/awd_package_parser_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service_test.go`
- Modify sample packages under:
  - `challenges/awd/ctf-1/awd-campus-drive/`
  - `challenges/awd/ctf-1/awd-iot-hub/`
  - `challenges/awd/ctf-1/awd-supply-ticket/`
  - `challenges/awd/ctf-1/awd-tcp-length-gate/`

**Review focus:** parser 是否完全切到目录级 `defense_workspace`，以及样例题包是否已经把 runtime/workspace/check 物理拆开。

- [x] **Step 1: 先写失败测试，锁定新契约**

在 `awd_package_parser_test.go` 中补覆盖：

- 目录级 `defense_workspace` happy path
- `workspace_roots` / `writable_roots` / `readonly_roots` 互斥与覆盖关系
- 拒绝 `docker/runtime`、`docker/check`、`challenge.yml` 进入工作区
- 拒绝单文件 `docker/challenge_app.py` 作为新契约主边界

- [x] **Step 2: 迁移仓库内 AWD 样例题包**

把四个样例包的 `docker/` 目录改成 `runtime/`, `workspace/`, `check/` 三层，并更新 `challenge.yml` 为 `defense_workspace` 目录级配置。

- [x] **Step 3: 更新 parser 和导入校验**

在 `awd_package_parser.go` 中新增目录级 `defense_workspace` 解析与校验，停止把 `defense_scope.editable_paths` 当成正式学生边界输入；`defense_scope` 只保留内部保护元数据角色。

- [x] **Step 4: 更新 import service 测试夹具**

把 `awd_challenge_import_service_test.go` 中依赖旧 `defense_scope` 的 fixture、断言和错误消息切到新契约。

- [x] **Step 5: 跑题包解析与导入测试**

Run:

```bash
cd code/backend
go test ./internal/module/challenge/domain ./internal/module/challenge/application/commands -run 'AWD|Package|Import' -count=1
```

Expected: PASS

- [x] **Step 6: Commit**

```bash
git add code/backend/internal/module/challenge challenges/awd
git commit -m "feat(awd): 切换目录级 defense workspace 契约"
```

### Task 2: 建立 defense workspace 状态模型

**Files:**

- Create: `code/backend/internal/model/awd_defense_workspace.go`
- Create: `code/backend/migrations/000006_create_awd_defense_workspaces.up.sql`
- Create: `code/backend/migrations/000006_create_awd_defense_workspaces.down.sql`
- Modify: `code/backend/internal/module/runtime/infrastructure/repository.go`
- Modify: `code/backend/internal/module/runtime/ports/http.go`
- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Modify: `code/backend/internal/module/contest/testsupport/db.go`

**Review focus:** workspace 状态是否从实例 runtime 状态中独立出来，能否表达 `workspace revision`、container id 和 reseed 生命周期。

- [x] **Step 1: 先定义迁移和唯一键**

新增 `awd_defense_workspaces`，至少包含：

- `contest_id`
- `team_id`
- `service_id`
- `instance_id`
- `workspace_revision`
- `status`
- `container_id`
- `seed_signature` 或等价 snapshot 标识
- `created_at / updated_at`

并为 `contest_id + team_id + service_id` 建唯一约束。

- [x] **Step 2: 写 model 和 ports**

新增 `AWDDefenseWorkspace` model，并把 runtime / practice 所需的读写接口补到对应 ports 中。

- [x] **Step 3: 实现仓储 CRUD**

在 runtime repository 中实现：

- 查找当前队伍服务的 workspace scope
- upsert workspace 记录
- bump `workspace_revision`
- runtime restart 与 reseed 时的状态更新

- [x] **Step 4: 补仓储与 AutoMigrate 测试**

为仓储行为补测试，并把新 model 接入 `contest/testsupport/db.go`，保证现有 AWD 测试基线可继续起库。

- [x] **Step 5: 跑状态模型相关测试**

Run:

```bash
cd code/backend
go test ./internal/model ./internal/module/runtime/infrastructure ./internal/module/contest/testsupport -run 'AWDDefenseWorkspace|AutoMigrate' -count=1
```

Expected: PASS

- [x] **Step 6: Commit**

```bash
git add code/backend/internal/model code/backend/migrations code/backend/internal/module/runtime code/backend/internal/module/practice/ports code/backend/internal/module/contest/testsupport
git commit -m "feat(awd): 增加 defense workspace 状态模型"
```

### Task 3: 实现 root 粒度挂载与 companion workspace provisioning

**Files:**

- Modify: `code/backend/internal/model/container.go`
- Modify: `code/backend/internal/module/runtime/ports/topology.go`
- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Modify: `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- Modify: `code/backend/internal/module/runtime/infrastructure/engine.go`
- Modify: `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
- Modify: `code/backend/internal/module/practice/application/commands/instance_provisioning.go`
- Modify: `code/backend/internal/module/practice/application/commands/instance_start_service.go`
- Modify tests:
  - `code/backend/internal/module/runtime/service_test.go`
  - `code/backend/internal/module/practice/application/commands/runtime_container_create_test.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
  - `code/backend/internal/module/practice/application/commands/instance_start_service_test.go`

**Review focus:** runtime restart 是否只重建 service runtime，workspace roots 和 workspace container 是否默认保留。

- [x] **Step 1: 先写 mount-aware 失败测试**

补测试覆盖：

- `ContainerConfig` / `TopologyCreateNode` 支持目录挂载
- create/provision 路径会额外创建 workspace container
- restart 不清空 workspace revision 和 workspace roots

- [x] **Step 2: 给 runtime container config 增加挂载能力**

在 `container.go`、`topology.go` 和 `practice/ports.go` 增加 mount 描述结构，支持 `rw / ro` 模式和 companion node 不对外发布端口。

- [x] **Step 3: 实现 Docker engine 与 provisioning service 的目录挂载**

在 `engine.go` 和 `provisioning_service.go` 中把 mount 配置真正映射到 Docker create request，并支持创建仅供 SSH 使用的 workspace container。

- [x] **Step 4: 接入 practice 实例生命周期**

在 `runtime_container_create.go`、`instance_provisioning.go`、`instance_start_service.go` 中：

- 首次 provision 时初始化 workspace roots 和 workspace container
- 普通 restart 仅 cleanup/recreate runtime container
- `reseed / recreate` 路径才 bump workspace revision 并重建 workspace container

- [x] **Step 5: 跑 provisioning 与 restart 测试**

Run:

```bash
cd code/backend
go test ./internal/module/runtime/... ./internal/module/practice/application/commands -run 'DefenseWorkspace|Provision|CreateTopology|RestartContestAWDService' -count=1
```

Expected: PASS

- [x] **Step 6: Commit**

```bash
git add code/backend/internal/model/container.go code/backend/internal/module/runtime code/backend/internal/module/practice/application/commands code/backend/internal/module/practice/ports
git commit -m "feat(awd): 实现 defense workspace companion provisioning"
```

### Task 4: 把 SSH 网关切到 workspace scope，并下线学生态浏览器防守接口

**Files:**

- Modify: `code/backend/internal/module/runtime/application/queries/proxy_ticket_service.go`
- Modify: `code/backend/internal/module/runtime/application/proxy_ticket_service_test.go`
- Modify: `code/backend/internal/module/runtime/infrastructure/awd_target_proxy_repository.go`
- Modify: `code/backend/internal/module/runtime/infrastructure/awd_target_proxy_repository_test.go`
- Modify: `code/backend/internal/module/runtime/runtime/adapters.go`
- Modify: `code/backend/internal/module/runtime/runtime/module.go`
- Modify: `code/backend/internal/module/runtime/api/http/handler.go`
- Modify: `code/backend/internal/module/runtime/api/http/handler_test.go`
- Modify: `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
- Modify: `code/backend/internal/app/router_routes.go`
- Modify: `code/backend/internal/app/router_test.go`

**Review focus:** 学生侧是否已经失去浏览器文件/命令入口，且 SSH 票据只会指向当前 workspace revision。

- [x] **Step 1: 先写 scope 与路由失败测试**

补测试覆盖：

- `FindAWDDefenseSSHScope` 返回 workspace container id 而不是 runtime `container_id`
- SSH ticket 绑定 `workspace_revision`
- 学生路由不再注册 `defense/files`、`defense/directories`、`defense/commands`

- [x] **Step 2: 改 proxy ticket 和 repository scope**

让 proxy ticket 在签发和鉴权时读取 `workspace_revision`，repository scope 返回 workspace container 及其状态。

- [x] **Step 3: 改 SSH gateway**

在 `awd_defense_ssh_gateway.go` 中：

- scope lookup 目标改成 workspace container
- 默认工作目录切到 `/workspace`
- 拒绝 runtime container fallback

- [x] **Step 4: 删除学生态浏览器防守接口**

从 `runtime` handler / adapter / router 中移除或硬禁用：

- `GET /defense/files`
- `GET /defense/directories`
- `PUT /defense/files`
- `POST /defense/commands`

只保留 `POST /defense/ssh`。

- [x] **Step 5: 跑 runtime API 与 router 测试**

Run:

```bash
cd code/backend
go test ./internal/module/runtime/... ./internal/app -run 'DefenseSSH|ProxyTicket|Router' -count=1
```

Expected: PASS

- [x] **Step 6: Commit**

```bash
git add code/backend/internal/module/runtime code/backend/internal/app
git commit -m "feat(awd): 切换 defense ssh 到 workspace scope"
```

### Task 5: 清理学生态 AWD workspace DTO，去掉 defense_scope 文件提示

**Files:**

- Modify: `code/backend/internal/module/contest/ports/awd.go`
- Modify: `code/backend/internal/module/contest/infrastructure/awd_contest_relation_repository.go`
- Modify: `code/backend/internal/module/contest/application/queries/awd_workspace_query.go`
- Modify: `code/backend/internal/module/contest/application/queries/awd_workspace_result.go`
- Modify: `code/backend/internal/module/contest/application/queries/awd_service_test.go`
- Modify: `code/backend/internal/dto/contest_awd_workspace.go`
- Modify: `code/backend/internal/module/contest/api/http/request_mapper.go`
- Modify if generated mapper is committed:
  - `code/backend/internal/module/contest/api/http/request_mapper_gen.go`

**Review focus:** 学生态 workspace 返回的是否只剩“连接入口状态”和 battle 必需事实，不再夹带任何文件级修补线索。

- [ ] **Step 1: 先写 query/DTO 失败测试**

在 `awd_service_test.go` 中改断言：

- 不再出现 `defense_scope`
- 自己队伍的服务会返回 `defense_connection` 或等价 workspace 摘要
- 未入队学生仍看不到目标目录

- [ ] **Step 2: 改 ports 和 infrastructure**

把 `AWDServiceDefinition` / repository 映射从 `DefenseScope` 改成最小 `DefenseWorkspaceSummary`，字段只保留：

- `entry_mode`
- `workspace_status`
- `workspace_revision` 或 `updated_at`

- [ ] **Step 3: 改 query result、DTO 和 HTTP mapper**

从 `awd_workspace_result.go`、`contest_awd_workspace.go` 和 mapper 中移除 `editable_paths`、`protected_paths`、`service_contracts`，改成新的连接摘要结构。

- [ ] **Step 4: 跑 contest workspace 查询测试**

Run:

```bash
cd code/backend
go test ./internal/module/contest/application/queries ./internal/module/contest/api/http -run 'GetUserWorkspace|AWDWorkspace' -count=1
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add code/backend/internal/module/contest code/backend/internal/dto
git commit -m "feat(awd): 清理学生态 defense workspace DTO"
```

### Task 6: 前端 battle 页改成“连接入口”模式，并删除浏览器 workbench

**Files:**

- Modify: `code/frontend/src/api/contracts.ts`
- Modify: `code/frontend/src/api/contest.ts`
- Modify: `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- Modify: `code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.ts`
- Modify: `code/frontend/src/router/routes/studentRoutes.ts`
- Delete: `code/frontend/src/views/contests/ContestAWDDefenseWorkbench.vue`
- Delete: `code/frontend/src/features/contest-awd-workspace/model/useContestAwdDefenseWorkbenchPage.ts`
- Modify or delete affected tests:
  - `code/frontend/src/api/__tests__/contest.test.ts`
  - `code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`
  - `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
  - `code/frontend/src/views/contests/__tests__/ContestAWDDefenseWorkbench.test.ts`
  - `code/frontend/src/features/contest-awd-workspace/model/useContestAwdDefenseWorkbenchPage.test.ts`

**Review focus:** 学生侧是否只剩 battle panel 内的 SSH 连接信息展示，不再有“打开防守文件页”的路由和 API 调用。

- [ ] **Step 1: 先改测试断言**

让前端测试先表达新行为：

- 不再存在 `ContestAWDDefenseWorkbench` route
- battle panel 不再调用 `requestContestAWDDefenseDirectory` / `requestContestAWDDefenseFile` / `requestContestAWDDefenseFileSave`
- defense 区只依赖 workspace summary + `requestContestAWDDefenseSSH`

- [ ] **Step 2: 更新 API contracts**

把 `ContestAWDWorkspaceServiceData` 从 `defense_scope` 切换到新的连接摘要字段，同时删除浏览器 defense file API 类型依赖。

- [ ] **Step 3: 清理 battle panel 与展示模型**

在 `ContestAWDWorkspacePanel.vue` 和 `awdDefensePresentation.ts` 中：

- 去掉“打开防守工作台”按钮
- 保留 SSH 命令、连接状态和 restart 反馈
- 保证页面仍然能快速打开本队服务和目标服务

- [ ] **Step 4: 删除路由和 workbench 文件**

删掉独立 `ContestAWDDefenseWorkbench` 页面、对应 composable 和学生路由注册；同步清理相关测试。

- [ ] **Step 5: 跑前端 AWD 相关测试**

Run:

```bash
cd code/frontend
npx vitest run src/api/__tests__/contest.test.ts src/views/contests/__tests__/ContestDetail.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add code/frontend/src
git commit -m "feat(awd): 收敛学生侧 defense workspace 入口"
```

### Task 7: 集成验证、文档回收与独立 review

**Files:**

- Modify if implementation differs from design:
  - `docs/architecture/features/awd-defense-workspace-boundary-design.md`
  - `docs/architecture/features/awd-student-battle-workspace-design.md`
  - `challenges/awd/challenge-package-contract.md`
- Create review archive:
  - `docs/reviews/architecture/2026-05-06-awd-defense-workspace-review.md`

**Review focus:** 设计事实源、实现计划和最终代码是否仍然对齐，是否还遗留任何学生可见的防守线索接口。

- [ ] **Step 1: 跑后端集成验证**

Run:

```bash
cd code/backend
go test ./internal/module/challenge/... ./internal/module/practice/... ./internal/module/runtime/... ./internal/module/contest/... ./internal/app -run 'AWD|Defense|Workspace|Router' -count=1
```

Expected: PASS

- [ ] **Step 2: 跑前端集成验证**

Run:

```bash
cd code/frontend
npx vitest run src/api/__tests__/contest.test.ts src/views/contests/__tests__ src/features/contest-awd-workspace
```

Expected: PASS

- [ ] **Step 3: 跑 harness 一致性检查**

Run:

```bash
cd /home/azhi/workspace/projects/ctf
bash scripts/check-consistency.sh
```

Expected: PASS

- [ ] **Step 4: 写独立 review 归档**

按 `docs/reviews/architecture/` 约定记录：

- student-facing hints 是否已全部收口
- restart / reseed 语义是否符合设计
- workspace revision 与 SSH ticket 是否有遗漏

- [ ] **Step 5: Commit**

```bash
git add docs/reviews/architecture docs/architecture/features challenges/awd/challenge-package-contract.md
git commit -m "docs(awd): 归档 defense workspace review 结论"
```

## 计划自检

- 目标边界是否明确：是。题包契约、workspace 状态、runtime 装配、学生 API、前端入口都分别落到了具体 owner 文件。
- 是否只修输出行为、却把结构收敛留到后面：否。计划把 `defense_workspace`、workspace 状态表、mount 模型和前端 route 清理放在同一流水线里，没有把结构问题留成“后续再拆”。
- 是否命中了已知债面却没有收口：没有。当前已知债面就是 `editable_paths`、浏览器 workbench、SSH 直连 runtime，这几个都被列入主任务，不作为 residual risk 留下。
- 是否存在“实现完成后立刻还要二次重构”的高概率点：唯一风险是 Docker mount 能力需要在 runtime engine 和 practice/runtime ports 同时补齐；因此计划专门把 mount 模型单独放在 Task 3，不允许在后续任务里再偷偷补第二套装配逻辑。
