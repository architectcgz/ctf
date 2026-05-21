# AWD 首次成功后立即轮换 Flag 实施计划

## 目标

- 修复 AWD 比赛中“同一轮首个成功提交后，其他队还能复用旧 flag 再次提交成功”的问题。
- 将本轮同一受害队伍、同一服务的 flag 在首个成功提交后立即轮换，并同步写回运行实例。
- 保持现有计分语义不变：旧 flag 失效；后续如果再次拿到新 flag，仍按现有成功提交逻辑处理。

## 非目标

- 不改动现有 `CountSuccessfulAttacks` 的记分去重范围。
- 不引入跨轮或跨服务的全局首杀判定。
- 不改动 round updater 的整轮 flag 同步策略。

## 输入事实源

- `code/backend/internal/module/contest/application/commands/awd_attack_submit_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_attack_submit_support.go`
- `code/backend/internal/module/contest/application/commands/awd_attack_log_commands.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_flag_sync.go`
- `code/backend/internal/module/contest/infrastructure/awd_round_state_store.go`
- `code/backend/internal/module/contest/infrastructure/awd_docker_flag_injector.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/application/commands/awd_service_test.go`

## 方案结论

- 在 `SubmitAttack` 链路内，当提交命中当前可接受的当前轮 flag 时，先对 Redis 当前轮 flag 做原子 claim，再进行轮换与注入。
- 新 flag 使用随机 nonce 配合现有 `flagcrypto.GenerateDynamicFlag` 生成，避免继续使用确定性 `BuildAWDRoundFlag`。
- 原子 claim 使用 `AWDRoundStateStore` 的 compare-and-swap 接口，保证同一旧 flag 在竞争窗口里只能被一个提交换走。
- 轮换需要同时完成两步：原子更新 `AWDRoundStateStore` 当前轮 flag 字段、调用 `AWDFlagInjector` 把新 flag 写回容器 `/flag`。
- 如果注入或后续落库失败，则整次提交返回错误，并对 Redis 与容器做 best-effort 回滚，避免“数据库已记成功但 flag 仍未轮换”的半成功状态。

## 改动边界

- `application/commands`
  - `AWDService` 新增轮换依赖与提交后轮换逻辑。
  - 抽出“首个成功后轮换”的最小辅助函数，保持 `SubmitAttack` 主流程可读。
- `ports`
  - `AWDRoundStateStore` 增加单条 round flag 覆写接口。
- `infrastructure`
  - Redis state store 实现单条 flag 覆写。
- `runtime`
  - `buildAWDHandler` 为 `AWDService` 注入现有 Docker flag injector。
- `tests`
  - 增加红灯测试覆盖即时轮换、旧 flag 失效、容器写回与失败回滚。

## 任务切片

### 切片 1：计划与复用决策

- 补 implementation plan。
- 补 `.harness/reuse-decisions/awd-flag-immediate-rotation.md`。

验证：

- `bash scripts/check-consistency.sh`

### 切片 2：测试先行

- 在 `awd_service_test.go` 增加失败测试：
  - 首个成功提交后 Redis 当前轮 flag 被替换。
  - 使用旧 flag 再次提交直接失败。
  - 注入器收到新的单条 flag 写入。

验证：

- `go test ./internal/module/contest/application/commands -run 'TestAWDServiceSubmitAttack.*Rotate' -count=1`

### 切片 3：实现即时轮换

- 扩展 `AWDRoundStateStore` 接口与 Redis 实现。
- 给 `AWDService` 注入 `AWDFlagInjector`。
- 在成功提交链路里做“先原子 claim 并轮换、后落日志”。

验证：

- `go test ./internal/module/contest/application/commands -run 'TestAWDServiceSubmitAttack.*Rotate' -count=1`

### 切片 4：回归验证

- 跑受影响 commands 包测试。
- 跑 contest 模块最小充分测试。
- 跑仓库一致性检查。

验证：

- `go test ./internal/module/contest/application/commands -count=1`
- `go test ./internal/module/contest/... -count=1`
- `bash scripts/check-consistency.sh`

## 数据与兼容性影响

- Redis 当前轮 flag hash 将允许单字段覆写，键结构不变。
- 数据库结构不变。
- 现有 API 入参与响应结构不变，但旧 flag 的提交结果会从“成功但可能 0 分”变为“失败”。

## 风险与回退

- 风险：轮换逻辑引入 Redis 与容器写入的额外失败点。
- 应对：轮换失败时直接返回错误，不写成功日志。
- 回退：撤销本次 `SubmitAttack` 轮换逻辑与 `SetAWDRoundFlag` 接口增量即可恢复现状。

## Review 关注点

- 成功判定与记分判定是否仍保持单一 owner。
- Redis 与容器 flag 是否始终同步更新。
- 失败路径是否留下成功日志或错误事件。
- 旧测试中基于“第二次旧 flag 仍成功”的断言是否已同步改正。
