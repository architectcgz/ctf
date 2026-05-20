# CTF Infra Registry Reconciliation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 `ctf` 项目的 PostgreSQL、Redis、Registry 相关 infra 入口收口到统一目录，重部署 `ctf-registry`，并把当前需要的容器题镜像重新推送与同步回数据库。

**Architecture:** 保留 `docker/ctf/docker-compose.dev.yml` 作为单一 Compose 入口，不拆成多份 Compose。通过在 `docker/ctf/infra/` 下集中承载 registry env、registry runtime 目录说明和 infra 文档，解决配置散落与事实源重复的问题；题目镜像回灌继续复用现有 registry build 脚本，并补一个数据库同步脚本来统一更新 `images`、`challenges`、`awd_challenges`、`contest_awd_services` 引用。

**Tech Stack:** Docker Compose、Bash、Python 3、PostgreSQL、Go 文档/配置更新。

---

### Task 1: 统一 `docker/ctf/infra` 目录与 registry 配置事实源

**Files:**
- Create: `docker/ctf/infra/README.md`
- Create: `docker/ctf/infra/registry/ctf-platform-registry.env.example`
- Modify: `docker/ctf/docker-compose.dev.yml`
- Modify: `scripts/registry/deploy-private-registry.sh`
- Modify: `scripts/registry/deploy-private-registry.conf.example`
- Modify: `scripts/registry/build-and-push-challenge-image.sh`
- Modify: `.gitignore`
- Modify: `README.md`
- Modify: `docs/architecture/backend/03-container-architecture.md`

- [ ] **Step 1: 确认新目录与默认路径**

要求：
- `ctf-api` 只从 `docker/ctf/infra/registry/ctf-platform-registry.env` 读取 registry 配置。
- registry 默认数据/认证目录落到 `docker/ctf/infra/registry/runtime/` 下。
- 旧的 `$HOME/ctf-registry` 与 `docker/ctf/.env.registry` 只作为迁移来源，不再作为默认事实源。

- [ ] **Step 2: 更新 compose 和脚本默认值**

要求：
- `docker-compose.dev.yml` 的 `env_file` 改到新路径。
- registry compose volume 默认值改到新 infra 目录。
- `deploy-private-registry.sh` 能在不显式传密码时复用旧 env 中已有凭据。
- 若旧 registry 数据目录存在且新目录为空，脚本自动迁移已有认证/镜像数据。

- [ ] **Step 3: 统一题目镜像构建脚本入口**

要求：
- `build-and-push-challenge-image.sh` 默认读取新的 registry env 文件。
- 同时支持：
  - Jeopardy: `docker/Dockerfile`
  - AWD: `docker/runtime/Dockerfile`，build context 为 `docker/`

- [ ] **Step 4: 补最小必要文档**

Run:
```bash
bash scripts/registry/deploy-private-registry.sh --help
bash scripts/registry/build-and-push-challenge-image.sh --help
docker compose --profile registry -f docker/ctf/docker-compose.dev.yml config
git diff --check
```

Expected:
- 脚本帮助与 compose config 都能通过。
- 文档明确 `docker/ctf/infra` 是新的 infra 入口。

### Task 2: 补题目镜像回灌与数据库同步脚本

**Files:**
- Create: `scripts/registry/sync-challenge-registry-state.py`
- Optional Modify: `scripts/registry/build-and-push-challenge-image.sh`
- Modify: `README.md`
- Modify: `docs/architecture/features/题包Registry交付架构.md`

- [ ] **Step 1: 定义回灌范围**

范围：
- AWD 容器题：`challenges/awd/ctf-1/*`
- Jeopardy 容器题：`challenges/jeopardy/packs/*` 中存在 Dockerfile 的 5 题
- `runtime.type=none` 和 `templates/` 不进入 registry 回灌

- [ ] **Step 2: 实现数据库同步脚本**

要求：
- 输入为题包目录与最终 `IMAGE_REF` / digest。
- `images` 表按 `(name, tag)` upsert，保留已有 `source_type` 优先，不强制改写历史来源。
- Jeopardy 同步 `challenges.image_id`。
- AWD 同步 `awd_challenges.runtime_config.image_id/image_ref`，并级联同步 `contest_awd_services.runtime_config.challenge_runtime.image_id/image_ref`。
- 不删除 `legacy-awd-*` 与 `e2e-*` 历史记录。

- [ ] **Step 3: 准备数据库备份命令**

Run:
```bash
mkdir -p backups
docker exec ctf-postgres pg_dump -U postgres -d ctf > backups/ctf-registry-sync-20260508.sql
```

Expected:
- 导出成功，可用于回退数据库内容。

### Task 3: 运行侧重部署与回灌验证

**Files:**
- Runtime only: Docker containers / registry state / PostgreSQL records

- [ ] **Step 1: 串行重部署 registry**

Run:
```bash
bash scripts/registry/deploy-private-registry.sh --force-recreate
docker ps --format '{{.Names}}\t{{.Label "com.docker.compose.project"}}\t{{.Label "com.docker.compose.service"}}' | rg '^ctf-registry'
```

Expected:
- `ctf-registry` 归到 `ctf / ctf-registry`。

- [ ] **Step 2: 批量 build + push 9 个容器题**

Run:
```bash
# AWD 4 题 + Jeopardy 5 题，逐个执行并记录 IMAGE_REF
```

Expected:
- 每个题包都能得到目标 `IMAGE_REF=<registry>/<mode>/<slug>:<tag>`。

- [ ] **Step 3: 批量同步数据库**

Run:
```bash
python3 scripts/registry/sync-challenge-registry-state.py ...
```

Expected:
- `images`、`challenges`、`awd_challenges`、`contest_awd_services` 都切到新的 registry 引用。

- [ ] **Step 4: 做回归验证**

Run:
```bash
docker exec ctf-postgres psql -U postgres -d ctf -P pager=off -c "SELECT id,name,tag,status,digest FROM public.images ORDER BY id;"
docker exec ctf-postgres psql -U postgres -d ctf -P pager=off -c "SELECT id,package_slug,image_id FROM public.challenges ORDER BY id;"
docker exec ctf-postgres psql -U postgres -d ctf -P pager=off -c "SELECT id,slug,runtime_config FROM public.awd_challenges ORDER BY id;"
```

Expected:
- Jeopardy 容器题全部有 `image_id`。
- AWD 当前正式题与历史题都引用可访问的 `127.0.0.1:5000/awd/...` 镜像。

### Task 4: 实现后 review 与收尾

**Files:**
- Create: `docs/reviews/general/2026-05-08-ctf-infra-registry-reconciliation-review.md`

- [ ] **Step 1: 切换到独立 review 视角**

Review focus:
- registry env 是否只保留一个事实源。
- infra 目录是否足够清晰，且没有把运行态敏感文件误提交。
- 数据库同步是否遗漏 `contest_awd_services` 这类运行时快照引用。
- 是否保留了 `legacy-awd-*` 的现有比赛价值。

- [ ] **Step 2: 修复 findings 后重跑受影响验证**

Run:
```bash
git diff --check
```

Expected:
- review finding 收口后工作树干净，可继续提交。
