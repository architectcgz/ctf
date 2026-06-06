# CTF 平台项目

本仓库包含平台代码、架构与契约文档、题目与题包，以及开发过程中沉淀的规则和资料。

文档入口：

- 架构和页面设计：`docs/architecture/`
- 接口与题包契约：`docs/contracts/`
- 题目与题包：`challenges/`

## 当前技术栈

- 后端：Go 1.24、Gin、GORM、pgx、Viper、PostgreSQL、Redis、Zap
- 前端：Vue 3.5、Vite 7、TypeScript 5、Pinia、Vue Router、Tailwind CSS 4、Axios、VueUse、Vitest
- 运行与联调：Docker Compose、PostgreSQL 16、Redis 7、可选本地 registry

## 推荐启动方式

推荐日常开发使用“依赖容器 + 本地前后端”。准备好 Docker、Go 1.24、Node.js 之后，通常只需要两条命令：

1. 启动后端热重载

```bash
go install github.com/air-verse/air@latest
cd code/backend
./scripts/dev-run.sh --infra --migrate --hot
```

2. 启动前端

```bash
cd code/frontend
npm install
npm run dev
```

启动后默认访问：

- 前端：`http://127.0.0.1:5173`
- 后端：`http://127.0.0.1:8080`

说明：

- `./scripts/dev-run.sh --infra --migrate --hot` 会自动启动本地 PostgreSQL / Redis 容器，补齐开发环境变量，并在启动前执行数据库迁移
- 如果 `8080` 已被占用，脚本会自动把后端切到 `18080`
- 如果不需要热重载，也可以用 `cd code/backend && ./scripts/dev-run.sh --infra --migrate`
- 如果只想直接运行后端，也可以用 `cd code/backend && APP_ENV=dev go run ./cmd/api`，但这时数据库和 Redis 相关环境变量需要自己提供
- 这是默认推荐路径。当前架构假设 API 迟早会出现实现或配置缺陷，因此默认开发链路不让 `ctf-api` 容器直接持有宿主 Docker daemon 控制权；部署边界说明见 `docs/architecture/backend/01-system-architecture.md` 的“7.5 安全边界设计”

默认开发账号由初始迁移写入，密码均为 `Password123`：

- `admin`：管理员账号
- `teacher`：教师账号，班级 `CTF-1`
- `student`：学员账号，班级 `CTF-1`
- `student2`：学员账号，班级 `CTF-1`

## 全容器联调

只有在需要验证完整容器编排、Nginx 反代、容器网络或运行时行为时，再启动整套容器：

```bash
CTF_HOST_ROOT="$(pwd)" docker compose -f docker/ctf/docker-compose.dev.yml up -d --build
```

这条路径下，`ctf-api` 容器会在启动应用前先执行一次 `/app/migrations` 里的正式 SQL migration；如需临时关闭，可给 `ctf-api` 设置 `CTF_AUTO_MIGRATE=false`。

这条路径默认只适合单用户、本机临时联调，不适合作为共享开发、演示或正式部署方案。原因是 `docker/ctf/docker-compose.dev.yml` 里的 `ctf-api` 会直接访问宿主 Docker daemon，用来管理靶机、checker sandbox 和运行时网络；如果 API 容器失陷，攻击者通常可以继续控制宿主 Docker 运行时。因此：

- 日常开发优先使用上面的“依赖容器 + 本机前后端”
- 需要多人长期使用时，至少把 API 改成宿主机进程运行
- 正式比赛或共享环境，推荐把 API 主机与靶机 Docker 主机拆开，由 API 通过 `runtime-agent` + mTLS 调用执行面；`DOCKER_HOST` 只能作为底层 Docker client 连接参数，不再等同于完整多机方案

更完整的威胁模型与部署建议见 `docs/architecture/backend/01-system-architecture.md` 的“7.5 安全边界设计”；运行模式和最小配置见 `docs/operations/runtime-agent-deployment.md`。

默认端口：

- `ctf-frontend`：`5173`
- `ctf-api`：`8080`
- `ctf-api` AWD 防守 SSH 网关：`2222`
- `ctf-postgres`：`15432`
- `ctf-redis`：`16379`

如果 `5173` 已被占用，可以改前端容器端口：

