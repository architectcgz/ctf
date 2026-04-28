# CTF 网络攻防靶场平台 — 难点分析与解决方案（实现向）

> 版本：v1.0 | 日期：2026-03-03 | 状态：草稿
> 适用范围：`ctf/` 项目（Go + Docker + PostgreSQL + Redis / Vue 3）
> 参考：`ctf/docs/architecture/backend/*`、`ctf/docs/architecture/frontend/*`、`ctf/docs/tasks/ctf-platform-task-breakdown.md`

---

## 0. 目标与边界

本文档不做新的架构决策，只把现有设计文档中隐含的“实现难点”显式化，并给出可执行的落地方案与检查清单，方便：

- 做任务拆分与验收时快速定位风险点
- 实现过程中避免典型安全与稳定性回归
- 评审时用统一语言对齐“做到什么程度算完成”

---

## 1. 系统级难点：靶机容器是“带漏洞的服务”，平台必须更安全

### 难点 1.1：安全边界与网络隔离落地

靶机容器天然可被选手拿到任意代码执行/提权（题目需要），所以隔离失败的后果通常是：

- 选手通过靶机攻击平台 API/DB/Redis（篡分、读题库、偷 flag secret）
- 选手在同宿主机横向攻击其他用户靶机（泄漏/抢答/干扰）
- 容器逃逸打到宿主机（最坏情况）

**解决方案（落地要点）**

- 网络隔离模型：默认“每个实例独立 network”，并确保靶机网络与平台服务网络不可达。
- 仅暴露必要端口到宿主机：动态端口映射，避免固定端口被扫描利用或冲突。
- 禁止靶机访问宿主机管理面：阻断到宿主机 IP（包括 docker bridge 网关）以及 Docker API/SSH 等端口的流量。
- AWD 模式例外：同队伍共享 network（按设计文档的例外规则），但仍需与平台服务网隔离。

**验收/检查清单**

- 容器 A ping 不通容器 B（不同用户/不同实例）。
- 容器内无法访问平台服务（API/PG/Redis 所在网段或容器）。
- 容器内无法访问宿主机关键端口（SSH、Docker API、数据库端口等）。

对应设计：`ctf/docs/architecture/backend/01-system-architecture.md` “安全边界设计/网络隔离模型”。

### 难点 1.2：容器资源限制与宿主机稳定性

靶机容器可被故意压测（fork bomb、内存爆、磁盘写满、CPU 忙循环），必须保证：

- 单容器失控不影响平台与其他容器
- 宿主机资源可控，竞赛期间不会“雪崩”

**解决方案（落地要点）**

- 强制资源上限：CPU、内存、pids、存储（以及可选的 IO/带宽）。
- 强制禁用 `--privileged`，capabilities 走最小集合（优先 Drop ALL）。
- 以非 root 用户运行靶机容器；需要 root 的题目必须显式标注并隔离更强的运行参数（实现时作为例外处理）。
- 只读根文件系统（按需用 tmpfs 提供可写目录）。

**验收/检查清单**

- fork bomb 仅导致容器内失败或被 kill，宿主机可用。
- 内存超限触发 OOM kill，不影响 API/DB。
- 容器无法获得超出预期的 capabilities（最小化白名单）。

对应设计：`ctf/docs/architecture/backend/03-container-architecture.md`、任务 T6/T7（容器隔离）。

---

## 2. 业务级难点：Flag 的生命周期、唯一性与反作弊

### 难点 2.1：Flag 策略要同时满足“教学可用”和“竞赛安全”

设计中包含三类题目：`static` / `dynamic` / `manual`。实现难点在于：

- 不能把“正确答案”暴露给题目容器（容器可被拿下）
- 动态 Flag 要做到“每用户/每实例唯一”，同时校验可重算、无需落库存明文
- 竞赛需要封榜、回放与审计，Flag/提交记录要可追溯

**解决方案（落地要点）**

- 静态 Flag：服务端存储（DB + 缓存），校验在服务端完成；避免写入前端或容器镜像。
- 动态 Flag：按设计采用 HMAC（global secret + contest salt + instance nonce）确定性生成，服务端校验时重算比对。
- contest salt 等敏感字段：加密存储，解密密钥来自环境变量（禁止写入日志/数据库明文字段）。
- Flag 格式统一并正则校验，先过滤非法格式再进入校验/计分逻辑，降低攻击面。

