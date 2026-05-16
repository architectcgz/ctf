# AWD Runtime 后续待办

- Project: `/home/azhi/workspace/projects/ctf`
- Created: `2026-05-16T21:15+08:00`

## 背景

今天已经补了 runtime Docker 封装收口、赛中 AWD 配置冻结、desired reconcile 抑噪、`CTF_CONTAINER_FLAG_GLOBAL_SECRET` 自动持久化，以及 AWD 控制面缺口。这里仅保留 AWD 运行恢复链路上仍未完整展开的后续项，作为单独 backlog 跟踪。

## 当前待办

- [ ] P0：做一次真实宿主重启的端到端恢复演练
  - 目标不是继续补单测或组合测试，而是验证“API + Docker 宿主一起起来”后的真实恢复行为
  - 演练前要先明确回放步骤、验收口径和要保留的证据

## 备注

- 这份 todo 只记录今天明确确认、并且目前仍未完成的后续项。
