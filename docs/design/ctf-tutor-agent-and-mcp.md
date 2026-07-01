# CTF 伴学 Agent 与 MCP 边界设计

> 状态：Draft
> 事实源：2026-06-29 讨论结论
> 替代：外部 MCP 认证部分已由 `docs/contracts/api-contract-v1.md` 的 OAuth 浏览器授权契约和 `docs/operations/mcp-oauth-login.md` 替代；其余 tutor-agent / RAG 边界仍为 Draft。

## 定位

| 本设计负责 | 本设计不负责（见其他文档） |
| --- | --- |
| 记录 CTF 平台 MCP、内部 Python 伴学 Agent、RAG / 向量化知识库、本地工具调用之间的职责边界。 | 不定义当前已落地后端模块事实，已落地事实以 `docs/architecture/backend/` 为准。 |
| 给出 `tutor-agent` 与 `backend` 同级放置的目标目录建议。 | 不替代 OpenAPI、数据库 schema、题包契约或容器 runtime 架构。 |
| 给出 MCP 是否暴露端口、暴露到公网还是内网的部署判断；其中外部 MCP 认证已收敛为 OAuth browser authorization。 | 不定义具体模型供应商、Prompt 文案和 UI 交互细节。 |
| 约束伴学 Agent 的安全边界：教学、提示、复盘，不自动攻击真实目标。 | 不提供真实公网目标扫描、漏洞利用或绕过权限能力。 |

## 背景

MCP 是 Model Context Protocol，用于给 AI 应用或 Agent 提供额外工具和上下文。它本身不是 Agent，不负责规划、记忆、自主循环或多 Agent 协作；这些属于 Host / Agent 服务。

在本项目里，MCP 的价值不是让 Agent 第一次拥有项目上下文。Claude / Codex 这类开发 Agent 已经可以直接读取当前仓库文件。MCP 更适合封装外部、动态、权限受控或需要多客户端复用的能力，例如题目列表、分值、附件、Hint、Flag 提交、Scoreboard、线上状态和本地分析工具。

本设计把目标拆成三类运行单元：

```text
CTF Backend
  提供平台领域能力、权限、审计、题目、提交、计分和附件管理

CTF Platform MCP
  作为 CTF Backend 的协议适配入口，暴露 Agent 可调用的标准工具

Tutor Agent
  Python 内部服务，负责伴学对话、分级提示、知识检索和工具编排
```

## 当前设计

### 组件边界

| 组件 | 负责 | 不负责 |
| --- | --- | --- |
| `code/backend` | 题目、分值、附件、Hint、Flag 提交、Scoreboard、竞赛状态、权限、审计、限流和领域规则。 | 不生成伴学对话，不直接内嵌 LLM 编排和 RAG 检索流程。 |
| `code/backend` 内的 MCP 入口 | 作为平台能力的 MCP 协议适配层，复用后端 application service 和用户上下文。 | 不直接查数据库绕过 application service，不承担教学策略。 |
| `tutor-agent` | Python 内部服务，负责对话编排、分级提示、复盘总结、RAG 检索、工具选择和安全策略。 | 不直接读写 CTF 数据库，不绕过平台权限提交 Flag、扣 Hint 分或下载附件。 |
| `ctf-tools-mcp`（可选后续服务） | 封装本地分析工具，例如 `file`、`strings`、`exiftool`、`binwalk`、`checksec`、`readelf`。 | 不访问平台数据库，不承担题目权限和计分规则。 |
| `ctf-knowledge-rag`（可内嵌或独立） | 检索 CTF 知识库、历史题解、漏洞模式和学习资料。 | 不替代平台业务数据源，不给出未经引用的当前题目事实。 |

### 目标目录

内部伴学 Agent 放在与现有后端同级的位置，作为独立 Python 服务：

```text
ctf/
  code/
    backend/
    frontend/
  tutor-agent/
    app/
      api/
      agent/
      ctf/
      memory/
      prompts/
      rag/
      tools/
    tests/
    pyproject.toml
    Dockerfile
```

如果后续仓库决定所有可部署单元统一放在 `services/`，可以迁移为：

```text
ctf/
  services/
    tutor-agent/
```

当前约定优先使用仓库根下的 `tutor-agent/`，因为用户已确认它应与 `backend` 同级，而不是放入 `code/backend/internal/module/*` 这样的后端领域模块内部。

### 平台 MCP 放置

平台 MCP 应放在 CTF 后端的接口适配层，与 HTTP / WebSocket 入口同级：

```text
code/backend/
  internal/
    app/
    module/
    interfaces/
      http/
      mcp/
```

