# Challenge Pack Specification v1 (`challenge-pack-v1`)

> 适用范围：CTF 靶场平台题目制作、审计、归档、导入与发布前校验
> 文档目的：定义“题目源包（Zip）”的结构与 `challenge.yml` 规范，并明确当前平台如何导入、校验与发布题目。
> 版本：v1.4 | 日期：2026-04-21

> 说明：截至 2026-04-21，仓库内已有题目、镜像、Hint、拓扑、Writeup 与 AWD 服务模板等后端能力，且已新增后台 `challenge.yml` 题目包上传预览、确认导入、同步自检与发布自检队列接口，以及 AWD 服务模板包预览/确认导入接口与复用同一解析核心的 CLI 导入脚本。
> 因此本规范既是“作者侧源包规范”，也是当前实现中的导入契约；文中会显式标注哪些扩展字段仍处于保留状态。

---

## 1. 结论摘要

现有 v1.0 规范不够合理，主要问题有三类：

- 把“当前平台已支持的字段”和“未来可能支持的导入能力”混写成同一层要求，容易让出题人误以为平台今天已经支持 Zip 导入、在线构建、tar 导入、manual flag 等能力。
- 把 `docker/Dockerfile` 写成所有题目包的必填项，和当前平台“题目通过已注册镜像运行”的现实不一致，也不适用于纯镜像交付或题目源码暂不开放的场景。
- 字段枚举与当前后端不完全一致，例如难度实际支持 `insane`，而不是 `hell`；Flag 实际只支持 `static` / `dynamic`，不支持 `manual`。

这版 v1.1 做了两件事：

- 收敛为“当前平台可执行”的最小核心字段。
- 保留拓扑、Writeup、源码构建材料等扩展能力，但不再把它们写成 v1 必填或当前平台必然生效。

---

## 2. 术语

- Challenge：平台中的题目实体。
- Challenge Pack：题目源包，通常为一个 Zip 文件，用于题目制作、归档、审计和未来导入。
- Manifest：题目源包根目录下的 `challenge.yml`。
- 当前平台支持：指当前仓库中的后端模型/API 能直接承接该字段。
- 保留扩展：指字段设计合理，但当前平台没有公开导入器或没有端到端自动落库路径。

---

## 3. 设计原则

- 可复现：题目环境、题面、提示、附件来源应可追溯。
- 不泄露：题目包不得包含平台全局 secret、数据库口令、生产令牌等敏感信息。
- 可映射：v1 核心字段必须能明确映射到当前平台数据模型。
- 可降级：扩展字段即使暂不自动导入，也不应阻塞题目源包归档和审计。
- 最小惊讶：规范不得承诺平台当前并不存在的行为。

---

## 4. 当前平台能力边界

截至 2026-04-25，当前仓库内与题目包最相关的后端事实如下：

- 题目基础字段已支持：`title`、`description`、`category`、`difficulty`、`points`、`image_id`、`attachment_url`、`hints`
- 题目发布已支持：`draft` -> `published`
- 题目发布前自检已支持：
  - 同步自检：`POST /api/v1/authoring/challenges/:id/self-check`
  - 异步发布自检入队：`POST /api/v1/authoring/challenges/:id/publish-requests`
  - 查询最近一次发布自检：`GET /api/v1/authoring/challenges/:id/publish-requests/latest`
- Flag 已支持：`static` / `dynamic`
- 动态 Flag 注入已支持：实例启动时通过环境变量 `FLAG` 注入容器；拓扑节点可按 `inject_flag` 控制是否注入
- 镜像已支持：平台当前通过“已注册镜像”运行题目；运行时可按 `container.registry` 配置为匹配的私有 registry 拉取镜像，但不支持公开的在线 Dockerfile 构建导入
- 拓扑已支持：挑战拓扑、环境模板、多节点网络与 ACL 已有独立 API
- Writeup 已支持：挑战 Writeup 已有独立 API
- 标签已支持：通过独立 Tag API 和 `challenge_tags` 关系维护，不在当前创建题目 API 内直接写入

当前已支持：

