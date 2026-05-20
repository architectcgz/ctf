# Harness Engineering 初始化实施计划

## Objective

把 `ctf` 仓库改造成可被 AI agent 稳定读取、执行和验证的项目级 harness，并沉淀一个可复用的 `harness-engineering` skill。

## Non-goals

- 不重排现有业务代码、架构目录或文档事实源。
- 不修改当前未提交的前端业务改动。

## Source Inputs

- 参考仓库：`https://github.com/deusyu/harness-engineering`
- 当前项目入口：`AGENTS.md`
- 当前文档入口：`docs/README.md`、`docs/architecture/README.md`
- 当前反馈目录：`feedback/`
- 当前 review / plan 目录：`docs/reviews/`、`docs/plan/impl-plan/`
- 当前 hook：`.githooks/pre-commit`、`scripts/install-githooks.sh`

## Chosen Direction

严格参考 `deusyu/harness-engineering` 的学习档案目录结构，而不是把 harness 折叠到现有 `docs/` 体系。

参考仓库的可复用部分是：

- 仓库即记录系统
- `AGENTS.md` 渐进式导航
- 反馈记录
- 机械化一致性检查

在 `ctf` 中，这些能力分别落到顶层目录：

- `concepts/`：概念笔记
- `thinking/`：独立思考
- `practice/`：动手实践
- `feedback/`：反馈记录
- `works/`：作品输出
- `prompts/`：提示词积累
- `references/`：外部资源索引
- `scripts/check-consistency.sh` 与 `.githooks/pre-commit`：机械化检查

## Task Slices

1. 创建 `harness-engineering` skill。
   - 变更位置：`/home/azhi/.codex/skills/harness-engineering`
   - 验证：`quick_validate.py`、`verify_skill_references.py`

2. 用 skill 初始化 `ctf` harness。
   - 变更位置：`concepts/`、`thinking/`、`practice/`、`feedback/`、`works/`、`prompts/`、`references/`、`scripts/check-consistency.sh`
   - 验证：`bash scripts/check-consistency.sh`

3. 接入入口导航和 hook。
   - 变更位置：`AGENTS.md`、`docs/README.md`、`.githooks/pre-commit`、`.githooks/README.md`
   - 验证：harness 检查脚本确认链接和 hook 接入

4. 做 review 记录。
   - 变更位置：`docs/reviews/general/2026-05-05-harness-engineering-initialization-review.md`
   - 验证：记录检查命令、结论和残余风险

5. 迁移既有 agent 积累到严格 harness 顶层目录。
   - 变更位置：`feedback/`、`practice/`、`prompts/`、`references/`、`works/`
   - 输入来源：历史 improvements、skills、superpowers、planning 与 refs 积累
   - 验证：`scripts/check-consistency.sh` 的 C6/C7 检查迁移索引存在且可从目录 `AGENTS.md` 发现

## Compatibility Impact

- 不改变运行时代码和 API。
- 新增 pre-commit 检查会在启用 `.githooks` 后执行；检查只读仓库文件，不生成产物。
- 保留原有 API 合同同步 hook。

## Validation

```bash
python3 /home/azhi/.codex/skills/.system/skill-creator/scripts/quick_validate.py /home/azhi/.codex/skills/harness-engineering
python3 /home/azhi/.codex/skills/.system/skill-creator/scripts/verify_skill_references.py /home/azhi/.codex/skills/harness-engineering
python3 /home/azhi/.codex/skills/harness-engineering/scripts/init_harness_project.py --repo ctf --project-name ctf --profile ctf-platform --mode strict-reference
bash scripts/check-consistency.sh
bash .githooks/pre-commit
```

## Review Focus

- 是否严格参考参考仓库的顶层 harness 目录，而不是继续使用旧的 docs 内适配层。
- 是否避免覆盖用户已有规则和未提交业务改动。
- 是否有最小可执行检查，能防止 harness 导航和反馈目录漂移。

## Rollback

本次变更不涉及数据库、运行时代码或外部服务。回退时移除新增顶层 harness 目录、`scripts/check-consistency.sh`，并删除标记块 `BEGIN HARNESS ENGINEERING` 到 `END HARNESS ENGINEERING`。
