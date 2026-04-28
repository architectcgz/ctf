# CTF 初学者题目示例

> 版本：v1.0 | 日期：2026-03-03 | 状态：整理中
> 适用对象：CTF 新手（0-3 个月）

---

## 0. 说明

本文档收集适合初学者的 CTF 题目示例，包括题目描述、解题思路和关键知识点。

---

## 1. Web 安全入门题

### 题目 1.1：SQL 注入基础

**难度**：⭐☆☆☆☆

**题目描述**
- 登录页面，输入用户名和密码
- 目标：绕过登录验证

**知识点**
- SQL 注入原理
- 常见 Payload：`' OR '1'='1`
- 注释符：`--`、`#`

**解题思路**
1. 尝试输入 `admin' OR '1'='1'--` 作为用户名
2. 密码随意输入
3. 后端 SQL 变为：`SELECT * FROM users WHERE username='admin' OR '1'='1'--' AND password='xxx'`
4. `OR '1'='1'` 永远为真，`--` 注释掉后面的密码验证
5. 成功登录，获取 flag

**防御措施**
- 使用参数化查询（Prepared Statement）
- 输入验证和过滤
- 最小权限原则

---

### 题目 1.2：查看页面源代码

**难度**：⭐☆☆☆☆

**题目描述**
- 一个简单的网页
- 提示：flag 就在这个页面上

**知识点**
- HTML 源代码查看
- 浏览器开发者工具
- 注释中的信息

**解题思路**
1. 右键 → 查看网页源代码（或按 Ctrl+U）
2. 在 HTML 注释中找到 flag：`<!-- flag{hidden_in_source} -->`
3. 或者在 JavaScript 代码中找到 flag

**扩展**
- 检查 robots.txt
- 检查 sitemap.xml
- 查看 Cookie 和 LocalStorage

---

### 题目 1.3：Cookie 篡改

**难度**：⭐⭐☆☆☆

**题目描述**
- 登录后显示 "You are guest"
- 目标：获取管理员权限

**知识点**
- Cookie 的作用
- 浏览器开发者工具修改 Cookie
- 权限验证机制

**解题思路**
1. 登录后查看 Cookie：`role=guest`
2. 使用开发者工具修改 Cookie：`role=admin`
3. 刷新页面，获取管理员权限
4. 页面显示 flag

**防御措施**
- 使用服务端 Session 管理权限
- Cookie 签名验证
- 加密敏感信息

---

## 2. 密码学入门题

### 题目 2.1：凯撒密码

**难度**：⭐☆☆☆☆

**题目描述**
- 密文：`synt{pnrfne_vf_rnfl}`
- 提示：ROT13

**知识点**
- 凯撒密码原理（字母移位）
- ROT13（移位 13）

**解题思路**
1. 识别为凯撒密码
2. 尝试所有移位（0-25），或直接用 ROT13
3. ROT13 解密：`flag{caesar_is_easy}`

**工具**
- CyberChef
- 在线 ROT13 解密器
- Python 脚本：
```python
import codecs
cipher = "synt{pnrfne_vf_rnfl}"
plain = codecs.decode(cipher, 'rot_13')
print(plain)  # flag{caesar_is_easy}
```

---

### 题目 2.2：Base64 编码

**难度**：⭐☆☆☆☆

**题目描述**
- 密文：`ZmxhZ3tiYXNlNjRfaXNfZW5jb2Rpbmd9`

**知识点**
- Base64 编码识别（字符集：A-Z, a-z, 0-9, +, /，末尾可能有 =）
- Base64 不是加密，只是编码

**解题思路**
1. 识别为 Base64（末尾有 = 或长度是 4 的倍数）
2. 使用工具解码
3. 得到：`flag{base64_is_encoding}`

**工具**
```bash
echo "ZmxhZ3tiYXNlNjRfaXNfZW5jb2Rpbmd9" | base64 -d
```

---

### 题目 2.3：多重编码

**难度**：⭐⭐☆☆☆

**题目描述**
- 密文：`NjY2YzYxNjc3YjY4NjU3ODVmNjQ2NTYzNmY2NDY1N2Q=`

**知识点**
- 识别多层编码
- Base64 + Hex

**解题思路**
1. 第一层：Base64 解码 → `666c61677b6865785f6465636f64657d`
2. 第二层：识别为十六进制（只有 0-9, a-f）
3. Hex 解码 → `flag{hex_decode}`

**工具**
- CyberChef（可以自动识别多层编码）
- Python：
```python
import base64
cipher = "NjY2YzYxNjc3YjY4NjU3ODVmNjQ2NTYzNmY2NDY1N2Q="
step1 = base64.b64decode(cipher).decode()
step2 = bytes.fromhex(step1).decode()
print(step2)  # flag{hex_decode}
```

---

## 3. 取证入门题

### 题目 3.1：图片隐写（EXIF）

**难度**：⭐☆☆☆☆

**题目描述**
- 给定一张图片 `image.jpg`
- 提示：图片里藏着秘密

**知识点**
- EXIF 元数据
- 图片文件结构

**解题思路**
1. 使用 `exiftool` 查看元数据
2. 在 Comment 或 Description 字段找到 flag
3. 或者使用 `strings` 命令查找可打印字符串

**工具**
```bash
exiftool image.jpg
strings image.jpg | grep flag
```

---

### 题目 3.2：文件头修复

**难度**：⭐⭐☆☆☆

**题目描述**
- 给定文件 `broken.png`，无法打开
- 提示：文件头损坏

**知识点**
- 常见文件头（Magic Number）
- PNG 文件头：`89 50 4E 47 0D 0A 1A 0A`
- 十六进制编辑器使用

