# Registry Access Endpoint Implementation Plan

## Objective

修复本地 dev 下题包镜像链路里 `build/push` 与 `verify` 处在不同网络视角导致的 registry 校验失败问题，同时保持现有镜像 ref、Docker daemon 拉取语义和运行时消费语义不变。

目标收口为：

- `container.registry.server` 继续表示镜像 ref 与 Docker daemon 使用的 canonical registry server
- 新增一个仅供后端进程直连 registry API 的可选访问地址
- 本地 `deploy-private-registry.sh` 能自动写出这组配置，避免再靠容器内临时端口转发

## Non-goals

- 不把现有 `images.name/tag`、`image_build_jobs.target_ref` 或 AWD `runtime_config.image_ref` 批量改写成 `ctf-registry:*`
- 不要求宿主 Docker daemon 额外配置 `insecure-registries`
- 不改题包导入、镜像构建、实例运行的主状态机
- 不引入新的 registry service、队列或跨模块 contract

## Inputs

- `code/backend/internal/config/config.go`
- `code/backend/internal/config/config_test.go`
- `code/backend/internal/module/challenge/infrastructure/registry_client.go`
- `code/backend/internal/module/challenge/infrastructure/registry_client_test.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/configs/config.yaml`
- `docker/ctf/infra/registry/ctf-platform-registry.env`
- `docker/ctf/infra/registry/ctf-platform-registry.env.example`
- `scripts/registry/deploy-private-registry.sh`
- `scripts/registry/deploy-private-registry.conf.example`
- `docs/architecture/features/题包Registry交付架构.md`
- `docs/architecture/backend/03-container-architecture.md`
- `docker/ctf/infra/README.md`
- `docker/ctf/infra/registry/README.md`

## Current Baseline

- 平台当前只配置一个 `container.registry.server`
- 该字段同时被：
  - 平台镜像 ref 生成使用
  - Docker auth/build/push/pull 使用
  - registry manifest verifier 的 HTTP URL 使用
- 本地 dev 默认值是 `127.0.0.1:5000`
- `docker build/push/pull` 通过 `ctf-api` 挂入的 `/var/run/docker.sock` 实际委托宿主 Docker daemon 执行
- registry manifest `HEAD /v2/.../manifests/...` 则由 `ctf-api` 容器内 Go 进程自己发出
- 结果是：`127.0.0.1:5000` 对 daemon 可达，但对容器内 Go 进程不可达

## Brainstorming Summary

候选方案：

1. 把 canonical registry server 改成 `ctf-registry:5000`
   - 拒绝：容器内 verifier 能通，但宿主 Docker daemon 通常无法把 compose service 名当作 registry 地址使用，运行时拉镜像也会受影响
2. 把 canonical registry server 改成非 localhost 的宿主地址
   - 拒绝：当前 dev registry 是 HTTP，本机会立刻牵涉 Docker daemon 的 `insecure-registries` 或 TLS，超出本次最小修复
3. 保留 canonical server，新增 verifier access endpoint
   - 采用：最小改动即可消除网络视角混用，同时不改现有 DB 和镜像 ref

## Chosen Direction

新增 `container.registry.access_server`：

- `server`
  - 负责：镜像 ref 前缀、Docker daemon registry 域名匹配、auth 目标 server
  - 不负责：容器内 Go 进程直连 registry API 的网络可达性
- `access_server`
  - 负责：registry verifier 在当前进程网络视角下访问 registry API 的 host:port
  - 不负责：生成镜像 ref、改写 DB、影响 Docker auth server 匹配

运行规则：

- 如果未配置 `access_server`，verifier 继续回退到 `server`
- verifier 仍然用 `server` 判断 image ref 是否属于当前 registry
- 只有实际发 HTTP 请求时才改为连 `access_server`

## Ownership Boundary

- `internal/config`
  - 负责：暴露 `access_server` 配置并提供默认值
  - 不负责：决定 HTTP verifier 如何拼接请求
- `challenge/infrastructure/registry_client.go`
  - 负责：在校验 image ref 归属时使用 canonical `server`，在直连 registry API 时使用 `access_server`
  - 不负责：改写 image ref 或 Docker auth 逻辑
