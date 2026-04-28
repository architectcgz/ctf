# ctf-platform 代码 Review（container-security 第 2 轮）：容器资源限制与安全加固

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | container-security |
| 轮次 | 第 2 轮（首次审查） |
| 审查范围 | commit e8424e6，4 个文件，130 行新增 |
| 变更概述 | 实现容器资源限制（CPU/内存/PID）和安全加固（特权禁用/Capabilities/只读根文件系统） |
| 审查基准 | docs/tasks/backend-task-breakdown.md B10 任务验收标准 |
| 审查日期 | 2026-03-05 |
| 上轮问题数 | N/A（首次提交代码） |

## 问题清单

### 🔴 高优先级

#### [H1] CPU 配额计算错误导致资源限制失效

- **文件**：`code/backend/internal/module/container/engine.go:69`
- **问题描述**：
  ```go
  resources.NanoCPUs = int64(cfg.Resources.CPUQuota * 1e9)
  ```
  当前代码将 `CPUQuota` 直接乘以 1e9，但配置文件中 `default_cpu_quota: 50000` 的单位含义不明确。

  根据 Docker API 文档，`NanoCPUs` 的单位是纳秒级 CPU 时间（1e9 = 1 核心）。如果 `CPUQuota` 表示"CPU 核心数"，那么 0.5 核应该是 `0.5`，而不是 `50000`。

  当前配置 `50000 * 1e9 = 5e13`，相当于 50000 个 CPU 核心，完全失去了资源限制的意义。

- **影响范围/风险**：容器可以无限制使用 CPU 资源，导致宿主机资源耗尽，无法通过验收标准"容器内执行 Fork Bomb 不影响其他容器"
- **修正建议**：
  1. 统一 `CPUQuota` 的语义为"CPU 核心数"（浮点数，如 0.5 表示半核）
  2. 修改配置文件：`default_cpu_quota: 0.5`
  3. 修改计算逻辑：
     ```go
     resources.NanoCPUs = int64(cfg.Resources.CPUQuota * 1e9)
     ```
  4. 在 model 中添加注释说明单位

---

#### [H2] 配置项定义但完全未使用

- **文件**：`code/backend/internal/config/config.go:102-107`、`engine.go:73-89`
- **问题描述**：
  `config.go` 中定义了 `ContainerConfig` 结构体，包含 5 个配置项：
  - `DefaultCPUQuota`
  - `DefaultMemory`
  - `DefaultPidsLimit`
  - `ReadonlyRootfs`
  - `RunAsUser`

  但在 `engine.go` 的 `CreateContainer` 方法中，这些配置项完全没有被使用。安全配置全部硬编码：
  ```go
  Privileged:     false,
  ReadonlyRootfs: true,  // 硬编码为 true
  CapDrop:        []string{"ALL"},
  CapAdd:         []string{"NET_BIND_SERVICE"},
  SecurityOpt:    []string{"no-new-privileges"},
  ```

- **影响范围/风险**：
  1. 配置文件修改不生效，运维无法调整安全策略
  2. 不同题目无法使用不同的安全配置（如某些题目需要写文件）
  3. 违反"资源配置可按题目自定义"的验收标准

- **修正建议**：
  1. `Engine` 结构体中注入 `config.ContainerConfig`
  2. `CreateContainer` 中使用配置项而非硬编码
  3. 将 `DefaultSecurityConfig()` 函数改为从配置读取

---

#### [H3] 安全配置缺少 Seccomp 和完整的 Capabilities 控制

- **文件**：`code/backend/internal/module/container/engine.go:82-84`、`internal/config/config.go:102-107`
- **问题描述**：
  1. **Seccomp 缺失**：代码中只有 `SecurityOpt: []string{"no-new-privileges"}`，没有配置 Seccomp 配置文件（如 `seccomp=unconfined` 或自定义 profile）
  2. **Capabilities 硬编码**：只保留了 `NET_BIND_SERVICE`，但任务要求"仅保留必要项"应该是可配置的
  3. **配置结构缺失**：`config.ContainerConfig` 中没有 `CapDrop`、`CapAdd`、`SecurityOpt` 字段

