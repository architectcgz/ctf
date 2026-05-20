# Docker Project Grouping Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 `ctf` 相关容器在 Docker 中按统一项目归属展示，AWD 题目实例归到 `ctf/awd`，普通题目实例归到 `ctf/jeopardy`，私有 registry 不再脱离 `ctf` Compose 项目单独漂在外面。

**Architecture:** 保留现有 runtime engine 直接调用 Docker API 的创建链路，不引入“实例改为 Compose 编排”的结构性重写；对动态题目容器补充 Compose 风格项目/服务标签，对 registry 部署脚本改为调用 `docker compose` 管理同名 service。这样可以最小改动收口 Docker 展示归属，同时不动实例生命周期、ACL、端口和清理逻辑。

**Tech Stack:** Go、Docker Engine API、Docker Compose、Bash、Go test。

---

### Task 1: 明确动态容器的分组判定与标签落点

**Files:**
- Modify: `code/backend/internal/module/runtime/domain/managed_resource_labels.go`
- Modify: `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- Modify: `code/backend/internal/module/contest/infrastructure/docker_checker_runner.go`
- Test: `code/backend/internal/module/runtime/service_test.go`
- Test: `code/backend/internal/module/contest/infrastructure/docker_checker_runner_test.go`

- [ ] **Step 1: 写出分组规则**

记录并实现以下判定：
- AWD：共享比赛网络、AWD 工作区伴生容器、AWD 镜像/别名/命名特征。
- Jeopardy：其余题目实例与自检试跑容器默认归为 `jeopardy`。
- Registry：单独作为 `ctf-registry` service，由 Compose 真正管理。

- [ ] **Step 2: 在 runtime 标签构建里补 Compose 风格项目/服务标签**

要求：
- 保留现有 `ctf.project`、`managed-by`、`ctf-component`。
- 新增 Docker Desktop/Compose 可识别的项目标签，至少覆盖 `project=ctf` 与 `service=awd|jeopardy`。
- 不改现有容器名、网络名、清理过滤条件。

- [ ] **Step 3: 给 AWD checker sandbox 补同类标签**

要求：
- checker 容器继续保留现有业务标签。
- 追加 `ctf/awd` 分组所需标签，避免 AWD 运行时里只有主容器归组、checker 继续散落。

- [ ] **Step 4: 补充单元测试**

Run:
```bash
go test ./internal/module/runtime ./internal/module/contest/infrastructure -run 'CreateContainer|CreateTopology|DockerChecker' -count=1
```

Expected:
- 现有 runtime 创建测试继续通过。
- 新增断言能验证 `awd` / `jeopardy` / checker 的标签结果。

### Task 2: 把私有 registry 收口到 ctf Compose 项目

**Files:**
- Modify: `docker/ctf/docker-compose.dev.yml`
- Modify: `scripts/registry/deploy-private-registry.sh`
- Modify: `scripts/registry/deploy-private-registry.conf.example`
- Modify: `.gitignore` if generated compose env file needs ignore
- Test: `scripts/registry/deploy-private-registry.sh`

- [ ] **Step 1: 在 `docker/ctf` 下定义 registry service**

要求：
- service 归属于 `name: ctf`。
- 仍支持自定义 `name/port/data-dir/auth-dir/username/password`。
- 不影响默认 `ctf-postgres` / `ctf-redis` / `ctf-api` 的启动路径。

- [ ] **Step 2: 改 registry 部署脚本为 Compose 管理**

要求：
- 继续生成 htpasswd 与平台 env 文件。
- 不再用手工 `docker run` / `docker rm -f` 创建主 registry 容器。
- `--force-recreate` 语义保留。

- [ ] **Step 3: 用最小命令验证脚本产物**

Run:
```bash
bash scripts/registry/deploy-private-registry.sh --help
docker compose -f docker/ctf/docker-compose.dev.yml config
```

Expected:
- 脚本帮助正常输出。
- Compose 配置可解析，registry service 合法。

### Task 3: 同步最小必要文档与回归验证

**Files:**
- Modify: `README.md`
- Modify: `docs/architecture/backend/03-container-architecture.md`

- [ ] **Step 1: 更新本地运行说明**

说明 registry 现在由 `ctf` Compose 项目管理，动态题目实例通过 `awd/jeopardy` 分组标签归类，不再只靠裸 `docker run`。

- [ ] **Step 2: 跑受影响验证**

Run:
```bash
go test ./internal/module/runtime ./internal/module/practice/application/commands ./internal/module/contest/infrastructure -run 'CreateContainer|CreateTopology|AWD|DockerChecker' -count=1
bash scripts/registry/deploy-private-registry.sh --help
docker compose -f docker/ctf/docker-compose.dev.yml config >/tmp/ctf-docker-compose-config.out
```

Expected:
- Go 侧标签/命名相关测试通过。
- registry 脚本帮助与 compose config 通过。

- [ ] **Step 3: 做实现后 review**

Review focus:
- 是否只改了 Docker 展示归属，而没有破坏实例生命周期。
- 是否保留现有清理过滤条件与容器命名。
- registry 是否彻底摆脱手工 `docker run` 主路径。
