# 真实题源清单

> 格式：challenge-pack-v1
> 说明：以下题包均基于公开 CTF 平台真实题目改编整理，仅保留中文题卡和来源信息，便于后续补全附件或容器。

## Web（5 题）

### 1. web-rootme-sqli-authentication
- **标题**：SQL 注入：认证绕过（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：SQL injection - Authentication
- **难度**：`easy`
- **链接**：https://www.root-me.org/en/Challenges/Web-Server/SQL-injection-authentication
- **概述**：给定一个存在认证逻辑缺陷的登录入口，目标是在不掌握有效凭据的情况下进入受保护区域。

### 2. web-rootme-php-command-injection
- **标题**：命令注入：PHP 执行链（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：PHP - Command injection
- **难度**：`medium`
- **链接**：https://www.root-me.org/en/Challenges/Web-Server/PHP-Command-injection
- **概述**：给定一个会把用户输入传入系统命令的 PHP 页面，目标是借助命令拼接拿到敏感输出。

### 3. web-rootme-http-open-redirect
- **标题**：开放重定向：跳转污染（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：HTTP - Open redirect
- **难度**：`easy`
- **链接**：https://www.root-me.org/en/Challenges/Web-Server/HTTP-Open-redirect
- **概述**：应用会根据用户可控参数执行跳转，目标是构造恶意地址并利用跳转逻辑拿到目标信息。

### 4. web-rootme-file-upload-double-ext
- **标题**：文件上传：双扩展绕过（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：File upload - Double extensions
- **难度**：`medium`
- **链接**：https://www.root-me.org/en/Challenges/Web-Server/File-upload-Double-extensions
- **概述**：站点限制脚本文件上传，但校验逻辑不完整，目标是借助双扩展名或解析差异绕过限制。

### 5. web-rootme-http-improper-redirect
- **标题**：不安全跳转：状态流篡改（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：HTTP - Improper redirect
- **难度**：`easy`
- **链接**：https://www.root-me.org/en/Challenges/Web-Server/HTTP-Improper-redirect
- **概述**：应用通过重定向串联关键业务步骤，目标是分析跳转流程并直接访问被错误保护的页面。

## Crypto（5 题）

### 1. crypto-rootme-encoding-ascii
- **标题**：编码识别：ASCII 还原（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：Encoding - ASCII
- **难度**：`beginner`
- **链接**：https://www.root-me.org/en/Challenges/Cryptanalysis/Encoding-ASCII
- **概述**：给定一段经过 ASCII 数值化处理的文本，目标是还原原文并识别其中的关键信息。

### 2. crypto-rootme-hash-md5
- **标题**：摘要识别：MD5 基础（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：Hash - Message Digest 5
- **难度**：`beginner`
- **链接**：https://www.root-me.org/en/Challenges/Cryptanalysis/Hash-Message-Digest-5
- **概述**：给出一段常见散列值，目标是识别算法并用最合适的方法恢复原始明文。

### 3. crypto-rootme-known-plaintext-xor
- **标题**：异或分析：已知明文（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：Known plaintext - XOR
- **难度**：`easy`
- **链接**：https://www.root-me.org/en/Challenges/Cryptanalysis/Known-plaintext-XOR
- **概述**：密文由 XOR 方案生成，并给出部分已知明文线索，目标是恢复密钥或完整内容。

### 4. crypto-rootme-vigenere
- **标题**：多表代换：维吉尼亚密码（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：Polyalphabetic substitution - Vigenere
- **难度**：`medium`
- **链接**：https://www.root-me.org/en/Challenges/Cryptanalysis/Polyalphabetic-substitution-Vigenere
- **概述**：目标是识别多表代换方案并通过密钥长度与频率特征还原原文。

### 5. crypto-rootme-shift-cipher
- **标题**：移位密码：凯撒变体（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：Shift cipher
- **难度**：`beginner`
- **链接**：https://www.root-me.org/en/Challenges/Cryptanalysis/Shift-cipher
- **概述**：给出一段经过固定偏移处理的密文，目标是恢复可读明文并识别隐藏信息。

## Reverse（5 题）

