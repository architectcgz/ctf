# 解法

这题是两段链：

1. 利用抓取器的 SSRF
2. 拿到会话签名密钥后伪造管理员 Cookie

抓取器只做了很弱的字符串前缀检查：

```text
url.startswith("http://relay.local")
```

它没有真正解析主机，所以可以用带 `userinfo` 的 URL：

```text
http://relay.local@127.0.0.1:8080/internal/secret
```

这条 URL 仍然满足前缀检查，但真实连接目标已经变成了 `127.0.0.1:8080`。因为请求是服务端自己发起的，所以内部接口会把 `SESSION_SECRET` 返回出来。

拿到 secret 以后，直接按题目现有会话格式伪造：

```text
base64url(json_payload).hex_hmac_sha256
```

把 `role` 改成 `admin`，再带着伪造 Cookie 请求 `/admin/export`，就会返回动态 Flag。
