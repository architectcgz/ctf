<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# 多实例动态 Flag Secret 契约重设计 Implementation Plan

**Goal:** Make dynamic Flag validation safe across multiple API instances by turning the global Flag secret into a cluster-verified runtime contract and persisting the key id used by each instance.

**Architecture:** Keep secret material outside the database. The config layer loads a keyring, the cluster-secret registry stores only key metadata and fingerprints, and the practice instance flow persists `instances.flag_key_id` alongside `nonce` so old running instances can still be validated after active key rotation.

**Tech Stack:** Go, GORM, PostgreSQL migrations, Viper config, existing practice/instance modules, existing health service.

---

## Task Metadata

- Task Slug: `2026-06-08-multi-instance-flag-secret-contract`
- Started At: `2026-06-08T07:15:23Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-multi-instance-flag-secret-contract`
- Branch: `task/2026-06-08-multi-instance-flag-secret-contract`

## Objective And Non-Goals

- Objective:
  - Production-like deployments must not silently auto-generate per-instance `container.flag_global_secret`.
  - API startup/readiness must detect active dynamic Flag secret mismatches across instances.
  - Dynamic practice instances must persist the key id used to generate their Flag.
  - Dynamic practice submissions must validate against the instance key id, while current instances without `flag_key_id` remain valid through the controlled `default` key.
- Non-Goals:
  - Do not store raw secrets in PostgreSQL.
  - Do not introduce Vault/KMS integration in this slice.
  - Do not redesign contest AWD round Flag rotation beyond reusing the same keyring API where practical.
  - Do not rewrite all health/readiness routing.

## Inputs

- Source docs:
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`
- Related architecture/contracts:
  - `code/backend/internal/config/config.go`
  - `code/backend/internal/service/health/service.go`
  - `code/backend/internal/module/practice/application/commands/submission_service.go`
  - `code/backend/internal/module/practice/application/commands/instance_start_service.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning.go`
  - `code/backend/internal/module/practice/infrastructure/repository.go`
  - `code/backend/internal/module/instance/entity/instance.go`
- Related prior work:
  - Existing dynamic Flag generation uses `FlagGlobalSecret + nonce`.
  - Existing config already persists generated local secret to `container.flag_global_secret_file`.
  - Existing health service already has `/ready` dependency aggregation.

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - Changes config semantics, migration/schema, runtime startup validation, health readiness, and dynamic Flag validation.
  - Requires TDD and migration/documentation updates.

## Files

- Create:
  - `code/backend/internal/platform/clustersecret/cluster_secret.go`
  - `code/backend/internal/platform/clustersecret/cluster_secret_test.go`
  - `code/backend/migrations/000013_add_runtime_cluster_secret_and_flag_key_id.up.sql`
  - `code/backend/migrations/000013_add_runtime_cluster_secret_and_flag_key_id.down.sql`
- Modify:
  - `code/backend/internal/config/config.go`
  - `code/backend/internal/config/config_test.go`
  - `code/backend/internal/app/composition/root.go`
  - `code/backend/internal/app/router.go`
  - `code/backend/internal/service/health/service.go`
  - `code/backend/internal/service/health/service_test.go`
  - `code/backend/internal/module/instance/entity/instance.go`
  - `code/backend/internal/module/practice/application/commands/submission_service.go`
  - `code/backend/internal/module/practice/application/commands/instance_start_service.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning.go`
  - `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
  - `code/backend/internal/module/practice/application/commands/instance_start_service_test.go`
  - `code/backend/internal/module/practice/infrastructure/repository.go`
  - `code/backend/configs/config.yaml`
  - `code/backend/configs/config.prod.yaml`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
- Review:
  - Secret mismatch failure modes.
  - Backward compatibility for existing `instances` rows with empty `flag_key_id`.
  - No raw secret material in DB/log/API response.
- Test:
  - Config unit tests.
  - Cluster secret registry unit tests.
  - Practice command dynamic Flag tests.
  - Health service readiness tests.
  - Migration contract tests if existing migration checks require new files.

## 复用与 Owner 决策

- Existing patterns searched:
  - `resolveContainerFlagGlobalSecret` already owns local secret file/env resolution.
  - `health.Service` already owns readiness dependency aggregation.
  - `practice.Service` already owns instance nonce creation and dynamic Flag validation.
  - Formal schema changes live in `code/backend/migrations`, not runtime `AutoMigrate`.
- Reuse / extend / split / create-new decision:
  - Extend config with keyring parsing and production auto-generation policy.
  - Create `internal/platform/clustersecret` because cluster-wide secret fingerprint registration is cross-module runtime infrastructure, not practice-only domain logic.
  - Extend `instances` with `flag_key_id` instead of deriving from nonce or challenge because key rotation is a runtime secret concern.
- Owner boundary:
  - Config owns loading raw key material into memory.
  - Cluster-secret registry owns DB fingerprint registration/verification only.
  - Practice instance flow owns choosing and persisting the key id for a new instance.
  - Practice submission flow owns resolving the instance key id for validation.
- Why this is the narrowest safe surface:
  - It fixes the observed multi-instance correctness risk without moving secret material into DB or replacing the runtime/provisioning architecture.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - The request is a design change across config, persistence, runtime startup, and validation behavior.
- grill-with-docs findings:
  - Existing docs currently state that file-missing auto-generation is acceptable, which conflicts with multi-instance production safety.
  - Existing dynamic Flag validation depends on `instances.nonce`; adding `flag_key_id` is the smallest persistence extension for rotation.
  - Readiness should expose secret mismatch, but startup should fail fast when registry verification fails.
- Plan adjustments after challenge:
  - Store only fingerprints and key ids in DB.
  - Keep development auto-generation, but disallow production auto-generation without explicit secret.
  - Preserve existing rows by treating empty `instances.flag_key_id` as `default`, and make startup/readiness reject API instances whose keyring lacks keys still referenced by effective instances.

## Validation

- Commands:
  - `go test ./internal/config -run 'TestLoad|TestResolveContainerFlagGlobalSecret|TestContainerFlagSecretKeyring' -count=1`
  - `go test ./internal/platform/clustersecret -count=1`
  - `go test ./internal/service/health -count=1`
  - `go test ./internal/module/practice/application/commands -run 'TestSubmitFlag|TestStartChallenge|TestBuildProvisioningFlag' -count=1`
  - `go test ./internal/app -run 'Test.*Migration|Test.*Health|Test.*Router' -count=1`
  - `bash scripts/run-workflow-stage.sh pre-commit-quick`
- Manual checks:
  - Inspect DB migration for reversible `runtime_cluster_secrets` and `instances.flag_key_id`.
  - Inspect docs for no stale claim that production can safely auto-generate per-instance secret.
- Review focus:
  - No raw secret in DB/logs.
  - Production misconfiguration fails before accepting traffic.
  - Dynamic Flag validation uses the instance key id and keeps existing rows safe.
