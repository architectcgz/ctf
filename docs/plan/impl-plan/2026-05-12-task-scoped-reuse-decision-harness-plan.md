# 2026-05-12 task-scoped reuse decision harness plan

## 目标

修正 reuse-first harness 仍以单个 `.harness/reuse-decision.md` 作为当前任务证据，导致同一工作区内多个受保护任务会互相覆盖的问题。

这次改造把当前任务证据改成 task-scoped 目录模型，让多个 reuse decision 文档可以并存，并由检查器按“覆盖当前 diff 中的受保护文件”来判定是否通过。

## 非目标

- 不重写相似页面、重复 hook、重复 API wrapper、后端复用检查的匹配算法。
- 不自动生成 task slug 或自动追加 durable reuse history。
- 不修改业务代码。

## 输入依据

- `AGENTS.md` 现有 reuse-first workflow 约束。
- `harness/checks/common.py` 与 `harness/checks/check-reuse-decision.py` 当前只读取单个 `.harness/reuse-decision.md`。
- `scripts/check-consistency.sh`、`scripts/doctor-local-harness.sh`、`scripts/check-skill-sync-reminder.sh` 当前都把单文件视为唯一入口。
- `harness/prompts/coding-agent-system-prompt.md` 与 `harness/templates/reuse-decision.md` 当前仍指导 agent 写入单文件。
- `.harness/reuse-decision.md` 当前已有活跃任务内容，说明单文件模型已经和实际并行任务冲突。

## 问题定义

当前模型把“当前任务证据”定义成全局唯一文件，会带来三个问题：

1. 两个受保护任务在同一工作区并行推进时，后写入的 reuse decision 会覆盖前一个任务。
2. staged diff 同时包含多个任务时，检查器只能要求一个文件提到所有受保护路径，导致任务边界被迫混在一起。
3. agent 即使知道 `harness/reuse/index.yaml` 和 `history.md` 已经分离，仍然必须争用同一个 scratch 文件。

## 目标状态

### 目录模型

- `.harness/reuse-decisions/`：当前工作区内按任务拆分的短期 reuse evidence 目录。
- `.harness/reuse-decisions/<task-slug>.md`：单个任务的 reuse decision 文档。
- `harness/reuse/index.yaml`：长期可检索复用模式。
- `harness/reuse/history.md`：append-only 历史摘要。

`.harness/` 仍然只保存当前任务状态和短期执行证据，不承载 durable 索引和历史。

### 校验模型

- `check-reuse-decision.py` 不再要求“一个文件覆盖所有受保护改动”。
- 新规则改为：当前 diff 中每个受保护文件，都必须被至少一个有效的 task-scoped reuse decision 文档引用。
- 单个 reuse decision 文档仍需满足原有结构要求：
  - `## Change type`
  - `## Existing code searched`
  - `## Similar implementations found`
  - `## Decision`
  - `## Reason`
  - `## Files to modify`

### 迁移策略

- 读取参考文本只汇总 `.harness/reuse-decisions/*.md`，并继续附加 `harness/reuse/index.yaml` 与 `harness/reuse/history.md`。
- 当前旧 `.harness/reuse-decision.md` 内容先迁入对应 task-scoped 文件，再移除旧文件。
- 仓库导航、prompt、模板、doctor 和 consistency check 全部切换到新目录模型，并在看到旧单文件时直接失败。

## 实施范围

1. 更新 `AGENTS.md` 的 reuse-first workflow 描述。
2. 调整 `harness/checks/common.py`：
   - 统一定义 task-scoped decision 目录。
   - 加载多个 reuse decision 文档。
   - 提供按 changed paths 计算覆盖关系的公共逻辑。
3. 调整 `harness/checks/check-reuse-decision.py`，按“路径被至少一个有效文档覆盖”校验。
4. 更新 `check-similar-pages.py`、`check-duplicate-hooks.py`、`check-api-wrapper-duplication.py`、`check-backend-reuse.py` 的提示文案。
5. 更新 `harness/prompts/coding-agent-system-prompt.md` 与 `harness/templates/reuse-decision.md`。
6. 更新 `scripts/check-consistency.sh`、`scripts/doctor-local-harness.sh`、`scripts/check-skill-sync-reminder.sh`。
7. 新增 `.harness/reuse-decisions/.gitkeep`，让目录成为稳定入口。

## 验证

- `python3 -m py_compile harness/checks/common.py harness/checks/check-reuse-decision.py harness/checks/check-similar-pages.py harness/checks/check-duplicate-hooks.py harness/checks/check-api-wrapper-duplication.py harness/checks/check-backend-reuse.py`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-reuse-first.sh --staged`

## Review Focus

- 多个 reuse decision 文档并存时，是否仍能严格覆盖当前 diff 中的所有受保护文件。
- 新入口是否已经替换掉所有用户可见的旧单文件指导。
- 旧 `.harness/reuse-decision.md` 是否已经彻底退出读写路径，并由 guardrail 明确阻止回流。
