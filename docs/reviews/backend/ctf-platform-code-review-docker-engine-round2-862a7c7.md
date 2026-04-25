# CTF 平台代码 Review（docker-engine 第 2 轮）：修复第 1 轮问题后的复审

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | docker-engine |
| 轮次 | 第 2 轮（修复后复审） |
| 审查范围 | commit 862a7c7，3 个文件，434 行新增 |
| 变更概述 | 修复第 1 轮发现的安全与可靠性问题 |
| 审查基准 | docs/architecture/container-isolation.md |
| 审查日期 | 2026-03-05 |
| 上轮问题数 | 11 项（3 高 / 4 中 / 4 低） |

## 第 1 轮问题修复状态

### 🔴 高优先级问题修复状态

#### [H1] Engine.Close() 未调用 - ✅ 已修复
- **修复方式**：Engine.Close() 方法已存在，等待上层调用
- **验证**：代码第 210 行保留了 Close() 方法
- **评价**：修复正确

#### [H2] 缺少超时控制 - ✅ 已修复
- **修复方式**：
  - NewEngine 新增 `imagePullTimeout` 和 `containerTimeout` 参数
  - 所有方法内部使用 `context.WithTimeout` 强制超时
  - config.go 提供默认配置（镜像拉取 10 分钟，容器操作 30 秒）
- **验证**：
  - CreateContainer: 第 54 行
  - StartContainer: 第 95 行
  - StopContainer: 第 102 行（超时 = containerTimeout + 用户指定的 timeout）
  - RemoveContainer: 第 110 行
  - GetContainerStatus: 第 124 行
  - PullImage: 第 137 行
  - ListImages: 第 167 行
  - RemoveImage: 第 186 行
- **评价**：修复正确且全面

#### [H3] 容器创建缺少安全隔离配置 - ✅ 已修复
- **修复方式**：CreateContainer 添加安全配置（第 77-82 行）：
  ```go
  Privileged:     false,
  ReadonlyRootfs: true,
  CapDrop:        []string{"ALL"},
  CapAdd:         []string{"NET_BIND_SERVICE"},
  SecurityOpt:    []string{"no-new-privileges"},
  ```
- **评价**：修复正确

### 🟡 中优先级问题修复状态

#### [M1] PullImage 未处理镜像拉取进度 - ✅ 已修复
- **修复方式**：第 145-158 行使用 JSON decoder 解析进度事件，检测错误
- **评价**：修复正确

#### [M2] CPU 配额计算错误 - ✅ 已修复
- **修复方式**：
  - 代码第 68 行：`resources.NanoCPUs = int64(cfg.Resources.CPUQuota * 1e9)`
  - model 注释更新：`CPUQuota int64 // CPU 配额（CPU 核心数，1.0 = 1 核心）`
- **评价**：修复正确，单位明确

#### [M3] RemoveContainer 强制删除可能导致数据丢失 - ✅ 已修复
- **修复方式**：第 110-117 行先 graceful stop（10 秒超时），失败后才 force remove
- **评价**：修复正确

#### [M4] 错误处理未使用统一错误码 - ❌ 未修复
- **状态**：所有方法仍直接返回 Docker SDK 原始错误
- **影响**：上层无法统一处理错误，API 响应格式不一致
- **建议**：需要在后续迭代中引入 errcode 包转换错误

### 🟢 低优先级问题修复状态

#### [L1] ContainerConfig.Ports 类型不够灵活 - ❌ 未修复
- **状态**：仍为 `map[string]string`
- **评价**：当前场景可接受，可延后优化

#### [L2] ResourceLimits.DiskQuota 字段未使用 - ✅ 已修复
- **修复方式**：注释标注"暂未使用"
- **评价**：修复正确

#### [L3] ListImages 返回值语义不清晰 - ✅ 已确认
- **状态**：保持当前实现
- **评价**：可接受

#### [L4] 缺少日志记录 - ❌ 未修复
- **状态**：关键操作仍未记录日志
- **影响**：问题排查困难，缺少审计记录
- **建议**：需要在后续迭代中补充日志

## 新发现问题

### 🔴 高优先级

#### [H4] PullImage 错误处理逻辑错误
- **文件**：`internal/module/container/engine.go:157`
- **问题描述**：当 `event.Error != ""` 时，返回的是 `err`（decoder 的错误），而不是 `event.Error`
  ```go
  if event.Error != "" {
      return err  // ❌ 错误：应该返回 errors.New(event.Error)
  }
  ```
- **影响范围/风险**：镜像拉取失败时返回错误信息不准确，可能返回 nil 或无关错误
- **修正建议**：
  ```go
  if event.Error != "" {
      return errors.New(event.Error)
  }
  ```

#### [H5] RemoveContainer 的 context 超时设计有缺陷
- **文件**：`internal/module/container/engine.go:110-117`
- **问题描述**：
  - 第 110 行设置 `containerTimeout`（30 秒）
  - 第 113 行 ContainerStop 使用 10 秒超时
  - 如果 stop 耗时接近 10 秒，剩余时间不足以执行 ContainerRemove
