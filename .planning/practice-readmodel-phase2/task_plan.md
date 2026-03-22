# Task Plan

## Goal

Hard-migrate student progress and timeline reads from `practice` to `practice_readmodel`.

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. Analyze remaining query ownership | complete | phase 2 slice narrowed to progress/timeline only |
| 2. Write spec and plan | complete | design approved |
| 3. Implement `practice_readmodel` | complete | module split into api/application/infrastructure/contracts |
| 4. Rewire router/composition | complete | direct ownership switch, no wrapper |
| 5. Verify tests | complete | focused + full backend passed |
| 6. Align with backend architecture rules | complete | readmodel refactored to consume module contract and minimal repo contract |

## Constraints

- No compatibility layer
- No external API path changes
- Keep instance runtime command flow untouched in this slice