- `POST /api/v1/authoring/challenge-imports`
- `GET /api/v1/authoring/challenge-imports/:id`
- `POST /api/v1/authoring/challenge-imports/:id/commit`
- `POST /api/v1/authoring/awd-service-template-imports`
- `GET /api/v1/authoring/awd-service-template-imports/:id`
- `POST /api/v1/authoring/awd-service-template-imports/:id/commit`
- `POST /api/v1/authoring/challenges/:id/self-check`
- `POST /api/v1/authoring/challenges/:id/publish-requests`
- `GET /api/v1/authoring/challenges/:id/publish-requests/latest`
- CLI `cmd/import-challenge-packs` 复用同一 `challenge.yml` 解析逻辑

当前明确不应在 v1 核心规范中写成“已支持”的能力：

- `manual` Flag 判题模式
- `runtime.type = none`
- 通过题目包直接在线构建 Dockerfile
- 通过题目包直接导入镜像 tar
- 多附件自动映射到题目详情页

同时约定：

- **Jeopardy / 解题赛题目包**：继续使用默认 `kind: challenge`，并省略 `meta.mode` 或显式写为 `jeopardy`
- **AWD 服务模板包**：仍复用 `challenge-pack-v1` 外壳，但必须额外声明 `meta.mode: awd`，并在 `extensions.awd` 内提供攻防运行定义
- 普通题目导入入口会拒绝 `meta.mode: awd` 的题目包，防止把 AWD 包误导入成 Jeopardy 题

### 4.1 AWD 镜像交付边界

AWD 服务模板包只声明运行镜像引用，不承担镜像构建职责。`runtime.image.ref` 必须指向平台运行节点可拉取的镜像；题目包内的 `docker/` 目录只作为源码、构建上下文和审计材料保留，当前平台导入 AWD 模板时不会自动执行 `docker build`。

推荐交付链路固定为：

```text
题目作者机器 / CI
  -> docker build
  -> docker scan / smoke test
  -> docker push registry

平台后台
  -> 上传 AWD 服务模板包
  -> 解析 runtime.image.ref
  -> 创建 AWD 模板
  -> 管理员加入比赛服务
  -> 队伍启动共享实例
  -> 平台轮次注入 Flag / Check / 计分
```

这条边界的含义是：

- 镜像构建失败、基础镜像不可拉取、`pip` / `apt` / `npm` / `composer` 等依赖源不可用，应在题目作者机器或 CI 阶段暴露并修复。
- 平台导入只校验题目包结构和 AWD 运行定义，不在上传请求中执行不受控构建。
- 平台运行只依赖最终镜像仓库，避免比赛现场临时访问 Docker Hub、PyPI、apt 源等外部服务。
- 管理员导入后仍应执行 checker preview；preview 通过后再将模板加入比赛服务并开赛。
- 如果后续要支持“上传源码后平台自动构建”，应作为独立的镜像构建任务系统设计，例如 `image_build_jobs`、隔离 builder、构建日志、超时、资源限制、registry 凭据和失败重试，而不是混入 AWD 模板导入接口。

---

## 5. 题目源包结构

### 5.1 根目录识别

为兼容常见打包方式，题目源包根目录按如下规则识别：

- 若 Zip 顶层直接包含 `challenge.yml`，则 Zip 顶层即根目录。
- 否则，Zip 顶层必须只包含一个目录，该目录视为根目录。

### 5.2 根目录必需文件

v1 核心最小要求：

- `challenge.yml`
- `statement.md`

### 5.3 根目录可选目录

- `attachments/`
- `docker/`
- `writeup/`

其中：

- `attachments/` 用于附件原始材料
- `docker/` 用于容器源码、Dockerfile、构建上下文和审计材料
- `writeup/` 用于题解原稿，不代表一定会自动导入平台

### 5.4 结构示例

```text
web-sqli-01.zip
  web-sqli-01/
    challenge.yml
    statement.md
    attachments/
      data.sql
    docker/
      Dockerfile
      src/...
    writeup/
      solution.md
```

---

## 6. challenge.yml 规范

### 6.1 通用要求

- 编码：UTF-8
- 文件名：必须为 `challenge.yml`
- 顶层必须包含：
  - `api_version: v1`
  - `kind: challenge`

### 6.2 v1 核心字段

以下字段属于“当前平台建议直接对齐”的核心字段。

