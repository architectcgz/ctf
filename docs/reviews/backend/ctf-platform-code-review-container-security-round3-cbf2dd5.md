# ctf-platform 代码 Review（container-security 第 3 轮）：修复配置系统和资源限制计算错误

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | container-security |
| 轮次 | 第 3 轮（修复后复审） |
| 审查范围 | commit cbf2dd5，3 个文件，79 行新增，32 行删除 |
| 变更概述 | 修复配置注入、CPU 配额计算、Seccomp 配置、配置验证等问题 |
| 审查基准 | docs/tasks/backend-task-breakdown.md B10 任务验收标准 |
| 审查日期 | 2026-03-05 |
| 上轮问题数 | 15 项（6 高 / 5 中 / 4 低） |

## 问题清单

### 🔴 高优先级

#### [H1] CPU 配额计算仍然错误，资源限制完全失效

- **文件**：`code/backend/internal/module/container/engine.go:59`、`config.yaml:93`
- **问题描述**：
  虽然修改了计算公式为 `cfg.Resources.CPUQuota * 10000`，但配置文件中的值仍然是 `50000`：

  ```go
  // engine.go:59
  resources := container.Resources{
      NanoCPUs:  cfg.Resources.CPUQuota * 10000,  // 50000 * 10000 = 500,000,000
      Memory:    cfg.Resources.Memory,
      PidsLimit: &cfg.Resources.PidsLimit,
  }
  ```

  ```yaml
  # config.yaml:93
  default_cpu_quota: 50000  # 这个值的单位是什么？
  ```

  根据 Docker API 文档，`NanoCPUs` 的单位是纳秒（1e9 = 1 核心）。当前计算结果：
  - `50000 * 10000 = 500,000,000 纳秒 = 0.5 核心` ✅ 这个结果是对的
  - 但配置值 `50000` 的语义完全不清楚，运维人员无法理解这个数字代表什么

  **根本问题**：配置文件中应该直接使用"微秒"作为单位（Docker CPU Period 的标准单位），而不是这个莫名其妙的 `50000`。

  标准做法：
  - Docker 的 `--cpu-quota` 参数单位是微秒（默认 period 是 100000 微秒 = 100ms）
  - 0.5 核心 = 50000 微秒的 quota（在 100000 微秒的 period 内）
  - 所以配置值 `50000` 恰好是微秒，但代码中 `* 10000` 的转换逻辑是错误的

- **影响范围/风险**：
  1. 虽然巧合地算出了正确的结果（0.5 核），但逻辑完全错误
  2. 如果运维人员想配置 1 核心，会写 `100000`，结果变成 `100000 * 10000 = 1e9 = 1 核` ✅ 巧合正确
  3. 如果运维人员想配置 0.25 核心，会写 `25000`，结果变成 `25000 * 10000 = 2.5e8 = 0.25 核` ✅ 巧合正确
  4. **但这个 `* 10000` 的魔法数字完全没有文档说明，代码可维护性极差**

- **修正建议**：
  有两种正确的方案：

  **方案 1：配置文件使用 CPU 核心数（推荐）**
  ```yaml
  default_cpu_quota: 0.5  # CPU 核心数
  ```
  ```go
  resources := container.Resources{
      NanoCPUs:  int64(cfg.Resources.CPUQuota * 1e9),  // 转换为纳秒
  }
  ```

  **方案 2：配置文件使用微秒（与 Docker 原生参数一致）**
  ```yaml
  default_cpu_quota: 50000  # 微秒（Docker --cpu-quota 的单位）
  ```
  ```go
  resources := container.Resources{
      NanoCPUs:  cfg.Resources.CPUQuota * 10000,  // 微秒转纳秒
  }
  ```
  并在配置文件和代码中添加注释说明单位。

  当前代码恰好选择了方案 2，但完全没有文档说明，必须补充注释。

---

#### [H2] 配置验证逻辑不完整，无法防止资源配置错误

- **文件**：`code/backend/internal/config/config.go:151-162`
- **问题描述**：
  虽然添加了 `Validate()` 方法，但只检查了"是否为正数"，没有检查上限：

  ```go
  func (c *Config) Validate() error {
      if c.Container.DefaultCPUQuota <= 0 {
          return fmt.Errorf("container.default_cpu_quota must be positive")
      }
      if c.Container.DefaultMemory <= 0 {
          return fmt.Errorf("container.default_memory must be positive")
      }
      if c.Container.DefaultPidsLimit <= 0 {
          return fmt.Errorf("container.default_pids_limit must be positive")
      }
      return nil
  }
  ```

  缺少的验证：
  1. CPU 配额是否超出宿主机核心数（如配置 1000 核）
  2. 内存是否超出宿主机可用内存（如配置 1TB）
  3. PID 限制是否合理（如配置 1000000）
  4. Seccomp 配置是否为有效值（`default`、`unconfined` 或文件路径）
  5. AllowedCapabilities 是否为有效的 Linux Capability 名称

