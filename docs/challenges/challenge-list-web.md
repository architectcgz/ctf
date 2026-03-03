# CTF 题目库 - Web 安全类（50题）

> 来源：基于 picoCTF、CTF101 等平台题目整理
> 格式：challenge-pack-v1

---

## 入门级 (beginner) - 15题

### 1. web-view-source
- **标题**：查看页面源代码
- **描述**：flag 隐藏在 HTML 注释中
- **标签**：`kp:html-source`, `kp:devtools`
- **容器**：静态 HTML

### 2. web-cookie-edit
- **标题**：Cookie 篡改
- **描述**：修改 Cookie 中的 role 字段获取管理员权限
- **标签**：`vuln:auth-bypass`, `kp:cookie`
- **容器**：PHP

### 3. web-robots-txt
- **标题**：robots.txt 信息泄露
- **描述**：检查 robots.txt 找到隐藏路径
- **标签**：`kp:recon`, `kp:robots`
- **容器**：静态 HTML

### 4. web-sqli-login-bypass
- **标题**：SQL 注入登录绕过
- **描述**：使用 OR 1=1 绕过登录验证
- **标签**：`vuln:sqli`, `kp:auth-bypass`
- **容器**：PHP + MySQL

### 5. web-xss-reflected
- **标题**：反射型 XSS
- **描述**：搜索框未过滤，注入 script 标签
- **标签**：`vuln:xss`, `kp:reflected-xss`
- **容器**：PHP

### 6. web-directory-listing
- **标题**：目录遍历
- **描述**：Web 服务器开启了目录列表，浏览文件找 flag
- **标签**：`vuln:info-disclosure`, `kp:directory-listing`
- **容器**：Nginx

### 7. web-hidden-input
- **标题**：隐藏表单字段
- **描述**：表单中有 hidden input 包含 flag
- **标签**：`kp:html-source`, `kp:form`
- **容器**：静态 HTML

### 8. web-js-source
- **标题**：JavaScript 源码分析
- **描述**：flag 硬编码在 JS 文件中
- **标签**：`kp:js-analysis`, `kp:devtools`
- **容器**：静态 HTML

### 9. web-get-parameter
- **标题**：GET 参数传递
- **描述**：通过 URL 参数传递特定值获取 flag
- **标签**：`kp:http-get`, `kp:url-params`
- **容器**：PHP

### 10. web-post-request
- **标题**：POST 请求
- **描述**：需要发送 POST 请求而非 GET
- **标签**：`kp:http-post`, `kp:burp`
- **容器**：PHP

### 11. web-user-agent
- **标题**：User-Agent 检测
- **描述**：修改 User-Agent 头通过验证
- **标签**：`kp:http-headers`, `kp:user-agent`
- **容器**：PHP

### 12. web-referer-check
- **标题**：Referer 检查绕过
- **描述**：伪造 Referer 头访问受限页面
- **标签**：`kp:http-headers`, `kp:referer`
- **容器**：PHP

### 13. web-weak-password
- **标题**：弱密码爆破
- **描述**：使用常见密码字典爆破登录
- **标签**：`vuln:weak-auth`, `kp:brute-force`
- **容器**：PHP + MySQL

### 14. web-base64-param
- **标题**：Base64 编码参数
- **描述**：URL 参数是 Base64 编码，解码后修改
- **标签**：`kp:base64`, `kp:encoding`
- **容器**：PHP

### 15. web-redirect
- **标题**：重定向跟踪
- **描述**：跟踪多次重定向找到最终 flag
- **标签**：`kp:http-redirect`, `kp:curl`
- **容器**：PHP

---

## 简单级 (easy) - 15题

### 16. web-sqli-union
- **标题**：SQL 注入 UNION 查询
- **描述**：使用 UNION SELECT 查询其他表数据
- **标签**：`vuln:sqli`, `kp:union-select`
- **容器**：PHP + MySQL

### 17. web-xss-stored
- **标题**：存储型 XSS
- **描述**：留言板存储 XSS，窃取管理员 Cookie
- **标签**：`vuln:xss`, `kp:stored-xss`
- **容器**：PHP + MySQL

