# Findings

- `runtime` 已完成 composition 标准化，但其他模块仍通过 `runtime.practice.* / runtime.ops.* / runtime.challenge.* / runtime.contest.*` 读取私有嵌套字段。
- 这些依赖本质上都是稳定的跨模块 contract，继续通过私有字段读取会让边界表达和 `identity.Users` 这类公开 contract 不一致。
- 这轮适合把 `runtime` 对外 contract 正式公开，再把 `challenge / ops / practice / contest` 的装配切到公开字段。