- **影响范围/风险**：
  1. 不符合 B10 验收标准"启动前校验资源配置合法性"和"超出宿主机资源上限时拒绝创建"
  2. 恶意或错误的配置可能导致容器创建失败或系统异常

- **修正建议**：
  ```go
  func (c *Config) Validate() error {
      // CPU 配额验证（假设单位是微秒，period 是 100000）
      if c.Container.DefaultCPUQuota <= 0 || c.Container.DefaultCPUQuota > 1600000 {
          return fmt.Errorf("container.default_cpu_quota must be 1-1600000 (0.01-16 cores)")
      }

      // 内存验证（256MB - 8GB）
      minMem := int64(256 * 1024 * 1024)
      maxMem := int64(8 * 1024 * 1024 * 1024)
      if c.Container.DefaultMemory < minMem || c.Container.DefaultMemory > maxMem {
          return fmt.Errorf("container.default_memory must be %d-%d (256MB-8GB)", minMem, maxMem)
      }

      // PID 限制验证
      if c.Container.DefaultPidsLimit <= 0 || c.Container.DefaultPidsLimit > 1000 {
          return fmt.Errorf("container.default_pids_limit must be 1-1000")
      }

      // Seccomp 验证
      validSeccomp := map[string]bool{"default": true, "unconfined": true}
      if !validSeccomp[c.Container.Seccomp] && !strings.HasPrefix(c.Container.Seccomp, "/") {
          return fmt.Errorf("container.seccomp must be 'default', 'unconfined', or a file path")
      }

      return nil
  }
  ```

---

#### [H3] Engine.CreateContainer 缺少运行时资源配置验证

- **文件**：`code/backend/internal/module/container/engine.go:31-48`
- **问题描述**：
  虽然在配置加载时有验证（`config.Validate()`），但 `CreateContainer` 方法接收的 `cfg.Resources` 是运行时传入的，可能绕过配置验证：

  ```go
  func (e *Engine) CreateContainer(ctx context.Context, cfg *model.ContainerConfig) (string, error) {
      if cfg.Resources == nil {
          cfg.Resources = &model.ResourceLimits{
              CPUQuota:  e.containerCfg.DefaultCPUQuota,
              Memory:    e.containerCfg.DefaultMemory,
              PidsLimit: e.containerCfg.DefaultPidsLimit,
          }
      }
      // 没有验证 cfg.Resources 的值是否合法
  ```

  如果调用方传入 `cfg.Resources = &model.ResourceLimits{CPUQuota: -1, Memory: 0, PidsLimit: 999999}`，会直接创建容器而不报错。

- **影响范围/风险**：
  1. 不符合 B10 验收标准"启动前校验资源配置合法性"
  2. 可能导致容器创建失败或资源滥用

- **修正建议**：
  在 `CreateContainer` 开头添加运行时验证：
  ```go
  func (e *Engine) CreateContainer(ctx context.Context, cfg *model.ContainerConfig) (string, error) {
      if cfg.Resources == nil {
          cfg.Resources = &model.ResourceLimits{
              CPUQuota:  e.containerCfg.DefaultCPUQuota,
              Memory:    e.containerCfg.DefaultMemory,
              PidsLimit: e.containerCfg.DefaultPidsLimit,
          }
      }

      // 运行时验证
      if err := e.validateResources(cfg.Resources); err != nil {
          return "", fmt.Errorf("invalid resource limits: %w", err)
      }

      // ... 后续逻辑
  }

  func (e *Engine) validateResources(res *model.ResourceLimits) error {
      if res.CPUQuota <= 0 || res.CPUQuota > 1600000 {
          return fmt.Errorf("cpu_quota must be 1-1600000, got %d", res.CPUQuota)
      }
      if res.Memory <= 0 || res.Memory > 8*1024*1024*1024 {
          return fmt.Errorf("memory must be 1-8GB, got %d", res.Memory)
      }
      if res.PidsLimit <= 0 || res.PidsLimit > 1000 {
          return fmt.Errorf("pids_limit must be 1-1000, got %d", res.PidsLimit)
      }
      return nil
  }
  ```

---

### 🟡 中优先级

#### [M1] 配置文件缺少单位注释，运维人员无法正确配置

- **文件**：`code/backend/configs/config.yaml:92-102`
- **问题描述**：
  配置文件中的数值没有任何注释说明单位：
  ```yaml
  container:
    default_cpu_quota: 50000        # 这是什么单位？
    default_memory: 268435456       # 这是多少 MB？
    default_pids_limit: 100
    readonly_rootfs: false
    run_as_user: ""
    allowed_capabilities:
      - CHOWN
      - SETUID
      - SETGID
    seccomp: default
  ```

