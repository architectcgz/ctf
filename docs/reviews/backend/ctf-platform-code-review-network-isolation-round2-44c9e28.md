# CTF 平台代码 Review（network-isolation 第 2 轮）：修复端口管理并发安全和网络连接问题

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | network-isolation |
| 轮次 | 第 2 轮（修复后复审） |
| 审查范围 | commit 44c9e28，4 个文件，49 行新增 / 33 行删除 |
| 变更概述 | 修复第 1 轮发现的高优先级问题：配置冲突、端口管理并发安全、网络连接实现、网络名称冲突 |
| 审查基准 | docs/reviews/ctf-platform-code-review-network-isolation-round1-bd16abe.md |
| 审查日期 | 2026-03-05 |
| 上轮问题数 | 15 项（4 高 / 6 中 / 5 低） |

## 第 1 轮问题修复情况

### 🔴 高优先级问题

#### [H1] 端口管理器并发安全缺陷 ✅ 已修复
- **原问题**：`isPortAvailable` 方法存在 TOCTOU 竞争条件，且在持有锁时执行网络 IO
- **修复情况**：
  - ✅ 新增 `selectPort()` 方法，在锁内选择未使用的端口
  - ✅ 将 `isPortAvailable()` 检查移到锁外执行
  - ✅ 采用"先选择后验证"策略，避免持有锁期间执行网络 IO
- **验证**：`internal/module/container/port_manager.go:28-60`
  ```go
  func (pm *PortManager) AllocatePort() (int, error) {
      for i := 0; i < 100; i++ {
          port := pm.selectPort()  // 锁内选择
          if port == 0 {
              return 0, fmt.Errorf("无法分配可用端口")
          }
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
  ```

#### [H2] 配置文件冲突 ✅ 已修复
- **原问题**：存在两个独立的 `config.go` 文件，端口范围硬编码
- **修复情况**：
  - ✅ 删除 `internal/config/config.go`，统一使用 `code/backend/internal/config/config.go`
  - ✅ 在 `ContainerConfig` 中添加 `PortRangeStart` 和 `PortRangeEnd` 字段
  - ✅ 在 `setDefaults()` 中设置默认值 30000-40000
- **验证**：`code/backend/internal/config/config.go:108-109, 208-209`
  ```go
  type ContainerConfig struct {
      // ...
      PortRangeStart   int    `mapstructure:"port_range_start"`
      PortRangeEnd     int    `mapstructure:"port_range_end"`
  }

  func setDefaults(v *viper.Viper) {
      // ...
      v.SetDefault("container.port_range_start", 30000)
      v.SetDefault("container.port_range_end", 40000)
  }
  ```

#### [H3] 网络名称冲突处理缺失 ✅ 已修复
- **原问题**：`CreateNetwork` 未处理网络名称冲突，可能导致创建失败
- **修复情况**：
  - ✅ 创建网络前先检查是否存在同名网络
  - ✅ 如存在则先删除旧网络，再创建新网络
