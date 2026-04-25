# CTF 平台代码 Review（network-isolation 第 1 轮）：容器网络隔离功能实现

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | network-isolation |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | commit bd16abe，9 个文件，596 行新增 |
| 变更概述 | 实现容器网络隔离功能，包括网络管理、端口分配、网络配置 |
| 审查基准 | docs/tasks/backend-task-breakdown.md 中的 B11 任务定义 |
| 审查日期 | 2026-03-05 |
| 上轮问题数 | - |

## 问题清单

### 🔴 高优先级

#### [H1] 端口管理器缺少并发安全的端口占用检测
- **文件**：`internal/module/container/port_manager.go:58-66`
- **问题描述**：`isPortAvailable` 方法存在 TOCTOU（Time-of-Check-Time-of-Use）竞争条件。在检查端口可用后到实际使用前，端口可能被其他进程占用。且该方法在持有锁的情况下执行网络 IO，会严重影响性能。
- **影响范围/风险**：
  - 高并发场景下可能分配到已被占用的端口
  - 持有锁期间执行 `net.Listen` 会阻塞其他端口分配请求
  - 可能导致容器启动失败
- **修正建议**：
  1. 将 `isPortAvailable` 检查移到锁外执行
  2. 采用"先分配后验证"策略：分配端口 → 释放锁 → 验证可用性 → 如失败则重试
  3. 或者依赖 Docker 的端口绑定失败机制，在容器创建失败时回收端口

```go
// 建议实现
func (pm *PortManager) AllocatePort() (int, error) {
    for i := 0; i < 100; i++ {
        port := pm.selectPort() // 在锁内选择端口

        // 锁外验证
        if pm.isPortAvailable(port) {
            pm.mu.Lock()
            pm.usedPorts[port] = true
            pm.mu.Unlock()
            return port, nil
        }
    }
    return 0, fmt.Errorf("无法分配可用端口")
}

func (pm *PortManager) selectPort() int {
    pm.mu.Lock()
    defer pm.mu.Unlock()

    for i := 0; i < 100; i++ {
        port := pm.rangeStart + rand.Intn(pm.rangeEnd-pm.rangeStart)
        if !pm.usedPorts[port] {
            return port
        }
    }
    return 0
}
```

#### [H2] 端口范围配置冲突和硬编码
- **文件**：
  - `internal/config/config.go:16-20`（端口范围 30000-40000）
  - `code/backend/internal/config/config.go:102`（无端口配置）
  - `code/backend/configs/config.yaml`（无端口配置）
- **问题描述**：
  1. 存在两个独立的 `config.go` 文件（`internal/config/config.go` 和 `code/backend/internal/config/config.go`），配置结构不一致
  2. 端口范围在 `internal/config/config.go` 中硬编码为 30000-40000，未通过配置文件注入
  3. `code/backend/configs/config.yaml` 中没有端口范围配置项
- **影响范围/风险**：
  - 配置管理混乱，不清楚哪个配置文件生效
  - 端口范围无法通过配置文件调整，违反外部化原则
  - 不同环境（dev/prod）无法使用不同端口范围
- **修正建议**：
  1. 删除 `internal/config/config.go`，统一使用 `code/backend/internal/config/config.go`
  2. 在 `code/backend/internal/config/config.go` 的 `ContainerConfig` 中添加端口配置
  3. 在 `config.yaml` 和 `setDefaults` 中添加端口范围配置

```yaml
# config.yaml
container:
  port_range_start: 30000
  port_range_end: 40000
```

```go
// code/backend/internal/config/config.go
type ContainerConfig struct {
    DefaultCPUQuota  int64  `mapstructure:"default_cpu_quota"`
    DefaultMemory    int64  `mapstructure:"default_memory"`
    DefaultPidsLimit int64  `mapstructure:"default_pids_limit"`
    ReadonlyRootfs   bool   `mapstructure:"readonly_rootfs"`
    RunAsUser        string `mapstructure:"run_as_user"`
    PortRangeStart   int    `mapstructure:"port_range_start"`
    PortRangeEnd     int    `mapstructure:"port_range_end"`
}
```

