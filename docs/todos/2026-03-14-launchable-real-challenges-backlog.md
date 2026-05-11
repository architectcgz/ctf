# 可启动真实题目补充待办

更新日期：2026-03-14

## 背景

当前已经完成一批“基于真实题源改编、可直接启动容器实例”的题目包，用于给平台先补一部分可实际开靶的内容。

本待办只记录后续未完成的扩充工作，避免明天继续时重复梳理上下文。

## 已完成（2026-03-14）

1. [x] 补齐 10 个可启动真实题目包
   - 覆盖方向：
     - Web 2 题
     - Crypto 2 题
     - Reverse 2 题
     - Pwn 1 题
     - Forensics 2 题
     - Misc 1 题
   - 题包目录：
     - `docs/challenges/packs/`
   - 汇总清单：
     - `docs/challenges/challenge-list-launchable-real-sourced.md`

2. [x] 为上述题目补齐容器运行内容
   - 已补内容：
     - `challenge.yml`
     - `statement.md`
     - `docker/Dockerfile`
     - Web / Crypto / Reverse / Forensics / Misc 的附件分发站点
     - Pwn 的 TCP 服务容器

3. [x] 完成基础可用性验证
   - 已通过：
     - `python3 docs/challenges/test_manifests.py`
     - `python3 docs/challenges/test_all_structure.py`
     - `python3 docs/challenges/test_zips.py`
     - 新增 10 题全部 `docker build` 成功
   - 已抽检运行：
     - Web HTTP 页面可访问
     - Crypto 附件分发页可访问
     - Pwn TCP 服务可连接

4. [x] 补齐批量生成脚本
   - 脚本位置：
     - `docs/challenges/generate_launchable_real_source_packs.py`

## 明日继续待办

### P1：继续扩大“可直接启动”的真实题目覆盖面

1. [ ] 把剩余真实题源题卡继续扩成可启动容器题
   - 优先顺序建议：
     1. Web
     2. Pwn
     3. Reverse
     4. Crypto
     5. Forensics
     6. Misc

2. [ ] 优先补齐 Web 方向的真实题源容器题
   - 建议优先转化：
     - `web-rootme-file-upload-double-ext`
     - `web-rootme-http-open-redirect`
     - `web-rootme-http-improper-redirect`
   - 目标：
     - 每题都具备真实漏洞入口
     - 页面交互、flag 获取路径和原题核心解法保持一致

3. [ ] 继续补齐 Pwn 方向容器题
   - 当前仅有 1 题可直接启动
   - 下一批建议优先做：
     - 更标准的 ret2win / shellcode / canary 绕过基础题
     - 保留真实题源思路，但样本与 flag 本地化

### P1：提升题目“可直接导入平台”的完整度

1. [ ] 为新增容器题补更明确的附件与解题入口说明
   - 目标：
     - 学员打开题面后，能明确知道是访问 HTTP、下载样本还是连 TCP

2. [ ] 为容器题补更稳定的动态 flag 或 flag 注入策略
   - 当前情况：
     - 已有部分题使用 `FLAG` 环境变量
   - 后续目标：
     - 统一不同题型的 flag 注入口径

3. [ ] 评估是否需要为附件型题目增加专门的下载页模板
   - 当前是按题单独生成静态页
   - 可考虑后续统一模板，减少重复内容

### P2：质量与维护性收尾

1. [ ] 把新增可启动题目纳入一个固定批量验证脚本
   - 目标：
     - 一次跑完 manifest / structure / zip / docker build / 基础 runcheck

2. [ ] 清理和复查 `docs/challenges/packs/packs/` 这类历史嵌套目录
   - 说明：
     - 当前题库里存在一批历史生成结果落在嵌套目录
     - 后续需要判断是否保留、迁移或统一整理

3. [ ] 统一真实题源记录格式
   - 当前已记录：
     - 平台
     - 原题名
     - 原始链接
   - 后续可补：
     - 改编说明
     - 本地化差异
     - 是否已具备完整靶机

## 明日继续入口

1. 先看：
   - `docs/challenges/challenge-list-launchable-real-sourced.md`
   - `docs/challenges/generate_launchable_real_source_packs.py`

2. 再从以下目录继续扩题：
   - `docs/challenges/packs/`

3. 继续后优先复跑：
   - `python3 docs/challenges/test_manifests.py`
   - `python3 docs/challenges/test_all_structure.py`
   - `python3 docs/challenges/test_zips.py`
   - 新增题目的 `docker build`

## 备注

1. `docs/challenges` 当前是仓库忽略目录，题库文件不会进入 git 正常跟踪
2. 后续继续做题时，默认只改 `docs/challenges` 与 `docs/todos`，不要顺手触碰当前其他业务代码
