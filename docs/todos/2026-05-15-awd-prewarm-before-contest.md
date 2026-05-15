# AWD 赛前预热能力

- 项目路径：`/home/azhi/workspace/projects/ctf`
- 创建时间：`2026-05-15 00:00:00 Asia/Shanghai`
- 完成时间：`2026-05-15`
- 背景：当前这轮只把 AWD 队伍服务实例生命周期收口到 `contest.end_time`，没有解决“演示或比赛前统一拉起实例”的运营需求。若宿主机前一晚关机、Docker 重启或 runtime 清空，比赛开始前仍需要一次显式预热。

## 已完成

- 管理员端新增 `POST /admin/contests/:id/awd/instances/prewarm`，支持整场或单队赛前预热
- 预热按 `team × visible service` 返回结果矩阵，区分 `started / reused / failed`
- 已有活跃实例直接复用；无实例或已失效目标才走启动链路
- 报名阶段的“启动本队 / 启动全部”前端操作已切到批量预热接口
- 单格启动入口已允许 `registration`，便于赛前对失败单元逐个重试

## 后续项

- 如果后续需要“管理员一键全量重启”，应在预热能力之上单独评估是否复用同一入口，而不是直接把比赛中的运行态重启语义混进赛前准备链路
