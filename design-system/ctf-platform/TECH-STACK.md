# 技术栈使用规范

> CTF 平台前端技术栈协调方案

---

## 1. 技术栈组成

- **Vue 3** - 框架
- **Element Plus** - UI 组件库（主要）
- **Tailwind CSS** - 样式系统（辅助）

---

## 2. 使用原则

### 2.1 优先使用 Element Plus

**所有页面默认使用 Element Plus 组件**

Element Plus 提供：
- 完整的交互逻辑
- 无障碍支持
- 响应式设计
- 主题定制能力

### 2.2 Tailwind CSS 辅助场景

**使用 Tailwind CSS 的场景：**
- 布局（flex, grid, spacing, positioning）
- Element Plus 未覆盖的样式
- 自定义间距和尺寸
- 响应式断点
- 工具类（text-center, hidden 等）

---

## 3. 组件使用规范

### 3.1 优先使用 Element Plus 组件

**基础组件：**
- `<ElButton>` - 按钮
- `<ElInput>` - 输入框
- `<ElSelect>` - 下拉选择
- `<ElTable>` - 表格
- `<ElCard>` - 卡片
- `<ElTag>` - 标签
- `<ElPagination>` - 分页

**复杂组件：**
- `<ElDialog>` - 弹窗
- `<ElForm>` + `<ElFormItem>` - 表单
- `<ElDatePicker>` - 日期选择
- `<ElUpload>` - 文件上传
- `<ElSwitch>` - 开关
- `<ElDrawer>` - 抽屉

### 3.2 Tailwind CSS 辅助使用

**布局类：**
```vue
<div class="flex items-center justify-between gap-4">
  <ElButton type="primary">按钮</ElButton>
</div>
```

**间距类：**
```vue
<div class="space-y-6 p-6">
  <ElCard>内容</ElCard>
</div>
```

**响应式：**
```vue
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
  <ElCard v-for="item in list" :key="item.id">
    {{ item.title }}
  </ElCard>
</div>
```

---

## 4. 样式覆盖策略

### 4.1 Element Plus 主题覆盖

```css
/* src/assets/styles/element-override.css */
:root {
  --el-color-primary: #0891b2;
  --el-bg-color: #161b22;
  --el-text-color-primary: #e6edf3;
  --el-border-color: #30363d;
}
```

### 4.2 组件级覆盖

```vue
<ElDialog
  class="!bg-[#161b22] !border-[#30363d]"
  :close-on-click-modal="false"
>
```

---

## 5. 迁移检查清单

**展示型页面：**
- [ ] 移除所有 `El*` 组件
- [ ] 使用 Tailwind 实现所有样式
- [ ] 使用 design-system 定义的颜色

**管理型页面：**
- [ ] 保留 ElDialog/ElForm/ElSelect
- [ ] 替换 ElButton/ElTable/ElTag
- [ ] 覆盖 Element Plus 主题色
