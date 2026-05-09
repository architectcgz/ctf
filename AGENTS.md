## Project Harness Intake

- 本仓库默认先进入 harness：除非任务明显简单、局部、可逆且不需要沉淀经验，否则开始前必须先按 `harness-router` 判断 `SIMPLE` / `HARNESS`。
- 路线为 `HARNESS` 时，先读本文件和相关 harness 入口，再决定是否需要计划、review、验证或更新 `feedback/`、`prompts/`、`references/`、`works/`。
- 可复用提示词见 `prompts/harness-router.md`；机械化检查使用 `bash scripts/check-consistency.sh`。
- 对 `API / filter / sort / pagination` 契约改动，harness 的 plan review 必须显式写清 `normalize / default / validate` 的唯一 owner，至少覆盖 `handler -> application -> repository` 三层；若 owner 不唯一，不得直接进入实现。
- 对同一输入语义出现跨层重复 `normalize/default/validate` 的情况，默认不视为“安全兜底就可以接受”。除非其中一层承担明确的 trust-boundary 防御且理由写清，否则应在实现阶段继续收口成单点 owner。
- harness 的 review gate 必须把 touched surface 上的跨层重复归一化、双重默认值兜底、repo 接收未收敛裸字符串排序键这类问题视为结构性 blocker，而不是普通建议或事后优化项。
- 对内部 `filter / sort / pagination` contract，单点 owner 还不够；plan review 还必须说明“收口后的语义如何变成 downstream 不易误用的表示”。默认优先 opaque value object、受控构造器或不可直接手工拼装的字段，而不是保留宽松导出 enum / struct 再指望下游自己别用错。
- 若 touched surface 仍允许调用方手工构造无效内部状态，并且要靠 repository 的 panic、fallback 或 defensive branch 才能发现，review gate 视为 contract 仍未收口，不得当作已完成的结构性修复。

## Design Context

### Users
这是一个学校使用的 CTF 平台，核心用户包含学生、教师和管理员，但前端设计优先级以学生刷题体验为先。

- 学生的主要任务是浏览题目、启动靶机、提交 Flag、查看题解与复盘自己的解题过程。
- 教师的主要任务是查看学生进展、分析训练表现、筛选题解与进行教学干预。
- 管理员的主要任务是维护题目、竞赛、用户和平台运行秩序。

设计决策默认优先保证学生端的信息获取效率、做题专注度和操作连贯性，再兼顾教师与管理员的管理效率。

### Brand Personality
品牌气质关键词：技术、专业、克制。

- Voice / Tone：冷静、可信、清晰，不做夸张情绪化表达。
- Emotional Goals：让学生感到专注、可靠、可控，让教师感到信息清楚、判断高效，让管理员感到系统稳健、秩序明确。
- 产品不追求“娱乐化竞赛感”，而是强调教学场景中的技术训练平台属性。

### Aesthetic Direction
整体视觉方向：极简专业。

- 参考方向：
  - https://ctf.bugku.com/game/index.html
  - https://www.ctfplus.cn
  - https://www.ctfhub.com/
- 默认主题：light 优先，同时支持 dark 切换。
- 应保持技术平台气质，但避免“安全圈站点模板味”与装饰性堆砌。

明确的反方向：
- 不要二次元感
- 不要游戏商城感
- 不要炫酷霓虹
- 不要企业 OA 感

### Design Principles
1. 学生刷题优先
页面结构、信息层级和交互反馈优先服务学生的连续做题流程，减少打断与噪音。

2. 极简但不空
优先使用清晰的排版、留白、分隔线和层级组织信息，而不是依赖堆叠卡片、重阴影和装饰性容器。

3. 技术感来自秩序，不来自特效
技术气质应通过严谨的布局、稳定的对齐、克制的色彩和明确的状态表达建立，而不是通过霓虹色、夸张渐变或“黑客风”视觉符号制造。

4. 轻量教学场景
教师和管理员相关界面应体现分析、判断、筛选和治理效率，避免做成通用 OA 或后台模板风格。

5. 可操作性优先于形式
所有关键流程默认要求键盘可操作，交互状态必须清晰，重要按钮和切换控件要具备明确的可达性和反馈。

### Copy Preferences
前端文案需要区分“结构性标识”“功能性说明”和“设计介绍式文案”，默认按以下规则处理：

- 保留结构性小标题/eyebrow。
  例如 `Progress Signal`、`Trend Review`、`Today Focus` 这类用于标识分区结构和阅读层级的短标签，应保留。
