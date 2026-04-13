# CTF 平台实例生命周期调研（2026-04-13）

## 背景

这份调研聚焦一个很具体的问题：题目实例、靶机或云端工作台在用户解题后通常怎么处理，平台会不会自动关闭，是否支持延时，是否要求用户手动终止，以及平台如何向用户提示这些资源仍在运行。

这个问题和当前项目的运行方式直接相关。按照 [毕业设计课题](../毕业设计课题.md) 的要求，平台不仅要提供靶场和攻防对抗能力，还要对实例运行、攻击步骤与教学反馈做集中管理。因此，实例生命周期策略不能只看资源回收，还要兼顾学生体验、教学连续性和教师侧的可管理性。

## 调研范围

本次只收集公开可访问、且能直接说明实例生命周期的官方资料，样本包括：

- CTFd `Application Target`
- Hack The Box `Challenges` / `Machines`
- TryHackMe `AttackBox` / 目标机器

没有纳入非官方博客、论坛经验帖，也没有把当前无法稳定访问正文的资料作为结论依据。

## 平台样本

### 1. CTFd：原生支持按用户部署实例，但官方页没有强调“解题后自动销毁”

CTFd 的 `Application Target` 文档明确说明，这类题目允许“按账户部署一个应用实例”，实例可以由拥有者部署、重新部署和访问。文档同时说明端口会映射到公开 URL，并支持按用户生成随机 Flag，用于一定程度的防作弊。

从这页官方文档能确认的重点是：

- 平台原生支持 `per-account instance`
- 支持 `deploy` / `redeploy` / `access`
- 支持按用户注入唯一 Flag
- 官方页重点在“如何把题目实例跑起来”和“如何给实例发唯一 Flag”

这页文档没有看到“用户提交正确 Flag 后立即销毁实例”或“解题后自动收缩生命周期”的说明。这里不能直接推断 CTFd 完全没有这类能力，只能说至少在其原生 `Application Target` 文档中，这不是被突出表达的默认策略。

来源：