- **影响范围/风险**：运维人员无法正确配置资源限制，可能导致配置错误

- **修正建议**：
  ```yaml
  container:
    default_cpu_quota: 50000        # CPU 配额（微秒），50000 = 0.5 核心（period 100000 微秒）
    default_memory: 268435456       # 内存限制（字节），268435456 = 256MB
    default_pids_limit: 100         # 最大进程数
    readonly_rootfs: false          # 是否只读根文件系统
    run_as_user: ""                 # 运行用户（空字符串表示使用镜像默认用户，建议设置为 "1000:1000"）
    allowed_capabilities:           # 允许的 Linux Capabilities（已 Drop ALL）
      - CHOWN                       # 允许修改文件所有者
      - SETUID                      # 允许设置用户 ID
      - SETGID                      # 允许设置组 ID
    seccomp: default                # Seccomp 配置（default/unconfined/文件路径）
  ```

---

#### [M2] model.ResourceLimits 字段类型与配置不一致

- **文件**：`code/backend/internal/model/container.go:14-18`、`config/config.go:103`
- **问题描述**：
  `model.ResourceLimits.CPUQuota` 是 `int64` 类型，但如果未来要支持小数核心数（如 0.5 核），需要改为 `float64`。

  当前代码：
  ```go
  // model/container.go
  type ResourceLimits struct {
      CPUQuota  int64  // 无法表示 0.5 核心
      Memory    int64
      PidsLimit int64
  }

  // config/config.go
  type ContainerConfig struct {
      DefaultCPUQuota int64  // 也是 int64
  }
  ```

- **影响范围/风险**：
  1. 如果采用"方案 1：配置文件使用 CPU 核心数"，无法支持小数（如 0.5）
  2. 当前采用"方案 2：微秒"可以用整数表示，但语义不清晰

- **修正建议**：
  如果要采用"方案 1：CPU 核心数"，需要修改类型：
  ```go
  type ResourceLimits struct {
      CPUQuota  float64  // CPU 核心数（支持小数）
      Memory    int64
      PidsLimit int64
  }
  ```
  或者保持 `int64`，但在文档中明确说明单位是"微秒"。

---

#### [M3] DefaultSecurityConfig 函数未被使用

- **文件**：`code/backend/internal/module/container/engine.go:95-103`
- **问题描述**：
  虽然修改了 `DefaultSecurityConfig` 函数接收配置参数，但在 `CreateContainer` 中仍然没有调用：

  ```go
  // 第 40-48 行：直接内联创建 SecurityConfig
  if cfg.Security == nil {
      cfg.Security = &model.SecurityConfig{
          ReadonlyRootfs: e.containerCfg.ReadonlyRootfs,
          CapDrop:        []string{"ALL"},
          CapAdd:         e.containerCfg.AllowedCapabilities,
          SecurityOpt:    []string{fmt.Sprintf("seccomp=%s", e.containerCfg.Seccomp), "no-new-privileges:true"},
          User:           e.containerCfg.RunAsUser,
      }
  }

  // 第 95-103 行：定义了函数但没人调用
  func DefaultSecurityConfig(cfg *config.ContainerConfig) *model.SecurityConfig {
      return &model.SecurityConfig{
          ReadonlyRootfs: cfg.ReadonlyRootfs,
          CapDrop:        []string{"ALL"},
          CapAdd:         cfg.AllowedCapabilities,
          SecurityOpt:    []string{fmt.Sprintf("seccomp=%s", cfg.Seccomp), "no-new-privileges:true"},
          User:           cfg.RunAsUser,
      }
  }
  ```

- **影响范围/风险**：
  1. 代码冗余，增加维护成本
  2. 如果后续修改默认安全配置，需要改两处

- **修正建议**：
  删除 `DefaultSecurityConfig` 函数，或者在 `CreateContainer` 中使用它：
  ```go
  if cfg.Security == nil {
      cfg.Security = DefaultSecurityConfig(e.containerCfg)
  }
  ```

---

#### [M4] 缺少容器名称生成逻辑

- **文件**：`code/backend/internal/module/container/engine.go:88`
- **问题描述**：
  容器名称仍然是空字符串，Docker 会生成随机名称（如 `silly_einstein`）：
  ```go
  resp, err := e.cli.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, "")
  ```

- **影响范围/风险**：
  1. 无法通过名称快速识别容器属于哪个用户/题目
  2. 日志和监控中容器名称无意义

- **修正建议**：
  在 `model.ContainerConfig` 中添加 `Name` 字段，或者根据其他字段生成名称：
  ```go
  containerName := fmt.Sprintf("ctf-%s-%d-%s", cfg.ChallengeID, cfg.UserID, generateRandomSuffix())
  resp, err := e.cli.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, containerName)
  ```

