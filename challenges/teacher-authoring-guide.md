# 教师出题指南

本文面向要在当前仓库中新增或维护 CTF 题目的老师。目标不是讲解某一类漏洞怎么设计，而是说明在这套题库里，一道题从构思到交付的标准做法。

## 1. 先理解当前交付形态

当前仓库只认可两类正式产物：

- `jeopardy/packs/<slug>/`：Jeopardy 题目真源目录，所有编辑都应在这里完成
- `jeopardy/dist/<slug>.zip`：Jeopardy 对外交付的题目包 zip

当前约定：

- Jeopardy 题目只改 `challenges/jeopardy/packs/<slug>/`
- Jeopardy 分发包只交付 `challenges/jeopardy/dist/<slug>.zip`
- AWD 题目只改 `challenges/awd/<contest-or-group>/<slug>/`
- 题目目录里的 `challenge.zip`、`case-evidence.zip` 这类文件，只是题内附件，不是外层交付包
- 平台当前没有公开的“题目包一键导入”接口，因此老师交付的是规范源包和分发 zip，不是直接在平台上点上传即可完成全部导入

如果是 AWD 题目，先看这份契约再落题：

- [awd/challenge-package-contract.md](/home/azhi/workspace/projects/ctf/challenges/awd/challenge-package-contract.md)

如果不想从空目录开始，可以直接使用模板：

- [jeopardy/templates/README.md](/home/azhi/workspace/projects/ctf/challenges/jeopardy/templates/README.md)
- [offline-static-template](/home/azhi/workspace/projects/ctf/challenges/jeopardy/templates/offline-static-template)
- [container-web-template](/home/azhi/workspace/projects/ctf/challenges/jeopardy/templates/container-web-template)
- [container-pwn-template](/home/azhi/workspace/projects/ctf/challenges/jeopardy/templates/container-pwn-template)

## 2. 先决定题型

在当前仓库里，题目大体分成两种。

### 2.1 离线题

适合：

- 附件分析类题目
- 编码、密码、逆向、取证静态样本
- 学生不需要启动靶机，只需要下载材料本地解题

典型特征：

- `runtime.type: none`
- 附件放在 `attachments/`
- flag 一般是 `static`

### 2.2 容器题

适合：

- Web、Pwn、交互式 Crypto、远程 Shell、上传/注入/绕过类题目
- 学生需要启动实例，访问 HTTP/TCP/SSH 服务

典型特征：

- `runtime.type: container`
- 题目运行材料放在 `docker/`
- flag 一般是 `dynamic`

如果是 AWD 容器题，再额外记一条：

- 平台运行归属看 label，不看老师本地进入哪个目录执行 `docker compose`

选择原则很简单：

- 如果题目的核心体验是“分析附件”，做离线题
- 如果题目的核心体验是“访问服务并利用”，做容器题

## 3. 出题的最低质量标准

无论哪种题型，都必须满足下面几条。

### 3.1 题目必须真的能做

不能出现：

- 空附件
- `flag{placeholder}`
- `FLAG_PLACEHOLDER`
- 只有题面，没有有效材料
- 容器能 build 但按题面无法拿到 flag

### 3.2 题目必须可重复验证

老师自己至少要能稳定复现：

- 题目如何启动
- 题目如何访问
- 题目如何拿到 flag

### 3.3 容器题必须自包含

容器题默认要求：

- 不依赖仓库外部数据库、外部缓存、外部 API
- 单独 `docker build` 后可运行
- 启动后能直接按题面访问

如果题目需要数据库，也应在容器内自初始化，或者通过题包自身材料完成，不要依赖宿主机另起 MySQL/Postgres。

### 3.4 题面必须说清楚“学生该做什么”

题面至少要回答：

- 学生拿到什么
- 学生怎么访问
- 学生目标是什么
- 至少一条有效提示

## 4. 目录结构

一个标准题目目录最少如下：

```text
jeopardy/packs/<slug>/
├── challenge.yml
├── statement.md
├── attachments/    # 可选
├── docker/         # 容器题必需
└── writeup/        # 可选
```

说明：

- `challenge.yml`：题目元数据
- `statement.md`：题面
- `attachments/`：学生下载的附件
- `docker/`：容器题运行上下文；容器入口统一为 `docker/Dockerfile`
- `writeup/`：官方题解，可后补

## 5. slug、分类、难度约定

### 5.1 slug 命名

建议格式：

- `<category>-<topic>`
- 例如：`web-sqli-login-bypass`
- 例如：`pwn-ret2text`
- 例如：`forensics-browser-history`

