# AWD Defense Workspace Shell Environment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 AWD defense workspace 的 SSH 会话默认具备稳定的 UTF-8 终端环境和最基本的编辑/比对工具，避免中文输入退化成 `...`，同时补齐当前 workspace 复用逻辑的编译缺口。

**Architecture:** 保持现有“runtime 容器 + 独立 workspace companion 容器”的结构，不引入新的镜像分发流程。workspace companion 继续使用轻量基础镜像，但在创建时注入固定的 `LANG/LC_ALL/TERM`，并通过启动命令补齐 `git`、`vim`、`nano` 后再进入常驻空闲态；同一切片内修复 `RuntimeInstanceService` 缺失 `InspectManagedContainer` 方法导致的基线编译失败，确保 workspace 复用/重建逻辑可验证。

**Tech Stack:** Go、项目现有 practice runtime topology create flow、Docker 容器启动命令与环境变量、Go tests。

---

## 输入文档

- `code/backend/internal/module/practice/application/commands/awd_defense_workspace_support.go`
- `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/application/commands/{service_test.go,runtime_container_create_test.go,instance_start_service_test.go,contest_instance_service_test.go}`
- `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`

## 文件结构

- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Modify: `code/backend/internal/module/practice/application/commands/awd_defense_workspace_support.go`
- Modify: `code/backend/internal/module/runtime/runtime/adapters.go`
- Modify: `code/backend/internal/module/runtime/runtime/adapters_test.go`
- Modify: `code/backend/internal/module/runtime/runtime/module.go`
- Modify: `code/backend/internal/module/runtime/service_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/service_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/runtime_container_create_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/instance_start_service_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/contest_instance_service_test.go`
- Create: `docs/reviews/backend/2026-05-07-awd-defense-workspace-shell-environment-review.md`

## 非目标

- 不在本次切片中引入预构建 `defense workspace shell` 镜像或 registry 分发流程
- 不修改题包结构、题包导入契约或学生端 AWD 页面
- 不重做 SSH gateway 的 `env` 请求透传协议

## 现状与风险

- 当前基线已经存在编译问题：`awd_defense_workspace_support.go` 调用了 `s.runtimeService.InspectManagedContainer(...)`，但 `practice/ports.RuntimeInstanceService` 尚未声明该方法；`go test ./internal/module/practice/application/commands ...` 在修改前即失败。
- workspace companion 当前只起 `python:3.12-alpine` + `tail -f /dev/null`，没有固定 `LANG/LC_ALL/TERM`，也没有 `git` / `vim` / `nano`。
- 本次最小实现会把工具安装放到 workspace companion 启动阶段，因此首次创建 companion 仍依赖 Alpine 仓库可达；这是本切片明确接受的残余风险，后续若要消除，应另起“预构建 shell 镜像”方案。

### Task 1: 修复 workspace 复用链路的接口基线

**Files:**

- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Modify: `code/backend/internal/module/runtime/runtime/adapters.go`
- Modify: `code/backend/internal/module/runtime/runtime/adapters_test.go`
- Modify: `code/backend/internal/module/runtime/runtime/module.go`
- Modify: `code/backend/internal/module/practice/application/commands/service_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/contest_instance_service_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/instance_start_service_test.go`

**Review focus:** `RuntimeInstanceService` 的抽象边界是否仍只暴露 practice 命令侧真正需要的方法；practice -> runtime adapter 是否完整透传 `ContainerName / Command / WorkingDir / Mounts`；测试 stub 与 adapter tests 是否完整覆盖新增接口。

- [x] **Step 1: 补齐 `RuntimeInstanceService` 所需的容器检查方法**

要求：

- 在 `practice/ports.RuntimeInstanceService` 中显式声明 `InspectManagedContainer`
- 复用已有 `ManagedContainerState`，不新造第二套状态结构

- [x] **Step 2: 对齐 runtime adapter 装配**

要求：

- `runtime/runtime/adapters.go` 中的 `runtimePracticeServiceAdapter` 实现新增方法
- `toRuntimeTopologyCreateRequest(...)` 必须继续透传 `Env`，并补齐 `Command`、`WorkingDir`、`Mounts`、`ContainerName`
- `runtime/runtime/module.go` 继续通过现有 runtime engine 注入该能力，不新增第二套 container inspector owner

- [x] **Step 3: 对齐命令层测试 stub**

要求：

