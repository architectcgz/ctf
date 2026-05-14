# Challenge Phase5 Tx Runner 遗留收口评审

## Review Target

- 仓库：`/home/azhi/workspace/projects/ctf`
- 分支：`main`
- diff 来源：当前工作区相对 `HEAD` 的未提交变更
- 对应实现方案：
  - `docs/plan/impl-plan/2026-05-14-challenge-tx-runner-phase5-remaining-implementation-plan.md`
  - `.harness/reuse-decisions/challenge-tx-runner-phase5-remaining.md`
- 重点复核文件：
  - `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
  - `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_package_revision_service.go`
  - `code/backend/internal/module/challenge/ports/ports.go`
  - `code/backend/internal/module/challenge/runtime/import_tx_bridge.go`
  - `code/backend/internal/module/challenge/runtime/awd_import_tx_bridge.go`
  - `code/backend/internal/module/challenge/runtime/package_export_tx_bridge.go`
  - `code/backend/internal/module/challenge/runtime/module.go`
  - `code/backend/internal/module/challenge/infrastructure/repository.go`
  - `code/backend/internal/module/architecture_allowlist_test.go`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`

## Classification Check

- 任务类型：migration / refactor
- 复杂度：non-trivial
- 结论：`agree`

## Gate Verdict

- `pass with minor issues`
- 结论：代码层面未发现 blocker；`challenge` 这块 tx-heavy concrete 依赖收口已经闭合。唯一保留项是本次归档属于实现后同上下文复核，不应当把它当成独立 reviewer gate。

## Findings

### 1. 未发现代码层面的 material finding

- `challenge_import_service.go` 现在只保留业务编排，事务 owner 与 `*gorm.DB` / `gorm.ErrRecordNotFound` / `clause.OnConflict` 已经下沉到 `ChallengeImportTxRunner` 与 runtime bridge。
- `awd_challenge_import_service.go` 与 `challenge_package_revision_service.go` 也改成窄边界 tx store，application 不再直接持有 raw tx 细节。
- `runtime/import_tx_bridge.go`、`runtime/awd_import_tx_bridge.go`、`runtime/package_export_tx_bridge.go` 把 concrete DB 读写和 image build tx store 适配收在 runtime；`runtime/module.go` 完成 wiring。
- `ports/ports.go` 新增的三个 use-case 级 tx runner / tx store 边界是收窄的，没有为了这次收口引入“全模块通用事务超接口”。
- `architecture_allowlist_test.go` 和 `07-modular-monolith-refactor.md` 与代码现状一致，没有留下“代码已迁、allowlist / facts source 未回收”的漂移。

## Material Findings

- 无。

## Senior Implementation Assessment

- 当前实现是这次需求下更低风险的落点：application 继续拥有 import / export 编排，runtime 只承接事务和 concrete adapter，符合模块边界。
- use case 级 tx runner 比“泛化事务服务”更容易维护，也更不容易把 unrelated repository 能力重新耦回 command service。
- 这刀同时收掉了 touched surface 上已经记录的结构债，没有把 `challenge` import/export 的 tx-heavy concrete 依赖继续留成 follow-up。

## Required Re-validation

```bash
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/commands -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/runtime -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/infrastructure -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module -run 'TestModuleArchitectureBoundaries|TestApplicationConcreteDependencyAllowlistIsCurrent|TestTransactionBoundaryAllowlistIsCurrent' -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/app -run 'TestChallengeImportPreviewAndCommitFlow|TestPracticeSubmissionFlowChallengeActivation' -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh
```

## Residual Risk

- 本文是实现后同上下文复核归档，不是独立 reviewer 进程产出的 review gate；如果要严格对齐 pipeline，还应补一次独立 reviewer 复核。
- 本次只复核 `challenge` 模块这块遗留 tx-heavy concrete surface；“整个 phase5 是否还有其他模块残留迁移”需要结合仓库内 plan / review / allowlist 再单独盘点。

## Touched Known-Debt Status

- touched surface：`challenge` import / AWD import / package export 的 tx-heavy concrete 依赖。
- 状态：已在本次 slice 内完成收口。
- 依据：
  - application 不再直接持有 `*gorm.DB`、`.Transaction(...)`、`gorm.ErrRecordNotFound`、`clause.OnConflict`
  - runtime bridge 已接管事务与 concrete persistence
  - allowlist 与架构事实源已经同步回收
