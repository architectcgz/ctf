# 暗色模式下表单控件不能写死浅色内高光

## 问题描述

`/challenges` 页面的分类、难度和搜索控件在暗色模式下出现了一道明显的白线。问题不在共享 `ui-control` 基础样式本身，而在业务层对筛选控件外壳额外加了一层固定浅色的 `inset` 高光。

本次触发点在：

- `code/frontend/src/components/challenge/ChallengeDirectoryPanel.vue`

## 原因分析

学生端页面已经接入 `journal-shell-user` 主题变量，控件背景、边框、文字和 caret 都应当随主题切换。

但如果业务组件又在局部样式里补一条类似下面的固定浅色内高光：

```css
box-shadow: inset 0 1px 0 color-mix(in srgb, white 30%, transparent);
```

那么 light 模式下它看起来像轻微高光，到了 dark 模式就会直接变成一条白线。这个问题有几个特征：

- 它绕过了共享主题变量，视觉结果只对 light mode 成立。
- 它不是文字颜色类问题，容易在只检查文字/背景时漏掉。
- 它常出现在搜索框、筛选栏、select 包裹层这类“业务组件想再做一点质感”的位置。

## 解决方案

以后凡是表单控件容器涉及边框、背景、内高光、placeholder、caret、focus ring、adornment：

- 禁止写死 `white` 或浅色 `rgba(...)` 作为局部高光。
- 必须从 `journal-border`、`journal-surface`、`color-border-default` 这类主题语义变量推导。
- dark mode 检查时，除了文字和光标，还要额外检查：
  - 顶部分隔线
  - `inset` 阴影
  - prefix / suffix 图标分隔感
  - focus ring 是否过亮

本次项目级硬规则已补到：

- `AGENTS.md` 的 `Frontend Guardrails`

本次实际修复点：

- `code/frontend/src/components/challenge/ChallengeDirectoryPanel.vue`
- `code/frontend/src/views/challenges/__tests__/ChallengeList.test.ts`

## 收获

暗色模式问题不能只看“文字能不能看见”。表单类控件还要检查所有基于亮面假设的细节层：

- 内高光
- 描边叠色
- 半透明白色投影
- placeholder / caret / icon 的对比度

如果某个局部样式是在共享主题控件之上再叠加“质感”，默认就应当怀疑它是否只对 light mode 成立。先用主题变量表达，再做 dark mode 目测或断言，才能避免同类白线问题反复出现。
