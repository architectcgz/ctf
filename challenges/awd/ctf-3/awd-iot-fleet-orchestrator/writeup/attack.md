# Attack Writeup

## 漏洞点

这题的公网入口 `fleet-web` 会把用户给出的 `device / key / op` 转发给内部 `fleet-agent`。而 `fleet-agent` 允许使用固定支持 key `fleet-support-2026`。

## 利用思路

1. 访问自己的实例源码或业务逻辑，可以确认 `fleet-agent` 存在固定支持 key：

```text
fleet-support-2026
```

2. 通过公网入口调用：

```text
/dispatch?device=demo-sensor&key=fleet-support-2026&op=pull-config
```

3. `fleet-web` 会把请求转发给 `fleet-agent`。
4. `fleet-agent` 接受默认 key，并继续访问 `config-vault` 读取设备配置。
5. `config-vault` 返回的数据里直接包含当前动态 Flag。

## 关键边界

- `/api/flag` 是 checker 私有通道，不应作为选手攻击面。
- 攻击面应聚焦 `docker/workspace/src/challenge_app.py` 中默认 key 与内部调度链路。
- 这题需要 `fleet-web -> fleet-agent -> config-vault` 的真实拓扑边界；压成单容器后，控制面与配置面的边界会失真。
