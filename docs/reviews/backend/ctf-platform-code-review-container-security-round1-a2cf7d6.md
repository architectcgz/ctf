# CTF 平台代码 Review（container-security 第 1 轮）：B10 任务未实现

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | container-security |
| 轮次 | 第 1 轮 |
| 审查范围 | worktree agent-a273364c，最新 commit a2cf7d6 |
| 变更概述 | 检查 B10（容器资源限制与安全加固）的实现情况 |
| 审查基准 | `docs/tasks/backend-task-breakdown.md` B10 任务定义 |
| 审查日期 | 2026-03-05 |
| 上轮问题数 | N/A（首次审查） |

## 核心发现

**B10 任务尚未实现**。当前 worktree 中最新的 commit 是 B9 的修复（a2cf7d6），而 B10 要求的功能并未完成。

当前代码在 `internal/module/container/engine.go` 中硬编码了部分安全配置，但不符合 B10 的要求。

## 问题清单

### 🔴 高优先级

#### [H1] B10 任务未实现
- **文件**：整个项目
- **问题描述**：B10 任务定义的所有交付物均未实现
- **影响范围/风险**：
  - 资源限制无法按题目自定义
  - 安全配置硬编码在代码中，无法灵活调整
  - 缺少 Seccomp 配置
  - 缺少配置验证机制
- **修正建议**：按照 B10 任务定义完整实现以下内容

#### [H2] 缺少 SecurityConfig 模型定义
- **文件**：`internal/model/container.go`
- **问题描述**：模型中没有 `SecurityConfig` 结构体定义，但 `engine.go` 中引用了该类型
- **影响范围/风险**：代码无法编译通过
- **修正建议**：
```go
// SecurityConfig 安全配置
type SecurityConfig struct {
    Privileged      bool     // 是否特权模式（必须为 false）
    ReadonlyRootfs  bool     // 只读根文件系统
    NoNewPrivileges bool     // 禁止提升权限
    CapDrop         []string // 移除的 Capabilities
    CapAdd          []string // 添加的 Capabilities
    SecurityOpt     []string // 安全选项（如 seccomp, apparmor）
    User            string   // 运行用户（非 root）
}
```

#### [H3] 资源限制配置未外部化
- **文件**：`configs/config.example.yaml`
- **问题描述**：配置文件中只有超时配置，缺少资源限制的默认值配置
- **影响范围/风险**：
  - 无法通过配置文件调整默认资源限制
  - 不同环境（开发/生产）无法使用不同的资源配额
- **修正建议**：
```yaml
docker:
  image_pull_timeout: 10m
  container_timeout: 30s

  # 默认资源限制
  default_resources:
    cpu_quota: 0.5        # CPU 核心数
    memory: 268435456     # 256MB（字节）
    pids_limit: 100       # 最大进程数
    disk_quota: "1G"      # 磁盘配额（暂未实现）

  # 默认安全配置
  default_security:
    privileged: false
    readonly_rootfs: false
    no_new_privileges: true
    cap_drop: ["ALL"]
    cap_add: ["CHOWN", "SETUID", "SETGID", "NET_BIND_SERVICE"]
    seccomp_profile: "default"  # default | unconfined | custom
    user: ""  # 空表示使用镜像默认用户
```

#### [H4] 安全配置硬编码
- **文件**：`internal/module/container/engine.go:73-88`
- **问题描述**：安全配置直接硬编码在 `CreateContainer` 方法中
```go
// 当前代码（硬编码）
&container.HostConfig{
    Privileged:     false,
    ReadonlyRootfs: true,
    CapDrop:        []string{"ALL"},
    CapAdd:         []string{"NET_BIND_SERVICE"},
    SecurityOpt:    []string{"no-new-privileges"},
}
```
- **影响范围/风险**：
  - 所有容器使用相同的安全配置，无法按题目需求调整
  - 某些题目可能需要写文件系统，但被强制只读
  - 某些题目可能需要特定 Capabilities，但无法配置
- **修正建议**：
```go
// 应该从 cfg.Security 读取配置
hostCfg := &container.HostConfig{
    PortBindings: portBindings,
    Resources:    resources,
    NetworkMode:  container.NetworkMode(cfg.Network),
}

// 应用安全配置（如果提供）
if cfg.Security != nil {
    hostCfg.Privileged = cfg.Security.Privileged
    hostCfg.ReadonlyRootfs = cfg.Security.ReadonlyRootfs
    hostCfg.CapDrop = cfg.Security.CapDrop
    hostCfg.CapAdd = cfg.Security.CapAdd
    hostCfg.SecurityOpt = cfg.Security.SecurityOpt

    if cfg.Security.User != "" {
        containerCfg.User = cfg.Security.User
    }

    // 只读根文件系统需要挂载 tmpfs
    if cfg.Security.ReadonlyRootfs {
        hostCfg.Tmpfs = map[string]string{
            "/tmp": "rw,noexec,nosuid,size=65536k",
        }
    }
} else {
    // 使用默认安全配置
    hostCfg.Privileged = false
    hostCfg.CapDrop = []string{"ALL"}
    hostCfg.SecurityOpt = []string{"no-new-privileges"}
}

// 强制禁用特权模式（安全保障）
if hostCfg.Privileged {
    return "", fmt.Errorf("privileged mode is not allowed")
}
```

