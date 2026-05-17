# Reuse Decision

## Change type
test contract alignment

## Existing code searched
- code/backend/internal/module/challenge/runtime/module_import_test.go
- code/backend/internal/module/challenge/contracts/challenge_import.go
- code/backend/internal/module/challenge/contracts/awd_challenge.go

## Similar implementations found
- code/backend/internal/module/challenge/api/http/challenge_import_handler_test.go

## Decision
refactor_existing

## Reason
runtime 导入集成测试仍使用全局 dto 响应类型，已与模块 contracts 收口方向不一致。直接切到 `challenge/contracts`，保持测试断言与当前 handler 真实边界一致。

## Files to modify
- code/backend/internal/module/challenge/runtime/module_import_test.go
