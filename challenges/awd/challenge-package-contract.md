# AWD 题目包契约

本文定义 AWD 题目包的下一阶段契约，用于后续 `defense workspace` 方案实现与题包迁移。新增或重构 AWD 题目时，应以本文为准，不再继续沿用“`challenge_app.py` 单文件暴露给学生”的老约定。

这份契约只约束三件事：

- 题目包本身必须提供什么
- 学生防守工作区和平台运行契约如何物理分层
- 平台运行时如何识别“这是不是 ctf 项目下受管的资源”

它不把宿主机目录名、老师在哪个子目录执行 `docker compose`、本地 compose project name 或容器名，当作平台运行归属的事实来源。

## 1. 设计边界

AWD 题包需要同时满足下面四点：

1. 学生通过防守入口拿到的是工作区，不是平台圈定好的答案文件。
2. 学生可以自己侦察、定位和修补业务代码，但不能接触 flag 注入代码、checker secret 或平台运行契约。
3. 运行中的正式服务由 `service runtime` 承载，学生 SSH 进入的是独立 `defense workspace`，不是 runtime 容器本体。
4. 对手的攻击面是业务服务对外暴露的行为，而不是 `ctf_runtime.py`、`check.py` 或平台控制面脚本。

## 2. 题目包结构

一个可导入的 AWD 题目包至少包含：

```text
<slug>/
├── challenge.yml
├── statement.md
├── docker/
│   ├── docker-compose.yml
│   ├── runtime/
│   │   ├── Dockerfile
│   │   ├── app.py
│   │   ├── ctf_runtime.py
│   │   └── entrypoint.sh
│   ├── workspace/
│   │   ├── src/
│   │   ├── templates/
│   │   ├── static/
│   │   └── data/
│   └── check/
│       └── check.py
└── writeup/
```

约束如下：

- `challenge.yml` 必须声明 `meta.mode: awd`
- `runtime.type` 必须为 `container`
- 平台构建模式下可以省略 `runtime.image.ref`，平台会生成 `<registry>/awd/<slug>:<tag>`；外部镜像引用模式下才必须填写完整 `runtime.image.ref`
- 本地调试入口固定放在 `docker/`
- 运行入口固定放在 `docker/runtime/app.py`
- 平台运行契约代码放在 `docker/runtime/ctf_runtime.py`，不要和业务漏洞代码混写
- 学生可侦察、可修补的业务资产统一放在 `docker/workspace/`
- checker 脚本和相关资产统一放在 `docker/check/`
- 本地 `docker/docker-compose.yml` 只用于老师 build、调试、验题，不作为平台实例归属判断依据

下面这种 legacy 结构不再推荐作为新题目基线：

```text
docker/
  app.py
  ctf_runtime.py
  challenge_app.py
```

原因很简单：

- 业务代码和平台契约混在同一层
- 学生一旦拿到容器 shell，就很难把 `ctf_runtime.py`、checker 资产和业务代码物理隔开
- `challenge_app.py` 单文件本身就在提示修补范围

## 3. 平台归属契约

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

后续无论题目源码目录怎么整理，只要运行资源满足这组 label，就应视为 `ctf` 项目下的正式受管资源。

## 4. challenge.yml 运行契约

`challenge.yml` 至少应保证：

- 平台构建模式提供 `docker/runtime/Dockerfile`
- 外部镜像引用模式下，`runtime.image.ref` 对应的镜像就是平台会用于创建实例的镜像
- `extensions.awd.runtime_config` 只描述服务访问、checker 和防守工作区边界
- 不把“平台必须在哪个宿主机目录下启动容器”写进题目包字段

这里要刻意区分两层：

- 题目包关心“镜像是什么、业务工作区怎么装配、checker 怎么校验”
- 平台关心“实例容器怎么命名、怎么打 label、怎么回收”

这两层不要混在一个字段里。

### 4.1 defense_workspace

学生可进入的防守范围，应通过目录级 `defense_workspace` 描述，而不是单文件 `editable_paths`。

推荐字段如下：

