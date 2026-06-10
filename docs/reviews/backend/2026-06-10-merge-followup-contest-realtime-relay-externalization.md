# Merge follow-up: contest realtime relay externalization

## Review source

- Branch merged: `task/2026-06-07-contest-realtime-relay-externalization`
- Merge commit: `2189ff7f7`
- Follow-up context: the branch worktree still had an untracked independent review note:
  - `docs/reviews/backend/2026-06-08-independent-review-contest-realtime-relay-externalization.md`

## Findings handled during merge

### Blocker: announcement list / cursor race

- Risk: both admin announcement management and contest detail loaded the announcement list before anchoring the sync cursor.
- Impact: an announcement created between those two requests could be absent from the full list while the later cursor anchor skipped past its outbox event.
- Resolution:
  - `useContestAnnouncementManagement.loadAnnouncements()` now anchors sync cursor before reading the list.
  - `useContestDetailDataLoader` now anchors sync cursor before reading the list during initial load and refresh.
  - Added frontend tests that assert cursor anchoring happens before list reads.

### Blocker: scoreboard freeze / unfreeze outbox transaction

- Risk: manual freeze / unfreeze committed contest state before writing the realtime outbox row.
- Impact: outbox insert failure could leave scoreboard state committed without durable relay intent.
- Resolution:
  - Added contest repository methods that update contest state and enqueue realtime relay in one database transaction.
  - `ScoreboardAdminService` now uses the transaction owner when realtime outbox is configured.
  - Added repository rollback coverage for outbox enqueue failure during status transition.

## Verification

```bash
cd code/backend && go test ./internal/module/contest/application/commands -run TestScoreboardAdminService -count=1
cd code/backend && go test ./internal/module/contest/application/queries -run TestParticipationServiceSyncAnnouncements -count=1
cd code/backend && go test ./internal/module/contest/infrastructure -run 'Test(RepositoryUpdateContestWithStatusTransitionAndRealtimeRelayRollsBackOnOutboxError|RealtimeOutboxRepository|ParticipationRepository)' -count=1
cd code/backend && go test ./internal/module/contest/... ./internal/module/ops/... -count=1
cd code/frontend && pnpm run typecheck
cd code/frontend && pnpm exec vitest run src/features/__tests__/featureBoundaries.test.ts src/features/contest-announcements/model/useAnnouncementSubscription.test.ts src/features/contest-announcements/model/useContestAnnouncementManagement.test.ts src/features/contest-detail/model/useContestDetailDataLoader.test.ts
```
