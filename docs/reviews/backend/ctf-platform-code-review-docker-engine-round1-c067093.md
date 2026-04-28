# CTF 平台代码 Review（docker-engine 第 1 轮）：Docker 引擎封装实现

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | docker-engine |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | commit c067093，3 个文件，201 行新增 |
| 变更概述 | 实现 Docker 引擎封装，包括容器生命周期管理、镜像操作、资源限制 |
| 审查基准 | docs/architecture/container-isolation.md |
| 审查日期 | 2026-03-05 |
| 上轮问题数 | - |

## 问题清单

### 🔴 高优先级

#### [H1] Engine.Close() 未在任何地方调用，存在连接泄漏风险
- **文件**：`internal/module/container/engine.go:138-140`
- **问题描述**：Engine 提供了 Close() 方法关闭 Docker 客户端连接，但在整个生命周期中没有调用点，会导致连接泄漏
- **影响范围/风险**：长时间运行后可能耗尽文件描述符，影响系统稳定性
- **修正建议**：
  1. 如果 Engine 是单例模式，应在应用关闭时调用 Close()
  2. 如果是短生命周期实例，应使用 defer 模式：
  ```go
  engine, err := NewEngine("")
  if err != nil {
      return err
  }
  defer engine.Close()
  ```

#### [H2] 所有 Docker 操作缺少超时控制
- **文件**：`internal/module/container/engine.go` 全部方法
- **问题描述**：虽然方法接收 context.Context，但调用方可能传入 context.Background()，导致操作无限期阻塞
- **影响范围/风险**：
  - 镜像拉取可能因网络问题永久挂起
  - 容器停止可能因进程无响应永久等待
  - 影响整个平台的可用性
- **修正建议**：在方法内部强制设置超时：
  ```go
  func (e *Engine) PullImage(ctx context.Context, imageName string) error {
      ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
      defer cancel()

      reader, err := e.cli.ImagePull(ctx, imageName, image.PullOptions{})
      // ...
  }
  ```

#### [H3] 容器创建缺少安全隔离配置
- **文件**：`internal/module/container/engine.go:46-70`
- **问题描述**：CreateContainer 未设置以下关键安全参数：
  - 未禁用特权模式（Privileged）
  - 未限制 Capabilities
  - 未设置只读根文件系统（ReadonlyRootfs）
  - 未配置 Seccomp/AppArmor
- **影响范围/风险**：容器可能逃逸，攻击宿主机系统
- **修正建议**：
  ```go
  &container.HostConfig{
      PortBindings: portBindings,
      Resources:    resources,
      NetworkMode:  container.NetworkMode(cfg.Network),
      Privileged:   false,  // 强制禁用特权模式
      ReadonlyRootfs: true, // 只读根文件系统
      CapDrop:      []string{"ALL"}, // 移除所有 capabilities
      CapAdd:       []string{"NET_BIND_SERVICE"}, // 按需添加
      SecurityOpt:  []string{"no-new-privileges"}, // 禁止提权
  }
  ```

### 🟡 中优先级

#### [M1] PullImage 未处理镜像拉取进度，无法感知卡死
- **文件**：`internal/module/container/engine.go:101-109`
- **问题描述**：`io.Copy(io.Discard, reader)` 直接丢弃拉取进度，无法监控是否卡死或失败
- **影响范围/风险**：镜像拉取失败时无法及时发现，影响调试和监控
- **修正建议**：
  ```go
  func (e *Engine) PullImage(ctx context.Context, imageName string) error {
      reader, err := e.cli.ImagePull(ctx, imageName, image.PullOptions{})
      if err != nil {
          return err
      }
      defer reader.Close()

      // 解析进度并记录日志
      decoder := json.NewDecoder(reader)
      for {
          var event struct {
              Status string `json:"status"`
              Error  string `json:"error"`
          }
          if err := decoder.Decode(&event); err != nil {
              if err == io.EOF {
                  break
              }
              return err
          }
          if event.Error != "" {
              return errors.New(event.Error)
          }
          // 可选：记录进度日志
      }
      return nil
  }
  ```

#### [M2] CPU 配额计算错误
- **文件**：`internal/module/container/engine.go:57`
- **问题描述**：`resources.NanoCPUs = cfg.Resources.CPUQuota * 10000` 单位转换错误
  - NanoCPUs 单位是纳秒（1e9），表示 1 个 CPU 核心 = 1e9
  - CPUQuota 如果是微秒，应该乘以 1000，而不是 10000