**验收/检查清单**

- 同一用户同一题目的不同实例，Flag 不同。
- 不同用户同一题目同一实例，不存在（实例归属唯一）。
- 题目容器内无法推导出其他用户的 Flag（关键 secret 不在容器侧）。

对应设计：`ctf/docs/architecture/backend/01-system-architecture.md` “动态 Flag 生成策略（ADR-004）”。

### 难点 2.2：Flag 提交的并发、幂等与防刷

典型坑：

- 并发提交导致重复计分（或重复插入 solve 记录）
- 暴力猜 flag/撞库（需要限流、锁与审计）
- 错误返回包含过多内部细节（泄漏题目判题逻辑/盐/密钥线索）

**解决方案（落地要点）**

- 提交链路必须幂等：以 `(user_id, challenge_id)` 或 `(team_id, contest_challenge_id)` 作为唯一约束，重复提交返回“已解出”而不是再次计分。
- 防刷：滑动窗口限流（按用户+题目），并在服务端记录每次提交结果用于风控/审计。
- 并发保护：提交入口加短 TTL 分布式锁，或在 DB 侧用唯一约束 + 事务保证最终一致。
- 错误文案：对外只给可理解提示 + request_id，避免透传内部异常栈/明文配置。

对应设计：`ctf/docs/architecture/backend/04-api-design.md`、`ctf/backend/internal/pkg/redis/keys.go`（提交限流/锁的 key 设计）。

---

## 3. 工程级难点：容器实例生命周期与端口分配

### 难点 3.1：实例创建/销毁需要“强一致的状态机”

实例系统通常是事故高发区：创建超时、容器实际已起但 DB 标记失败、销毁失败遗留网络/端口占用等。

**解决方案（落地要点）**

- 把实例生命周期显式建模：`pending -> running -> stopping -> stopped -> deleted`（或设计文档中的状态），每个状态迁移要可重试。
- 创建流程建议拆为：预占端口/资源 -> 创建容器与网络 -> 启动 -> 健康探测（可选）-> 标记 running。
- 失败要可回滚：任何一步失败都要进入“补偿删除”流程（删除容器、删除网络、释放端口占用、回滚计数器）。
- 定时回收：TTL 到期扫描 + 强制删除，保证最终清理。

对应设计：`ctf/docs/architecture/backend/05-key-flows.md`、任务 T5（生命周期管理）。

### 难点 3.2：动态端口分配的冲突与安全

动态端口一旦出错，常见症状是“偶发无法访问”或“串线访问到别人的容器”。

**解决方案（落地要点）**

- 端口分配必须可追踪：端口范围配置化，落库记录，回收可验证。
- 避免并发冲突：端口分配需加锁（Redis 分布式锁或原子集合/位图），或用数据库唯一约束兜底。
- 端口暴露最小化：只映射题目需要的端口，避免把管理端口、调试端口暴露到宿主机。

对应设计：`ctf/docs/architecture/backend/03-container-architecture.md`、任务 T7（网络隔离）。

---

## 4. 鉴权与前端安全：Token 与富文本是典型高危点

### 难点 4.1：前端 XSS 直接升级为账号接管

CTF 平台需要展示题目描述/公告/通知等富文本内容，若渲染不做 sanitize，任意 HTML/脚本注入即可窃取 token 或劫持操作。

**解决方案（落地要点）**

- Markdown/富文本渲染：必须走白名单 sanitize，禁用事件属性与危险 URL scheme。
- 不把认证 token 暴露给前端：浏览器只保留 HttpOnly session cookie；前端通过 `/auth/profile` 恢复登录态。
- 前端错误提示不要透传后端 message（避免内部信息泄漏），统一映射错误码。

对应设计：`ctf/docs/architecture/frontend/01-architecture-overview.md` “安全基线（必须）”。

### 难点 4.2：WebSocket 认证与泄漏面

竞赛/排行榜/通知常用 WebSocket 推送，但认证如果用 query ticket，容易在代理/日志链路泄露。

**解决方案（落地要点）**

