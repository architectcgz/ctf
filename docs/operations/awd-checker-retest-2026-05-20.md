# AWD Checker 复测记录 2026-05-20

## 实验目的

- 补做论文第 5 章所需的“当前版本 AWD checker 复测”。
- 重新验证 `AWD Checker E2E 20260429` 对应链路在 2026-05-20 本地环境中的实际表现。
- 区分“当前源码态”与“当前本地已发布镜像态”的结果，避免把历史结果直接外推到当前版本。

## 环境

- 日期：2026-05-20
- 仓库：`ctf 仓库根目录`
- API：临时启动 `ctf-api` 容器，访问 `http://127.0.0.1:8080`
- PostgreSQL：`ctf-postgres`
- Redis：`ctf-redis`
- 管理员账号：`admin / Password123`
- 复测对象：
  - 赛事：`contest_id=10`，标题 `AWD Checker E2E 20260429`
  - 服务：`service_id=29`
  - AWD 题目：`awd_challenge_id=9`，标题 `TCP 长度门禁`

## 基线

复测前数据库中 `contest_awd_services.id=29` 的 `awd_checker_validation_state=pending`，尚未形成当前版本下的复测通过证据。

## 实验一：当前本地已发布镜像态的自动 preview

### 操作

1. 启动现有 `ctf-api` 容器。
2. 登录管理员账号。
3. 对 `contest_id=10 / service_id=29` 调用：

```bash
curl -b "$COOKIE" \
  -X POST http://127.0.0.1:8080/api/v1/admin/contests/10/awd/checker-preview \
  -H "Content-Type: application/json" \
  --data @preview-body.json
```

请求体核心字段：

```json
{
  "service_id": 29,
  "awd_challenge_id": 9,
  "checker_type": "tcp_standard",
  "checker_config": {
    "timeout_ms": 3000,
    "steps": [
      {"send": "PING\n", "expect_contains": "PONG"},
      {"send_template": "SET_FLAG {{CHECKER_TOKEN}} {{FLAG}}\n", "expect_contains": "OK"},
      {"send_template": "GET_FLAG {{CHECKER_TOKEN}}\n", "expect_contains": "{{FLAG}}"}
    ]
  },
  "preview_flag": "flag{preview-20260520}",
  "preview_request_id": "paper-checker-retest-20260520"
}
```

### 结果

- preview 返回 `service_status=down`
- `preview_summary=0/3 通过`
- 错误码：`tcp_expectation_failed`
- 自动试跑生成的访问地址是 `http://host-gateway.internal:30000`

### 结论

- 当前本地已发布镜像态下，直接调用平台自动 preview 未通过。
- 这一步暴露出一个实现问题：单容器 preview 适配器会把试跑入口固定为 HTTP 访问地址，而当前题目 `awd-tcp-length-gate` 的 checker 类型是 `tcp_standard`。

## 实验二：当前本地已发布镜像与当前源码的一致性检查

### 操作

直接读取镜像内文件，对比当前仓库源码：

```bash
docker run --rm ctf/awd-tcp-length-gate:latest sh -lc 'sed -n "1,220p" /app/app.py; sed -n "1,220p" /app/ctf_runtime.py'
sed -n '1,240p' challenges/awd/ctf-1/awd-tcp-length-gate/docker/runtime/app.py
sed -n '1,240p' challenges/awd/ctf-1/awd-tcp-length-gate/docker/runtime/ctf_runtime.py
```

### 结果

- 当前源码中的 `app.py / ctf_runtime.py` 已要求 `SET_FLAG {{CHECKER_TOKEN}} {{FLAG}}` 与 `GET_FLAG {{CHECKER_TOKEN}}`
- 本地已发布镜像 `ctf/awd-tcp-length-gate:latest` 与 `ctf/awd/awd-tcp-length-gate:latest` 内仍是旧协议：
  - 只支持 `SET_FLAG <flag>`
  - 只支持 `GET_FLAG`
  - 不支持带 token 的 `GET_FLAG <token>`

### 结论