```yaml
api_version: v1
kind: challenge

meta:
  slug: web-sqli-login-01
  title: "Web-02 SQL Injection: Login Bypass"
  category: web
  difficulty: easy
  points: 100

content:
  statement: statement.md
  attachments:
    - path: attachments/data.sql
      name: data.sql

flag:
  type: dynamic
  prefix: flag

hints:
  - level: 1
    title: Hint 1
    cost_points: 30
    content: "登录接口把用户输入直接拼进了 SQL。"

runtime:
  type: container
  image:
    ref: "registry.example.edu/ctf/web-sqli-login:20260312"
```

字段说明：

- `meta.slug`：必填，建议使用小写字母、数字、`-`；当前平台会持久化到 `challenges.package_slug`，作为导入稳定 upsert 标识
- `meta.title`：必填，对应平台 `challenge.title`
- `meta.category`：必填，建议枚举 `web` `pwn` `reverse` `crypto` `misc` `forensics`
- `meta.difficulty`：必填，当前平台支持 `beginner` `easy` `medium` `hard` `insane`
- `meta.points`：必填，正整数
- `content.statement`：必填，v1 固定为 `statement.md`
- `content.attachments[]`：可选，对应附件材料；当前平台会保留附件并对外生成单一下载入口
- `hints[]`：可选，对应平台 `challenge_hints`
  - `level`：提示级别，未填写时导入器可按顺序补齐
  - `content`：提示内容
  - `cost_points`：提示代价
- `flag.type`：必填，仅允许 `static` / `dynamic`
- `flag.value`：静态题目必填
- `flag.prefix`：可选，对应平台 `flag_prefix`，默认建议 `flag`
- `runtime.type`：当前首版仅支持 `container`
- `runtime.image.ref`：容器题必填，表示最终运行镜像引用

### 6.3 v1 建议字段

以下字段设计合理，但当前平台不一定自动导入，应作为“作者侧规范”保留：

- `tags`
  - 建议继续使用字符串数组
  - 当前平台标签是独立实体，导入器未来应做映射或创建
- `writeup`
  - 当前平台有 Writeup 能力，但没有题目包自动导入链路

示例：

```yaml
meta:
  slug: web-sqli-login-01
  tags:
    - vuln:sqli
    - stack:php
    - kp:auth-bypass

content:
  attachments:
    - path: attachments/data.sql
      name: data.sql
      sha256: "<hex>"

writeup:
  file: writeup/solution.md
  visibility: private
```

### 6.4 v1 拓扑扩展

当前平台已经支持挑战拓扑和环境模板，因此题目源包可以定义拓扑；但它属于扩展能力，不应阻塞普通单容器题目。

建议结构：

```yaml
topology:
  entry_node_key: web
  networks:
    - key: public
      name: Public
    - key: internal
      name: Internal
      internal: true
  nodes:
    - key: web
      name: Web
      service_port: 8080
      inject_flag: true
      network_keys: [public, internal]
    - key: db
      name: Database
      network_keys: [internal]
      inject_flag: false
  policies:
    - source_node_key: web
      target_node_key: db
      action: allow
      protocol: tcp
      ports: [3306]
```

约束建议：

- `entry_node_key` 必填
- 至少一个 `node`
- `nodes[].inject_flag` 建议显式声明
- 若题目是单节点单容器题，可完全省略 `topology`

当前平台中与拓扑直接对齐的字段包括：

- `entry_node_key`
- `networks[]`
- `nodes[].service_port`
- `nodes[].inject_flag`
- `nodes[].tier`
- `nodes[].network_keys`
- `nodes[].env`
- `nodes[].resources`
- `policies[]`

### 6.5 Docker 源码材料

v1.0 把 `docker/Dockerfile` 规定为所有题目包必填，这不合理。v1.1 调整为：

- 如果题目包以“源码交付、审计复现”为目标，强烈建议包含 `docker/`
- 如果题目包只用于归档已经构建好的镜像题，可不包含 `docker/`
- 只有当 `runtime.build.source = dockerfile` 时，`docker/Dockerfile` 才是必填

建议写法：

```yaml
runtime:
  image:
    ref: "registry.example.edu/ctf/web-sqli-login:20260312"
  build:
    source: dockerfile
    context_dir: docker
    dockerfile: Dockerfile
```

说明：

- `runtime.image.ref` 是运行事实
- `runtime.build.*` 是复现材料
- 当前平台运行时只需要 `runtime.image.ref`，不会直接消费 `runtime.build.*`

