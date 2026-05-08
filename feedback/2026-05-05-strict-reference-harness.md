# Strict Reference Harness

## 问题描述

第一版初始化偏向适配现有 CTF 文档体系，使用了 docs 内适配层作为项目内索引。

## 原因分析

该做法符合“项目适配”，但不符合“严格参考 harness-engineering 仓库结构”的要求。

## 解决方案

改为创建参考仓库同构的顶层目录：`concepts/`、`thinking/`、`practice/`、`feedback/`、`works/`、`prompts/`、`references/`，并用 `scripts/check-consistency.sh` 检查这些目录和导航。

## 收获

当用户要求严格参考某个 harness 时，不能先把它折叠进现有 docs 体系；应优先保持参考项目的结构形态。
