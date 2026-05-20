# Jeopardy 生成器拆分 Implementation Plan

## Objective

把当前 `scripts/challenges/generate_jeopardy_expansion_batch.py` 从单文件 `5422` 行拆成可维护模块，同时把 `scripts/challenges/verify_jeopardy_packs.py` 收口到同一套模型和路径定义下，要求：

- 保持原有 CLI 入口不变：
  - `python3 scripts/challenges/generate_jeopardy_expansion_batch.py ...`
  - `python3 scripts/challenges/verify_jeopardy_packs.py ...`
- 保持 `jeopardy_real_training_targets.json` 输入格式不变
- 保持 pack 生成结构、`challenge.yml` 渲染规则、验证逻辑和报告输出口径不变
- 后续继续补题、验题时，不再往单文件里堆 5000+ 行

## Non-goals

- 本切片不改 `65` 道新增题的题型设计、flag 设计和题解口径
- 本切片不改现有 `targets json` 字段结构
- 本切片不重做平台导入链路、题包格式或运行时定义
- 本切片不追求一次性把所有题重新生成到底，只做最小充分 smoke 验证

## Inputs

- `scripts/challenges/generate_jeopardy_expansion_batch.py`
- `scripts/challenges/verify_jeopardy_packs.py`
- `scripts/challenges/data/jeopardy_real_training_targets.json`
- `docs/plan/impl-plan/2026-05-09-jeopardy-80-pack-expansion-implementation-plan.md`
- `docs/design/jeopardy-80-真实训练题去重矩阵.md`

## Current Baseline

- 生成器当前是单文件 `5422` 行
- 验证器当前是单文件 `120` 行
- 生成器当前把以下职责堆在一个文件中：
  - CLI 参数解析
  - 目标加载、配额校验、去重矩阵同步
  - pack 写入与 dist zip 写入
  - `challenge.yml` 渲染
  - 通用构建工具（编译、压缩、pcap、sqlite、crypto helper 等）
  - 六大分类 `65` 个 builder
  - builder registry
- 六类 builder 具备明显天然边界：
  - `crypto` 11 个
  - `forensics` 11 个
  - `misc` 11 个
  - `reverse` 11 个
  - `pwn` 10 个
  - `web` 11 个
- 其中 `web` builder 区段单独就有约 `1147` 行，是当前最大的维护热点

## Chosen Direction

新增包目录 `scripts/challenges/jeopardy_batch/`，把“流程层、公共层、分类 builder”拆开，保留原脚本文件作为薄入口。

建议模块边界如下：

- `jeopardy_batch/models.py`
  - `Target`
  - `BuildResult`
  - `Builder`
- `jeopardy_batch/paths.py`
  - `REPO_ROOT`
  - `PACKS_DIR`
  - `DIST_DIR`
  - `TARGETS_FILE`
  - `MATRIX_DOC`
  - `REPORT_DOC`
  - 常量如 `IMAGE_TAG`、`WEB_PORT`
- `jeopardy_batch/targets.py`
  - `load_targets`
  - `validate_targets`
  - `filter_targets`
  - `print_summary`
  - `sync_matrix_doc`
- `jeopardy_batch/pack_io.py`
  - `render_challenge_yml`
  - `write_pack`
  - `write_dist_zip`
  - `write_text`
  - `yaml_quote`
- `jeopardy_batch/helpers.py`
  - `slug_flag`
  - `rng_for`
  - 编译、压缩、sqlite、pcap、crypto helper 等通用函数
- `jeopardy_batch/builders_crypto.py`
- `jeopardy_batch/builders_forensics.py`
- `jeopardy_batch/builders_misc.py`
- `jeopardy_batch/builders_reverse.py`
- `jeopardy_batch/builders_pwn.py`
- `jeopardy_batch/builders_web.py`
- `jeopardy_batch/registry.py`
  - 汇总 `BUILDERS`
- `jeopardy_batch/generate.py`
  - `build_target`
  - 生成 CLI `main`
- `jeopardy_batch/verify.py`
  - 验证 CLI `main`
  - `verify_target`
  - `write_report`

顶层入口保持不变：

