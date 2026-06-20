# Runtime Agent Only Execution Plane 实现 Review

**Review 日期**: 2026-06-20
**实现分支**: `task/2026-06-20-runtime-agent-only-execution-plane`
**实现计划**: `docs/plan/impl-plan/2026-06-20-runtime-agent-only-execution-plane-implementation-plan.md`
**Review 人员**: Claude (Opus 4.8)

---

## Executive Summary

本次实现成功完成了 runtime execution 执行面收口到 runtime-agent 的架构目标。所有 6 个 execution slices 均已完成，验证证据充分，代码质量整体良好。

### 核心成就

✅ **配置层安全默认值**：`runtime_agent.allow_local_fallback` 默认 `false`，生产环境强制拒绝
✅ **Composition 层守卫**：非 test 环境默认拒绝构造本地 Docker executor，除非显式 fallback
✅ **本地开发链路升级**：`dev-run.sh` 默认启动本机 runtime-agent，API/gateway 只持有 mTLS client 配置
✅ **Docker Compose 权限收窄**：docker.sock 只交给 `ctf-runtime-agent` 服务，API/gateway 已移除
✅ **文档同步更新**：架构文档、operations 手册、README、security todo 均已更新
✅ **测试覆盖充分**：config validation、composition 守卫、test runtime engine 隔离均有单测验证

### 安全边界改进

**Before**:
- API/gateway 容器直接挂载 `/var/run/docker.sock`
- 失陷后攻击者直接获得宿主 Docker daemon root 权限
- 默认配置训练开发者习惯不安全路径

**After**:
- API/gateway 只持有 runtime-agent mTLS client 配置
- docker.sock 只在 runtime-agent 进程中可见
- 本地 dev 默认也走 agent 边界，与生产一致
- 显式 fallback 只允许非生产 + 手动开启

---

## 实现质量评估

### 1. 架构一致性 ✅ PASS

实现严格遵循洋葱架构和既有分层约束：

**Config 层**:
- `RuntimeAgentConfig.AllowLocalFallback` 字段位置合理
- 默认值在 `setDefaults()` 统一设置
- 生产拒绝逻辑放在 `Validate()` 通用校验中
- 测试覆盖默认值和生产拒绝两个维度

```go
// code/backend/internal/config/validate.go:267
if c.RuntimeAgent.AllowLocalFallback {
    return fmt.Errorf("runtime_agent.allow_local_fallback must be false in prod")
}
```

**Composition 层**:
- 守卫逻辑前置到 `buildRuntimeNodeClientFromNode()`
- 三个小函数职责清晰：
  - `runtimeUsesTestEngine()` - 判断 test 环境
  - `runtimeAllowsLocalFallback()` - 判断是否允许 local executor
  - `usesLocalRuntimeNode()` - 判断 node 是否 local
- 错误消息清晰指向修复路径

```go
// code/backend/internal/app/composition/runtime_node_execution_router.go:119-124
if usesLocalRuntimeNode(node) {
    if !runtimeAllowsLocalFallback(cfg) {
        return nil, errLocalRuntimeFallbackDisabled
    }
    // ... build local executor
}
```

**边界守卫顺序正确**:
1. test 环境直接返回 test engine（无 Docker 依赖）
2. 非 test 环境检查 local node 时，先验证 fallback 开关
3. 只有通过验证才构造 local executor

### 2. 测试质量 ✅ PASS

**Config 层测试**（2 个新增）:
- `TestRuntimeAgentLocalFallbackDefaultsFalse` - 验证默认值
- `TestValidateRejectsRuntimeAgentLocalFallbackInProduction` - 验证生产拒绝

**Composition 层测试**（3 个新增）:
- `TestBuildContainerRuntimeModuleRejectsLocalRuntimeWithoutExplicitFallback` - 验证默认拒绝
- `TestBuildContainerRuntimeModuleAllowsExplicitLocalRuntimeFallback` - 验证显式允许
- `TestRuntimeNodeClientAllowsTestRuntimeEngineWithoutFallback` - 验证 test 环境隔离

**测试设计亮点**:
1. 使用 stub injection 验证 local executor 是否被构造
2. 错误消息断言包含 `runtime_agent.allow_local_fallback` 指引
3. test 环境测试验证不会意外触发 local Docker runner

