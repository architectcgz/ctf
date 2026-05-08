# 防守题解

## 可修改位置

- 平台防守入口对应的业务代码在 `/workspace/src/challenge_app.py`
- 题目包内源文件位置是 `docker/workspace/src/challenge_app.py`
- 受保护的 `runtime` 与 `check` 代码不要动

## 主要漏洞点

### 1. URL 黑名单只看字符串

当前实现只检查 `target_url` 里是否直接出现 `127.0.0.1 / localhost / 0.0.0.0`，这对十进制 IP、IPv6、DNS rebinding 风格的写法都不可靠。

推荐修法：

- 先用 `urllib.parse` 拆出 hostname
- 做 DNS 解析
- 显式拒绝回环地址、私有地址和链路本地地址
- 最稳妥的是改成“只允许访问你维护的固定 allowlist 域名”

### 2. 内部调试接口没有额外隔离

`/internal/snapshot` 只靠 `remote_addr` 限制，一旦 SSRF 成功就会把 Flag 直接回显。

推荐修法：

- 直接删除这个调试接口
- 或者至少让它不返回敏感数据

## 推荐修改函数

- `preview()`
- `internal_snapshot()`

## 保活约束

- `/health` 正常返回 200
- `/api/flag` 继续支持 checker 的 PUT / GET
- `/manifest/demo` 仍然可用
- 预览入口仍可抓取合法外部地址

## 交付判断

1. 十进制 / 变体地址不能再绕过本地拦截
2. 内部调试接口不再直接暴露 Flag
3. `/health` 和 `/api/flag` 仍保持可用