要求：

- 全小写
- 用 `-` 分词
- 稳定，不要频繁改名

### 5.2 分类

当前题库使用：

- `crypto`
- `forensics`
- `misc`
- `pwn`
- `reverse`
- `web`

### 5.3 难度

当前仓库中常见值：

- `beginner`
- `easy`
- `medium`
- `hard`
- `insane`

### 5.4 标签

`tags` 用来补充“分类”和“难度”之外的检索信息。

不要把下面这些内容再写进 `tags`：

- 分类，例如 `web`、`pwn`、`crypto`
- 难度，例如 `easy`、`medium`
- 状态，例如“已解出”“待攻克”

推荐把标签控制在 `2` 到 `4` 个，优先表达下面几类信息：

- `topic:` 题型或漏洞点，例如 `topic:xss`、`topic:file-upload`、`topic:sqli`
- `kp:` 知识点，例如 `kp:rsa`、`kp:shift`、`kp:ret2libc`
- `stack:` 技术栈或运行环境，例如 `stack:php`、`stack:flask`、`stack:linux`
- `tool:` 关键工具，例如 `tool:burp`、`tool:gdb`、`tool:ida`
- `src:` 来源或专题，例如 `src:root-me`、`src:school-match`

建议顺序：

- 先写 `topic`
- 再写 `kp`
- 再写 `stack`
- 最后按需补 `tool` 或 `src`

命名约定：

- 全小写
- 简短、可检索
- 多词时使用 `-` 连接
- 不写整句说明
- 不重复表达同一个意思

示例：

```yaml
tags:
  - topic:file-upload
  - topic:bypass
  - stack:php
  - src:root-me
```

```yaml
tags:
  - kp:rsa
  - kp:crt
  - tool:sage
  - src:crypto-training
```

## 6. challenge.yml 怎么写

### 6.1 离线题最小示例

```yaml
api_version: v1
kind: challenge

meta:
  slug: crypto-ascii-shift
  title: ASCII 移位
  category: crypto
  difficulty: beginner
  points: 100
  tags:
    - kp:shift

content:
  statement: statement.md
  attachments:
    - path: attachments/challenge.txt
      name: challenge.txt

flag:
  type: static
  value: flag{replace-with-real-flag}
  prefix: flag

hints:
  - level: 1
    title: Hint 1
    content: 先确认字符与整数之间的映射关系。
```

### 6.2 容器题最小示例

```yaml
api_version: v1
kind: challenge

meta:
  slug: web-rootme-file-upload-double-ext
  title: 文件上传：双扩展绕过（Root-Me）
  category: web
  difficulty: medium
  points: 100
  tags:
    - src:root-me
    - topic:file-upload
    - topic:bypass
    - stack:web

content:
  statement: statement.md
  attachments: []

flag:
  type: dynamic
  prefix: flag

hints:
  - level: 1
    title: Hint 1
    content: 前端只看最后一个扩展名，但后端存储文件名时可能做了不同的截断。

runtime:
  type: container
  image:
    ref: registry.example.edu/ctf/web-rootme-file-upload-double-ext:latest
```

### 6.3 关键字段说明

- `api_version`
  - 当前固定为 `v1`
- `kind`
  - 当前固定为 `challenge`
- `meta.slug`
  - 题目唯一标识
- `meta.title`
  - 展示标题
- `meta.category`
  - 六大类之一
- `meta.difficulty`
  - 难度等级
- `meta.points`
  - 题目分值
- `meta.tags`
  - 知识点、来源、技术栈
- `content.statement`
  - 通常固定指向 `statement.md`
- `content.attachments`
  - 需要下载样本的题目应写清附件路径和展示文件名
- `hints`
  - 提示列表，建议至少 1 条
- `flag.type`
  - `static`：固定答案
  - `dynamic`：运行时注入或服务端生成
- `flag.value`
  - `static` 题必填
- `flag.prefix`
  - 动态 Flag 前缀，默认可用 `flag`
- `runtime.type`
  - 容器题用 `container`
- `runtime.image.ref`
  - 题目实际运行镜像引用

## 7. statement.md 怎么写

`statement.md` 是展示在题目详情页“题目描述”区块里的正文，不是独立文章页。默认按下面的规则写：

- 不要再写题目名。页面外层已经展示标题，`statement.md` 里重复写 `# 题目名` 会出现两次标题。
- 不要再写 `## 题目描述`。页面外层已经有“题目描述”区块标题，再写一次会形成重复结构。
- 开头直接写 1 到 2 段正文，说明背景、入口和目标。
- 优先使用 `## 目标`、`## 访问入口` / `## 获取方式`、`## 补充说明` 这类真正帮助学生操作的段落。
- `hints` 已经有独立字段时，尽量不要在 `statement.md` 再写一个 `## 提示` 段，避免信息重复。

