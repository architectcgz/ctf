# Jeopardy 题目包

这里保留普通 Jeopardy 题的仓库内可维护源、模板和分发 zip。平台导入后的题包源、导出包和镜像构建 source 由后端 `challenge` 模块的存储适配器持久化，不以本目录作为运行时存储位置。

题包正式契约见：

- [challenge-pack-v1.md](../../docs/contracts/challenge-pack-v1.md)

## 目录结构

```text
jeopardy/
├── packs/
│   └── <slug>/
│       ├── challenge.yml
│       ├── statement.md
│       ├── attachments/    # 可选
│       ├── docker/         # 可选
│       └── writeup/        # 可选
├── dist/
│   └── <slug>.zip
└── templates/
```

## 当前约定

- `packs/<slug>/` 是仓库内可维护源；修改题面、附件、Dockerfile、writeup 时只改这里。
- `dist/<slug>.zip` 是外层分发包；导入平台或对外交付时使用这里。
- 题目目录内部的 `challenge.zip` 只表示题内附件，不是外层题包。
- 模板只从 `templates/` 复制起步，不作为正式题目导入。

## 平台导入与存储边界

普通题包导入入口：

```text
POST /api/v1/authoring/challenge-imports
POST /api/v1/authoring/challenge-imports/:id/commit
```

导入时的服务端存储由这些 owner 承接：

- `ChallengeImportPreviewStore`：保存上传 zip、解包预览 workspace 和 preview JSON。
- `ChallengePackageStorage`：提交导入后持久化题包 source、导出 archive 和 image build source。
- `ImageBuildService`：对 `platform_build` 题包创建并推进 `images` / `image_build_jobs` 状态机。

仓库内 `packs/` 和 `dist/` 不参与平台运行时读取；平台运行态只依赖导入后落到数据库和 storage adapter 管理目录里的事实。

## 维护方式

复制模板：

```bash
cd challenges/jeopardy
cp -R templates/offline-static-template packs/<your-slug>
```

刷新外层分发包：

```bash
cd challenges/jeopardy
mkdir -p dist
cd packs
zip -r ../dist/<your-slug>.zip <your-slug>
```

容器题最低自测：

```bash
cd challenges/jeopardy
docker build -t test-<your-slug> packs/<your-slug>/docker
docker run --rm -p 18080:80 test-<your-slug>
```

更完整的模板说明见 [templates/README.md](templates/README.md)，教师出题流程见 [../teacher-authoring-guide.md](../teacher-authoring-guide.md)。
