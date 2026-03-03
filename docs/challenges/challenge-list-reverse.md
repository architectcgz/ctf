# CTF 题目库 - 逆向工程类（50题）

> 来源：基于 picoCTF、Crackmes 等平台题目整理
> 格式：challenge-pack-v1

---

## 入门级 (beginner) - 15题

### 1. reverse-strings-basic
- **标题**：字符串查找
- **描述**：使用 strings 命令找到 flag
- **标签**：`kp:strings`, `kp:static-analysis`

### 2. reverse-ltrace-basic
- **标题**：库函数跟踪
- **描述**：使用 ltrace 查看 strcmp 参数
- **标签**：`kp:ltrace`, `kp:dynamic-analysis`

### 3. reverse-strace-basic
- **标题**：系统调用跟踪
- **描述**：使用 strace 查看文件操作
- **标签**：`kp:strace`, `kp:syscall`

### 4. reverse-hex-editor
- **标题**：十六进制编辑
- **描述**：修改二进制文件中的检查逻辑
- **标签**：`kp:hex-edit`, `kp:patching`

### 5. reverse-simple-xor
- **标题**：简单 XOR 加密
- **描述**：逆向 XOR 加密算法
- **标签**：`kp:xor`, `kp:algorithm`

### 6. reverse-base64-check
- **标题**：Base64 验证
- **描述**：程序对输入进行 Base64 编码后比较
- **标签**：`kp:base64`, `kp:encoding`

### 7. reverse-password-check
- **标题**：密码验证
- **描述**：简单的字符串比较
- **标签**：`kp:strcmp`, `kp:password`

### 8. reverse-flag-format
- **标题**：Flag 格式检查
- **描述**：检查输入是否符合 flag{} 格式
- **标签**：`kp:format-check`, `kp:regex`

### 9. reverse-ascii-art
- **标题**：ASCII 艺术
- **描述**：flag 隐藏在 ASCII 艺术中
- **标签**：`kp:ascii`, `kp:visual`

### 10. reverse-elf-header
- **标题**：ELF 文件头
- **描述**：flag 隐藏在 ELF 文件头中
- **标签**：`kp:elf`, `kp:file-format`

### 11. reverse-upx-packed
- **标题**：UPX 加壳
- **描述**：使用 UPX 解壳
- **标签**：`kp:upx`, `kp:packer`

### 12. reverse-simple-loop
- **标题**：简单循环
- **描述**：逆向简单的 for 循环加密
- **标签**：`kp:loop`, `kp:algorithm`

### 13. reverse-array-check
- **标题**：数组检查
- **描述**：输入与硬编码数组逐字节比较
- **标签**：`kp:array`, `kp:comparison`

### 14. reverse-function-call
- **标题**：函数调用
- **描述**：跟踪函数调用找到验证逻辑
- **标签**：`kp:function`, `kp:call-stack`

### 15. reverse-gdb-basic
- **标题**：GDB 调试基础
- **描述**：使用 GDB 设置断点查看变量
- **标签**：`kp:gdb`, `kp:debugging`

---

## 简单级 (easy) - 15题

### 16. reverse-ida-basic
- **标题**：IDA 反汇编
- **描述**：使用 IDA 分析控制流
- **标签**：`kp:ida`, `kp:disassembly`

### 17. reverse-ghidra-decompile
- **标题**：Ghidra 反编译
- **描述**：使用 Ghidra 查看伪代码
- **标签**：`kp:ghidra`, `kp:decompile`

### 18. reverse-anti-debug
- **标题**：反调试检测
- **描述**：绕过 ptrace 反调试
- **标签**：`kp:anti-debug`, `kp:ptrace`

### 19. reverse-obfuscation
- **标题**：代码混淆
- **描述**：去混淆还原原始逻辑
- **标签**：`kp:obfuscation`, `kp:deobfuscate`

### 20. reverse-vm-basic
- **标题**：虚拟机保护
- **描述**：简单的字节码虚拟机
- **标签**：`kp:vm`, `kp:bytecode`

### 21. reverse-z3-solver
- **标题**：符号执行
- **描述**：使用 Z3 求解约束
- **标签**：`kp:z3`, `kp:smt`

### 22. reverse-angr-basic
- **标题**：Angr 自动化
- **描述**：使用 Angr 自动求解路径
- **标签**：`kp:angr`, `kp:symbolic-execution`

### 23. reverse-golang-binary
- **标题**：Go 语言逆向
- **描述**：逆向 Go 编译的二进制
- **标签**：`kp:golang`, `kp:language-specific`

### 24. reverse-rust-binary
- **标题**：Rust 语言逆向
- **描述**：逆向 Rust 编译的二进制
- **标签**：`kp:rust`, `kp:language-specific`

### 25. reverse-dotnet-decompile
- **标题**：.NET 反编译
- **描述**：使用 dnSpy 反编译 C# 程序
- **标签**：`kp:dotnet`, `kp:csharp`