#### [H5] 缺少 Seccomp 配置
- **文件**：`internal/module/container/engine.go`
- **问题描述**：B10 要求支持 Seccomp 配置，但当前代码未实现
- **影响范围/风险**：
  - 无法限制容器可调用的系统调用
  - 容器可能执行危险的系统调用（如 reboot, mount）
- **修正建议**：
```go
// 在 SecurityConfig 中添加 Seccomp 支持
type SecurityConfig struct {
    // ... 其他字段
    SeccompProfile string // "default" | "unconfined" | 自定义 profile 路径
}

// 在 CreateContainer 中应用 Seccomp
if cfg.Security != nil && cfg.Security.SeccompProfile != "" {
    if cfg.Security.SeccompProfile == "unconfined" {
        hostCfg.SecurityOpt = append(hostCfg.SecurityOpt, "seccomp=unconfined")
    } else if cfg.Security.SeccompProfile != "default" {
        // 自定义 profile
        hostCfg.SecurityOpt = append(hostCfg.SecurityOpt,
            fmt.Sprintf("seccomp=%s", cfg.Security.SeccompProfile))
    }
    // "default" 不需要显式指定，Docker 会使用默认 profile
}
```

#### [H6] 缺少资源配置验证
- **文件**：`internal/module/container/engine.go`
- **问题描述**：B10 要求"启动前校验资源配置合法性，超出宿主机资源上限时拒绝创建"，但当前代码未实现
- **影响范围/风险**：
  - 可能创建超出宿主机能力的容器
  - 可能导致宿主机资源耗尽
  - 恶意用户可能通过配置超大资源限制进行 DoS 攻击
- **修正建议**：
```go
// 添加资源验证方法
func (e *Engine) ValidateResources(cfg *model.ResourceLimits, maxCPU float64, maxMemory int64) error {
    if cfg == nil {
        return nil
    }

    if cfg.CPUQuota <= 0 || cfg.CPUQuota > maxCPU {
        return fmt.Errorf("invalid CPU quota: %.2f (max: %.2f)", cfg.CPUQuota, maxCPU)
    }

    if cfg.Memory <= 0 || cfg.Memory > maxMemory {
        return fmt.Errorf("invalid memory: %d (max: %d)", cfg.Memory, maxMemory)
    }

    if cfg.PidsLimit <= 0 || cfg.PidsLimit > 1000 {
        return fmt.Errorf("invalid pids limit: %d (max: 1000)", cfg.PidsLimit)
    }

    return nil
}

// 在 CreateContainer 中调用验证
func (e *Engine) CreateContainer(ctx context.Context, cfg *model.ContainerConfig) (string, error) {
    // 验证资源配置
    if err := e.ValidateResources(cfg.Resources, e.maxCPU, e.maxMemory); err != nil {
        return "", err
    }

    // ... 继续创建容器
}
```

### 🟡 中优先级

#### [M1] ContainerConfig 缺少 Security 字段
- **文件**：`internal/model/container.go:8-14`
- **问题描述**：`ContainerConfig` 结构体中没有 `Security` 字段
```go
// 当前定义
type ContainerConfig struct {
    Image     string
    Env       []string
    Ports     map[string]string
    Resources *ResourceLimits
    Network   string
}
```
- **影响范围/风险**：无法传递安全配置到容器创建方法
- **修正建议**：
```go
type ContainerConfig struct {
    Image     string
    Env       []string
    Ports     map[string]string
    Resources *ResourceLimits
    Security  *SecurityConfig  // 新增
    Network   string
}
```

#### [M2] 配置结构体缺少资源限制字段
- **文件**：`internal/config/config.go`
- **问题描述**：配置结构体中缺少 Docker 资源限制和安全配置的字段定义
- **影响范围/风险**：无法从配置文件读取资源限制和安全配置
- **修正建议**：
```go
type Config struct {
    // ... 其他字段
    Docker DockerConfig `mapstructure:"docker"`
}

type DockerConfig struct {
    ImagePullTimeout time.Duration      `mapstructure:"image_pull_timeout"`
    ContainerTimeout time.Duration      `mapstructure:"container_timeout"`
    DefaultResources ResourceLimits     `mapstructure:"default_resources"`
    DefaultSecurity  SecurityConfig     `mapstructure:"default_security"`
    MaxCPU           float64            `mapstructure:"max_cpu"`
    MaxMemory        int64              `mapstructure:"max_memory"`
}

type ResourceLimits struct {
    CPUQuota  float64 `mapstructure:"cpu_quota"`
    Memory    int64   `mapstructure:"memory"`
    PidsLimit int64   `mapstructure:"pids_limit"`
    DiskQuota string  `mapstructure:"disk_quota"`
}

type SecurityConfig struct {
    Privileged       bool     `mapstructure:"privileged"`
    ReadonlyRootfs   bool     `mapstructure:"readonly_rootfs"`
    NoNewPrivileges  bool     `mapstructure:"no_new_privileges"`
    CapDrop          []string `mapstructure:"cap_drop"`
    CapAdd           []string `mapstructure:"cap_add"`
    SeccompProfile   string   `mapstructure:"seccomp_profile"`
    User             string   `mapstructure:"user"`
}
```

