# reuse decision history Review

## Review Target

- Repository: `ctf`
- Worktree: `.worktrees/fix/reuse-decision-history`
- Scope:
  - `.harness/reuse-decision.md`
  - `.harness/reuse-index.yaml`
  - `.harness/reuse-history.md`
  - `AGENTS.md`
  - `harness/checks/*`
  - `harness/templates/reuse-decision.md`
  - `harness/prompts/coding-agent-system-prompt.md`
  - `scripts/check-consistency.sh`
  - `feedback/2026-05-10-reuse-first-harness.md`
  - `docs/plan/impl-plan/2026-05-10-reuse-first-harness-implementation-plan.md`
  - `docs/plan/impl-plan/2026-05-11-reuse-decision-history-implementation-plan.md`
- Classification: non-trivial harness workflow fix

## Gate Verdict

Pass.

## Findings

No material findings.

## Assessment

The patch keeps the original current-task gate intact: `check-reuse-decision.py` still validates `.harness/reuse-decision.md` against the protected files in the active diff.

The new durable files solve the overwrite problem without turning the current decision file into a historical index:

- `.harness/reuse-decision.md` remains the current task scratchpad.
- `.harness/reuse-index.yaml` stores searchable reusable entries.
- `.harness/reuse-history.md` stores append-only summaries.
- Similar page/hook/API checks now read the combined reference text, so previously indexed candidates can be reused as evidence without rereading old overwritten decisions.

## Validation Reviewed

- `python3 -m py_compile harness/checks/common.py harness/checks/check-reuse-decision.py harness/checks/check-similar-pages.py harness/checks/check-duplicate-hooks.py harness/checks/check-api-wrapper-duplication.py`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-reuse-first.sh --staged`
- Manual Python assertion that durable index/history are included in reference text while current decision validation still fails for a missing changed file.

## Residual Risk

- The durable index is still manually maintained. The harness now tells agents where to put reusable decisions, but it does not automatically append entries after every task.
- Review was performed in the same session context, matching the project-level review rule that does not require a different reviewer for this workflow fix.
