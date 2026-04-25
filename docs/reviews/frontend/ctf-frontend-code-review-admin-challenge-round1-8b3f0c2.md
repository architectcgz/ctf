# CTF 前端代码 Review（admin-challenge 第 1 轮）：靶场管理页面功能完善

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | admin-challenge |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | commit 8b3f0c2，5 个文件，+158/-4 行 |
| 变更概述 | 完善管理员端靶场管理功能，添加描述、镜像、Flag、标签字段及发布前校验 |
| 审查基准 | docs/tasks/frontend-task-breakdown.md（FE-T2 任务要求） |
| 审查日期 | 2026-03-04 |
| 上轮问题数 | N/A（首次审查） |

## 变更文件

1. `code/frontend/src/api/contracts.ts` - 扩展 AdminChallengeListItem 类型
2. `code/frontend/src/api/admin.ts` - 添加 getChallengeDetail 接口
3. `code/frontend/src/views/admin/ChallengeManage.vue` - 增强表单字段和校验
4. `code/frontend/src/views/admin/ChallengeDetail.vue` - 新建详情页
5. `code/frontend/src/router/index.ts` - 添加详情页路由

## 问题列表

### 🟡 中优先级问题（3 项）

#### M1. 缺少提示（hints）字段支持

**位置**：`ChallengeManage.vue`

**问题**：
- 任务要求支持"提示列表（可添加多条，每条可编辑/删除）"
- 当前实现缺少 hints 字段的表单输入

**影响**：功能不完整，无法满足 FE-T2 验收标准

**建议**：
添加动态提示列表组件，支持添加/删除多条提示

---

#### M2. 缺少资源限制（resource_limits）字段

**位置**：`ChallengeManage.vue`

**问题**：
- 任务要求支持"资源限制（CPU 核数、内存 MB）"
- 当前实现缺少 resource_limits 字段

**影响**：无法配置容器资源限制

**建议**：
添加 CPU 和内存输入框：
```vue
<ElFormItem label="CPU 限制">
  <ElInputNumber v-model="form.cpu" :min="0.1" :max="4" :step="0.1" />
</ElFormItem>
<ElFormItem label="内存限制(MB)">
  <ElInputNumber v-model="form.memory" :min="128" :max="4096" :step="128" />
</ElFormItem>
```

---

#### M3. 镜像加载失败无提示

**位置**：`ChallengeManage.vue` - `loadImages()`

**问题**：
```typescript
async function loadImages() {
  try {
    const res = await getImages({ page: 1, page_size: 100 })
    images.value = res.list.filter(img => img.status === 'ready')
  } catch (error) {
    console.error('加载镜像列表失败', error)  // 只有控制台输出
  }
}
```

**影响**：用户无法感知镜像加载失败

**建议**：
添加用户提示：
```typescript
catch (error) {
  toast.error('加载镜像列表失败')
}
```

---

### 🟢 低优先级问题（4 项）

#### L1. 类型定义不够精确

**位置**：`ChallengeManage.vue` - `openDialog(row?: any)`

**问题**：参数类型使用 `any`，降低类型安全

**建议**：
```typescript
function openDialog(row?: AdminChallengeListItem) {
```

---

#### L2. 详情页缺少编辑功能

**位置**：`ChallengeDetail.vue`

**问题**：
- 任务要求"支持编辑模式切换"
- 当前只能查看，无法编辑

**影响**：用户体验不完整

**建议**：
添加"编辑"按钮，点击后跳转到编辑对话框或切换为编辑模式

---

#### L3. 标签输入体验不佳

**位置**：`ChallengeManage.vue` - 标签字段

**问题**：
- 使用逗号分隔的字符串输入
- 无法可视化管理标签（添加/删除）

**建议**：
使用 Element Plus 的 `ElTag` 组件实现标签输入：
```vue
<ElFormItem label="标签">
  <div class="flex flex-wrap gap-2">
    <ElTag v-for="tag in formTags" :key="tag" closable @close="removeTag(tag)">
      {{ tag }}
    </ElTag>
    <ElInput v-model="newTag" size="small" @keyup.enter="addTag" placeholder="输入后按回车" />
  </div>
</ElFormItem>
```

---

#### L4. 详情页缺少加载错误处理

**位置**：`ChallengeDetail.vue`

**问题**：
```typescript
} catch (error) {
  console.error('加载失败', error)  // 只有控制台输出
}
```

**影响**：用户无法感知加载失败

**建议**：
显示错误提示或空状态页面

---

## 优点

1. ✅ 发布前校验逻辑正确（镜像、Flag、标签必填）
2. ✅ 类型定义扩展合理
3. ✅ 代码风格与现有代码一致
4. ✅ 最小化实现，避免冗余
5. ✅ 镜像选择下拉框正确过滤 ready 状态

## 验收标准对照

| 标准 | 状态 | 说明 |
|------|------|------|
| 管理员可创建包含完整信息的靶机 | ⚠️ 部分完成 | 缺少 hints 和 resource_limits |
| 管理员可编辑靶机的所有字段 | ⚠️ 部分完成 | 列表页可编辑，详情页不可编辑 |
| 管理员可查看靶机详情页 | ✅ 完成 | 详情页已创建 |
| 发布前校验生效 | ✅ 完成 | 镜像、Flag、标签校验正确 |
| 镜像选择下拉框正确显示 | ✅ 完成 | 过滤 ready 状态镜像 |
| 提示和标签可动态添加/删除 | ❌ 未完成 | 标签使用字符串输入，hints 未实现 |

## 总结

本轮改动完成了靶场管理页面的核心功能，包括字段扩展、发布前校验和详情页创建。主要问题是缺少 hints 和 resource_limits 字段，以及部分用户体验细节需要优化。

**建议**：
1. 优先修复 M1、M2（缺失字段）
2. M3 和 L4（错误提示）可快速修复
3. L2、L3（体验优化）可后续迭代

**是否需要修复后复审**：是（存在中优先级问题）
