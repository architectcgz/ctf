# CTF 题目包列表

> 版本：v1.0 | 日期：2026-03-03
> 格式：challenge-pack-v1

---

## 已创建题目

### 1. web-sqli-login-bypass - SQL 注入：登录绕过

**分类**：Web 安全
**难度**：⭐☆☆☆☆ (beginner)
**知识点**：SQL 注入、登录绕过、SQL 注释符
**容器**：PHP + MySQL
**来源**：基于 [picoCTF Web Exploitation](https://motasem-notes.net/web-hacking-101-with-picoctf-ctf-walkthrough/)

**题目描述**：
简单的登录系统存在 SQL 注入漏洞，需要绕过登录验证以管理员身份登录。

**解题思路**：
- 在用户名输入 `admin' OR '1'='1'--`
- 密码随意输入
- 成功绕过验证获取 flag

---

### 2. web-xss-reflected - XSS：反射型跨站脚本

**分类**：Web 安全
**难度**：⭐☆☆☆☆ (beginner)
**知识点**：反射型 XSS、JavaScript 注入
**容器**：PHP
**来源**：基于常见 CTF XSS 题目

**题目描述**：
搜索页面未对用户输入进行过滤，存在反射型 XSS 漏洞。

**解题思路**：
- 在搜索框输入 `<script>alert(1)</script>`
- 触发 XSS 弹窗
- 查看页面源码获取 flag

---

### 3. crypto-base64-decode - 密码学：Base64 解码

**分类**：密码学
**难度**：⭐☆☆☆☆ (beginner)
**知识点**：Base64 编码识别与解码
**容器**：无（纯文本题）
**来源**：基于 [CTF 101](https://ctf101.org/)

**题目描述**：
下载附件 encoded.txt，对 Base64 编码的消息进行解码。

**解题思路**：
```bash
echo "ZmxhZ3tiYXNlNjRfaXNfZWFzeV90b19kZWNvZGV9" | base64 -d
# 输出：flag{base64_is_easy_to_decode}
```

---

## 题目包结构说明

每个题目包都遵循 `challenge-pack-v1` 规范，包含：

```
题目目录/
├── manifest.yml          # 题目元数据
├── statement.md          # 题面描述
├── attachments/          # 附件（可选）
└── docker/              # 容器构建上下文（可选）
    ├── Dockerfile
    └── src/
```

---

## 使用方法

### 教师/管理员上传

1. 将题目目录打包为 Zip 文件
2. 在平台管理后台上传题目包
3. 平台自动校验并导入

### 学员练习

1. 浏览题目列表，选择感兴趣的题目
2. 点击"开始挑战"启动容器实例
3. 访问分配的靶机地址
4. 提交 flag 获取分数

---

## 参考资源

- [picoCTF Web Exploitation](https://motasem-notes.net/web-hacking-101-with-picoctf-ctf-walkthrough/)
- [PicoCTF 2024 Writeups](https://blog.qz.sg/picoctf-2024-web-exploitation-writeups/)
- [CTF 101 - SQL Injection](https://ctf101.org/web-exploitation/sql-injection/what-is-sql-injection/)
- [Web Exploitation Techniques](https://toxigon.com/web-exploitation-techniques-for-ctf)