### 1. reverse-rootme-elf-x86-basic
- **标题**：ELF 基础逆向（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：ELF x86 - Basic
- **难度**：`beginner`
- **链接**：https://www.root-me.org/en/Challenges/Cracking/ELF-x86-Basic
- **概述**：给出一个 Linux ELF 可执行文件，目标是通过静态或动态分析恢复校验逻辑。

### 2. reverse-rootme-pe-x86-0-protection
- **标题**：PE 基础逆向（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：PE x86 - 0 protection
- **难度**：`easy`
- **链接**：https://www.root-me.org/en/Challenges/Cracking/PE-x86-0-protection
- **概述**：给定一个无额外保护的 Windows PE 程序，目标是还原其输入校验或关键分支条件。

### 3. reverse-rootme-elf-x86-keygenme
- **标题**：KeygenMe：序列号算法（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：ELF x86 - KeygenMe
- **难度**：`medium`
- **链接**：https://www.root-me.org/en/Challenges/Cracking/ELF-x86-KeygenMe
- **概述**：程序会对输入用户名和序列号进行组合校验，目标是逆向出生成规则。

### 4. reverse-rootme-apk-introduction
- **标题**：Android APK 入门逆向（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：APK - Introduction
- **难度**：`easy`
- **链接**：https://www.root-me.org/en/Challenges/Cracking/APK-Introduction
- **概述**：目标是对 Android APK 进行反编译和代码阅读，定位隐藏在客户端中的校验信息。

### 5. reverse-rootme-unity-mono-basic-game-hacking
- **标题**：Unity Mono 游戏逆向（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：Unity - Mono - Basic Game Hacking
- **难度**：`medium`
- **链接**：https://www.root-me.org/en/Challenges/Cracking/Unity-Mono-Basic-Game-Hacking
- **概述**：给出一款基于 Unity Mono 的小游戏，目标是定位内置逻辑并修改或理解核心校验流程。

## Pwn（5 题）

### 1. pwn-rootme-elf-x64-basic-heap-overflow
- **标题**：堆溢出基础（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：ELF x64 - Basic heap overflow
- **难度**：`hard`
- **链接**：https://www.root-me.org/en/Challenges/App-System/ELF-x64-Basic-heap-overflow
- **概述**：给出一个存在堆区越界写的 64 位程序，目标是控制程序执行路径或修改关键对象。

### 2. pwn-rootme-riscv-intro-rop
- **标题**：RISC-V ROP 入门（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：ELF RISC-V - Intro let's do the ROP
- **难度**：`hard`
- **链接**：https://www.root-me.org/en/Challenges/App-System/ELF-RISC-V-Intro-let-s-do-the-ROP
- **概述**：给定一份 RISC-V ELF 程序，目标是在非 x86 架构下完成一次基础 ROP 链构造。

### 3. pwn-rootme-arm64-multithreading
- **标题**：多线程竞争利用（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：ELF ARM64 - Multithreading
- **难度**：`hard`
- **链接**：https://www.root-me.org/en/Challenges/App-System/ELF-ARM64-Multithreading
- **概述**：目标是在 ARM64 程序中分析并利用由多线程访问时序引入的缺陷。

### 4. pwn-rootme-arm64-heap-underflow
- **标题**：堆下溢：边界反向破坏（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：ELF ARM64 - Heap Underflow
- **难度**：`hard`
- **链接**：https://www.root-me.org/en/Challenges/App-System/ELF-ARM64-Heap-Underflow
- **概述**：程序存在向前越界写的问题，目标是利用下溢破坏相邻堆元数据或关键对象。

### 5. pwn-rootme-elf-x86-stack-overflow-basic-1
- **标题**：栈溢出基础 1（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：ELF x86 - Stack buffer overflow basic 1
- **难度**：`medium`
- **链接**：https://www.root-me.org/en/Challenges/App-System/ELF-x86-Stack-buffer-overflow-basic-1
- **概述**：给定一个基础 32 位程序，目标是利用栈缓冲区溢出覆盖返回地址，进入目标函数或输出敏感信息。

## Forensics（5 题）