- `scripts/challenges/generate_jeopardy_expansion_batch.py` 只保留薄包装，转发到 `jeopardy_batch.generate.main`
- `scripts/challenges/verify_jeopardy_packs.py` 只保留薄包装，转发到 `jeopardy_batch.verify.main`

## Ownership Boundary

- `targets.py` 只负责“目标集合语义”
- `pack_io.py` 只负责“题包输出语义”
- `helpers.py` 只负责“builder 可复用底层工具”
- `builders_*` 只负责“单分类题包内容构造”
- `registry.py` 是唯一 builder 装配点
- `generate.py` / `verify.py` 是唯一 CLI 流程 owner

## Task Slices

### Slice 1: 建立公共包骨架并迁移共享模型/路径/目标加载

目标：

- 建立 `jeopardy_batch` 包结构
- 抽出 `Target` / `BuildResult` / 常量路径 / targets 相关函数
- 让生成器和验证器共享同一套数据模型

Validation:

- `python3 -m py_compile` 覆盖新包与两个入口脚本
- `--sync-docs` 能正常运行

Review focus:

- 导入关系是否简单直接
- 路径层是否没有循环依赖
- 验证器是否不再复制 `Target` 定义

### Slice 2: 迁移 pack 写入与通用 helper

目标：

- 抽出 `pack_io.py` 与 `helpers.py`
- 保持 `challenge.yml` 渲染、pack 原子写入、zip 原子写入逻辑不变

Validation:

- 选取 1 个 `crypto` slug 做 `--write` smoke
- 检查生成目录结构和 zip 根目录是否正确

Review focus:

- 原子写入语义是否保留
- `runtime/container` 渲染逻辑是否保持兼容

### Slice 3: 按分类迁移 65 个 builder 与 registry

目标：

- 将六类 builder 拆到独立模块
- `registry.py` 统一装配

Validation:

- `python3 scripts/challenges/generate_jeopardy_expansion_batch.py --slug <sample>` 对六个分类各抽 1 题 smoke
- `build_target()` 对全量 target 进行无落盘构建 smoke

Review focus:

- builder 名称与 `kind` 映射是否完全一致
- 分类模块是否只依赖公共 helper，而不是再次交叉耦合

### Slice 4: 收口生成/验证 CLI 并跑最小充分验证

目标：

- 顶层旧脚本改为薄入口
- 验证器切到共享模型与路径模块
- 保证“生成器拆分后仍能继续完成后续 80 题工作流”

Validation:

- `python3 -m py_compile scripts/challenges/generate_jeopardy_expansion_batch.py scripts/challenges/verify_jeopardy_packs.py`
- `python3 scripts/challenges/generate_jeopardy_expansion_batch.py --sync-docs`
- `python3 scripts/challenges/generate_jeopardy_expansion_batch.py --slug crypto-repeating-key-xor`
- `python3 scripts/challenges/verify_jeopardy_packs.py --slug crypto-repeating-key-xor`

Review focus:

- CLI 参数和输出口径是否保持兼容
- 失败信息是否仍然足够定位

### Slice 5: 独立 review 与工作流收尾

目标：

- 对拆分结果做独立 review
- 记录残余风险与后续继续生成 65 题时的注意点

Validation:

- 独立 review 归档到 `docs/reviews/general/`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

Review focus:

- 是否只是把大文件机械切开，还是把 owner 真正收口清楚
- 是否仍有明显的单点继续失控位置

## Risks

- 拆分过程中最容易引入的是相对导入、路径层和 registry 映射错误
- `web` / `pwn` builder 依赖本地工具与多行嵌入脚本，迁移时最容易发生缩进或字符串损坏
- 若直接全量重跑 `--write`，会受已有 pack / zip 存在状态影响，因此本切片验证以“编译 + 抽样生成 + 抽样验证 + 全量无落盘构建”为主

## Done Criteria

- 生成器不再是单文件 5000+ 行
- 顶层入口兼容旧命令
- 生成与验证共享模型/路径定义
- 六类 builder 已按分类落到独立模块
- 至少完成一轮可复现的编译与抽样 smoke 验证
- 已归档独立 review 和最终工作流检查结果
