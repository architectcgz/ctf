# AWD Runtime Workspace Src Import Priority Implementation Plan

**Goal:** 修复 AWD 题运行容器中 `challenge_app` 的导入优先级，让蓝方工作区 `/workspace/src` 的补丁在重启后稳定覆盖镜像内 `/app/challenge_app.py`，并用机械检查防止后续题目再次回归。

**Architecture:** 保持现有 AWD runtime 结构不变，不改题目业务逻辑和实例代理链路；只修运行入口的 `sys.path` 排序语义，并在仓库一致性检查中固化这个约束。

**Tech Stack:** Python runtime entrypoint, Docker challenge runtime, shell consistency checks

---

## Objective

- 修复所有使用 `/workspace/src` 覆盖机制的 AWD runtime 入口，让 `/workspace/src` 无条件排到 `sys.path[0]`
- 保证蓝方在容器内替换 `workspace/src/challenge_app.py` 后，重启服务即可加载最新补丁
- 在 `scripts/check-consistency.sh` 中新增 AWD runtime 导入优先级检查，阻止后续新题继续使用有缺陷的写法

## Non-goals

- 不修改任何 AWD 题目的业务代码、flag 逻辑、数据库内容或题目漏洞本身
- 不调整代理鉴权、实例启动、轮次切换或 SSH 防守链路
- 不移除镜像中的 `/app/challenge_app.py` 备份副本，也不重构现有 Dockerfile 模式

## Inputs

- `challenges/awd/ctf-1/awd-supply-ticket/docker/runtime/app.py`
- `challenges/awd/ctf-1/awd-campus-drive/docker/runtime/app.py`
- `challenges/awd/ctf-1/awd-iot-hub/docker/runtime/app.py`
- `challenges/awd/ctf-1/awd-tcp-length-gate/docker/runtime/app.py`
- `challenges/awd/ctf-2/awd-passkey-sync-gateway/docker/runtime/app.py`
- `challenges/awd/ctf-2/awd-webhook-inspector/docker/runtime/app.py`
- `challenges/awd/ctf-3/awd-forwarded-admin-gateway/docker/runtime/app.py`
- `challenges/awd/ctf-3/awd-iot-fleet-orchestrator/docker/runtime/app.py`
- `challenges/awd/ctf-3/awd-patch-signing-gateway/docker/runtime/app.py`
- `challenges/awd/ctf-3/awd-preview-render-farm/docker/runtime/app.py`
- `challenges/awd/ctf-3/awd-webhook-relay-hub/docker/runtime/app.py`
- `scripts/check-consistency.sh`

## Ownership Boundary

- `challenges/awd/**/docker/runtime/app.py`
  - 负责：启动 Flask runtime 前整理 Python 模块搜索路径并导入 `challenge_app`
  - 不负责：题目业务修复、数据库迁移、实例路由或 ticket 鉴权