```yaml
extensions:
  awd:
    runtime_config:
      defense_workspace:
        entry_mode: ssh
        seed_root: docker/workspace
        workspace_roots:
          - docker/workspace/src
          - docker/workspace/templates
          - docker/workspace/static
          - docker/workspace/data
        writable_roots:
          - docker/workspace/src
          - docker/workspace/templates
          - docker/workspace/static
        readonly_roots:
          - docker/workspace/data
        runtime_mounts:
          - source: docker/workspace/src
            target: /workspace/src
            mode: rw
          - source: docker/workspace/templates
            target: /workspace/templates
            mode: rw
          - source: docker/workspace/static
            target: /workspace/static
            mode: rw
          - source: docker/workspace/data
            target: /workspace/data
            mode: rw
```

字段语义：

- `entry_mode`
  - 当前固定为 `ssh`
- `seed_root`
  - 工作区初始化快照来源目录
- `workspace_roots`
  - 学生在 defense workspace 内可见的业务目录根
- `writable_roots`
  - 学生可修改的目录根
- `readonly_roots`
  - 学生可查看但不可修改的目录根
- `runtime_mounts`
  - 这些业务根在 `service runtime` 内的挂载目标与模式

### 4.2 defense_scope

`defense_scope` 在新契约里不再承担“学生文件浏览边界”的职责。它如果存在，只能作为平台内部保护元数据使用，不直接下发给学生端。

推荐字段如下：

```yaml
extensions:
  awd:
    runtime_config:
      defense_scope:
        protected_paths:
          - docker/runtime/app.py
          - docker/runtime/ctf_runtime.py
          - docker/check/check.py
          - challenge.yml
        service_contracts:
          - /health 必须返回 200
          - checker 依赖的取旗接口必须保留鉴权
```

约束：

- `protected_paths` 只描述必须留在平台保护边界内的文件或目录
- `service_contracts` 只描述运行契约，不写漏洞位置、修复步骤或 payload
- 新题包不再使用 `defense_scope.editable_paths`

### 4.3 何时使用 `topology`

`topology` 不是 AWD 新题的默认标配。只有当攻击面、防守面或运行契约本身依赖多节点边界时，才应该把题目做成 `extensions.awd.deployment_mode: topology`；否则默认保持 `single_container`。

判断规则按下面两条收口：

- 如果把题目压成单容器后，漏洞成因、利用路径、修补点和 checker 语义仍然成立，就不要上 `topology`
- 如果压成单容器后，会把网络边界、节点信任关系、横向移动路径、服务间 ACL 或 flag 注入位置做假，就必须上 `topology`

适合使用 `topology` 的典型场景：

- `Web/API + DB/Cache`，并且利用或修补依赖公网与内网分层，例如 SSRF 打内网管理口、Web 节点能访问 Redis/DB 但外部不能直连
- `Reverse Proxy / Gateway + App`，漏洞或防守点在入口层，例如错误转发、Host 信任、路径重写穿透、`X-Forwarded-*` 误用
- `公开站点 + 内部管理面`，攻击链依赖本不应暴露到公网的内部管理接口、调试接口或探针接口
- `Worker / Queue / Broker / Callback`，攻击链依赖异步消费、回调目标、消息签名、重放窗口或消费侧权限边界
- `文件处理链`，例如上传服务、预览/转码服务和元数据存储分离，漏洞点不在主站入口而在内部处理节点
- `IoT / Agent / 控制面`，例如设备网关、消息代理、控制台三者之间的默认密钥、Topic 越权或内部 API 误暴露
- `Binary 网关 + 后端服务`，对外是 TCP 协议入口，但真实漏洞来自网关到内部鉴权、签名或状态服务的信任错误
- `落点后再横向移动`，需要先拿一个外部 foothold，再打第二个内部节点
- `防守点本身在网络策略`，例如应关闭不必要互通、限制节点间端口，而不是只改业务代码

不应该为了“看起来更真实”就使用 `topology`。下面这些场景默认仍应保持 `single_container`：

