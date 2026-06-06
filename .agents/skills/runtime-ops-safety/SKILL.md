---
name: runtime-ops-safety
description: Use when running tests, scripts, recursive scans, background jobs, or performance checks that may consume significant resources, outlive the current turn, or leave residual processes.
---

# Runtime Ops Safety

本仓库里的这份 `runtime-ops-safety` 是项目补充入口，不再作为通用运行安全规则的主体。

## 主体来源

- 通用主体：`~/.agents/skills/runtime-ops-safety`
- 项目补充：
  - `development-pipeline`
  - `security-vulnerability-scan`

## 在 CTF 仓库中的使用方式

跑测试、脚本、扫描、临时服务或长命令前：

1. 先使用全局 `runtime-ops-safety`
2. 再结合本仓库入口判断 startup gate、守卫脚本和收尾动作
3. 如果会留下后台进程、端口或扫描结果，完成后立刻清理本轮启动的残留进程

## 本地补充关注点

- 任务开始先跑 `bash scripts/check-task-intake.sh`
- 改了 harness、入口、skill、hook 或文档导航后，收尾要跑 `bash scripts/check-consistency.sh`
- 本仓库有前端守卫、文档守卫和 agent 入口检查，验证顺序要按依赖串行执行，不要并发糊过去

## 项目内附带参考

- `scripts/check-task-intake.sh`
- `scripts/check-consistency.sh`
- `scripts/check-agent-entrypoints.sh`