### 18. web-file-upload
- **标题**：文件上传漏洞
- **描述**：绕过文件类型检查上传 PHP 文件
- **标签**：`vuln:file-upload`, `kp:bypass`
- **容器**：PHP

### 19. web-lfi-basic
- **标题**：本地文件包含
- **描述**：通过 LFI 读取 /etc/passwd
- **标签**：`vuln:lfi`, `kp:path-traversal`
- **容器**：PHP

### 20. web-command-injection
- **标题**：命令注入
- **描述**：ping 功能存在命令注入
- **标签**：`vuln:command-injection`, `kp:shell`
- **容器**：PHP

### 21. web-xxe-basic
- **标题**：XXE 外部实体注入
- **描述**：XML 解析器未禁用外部实体
- **标签**：`vuln:xxe`, `kp:xml`
- **容器**：PHP

### 22. web-csrf-basic
- **标题**：CSRF 跨站请求伪造
- **描述**：构造恶意页面触发管理员操作
- **标签**：`vuln:csrf`, `kp:token`
- **容器**：PHP

### 23. web-idor
- **标题**：越权访问
- **描述**：修改用户 ID 参数访问他人数据
- **标签**：`vuln:idor`, `kp:access-control`
- **容器**：PHP + MySQL

### 24. web-jwt-weak
- **标题**：JWT 弱密钥
- **描述**：JWT 使用弱密钥，可被爆破
- **标签**：`vuln:jwt`, `kp:weak-secret`
- **容器**：Node.js

### 25. web-ssti-basic
- **标题**：服务端模板注入
- **描述**：Jinja2 模板注入执行代码
- **标签**：`vuln:ssti`, `stack:flask`
- **容器**：Python Flask

### 26. web-ssrf-basic
- **标题**：SSRF 服务端请求伪造
- **描述**：利用 SSRF 访问内网服务
- **标签**：`vuln:ssrf`, `kp:internal-network`
- **容器**：PHP

### 27. web-deserialization
- **标题**：反序列化漏洞
- **描述**：PHP unserialize 导致代码执行
- **标签**：`vuln:deserialization`, `stack:php`
- **容器**：PHP

### 28. web-race-condition
- **标题**：条件竞争
- **描述**：并发请求绕过余额检查
- **标签**：`vuln:race-condition`, `kp:concurrency`
- **容器**：PHP + MySQL

### 29. web-open-redirect
- **标题**：开放重定向
- **描述**：利用重定向参数进行钓鱼
- **标签**：`vuln:open-redirect`, `kp:phishing`
- **容器**：PHP

### 30. web-path-traversal
- **标题**：路径遍历
- **描述**：使用 ../ 读取任意文件
- **标签**：`vuln:path-traversal`, `kp:file-read`
- **容器**：PHP

---

## 中等级 (medium) - 12题

### 31. web-sqli-blind
- **标题**：盲注 SQL 注入
- **描述**：基于时间的盲注提取数据
- **标签**：`vuln:sqli`, `kp:blind-sqli`
- **容器**：PHP + MySQL

### 32. web-nosql-injection
- **标题**：NoSQL 注入
- **描述**：MongoDB 查询注入
- **标签**：`vuln:nosql-injection`, `stack:mongodb`
- **容器**：Node.js + MongoDB

### 33. web-graphql-idor
- **标题**：GraphQL 越权
- **描述**：GraphQL 查询绕过权限检查
- **标签**：`vuln:idor`, `stack:graphql`
- **容器**：Node.js

### 34. web-prototype-pollution
- **标题**：原型链污染
- **描述**：JavaScript 原型链污染导致权限提升
- **标签**：`vuln:prototype-pollution`, `stack:nodejs`
- **容器**：Node.js

### 35. web-jwt-algorithm-confusion
- **标题**：JWT 算法混淆
- **描述**：将 RS256 改为 HS256 伪造 token
- **标签**：`vuln:jwt`, `kp:algorithm-confusion`
- **容器**：Node.js

