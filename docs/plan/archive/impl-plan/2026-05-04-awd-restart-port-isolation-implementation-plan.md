# 2026-05-04 AWD 重启端口隔离实现计划

## Plan Summary

- Objective
  - 修复 AWD 比赛中用户重启防守服务时，历史 `host_port` 残留导致重启失败的问题。
  - 保证 AWD 服务的稳定访问目标由赛内网络别名和服务端口承担，重启容器不分配、不复用宿主机端口。
  - 重启历史 AWD 实例时释放并清空旧 `host_port`，避免 `port_allocations` 中其他实例的合法占用阻断重启。
- Non-goals
  - 不改变普通 Jeopardy/个人靶机实例的宿主机端口分配和重启保留语义。
  - 不改变 AWD checker、轮次调度、服务状态写回逻辑。
  - 不做破坏性批量数据修复脚本；本次先让代码路径具备自恢复能力。
- Source architecture or design docs
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
  - 当前代码路径：
    - `code/backend/internal/module/practice/application/commands/instance_start_service.go`
    - `code/backend/internal/module/practice/infrastructure/repository.go`
    - `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
- Dependency order
  - 先调整端口分配判定，再调整重启 reset 契约，最后补测试和验证。
- Expected specialist skills
  - `development-pipeline`
  - `architect-agent`
  - `backend-engineer`
  - `go-backend`
  - `test-engineer`
  - `code-reviewer`

## Task 1

- Status
  - [x] AWD 实例创建路径不再申请宿主机端口。
  - [x] `StartContestAWDService` 测试覆盖不调用端口 reservation，实例 `HostPort == 0`。
- Goal
  - 让 AWD 实例创建路径不再申请宿主机端口。
- Touched modules or boundaries
  - `instance_start_service.go`
  - 相关 start/restart service tests
- Dependencies
  - 依赖 `InstanceScope.ContestMode` 和 `ServiceID` 已能识别 AWD service scope。
- Validation
  - 新增或更新 `StartContestAWDService` 测试，断言 AWD 创建不调用端口 reservation，实例 `HostPort == 0`。
- Review focus
  - 普通靶机是否仍然保留宿主机端口。
  - AWD service scope 判断是否过宽。
- Risk notes
  - 如果某些非 AWD 竞赛路径也带 `ServiceID`，不能被误判；优先使用 `ContestMode == awd`。

## Task 2

- Status
  - [x] 重启 reset 契约支持按 scope 选择是否保留 `host_port`。
  - [x] AWD 重启传入“不保留 host port”。
  - [x] repository reset 在不保留端口时清空 `instances.host_port`。
  - [x] repository reset 不删除其他实例持有的 `port_allocations`。
  - [x] 普通重启保留 `host_port` 并校验 allocation 所属实例。
- Goal
  - 重启 AWD 历史实例时释放并清空旧 `host_port`，避免复用脏端口。
- Touched modules or boundaries
  - `ports.go`
  - `repository.go`
  - `repository_stub_test.go`
  - `instance_context_contract_test.go`
  - restart service tests
- Dependencies
  - Task 1 明确 AWD 不再需要 `host_port`。
- Validation
  - 单元测试覆盖：
    - AWD 重启传入“不保留 host port”。
    - repository reset 在不保留端口时删除仅属于该实例的 `port_allocations` 绑定并将 `instances.host_port` 置 0。
    - 普通重启仍保留 `host_port` 并校验 allocation 所属实例。
- Review focus
  - 事务内释放端口和清空实例字段是否一致。
  - 不应删除属于其他实例的端口分配记录。
  - failed 历史实例被重启时能从脏数据恢复。
- Risk notes
  - 当前生产数据可能已有 failed AWD 实例 `host_port` 指向其他实例；新路径应清空旧值时不要求它拥有 allocation。

## Integration Checks

- [x] `go test ./internal/module/practice/application/commands`
- [x] `go test ./internal/module/practice/infrastructure`
- [x] `go test ./internal/module/practice/ports`
- [x] `go test ./internal/module/practice/...`

## Rollback / Recovery Notes

- 代码回滚可独立 revert。
- 不引入 schema migration。
- 对已失败的 AWD instance，修复后再次点击重启应通过 reset 路径自恢复 `host_port=0`。
- 若需要即时修复当前演示数据，可在代码修复后重新触发重启；不默认执行手工 SQL。

## Residual Risks

- 如果外部防守入口依赖宿主机端口而不是 AWD alias/代理，该路径需要另行设计；当前代码和日志显示 AWD 使用 `http://awd-c...` 稳定别名。
- 历史已经 running 且带 host_port 的 AWD 实例不会被本次主动批量清理；它们下次重启时会进入自恢复路径。
