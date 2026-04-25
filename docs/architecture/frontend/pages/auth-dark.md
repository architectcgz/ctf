# 页面设计：认证页面 (Auth)

> 继承：../design-system/MASTER.md | 技术栈：Element Plus + Tailwind CSS

---

## 1. 登录页面 (LoginView)

### 布局结构

```
┌──────────────────────────────────────────────┐
│  [Logo] CTF 网络攻防靶场平台                   │
│                                               │
│  ┌─────────────────────────────────────────┐ │
│  │  登录                                    │ │
│  │  ─────────────────────────────────────  │ │
│  │  用户名/邮箱 [____________]              │ │
│  │  密码       [____________] [显示]        │ │
│  │  [ ] 记住我                              │ │
│  │  [登录按钮]                              │ │
│  │  还没有账号？[立即注册]                   │ │
│  └─────────────────────────────────────────┘ │
└──────────────────────────────────────────────┘
```

### 技术栈使用

**Element Plus 组件：**
- `<ElForm>` + `<ElFormItem>` - 表单容器
- `<ElInput>` - 用户名和密码输入框
- `<ElCheckbox>` - 记住我选项
- `<ElButton type="primary">` - 登录按钮

**Tailwind CSS：**
- 居中布局：`flex items-center justify-center min-h-screen`
- 卡片容器：`max-w-md w-full space-y-6 p-8`
- 间距：`space-y-4`, `gap-4`

### 代码示例

```vue
<template>
  <div class="flex min-h-screen items-center justify-center bg-[#0f1117]">
    <div class="w-full max-w-md space-y-6 rounded-lg border border-[#30363d] bg-[#161b22] p-8">
      <div class="text-center">
        <h1 class="text-2xl font-bold text-[#e6edf3]">登录</h1>
      </div>

      <ElForm :model="form" :rules="rules" ref="formRef">
        <ElFormItem prop="username">
          <ElInput
            v-model="form.username"
            placeholder="用户名或邮箱"
            size="large"
          />
        </ElFormItem>

        <ElFormItem prop="password">
          <ElInput
            v-model="form.password"
            type="password"
            placeholder="密码"
            size="large"
            show-password
          />
        </ElFormItem>

        <ElFormItem>
          <ElCheckbox v-model="form.remember">记住我</ElCheckbox>
        </ElFormItem>

        <ElButton type="primary" size="large" class="w-full" @click="handleLogin">
          登录
        </ElButton>
      </ElForm>

      <div class="text-center text-sm text-[#8b949e]">
        还没有账号？
        <router-link to="/register" class="text-[#0891b2] hover:text-[#06b6d4]">
          立即注册
        </router-link>
      </div>
    </div>
  </div>
</template>
```

---

## 2. 注册页面 (RegisterView)

### 布局结构

```
┌──────────────────────────────────────────────┐
│  [Logo] CTF 网络攻防靶场平台                   │
│                                               │
│  ┌─────────────────────────────────────────┐ │
│  │  注册                                    │ │
│  │  ─────────────────────────────────────  │ │
│  │  用户名     [____________]               │ │
│  │  邮箱       [____________]               │ │
│  │  密码       [____________] [显示]        │ │
│  │  确认密码   [____________] [显示]        │ │
│  │  [ ] 我已阅读并同意服务条款               │ │
│  │  [注册按钮]                              │ │
│  │  已有账号？[立即登录]                     │ │
│  └─────────────────────────────────────────┘ │
└──────────────────────────────────────────────┘
```

### 技术栈使用

**Element Plus 组件：**
- `<ElForm>` + `<ElFormItem>` - 表单容器
- `<ElInput>` - 所有输入框
- `<ElCheckbox>` - 服务条款同意
- `<ElButton type="primary">` - 注册按钮

**Tailwind CSS：**
- 布局同登录页
- 间距：`space-y-4`

### 表单验证规则

```typescript
const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度为 3-20 个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少 6 个字符', trigger: 'blur' }
  ]
}
```

---

## 3. Element Plus 主题覆盖

```css
/* 认证页面特定样式 */
.el-form-item {
  margin-bottom: 20px;
}

.el-input__wrapper {
  background-color: #0f1117;
  border-color: #30363d;
}

.el-input__wrapper:hover {
  border-color: #0891b2;
}

.el-input__wrapper.is-focus {
  border-color: #0891b2;
  box-shadow: 0 0 0 1px rgba(8, 145, 178, 0.3);
}

.el-button--primary {
  background-color: #0891b2;
  border-color: #0891b2;
}

.el-button--primary:hover {
  background-color: #06b6d4;
  border-color: #06b6d4;
}
```

---

## 4. 响应式设计

```vue
<div class="flex min-h-screen items-center justify-center bg-[#0f1117] px-4">
  <div class="w-full max-w-md space-y-6 rounded-lg border border-[#30363d] bg-[#161b22] p-6 sm:p-8">
    <!-- 内容 -->
  </div>
</div>
```

- 移动端：`px-4`, `p-6`
- 桌面端：`sm:p-8`