### 36. web-oauth-redirect
- **标题**：OAuth 重定向劫持
- **描述**：劫持 OAuth 回调获取 token
- **标签**：`vuln:oauth`, `kp:redirect-uri`
- **容器**：PHP

### 37. web-xml-bomb
- **标题**：XML 炸弹
- **描述**：构造 XML 炸弹导致 DoS
- **标签**：`vuln:xxe`, `kp:xml-bomb`
- **容器**：Java

### 38. web-type-juggling
- **标题**：PHP 类型混淆
- **描述**：利用 PHP 弱类型比较绕过验证
- **标签**：`vuln:type-juggling`, `stack:php`
- **容器**：PHP

### 39. web-mass-assignment
- **标题**：批量赋值漏洞
- **描述**：修改不应暴露的对象属性
- **标签**：`vuln:mass-assignment`, `kp:parameter-pollution`
- **容器**：Ruby on Rails

### 40. web-http-smuggling
- **标题**：HTTP 请求走私
- **描述**：利用前后端解析差异走私请求
- **标签**：`vuln:http-smuggling`, `kp:cl-te`
- **容器**：Nginx + Gunicorn

### 41. web-cache-poisoning
- **标题**：缓存投毒
- **描述**：污染 CDN 缓存注入恶意内容
- **标签**：`vuln:cache-poisoning`, `kp:cdn`
- **容器**：Nginx + Varnish

### 42. web-websocket-hijack
- **标题**：WebSocket 劫持
- **描述**：劫持 WebSocket 连接窃取数据
- **标签**：`vuln:websocket`, `kp:hijacking`
- **容器**：Node.js

---

## 困难级 (hard) - 6题

### 43. web-sqli-waf-bypass
- **标题**：WAF 绕过 SQL 注入
- **描述**：绕过 ModSecurity 规则进行注入
- **标签**：`vuln:sqli`, `kp:waf-bypass`
- **容器**：PHP + MySQL + ModSecurity

### 44. web-ssti-sandbox-escape
- **标题**：SSTI 沙箱逃逸
- **描述**：绕过 Jinja2 沙箱限制执行命令
- **标签**：`vuln:ssti`, `kp:sandbox-escape`
- **容器**：Python Flask

### 45. web-dom-xss
- **标题**：DOM 型 XSS
- **描述**：纯前端 XSS，需分析 JavaScript 逻辑
- **标签**：`vuln:xss`, `kp:dom-xss`
- **容器**：静态 HTML + JS

### 46. web-csp-bypass
- **标题**：CSP 绕过
- **描述**：绕过内容安全策略执行 XSS
- **标签**：`vuln:xss`, `kp:csp-bypass`
- **容器**：PHP

### 47. web-polyglot-upload
- **标题**：多语言文件上传
- **描述**：构造同时是图片和 PHP 的文件
- **标签**：`vuln:file-upload`, `kp:polyglot`
- **容器**：PHP

### 48. web-advanced-ssrf
- **标题**：高级 SSRF
- **描述**：绕过黑名单访问云元数据服务
- **标签**：`vuln:ssrf`, `kp:cloud-metadata`
- **容器**：Python

---

## 地狱级 (hell) - 2题

### 49. web-0day-exploit
- **标题**：框架 0day 利用
- **描述**：利用真实框架漏洞（如 Spring4Shell）
- **标签**：`vuln:rce`, `kp:0day`
- **容器**：Java Spring

### 50. web-full-chain
- **标题**：完整攻击链
- **描述**：从信息收集到 RCE 的完整渗透
- **标签**：`kp:pentest`, `kp:full-chain`
- **容器**：多容器网络拓扑

---

## 参考来源

- [picoCTF Web Exploitation](https://motasem-notes.net/web-hacking-101-with-picoctf-ctf-walkthrough/)
- [CTF Categories Guide](https://hackerdna.com/blog/ctf-categories)
- [CTF 101](https://ctf101.org/web-exploitation/sql-injection/what-is-sql-injection/)
