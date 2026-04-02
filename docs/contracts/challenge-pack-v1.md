# Challenge Pack Specification v1 (`challenge-pack-v1`)

> 适用范围：CTF 题目制作、归档、审核、后台导入与 CLI 导入  
> 文档目的：定义当前仓库实际采用的题目包目录结构、`challenge.yml` 规范、导入行为与校验边界  
> 版本：v1.3 | 日期：2026-04-01

---

## 1. 当前基线

`challenge-pack-v1` 是当前仓库内已经落地的题目包契约。

当前实现已支持：

- 后台上传 Zip 题目包并生成导入预览
- 后台确认导入并创建或更新 Challenge 草稿
- CLI `cmd/import-challenge-packs` 复用同一套 `challenge.yml` 解析逻辑
- 解析题目元数据、题面、附件、提示、Flag、运行时镜像信息
- 透传 `extensions.topology` 到导入预览结果

当前相关接口：

- `POST /api/v1/admin/challenge-imports`
- `GET /api/v1/admin/challenge-imports/:id`
- `POST /api/v1/admin/challenge-imports/:id/commit`

---

## 2. 设计目标

- 题目包必须能独立描述题目元数据、题面、附件和运行时引用
- 同一题目包既可用于归档，也可直接进入当前导入链路
- 核心字段必须与现有平台数据模型明确对齐
- 扩展字段可以保留，但不能伪装成当前已经自动落库的能力
- 规范以当前仓库实现为准，不额外承诺未实现行为

---

## 3. 题目包结构

### 3.1 Zip 根目录识别

导入器按以下规则识别题目包根目录：

- 如果 Zip 顶层直接包含 `challenge.yml`，则 Zip 顶层就是题目包根目录
- 否则，Zip 顶层必须只包含一个目录，并且该目录下存在 `challenge.yml`

不满足上述规则时，导入会失败。

### 3.2 根目录必需文件

最小可导入题目包应包含：

- `challenge.yml`
- 题面文件

默认推荐题面文件为 `statement.md`。如果 `content.statement` 为空，解析器会回落到 `statement.md`。

### 3.3 根目录推荐目录

以下目录不是导入器强制要求，但符合当前仓库示例与制作习惯：

- `attachments/`
- `docker/`
- `writeup/`

约定用途：

- `attachments/`：题目附件
- `docker/`：容器源码、Dockerfile、构建上下文与审计材料
- `writeup/`：题解原稿或内部材料，当前不会随题目包自动导入

### 3.4 推荐结构示例

```text
web-sqli-login-01.zip
  web-sqli-login-01/
    challenge.yml
    statement.md
    attachments/
      data.sql
      readme.txt
    docker/
      Dockerfile
      app.py
    writeup/
      solution.md
```

---

## 4. `challenge.yml` 总体要求

### 4.1 文件要求

- 编码：UTF-8
- 文件名：必须为 `challenge.yml`
- 顶层必须包含：
  - `api_version: v1`
  - `kind: challenge`

### 4.2 顶层结构

当前实现识别以下顶层字段：

```yaml
api_version: v1
kind: challenge
meta: ...
content: ...
flag: ...
hints: ...
runtime: ...
extensions: ...
```

---

## 5. 核心示例

```yaml
api_version: v1
kind: challenge

meta:
  slug: web-sqli-login-01
  title: "Web-02 SQL Injection: Login Bypass"
  category: web
  difficulty: easy
  points: 100
  tags:
    - vuln:sqli
    - stack:sqlite
    - kp:auth-bypass

content:
  statement: statement.md
  attachments:
    - path: attachments/data.sql
      name: data.sql
    - path: attachments/readme.txt
      name: readme.txt

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
    ref: registry.example.edu/ctf/web-sqli-login-01:20260312

extensions:
  topology:
    source: docker/topology.yml
    enabled: false
```

---

## 6. 字段规范

### 6.1 `meta`

```yaml
meta:
  slug: web-sqli-login-01
  title: "Web-02 SQL Injection: Login Bypass"
  category: web
  difficulty: easy
  points: 100
  tags:
    - vuln:sqli
```

字段说明：

- `meta.slug`
  - 必填
  - 建议只使用小写字母、数字和 `-`
  - 导入后持久化到 `challenges.package_slug`
  - 是当前导入链路的稳定 upsert 标识
