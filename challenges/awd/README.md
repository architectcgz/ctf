# AWD 题目包

这里放的是毕设演示用 AWD 题目包。每个目录都按平台 `challenge-pack-v1` 的 AWD 扩展格式组织，包含题面、元数据、Docker 启动文件、本地 Checkbot 示例和攻防说明。

题目包正式契约见：

- [AWD 题目包契约](./challenge-package-contract.md)

AWD 题目包必须在 `challenge.yml` 的 `extensions.awd.checker` 中声明平台正式 checker。管理员导入题目包后，赛事服务默认继承这个 checker；后台编辑只用于赛事级覆盖、试跑和保存校验结果，不作为题目包 checker 的替代来源。

当前 `ctf-1/` 期次题目：

- `ctf-1/awd-supply-ticket/`：供应链工单系统，Flask + SQLite，核心漏洞为模板注入和默认口令。
- `ctf-1/awd-campus-drive/`：校园网盘，Flask 文件上传与预览，核心漏洞为路径穿越和上传校验绕过。
- `ctf-1/awd-iot-hub/`：IoT 设备管理平台，Flask 模拟 MQTT Topic，核心漏洞为默认设备密钥和 Topic 前缀越权。

当前 `ctf-2/` 期次题目：

- `ctf-2/awd-webhook-inspector/`：Webhook 文档预览器，核心漏洞为 SSRF 与粗糙本地地址黑名单。
- `ctf-2/awd-passkey-sync-gateway/`：Passkey 同步网关，核心漏洞为调试导出命令里的固定支持密钥。

当前 `ctf-3/` 期次题目：

- `ctf-3/awd-webhook-relay-hub/`：三容器 webhook relay，核心漏洞为公网预览器可打内部 replay，再横向取 archive 中的 Flag。
- `ctf-3/awd-forwarded-admin-gateway/`：公网 gateway + 内部 admin + audit，核心漏洞为伪造 `X-Forwarded-*` 头穿透内部导出链路。
- `ctf-3/awd-preview-render-farm/`：预览站点 + renderer + asset cache，核心漏洞为素材路径直拼导致命中内部 debug asset。
- `ctf-3/awd-iot-fleet-orchestrator/`：公网控制台 + fleet agent + config vault，核心漏洞为默认支持 key 可调用内部高权限设备操作。
- `ctf-3/awd-patch-signing-gateway/`：公网 patch gateway + signer + key vault，核心漏洞为客户端可伪造内部 signer 角色头获取高权限 bundle。

每期比赛单独建立子目录，例如 `ctf-1/`、`ctf-2/`。只把可导入、可运行的完整题目包放入期次目录；顶层 Markdown 可保留为设计说明或候选题草案。

## 目录结构

```text
<period>/
├── <slug>/
│   ├── challenge.yml
│   ├── statement.md
│   ├── docker/
│   │   ├── docker-compose.yml
│   │   ├── runtime/
│   │   ├── workspace/
│   │   └── check/
│   └── writeup/
│       ├── attack.md
│       └── defense.md
└── dist/
    └── <slug>.zip
```

补充约定：

- `docker/` 是题目本地调试入口，不是平台运行归属依据
- Web 题固定入口为 `docker/runtime/app.py`，学生主要审计和修补的业务代码位于 `docker/workspace/src/`
- TCP 题固定入口为 `docker/runtime/app.py`，学生主要审计和修补的业务逻辑位于 `docker/workspace/src/`
- `ctf_runtime.py` 承载 `/health`、`/api/flag`、checker token、动态 Flag 写入读取等平台运行契约，默认放入 `protected_paths`
- `challenge.yml` 的 `extensions.awd.runtime_config.defense_scope` 必须声明 `editable_paths`、`protected_paths`、`service_contracts`
- 平台是否把实例识别为 `ctf` 项目下的受管容器，只看运行时 label，不看题目放在哪个目录、老师从哪个目录执行 compose
- `runtime` 镜像必须内置 `/workspace/` 种子内容，否则平台首次挂载 named volume 时会得到空工作区

## 平台导入

这些包使用 `meta.mode: awd` 和 `extensions.awd`，应通过 AWD 题目导入入口导入：

```text
POST /api/v1/authoring/awd-challenge-imports
POST /api/v1/authoring/awd-challenge-imports/:id/commit
```

