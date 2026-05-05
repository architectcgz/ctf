# Challenge Package Import Design

## 背景

当前挑战管理主流程仍以手工表单创建题目为主，管理员在管理端直接填写标题、分类、难度、分值、描述、镜像、附件地址、提示和 Flag 配置。与此同时，后端已经存在一个离线的 `import-challenge-packs` 脚本，但它没有接入管理端主流程，也不符合当前统一的 `challenge.yml` 题目包思路。

结合 [docs/毕业设计课题.md](/home/azhi/workspace/projects/ctf/.worktrees/feat-challenge-package-import/docs/毕业设计课题.md) 中“使用 Docker 技术快速生成靶机”“靶场管理支持自定义拓扑与难度分级”的要求，本项目的题目管理不能继续停留在“元数据录入表单”，而应收敛到“可复用、可迁移、可部署”的题目包模型。

## 目标

将挑战管理主流程从“手工创建挑战”切换为“导入题目包”，并以 `challenge.yml` 作为唯一主规范：

- 管理端不再以手工录题表单作为主入口
- 管理员通过上传 `.zip` 题目包完成导入
- 后端统一解析 `challenge.yml`
- 首版覆盖：
  - 题目元数据
  - 题目说明
  - 附件
  - hints
  - flag
  - runtime 镜像/容器信息
- `topology` 在包规范中预留扩展位，但首版不强制自动导入为拓扑编排数据
- `writeup` 暂不纳入首版题目包导入

## 为什么做到这一层

### 不选择“只导入元数据和附件”

该方案太接近现有手工建题流程，只是换了录入入口，难以体现课题要求中的 Docker 靶机生成与靶场运行时配置能力，也不符合 CTFd / rCTF 常见的工程化题目包思路。

### 不首版直接把 topology 和 writeup 全部并入

`topology` 和 `writeup` 在当前系统中已经是独立复杂能力：

- 拓扑：独立页面、独立数据模型、独立编排逻辑
- 题解：独立编辑和发布能力

首版直接把它们都并进导入规范，会显著扩大改造面，增加跨模块耦合、测试成本与失败场景，不利于毕业设计周期内保证“完整、稳定、可演示、可答辩”的主线结果。

## 新的题目包规范

建议首版使用如下目录约定：

```text
challenge-package.zip
  challenge.yml
  statement.md
  attachments/
  dist/
  src/
  docker/
    Dockerfile
    docker-compose.yml
```

建议 `challenge.yml` 首版字段：

```yaml
api_version: v1
kind: challenge

meta:
  slug: web-sqli-101
  title: SQL Injection 101
  category: web
  difficulty: easy
  points: 100
  tags:
    - sqli
    - mysql

content:
  statement: statement.md
  attachments:
    - path: attachments/web-sqli-101.zip
      name: web-sqli-101.zip

flag:
  type: static
  value: flag{example}
  prefix: flag

hints:
  - level: 1
    title: Hint 1
    cost_points: 0
    content: 先关注查询参数拼接方式

runtime:
  type: container
  image:
    ref: ctf/web-sqli-101:latest
  expose:
    - port: 80
      protocol: http

extensions:
  topology:
    source: docker/topology.yml
    enabled: false
```

约束说明：

- `meta / content / flag / hints / runtime` 是首版强约束
- `extensions.topology` 是预留字段，不要求首版真正落库
- 题目附件以包内相对路径声明，由平台接管存储和下载分发
- 运行时首版只要求容器镜像引用，不要求在线构建 Dockerfile

## 管理端交互设计

挑战管理页改为三段式：

1. 导入入口
   - 主 CTA 从“创建挑战”改为“导入题目包”
   - 支持 `.zip` 上传
   - 支持拖拽 / 点击选择文件
2. 导入结果
   - 展示题目包解析结果
   - 展示元数据、附件、flag、runtime、warnings
   - 支持确认导入
3. 已有题目列表
   - 保留“查看 / 编排 / 题解 / 发布 / 删除”
   - 取消手工创建与手工编辑元数据对话框作为主路径

设计取向：

- 题目包导入是主推荐路径
- 已导入题目后续仍可以在详情、编排、题解页面继续完善
- 页面视觉继续沿用当前管理端风格，不另起新路由

