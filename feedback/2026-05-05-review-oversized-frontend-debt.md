# Review touching oversized frontend files must record residual debt

## Migrated From

历史 improvement 条目 `2026-05-05-review-touching-oversized-frontend-files-must-record-residual-debt`

## Status

superseded-by-agent-recorded-rule

## 问题描述

AWD defense content page review 触达了已在前端 audit TD-1 oversized-component backlog 中列出的 `ContestAWDWorkspacePanel.vue`，但初始 review archive 只写了 `No material findings`，没有显式记录已知拆分债。

## 原因分析

当 review 命中已知超大 owner 文件，却只检查当前功能 diff 的正确性，后续读 review 结论的人会误以为该文件已经适合长期承载新需求。

## 解决方案

review 归档触达已知超大组件时必须写清：

- 当前切片是否降低 owner 混杂风险。
- 原有拆分债是否仍然存在，以及是否阻塞本次合并。
- 若不阻塞，本次为何只做最小切片而不继续展开结构重排。

## 收获

该条已被更强规则 `Touched structural debt must be closed in pipeline` 覆盖，但仍保留为具体事故样本。