- `meta.title`
  - 必填
  - 映射到 `challenges.title`
- `meta.category`
  - 建议使用 `web` `pwn` `reverse` `crypto` `misc` `forensics`
  - 当前解析器会将未知值归一化为 `misc`
- `meta.difficulty`
  - 建议使用 `beginner` `easy` `medium` `hard` `insane`
  - 当前解析器会做如下归一化：
    - `hell` -> `insane`
    - 未识别值 -> `easy`
- `meta.points`
  - 必填
  - 必须大于 `0`
- `meta.tags`
  - 可选
  - 当前解析器会读取该字段
  - 当前导入流程不会自动写入挑战标签关系

### 6.2 `content`

```yaml
content:
  statement: statement.md
  attachments:
    - path: attachments/data.sql
      name: data.sql
```

字段说明：

- `content.statement`
  - 可选
  - 为空时默认读取 `statement.md`
  - 允许指向题目包根目录内任意相对路径文件
  - 对应题目描述 Markdown 原文
  - 读取后不能为空
- `content.attachments`
  - 可选
  - 每项支持：
    - `path`：必填，相对于题目包根目录
    - `name`：可选，下载时显示名称；为空则取文件名
  - 当前实现要求目标必须是包内真实文件，且不能越出题目包根目录

附件导入行为：

- 无附件：不生成 `attachment_url`
- 单附件：按原文件保存，并生成单文件下载地址
- 多附件：自动打包为一个 zip，再生成统一下载地址

### 6.3 `flag`

```yaml
flag:
  type: static
  value: flag{example}
  prefix: flag
```

当前实现支持的 `flag.type`：

- `static`
- `dynamic`
- `regex`
- `manual_review`

字段说明：

- `flag.type`
  - 必填
  - 仅支持上述四种取值
- `flag.value`
  - `static` 必填
  - `regex` 必填
  - `dynamic` 和 `manual_review` 可省略
- `flag.prefix`
  - 可选
  - 默认值为 `flag`

导入行为：

- `static`：存储哈希值，不回存明文
- `dynamic`：配置为动态 Flag 模式
- `regex`：会先编译校验正则，再写入 `flag_regex`
- `manual_review`：配置为人工复核模式

### 6.4 `hints`

```yaml
hints:
  - level: 1
    title: Hint 1
    cost_points: 10
    content: "先确认入口参数。"
```

字段说明：

- `hints` 可选
- 每个提示的 `content` 必填且不能为空
- `level` 可选
  - 小于等于 `0` 或缺失时，按顺序自动补为 `1..n`
  - 不允许重复
- `title` 可选
  - 为空时自动补为 `Hint <level>`
- `cost_points` 可选
  - 小于 `0` 时会归一化为 `0`

导入时会覆盖当前题目的全部 Hint 记录。

### 6.5 `runtime`

```yaml
runtime:
  type: container
  image:
    ref: ctf/web-sqli-login-01:latest
```

字段说明：

- `runtime.type`
  - 当前推荐使用 `container`
  - 只有 `runtime.type=container` 时，解析器才会生成运行时镜像引用
- `runtime.image.ref`
  - 优先使用
  - 例如：`registry.example.edu/ctf/web:20260401`
- `runtime.image.name` + `runtime.image.tag`
  - 可替代 `ref`
  - 如果只给 `name`，则 `tag` 默认补为 `latest`

镜像解析行为：

- 导入提交时，会按 `name + tag` 查询或创建平台 `images` 记录
- 如果镜像记录已存在，会复用并恢复为可用状态
- 如果未提供有效镜像引用，当前导入流程仍可继续，但题目不会自动关联镜像

### 6.6 `extensions`

```yaml
extensions:
  topology:
    source: docker/topology.yml
    enabled: false
```

当前实现识别：

- `extensions.topology.source`
- `extensions.topology.enabled`

当前行为：

- 这两个字段会进入导入预览结果
- 当前导入提交不会自动写入挑战拓扑数据表
- 适合作为作者侧扩展声明和后续人工编排提示

---

## 7. 字段映射

