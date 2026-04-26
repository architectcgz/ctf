# CTF 题目包

当前目录主要保留题目包产物与相关说明文档：

- `packs/<slug>/`：规范题目包目录，仓库内唯一真源
- `dist/<slug>.zip`：规范外层 zip 分发包

`challenge-pack-v1` 的最小必需文件仍是：

```text
<slug>/
├── challenge.yml
├── statement.md
├── attachments/    # 可选
├── docker/         # 可选
└── writeup/        # 可选
```

## 当前约定

- 修改题目内容时，只改 `packs/<slug>/`
- 对外交付或导入平台时，只使用 `dist/<slug>.zip`
- 题目目录内部如果出现 `challenge.zip`，那是题内附件，不是外层题目包
- 不再使用 `docs/challenges/*.zip` 或 `docs/challenges/packs/*.zip` 作为正式分发位置

## 维护方式

- 题目上传、录入与管理以平台为准，不再保留仓库内批量生成/校验脚本
- `packs/<slug>/` 继续作为已整理题目包的目录
- `dist/<slug>.zip` 继续作为对外交付或平台导入时使用的外层分发包
- 如需刷新某道题的外层 zip，可直接对 `packs/<slug>/` 重新打包生成同名 zip

## 保留文档

- `teacher-authoring-guide.md`：教师出题完整指南，包含离线题 / 容器题制作、验证和交付要求
- `templates/`：教师可直接复制的出题模板目录，包含离线题模板、容器 Web 题模板和容器 Pwn 题模板
- `challenge-list-real-sourced.md`：真实题源整理清单
- `challenge-list-launchable-real-sourced.md`：可启动真实题源清单
- `non-container-pack-audit.md`：非容器题目包的“离线可发布 / 仅题卡”分类报告
- `card-only-target-mode-audit.md`：170 个题卡建议补成“离线题 / 容器题”的分流报告

旧的进度文档、概览文档和 50 题构思清单已移除，避免继续和当前产物状态冲突。