### 6.6 AWD 扩展

AWD 不应另起一套完全独立的包格式；但也不能直接拿 Jeopardy 的最小字段原样复用。推荐做法是：

- 继续复用 `challenge-pack-v1` 的根目录结构、`challenge.yml` 外壳与 `statement.md`
- 通过 `meta.mode: awd` 明确声明“这是 AWD 服务模板包”
- 通过 `extensions.awd` 补齐服务类型、Checker、Flag 策略、防守入口、访问端口与运行参数

最小可用示例：

```yaml
api_version: v1
kind: challenge

meta:
  mode: awd
  slug: awd-bank-portal-01
  title: "AWD Web-01 Bank Portal"
  category: web
  difficulty: hard
  points: 500

content:
  statement: statement.md

flag:
  type: dynamic
  prefix: awd

runtime:
  type: container
  image:
    ref: "registry.example.edu/ctf/awd-bank-portal:v1"

extensions:
  awd:
    service_type: web_http
    deployment_mode: single_container
    version: "v2026.04"
    checker:
      type: http_standard
      config:
        put_flag:
          method: PUT
          path: /api/flag
          expected_status: 200
          body_template: "{{FLAG}}"
        get_flag:
          method: GET
          path: /api/flag
          expected_status: 200
          expected_substring: "{{FLAG}}"
        havoc:
          method: GET
          path: /healthz
          expected_status: 200
    flag_policy:
      mode: dynamic_team
      config:
        flag_prefix: awd
        rotate_interval_sec: 120
    defense_entry:
      mode: http
    access_config:
      public_base_url: "http://{{TEAM_HOST}}:8080"
      service_port: 8080
      exposed_ports:
        - port: 8080
          protocol: tcp
          purpose: http
    runtime_config:
      instance_sharing: per_team
      service_port: 8080
```

字段约定：

- `meta.mode`
  - 必填，AWD 固定为 `awd`
- `extensions.awd.service_type`
  - 必填，对应 `awd_service_templates.service_type`
  - 当前枚举：`web_http` `binary_tcp` `multi_container`
- `extensions.awd.deployment_mode`
  - 必填，对应 `awd_service_templates.deployment_mode`
  - 当前枚举：`single_container` `topology`
- `extensions.awd.version`
  - 可选，对应模板版本字符串；未填写时平台默认回落到 `v1`
- `extensions.awd.checker.type`
  - 必填，对应 `checker_type`
  - 当前枚举：`legacy_probe` `http_standard`
- `extensions.awd.checker.config`
  - 建议必填，对应 `checker_config`
  - 对 `http_standard`，建议至少显式提供 `put_flag / get_flag / havoc`
- `extensions.awd.flag_policy.mode`
  - 必填，对应 `flag_mode`
  - 例如 `dynamic_team`
- `extensions.awd.flag_policy.config`
  - 可选，对应 `flag_config`
  - 可放 `flag_prefix`、轮转周期等模板级 Flag 策略
- `extensions.awd.defense_entry.mode`
  - 必填，对应 `defense_entry_mode`
  - 例如 `http`
- `extensions.awd.access_config`
  - 必填，对应 `access_config`
  - 建议至少包含 `service_port`
  - 建议把 `public_base_url`、`exposed_ports` 等攻防展示信息统一放在这里
- `extensions.awd.runtime_config`
  - 可选，对应 `runtime_config`
  - 建议写 `instance_sharing`、`service_port`、拓扑或运行时参数
  - 平台导入时会额外把 `runtime.image.ref` 解析成 `image_ref` 与 `image_id` 并补进模板运行配置

边界说明：

- `meta.points` 在 AWD 包中只作为**建议分值**保留，当前不会直接写入 AWD 模板；真正比赛分值仍在管理员配置比赛时设置
- AWD 模板导入成功后，平台会直接生成 `published` 状态模板，便于管理员在比赛题池里立即选题
- 比赛里的 Checker 覆盖、分值、顺序、可见性仍然属于**比赛级配置**，不应反向写回题库模板
- `runtime.image.ref` 必须是已构建并已推送到平台可访问 registry 的最终镜像引用；导入 AWD 模板不会自动构建 `docker/` 目录中的 Dockerfile

---

## 7. 字段映射矩阵

