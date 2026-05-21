# assessment 维度总分缓存实现方案

## Objective

把能力画像计算里“已发布题目各维度总分”的重复聚合结果缓存下来，避免给每个学生重算一次 `SUM(challenges.points)`，并且在题库变更后借助事件立即失效，避免只靠 TTL 产生陈旧窗口。

## Non-goals

- 不改动 `SkillProfile` 的评分公式
- 不重构学生做题得分统计逻辑
- 不在本轮补齐“题库变更后全量重建所有画像”的流程
- 不把 challenge 写路径和 assessment 缓存读取逻辑耦合到同一个模块

## Inputs

- `code/backend/internal/module/assessment/application/commands/profile_service.go`
- `code/backend/internal/module/assessment/infrastructure/repository.go`
- `code/backend/internal/module/assessment/infrastructure/state_store.go`
- `code/backend/internal/module/assessment/runtime/module.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/config/config.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- `code/backend/internal/module/challenge/contracts/events.go`

## Ownership Evaluation

- owner 明确：已发布题目维度总分的读取 owner 仍在 `assessment/infrastructure/repository.go`
- 复用点明确：复用 assessment 已有 Redis state store 模式，而不是新增统计表或并行存储
- 落点明确：repository 负责读缓存与回源数据库，challenge command service 负责发题库变更事件，assessment runtime 负责注册缓存失效 consumer
- 已知限制明确：TTL 只作为 Redis 失效兜底，不再承担题库变更后的主失效路径

## Task slices

1. 新增 assessment 维度总分缓存接口、Redis store 和 cache key
2. 改造 assessment repository，把维度总分读路径拆成“缓存总分 + 查询学生得分”
3. 在 runtime 装配缓存配置，并补 assessment 配置默认值
4. 在 challenge 写路径补“已发布题库变更”事件，并在 assessment 侧订阅后删除维度总分缓存
5. 补 repository / service / challenge command 相关测试
6. 跑最小充分验证

## Validation

- `go test ./internal/module/challenge/application/commands -count=1`
- `go test ./internal/module/assessment/infrastructure -count=1`
- `go test ./internal/module/assessment/application/commands -count=1`
- `go test ./internal/module/assessment/application/queries -count=1`
- `go test ./internal/module/assessment/runtime -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- 维度总分是否真正只回源一次并进入缓存
- repository 是否仍保持 assessment 读路径 owner，不把缓存判断扩散到 service
- 缓存 miss / hit / decode 失败时是否都能安全回退到数据库
- challenge 写路径是否覆盖到 `Update / Delete / Publish / CommitChallengeImport`
- 题库变更后是否会马上删掉 assessment 维度总分缓存，而不是继续等待 TTL

## Rollback

如有回归，可移除 challenge 事件订阅和 repository 上的维度总分缓存 option，恢复到原先的 SQL 聚合路径；不会涉及 schema 变更。
