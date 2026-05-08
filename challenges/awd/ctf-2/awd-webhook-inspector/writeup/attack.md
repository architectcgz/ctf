# Attack Writeup

## 漏洞点

业务入口 `/preview?url=` 只在原始 URL 字符串里黑名单拦截：

- `127.0.0.1`
- `localhost`
- `0.0.0.0`

但实际请求是由服务端发起的，只要构造一个“字符串里不像本地地址、解析后仍然落到回环地址”的 URL，就可以让它代我们访问内部调试接口。

## 利用思路

1. 访问 `/manifest/demo` 确认预览器行为正常。
2. 构造十进制回环地址：`2130706433`。
3. 请求：

```text
/preview?url=http://2130706433:8080/internal/snapshot
```

4. 从返回内容中提取当前实例的动态 Flag。

## 关键边界

- `/api/flag` 是 checker 私有通道，不应作为选手攻击面。
- 攻击面应聚焦 `docker/workspace/src/challenge_app.py` 中的业务预览逻辑。
