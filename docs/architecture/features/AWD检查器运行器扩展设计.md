# AWD 检查器运行器扩展架构

## 文档元信息

- 状态：`implemented`
- 事实源级别：`final`
- 适用范围：`backend`、`contracts`、`frontend`
- 关联模块：
  - `internal/module/contest/application/jobs`
  - `internal/module/contest/application/commands`
  - `internal/module/contest/infrastructure`
  - `frontend/src/components/platform/contest`
- 过程追溯：`practice/superpowers-plan-index.md` 中的 `2026-04-29-awd-checker-runner-extension`
- 最后更新：`2026-05-07`

## 1. 背景与问题

AWD checker 已经不再只有 `legacy_probe` 和 `http_standard` 两种执行语义。当前代码已经把 TCP 场景、脚本场景和独立沙箱 runner 接进同一条 preview / readiness / 赛中巡检链路，文档需要明确以下事实：

- checker 类型扩展仍然以 `AWDRoundUpdater` 为统一执行入口
- `tcp_standard` 与 `script_checker` 已经是当前实现的一部分，不再只是候选方案
- 脚本 checker 不能在 API 进程或轮次调度进程里直接执行，必须经过独立 runner

## 2. 架构结论

- AWD checker 的正式类型集合是 `legacy_probe`、`http_standard`、`tcp_standard`、`script_checker`。
- 赛前试跑和赛中轮次都复用 `AWDRoundUpdater` 的按类型分派逻辑，不存在第二套平行 runner。
- `tcp_standard` 在后端进程内执行受控 TCP 步骤，不提供任意代码执行能力。
- `script_checker` 只能通过 `DockerCheckerRunner` 执行，且脚本文件必须来自题目包导入时保存的私有 artifact。
- `CHECKER_TOKEN` 已经是当前 TCP / 脚本 checker 的正式模板变量和运行时 secret，不属于学生可见攻击面。

## 3. 模块边界与职责

### 3.1 模块清单

- `AWDRoundUpdater`
  - 负责：根据 `checker_type` 分派到 HTTP、TCP、脚本或兼容探活路径
  - 不负责：持久化 preview token 或维护前端草稿状态

- `DockerCheckerRunner`
  - 负责：为 `script_checker` 创建一次性容器、注入文件与环境变量、收集退出码和输出
  - 不负责：解析 AWD 业务语义或决定服务状态

- `ContestAWDServiceService`
  - 负责：保存赛事 service 的 `checker_type`、`checker_config` 和 challenge runtime 扩展信息
  - 不负责：直接执行 checker

- `AWDChallengeConfigDialog`
  - 负责：暴露四种 checker 类型的结构化配置入口
  - 不负责：定义后端执行语义

### 3.2 事实源与所有权

- checker 类型与配置事实源：`contest_awd_services.runtime_config`
- 脚本 artifact 事实源：`AWD_CHECKER_ARTIFACT_DIR/<slug>/<digest>/...`
- 赛中检查结果事实源：`awd_team_services.check_result`
- preview 与 runtime token 算法事实源：`internal/module/contest/domain/awd_checker_token.go`

## 4. 关键模型与不变量

### 4.1 核心实体

- `checker_type`
  - 取值：`legacy_probe`、`http_standard`、`tcp_standard`、`script_checker`

- `checker_config`
  - `tcp_standard`：`timeout_ms`、`connect`、`steps`、`havoc`
  - `script_checker`：`runtime`、`entry`、`args`、`env`、`timeout_sec`、`output`、`artifact`

- `CHECKER_TOKEN`
  - preview 场景通过 `BuildAWDCheckerPreviewToken`
  - 赛中轮次通过 `BuildAWDCheckerToken`

### 4.2 不变量

- `script_checker` 缺少 `checkerRunner` 时直接视为 `checker_runner_unavailable`，不会退回本机执行。
- `script_checker` 的 `entry` 和 `files` 必须位于题目包相对路径下，不能是绝对路径，也不能穿越到 artifact root 外。
- `tcp_standard` 只支持 `send`、`send_template`、`send_hex`、`expect_contains`、`expect_regex` 这类结构化步骤。
- `CHECKER_TOKEN` 只在题目声明 `checker_token_env` 时注入运行时和模板变量。
- target audit 中的 flag、checker token、stdout/stderr 都必须先脱敏再写回结果。

