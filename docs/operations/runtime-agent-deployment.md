# runtime-agent 部署说明

## 定位

这份说明覆盖 runtime control plane / agent split 落地后的运行模式、最小部署步骤和配置入口。

- 负责：说明本地单机模式、API + 单 remote agent 模式、正式多 node 模式的差异，以及 `runtime_agent` / `runtime_nodes` 的最小配置方式。
- 不负责：替代完整的比赛运维手册、容量规划、证书签发体系、容量智能分配方案或 live migration 方案。

## 当前模式

### 1. 本地单机开发

适用：本机开发、最小联调、排查后端与题目实例基础链路。

- API 以宿主机进程运行，`code/backend/scripts/dev-run.sh` 默认同时启动本机 `runtime-agent`
- API 侧 `runtime_agent.enabled: true`，endpoint 指向 `127.0.0.1:<本机 agent 端口>`
- 未配置 `runtime_agent.node_name` 时，`runtime_nodes` 默认写入 `agent-default`
- 本机 Docker executor 只允许在 `APP_ENV=test`，或非生产环境显式设置 `runtime_agent.allow_local_fallback: true` 时作为 fallback；生产环境会拒绝该 fallback 配置

### 2. API + 单 remote agent

适用：共享开发、内网试运行、先把控制面与执行面拆开的环境。

- API 主机运行 `ctf-api`
- 靶机宿主运行 `runtime-agent`
- API 侧 `runtime_agent.enabled: true`
- 建议显式配置同一个稳定名称：API 侧 `runtime_agent.node_name`、agent 侧 `runtime_agent.server.node_name`、数据库 `runtime_nodes.name`
- 数据库 `runtime_nodes.endpoint` 写 API 访问 runtime-agent 的控制面地址；`runtime_nodes.public_host/access_host` 写该 node 发布端口的数据面访问地址
- 未配置 `runtime_agent.node_name` 时，`runtime_nodes` 默认写入 `agent-default`
- 实例创建、清理、checker、AWD 文件写入和容器 maintenance 都按 `node_id` 路由到对应 `runtime_nodes.name`

### 3. 正式多 node

适用：多个靶机宿主共同承载实例和 AWD service。

- 每台靶机宿主运行一个 `runtime-agent`
- API 侧在 `runtime_nodes` 表中登记多个 schedulable node，并为每个 node 维护 `endpoint`、`public_host`、`access_host`
- 每个 `runtime_nodes.name` 必须对应一个 runtime-agent 的 `runtime_agent.server.node_name`
- `runtime_node_health` 后台任务会探测每个 runtime node，默认新调度只选择 `schedulable + ready / degraded` 且心跳未过期的 node
- node 离线后，该 node 上未过期的普通 `creating / running` 实例会重新进入 `pending`，由现有 scheduler 在健康节点上重建；AWD service instance 遵守 `contest_runtime_placements`，绑定 node 不可用时等待 / backoff，不静默漂移到其他 node
- 当前不声明容量智能分配、live migration 或已有 SSH/WebSocket 会话透明迁移已经落地

## 打包和启动 runtime-agent

`runtime-agent` 入口在 `code/backend/cmd/runtime-agent/main.go`，启动逻辑在 `code/backend/internal/bootstrap/runtime_agent.go`。

部署时优先把它编译成宿主机二进制，并交给 systemd 或等价进程管理器运行。仓库提供最小构建脚本：

```bash
cd code/backend
./scripts/build-runtime-agent.sh
```

默认输出为 `code/backend/bin/ctf-runtime-agent`，目标平台为 `linux/amd64`，可通过 `GOOS`、`GOARCH`、`CGO_ENABLED`、`OUT_DIR` 或第一个位置参数覆盖。例如：

```bash
cd code/backend
OUT_DIR=/tmp/ctf-release ./scripts/build-runtime-agent.sh
```

推荐的宿主机发布目录是 `/opt/ctf/backend`，因为后端配置加载会从当前工作目录下的 `configs/` 读取 `config.yaml` 和 `config.prod.yaml`。下面命令从仓库根目录执行：