| 题目包字段 | 当前平台状态 | 说明 |
|---|---|---|
| `meta.title` | 已支持 | 直接映射题目标题 |
| `content.statement=statement.md` | 已支持 | 可直接映射到 `description` Markdown 原文 |
| `meta.category` | 已支持 | 建议使用既有分类 |
| `meta.difficulty` | 已支持 | 仅 `beginner/easy/medium/hard/insane` |
| `meta.points` | 已支持 | 直接映射 |
| `hints[].content/cost_points` | 已支持 | `level/title/cost_points/content` 全部可映射 |
| `flag.type=static/dynamic` | 已支持 | `manual` 不支持 |
| `flag.prefix` | 已支持 | 对应 `flag_prefix` |
| `runtime.image.ref` | 已支持 | 导入时可自动关联或创建镜像记录 |
| `content.attachments` | 已支持 | 平台会生成统一下载入口 |
| `tags` | 部分支持 | 平台有 Tag 能力，但不是题目创建时内联字段 |
| `extensions.topology` | 已保留 | 首版只做预览与扩展提示，不自动落现有拓扑表 |
| `writeup` | 已支持但需独立落库 | 有独立 API，不是题目创建接口内联字段 |
| `meta.slug` | 已支持 | 持久化到 `challenges.package_slug`，并作为导入稳定 upsert 标识与附件存储目录 |
| `runtime.type=none` | 不支持 | 当前创建题目必须依赖镜像 |
| `flag.type=manual` | 不支持 | 当前仅 static / dynamic |
| `runtime.build.source=dockerfile` | 未公开导入支持 | 可作为审计材料保留 |
| `runtime.build.source=tar` | 未公开导入支持 | 不应写成当前能力 |

### 7.1 AWD 扩展字段映射

| AWD 包字段 | 当前平台状态 | 说明 |
|---|---|---|
| `meta.mode=awd` | 已支持 | 作为 AWD 包入口识别条件 |
| `extensions.awd.service_type` | 已支持 | 映射到 `awd_service_templates.service_type` |
| `extensions.awd.deployment_mode` | 已支持 | 映射到 `awd_service_templates.deployment_mode` |
| `extensions.awd.version` | 已支持 | 映射到模板 `version` |
| `extensions.awd.checker.type` | 已支持 | 映射到 `checker_type` |
| `extensions.awd.checker.config` | 已支持 | 映射到 `checker_config` |
| `extensions.awd.flag_policy.mode` | 已支持 | 映射到 `flag_mode` |
| `extensions.awd.flag_policy.config` | 已支持 | 映射到 `flag_config` |
| `extensions.awd.defense_entry.mode` | 已支持 | 映射到 `defense_entry_mode` |
| `extensions.awd.access_config` | 已支持 | 映射到 `access_config`，建议承载可攻击端口与展示入口 |
| `extensions.awd.runtime_config` | 已支持 | 映射到 `runtime_config`，导入时会额外补齐 `image_ref/image_id` |
| `runtime.image.ref` | 已支持 | 导入 AWD 模板时会自动关联或创建镜像记录 |
| `meta.points` | 部分支持 | 当前只作为建议分值，不直接落 AWD 模板 |

---

## 8. 校验规则

### 8.1 源包静态校验

- 必须存在 `challenge.yml`
- 必须存在 `statement.md`
- `content.statement` 必须等于 `statement.md` 或为空时默认回落到 `statement.md`
- `meta.difficulty` 必须属于 `beginner/easy/medium/hard/insane`
- `flag.type` 必须属于 `static/dynamic`
- 若存在 `attachments[*].path`，必须位于 `attachments/` 下
- 若存在 `runtime.build.source=dockerfile`，则必须存在 `docker/Dockerfile`

### 8.2 安全校验

- 拒绝 Zip Slip：禁止绝对路径与 `..`
- 拒绝 symlink
- 限制最大解包文件数、总大小、单文件大小
- 不允许在 manifest、题面、附件、Dockerfile 中内嵌平台 secret

### 8.3 平台导入前置条件

即使未来实现导入器，也必须先满足：

