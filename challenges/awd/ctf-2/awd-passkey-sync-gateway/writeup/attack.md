# Attack Writeup

## 漏洞点

业务协议里保留了 `EXPORT <support_key>` 调试命令。当前实现把默认支持密钥硬编码成了固定常量 `sync-support-2026`，只要知道这串值，就能把当前快照直接导出。

## 利用思路

1. 连接 TCP 服务并确认 `PING` 返回 `PONG`。
2. 发送：

```text
EXPORT sync-support-2026
```

3. 服务会直接返回：

```text
SNAPSHOT flag{...}
```

## 关键边界

- `SET_FLAG` / `GET_FLAG` 属于 checker 私有管理通道，不应作为选手攻击面。
- 攻击面应聚焦 `docker/workspace/src/challenge_app.py` 中的业务命令处理。
