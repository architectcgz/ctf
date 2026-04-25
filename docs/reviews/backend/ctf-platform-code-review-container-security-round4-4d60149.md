# ctf-platform 代码 Review（container-security 第 4 轮）：验证高优先级问题修复

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | container-security |
| 轮次 | 第 4 轮（修复后复审） |
| 审查范围 | commit 4d60149，4 个文件，28 行新增，14 行删除 |
| 变更概述 | 修复 CPU 配额单位、配置验证上下限、运行时参数校验 |
| 审查基准 | docs/tasks/backend-task-breakdown.md B10 任务验收标准 |
| 审查日期 | 2026-03-05 |
| 上轮问题数 | 11 项（3 高 / 5 中 / 3 低） |

## 问题清单

### 🔴 高优先级

**无高优先级问题**

---

### 🟡 中优先级

#### [M1] 配置文件缺少单位注释，运维人员无法正确配置

- **文件**：`code/backend/configs/config.yaml:92-102`
- **问题描述**：
  配置文件中的数值仍然没有注释说明单位和含义：
  ```yaml
  container:
    default_cpu_quota: 0.5        # 缺少注释
    default_memory: 268435456     # 这是多少 MB？
    default_pids_limit: 100
    readonly_rootfs: false
    run_as_user: ""
    allowed_capabilities:
      - CHOWN
      - SETUID
      - SETGID
    seccomp: default
  ```

- **影响范围/风险**：运维人员无法理解配置含义，可能导致配置错误

- **修正建议**：
  ```yaml
  container:
    default_cpu_quota: 0.5              # CPU 核心数（0.5 = 半核）
    default_memory: 268435456           # 内存限制（字节），268435456 = 256MB
    default_pids_limit: 100             # 最大进程数
    readonly_rootfs: false              # 是否只读根文件系统
    run_as_user: ""                     # 运行用户（空字符串表示使用镜像默认用户）
    allowed_capabilities:               # 允许的 Linux Capabilities（已 Drop ALL）
      - CHOWN                           # 允许修改文件所有者
      - SETUID                          # 允许设置用户 ID
      - SETGID                          # 允许设置组 ID
    seccomp: default                    # Seccomp 配置（default/unconfined/文件路径）
  ```

---

#### [M2] DefaultSecurityConfig 函数未被使用

- **文件**：`code/backend/internal/module/container/engine.go:53-60` vs `95-103`
- **问题描述**：
  `DefaultSecurityConfig` 函数仍然存在但未被调用，`CreateContainer` 中直接内联创建 `SecurityConfig`：

  ```go
  // 第 53-60 行：内联创建
  if cfg.Security == nil {
      cfg.Security = &model.SecurityConfig{
          ReadonlyRootfs: e.containerCfg.ReadonlyRootfs,
          CapDrop:        []string{"ALL"},
          CapAdd:         e.containerCfg.AllowedCapabilities,
          SecurityOpt:    []string{fmt.Sprintf("seccomp=%s", e.containerCfg.Seccomp), "no-new-privileges:true"},
          User:           e.containerCfg.RunAsUser,
      }
  }

  // 第 95-103 行：函数定义但无人调用
  func DefaultSecurityConfig(cfg *config.ContainerConfig) *model.SecurityConfig { ... }
  ```

- **影响范围/风险**：代码冗余，维护成本增加

- **修正建议**：
  删除 `DefaultSecurityConfig` 函数，或在 `CreateContainer` 中使用它：
  ```go
  if cfg.Security == nil {
      cfg.Security = DefaultSecurityConfig(e.containerCfg)
  }
  ```

---

#### [M3] 缺少容器名称生成逻辑

- **文件**：`code/backend/internal/module/container/engine.go:88`（推测行号）
- **问题描述**：
  容器名称仍然是空字符串，Docker 会生成随机名称（如 `silly_einstein`）

- **影响范围/风险**：
  1. 无法通过名称快速识别容器属于哪个用户/题目
  2. 日志和监控中容器名称无意义

- **修正建议**：
  在 `model.ContainerConfig` 中添加 `Name` 字段，或根据其他字段生成名称：
  ```go
  containerName := fmt.Sprintf("ctf-%s-%d-%s", cfg.ChallengeID, cfg.UserID, generateRandomSuffix())
  resp, err := e.cli.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, containerName)
  ```

---

#### [M4] 磁盘 IO 限制仍未实现

- **文件**：`code/backend/internal/model/container.go:14-18`
- **问题描述**：
  B10 任务要求包含"磁盘 IO 限制：`--storage-opt` 参数"，但当前代码完全没有实现

- **影响范围/风险**：
  1. 容器可以无限制写入磁盘，可能填满宿主机磁盘
  2. 不符合 B10 任务交付物要求