**测试执行结果**:
```
✓ go test ./internal/config -run 'TestValidate.*RuntimeAgent|TestRuntimeAgentLocalFallback' -count=1
  → ok ctf-platform/internal/config 0.004s

✓ go test ./internal/app/composition -run 'TestBuildContainerRuntimeModule|TestRuntimeNodeExecutionRouter|TestRuntimeNodeClient' -count=1
  → ok ctf-platform/internal/app/composition 0.803s
```

### 3. Dev 脚本实现 ✅ PASS

`code/backend/scripts/dev-run.sh` 新增 249 行逻辑：

**证书生成逻辑**:
- `ensure_local_runtime_agent_certs()` - 幂等证书生成
- `create_runtime_agent_leaf_cert()` - 标准 OpenSSL 流程
- 证书放在 `docker/runtime/runtime-agent-certs/`
- 已存在时跳过重新生成

**启动决策逻辑**:
- `should_start_local_runtime_agent()` - 三重检查：
  1. `CTF_DEV_RUNTIME_AGENT=false` → 不启动
  2. `CTF_RUNTIME_AGENT_ALLOW_LOCAL_FALLBACK=true` → 不启动（走 fallback）
  3. `CTF_RUNTIME_AGENT_ENDPOINT` 已配置 → 不启动（使用远端）
- 默认行为：启动本机 runtime-agent

**端口选择**:
- `choose_local_runtime_agent_port()` - 智能端口选择
- 优先使用 `CTF_RUNTIME_AGENT_SERVER_PORT` 环境变量
- 默认 `19443`，如果占用则 `+1` 直到找到可用端口

**环境变量导出**:
```bash
export CTF_RUNTIME_AGENT_ENABLED=true
export CTF_RUNTIME_AGENT_ENDPOINT=127.0.0.1:${port}
export CTF_RUNTIME_AGENT_SERVER_NAME=runtime-agent.local
export CTF_RUNTIME_AGENT_CA_FILE=${cert_dir}/ca.pem
export CTF_RUNTIME_AGENT_CERT_FILE=${cert_dir}/client.pem
export CTF_RUNTIME_AGENT_KEY_FILE=${cert_dir}/client-key.pem
```

**生命周期管理**:
- foreground 模式：API 退出时自动停止 runtime-agent
- background 模式：PID 写入文件供外部管理
- 信号处理：INT → TERM → KILL 梯度优雅停止

**脚本验证**:
```
✓ bash -n code/backend/scripts/dev-run.sh
  → 语法检查通过
```

### 4. Docker Compose 改动 ✅ PASS

新增 149 行，主要包含三个新服务：

**`ctf-runtime-agent-certs` 服务**:
- 一次性证书生成服务（`alpine:3.22` + `openssl`）
- 生成 CA、server、client 三对密钥对
- 证书存储在共享 volume `runtime-agent-certs`
- 幂等设计：证书已存在时直接退出

**`ctf-runtime-agent` 服务**:
- 使用 `ctf-backend:local` 镜像
- **唯一挂载 docker.sock 的服务**
- 监听 `0.0.0.0:9443`（容器内）
- mTLS server 配置完整
- healthcheck 使用 `openssl s_client` 验证 mTLS 握手、CA 和 hostname
- `depends_on: ctf-runtime-agent-certs` 确保证书就绪

**API/gateway 改动**:
- ❌ **移除**: `/var/run/docker.sock` volume mount
- ❌ **移除**: `DOCKER_HOST` 环境变量
- ✅ **新增**: runtime-agent client 配置
  - `CTF_RUNTIME_AGENT_ENABLED=true`
  - `CTF_RUNTIME_AGENT_ENDPOINT=ctf-runtime-agent:9443`
  - `CTF_RUNTIME_AGENT_SERVER_NAME=runtime-agent.local`
  - CA / cert / key 文件路径
- ✅ **新增**: `depends_on: ctf-runtime-agent` 健康检查

**权限验证**:
```bash
$ rg "/var/run/docker.sock|DOCKER_HOST" /tmp/ctf-compose-dev-config.yaml
286:      DOCKER_HOST: unix:///var/run/docker.sock     # 只在 ctf-runtime-agent
316:        source: /var/run/docker.sock                # 只在 ctf-runtime-agent
317:        target: /var/run/docker.sock
```

✅ 确认 API/gateway 不再持有 docker.sock