### 26. reverse-java-decompile
- **标题**：Java 反编译
- **描述**：使用 JD-GUI 反编译 JAR
- **标签**：`kp:java`, `kp:jar`

### 27. reverse-python-pyc
- **标题**：Python 字节码
- **描述**：反编译 .pyc 文件
- **标签**：`kp:python`, `kp:bytecode`

### 28. reverse-apk-basic
- **标题**：Android APK 逆向
- **描述**：使用 jadx 反编译 APK
- **标签**：`kp:android`, `kp:apk`

### 29. reverse-ios-binary
- **标题**：iOS 二进制逆向
- **描述**：逆向 Mach-O 格式文件
- **标签**：`kp:ios`, `kp:macho`

### 30. reverse-shellcode-analysis
- **标题**：Shellcode 分析
- **描述**：分析并理解 shellcode 功能
- **标签**：`kp:shellcode`, `kp:assembly`

---

## 中等级 (medium) - 12题

### 31. reverse-control-flow-flatten
- **标题**：控制流平坦化
- **描述**：还原被平坦化的控制流
- **标签**：`kp:control-flow`, `kp:obfuscation`

### 32. reverse-opaque-predicate
- **标题**：不透明谓词
- **描述**：识别并去除不透明谓词
- **标签**：`kp:opaque-predicate`, `kp:obfuscation`

### 33. reverse-custom-packer
- **标题**：自定义加壳
- **描述**：分析自定义加壳算法
- **标签**：`kp:packer`, `kp:unpacking`

### 34. reverse-tls-callback
- **标题**：TLS 回调
- **描述**：分析 TLS 回调中的反调试
- **标签**：`kp:tls`, `kp:anti-debug`

### 35. reverse-exception-handler
- **标题**：异常处理
- **描述**：利用异常处理隐藏逻辑
- **标签**：`kp:seh`, `kp:exception`

### 36. reverse-inline-hook
- **标题**：Inline Hook
- **描述**：检测并绕过 inline hook
- **标签**：`kp:hook`, `kp:anti-analysis`

### 37. reverse-import-obfuscation
- **标题**：导入表混淆
- **描述**：动态解析 API 地址
- **标签**：`kp:iat`, `kp:obfuscation`

### 38. reverse-string-encryption
- **标题**：字符串加密
- **描述**：运行时解密字符串
- **标签**：`kp:string-encryption`, `kp:obfuscation`

### 39. reverse-code-virtualization
- **标题**：代码虚拟化
- **描述**：分析 VMProtect 保护
- **标签**：`kp:virtualization`, `kp:vmprotect`

### 40. reverse-llvm-obfuscation
- **标题**：LLVM 混淆
- **描述**：去除 LLVM-Obfuscator 混淆
- **标签**：`kp:llvm`, `kp:obfuscation`

### 41. reverse-firmware-analysis
- **标题**：固件分析
- **描述**：提取并分析嵌入式固件
- **标签**：`kp:firmware`, `kp:embedded`

### 42. reverse-protocol-analysis
- **标题**：协议逆向
- **描述**：逆向自定义网络协议
- **标签**：`kp:protocol`, `kp:network`

---

## 困难级 (hard) - 6题

### 43. reverse-kernel-driver
- **标题**：内核驱动逆向
- **描述**：分析 Windows 内核驱动
- **标签**：`kp:kernel`, `kp:driver`

### 44. reverse-hypervisor
- **标题**：虚拟化保护
- **描述**：分析基于虚拟化的保护
- **标签**：`kp:hypervisor`, `kp:vt-x`

### 45. reverse-sgx-enclave
- **标题**：SGX Enclave
- **描述**：分析 Intel SGX 保护的代码
- **标签**：`kp:sgx`, `kp:tee`

### 46. reverse-malware-analysis
- **标题**：恶意软件分析
- **描述**：完整分析真实恶意软件样本
- **标签**：`kp:malware`, `kp:analysis`

### 47. reverse-bootloader
- **标题**：Bootloader 逆向
- **描述**：分析 UEFI Bootloader
- **标签**：`kp:bootloader`, `kp:uefi`

### 48. reverse-hardware-crypto
- **标题**：硬件加密
- **描述**：绕过硬件加密芯片
- **标签**：`kp:hardware`, `kp:crypto`

---

## 地狱级 (hell) - 2题

### 49. reverse-full-obfuscation
- **标题**：完全混淆
- **描述**：多层混淆+虚拟化+反调试
- **标签**：`kp:full-protection`, `kp:advanced`

### 50. reverse-zero-day-analysis
- **标题**：0day 漏洞分析
- **描述**：在真实软件中发现未知漏洞
- **标签**：`kp:0day`, `kp:vulnerability`

---

## 参考来源

- [Crackmes.one](https://crackmes.one/)
- [picoCTF Reverse Engineering](https://picoctf.org/)
