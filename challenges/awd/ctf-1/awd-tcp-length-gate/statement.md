# TCP 长度门禁

这是一个 AWD 模式的原生 TCP 服务。服务提供简单命令协议，队伍需要保持服务可用，同时修复长度门禁逻辑中的弱点。

## 连接方式

启动实例后使用平台提供的 TCP 连接命令访问服务。

## 协议摘要

```text
PING
CHECK <payload>
```

公开攻击面只保留 `PING` 和 `CHECK`。平台 checker 还会通过私有管理通道写入并回读本轮 flag，但该通道不属于选手防守范围，也不会下发到工作区。

选手可修改的业务逻辑位于工作区挂载的 `docker/workspace/src/challenge_app.py`。题目的目标是保留服务可用，同时让 `CHECK` 不再泄露当前保存的 flag。
