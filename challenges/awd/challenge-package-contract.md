# AWD 题目包契约

本文是 AWD 题目包在当前仓库中的正式契约。老师新增或维护 AWD 题目时，`challenge.yml`、`docker/` 和本地联调方式都应以这里为准。

这份契约只约束两件事：

- 题目包本身必须提供什么
- 平台运行时如何识别“这是不是 ctf 项目下受管的容器”

它不把宿主机上的目录名、老师本地进入哪个子目录执行 `docker compose`，当作平台运行归属的事实来源。

## 1. 题目包边界

一个可导入的 AWD 题目包至少包含：

```text
<slug>/
├── challenge.yml
├── statement.md
├── docker/
│   ├── docker-compose.yml
│   ├── Dockerfile
│   ├── app/
│   └── check/
└── writeup/
```

约束如下：

- `challenge.yml` 必须声明 `meta.mode: awd`
- `runtime.type` 必须为 `container`
- `runtime.image.ref` 必须指向平台实际会拉起的题目镜像
- 本地调试入口固定放在 `docker/`
- 本地 `docker/docker-compose.yml` 只用于老师构建、调试、验题，不作为平台实例归属判断依据

## 2. 平台归属契约

AWD 题目在平台内启动后，容器是否归属于 `ctf` 项目，不看下面这些信息：

- 题目源码放在 `challenges/awd/...` 还是别的目录
- 老师是在仓库根目录、`docker/` 目录还是其它目录执行 `docker compose`
- 本地 compose project name 是什么
- 容器名是否碰巧带有 `ctf`、`docker` 或题目 slug

平台只看运行时 label。

当前正式约定：

- 平台受管题目实例容器必须带：
  - `managed-by=ctf-platform`
  - `ctf-component=challenge-instance`
- 平台受管题目实例网络必须带：
  - `managed-by=ctf-platform`
  - `ctf-component=challenge-instance`
- 平台受管 checker sandbox 容器必须带：
  - `ctf.role=checker-sandbox`

含义可以这样理解：

- `managed-by=ctf-platform` 表示这个资源由平台接管生命周期
- `ctf-component=challenge-instance` 表示这是题目实例，而不是数据库、后台服务或其它基础设施
- `ctf.role=checker-sandbox` 表示这是平台为判题临时拉起的 checker 容器

后续无论题目源码目录怎么整理，只要运行资源满足这组 label，就应视为 `ctf` 项目下的正式受管容器。

## 3. 出题人需要遵守的规则

老师在编写题目包时，默认遵守下面几条：

- 不要把“目录位置”当成平台运行归属的约束条件
- 不要依赖本地 compose project name、容器名或目录名，让平台推断题目归属
- 不要在题目自带的 `docker-compose.yml` 中发明另一套与平台冲突的受管 label 语义
- 题目本地 compose 可以保留最小调试配置，但它不是平台运行契约

更具体地说：

- `docker/docker-compose.yml` 可以继续只表达本地如何 build 和启动
- 平台真正比赛实例的 label、网络和容器命名由平台运行时统一注入
- AWD 比赛实例容器名当前统一为 `ctf-instance-<challenge-name>-c<contest-id>-t<team-id>`
- 这里的 `<challenge-name>` 由题目包 slug 优先生成；没有 slug 时回退到题目标题清洗后的结果
- 如果本地 compose 需要额外标签，也只能作为老师本地排查辅助，不能替代平台受管 label

## 4. challenge.yml 里和容器归属相关的最小要求

`challenge.yml` 至少应保证：

- `runtime.image.ref` 对应的镜像就是平台会用于创建实例的镜像
- `extensions.awd.runtime_config` 只描述服务访问、实例共享、checker 所需运行参数
- 不把“平台必须在哪个宿主机目录下启动容器”写进题目包字段

这里要刻意区分两层：

- 题目包关心“镜像是什么、服务怎么访问、checker 怎么校验”
- 平台关心“实例容器怎么命名、怎么打 label、怎么回收”

这两层不要混在一个字段里。

## 5. 本地调试与平台运行的关系

老师本地通常会这样做：

```bash
cd challenges/awd/<period>/<slug>/docker
docker compose up --build
```

这只是题目开发和验题动作。它的目的只有两个：

- 确认镜像能正常构建
- 确认题面、服务链路和 checker 示例能跑通

平台正式运行时：

- 不依赖老师当前在哪个目录执行过 compose
- 不依赖本地 compose 生成的 project 名
- 统一由平台按照运行时 label 识别和管理题目实例

## 6. 推荐检查项

老师提交 AWD 题目之前，至少自查：

- `challenge.yml` 已声明 `meta.mode: awd`
- `runtime.image.ref` 与本地 build 出来的镜像一致
- 本地 `docker compose up --build` 可以启动
- 本地 `check/check.py` 可以跑通
- 没有把目录名、compose project name 或容器名当成平台归属契约
- 题目需要的平台归属语义，已经明确交给平台 label 管理
