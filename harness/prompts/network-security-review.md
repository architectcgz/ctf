# 网络安全防护审查提示词（CTF 入口）

本提示词用于审查 CTF 平台代码、配置、容器编排和前端依赖中的网络安全防护风险。它是项目本地入口；跨项目通用方法优先沉淀到 `security-vulnerability-scan` skill。

## 参考来源

- OWASP Top 10：访问控制、加密失败、注入、安全配置错误、易受攻击组件等高频风险分类。
- OWASP Web Security Testing Guide：认证、会话、授权、输入验证、配置、客户端安全的测试组织方式。
- OWASP Application Security Verification Standard：把认证、会话、访问控制、输入验证、文件、API、配置等拆成可验证控制项。
- OWASP Cheat Sheet Series：CSRF、XSS、SSRF、SQL Injection、File Upload 等专题防护要点。

## 使用前置

1. 先读取 `AGENTS.md`、`docs/文档规范.md`、`harness/prompts/AGENTS.md`。
2. 明确本轮目标是平台安全审查，不是题目漏洞审查。
3. 默认排除这些有意脆弱内容，除非任务明确要求审题：
   - `challenges/`
   - `code/backend/data/challenge-attachments/`
   - `docs/contracts/examples/`
   - `scripts/challenges/`
4. 如果发现真实平台风险，先区分：
   - 可立即修复的代码缺陷
   - 需要进入 `docs/todos/2026-06-02-security-review-findings.md` 的未收口安全项
   - 仅由缺少扫描工具造成的证据缺口

## 提示词

````text
你是资深应用安全审查者。请审查本地仓库 `/home/azhi/workspace/projects/ctf` 的网络安全防护状态，目标是找出真实平台风险，而不是把 CTF 题目中故意保留的漏洞当成平台漏洞。

审查前必须读取：
1. `/home/azhi/workspace/projects/ctf/AGENTS.md`
2. `/home/azhi/workspace/projects/ctf/docs/todos/2026-06-02-security-review-findings.md`
3. `/home/azhi/workspace/projects/ctf/code/backend/configs/config.yaml`
4. `/home/azhi/workspace/projects/ctf/code/backend/configs/config.dev.yaml`
5. `/home/azhi/workspace/projects/ctf/code/backend/configs/config.prod.yaml`
6. `/home/azhi/workspace/projects/ctf/docker/docker-compose.dev.yml`
7. `/home/azhi/workspace/projects/ctf/code/frontend/package.json`
8. `/home/azhi/workspace/projects/ctf/code/frontend/pnpm-lock.yaml`

默认排除：
- `challenges/`
- `code/backend/data/challenge-attachments/`
- `docs/contracts/examples/`
- `scripts/challenges/`
- `code/frontend/node_modules/`
- `code/frontend/dist/`
- `code/backend/tmp/`
- `code/backend/.tmp/`
- `output/`

先运行 bounded 扫描：

```bash
timeout 180 bash ~/.agents/skills/security-vulnerability-scan/scripts/run_security_vulnerability_scan.sh /home/azhi/workspace/projects/ctf
```

