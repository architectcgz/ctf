# Web-02 SQL Injection: Login Bypass

## 背景

这是一个典型的登录 SQL 注入题。应用使用 SQLite，并且登录时直接把用户名和密码拼接进 SQL 语句里。

服务启动时，会把容器运行环境变量 `CTF_FLAG` 写入数据库表 `secrets` 中（字段 `secret`），选手需要通过 SQL 注入把该值读出来并提交。

## 目标

1. 找到登录注入点并绕过认证
2. 读出数据库表 `secrets.secret` 的值（即 Flag）
3. 在平台提交该 Flag

## 访问入口

- `/` 首页
- `/login` 登录页
- `/me` 登录后查看当前用户信息（用于验证你已经绕过认证）

## 题目说明

- 默认存在用户：`guest / guest`
- Flag 不在源码内，来自 `CTF_FLAG` 环境变量（由平台在启动实例时注入）