- `service_test.go` 中的 `stubPracticeRuntimeService` 支持 `InspectManagedContainer`
- `contest_instance_service_test.go` 中的实现同步满足接口
- `instance_start_service_test.go` 中受 restart/workspace 复用影响的 stub 一并满足接口 rollout
- `runtime/runtime/adapters_test.go` 显式覆盖 topology 字段透传与 `InspectManagedContainer` delegation

- [x] **Step 4: 跑 workspace 命令包基线编译检查**

Run:

```bash
cd code/backend
go test ./internal/module/practice/application/commands ./internal/module/runtime/runtime -run 'TestCreateSingleAWDContainerUsesPrivateTopology|TestRestartContestAWDServiceReusesExistingDefenseWorkspaceContainer|TestRestartContestAWDServiceRecreatesMissingDefenseWorkspaceContainer|Practice|Topology' -count=1
```

Expected:

- 至少通过编译阶段，不再出现 `InspectManagedContainer undefined`

### Task 2: 给 workspace companion 注入 UTF-8 环境与基础工具

**Files:**

- Modify: `code/backend/internal/module/practice/application/commands/awd_defense_workspace_support.go`
- Modify: `code/backend/internal/module/practice/application/commands/runtime_container_create_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/instance_start_service_test.go`
- Modify: `code/backend/internal/module/runtime/service_test.go`

**Review focus:** companion 容器命令是否清晰、可重复推导；环境变量是否直接落在容器配置而不是依赖 SSH 客户端；失败模式是否和现有 workspace 创建流程兼容。

- [x] **Step 1: 抽出 workspace shell 启动常量**

要求：

- 保持现有 companion 镜像与 `/workspace` working dir
- 新增固定环境：
  - `LANG=C.UTF-8`
  - `LC_ALL=C.UTF-8`
  - `TERM=xterm-256color`
- 新增 companion 启动命令，先补 `git`、`vim`、`nano`，再进入常驻空闲态

- [x] **Step 2: 把环境和命令注入 workspace companion create request**

要求：

- 只影响 workspace companion，不改 runtime 容器
- 继续保留共享 network alias、mount、working dir 与 `DisableEntryPortPublishing`

- [x] **Step 3: 更新定向测试断言**

覆盖：

- `runtime_container_create_test.go` 断言 workspace 节点包含新的 env 和 command
- `instance_start_service_test.go` 断言重建 workspace 时仍走同一 companion 形状

### Task 3: 最小充分验证与独立评审

**Files:**

- Test: `code/backend/internal/module/practice/application/commands/*.go`
- Create: `docs/reviews/backend/2026-05-07-awd-defense-workspace-shell-environment-review.md`

- [x] **Step 1: 跑定向测试**

Run:

```bash
cd code/backend
go test ./internal/module/practice/application/commands ./internal/module/runtime/runtime -run 'TestCreateSingleAWDContainerUsesPrivateTopology|TestRestartContestAWDServiceReusesExistingDefenseWorkspaceContainer|TestRestartContestAWDServiceRecreatesMissingDefenseWorkspaceContainer|Practice|Topology' -count=1
```

Expected: PASS

- [x] **Step 2: 跑更宽一层的 workspace 相关命令包验证**

Run:

```bash
cd code/backend
go test ./internal/module/practice/application/commands -run 'AWD|Workspace|ContestInstance' -count=1
```

Expected: PASS

- [x] **Step 3: 归档独立 review**

输出：

- `docs/reviews/backend/2026-05-07-awd-defense-workspace-shell-environment-review.md`

Review focus：

- runtime 容器与 workspace companion 的职责边界是否保持清晰
- UTF-8 环境是否真正落在 companion 容器配置，而不是只停留在文案或注释
- workspace 启动命令是否引入新的明显回归风险
- 基线编译缺口是否在本切片内收口

## Plan Self-Review

- 目标边界清晰：本切片只补 workspace companion 的 shell 可用性，不把问题扩成新的镜像构建体系。
- 复用点清晰：继续复用现有 `CreateTopology`、workspace row、SSH gateway 与 mount 模型，不新增第二套 workspace 生命周期。
- 结构收敛清晰：把已暴露的 `RuntimeInstanceService` 编译缺口和 shell 环境缺口放在同一切片收口，避免一边加能力一边继续留着当前 touched surface 无法编译。
- 残余风险显式：首次 companion 启动依赖 Alpine 仓库可达，这次接受；若后续需要彻底去掉外部依赖，应另起“预构建 shell 镜像”方案，而不是在本切片里继续堆大。
