# 可启动真实题目清单

> 说明：以下题目基于真实 CTF 题源改编，当前版本已经补成可直接启动的容器题。

## web（2 题）

### web-rootme-sqli-authentication
- 标题：SQL 注入：认证绕过（Root-Me）
- 来源：Root-Me / SQL injection - Authentication
- 启动方式：容器，端口 `80/tcp`
- 访问链接：https://www.root-me.org/en/Challenges/Web-Server/SQL-injection-authentication

### web-rootme-php-command-injection
- 标题：命令注入：PHP 执行链（Root-Me）
- 来源：Root-Me / PHP - Command injection
- 启动方式：容器，端口 `80/tcp`
- 访问链接：https://www.root-me.org/en/Challenges/Web-Server/PHP-Command-injection

## crypto（2 题）

### crypto-rootme-encoding-ascii
- 标题：编码识别：ASCII 还原（Root-Me）
- 来源：Root-Me / Encoding - ASCII
- 启动方式：容器，端口 `80/tcp`
- 访问链接：https://www.root-me.org/en/Challenges/Cryptanalysis/Encoding-ASCII

### crypto-rootme-known-plaintext-xor
- 标题：异或分析：已知明文（Root-Me）
- 来源：Root-Me / Known plaintext - XOR
- 启动方式：容器，端口 `80/tcp`
- 访问链接：https://www.root-me.org/en/Challenges/Cryptanalysis/Known-plaintext-XOR

## reverse（2 题）

### reverse-rootme-elf-x86-basic
- 标题：ELF 基础逆向（Root-Me）
- 来源：Root-Me / ELF x86 - Basic
- 启动方式：容器，端口 `80/tcp`
- 访问链接：https://www.root-me.org/en/Challenges/Cracking/ELF-x86-Basic

### reverse-rootme-elf-x86-keygenme
- 标题：KeygenMe：序列号算法（Root-Me）
- 来源：Root-Me / ELF x86 - KeygenMe
- 启动方式：容器，端口 `80/tcp`
- 访问链接：https://www.root-me.org/en/Challenges/Cracking/ELF-x86-KeygenMe

## pwn（1 题）

### pwn-rootme-elf-x86-stack-overflow-basic-1
- 标题：栈溢出基础 1（Root-Me）
- 来源：Root-Me / ELF x86 - Stack buffer overflow basic 1
- 启动方式：容器，端口 `9999/tcp`
- 访问链接：https://www.root-me.org/en/Challenges/App-System/ELF-x86-Stack-buffer-overflow-basic-1

## forensics（2 题）

### forensics-rootme-find-me-on-android
- 标题：Android 取证定位（Root-Me）
- 来源：Root-Me / Find me on Android
- 启动方式：容器，端口 `80/tcp`
- 访问链接：https://www.root-me.org/en/Challenges/Forensic/Find-me-on-Android

### forensics-rootme-ios-introduction
- 标题：iOS 取证入门（Root-Me）
- 来源：Root-Me / iOS - Introduction
- 启动方式：容器，端口 `80/tcp`
- 访问链接：https://www.root-me.org/en/Challenges/Forensic/iOS-Introduction

## misc（1 题）

### misc-overthewire-bandit12
- 标题：Bandit 12：多层压缩拆包（OverTheWire）
- 来源：OverTheWire / Bandit Level 12
- 启动方式：容器，端口 `80/tcp`
- 访问链接：https://overthewire.org/wargames/bandit/bandit12.html

总计：10 题

目录位置：`ctf/challenges/packs/`
Zip 位置：`ctf/challenges/dist/`
