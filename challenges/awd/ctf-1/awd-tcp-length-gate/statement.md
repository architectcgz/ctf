# TCP 长度门禁

这是一个 AWD 模式的原生 TCP 服务。服务提供简单命令协议，队伍需要保持服务可用，同时修复长度门禁逻辑中的弱点。

## 连接方式

启动实例后使用平台提供的 TCP 连接命令访问服务。

## 协议摘要

```text
PING
SET_FLAG <flag>
GET_FLAG
CHECK <payload>
```

平台 checker 会使用 `SET_FLAG` 写入本轮 flag，再通过 `GET_FLAG` 回读校验。选手攻击面位于 `CHECK` 命令，满足隐藏条件时服务会返回当前保存的 flag。
