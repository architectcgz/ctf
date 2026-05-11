# Jeopardy 容器验证边界修复 Implementation Plan

## Objective

修复当前 Jeopardy `80` 题验证链里的“旧容器题只用本地材料模拟验证”边界，要求：

- 旧 `11` 道容器题必须通过真实 `docker build` / `docker run` 运行面完成验证
- `writeup/solve.py` 不再直接 `import docker/app.py`、本地起服务，或只在宿主机本地编译运行样本
- `scripts/challenges/verify_jeopardy_packs.py` 要能统一拉起容器、传入目标地址并清理运行资源
- 题解与 pack 材料要尽量对齐真实训练路径，不能继续依赖只有仓库内部才有的“验证捷径”

## Non-goals

- 不重做 `65` 道新增题的题型设计、flag 设计或题面文案
- 不改平台题目导入、实例编排、镜像 registry 或真实部署链
- 不把全部 `80` 题都改成容器题；本轮只收口现有旧容器题验证边界
- 不重写整套 pack 生成器；继续沿用已拆分好的 `scripts/challenges/jeopardy_batch/`

## Inputs

- `docs/plan/impl-plan/2026-05-09-jeopardy-80-pack-expansion-implementation-plan.md`
- `docs/plan/impl-plan/2026-05-09-jeopardy-generator-split-implementation-plan.md`
- `scripts/challenges/jeopardy_batch/verify.py`
- `challenges/jeopardy/packs/{crypto-stream-backup-ticket,forensics-ci-preview-artifact,misc-zero-width-briefing,pwn-length-gate,pwn-passkey-recovery-relay,pwn-ret2win-warmup,reverse-agent-cache-key,web-header-door,web-mcp-manifest-ssrf,web-notes-download,web-source-audit-double-wrap-01}/`

## Current Baseline

- 当前全量验证已经能得到 `80/80 ok`
- 但旧 `11` 道容器题里，存在以下假验证路径：
  - `web-header-door` / `pwn-length-gate` 直接导入 `docker/app.py` 或 `docker/server.py` 后在宿主机本地起服务
  - `crypto-stream-backup-ticket` / `forensics-ci-preview-artifact` / `misc-zero-width-briefing` / `reverse-agent-cache-key` / `web-mcp-manifest-ssrf` / `web-notes-download` 由 `solve.py` 直接在宿主机本地执行 `docker/app.py`
  - `pwn-ret2win-warmup` 直接在宿主机本地编译 `docker/src/challenge.c`
  - `pwn-passkey-recovery-relay` 直接本地运行附件 `relay.bin`
  - `web-source-audit-double-wrap-01` 只读静态附件 `source.html`，没有走容器暴露页面
- `pwn-ret2win-warmup` 的题面要求“下载或导出题目二进制”，但当前 pack 没有提供对应附件，训练面与验证面不一致

## Chosen Direction

这次不再让 `solve.py` 负责“本地伪造服务”，而是把容器生命周期统一收口到验证器里：

1. 在 `scripts/challenges/jeopardy_batch/` 新增容器运行辅助层，负责：
   - 识别 pack 是否属于 `runtime.type=container`
   - 基于 pack 的 `docker/` 目录构建镜像
   - 为宿主机分配临时端口并运行容器
   - 注入统一验证环境变量，例如 `FLAG`、`CTF_FLAG`
   - 等待服务 ready
   - 把 `BASE_URL`、`HOST`、`PORT`、`SERVICE_PORT` 等参数传给 `solve.py`
   - 在验证结束后强制清理容器和临时镜像
2. `solve.py` 只保留“解题动作”：
   - web / crypto / reverse / misc / forensics 类容器题：读 `BASE_URL`
   - pwn 类容器题：读 `HOST` 与 `PORT`，真实连到容器 TCP 服务
   - SSRF 类题额外读 `SERVICE_PORT`，区分“宿主机映射端口”和“容器内自访问端口”
3. 收口 touched surface 上的训练面不一致：
   - `pwn-ret2win-warmup` 补齐可分析的二进制附件，并让 `solve.py` 用附件推导目标地址后攻击真实容器服务
   - `web-source-audit-double-wrap-01` 改为从真实容器页面拉源码，再还原双层编码 flag
4. 最终验证口径改为：
   - 旧容器题必须通过“真实容器运行面 + 真实 solve 利用链”通过
   - 不再接受“本地 Python 进程模拟容器”或“只运行宿主机本地二进制”作为验题结果

## Ownership Boundary

- `scripts/challenges/jeopardy_batch/verify.py`
  - 保留为唯一 CLI 流程 owner
  - 决定是否进入容器验证路径
- `scripts/challenges/jeopardy_batch/container_runtime.py`
  - 只负责容器构建、运行、ready 检查、环境变量注入和清理
- 各 pack `writeup/solve.py`
  - 只负责题目利用与 flag 恢复
  - 不再负责服务启动、源码导入或宿主机本地构建运行面
- `pwn-ret2win-warmup` pack 内容
  - 补齐学生可分析的二进制附件，收口题面与验证面不一致

## Change Surface

- Create: `scripts/challenges/jeopardy_batch/container_runtime.py`
- Modify: `scripts/challenges/jeopardy_batch/verify.py`
- Modify: `challenges/jeopardy/packs/<11 old container slugs>/writeup/solve.py`
- Modify: `challenges/jeopardy/packs/pwn-ret2win-warmup/challenge.yml`
- Modify: `challenges/jeopardy/packs/pwn-ret2win-warmup/statement.md`
- Modify: `challenges/jeopardy/packs/pwn-ret2win-warmup/writeup/solution.md`
- Create: `challenges/jeopardy/packs/pwn-ret2win-warmup/attachments/challenge.bin`
- Modify: `docs/reports/2026-05-09-jeopardy-80-pack-verification.md`
- Modify: `docs/reviews/general/2026-05-09-jeopardy-80-pack-expansion-review.md`

