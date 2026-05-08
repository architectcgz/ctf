# Batch Modern Challenge Packs Implementation Plan

**Goal:** 按当前题目包规范新增一批“与时俱进”的 Jeopardy 与 AWD 题目，覆盖 `web / pwn / reverse / crypto / misc / forensics` 六类 Jeopardy，并补至少两道 AWD 题；所有容器题都完成本地 build、registry push、导入或同步、启动验证与可解性验证，所有题都补齐官方题解。

**Non-goals:**
- 不改题包解析器、registry 脚本或平台导入契约。
- 不做新的题目分类扩展，继续沿用现有六大类和 AWD 扩展格式。
- 不补整套新的比赛编排或 AWD 轮次数据，只验证题包本身与容器运行链路。

**Chosen direction:** 为了满足“注册到 registry、可以启动、可以成功做题”的交付要求，这一批 Jeopardy 统一做成 `runtime.type=container` 的可运行题，不再混入 `runtime.type=none` 的离线题。不同分类通过“下载样本 + 服务利用”组合实现：`reverse / crypto / misc / forensics` 仍保留各自分析特点，但最终都通过运行中的服务兑换动态 Flag。

**Inputs / evidence:**
- `challenges/teacher-authoring-guide.md`
- `challenges/jeopardy/templates/README.md`
- `challenges/awd/README.md`
- `challenges/awd/challenge-package-contract.md`
- `docs/contracts/challenge-pack-v1.md`
- `docs/architecture/features/题包Registry交付架构.md`
- `scripts/registry/build-and-push-challenge-image.sh`
- `scripts/registry/sync-challenge-registry-state.py`

**Target packs:**
- Jeopardy
  - `web-mcp-manifest-ssrf`
  - `pwn-passkey-recovery-relay`
  - `reverse-agent-cache-key`
  - `crypto-stream-backup-ticket`
  - `misc-zero-width-briefing`
  - `forensics-ci-preview-artifact`
- AWD
  - `awd-webhook-inspector`
  - `awd-passkey-sync-gateway`

**Tech stack:** Python 3 + Flask、C / GCC、Docker、zip、平台现有 registry / import API。

---

### Task 1: 先补方案文档与题型边界

**Files:**
- Create: `docs/plan/impl-plan/2026-05-08-batch-modern-challenge-packs-implementation-plan.md`

- [ ] **Step 1: 固定题目分布与主题**

要求：
- Jeopardy 六类都覆盖到。
- 题面主题尽量贴近 `AI agent / MCP / passkey / CI artifact / webhook / browser backup` 等近两年的常见技术语境。
- 每题都只有一条主解法，难度控制在 `easy ~ medium`，优先保证可交付和可验证。

- [ ] **Step 2: 固定运行与验证路线**

要求：
- Jeopardy 全部采用容器题，保证都能 build / push / start。
- AWD 保持 `runtime / workspace / check` 三层结构与 `defense_workspace` 契约。
- 每题都补一份题解文档，并准备可重放的最小利用命令或脚本。

### Task 2: 批量新增 Jeopardy 题目包

**Files:**
- Create under `challenges/jeopardy/packs/`:
  - `web-mcp-manifest-ssrf/**`
  - `pwn-passkey-recovery-relay/**`
  - `reverse-agent-cache-key/**`
  - `crypto-stream-backup-ticket/**`
  - `misc-zero-width-briefing/**`
  - `forensics-ci-preview-artifact/**`
- Create under `challenges/jeopardy/dist/`:
  - `<slug>.zip` for all 6 packs

- [ ] **Step 1: Web / Pwn 题**

要求：
- `web-mcp-manifest-ssrf` 用 SSRF 绕过本地地址黑名单，读取内部 flag 接口。
- `pwn-passkey-recovery-relay` 提供可下载二进制或等价审计材料，利用 `ret2win` 获取动态 Flag。

验证：
- `docker build` / `docker run`
- 用 `curl` 或 `python` exploit 实际拿到动态 Flag

- [ ] **Step 2: Reverse / Crypto 题**

要求：
- `reverse-agent-cache-key` 通过下载到的 `pyc` / 二进制样本逆向出兑换码，再兑换动态 Flag。
- `crypto-stream-backup-ticket` 围绕流密码 key/nonce 复用恢复兑换码，再兑换动态 Flag。

验证：
- 脚本化恢复兑换码
- 命中兑换接口并拿到动态 Flag

- [ ] **Step 3: Misc / Forensics 题**