### 1. forensics-rootme-deleted-file
- **标题**：已删除文件恢复（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：Deleted file
- **难度**：`easy`
- **链接**：https://www.root-me.org/en/Challenges/Forensic/Deleted-file
- **概述**：提供一个含删除痕迹的介质样本，目标是恢复关键文件并提取其中的线索。

### 2. forensics-rootme-ios-introduction
- **标题**：iOS 取证入门（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：iOS - Introduction
- **难度**：`medium`
- **链接**：https://www.root-me.org/en/Challenges/Forensic/iOS-Introduction
- **概述**：目标是在 iOS 备份或应用痕迹中定位账户、消息或其他关键证据。

### 3. forensics-rootme-find-me-on-android
- **标题**：Android 取证定位（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：Find me on Android
- **难度**：`medium`
- **链接**：https://www.root-me.org/en/Challenges/Forensic/Find-me-on-Android
- **概述**：在 Android 设备数据中寻找指定线索，目标是通过应用缓存、数据库或系统记录完成定位。

### 4. forensics-rootme-lost-case-mobile
- **标题**：失踪案件：移动调查（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：The Lost Case - Investigation Mobile
- **难度**：`hard`
- **链接**：https://www.root-me.org/en/Challenges/Forensic/The-Lost-Case-Investigation-Mobile
- **概述**：题目以事件调查为背景，需要从多源移动端证据中还原时间线并锁定关键事实。

### 5. forensics-rootme-malicious-word-macro
- **标题**：恶意 Word 宏分析（Root-Me）
- **来源平台**：Root-Me
- **原题名称**：Malicious Word macro
- **难度**：`medium`
- **链接**：https://www.root-me.org/en/Challenges/Forensic/Malicious-Word-macro
- **概述**：给定一份带宏的 Office 文档，目标是提取宏代码并识别其中的恶意行为或隐藏数据。

## Misc（5 题）

### 1. misc-overthewire-bandit0
- **标题**：Bandit 0：SSH 连接入门（OverTheWire）
- **来源平台**：OverTheWire
- **原题名称**：Bandit Level 0
- **难度**：`beginner`
- **链接**：https://overthewire.org/wargames/bandit/bandit0.html
- **概述**：目标是使用给定账号密码通过 SSH 连入靶机，为后续关卡打基础。

### 2. misc-overthewire-bandit1
- **标题**：Bandit 1：特殊文件名读取（OverTheWire）
- **来源平台**：OverTheWire
- **原题名称**：Bandit Level 1
- **难度**：`beginner`
- **链接**：https://overthewire.org/wargames/bandit/bandit1.html
- **概述**：目标是读取一个文件名对 shell 来说具有特殊含义的文件，从而取得下一步线索。

### 3. misc-overthewire-bandit5
- **标题**：Bandit 5：批量检索目标文件（OverTheWire）
- **来源平台**：OverTheWire
- **原题名称**：Bandit Level 5
- **难度**：`easy`
- **链接**：https://overthewire.org/wargames/bandit/bandit5.html
- **概述**：需要在大量目录和文件中筛出满足特定大小、类型与权限条件的唯一目标文件。

### 4. misc-overthewire-bandit9
- **标题**：Bandit 9：字符串筛选（OverTheWire）
- **来源平台**：OverTheWire
- **原题名称**：Bandit Level 9
- **难度**：`easy`
- **链接**：https://overthewire.org/wargames/bandit/bandit9.html
- **概述**：题目把线索藏在一个二进制样本的可打印字符串中，目标是结合筛选条件找到正确内容。

### 5. misc-overthewire-bandit12
- **标题**：Bandit 12：多层压缩拆包（OverTheWire）
- **来源平台**：OverTheWire
- **原题名称**：Bandit Level 12
- **难度**：`medium`
- **链接**：https://overthewire.org/wargames/bandit/bandit12.html
- **概述**：给出一份经过十六进制转储和多层压缩处理的文件，目标是逐层还原并提取最终线索。

---

总计：30 题

目录位置：`ctf/challenges/packs/`
Zip 位置：`ctf/challenges/dist/`
