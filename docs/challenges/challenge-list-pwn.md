# CTF 题目库 - 二进制利用类（50题）

> 来源：基于 pwnable.kr、pwnable.tw 等平台题目整理
> 格式：challenge-pack-v1

---

## 入门级 (beginner) - 15题

### 1. pwn-buffer-overflow-basic
- **标题**：栈溢出基础
- **描述**：覆盖返回地址跳转到 win 函数
- **标签**：`kp:stack-overflow`, `kp:ret2win`

### 2. pwn-shellcode-basic
- **标题**：Shellcode 注入
- **描述**：注入并执行 shellcode
- **标签**：`kp:shellcode`, `kp:code-injection`

### 3. pwn-format-string-leak
- **标题**：格式化字符串泄露
- **描述**：使用 %p 泄露栈上数据
- **标签**：`kp:format-string`, `kp:leak`

### 4. pwn-integer-overflow
- **标题**：整数溢出
- **描述**：整数溢出导致缓冲区溢出
- **标签**：`kp:integer-overflow`, `kp:arithmetic`

### 5. pwn-off-by-one
- **标题**：Off-by-One
- **描述**：单字节溢出覆盖关键数据
- **标签**：`kp:off-by-one`, `kp:overflow`

### 6. pwn-use-after-free-basic
- **标题**：UAF 基础
- **描述**：简单的 Use After Free
- **标签**：`kp:uaf`, `kp:heap`

### 7. pwn-double-free
- **标题**：Double Free
- **描述**：重复释放导致的漏洞
- **标签**：`kp:double-free`, `kp:heap`

### 8. pwn-null-byte-overflow
- **标题**：空字节溢出
- **描述**：strcpy 导致的空字节覆盖
- **标签**：`kp:null-byte`, `kp:overflow`

### 9. pwn-canary-leak
- **标题**：Canary 泄露
- **描述**：泄露并绕过 stack canary
- **标签**：`kp:canary`, `kp:leak`

### 10. pwn-ret2text
- **标题**：返回到代码段
- **描述**：返回到已有的 system 调用
- **标签**：`kp:ret2text`, `kp:rop`

### 11. pwn-ret2shellcode
- **标题**：返回到 Shellcode
- **描述**：跳转到注入的 shellcode
- **标签**：`kp:ret2shellcode`, `kp:nx-bypass`

### 12. pwn-got-overwrite
- **标题**：GOT 覆写
- **描述**：覆写 GOT 表劫持函数
- **标签**：`kp:got`, `kp:hijack`

### 13. pwn-stack-pivot
- **标题**：栈迁移
- **描述**：迁移栈到可控区域
- **标签**：`kp:stack-pivot`, `kp:rop`

### 14. pwn-one-gadget
- **标题**：One Gadget
- **描述**：使用 libc 中的 one gadget
- **标签**：`kp:one-gadget`, `kp:libc`

### 15. pwn-partial-overwrite
- **标题**：部分覆写
- **描述**：只覆写地址的低位字节
- **标签**：`kp:partial-overwrite`, `kp:aslr`

---

## 简单级 (easy) - 15题

### 16. pwn-ret2libc
- **标题**：返回到 libc
- **描述**：泄露 libc 地址调用 system
- **标签**：`kp:ret2libc`, `kp:aslr`

### 17. pwn-rop-chain
- **标题**：ROP 链构造
- **描述**：构造 ROP 链绕过 NX
- **标签**：`kp:rop`, `kp:gadget`

### 18. pwn-sigrop
- **标题**：SIGROP
- **描述**：Signal Return Oriented Programming
- **标签**：`kp:sigrop`, `kp:signal`

### 19. pwn-heap-overflow
- **标题**：堆溢出
- **描述**：堆溢出覆盖相邻 chunk
- **标签**：`kp:heap-overflow`, `kp:heap`

### 20. pwn-fastbin-attack
- **标题**：Fastbin Attack
- **描述**：利用 fastbin 机制
- **标签**：`kp:fastbin`, `kp:heap`

### 21. pwn-tcache-poisoning
- **标题**：Tcache Poisoning
- **描述**：污染 tcache 链表
- **标签**：`kp:tcache`, `kp:heap`

### 22. pwn-unlink-attack
- **标题**：Unlink 攻击
- **描述**：利用 unlink 机制写任意地址
- **标签**：`kp:unlink`, `kp:heap`

### 23. pwn-house-of-spirit
- **标题**：House of Spirit
- **描述**：伪造 chunk 结构
- **标签**：`kp:house-of-spirit`, `kp:heap`

### 24. pwn-house-of-force
- **标题**：House of Force
- **描述**：修改 top chunk size
- **标签**：`kp:house-of-force`, `kp:heap`

### 25. pwn-house-of-lore
- **标题**：House of Lore
- **描述**：污染 small bin
- **标签**：`kp:house-of-lore`, `kp:heap`

