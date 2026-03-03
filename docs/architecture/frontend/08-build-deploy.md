# 前端构建、部署与开发规范

---

## 1. 构建配置

### 1.1 环境变量

```bash
# .env.development
# 推荐：开发环境走 Vite Proxy，保持同源（避免 CORS，并支持基于 Cookie 的刷新方案）
VITE_API_BASE_URL=/api/v1
VITE_WS_BASE_URL=/ws

# .env.production
VITE_API_BASE_URL=/api/v1
# 推荐：生产环境使用相对路径，自动继承当前站点协议（https -> wss）
VITE_WS_BASE_URL=/ws
```

> 备注：如果开发环境必须直连后端（不同域/不同端口），需后端开启 CORS，并在刷新方案采用 HttpOnly Cookie 时确保 `withCredentials` + `SameSite` 配置正确。

### 1.2 Vite 配置要点

```ts
// vite.config.ts
export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) }
  },
  server: {
    proxy: {
      '/api': { target: 'http://localhost:8080', changeOrigin: true },
      '/ws': { target: 'ws://localhost:8080', ws: true }
    }
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'echarts': ['echarts/core', 'echarts/charts', 'echarts/components', 'echarts/renderers', 'vue-echarts'],
          'vendor': ['vue', 'vue-router', 'pinia', '@vueuse/core', 'axios', 'element-plus']
        }
      }
    }
  }
})
```

### 1.2.1 Element Plus（按需引入建议）

- 推荐使用自动按需引入（减少手动 import、支持 tree-shaking）：`unplugin-auto-import` + `unplugin-vue-components` + `ElementPlusResolver`。
- 主题建议走 CSS 变量：在全局样式中用项目的 `@theme` token 映射 Element Plus 变量（例如 `--el-color-primary` 等），避免双套主题体系。

### 1.3 分包策略

| Chunk | 包含内容 | 预估大小 |
|-------|----------|----------|
| vendor | Vue + Router + Pinia + Axios + Element Plus | ~80KB gzip |
| echarts | ECharts 按需模块 | ~120KB gzip |
| views/* | 各页面按路由自动分割 | 各 5-20KB |
| index | 入口 + 全局组件 + 样式 | ~30KB gzip |

---

## 2. 部署方案

### 2.1 Nginx 配置

```nginx
server {
    listen 80;
    server_name ctf.campus.edu;
    root /var/www/ctf/dist;
    index index.html;

    # SPA fallback
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API 代理
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Request-ID $request_id;
    }

    # WebSocket 代理
    location /ws/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 86400s;
    }

    # 静态资源缓存
    location /assets/ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

### 2.2 构建部署流程

```
npm run build → dist/ → 复制到 Nginx 静态目录 → reload Nginx
```

---

## 3. 开发规范

### 3.1 文件命名

| 类型 | 命名规则 | 示例 |
|------|----------|------|
| 视图组件 | PascalCase + View/List/Detail 后缀 | `DashboardView.vue` |
| 通用组件 | PascalCase + App 前缀 | `AppButton.vue` |
| Composable | camelCase + use 前缀 | `useCountdown.ts` |
| Store | camelCase | `auth.ts` |
| API 模块 | camelCase | `challenge.ts` |
| 工具函数 | camelCase | `format.ts` |

### 3.2 组件规范

- 使用 `<script setup>` 语法
- 统一使用 `<script setup lang="ts">`，并在同目录内避免同时存在同名 `.js/.ts` 文件
- Props 使用 `defineProps` 声明类型
- Emits 使用 `defineEmits` 显式声明
- 模板中使用 PascalCase 引用组件
- 样式使用 Tailwind 类名，避免 `<style>` 块（除非需要深度选择器）
- Element Plus 组件优先使用 CSS 变量/Props 控制样式，避免大面积 `:deep()` 覆盖内部样式（升级成本高）

### 3.4 类型检查（建议加入 CI）

- `vue-tsc --noEmit`：做 SFC/TS 类型检查（不产物）

### 3.3 Git 分支与提交

- 分支: `feat/页面名` 或 `fix/问题描述`
- 提交: 遵循全局 CLAUDE.md 中的 commit 规范
- 每个页面可独立提交，保持原子性

---

## 4. 性能优化策略

| 策略 | 实现方式 |
|------|----------|
| 路由懒加载 | 所有视图 `() => import()` |
| ECharts 按需引入 | 仅导入 radar/line/bar/gauge + canvas renderer |
| 图片懒加载 | `loading="lazy"` 属性 |
| 虚拟滚动 | 审计日志等大列表场景（Phase 2） |
| 请求去重 | Axios 拦截器对相同 GET 请求去重 |
| 骨架屏 | 页面级 AppSkeleton 替代白屏 |
| prefers-reduced-motion | 所有动画降级为 instant |