- ticket 一次性、短 TTL；网关/反代禁止记录 querystring（或改为 header/cookie 认证）。
- 断线重连要做节流与指数退避，避免网络抖动时雪崩式重连。
- 服务端要做背压与连接上限控制，防止单用户/单 IP 打爆推送通道。

对应设计：`ctf/docs/architecture/frontend/05-websocket-composables.md`、后端 WS 设计章节（如有）。

---

## 5. 数据一致性与可审计：竞赛结算靠“可回放的事实表”

### 难点 5.1：计分、封榜与复盘

竞赛期间常见诉求：

- 封榜前后分数展示逻辑不同
- 有争议时需要复盘“谁在什么时候提交了什么，为什么判错/判对”

**解决方案（落地要点）**

- 以“提交记录/解题记录”为事实表，全量落库，不依赖缓存作为事实来源。
- 排行榜可缓存（Redis ZSET 等），但要能从事实表重建。
- 封榜生成快照并保留快照来源（截止时间点），防止赛后无法解释争议。

对应设计：`ctf/docs/architecture/backend/02-database-design.md`、`ctf/docs/architecture/backend/05-key-flows.md`。

---

## 6. 运维与事故预案：竞赛窗口里只允许“自动恢复”

### 难点 6.1：竞赛期间的可用性保障

竞赛对“可用性”的敏感度高于日常训练，手工干预越多越容易二次事故。

**解决方案（落地要点）**

- 健康检查与自动重启：systemd/进程守护 + `/health` 探测。
- 日志结构化并带 request_id，关键操作（登录/权限变更/题目修改/容器操作）写审计日志。
- 资源监控：容器数量、失败率、回收延迟、端口耗尽、Redis/PG 连接池等指标要可观测。

对应设计：`ctf/docs/architecture/backend/01-system-architecture.md`（可用性目标、审计日志、健康检查）。

---

## 7. 面向任务的“高风险点”提醒（按交付链路）

本节用于评审/实现时快速点名风险，避免“功能做完了但一上线就出事故”。

- T3（认证/RBAC）：服务端 session 存储、强制下线、登录失败锁定、越权访问回归测试。
- T5~T7（容器引擎）：创建/销毁幂等、超时补偿、网络隔离验证、端口分配冲突与回收。
- T8~T11（靶机管理）：题目发布前校验、附件上传安全（路径穿越/类型伪造/超大文件）、题目描述富文本 XSS。
- Flag 提交与计分：并发重复计分、限流生效、错误信息泄漏、审计数据完整性。

---

## 8. 开发态自检建议（最小集合）

不引入新工具的前提下，建议在每次改动关键链路后至少做以下自检：

- 容器隔离：起两个实例互相 `ping`/`curl` 验证隔离规则。
- 提交幂等：并发请求提交同一 flag，多次响应不应重复计分。
- XSS 基线：在题目描述插入典型 payload，前端必须被 sanitize 掉且不影响正常 Markdown。

---

## 9. 设计缺口（待确认）：教师上传题目格式与比赛题目格式

你提到的两个“格式未定”，属于会向下游扩散的关键设计缺口。在未确定前，相关实现建议暂停到“接口/数据模型/校验规则”层面，避免返工。

### 9.0 建议结论（先给默认选型）

- 教师上传题目：采用 **单文件 Zip 题目包**（`challenge-pack-v1`），包内用 **YAML manifest + 资源文件** 描述题目元数据、附件、容器运行参数与可选构建方式。
- 比赛题目：采用 **YAML 竞赛定义**（`contest-spec-v1`）引用题目（`challenge_id` 或 `challenge_slug`），并允许“竞赛内覆写”（分值、开放时间、flag 模式、实例策略等）；同时明确支持 **队伍对抗**：
  - Jeopardy 团队赛（提交与计分归属到 team）
  - AWD（Attack With Defense）团队对抗（轮次、flag 轮转、服务存活检测、攻防计分）

这样做的核心原因是：Zip 包便于版本化与离线制作，manifest 便于静态校验；竞赛定义与题目分离，避免“把竞赛配置写进题目包”导致复用困难与数据污染。

### 9.1 教师上传题目格式（导入/发布格式）

**直接影响**