```bash
sudo install -d -m 0755 /opt/ctf/backend
sudo install -m 0755 code/backend/bin/ctf-runtime-agent /opt/ctf/backend/ctf-runtime-agent
sudo install -d -m 0755 /opt/ctf/backend/configs
sudo cp -R code/backend/configs/. /opt/ctf/backend/configs/
```

systemd 模板位于：

- `code/backend/deploy/systemd/ctf-runtime-agent.service`
- `code/backend/deploy/systemd/ctf-runtime-agent.env.example`

最小安装流程：

```bash
sudo install -d -m 0750 /etc/ctf/runtime-agent
sudo install -m 0640 code/backend/deploy/systemd/ctf-runtime-agent.env.example /etc/ctf/runtime-agent/runtime-agent.env
sudo install -m 0644 code/backend/deploy/systemd/ctf-runtime-agent.service /etc/systemd/system/ctf-runtime-agent.service
sudo systemctl daemon-reload
sudo systemctl enable --now ctf-runtime-agent
```

启动前必须编辑 `/etc/ctf/runtime-agent/runtime-agent.env`，至少确认 `CTF_RUNTIME_AGENT_SERVER_*` 证书路径、监听地址、端口，以及 Docker client 连接方式。默认不设置 `DOCKER_HOST` 时，runtime-agent 会通过 Docker SDK 访问本机 `/var/run/docker.sock`。

`ctf-runtime-agent` 使用 runtime-agent 专用配置加载入口：它仍读取 `configs/` 和 `CTF_*` 环境覆盖，但只校验 agent server、Docker runtime 默认值、registry 拉取凭据和 checker sandbox。Docker 物理机不需要部署 API 进程使用的 PostgreSQL、Redis、CORS 或动态 Flag secret 配置。

后端 Dockerfile 也会把 runtime-agent 编进镜像，路径为 `/app/ctf-runtime-agent`。这只用于镜像分发或受控 smoke，不作为生产首选：容器化 runtime-agent 仍需要宿主 Docker 和 iptables 权限，通常会重新引入 docker socket、host network、`NET_ADMIN` 或 privileged 等高权限容器问题。生产默认仍推荐宿主机二进制 + systemd。

## 启动 AWD defense SSH gateway

AWD 防守 SSH 入口已经从 `ctf-api` 进程拆成独立命令，入口在 `code/backend/cmd/awd-defense-ssh-gateway/main.go`，启动逻辑在 `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go`。

开发或部署时可以直接启动：

```bash
cd code/backend
APP_ENV=prod go run ./cmd/awd-defense-ssh-gateway
```

要求：

- gateway 节点需要访问与 API 相同的 PostgreSQL / Redis
- 默认本地单机模式下，gateway 与 API 一样通过本机 runtime-agent 路由进入目标工作区容器
- 只有非生产环境显式设置 `runtime_agent.allow_local_fallback: true` 时，gateway 才会回到本机 Docker executor fallback
- `runtime_agent.enabled: true` 时，gateway 与 API 一样通过 runtime node / agent 路由进入目标工作区容器
- 多副本部署时，API 与 gateway 还必须共享同一份 `shared_storage.shared_fs.root` 挂载；当前至少会用到 `reports/*`、`challenge-attachments/*`、`runtime/flag-global-secret`、`runtime/awd-defense-ssh-host-key.pem`
- `container.defense_ssh_host` 需要配置成学生客户端实际访问的地址，通常是 TCP LB / ingress 地址；gateway 进程本身仍监听 `:container.defense_ssh_port`
- `container.defense_ssh_host_key_path` 必须在所有 gateway 副本上指向同一份预置 host key 文件；相对路径按 `shared_storage.shared_fs.root` 解析，gateway 启动时不会再为缺失路径自动生成新 key
- 需要摘流时，先对目标 gateway 进程发送 `SIGTERM`，让它先进入 draining 停止接新连接，再在 shutdown timeout 内完成 hard stop；LB 健康探测应直接看 `container.defense_ssh_port` 的 TCP 连通性，而不是额外依赖 HTTP `/ready`