- 保留功能性说明。
  例如空状态说明、状态提示、数据上下文说明、面板副标题、字段含义解释，这些直接帮助用户理解当前内容或下一步操作，应保留。
- 移除设计介绍式文案。
  例如解释布局意图、说明“这一块为什么这样设计”、描述“帮助教师快速判断下一步教学策略”这类偏设计汇报口吻的句子，默认不应出现在正式页面中。
- 优先短句和结果导向表达。
  页面文案应直接服务当前任务，不写“为你提供”“帮助你理解”“在同一块区域里”这类解释 UI 组织方式的句子，除非它本身是必要操作说明。
- 组件默认文案也遵守同一规则。
  `modal / drawer / popover / empty state / card` 的默认 `subtitle`、`description`、`helper`、`placeholder` 不得写“在这里放置…”、“该区域用于…”、“支持长表单…”这类实现说明；默认应为空，只有调用方显式传入真实业务文案时才显示。

### Frontend Guardrails
- 实现型 agent（如 `backend-engineer`、`frontend-engineer`、`code-agent`）在提交代码时，凡是涉及锁续租、并发收敛、状态机分支、跨上下文释放资源、兼容性兜底这类不看上下文就容易误判的关键实现点，必须补上简洁注释，说明“为什么这样做”，不能只让读代码的人自行猜测。
- 避免复用容易受全局样式影响的通用类名，尤其是类似 `.overline` 这类在多个页面重复出现的样式标识；新页面优先使用局部命名，减少样式串扰。
- 暗色模式兼容不是收尾项。凡是 `input / select / textarea / 搜索框 / 筛选控件` 这类表单容器涉及边框、背景、内高光、placeholder、caret、focus ring、adornment 时，禁止写死 `white`、浅色 `rgba(...)` 或只在 light mode 成立的 `inset` 高光；必须走主题语义变量，并在 dark mode 下逐项检查文字、光标、分隔线、内阴影是否仍然可见且不会产生白线。
- 教师端页面允许保留必要的教学语义和状态说明，但不应出现产品设计说明、布局介绍或“workspace 理念解释”类文案。
- 模板组件不得自带会直接渲染到 UI 的脚手架说明文案。
  通用 `modal / drawer / panel` 若需要 `subtitle`、`description` 或 helper 区域，默认只提供结构，不提供演示型说明文字；示例文案只能留在测试、文档或注释里，不能作为运行时默认值进入页面。
- 前端实现禁止在同一个 `.vue` 页面文件里持续堆叠路由状态、接口调用、派生数据、交互流程和大段模板判断。
  当页面已经承担 2 个以上独立职责时，应优先拆分为 composable 或子组件，而不是继续把逻辑塞回当前文件。
- 对于以下类型的逻辑，默认优先考虑抽成 composable：
  - 路由级 tab / panel 的 query 同步与键盘切换
  - 页面级异步加载、提交、轮询、下载、保存等任务流
  - 面向模板的大量 `computed` 派生数据和聚合统计
  - 可复用的数据整形、筛选、排序、映射逻辑
  - 一组彼此关联的状态机式交互，例如题解切换、写作区保存、审核流、导出流
- 大文件不等于一定要拆，但必须先判断“大”的来源：
  - 如果主要是样式或长模板，优先拆子组件或共享样式
  - 如果主要是 `<script setup>` 中的状态、`computed`、`watch`、事件处理和接口编排，优先拆 composable
- 当已有页面文件超过约 `500` 行，或 `<script setup>` 已明显变成“页面控制器”时，新增功能前必须先评估是否应该抽离。
  不允许在 `700+` 行的页面文件中继续直接追加一整组新状态和流程，除非能够说明这段逻辑只会在当前文件出现且拆分反而更差。
- composable 的职责要按“一个页面能力域”切分，不要把所有页面杂糅成新的万能 util。
  推荐拆分方向：
  - `useXxxTabs`
  - `useXxxDetail`
  - `useXxxActions`
  - `useXxxMetrics`
  - `useXxxPreview`
- 优雅实现优先于表面复用：
  - 不为减少几行代码而抽象
  - 不把简单展示逻辑强行搬进 composable
  - 只在能明显降低页面认知负担、测试复杂度和回归风险时才提取
- 高复杂度页面应按职责优先拆分，避免在同一文件持续累积路由状态、数据编排和交互流程；优先抽离为 composable 或子组件后再扩展功能。

