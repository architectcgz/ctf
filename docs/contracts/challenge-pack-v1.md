# Challenge Pack Specification v1 (challenge-pack-v1)

> 适用范围：CTF 靶场平台题目上传（教师/管理员）
> 文档目的：定义“题目包（Zip）”的结构与 `manifest.yml` 规范，以及平台在导入时必须执行的校验与安全要求。
> 版本：v1.0 | 日期：2026-03-03

> 存储与下载设计：见 `ctf/docs/architecture/backend/06-file-storage.md`

---

## 1. 术语

- Challenge（题目）：平台中可被练习/竞赛引用的题目实体。
- Challenge Pack（题目包）：教师/管理员上传的 Zip 包，包含题目元数据、题面、附件与容器构建上下文。
- Manifest：题目包根目录的 `manifest.yml`，描述题目元数据与运行参数。

---

## 2. 上传对象与原则

### 2.1 谁上传

- 教师：制作并上传题目包，提交审核或直接发布（取决于平台权限策略）。
- 管理员：可上传/修改/发布题目包，且可执行“在线构建”等高权限操作（如果平台支持）。

### 2.2 设计原则（必须）

- **可复现**：同一个题目包在受控环境下应能构建出一致的运行环境。
- **不泄露**：题目包不得包含平台密钥、数据库地址/口令、全局 flag secret 等敏感信息。
- **可校验**：导入前可静态校验字段与文件结构，导入失败必须给出可定位错误（字段/文件级）。
- **最小暴露**：容器仅暴露题目需要的端口，禁止把管理面/调试端口暴露给选手。

---

## 3. 题目包（Zip）结构规范

### 3.0 “题目包根目录”的定义（必须读）

实际制作 Zip 时，很多人会直接把一个文件夹打包（Zip 内会多一层顶层目录）。为减少踩坑，本规范把“题目包根目录”定义为：

- 若 Zip 顶层直接包含 `manifest.yml`，则 Zip 顶层就是根目录。
- 否则，Zip 顶层必须**只包含一个目录**（称为顶层目录），该顶层目录的内容视为根目录。

平台导入时必须先解析出根目录，再按后续规则做校验。

### 3.1 根目录文件（必须）

题目包 Zip 解压并定位到“根目录”后，必须包含：

- `manifest.yml`：题目元数据与运行参数
- `statement.md`：题面（Markdown）
- `docker/`：容器构建上下文目录（用于可复现与审计）
  - `docker/Dockerfile`（必须）

> 注：即便平台最终运行使用的是 registry 镜像，仍要求题目包携带 Dockerfile 作为“源码与审计材料”。

### 3.2 附件目录（可选）

- `attachments/`：题目附件目录（静态文件）

### 3.3 示例结构

```text
web-sqli-01.zip                   # 允许两种形式
  web-sqli-01/                    # 形式 A：带一个顶层目录（更常见）
    manifest.yml
    statement.md
    attachments/
      data.sql
      hint.png
    docker/
      Dockerfile
      src/...
  (或直接在 Zip 顶层放这些文件)  # 形式 B：不带顶层目录
```

---

## 4. manifest.yml 规范

### 4.1 基本约束

- 编码：UTF-8
- 格式：YAML（推荐 2 空格缩进）
- 文件名：必须为 `manifest.yml`
- 顶层字段 `spec_version` 必须为 `challenge-pack-v1`

### 4.2 字段说明（v1）

以下字段为规范定义；平台实现可增加额外字段，但不得改变已定义字段语义。

#### 顶层元数据

- `spec_version`（必填）：固定值 `challenge-pack-v1`
- `slug`（必填）：题目唯一标识
  - 仅允许小写字母、数字与 `-`
  - 长度建议 3~64
- `title`（必填）：题目标题（展示用）
- `category`（必填）：题目分类（建议枚举如下）
  - `web` `pwn` `reverse` `crypto` `misc` `forensics`
- `difficulty`（必填）：难度（建议枚举如下）
  - `beginner` `easy` `medium` `hard` `hell`
- `tags`（可选）：标签数组（建议使用“前缀约定”做分类）
  - `vuln:<name>`：漏洞类型，例如 `vuln:sqli` `vuln:xss` `vuln:ssrf`
  - `stack:<name>`：技术栈，例如 `stack:flask` `stack:gin` `stack:nginx` `stack:sqlite`
  - `kp:<name>`：知识点/能力点，例如 `kp:union-select` `kp:auth-bypass`

> 平台实现层仍可把 `tags` 当作字符串集合存储；前缀约定用于“教师侧制作规范”和“平台侧展示/筛选”的一致性。

#### 题面与提示

- `description.file`（必填）：题面文件路径，必须为 `statement.md`
- `hints`（可选）：提示数组
  - `text`（必填）：提示内容
  - `cost`（必填）：提示花费（分/币/点数，具体由平台解释）

#### 附件

- `attachments`（可选）：附件数组
  - `path`（必填）：相对路径，必须位于 `attachments/` 下
  - `name`（必填）：展示文件名
  - `sha256`（必填）：附件内容 sha256（hex 小写）

#### Flag 策略

- `flag.mode`（必填）：`static` / `dynamic` / `manual`
  - `static`：静态 flag（不建议用于竞赛）
  - `dynamic`：动态 flag（推荐用于线上练习与竞赛）
  - `manual`：人工评判（例如报告题/取证题）