### 5. 文档更新 ✅ PASS

**架构文档** (`docs/architecture/backend/01-system-architecture.md`):
- 7.2 节更新：`DOCKER_HOST` 只属于 runtime-agent 执行节点
- 7.5 节安全边界表格更新：
  - 原"API 直接跑宿主 Docker" → "API + 本机 runtime-agent"
  - 明确本地 fallback 只在非生产 + 显式开启时允许
- 7.6 节配置示例更新：
  - `runtime_agent.enabled: true`（非 false）
  - 新增 `allow_local_fallback: false` 说明
  - endpoint/certs 示例完整

**Operations 手册** (`docs/operations/runtime-agent-deployment.md`):
- 模式 1 更新：本机开发默认启动 runtime-agent
- 模式 5 更新：gateway 默认走 runtime-agent
- Client 配置章节：新增 `allow_local_fallback` 说明

**README.md**:
- 默认开发路径说明：dev-run.sh 启动本机 runtime-agent
- Docker Compose 路径说明：
  - docker.sock 只交给 `ctf-runtime-agent`
  - API/gateway 通过 mTLS 访问 agent
  - fallback 只在非生产 + 显式开启时允许
- 端口列表新增：`ctf-runtime-agent: 9443`（内部）

**Security TODO** (`docs/todos/2026-06-02-security-review-findings.md`):
- P1 高风险项标记完成：`[x]`
- 修复说明：docker.sock 已从 API/gateway 移除
- 关联引用更新

---

## 发现的问题

### Critical Issues

无。

### High Priority Issues

无。

### Medium Priority Issues

#### M1: Dev 脚本证书生成缺少错误处理

**Post-review 状态**: 已处理。`dev-run.sh` 已将本机证书生成封装为带上下文的命令执行，失败时输出具体步骤、证书目录提示和 openssl stderr。

**位置**: `code/backend/scripts/dev-run.sh:385-430`

**问题**:
`ensure_local_runtime_agent_certs()` 和 `create_runtime_agent_leaf_cert()` 中的 `openssl` 命令使用了 `>/dev/null 2>&1` 重定向，但没有检查返回码。如果证书生成失败（如权限问题、openssl 版本不兼容），脚本会继续执行，导致后续 runtime-agent 启动失败。

**建议**:
```bash
openssl req -x509 ... || {
    echo "生成 CA 证书失败" >&2
    return 1
}
```

**影响**: Medium（证书生成失败时诊断困难）

**当前缓解**: 脚本在证书生成前已检查 `openssl` 命令存在性

#### M2: Compose 健康检查可能误报

**Post-review 状态**: 已处理。`ctf-runtime-agent` healthcheck 已从端口探测改为 `openssl s_client` mTLS 握手校验；运行镜像补充 `openssl` 依赖。

**位置**: `docker/docker-compose.dev.yml:120-125`

**问题**:
`ctf-runtime-agent` 的 healthcheck 使用 `nc -z 127.0.0.1 9443` 只检查端口监听，不验证 TLS 握手是否正常。如果 server 启动但证书配置错误，健康检查仍会通过，导致 API 启动后 mTLS 握手失败。

**建议**:
```yaml
test:
  - CMD-SHELL
  - |
    openssl s_client -connect 127.0.0.1:9443 \
      -CAfile /app/runtime-agent-certs/ca.pem \
      -cert /app/runtime-agent-certs/client.pem \
      -key /app/runtime-agent-certs/client-key.pem \
      </dev/null 2>/dev/null | grep -q 'Verify return code: 0'
```

**影响**: Medium（证书配置错误时诊断延后）

**当前缓解**: 证书生成逻辑已验证通过，API 启动失败日志会暴露 mTLS 错误

### Low Priority Issues

#### L1: 环境变量命名不一致

**位置**: `code/backend/scripts/dev-run.sh:21-28`

**问题**:
使用了 `CTF_DEV_RUNTIME_AGENT_*` 前缀用于脚本内部变量，但实际配置使用 `CTF_RUNTIME_AGENT_*`。虽然逻辑正确（内部变量不应与配置冲突），但命名空间稍显复杂。

**建议**: 保持现状，或在注释中明确说明两套前缀的用途区别。

**影响**: Low（可读性）

#### L2: 端口冲突处理未记录