- [CTFd Docs - Application Target](https://docs.ctfd.io/docs/custom-challenges/application-target/)

### 2. Hack The Box：更偏固定生命周期、可延时、手动停止，Flag 提交与实例关闭分离

Hack The Box 在 `How to Play Machines` 文档里，对实例生命周期写得比较明确：

- VIP 用户需要按需启动目标机器实例
- 机器实例有固定生命周期，到期后会自动关闭
- 平台支持延长实例时间，最长可延到 24 小时
- 用户完成当前机器后，需要主动点击 `Stop Machine`
- 平台不允许同时保有两个活跃机器实例

同一篇文档里，`Submitting found flags` 和 `Stopping a Machine` 是两个独立步骤。也就是说，官方把“提交 Flag”与“停止实例”设计成两条分离链路，而不是“解题成功后立刻自动停机”。

在 `How to Play Challenges` 文档中，HTB 还明确说明某些题型需要连接到 Docker 容器。这说明在挑战模式下，它同样存在“题目资源是运行时实例”的设计，只是帮助文档没有把生命周期策略展开到机器模式那样详细。

从 HTB 的公开帮助文档看，比较稳定的产品模式是：

- 有实例生命周期
- 允许延时
- 允许手动停止
- 提交 Flag 与停止实例解耦

来源：

- [Hack The Box Help - How to Play Machines](https://help.hackthebox.com/en/articles/5185338-how-to-play-machines)
- [Hack The Box Help - How to Play Challenges](https://help.hackthebox.com/en/articles/5185436-how-to-play-challenges)

### 3. TryHackMe：强调默认时长、手动关闭、关闭页面不等于停机，以及成本驱动的终止/休眠

TryHackMe 的资料对“资源仍在占用”这件事提示得更直接。

`The AttackBox Explained` 文档明确写到：

- AttackBox 默认会话时长是 2 小时
- 用户可以 `Extend time`
- 用户完成后可以 `Shut down`
- 关闭房间页面不会自动关闭 AttackBox

这说明 TryHackMe 把“页面退出”和“资源释放”明确拆开，并直接告诉用户：如果你只是离开页面，云端环境还在运行。

`Instance Termination & Hibernation` 文档进一步说明：

- 平台可能因为成本控制而终止或休眠实例
- 终止会丢失未保存进度，需要重新启动新实例
- 休眠会保留状态，恢复后从原位置继续
- 平台会通过弹窗通知用户实例被终止或休眠
- 用户可以在房间里使用 `Terminate Machine` 或 `Resume Machine`

TryHackMe 这组设计体现得很清楚：平台把实例管理当成一个显式的用户交互问题，而不是隐式后台细节。实例何时会结束、结束后会丢什么、怎样恢复，都有直接提示。

来源：

- [TryHackMe Help - The AttackBox Explained](https://help.tryhackme.com/en/articles/6721845-the-attackbox-explained)
- [TryHackMe Help - Instance Termination & Hibernation](https://help.tryhackme.com/en/articles/6498318-instance-termination-hibernation)

## 这几类平台的共性

结合上面的官方资料，这次抽样能看到几个比较稳定的共性。

### 1. 更常见的是“显式生命周期管理”，不是“解题即销毁”

至少在本次纳入的官方样本里，没有看到哪家把“提交正确 Flag 后立刻销毁实例”当成主流默认路径来宣传。更常见的是：

- 实例有固定 TTL
- 用户可以手动停止
- 部分平台支持延长时间
- 到期后自动关闭或被平台终止/休眠

### 2. Flag 提交与实例生命周期通常不是同一个动作

HTB 把 `submit flag` 和 `stop machine` 分开；THM 则把“完成学习任务”和“关闭 AttackBox/Terminate Machine”分开。这个设计说明平台默认认为：

- 解题成功不等于用户已经不需要环境
- 用户可能还要查看文件、复盘、验证 payload、写题解
- 因此实例回收通常需要一个单独的生命周期策略，而不是直接绑在“题目 solved”事件上

### 3. 平台会明确提醒用户资源仍在运行

TryHackMe 明确提示“关闭页面不会关闭 AttackBox”，也提供 `Shut down`、`Terminate Machine`、`Resume Machine` 等动作；HTB 也把 `Stop Machine` 和 `Extend` 做成显式操作。成熟平台不会假设用户天然知道云端实例的状态，而是主动暴露：

- 当前是否仍在运行
- 还能运行多久
- 是否可以延时
- 用户现在该点击哪里关闭

### 4. 自动回收通常服务于资源与成本控制

THM 直接把终止/休眠和成本控制联系起来；HTB 则通过实例寿命和活跃实例数限制来约束资源。也就是说，生命周期设计不只是后端运维问题，还是平台产品层面的资源分配机制。

## 对当前仓库的直接对照

当前仓库里，题目提交成功后并不会对实例生命周期做额外处理。

### 1. 当前提交成功链路

后端 `SubmitFlagWithContext(...)` 在判定正确后会：

- 记录正确提交
- 清理用户进度缓存
- 发布 `FlagAccepted` 事件
- 返回“恭喜你，Flag 正确！”以及分数

但这里没有看到对实例 `expires_at` 的收缩，也没有实例关闭或排队清理动作。

相关位置：

- `code/backend/internal/module/practice/application/commands/service.go`
- `code/backend/internal/dto/submission.go`

前端 `submitFlagHandler()` 在提交成功后目前只会：

- 弹出 `Flag 正确！`
- 把当前题目标记为已解决
- 加载题解

没有看到“实例将在 N 分钟后自动关闭”之类的反馈。

相关位置：

- `code/frontend/src/composables/useChallengeDetailInteractions.ts`

### 2. 当前实例生命周期

当前实例本身已经有 TTL 和后台清理机制：

- `container.default_ttl: 2h`
- `container.cleanup_interval: "*/5 * * * *"`
- 后台维护任务会扫描过期实例并清理运行时资源，然后把状态更新为 `expired`

相关位置：

- `code/backend/configs/config.yaml`
- `code/backend/internal/module/runtime/application/commands/runtime_maintenance_service.go`

这意味着当前平台并不是完全没有生命周期能力，而是：

- 已有“统一 TTL + 定时清理”
- 还没有“解题成功后收口生命周期”的策略
- 也缺少面向学生的明确提示

## 对本项目的启发

结合外部样本和当前仓库状态，更合理的方向不是“Flag 正确后立刻硬关闭实例”，而是“解题后进入一个短暂缓冲期，然后自动回收”。

推荐原因有三点：

- 这更接近 HTB / THM 这类平台把“解题”和“资源回收”分开的做法
- 学生在解题成功后通常还需要几分钟查看环境、核对利用链、整理题解
- 当前仓库已经有 `expires_at + cleanup job` 基础，只需要补一层“成功后收缩 TTL”和前端提示，不需要重做整个运行时模型

如果后续要往实现推进，比较自然的产品语义会是：

- 题目首次解出后，将该实例的剩余生命周期收缩到一个较短的 `grace period`
- 前端明确提示“题目已解决，实例将在 10 分钟后自动关闭”
- 在实例页或题目页持续显示倒计时
- 如果实例实际启动失败、崩溃或已过期，应直接展示状态，而不是继续给用户一个看似可访问的地址

这样既能减少资源占用，也符合毕业设计里“集中监控与管理”“记录过程并反馈”的目标。

## 结论

本次抽样没有发现“主流 CTF 平台普遍在解题成功后立即销毁实例”的证据。更常见的产品模式是：

- 固定 TTL
- 允许手动停止
- 支持延时或恢复
- 到期后自动关闭、终止或休眠
- 通过显式提示告诉用户实例仍在运行以及何时结束

对当前项目来说，最稳妥的方向不是把实例关闭绑定成一个瞬时动作，而是把“解题成功”变成“实例生命周期收口”的触发点，再通过清晰的 UI 提示把这个变化告诉学生。