## Task Slices

### Slice 1: 建立统一容器验证辅助层

目标：

- 新增容器辅助模块
- 支持镜像构建、容器运行、随机端口映射、环境变量注入、ready 探测与清理
- 为 `solve.py` 提供统一环境变量输入

Files / modules:

- `scripts/challenges/jeopardy_batch/container_runtime.py`
- `scripts/challenges/jeopardy_batch/verify.py`

Validation:

- `docker --version`
- 以 `web-header-door` 和 `pwn-length-gate` 做真实 `docker build/run` smoke

Review focus:

- 资源清理是否完整
- HTTP 与 TCP ready 探测是否区分明确
- 容器外访问地址和容器内自访问端口是否没有混淆

### Slice 2: 改造 11 个旧容器题 solve 入口

目标：

- 删除本地起服务、本地导入源码、本地编译执行路径
- 统一改为读取验证器注入的目标地址
- 保留每道题原本的真实解题动作

Files / modules:

- `challenges/jeopardy/packs/<11 old container slugs>/writeup/solve.py`

Validation:

- 对每一类至少抽 1 题做单题验证
- `web-mcp-manifest-ssrf` 需要确认 `BASE_URL` 与 `SERVICE_PORT` 路径正确
- `pwn-passkey-recovery-relay` 需要确认仍然通过真实 TCP 利用链拿到 flag

Review focus:

- solve 是否还残留本地 `subprocess.Popen(app.py)`、`importlib` 起服务或本地二进制执行路径
- 题解动作是否仍然对应学生真实训练路径

### Slice 3: 收口 pwn-ret2win-warmup 的材料边界

目标：

- 为 `pwn-ret2win-warmup` 补齐可分析的二进制附件
- 更新 pack 元数据、题面和题解，让“分析附件 -> 打远端服务”成为一致路径

Files / modules:

- `challenges/jeopardy/packs/pwn-ret2win-warmup/attachments/challenge.bin`
- `challenges/jeopardy/packs/pwn-ret2win-warmup/challenge.yml`
- `challenges/jeopardy/packs/pwn-ret2win-warmup/statement.md`
- `challenges/jeopardy/packs/pwn-ret2win-warmup/writeup/solution.md`
- `challenges/jeopardy/packs/pwn-ret2win-warmup/writeup/solve.py`

Validation:

- 对附件运行 `nm -an` 能取到 `win`
- `solve.py` 使用附件地址并真实攻击容器端口成功

Review focus:

- 题面、附件和验证链是否一致
- 是否继续依赖仅仓库内部可见的 `docker/src/challenge.c`

### Slice 4: 重跑全量验证并刷新报告

目标：

- 让 `verify_jeopardy_packs.py` 默认全量验证时真正覆盖真实容器链
- 刷新报告，去掉“旧容器题只做本地模拟”的残余口径

Files / modules:

- `scripts/challenges/jeopardy_batch/verify.py`
- `docs/reports/2026-05-09-jeopardy-80-pack-verification.md`

Validation:

- `python3 scripts/challenges/verify_jeopardy_packs.py --slug web-header-door`
- `python3 scripts/challenges/verify_jeopardy_packs.py --slug pwn-ret2win-warmup`
- `python3 scripts/challenges/verify_jeopardy_packs.py --write-report`

Review focus:

- 报告是否只记录真实跑过的结果
- 旧容器题是否全部进入容器验证路径

### Slice 5: Review 收尾

目标：

- 用 review 心智检查边界是否真的被收口
- 刷新 review 文档并跑仓库收尾检查

Files / modules:

- `docs/reviews/general/2026-05-09-jeopardy-80-pack-expansion-review.md`

Validation:

- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

Review focus:

- 是否仍有 pack 的 solve 依赖仓库内部运行捷径
- 是否仍有 touched surface 上的训练面与验证面不一致

## Risks

- `web-mcp-manifest-ssrf` 需要同时处理宿主机映射端口与容器内回环端口，最容易把 SSRF 目标打错
- `pwn-ret2win-warmup` 若只改验证器、不补附件，会继续保留题面与训练面的不一致
- 全量 `80` 题重验会消耗更多时间，必须串行执行并保证每题容器都能清理

## Verification Plan

1. `python3 -m py_compile scripts/challenges/generate_jeopardy_expansion_batch.py scripts/challenges/verify_jeopardy_packs.py scripts/challenges/jeopardy_batch/*.py`
2. `python3 scripts/challenges/verify_jeopardy_packs.py --slug web-header-door`
3. `python3 scripts/challenges/verify_jeopardy_packs.py --slug web-mcp-manifest-ssrf`
4. `python3 scripts/challenges/verify_jeopardy_packs.py --slug pwn-passkey-recovery-relay`
5. `python3 scripts/challenges/verify_jeopardy_packs.py --slug pwn-ret2win-warmup`
6. `python3 scripts/challenges/verify_jeopardy_packs.py --write-report`
7. `bash scripts/check-consistency.sh`
8. `bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- 目标边界明确：容器生命周期 owner 在验证器，solve 只保留解题逻辑
- 共享能力 landing zone 明确：统一落到 `scripts/challenges/jeopardy_batch/container_runtime.py`
- 本次不是只修“输出看起来 80/80 ok”，而是收口“旧容器题验证不走真实运行面”的结构问题
- touched surface 上已知的 `pwn-ret2win-warmup` 材料不一致被纳入同一切片，不留作 follow-up
- 完成后不应再需要马上做第二轮“把本地模拟改回真容器”的重复改造