开发态补充：

- `docker/docker-compose.dev.yml` 通过一次性的 `ctf-awd-defense-ssh-host-key` service 在共享 `/app/storage/runtime` 下预置 `awd-defense-ssh-host-key.pem`，随后 `ctf-awd-defense-ssh-gateway` 再按 load-only 契约启动
- 同一份 compose 同时通过 `CTF_SHARED_STORAGE_SHARED_FS_ROOT=/app/storage` 把 shared root 接到宿主持久化目录；相对配置项 `runtime/flag-global-secret`、`runtime/awd-defense-ssh-host-key.pem` 会统一落到 `/app/storage/runtime/*`
- 这只是本地联调的预置 owner。生产或共享环境仍应由部署层显式分发同一份 host key 文件，而不是把生成逻辑放回 gateway 运行时

## 最小配置

### runtime-agent 节点

runtime-agent 所在宿主机需要打开 `runtime_agent.server.*`：

```yaml
runtime_agent:
  server:
    enabled: true
    node_name: runtime-node-a
    host: 0.0.0.0
    port: 9443
    cert_file: /etc/ctf/runtime-agent/server.pem
    key_file: /etc/ctf/runtime-agent/server-key.pem
    client_ca_file: /etc/ctf/runtime-agent/ca.pem
    shutdown_timeout: 10s
```

要求：

- `runtime_agent.server.enabled` 为 `true`
- `runtime_agent.server.node_name` 是该 Docker 宿主的稳定逻辑名称，应匹配 API 侧登记的 `runtime_nodes.name`
- `cert_file` / `key_file` / `client_ca_file` 必须同时提供
- runtime-agent 进程所在节点本机仍直接连接 Docker Engine，并在本机执行 checker sandbox、ACL 和容器文件 copy / exec

### API 主机

API 侧开启 client 配置：

```yaml
runtime_agent:
  enabled: true
  allow_local_fallback: false
  node_name: runtime-node-a
  endpoint: 10.0.1.2:9443
  dial_timeout: 5s
  server_name: runtime-agent.internal
  ca_file: /etc/ctf/runtime-agent/ca.pem
  cert_file: /etc/ctf/runtime-agent/client.pem
  key_file: /etc/ctf/runtime-agent/client-key.pem
```

要求：

- `runtime_agent.enabled` 为 `true` 时，`endpoint`、`dial_timeout`、`server_name`、`ca_file`、`cert_file`、`key_file` 都必须存在
- `runtime_agent.node_name` 是 API 启动时默认 bootstrap / selector 使用的 `runtime_nodes.name`；配置后，远端 agent 的 Health 自报 `node_name` 必须与该 node row 名称一致，否则 API 不会缓存或使用这个 remote client
- `runtime_agent.allow_local_fallback` 默认为 `false`；只有非生产故障排查时才允许临时打开，生产配置会拒绝该值为 `true`
- `server_name` 要与目标 node 的 `tls_identity` 一致
- API 主机不再依赖“切 `DOCKER_HOST` 就完成多机”的假设；完整执行 authority 来自 agent 协议和 `node_id`
- API 仍负责签发 AWD defense SSH ticket，但 `2222` listener 由独立的 `awd-defense-ssh-gateway` 进程持有
- 如果部署了多个 `awd-defense-ssh-gateway` 副本，LB 只需要把新连接导向健康副本；已有 SSH 会话在某个 gateway 或 runtime node 故障时允许中断，客户端后续重连会重新走 ticket + scope 校验

## runtime node 访问地址与资源池

当前 v1 没有 runtime node management UI/API。operator 通过配置 bootstrap、数据库 seed、迁移脚本或受控 ops seed path 维护 `runtime_nodes` 行，不在学生侧或管理员 Web 页面里直接编辑 node host 字段。

字段语义：

- `runtime_nodes.endpoint`：控制面地址，API 用它拨号到对应 `runtime-agent`。
- `runtime_nodes.public_host`：学生访问该 node 发布端口时使用的 host。
- `runtime_nodes.access_host`：API、gateway、checker 或 readiness probe 访问该 node 发布端口时使用的 host。