| 题目包字段 | 当前导入行为 |
|---|---|
| `meta.slug` | 持久化到 `challenges.package_slug`，并作为 upsert 主标识 |
| `meta.title` | 映射到 `challenges.title` |
| `meta.category` | 映射到 `challenges.category`，未知值归一化为 `misc` |
| `meta.difficulty` | 映射到 `challenges.difficulty`，未知值归一化为 `easy` |
| `meta.points` | 映射到 `challenges.points` |
| `content.statement` | 文件内容写入 `challenges.description` |
| `content.attachments` | 生成平台附件下载地址并写入 `attachment_url` |
| `hints[]` | 覆盖写入 `challenge_hints` |
| `flag.type/value/prefix` | 配置题目 Flag 模式与摘要字段 |
| `runtime.image.*` | 解析镜像引用并关联或创建 `images` 记录 |
| `meta.tags` | 当前读取但不自动落库标签关系 |
| `extensions.topology.*` | 当前仅进入导入预览，不自动导入拓扑 |

---

## 8. 导入行为

### 8.1 预览阶段

上传 Zip 后，系统会：

1. 校验压缩包结构与安全限制
2. 解压并识别题目包根目录
3. 解析 `challenge.yml`
4. 读取题面文件
5. 校验附件、提示、Flag、镜像引用格式
6. 返回导入预览结果

当前预览结果包含：

- 题目基础信息
- 题面内容
- 附件列表
- Hint 列表
- Flag 摘要
- Runtime 摘要
- `extensions.topology`
- `warnings`

说明：

- 当前解析器保留 `warnings` 字段，但当前实现通常返回空列表

### 8.2 提交阶段

确认导入后，系统会：

1. 重新解析题目包目录
2. 保存附件或附件打包产物
3. 解析并关联镜像记录
4. 创建或更新 Challenge
5. 同步 Hint
6. 配置 Flag
7. 将题目保存为 `draft`

### 8.3 Upsert 规则

导入提交按以下顺序查找已有题目：

1. 优先按 `package_slug = meta.slug`
2. 如果不存在，再回退到：
   - `package_slug` 为空
   - `title` 相同
   - `category` 相同

这个回退规则用于复用尚未绑定 `package_slug` 的已有题目记录。

---

## 9. 校验与安全约束

### 9.1 Manifest 校验

- `api_version` 必须为 `v1`
- `kind` 必须为 `challenge`
- `meta.slug` 必填
- `meta.title` 必填
- `meta.points` 必须大于 `0`
- 题面文件必须存在且内容非空
- `flag.type` 必须属于 `static` `dynamic` `regex` `manual_review`
- `static` 和 `regex` 必须提供 `flag.value`
- 每个附件都必须存在且是文件
- Hint 级别不能重复

### 9.2 路径校验

所有包内文件路径都必须满足：

- 不能为空
- 必须是题目包根目录内的相对路径
- 不允许通过 `..` 逃逸出题目包根目录

### 9.3 Zip 安全限制

当前导入器执行以下限制：

- 禁止符号链接
- 单个压缩包最多 `128` 个文件
- 单文件解压后最大 `16 MiB`
- 总解压体积最大 `64 MiB`
- Zip 条目路径必须安全，不能产生 Zip Slip

---

## 10. 推荐制作约束

以下内容不是导入器硬性限制，但符合当前平台和示例的最佳实践：

- 题面固定使用 `statement.md`
- 附件集中放在 `attachments/`
- 运行时统一使用 `runtime.type: container`
- 镜像引用使用不可变标签，而不是长期依赖 `latest`
- `docker/` 仅作为源码复现和审计材料，不作为在线构建输入
- 不在题目包中放入平台级密钥、数据库口令或生产环境凭据
- `extensions.topology` 只声明来源和启用意图，实际拓扑编排仍走独立能力

---

## 11. 示例位置

仓库内示例目录：

- `ctf/docs/contracts/examples/challenge-pack-v1/web-hello-01/`
- `ctf/docs/contracts/examples/challenge-pack-v1/web-sqli-login-01/`

这些示例用于说明当前推荐的作者侧题目包组织方式。

---

## 12. 结论

当前 `challenge-pack-v1` 可以视为“作者侧源包结构 + 平台导入契约”的统一规范：

- 目录结构以 `challenge.yml + 题面文件 + 可选附件/源码材料` 为中心
- 平台已经支持预览和确认导入
- 当前真正自动落地的范围包括题目元数据、题面、附件、Hint、Flag 和镜像引用
- `tags` 与 `extensions.topology` 仍属于保留扩展，当前不会自动完整落库
