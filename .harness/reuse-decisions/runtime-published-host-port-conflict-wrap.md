# Reuse Decision

## Change type
service / port / runtime / test

## Existing code searched
- `code/backend/internal/module/runtime/infrastructure/engine.go`
- `code/backend/internal/module/runtime/infrastructure/engine_test.go`
- `code/backend/internal/module/runtime/ports/container_runtime.go`
- `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
- `code/backend/internal/module/practice/application/commands/runtime_container_create_test.go`
- `code/backend/internal/module/runtime/infrastructure/cleaner.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`

## Similar implementations found
- `runtime/infrastructure/engine.go`
  - 已经是 Docker client error 归一化 owner，published host port 冲突识别继续放在这里最合理。
- `practice/application/commands/runtime_container_create.go`
  - 上层已经通过 `errors.Is(err, runtimeports.ErrPublishedHostPortConflict)` 做重绑分支，说明这次应该继续扩展现有 typed sentinel，而不是改上层分支条件。
- `runtime/infrastructure/*_test.go`
  - runtime 基础设施层已有针对 helper/adapter 的 package-local 单测，适合补一组 wrapper 回归测试，而不是只在更高层做间接验证。

## Decision
refactor_existing

## Reason
这次不是新增新的 Docker 错误模型，而是把现有的 published host port conflict 识别从“只暴露 typed sentinel + 拼接文本”收口成“typed sentinel 仍可 `errors.Is`，同时完整保留底层 Docker error unwrap 链”。继续沿用 `runtime/ports` 里的现有 error contract 和 `runtime/infrastructure/engine.go` 的归一化入口，改动最小，也不会把 Docker 细节扩散到 practice/application。

## Files to modify
- `code/backend/internal/module/runtime/ports/errors.go`
- `code/backend/internal/module/runtime/infrastructure/engine.go`
- `code/backend/internal/module/runtime/infrastructure/engine_error_test.go`

## After implementation
- 如果后续还要识别其他 Docker daemon 特定冲突，优先继续在 `runtime/ports` 暴露模块内 typed wrapper，并在 `engine.go` 做单点归一化，不把字符串判定或 Docker errdefs 判断扩散到上层模块。
- 如果 Docker SDK 将来暴露更稳定的结构化业务错误字段，再把当前 message-based 识别替换掉，但保持“上层只依赖 runtime typed error、底层 cause 不丢”的契约不变。
