# CTF Instance Lifecycle Research

## Migrated From

历史 refs 条目 `ctf-platform-instance-lifecycle-research-2026-04-13`

## 主题

题目实例、靶机或云端工作台在用户解题后通常如何处理：是否自动关闭、是否支持延时、是否要求用户手动终止，以及平台如何提示资源仍在运行。

## 样本

- CTFd `Application Target`
- Hack The Box `Challenges` / `Machines`
- TryHackMe `AttackBox` / 目标机器

## 关键结论

- 更常见的是显式生命周期管理，不是“解题即销毁”。
- Flag 提交与实例生命周期通常不是同一个动作。
- 平台会明确提醒用户资源仍在运行。
- 自动回收通常服务于资源与成本控制。

## 对本项目的含义

当前提交成功链路不应默认收缩实例 `expires_at` 或直接关闭实例。更稳妥的设计是让实例运行状态、TTL、延时、手动停止和资源提醒成为显式产品交互，并由后端生命周期策略兜底回收。

## 原始资料

完整调研仍保留在原路径；本文件作为严格 harness `references/` 下的资料入口。
