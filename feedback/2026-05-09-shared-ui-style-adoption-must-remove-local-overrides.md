# 采用共享 UI 样式时必须移除同类局部覆盖

## 问题描述

在统一管理员和教师端 header 操作按钮时，按钮 class 已经从旧的局部体系切到共享 `header-btn`，但 `/platform/challenges` 的“导入题目”按钮仍保留了 `challenge-manage-import-button` 及一组局部 `--header-btn-*` 变量覆盖。

结果是代码表面上已经复用了共享原语，实际视觉仍然偏离其它 header 主按钮，形成“换了 class 名但没有真正统一样式”的问题。

## 原因分析

- 只检查了旧类名是否清空，没有继续检查同类局部覆盖是否仍在改变共享原语的视觉结果。
- 把“接入共享 class”误当成完成标准，而没有把“移除冲突的页面私有变量、私有 modifier、私有尺寸/颜色覆盖”纳入完成标准。
- 测试一开始还在断言局部按钮变量，反而固化了不一致样式。

## 解决方案

- 做共享 UI 原语统一时，完成标准必须包含：
  - class / 组件引用已经切到共享原语；
  - 同一元素上的页面私有 class 已移除，除非它只负责布局定位且不覆盖视觉变量；
  - 同类局部 CSS 变量覆盖、尺寸覆盖、颜色覆盖、hover/focus 覆盖已删除；
  - 测试断言应验证共享原语继承关系，而不是继续要求页面私有变量存在。
- 如果某个页面确实需要保留局部覆盖，必须能说明它是业务语义差异，而不是样式漂移。

## 收获

共享样式统一不是“把 class 名换成共享名”就结束。真正的统一要同时清理旧局部样式入口，否则私有变量仍会让页面视觉继续分叉。

后续遇到按钮、tag、列表、toolbar、表格等设计系统原语统一任务时，应默认扫描并移除同类局部覆盖。

## 交叉链接

- `code/frontend/src/style.css`
- `code/frontend/src/components/platform/challenge/ChallengeManageHeroPanel.vue`
- `works/harness-good-practices.md`
