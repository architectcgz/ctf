# AWD Defense Workbench Task 1-3 Review

## Scope

- Frontend Task 1-3 from `docs/plan/impl-plan/2026-05-04-awd-defense-workbench-design-implementation-plan.md`.
- Reviewed current worktree diff for SSH / VS Code UX, defense service presentation model, service selection, and component extraction.
- Backend Phase 2 defense file / directory / command routes were out of scope and not enabled.

## Independent Review

Reviewer: subagent `019df2bb-cf9f-7b03-b2e0-2a375c1746d6`

### Findings

1. `P1` `AWDDefenseConnectionPanel.vue`: connection panel did not display `access.expires_at`, while the plan requires an SSH ticket expiration hint.
2. `P2` `ContestAWDWorkspacePanel.vue`: unsupported Clipboard API path showed browser capability text instead of the required manual-copy fallback message.

### Initial Verdict

Blocked until the `P1` finding is fixed.

## Fixes Applied

- `AWDDefenseConnectionPanel.vue` now shows `票据将在 ... 过期` using the existing `formatTime` helper.
- Clipboard unsupported and rejected write paths both show `复制失败，请手动选择文本`.
- Added `AWDDefenseConnectionPanel.test.ts` to cover command/config/expiration rendering and copy emits.
- Updated raw source guard to keep the expiration hint and fallback copy text covered.

## Verification

```bash
cd code/frontend
npm run test:run -- src/components/contests/awd/__tests__/AWDDefenseConnectionPanel.test.ts src/features/contest-awd-workspace/model/awdDefensePresentation.test.ts src/features/contest-awd-workspace/model/useAwdDefenseServiceSelection.test.ts src/features/contest-awd-workspace/model/sshAccessPresentation.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts
npm run typecheck
npm run check:theme-tail
npm run test:run -- src/views/contests/__tests__/ContestDetail.test.ts
```

Result: all commands passed locally.

## Residual Risk

- Manual VS Code Remote-SSH GUI verification was not run in this environment.
- Component tests cover the connection panel directly; the full AWD panel still relies mostly on integration/source tests rather than browser-level interaction tests.

## Re-review

Reviewer: subagent `019df2c1-0cc6-7f70-9a67-f85ebafbdcde`

Verdict: approved.

Notes:

- Confirmed the SSH ticket expiration hint is rendered from `access.expires_at`.
- Confirmed unsupported Clipboard API and rejected `writeText` both show `复制失败，请手动选择文本`.
- No new findings were reported.
