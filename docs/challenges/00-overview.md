# CTF 题目库总览

> 版本：v1.0 | 日期：2026-03-03
> 格式：challenge-pack-v1
> 总计：300 道题目（6 个类别 × 50 题）

---

## 题目分类统计

| 类别 | 英文 | 题目数 | 难度分布 |
|------|------|--------|----------|
| Web 安全 | Web Exploitation | 50 | 入门15 / 简单15 / 中等12 / 困难6 / 地狱2 |
| 密码学 | Cryptography | 50 | 入门15 / 简单15 / 中等12 / 困难6 / 地狱2 |
| 逆向工程 | Reverse Engineering | 50 | 入门15 / 简单15 / 中等12 / 困难6 / 地狱2 |
| 二进制利用 | Pwn/Binary Exploitation | 50 | 入门15 / 简单15 / 中等12 / 困难6 / 地狱2 |
| 取证分析 | Forensics | 50 | 入门15 / 简单15 / 中等12 / 困难6 / 地狱2 |
| 杂项 | Miscellaneous | 50 | 入门15 / 简单15 / 中等12 / 困难6 / 地狱2 |

**总计**：300 道题目

---

## 文件结构

```
ctf/docs/challenges/
├── README.md                           # 本文件
├── challenge-list-web.md               # Web 安全 50 题
├── challenge-list-crypto.md            # 密码学 50 题
├── challenge-list-reverse.md           # 逆向工程 50 题
├── challenge-list-pwn.md               # 二进制利用 50 题
├── challenge-list-forensics.md         # 取证分析 50 题
├── challenge-list-misc.md              # 杂项 50 题
├── 01-ctf-platforms.md                 # 平台列表
├── 02-ctf-categories.md                # 分类说明
├── 03-beginner-challenges.md           # 初学者示例
└── packs/                              # 题目包示例
    ├── web-sqli-login-bypass/
    ├── web-xss-reflected/
    └── crypto-base64-decode/
```

---

## 快速导航

### 按难度选择

- **入门级 (beginner)**：适合 CTF 新手，0-3 个月经验
  - 每个类别 15 题，共 90 题
  - 基础工具使用、简单漏洞利用

- **简单级 (easy)**：适合有基础的选手，3-6 个月经验
  - 每个类别 15 题，共 90 题
  - 常见攻击技术、工具组合使用

- **中等级 (medium)**：适合进阶选手，6-12 个月经验
  - 每个类别 12 题，共 72 题
  - 复杂利用链、绕过防护机制

- **困难级 (hard)**：适合高级选手，1 年以上经验
  - 每个类别 6 题，共 36 题
  - 高级技术、真实场景模拟

- **地狱级 (hell)**：适合专家级选手
  - 每个类别 2 题，共 12 题
  - 0day 利用、完整攻击链

### 按类别选择

1. **Web 安全** - [查看详情](./challenge-list-web.md)
   - SQL 注入、XSS、SSRF、反序列化等
   - 适合 Web 开发者和渗透测试人员

2. **密码学** - [查看详情](./challenge-list-crypto.md)
   - 古典密码、RSA、AES、哈希碰撞等
   - 适合对数学和密码学感兴趣的选手

3. **逆向工程** - [查看详情](./challenge-list-reverse.md)
   - 二进制分析、反汇编、反混淆等
   - 适合对底层原理感兴趣的选手

4. **二进制利用** - [查看详情](./challenge-list-pwn.md)
   - 栈溢出、堆利用、ROP 链等
   - 适合对漏洞利用感兴趣的选手

5. **取证分析** - [查看详情](./challenge-list-forensics.md)
   - 文件分析、内存取证、流量分析等
   - 适合对数字取证感兴趣的选手

6. **杂项** - [查看详情](./challenge-list-misc.md)
   - OSINT、编程、区块链、AI 等
   - 适合综合能力强的选手

---

## 使用说明

### 教师/管理员

1. **选择题目**：从题目列表中选择合适的题目
2. **制作题目包**：按照 `challenge-pack-v1` 规范制作
3. **上传平台**：通过管理后台上传题目包
4. **发布使用**：配置到练习或竞赛中

### 学员

1. **浏览题目**：在平台上浏览可用题目
2. **开始挑战**：点击"开始挑战"启动实例
3. **解题提交**：找到 flag 后提交验证
4. **查看进度**：跟踪个人解题进度

---

## 题目来源

所有题目基于以下平台和资源整理：

- [picoCTF](https://picoctf.org/) - 教育导向的 CTF 平台
- [CTF101](https://ctf101.org/) - CTF 入门教程
- [CryptoHack](https://cryptohack.org/) - 密码学专项平台
- [pwnable.kr](https://pwnable.kr/) - 二进制利用平台
- [pwnable.tw](https://pwnable.tw/) - 高级 Pwn 挑战
- [Crackmes.one](https://crackmes.one/) - 逆向工程挑战
- [HackerDNA CTF Guide](https://hackerdna.com/blog/ctf-categories)
- [CTF for Beginners 2026](https://hackerdna.com/blog/ctf-for-beginners)

---

## 下一步计划

- [ ] 为每道题目创建完整的题目包
- [ ] 编写详细的 Writeup
- [ ] 制作视频讲解
- [ ] 建立题目难度评级系统
- [ ] 收集用户反馈优化题目

---

## 参考资源

- [picoCTF Web Exploitation](https://motasem-notes.net/web-hacking-101-with-picoctf-ctf-walkthrough/)
- [CTF Categories Guide](https://hackerdna.com/blog/ctf-categories)
- [Web Exploitation Techniques](https://toxigon.com/web-exploitation-techniques-for-ctf)
- [CTF 101 - SQL Injection](https://ctf101.org/web-exploitation/sql-injection/what-is-sql-injection/)
