# runtime-agent 部署说明

## 定位

这份说明覆盖 runtime control plane / agent split 落地后的运行模式、最小部署步骤和配置入口。

- 负责：说明本地单机模式、API + 单 remote agent 模式、正式多 node 模式的差异，以及 `runtime_agent` / `runtime_nodes` 的最小配置方式。
- 不负责：替代完整的比赛运维手册、容量规划、证书签发体系或多机自动均衡方案。

## 当前模式

### 1. 本地单机开发

适用：本机开发、最小联调、排查后端与题目实例基础链路。

- API 以宿主机进程运行，`runtime_agent.enabled: false`
- `runtime_nodes` 默认写入 `local-default`
- API 直接复用本机 executor；这是开发 fallback，不是正式多机边界

### 2. API + 单 remote agent

适用：共享开发、内网试运行、先把控制面与执行面拆开的环境。

- API 主机运行 `ctf-api`
- 靶机宿主运行 `runtime-agent`
- API 侧 `runtime_agent.enabled: true`
- `runtime_nodes` 默认写入 `agent-default`
- 实例创建、清理、checker、AWD 文件写入和容器 maintenance 都按 `node_id` 路由到 `agent-default`

### 3. 正式多 node

适用：多个靶机宿主共同承载实例和 AWD service。

- 每台靶机宿主运行一个 `runtime-agent`
- API 侧在 `runtime_nodes` 表中登记多个 schedulable node
- 当前调度只保证“显式绑定或单节点默认”正确，不声明自动均衡、故障迁移或容量智能分配已经落地

## 启动 runtime-agent

`runtime-agent` 入口在 `code/backend/cmd/runtime-agent/main.go`，启动逻辑在 `code/backend/internal/bootstrap/runtime_agent.go`。

开发或部署时可以直接启动：

```bash
cd code/backend
APP_ENV=prod go run ./cmd/runtime-agent
```

生产环境更推荐编译成独立二进制后交给 systemd 或等价进程管理器运行。

## 启动 AWD defense SSH gateway

AWD 防守 SSH 入口已经从 `ctf-api` 进程拆成独立命令，入口在 `code/backend/cmd/awd-defense-ssh-gateway/main.go`，启动逻辑在 `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go`。

开发或部署时可以直接启动：

```bash
cd code/backend
APP_ENV=prod go run ./cmd/awd-defense-ssh-gateway
```

要求：

- gateway 节点需要访问与 API 相同的 PostgreSQL / Redis
- 本地单机模式下，gateway 节点本机需要直接访问 Docker Engine
- `runtime_agent.enabled: true` 时，gateway 与 API 一样通过 runtime node / agent 路由进入目标工作区容器
- `container.defense_ssh_host` 需要配置成学生客户端实际访问的地址，通常是 TCP LB / ingress 地址；gateway 进程本身仍监听 `:container.defense_ssh_port`
- `container.defense_ssh_host_key_path` 必须在所有 gateway 副本上指向同一份预置 host key 文件；gateway 启动时不会再为缺失路径自动生成新 key
- 需要摘流时，先对目标 gateway 进程发送 `SIGTERM`，让它先进入 draining 停止接新连接，再在 shutdown timeout 内完成 hard stop；LB 健康探测应直接看 `container.defense_ssh_port` 的 TCP 连通性，而不是额外依赖 HTTP `/ready`

开发态补充：

- `docker/docker-compose.dev.yml` 通过一次性的 `ctf-awd-defense-ssh-host-key` service 在共享 `/app/storage/runtime` 下预置 `awd-defense-ssh-host-key.pem`，随后 `ctf-awd-defense-ssh-gateway` 再按 load-only 契约启动
- 这只是本地联调的预置 owner。生产或共享环境仍应由部署层显式分发同一份 host key 文件，而不是把生成逻辑放回 gateway 运行时

## 最小配置

### runtime-agent 节点

runtime-agent 所在宿主机需要打开 `runtime_agent.server.*`：