访问 host fallback：

- public：node `public_host` -> global `container.public_host`
- access：node `access_host` -> node `public_host` -> global `container.access_host` -> global `container.public_host`

运维建议：

- 单 remote agent 环境也建议填写 node `public_host/access_host`，避免未来扩到多 node 后学生侧 URL 仍落到全局 host。
- `endpoint` 可以是内网 DNS、IP 或 agent TLS 地址；不要把它当作学生访问靶机的地址。
- `public_host` 通常是学生浏览器可达的域名 / LB / 公网地址；`access_host` 通常是 API / gateway 容器可达的内网地址。
- 如果某个 node 暂时不接收新实例，优先把 `schedulable=false`；不要删除 node row，否则会影响仍绑定该 node 的 runtime 清理和排障。

端口和子网资源池按 node 管理：

- `runtime_port_pool(runtime_node_id, port)` 控制宿主发布端口，`runtime_subnet_pool(runtime_node_id, subnet)` 控制 Docker bridge 子网。
- 同一个端口号或子网可以出现在不同 node；同一 node 内不能重复绑定。
- 启动时会按 `container.port_range_*` 和 `container.network.*` 为 node seed pool rows；已 `reserved/bound/quarantined` 的行不会被 seed 覆盖。
- Docker 返回 host port conflict 或 subnet overlap 时，冲突资源会留在对应 node pool 的 `quarantined` 状态，不释放回 `available`；operator 排查宿主机占用、Docker 残留 network 或配置冲突后，再通过受控运维脚本恢复资源状态。

### runtime node health

API 侧默认开启 runtime node 健康探测：

```yaml
container:
  runtime_node_health:
    enabled: true
    poll_interval: 10s
    probe_timeout: 2s
    stale_after: 30s
    failure_threshold: 3
```

字段含义：

- `enabled`: 开启后，默认 selector 和 execution router 会使用健康过滤；关闭后回退到旧的 `schedulable=true` 查找语义，适合作为生产回滚开关。
- `poll_interval`: 每轮探测间隔。
- `probe_timeout`: 单个 node 探测超时，探测入口复用 runtime-agent 的 managed container stats 能力。
- `stale_after`: `runtime_nodes.last_seen_at` 超过该阈值后视为心跳过期。
- `failure_threshold`: 连续探测失败达到该次数后标记 node `offline`。

运行事实：

- 成功探测会写 `runtime_nodes.health_status=ready`、`last_seen_at=now` 和 `capacity_snapshot`。
- `ready / degraded` 且心跳未过期的 schedulable node 可被新实例选择。
- `schedulable=false` 用于 cordon 新调度，不会停止健康探测；显式 `node_id` 绑定的旧容器操作在该 node 健康且心跳新鲜时仍继续路由到原 node。
- `unknown / offline`、从未成功探测或心跳过期的 node 不会被默认新调度选择。
- 显式绑定到离线或心跳过期 node 的旧容器操作会失败，不会透明切换到其他 node。
- 离线 node 上未过期的 `creating / running` 实例会被置回 `pending` 并清空旧 `node_id / container_id / network_id / runtime_details / access_url`，后续由现有 provisioning scheduler 在健康 node 上重建。

## PostgreSQL / Redis 连接基线

- PostgreSQL 连接串由后端统一生成 keyword/value DSN，并显式带 `TimeZone=UTC`；如果生产侧通过代理、VIP 或 DNS 提供单主写 HA，业务层不需要额外改配置或改 `/ready` 逻辑。
- Redis 默认仍可用 `redis.mode: single` 直连单地址；需要哨兵时改为 `redis.mode: sentinel`，并配置 `master_name`、`sentinel_addrs`，可按部署情况补 `sentinel_username`、`sentinel_password`。
- 配置里额外预留了 `redis.cluster.addrs`、`route_by_latency`、`route_randomly` 供未来 cluster 接线使用，但当前 `redis.mode=cluster` 仍会被启动校验拒绝，不能把这些字段当成已支持部署方式。
- 当前控制面只支持 Redis Sentinel 单 master failover，不把 Redis Cluster 视为本阶段部署目标。现有 Lua、pipeline、多 key 锁和排行榜逻辑都按这一前提运行。
- `/ready` 继续按 live Ping PostgreSQL / Redis 表达当前依赖可用性；依赖恢复后服务会自然重新 ready，不需要手工清理额外状态。

