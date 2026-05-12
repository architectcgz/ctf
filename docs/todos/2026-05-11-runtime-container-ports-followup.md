# Runtime / Container Ports 后续待办

- Project: `/home/azhi/workspace/projects/ctf`
- Created: `2026-05-11T22:25+08:00`

## 背景

`runtime/application/*` compat wrapper 已经压成薄层，`runtime` 与 `instance` 的当前依赖图也补到了架构文档里。下一轮重点不再是“证明边界应该怎么画”，而是继续把剩余的实例业务与容器适配混合面切干净。

## 当前待办

- [x] P0：先处理 `internal/app/composition/runtime_adapter_compat.go` 中 AWD defense workbench 的容器文件 / 命令访问
  - 在 `instance` owner 侧定义明确 contract
  - 把 scope、可编辑路径、敏感路径、backup、命令超时这些判断迁到 owner 应用服务
  - composition 只保留 wiring，底层容器文件 / 命令能力通过 container runtime port 注入
- [x] P1：继续把 Docker / ACL / 文件操作往 container runtime ports 收口
  - 清掉 `internal/module/runtime/*` 里仍混住的“实例业务 + 容器适配”残留
- [x] P1：继续缩小 `runtime` 物理模块职责
  - 保持 `runtime` 只承接 container-facing capability，不再回流 repo / config / engine 级构造逻辑
- [x] P1：决定 `runtime_adapter_compat.go` 是继续保留还是删除
  - 如果 compat path 继续保留，只允许保持薄 wrapper
  - 如果仓库内生产调用已经迁空，下一刀就是删除 compat 文件本体
- [x] P2：补一轮新的 architecture guardrail
  - 防止 compat 逻辑、宽 engine 依赖和 owner 混住再次回流

## 本轮完成标准

- `runtime_adapter_compat.go` 已删除，仍然需要的 runtime HTTP facade 现在位于 `internal/app/composition/runtime_http_service_adapter.go`
- `instance` owner 对 AWD defense workbench 的 contract、实现和测试已能独立表达，但当前生产路由只保留 AWD defense SSH，不再注入浏览器 workbench facade
- `code/backend/internal/app/composition/architecture_test.go` 现在会阻止 `runtime_adapter_compat.go` 和 `AWDDefenseWorkbenchService` 注入回流
- `code/backend/internal/module/runtime/architecture_test.go` 现在会阻止 `runtime/runtime.Module` / `Deps` 回到宽 `Engine` 结构面，并阻止 runtime HTTP handler 重新声明 retired defense workbench 方法
