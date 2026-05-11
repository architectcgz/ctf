# AWD 记分榜闭环待办

更新日期：2026-03-11

## 目标

继续围绕“竞赛管理”主线，补齐 AWD 攻防得分到实时排行榜的真实回写链路。

## 本轮任务

1. [x] 基于 `awd_team_services.defense_score` 与 `awd_attack_logs.score_gained` 重算 AWD 赛事队伍总分
2. [x] 在手工服务检查、轮次巡检、攻击日志写入后同步回写 `teams.total_score`
3. [x] 同步刷新竞赛排行榜 Redis 缓存，打通现有 `/scoreboard` 主链路
4. [x] 运行 `go test ./internal/module/contest` 验证 AWD 计分与排行榜闭环
