# Findings

- `challenge` 已完成模块内部物理分层，但 `ports/ports.go` 仍保留宽 `ChallengeRepository`，命令、查询、flag、writeup、topology 共用同一仓储接口。
- [`challenge_module.go`](/home/azhi/workspace/projects/ctf/code/backend/internal/app/composition/challenge_module.go) 仍在单个 builder 中直接装配 `challengeinfra.Repository`、`ImageRepository`、`TemplateRepository`，缺少与 `contest phase2` 对齐的 typed deps 收口。
- `ImageRepository`、`EnvironmentTemplateRepository`、`TagRepository` 已天然按职责拆开；本次重点是把 challenge 主仓储按用例边界拆成窄接口。
- `practice` 与 `runtime` 仍有更重的跨模块适配边界，本次不扩散到这两块。
