# Attack Writeup

## 漏洞点

业务入口在 `CHECK <payload>`。当前实现会根据长度和关键字进入错误分支，并把运行中的 flag 直接拼回响应里。

## 利用思路

1. 连接 TCP 服务并确认 `PING` 返回 `PONG`。
2. 构造满足长度门禁条件的 `CHECK` 请求。
3. 从返回内容中提取 flag。

## 关键边界

- `SET_FLAG` / `GET_FLAG` 属于 checker 私有管理通道，不应作为选手攻击面。
- 攻击应聚焦 `challenge_app.py` 中的业务判断与回显逻辑，而不是修改受保护的 runtime / checker 代码。
