# 2026-05-11 reuse decision history implementation plan

## 目标

修正 reuse-first harness 中 `.harness/reuse-decision.md` 被反复覆盖后容易丢失历史复用线索的问题，把当前任务决策、长期索引和历史记录拆成三层。

## 非目标

- 不重写现有 reuse-first 检查器的相似度算法。
- 不迁移历史所有任务记录；这一轮只把已有当前决策作为第一条历史样例沉淀。
- 不修改业务代码。

## 输入依据

- `AGENTS.md` 中 reuse-first harness 约束。
- `harness/checks/common.py` 与 `scripts/check-reuse-first.sh` 的当前检查方式。
- `.harness/reuse-decision.md` 当前实际内容。
- `docs/plan/impl-plan/2026-05-10-reuse-first-harness-implementation-plan.md` 对 `.harness/reuse-decision.md` 的“本轮决策证据”定位。

## 方案

1. 保留 `.harness/reuse-decision.md` 作为当前任务的可覆盖 scratchpad。
2. 新增 `.harness/reuse-index.yaml` 作为长期复用索引，记录可直接复用的模式、入口和搜索关键词。
3. 新增 `.harness/reuse-history.md` 作为 append-only 历史日志，保留每轮完整决策摘要。
4. 更新检查公共逻辑，让历史/索引作为复用参考文本可被相似实现检查读取，但当前任务是否完整仍由 `.harness/reuse-decision.md` 校验。
5. 更新模板、prompt、AGENTS、反馈和一致性脚本，明确三层职责，避免把当前决策文件误当长期索引。

## 验证

- `python3 -m py_compile harness/checks/common.py harness/checks/check-reuse-decision.py harness/checks/check-similar-pages.py harness/checks/check-duplicate-hooks.py harness/checks/check-api-wrapper-duplication.py`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-reuse-first.sh --staged`

## Review Focus

- 当前任务决策校验是否仍然严格。
- 历史索引是否不会被当前任务覆盖。
- 检查器读取历史参考后是否降低了不必要的重复阅读成本，同时不允许完全跳过当前任务决策。
