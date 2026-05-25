# 前端 review 单轮快照在主索引吸收后应及时清理

## 问题描述

前端专项审查曾长期保留一批 `docs/reviews/frontend/ctf-frontend-code-review-*.md` 单轮快照。主索引已经吸收这些快照里的当前结论后，活动目录里仍继续保留原文件，导致后续 agent 容易把旧文件里的“未修复”状态误读成当前待办。

## 原因分析

- 第七十四轮只完成了“当前事实源 + 历史读取规则”的索引治理，没有继续收口原始快照文件。
- 单轮快照保留在活动目录后，文件名看起来仍像可直接引用的当前 review 证据。
- 旧快照里的行号、测试数量和修复状态天然会过期，继续保留比继续索引更容易制造漂移。

## 解决方案

- 把仍有价值的当前结论收敛到 `docs/reviews/frontend/ctf-frontend-audit-20260422.md` 与 `docs/reviews/frontend/README.md`。
- 在索引里明确：主索引是当前事实源，早期单轮快照只在 Git 历史中回溯，不再作为活动目录文件保留。
- 清理已被吸收的 `ctf-frontend-code-review-*` 单轮快照，避免 review 证据面继续膨胀。

## 收获

- review 目录里“当前事实源”和“历史原文”边界更清楚，后续判断待修问题时不会再默认翻到过期快照。
- 旧审查结论的保留方式从“活动目录堆文件”改成“主索引摘要 + Git 历史回溯”，维护成本更低。
- 后续新增前端 review 时，可以先判断是否真的需要独立证据文件，而不是默认继续累积 round 文档。

## 沉淀状态

- 状态：已沉淀
- Owner：`docs/reviews/frontend/README.md`、`docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 链接：`feedback/2026-05-23-frontend-review-snapshots-can-be-pruned-after-audit-absorption.md`
