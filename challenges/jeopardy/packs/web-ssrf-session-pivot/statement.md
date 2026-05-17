这是一个给内部员工用的支持台。普通用户可以登录后使用“抓取器”拉取指定 relay 的页面片段，管理员才能导出审计数据。

题目里至少有两层问题，但没有任何现成管理员口令。

## 目标

1. 以普通用户身份观察抓取器的 URL 校验逻辑。
2. 找到一种方式访问只允许本机调用的内部接口。
3. 恢复会话签名材料，伪造管理员会话。
4. 访问导出接口并拿到动态 Flag。

## 访问方式

- 容器服务端口：`8080/http`
- 本地测试示例：

```bash
docker build -t test-web-ssrf-session-pivot docker
docker run --rm -p 18080:8080 -e FLAG='flag{local_test}' test-web-ssrf-session-pivot
curl http://127.0.0.1:18080/
```