导入后，在 AWD 竞赛后台把 AWD 题目添加为 `contest_awd_services`，再由队伍在学生端启动队伍共享实例。

上传预览阶段会校验 `defense_scope` 的结构、路径存在性、可编辑/受保护路径是否交叉，以及固定入口是否被错误放入可编辑范围；不符合约定的包不会进入待确认队列。

添加到赛事时，管理端会读取题目包导入得到的 `checker_type / checker_config` 作为默认草稿。只有当前赛事需要临时调整裁判规则时，才在赛事配置里覆盖 checker；覆盖结果只写入该赛事的 `contest_awd_services.runtime_config`，不会反向修改题目包或 AWD 题库模板。

## 镜像准备

`challenge.yml` 中的 `runtime.image.ref` 使用本地演示镜像名，例如 `ctf/awd/awd-supply-ticket:latest`。本地可先构建：

```bash
cd challenges/awd/ctf-1/awd-supply-ticket/docker
docker compose build
```

如需推送到私有仓库，把 `runtime.image.ref` 改成实际镜像地址，并同步修改 `docker-compose.yml` 的 `image`。

## 本地启动

以 `awd-supply-ticket` 为例：

```bash
cd challenges/awd/ctf-1/awd-supply-ticket/docker
docker compose up --build
```

启动后访问 `http://127.0.0.1:18081/`。其他题目端口见各自 `docker/docker-compose.yml`。

对 `ctf-3/` 这类多容器 topology 题，本地验题不要只用 `up -d`。推荐直接等待服务健康：

```bash
cd challenges/awd/ctf-3/awd-webhook-relay-hub/docker
docker compose up -d --build --wait --wait-timeout 60
python3 check/check.py http://127.0.0.1:18131
docker compose down -v
```

原因是 `up -d` 只能保证容器进入 `running`，不能保证内部依赖已经 ready；如果 checker 覆盖真实攻击链，公网入口过早可用时会打到尚未就绪的内部节点。

## Checkbot 示例

```bash
python3 challenges/awd/ctf-1/awd-supply-ticket/docker/check/check.py http://127.0.0.1:18081
```

本地 `check/check.py` 覆盖服务可用性、主业务链路和 `/api/flag` 标准检查接口。正式比赛中的轮次检查由平台 `http_standard` checker 执行。

`check/check.py` 是出题人本地验证与审计材料，不会作为学生附件公开，也不会被当前平台当作任意脚本 checker 调度。若以后接入脚本型 checker，应作为管理员私有导入内容部署到 checker runner，而不是进入选手可下载资源。

## 分制规范

AWD 题目包只声明题目建议分值，实际每轮计分由赛事服务编排写入 `contest_awd_services.score_config` 和 `awd_rounds`。

- Drill：建议 12-24 轮，总分量级 300-800。
- 正式赛：建议 24-48 轮，总分量级 1000-3000。
- 长时赛：建议总分量级 3000-8000，需降低轮频或服务分，避免纯 SLA 堆分。
- 服务默认每轮 `SLA = 1`、`防御 = 2`；单项上限均为 5。
- 轮次默认 `攻击 = 30`、`防御兜底 = 3`；攻击分上限 100，轮次防御兜底上限 10。
- `meta.points` 不参与 AWD 每轮累计，只作为导入预览和服务展示建议值。

## AWD 题目编写要求

- `challenge.yml` 必须提供 `extensions.awd.checker.type` 和 `extensions.awd.checker.config`；`http_standard` 至少应声明 `put_flag` 与 `get_flag`，建议同时声明 `havoc`。
- `PUT /api/flag` 不能只返回成功，必须把请求体中的 Flag 写入题目服务实际读取的位置，例如文件、数据库或内存状态。
- `GET /api/flag` 必须返回最近一次由 `PUT /api/flag` 写入的 Flag；不能只返回容器启动时的 `FLAG` 环境变量默认值。
- 本地 `check/check.py` 必须覆盖 `PUT /api/flag` 后再 `GET /api/flag` 的闭环，并断言读取结果包含刚写入的 Flag。
- 平台 `http_standard` checker 会按 `put_flag -> get_flag -> havoc` 顺序执行检查；如果 `PUT` 成功但 `GET` 读不到同一个 Flag，会判定为 `flag_mismatch`，题目 readiness 不会通过。
