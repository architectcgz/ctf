# 2026-05-14 教学复盘脏分支人工提炼 Review

## Scope

- 脏工作区分支：`feat/teaching-review-strategy-tuning`
- 对比基线：`main` at `6eb007a2`
- 已提炼提交：
  - `7b23db2c` `重构(challenge): 收口 image build tx store concrete`
  - `6eb007a2` `docs(plan): 记录 challenge tx runner 剩余收口计划`

## Findings

### 1. `code/backend/cmd/seed-teaching-review-data/main.go` 的残留脏改动不能直接并回 `main`

严重度：`blocker`

原因：

- 当前 `main` 已经保留 richer AWD seed：
  - 多 `team`
  - 多 `round`
  - `awd_team_services`
  - `awd_traffic_events`
  - richer `teacher AWD review archive` 证据
- 脏分支里的对应文件是更早版本，会把 AWD 样本缩回到更简单的 `1 round / 2 teams / attack-only` 形态。
- 当前 `main` 还包含 `nextCoverageDimension(...)` 和 sparse catalog 回退逻辑；脏分支版本会退回“默认取相邻维度”的更弱实现。

结论：

- 该文件没有可直接搬运的剩余增量。
- 后续若还要补 seed，只能基于当前 `main` 继续做增量，不应从脏分支覆盖。

### 2. `code/backend/internal/teaching/advice/advice.go` 与 `advice_test.go` 的残留脏改动属于旧策略基线

严重度：`blocker`

原因：

- 当前 `main` 已经包含更完整的个人复盘口径：
  - `low_activity`
  - 更细的 `submission_stability`
  - “早期成功样本不直接放大成高置信度弱项”
  - 显式 `challenge_success / submission_success / submission_failure` 统计口径
  - 更细的观察文案分支
- 脏分支中的 `advice.go` / `advice_test.go` 是早于这轮收口的旧状态。直接合并会回退现有行为，而不是补强现有行为。

结论：

- 这组文件不应继续并回 `main`。
- 需要保留的策略事实，应该以当前 `main` 为准，而不是以脏分支残留为准。

### 3. 教学复盘架构文档与 plan / reuse decision 的脏改动同样是旧版本

严重度：`medium`

涉及：

- `docs/architecture/features/教学复盘建议生成架构.md`
- `.harness/reuse-decisions/teaching-review-strategy-tuning.md`
- `.harness/reuse-decisions/teaching-review-training-data-expansion.md`
- `docs/plan/impl-plan/2026-05-13-teaching-review-strategy-tuning-implementation-plan.md`
- `docs/plan/impl-plan/2026-05-13-teaching-review-training-data-expansion-implementation-plan.md`

原因：

- 当前 `main` 这些文档已经包含更晚的策略边界与 richer AWD / recommendation / archive 事实。
- 脏分支版本内容范围更窄，而且会丢失当前 `main` 已有的现状描述。

结论：

- 这组文档不应从脏分支回灌到 `main`。
- 后续如需修订文档，应直接在当前 `main` 版本上修改。

### 4. `challenge` 残留里真正有价值的部分已经提炼进 `main`

严重度：`info`

已提炼内容：

- `image build tx store` concrete 收口
- `ImageBuildRepository` 基础设施适配
- `allowlist` 删除
- 对应 slice42 reuse decision / implementation plan
- phase5 剩余 tx runner 收口计划文档

结论：

- `challenge` 方向没有需要继续从脏分支提炼的有效代码残留。

### 5. 其余未跟进文件暂不应并入

严重度：`info`

涉及：

- `.harness/reuse-decisions/echarts-mount-gate.md`
- `code/backend/data/challenge-attachments/imports/*`

原因：

- 当前没有与之配套的同轮实现或测试引用。
- 混入本轮 backend / teaching-review 收口不会增加 correctness，只会扩大噪音面。

结论：

- 这组文件不纳入本次合并。

## Distillation Result

- 已安全提炼并进入 `main` 的内容：`challenge` slice42 收口和 phase5 剩余计划文档。
- 教学复盘相关残留脏改动经人工对比后，确认主要是旧基线，不存在适合继续并回 `main` 的直接增量。
- 后续处理建议是清理脏分支残留，而不是继续尝试合并这些旧文件。

## Validation Evidence

- `git diff --no-index` 对比脏工作区与当前 `main` 同名文件
- `go test ./internal/module/challenge/... ./internal/module -run 'TestApplicationConcreteDependencyAllowlistIsCurrent' -count=1 -timeout 300s`
- `go test ./internal/module/challenge/application/commands ./internal/module/challenge/infrastructure ./internal/module/challenge/runtime -count=1 -timeout 300s`
- `bash scripts/check-consistency.sh`
- `python3 scripts/check-docs-consistency.py`
