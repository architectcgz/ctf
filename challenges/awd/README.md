# AWD 题目包

这里放的是毕设演示用 AWD 服务模板题目包。每个目录都按平台 `challenge-pack-v1` 的 AWD 扩展格式组织，包含题面、元数据、Docker 启动文件、本地 Checkbot 示例和攻防说明。

当前 `ctf-1/` 期次题目：

- `ctf-1/awd-supply-ticket/`：供应链工单系统，Flask + SQLite，核心漏洞为模板注入和默认口令。
- `ctf-1/awd-campus-drive/`：校园网盘，Flask 文件上传与预览，核心漏洞为路径穿越和上传校验绕过。
- `ctf-1/awd-iot-hub/`：IoT 设备管理平台，Flask 模拟 MQTT Topic，核心漏洞为默认设备密钥和 Topic 前缀越权。

每期比赛单独建立子目录，例如 `ctf-1/`、`ctf-2/`。只把可导入、可运行的完整题目包放入期次目录；顶层 Markdown 可保留为设计说明或候选题草案。

## 目录结构

```text
<period>/
├── <slug>/
│   ├── challenge.yml
│   ├── statement.md
│   ├── docker/
│   │   ├── docker-compose.yml
│   │   ├── app/
│   │   └── check/
│   └── writeup/
│       ├── attack.md
│       └── defense.md
└── dist/
    └── <slug>.zip
```

## 平台导入

这些包使用 `meta.mode: awd` 和 `extensions.awd`，应通过 AWD 服务模板导入入口导入：

```text
POST /api/v1/authoring/awd-service-template-imports
POST /api/v1/authoring/awd-service-template-imports/:id/commit
```

导入后，在 AWD 竞赛后台把模板添加为 `contest_awd_services`，再由队伍在学生端启动队伍共享实例。

## 镜像准备

`challenge.yml` 中的 `runtime.image.ref` 使用本地演示镜像名，例如 `ctf/awd-supply-ticket:latest`。本地可先构建：

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

## Checkbot 示例

```bash
python3 challenges/awd/ctf-1/awd-supply-ticket/docker/check/check.py http://127.0.0.1:18081
```

本地 `check/check.py` 覆盖服务可用性、主业务链路和 `/api/flag` 标准检查接口。正式比赛中的轮次检查由平台 `http_standard` checker 执行。

## AWD 题目编写要求

- `PUT /api/flag` 不能只返回成功，必须把请求体中的 Flag 写入题目服务实际读取的位置，例如文件、数据库或内存状态。
- `GET /api/flag` 必须返回最近一次由 `PUT /api/flag` 写入的 Flag；不能只返回容器启动时的 `FLAG` 环境变量默认值。
- 本地 `check/check.py` 必须覆盖 `PUT /api/flag` 后再 `GET /api/flag` 的闭环，并断言读取结果包含刚写入的 Flag。
- 平台 `http_standard` checker 会按 `put_flag -> get_flag -> havoc` 顺序执行检查；如果 `PUT` 成功但 `GET` 读不到同一个 Flag，会判定为 `flag_mismatch`，题目 readiness 不会通过。