如果当前代码尚未使用 `interfaces/` 目录，落地时应按后端既有分层命名选择最接近的协议入口层。关键约束是：MCP handler 只能复用 application service 和已存在的权限 / 审计 / 限流机制，不直接访问 repository 或数据库表。

第一版平台 MCP 目标工具：

```text
list_challenges(category?, difficulty?)
get_challenge(challenge_id)
download_attachment(challenge_id, file_id)
get_hints(challenge_id)
unlock_hint(challenge_id, hint_id)
submit_flag(challenge_id, flag)
get_my_solves()
get_scoreboard(limit?)
```

读操作与副作用操作必须分开审计：

| 类型 | 工具 | 要求 |
| --- | --- | --- |
| 只读 | `list_challenges`、`get_challenge`、`get_hints`、`get_my_solves`、`get_scoreboard` | 必须绑定当前用户 / 队伍权限，只返回用户可见数据。 |
| 文件读取 | `download_attachment` | 必须复用附件访问权限、比赛状态和下载审计规则。 |
| 有副作用 | `unlock_hint`、`submit_flag` | 必须复用扣分、判题、计分、限流、审计和幂等规则。 |

### Tutor Agent 服务边界

`tutor-agent` 是应用服务，不是 CTF 领域模型的一部分。它通过内网调用平台 MCP 或后端 API 获取事实和执行动作：

```text
Frontend
  -> Backend /api/tutor/chat
  -> tutor-agent:9000
  -> backend /mcp
  -> application service
```

推荐由主后端代理前端到 Agent 的请求：

```text
浏览器
  -> https://ctf.example.com/api/tutor/chat
  -> backend
  -> http://tutor-agent:9000/chat
```

这样认证、用户身份、班级 / 队伍上下文、比赛状态和平台审计更容易复用。前端直接访问 `tutor-agent` 只作为后续可选方案，不作为第一版默认路径。

`tutor-agent` 内部建议拆分：

```text
tutor-agent/app/
  main.py
  api/chat.py
  agent/orchestrator.py
  agent/policy.py
  agent/prompts.py
  ctf/client.py
  ctf/models.py
  memory/store.py
  rag/ingest.py
  rag/retriever.py
  tools/sandbox.py
```

### MCP 与 Agent 的调用关系

```text
Tutor Agent
  决定下一步教学动作、提示等级和工具选择
    |
    | MCP / HTTP
    v
CTF Platform MCP
  获取题目、分值、附件、Hint、提交 Flag、Scoreboard
    |
    v
CTF Backend Application Service
  权限、审计、计分、判题、附件、比赛状态
```

MCP Server 不应生成教学话术，也不应判断用户学习阶段。它只提供事实和受控动作。教学策略属于 `tutor-agent`。

### RAG 与向量化边界

RAG 用于给伴学 Agent 提供知识库检索能力。流程是把 CTF 知识文档、历史题解、漏洞模式和课程资料切分成 chunk，通过 embedding 模型向量化后存入向量库。用户提问时，Agent 先检索相关片段，再基于片段回答。

RAG 适合回答：

```text
SQL 注入的布尔盲注思路是什么？
ELF 文件先看哪些信息？
常见 XOR 题怎么判断 key 长度？
这类文件隐写题有哪些排查步骤？
```

RAG 不适合替代平台事实源。题目分值、用户是否已解出、Hint 是否已解锁、Flag 是否正确，都必须从 CTF Backend / Platform MCP 获取。

### 部署与端口

部署时是否暴露 MCP 端口，取决于调用方在哪里。

| 场景 | MCP 暴露方式 | 推荐 |
| --- | --- | --- |
| 只给内部 `tutor-agent` 调用 | 只开放内网地址，例如 `http://backend:8000/mcp` | 推荐第一版 |
| 给外部 AI 客户端直接连接 | 通过现有 HTTPS 暴露 `/mcp`，例如 `https://ctf.example.com/mcp` | 已采用 OAuth 2.1 Authorization Code + PKCE；不再使用手工 MCP token |
| MCP 独立部署 | 内部服务端口如 `ctf-mcp:8000`，外部经 Ingress / Nginx 复用 443 | 适合后续拆服务 |
| Agent 内部服务 | `tutor-agent:9000` 仅内网开放 | 推荐第一版 |

推荐第一版部署拓扑：

```text
公网:
  https://ctf.example.com/api/tutor/chat

内网:
  backend -> tutor-agent:9000
  tutor-agent -> backend:8000/mcp
```

Claude Desktop、Cursor、Codex 或 VS Code 直接接入平台 MCP 时，`/mcp` 通过公网 HTTPS 暴露，并使用 `/.well-known/oauth-protected-resource`、`/.well-known/oauth-authorization-server`、`/api/v1/oauth/register`、`/api/v1/oauth/authorize` 和 `/api/v1/oauth/token` 完成用户级 OAuth 授权；不再配置 `CTF_MCP_TOKEN` 或手工签发 Bearer token。

