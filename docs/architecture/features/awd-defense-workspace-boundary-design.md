# AWD 防守工作区与边界设计

## 状态

已采用。

本设计替代以下两份阶段性文档，作为当前 AWD 防守能力的事实源：

- `awd-web-defense-workbench-design.md`
- `awd-defense-content-page-design.md`

## 1. 设计目标

根据当前课题目标和前序讨论，AWD 防守能力需要同时满足下面四个要求：

1. 学生需要自己判断该看哪里、该改哪里、该如何验证，平台不能直接给出修补目标。
2. 学生不能通过防守入口接触平台运行契约、checker secret、动态 flag 存储等受保护资产。
3. 平台仍要提供可操作的防守入口，避免把“能防守”完全退化成线下自建环境。
4. 设计要能写进毕业设计：赛中是真实攻防，赛后才做教学复盘，而不是赛中由平台指导修洞。

## 2. 问题重述

当前仓库里已经尝试过两条路线，但都不适合作为最终方案。

### 2.1 浏览器受控片段

“风险片段 / 受控 patch”方向虽然不再直接暴露完整源码，但本质上仍然在给学生提示：

- 哪个逻辑位置值得优先看
- 哪段代码和漏洞更相关
- 平台认为哪些点属于“修补区”

这与 AWD 防守方应自行定位问题的目标冲突。

### 2.2 浏览器文件工作台

“目录树 + 文件内容页”方向进一步放开后，会直接暴露：

- 真实文件路径
- 真实文件名
- 受控根目录结构
- 受控文件内容

即使只暴露 `editable_paths`，只要平台告诉学生“重点看这几个文件”，本质上仍然是在收缩答案空间。

### 2.3 直接 SSH 进入服务容器

“不给浏览器入口，只保留 SSH”也不能直接成立。如果 SSH 直接进入当前服务容器本体，学生仍然可能接触：

- `ctf_runtime.py`
- `check/check.py`
- `/flag`
- `CHECKER_TOKEN`
- 其他平台运行契约、secret、临时文件和日志

这会把“防守入口”变成“平台把控制面秘密也带进学生作战面”。

结论不是“不要 SSH”，而是：

- 不能继续让学生直接进入持有 secret 的服务容器本体
- 不能继续把文件级修补边界下发给学生

## 3. 核心设计结论

最终采用以下边界：

1. 学生端默认不暴露 `editable_paths`、`protected_paths`、`service_contracts`、文件树、文件内容和受控片段。
2. 学生端保留战场态势、服务状态、打开服务、重启服务、最近事件和防守连接入口。
3. 防守连接入口不再进入服务容器本体，而是进入独立的 `defense workspace`。
4. `defense workspace` 只挂载业务工作区，不挂载平台运行契约、checker、flag 或 secret。
5. 平台限制的是“工作区边界”和“可写挂载”，不是“告诉学生哪个文件该改”。

一句话说：

AWD 防守应当设计成“进入本队独立防守工作区后自行侦察和修补”，而不是“平台直接给修补线索”。

## 4. 方案比较

### 方案 A：浏览器风险片段 / 受控 patch

做法：

- 平台展示漏洞相关片段
- 平台提供受控 patch 入口

优点：

- 平台控制力强
- 对低年级教学友好

缺点：

- 平台直接提供防守线索
- 更像教学修洞，不像攻防对抗
- 不适合作为毕设里的 AWD 主方案

不采用为默认竞赛模式。

### 方案 B：直接 SSH 到服务容器

做法：

- 平台只发 SSH 凭据
- 学生直接进入当前 service 容器

优点：

- 交互最简单
- 最接近传统运维入口

缺点：

- 无法可靠隔离 `ctf_runtime.py`、flag 和 checker token
- 文件级限制在同一容器 shell 中不是强安全边界
- 仍然容易把平台控制面秘密带给学生

不采用。

### 方案 C：独立防守工作区容器

做法：

- 平台为每个 `contest + team + service` 提供独立 `defense workspace`
- 学生通过 SSH/VPN 进入该工作区，而不是进入 service 容器本体
- service 容器与 workspace 通过共享业务卷协同

优点：

- 能把 secret 和运行契约留在 service 容器
- 学生仍能真实使用 shell、编辑器和本地工具
- 平台控制的是“工作区边界”，不是“答案文件”
- 最符合“赛中真实攻防、赛后教学复盘”的课题定位

缺点：

- 题目包结构和运行时装配需要调整
- 实现复杂度高于直接 SSH

推荐采用。

## 5. 推荐架构

### 5.1 三个边界

最终防守能力拆成三个清晰边界：

1. `battle workspace`
   只显示战场态势、攻击目标、本队状态、最近事件、连接入口和重启动作。

2. `defense workspace`
   学生真正进入的防守作业区，只包含业务代码、模板、静态资源、业务数据和必要日志。