## 后端设计

### 模块边界

- `challenge` 模块 owner：
  - 题目包解析
  - 导入草稿
  - challenge / hint / flag / attachment 元数据落库
- `runtime` 模块继续 owner：
  - 运行时实例能力
  - 镜像探测与运行时探针
- `topology` 首版不跨模块自动写入，只作为 challenge 导入结果里的扩展信息保留

### 新流程

1. 管理员上传 `.zip`
2. 后端解压到临时目录
3. 解析 `challenge.yml`
4. 读取 `statement.md`
5. 校验附件相对路径、flag、runtime image、hint 结构
6. 生成导入预览结果
7. 管理员确认后执行导入
8. 创建或更新 challenge、hint、flag、附件引用、image 关联

### API 设计

- `POST /api/v1/admin/challenge-imports`
  - multipart 上传题目包
  - 返回导入预览结果和临时导入 ID
- `GET /api/v1/admin/challenge-imports/:id`
  - 查询导入预览详情
- `POST /api/v1/admin/challenge-imports/:id/commit`
  - 确认导入
  - 创建或更新 challenge 草稿

### 导入预览持久化策略

首版不新增 `challenge_imports` 数据表，而是采用“文件工作目录 + 预览 JSON”的临时持久化方案：

- 上传后将 zip 保存到服务端导入工作目录
- 解压后的题目包和导入预览 JSON 以 `import_id` 为键放在同一临时目录
- `GET /admin/challenge-imports/:id` 与 `POST /admin/challenge-imports/:id/commit` 直接读取该目录
- 确认导入后删除临时预览目录

这样做的原因：

- 预览是短生命周期过程数据，不是稳定业务实体
- 首版无需为一次性草稿预览引入额外表结构、迁移和清理任务
- 仍然保留了 `preview -> review -> commit` 的双阶段导入体验
- 更适合当前毕业设计阶段的“闭环完整、结构清晰、实现成本可控”目标

首版不支持：

- 在服务端直接构建 Dockerfile
- 自动发布
- 自动导入 writeup
- 自动把 topology 写入现有编排表

## 与现有实现的衔接

- 现有 `code/backend/cmd/import-challenge-packs/main.go` 不直接删除
- 将其中“读取包、解析 manifest、构建 challenge spec”的逻辑抽到 challenge 模块内共享解析器
- CLI 导入脚本改为复用 `challenge.yml` 解析器
- 管理端在线上传导入与 CLI 离线导入使用同一套解析核心

## 风险

### 高风险

- 题目包导入约定统一到 `challenge.yml`
- 附件存储与下载路径兼容
- 导入预览与真正提交之间的临时文件生命周期管理

### 中风险

- 旧管理端测试基本围绕手工表单，需要整体改写
- 现有 challenge DTO 主要面向手工建题，需要补导入 DTO 与返回结构

### 低风险

- 前端视觉改造
- 导入记录列表

## 验收标准

- 管理端主入口已从“手工创建挑战”切换为“导入题目包”
- 上传合法题目包后可以看到导入预览
- 确认导入后可创建 challenge 草稿
- 导入结果至少正确落地：
  - title
  - description
  - category
  - difficulty
  - points
  - attachment
  - hints
  - flag
  - image/runtime 引用
- 已导入 challenge 可继续进入详情、编排、题解和发布流程
- CLI 导入能力与管理端导入共用 `challenge.yml` 解析逻辑

## 追加记录（2026-05-03）

### 已完成：重复 slug 导入改为拒绝，而不是覆盖更新

- 普通题在线导入与 CLI 导入：
  - 若 `meta.slug` 已被已有题目占用，commit / import 会直接拒绝，不再按 slug 覆盖更新。
  - 仍保留一层兼容：历史上尚未绑定 `package_slug` 的旧题，可在首次导入同标题同分类题包时被接管并补齐 `package_slug`。
- AWD 在线导入：
  - 若 `slug` 已存在，commit 会直接返回冲突，不再覆盖已有 AWD 题目。
- 前端导入页：
  - commit 失败时保留并透传后端冲突文案，管理员可直接看到 “slug 已被占用，请改用题目编辑入口更新”。
