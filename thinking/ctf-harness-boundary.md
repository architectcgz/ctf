# CTF Harness Boundary

## 论点

本项目严格采用参考 harness 的顶层目录形态，但业务事实源仍保留在 CTF 原有代码和文档中。

## 项目证据

重点关注 API 合同、后端 UTC/context 契约、前端路由命名空间、AWD 运行时、超大页面职责堆叠与验证闭环。

## 判断

Harness 层负责让 agent 找到事实源和反馈，不把所有 CTF 架构内容复制进 harness 目录。