3. `service runtime`
   平台正式裁判和对外服务所在的运行容器，持有 flag、checker token、运行契约和健康检查逻辑。

学生可以进入 1、2，但不能直接进入 3。

### 5.2 文件系统与挂载模型

推荐运行模型不是“一个容器里做路径过滤”，而是“按 root 粒度拆工作区挂载”：

- `workspace seed snapshot`
  - 来源于题包 `docker/workspace/`
  - 在实例首次 provision 时按 `defense_workspace` 声明的目录根拆分并初始化
  - 每个 root 应对应独立 volume 或等价独立挂载，不能退回成“单 volume + shell 限制”

- `service runtime container`
  - 只挂载题目声明的业务 roots
  - 业务根推荐统一映射到 `/workspace/...`
  - 运行契约位于镜像内或单独只读挂载，例如 `/opt/ctf/runtime`
  - flag、checker token、运行时 secret 只存在于该容器或其 runtime-only 挂载中

- `defense workspace container`
  - 只挂载 `defense_workspace.workspace_roots`
  - `writable_roots` 以 `rw` 挂载
  - `readonly_roots` 以 `ro` 挂载
  - 可选附加只读日志根，例如 `/workspace-logs`
  - 不挂载 `/opt/ctf/runtime`、`docker/check/`、flag path 和平台 secret

- `platform control plane`
  - 负责创建 workspace roots、runtime 容器、workspace 容器和访问凭据
  - 负责重启、reseed、审计、计分和复盘

### 5.3 容器拓扑与生命周期

这里采用的是独立 companion container，而不是“直接进 service 容器”或“跟 service 主进程绑死的 sidecar shell”。

推荐生命周期如下：

1. 平台为每个 `contest + team + service` 创建一组逻辑资源：
   - 一份 `workspace seed snapshot`
   - 一组 root 级业务挂载
   - 一个 `defense workspace container`
   - 一个 `service runtime container`
2. `defense workspace container` 与 `service runtime container` 共享同一组业务 roots，但不共享 runtime-only 资产。
3. 学生 SSH 永远落到 `defense workspace container`，不会直接落到 `service runtime container`。
4. 如果 `defense workspace container` 异常退出，平台可以基于同一组业务 roots 重建它，而不覆盖学生已经做出的业务改动。
5. `service runtime container` 可以在比赛中多次重启或重建；默认不连带销毁 `defense workspace container` 和工作区 roots。

这样做的核心收益是：

- 防守 shell 与平台 secret 物理分离
- 服务重启不打断学生持续修补
- 可写边界由挂载模型保证，而不是由 UI 提示保证

### 5.4 变更生效、重启与持久化语义

学生修改发生在共享业务 roots 中，正式裁判语义如下：

- 学生在工作区内完成的修改，写入的是共享业务挂载，不是 service 容器根文件系统。
- `service runtime container` 会看到同一组业务 roots 上的最新内容。
- 平台只保证“在执行 `instances/restart` 后，新 runtime 一定加载最新工作区内容”。
- 是否支持热加载属于题目自己的实现细节，不作为平台官方判题契约。

重启与清空边界要明确区分：

- `restart`
  - 只重建 `service runtime container`
  - 重新注入 runtime-only secret
  - 不重建 workspace roots
  - 不清空学生已改动的业务文件
  - 不默认销毁 `defense workspace container`

- `recreate / reset / reseed`
  - 重新从题包快照初始化 workspace roots
  - 清空学生已有业务改动
  - 重建 `defense workspace container`
  - 使旧 SSH 会话和旧连接票据失效

为了避免“运行中悄悄换题包”带来的不公平，题包 revision 更新不应热覆盖到正在比赛的实例：

- active contest 中已有实例默认继续绑定原 workspace snapshot
- 只有管理员显式 reprovision，或新一轮/新赛事重新发放实例时，才切换到新的 snapshot

### 5.5 关键原则

1. 学生修改发生在共享业务卷中，而不是直接改 service 容器根文件系统。
2. 平台契约文件与业务代码文件物理分离，避免“只要进容器就能看见 secret”。
3. 防守入口默认落在目录级工作区，不以单文件为粒度。

## 6. 题目包设计调整

防守边界要成立，题目包结构必须跟着调整。

### 6.1 不再推荐的结构

下面这种结构不适合作为最终 AWD 包约定：

```text
docker/
  app.py
  ctf_runtime.py
  challenge_app.py
```

原因：

- 业务代码和平台运行契约混在同一层
- 一旦给学生容器 shell，很难把 `ctf_runtime.py` 和 `challenge_app.py` 物理隔开
- `challenge_app.py` 作为单文件入口，本身就在暗示修补目标

### 6.2 推荐结构

推荐把运行契约层和业务工作区层拆开：

