# AWD Checker Tokenized Management Review

## Review Target

- Repository: `ctf`
- Diff source: 本地未提交改动，范围覆盖 AWD checker token 注入链路、preview/runtime freshness、practice runtime 注入、4 个 AWD 样例题包与实现计划
- Files reviewed:
  - `code/backend/internal/module/contest/application/{commands,jobs}/**`
  - `code/backend/internal/module/contest/{domain,infrastructure,ports}/**`
  - `code/backend/internal/module/practice/application/commands/{awd_defense_workspace_support.go,runtime_container_create.go,*_test.go}`
  - `challenges/awd/ctf-1/{awd-campus-drive,awd-iot-hub,awd-supply-ticket,awd-tcp-length-gate}/**`
  - `docs/plan/impl-plan/2026-05-07-awd-checker-tokenized-management-implementation-plan.md`

## Classification Check

同意按结构性改动处理。该变更同时影响 checker 契约、preview 保存语义、practice runtime 注入、题包默认 checker 配置和样例题防守边界，必须经过验证与独立 review。

## Gate Verdict

Pass with minor residual risk.

## Findings

- No material blocker found.
- `checker_token_env` 已被纳入 preview token 匹配和 validation stale 判定，避免了“checker 配置没变但 runtime 管理语义已变”时继续沿用旧 preview 结果。
- preview/runtime/checker 三段现在共享同一 token 语义，`tcp_standard` 与 `script_checker` 的模板渲染、practice runtime 注入、样例题 checker 配置已经对齐。
- `awd-tcp-length-gate` 的公网协议面已收敛到 `CHECK` 业务入口；`SET_FLAG/GET_FLAG` 改为 token 鉴权私有管理通道，设计错误已被实质关闭。

## Material Findings

- None.

## Senior Implementation Assessment

当前实现保持了最小改动面，没有引入新的通用 secret 中心，也没有重做 AWD runtime contract。核心思路是把 token 继续附着在现有 `runtime_config` / preview token / runtime env 这三条链路里，并只对需要的样例题做兼容修复。这比新建独立 secret 模块或改写 checker 类型更低风险，也更符合这次“收口设计错误”的目标。

## Required Re-validation

已执行：

```bash
go test ./internal/module/contest/application/commands ./internal/module/contest/application/jobs ./internal/module/practice/application/commands -run 'AWD|Checker|TCP|Script' -count=1
python3 -m py_compile challenges/awd/ctf-1/awd-campus-drive/docker/check/check.py challenges/awd/ctf-1/awd-iot-hub/docker/check/check.py challenges/awd/ctf-1/awd-supply-ticket/docker/check/check.py challenges/awd/ctf-1/awd-tcp-length-gate/docker/check/check.py challenges/awd/ctf-1/awd-tcp-length-gate/docker/runtime/app.py challenges/awd/ctf-1/awd-tcp-length-gate/docker/runtime/ctf_runtime.py
env PYTHONPATH=/home/azhi/workspace/projects/ctf/challenges/awd/ctf-1/awd-tcp-length-gate/docker/workspace/src PORT=18080 CHECKER_TOKEN=local-preview-token FLAG='flag{local_tcp_check}' python3 challenges/awd/ctf-1/awd-tcp-length-gate/docker/runtime/app.py
env CHECKER_TOKEN=local-preview-token python3 challenges/awd/ctf-1/awd-tcp-length-gate/docker/check/check.py 127.0.0.1 18080
git diff --check -- code/backend/internal/module/contest code/backend/internal/module/practice/application/commands challenges/awd/ctf-1/awd-campus-drive challenges/awd/ctf-1/awd-iot-hub challenges/awd/ctf-1/awd-supply-ticket challenges/awd/ctf-1/awd-tcp-length-gate docs/plan/impl-plan/2026-05-07-awd-checker-tokenized-management-implementation-plan.md docs/plan/impl-plan/2026-05-06-awd-defense-workspace-boundary-implementation-plan.md
```

建议补跑：

```bash
docker build -q -t ctf/awd-campus-drive:latest challenges/awd/ctf-1/awd-campus-drive/docker
docker build -q -t ctf/awd-iot-hub:latest challenges/awd/ctf-1/awd-iot-hub/docker
docker build -q -t ctf/awd-supply-ticket:latest challenges/awd/ctf-1/awd-supply-ticket/docker
docker build -q -t ctf/awd-tcp-length-gate:latest challenges/awd/ctf-1/awd-tcp-length-gate/docker
```

## Residual Risk

- 本 review 没有覆盖“重新 build 镜像并重新入库/registry 后，再由真实比赛实例拉起”的端到端证据。
- 三个 HTTP 样例题这次只做了 checker 配置与本地脚本兼容修复，没有补跑完整 live roundtrip；当前证据以 `py_compile`、后端定向测试和契约一致性检查为主。

## Touched Known-Debt Status

- 已触达并关闭 `awd-tcp-length-gate` “公网暴露 flag 管理命令但学生无法在 workspace 内修补”的已知设计债。
- 已触达并关闭 3 个 HTTP 样例题“题包声明了 `checker_token_env`，但 checker 仍硬编码 demo token”的兼容债。
