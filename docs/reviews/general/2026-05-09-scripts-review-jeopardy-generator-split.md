# Review: Jeopardy 生成器拆分

## Findings

- 无 material finding。

## Review Target

- Repository: `ctf`
- Scope: `scripts/challenges/generate_jeopardy_expansion_batch.py` 拆分到 `scripts/challenges/jeopardy_batch/*`，并收口 `verify_jeopardy_packs.py`
- Files reviewed:
  - `scripts/challenges/generate_jeopardy_expansion_batch.py`
  - `scripts/challenges/verify_jeopardy_packs.py`
  - `scripts/challenges/jeopardy_batch/*`
  - `docs/plan/impl-plan/2026-05-09-jeopardy-generator-split-implementation-plan.md`

## Classification Check

- 同意 `non-trivial / structural` 分类。
- touched surface 是单文件 `5422` 行脚本，本次已经把分类 builder、共享 helper、目标加载、pack 输出和 CLI 入口分开，未继续把结构债留在原文件上。

## Gate Verdict

- `pass with minor issues`

## Material Findings

- None.

## Senior Implementation Assessment

- 当前方案是这次任务的更低风险实现：
  - 保留原 CLI 路径，外部调用不需要改
  - 共享 `Target` / 路径常量，避免生成器和验证器重复定义
  - 按分类拆开 builder，后续继续补题时不会再回到单文件堆叠
- 相比直接在 5000+ 行文件里继续追加 builder，这个版本更容易定位问题，也更容易只验证某一类题型。

## Required Re-validation

- 已执行：
  - `python3 -m py_compile scripts/challenges/generate_jeopardy_expansion_batch.py scripts/challenges/verify_jeopardy_packs.py scripts/challenges/jeopardy_batch/*.py`
  - `python3 scripts/challenges/generate_jeopardy_expansion_batch.py --sync-docs`
  - `python3 scripts/challenges/generate_jeopardy_expansion_batch.py --slug crypto-repeating-xor-ledger --write`
  - `python3 scripts/challenges/verify_jeopardy_packs.py --slug crypto-repeating-xor-ledger`
  - `python3 - <<'PY' ... build_target(target) for all 65 targets ... PY`
  - `python3 scripts/challenges/generate_jeopardy_expansion_batch.py --slug web-sqli-auth-bypass --write`
  - `python3 scripts/challenges/verify_jeopardy_packs.py --slug web-sqli-auth-bypass`
- 仍建议后续在继续题包扩容前，跑完整的新增题逐题验证。

## Residual Risk

- 当前只对 `crypto`、`web` 各抽了 1 题做真实落盘和 verify smoke，尚未对全部 `65` 题做端到端写盘后验证。
- 现有老 `15` 道题仍未纳入这次统一验证链。
- 本 review 为当前会话内归档自审，不是独立 reviewer 子会话。

## Touched Known-debt Status

- 已关闭本次 touched surface 上“单文件 5000+ 行生成器继续膨胀”的结构债。