```yaml
runtime_agent:
  server:
    enabled: true
    host: 0.0.0.0
    port: 9443
    cert_file: /etc/ctf/runtime-agent/server.pem
    key_file: /etc/ctf/runtime-agent/server-key.pem
    client_ca_file: /etc/ctf/runtime-agent/ca.pem
    shutdown_timeout: 10s
```

要求：

- `runtime_agent.server.enabled` 为 `true`
- `cert_file` / `key_file` / `client_ca_file` 必须同时提供
- runtime-agent 进程所在节点本机仍直接连接 Docker Engine，并在本机执行 checker sandbox、ACL 和容器文件 copy / exec

### API 主机

API 侧开启 client 配置：

```yaml
runtime_agent:
  enabled: true
  endpoint: 10.0.1.2:9443
  dial_timeout: 5s
  server_name: runtime-agent.internal
  ca_file: /etc/ctf/runtime-agent/ca.pem
  cert_file: /etc/ctf/runtime-agent/client.pem
  key_file: /etc/ctf/runtime-agent/client-key.pem
```

要求：

- `runtime_agent.enabled` 为 `true` 时，`endpoint`、`dial_timeout`、`server_name`、`ca_file`、`cert_file`、`key_file` 都必须存在
- `server_name` 要与目标 node 的 `tls_identity` 一致
- API 主机不再依赖“切 `DOCKER_HOST` 就完成多机”的假设；完整执行 authority 来自 agent 协议和 `node_id`
- API 仍负责签发 AWD defense SSH ticket，但 `2222` listener 由独立的 `awd-defense-ssh-gateway` 进程持有
- 如果部署了多个 `awd-defense-ssh-gateway` 副本，LB 只需要把新连接导向健康副本；已有 SSH 会话在某个 gateway 或 runtime node 故障时允许中断，客户端后续重连会重新走 ticket + scope 校验

## PostgreSQL / Redis 连接基线

- PostgreSQL 连接串由后端统一生成 keyword/value DSN，并显式带 `TimeZone=UTC`；如果生产侧通过代理、VIP 或 DNS 提供单主写 HA，业务层不需要额外改配置或改 `/ready` 逻辑。
- Redis 默认仍可用 `redis.mode: single` 直连单地址；需要哨兵时改为 `redis.mode: sentinel`，并配置 `master_name`、`sentinel_addrs`，可按部署情况补 `sentinel_username`、`sentinel_password`。
- 配置里额外预留了 `redis.cluster.addrs`、`route_by_latency`、`route_randomly` 供未来 cluster 接线使用，但当前 `redis.mode=cluster` 仍会被启动校验拒绝，不能把这些字段当成已支持部署方式。
- 当前控制面只支持 Redis Sentinel 单 master failover，不把 Redis Cluster 视为本阶段部署目标。现有 Lua、pipeline、多 key 锁和排行榜逻辑都按这一前提运行。
- `/ready` 继续按 live Ping PostgreSQL / Redis 表达当前依赖可用性；依赖恢复后服务会自然重新 ready，不需要手工清理额外状态。

## node 绑定说明

- 默认节点由启动时的 `runtime_agent.enabled` 决定：
  - `false` -> `local-default`
  - `true` -> `agent-default`
- 新实例启动时会把选中的 `node_id` 持久化到实例记录
- checker job metadata、AWD service instance 和容器文件写入路径都会显式携带或反查 `node_id`
- 容器清理、checker 执行、AWD 防守工作区写入和 AWD defense SSH interactive exec 都不再以“当前 API 进程连着哪台宿主机”作为 authority

## 运维限制

- 现阶段不把 `DOCKER_HOST` 直连远端 daemon 当成正式多机方案；它只能解决 Docker client 连接目标，不能替代 agent 对 ACL、checker sandbox 和容器文件副作用的 owner 收口
- 当前多 node 只保证绑定正确，不保证自动均衡、故障转移和跨 node 重调度
- 如果更新 node 的 endpoint 或 TLS 身份，按当前实现需要让 API 进程重新建立该节点的 runtime client；不要把在线热切节点配置写成已落地能力