要求：
- `misc-zero-width-briefing` 通过零宽字符等隐藏编码恢复兑换码。
- `forensics-ci-preview-artifact` 通过下载工件、分析日志或元数据恢复访问 token。

验证：
- 下载材料有效
- 题解命令能稳定恢复 token / code
- 命中服务后拿到动态 Flag

- [ ] **Step 4: 补题解与分发包**

要求：
- 每题至少包含 `challenge.yml`、`statement.md`、`writeup/solution.md`。
- 如需额外可复现实验脚本，可放入 `writeup/` 或 `attachments/`，但题解正文要能独立解释清楚。
- `dist/<slug>.zip` 解压后根目录仍为 `<slug>/`。

### Task 3: 批量新增 AWD 题目包

**Files:**
- Create under `challenges/awd/ctf-2/`:
  - `awd-webhook-inspector/**`
  - `awd-passkey-sync-gateway/**`
- Modify: `challenges/awd/README.md`

- [ ] **Step 1: Web AWD**

要求：
- `awd-webhook-inspector` 使用现代 Web 运维主题，业务漏洞放在 `docker/workspace/src/`。
- `writeup/attack.md` 描述攻击路径，`writeup/defense.md` 描述最小修补思路。
- `docker/check/check.py` 覆盖健康检查与 `put_flag -> get_flag` 闭环。

- [ ] **Step 2: TCP AWD**

要求：
- `awd-passkey-sync-gateway` 使用文本 TCP 协议或等价的单端口服务。
- 业务漏洞不能依赖 checker 私有通道；攻击面必须在业务命令内。
- 同样补齐 `attack.md / defense.md / check.py`。

验证：
- `docker compose up --build`
- 本地 checker 通过
- 本地攻击脚本或手工命令能拿到动态 Flag

### Task 4: Registry、导入与运行验证

**Files:**
- Runtime only: local registry state, local API import records, PostgreSQL runtime records

- [ ] **Step 1: 生成 zip 与构建镜像**

Run:
```bash
# 逐题执行
zip -r challenges/jeopardy/dist/<slug>.zip challenges/jeopardy/packs/<slug>
bash scripts/registry/build-and-push-challenge-image.sh <pack_dir> --tag 20260508
```

要求：
- Jeopardy 容器题与 AWD 题都推送到 `127.0.0.1:5000/<mode>/<slug>:20260508`
- `challenge.yml` 中的 `runtime.image.ref` 与最终交付一致

- [ ] **Step 2: 导入或同步题目记录**

要求：
- Jeopardy 包通过 `POST /api/v1/authoring/challenge-imports` 与 `/commit` 导入。
- AWD 包通过 `POST /api/v1/authoring/awd-challenge-imports` 与 `/commit` 导入。
- 如导入后镜像引用未完全回灌，再补跑 `python3 scripts/registry/sync-challenge-registry-state.py <pack_dir>`。

- [ ] **Step 3: 启动与做题验证**

要求：
- Jeopardy 至少抽每个分类 1 题，走实际启动实例与提交 Flag 链路。
- AWD 至少验证容器启动、checker 成功、攻击路径能读取动态 Flag。
- 只记录实际运行过的命令和结果，不写“理论可解”。

### Task 5: Review 与交付说明

**Files:**
- Create: `docs/reviews/general/2026-05-08-batch-modern-challenge-packs-review.md`

- [ ] **Step 1: 切到 review 心智**

Review focus:
- 题面是否和实际入口一致。
- 题解是否真的能复现，不依赖仓库内源码捷径。
- registry / import / startup 验证是否真实覆盖到了新增题。
- AWD 是否把业务漏洞和受保护运行时代码物理隔离开。

- [ ] **Step 2: findings 修复后重跑受影响验证**

Run:
```bash
git diff --check
```

Expected:
- 没有格式性问题。
- 最终说明里明确哪些链路已跑，哪些因为平台边界只做了本地验证。

---

## Rollback / recovery

- 题目目录级回退：删除本次新增的 `packs/<slug>`、`dist/<slug>.zip`、`ctf-2/<slug>`。
- registry 回退：删除本次新 push 的 `127.0.0.1:5000/jeopardy/<slug>:20260508`、`127.0.0.1:5000/awd/<slug>:20260508` 标签。
- 导入记录回退：通过数据库按新增 `slug` 清理对应 `challenges / awd_challenges / images` 记录，必要时先做 `pg_dump` 备份。