- **影响范围/风险**：
  1. 无法通过配置调整 Capabilities 白名单
  2. 缺少 Seccomp 防护，容器可能执行危险系统调用
  3. 不符合 B10 任务"Capabilities 白名单：Drop ALL，仅保留必要项"的要求（必要项应可配置）

- **修正建议**：
  在 `config.ContainerConfig` 中添加：
  ```go
  type ContainerConfig struct {
      // ... 现有字段
      CapDrop     []string `mapstructure:"cap_drop"`
      CapAdd      []string `mapstructure:"cap_add"`
      SecurityOpt []string `mapstructure:"security_opt"`
  }
  ```
  配置文件示例：
  ```yaml
  container:
    cap_drop: ["ALL"]
    cap_add: ["NET_BIND_SERVICE", "CHOWN", "SETUID", "SETGID"]
    security_opt: ["no-new-privileges:true", "seccomp=unconfined"]
  ```

---

#### [H4] tmpfs 挂载逻辑错误

- **文件**：`code/backend/internal/module/container/engine.go:86-90`
- **问题描述**：
  ```go
  if cfg.Security.ReadonlyRootfs {
      hostCfg.Tmpfs = map[string]string{
          "/tmp": "rw,noexec,nosuid,size=65536k",
      }
  }
  ```
  当前代码只在 `ReadonlyRootfs=true` 时才挂载 tmpfs，但实际上：
  1. 代码中 `ReadonlyRootfs` 硬编码为 `true`（第 81 行）
  2. 配置文件中 `readonly_rootfs: false`
  3. 配置项未被使用，导致 tmpfs 永远不会被挂载

- **影响范围/风险**：
  1. 如果 `ReadonlyRootfs=true` 但没有 tmpfs，容器无法写入 `/tmp`，很多程序会崩溃
  2. 如果 `ReadonlyRootfs=false`，tmpfs 不挂载，但此时容器可以写入根文件系统任意位置，失去安全防护

- **修正建议**：
  1. 修复配置注入问题（见 H2）
  2. 无论 `ReadonlyRootfs` 是否启用，都应该挂载 tmpfs 到 `/tmp`，只是权限不同：
     ```go
     hostCfg.Tmpfs = map[string]string{
         "/tmp": "rw,noexec,nosuid,size=65536k",
     }
     if cfg.Security.ReadonlyRootfs {
         hostCfg.ReadonlyRootfs = true
     }
     ```

---

#### [H5] 缺少资源配置验证

- **文件**：`code/backend/internal/module/container/engine.go:56-71`
- **问题描述**：
  `CreateContainer` 方法直接使用传入的 `cfg.Resources`，没有任何校验：
  - CPU 配额是否为负数或超出宿主机核心数
  - 内存是否为负数或超出宿主机可用内存
  - PID 限制是否合理（如 0 或负数）

- **影响范围/风险**：
  1. 恶意配置可能导致容器创建失败或行为异常
  2. 不符合 B10 验收标准"启动前校验资源配置合法性"和"超出宿主机资源上限时拒绝创建"

- **修正建议**：
  在 `CreateContainer` 开头添加校验逻辑：
  ```go
  if cfg.Resources != nil {
      if cfg.Resources.CPUQuota <= 0 || cfg.Resources.CPUQuota > 16 {
          return "", fmt.Errorf("invalid CPU quota: %.2f (must be 0-16)", cfg.Resources.CPUQuota)
      }
      if cfg.Resources.Memory <= 0 || cfg.Resources.Memory > 8*1024*1024*1024 {
          return "", fmt.Errorf("invalid memory: %d (must be 0-8GB)", cfg.Resources.Memory)
      }
      if cfg.Resources.PidsLimit <= 0 || cfg.Resources.PidsLimit > 1000 {
          return "", fmt.Errorf("invalid pids limit: %d (must be 0-1000)", cfg.Resources.PidsLimit)
      }
  }
  ```

---

#### [H6] model.ContainerConfig 缺少 Security 字段但代码中使用

