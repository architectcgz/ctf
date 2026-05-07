# Reuse Decision

在新增或修改 `page / component / hook / service / store / api / form / table / modal / layout / schema` 之前，先更新本文件。

## Change type
- schema
- content
- service

## Existing code searched
- code/frontend/src/views
- code/frontend/src/components
- code/frontend/src/features
- code/frontend/src/widgets
- code/frontend/src/composables
- code/frontend/src/api
- code/frontend/src/stores
- code/backend/internal
- challenges/jeopardy/packs
- code/backend/data/challenge-attachments/imports
- scripts/challenges
- scripts/challenges/jeopardy_batch
- code/backend/internal/module/practice/application/commands
- code/backend/internal/module/runtime/runtime

## Similar implementations found
- challenges/jeopardy/packs/crypto-caesar-postcard/challenge.yml
- challenges/jeopardy/packs/crypto-stream-backup-ticket/challenge.yml
- challenges/jeopardy/packs/forensics-ci-preview-artifact/challenge.yml
- challenges/jeopardy/packs/misc-hidden-comment/challenge.yml
- challenges/jeopardy/packs/pwn-length-gate/challenge.yml
- challenges/jeopardy/packs/reverse-agent-cache-key/challenge.yml
- challenges/jeopardy/packs/web-header-door/challenge.yml
- scripts/challenges/jeopardy_batch/generate.py
- scripts/challenges/jeopardy_batch/registry.py
- scripts/challenges/verify_jeopardy_packs.py
- code/backend/internal/module/practice/application/commands/awd_defense_workspace_support.go
- code/backend/internal/module/practice/application/commands/runtime_container_create.go
- code/backend/internal/module/practice/application/commands/runtime_container_create_test.go
- code/backend/internal/module/runtime/runtime/adapters.go

## Decision
- create_new_with_reason
- extend_existing

## Reason
- This change expands the Jeopardy pack catalog, but it does not introduce a parallel pack format.
- The new packs reuse the existing challenge.yml plus statement/writeup/attachments structure and the shared generator plus verifier workflow under scripts/challenges.
- Creating new challenge slugs is appropriate here because the repo is intentionally adding more training content, while still following the existing pack schema and import layout.
- The AWD stale-workspace repair reuses the existing workspace companion bootstrap and runtime adapter flow.
- We are extending the current `practice` service tests and runtime ports to cover the missing-companion recovery path instead of introducing a separate restart workflow.