- `challenge/runtime/module.go`
  - 负责：把 registry config 正确装配给 verifier
  - 不负责：生成 dev env 文件
- `scripts/registry/deploy-private-registry.sh`
  - 负责：在本地 `127.0.0.1:*` 场景自动生成 `access_server=ctf-registry:5000`
  - 不负责：替宿主 Docker daemon 配置 insecure registry

## Change Surface

- Add: `.harness/reuse-decisions/registry-access-endpoint.md`
- Add: `docs/plan/impl-plan/2026-05-14-registry-access-endpoint-implementation-plan.md`
- Modify: `code/backend/internal/config/config.go`
- Modify: `code/backend/internal/config/config_test.go`
- Modify: `code/backend/internal/module/challenge/infrastructure/registry_client.go`
- Modify: `code/backend/internal/module/challenge/infrastructure/registry_client_test.go`
- Modify: `code/backend/internal/module/challenge/runtime/module.go`
- Modify: `code/backend/configs/config.yaml`
- Modify: `docker/ctf/infra/registry/ctf-platform-registry.env`
- Modify: `docker/ctf/infra/registry/ctf-platform-registry.env.example`
- Modify: `scripts/registry/deploy-private-registry.sh`
- Modify: `scripts/registry/deploy-private-registry.conf.example`
- Modify: `docs/architecture/features/题包Registry交付架构.md`
- Modify: `docs/architecture/backend/03-container-architecture.md`
- Modify: `docker/ctf/infra/README.md`
- Modify: `docker/ctf/infra/registry/README.md`

## Task Slices

- [x] Slice 1: 后端配置与 verifier 语义收口
  - 目标：新增 `access_server`，不破坏现有 `server` 语义
  - 验证：
    - `cd code/backend && go test ./internal/config -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/challenge/infrastructure -run RegistryClient -count=1 -timeout 5m`
  - Review focus：ref ownership 判断仍是否绑定 canonical `server`；只有 HTTP 访问 host 被 override

- [x] Slice 2: 本地 deploy/env 默认值收口
  - 目标：本地 registry 脚本和 env 示例能表达 `server + access_server`
  - 验证：
    - `bash scripts/registry/deploy-private-registry.sh --help`
    - 检查 env 输出模板
  - Review focus：localhost 场景是否自动写出可用的 `access_server`；非 localhost 场景是否不会强行写 override

- [x] Slice 3: 文档同步
  - 目标：架构事实、容器运行说明和 infra README 同步更新
  - 验证：
    - `python3 scripts/check-docs-consistency.py`
    - `bash scripts/check-consistency.sh`
  - Review focus：当前事实是否明确区分 canonical registry server 与 verifier access endpoint

## Risks

- 如果 verifier 用错 `server` / `access_server` 的职责，可能出现“HTTP 可达但 image ref 归属判断失真”
- 如果 deploy 脚本对非 localhost 场景也自动写 `access_server=ctf-registry:5000`，会把远端部署写坏
- 如果文档没同步，后续仍容易有人直接把 DB ref 改成 compose service 名

## Verification Plan

1. `cd code/backend && go test ./internal/config -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/challenge/infrastructure -run RegistryClient -count=1 -timeout 5m`
3. `python3 scripts/check-docs-consistency.py`
4. `bash scripts/check-consistency.sh`
5. `bash scripts/registry/deploy-private-registry.sh --help`
6. 重新构建并启动 `ctf-api`，确认本轮没有再依赖 `socat`
7. 在 `ctf-api` 容器内执行：
   - `wget -S -O- http://ctf-registry:5000/v2/` 预期 401
   - `docker pull 127.0.0.1:5000/...` 预期仍可通过宿主 daemon 拉取

## Architecture-Fit Evaluation

- owner 明确：镜像身份仍由 canonical `server` 表达，网络可达性问题只落在 verifier adapter
- reuse point 明确：复用现有 `container.registry`、`RegistryVerifier` 和 deploy-private-registry 脚本，不新建并行配置体系
- 这刀同时解决行为与结构：既消掉本地 verify 网络视角错误，又避免把 compose service 名写进 DB / image ref