#### [H3] 网络名称可能冲突
- **文件**：`internal/module/container/network_test.go:22`
- **问题描述**：网络名称格式为 `ctf-{challenge_id}-{instance_id}`，但测试代码中使用 `ctf-test-1-1` 这种固定名称，且 `CreateNetwork` 方法虽然设置了 `CheckDuplicate: true`，但没有处理重名冲突的逻辑。
- **影响范围/风险**：
  - 如果实例重建或测试重复运行，会因网络名称冲突导致创建失败
  - 没有清理孤儿网络的机制，可能导致网络资源泄漏
- **修正建议**：
  1. 网络名称添加时间戳或随机后缀确保唯一性：`ctf-{challenge_id}-{instance_id}-{timestamp}`
  2. 在 `CreateNetwork` 前先检查同名网络是否存在，如存在则先删除
  3. 实现定时清理孤儿网络的机制（B11 任务要求）

```go
func (e *Engine) CreateNetwork(ctx context.Context, name string) (string, error) {
    // 检查是否存在同名网络
    networks, err := e.cli.NetworkList(ctx, types.NetworkListOptions{})
    if err != nil {
        return "", err
    }

    for _, net := range networks {
        if net.Name == name {
            // 删除旧网络
            if err := e.cli.NetworkRemove(ctx, net.ID); err != nil {
                return "", fmt.Errorf("删除旧网络失败: %w", err)
            }
        }
    }

    resp, err := e.cli.NetworkCreate(ctx, name, types.NetworkCreate{
        Driver: "bridge",
        CheckDuplicate: true,
    })
    return resp.ID, err
}
```

#### [H4] 容器创建时未连接到指定网络
- **文件**：`internal/module/container/engine.go:92`
- **问题描述**：`CreateContainer` 方法中使用 `NetworkMode: container.NetworkMode(cfg.Network)` 设置网络，但这只是设置了网络模式，并未真正将容器连接到指定的 Docker Network。根据 Docker SDK 文档，应该在 `NetworkingConfig` 中指定网络连接。
- **影响范围/风险**：
  - 容器可能无法正确连接到隔离网络
  - 网络隔离功能失效，容器间可能互通
  - 测试可能通过但实际功能不正确
- **修正建议**：
  使用 `NetworkingConfig` 参数正确连接网络

```go
func (e *Engine) CreateContainer(ctx context.Context, cfg *model.ContainerConfig) (string, error) {
    // ... 前面代码保持不变 ...

    // 构建网络配置
    networkConfig := &network.NetworkingConfig{}
    if cfg.Network != "" {
        networkConfig.EndpointsConfig = map[string]*network.EndpointSettings{
            cfg.Network: {},
        }
    }

    // 创建容器时传入网络配置
    resp, err := e.cli.ContainerCreate(ctx, containerCfg, hostCfg, networkConfig, nil, "")
    if err != nil {
        return "", err
    }

    return resp.ID, nil
}
```

### 🟡 中优先级

#### [M1] 端口管理器未初始化随机数种子
- **文件**：`internal/module/container/port_manager.go:33`
- **问题描述**：`AllocatePort` 方法使用 `rand.Intn()` 生成随机端口，但未初始化随机数种子，导致每次程序启动时端口分配序列相同。
- **影响范围/风险**：
  - 多个实例同时启动时可能分配到相同端口
  - 端口分配可预测，存在安全隐患
- **修正建议**：
  1. 在 `NewPortManager` 中初始化随机数种子
  2. 或使用 `crypto/rand` 生成更安全的随机数

```go
func NewPortManager(start, end int) *PortManager {
    rand.Seed(time.Now().UnixNano())
    return &PortManager{
        rangeStart: start,
        rangeEnd:   end,
        usedPorts:  make(map[int]bool),
    }
}
```

