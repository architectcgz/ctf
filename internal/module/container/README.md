# Docker 引擎封装模块

## 功能说明

本模块封装了 Docker SDK for Go，提供容器和镜像管理的核心功能。

## 核心接口

### 容器生命周期管理
- `CreateContainer(ctx, config)` - 创建容器
- `StartContainer(ctx, containerID)` - 启动容器
- `StopContainer(ctx, containerID, timeout)` - 停止容器
- `RemoveContainer(ctx, containerID)` - 删除容器
- `GetContainerStatus(ctx, containerID)` - 查询容器状态

### 镜像管理
- `PullImage(ctx, imageName)` - 拉取镜像
- `ListImages(ctx)` - 列出镜像
- `RemoveImage(ctx, imageID)` - 删除镜像

## 连接方式

支持两种连接方式：
- Unix Socket（默认）：`NewEngine("")`
- TCP：`NewEngine("tcp://host:port")`

## 资源限制

通过 `ResourceLimits` 配置：
- CPU 配额（微秒）
- 内存限制（字节）
- 进程数限制
- 磁盘配额

## 超时控制

所有操作必须传入 `context.Context`，建议使用 `context.WithTimeout` 设置超时。