- `scripts/check-consistency.sh`
  - 负责：拦截 AWD runtime 入口的已知错误模式，保证后续新增或修改题目时不回归
  - 不负责：验证题目漏洞是否被蓝方业务补丁修复

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-15-awd-runtime-workspace-src-import-priority-implementation-plan.md`
- Modify: `challenges/awd/ctf-1/awd-campus-drive/docker/runtime/app.py`
- Modify: `challenges/awd/ctf-1/awd-iot-hub/docker/runtime/app.py`
- Modify: `challenges/awd/ctf-1/awd-supply-ticket/docker/runtime/app.py`
- Modify: `challenges/awd/ctf-1/awd-tcp-length-gate/docker/runtime/app.py`
- Modify: `challenges/awd/ctf-2/awd-passkey-sync-gateway/docker/runtime/app.py`
- Modify: `challenges/awd/ctf-2/awd-webhook-inspector/docker/runtime/app.py`
- Modify: `challenges/awd/ctf-3/awd-forwarded-admin-gateway/docker/runtime/app.py`
- Modify: `challenges/awd/ctf-3/awd-iot-fleet-orchestrator/docker/runtime/app.py`
- Modify: `challenges/awd/ctf-3/awd-patch-signing-gateway/docker/runtime/app.py`
- Modify: `challenges/awd/ctf-3/awd-preview-render-farm/docker/runtime/app.py`
- Modify: `challenges/awd/ctf-3/awd-webhook-relay-hub/docker/runtime/app.py`
- Modify: `scripts/check-consistency.sh`

## Task 1: 统一修正 AWD runtime 导入优先级

- [ ] 将所有受影响 `docker/runtime/app.py` 中的“仅在 `sys.path` 缺失时才插入 `/workspace/src`”改为“始终把 `/workspace/src` 提到最前”
- [ ] 保持现有运行入口结构和导入目标不变，只修路径优先级

验证：

- `cd /home/azhi/workspace/projects/ctf && rg -n 'if str\\(WORKSPACE_SRC\\) not in sys.path' challenges/awd`
- `cd /home/azhi/workspace/projects/ctf && rg -n 'sys.path = \\[workspace_src\\] \\+ \\[p for p in sys.path if p != workspace_src\\]' challenges/awd`

Review focus：

- 是否所有 AWD runtime 入口都统一成同一修复模式
- 是否避免因为重复插入造成 `sys.path` 中出现多个 `/workspace/src`

## Task 2: 增加一致性检查 guardrail

- [ ] 在 `scripts/check-consistency.sh` 中新增 AWD runtime 入口检查
- [ ] 只扫描真实 runtime 入口文件，避免题解或文档中的示例代码误报
- [ ] 检查失败时给出明确报错，便于后续新增题目时快速定位

验证：

- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`

Review focus：

- 检查范围是否刚好覆盖所有 AWD runtime 入口
- 是否能同时拦住旧写法和缺失强制置顶逻辑的变体

## Task 3: 运行级验证当前复现题

- [ ] 重建当前 `awd-supply-ticket` 对应实例容器或运行环境
- [ ] 在运行容器中确认 `challenge_app` 实际导入来源变为 `/workspace/src/challenge_app.py`
- [ ] 通过真实代理链路复现旧漏洞路径，确认不再从镜像内旧 `/app/challenge_app.py` 取到 flag

验证：

- `docker exec ctf-instance-challenge-c35-t69-s30 python -c "import importlib.util; spec=importlib.util.find_spec('challenge_app'); print(spec.origin)"`
- `cd /home/azhi/workspace/projects/ctf && docker compose -f docker/ctf/docker-compose.dev.yml up -d --build ctf-api`
- 复用 contest proxy 链路验证 `/api/v1/contests/35/awd/services/30/targets/69/proxy/notify/2`

Review focus：

- 失败时要区分“运行时仍导入旧文件”和“数据库里残留旧 payload”两类问题
- 不能只做静态 grep，就声称蓝方补丁在真实运行里已生效

## Risks

- 如果只修当前题而漏掉其他 runtime 入口，后续 AWD 新比赛仍会继续出现同类问题
- 如果 guardrail 扫描范围过宽，题解中的示例代码可能导致误报
- 如果只确认文件已替换、不确认 Python 实际导入来源，仍可能误判为“重启未生效”

## Verification Plan

1. `cd /home/azhi/workspace/projects/ctf && git diff --check`
2. `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
3. `cd /home/azhi/workspace/projects/ctf && docker compose -f docker/ctf/docker-compose.dev.yml up -d --build ctf-api`
4. `docker exec ctf-instance-challenge-c35-t69-s30 python -c "import importlib.util; spec=importlib.util.find_spec('challenge_app'); print(spec.origin)"`
5. 通过当前比赛代理链路复验 `target 69` 的 `/notify/2` 行为

## Architecture-Fit Evaluation

- owner 明确：运行入口只负责导入优先级，不把“蓝方补丁是否生效”分散到题目业务代码或平台代理层兜底
- reuse point 明确：所有 AWD runtime 入口沿用同一 `sys.path` 置顶模式，而不是每道题单独解释或单独补救
- 结构收敛明确：把“工作区补丁优先于镜像初始文件”固化成运行时不变量和一致性检查，而不是继续依赖人工重启排查