- **验证**：`internal/module/container/engine.go:207-221`
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
                  return "", err
              }
              break
          }
      }

      resp, err := e.cli.NetworkCreate(ctx, name, types.NetworkCreate{
          Driver: "bridge",
          CheckDuplicate: true,
      })
      return resp.ID, err
  }
  ```

#### [H4] 容器创建时未连接到指定网络 ✅ 已修复
- **原问题**：使用 `NetworkMode` 设置网络，未真正连接到 Docker Network
- **修复情况**：
  - ✅ 移除 `HostConfig` 中的 `NetworkMode` 配置
  - ✅ 使用 `NetworkingConfig` 正确连接网络
  - ✅ 在 `EndpointsConfig` 中指定网络名称
- **验证**：`internal/module/container/engine.go:112-121`
  ```go
  // 构建网络配置
  networkCfg := &network.NetworkingConfig{}
  if cfg.Network != "" {
      networkCfg.EndpointsConfig = map[string]*network.EndpointSettings{
          cfg.Network: {},
      }
  }

  // 创建容器
  resp, err := e.cli.ContainerCreate(ctx, containerCfg, hostCfg, networkCfg, nil, "")
  ```

### 🟡 中优先级问题

#### [M1] 端口管理器未初始化随机数种子 ⚠️ 未修复
- **状态**：仍存在问题
- **当前代码**：`internal/module/container/port_manager.go:19-25`
  ```go
  func NewPortManager(start, end int) *PortManager {
      return &PortManager{
          rangeStart: start,
          rangeEnd:   end,
          usedPorts:  make(map[int]bool),
      }
  }
  ```
- **问题**：未调用 `rand.Seed(time.Now().UnixNano())`，导致每次程序启动时端口分配序列相同

#### [M2] 缺少端口范围合法性校验 ⚠️ 未修复
- **状态**：仍存在问题
- **问题**：`NewPortManager` 未校验 `start < end` 和端口在 1-65535 范围内

#### [M3] 端口分配失败时错误信息不明确 ⚠️ 未修复
- **状态**：仍存在问题
- **当前代码**：`internal/module/container/port_manager.go:45`
  ```go
  return 0, fmt.Errorf("无法分配可用端口")
  ```
- **问题**：未说明失败原因（端口耗尽、系统限制等）

#### [M4] 网络配置模型未使用 ⚠️ 未修复
- **状态**：仍存在问题
- **问题**：`internal/model/container.go` 中定义的 `NetworkConfig` 结构体未使用

#### [M5] 测试用例缺少清理逻辑 ⚠️ 未修复
- **状态**：仍存在问题
- **问题**：测试失败时 `defer` 不会执行，导致资源泄漏

#### [M6] 缺少网络隔离验证 ⚠️ 未修复
- **状态**：仍存在问题
- **问题**：测试未验证容器间是否真的网络隔离

### 🟢 低优先级问题

#### [L1] 示例代码未删除 ⚠️ 未修复
- **状态**：仍存在问题
- **问题**：`internal/module/container/example.go` 仍在代码库中

#### [L2] README 文档不完整 ⚠️ 未修复
- **状态**：仍存在问题
- **问题**：README 未包含网络管理和端口管理的说明

#### [L3] 容器配置缺少容器名称字段 ⚠️ 未修复
- **状态**：仍存在问题
- **问题**：`ContainerConfig` 未包含 `Name` 字段

#### [L4] 端口映射使用 map 类型不稳定 ⚠️ 未修复
- **状态**：仍存在问题
- **问题**：`Ports map[string]string` 遍历顺序不确定

#### [L5] 缺少日志记录 ⚠️ 未修复
- **状态**：仍存在问题
- **问题**：网络创建、端口分配等关键操作未记录日志

## 新发现问题

### 🟡 中优先级

#### [M7] 配置文件缺少端口范围配置项
- **文件**：`code/backend/configs/config.yaml:92-98`
- **问题描述**：虽然代码中已添加端口范围配置字段和默认值，但 `config.yaml` 中未添加对应配置项，用户无法通过配置文件自定义端口范围
- **影响范围/风险**：
  - 用户无法通过配置文件调整端口范围
  - 配置文件与代码不一致，容易混淆
  - 不同环境无法使用不同端口范围
- **修正建议**：
  在 `config.yaml` 的 `container` 部分添加端口配置
  ```yaml
  container:
    default_cpu_quota: 50000
    default_memory: 268435456
    default_pids_limit: 100
    readonly_rootfs: false
    run_as_user: ""
    port_range_start: 30000
    port_range_end: 40000
  ```

## 问题统计

| 类别 | 第 1 轮 | 已修复 | 未修复 | 新增 |
|------|---------|--------|--------|------|
| 🔴 高优先级 | 4 | 4 | 0 | 0 |
| 🟡 中优先级 | 6 | 0 | 6 | 1 |
| 🟢 低优先级 | 5 | 0 | 5 | 0 |
| **合计** | **15** | **4** | **11** | **1** |

## 总体评价

本轮修复成功解决了所有 4 个高优先级问题，代码质量显著提升：

**✅ 已修复的关键问题**：
1. 端口管理器并发安全问题已彻底解决，采用了正确的锁分离策略
2. 配置文件冲突已消除，统一使用 `code/backend/internal/config/config.go`
3. 网络名称冲突处理机制已实现，避免重复创建失败
4. 容器网络连接实现已修正，使用正确的 `NetworkingConfig` API

**⚠️ 仍需修复的问题**：
- 6 个中优先级问题（包括 1 个新发现问题）
- 5 个低优先级问题

**建议修复优先级**：
1. **M7（新增）**：在 `config.yaml` 中添加端口范围配置项（5 分钟）
2. **M1**：初始化随机数种子（2 分钟）
3. **M2**：添加端口范围合法性校验（5 分钟）
4. **M3**：改进错误信息（5 分钟）
5. 其他中低优先级问题可在后续迭代中修复

**验收建议**：
当前代码已具备生产可用的基础功能，建议：
- 补充并发端口分配测试（100+ 并发）
- 补充网络隔离验证测试（容器间不可达）
- 补充端口冲突恢复测试
- 修复 M7 后即可进入集成测试阶段
