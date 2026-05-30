# CTF Project AGENTS

## Project Harness Intake

- 本仓库默认先进入 harness：除非任务明显简单、局部、可逆且不需要沉淀经验，否则开始前必须先按 `harness-router` 判断 `SIMPLE` / `HARNESS`。
- 路线为 `HARNESS` 时，先读本文件和相关 harness 入口，再决定是否需要计划、review、验证或更新 `feedback/`、`harness/prompts/`、`references/`、`works/`。
- 任务分流使用全局 `harness-router` skill；机械化检查使用 `bash scripts/check-consistency.sh`。
- 开始新任务前，先运行 `bash scripts/check-task-intake.sh`；它会调用 `bash scripts/check-open-todos.sh --quiet-if-empty` 提示 `docs/todos/` 里未收口事项。若将进入受保护实现面，必须在动代码前运行 `bash scripts/check-task-intake.sh --reuse-decision <task-slug>`，先让 task-scoped reuse decision 过 startup gate，再进入实现。
- 对 `API / filter / sort / pagination` 契约改动，plan review 必须写清 `normalize / default / validate` 的唯一 owner，至少覆盖 `handler -> application -> repository` 三层。
- 对同一输入语义出现跨层重复 `normalize/default/validate` 的情况，默认不视为“安全兜底”。除非其中一层承担明确 trust-boundary 防御且理由写清，否则应继续收口成单点 owner。
- 内部 `filter / sort / pagination` contract 应收口成 downstream 不易误用的表示，优先 opaque value object、受控构造器或不可直接手工拼装的字段。
- 测试失败、review finding 或契约漂移时，先判断是不是 owner / contract / 架构未收口；不得为了让测试变绿而放宽断言、修改 fixture / mock 迁就错误实现。

## Product And Design Context

- 这是学校使用的 CTF 平台，核心用户包括学生、教师和管理员；前端设计优先级默认以学生刷题体验为先。
- 品牌气质是技术、专业、克制。学生侧强调专注、可靠、可控；教师侧强调信息清楚、判断高效；管理员侧强调系统稳健、秩序明确。
- 整体视觉方向是 light 优先、支持 dark、极简专业。技术感来自布局秩序、稳定对齐、克制色彩和明确状态，而不是霓虹、黑客风、游戏商城感或企业 OA 感。
- 页面文案保留结构性标识和必要功能说明；移除设计介绍式文案、布局解释、实现说明和脚手架文案。

## CTF Frontend Local Rules

- 通用前端工程守则由 `frontend-engineer` skill 承接；本节只保留 CTF 项目特有约束。
- 管理员端“目录标题 + `WorkspaceDirectoryToolbar` + 列表/表格/空状态/分页”结构的垂直节奏由页面目录 section 统一控制，避免 section `gap` 与 toolbar `margin-bottom` 叠加。
- 前端页面路由命名空间：教师端只使用 `/academy/*`，管理员端只使用 `/platform/*`；不再新增 `/teacher/*`、`/admin/*` 页面路由、重定向或兼容入口。
- Vue 单文件组件统一采用 `template -> style -> script` 块顺序；`<script setup>` 或 `<script>` 必须放在文件最下方，不再写在 `<template>` 上方。
- `entities/*` 只放稳定业务对象表达，不放页面流程 owner。
  - 适合进入 `entities/*` 的内容：某个业务对象的共享展示组件、状态/文案映射、轻量类型或多个 feature 都会复用且语义仍然明显属于该对象的 UI。
  - 不适合进入 `entities/*` 的内容：上传/提交/导入/发布/筛选/保存/跳转这类用户动作流程，或依赖 route、弹窗编排、异步工作流的页面壳；这些应留在 `features/*`。
  - 判断规则：如果代码主要在回答“这个业务对象是什么、如何稳定展示”，优先放 `entities/*`；如果主要在回答“用户在这里要完成什么动作”，优先放 `features/*`。
  - `shared/*` 不承载带明显 challenge / contest / class / image / writeup 等业务语义的展示块；这类内容优先判断是否应落在对应 `entities/*`。
  - `features/*` 可以依赖 `entities/*`，但 `entities/*` 不能反向依赖具体 feature 的 workflow、route state 或页面壳。