- 当前本地已发布镜像态与当前源码存在协议漂移。
- 仅引用 2026-04-29 历史结果已经不足以说明“当前版本 checker 链路仍然成立”。

## 实验三：按当前源码临时重建镜像后的对照复测

### 操作

1. 用当前源码临时构建镜像：

```bash
docker build \
  -f challenges/awd/ctf-1/awd-tcp-length-gate/docker/runtime/Dockerfile \
  -t ctf/awd-tcp-length-gate:retest-20260520 \
  challenges/awd/ctf-1/awd-tcp-length-gate/docker
```

2. 用当前源码自带本地 checker 做控制实验：

```bash
docker run -d --rm --name checker-preview-rebuilt-host-20260520 \
  -p 18084:8080 \
  -e CHECKER_TOKEN=<preview-checker-token> \
  ctf/awd-tcp-length-gate:retest-20260520

CHECKER_TOKEN=<preview-checker-token> \
python3 challenges/awd/ctf-1/awd-tcp-length-gate/docker/check/check.py 127.0.0.1 18084
```

控制实验输出：

```text
ok
```

3. 再在 `ctf-network` 内启动同一临时镜像，显式使用 `tcp://` 地址做平台 preview：

```bash
docker run -d --rm --network ctf-network --name checker-preview-rebuilt-net-20260520 \
  -e CHECKER_TOKEN=<preview-checker-token> \
  ctf/awd-tcp-length-gate:retest-20260520

curl -b "$COOKIE" \
  -X POST http://127.0.0.1:8080/api/v1/admin/contests/10/awd/checker-preview \
  -H "Content-Type: application/json" \
  --data @rebuilt-preview-body.json
```

4. 将 preview 返回的 `preview_token` 写回现有服务：

```bash
curl -b "$COOKIE" \
  -X PUT http://127.0.0.1:8080/api/v1/admin/contests/10/awd/services/29 \
  -H "Content-Type: application/json" \
  --data '{"awd_checker_preview_token":"<preview_token>"}'
```

5. 查询 readiness：

```bash
curl -b "$COOKIE" http://127.0.0.1:8080/api/v1/admin/contests/10/awd/readiness
```

### 结果

- 控制实验 `check.py` 返回 `ok`
- preview 返回：
  - `service_status=up`
  - `preview_summary=3/3 通过`
  - `preview_pass_count=3`
  - `preview_required_count=2`
  - `access_url=tcp://checker-preview-rebuilt-net-20260520:8080`
- 写回 `preview_token` 后：
  - `service_id=29` 的 `validation_state=passed`
  - readiness 返回：
    - `ready=true`
    - `passed_challenges=1`
    - `blocking_count=0`

## 复测结论

本轮复测应拆成两层结论：

1. 当前源码态结论
   - 使用 2026-05-20 当前源码临时重建的 `awd-tcp-length-gate` 镜像后，`tcp_standard` preview 可以 `3/3` 通过。
   - `preview_token` 写回后，`contest_id=10` 的 readiness 可恢复为 `ready=true`。
   - 这说明当前源码中的 token 化 checker 协议与 readiness 保存链路本身仍然成立。

2. 当前本地已发布镜像态结论
   - 直接使用本地现有镜像与自动试跑路径时，preview 会失败。
   - 失败原因至少包括两点：
     - 单容器 preview 适配器当前固定按 HTTP 入口拉起试跑实例，不适配 `tcp_standard`。
     - 本地现有 `awd-tcp-length-gate` 镜像仍是旧协议，和当前源码中的 token 化 checker 契约不一致。

## 论文可引用口径

如果要把本轮结果写进论文，建议用下面的保守口径：

- 2026-05-20 对当前源码态重新执行了 `tcp_standard` checker 复测，显式 `tcp://` 目标下 preview `3/3` 通过，写回 preview token 后 readiness 恢复为 `ready=true`。
- 同时发现当前本地已发布镜像与自动试跑路径存在协议漂移与 TCP 入口适配问题，因此历史 2026-04-29 结果不能直接视为当前发布镜像态的自动继承结论。

## 产出文件

- 原始请求与响应保存在：`/tmp/awd-checker-retest-2026-05-20`
