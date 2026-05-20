# AWD Topology Five Packs Implementation Plan

## Plan Summary
- Objective
  - 为 AWD 补齐 `topology` 题的最小可用闭环，并基于该闭环新增 5 道三容器 AWD 题。
  - 这 5 道题都必须满足：题包可导入、拓扑配置可被平台读取、运行链路不会把三容器题退化成单容器题、题解与本地验题材料完整。
- Non-goals
  - 不在这一轮扩到 20 题总量。
  - 不重做通用 `challenge topology studio` 的普通题导入逻辑。
  - 不把 AWD 预览自动拉起能力一次扩成所有复杂部署模式，只补本次 topology AWD 所需的最小闭环。
- Source architecture or design docs
  - `challenges/awd/challenge-package-contract.md`
  - `docs/contracts/challenge-pack-v1.md`
  - `docs/architecture/features/题包拓扑同步与导出架构.md`
  - `docs/architecture/features/校园级CTF-AWD模式完整设计.md`
- Dependency order
  - 先确认 AWD 当前 `topology` 缺口与 owner
  - 再补 AWD 题包解析/导入/运行定义
  - 再批量落 5 道三容器题
  - 最后做本地 build / import / checker / 文档验证
- Expected specialist skills
  - `development-pipeline`
  - `go-backend`
  - `test-engineer`
  - `runtime-ops-safety`
  - `code-reviewer`

## Task 1
- Goal
  - 收口 AWD topology 的事实源与 owner，避免出现“题包是三容器，平台运行仍按单容器”的错位。
- Touched modules or boundaries
  - `code/backend/internal/module/challenge/domain`
  - `code/backend/internal/module/challenge/application/commands`
  - `code/backend/internal/module/contest/application/commands`
  - `code/backend/internal/module/runtime`
  - `docs/contracts`
- Dependencies
  - 需要先确认当前 AWD import、preview、runtime launch 各自消费什么字段。
- Validation
  - 相关单测至少覆盖：
    - AWD 题包能解析 `docker/topology.yml`
    - AWD 导入能保留 topology 定义
    - 非 topology 题不会被回归影响
- Review focus
  - `topology` 唯一 owner 是否明确
  - 是否又造了一套与普通 challenge topology 并行但不兼容的表示
- Risk notes
  - 当前已确认普通题有 `challenge_topologies`，AWD 未必直接复用；如果 owner 不收口，后面 5 题只能是素材目录，不是可赛用题。

## Task 2
- Goal
  - 补 AWD 题包导入与运行时最小闭环，使 `deployment_mode: topology` 的 AWD 题能真正保存并被后续运行链路消费。
- Touched modules or boundaries
  - `code/backend/internal/module/challenge/domain/awd_package_parser.go`
  - `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
  - `code/backend/internal/module/challenge/api/http/awd_challenge_handler*.go`
  - `code/backend/internal/module/contest/application/commands/awd_preview_runtime_support.go`
  - 必要时补 `runtime_config` 结构读取或新增 topology support helper
- Dependencies
  - 依赖 Task 1 的 owner 决策。
- Validation
  - `go test` 覆盖 AWD import、AWD preview/runtime support、相关 handler/service 测试。
  - 至少一个 topology AWD 包通过导入预览与 commit 测试。
- Review focus
  - 平台对 topology AWD 的保存格式是否单点 owner
  - 自动 preview 不支持的边界是否明确报错而不是静默退化
- Risk notes
  - 若正式比赛实例创建链本轮无法完整接入 topology，必须明确阻断或显式标成仅手动接入，不能假装“已支持”。

## Task 3
- Goal
  - 新增 5 道三容器 topology AWD 题，统一结构、统一验题方式、统一 writeup 规范。
- Touched modules or boundaries
  - `challenges/awd/ctf-3/*`
  - `challenges/awd/README.md`
  - 必要的 `dist/*.zip`
- Dependencies
  - 依赖 Task 2 让 topology AWD 至少能导入。
- Validation
  - 每题至少完成：
    - `docker compose build`
    - 本地 `check/check.py`
    - 攻击路径可实际取到动态 flag
  - 至少抽 1 题做平台导入预览与 commit 验证。
- Review focus
  - 三容器边界是否真的服务于攻击面/防守面
  - 题解是否依赖源码捷径而不是服务行为
  - checker 是否只依赖公开入口与私有 checker token，不偷走业务漏洞路径
- Risk notes
  - 5 题必须复用一套最小基座，否则实现与验题成本会失控。

## Task 4
- Goal
  - 记录作者侧规则与题包清单，补最小必要文档。
- Touched modules or boundaries
  - `challenges/awd/challenge-package-contract.md`
  - `challenges/awd/README.md`
  - 必要时 `docs/reviews/general/*`
- Dependencies
  - 依赖前面实现结果稳定。
- Validation
  - 文档与实际目录、字段、验证命令一致。
- Review focus
  - 事实源是否和真实实现一致
  - 没有把“未来计划”写成“当前已支持”
- Risk notes
  - AWD topology 现状与普通题 topology 现状很容易被写混，必须显式区分。

## Integration Checks
- AWD `deployment_mode: topology` 的题包预览、commit、存储、preview/runtime support 是否语义一致。
- 运行时若生成 alias access URL，checker 与 target proxy 是否仍能解析到入口节点。
- 三容器题的 `defense_workspace`、`protected_paths`、checker token 注入不能被 topology 改动打断。
- 新增题包是否都保留 `writeup/attack.md`、`writeup/defense.md` 与本地 `check/check.py` 闭环。

## Rollback / Recovery Notes
- 后端支持层与题包目录分开回退：
  - 若 topology AWD 支持未收口，可先回退代码支持，不提交 5 道题目录。
  - 若代码支持稳定但个别题不合格，可单独回退对应题目目录与 zip。
- 任何 registry push、导入记录或本地实例验证都必须按题目 slug 可独立清理。

## Residual Risks
- 当前最大的残余风险是 AWD topology 运行链可能不只缺 import，还缺正式比赛实例创建的消费侧。如果证据坐实这一点，本计划必须在编码前回到 Task 1/2 收口，而不能先交 5 个目录。