如果审查前端依赖，以项目实际包管理器为准再运行：

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend && pnpm audit --json
```

人工复核必须覆盖这些面：

1. 认证、会话、Cookie、CSRF、CORS
   - session cookie 是否 `HttpOnly`、生产是否 `Secure`、`SameSite` 是否符合跨站请求风险。
   - state-changing route 是否只在认证分组下，或是否具备等价 ticket / Origin / Referer / CSRF token 防护。
   - CORS 是否存在 `allow_credentials=true` 与空白名单、通配符、Origin 反射组合。
   - WebSocket 是否绕过 HTTP auth；如果绕过，handler 内是否消费短期一次性 ticket。

2. 授权与对象级访问控制
   - 学生、教师、平台管理员三类角色是否通过中间件和 service 层共同约束。
   - `contest_id`、`team_id`、`instance_id`、`service_id`、`user_id` 是否校验当前用户可访问范围。
   - 下载、导出、实例代理、AWD defense workbench、通知、报表等是否存在 IDOR。

3. 注入与危险执行
   - SQL 是否使用参数化；`fmt.Sprintf` 拼接 SQL 时参数是否严格来自白名单常量。
   - Docker、iptables、SSH、容器 exec、脚本运行是否有命令参数白名单和上下文隔离。
   - 模板、Markdown、HTML、PDF 生成是否存在模板注入或未净化输出。

4. 文件、归档与路径
   - 题包导入、附件下载、writeup、导出包、AWD defense 文件读写是否限制根目录。
   - zip/tar 解包是否防 zip slip、过大文件、过多文件、隐藏敏感路径。
   - `.env`、SSH key、Docker socket、`/proc`、`/sys`、`/run/secrets` 是否被读取或写入。

5. SSRF、出站请求与代理
   - 后端或题包导入流程是否根据用户输入请求 URL。
   - Registry、CAS、runtime-agent、Docker daemon、instance proxy、webhook 相关 URL 是否有 scheme / host / IP 段 allowlist。
   - 是否防止 loopback、link-local、private CIDR、IPv6、DNS rebinding、重定向绕过。

6. 容器、网络和部署配置
   - `docker.sock` 挂载、privileged、host network、root 用户、volume、端口绑定是否只限本机开发。
   - 生产配置是否禁止默认密码、`change_me`、弱 secret、调试端口、过宽 CORS。
   - runtime agent / Docker API / registry 是否具备 TLS、认证、最小权限与网络分段。

7. 前端 XSS、依赖和供应链
   - `v-html` 是否只消费经过 DOMPurify 或等价净化的 HTML。
   - Markdown 渲染后是否再净化，链接协议和 `target` 是否安全。
   - `pnpm audit` 中生产依赖、开发工具、测试 UI 的风险是否按真实暴露面分级。
   - 如果 `package-lock.json` 与 `pnpm-lock.yaml` 漂移，说明 audit 证据以哪个锁文件为准。

输出格式：

## 结论
说明是否发现新增可信平台漏洞；如果只有已知待办或工具盲区，直接说明。

## Findings
按严重程度排序。每条必须包含：
- Severity
- 文件和行号
- 信任边界
- 攻击路径
- 为什么不是题目有意漏洞
- 修复方向
- 建议验证命令

## 已知待办对齐
列出现有 `docs/todos/2026-06-02-security-review-findings.md` 中仍适用的项目，不重复伪造新 finding。

## 扫描证据
列出实际执行的命令、结果目录、跳过的工具和原因。

## 盲区
列出未安装工具、未启动 DAST、本轮未覆盖目录、需要人工复核的运行时前提。

要求：
- 不要把 `challenges/`、题包附件、解题脚本里的漏洞算作平台漏洞。
- 不要只贴扫描器输出；必须确认可达入口、攻击者可控输入和真实 sink。
- 不要把 dev-only 依赖漏洞直接升级为生产漏洞；必须说明暴露条件。
- 如果发现平台高风险问题，优先给最小修复和最小验证，而不是泛泛建议“加强安全”。
````

## 本仓库重点提示

- WebSocket 入口挂在 `engine.GET("/ws/...")` 时，要继续确认 handler 内是否消费 `/api/v1/auth/ws-ticket` 签发的短期一次性 ticket。
- `docker/docker-compose.dev.yml` 面向本机开发；`docker.sock` 和固定开发密码不能被解释成生产安全边界。
- 前端依赖审查以 `pnpm-lock.yaml` 为准；`package-lock.json` 的 audit 结果只能作为漂移信号。
- `v-html` 风险要从渲染点反查到 `useSanitize()`，确认 Markdown 先 render、后 sanitize。
- `fmt.Sprintf` 生成 SQL 只有在输入来自受控白名单或常量时才可降级；否则按 SQL 注入候选处理。