## 5. 关键链路

### 5.1 配置与导入链路

1. 题目包导入时解析 checker 类型与配置。
2. 对 `script_checker`，服务端把声明文件保存到 `AWD_CHECKER_ARTIFACT_DIR` 下的私有 artifact 目录。
3. 赛事 service 保存时，把 `checker_type`、`checker_config` 和 `challenge_runtime` 写入 `runtime_config`。

### 5.2 赛前试跑链路

1. `POST /admin/contests/:id/awd/checker-preview` 进入 `AWDService.PreviewChecker`。
2. `PreviewChecker` 校验 checker 类型与配置，准备 `preview` 作用域的 `CHECKER_TOKEN`。
3. `AWDRoundUpdater.PreviewServiceCheck` 根据 checker 类型分派到 HTTP、TCP、脚本执行路径。
4. `script_checker` 通过 `DockerCheckerRunner.RunChecker` 在一次性容器中执行。
5. 结果聚合后返回 `preview_token`，供后续保存配置时消费。

### 5.3 赛中轮次链路

1. `RunRoundServiceChecks` 读取轮次、队伍、service 定义和目标实例。
2. `http_standard`、`tcp_standard`、`script_checker` 共享 `AWDRoundUpdater` 的结果聚合与服务状态判定。
3. 赛中路径使用 runtime scope 的 `CHECKER_TOKEN`，并把审计摘要写回 `check_result.targets[].audit`。

## 6. 接口与契约

### 6.1 配置契约

- `checker_type` 只接受 `legacy_probe http_standard tcp_standard script_checker`
- `script_checker.output` 支持 `exit_code` 或 `json`
- `tcp_standard` 模板变量支持：
  - `{{TARGET_URL}}`
  - `{{TARGET_HOST}}`
  - `{{TARGET_PORT}}`
  - `{{FLAG}}`
  - `{{ROUND}}`
  - `{{TEAM_ID}}`
  - `{{CHALLENGE_ID}}`
  - `{{CHECKER_TOKEN}}`

### 6.2 Runner 契约

- `DockerCheckerRunner` 默认使用只读 rootfs、`cap_drop=ALL`、`no-new-privileges`
- 默认 network mode 为 `none`；脚本场景由 `TargetAllowlist` 生成最小目标地址白名单
- 超时、输出超限、非零退出码、无效 JSON 都会转成结构化失败原因

## 7. 兼容与迁移

- `legacy_probe` 仍保留给旧 service 配置，但不再是扩展 checker 的主路径。
- `tcp_standard` 当前仍是轻量结构化 TCP runner，不支持复杂二进制状态机或动态分支。
- `script_checker` 当前依赖预导入 artifact 和内置运行时镜像，不支持赛时联网安装依赖。
- 新 checker 类型若要接入，必须同时补：
  - `NormalizeAWDCheckerType`
  - 前端结构化编辑器
  - `AWDRoundUpdater` 分派逻辑
  - readiness / preview / 结果展示测试

## 8. 代码落点

- `code/backend/internal/module/contest/application/jobs/awd_checker_preview.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_runner.go`
- `code/backend/internal/module/contest/application/jobs/awd_tcp_checker_runner.go`
- `code/backend/internal/module/contest/application/jobs/awd_script_checker_runner.go`
- `code/backend/internal/module/contest/infrastructure/docker_checker_runner.go`
- `code/backend/internal/module/contest/application/commands/awd_service_run_commands.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_support.go`
- `code/backend/internal/module/contest/domain/awd_checker_token.go`
- `code/frontend/src/components/platform/contest/AWDChallengeConfigDialog.vue`
- `code/frontend/src/components/platform/contest/awdCheckerConfigSupport.ts`

## 9. 验证标准

- `tcp_standard` 试跑和赛中巡检都能输出结构化 target 结果与 audit 摘要。
- `script_checker` 能从 artifact root 加载文件，并在容器内只读执行。
- `CHECKER_TOKEN` 仅在声明了 `checker_token_env` 的题目上注入并渲染。
- 输出超限、timeout、invalid JSON、artifact 越界路径都能得到稳定错误码，而不是退回不受控执行。
