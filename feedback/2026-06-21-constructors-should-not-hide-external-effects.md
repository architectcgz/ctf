# 构造函数不应隐藏外部依赖读取和副作用

## 问题描述

后端实现中容易把“顺手补一个默认值”写进 `NewService` / `New...` 构造函数，例如读取 hostname、环境变量、文件、证书、网络状态、时间或随机数。这样会让一个看起来只是组装依赖的函数隐含外部世界访问，调用方和测试都不容易意识到这里可能失败、阻塞或受运行环境影响。

本次 runtime-agent identity 中的 `os.Hostname()` 只是一个例子；同类问题还包括在构造函数里读取 env、打开文件、探测网络、创建真实 client、读取当前时间生成默认 identity、静默生成随机 secret 等。

## 原因分析

- 构造函数的默认职责是接收依赖和配置，形成可测试对象；外部依赖读取应由进程入口、bootstrap、composition、factory 或明确 owner 层负责。
- 隐藏外部访问会模糊错误 owner：失败是配置问题、运行环境问题，还是对象构造问题，调用方看不出来。
- 单元测试会被机器环境污染，必须额外 patch OS/env/time/random 才能验证本该纯粹的对象行为。
- 用 `_` 忽略这类错误会丢掉运维信号，后续只能看到空字段或默认值，无法判断是未配置、读取失败还是 fallback 生效。

## 解决方案

- 普通 `New...` / `NewService` 构造函数只接收已经解析好的依赖、配置和 identity，不主动访问外部世界。
- 需要外部事实时，放到名字和职责明确的 owner 层：`bootstrap`、`composition`、`Load*`、`Dial*`、`Open*`、`Build*`、进程入口或显式 factory。
- 外部读取失败必须可观察：返回错误、打结构化日志，或由调用方明确决定 fallback；不要在构造函数里静默吞掉。
- 测试通过依赖注入传入 identity、clock、random、hostname、client 或文件内容，不把本机环境变成行为断言的一部分。

## 收获

- 看到构造函数里出现 `os.*`、`net.*`、文件读取、环境变量、时间、随机数、真实 client dial/open 时，先停下来问：这是不是外部依赖 owner 错位？
- 如果函数名已经表达副作用，例如 `LoadServerTLSConfig`、`DialContext`、`OpenStore`，可以访问外部世界，但错误必须向上返回或可观察。
- “为了方便默认值”不是把外部世界访问塞进构造函数的理由；默认值应来自配置归一化、bootstrap 或显式 factory。

## 沉淀状态

- 状态：archived
- Owner：`.agents/skills/ctf-backend-patterns/SKILL.md`
- 链接：`.agents/skills/ctf-backend-patterns/SKILL.md` 的 `Construction & External Effects` 检查点
