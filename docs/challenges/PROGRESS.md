# CTF 题目包创建进度

> 更新时间：2026-03-03

---

## 已完成题目（5 道）

### Web 安全类（3 道）

1. **web-sqli-login-bypass** ✅
   - SQL 注入登录绕过
   - 包含：PHP + MySQL 完整代码
   - ZIP: `web-sqli-login-bypass.zip`

2. **web-cookie-edit** ✅
   - Cookie 篡改获取管理员权限
   - 包含：PHP 完整代码
   - ZIP: `web-cookie-edit.zip`

3. **web-robots-txt** ✅
   - robots.txt 信息泄露
   - 包含：静态 HTML + Nginx
   - ZIP: `web-robots-txt.zip`

### Web 安全类（2 道 - 之前创建）

4. **web-xss-reflected** ✅
   - 反射型 XSS
   - 包含：PHP 完整代码

### 密码学类（1 道）

5. **crypto-base64-decode** ✅
   - Base64 解码
   - 包含：附件文件

---

## 文件位置

所有题目包位于：`/home/azhi/workspace/projects/ctf/docs/challenges/packs/`

每个题目包包含：
- `manifest.yml` - 题目元数据
- `statement.md` - 题面描述
- `docker/Dockerfile` - 容器构建文件
- `docker/src/` - 源代码
- `{题目名}.zip` - 打包文件（可直接上传平台）

---

## 下一步建议

考虑到 300 道题目的工作量，建议：

1. **优先级方案**：告诉我你最需要哪些类别/难度的题目，我优先创建
2. **模板复用**：基于已有的 5 道题目模板，你可以快速修改生成其他题目
3. **分批创建**：我可以继续每次创建 5-10 道完整题目

---

## 题目列表文档

已创建 6 个类别的题目列表（每个 50 题），包含：
- `challenge-list-web.md` - Web 50 题
- `challenge-list-crypto.md` - 密码学 50 题
- `challenge-list-reverse.md` - 逆向 50 题
- `challenge-list-pwn.md` - Pwn 50 题
- `challenge-list-forensics.md` - 取证 50 题
- `challenge-list-misc.md` - 杂项 50 题

每道题都有标题、描述、难度、标签等元数据，可作为创建完整题目包的参考。