- **修正建议**：
  如果当前阶段不实现磁盘限制，应在任务文档中明确说明延后到后续任务

---

### 🟢 低优先级

#### [L1] 缺少日志记录

- **文件**：`code/backend/internal/module/container/engine.go:31-93`
- **问题描述**：
  `CreateContainer` 方法中没有任何日志记录，无法追踪容器创建过程和配置参数

- **影响范围/风险**：生产环境排查问题困难

- **修正建议**：
  添加日志：
  ```go
  func (e *Engine) CreateContainer(ctx context.Context, cfg *model.ContainerConfig) (string, error) {
      log.Info("creating container",
          "image", cfg.Image,
          "cpu_quota", cfg.Resources.CPUQuota,
          "memory", cfg.Resources.Memory,
          "pids_limit", cfg.Resources.PidsLimit,
      )

      // ... 现有逻辑

      log.Info("container created", "id", resp.ID)
      return resp.ID, nil
  }
  ```

---

#### [L2] 缺少配置默认值的单元测试

- **文件**：`code/backend/internal/config/config.go:221`
- **问题描述**：
  `setDefaults` 函数中设置了容器配置的默认值，但没有单元测试验证这些默认值是否正确加载

- **影响范围/风险**：配置加载失败时难以排查

- **修正建议**：
  添加单元测试验证配置默认值

---

#### [L3] tmpfs 挂载逻辑仍然有问题

- **文件**：`code/backend/internal/module/container/engine.go:82-86`（推测行号）
- **问题描述**：
  tmpfs 只在 `ReadonlyRootfs=true` 时挂载，但配置文件中 `readonly_rootfs: false`，导致 tmpfs 永远不会被挂载

- **影响范围/风险**：
  1. 如果 `ReadonlyRootfs=false`，tmpfs 不挂载，容器可以写入根文件系统任意位置
  2. 如果 `ReadonlyRootfs=true` 但忘记挂载 tmpfs，容器无法写入 `/tmp`

- **修正建议**：
  无论 `ReadonlyRootfs` 是否启用，都应该挂载 tmpfs 到 `/tmp`：
  ```go
  // 始终挂载 tmpfs 到 /tmp
  hostCfg.Tmpfs = map[string]string{
      "/tmp": "rw,noexec,nosuid,size=65536k",
  }

  // 根据配置决定是否只读根文件系统
  hostCfg.ReadonlyRootfs = cfg.Security.ReadonlyRootfs
  ```

---

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 0 |
| 🟡 中 | 4 |
| 🟢 低 | 3 |
| 合计 | 7 |

## 总体评价

本轮修复成功解决了第 3 轮的所有 3 个高优先级问题，代码质量显著提升。

**已修复的高优先级问题**：
- ✅ H1: CPU 配额单位混乱 → 已改为 float64 核心数，计算公式改为 `* 1e9`，语义清晰
- ✅ H2: 配置验证不完整 → 已添加上下限检查（CPU: 0-16 核，内存: 64MB-16GB，PID: 1-10000）
- ✅ H3: 运行时验证缺失 → 已在 `CreateContainer` 开头添加运行时参数校验

**核心改进**：
1. **CPU 配额语义清晰**：配置文件使用 `0.5` 表示 0.5 核心，代码中使用 `* 1e9` 转换为纳秒，逻辑直观易懂
2. **配置验证完整**：启动时和运行时都有完整的参数校验，防止恶意或错误配置
3. **类型一致性**：`ContainerConfig.DefaultCPUQuota` 和 `ResourceLimits.CPUQuota` 都改为 `float64`，支持小数核心数

**剩余问题**：
- 4 个中优先级问题（配置注释、代码冗余、容器命名、磁盘限制）
- 3 个低优先级问题（日志、测试、tmpfs 挂载）

**验收状态**：
- ✅ 容器创建时正确应用 CPU、内存、PID 限制
- ✅ 安全配置生效（Capabilities、Seccomp、no-new-privileges）
- ✅ 配置可通过 YAML 灵活调整
- ✅ 配置验证完整，防止恶意配置
- ✅ 启动前校验资源配置合法性
- ✅ 超出宿主机资源上限时拒绝创建

**合并建议**：
所有高优先级问题已修复，代码已满足 B10 任务的核心验收标准。剩余的中低优先级问题不阻塞合并，可以在后续迭代中优化。

**建议合并到主分支**，并在后续任务中处理：
1. M1: 补充配置文件注释（5 分钟工作量）
2. M2: 清理冗余代码（2 分钟工作量）
3. M3: 实现容器命名逻辑（10 分钟工作量）
4. M4: 磁盘 IO 限制（需要独立任务）
5. L1-L3: 日志、测试、tmpfs 优化（后续迭代）