#### [M3] Engine 初始化缺少资源上限参数
- **文件**：`internal/module/container/engine.go:26-44`
- **问题描述**：`NewEngine` 方法缺少 `maxCPU` 和 `maxMemory` 参数，无法进行资源验证
- **影响范围/风险**：无法实现资源配置验证功能
- **修正建议**：
```go
type Engine struct {
    cli              *client.Client
    imagePullTimeout time.Duration
    containerTimeout time.Duration
    maxCPU           float64  // 新增
    maxMemory        int64    // 新增
}

func NewEngine(host string, imagePullTimeout, containerTimeout time.Duration,
               maxCPU float64, maxMemory int64) (*Engine, error) {
    // ... 创建客户端

    return &Engine{
        cli:              cli,
        imagePullTimeout: imagePullTimeout,
        containerTimeout: containerTimeout,
        maxCPU:           maxCPU,
        maxMemory:        maxMemory,
    }, nil
}
```

### 🟢 低优先级

#### [L1] 配置文件缺少注释说明
- **文件**：`configs/config.example.yaml`
- **问题描述**：配置项缺少详细的说明和单位标注
- **影响范围/风险**：用户可能不清楚如何正确配置资源限制
- **修正建议**：
```yaml
docker:
  image_pull_timeout: 10m  # 镜像拉取超时时间
  container_timeout: 30s   # 容器操作超时时间

  # 宿主机资源上限（用于验证容器配置）
  max_cpu: 4.0             # 最大 CPU 核心数
  max_memory: 8589934592   # 最大内存 8GB（字节）

  # 默认资源限制（单个容器）
  default_resources:
    cpu_quota: 0.5         # CPU 核心数（0.5 = 50% 单核）
    memory: 268435456      # 内存限制 256MB（字节）
    pids_limit: 100        # 最大进程数
    disk_quota: "1G"       # 磁盘配额（暂未实现）

  # 默认安全配置
  default_security:
    privileged: false                    # 禁用特权模式（强制）
    readonly_rootfs: false               # 只读根文件系统（按题目配置）
    no_new_privileges: true              # 禁止提升权限
    cap_drop: ["ALL"]                    # 移除所有 Capabilities
    cap_add:                             # 添加必要的 Capabilities
      - "CHOWN"
      - "SETUID"
      - "SETGID"
      - "NET_BIND_SERVICE"
    seccomp_profile: "default"           # Seccomp 配置：default | unconfined
    user: ""                             # 运行用户（空=使用镜像默认）
```

#### [L2] 缺少资源限制测试用例
- **文件**：`internal/module/container/engine_test.go`（不存在）
- **问题描述**：B10 验收标准要求"容器内存超限后被 OOM Kill"、"Fork Bomb 不影响其他容器"，但缺少对应的测试
- **影响范围/风险**：无法验证资源限制是否生效
- **修正建议**：添加集成测试验证资源限制功能

#### [L3] 缺少安全配置文档
- **文件**：`docs/architecture/backend/`
- **问题描述**：缺少容器安全配置的架构文档说明
- **影响范围/风险**：开发者和运维人员不清楚安全配置的设计意图和使用方法
- **修正建议**：补充容器安全配置的架构文档

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 6 |
| 🟡 中 | 3 |
| 🟢 低 | 3 |
| 合计 | 12 |

## 总体评价

**B10 任务尚未实现**。当前代码仅包含 B9 的基础 Docker 引擎封装和部分硬编码的安全配置，不符合 B10 的要求。

### 主要缺陷

1. **配置外部化缺失**：资源限制和安全配置硬编码在代码中，无法通过配置文件调整
2. **模型定义不完整**：缺少 `SecurityConfig` 结构体定义
3. **安全功能不足**：缺少 Seccomp 配置、资源验证、特权模式强制禁用
4. **灵活性不足**：无法按题目需求自定义资源限制和安全配置

### 修复建议

按照 B10 任务定义，依次完成：

1. **定义模型**：在 `internal/model/container.go` 中添加 `SecurityConfig` 结构体
2. **外部化配置**：在 `configs/config.example.yaml` 和 `internal/config/config.go` 中添加资源和安全配置
3. **重构 Engine**：修改 `CreateContainer` 方法，从 `cfg.Security` 读取配置而非硬编码
4. **添加验证**：实现 `ValidateResources` 方法，验证资源配置合法性
5. **支持 Seccomp**：添加 Seccomp profile 配置支持
6. **强制安全**：确保特权模式被强制禁用

### 下一步

建议 backend-engineer 重新实现 B10 任务，完成上述所有高优先级和中优先级问题的修复。