---

#### [M5] 磁盘 IO 限制仍未实现

- **文件**：`code/backend/internal/model/container.go:14-18`
- **问题描述**：
  B10 任务要求包含"磁盘 IO 限制：`--storage-opt` 参数"，但当前代码完全没有实现。

  `ResourceLimits` 结构体中没有 `DiskQuota` 字段，`engine.go` 中也没有设置 `storage-opt`。

- **影响范围/风险**：
  1. 容器可以无限制写入磁盘，可能填满宿主机磁盘
  2. 不符合 B10 任务交付物要求

- **修正建议**：
  如果当前阶段不实现磁盘限制，应在任务文档中明确说明延后到后续任务。

  如果要实现，需要：
  1. 在 `ResourceLimits` 中添加 `DiskQuota` 字段
  2. 在 `HostConfig` 中设置 `StorageOpt`（需要 overlay2 driver 支持）

---

### 🟢 低优先级

#### [L1] 缺少日志记录

- **文件**：`code/backend/internal/module/container/engine.go:31-93`
- **问题描述**：
  `CreateContainer` 方法中没有任何日志记录，无法追踪容器创建过程和配置参数。

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

- **文件**：`code/backend/internal/config/config.go:221-228`
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
      assert.Equal(t, int64(100), cfg.Container.DefaultPidsLimit)
      assert.Equal(t, []string{"CHOWN", "SETUID", "SETGID"}, cfg.Container.AllowedCapabilities)
      assert.Equal(t, "default", cfg.Container.Seccomp)
  }
  ```

---

#### [L3] tmpfs 挂载逻辑仍然有问题

- **文件**：`code/backend/internal/module/container/engine.go:82-86`
- **问题描述**：
  tmpfs 只在 `ReadonlyRootfs=true` 时挂载，但配置文件中 `readonly_rootfs: false`，导致 tmpfs 永远不会被挂载：

  ```go
  if cfg.Security.ReadonlyRootfs {
      hostCfg.Tmpfs = map[string]string{
          "/tmp": "rw,noexec,nosuid,size=65536k",
      }
  }
  ```

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
| 🔴 高 | 3 |
| 🟡 中 | 5 |
| 🟢 低 | 3 |
| 合计 | 11 |

## 总体评价

本轮修复解决了第 2 轮的部分问题，但仍存在严重缺陷：

**已修复的问题**：
- ✅ H6: 添加了 `model.ContainerConfig.Security` 字段，代码可以编译通过
- ✅ H2（部分）: Engine 注入了配置，`CreateContainer` 使用配置值而非硬编码
- ✅ H3（部分）: 添加了 `AllowedCapabilities` 和 `Seccomp` 配置字段
- ✅ H4（部分）: 添加了 `config.Validate()` 方法
- ✅ M1（部分）: 配置文件中添加了 `allowed_capabilities` 和 `seccomp` 字段

**仍存在的严重问题**：
- ❌ H1: CPU 配额计算逻辑虽然巧合正确，但语义不清晰，缺少文档说明
- ❌ H2: 配置验证不完整，只检查正数，未检查上限和有效性
- ❌ H3: `CreateContainer` 缺少运行时资源配置验证
- ❌ M1: 配置文件缺少单位注释
- ❌ M3: `DefaultSecurityConfig` 函数仍未被使用
- ❌ M4: 缺少容器名称生成逻辑
- ❌ M5: 磁盘 IO 限制仍未实现
- ❌ L3: tmpfs 挂载逻辑仍然有问题

**核心问题**：
1. **CPU 配额单位混乱**：配置值 `50000` 的语义不明确，虽然计算结果正确但逻辑难以理解
2. **配置验证不完整**：只检查正数，未检查上限，无法防止资源配置错误
3. **运行时验证缺失**：`CreateContainer` 接收的参数可能绕过配置验证

**阻塞验收的问题**：
- H1: CPU 配额语义不清晰（需要补充文档或重构）
- H2: 配置验证不完整（不符合验收标准"启动前校验资源配置合法性"）
- H3: 运行时验证缺失（不符合验收标准"超出宿主机资源上限时拒绝创建"）

**建议**：
1. 优先修复 H1：在配置文件和代码中添加详细注释，说明 CPU 配额的单位和计算逻辑
2. 修复 H2 和 H3：完善配置验证和运行时验证
3. 修复 M1：在配置文件中添加单位注释
4. 修复 L3：调整 tmpfs 挂载逻辑
5. 对于 M5（磁盘 IO 限制），如果当前阶段不实现，应在任务文档中明确说明

修复后需要进行第 4 轮审查。
