# 带顶部 tab 的面板不应重复渲染 eyebrow

## 问题描述

带顶部 tab 的页面已经通过 tab 文案提供了当前内容分区语义，但部分 tab 面板内部又在列表 header 或内容区顶部放了 `workspace-overline`、`student-directory-list-heading__eyebrow` 或局部 kicker。

这会造成视觉层级重复：用户先看到 tab，再看到一个含义接近的 eyebrow，页面显得啰嗦，也容易让后续实现继续复制类似结构。

## 原因分析

之前统一列表标题区时，只对普通列表页面建立了 header 结构约束，没有区分“页面主体列表”和“tab 面板内列表”。

结果是 `/scoreboard` 这类带顶部 tab 的页面，在 tab 面板下继续复用带 eyebrow 的列表 header，形成了重复标识。

## 解决方案

- 带顶部 tab 的页面，tab 面板内的列表或 section header 默认不再渲染 eyebrow。
- 保留标题本身，例如 `竞赛排行列表`、`积分排行列表`。
- 保留真正的数据字段 label、指标卡 label、表单 label；这些不是分区 eyebrow。
- 新增/修改 top-tabs 页面时，先检查 tab 文案是否已经表达当前分区，避免在 tab 内容第一层继续放 `workspace-overline`、`journal-note-label`、`student-directory-list-heading__eyebrow` 或等价局部 kicker。
- 如果父级页面通过动态组件渲染 tab panel，不能只扫描父级 `.vue`；必须把注册表中的子面板组件一起纳入检查，学生 dashboard 这类页面的重复 eyebrow 往往藏在子面板首屏。
- 用静态测试约束已知页面，防止 `/scoreboard`、能力画像、教师班级学生页、教师仪表盘再次出现重复 eyebrow。
- 静态测试应覆盖已知动态子面板，防止只约束父级页面后再次漏掉 `/student/dashboard`。

## 收获

页面结构统一不能只复用 class，还要判断所在信息架构层级。普通列表页需要 eyebrow 来建立分区；tab 面板内通常不需要，因为 tab 本身已经承担了分区导航和语义说明。
