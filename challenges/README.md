# CTF 题目包入口

本目录保留仓库内可维护的题包源、分发 zip、出题模板和题目说明。平台导入后的题包源、导出包、镜像构建 source 和私有 checker artifact 不以本目录为运行时事实源，而由后端 `challenge` 模块的存储适配器收口。

当前入口：

- [jeopardy/README.md](jeopardy/README.md)：Jeopardy / 普通题题包源、模板和分发包入口。
- [awd/README.md](awd/README.md)：AWD 题目包、期次目录、本地验题和 checker 约定入口。
- [teacher-authoring-guide.md](teacher-authoring-guide.md)：教师出题完整指南，包含离线题 / 容器题制作、验证和交付要求。

仓库内目录和平台存储边界：

- 仓库内 Jeopardy 题包源：`jeopardy/packs/<slug>/`
- 仓库内 Jeopardy 外层分发包：`jeopardy/dist/<slug>.zip`
- 仓库内 AWD 题包源：`awd/<period>/<slug>/`
- 仓库内 AWD 外层分发包：`awd/dist/<slug>.zip`
- 平台导入预览：`ChallengeImportPreviewStore` / `AWDChallengeImportPreviewStore`
- 平台题包持久化：`ChallengePackageStorage`
- 平台 AWD 私有 checker artifact：`AWDCheckerArtifactStore`

对应代码事实源：

- `code/backend/internal/module/challenge/ports/ports.go` 的 `ChallengePackageStorage` port
- `code/backend/internal/module/challenge/infrastructure/challenge_package_storage.go`
- `code/backend/internal/module/challenge/infrastructure/challenge_import_preview_store.go`
- `code/backend/internal/module/challenge/infrastructure/awd_checker_artifact_store.go`
- `docs/architecture/backend/modules/challenge.md`
- `docs/architecture/features/题包Registry交付架构.md`

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

- 修改题目内容时，只改对应题型的源码目录：`jeopardy/packs/<slug>/` 或 `awd/<period>/<slug>/`
- 对外交付或导入平台时，只使用对应 `dist/<slug>.zip`
- 题目目录内部如果出现 `challenge.zip`，那是题内附件，不是外层题目包
- 不再使用 `docs/challenges/*.zip` 或 `docs/challenges/packs/*.zip` 作为正式分发位置

## 维护方式

- 题目上传、录入、导入预览、提交导入、导出修订与镜像构建以平台为准。
- 仓库内 `packs` / `ctf-*` 目录只作为可维护源和演示资产。
- `dist/<slug>.zip` 继续作为对外交付或平台导入时使用的外层分发包。
- 如需刷新某道题的外层 zip，按对应题型 README 从源码目录重新打包生成同名 zip。

## 保留文档

- `teacher-authoring-guide.md`：教师出题完整指南，包含离线题 / 容器题制作、验证和交付要求
- `jeopardy/templates/`：教师可直接复制的 Jeopardy 出题模板目录，包含离线题模板、容器 Web 题模板和容器 Pwn 题模板
- `challenge-list-real-sourced.md`：真实题源整理清单
- `challenge-list-launchable-real-sourced.md`：可启动真实题源清单
- `non-container-pack-audit.md`：非容器题目包的“离线可发布 / 仅题卡”分类报告
- `card-only-target-mode-audit.md`：170 个题卡建议补成“离线题 / 容器题”的分流报告

旧的进度文档、概览文档和 50 题构思清单已移除，避免继续和当前产物状态冲突。
