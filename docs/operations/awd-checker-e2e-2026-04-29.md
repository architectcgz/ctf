# AWD Checker E2E 验证记录 2026-04-29

## 环境

- 后端：本地 `go run ./cmd/api`
- API：`http://127.0.0.1:18080`
- PostgreSQL：`ctf-postgres`，`127.0.0.1:15432`
- Redis：`ctf-redis`，`127.0.0.1:16379`
- `AWD_CHECKER_ARTIFACT_DIR=/tmp/ctf-awd-checker-artifacts`
- `CTF_CONTAINER_DEFENSE_SSH_ENABLED=false`

## 题目包导入

已通过真实 HTTP API 完成：

- `awd-tcp-length-gate.zip`
  - 路由：`POST /api/v1/authoring/awd-challenge-imports`
  - preview：`checker_type=tcp_standard`，`service_type=binary_tcp`
  - commit：生成 AWD challenge，ID `9`

- `script-checker-files.zip`
  - 路由：`POST /api/v1/authoring/awd-challenge-imports`
  - preview：`checker_type=script_checker`，`checker_config.files` 包含 `docker/check/check.py` 和 `docker/check/protocol.py`
  - commit：生成 AWD challenge，ID `10`
  - artifact：写入 2 个私有 checker 文件

## TCP Checker 预检

使用镜像：

```bash
docker build -t ctf/awd-tcp-length-gate:latest challenges/awd/ctf-1/awd-tcp-length-gate/docker
docker run -d --rm -p 18081:8080 ctf/awd-tcp-length-gate:latest
```

验证链路：

- 创建 AWD contest：ID `10`
- `POST /api/v1/admin/contests/10/awd/checker-preview`
  - `checker_type=tcp_standard`
  - `access_url=tcp://127.0.0.1:18081`
  - 结果：`service_status=up`
- `POST /api/v1/admin/contests/10/awd/services`
  - 携带 `awd_checker_preview_token`
  - 结果：`validation_state=passed`
- `GET /api/v1/admin/contests/10/awd/readiness`
  - 结果：`ready=true`
  - `passed_challenges=1`
  - `blocking_count=0`

验证后已停止本轮启动的 TCP 服务容器。