### 认证与安全边界

MCP 与 Tutor Agent 都不能使用管理员 token 代表所有用户执行操作。每次调用必须带当前用户 / 队伍上下文：

```text
UserContext(user_id, team_id, role, contest_id?)
```

必须按用户身份检查：

```text
题目是否可见
附件是否可下载
Hint 是否可解锁
提交是否在比赛时间内
Scoreboard 是否对当前角色开放
实例是否属于当前用户或队伍
```

伴学 Agent 的安全边界：

| 允许 | 不允许 |
| --- | --- |
| 本地 CTF 附件分析 | 扫描公网目标 |
| 靶场 / 比赛环境内的授权题目 | 绕过真实系统权限 |
| 分级提示、概念解释、复盘总结 | 生成持久化后门或真实攻击链 |
| 调用受限工具分析题目文件 | 无限制执行任意 shell 命令 |
| 提交用户明确提供或推导出的 Flag | 代替用户自动化刷题和批量提交 |

## 分阶段落地建议

第一阶段先做平台 MCP 和最小 Agent 闭环：

1. 在后端新增内网 MCP 入口，提供题目详情、分值、附件、Hint、提交 Flag 和个人解题状态。
2. 在仓库根新增 `tutor-agent/` Python 服务，只实现 `/chat` 和平台 MCP client。
3. 前端通过后端 `/api/tutor/chat` 调用内部 Agent。
4. Agent 只提供分级提示和概念解释，不接本地命令执行。

第二阶段加入 RAG：

1. 建立 CTF 知识库文档目录。
2. 增加 ingest / retriever。
3. Agent 回答知识点时必须带可追溯片段。

第三阶段加入工具 MCP：

1. 封装 `file`、`strings`、`exiftool`、`binwalk`、`checksec` 等本地工具。
2. 限制执行目录、超时、输出大小和危险命令。
3. 工具调用结果只作为教学辅助，不自动提交攻击结果。

第四阶段考虑外部 MCP 开放：

1. 公网暴露 `https://ctf.example.com/mcp`。
2. 使用已落地的 OAuth browser authorization、`mcp:challenge:read` scope、rate limit 和审计日志。
3. 允许外部 AI 客户端读取题目和提交 Flag，但默认禁用高风险操作或要求二次确认。

## 关键决策

| 决策 | 结论 | 原因 |
| --- | --- | --- |
| MCP Server 由谁提供 | 平台侧提供；如果平台没有实现，Agent 侧只能临时写适配器。 | 你是 CTF 平台开发者，平台最清楚权限、题目、分值、Hint、提交和审计规则。 |
| Tutor Agent 放哪里 | 放在与 `backend` 同级的 `tutor-agent/`。 | Agent 是独立应用服务，不属于后端领域模块。 |
| Tutor Agent 是否直连数据库 | 不直连。 | 直连会绕过权限、扣分、比赛状态、审计和限流。 |
| MCP 是否必须公网暴露 | 不必须。 | 内部伴学 Agent 使用时只需要内网访问；外部 AI 客户端接入时才通过 HTTPS 暴露。 |
| MCP 是否负责教学 | 不负责。 | MCP 是工具 / 上下文接口，教学策略属于 Agent。 |
| 第一版是否自动解题 | 不做自动解题。 | 项目定位是伴学，不是代打；需要保护教学价值和安全边界。 |

## 已知限制

- 当前文档是设计稿，尚未对应代码落地。
- 外部 MCP 浏览器授权已落地到 `code/backend/internal/module/auth/` 和 `code/backend/internal/interfaces/mcp/`；本文仍不作为当前契约事实源。
- 具体 MCP SDK、内部 tutor-agent 传输方式和模型供应商仍待实现计划确认。
- RAG 向量库选型、知识库目录和 chunk 策略尚未确定。
- 工具 MCP 的沙箱策略需要单独设计，不能只靠 Prompt 限制。

## Guardrail

- 后续实现 `tutor-agent` 前，必须先写正式 implementation plan，并明确：
  - Agent 与 Backend 的调用契约。
  - 用户身份如何传递到 MCP。
  - 哪些 tool 有副作用，哪些需要二次确认。
  - RAG 知识库与平台事实源的优先级。
  - 本地工具执行的目录、超时、网络和命令 allowlist。
- 后续若把本 Draft 采纳为当前事实，必须回收到 `docs/architecture/` 或 `docs/contracts/`，并把本文标记为 `Superseded`。
