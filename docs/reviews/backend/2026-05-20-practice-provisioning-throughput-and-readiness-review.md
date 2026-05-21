# Practice Provisioning Throughput And Readiness Review

- Review target: `ctf` repo，本地 `main` 工作区；review 范围为 practice 实例创建吞吐修复相关 diff
- Files reviewed:
  - `code/backend/configs/config.dev.yaml`
  - `code/backend/internal/module/practice/infrastructure/instance_readiness_probe.go`
  - `code/backend/internal/module/practice/infrastructure/instance_readiness_probe_test.go`
  - `code/backend/internal/config/config_test.go`
  - `code/backend/internal/app/composition/instance_practice_runtime_adapter.go`
  - `code/backend/internal/app/composition/instance_practice_runtime_adapter_test.go`
  - `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
  - `code/backend/internal/module/practice/application/commands/runtime_container_create_test.go`
  - `code/backend/internal/module/practice/ports/ports.go`
  - `docs/plan/impl-plan/2026-05-20-practice-provisioning-throughput-and-readiness-implementation-plan.md`
- Classification check: non-trivial backend implementation + config override + runtime validation
- Gate verdict: passed after fix

## Findings

1. 初始实现中的 blocker：`instance_readiness_probe.go` 一度使用 `strings.Contains(err.Error(), "malformed HTTP")` 触发 TCP fallback。
   - 风险：会把“任意端口可连但 HTTP 协议不匹配”的场景也误判成 ready，超出实施计划中“只兼容 malformed HTTP status code” 的边界。
   - 处理结果：已收窄为只匹配 `malformed HTTP status code`，并补了负向回归测试，确认普通 `malformed HTTP response` 仍返回错误。

## Final Assessment

- `config.dev.yaml` 的 scheduler 覆盖只作用于 `dev`，没有扩散到通用 / 生产默认值。
- readiness fallback 现在与计划承诺一致，且新增测试覆盖了正向与负向路径。
- scheduler 并发阈值从 `4` 提升到 `12` 后，创建链路的首要问题不再是调度器本身，而是子网分配路径是否把不同题型正确分流到各自地址池。
- practice 单容器题修复 `SubnetPool` 透传缺口后，100 个独立实例复测已达到 `100/100` 创建成功并全部进入 `running`，说明这轮创建失败的主因已经被定位并收口。

## Root Cause And Fix

- 旧问题集中出现在 practice 单容器题路径：运行时创建请求没有显式透传 `SubnetPool`，导致单容器题默认回退到 topology 地址池。
- 直接证据是修复前容器创建日志统一报错：`cause: no available subnet in 10.10.0.0/16 with /24`。
- 这意味着本应占用 `10.11.* /29` 的单容器实例，误用了 `10.10.0.0/16` 下仅供多容器 topology 使用的 `/24` 候选。
- 在已有 `182` 个 `10.10.*` 网络的前提下，当轮又成功分配了 `74` 个 `/24` 子网，恰好打满 `256` 个候选，因此剩余 `26` 个创建请求全部失败。
- 修复点共 3 处：
  - `code/backend/internal/module/practice/ports/ports.go`：为 `TopologyCreateRequest` 增加 `SubnetPool` 字段，并补 `SubnetPoolTopology`、`SubnetPoolSingleContainer` 常量别名。
  - `code/backend/internal/app/composition/instance_practice_runtime_adapter.go`：把 practice 侧请求中的 `SubnetPool` 透传到 runtime adapter。
  - `code/backend/internal/module/practice/application/commands/runtime_container_create.go`：两个单容器 practice 创建路径显式设置 `SubnetPool: runtimeports.SubnetPoolSingleContainer`。

## Retest Evidence

- 修复后以 20 个学生账号、每人 5 道不同单容器题的方式复测，总计 `100` 个独立实例；题目编号为 `6, 13, 14, 15, 16`。
- `100/100` 创建 API 返回成功，最终 `100/100` 实例进入 `running`。
- 创建窗口收口时，观测结果为 `create_poll elapsed=326.5s counts={"running": 100} global_active=101`。
- 保持运行约 `22.4s` 后再次采样，结果为 `hold_sample elapsed=22.4s counts={"running": 100} global_active=101`，窗口内未出现掉线或回退状态。
- 子网统计结果为 `subnet_prefix_counts={"10.11": 100}`，说明 practice 单容器实例已全部切换到 `10.11.* /29` 单容器地址池，没有再误占 `10.10.* /24` topology 池。
- 销毁阶段按 10 波渐进执行，每波 10 个实例；最终 `100/100` 实例均推进到 `stopped`，并回到 `final_global_active=1` 的基线。
- 本轮最终摘要如下：
  - `SUMMARY={"baseline_global_active": 1, "create_api_ok": 100, "create_final_counts": {"running": 100}, "final_global_active": 1, "requested": 100, "subnet_prefix_counts": {"10.11": 100}, "unique_instance_records": 100}`

## Validation Evidence

- `cd code/backend && go test ./internal/module/practice/infrastructure -count=1`
- `cd code/backend && go test ./internal/config -count=1`
- `cd code/backend && go test ./internal/module/practice/application/commands -count=1`
- `cd code/backend && go test ./internal/app/composition ./internal/module/practice/application/commands -count=1`
- `bash scripts/check-consistency.sh`
- 本地运行压测：
  - 修复前 100 规模：`100/100` 创建 API 成功，但最终只有 `74 running / 26 failed`；失败日志统一为 `no available subnet in 10.10.0.0/16 with /24`
  - 修复后 100 规模：`100/100` 创建 API 成功，最终 `100 running`；渐进销毁后 `100 stopped`，`global_active` 回到基线 `1`
  - 子网分布：`subnet_prefix_counts={"10.11": 100}`

## Residual Risk

- readiness fallback 只覆盖了当前已确认的 `malformed HTTP status code` 路径；如果后续还有其他题目把非 HTTP 服务错误标成 `http`，仍应优先修正题目元数据。
- 这轮 100 实例复测只覆盖了单机、单节点、单容器 practice 题路径；多容器 topology、长时间稳态运行和混合读写负载仍需单独补测。
- 当前结论证明了“单容器 practice 不再误用 topology 子网池”，但不等于所有题型在更高规模下都已经得到同等强度的容量验证。