- 后端：题目创建/更新 API、附件上传、镜像构建与校验、flag 策略配置、题目版本管理与审核流（若有）
- 前端：教师/管理员题目创建页的表单结构、校验规则、错误提示与预览能力
- 运维：导入失败的可诊断性（日志/构建日志/校验错误）、大文件与并发导入的资源保护

**必须拍板的问题清单（建议先回答这些）**

- 载体是什么：单个 `zip` / 单 `yaml+assets` / 纯表单（无文件包）/ 与镜像构建分离？
- 题目元数据最小字段：`title`、`category`、`difficulty`、`tags`、`description(md)`、`hints`、`attachments`、`author`、`version` 是否必填？
- flag 配置形态：`static/dynamic/manual` 三类如何在格式里表达？是否允许题目自带 `flag`（强烈倾向服务端配置而不是包内明文）？
- 容器题目如何描述运行参数：暴露端口、健康检查、资源配额、启动命令、只读/非 root 等安全参数是“题目包里声明”还是“平台侧统一模板+少量可配项”？
- 版本与兼容：同一题目多版本如何区分？是否支持“替换版本但保留历史提交记录/结算规则”？
- 校验与审核：导入前校验哪些规则（镜像存在、端口合法、描述字段长度、附件大小/类型）？是否需要管理员审核后才能发布？
- 附件安全：允许的文件类型白名单、单文件/总大小上限、是否做解压炸弹防护、是否做路径穿越防护（zip slip）？

**建议的验收口径（先写成 checklist，格式定了再补细节）**

- 导入失败能给出“可定位到字段/文件”的错误（而不是泛化 500）。
- 任何题目包内容都不能让选手侧推导平台密钥或其他题答案（flag/secret 不落包）。
- 同一题目重复导入的幂等策略明确（覆盖/新建版本/拒绝）。

#### 推荐格式：`challenge-pack-v1`（Zip 题目包）

规范文档：

- `ctf/docs/contracts/challenge-pack-v1.md`
- `ctf/docs/architecture/backend/06-file-storage.md`（文件存储与导入/下载安全约束）

文件结构建议（示例）：

```text
web-sqli-01.zip
  challenge.yml
  statement.md
  attachments/
    hint.png
    data.sql
  docker/                # 建议必带：用于复现/审计（不等于平台必须在线构建）
    Dockerfile
    src/...
```

`challenge.yml` 最小字段建议（示意，不是最终 schema）：

```yaml
api_version: v1
kind: challenge
meta:
  slug: web-sqli-01
  title: "Web-01 SQLi 入门"
  category: web
  difficulty: easy
  tags: ["sqli", "mysql"]
content:
  statement: statement.md
  attachments:
    - path: attachments/data.sql
      name: "data.sql"
flag:
  type: dynamic            # static | dynamic
runtime:
  type: container
  image:
    ref: "ghcr.io/ctf/web-sqli-01:1.0.0"
```

仓库内示例题目包（未压缩形态）：

- `ctf/docs/contracts/examples/challenge-pack-v1/web-hello-01/`（包含 `challenge.yml`、`statement.md`、`docker/Dockerfile`）
- `ctf/docs/contracts/examples/challenge-pack-v1/web-sqli-login-01/`（更接近真实题：SQL 注入读取 `secrets` 表中的动态 Flag）

强制校验建议（安全与运维优先）：

- Zip 解包：拒绝绝对路径/`..` 路径穿越、拒绝 symlink、限制文件数/总大小/单文件大小，防止 zip bomb 与写任意路径。
- `attachments`：要求 sha256，导入时校验一致性；导出/下载只允许在已知目录内读取。
- `runtime`：建议支持三种来源，但要求“至少选一种且只能选一种”：`registry(ref)` / `dockerfile(context)` / `tar(image)`。
- `runtime(dockerfile)`：不建议“教师上传即同步构建并发布”。推荐走异步构建队列，且构建必须在隔离的 builder 上执行（资源限额 + 超时 + 无平台内网访问），否则会显著扩大供应链风险与资源消耗。
- 如果你们希望“题目包必须自带 Dockerfile”来保证可复现：建议把 Dockerfile 作为源码与审计材料，但运行产物仍要求提供 `registry(ref)` 或 `tar(image)` 之一，平台默认只做 pull/load 而不做 build（把 build 风险留在受控流水线或教师本地）。
- `flag`：强烈建议竞赛与在线靶机一律用 `dynamic/manual`，避免静态 flag 被转发共享。