> 题目包内禁止存放平台级 secret。`static` 的具体 flag 值如果必须配置，应在平台管理端单独输入并加密存储，不放入 Zip。

#### 运行与实例

- `runtime.type`（必填）：`container` / `none`
  - `container`：需要启动容器实例
  - `none`：无容器（线下题/纯文本题等）

- `runtime.image.source`（必填）：`registry` / `dockerfile` / `tar`
  - `registry`：使用已推送镜像（推荐生产默认）
  - `dockerfile`：由 `docker/` 构建（推荐用于可复现；是否平台在线构建取决于管理员策略）
  - `tar`：使用镜像 tar 包导入（适合离线环境）

- `runtime.image` 字段（按 source 不同要求不同）：
  - source=registry：
    - `ref`（必填）：镜像引用（禁止 `latest`，必须使用明确 tag 或 digest）
  - source=dockerfile：
    - `context_dir`（必填）：构建上下文目录，固定为 `docker`
    - `dockerfile`（必填）：固定为 `Dockerfile`
    - `build_args`（可选）：构建参数（必须是非敏感参数；禁止口令/令牌）
  - source=tar：
    - `path`（必填）：tar 文件相对路径（建议位于 `docker/` 下）
    - `image_name`（必填）：导入后的镜像名:tag（禁止 `latest`）

- `runtime.expose`（必填，container 才需要）：暴露端口数组
  - `container_port`（必填）：容器内端口
  - `protocol`（必填）：`tcp` / `udp`

- `instance`（可选）：实例策略
  - `ttl`：默认存活时间（例如 `2h`）
  - `max_per_user`：每用户最大实例数（建议 1）
  - `max_per_team`：每队最大实例数（竞赛模式下使用）

- `resources`（可选）：资源上限
  - `cpu`：例如 `0.5`
  - `memory`：例如 `256m`
  - `pids`：例如 `100`

### 4.3 最小示例（示意）

```yaml
spec_version: challenge-pack-v1
slug: web-hello-01
title: "Web-01 Hello Container"
category: web
difficulty: easy
tags: ["stack:docker"]

description: { file: statement.md }
hints:
  - text: "先确认服务端口与路由是否正常。"
    cost: 10

attachments: []

flag:
  mode: dynamic

runtime:
  type: container
  image:
    source: dockerfile
    context_dir: docker
    dockerfile: Dockerfile
  expose:
    - container_port: 8080
      protocol: tcp

instance: { ttl: "2h", max_per_user: 1 }
resources: { cpu: 0.5, memory: "256m", pids: 100 }
```

仓库内示例（未压缩形态）：

- `ctf/docs/contracts/examples/challenge-pack-v1/web-hello-01/`
- `ctf/docs/contracts/examples/challenge-pack-v1/web-sqli-login-01/`（更接近真实 CTF 的 Web 注入题）

仓库内示例（可直接上传的 Zip）：

- `ctf/docs/contracts/examples/challenge-pack-v1/web-sqli-login-01.zip`
- `ctf/docs/contracts/examples/challenge-pack-v1/README.md`（按分类组织示例的说明）

---

## 5. 平台导入校验（必须）

### 5.1 Zip 解包安全（必须）

- 拒绝绝对路径与包含 `..` 的路径（防止写任意位置）。
- 拒绝 symlink（防止绕过目录约束读取/写入）。
- 限制最大文件数、最大单文件大小、最大总大小（防 zip bomb）。
- 解包到隔离目录，导入完成后再移动到受控存储路径。

### 5.2 内容校验（必须）

- 必须存在 `manifest.yml`、`statement.md`、`docker/Dockerfile`。
- `description.file` 必须指向存在的文件且为 `statement.md`。
- `attachments[*].path` 必须位于 `attachments/` 下，且 sha256 校验通过。
- `runtime.expose[*].container_port` 必须是 1~65535；协议必须是 `tcp/udp`。
- 镜像引用禁止 `latest`（registry 与 tar）。
- 禁止在 manifest/附件中出现平台级敏感字段（例如 DB 密码、全局 secret）。平台可用关键字扫描做兜底，但不得以扫描作为唯一防线。

### 5.3 构建策略（强烈建议）

如果平台支持 `dockerfile` 在线构建：

- 构建必须异步化（队列），并设置超时与资源限额。
- builder 必须与平台服务网络隔离，避免 Dockerfile 在构建期访问平台内网服务。
- 构建过程中禁止注入任何生产 secret（包括 registry 凭据需最小权限且短期 token）。

默认策略建议：

- 教师/CI 本地构建并推送到学校私有 registry（`source=registry`），平台只负责拉取并运行。
- 题目包仍携带 Dockerfile 以便审计与后续复现。

---

## 6. 常见错误与处理建议

- 题目包导入失败但无定位信息：
  - 平台应返回字段级错误（例如 `attachments[0].sha256 mismatch`）并附带 `request_id`。
- 镜像构建偶发失败：
  - 平台应保留构建日志并可下载；并提供重试入口（需权限控制）。
- 选手能访问到不该访问的端口：
  - 检查 `runtime.expose` 仅包含必要端口；实例创建时只映射白名单端口。