#### [M2] 缺少端口范围合法性校验
- **文件**：`internal/module/container/port_manager.go:19-24`
- **问题描述**：`NewPortManager` 未校验端口范围的合法性（start < end，端口在 1-65535 范围内）。
- **影响范围/风险**：
  - 配置错误时可能导致死循环或 panic
  - 端口范围过小时分配失败率高
- **修正建议**：

```go
func NewPortManager(start, end int) *PortManager {
    if start < 1 || end > 65535 || start >= end {
        panic(fmt.Sprintf("无效的端口范围: %d-%d", start, end))
    }
    if end - start < 100 {
        log.Warn("端口范围过小，建议至少 1000 个端口")
    }
    rand.Seed(time.Now().UnixNano())
    return &PortManager{
        rangeStart: start,
        rangeEnd:   end,
        usedPorts:  make(map[int]bool),
    }
}
```

#### [M3] 端口分配失败时错误信息不明确
- **文件**：`internal/module/container/port_manager.go:47`
- **问题描述**：端口分配失败时只返回"无法分配可用端口"，未说明原因（端口耗尽、系统限制等）。
- **影响范围/风险**：
  - 排查问题困难
  - 无法区分是端口耗尽还是系统问题
- **修正建议**：

```go
func (pm *PortManager) AllocatePort() (int, error) {
    pm.mu.Lock()
    usedCount := len(pm.usedPorts)
    totalRange := pm.rangeEnd - pm.rangeStart
    pm.mu.Unlock()

    if usedCount >= totalRange {
        return 0, fmt.Errorf("端口池已耗尽 (%d/%d)", usedCount, totalRange)
    }

    // ... 分配逻辑 ...

    return 0, fmt.Errorf("尝试 100 次后仍无法分配可用端口（已用: %d/%d）", usedCount, totalRange)
}
```

#### [M4] 网络配置模型未使用
- **文件**：`internal/model/container.go:15-19`
- **问题描述**：定义了 `NetworkConfig` 结构体，但在实际代码中未使用，`ContainerConfig.Network` 直接使用 string 类型。
- **影响范围/风险**：
  - 代码不一致，容易混淆
  - 未来扩展网络配置时需要重构
- **修正建议**：
  1. 如果暂不需要复杂网络配置，删除 `NetworkConfig` 结构体
  2. 或者将 `ContainerConfig.Network` 改为 `*NetworkConfig` 类型

#### [M5] 测试用例缺少清理逻辑
- **文件**：`internal/module/container/network_test.go:79-80`
- **问题描述**：`TestContainerWithNetwork` 使用 `defer` 清理资源，但如果测试失败（`t.Fatalf`），defer 不会执行，导致资源泄漏。
- **影响范围/风险**：
  - 测试失败后残留容器和网络
  - 影响后续测试运行
- **修正建议**：

```go
func TestContainerWithNetwork(t *testing.T) {
    // ... 初始化代码 ...

    // 创建网络
    networkID, err := engine.CreateNetwork(ctx, networkName)
    if err != nil {
        t.Fatalf("创建网络失败: %v", err)
    }
    // 确保清理，即使测试失败
    t.Cleanup(func() {
        engine.RemoveNetwork(context.Background(), networkID)
    })

    // ... 其余代码 ...
}
```

#### [M6] 缺少网络隔离验证
- **文件**：`internal/module/container/network_test.go`
- **问题描述**：测试用例只验证了网络创建和容器创建，未验证网络隔离效果（容器间是否真的不可达）。
- **影响范围/风险**：
  - 无法确认网络隔离功能是否生效
  - 可能存在配置错误但测试通过
- **修正建议**：
  添加网络隔离验证测试：创建两个独立网络的容器，验证它们无法互相 ping 通

### 🟢 低优先级

#### [L1] 示例代码未删除
- **文件**：`internal/module/container/example.go`
- **问题描述**：文件注释说明"仅供参考，实际使用时删除"，但代码已提交到仓库。
- **影响范围/风险**：
  - 代码库冗余
  - 可能被误用
- **修正建议**：
  删除 `example.go` 文件，或移动到 `examples/` 目录

