# 后端测试要按 owner 分层，PDF 断言不能绑定脆弱文案

## 问题描述

这轮后端测试架构拆分后，`internal/app` 的 full-router 集成测试里仍保留了一段旧式 PDF 断言：

- 直接断历史英文标题，例如 `Teacher AWD Review Report`
- 使用落后的 PDF helper，只能匹配原始字节和 zlib 解压后的 ASCII 文本

结果是模块内 `assessment/application/commands` 的 PDF 渲染测试一直是绿的，但 full-router 既有用例持续失败。表面看像“报告文案改了”，实际是两层问题混在一起：

- full-router 测试断言越过了它自己的职责边界，开始绑定 renderer 文案细节
- full-router helper 没有跟模块级 `pdfContainsText` 对齐，缺少 UTF-16 BE 文本匹配

## 原因分析

- 模块内测试和 full-router 测试的 owner 没有严格区分，导致集成层去断本该由模块层负责的渲染细节。
- 同一类 PDF 文本提取 helper 在不同测试层各自演化，没有把“可提取文本的编码能力”当成共享契约维护。
- 过去用英文标题做断言时能通过，后来 renderer 改成中文后，这类脆弱断言就暴露出来了。

## 解决方案

- 模块内测试按 owner 断言：
  - `application/queries` 测 archive / selected round / filter / snapshot 语义
  - `application/commands` 测导出生命周期、builder 选择逻辑、zip/json 字段保真、PDF 渲染内容
- full-router / `internal/app` 只兜集成层事实：
  - 路由可达
  - 权限正确
  - 导出状态流正确
  - 下载元信息正确
  - 二进制内容至少包含当前稳定可提取的结构标记
- PDF 断言优先断稳定 token，而不是脆弱文案：
  - 优先 payload token、section heading、字段名、路径、编号等
  - 避免把标题文案、措辞、语言版本直接当成集成层唯一断言
- 同类 helper 要对齐能力：
  - 模块层 `pdfContainsText` 如果已经支持 UTF-16 BE，full-router 对应 helper 也必须同步
  - 不能一边修模块层，一边让集成层继续用落后的解析能力

## 收获

- 测试分层不是“目录分开”就结束，断言内容也要跟着 owner 分层。
- full-router 失败而模块内测试通过时，先排查 helper 漂移和测试职责越界，不要先怀疑生产逻辑。
- 对 PDF、ZIP、导出文件这类二进制产物，集成层应验证“下载链路 + 稳定结构信号”，细节内容由模块层负责更精确的回归保护。

## 交叉链接

- `.harness/reuse-decisions/backend-test-architecture-phase15.md`
- `.harness/reuse-decisions/assessment-awd-review-report-pdf-fix.md`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- `code/backend/internal/module/assessment/application/queries/teacher_awd_review_service_test.go`
- `code/backend/internal/app/full_router_teacher_state_matrix_test.go`

## 沉淀状态

- 状态：仅项目保留
- Owner：`feedback/` 项目 harness
- 链接：`feedback/2026-06-03-backend-test-layering-and-pdf-assertion-ownership.md`
