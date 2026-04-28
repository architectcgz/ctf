# CTF 平台项目骨架

按 `docs/architecture` 的前后端说明初始化的项目骨架：

- 后端：位于 `code/backend/`，技术栈为 Go + Gin + Viper + Zap，分层目录为 `code/backend/cmd/`、`code/backend/internal/handler/`、`code/backend/internal/service/`、`code/backend/internal/repository/`、`code/backend/internal/model/`、`code/backend/internal/middleware/`、`code/backend/pkg/`
- 前端：Vue 3 + Vite + TypeScript + Pinia + Vue Router + Element Plus
- 开发依赖：可复用根目录 `infra/` 的共享 PostgreSQL + Redis，也可用 `docker/ctf/docker-compose.dev.yml` 启动 CTF 完整容器栈（`ctf-api` + `ctf-postgres` + `ctf-redis`）

## 启动方式

后端：

```bash
cd code/backend && APP_ENV=dev go run ./cmd/api
```

默认开发账号由初始迁移写入，密码均为 `Password123`：

- `admin`：管理员账号
- `teacher`：教师账号，班级 `CTF-1`
- `student`：学员账号，班级 `CTF-1`
- `student2`：学员账号，班级 `CTF-1`

后端热重载开发（推荐）：

```bash
go install github.com/air-verse/air@latest
cd code/backend && ./scripts/dev-run.sh --infra-shared --hot
```

后端后台启动并保留 Codex 可读日志：

```bash
cd code/backend && ./scripts/dev-run.sh --infra-shared --background
tail -f /tmp/ctf-backend.log
```

脚本会在以下场景自动补齐本地开发环境变量：

- 复用 Docker 中的 PostgreSQL / Redis 时，默认切到 `127.0.0.1:15432` 与 `127.0.0.1:16379`，并注入开发默认密码
- 如果 `8080` 已被 `ctf-api` 容器占用，默认把本地 API 端口切到 `18080`
- 默认把后端输出写入 `CTF_BACKEND_LOG`，未指定时为 `/tmp/ctf-backend.log`

如果共享依赖已经在跑，也可以直接：

```bash
cd code/backend && \
  APP_ENV=dev \
  CTF_POSTGRES_PORT=15432 \
  CTF_POSTGRES_PASSWORD=postgres123456 \
  CTF_REDIS_ADDR=127.0.0.1:16379 \
  CTF_REDIS_PASSWORD=redis123456 \
  CTF_HTTP_PORT=18080 \
  air -c .air.toml
```

前端：

```bash
cd code/frontend && npm run dev
```

如果是新建 worktree，且其他 worktree 已经装过前端依赖，可以先执行：

```bash
./scripts/bootstrap-frontend-deps.sh
```

脚本会优先复用 `package-lock.json` 一致的 `code/frontend/node_modules`，找不到可复用依赖时再回退到 `npm ci --prefer-offline`。

开发容器栈（仅在需要整套容器联调时使用）：

```bash
docker compose -f docker/ctf/docker-compose.dev.yml up -d --build
```

复用共享基础设施（推荐，避免重复占用资源）：

```bash
docker compose -f docker/ctf/docker-compose.dev.yml up -d ctf-postgres ctf-redis
```

建议日常 Go 开发使用“依赖容器 + 本地热重载”：

- PostgreSQL / Redis 继续跑在 Docker 中
- Go 后端直接在宿主机执行 `air -c .air.toml`
- 只有需要验证镜像、容器网络或完整编排时再启动 `ctf-api` 容器

`docker/ctf/docker-compose.dev.yml` 默认端口如下，并且仅绑定到 `127.0.0.1`，避免开发态暴露到局域网：

- `ctf-api`: `8080`
- `ctf-postgres`: `15432`
- `ctf-redis`: `16379`

Docker 编排规范见：`docs/docker-compose-rules.md`。

强制要求摘要：

- CTF 相关容器统一放在 `docker/ctf/` 下
- CTF 相关容器统一由一个 Compose 项目管理（建议 `name: ctf`）
- CTF 内部统一使用 `ctf-network`
- 禁止 CTF 容器混用 Compose 与手工 `docker run`

## 当前骨架范围

- 已提供统一响应结构、请求 ID 中间件、访问日志、中断优雅退出、健康检查接口
- 已在 `backend/internal/` 下预留 `auth`、`challenge`、`practice`、`contest`、`assessment`、`system`、`container` 分层目录
- 前端现有路由、状态管理、API 封装与布局结构保留，并补充项目级说明文件