## Files to modify
- challenges/jeopardy/packs/crypto-affine-badge/challenge.yml
- challenges/jeopardy/packs/crypto-columnar-archive/challenge.yml
- challenges/jeopardy/packs/crypto-dsa-nonce-reuse/challenge.yml
- challenges/jeopardy/packs/crypto-hill-checkin/challenge.yml
- challenges/jeopardy/packs/crypto-lcg-telemetry/challenge.yml
- challenges/jeopardy/packs/crypto-mt-reset-link/challenge.yml
- challenges/jeopardy/packs/crypto-repeating-xor-ledger/challenge.yml
- challenges/jeopardy/packs/crypto-rsa-broadcast-notice/challenge.yml
- challenges/jeopardy/packs/crypto-rsa-common-modulus-mail/challenge.yml
- challenges/jeopardy/packs/crypto-rsa-fermat-ledger/challenge.yml
- challenges/jeopardy/packs/crypto-vigenere-courier/challenge.yml
- challenges/jeopardy/packs/forensics-authlog-bruteforce/challenge.yml
- challenges/jeopardy/packs/forensics-browser-history-lab/challenge.yml
- challenges/jeopardy/packs/forensics-docker-layer-leak/challenge.yml
- challenges/jeopardy/packs/forensics-email-forward-chain/challenge.yml
- challenges/jeopardy/packs/forensics-git-reflog-stash/challenge.yml
- challenges/jeopardy/packs/forensics-memory-env-snapshot/challenge.yml
- challenges/jeopardy/packs/forensics-office-review-trace/challenge.yml
- challenges/jeopardy/packs/forensics-pcap-basic-auth/challenge.yml
- challenges/jeopardy/packs/forensics-pcap-dns-exfil/challenge.yml
- challenges/jeopardy/packs/forensics-registry-runmru/challenge.yml
- challenges/jeopardy/packs/forensics-sqlite-wal-chat/challenge.yml
- challenges/jeopardy/packs/misc-bmp-lsb-note/challenge.yml
- challenges/jeopardy/packs/misc-cron-spool-inspection/challenge.yml
- challenges/jeopardy/packs/misc-env-precedence-merge/challenge.yml
- challenges/jeopardy/packs/misc-find-target-file/challenge.yml
- challenges/jeopardy/packs/misc-hexdump-archive-chain/challenge.yml
- challenges/jeopardy/packs/misc-makefile-override/challenge.yml
- challenges/jeopardy/packs/misc-pdf-embedded-file/challenge.yml
- challenges/jeopardy/packs/misc-png-text-chunk/challenge.yml
- challenges/jeopardy/packs/misc-special-filename/challenge.yml
- challenges/jeopardy/packs/misc-strings-signal/challenge.yml
- challenges/jeopardy/packs/misc-terminal-cast-replay/challenge.yml
- challenges/jeopardy/packs/pwn-format-string-write/challenge.yml
- challenges/jeopardy/packs/pwn-function-pointer-smash/challenge.yml
- challenges/jeopardy/packs/pwn-heap-adjacent-overflow/challenge.yml
- challenges/jeopardy/packs/pwn-integer-wrap-bypass/challenge.yml
- challenges/jeopardy/packs/pwn-off-by-one-auth/challenge.yml
- challenges/jeopardy/packs/pwn-partial-pointer-overwrite/challenge.yml
- challenges/jeopardy/packs/pwn-signed-index-leak/challenge.yml
- challenges/jeopardy/packs/pwn-struct-auth-flip/challenge.yml
- challenges/jeopardy/packs/pwn-table-index-overwrite/challenge.yml
- challenges/jeopardy/packs/pwn-uaf-callback/challenge.yml
- challenges/jeopardy/packs/reverse-batch-substring-gate/challenge.yml
- challenges/jeopardy/packs/reverse-js-array-mapper/challenge.yml
- challenges/jeopardy/packs/reverse-native-const-array/challenge.yml
- challenges/jeopardy/packs/reverse-native-crc32-gate/challenge.yml
- challenges/jeopardy/packs/reverse-native-keygen/challenge.yml
- challenges/jeopardy/packs/reverse-native-state-machine/challenge.yml
- challenges/jeopardy/packs/reverse-native-string-table/challenge.yml
- challenges/jeopardy/packs/reverse-powershell-xor-chain/challenge.yml
- challenges/jeopardy/packs/reverse-protocol-frame/challenge.yml
- challenges/jeopardy/packs/reverse-shell-parameter-maze/challenge.yml
- challenges/jeopardy/packs/reverse-vm-bytecode/challenge.yml
- challenges/jeopardy/packs/web-command-injection-panel/challenge.yml
- challenges/jeopardy/packs/web-cookie-json-tamper/challenge.yml
- challenges/jeopardy/packs/web-file-upload-double-ext/challenge.yml
- challenges/jeopardy/packs/web-idor-export-center/challenge.yml
- challenges/jeopardy/packs/web-jwt-weak-secret/challenge.yml
- challenges/jeopardy/packs/web-pickle-session-lab/challenge.yml
- challenges/jeopardy/packs/web-reset-token-predictable/challenge.yml
- challenges/jeopardy/packs/web-sqli-auth-bypass/challenge.yml
- challenges/jeopardy/packs/web-ssti-render-lab/challenge.yml
- challenges/jeopardy/packs/web-workflow-step-bypass/challenge.yml
- challenges/jeopardy/packs/web-xxe-local-reader/challenge.yml
- code/backend/internal/module/practice/application/commands/instance_start_service_test.go
- code/backend/internal/module/practice/application/commands/service_test.go
- code/backend/internal/module/practice/ports/ports.go
