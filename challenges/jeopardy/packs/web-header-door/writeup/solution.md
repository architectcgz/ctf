# Web 入门：门禁请求头题解

## 解法

1. 访问首页，提示要先看 crawler rules。
2. 访问 `/robots.txt`，发现隐藏路径：

```text
Disallow: /staff-door
```

3. 访问 `/staff-door`，页面返回缺少请求头：

```text
Required header: X-CTF-Token: open-sesame
```

4. 使用 curl 添加请求头：

```bash
curl -H 'X-CTF-Token: open-sesame' http://<host>:<port>/staff-door
```

得到 flag：

```text
flag{web_header_door_opened}
```