**解题思路**
1. 使用 `file` 命令检查：显示 "data"
2. 使用十六进制编辑器打开
3. 发现文件头不是 PNG 标准头
4. 修改为正确的 PNG 文件头
5. 保存后打开图片，获取 flag

**工具**
```bash
hexdump -C broken.png | head
# 使用 hexedit 或 010 Editor 修改文件头
```

---

### 题目 3.3：压缩包密码破解

**难度**：⭐⭐☆☆☆

**题目描述**
- 给定加密的 ZIP 文件 `secret.zip`
- 提示：密码是 4 位数字

**知识点**
- ZIP 密码破解
- 字典攻击/暴力破解

**解题思路**
1. 尝试常见密码（1234、0000、password）
2. 使用工具暴力破解 4 位数字（0000-9999）
3. 找到密码后解压，获取 flag

**工具**
```bash
# 使用 fcrackzip
fcrackzip -b -c '1' -l 4-4 -u secret.zip

# 使用 John the Ripper
zip2john secret.zip > hash.txt
john hash.txt
```

---

## 4. 逆向工程入门题

### 题目 4.1：字符串查找

**难度**：⭐☆☆☆☆

**题目描述**
- 给定二进制文件 `crackme`
- 目标：找到 flag

**知识点**
- `strings` 命令
- 二进制文件中的可打印字符串

**解题思路**
1. 直接使用 `strings` 命令查看所有可打印字符串
2. 使用 `grep` 过滤包含 "flag" 的行
3. 找到 flag

**工具**
```bash
strings crackme | grep flag
strings crackme | grep -i ctf
```

---

### 题目 4.2：简单密码验证

**难度**：⭐⭐☆☆☆

**题目描述**
- 程序要求输入密码
- 输入正确密码后显示 flag

**知识点**
- 静态分析
- 字符串比较逻辑

**解题思路**
1. 使用 `strings` 查看字符串，可能直接看到密码
2. 或使用 `ltrace` 跟踪库函数调用
3. 观察 `strcmp` 函数的参数
4. 找到正确密码，输入后获取 flag

**工具**
```bash
ltrace ./crackme
# 输入测试密码，观察 strcmp 的参数
```

---

### 题目 4.3：简单异或加密

**难度**：⭐⭐☆☆☆

**题目描述**
- 程序对输入进行异或运算后与固定值比较

**知识点**
- 异或运算性质：`A ^ B = C` 则 `C ^ B = A`
- 简单加密算法逆向

**解题思路**
1. 使用反汇编工具查看逻辑
2. 找到异或的密钥
3. 对加密后的数据进行异或解密
4. 得到 flag

**Python 脚本示例**
```python
encrypted = [0x66, 0x6d, 0x60, 0x68, 0x7c, 0x78, 0x6f, 0x73]
key = 0x05
flag = ''.join(chr(c ^ key) for c in encrypted)
print(flag)  # flag{xor}
```

---

## 5. 杂项入门题

### 题目 5.1：二维码识别

**难度**：⭐☆☆☆☆

**题目描述**
- 给定一张包含二维码的图片

**知识点**
- 二维码扫描
- 图像处理

**解题思路**
1. 使用手机或在线工具扫描二维码
2. 获取 flag

**工具**
- 手机扫码 App
- 在线二维码解码器
- `zbarimg` 命令行工具

```bash
zbarimg qrcode.png
```

---

### 题目 5.2：摩斯密码

**难度**：⭐☆☆☆☆

**题目描述**
- 密文：`.... . .-.. .-.. --- / -.-. - ..-.`

**知识点**
- 摩斯密码（点和划）
- 空格分隔字母，斜杠分隔单词

**解题思路**
1. 识别为摩斯密码
2. 使用在线解码器或查表
3. 得到：`HELLO CTF`

**工具**
- CyberChef
- 在线摩斯密码解码器

---

### 题目 5.3：Python 脚本题

**难度**：⭐⭐☆☆☆

**题目描述**
- 服务器每秒发送一个数学题
- 需要在 1 秒内回答 100 道题

**知识点**
- Socket 编程
- 自动化脚本

**解题思路**
1. 使用 Python 连接服务器
2. 接收题目，解析数学表达式
3. 计算结果并发送
4. 重复 100 次后获取 flag

**Python 脚本示例**
```python
from pwn import *

conn = remote('server.com', 1234)

for i in range(100):
    question = conn.recvline().decode().strip()
    # 解析 "1 + 2 = ?"
    result = eval(question.replace('= ?', ''))
    conn.sendline(str(result).encode())

flag = conn.recvline().decode()
print(flag)
```

---

## 6. 学习建议

### 6.1 循序渐进
- 从最简单的题目开始
- 每个类别至少做 5-10 道题
- 理解原理比做题数量更重要

### 6.2 记录总结
- 记录每道题的解题思路
- 整理常用工具和命令
- 建立自己的知识库

### 6.3 阅读 Writeup
- 做不出来时看 Writeup 学习
- 理解别人的思路和技巧
- 尝试复现解题过程

### 6.4 动手实践
- 搭建本地环境练习
- 修改题目参数加深理解
- 编写自己的工具脚本

---

## 7. 推荐练习平台

- **picoCTF**：最适合新手，题目循序渐进
- **OverTheWire**：命令行和 Linux 基础
- **CryptoHack**：密码学专项练习
- **RingZer0 CTF**：各类题目，有难度标注
- **HackThisSite**：Web 安全入门

---

## 8. 参考资源

- [CTF Wiki](https://ctf-wiki.org/)
- [CTF 101](https://ctf101.org/)
- [picoCTF Resources](https://picoctf.org/resources)
- [Awesome CTF](https://github.com/apsdehal/awesome-ctf)
