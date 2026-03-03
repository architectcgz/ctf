# CTF 平台项目骨架

按 `docs/architecture` 的前后端说明初始化的项目骨架：

- 后端：位于 `code/backend/`，技术栈为 Go + Gin + Viper + Zap，分层目录为 `code/backend/cmd/`、`code/backend/internal/handler/`、`code/backend/internal/service/`、`code/backend/internal/repository/`、`code/backend/internal/model/`、`code/backend/internal/middleware/`、`code/backend/pkg/`
- 前端：Vue 3 + Vite + TypeScript + Pinia + Vue Router + Element Plus
- 开发依赖：可复用根目录 `infra/` 的共享 PostgreSQL + Redis，也可用 `code/backend/docker-compose.dev.yml` 单独启动一套

## 启动方式

后端：

```bash
cd code/backend && make run
```

前端：

```bash
cd code/frontend && npm run dev
```

开发基础设施：

```bash
cd code/backend && make infra-up
```

复用共享基础设施（推荐，避免重复占用资源）：

```bash
cd code/backend && make infra-up-shared
```

`code/backend/docker-compose.dev.yml` 的 PostgreSQL/Redis 端口仅绑定到 `127.0.0.1`，避免开发态暴露到局域网。

## 当前骨架范围

- 已提供统一响应结构、请求 ID 中间件、访问日志、中断优雅退出、健康检查接口
- 已在 `backend/internal/` 下预留 `auth`、`challenge`、`practice`、`contest`、`assessment`、`system`、`container` 分层目录
- 前端现有路由、状态管理、API 封装与布局结构保留，并补充项目级说明文件
