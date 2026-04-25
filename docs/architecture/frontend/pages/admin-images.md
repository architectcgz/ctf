# 页面设计：镜像管理 (Image Management)

> 继承：../design-system/MASTER.md | 角色：管理员 | 位置：管理后台子页
> 技术栈：Element Plus (主) + Tailwind CSS (辅)

---

## 技术栈

**Element Plus：** `<ElTable>`, `<ElDialog>`, `<ElForm>`, `<ElTag>`
**Tailwind CSS：** 布局、间距

---

## 布局结构

```
┌──────────────────────────────────────────────┐
│  "镜像管理"                      [上传镜像]   │
│  [搜索] [状态▼]                               │
├──────────────────────────────────────────────┤
│  表格                                         │
│  □ | 镜像名称 | Tag | 大小 | 关联靶场 | 状态 | 操作│
│  ─────────────────────────────────────────── │
│  □  ctf/sqli-basic  latest  128MB  2  可用  [▼]│
│  □  ctf/xss-lab     v1.2    256MB  1  可用  [▼]│
│  □  ctf/pwn-stack   latest  512MB  0  未用  [▼]│
│  ─────────────────────────────────────────── │
│  已选 0 项  [批量删除]                         │
│  [分页]                                       │
└──────────────────────────────────────────────┘
```

### 组件细节

- 镜像名称：`font-mono text-sm`
- 大小：`font-mono text-secondary`
- 状态：可用 `text-success` / 未用 `text-muted` / 拉取中 `text-primary`
- 操作菜单：查看详情、删除（关联靶场 >0 时禁用删除，tooltip 提示原因）
- 上传镜像：Dialog，支持输入镜像地址从 Registry 拉取，显示拉取进度条
