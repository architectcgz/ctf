# runtime ACL authority hardening review

## Review target

- Repository: `ctf`
- Branch: `main`
- Diff source: local working tree on `2026-06-02`
- Dominant scope: runtime ACL authority hardening + security headers/CSP follow-up
- Files reviewed:
  - `code/backend/internal/module/runtime/contracts/runtime_details.go`
  - `code/backend/internal/module/runtime/domain/resources.go`
  - `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
  - `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
  - `code/backend/internal/module/runtime/infrastructure/acl.go`
  - `code/backend/internal/module/runtime/infrastructure/engine_provisioning.go`
  - `code/backend/internal/module/runtime/ports/container_runtime.go`
  - `code/backend/internal/module/runtime/service_test.go`
  - `code/backend/internal/module/runtime/service_acl_test.go`
  - `code/backend/internal/module/runtime/infrastructure/acl_test.go`
  - `code/backend/internal/app/router.go`
  - `code/backend/internal/middleware/security_headers.go`
  - `code/frontend/nginx/default.conf`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/todos/2026-06-02-security-review-findings.md`

## Classification check

- Route chosen: `HARNESS`
- Review classification: agree this is non-trivial security/runtime work

## Gate verdict

- `blocked`

## Findings

### 1. Blocker: legacy cleanup fallback can no longer delete old ACL rules because comment canonicalization changed the match key

- Location:
  - `code/backend/internal/module/runtime/infrastructure/acl.go:53-65`
  - `code/backend/internal/module/runtime/infrastructure/acl.go:332-370`
  - `code/backend/internal/module/runtime/domain/topology_acl.go:233-249`
  - `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go:106-108`
- Risk:
  - The new rollout explicitly keeps `runtime_details.acl_rules` as the cleanup fallback for old instances.
  - However `removeACLRules()` now runs every stored rule through `validateAndCanonicalizeACLRule()`, which unconditionally rewrites `Comment` using `systemACLComment()`.
  - Old rules were inserted with `buildRuntimeACLComment()`, which hashes `source_container_id/target_container_id/source_ip/target_ip/action/protocol/ports`.
  - `iptables -D` needs the same match tuple that was used when the rule was inserted. Rebuilding the comment into a different format means the fallback delete path will miss the old rule and leave stale ACL entries behind.
- Fix direction:
  - Preserve the stored legacy comment during the old-rule fallback path, or introduce a dedicated `validateLegacyACLRuleForDelete()` that validates IP/protocol/action/ports without rewriting `Comment`.
  - Add an infrastructure-level regression test that proves a pre-migration rule shape can still be deleted.

### 2. Blocker: `ApplyACL` failure leaves orphan instance chains on the host because provisioning rollback does not clean up partially created ACL resources

- Location:
  - `code/backend/internal/module/runtime/infrastructure/acl.go:124-138`
  - `code/backend/internal/module/runtime/application/commands/provisioning_service.go:405-413`
- Risk:
  - `applyInstanceACL()` creates or flushes the instance chain before writing rules, and only attaches the `DOCKER-USER` jump at the end.
  - If rule append or jump insertion fails after the chain is created, the function returns directly without calling `removeInstanceACL()`.
  - `CreateTopology()` then falls back to `cleanupTopologyResources()`, but at that point `details.ACL` has not been persisted yet and the cleanup path only removes containers and networks. The partially created chain is left on the host.
  - Repeated failures can accumulate orphan chains and stale partial rules, which is exactly the kind of host-level drift this refactor was supposed to reduce.
- Fix direction:
  - Make `applyInstanceACL()` internally transactional: on any error after chain creation, call `removeInstanceACL()` before returning.
  - Alternatively, set the handle earlier and explicitly invoke `RemoveACL(handle)` from the provisioning rollback branch.
  - Add a test that simulates failure during rule append or jump creation and asserts the chain cleanup path runs.

## Material findings

- Keep legacy `acl_rules` cleanup compatible with pre-migration comments.
- Ensure `ApplyACL` rolls back host ACL resources when provisioning fails mid-flight.

## Senior implementation assessment

- The overall direction is right: moving cleanup authority from `runtime_details.acl_rules` to an instance-level ACL handle is the correct architectural move.
- The current patch still leaves two lifecycle gaps at the infrastructure boundary:
  - migration compatibility for existing persisted rules
  - transactional cleanup when host ACL resource creation only partially succeeds
- A senior maintainer would keep the new handle model, but would make the infrastructure adapter fully self-rolling-back and would split “new canonical rule validation” from “legacy persisted rule delete validation” instead of routing both through the same comment rewrite.

## Required re-validation

- `cd code/backend && go test ./internal/module/runtime/infrastructure -count=1`
  - add a case that deletes a legacy hashed-comment rule successfully
  - add a case that `ApplyACL` rolls back the chain on append/jump failure
- `cd code/backend && go test ./internal/module/runtime/... -count=1`
  - confirm provisioning + cleanup behavior still passes with the new rollback path

## Residual risk

- I did not run browser-level validation for the frontend CSP and API header changes.
- The runtime package test suite passes, but it currently does not cover the two failure modes above, so passing tests are not evidence those paths are safe.

## Touched known-debt status

- This diff directly touches the security todo item about `iptables` parameter hardening and runtime ACL authority.
- The authority shift is a net improvement, but the touched surface is not fully closed because the rollout-compatible cleanup path and the partial-failure rollback path are still unsafe.

## Validation evidence

- `bash scripts/check-task-intake.sh`
- `cd code/backend && go test ./internal/module/runtime/... -count=1`
