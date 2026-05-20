> 状态：Current
> 事实源：`docker/ctf/docker-compose.dev.yml`、`code/backend/internal/module/practice/`、`code/backend/internal/module/runtime/`
> 替代：无

# Docker Runtime Access Host Implementation Plan

## Objective

让 `ctf-api` 运行在 Docker 容器里时，容器题实例仍能完整走通：

- 题包上传
- 镜像 build / push
- 学生开题拉镜像
- TCP 直连或 HTTP 代理访问

## Non-goals

- 不改数据库结构。
- 不引入新的部署拓扑。
- 不把用户侧访问地址统一改成内部宿主网关地址。

## Inputs

- `docs/architecture/backend/03-container-architecture.md`
- `docker/ctf/docker-compose.dev.yml`
- `docker/ctf/infra/README.md`
- `code/backend/internal/module/practice/application/commands/instance_provisioning.go`
- `code/backend/internal/module/runtime/api/http/handler.go`

## Task slices

1. 配置切片
- 新增 `container.access_host`，保留 `container.public_host` 作为学生侧地址。
- 本地 compose dev 注入 `CTF_CONTAINER_ACCESS_HOST=host-gateway.internal`，并用 `extra_hosts` 显式映射。

2. 运行时切片
- 容器创建后，内部保存可由后端访问的 published access URL。
- 就绪探测和 HTTP 代理继续使用内部 access URL。
- 对外响应在 TCP / 实例列表 / 实例详情场景重写回 `public_host`。

3. 文档切片
- 更新 infra README、容器编排架构文档和本地测试账号文档。

## Completion checklist

- [x] Slice 1: 配置切片已完成
- [x] Slice 2: 运行时切片已完成
- [x] Slice 3: 文档切片已完成

## Expected files

- `code/backend/internal/config/config.go`
- `code/backend/internal/config/config_test.go`
- `code/backend/internal/model/topology.go`
- `code/backend/internal/model/topology_runtime_test.go`
- `code/backend/internal/module/practice/application/commands/*`
- `code/backend/internal/module/runtime/*`
- `code/backend/internal/module/instance/*`
- `docker/ctf/docker-compose.dev.yml`
- `docker/ctf/infra/README.md`
- `docs/architecture/backend/03-container-architecture.md`
- `docs/requirements/local-dev-test-credentials.md`

## Validation

- `go test ./internal/config ./internal/model ./internal/module/instance/application/... ./internal/module/practice/application/... ./internal/module/runtime/...`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`
- 真实 Docker 验证：
  - `ctf-api` 容器健康检查通过
  - 管理员导入并发布容器题
  - 删除本地镜像缓存后，学生开题成功重新拉回镜像
  - `pwn-rop-register-vault-e2e-*` 返回 `tcp://127.0.0.1:*`
  - `web-ssrf-session-pivot-e2e-*` 返回平台代理 URL，并能通过代理读到题目首页

## Risks

- 如果宿主机的 Docker network 残留过多孤儿网络，实例创建会因为地址池耗尽失败，需要先清理 `ctf-net-*` 孤儿网络。
- `container.access_host` 只应给后端进程内部链路使用，不能直接作为学生侧公开地址。
