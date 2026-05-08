# CTF 平台项目

本仓库主要包含平台实现、架构与契约文档、题目与题包，以及开发过程中沉淀的规则和资料。

- 后端：位于 `code/backend/`，技术栈为 Go + Gin + Viper + Zap；当前按 `auth`、`identity`、`challenge`、`runtime`、`practice`、`contest`、`assessment`、`ops`、`practice_readmodel`、`teaching_readmodel` 等模块组织
- 前端：Vue 3 + Vite + TypeScript + Pinia + Vue Router + Element Plus
- 开发依赖：可复用根目录 `infra/` 的共享 PostgreSQL + Redis，也可用 `docker/ctf/docker-compose.dev.yml` 启动 CTF 完整容器栈（`ctf-api` + `ctf-postgres` + `ctf-redis`）
- 文档入口：架构和页面设计主要看 `docs/architecture/`，接口与题包契约主要看 `docs/contracts/`

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

开发容器栈（仅在需要整套容器联调时使用）：

```bash
CTF_HOST_ROOT="$(pwd)" docker compose -f docker/ctf/docker-compose.dev.yml up -d --build
```

复用共享基础设施（推荐，避免重复占用资源）：

```bash
CTF_HOST_ROOT="$(pwd)" docker compose -f docker/ctf/docker-compose.dev.yml up -d ctf-postgres ctf-redis
```

建议日常 Go 开发使用“依赖容器 + 本地热重载”：

- PostgreSQL / Redis 继续跑在 Docker 中
- Go 后端直接在宿主机执行 `air -c .air.toml`
- 只有需要验证镜像、容器网络或完整编排时再启动 `ctf-api` 容器

`docker/ctf/docker-compose.dev.yml` 默认端口如下，并且仅绑定到 `127.0.0.1`，避免开发态暴露到局域网：

- `ctf-api`: `8080`
- `ctf-postgres`: `15432`
- `ctf-redis`: `16379`

题包镜像构建需要先配置私有 registry。推荐通过脚本部署 registry 并生成 `docker/ctf/.env.registry`：

```bash
scripts/registry/deploy-private-registry.sh --force-recreate
CTF_HOST_ROOT="$(pwd)" docker compose -f docker/ctf/docker-compose.dev.yml up -d --build ctf-api
```

`docker/ctf/.env.registry` 只会被 `ctf-api` 加载，用于平台构建、推送、manifest 校验、pull/inspect 和镜像注册；不要把这些凭据写进题包或题目容器。

Docker 编排规范见：`docs/docker-compose-rules.md`。

强制要求摘要：

- CTF 相关容器统一放在 `docker/ctf/` 下
- CTF 相关容器统一由一个 Compose 项目管理（建议 `name: ctf`）
- CTF 内部统一使用 `ctf-network`
- 禁止 CTF 容器混用 Compose 与手工 `docker run`

## 仓库内容

- `code/backend/`：后端实现。业务代码主要在 `internal/module/`，进程级装配入口在 `internal/app/composition/`
- `code/frontend/`：前端实现。页面设计和信息结构说明见 `docs/architecture/frontend/`
- `docs/architecture/`：当前架构、页面设计和专题设计入口
- `docs/contracts/`：接口、事件和题包格式等契约
- `challenges/`：题目、题包、题面、源码、writeup 和防守说明
- `concepts/`、`thinking/`、`practice/`、`feedback/`、`works/`、`prompts/`、`references/`：项目在开发过程中沉淀的规则、实验、资料和可复用说明

<!-- BEGIN HARNESS ENGINEERING: readme-harness -->
## Harness Engineering

本项目按 `deusyu/harness-engineering` 建立顶层 harness 结构：

- `concepts/`：核心概念
- `thinking/`：独立思考
- `practice/`：实践记录
- `feedback/`：反馈闭环
- `works/`：作品输出
- `prompts/`：提示词积累
- `references/`：外部资料

一致性检查：

```bash
bash scripts/check-consistency.sh
```
<!-- END HARNESS ENGINEERING: readme-harness -->
