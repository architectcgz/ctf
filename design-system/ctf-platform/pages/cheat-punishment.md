# 页面设计：作弊检测增强 (Cheat Detection Enhancement)

> 继承：design-system/MASTER.md | 角色：管理员 | 任务：T28
> 技术栈：Element Plus (主) + Tailwind CSS (辅)
> 基于：admin-cheat.md（Phase 1 已有基础设计）
> 本文档仅描述 Phase 2 新增的处罚操作流程

---

## 技术栈

**Element Plus：** `<ElDialog>`, `<ElCheckbox>`, `<ElInput>`, `<ElButton>`
**Tailwind CSS：** 布局、间距

---

## 1. 处罚操作流程

在现有 admin-cheat.md 的"确认作弊"按钮后，新增处罚选择步骤。

### 确认作弊 + 选择处罚 Dialog

```
┌─────────────────────────────────────┐
│  "确认作弊并处罚"            [关闭]  │
├─────────────────────────────────────┤
│  涉及用户: zhangsan, lisi           │
│  涉及题目: Web-03 SQL注入进阶       │
│  涉及竞赛: 2026 春季 CTF 挑战赛     │
│                                     │
│  选择处罚措施                       │
│  ┌─────────────────────────────┐    │
│  │ ☑ 扣除该题得分              │    │
│  │   扣除 zhangsan 82pts       │    │
│  │   扣除 lisi 82pts           │    │
│  └─────────────────────────────┘    │
│  ┌─────────────────────────────┐    │
│  │ ☐ 禁赛                     │    │
│  │   禁止参加本场剩余比赛       │    │
│  └─────────────────────────────┘    │
│  ┌─────────────────────────────┐    │
│  │ ☐ 警告                     │    │
│  │   发送警告通知，不扣分       │    │
│  └─────────────────────────────┘    │
│                                     │
│  处罚说明 (可选)                    │
│  [短时间内提交相同Flag，判定共享__]  │
│                                     │
│  [取消]          [确认处罚(danger)]  │
└─────────────────────────────────────┘
```

- Dialog 宽度：`max-w-md`
- 处罚选项：checkbox 多选，至少选一项
- 默认勾选"扣除该题得分"

### 处罚选项卡片

```
每个选项:
  bg-surface border border-border rounded-lg p-3 mb-2
  hover:border-primary/50 cursor-pointer

选中态:
  border-primary bg-primary/5

checkbox: 左侧
标题: text-sm font-medium text-primary
描述: text-xs text-secondary mt-1
```

---

## 2. 处罚记录

在告警详情 Drawer 中新增"处罚记录"区域（已处理的告警）：

```
── 处罚记录 ──
处罚时间: 2026-03-01 15:45
处罚人: admin
措施: 扣除该题得分 + 警告
说明: "短时间内提交相同 Flag，判定共享"

影响:
  zhangsan: -82 pts (Web-03)
  lisi: -82 pts (Web-03)
```

### 样式

```
mt-4 pt-4 border-t border-border-subtle

标签: text-xs text-muted uppercase tracking-wider "处罚记录"
内容: text-sm text-secondary
影响列表: font-mono text-xs text-danger
```

---

## 3. 告警列表增强

在现有 admin-cheat.md 告警卡片中新增：

```
已处罚标签:
  bg-danger/10 text-danger text-xs px-2 py-0.5 rounded
  "已处罚: 扣分+警告"

替换原有的"已确认"状态，细化为:
  已确认-已处罚: bg-danger/10 text-danger "已处罚"
  已确认-仅标记: bg-border text-muted "已确认"
```