- **文件**：`code/backend/internal/model/container.go:5-11`、`engine.go:78-89`
- **问题描述**：
  `engine.go` 第 78 行使用了 `cfg.Security`：
  ```go
  if cfg.Security != nil {
      containerCfg.User = cfg.Security.User
      hostCfg.ReadonlyRootfs = cfg.Security.ReadonlyRootfs
      // ...
  }
  ```
  但 `model.ContainerConfig` 结构体中没有 `Security` 字段，只有 `Resources` 字段。

- **影响范围/风险**：
  1. 代码无法编译通过（`cfg.Security` 未定义）
  2. 说明代码未经过基本的编译测试

- **修正建议**：
  在 `model.ContainerConfig` 中添加：
  ```go
  type ContainerConfig struct {
      Image     string
      Env       []string
      Ports     map[string]string
      Resources *ResourceLimits
      Security  *SecurityConfig  // 添加此字段
      Network   string
  }
  ```

---

### 🟡 中优先级

#### [M1] 配置文件中 CPU 配额单位不明确

- **文件**：`code/backend/configs/config.yaml:93`
- **问题描述**：
  ```yaml
  default_cpu_quota: 50000
  ```
  这个值的单位是什么？是微秒？是 CPU 核心数的 1000 倍？还是其他单位？没有注释说明。

- **影响范围/风险**：运维人员无法正确配置 CPU 限制
- **修正建议**：
  添加注释并修正值：
  ```yaml
  container:
    default_cpu_quota: 0.5  # CPU 核心数（0.5 = 半核）
    default_memory: 268435456  # 字节（256MB）
    default_pids_limit: 100  # 最大进程数
  ```

---

#### [M2] 内存配置使用魔法数字

- **文件**：`code/backend/configs/config.yaml:94`
- **问题描述**：
  ```yaml
  default_memory: 268435456
  ```
  `268435456` 是 256MB，但直接写数字不易读。

- **影响范围/风险**：配置文件可读性差，容易配置错误
- **修正建议**：
  虽然 YAML 不支持表达式，但可以添加注释：
  ```yaml
  default_memory: 268435456  # 256MB (256 * 1024 * 1024)
  ```
  或者在代码中支持字符串解析（如 "256MB"），但这需要额外的解析逻辑。

---

#### [M3] DefaultSecurityConfig 函数未被使用

- **文件**：`code/backend/internal/module/container/engine.go:75-81`（git diff 中的位置）
- **问题描述**：
  代码中定义了 `DefaultSecurityConfig()` 函数，但在 `CreateContainer` 中完全没有调用，安全配置直接硬编码在 `HostConfig` 中。

- **影响范围/风险**：
  1. 函数定义无意义，增加代码冗余
  2. 如果后续想使用默认配置，需要修改两处代码

- **修正建议**：
  删除 `DefaultSecurityConfig()` 函数，或者修改 `CreateContainer` 使用它：
  ```go
  if cfg.Security == nil {
      cfg.Security = DefaultSecurityConfig()
  }
  ```

---

#### [M4] 缺少容器名称生成逻辑

- **文件**：`code/backend/internal/module/container/engine.go:73`
- **问题描述**：
  ```go
  resp, err := e.cli.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, "")
  ```
  最后一个参数是容器名称，当前传空字符串，Docker 会自动生成随机名称（如 `silly_einstein`）。

- **影响范围/风险**：
  1. 无法通过名称快速识别容器属于哪个用户/题目
  2. 日志和监控中容器名称无意义

- **修正建议**：
  生成有意义的容器名称：
  ```go
  containerName := fmt.Sprintf("ctf-%s-%d", challengeID, userID)
  resp, err := e.cli.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, containerName)
  ```
  需要在 `ContainerConfig` 中添加 `Name` 字段。

---

#### [M5] 缺少磁盘 IO 限制实现

- **文件**：`code/backend/internal/model/container.go:18`
- **问题描述**：
  `ResourceLimits` 中定义了 `DiskQuota string` 字段，注释为"暂未使用"，但 B10 任务要求包含"磁盘 IO 限制：`--storage-opt` 参数"。

- **影响范围/风险**：
  1. 容器可以无限制写入磁盘，可能填满宿主机磁盘
  2. 不符合 B10 任务交付物要求