- **影响范围/风险**：容器 CPU 限制不准确，可能导致资源争抢或浪费
- **修正建议**：
  ```go
  // 方案 1：明确 CPUQuota 单位为"CPU 核心数的百分比"
  resources.NanoCPUs = cfg.Resources.CPUQuota * 1e7 // 100% = 1e9

  // 方案 2：明确 CPUQuota 单位为"微秒"
  resources.NanoCPUs = cfg.Resources.CPUQuota * 1000

  // 建议在 model.ResourceLimits 中添加注释说明单位
  ```

#### [M3] RemoveContainer 强制删除可能导致数据丢失
- **文件**：`internal/module/container/engine.go:85-87`
- **问题描述**：`Force: true` 会强制删除运行中的容器，可能导致数据未保存
- **影响范围/风险**：用户容器内的临时数据可能丢失，影响用户体验
- **修正建议**：
  ```go
  func (e *Engine) RemoveContainer(ctx context.Context, containerID string, force bool) error {
      return e.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: force})
  }

  // 调用方根据场景决定是否强制删除
  ```

#### [M4] 错误处理未使用统一错误码
- **文件**：`internal/module/container/engine.go` 全部方法
- **问题描述**：所有方法直接返回 Docker SDK 的原始错误，未转换为业务错误码
- **影响范围/风险**：
  - 上层无法统一处理错误
  - API 响应格式不一致
  - 错误信息可能泄漏内部实现细节
- **修正建议**：
  ```go
  func (e *Engine) CreateContainer(ctx context.Context, cfg *model.ContainerConfig) (string, error) {
      resp, err := e.cli.ContainerCreate(...)
      if err != nil {
          // 转换为业务错误
          if client.IsErrNotFound(err) {
              return "", errcode.ErrNotFound("镜像")
          }
          return "", errcode.ErrInternal("容器创建失败: " + err.Error())
      }
      return resp.ID, nil
  }
  ```

### 🟢 低优先级

#### [L1] ContainerConfig.Ports 类型不够灵活
- **文件**：`internal/model/container.go:9`
- **问题描述**：`map[string]string` 无法表达多个宿主机端口映射到同一容器端口的场景
- **影响范围/风险**：功能受限，但当前场景可能不需要
- **修正建议**：
  ```go
  type ContainerConfig struct {
      // ...
      Ports map[string][]string // 容器端口 -> 多个宿主机端口
  }
  ```

#### [L2] ResourceLimits.DiskQuota 字段未使用
- **文件**：`internal/model/container.go:19`、`internal/module/container/engine.go:55-60`
- **问题描述**：DiskQuota 字段定义了但未在 CreateContainer 中使用
- **影响范围/风险**：配置无效，可能误导使用者
- **修正建议**：
  1. 如果暂不支持，应删除该字段或添加 TODO 注释
  2. 如果需要支持，应使用 Docker 的 StorageOpt 配置

#### [L3] ListImages 返回值可能为空切片，应明确语义
- **文件**：`internal/module/container/engine.go:112-129`
- **问题描述**：当没有镜像时返回空切片，调用方无法区分"查询成功但无结果"和"查询失败"
- **影响范围/风险**：语义不清晰，但不影响功能
- **修正建议**：保持当前实现即可，或在文档中明确说明

#### [L4] 缺少日志记录
- **文件**：`internal/module/container/engine.go` 全部方法
- **问题描述**：关键操作（创建/启动/停止/删除容器）未记录日志
- **影响范围/风险**：问题排查困难，缺少审计记录
- **修正建议**：
  ```go
  func (e *Engine) CreateContainer(ctx context.Context, cfg *model.ContainerConfig) (string, error) {
      log.Info("创建容器", "image", cfg.Image, "ports", cfg.Ports)
      resp, err := e.cli.ContainerCreate(...)
      if err != nil {
          log.Error("容器创建失败", "error", err)
          return "", err
      }
      log.Info("容器创建成功", "containerID", resp.ID)
      return resp.ID, nil
  }
  ```

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 3 |
| 🟡 中 | 4 |
| 🟢 低 | 4 |
| 合计 | 11 |

## 总体评价

代码结构清晰，基本功能完整，但存在以下关键问题：

1. **安全性不足**：缺少容器隔离配置，存在逃逸风险（H3）
2. **资源泄漏风险**：Engine 连接未关闭（H1）
3. **超时控制缺失**：所有操作可能无限期阻塞（H2）
4. **错误处理不规范**：未使用统一错误码（M4）

**建议**：修复所有高优先级和中优先级问题后再合并，低优先级问题可在后续迭代中优化。