```bash
CTF_HOST_ROOT="$(pwd)" CTF_FRONTEND_PORT=15173 docker compose -f docker/ctf/docker-compose.dev.yml up -d --build ctf-frontend ctf-api ctf-postgres ctf-redis
```

### 局域网访问部署

`docker/ctf/docker-compose.dev.yml` 默认面向“只在宿主机本机调试”的场景：

- `CTF_CONTAINER_PUBLIC_HOST=127.0.0.1` 会让 Jeopardy / TCP 题目的访问地址返回为宿主机本地地址。
- `ctf-frontend`、`ctf-api`、`ctf-api` 的 SSH 网关端口默认都只绑定在 `127.0.0.1`，局域网内其他机器无法直接访问。

如果需要让学生从局域网内其他机器访问平台和题目实例，至少要同时调整两类配置：

1. 把 `ctf-api` 环境变量里的 `CTF_CONTAINER_PUBLIC_HOST` 改成学生真实可访问到的宿主机地址，例如固定局域网 IP 或内网域名。
2. 把 `ctf-frontend`、`ctf-api`、`ctf-api` SSH 网关的端口绑定从 `127.0.0.1:...` 改成 `0.0.0.0:...` 或指定局域网 IP。

可直接按下面的方式修改 `docker/ctf/docker-compose.dev.yml`：

```yaml
services:
  ctf-frontend:
    ports:
      - "0.0.0.0:${CTF_FRONTEND_PORT:-5173}:80"

  ctf-api:
    environment:
      CTF_CONTAINER_PUBLIC_HOST: 192.168.1.50
      CTF_CONTAINER_ACCESS_HOST: host-gateway.internal
    ports:
      - "0.0.0.0:8080:8080"
      - "0.0.0.0:2222:2222"
```

选择 `CTF_CONTAINER_PUBLIC_HOST` 时，优先使用“学生机器实际能访问到的宿主机地址”，不要机械地使用当前 shell 里看到的任意 IPv4：

- Linux 服务器通常直接使用服务器网卡的局域网 IP，例如 `192.168.x.x` 或 `10.x.x.x`。
- 如果平台跑在 Windows + WSL + Docker Desktop 上，优先使用 Windows 宿主机的局域网 IP，不要直接使用 WSL 里的 `172.*` NAT 地址。
- 如果后续会切域名访问，建议直接填内网 DNS 名称，避免学生侧地址和后端配置再次切换。

修改后建议从另一台局域网机器验证：

```bash
curl http://<宿主机地址>:8080/health
```

再登录学生账号创建 Jeopardy 实例，确认页面展示的实例地址和“打开目标”返回地址已经不再是 `127.0.0.1`。

## 可选：本地 registry

只有在需要构建和推送动态题目镜像时，才需要本地 registry。平时开发平台前后端，不需要先处理它。

```bash
scripts/registry/deploy-private-registry.sh --force-recreate
CTF_HOST_ROOT="$(pwd)" docker compose -f docker/ctf/docker-compose.dev.yml up -d --build ctf-api
```

相关编排细则见 `docs/docker-compose-rules.md`。

## 仓库内容

- `code/backend/`：后端实现。业务代码主要在 `internal/module/`，进程级装配入口在 `internal/app/composition/`
- `code/frontend/`：前端实现。页面设计和信息结构说明见 `docs/architecture/frontend/`
- `docs/architecture/`：当前架构、页面设计和专题设计入口
- `docs/contracts/`：接口、事件和题包格式等契约
- `challenges/`：题目、题包、题面、源码、writeup 和防守说明
- `concepts/`、`thinking/`、`practice/`、`feedback/`、`works/`、`references/`：项目在开发过程中沉淀的规则、实验、资料和可复用说明；项目级 prompt 资产位于 `harness/prompts/`

<!-- BEGIN HARNESS ENGINEERING: readme-harness -->
## Harness Engineering

本项目按 `deusyu/harness-engineering` 建立顶层 harness 结构：

- `concepts/`：核心概念
- `thinking/`：独立思考
- `practice/`：实践记录
- `feedback/`：反馈闭环
- `works/`：作品输出
- `harness/prompts/`：已验证的项目级 agent 工作流 prompt
- `references/`：外部资料

一致性检查：

```bash
bash scripts/check-consistency.sh
```
<!-- END HARNESS ENGINEERING: readme-harness -->
