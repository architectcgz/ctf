# CTF 平台项目骨架

按 `docs/architecture` 的前后端说明初始化的项目骨架：

- 后端：Go + Gin + Viper + Zap，分层目录为 `cmd/`、`internal/handler/`、`internal/service/`、`internal/repository/`、`internal/model/`、`internal/middleware/`、`pkg/`
- 前端：Vue 3 + Vite + TypeScript + Pinia + Vue Router + Element Plus
- 开发依赖：`docker-compose.dev.yml` 提供 PostgreSQL 和 Redis

## 启动方式

后端：

```bash
make run
```

前端：

```bash
make frontend-dev
```

开发基础设施：

```bash
make infra-up
```

`docker-compose.dev.yml` 的 PostgreSQL/Redis 端口仅绑定到 `127.0.0.1`，避免开发态暴露到局域网。

## 当前骨架范围

- 已提供统一响应结构、请求 ID 中间件、访问日志、中断优雅退出、健康检查接口
- 已预留 `auth`、`challenge`、`practice`、`contest`、`assessment`、`system`、`container` 分层目录
- 前端现有路由、状态管理、API 封装与布局结构保留，并补充项目级说明文件