**Post-review 状态**: 已处理。`dev-run.sh` 在默认端口被占用并自动切换时会输出最终端口提示。

**位置**: `code/backend/scripts/dev-run.sh:361-369`

**问题**:
`choose_local_runtime_agent_port()` 自动递增端口避免冲突，但没有记录最终选择的端口与默认端口不同的情况。用户可能困惑为什么 endpoint 显示 `19444` 而不是预期的 `19443`。

**建议**:
```bash
if [[ "${port}" -ne "${CTF_DEV_RUNTIME_AGENT_DEFAULT_PORT}" ]]; then
    echo "注意：默认端口 ${CTF_DEV_RUNTIME_AGENT_DEFAULT_PORT} 被占用，已切换到 ${port}"
fi
```

**影响**: Low（用户体验）

---

## 验证证据总结

所有 validation plan 中的命令均已执行并通过：

| 验证项 | 命令 | 结果 |
|--------|------|------|
| Config 测试 | `go test ./internal/config -run 'TestValidate.*RuntimeAgent\|TestRuntimeAgentLocalFallback' -count=1` | ✅ PASS (0.004s) |
| Composition 测试 | `go test ./internal/app/composition -run 'TestBuildContainerRuntimeModule\|TestRuntimeNodeExecutionRouter\|TestRuntimeNodeClient' -count=1` | ✅ PASS (0.803s) |
| 脚本语法 | `bash -n code/backend/scripts/dev-run.sh` | ✅ PASS |
| Compose 配置 | `docker compose -f docker/docker-compose.dev.yml config` | ✅ PASS |
| Startup gate | `bash scripts/check-startup-gate.sh` | ✅ PASS |
| Completion gate | `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full` | ✅ PASS (180s) |
| Workflow governance | `bash scripts/check-workflow-governance.sh` | ✅ PASS (180s) |
| Architecture full | `bash scripts/check-architecture.sh --full` | ✅ PASS (180s) |
| Git diff check | `git diff --check` | ✅ PASS |

**手工验证**:
- ✅ `/var/run/docker.sock` 只出现在 `ctf-runtime-agent` 服务中
- ✅ API/gateway 不再持有 `DOCKER_HOST` 环境变量
- ✅ API/gateway 包含完整 runtime-agent client mTLS 配置

---

## 架构契合度评估

### 目标边界符合度 ✅

实现完全符合 implementation plan 中定义的 owner 边界：

| Owner | 职责 | 实现验证 |
|-------|------|----------|
| `config` | 解析 fallback 配置，生产安全校验 | ✅ `validate.go:267` |
| `app/composition` | 决定是否可构造 local executor | ✅ `runtime_node_execution_router.go:119-124` |
| `scripts/dev-run.sh` | 本地 dev 启动 runtime-agent | ✅ `dev-run.sh:470-478` |
| `docker-compose.dev.yml` | docker.sock 只交给 runtime-agent | ✅ 已验证 |
| `runtime-agent` | 唯一允许触碰 Docker 的 runtime 进程 | ✅ 已验证 |

### 复用点正确性 ✅

- ✅ 复用 `RuntimeAgentConfig` 添加 `AllowLocalFallback` 字段
- ✅ 复用 `buildRuntimeNodeClientFromNode()` 添加守卫逻辑
- ✅ 复用 `newLocalRuntimeHostRunner` 和 `newLocalSandboxExecutor`（只在 fallback 允许时）
- ✅ 复用 existing test runtime engine（与 Docker 隔离）

### Non-Goals 遵守度 ✅

实现正确遵守了所有 Non-Goals：

- ✅ 未触碰 challenge image build 权限（`docker_image_builder.go` 不在改动范围）
- ✅ 未引入生产证书自动化或完整 PKI workflow（dev 证书 30 天有效期）
- ✅ 未添加 runtime 容量调度、live migration、会话迁移
- ✅ 未移除 runtime-agent 的 Docker 访问权限

---

## 风险评估

### 回归风险: LOW

**理由**:
1. Test 环境仍使用独立的 test runtime engine，不依赖 Docker
2. Composition 层守卫在构造 executor 之前拦截，不影响已构造的 executor
3. 显式 fallback 保留了"紧急回到本地 Docker"的路径

**验证**:
- ✅ 单测验证 test 环境不走 local Docker executor
- ✅ 单测验证显式 fallback 可以构造 local executor