### 7.1 离线题示例结构

```md
下载附件后分析样本，恢复 flag。

## 获取方式

下载附件 `challenge.txt`，还原编码结果并恢复 flag。

## 目标

1. 下载附件并识别编码方式。
2. 还原原始内容并恢复 flag。

## 补充说明

- 先确认字符与整数之间的映射关系。
```

### 7.2 容器题示例结构

```md
启动实例后访问上传页面，构造双扩展文件绕过校验并读取 flag。

## 目标

1. 启动实例并确认上传入口。
2. 构造可绕过校验的文件并完成上传。
3. 读取服务中的动态 flag。

## 访问入口

启动实例后访问平台分配的 URL。

## 补充说明

- 先看服务端如何处理文件名。
```

## 8. 离线题怎么做

### 8.1 基本流程

1. 新建 `jeopardy/packs/<slug>/`
2. 写 `statement.md`
3. 准备 `attachments/`
4. 计算附件 `sha256`
5. 写 `challenge.yml`
6. 生成 `jeopardy/dist/<slug>.zip`
7. 做可做性验证

如果是第一次出题，建议直接复制 `jeopardy/templates/offline-static-template/` 作为起点。

### 8.2 离线题材料要求

附件必须是有效材料，而不是壳子。

有效材料示例：

- 编码后的文本
- 可逆向的二进制
- 可分析的 apk/jar/pyc/docm
- 真正有内容的 zip
- 数据库、日志、配置、取证样本

无效材料示例：

- 空 zip
- 只写一行 `flag{placeholder}`
- 只有文件头、没有有效内容

### 8.3 离线题自测要求

老师至少应验证：

- 附件能打开
- 题面描述和附件一致
- 能从附件中稳定恢复答案
- 如果是 `static` 题，答案与题面预期一致

## 9. 容器题怎么做

### 9.1 基本流程

1. 新建 `jeopardy/packs/<slug>/`
2. 写 `statement.md`
3. 在 `docker/` 下准备 `Dockerfile`、源码、静态资源、入口脚本
4. 写 `challenge.yml`
5. `docker build` 验证
6. `docker run` 验证真实利用路径
7. 生成 `jeopardy/dist/<slug>.zip`

如果是第一次做容器题：

- Web 题建议直接复制 `jeopardy/templates/container-web-template/`
- Pwn 题建议直接复制 `jeopardy/templates/container-pwn-template/`

### 9.2 Docker 目录建议

容器题的镜像入口固定放在 `docker/Dockerfile`。源码可以按题目复杂度放在 `docker/app/`、`docker/src/` 或其他子目录，但不要把 Dockerfile 下沉到这些子目录里。

常见写法：

```text
jeopardy/packs/<slug>/docker/
├── Dockerfile
├── entrypoint.sh          # 可选
├── src/                   # 题目源码 / 页面源码
├── site/                  # 静态页面 / 下载页
└── downloads/             # 若题目需要向学生提供样本
```

Pwn 题常见写法：

```text
jeopardy/packs/<slug>/docker/
├── Dockerfile
├── entrypoint.sh
├── src/
│   └── challenge.c
└── site/
    └── index.html
```

这类题通常：

- `80/tcp` 提供下载页
- `9999/tcp` 提供交互利用端口
- 在镜像构建时编译题目二进制
- 从 `FLAG` 环境变量读取动态 flag

### 9.3 容器题设计要求

- `docker build` 必须独立成功
- 容器启动后必须按题面访问到服务
- 服务行为必须和题面一致
- 如果题目需要下载样本，下载页必须真的能下载到
- 如果题目是动态 flag，运行时必须能正确注入
- 不能依赖仓库外的 MySQL、Redis、第三方 API

### 9.4 动态 flag 建议

容器题推荐做成动态 flag：

- Web/Pwn/交互式 Crypto 常用动态 flag
- 服务端从 `FLAG` 环境变量读取
- 如果未注入，也应有一个本地默认值，方便老师本地调试

示例：

```php
$flag = getenv('FLAG') ?: 'flag{web_demo_container}';
```

或：

```python
flag = os.getenv("FLAG", "flag{crypto_demo_container}")
```

或：

```c
const char *flag = getenv("FLAG");
puts(flag ? flag : "flag{pwn_demo_container}");
```

### 9.5 容器题自测要求