### Admin Directory Spacing Rules
- 对管理员端页面，凡是“目录标题 + `WorkspaceDirectoryToolbar` + 列表/表格/空状态/分页”这一类连续结构，垂直节奏默认由页面自己的目录 section 统一控制，不要让 section `gap` 与 toolbar 自带的 `margin-bottom` 叠加。
- 默认做法：
  - 目录 section 使用 `display: grid; gap: var(--space-4);`
  - 目录标题头部自身 `margin-bottom: 0;`
  - 页面内对 `WorkspaceDirectoryToolbar` 做局部收口：`:deep(.workspace-directory-toolbar) { margin-bottom: 0; }`
- 当确实需要更松的目录节奏时，优先调整 section 的 `gap`，不要重新放大 toolbar 的 `margin-bottom`。
- 新增或重构管理员端列表页时，需优先检查 `users / challenges / contests / classes / students / instances / audit-log / images` 这类目录页的现有节奏，保持一致后再提交。

### Frontend Workflow Preference
- 本节为 `ctf` 仓库内的覆盖规则；命中本节条件时，优先于上层“实现类改动默认使用 `git worktree add`”的通用约定。
- 当用户要求“更新前端 UI / 页面样式 / 视觉优化”时，默认直接在 `main` 分支修改，不创建新分支、不创建新 worktree。
- 只有用户明确要求“开新分支”“开 worktree”或“隔离开发”时，才切换到分支 / worktree 开发流程。
- 若本地已启动前端开发服务（如 `npm run dev`），默认复用当前服务以便用户立即看到页面变化，不主动中断。

### Branch Workflow Preference
- 本节为 `ctf` 仓库内的覆盖规则；命中本节条件时，优先于上层“实现类改动默认使用 `git worktree add`”的通用约定。
- 当用户没有明确指定“新分支 / worktree / 隔离开发”时，默认直接在当前主工作区的 `main` 分支修改。
- 即使任务较大、周期较长，若用户没有明确要求，也不要自行创建新分支或新 worktree。
- 只有用户明确要求开分支、开 worktree 或隔离开发时，才创建新分支或新 worktree。
- 若当前仓库实际不在 `main`，需要先说明当前所在分支，再按用户意图继续执行，避免无提示切到临时分支。

### Documentation Workflow Preference
- 纯文档编辑默认直接在当前分支修改，不创建新分支、不创建新 worktree。
- 这里的“纯文档编辑”仅指说明文档、架构文档、设计文档、README、注释性契约说明等不影响运行产物的文本改动。
- 即使文档修改需要和实现代码、脚本、配置变更一起提交，若用户没有明确要求，也默认不使用 worktree。
- 只有用户明确要求隔离开发、开分支或开 worktree 时，才使用 worktree。
- 不要把“文档也可能并发修改”当作默认进入 worktree 的理由；文档类任务默认走轻量流程。

### File Placement Rules
- 文档写作、目录归属、命名和验证细则以 `docs/文档规范.md` 为准；本节只保留 agent 执行时必须快速可见的归属规则。
- 规范化存量架构文档时，必须按 `docs/文档规范.md` 的“架构文档规范化流程”执行：先收口入口和状态，再按主题小提交处理正文、移动和删除。
- 生成架构图或让模型根据文档生成图片前，必须按 `docs/文档规范.md` 的“架构图生成规范”执行，先明确图表类型、事实来源、节点/边依据和可 review 的文本源。
- 文档、图表、prompt 输入包和 review 记录中的代码路径默认使用仓库根相对路径；敏感配置、token、cookie、密码和生产连接串不得进入图表源码或外发模型输入。
- 新增文件前，先判断它属于“最终事实”“中间设计”“实施过程”“评审证据”“长期积累”中的哪一类，不要因为顺手就丢进 `docs/architecture/` 或 `docs/plan/`。
- 架构文档文件名默认首选中文命名；只有用户明确指定其他语言，或必须与外部英文契约保持一致时，才使用英文文件名。
- `docs/architecture/`：当前架构与最终设计事实源。
  - `backend/`、`frontend/`：长期架构、运行模型、页面最终设计、设计系统。
  - `backend/design/`：已采用、但不适合并入总览的后端专题架构。
  - `features/`：面向产品能力或业务专题的最终架构事实，也可以承接该专题当前已经固定的内部边界结论；不要单独保留“迁移过程/收口过程”专题。
    - 单题题包设计、题面设计、解法设计不放这里。
    - 已落地的单题事实优先回到对应题包目录，例如 `challenge.yml`、`statement.md`、题目源码与测试。
    - 仍在推演的单题方案进入 `docs/design/`，不要伪装成架构专题。