### 部署风险: MEDIUM

**理由**:
1. 首次部署需要配置 runtime-agent mTLS 证书
2. 多节点环境需要在每个 runtime node 上部署 runtime-agent
3. `runtime_nodes` 表中 default node 从 `local-default` 变为 `agent-default`

**缓解**:
- ✅ `docs/operations/runtime-agent-deployment.md` 包含完整部署步骤
- ✅ 生产环境拒绝 `allow_local_fallback=true`，避免误用 fallback
- ✅ 本地 dev 脚本自动处理证书生成和 runtime-agent 启动

### 性能风险: LOW

**理由**:
1. 本机 dev runtime-agent 与 API 仍在同一宿主，只增加 gRPC + mTLS 开销
2. 远端 runtime-agent 的网络延迟已在既有 agent protocol 中存在

**观测**:
- 单测执行时间无明显增加（composition 测试 0.803s 正常）

---

## 建议

### 1. 合并建议: APPROVE WITH MINOR COMMENTS

**核心改动**: ✅ 质量达标，可以合并
**文档同步**: ✅ 完整更新
**测试覆盖**: ✅ 充分
**安全改进**: ✅ 显著

**合并前建议处理**:
- [ ] (Optional) M1: dev-run.sh 证书生成错误处理
- [ ] (Optional) M2: Compose healthcheck 改用 TLS 验证
- [ ] (Optional) L2: 端口冲突时输出提示

**可推迟到后续**:
- L1: 环境变量命名（非功能性，保持现状即可）

### 2. 后续工作建议

#### 短期（1-2 周内）:

1. **验证本地 dev 启动体验**
   - 在全新环境（无 `docker/runtime/runtime-agent-certs/`）运行 `dev-run.sh`
   - 验证证书生成 + runtime-agent 启动 + API 连接的完整链路
   - 创建实例、进入容器、运行 checker 验证执行面正常

2. **验证 Compose 全栈启动**
   - `docker compose -f docker/docker-compose.dev.yml up --build`
   - 验证 `ctf-runtime-agent-certs` → `ctf-runtime-agent` → API/gateway 依赖链
   - 验证容器内创建实例、AWD service 正常

#### 中期（1 个月内）:

3. **远端 runtime-agent 部署验证**
   - 按 `docs/operations/runtime-agent-deployment.md` 模式 2 部署
   - 验证 API 主机 → 远端 runtime node 的跨机执行
   - 验证 node health、failover、scheduler 逻辑

4. **Image build plane 架构决策**
   - 当前 challenge image build 仍持有 Docker 权限
   - 评估是否需要类似的"image-build-agent"边界
   - 或接受"image build 是信任面"的架构假设

#### 长期（后续迭代）:

5. **Runtime-agent 可观测性**
   - runtime-agent 自身的 metrics、日志聚合
   - 执行面延迟、失败率、容量利用率监控

6. **Runtime node 容量调度**
   - 当前只有健康过滤，无容量感知调度
   - `capacity_snapshot` 已采集但未用于调度决策

---

## 结论

本次实现**质量优秀**，完全符合 implementation plan 的架构目标，测试覆盖充分，文档同步完整。

**核心成就**:
- Runtime 执行面收口到 runtime-agent，API/gateway 只持有 mTLS client 配置
- 默认开发路径升级到与生产一致的 agent 边界
- Docker socket 权限收窄到单一 runtime-agent 服务
- 显式 fallback 保留了非生产排障路径，但受配置校验保护

**发现的 Medium 问题均不阻塞合并**，可在合并后作为改进项处理。

**Review 结论**: ✅ **APPROVE**

建议在合并后立即进行"短期验证"清单中的本地 dev 启动体验验证，确保首次使用者的平滑体验。

---

## Appendix: Implementation Plan Checklist Status

所有 6 个 execution slices 的 checklist 均已完成：

- [x] Slice 1: Config Contract For Explicit Local Fallback
- [x] Slice 2: Composition Guardrail Against Direct Docker Runtime
- [x] Slice 3: Local Dev Runtime-Agent Startup
- [x] Slice 4: Dev Compose Execution Plane
- [x] Slice 5: Docs And README Alignment
- [x] Slice 6: Integration Validation And Workflow Gate

Implementation plan 的 validation evidence 和 independent review handoff 章节已由实现者填写完整。
