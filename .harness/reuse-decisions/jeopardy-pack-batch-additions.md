# Reuse Decision

## Change type

challenge-pack / manifest / writeup / attachment

## Existing code searched

- `challenges/README.md`
- `challenges/teacher-authoring-guide.md`
- `challenges/jeopardy/templates/README.md`
- `scripts/challenges/jeopardy_batch/verify.py`
- `code/backend/internal/module/challenge/domain/package_parser.go`
- `challenges/jeopardy/packs/web-source-audit-double-wrap-01/challenge.yml`

## Similar implementations found

- `challenges/teacher-authoring-guide.md`
  - 已有 Jeopardy 题包目录结构、`challenge.yml` 字段和附件引用约束
- `challenges/jeopardy/templates/README.md`
  - 已有非容器题、容器题的包结构模板
- `scripts/challenges/jeopardy_batch/verify.py`
  - 已有题包验证 owner，可直接复用 `writeup/solve.py` 的恢复口径
- `package_parser.go`
  - 已有题包导入契约 owner，manifest 不应扩出平行字段体系

## Decision

extend_existing

## Reason

这次不是发明新的 Jeopardy 题包格式，而是在现有 contract 下新增一批具体题包，并把 `web-source-audit-double-wrap-01` 收口回“题包自包含、可被现有验证脚本直接求解”的模式。`challenge.yml` 继续服从现有导入契约；`writeup/solve.py` 继续作为验证脚本入口；容器题仍走现有 `docker/` 布局。

## Files to modify

- `challenges/jeopardy/packs/crypto-rsa-prime-mesh/challenge.yml`
- `challenges/jeopardy/packs/crypto-rsa-prime-mesh/statement.md`
- `challenges/jeopardy/packs/crypto-rsa-prime-mesh/attachments/mesh.json`
- `challenges/jeopardy/packs/crypto-rsa-prime-mesh/writeup/solution.md`
- `challenges/jeopardy/packs/crypto-rsa-prime-mesh/writeup/solve.py`
- `challenges/jeopardy/packs/forensics-mobile-timeline-fusion/challenge.yml`
- `challenges/jeopardy/packs/forensics-mobile-timeline-fusion/statement.md`
- `challenges/jeopardy/packs/forensics-mobile-timeline-fusion/attachments/case-evidence.zip`
- `challenges/jeopardy/packs/forensics-mobile-timeline-fusion/writeup/solution.md`
- `challenges/jeopardy/packs/forensics-mobile-timeline-fusion/writeup/solve.py`
- `challenges/jeopardy/packs/pwn-rop-register-chain/challenge.yml`
- `challenges/jeopardy/packs/pwn-rop-register-chain/statement.md`
- `challenges/jeopardy/packs/pwn-rop-register-chain/attachments/challenge.bin`
- `challenges/jeopardy/packs/pwn-rop-register-chain/writeup/solution.md`
- `challenges/jeopardy/packs/pwn-rop-register-chain/writeup/solve.py`
- `challenges/jeopardy/packs/pwn-rop-register-vault/challenge.yml`
- `challenges/jeopardy/packs/pwn-rop-register-vault/statement.md`
- `challenges/jeopardy/packs/pwn-rop-register-vault/attachments/challenge.bin`
- `challenges/jeopardy/packs/pwn-rop-register-vault/docker/Dockerfile`
- `challenges/jeopardy/packs/pwn-rop-register-vault/docker/challenge.bin`
- `challenges/jeopardy/packs/pwn-rop-register-vault/docker/entrypoint.sh`
- `challenges/jeopardy/packs/pwn-rop-register-vault/docker/src/challenge.c`
- `challenges/jeopardy/packs/pwn-rop-register-vault/writeup/solution.md`
- `challenges/jeopardy/packs/pwn-rop-register-vault/writeup/solve.py`
- `challenges/jeopardy/packs/reverse-vm-block-shuffle/challenge.yml`
- `challenges/jeopardy/packs/reverse-vm-block-shuffle/statement.md`
- `challenges/jeopardy/packs/reverse-vm-block-shuffle/attachments/challenge.bin`
- `challenges/jeopardy/packs/reverse-vm-block-shuffle/attachments/program.blk`
- `challenges/jeopardy/packs/reverse-vm-block-shuffle/writeup/solution.md`
- `challenges/jeopardy/packs/reverse-vm-block-shuffle/writeup/solve.py`
- `challenges/jeopardy/packs/web-ssrf-session-pivot/challenge.yml`
- `challenges/jeopardy/packs/web-ssrf-session-pivot/statement.md`
- `challenges/jeopardy/packs/web-ssrf-session-pivot/docker/Dockerfile`
- `challenges/jeopardy/packs/web-ssrf-session-pivot/docker/app.py`
- `challenges/jeopardy/packs/web-ssrf-session-pivot/writeup/solution.md`
- `challenges/jeopardy/packs/web-ssrf-session-pivot/writeup/solve.py`
- `challenges/jeopardy/packs/web-source-audit-double-wrap-01/challenge.yml`
- `challenges/jeopardy/packs/web-source-audit-double-wrap-01/writeup/solve.py`

## After implementation

- 后续批量新增 Jeopardy 题包继续复用现有 `challenge.yml + statement.md + writeup/solve.py (+ optional docker/)` 结构，不再为单题发散新的包布局
- 需要调整验证口径时，优先改 `scripts/challenges/jeopardy_batch/verify.py` 和模板文档，而不是在单题里各自发明入口