## node 绑定说明

- 默认节点由启动时的 `runtime_agent.enabled` 决定：
  - `false` -> `local-default`，但非 test 环境只有 `runtime_agent.allow_local_fallback: true` 时才能实际构造本地 executor
  - `true` 且配置 `runtime_agent.node_name` -> 该配置值
  - `true` 且未配置 `runtime_agent.node_name` -> `agent-default`
- `runtime_nodes.name` 是部署期稳定节点身份；`runtime_nodes.id` 和各业务表里的 `node_id` 仍是数据库内部路由键，不应配置到 runtime-agent
- `runtime_nodes.endpoint` 是 API 到 runtime-agent 的控制面地址；`runtime_nodes.public_host/access_host` 是发布端口的数据面访问地址，不改变 node identity
- runtime-agent 的 Health 会返回自报 `node_name` 和宿主 `hostname`。当远端 self-report 与 API 正在拨号的 `runtime_nodes.name` 不一致时，API 会在缓存 client 前失败，错误信息会包含 expected / reported node、endpoint 和 hostname
- 新实例启动时会把选中的 `node_id` 持久化到实例记录
- checker job metadata、AWD service instance 和容器文件写入路径都会显式携带或反查 `node_id`
- 容器清理、checker 执行、AWD 防守工作区写入和 AWD defense SSH interactive exec 都不再以“当前 API 进程连着哪台宿主机”作为 authority
- `runtime_nodes.last_seen_at` 是心跳事实源，不使用 `updated_at` 推断 node 健康状态

## runtime node failover 手工验证

正式多 node 部署完成后，可以用下面步骤做最小演练：

1. 准备 node A / node B 两个 `runtime-agent`，并在 `runtime_nodes` 中保持两个 node `schedulable=true`。
2. 确认 API 配置 `container.runtime_node_health.enabled: true`，并使用默认 `stale_after: 30s` 或按演练需要调小。
3. 在 node B 上启动一个普通实例或 AWD service runtime，确认实例记录写入 node B 的 `node_id`。
4. 停止 node B 的 `runtime-agent` 进程。
5. 等待超过 `stale_after`，或等待连续失败达到 `failure_threshold`。
6. 检查 node B 的 `runtime_nodes.health_status` 已变为 `offline`，新实例不再选择 node B。
7. 检查 node B 上未过期的 `running / creating` 实例变为 `pending`，旧 runtime identity 和 `node_id` 被清空。
8. 等待 `practice_instance_scheduler` 领取 pending 实例，确认它在 node A 或其他健康 node 上重新变为 `running`。
9. 对 AWD 比赛，确认缺失的 `team × visible service` 不再被旧 active 行阻塞，desired reconciler 会补齐新的 runtime。

预期限制：

- 原 node B 上的 TCP / HTTP / SSH / WebSocket 会话允许中断。
- 平台不会迁移旧容器或保留原会话；用户重新访问时会通过新实例的 access 入口进入重建后的 runtime。
- 如果 node B 恢复，下一轮成功探测会重新写 `ready` heartbeat，但已经重新入队并在其他 node 上创建的实例不会被迁回 node B。

## 运维限制

- 现阶段不把 `DOCKER_HOST` 直连远端 daemon 当成正式多机方案；它只能解决 Docker client 连接目标，不能替代 agent 对 ACL、checker sandbox 和容器文件副作用的 owner 收口
- 当前多 node 已支持健康过滤与 offline 后重建，但不保证容量智能均衡、live migration、原会话保持或跨 node 原地迁移
- 如果更新 node 的 endpoint 或 TLS 身份，按当前实现需要让 API 进程重新建立该节点的 runtime client；不要把在线热切节点配置写成已落地能力