#### [L2] README 文档不完整
- **文件**：`internal/module/container/README.md`
- **问题描述**：README 未包含网络管理和端口管理的说明，只有容器和镜像管理。
- **影响范围/风险**：
  - 文档与代码不一致
  - 新开发者难以理解网络隔离功能
- **修正建议**：
  添加网络管理章节：

```markdown
### 网络管理
- `CreateNetwork(ctx, name)` - 创建隔离网络
- `RemoveNetwork(ctx, networkID)` - 删除网络

### 端口管理
- `NewPortManager(start, end)` - 创建端口管理器
- `AllocatePort()` - 分配可用端口
- `ReleasePort(port)` - 释放端口
```

#### [L3] 容器配置缺少容器名称字段
- **文件**：`internal/model/container.go:6-12`
- **问题描述**：`ContainerConfig` 未包含容器名称字段，创建容器时使用空字符串作为名称。
- **影响范围/风险**：
  - 容器名称随机生成，不便于管理和调试
  - 无法通过名称快速定位容器
- **修正建议**：

```go
type ContainerConfig struct {
    Name      string // 容器名称
    Image     string
    Env       []string
    // ...
}
```

#### [L4] 端口映射使用 map 类型不稳定
- **文件**：`internal/model/container.go:9`
- **问题描述**：`Ports map[string]string` 使用 map 类型，遍历顺序不确定，可能影响端口绑定顺序。
- **影响范围/风险**：
  - 多端口映射时顺序不可控
  - 调试时输出顺序不一致
- **修正建议**：
  如需保持顺序，使用 slice：

```go
type PortMapping struct {
    ContainerPort string
    HostPort      string
}

type ContainerConfig struct {
    // ...
    Ports []PortMapping
}
```

#### [L5] 缺少日志记录
- **文件**：`internal/module/container/engine.go`, `port_manager.go`
- **问题描述**：网络创建、端口分配等关键操作未记录日志，排查问题困难。
- **影响范围/风险**：
  - 生产环境问题难以追踪
  - 无法审计网络和端口使用情况
- **修正建议**：
  添加日志记录：

```go
func (e *Engine) CreateNetwork(ctx context.Context, name string) (string, error) {
    log.Info("创建网络", zap.String("name", name))
    resp, err := e.cli.NetworkCreate(ctx, name, types.NetworkCreate{
        Driver: "bridge",
        CheckDuplicate: true,
    })
    if err != nil {
        log.Error("创建网络失败", zap.String("name", name), zap.Error(err))
        return "", err
    }
    log.Info("网络创建成功", zap.String("name", name), zap.String("id", resp.ID))
    return resp.ID, nil
}
```

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 4 |
| 🟡 中 | 6 |
| 🟢 低 | 5 |
| 合计 | 15 |

## 总体评价

本次变更实现了容器网络隔离的基础功能，包括网络管理和端口分配，代码结构清晰。但存在以下主要问题：

1. **高优先级问题（必须修复）**：
   - 端口管理器存在并发安全问题（TOCTOU 竞争）
   - 配置管理混乱，存在重复配置文件
   - 网络连接实现不正确，可能导致隔离失效
   - 网络名称冲突处理缺失

2. **中优先级问题（建议修复）**：
   - 随机数种子未初始化
   - 参数校验不完整
   - 错误信息不明确
   - 测试覆盖不足

3. **低优先级问题（可后续优化）**：
   - 文档不完整
   - 日志记录缺失
   - 示例代码未清理

**建议修复顺序**：
1. 先修复 H2（配置冲突）和 H4（网络连接），确保基础功能正确
2. 再修复 H1（并发安全）和 H3（网络冲突），确保生产可用
3. 最后修复中低优先级问题，提升代码质量

**验收建议**：
修复后需补充以下测试：
- 并发端口分配测试（100+ 并发）
- 网络隔离验证测试（容器间不可达）
- 端口冲突恢复测试
- 网络清理测试（孤儿网络检测）