### 26. pwn-fsop
- **标题**：FSOP 攻击
- **描述**：File Stream Oriented Programming
- **标签**：`kp:fsop`, `kp:file-structure`

### 27. pwn-ret2csu
- **标题**：ret2csu
- **描述**：利用 __libc_csu_init gadget
- **标签**：`kp:ret2csu`, `kp:rop`

### 28. pwn-ret2dlresolve
- **标题**：ret2dlresolve
- **描述**：利用动态链接解析
- **标签**：`kp:ret2dlresolve`, `kp:dynamic-linking`

### 29. pwn-srop-chain
- **标题**：SROP 链
- **描述**：构造完整的 SROP 利用链
- **标签**：`kp:srop`, `kp:syscall`

### 30. pwn-mprotect-rop
- **标题**：mprotect ROP
- **描述**：使用 mprotect 修改内存权限
- **标签**：`kp:mprotect`, `kp:rop`

---

## 中等级 (medium) - 12题

### 31. pwn-pie-bypass
- **标题**：PIE 绕过
- **描述**：绕过 PIE 保护
- **标签**：`kp:pie`, `kp:aslr`

### 32. pwn-seccomp-bypass
- **标题**：Seccomp 绕过
- **描述**：绕过 seccomp 沙箱
- **标签**：`kp:seccomp`, `kp:sandbox`

### 33. pwn-house-of-orange
- **标题**：House of Orange
- **描述**：无 free 的堆利用
- **标签**：`kp:house-of-orange`, `kp:heap`

### 34. pwn-house-of-einherjar
- **标题**：House of Einherjar
- **描述**：off-by-one 堆利用
- **标签**：`kp:house-of-einherjar`, `kp:heap`

### 35. pwn-house-of-roman
- **标题**：House of Roman
- **描述**：fastbin attack 变种
- **标签**：`kp:house-of-roman`, `kp:heap`

### 36. pwn-largebin-attack
- **标题**：Large Bin Attack
- **描述**：利用 large bin 机制
- **标签**：`kp:largebin`, `kp:heap`

### 37. pwn-unsorted-bin-attack
- **标题**：Unsorted Bin Attack
- **描述**：利用 unsorted bin
- **标签**：`kp:unsorted-bin`, `kp:heap`

### 38. pwn-io-file-exploit
- **标题**：IO_FILE 利用
- **描述**：劫持 FILE 结构体
- **标签**：`kp:io-file`, `kp:vtable`

### 39. pwn-tls-dtor-hijack
- **标题**：TLS Destructor 劫持
- **描述**：劫持线程局部存储析构函数
- **标签**：`kp:tls`, `kp:destructor`

### 40. pwn-ld-preload
- **标题**：LD_PRELOAD 利用
- **描述**：利用环境变量劫持
- **标签**：`kp:ld-preload`, `kp:hijack`

### 41. pwn-kernel-rop
- **标题**：内核 ROP
- **描述**：内核空间的 ROP 利用
- **标签**：`kp:kernel`, `kp:rop`

### 42. pwn-race-condition-heap
- **标题**：堆竞争条件
- **描述**：多线程堆利用
- **标签**：`kp:race-condition`, `kp:heap`

---

## 困难级 (hard) - 6题

### 43. pwn-glibc-2.35-exploit
- **标题**：Glibc 2.35+ 利用
- **描述**：新版 glibc 的堆利用
- **标签**：`kp:glibc-2.35`, `kp:heap`

### 44. pwn-safe-linking-bypass
- **标题**：Safe-Linking 绕过
- **描述**：绕过 safe-linking 保护
- **标签**：`kp:safe-linking`, `kp:heap`

### 45. pwn-tcache-key-bypass
- **标题**：Tcache Key 绕过
- **描述**：绕过 tcache double free 检测
- **标签**：`kp:tcache-key`, `kp:heap`

### 46. pwn-kernel-uaf
- **标题**：内核 UAF
- **描述**：内核 Use After Free 利用
- **标签**：`kp:kernel`, `kp:uaf`

### 47. pwn-browser-exploit
- **标题**：浏览器漏洞利用
- **描述**：V8 引擎漏洞利用
- **标签**：`kp:browser`, `kp:v8`

### 48. pwn-vm-escape
- **标题**：虚拟机逃逸
- **描述**：QEMU 虚拟机逃逸
- **标签**：`kp:vm-escape`, `kp:qemu`

---

## 地狱级 (hell) - 2题

### 49. pwn-full-chain-exploit
- **标题**：完整利用链
- **描述**：从沙箱逃逸到内核提权
- **标签**：`kp:full-chain`, `kp:sandbox-escape`

### 50. pwn-real-world-0day
- **标题**：真实 0day 利用
- **描述**：真实软件的 0day 漏洞利用
- **标签**：`kp:0day`, `kp:real-world`

---

## 参考来源

- [pwnable.kr](https://pwnable.kr/)
- [pwnable.tw](https://pwnable.tw/)