- `docs/contracts/`：API、OpenAPI、事件、题包格式、导入导出结构等契约。
- `docs/design/`：仍在推演的中间设计稿、设计索引和过期说明；这里不是最终事实源。
- `docs/plan/impl-plan/`：结构性实现方案、阶段计划、执行清单和验证步骤。
- `docs/reviews/`：代码、架构、UI、流程的 review 记录和 findings；它们是评审证据，不覆盖当前设计事实。
- `docs/requirements/`：需求基线、范围说明、差距分析和立项约束。
- `docs/tasks/`：任务分解、拆解清单和阶段性工作列表；不替代架构或 implementation plan。
- `docs/todos/`：明确尚未收口的 backlog、延期项和后续跟踪。
- `docs/operations/`：运维手册、联调步骤、演练记录、运行侧说明。
- `docs/reports/`：阶段性报告、汇总结论、差距报告和分析输出。
- `docs/Q&A/`：会被重复引用的问答式说明，适合解释“某个能力是怎么工作的”。
- `docs/thesis/`、`docs/weekly-reports/`、`docs/开题报告/`、`docs/文献/`、`docs/毕业设计文档相关/`：论文与学校材料，不要混入产品事实源。
- `docs/ui-theme/`：历史或补充型 UI 资料；当前最终 UI 事实源仍是 `docs/architecture/frontend/`，页面参考稿进入 `docs/architecture/frontend/pages/`。
- `concepts/`：项目级长期概念、原则和 harness 定义。
- `thinking/`：尚未落地、但需要保存的判断、质疑和取舍分析。
- `practice/`：实验记录、迁移过程、历史计划索引、实践说明；它不是最终设计入口。
- `feedback/`：agent 工作流、review、prompt、policy、协作方式相关的踩坑与修正规则。
- `works/`：可直接复用的模板、地图、教程、good practices 和说明稿。
- `prompts/`：已验证提示词、工作流和可复用 prompt 资产。
- `references/`：外部文章、仓库、工具和研究资料索引。
- 已稳定的结论要回收到对应事实源：
  - 架构结论回 `docs/architecture/`
  - 契约回 `docs/contracts/`
  - 页面最终稿回 `docs/architecture/frontend/pages/`
  - 旧中间稿在原位置标记 `Superseded by ...`
- 不再新增 `docs/improvements/`、`docs/superpowers/`、`docs/refs/`、`docs/skills/` 作为活动入口；对应内容分别进入 `feedback/`、`practice/`、`references/`、`prompts/`。

### Review Workflow Preference
- 需要独立 review，且使用 subagent 执行 review 时，默认使用 `gpt-5.5`，`reasoning_effort` 使用 `medium`。
- 只有用户明确指定别的 review 模型，或当前任务本身已经带有更强的模型约束时，才偏离这条默认值。

### Destructive Database Safety
- 对删库、重建 schema、重置 migration baseline、批量清空数据、不可逆结构调整这类破坏性数据库操作，执行前必须先产出一份可恢复备份，不能只口头说明“会备份”。
- 这里的“已备份”必须满足三个条件：
  - 备份命令已经实际执行完成
  - 备份文件或备份库的具体路径 / 名称已经明确可见
  - 已经说明最小可行恢复方式，例如 `psql` / `pg_restore` 的恢复命令或恢复步骤
- 在备份产物未确认存在前，不得执行 `drop database`、`drop schema`、重建开发库、覆盖式导入或等价的破坏性操作。
- 若用户明确要求放弃备份并继续，必须先明确说明“本次无可恢复备份，执行后原数据可能无法找回”，再继续执行；不得默认替用户承担这个取舍。
- 破坏性数据库操作完成后，默认保留本次备份，直到用户明确确认可以删除，不能在同一轮里顺手清理掉。