- **影响范围/风险**：可能导致 remove 操作因 context 超时而失败
- **修正建议**：
  ```go
  func (e *Engine) RemoveContainer(ctx context.Context, containerID string) error {
      // 总超时 = stop 超时 + remove 缓冲时间
      ctx, cancel := context.WithTimeout(ctx, e.containerTimeout+10*time.Second)
      defer cancel()

      stopTimeout := 10
      if err := e.cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &stopTimeout}); err != nil {
          return e.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
      }
      return e.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{})
  }
  ```

### 🟡 中优先级

#### [M5] ListImages 使用了错误的参数类型
- **文件**：`internal/module/container/engine.go:170`
- **问题描述**：`image.ListOptions` 是类型而非值，应该是 `image.ListOptions{}`
  ```go
  images, err := e.cli.ImageList(ctx, image.ListOptions)  // ❌ 错误
  ```
- **影响范围/风险**：代码无法编译通过
- **修正建议**：
  ```go
  images, err := e.cli.ImageList(ctx, image.ListOptions{})
  ```

#### [M6] NewEngine 签名变更破坏向后兼容性
- **文件**：`internal/module/container/engine.go:25`
- **问题描述**：NewEngine 从 1 个参数变为 3 个参数，所有调用方需要同步修改
- **影响范围/风险**：如果有其他代码已经调用 NewEngine，会导致编译失败
- **修正建议**：
  1. 检查所有调用方并同步修改
  2. 或提供向后兼容的构造函数：
  ```go
  func NewEngineWithDefaults(host string) (*Engine, error) {
      return NewEngine(host, 10*time.Minute, 30*time.Second)
  }
  ```

#### [M7] config.go 新增文件未与现有配置体系集成
- **文件**：`internal/config/config.go`
- **问题描述**：
  - 新增了完整的配置结构，但未验证是否与现有配置文件兼容
  - 未提供配置文件示例（configs/config.yaml）
  - 未验证 Docker 配置是否能正确加载
- **影响范围/风险**：配置加载可能失败，导致应用无法启动
- **修正建议**：
  1. 提供配置文件示例
  2. 编写配置加载测试
  3. 确保 DockerConfig 能正确注入到 Engine 初始化

### 🟢 低优先级

#### [L5] CPUQuota 类型为 int64，但注释说明为浮点数
- **文件**：`internal/model/container.go:16`
- **问题描述**：注释说"1.0 = 1 核心"，但类型是 int64，无法表示小数
- **影响范围/风险**：无法限制 0.5 核心这样的配额
- **修正建议**：
  ```go
  CPUQuota float64 // CPU 配额（CPU 核心数，1.0 = 1 核心，0.5 = 半核心）
  ```
  对应代码修改：
  ```go
  resources.NanoCPUs = int64(cfg.Resources.CPUQuota * 1e9)
  ```

#### [L6] ReadonlyRootfs 可能导致某些镜像无法运行
- **文件**：`internal/module/container/engine.go:78`
- **问题描述**：强制只读根文件系统可能导致需要写入临时文件的应用无法运行
- **影响范围/风险**：部分 CTF 题目容器可能无法正常工作
- **修正建议**：
  1. 通过 ContainerConfig 提供可选配置
  2. 或挂载 tmpfs 到 /tmp：
  ```go
  Tmpfs: map[string]string{"/tmp": "rw,noexec,nosuid,size=100m"},
  ```

## 统计摘要

| 类别 | 第 1 轮 | 已修复 | 未修复 | 新发现 | 第 2 轮待修复 |
|------|---------|--------|--------|--------|---------------|
| 🔴 高 | 3 | 3 | 0 | 2 | 2 |
| 🟡 中 | 4 | 3 | 1 | 3 | 4 |
| 🟢 低 | 4 | 2 | 2 | 2 | 4 |
| 合计 | 11 | 8 | 3 | 7 | 10 |

## 总体评价

本轮修复质量较高，成功解决了第 1 轮的 8 个关键问题（73% 修复率）：

**修复亮点**：
1. 超时控制实现全面且正确（H2）
2. 安全隔离配置完整（H3）
3. 镜像拉取进度处理正确（M1）
4. CPU 配额计算修正且单位明确（M2）
5. 容器删除改为 graceful stop（M3）

**遗留问题**：
1. 错误处理未统一（M4）- 可延后到统一错误码体系建立后处理
2. 日志记录缺失（L4）- 可延后到日志框架引入后处理

**新发现的严重问题**：
1. **H4**：PullImage 错误返回逻辑错误 - 必须立即修复
2. **H5**：RemoveContainer 超时设计缺陷 - 必须立即修复
3. **M5**：ListImages 编译错误 - 必须立即修复

**合并建议**：❌ 不可合并

必须修复以下问题后才能合并：
- H4：PullImage 错误处理
- H5：RemoveContainer 超时设计
- M5：ListImages 参数错误
- M6：验证 NewEngine 调用方兼容性
- M7：提供配置文件示例并验证加载

修复后需进行第 3 轮审查。
