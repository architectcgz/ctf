# Practice Provisioning Throughput And Readiness Review

- Review target: `ctf` repo，本地 `main` 工作区；review 范围为 practice 实例创建吞吐修复相关 diff
- Files reviewed:
  - `code/backend/configs/config.dev.yaml`
  - `code/backend/internal/module/practice/infrastructure/instance_readiness_probe.go`
  - `code/backend/internal/module/practice/infrastructure/instance_readiness_probe_test.go`
  - `code/backend/internal/config/config_test.go`
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
- 这轮压测表明 scheduler 并发阈值已从 `4` 提升到 `12` 并稳定生效，但创建吞吐仍主要受 Docker `containers/create` 超时和单实例启动链路耗时限制。

## Validation Evidence

- `cd code/backend && go test ./internal/module/practice/infrastructure -count=1`
- `cd code/backend && go test ./internal/config -count=1`
- `cd code/backend && go test ./internal/module/practice/application/commands -count=1`
- `bash scripts/check-consistency.sh`
- 本地运行压测：
  - 50 规模：`50/50` 创建 API 成功；240s 窗口内 `16 running / 28 failed / 6 creating`；删除 `50/50` 成功并最终 `stopped=50`
  - 100 规模：`100/100` 创建 API 成功；240s 窗口内 `9 running / 38 failed / 44 pending / 9 creating`；删除 `100/100` 成功并最终 `stopped=100`

## Residual Risk

- readiness fallback 只覆盖了当前已确认的 `malformed HTTP status code` 路径；如果后续还有其他题目把非 HTTP 服务错误标成 `http`，仍应优先修正题目元数据。
- 本轮主要瓶颈已从 scheduler 并发上限转移到 Docker `containers/create` 超时与题目启动成本；若要继续提升 50/100 规模成功率，需要进一步收口 Docker 创建尾延迟或分层限流策略。
