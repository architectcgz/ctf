# AWD Defense SSH Host Key Persistence Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 AWD defense SSH 网关对固定入口持久化服务端 host key，避免每次后端重启都触发客户端 host key 变更告警。

**Architecture:** 在 AWD defense workspace 既有 SSH 网关上新增独立的 host key 文件配置。启动时优先加载指定私钥；仅在文件缺失时首次生成并原子写回。host key 生命周期和用户短时 SSH 票据完全分离，普通重启不轮换，显式替换文件才视为 rotation。

**Tech Stack:** Go、`golang.org/x/crypto/ssh`、本地文件系统配置、项目现有 config/validation、Go tests。

---

## 输入文档

- `docs/architecture/features/awd-defense-workspace-boundary-design.md`
- `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
- `code/backend/internal/app/composition/runtime_module.go`
- `code/backend/internal/config/config.go`

## 文件结构

- Modify: `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
- Modify: `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`
- Modify: `code/backend/internal/app/composition/runtime_module.go`
- Modify: `code/backend/internal/config/config.go`
- Modify: `code/backend/internal/config/config_test.go`
- Modify: `code/backend/configs/config.dev.yaml`
- Modify: `docs/architecture/features/awd-defense-workspace-boundary-design.md`

## 回退与恢复约定

- 仅新增 host key 配置与加载逻辑，不改变 SSH 票据格式、端口、用户名约定和 workspace 路由。
- 已存在的 host key 文件优先复用；不做自动 rotation。
- 若启动时发现 key 文件损坏，直接报错并阻止启动，避免静默切换 host identity。

### Task 1: 配置与网关持久化落盘

**Files:**

- Modify: `code/backend/internal/config/config.go`
- Modify: `code/backend/internal/app/composition/runtime_module.go`
- Modify: `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`

**Review focus:** host key 是否和临时票据解耦；首次生成与后续复用是否清晰；损坏文件是否 fail fast。

- [x] **Step 1: 先补配置模型与校验**

新增 `container.defense_ssh_host_key_path` 配置项、默认值和校验规则。

- [x] **Step 2: 在网关里抽出 host key 加载/首次生成逻辑**

要求：

- 文件存在时加载既有私钥
- 文件不存在时生成新的私钥并写回
- 文件损坏或目录不可写时返回错误

- [x] **Step 3: 把配置注入 SSH gateway 装配**

`runtime_module.go` 创建网关时传入 host key 路径，不在网关内部硬编码路径。

### Task 2: 回归测试与配置样例

**Files:**

- Modify: `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`
- Modify: `code/backend/internal/config/config_test.go`
- Modify: `code/backend/configs/config.dev.yaml`

**Review focus:** 测试是否覆盖“首次生成”“重启复用”“损坏文件失败”“配置校验”四类关键路径。

- [x] **Step 1: 补网关测试**

覆盖：

- 首次启动会生成 key 文件
- 再次启动会复用同一 key
- 非法 key 内容会返回错误

- [x] **Step 2: 补配置测试**

覆盖：

- `defense_ssh_enabled=true` 且 path 为空时报错
- 合法 path 能通过校验

- [x] **Step 3: 更新开发配置样例**

在 `config.dev.yaml` 中显式给出 `defense_ssh_host_key_path`，便于本地环境直接复用。

### Task 3: 定向验证

**Files:**

- Test: `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`
- Test: `code/backend/internal/config/config_test.go`

- [x] **Step 1: 跑网关与配置测试**

Run:

```bash
cd code/backend
go test ./internal/app/composition ./internal/config -run 'AWDDefenseSSH|DefenseSSHHostKey|DefenseSSHEnabled' -count=1
```

Expected: PASS

- [x] **Step 2: 手工验证固定入口不再轮换 host key**

重启本地后端进程两次，分别记录 `ssh-keyscan -p 2222 127.0.0.1` 输出的指纹，预期两次一致。

## 执行记录

- `go test ./internal/app/composition ./internal/config -run 'AWDDefenseSSH|DefenseSSHHostKey|DefenseSSHEnabled' -count=1` 通过
- 本地临时启动 `ctf-api`，两次通过 `ssh-keyscan -p 2222 127.0.0.1 | ssh-keygen -lf - -E sha256` 取得同一指纹：
  - `SHA256:NLLieq4rzMQ0eMlSabFWXrn0uJUGokl19DaKz1J17Jo`
- 持久化 key 文件位于 `code/backend/storage/runtime/awd-defense-ssh-host-key.pem`，文件权限为 `0600`

## Review 结论

- 本次改动已把 SSH 网关服务端身份和短时登录票据解耦，普通重启不会再隐式轮换 host key。
- 当前未发现新的 blocker；剩余运维前提是同一 `defense_ssh_host:port` 若由多实例共同承载，必须共享同一份 host key 文件。
