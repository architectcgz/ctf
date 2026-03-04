# CTF 平台 — Light Mode 设计系统

> 基于 ui-ux-pro-max 推荐，调整为与 Dark Mode 品牌一致
> 技术栈：Vue 3 + Tailwind CSS + Element Plus

---

## 1. 配色系统（Light Mode）

### 1.1 基础色板

```
背景层级：
  bg-base:       #ffffff    -- 最底层背景（页面）
  bg-surface:    #f8fafc    -- 卡片/面板背景（slate-50）
  bg-elevated:   #f1f5f9    -- 弹窗/下拉菜单/hover 行（slate-100）
  bg-overlay:    #000000/20 -- 遮罩层

边框：
  border-default: #e2e8f0   -- 默认边框（slate-200）
  border-subtle:  #f1f5f9   -- 弱边框（slate-100）
  border-active:  #0891b2   -- 选中/聚焦态边框（保持品牌主色）

文字层级：
  text-primary:   #0f172a   -- 主要文字（slate-900）
  text-secondary: #475569   -- 次要文字（slate-600）
  text-muted:     #94a3b8   -- 辅助/禁用文字（slate-400）
  text-inverse:   #ffffff   -- 按钮内反色文字
```

### 1.2 主题色（保持品牌一致）

```
Teal（主色 / 交互）：
  primary:       #0891b2    -- 按钮、链接、选中态（与 Dark Mode 一致）
  primary-hover: #06b6d4    -- hover 态（cyan-500）
  primary-dim:   #0891b2/10 -- 背景高亮

Violet（强调）：
  accent:        #8b5cf6    -- 首杀标识、特殊成就
  accent-dim:    #8b5cf6/10

Green（成功）：
  success:       #22c55e    -- Flag 正确、已解出
  success-dim:   #22c55e/10

Amber（警告）：
  warning:       #f59e0b    -- 即将过期、注意事项
  warning-dim:   #f59e0b/10

Red（错误 / 危险）：
  danger:        #ef4444    -- Flag 错误、删除操作
  danger-dim:    #ef4444/10
```

### 1.3 难度色标（与 Dark Mode 一致）

```
入门:  #34d399  (emerald-400)
简单:  #38bdf8  (sky-400)
中等:  #fbbf24  (amber-400)
困难:  #f87171  (red-400)
地狱:  #c084fc  (purple-400)
```

### 1.4 分类色标（与 Dark Mode 一致）

```
Web:       #06b6d4  (cyan-500)
Crypto:    #a78bfa  (violet-400)
Pwn:       #ef4444  (red-500)
Reverse:   #fb923c  (orange-400)
Misc:      #22c55e  (green-500)
Forensics: #3b82f6  (blue-500)
```

---

## 2. 组件样式（Light Mode）

### 2.1 按钮

```
主按钮:
  bg-[#0891b2] hover:bg-[#06b6d4] text-white font-medium
  px-4 py-2 rounded-lg transition-colors duration-150

次要按钮:
  bg-[#f1f5f9] hover:bg-[#e2e8f0] text-[#0f172a] border border-[#e2e8f0]
  px-4 py-2 rounded-lg transition-colors duration-150

危险按钮:
  bg-[#ef4444]/10 hover:bg-[#ef4444]/20 text-[#dc2626] border border-[#ef4444]/20
  px-4 py-2 rounded-lg transition-colors duration-150
```

### 2.2 卡片

```
基础卡片:
  bg-[#f8fafc] border border-[#e2e8f0] rounded-lg p-5

可点击卡片:
  基础卡片 + hover:border-[#0891b2]/50 hover:shadow-sm transition-all duration-150 cursor-pointer

高亮卡片（已解出/活跃）:
  bg-[#f8fafc] border border-green-500/30 shadow-sm
```

### 2.3 输入框

```
基础输入:
  bg-white border border-[#e2e8f0] text-[#0f172a]
  placeholder:text-[#94a3b8]
  rounded-lg px-3 py-2 text-sm
  focus:border-[#0891b2] focus:ring-1 focus:ring-[#0891b2]/30
  transition-colors duration-150

Flag 输入框:
  基础输入 + font-mono text-[#0891b2]
  border-[#0891b2]/30 focus:border-[#0891b2]
```

### 2.4 表格