### Backend Time Contract
- 后端业务时间统一使用 UTC。凡是会写入数据库、进入 Redis / JSON / 审计日志、参与跨请求比较、作为 API 响应返回的 `time.Time`，创建时必须使用 `time.Now().UTC()`，从请求、数据库或第三方读取后输出前也必须归一到 UTC。
- PostgreSQL 连接必须显式声明 UTC 时区；使用 pgx / GORM DSN 时应包含 `TimeZone=UTC`，不要依赖数据库实例、容器、宿主机或开发机本地时区。
- 数据库迁移中新增时间列默认使用 `timestamp with time zone` 和 `now()`；不得新增 `timestamp without time zone DEFAULT CURRENT_TIMESTAMP`。若必须接入历史无时区列，读写边界要明确按 UTC 解释并在代码里转换。
- API / DTO / WebSocket / 事件 payload 中的时间默认输出 RFC3339 UTC。手动 `Format(time.RFC3339)` 前应先 `.UTC()`，除非字段明确是本地展示时间且文档已说明。
- 比赛、AWD、实例生命周期、登录锁定、报表过期、提交记录、公告、通知、审计、题目导入 / 发布、writeup 发布时间等业务字段都属于 UTC 业务时间。
- 只有纯运行时测量允许直接用 `time.Now()`：耗时统计、deadline、timer、随机后缀、临时文件名、单进程内限流窗口等不落库、不出 API、不跨进程比较的场景。
- Review 后端时间相关改动时，必须同时检查：
  - DSN / 数据库 session 是否显式 UTC
  - 写库时间是否使用 `time.Now().UTC()`
  - 请求输入时间和数据库读出时间是否在领域 mapper / DTO 边界转 UTC
  - migration 是否混入 `timestamp without time zone` 或 `CURRENT_TIMESTAMP`
  - 测试是否覆盖至少一个 UTC location 断言

### Backend Context Contract
- 后端业务代码不得自行创建 `context.Background()` 或 `context.TODO()`。服务、仓储、任务、网关、runner、checker、runtime 操作等链路必须从上游显式接收 `ctx context.Context` 并继续向下传递。
- `context.Background()` 只允许出现在真正的进程根、框架入口、命令行入口、测试根或架构测试明确白名单文件中；普通 application / domain / infrastructure 代码里需要上下文时，必须改函数签名传入 `ctx`。
- 派生上下文必须基于传入的 `ctx`，例如 `context.WithTimeout(ctx, ...)`、`context.WithCancel(ctx)`；不要用 `context.Background()` 作为逃避取消、超时、trace、审计链路的替代品。
- Review 后端改动时，如果看到新增 `context.Background()` / `context.TODO()`，默认视为需要修复，除非能证明该文件属于进程根或测试根。

### Frontend Route Namespace Rules
- 本节为 `ctf/code/frontend` 的页面路由命名空间规范，用于统一教师端与管理员端的 URL 语义。
- 教师端页面路由只使用 `/academy/*`。
- 管理员端页面路由只使用 `/platform/*`。
- `/teacher/*` 与 `/admin/*` 不再作为前端页面路由保留，也不再新增任何重定向、兼容入口或导航匹配逻辑。
- 若页面属于教师/管理员共享视图，允许复用同一组件或同一 route name 解析工具，但最终落到浏览器地址栏的路径仍必须分别归属 `/academy/*` 或 `/platform/*`。
- 后端 HTTP API 命名空间不受本节约束；本节只约束前端页面路由、导航、redirect、错误页返回入口与测试基线。
- 当修改前端路由时，必须同步检查以下位置是否仍引用旧命名空间：
  - `router/index.ts`
  - `config/backofficeNavigation.ts`
  - `utils/roleRoutes.ts`
  - 登录跳转、错误页返回入口、侧边栏/顶部导航测试

<!-- BEGIN HARNESS ENGINEERING: root-navigation -->
## Harness Engineering 学习档案

严格参考 `deusyu/harness-engineering` 的顶层结构：

| 目录 | 内容 | 说明 |
|------|------|------|
| `concepts/` | 概念笔记 | Harness 核心概念与 CTF 项目映射 |
| `thinking/` | 独立思考 | 对项目 harness 边界和取舍的判断 |
| `practice/` | 动手实践 | 初始化和后续实验记录 |
| `feedback/` | 反馈记录 | 踩坑、修正和可复用经验 |
| `works/` | 作品输出 | 可展示模板、报告和说明 |
| `prompts/` | 提示词积累 | 已验证提示词和工作流 |
| `references/` | 外部资源 | 文章、仓库和工具索引 |

机械化检查：`bash scripts/check-consistency.sh`。
<!-- END HARNESS ENGINEERING: root-navigation -->