- **修正建议**：
  1. 如果当前阶段不实现磁盘限制，应在任务文档中明确说明延后到后续任务
  2. 如果要实现，需要配置 `storage-driver` 和 `storage-opt`（如 overlay2 + size 限制）

---

### 🟢 低优先级

#### [L1] 配置项命名不一致

- **文件**：`code/backend/internal/config/config.go:102-107`
- **问题描述**：
  配置结构体字段使用驼峰命名（`DefaultCPUQuota`），但 YAML 配置文件使用下划线命名（`default_cpu_quota`），虽然通过 `mapstructure` 标签映射，但不够直观。

- **影响范围/风险**：代码可读性略差
- **修正建议**：
  保持现状即可，这是 Go 项目的常见做法。如果要统一，可以考虑配置文件也使用驼峰（但 YAML 社区更推荐下划线）。

---

#### [L2] 缺少配置默认值的单元测试

- **文件**：`code/backend/internal/config/config.go:202-206`
- **问题描述**：
  `setDefaults` 函数中设置了容器配置的默认值，但没有单元测试验证这些默认值是否正确加载。

- **影响范围/风险**：配置加载失败时难以排查
- **修正建议**：
  添加单元测试：
  ```go
  func TestContainerConfigDefaults(t *testing.T) {
      cfg, err := Load("test")
      assert.NoError(t, err)
      assert.Equal(t, int64(50000), cfg.Container.DefaultCPUQuota)
      assert.Equal(t, int64(268435456), cfg.Container.DefaultMemory)
  }
  ```

---

#### [L3] run_as_user 配置项为空字符串的语义不明确

- **文件**：`code/backend/configs/config.yaml:97`
- **问题描述**：
  ```yaml
  run_as_user: ""
  ```
  空字符串表示什么？是以 root 运行，还是使用镜像默认用户？

- **影响范围/风险**：配置语义不清晰
- **修正建议**：
  添加注释说明：
  ```yaml
  run_as_user: ""  # 空字符串表示使用镜像默认用户，建议设置为 "1000:1000" 或 "nobody"
  ```

---

#### [L4] 缺少日志记录

- **文件**：`code/backend/internal/module/container/engine.go:56-92`
- **问题描述**：
  `CreateContainer` 方法中没有任何日志记录，无法追踪容器创建过程和配置参数。

- **影响范围/风险**：生产环境排查问题困难
- **修正建议**：
  添加日志：
  ```go
  func (e *Engine) CreateContainer(ctx context.Context, cfg *model.ContainerConfig) (string, error) {
      log.Info("creating container", "image", cfg.Image, "cpu", cfg.Resources.CPUQuota, "memory", cfg.Resources.Memory)
      // ... 现有逻辑
      log.Info("container created", "id", resp.ID)
      return resp.ID, nil
  }
  ```

---

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 6 |
| 🟡 中 | 5 |
| 🟢 低 | 4 |
| 合计 | 15 |

## 总体评价

本次提交实现了容器资源限制和安全加固的基本框架，但存在多个严重问题：

**核心问题**：
1. **配置系统完全失效**：定义了配置项但完全未使用，所有安全配置都是硬编码
2. **资源限制计算错误**：CPU 配额单位混乱，当前配置相当于 50000 核 CPU
3. **代码无法编译**：使用了未定义的 `cfg.Security` 字段

**架构偏离**：
- 违反了"配置外部化"原则，安全策略无法通过配置文件调整
- 不符合 B10 任务"资源配置可按题目自定义"的验收标准

**必须修复的问题**（阻塞验收）：
- H1: CPU 配额计算错误
- H2: 配置项未使用
- H3: 缺少 Seccomp 和可配置的 Capabilities
- H5: 缺少资源配置验证
- H6: model 结构体缺少 Security 字段

**建议**：
1. 优先修复 H6（编译错误）和 H2（配置注入）
2. 重新设计配置结构，将所有硬编码的安全配置移到配置文件
3. 添加资源配置验证逻辑
4. 补充单元测试验证配置加载和资源限制生效

修复后需要进行第 3 轮审查。