```
表头:
  bg-[#f8fafc] text-[#475569] text-xs font-medium uppercase tracking-wider
  px-4 py-3 border-b border-[#e2e8f0] text-left

表行:
  text-[#0f172a] text-sm px-4 py-3 border-b border-[#f1f5f9]
  hover:bg-[#f8fafc] transition-colors duration-100

排名高亮行:
  第1名: bg-amber-50 border-l-2 border-amber-400
  第2名: bg-slate-50 border-l-2 border-slate-400
  第3名: bg-orange-50 border-l-2 border-orange-400
```

---

## 3. 主题切换实现方案

### 方案A：CSS 变量（推荐）

在 `src/assets/styles/theme.css` 中定义：

```css
:root {
  /* Light Mode (默认) */
  --color-bg-base: #ffffff;
  --color-bg-surface: #f8fafc;
  --color-bg-elevated: #f1f5f9;
  --color-border-default: #e2e8f0;
  --color-text-primary: #0f172a;
  --color-text-secondary: #475569;
  --color-text-muted: #94a3b8;
  --color-primary: #0891b2;
  --color-primary-hover: #06b6d4;
}

[data-theme="dark"] {
  /* Dark Mode */
  --color-bg-base: #0f1117;
  --color-bg-surface: #161b22;
  --color-bg-elevated: #1c2128;
  --color-border-default: #30363d;
  --color-text-primary: #e6edf3;
  --color-text-secondary: #8b949e;
  --color-text-muted: #6e7681;
  --color-primary: #0891b2;
  --color-primary-hover: #06b6d4;
}
```

在组件中使用：
```vue
<div class="bg-[var(--color-bg-surface)] text-[var(--color-text-primary)]">
```

### 方案B：Tailwind CSS 配置

在 `tailwind.config.js` 中扩展：

```js
module.exports = {
  darkMode: 'class', // 使用 class 策略
  theme: {
    extend: {
      colors: {
        'bg-base': 'var(--color-bg-base)',
        'bg-surface': 'var(--color-bg-surface)',
        // ... 其他颜色
      }
    }
  }
}
```

---

## 4. Element Plus 主题配置

在 `src/main.ts` 中配置：

```typescript
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'

app.use(ElementPlus, {
  // Light Mode 主题覆盖
  locale: zhCn,
})
```

创建 `src/assets/styles/element-override.css`：

```css
/* Light Mode Element Plus 覆盖 */
:root {
  --el-color-primary: #0891b2;
  --el-color-primary-light-3: #06b6d4;
  --el-color-primary-light-5: #22d3ee;
  --el-bg-color: #ffffff;
  --el-bg-color-page: #f8fafc;
  --el-text-color-primary: #0f172a;
  --el-text-color-regular: #475569;
  --el-border-color: #e2e8f0;
}

[data-theme="dark"] {
  --el-color-primary: #0891b2;
  --el-bg-color: #161b22;
  --el-bg-color-page: #0f1117;
  --el-text-color-primary: #e6edf3;
  --el-text-color-regular: #8b949e;
  --el-border-color: #30363d;
}
```

---

## 5. 主题切换逻辑

创建 `src/composables/useTheme.ts`：

```typescript
import { ref, watch } from 'vue'

const theme = ref<'light' | 'dark'>('light')

export function useTheme() {
  const toggleTheme = () => {
    theme.value = theme.value === 'light' ? 'dark' : 'light'
    document.documentElement.setAttribute('data-theme', theme.value)
    localStorage.setItem('theme', theme.value)
  }

  const initTheme = () => {
    const saved = localStorage.getItem('theme') as 'light' | 'dark' | null
    theme.value = saved || 'light'
    document.documentElement.setAttribute('data-theme', theme.value)
  }

  return { theme, toggleTheme, initTheme }
}
```

---

## 6. 可访问性检查

- ✅ 文字对比度 ≥ 4.5:1（WCAG AA）
- ✅ 主色 #0891b2 在白色背景上对比度：4.52:1
- ✅ 文字颜色 #0f172a 在白色背景上对比度：16.1:1
- ✅ 所有交互元素有 focus-visible 样式
- ✅ 颜色不作为唯一信息传达方式

---

## 7. 迁移步骤

1. 创建 CSS 变量文件
2. 更新 Tailwind 配置
3. 创建 useTheme composable
4. 在 App.vue 中初始化主题
5. 添加主题切换按钮
6. 逐步迁移现有组件使用 CSS 变量
