# practice score DTO 收口实现方案

## Objective

把 `internal/dto/score.go` 中属于 `practice` 计分与排行榜的类型拆回 owning module：

- `practice/contracts`：`UserScoreInfo`、`RankingItem`

## Non-goals

- 不在这一刀里处理 `dto.PageResult`
- 不扩到 `audit.go`、`notification.go`、`cheat_detection.go` 等其他 DTO 文件
- 不改变计分、排行榜和缓存的 JSON 字段语义

## Inputs

- `code/backend/internal/dto/score.go`
- `code/backend/internal/module/practice/**/*score*.go`
- `code/backend/internal/module/practice/ports/*.go`
- `code/backend/internal/module/practice/contracts/*.go`

## Task slices

1. `practice/contracts`
   - 新增 score contract 类型
   - query / command / ports / infra 统一依赖新 owner

2. consumers
   - practice handler interface、query mapper、score state store、context contract test 改为新 owner

3. cleanup
   - 删除 `internal/dto/score.go`
   - 跑 practice mapper 生成和受影响测试

## Expected changes

- `code/backend/internal/dto/score.go`
- `code/backend/internal/module/practice/contracts/**`
- `code/backend/internal/module/practice/application/commands/**`
- `code/backend/internal/module/practice/application/queries/**`
- `code/backend/internal/module/practice/infrastructure/**`
- `code/backend/internal/module/practice/ports/**`
- `code/backend/internal/module/practice/api/http/handler.go`

## Validation

- `go generate ./internal/module/practice/...`
- `go test ./internal/module/practice/... -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`

## Review focus

- `practice/contracts` 是否成为 score contract 唯一 owner
- ports / infra 是否没有继续引用全局 `dto.UserScoreInfo` / `dto.RankingItem`
- 生成 mapper 是否只改到 score 相关 owner 切换
