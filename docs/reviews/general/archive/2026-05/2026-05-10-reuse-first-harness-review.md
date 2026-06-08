# reuse-first harness Review

## Review Target

- Repository: `ctf`
- Scope:
  - `AGENTS.md`
  - `.githooks/pre-commit`
  - `.githooks/README.md`
  - `.harness/reuse-decision.md`
  - `harness/policies/*`
  - `harness/templates/*`
  - `harness/prompts/coding-agent-system-prompt.md`
  - `harness/checks/*`
  - `scripts/check-reuse-first.sh`
  - `scripts/check-consistency.sh`
  - `scripts/doctor-local-harness.sh`
  - `feedback/2026-05-10-reuse-first-harness.md`
  - `docs/plan/impl-plan/2026-05-10-reuse-first-harness-implementation-plan.md`
- Classification: non-trivial harness / workflow guardrail change

## Gate Verdict

Pass with residual review-process and environment risk.

## Findings

No material implementation defects found in same-context review.

## Assessment

这次实现形成了完整闭环，而不是只补一条 prompt：

1. `harness/policies/reuse-first.yaml` 定义受保护创建面和决策枚举。
2. `harness/policies/project-patterns.yaml` 用仓库真实页面、hook、API wrapper 建立模式索引。
3. `.harness/reuse-decision.md` 提供每次受保护改动前必须更新的证据文件。
4. `harness/checks/*` + `scripts/check-reuse-first.sh` 把规则落到可执行检查。
5. `.githooks/pre-commit` 覆盖本地提交前检查。
6. 若后续需要远端兜底，再单独接 CI；它不是这轮本地 workflow guardrail 的必要条件。
6. `AGENTS.md` 把 `Classify -> Search -> Decide -> Implement` 明确写成前置步骤。

最关键的取舍是脚本强约束先聚焦前端高频重复面：页面、hook、API wrapper、表格/表单/目录模式。backend `service/schema` 先纳入受保护类型和检索范围，但没有在这一轮引入更重的语义分析器；这个边界和本轮目标一致。

## Validation

- `bash scripts/check-consistency.sh`
- `bash scripts/check-reuse-first.sh --staged`
- `python3 -m py_compile harness/checks/common.py harness/checks/check-reuse-decision.py harness/checks/check-similar-pages.py harness/checks/check-duplicate-hooks.py harness/checks/check-api-wrapper-duplication.py`
- `python3` 调用 `validate_reuse_decision(...)` 验证占位版 `.harness/reuse-decision.md` 会因为缺少搜索根、决策值和受影响文件引用而失败
- `python3` 调用 `similarity_score(...)` 验证 `NotificationList.vue` 与 `StudentManagementPage.vue` 会命中共享关键词，得分 `17`

## Residual Risk

- 这次 review 在同一 agent 上下文完成，没有满足独立 subagent review gate。剩余风险在 review 独立性，不在已观察到的实现缺陷。
- `bash scripts/doctor-local-harness.sh` 在当前 worktree 中因为 `code/frontend/node_modules` 缺失而失败。这是环境准备项，不是 reuse-first harness 本身的逻辑错误。