```text
<slug>/
├── challenge.yml
├── docker/
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

其中：

- `docker/runtime/*` 属于平台运行契约和服务装配层
- `docker/workspace/*` 属于学生可侦察、可修补的业务工作区
- `docker/check/*` 属于 checker 资产

### 6.3 defense_scope 的角色调整

现有 `defense_scope` 不能再作为学生端文件浏览契约使用。

后续应调整为：

- 只作为平台内部元数据
- 用于导入校验、运行时装配、审计和保护边界
- 不再直接下发给学生端

也就是说：

- 可以保留 `defense_scope` 作为兼容字段
- 但它不再驱动学生侧文件树、片段区或提示文案

### 6.4 新的题包字段建议

后续题包契约建议新增目录级字段，例如：

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

约束建议：

- 只允许目录级路径，不允许单文件路径
- `workspace_roots` 必须完整覆盖学生可见业务根，`writable_roots` 与 `readonly_roots` 必须是它的互斥划分
- 不允许把 `runtime/`、`check/`、`challenge.yml`、flag 路径放进工作区
- `runtime_mounts.source` 必须来自 `workspace_roots`
- `seed_root` 只能指向题包内业务工作区目录，不能复用 `docker/` 根目录
- `workspace_roots` 必须是“有多文件语义的业务目录”，不能继续用 `docker/challenge_app.py` 这种单文件答案入口

## 7. 学生端产品设计

### 7.1 战场页展示

学生战场页只展示：

- 当前轮次
- 本队服务状态
- 对手目标目录
- 最近被打 / 命中反馈
- 打开服务
- 重启服务
- 进入防守工作区

不展示：

- 文件树
- 文件路径
- 文件内容
- 平台推断的漏洞片段
- `editable_paths`
- `service_contracts`

### 7.2 防守入口文案

学生端入口不应命名成：

- 查看源码
- 防守文件
- 修补目标

推荐命名：

- 进入工作区
- 防守终端
- 防守连接

这样表达的是“给你作战环境”，不是“给你答案线索”。

### 7.3 浏览器侧非目标

默认竞赛模式下，不再提供：

- `GET /defense/directories`
- `GET /defense/files`
- `POST /defense/commands`
- 浏览器文件树
- 浏览器 patch 编辑器
- 浏览器命令执行器

如果未来要做“教学引导模式”，应作为单独模式，不与标准 AWD 主路径混用。

## 8. 后端与 API 方向

### 8.1 保留

- `GET /api/v1/contests/:id/awd/workspace`
  - 学生态只读取 `defense_connection` 摘要
  - query 阶段通过 `contest_awd_services`、`instances` 和 `awd_defense_workspaces` 聚合 `entry_mode / workspace_status / workspace_revision`
  - 不再把 `defense_scope`、文件路径或容器标识下发给学生端
- `POST /api/v1/contests/:id/awd/services/:sid/defense/ssh`
  - 语义改为：签发 `defense workspace` 的连接凭据
- `POST /api/v1/contests/:id/awd/services/:sid/instances/restart`
  - 继续作为平台防守动作

### 8.2 下线或停用学生侧文件接口

不再把下面这些接口作为学生默认防守能力：

- `GET /defense/directories`
- `GET /defense/files`
- `PUT /defense/files`
- `POST /defense/commands`

如果兼容历史代码暂时保留，也应：

- 默认关闭
- 不作为当前设计事实源
- 不在学生端挂入口

### 8.3 SSH 网关调整

当前 SSH 网关直接把学生带进 service 容器本体，这一做法需要替换。

后续 SSH 网关应改为：

- 票据仍以 `contest + team + service` 归属签发
- 票据还应绑定当前 `workspace revision`
- 解析后连接到对应 `defense workspace container`
- 会话默认工作目录为 `/workspace`
- 不提供 service 容器 rootfs
- `service runtime` 的普通重启不轮换 `workspace revision`
- `reseed / recreate` 必须轮换 `workspace revision`
- 用户侧拿到的是短时票据，不是长期固定容器口令

#### SSH host key 持久化

SSH 网关的服务端 `host key` 必须和短时票据分开处理。

- `host key` 表示网关自身身份，属于稳定入口契约的一部分
- 短时票据只用于当前用户的临时认证，不承担“服务端身份稳定”的职责
- 对固定入口 `host:port` 而言，如果 `host key` 在每次进程重启后都重新生成，客户端会持续触发 `REMOTE HOST IDENTIFICATION HAS CHANGED`

因此这里的正式约束是：

1. SSH 网关启动时优先从配置指定路径读取持久化私钥。
2. 当私钥文件不存在时，只允许首次生成一次新 key，并立刻写回该路径。
3. 写回后的私钥文件权限必须收敛到仅当前进程用户可读写，例如 `0600`。
4. 后续普通重启、代码热更新、runtime 容器重建、workspace 容器重建，都不得隐式轮换 `host key`。
5. 只有管理员显式替换 key 文件，或执行明确的 key rotation 流程时，`host key` 才允许变化。

配置与部署约束：

- 新增独立配置项，例如 `container.defense_ssh_host_key_path`
- 默认值可落在后端本地持久化目录，例如 `storage/runtime/awd-defense-ssh-host-key.pem`
- 单机开发环境允许使用本地文件
- 如果未来同一个 `defense_ssh_host:port` 背后接入多个 API 实例，这些实例必须共享同一份 host key，而不是各自生成

故障与回退语义：

- 如果 `defense_ssh_enabled=true` 且配置路径上的 key 文件存在但不可解析，启动应直接失败，而不是静默回退到重新生成
- 如果目录不可写且文件又不存在，启动同样应失败，因为这意味着服务端身份无法稳定落盘
- 删除 key 文件等价于显式轮换，客户端 `known_hosts` 需要重新确认，这是可预期的运维行为，不应伪装成普通重启

### 8.4 审计语义

平台应记录“防守动作元数据”，但不应把工作区内的一切都伪装成官方裁判证据。

建议记录：

- workspace provision / recreate / reseed
- SSH 票据签发
- SSH 会话建立 / 关闭
- 学生触发的 restart 请求及其结果
- 访问失败、权限拒绝和异常重建

可以作为复盘辅助，但不应作为官方裁判依据的内容：

- 学生每一次文件编辑细节
- 学生 shell 中的完整命令历史
- 本地编辑器产生的临时文件、swap 文件、backup 文件
- “谁先改了哪一行”这类细粒度文件操作轨迹

正式裁判证据仍应以这些数据为准：

- attack 提交结果
- checker 结果
- 平台代理流量与状态记录

## 9. 安全与公平性原则

### 9.1 应限制什么

平台真正需要限制的是：

- 学生不能进入别队工作区
- 学生不能进入 service runtime 契约层
- 学生不能接触 flag、checker token 和平台 secret
- 学生不能通过平台入口直接拿到“该改哪个文件”的提示

### 9.2 不应限制成什么

平台不应把限制做成：

- 只给一个 `challenge_app.py`
- 只给几个平台挑好的片段
- 只允许学生看到“正确修补文件”

因为这样虽然更可控，但已经从攻防对抗退化成“带答案范围的修补训练”。

### 9.3 直接 shell 的边界

如果学生进入的是 service 容器本体，那么：

- 路径过滤不是强安全边界
- shell 内文件访问限制很容易和真实运维、真实程序调试相冲突
- secret 泄露风险会持续存在

所以真正的解法不是“在同一容器里把 shell 再限制细一点”，而是“不要把学生放进那个容器”。

## 10. 教学与论文表述

这套设计更适合你在论文中这样表述：

- 赛中：平台提供战场态势、连接入口、服务重启和官方裁判反馈，学生通过防守工作区与本地工具完成真实修补。
- 赛后：平台基于攻击日志、流量摘要、checker 结果和防守动作记录进行教学复盘。

不建议写成：

- 平台赛中自动指出漏洞位置
- 平台赛中提供完整源码修补指导
- 平台完整记录学生每一次防守文件操作并作为官方裁判依据

## 11. 迁移要求

要从当前实现迁到这套设计，至少需要完成：

1. 学生侧移除 `defense_scope` 展示与消费。
2. 学生侧不再挂接浏览器文件树与文件内容页。
3. 平台新增 `defense workspace` 状态与 `workspace revision` 持久化。
4. SSH 网关从 service 容器切到 `defense workspace` 容器。
5. `restart` 只重建 runtime，不清空 workspace；`reseed` 才重置工作区。
6. AWD 题包从“单文件业务入口”重构到“目录级业务工作区”。
7. `ctf_runtime.py`、checker、flag 存储和 secret 不再进入学生作战工作区。

## 12. 验收标准

- 学生端看不到 `editable_paths`、`protected_paths`、`service_contracts`、文件树和文件内容。
- 学生通过平台拿到的防守连接入口默认落在独立工作区，而不是 service 容器本体。
- 工作区的可写、只读和不可见边界由挂载模型保证，而不是只靠前端隐藏。
- 学生工作区中不存在 `/flag`、`CHECKER_TOKEN`、`ctf_runtime.py`、`check/check.py` 等受保护资产。
- 普通 `restart` 后学生已修改的业务文件仍然保留；只有 `reseed / recreate` 才会清空。
- `reseed / recreate` 后旧 SSH 票据和旧工作区会话失效。
- 学生仍能通过工作区自行定位问题、修改业务代码，并通过平台重启与 checker 反馈验证结果。
- 平台对“防守”提供的是受控接入面，而不是修补答案提示。