- `features/*` 的命名按角色归属区分：
  - 顶层 `features/*` 只放跨角色共享能力，或明显不绑定单一平台后台 / 教师后台的用户能力。
  - 只服务平台后台 `/platform/*` 的能力，统一放在 `features/platform/*`。
  - 只服务教师后台 `/academy/*` 的能力，统一放在 `features/teacher/*`。
  - 不再新增语义上属于平台后台、但路径仍落在顶层 `features/*` 的 feature。

## Reuse-First Workflow

- reuse-first harness 是前后端实现前置约束，不是建议项；具体受保护类型、搜索范围和前后端复用模式以 `harness/policies/reuse-first.yaml`、`harness/policies/project-patterns.yaml` 和 `harness/templates/reuse-decision.md` 为准。
- `scripts/check-task-intake.sh --reuse-decision <task-slug>` 是 reuse-first 的最早启动门禁；它检查 task-scoped reuse decision 是否已经存在且结构完整，然后才允许进入受保护实现。
- 触发受保护改动时，编码前必须在 `.harness/reuse-decisions/<task-slug>.md` 完成当前任务 reuse decision；当前 diff 里的每个受保护文件都必须被至少一个有效的 reuse decision 文档覆盖。
- `.harness/reuse-decision.md` 已废弃；仓库只接受 `.harness/reuse-decisions/` 下的 task-scoped reuse decision 文档。
- 长期复用线索改为保存在本地私有的 `.harness/reuse-index/`；仓库级入口用 `.harness/reuse-index/index.yaml`，模块级 / 模块内部二级索引用 `.harness/reuse-index/<source-path>/README.md` 镜像源码目录，不再把这类个人索引提交进仓库。
- 本地 workflow 是 reuse-first harness 的权威入口；`scripts/check-task-intake.sh` 负责实现前 startup gate，`scripts/check-reuse-first.sh` 和 `.githooks/pre-commit` 负责提交前终态检查。

## Workflow Overrides

- 本节为 `ctf` 仓库内的覆盖规则；命中时优先于上层通用 worktree 约定。
- 本仓库提交信息格式固定为 `英文类型(可选 scope): 中文描述`，例如 `fix(frontend): 修正拓扑页导出按钮禁用态`；`fix`、`refactor`、`docs`、`chore` 等类型前缀必须使用英文。
- 当用户没有明确指定“新分支 / worktree / 隔离开发”时，默认直接在当前主工作区的 `main` 分支修改。
- 纯文档编辑默认直接在当前分支修改，不创建新分支、不创建新 worktree。
- 更新前端 UI、页面样式、视觉优化时，默认直接在 `main` 分支修改；只有用户明确要求隔离开发时才创建新分支或 worktree。
- 若当前仓库实际不在 `main`，先说明当前所在分支，再按用户意图继续执行。
- 若本地已启动前端开发服务，默认复用当前服务以便用户立即看到页面变化，不主动中断。

### Completion Validation Gate

