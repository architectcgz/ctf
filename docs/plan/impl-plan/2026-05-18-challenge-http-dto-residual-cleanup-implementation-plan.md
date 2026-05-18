# challenge HTTP DTO 残留清理实现方案

## Objective

清掉 `internal/dto` 中已经被 `challenge/api/http` 接管的残留文件：

- `awd_challenge_import.go`
- `challenge.go`
- `challenge_import.go`
- `topology.go`
- `awd_challenge.go`

同时把 app 集成测试改为直接解码模块 owner 的 HTTP DTO。

## Non-goals

- 不在这一刀处理 `common.go`
- 不改 challenge / topology / awd challenge / package import 接口字段和行为
- 不调整 challenge 模块现有 contracts、mapper、handler 结构

## Inputs

- `code/backend/internal/dto/{awd_challenge_import.go,challenge.go,challenge_import.go,topology.go,awd_challenge.go}`
- `code/backend/internal/module/challenge/api/http/{challenge_request_types.go,challenge_response_types.go}`
- `code/backend/internal/app/{full_router_integration_test.go,full_router_state_matrix_integration_test.go}`
- `.harness/reuse-decisions/challenge-http-dto-residual-cleanup.md`

## Task slices

1. app 集成测试收口
   - `full_router_integration_test.go` 改为使用 `challenge/api/http` challenge 相关 DTO
   - `full_router_state_matrix_integration_test.go` 改为使用 `challenge/api/http` challenge 相关 DTO

2. cleanup
   - 删除 `internal/dto/awd_challenge_import.go`
   - 删除 `internal/dto/challenge.go`
   - 删除 `internal/dto/challenge_import.go`
   - 删除 `internal/dto/topology.go`
   - 删除 `internal/dto/awd_challenge.go`

3. verification
   - 跑受影响 app 用例、challenge 模块测试、全局 mapper guardrail 和 consistency check

## Expected changes

- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/dto/awd_challenge_import.go`
- `code/backend/internal/dto/challenge.go`
- `code/backend/internal/dto/challenge_import.go`
- `code/backend/internal/dto/topology.go`
- `code/backend/internal/dto/awd_challenge.go`

## Validation

- `go test ./internal/app -run 'TestFullRouter_TeacherCanBrowseArchivedAndDraftChallengesButOnlyManageOwnChallenges|TestFullRouter_ChallengeSelfCheckRunsPrecheckAndRuntime|TestFullRouter_ChallengeLifecycleAndContentStateMatrix|TestFullRouter_ChallengeWriteupsUseCommunitySemantics|TestFullRouter_AWDChallengeAdminStateMatrix' -count=1`
- `go test ./internal/module/challenge/... -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- app 测试是否彻底脱离 challenge 相关全局 DTO
- 删除四个文件后是否仍存在直接或间接引用遗漏
- 这一刀是否只做 owner 收口，没有改变 challenge 对外 HTTP 契约
