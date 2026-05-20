# Jeopardy 80 Pack Expansion Implementation Plan

## Objective

把当前 `challenges/jeopardy/packs/` 的 Jeopardy 题目总数从 `15` 道扩到约 `80` 道，最终收口为 `80` 道：

- 新增 `65` 道 Jeopardy 题目包
- 六个分类 `crypto / forensics / misc / pwn / reverse / web` 全覆盖
- 新增题全部补齐官方题解与可重放 solve 脚本
- 最终为整套 Jeopardy 包生成当前可引用的分类统计与验证报告

## Non-goals

- 不改平台题目导入、题包解析、registry、实例启动或 Flag 提交流程
- 不扩展新的题目分类，继续沿用现有六大类
- 不改 AWD 题目结构与 AWD 运行链路
- 不追求“80 道都做成容器题”；本轮以离线可复现题为主

## Inputs

- `AGENTS.md`
- `challenges/README.md`
- `challenges/teacher-authoring-guide.md`
- `challenges/jeopardy/templates/offline-static-template/*`
- `challenges/jeopardy/packs/*`
- `docs/plan/impl-plan/2026-05-08-batch-modern-challenge-packs-implementation-plan.md`
- `docs/reviews/general/2026-05-08-batch-modern-challenge-packs-review.md`

## Current Baseline

- 当前 `challenges/jeopardy/packs/` 实际只有 `15` 道题
- 现有分布：
  - `crypto`: 2
  - `forensics`: 2
  - `misc`: 2
  - `pwn`: 3
  - `reverse`: 2
  - `web`: 4
- 现有题包大多已有 `writeup/solution.md`，但没有统一的 `writeup/solve.py` 验证入口

## Chosen Direction

这次不做“附件里藏明文、写脚本抽一下就出 flag”的脚本题批量凑数，而是把新增 `65` 道题做成 **真实训练题**。`solve.py` 只保留为教师验题和验证报告的内部材料，不作为题目设计本体。

方向调整如下：

1. 目标不是把目录数堆到 `80`，而是把论文里的题量陈述收口到一个真实、可训练、可验证的规模。
2. `crypto / forensics / misc / reverse` 以离线样本题为主，但样本必须是真实训练材料：
   - `crypto`：真正需要做解密、密钥恢复或数学还原
   - `forensics`：优先 `sqlite / zip / eml / office / log / history / strings dump`
   - `misc`：优先文件结构、协议片段、隐藏信息、环境分析，不做纯谜语
   - `reverse`：给可执行文件、字节码或压缩脚本，不给明文源码
3. `web / pwn` 保持真实利用路径：
   - `web`：优先小型可运行题包或至少可本地复现实验链的源码包
   - `pwn`：优先本地 ELF 或最小可运行服务，做真正的溢出、格式化字符串、函数指针覆盖、整数问题
4. 题解和 `solve.py` 的职责是：
   - 证明题目真的能做
   - 给出教师可重复复现的验证证据
   - 不替代学生的实际训练路径
5. 新增题必须按“主训练点”去重：
   - 不允许只换题名、附件名或 flag 的重复题
   - 不允许同一分类里连续堆叠本质相同的解题动作
   - 每道题需要能标出唯一的主知识点、主解题动作和主训练目标

## Target Distribution

最终总数固定为 `80` 道，分类目标如下：

- `crypto`: 13
- `forensics`: 13
- `misc`: 13
- `pwn`: 13
- `reverse`: 13
- `web`: 15

对应新增配额：

- `crypto`: +11
- `forensics`: +11
- `misc`: +11
- `pwn`: +10
- `reverse`: +11
- `web`: +11

## Change Surface

- Create: `scripts/challenges/generate_jeopardy_expansion_batch.py`
- Create: `scripts/challenges/verify_jeopardy_packs.py`
- Create: `challenges/jeopardy/packs/<65 new slugs>/...`
- Create: `challenges/jeopardy/dist/<new-or-refreshed>.zip`
- Modify: existing Jeopardy packs that需要补齐 `writeup/solution.md` 或 `writeup/solve.py`
- Create: `docs/reports/2026-05-09-jeopardy-80-pack-verification.md`
- Create: `docs/reviews/general/2026-05-09-jeopardy-80-pack-expansion-review.md`

## Decision Notes

- 新增题不允许退化成“明文线索 + 提取脚本”的批量模板题。
- 新增题不允许做“换皮重复题”。如果两道题的主知识点、主解题动作和主训练目标一致，则保留一题、重做一题。
- `solve.py` 只作为内部验题入口，不能成为题目唯一价值所在。
- `pwn` 题使用本地 ELF 或可运行最小服务；solve 脚本需要真正打利用链。
- `reverse` 题使用可执行文件、字节码或压缩脚本；solve 脚本需要真正还原逻辑，而不是读取隐藏明文。
- `web` 题优先做可本地复现的最小应用；solve 脚本要走真实 HTTP 请求或真实利用路径。
- `forensics` 题使用日志、sqlite、zip、eml、office、history 等真实离线样本；solve 脚本直接从样本恢复答案。
- 新增题目每包至少包含：
  - `challenge.yml`
  - `statement.md`
  - `writeup/solution.md`
  - `writeup/solve.py`
  - 至少一个真实附件