### 9.2 比赛题目格式（Contest Challenge 定义）

这里的“比赛题目格式”建议明确是：

- 竞赛如何引用练习题（`challenge`）并形成竞赛内的 `contest_challenge`（包含竞赛专属配置，如分值、开放时间、动态 flag salt、解题上限、队伍模式等）。

**直接影响**

- 数据模型：`contest_challenge_id` 与 `challenge_id` 的关系、竞赛专属字段（分值、开放/关闭、封榜、队伍/个人）
- 计分与提交：同题在不同竞赛中的计分是否独立、封榜快照、作弊检测维度
- WebSocket 推送：竞赛题目列表/状态/公告/排行榜的推送 payload

**必须拍板的问题清单**

- 竞赛引用题目的粒度：是否允许同一 `challenge` 在同一竞赛内出现多次（不同配置）？
- 计分模型：固定分还是动态分（随解出人数衰减）？封榜前后显示规则是什么？
- 作答主体：个人赛/组队赛/混合？若组队，flag 校验与提交归属到 `team_id` 还是 `user_id`？
- 题目可见性：是否支持按时间/阶段解锁？是否支持赛中临时下线/替换题目与对应回放规则？
- 实例策略：竞赛题是否必须动态 flag？是否允许每队共享实例还是每人实例？实例 TTL/并发限制是否与练习模式一致？
- 反作弊与审计：需要记录哪些事件用于争议仲裁（提交原文、判题结果、请求 ID、IP/UA、容器实例 ID 等）？

**建议的最小验收口径**

- 竞赛内的题目配置是“竞赛专属”的，不会污染练习题配置。
- 竞赛结束后可从事实表重建排行，封榜快照可解释争议。

#### 推荐格式：`contest-spec-v1`（YAML 竞赛定义）

将“竞赛”视为对题目的引用与覆写（而非复制），避免竞赛配置污染题库。

Jeopardy 团队赛（示意）：

```yaml
spec_version: contest-spec-v1
mode: jeopardy_team
slug: "spring-ctf-2026"
title: "2026 春季校园 CTF 挑战赛"
time:
  start_at: "2026-04-01T12:00:00+08:00"
  end_at: "2026-04-01T18:00:00+08:00"
team:
  enabled: true
  size_min: 1
  size_max: 4
scoreboard:
  freeze_at: "2026-04-01T17:00:00+08:00"
challenges:
  - challenge_slug: web-sqli-01
    points: 100
    open_at: "2026-04-01T12:00:00+08:00"
    flag:
      mode: dynamic
    instance:
      max_per_team: 1
      ttl: "2h"
```

AWD 团队对抗（示意，二期/三期落地更合理）：

```yaml
spec_version: contest-spec-v1
mode: awd
slug: "awd-finals-2026"
title: "AWD 总决赛"
time: { start_at: "...", end_at: "..." }
team: { enabled: true, size_min: 3, size_max: 6 }
awd:
  round_duration: "5m"
  flag_rotation: "1m"              # 每轮/每分钟轮换（按你们规则）
  service_check:
    interval: "15s"
    timeout: "3s"
  services:
    - service_slug: "pwn-heap-01"
      image_ref: "ghcr.io/ctf/pwn-heap-01:1.0.0"
      expose: [{ container_port: 9999, protocol: tcp }]
      resources: { cpu: 1.0, memory: "512m" }
```

必须拍板的“对抗规则”（否则格式定了也无法实现）：

- 攻击得分来源：提交对手 flag、利用成功回执、还是以“夺旗记录”为准？
- 防守扣分/加分：服务存活（check）如何计分？被攻破是否扣分？
- flag 轮转策略：按轮次、按时间、按服务实例？过期判定与容错窗口是多少？
- 隔离策略：同队共享网络/是否允许互访；跨队强隔离与端口暴露规则。

> 这个缺口建议先产出一个“格式草案 + 3~5 条样例”，再反推 API 合同与前端表单结构；否则实现阶段会在字段命名、必填/可选、兼容上反复返工。
