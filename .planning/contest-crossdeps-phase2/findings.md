# Findings

- `contest` 已完成仓储端口收口，但 composition 仍把 `ChallengeModule`、`RuntimeModule` 整体保存到 `contestModuleDeps`。
- 当前真实跨模块依赖只有三项：挑战目录查询、flag 校验、AWD 容器写文件；直接依赖整个模块会放宽边界，削弱 phase2 typed deps 的收益。
- 这一轮只需要把 composition 内部依赖改成 typed contract，不改 `BuildContestModule` 的对外签名，也不触碰路由与业务逻辑。