- 最终验证口径以“solve 脚本实际跑通并完成真实解题链路”为准，而不是只做结构检查。

## Task Slices

### Slice 1: 固定 65 道新增题的题型梯度与生成入口

目标：

- 落一套可复用的批量生成脚本或构建工具
- 把 65 道题的 slug、标题、分类、难度、标签、提示、训练目标、附件模板和题解模板收进代码
- 明确哪些题是离线分析题，哪些题是本地可运行题
- 先产出题型去重矩阵，保证每道题对应唯一主训练点

Files / modules:

- `scripts/challenges/generate_jeopardy_expansion_batch.py`

Validation:

- 生成脚本 `--dry-run` 能输出稳定的分类配额和 slug 清单
- 不与现有 `15` 道 pack slug 冲突
- 新增 `65` 道题的主训练点矩阵中不存在重复项

Review focus:

- slug 命名是否稳定
- 分类配额是否精确达到目标
- 模板是否真正落到“学生需要分析 / 利用 / 调试”，而不是单纯提取
- 是否存在“题目名字不同，但核心训练点重复”的换皮题

### Slice 2: 批量生成新增 65 道真实训练题包与 dist zip

目标：

- 实际写入新增 pack 目录
- 生成外层分发 zip

Files / modules:

- `challenges/jeopardy/packs/<65 new slugs>/...`
- `challenges/jeopardy/dist/<65 new slugs>.zip`

Validation:

- pack 目录结构完整
- 附件非空
- zip 解压根目录仍为 `<slug>/`

Review focus:

- 题面、附件、flag、题解是否一致
- `pwn / reverse / web` 是否仍保留题型特征，而不是全部退化成同一种文本题
- `crypto / forensics / misc` 是否具备真实分析材料，而不是谜语壳

### Slice 3: 统一现有 15 道题的 solve 入口与缺失题解

目标：

- 为现有题补齐 `writeup/solve.py`
- 对缺失 `writeup/solution.md` 的 pack 补齐官方题解
- 如 pack 内容被补充，刷新对应 `dist/*.zip`

Files / modules:

- `challenges/jeopardy/packs/<existing slugs>/writeup/*`
- `challenges/jeopardy/dist/<existing slugs>.zip`

Validation:

- 现有 `15` 道题全部具备统一 solve 入口

Review focus:

- solve 是否真的依赖 pack 内材料
- 是否出现“题解走仓库源码捷径，但选手附件里拿不到同样信息”的问题

### Slice 4: 跑整套 80 道题的 solve 验证并生成当前报告

目标：

- 对最终 `80` 道 Jeopardy pack 跑逐题验证
- 产出论文可引用的当前统计与验证报告

Files / modules:

- `scripts/challenges/verify_jeopardy_packs.py`
- `docs/reports/2026-05-09-jeopardy-80-pack-verification.md`

Validation:

- 验证脚本逐题运行 `writeup/solve.py`
- solve 过程必须真的走解题链，而不是直接读取固定答案
- 最终分类统计与 pack 实际目录一致

Review focus:

- 报告里的总数和分类数是否与目录真实一致
- 是否只记录真实跑过的验证结果

### Slice 5: Review 收尾

目标：

- 用 review 心智检查题面真实性、题解独立性、验证覆盖率和文档口径

Files / modules:

- `docs/reviews/general/2026-05-09-jeopardy-80-pack-expansion-review.md`

Validation:

- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

Review focus:

- 新增 pack 是否存在明显重复模板却没有题型变化
- 现有题与新增题的“题包 + 题解 + solve 验证”口径是否统一
- 最终报告是否足以支撑“当前 Jeopardy 题目约 80 道”的论文表述

## Verification Plan

1. `python3 scripts/challenges/generate_jeopardy_expansion_batch.py --dry-run`
2. `python3 scripts/challenges/generate_jeopardy_expansion_batch.py --write`
3. `python3 scripts/challenges/verify_jeopardy_packs.py --all`
4. `bash scripts/check-consistency.sh`
5. `bash scripts/check-workflow-complete.sh`

## Rollback

- 删除本轮新增的 `challenges/jeopardy/packs/<slug>/`
- 删除本轮新增的 `challenges/jeopardy/dist/<slug>.zip`
- 回退对现有 `writeup/` 和 `dist/` 的补充
- 删除本轮新增的 `scripts/challenges/*.py`、验证报告和 review 文档