- 单体 Web 漏洞，例如 SSTI、SQLi、任意文件读、上传绕过、鉴权缺失，且攻击与修补都发生在同一服务
- 单体 API + 本地 SQLite / 本地文件存储，数据库只是实现细节，不是学生需要理解的网络边界
- 单个 TCP 服务，例如栈溢出、格式解析、逻辑口令、固定密钥、越界读写
- 只是额外挂一个数据库、缓存或静态服务，但攻击路径和防守路径并不依赖多节点关系

新增 AWD 题目时，出题人应先判断：

1. 题目的主要攻击路径是否依赖“某些节点只能被别的节点访问”
2. 学生的主要修补点是否包含入口代理、内网接口、节点互通或内部信任边界
3. 如果把它改成单容器，会不会把题目真实的攻击面或防守面简化成另一个题

只要上述问题里任一项答案为“会”，就应优先设计成 `topology`。

### 4.4 导入校验规则

管理员上传 AWD 包时，平台至少应校验：

- `defense_workspace` 必须存在，并且是对象
- `entry_mode` 当前只能为 `ssh`
- `seed_root` 必须是题目包内相对目录路径
- `workspace_roots`、`writable_roots`、`readonly_roots` 都必须是非空字符串数组
- 这些 roots 必须是题目包内相对目录路径，不能是绝对路径、单文件路径或 `..` 逃逸路径
- `writable_roots` 与 `readonly_roots` 不能重叠
- `workspace_roots` 必须完整覆盖 `writable_roots ∪ readonly_roots`
- `runtime_mounts.source` 必须来自 `workspace_roots`
- `workspace_roots` 不允许包含 `docker/runtime/`、`docker/check/`、`challenge.yml`、flag 路径或平台 secret 路径
- `defense_scope.protected_paths` 如果存在，必须至少包含 `docker/runtime/app.py`、`docker/runtime/ctf_runtime.py`、`docker/check/check.py`、`challenge.yml`
- 新题包不得再以 `docker/challenge_app.py` 这类单文件作为工作区主边界

## 5. 本地调试与平台运行的关系

老师本地通常会这样做：

```bash
cd challenges/awd/<period>/<slug>/docker
docker compose up --build
```

这只是题目开发和验题动作。它的目的只有两个：

- 确认镜像能正常构建
- 确认题面、服务链路和 checker 示例能跑通

本地 compose 可以把 `docker/workspace/` 挂进本地 runtime，方便老师调试，但这不等同于学生比赛时的 defense workspace 边界。

平台正式运行时：

- 不依赖老师当前在哪个目录执行过 compose
- 不依赖本地 compose 生成的 project 名
- 统一由平台按照运行时 label、runtime config 和 workspace contract 识别、装配和管理实例

## 6. 出题人自查清单

老师提交 AWD 题目之前，至少自查：

- `challenge.yml` 已声明 `meta.mode: awd`
- `docker/runtime/`、`docker/workspace/`、`docker/check/` 已物理拆开
- 平台构建模式已提供 `docker/runtime/Dockerfile`，外部镜像引用模式下 `runtime.image.ref` 与本地 build 出来的镜像一致
- 本地 `docker compose up --build` 可以启动
- 本地 `docker/check/check.py` 可以跑通
- 学生可见的是目录级工作区，不是单文件答案入口
- `ctf_runtime.py`、checker 脚本、flag 注入逻辑和平台 secret 不会进入学生工作区
- `defense_scope` 没有暴露漏洞提示
- 没有把目录名、compose project name 或容器名当成平台归属契约
- 题目需要的平台归属语义，已经明确交给平台 label 和 runtime config 管理

## 7. 迁移说明

当前仓库里如果还存在 legacy 平铺结构或 `defense_scope.editable_paths`，只能视为迁移兼容输入，不再作为新增题包设计基线。

迁移顺序建议是：

1. 先把题目包拆成 `runtime / workspace / check`
2. 再补 `defense_workspace` 目录级字段
3. 最后清理 legacy `challenge_app.py` 单文件暴露和学生侧文件工作台依赖