- `runtime.image.ref` 已可被平台所在 Docker/registry 环境访问；若使用私有 registry，需要在后端 `container.registry` 配置匹配的 `server` 与凭据
- 若要自动创建题目，导入器必须把 `runtime.image.ref` 先映射为平台 `images` 记录
- 若要自动导入拓扑，导入器必须单独调用挑战拓扑落库流程
- 若要自动导入 Writeup，导入器必须单独调用 Writeup 落库流程

私有 registry 拉取能力的验证边界是：使用带认证的 registry 推送测试镜像，删除运行节点本地同名 tag，然后通过后端 runtime engine 拉取镜像。该验证覆盖 `container.registry` 配置读取、registry 域名匹配、Docker `RegistryAuth` 传递和镜像拉取链路；它不覆盖镜像构建，也不代表平台会接管 registry 的部署和账号生命周期。

### 8.4 平台发布校验边界

平台当前把“题目是否可发布”拆成两层：

- 作者侧本地验证
  - 强烈建议先做，但不是上传题目包的强制前置条件
  - 目的：缩短反馈回路，尽早发现 Dockerfile、依赖、端口、初始化脚本或 Flag 配置问题
- 平台侧发布自检
  - 点击后台“发布”后，题目不会立刻公开，而是进入发布自检队列
  - worker 会执行预检和真实运行时拉起
  - 通过后自动发布
  - 失败则保持 `draft`，并通知题目发布者

平台自检负责确认“题目在平台环境里能被正确拉起和判定”，包括：

- Flag 配置是否完整
- 镜像引用是否可用
- 拓扑或单容器运行请求是否能成功创建
- 运行时资源是否能成功清理

平台自检不负责证明以下内容：

- 题目设计是否有教学价值
- 题面是否足够清晰
- 题目难度是否合适
- 预期解法是否唯一或最优
- 学生是否一定能按作者预期路径解出题目

---

## 9. 推荐制作流程

### 9.1 当前最稳妥流程

1. 本地制作题目源码、题面、Hint、附件与可选拓扑定义
2. 本地构建并验证镜像
3. 本地完成最小可用验证
4. 推送镜像或确保镜像在平台运行节点可见；私有 registry 镜像需要平台后端配置 `container.registry`
5. 生成 challenge pack 作为归档与审计材料
6. 在平台中分别创建：
   - 镜像
   - 题目
   - Flag
   - 可选拓扑
   - 可选 Writeup
7. 点击发布，进入平台发布自检队列
8. 等待平台通知结果
   - 通过：题目自动发布
   - 失败：根据失败摘要回修后重新提交发布检查

### 9.2 本地最小验证清单

推荐至少覆盖以下检查：

- 题目包结构检查
  - `challenge.yml`
  - `statement.md`
  - 附件路径和引用路径存在
- 镜像可用性检查
  - `docker build` 成功
  - 关键启动命令成功
  - 暴露端口与题面描述一致
- 运行时检查
  - 服务可访问
  - 初始化脚本能跑完
  - 题目不会启动即崩溃
- 判题检查
  - 静态 Flag 可提交
  - 动态 Flag 注入后能被题目读取
- 清理检查
  - 容器退出或删除后不会残留必须人工清理的关键资源

### 9.3 不推荐做法

- 在题目包中存放静态 Flag 明文
- 把 `latest` 当作正式镜像版本
- 用题目包假装描述 `manual` 判题，但平台端根本没有对应实现
- 用单一 v1 文档同时要求“必须有 Dockerfile”和“只交付 registry 镜像”
- 只在本地验证，从不走平台发布自检
- 完全跳过本地验证，把平台自检当成唯一调试手段

---

## 10. 示例说明

仓库中的示例目录：

- `ctf/docs/contracts/examples/challenge-pack-v1/web-hello-01/`
- `ctf/docs/contracts/examples/challenge-pack-v1/web-sqli-login-01/`

这些示例用于说明“作者侧源包长什么样”，不代表当前平台已经支持直接上传 Zip 并一键导入全部内容。

---

## 11. 后续建议

如果后续要把题目包真正做成平台能力，建议按顺序落地：

1. 先实现只读校验器：上传 Zip，返回 manifest/文件级错误
2. 再实现“镜像引用 + 题目元数据 + Hint”的最小导入
3. 然后补 Tag、附件上传、Writeup 导入
4. 最后补拓扑导入、模板引用、源码构建队列

这样比直接承诺“大而全的 v1 导入器”更稳。