- “完成”是强约束，不是描述性用语：未执行完约定验证并通过前，不得宣称完成、已修复或可提交。
- 涉及 mapper / DTO / contract 边界迁移的任务，完成前至少执行以下三类验证：
  - `go generate`：受影响包（至少覆盖本次改动触达的 mapper 包）
  - `go test`：受影响模块（最小充分范围）
  - 全局 mapper 委托约束：`go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- 回复中必须给出本次实际执行的验证命令与结果；不能用“理论上通过”“应当通过”替代执行证据。

## File Placement Rules

- 文档写作、目录归属、命名和验证细则以 `docs/文档规范.md` 为准；新增、移动、删除或修改文档前，必须先按其中“文档修改前置读取协议”和“新增路径登记协议”执行。
- 规范化存量架构文档时，必须按 `docs/文档规范.md` 的“架构文档规范化流程”执行。
- 生成架构图或让模型根据文档生成图片前，必须按 `docs/文档规范.md` 的“架构图生成规范”执行。
- 新增文件前，先判断它属于“最终事实”“中间设计”“实施过程”“评审证据”“长期积累”中的哪一类。
- `docs/architecture/`：当前架构与最终设计事实源；后端、前端、业务专题和最终页面设计进入这里。
- `docs/contracts/`：API、OpenAPI、事件、题包格式、导入导出结构等契约。
- OpenAPI v1 采用双层结构：`docs/contracts/openapi-v1/` 是拆分源，`docs/contracts/openapi-v1.yaml` 是稳定 bundle；修改 OpenAPI 时先改拆分源，再运行 `python3 scripts/sync_openapi_from_contract.py`。
- `docs/design/`：仍在推演的中间设计稿、设计索引和过期说明。
- `docs/plan/impl-plan/`：当前仍在执行或尚未归档的结构性实现方案、阶段计划、执行清单和验证步骤。
- `docs/plan/archive/impl-plan/`：历史实施计划目录入口；当前只保留目录说明，旧计划正文默认通过 Git 历史追溯，不作为当前事实源或默认读取入口。
- `docs/reviews/`：代码、架构、UI、流程的 review 记录和 findings。
- `docs/requirements/`：需求基线、范围说明、差距分析和立项约束。
- `docs/tasks/`：任务拆解清单和阶段性工作列表；不替代架构或 implementation plan。
- `docs/todos/`：明确尚未收口的 backlog、延期项和后续跟踪。
- `docs/operations/`：运维手册、联调步骤、演练记录和运行侧说明。
- `docs/reports/`：阶段性报告、汇总结论、差距报告和分析输出。
- `docs/Q&A/`：会被重复引用的问答式说明。
- `docs/thesis/`、`docs/weekly-reports/`、`docs/开题报告/`、`docs/文献/`、`docs/毕业设计文档相关/`：论文与学校材料，不混入产品事实源。
- `concepts/`、`thinking/`、`practice/`、`feedback/`、`works/`、`references/` 是 harness 顶层目录；各目录局部规则见对应 `AGENTS.md`。
- `.harness/` 保存当前任务状态、短期执行证据和用户本地私有索引；其中 `.harness/reuse-decisions/` 只放任务证据，`.harness/reuse-index/` 只放用户自用的长期复用索引，不进入版本控制。
- 已稳定的结论要回收到对应事实源；旧中间稿在原位置标记 `Superseded by ...`。
- 不再新增 `docs/improvements/`、`docs/superpowers/`、`docs/refs/`、`docs/skills/` 作为活动入口；对应内容分别进入 `feedback/`、`practice/`、`references/`、`harness/prompts/`。

## Backend Contracts

- 后端业务时间统一使用 UTC。凡是会写入数据库、进入 Redis / JSON / 审计日志、参与跨请求比较、作为 API 响应返回的 `time.Time`，创建时必须使用 `time.Now().UTC()`，从请求、数据库或第三方读取后输出前也必须归一到 UTC。
- PostgreSQL 连接必须显式声明 UTC 时区；新增 migration 时间列默认使用 `timestamp with time zone` 和 `now()`。
- API / DTO / WebSocket / 事件 payload 中的时间默认输出 RFC3339 UTC。
- 纯运行时测量可以直接用 `time.Now()`，例如耗时统计、deadline、timer、随机后缀、临时文件名、单进程内限流窗口。
- 后端业务代码不得自行创建 `context.Background()` 或 `context.TODO()`；服务、仓储、任务、网关、runner、checker、runtime 操作等链路必须从上游显式接收 `ctx context.Context` 并继续向下传递。
- `context.Background()` 只允许出现在进程根、框架入口、命令行入口、测试根或架构测试明确白名单文件中。

## Destructive Database Safety

- 对删库、重建 schema、重置 migration baseline、批量清空数据、不可逆结构调整这类破坏性数据库操作，执行前必须先产出一份可恢复备份。
- “已备份”必须满足：备份命令实际执行完成，备份文件或备份库路径 / 名称明确可见，已说明最小可行恢复方式。
- 在备份产物未确认存在前，不得执行 `drop database`、`drop schema`、重建开发库、覆盖式导入或等价破坏性操作。
- 破坏性数据库操作完成后，默认保留本次备份，直到用户明确确认可以删除。

## Harness Engineering Map

- `concepts/`：Harness 核心概念与 CTF 项目映射。
- `thinking/`：对项目 harness 边界和取舍的判断。
- `practice/`：初始化和后续实验记录。
- `feedback/`：踩坑、修正和可复用经验。
- `works/`：可展示模板、报告和说明。
- `harness/prompts/`：已验证、会复用的项目级 agent 工作流 prompt。
- `references/`：文章、仓库和工具索引。
- 机械化检查：`bash scripts/check-consistency.sh`。
