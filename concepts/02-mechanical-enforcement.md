# Mechanical Enforcement

机械化执行：文档负责解释，脚本和 hook 负责阻止漂移。

## 本项目落点

- `scripts/check-workflow-governance.sh` 检查 harness 目录、导航和计数声明；`scripts/check-agent-entrypoints.sh` 与 `scripts/check-shared-skills.sh` 负责 agent / shared skill 入口接线一致性。
- `.githooks/pre-commit` 在提交前执行一致性检查。
- 适合脚本化的规则应优先进入检查脚本，而不是只写进说明。
