# Harness Router Prompt

## 用途

让 agent 在仓库任务开始时默认进入 harness intake，除非任务明显简单、局部、可逆且不需要沉淀。

## Prompt

```text
请先使用 harness router 判断本任务是 SIMPLE 还是 HARNESS。

默认选择 HARNESS；只有在任务明显简单、局部、可逆、不会改变项目约定、无需验证闭环、无需沉淀经验时，才选择 SIMPLE。

如果是 HARNESS：
1. 先读 AGENTS.md。
2. 再按需读取 concepts/、practice/、feedback/、prompts/、references/、works/ 和 scripts/check-consistency.sh。
3. 说明本次路线：任务类型、复杂度、需要使用的 skill、是否需要计划/review/验证、是否需要更新 harness 文件。
4. 完成后运行必要验证；如果改了 harness，必须运行 bash scripts/check-consistency.sh。
5. 收尾前必须执行一次经验提炼判断：
   - 主动判断本次是否产生了可复用经验。
   - 如果有，明确告诉用户并判断应沉淀到 `feedback/`、`prompts/`、`AGENTS.md`、skill 或检查脚本。
   - 如果没有，也要明确告诉用户“本次没有新增可提炼经验”。
```

## 效果评价

该 prompt 用于把“默认进行 harness，简单任务除外”的偏好显式化，并补上任务收尾时主动判断经验沉淀机会的闭环。对应全局 skill 为 `/home/azhi/.codex/skills/harness-router/SKILL.md`。