老师至少要自己跑通一次完整链路：

- 能 build
- 能 run
- 能访问
- 能拿到 flag

只验证 build 成功不够。

例如：

```bash
docker build -t test-web-sqli-union jeopardy/packs/web-sqli-union/docker
docker run --rm -p 18080:80 -e FLAG='flag{demo}' test-web-sqli-union
```

然后再用浏览器或 `curl` 走实际利用路径。

如果是 Pwn 题，至少还应验证：

```bash
docker build -t test-pwn-demo jeopardy/packs/pwn-demo/docker
docker run --rm -p 18080:80 -p 19999:9999 -e FLAG='flag{demo}' test-pwn-demo
curl http://127.0.0.1:18080/
nc 127.0.0.1 19999
```

然后按题面给出的利用链实际打一遍，不要只验证端口能连通。

## 10. 出题推荐工作流

推荐按下面顺序执行。

### 10.1 先构思

先写清楚三件事：

- 学生看到什么
- 学生要做什么
- 正确解法是什么

### 10.2 再定题型

- 附件分析就做离线题
- 远程利用就做容器题

### 10.3 先做最小可用版

不要一开始就追求复杂。先做出：

- 最小题面
- 最小样本
- 最小 exploit path

### 10.4 先验证可做，再打磨美化

优先级应是：

1. 能做
2. 稳定
3. 题面清晰
4. 观感和包装

## 11. 题包自检建议

当前题目录入、上传与管理以平台为准，不再依赖仓库脚本批量生成题目包。

老师在本地至少应完成：

- 检查 `jeopardy/packs/<slug>/` 结构是否完整，`challenge.yml`、`statement.md`、附件或 `docker/` 是否齐全
- 确认题面、flag、提示、附件文件名和 `challenge.yml` 中的引用一致
- 如需生成外层分发包，重新打包 `jeopardy/packs/<slug>/`，确保 zip 根目录名仍为 `<slug>/`

容器题还应至少执行：

```bash
docker build -t test-<slug> jeopardy/packs/<slug>/docker
```

必要时再手工起容器验证一次核心解题链路。老师日常出题时，至少应验证自己新增或修改的那道题。

## 12. 题目完成后的交付标准

一道题可以认为“可交付”，至少要满足：

- `jeopardy/packs/<slug>/` 结构完整
- `challenge.yml` 正确
- `statement.md` 清晰
- 附件不是空壳
- 容器题已经 build 并实跑验证
- `jeopardy/dist/<slug>.zip` 已生成
- 题目可以被其他老师复现

## 13. 常见错误

最常见的问题如下。

### 13.1 空材料

例如：

- 空 zip
- 只有占位文本
- 只有文件头没有题意

### 13.2 题面和实现不一致

例如：

- 题面说访问 `/flag`，实际没有这个路径
- 题面说能上传拿 shell，实际上传后不可执行

### 13.3 容器不自包含

例如：

- 代码里连 `localhost:3306`
- 容器里没有数据库
- 依赖宿主机某个服务，但题包没说明

### 13.4 只测 build，不测 exploit path

这是容器题最常见的问题。能 build 不等于学生能做。

### 13.5 slug 和标题反复变化

会导致：

- zip 名称变动
- 平台侧记录混乱
- 老 writeup 和附件路径失效

## 14. 给老师的简化操作建议

如果是第一次出题，建议：

1. 先从离线题开始
2. 先做一个最小样本
3. 题面只写清楚“拿到什么、目标是什么、提示是什么”
4. 用仓库已有同类题做参考
5. 出完题后自己完整解一遍

如果要做容器题，建议：

1. 优先做单容器
2. 优先用 PHP/Python/静态文件这类简单运行时
3. 先保证 build 和 run 稳定
4. 再考虑多服务联动

## 15. 推荐交付给管理员的内容

老师交题时建议至少给管理员：

- `jeopardy/packs/<slug>/`
- `jeopardy/dist/<slug>.zip`
- 一段简短说明：
  - 题目类型
  - 预期难度
  - 是否容器题
  - 正确解法一句话
  - 是否已本地验证

如果题目较复杂，建议同时提供：

- 官方解题步骤
- 已验证的 payload 或关键命令
- 题目依赖说明

## 16. 一句话总结

在这套仓库里，老师出题的核心标准不是“把文件放进去”，而是：

- 有规范目录
- 有清晰题面
- 有真实材料
- 能本地验证
- 能稳定交付

只要按这个标准做，后续无论是平台接入、统一审计，还是学生实战启动，都会省很多返工成本。
