# contest early end guard 实施计划

## Objective

给 `UpdateContest` 补一层保护，避免运行中的比赛在未到结束时间时被普通手动更新直接切到 `ended`，从而触发运行实例清理并造成赛中实例消失。

## Non-goals

- 不重做竞赛状态机
- 不新增一套独立的“提前结束比赛”命令
- 不在这刀补充完整的操作审计落库
- 不回填已经受影响的历史比赛数据

## Inputs

- `code/backend/internal/module/contest/application/commands/contest_update_commands.go`
- `code/backend/internal/module/contest/application/commands/contest_update_support.go`
- `code/backend/internal/module/contest/application/commands/contest_service_test.go`
- `code/backend/internal/module/contest/contracts/errors.go`
- `code/backend/internal/module/contest/infrastructure/ended_contest_runtime_cleaner.go`

## Ownership evaluation

- `ContestService.UpdateContest` 是管理员手动改竞赛状态的唯一 owner，应在这里拦住“运行中提前结束”的危险输入。
- `ended contest runtime cleaner` 只负责在比赛已经进入 `ended` 后清理运行态，不负责判断这次结束是否合理。
- `StatusUpdater` 只负责基于时间窗口自动推进状态，不负责手动更新的风控。

## Task slices

1. 在 `contest_service_test.go` 增加回归用例，先证明当前代码会放行“运行中提前结束”。
2. 在 `UpdateContest` 校验路径中补保护：当目标状态是 `ended` 且当前时间尚未到 `end_time` 时，默认拒绝；只有显式 `force_override=true` 且 `override_reason` 非空时才允许。
3. 补充对应错误契约和放行用例，确认原有 AWD 开赛 override 逻辑不受影响。

## Data and compatibility impact

- 管理员在比赛尚未自然结束时，不能再无提示地把状态改成 `ended`。
- 如确需提前结束，必须显式携带 `force_override` 与 `override_reason`。
- 到达自然结束时间后的手动 `ended` 更新行为保持不变。

## Validation

- `go test ./internal/module/contest/application/commands -run TestContestServiceUpdateContest -count=1`
- `go test ./internal/module/contest/application/commands -count=1`

## Review focus

- 提前结束的保护是否只收口在手动更新入口，而没有影响自动状态推进
- `force_override` 的放行条件是否足够明确，空白理由是否仍被拒绝
- 既有 `running -> frozen`、`registration -> running` 测试是否保持通过

## Rollback

如果确认当前产品确实需要无保护地提前结束比赛，可回退这次校验与新错误契约，仅保留调查结论，再单独设计显式的“提前结束比赛”流程。
