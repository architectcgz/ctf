# 2026-05-21 AWD 复盘 PDF 导出修复实施计划

## 目标

- 修复教师 AWD 复盘 PDF 中文字体不可用的问题
- 让教师 AWD 复盘 PDF 在未显式选轮次时也能输出可复盘的实质内容
- 提高导出报告的信息密度，避免结果只剩赛事基础信息

## 非目标

- 不改教师 AWD 复盘详情页查询行为
- 不新增新的 report 任务类型
- 不重构个人 / 班级报告导出链路

## 输入事实源

- `docs/architecture/features/AWD教师复盘归档与报告导出设计.md`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/commands/report_pdf_fonts.go`
- `code/backend/internal/module/assessment/application/commands/awd_review_export_builder.go`
- `code/backend/internal/module/assessment/application/commands/awd_review_export_renderer.go`

## 任务切片

1. 修正 PDF 字体装配
   - 使用现有 report PDF 字体注册点
   - 明确 bold/regular 字体来源，避免回退到核心英文字体
   - 验证：相关 Go 单测

2. 收口 AWD 报告默认轮次选择
   - 导出 builder 在未指定 `round_number` 时自动挑选焦点轮次
   - 保持详情页查询契约不变，只在导出装配层补默认策略
   - 验证：builder / renderer 相关 Go 单测

3. 增强 AWD PDF 报告内容结构
   - 继续复用通用 report 骨架
   - 补充关键样本和分析段落，让教师可直接看到关键攻击、流量和建议
   - 验证：PDF 渲染相关 Go 单测

4. 同步架构文档
   - 只更新当前设计事实：默认选轮次、报告主要输出块
   - 验证：必要时运行一致性检查

## 风险与回退

- 风险：字体嵌入方式不正确会导致 PDF 仍回退到英文核心字体
- 风险：默认选轮次策略如果放错层，会影响详情页查询契约
- 回退：改动集中在 assessment report 导出实现和文档，可单独回退

## 预期验证

- `go test ./internal/module/assessment/application/commands -run 'Test(RenderAWDReviewReportPDF|WritePersonalPDF|TeacherAWDReviewExportBuilder)' -count=1`
- 若文档有改动：`bash scripts/check-consistency.sh`
