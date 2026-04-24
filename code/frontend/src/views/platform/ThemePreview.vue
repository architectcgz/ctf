<script setup lang="ts">
import { ref } from 'vue'
import {
  Bell,
  Book,
  CheckCircle,
  Database,
  Edit3,
  Eye,
  FileSearch,
  Filter,
  Layout,
  LogOut,
  MoreHorizontal,
  Search,
  Settings,
  ShieldCheck,
  Sidebar as SidebarIcon,
  Users,
  Zap,
  ChevronRight,
} from 'lucide-vue-next'

type PreviewLayout = 'variant1' | 'variant2' | 'variant3'
type PreviewTab = 'manage' | 'import' | 'queue'
type PreviewTone = 'primary' | 'success' | 'warning' | 'danger'

const currentLayout = ref<PreviewLayout>('variant2')
const activeTab = ref<PreviewTab>('manage')
const sidebarCollapsed = ref(false)

const layoutOptions: Array<{ key: PreviewLayout; label: string }> = [
  { key: 'variant1', label: '方案1：导航融合' },
  { key: 'variant2', label: '方案2：分层指挥' },
  { key: 'variant3', label: '方案3：沉浸画布' },
]

const navItems = [
  { label: '概览中心', icon: Layout },
  { label: '题目管理', icon: FileSearch },
  { label: '竞赛编排', icon: ShieldCheck },
  { label: '用户治理', icon: Users },
  { label: '系统设置', icon: Settings },
]

const tabs: Array<{ key: PreviewTab; label: string }> = [
  { key: 'manage', label: '题库管理' },
  { key: 'import', label: '上传题目包' },
  { key: 'queue', label: '导入队列' },
]

const stats = [
  { label: '题目总量', value: 256, icon: Book, trend: '+12%', tone: 'primary' },
  { label: '已发布', value: 184, icon: CheckCircle, trend: '72%', tone: 'success' },
  { label: '运行环境', value: 42, icon: Zap, trend: '正常', tone: 'warning' },
  { label: '待处理', value: 12, icon: Edit3, trend: '需审核', tone: 'danger' },
] as const satisfies ReadonlyArray<{
  label: string
  value: number
  icon: unknown
  trend: string
  tone: PreviewTone
}>

const challenges = [
  { id: 1, uuid: 'WEB-SSR-01', title: '内部笔记下载器：服务端请求伪造', category: 'Web', difficulty: '简单', points: 100 },
  { id: 2, uuid: 'PWN-HEAP-05', title: '堆溢出：Tcache Poisoning 攻击实战', category: 'Pwn', difficulty: '困难', points: 850 },
  { id: 3, uuid: 'MISC-TRAF-09', title: '流量分析：隐藏协议识别与提取', category: 'Misc', difficulty: '中等', points: 300 },
]

function getStatToneClass(tone: PreviewTone): string {
  return `theme-preview__stat-trend--${tone}`
}
</script>

<template>
  <div class="theme-preview min-h-screen flex overflow-hidden">
    <aside
      v-if="currentLayout !== 'variant3'"
      :class="[
        'theme-preview__sidebar flex flex-col transition-all duration-300 shrink-0',
        currentLayout === 'variant2' ? (sidebarCollapsed ? 'w-20' : 'w-64') : 'w-64',
      ]"
    >
      <div class="theme-preview__sidebar-brand h-16 flex items-center px-6 gap-3 shrink-0">
        <div class="theme-preview__brand-mark w-8 h-8 rounded-lg flex items-center justify-center shrink-0">
          <Database class="w-4 h-4" />
        </div>
        <span
          v-if="!sidebarCollapsed || currentLayout === 'variant1'"
          class="theme-preview__brand-text font-bold tracking-tight uppercase"
        >
          CTF<span class="theme-preview__brand-accent">Ops</span>
        </span>
      </div>

      <nav class="flex-1 p-4 space-y-2">
        <div
          v-for="item in navItems"
          :key="item.label"
          :class="[
            'theme-preview__nav-item flex items-center gap-3 px-4 py-3 transition-all cursor-pointer',
            item.label === '题目管理' ? 'theme-preview__nav-item--active' : '',
          ]"
        >
          <component
            :is="item.icon"
            class="w-5 h-5"
          />
          <span
            v-if="!sidebarCollapsed || currentLayout === 'variant1'"
            class="text-sm font-bold"
          >
            {{ item.label }}
          </span>
        </div>
      </nav>

      <div class="theme-preview__sidebar-footer p-4">
        <div class="theme-preview__logout-row flex items-center gap-3 px-4 py-3">
          <LogOut class="w-5 h-5" />
          <span
            v-if="!sidebarCollapsed || currentLayout === 'variant1'"
            class="text-sm font-bold"
          >
            退出系统
          </span>
        </div>
      </div>
    </aside>

    <div class="flex-1 flex flex-col min-w-0 h-screen overflow-hidden">
      <header class="theme-preview__layout-bar px-6 py-2 flex justify-between items-center shrink-0">
        <div class="flex items-center gap-4">
          <span class="theme-preview__layout-label">Layout Switcher</span>
          <div class="theme-preview__layout-switcher flex p-1">
            <button
              v-for="option in layoutOptions"
              :key="option.key"
              :class="[
                'theme-preview__layout-option px-3 py-1 transition-all',
                currentLayout === option.key ? 'theme-preview__layout-option--active' : '',
              ]"
              @click="currentLayout = option.key"
            >
              {{ option.label }}
            </button>
          </div>
        </div>
        <div class="theme-preview__layout-meta">
          PREVIEW MODE / {{ currentLayout.toUpperCase() }}
        </div>
      </header>

      <header class="theme-preview__page-header shrink-0">
        <div class="px-8 h-16 flex items-center justify-between">
          <div class="flex items-center gap-4">
            <button
              v-if="currentLayout === 'variant2'"
              class="theme-preview__icon-button theme-preview__icon-button--muted p-2"
              @click="sidebarCollapsed = !sidebarCollapsed"
            >
              <SidebarIcon class="w-5 h-5" />
            </button>
            <div class="flex items-center gap-2 text-sm">
              <span class="theme-preview__breadcrumb">资源管理</span>
              <ChevronRight class="theme-preview__breadcrumb-separator w-4 h-4" />
              <span class="theme-preview__page-title">题目管理中心</span>
            </div>
          </div>
          <div class="flex items-center gap-4">
            <button class="theme-preview__icon-button theme-preview__icon-button--muted p-2.5 relative">
              <Bell class="w-5 h-5" />
              <span class="theme-preview__alert-dot absolute top-2.5 right-2.5 w-2 h-2 rounded-full" />
            </button>
            <div class="theme-preview__avatar-shell w-10 h-10 rounded-2xl flex items-center justify-center overflow-hidden">
              <img
                src="https://api.dicebear.com/7.x/avataaars/svg?seed=Felix"
                alt="avatar"
              >
            </div>
          </div>
        </div>

        <div
          v-if="currentLayout === 'variant2'"
          class="px-8 flex gap-8"
        >
          <button
            v-for="tab in tabs"
            :key="tab.key"
            :class="[
              'theme-preview__header-tab pb-4 mt-2',
              activeTab === tab.key ? 'theme-preview__header-tab--active' : '',
            ]"
            @click="activeTab = tab.key"
          >
            {{ tab.label }}
          </button>
        </div>
      </header>

      <div class="theme-preview__content flex-1 overflow-y-auto p-8 custom-scrollbar">
        <nav
          v-if="currentLayout === 'variant1'"
          class="theme-preview__content-tabs flex gap-2 mb-8 p-1 w-fit"
        >
          <button
            v-for="tab in tabs"
            :key="tab.key"
            :class="[
              'theme-preview__content-tab px-6 py-2 transition-all',
              activeTab === tab.key ? 'theme-preview__content-tab--active' : '',
            ]"
            @click="activeTab = tab.key"
          >
            {{ tab.label }}
          </button>
        </nav>

        <div class="theme-preview__body mx-auto">
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-10">
            <div
              v-for="item in stats"
              :key="item.label"
              class="theme-preview__stat-card p-6 group relative overflow-hidden"
            >
              <div class="flex flex-col h-full relative z-10">
                <div class="theme-preview__stat-head flex justify-between items-start mb-6">
                  <span>{{ item.label }}</span>
                  <component
                    :is="item.icon"
                    class="w-4 h-4"
                  />
                </div>
                <div class="flex items-end justify-between gap-3">
                  <div class="flex flex-col">
                    <h3 class="theme-preview__stat-value">
                      {{ item.value }}
                    </h3>
                    <p class="theme-preview__stat-helper mt-3">
                      核心指标说明
                    </p>
                  </div>
                  <div
                    :class="['theme-preview__stat-trend', getStatToneClass(item.tone)]"
                  >
                    {{ item.trend }}
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="flex flex-col md:flex-row justify-between items-center gap-4 mb-6">
            <div class="relative w-full md:w-96">
              <Search class="theme-preview__search-icon absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4" />
              <input
                type="text"
                placeholder="检索资源..."
                class="theme-preview__search-input w-full pl-11 pr-4 py-3"
              >
            </div>
            <div class="flex items-center gap-2">
              <button class="theme-preview__secondary-button px-4 py-2.5 flex items-center gap-2">
                <Filter class="w-4 h-4" />
                高级筛选
              </button>
              <button class="theme-preview__primary-button px-6 py-2.5">
                + 新建题目
              </button>
            </div>
          </div>

          <div class="theme-preview__table-shell overflow-hidden">
            <table class="w-full text-left border-collapse table-fixed">
              <thead>
                <tr class="theme-preview__table-head">
                  <th class="theme-preview__table-title-col px-8 py-5">
                    题目资源名称
                  </th>
                  <th class="theme-preview__table-col px-6 py-5 text-center">
                    分类
                  </th>
                  <th class="theme-preview__table-col px-6 py-5 text-center">
                    难度
                  </th>
                  <th class="theme-preview__table-col theme-preview__table-points-col px-6 py-5 text-center">
                    分值
                  </th>
                  <th class="theme-preview__table-col px-8 py-5 text-right">
                    操作
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="challenge in challenges"
                  :key="challenge.id"
                  class="theme-preview__table-row"
                >
                  <td class="px-8 py-5">
                    <div class="flex flex-col gap-1">
                      <span class="theme-preview__challenge-title truncate">{{ challenge.title }}</span>
                      <span class="theme-preview__challenge-id">{{ challenge.uuid }}</span>
                    </div>
                  </td>
                  <td class="px-6 py-5 text-center">
                    <span class="theme-preview__category-pill">{{ challenge.category }}</span>
                  </td>
                  <td class="px-6 py-5 text-center">
                    <span class="theme-preview__difficulty-pill">{{ challenge.difficulty }}</span>
                  </td>
                  <td class="theme-preview__points px-6 py-5 text-center">
                    {{ challenge.points }}
                  </td>
                  <td class="px-8 py-5 text-right">
                    <div class="flex items-center justify-end gap-2">
                      <button class="theme-preview__icon-button theme-preview__icon-button--accent p-2">
                        <Eye class="w-4 h-4" />
                      </button>
                      <button class="theme-preview__icon-button theme-preview__icon-button--muted p-2">
                        <MoreHorizontal class="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.theme-preview {
  --preview-shell: var(--color-bg-base);
  --preview-surface: var(--color-bg-surface);
  --preview-surface-elevated: var(--color-bg-elevated);
  --preview-surface-strong: color-mix(in srgb, var(--color-bg-elevated) 84%, black);
  --preview-line: var(--color-border-default);
  --preview-line-strong: color-mix(in srgb, var(--color-border-default) 82%, var(--color-text-primary));
  --preview-text: var(--color-text-primary);
  --preview-text-secondary: var(--color-text-secondary);
  --preview-text-muted: var(--color-text-muted);
  --preview-accent: var(--color-primary);
  --preview-accent-soft: color-mix(in srgb, var(--color-primary) 14%, var(--color-bg-surface));
  --preview-accent-line: color-mix(in srgb, var(--color-primary) 26%, transparent);
  --preview-accent-strong: color-mix(in srgb, var(--color-primary) 80%, black);
  background:
    radial-gradient(circle at top, color-mix(in srgb, var(--color-primary) 10%, transparent), transparent 34%),
    var(--preview-shell);
  color: var(--preview-text);
}

.theme-preview__sidebar {
  background: color-mix(in srgb, var(--preview-surface-strong) 90%, black);
  color: var(--preview-text-muted);
  border-right: 1px solid color-mix(in srgb, var(--preview-line) 65%, black);
  box-shadow: 0 1.5rem 3rem color-mix(in srgb, black 36%, transparent);
}

.theme-preview__sidebar-brand {
  border-bottom: 1px solid color-mix(in srgb, var(--preview-line) 60%, black);
}

.theme-preview__brand-mark {
  background: var(--preview-accent);
  color: var(--color-text-inverse);
}

.theme-preview__brand-text {
  color: var(--color-text-inverse);
}

.theme-preview__brand-accent {
  color: color-mix(in srgb, var(--preview-accent) 76%, white);
}

.theme-preview__nav-item,
.theme-preview__logout-row {
  border: 1px solid transparent;
  border-radius: 1rem;
}

.theme-preview__nav-item:hover {
  background: color-mix(in srgb, var(--preview-accent) 8%, transparent);
  color: var(--color-text-inverse);
}

.theme-preview__nav-item--active {
  background: color-mix(in srgb, var(--preview-accent) 14%, transparent);
  border-color: var(--preview-accent-line);
  color: color-mix(in srgb, var(--preview-accent) 76%, white);
}

.theme-preview__sidebar-footer {
  border-top: 1px solid color-mix(in srgb, var(--preview-line) 60%, black);
}

.theme-preview__logout-row {
  color: var(--preview-text-muted);
}

.theme-preview__layout-bar {
  background: color-mix(in srgb, var(--preview-accent) 18%, var(--preview-surface-strong));
  border-bottom: 1px solid color-mix(in srgb, var(--preview-accent) 20%, transparent);
  color: var(--color-text-inverse);
}

.theme-preview__layout-label,
.theme-preview__layout-meta {
  font-size: var(--font-size-11);
  font-weight: 800;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.theme-preview__layout-label {
  opacity: 0.58;
}

.theme-preview__layout-meta {
  opacity: 0.72;
}

.theme-preview__layout-switcher {
  background: color-mix(in srgb, black 16%, transparent);
  border-radius: 0.75rem;
}

.theme-preview__layout-option {
  border: 1px solid transparent;
  border-radius: 0.5rem;
  color: color-mix(in srgb, var(--color-text-inverse) 90%, transparent);
  font-size: var(--font-size-11);
  font-weight: 700;
}

.theme-preview__layout-option:hover {
  background: color-mix(in srgb, var(--color-text-inverse) 10%, transparent);
}

.theme-preview__layout-option--active {
  background: var(--preview-accent);
  color: var(--color-text-inverse);
  box-shadow: 0 0.75rem 1.5rem color-mix(in srgb, var(--preview-accent) 20%, transparent);
}

.theme-preview__page-header {
  background: color-mix(in srgb, var(--preview-surface) 94%, var(--preview-shell));
  border-bottom: 1px solid var(--preview-line);
  backdrop-filter: blur(14px);
}

.theme-preview__icon-button {
  border: 1px solid transparent;
  border-radius: 0.9rem;
  transition: background-color 0.2s ease, border-color 0.2s ease, color 0.2s ease;
}

.theme-preview__icon-button--muted {
  color: var(--preview-text-muted);
}

.theme-preview__icon-button--muted:hover {
  background: color-mix(in srgb, var(--preview-accent) 8%, var(--preview-surface));
  border-color: color-mix(in srgb, var(--preview-accent) 16%, transparent);
  color: var(--preview-text);
}

.theme-preview__icon-button--accent {
  color: var(--preview-text-muted);
}

.theme-preview__icon-button--accent:hover {
  background: color-mix(in srgb, var(--preview-accent) 12%, var(--preview-surface));
  border-color: color-mix(in srgb, var(--preview-accent) 24%, transparent);
  color: var(--preview-accent);
}

.theme-preview__breadcrumb {
  color: var(--preview-text-muted);
  font-weight: 600;
}

.theme-preview__breadcrumb-separator {
  color: color-mix(in srgb, var(--preview-line) 86%, transparent);
}

.theme-preview__page-title {
  color: var(--preview-text);
  font-size: 1.125rem;
  font-weight: 800;
  letter-spacing: -0.02em;
}

.theme-preview__alert-dot {
  background: var(--color-danger);
  border: 2px solid var(--preview-surface);
}

.theme-preview__avatar-shell {
  background: color-mix(in srgb, var(--preview-accent) 14%, var(--preview-surface));
  border: 1px solid color-mix(in srgb, var(--preview-accent) 20%, transparent);
  box-shadow: 0 0.75rem 1.5rem color-mix(in srgb, black 12%, transparent);
}

.theme-preview__header-tab {
  border-bottom: 2px solid transparent;
  color: var(--preview-text-muted);
  font-size: var(--font-size-12);
  font-weight: 900;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.theme-preview__header-tab:hover {
  color: var(--preview-text-secondary);
}

.theme-preview__header-tab--active {
  border-bottom-color: var(--preview-accent);
  color: var(--preview-accent);
}

.theme-preview__content {
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--preview-surface-elevated) 64%, transparent),
    transparent 28%
  ), var(--preview-shell);
}

.theme-preview__content-tabs {
  background: color-mix(in srgb, var(--preview-surface-elevated) 88%, var(--preview-shell));
  border: 1px solid var(--preview-line);
  border-radius: 1rem;
}

.theme-preview__content-tab {
  border: 1px solid transparent;
  border-radius: 0.85rem;
  color: var(--preview-text-muted);
  font-size: var(--font-size-12);
  font-weight: 700;
}

.theme-preview__content-tab:hover {
  color: var(--preview-text-secondary);
}

.theme-preview__content-tab--active {
  background: color-mix(in srgb, var(--preview-surface) 96%, transparent);
  border-color: color-mix(in srgb, var(--preview-accent) 18%, transparent);
  color: var(--preview-accent);
  box-shadow: 0 0.75rem 1.5rem color-mix(in srgb, black 10%, transparent);
}

.theme-preview__body {
  max-width: 87.5rem;
}

.theme-preview__stat-card {
  background: color-mix(in srgb, var(--preview-surface) 96%, transparent);
  border: 1px solid var(--preview-line);
  border-radius: 1.25rem;
  box-shadow: 0 1rem 2.5rem color-mix(in srgb, black 10%, transparent);
  transition: border-color 0.2s ease, transform 0.2s ease, box-shadow 0.2s ease;
}

.theme-preview__stat-card:hover {
  border-color: color-mix(in srgb, var(--preview-accent) 34%, transparent);
  transform: translateY(-0.125rem);
  box-shadow: 0 1.25rem 2.75rem color-mix(in srgb, black 14%, transparent);
}

.theme-preview__stat-head {
  color: var(--preview-text-muted);
  font-size: var(--font-size-11);
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  transition: color 0.2s ease;
}

.theme-preview__stat-card:hover .theme-preview__stat-head {
  color: var(--preview-accent);
}

.theme-preview__stat-value {
  color: var(--preview-text);
  font-family: var(--font-family-mono);
  font-size: clamp(2rem, 4vw, 2.5rem);
  font-style: normal;
  font-weight: 900;
  letter-spacing: -0.05em;
  line-height: 1;
  margin: 0;
}

.theme-preview__stat-helper {
  color: var(--preview-text-muted);
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.08em;
  margin-bottom: 0;
}

.theme-preview__stat-trend {
  background: color-mix(in srgb, var(--preview-surface) 94%, transparent);
  border: 1px solid var(--preview-line);
  border-radius: 0.75rem;
  box-shadow: 0 0.5rem 1rem color-mix(in srgb, black 8%, transparent);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-11);
  font-weight: 900;
  padding: 0.35rem 0.7rem;
  text-align: right;
}

.theme-preview__stat-trend--primary {
  color: var(--preview-accent);
}

.theme-preview__stat-trend--success {
  color: var(--color-success);
}

.theme-preview__stat-trend--warning {
  color: var(--color-warning);
}

.theme-preview__stat-trend--danger {
  color: var(--color-danger);
}

.theme-preview__search-icon {
  color: var(--preview-text-muted);
}

.theme-preview__search-input {
  background: color-mix(in srgb, var(--preview-surface) 96%, transparent);
  border: 1px solid var(--preview-line);
  border-radius: 1rem;
  color: var(--preview-text);
  font-size: var(--font-size-14);
  outline: none;
  transition: border-color 0.2s ease, box-shadow 0.2s ease, background-color 0.2s ease;
}

.theme-preview__search-input::placeholder {
  color: var(--preview-text-muted);
}

.theme-preview__search-input:focus {
  background: var(--preview-surface);
  border-color: color-mix(in srgb, var(--preview-accent) 36%, transparent);
  box-shadow: 0 0 0 0.25rem color-mix(in srgb, var(--preview-accent) 8%, transparent);
}

.theme-preview__secondary-button,
.theme-preview__primary-button {
  border-radius: 0.9rem;
  font-size: var(--font-size-12);
  font-weight: 800;
  transition: background-color 0.2s ease, border-color 0.2s ease, color 0.2s ease, transform 0.2s ease;
}

.theme-preview__secondary-button {
  background: color-mix(in srgb, var(--preview-surface) 96%, transparent);
  border: 1px solid var(--preview-line);
  color: var(--preview-text-secondary);
}

.theme-preview__secondary-button:hover {
  background: color-mix(in srgb, var(--preview-accent) 8%, var(--preview-surface));
  border-color: color-mix(in srgb, var(--preview-accent) 18%, transparent);
}

.theme-preview__primary-button {
  background: var(--preview-accent);
  border: 1px solid transparent;
  box-shadow: 0 1rem 2rem color-mix(in srgb, var(--preview-accent) 18%, transparent);
  color: var(--color-text-inverse);
}

.theme-preview__primary-button:hover {
  background: var(--preview-accent-strong);
}

.theme-preview__primary-button:active {
  transform: scale(0.98);
}

.theme-preview__table-shell {
  background: color-mix(in srgb, var(--preview-surface) 96%, transparent);
  border: 1px solid var(--preview-line);
  border-radius: 1.5rem;
  box-shadow: 0 1rem 2.5rem color-mix(in srgb, black 12%, transparent);
}

.theme-preview__table-head {
  background: color-mix(in srgb, var(--preview-surface-elevated) 92%, transparent);
  border-bottom: 1px solid var(--preview-line);
  color: var(--preview-text-muted);
}

.theme-preview__table-col,
.theme-preview__table-title-col {
  font-size: var(--font-size-11);
  font-weight: 900;
  letter-spacing: 0.2em;
  text-transform: uppercase;
}

.theme-preview__table-title-col {
  width: 40%;
}

.theme-preview__table-points-col {
  width: 6rem;
}

.theme-preview__table-row {
  border-bottom: 1px solid color-mix(in srgb, var(--preview-line) 68%, transparent);
  transition: background-color 0.2s ease;
}

.theme-preview__table-row:last-child {
  border-bottom: none;
}

.theme-preview__table-row:hover {
  background: color-mix(in srgb, var(--preview-accent) 5%, var(--preview-surface));
}

.theme-preview__challenge-title {
  color: var(--preview-text);
  font-weight: 800;
  transition: color 0.2s ease;
}

.theme-preview__table-row:hover .theme-preview__challenge-title {
  color: var(--preview-accent);
}

.theme-preview__challenge-id {
  color: var(--preview-text-muted);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-11);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.theme-preview__category-pill {
  background: color-mix(in srgb, var(--preview-accent) 12%, var(--preview-surface));
  border: 1px solid color-mix(in srgb, var(--preview-accent) 20%, transparent);
  border-radius: 0.75rem;
  color: var(--preview-accent);
  display: inline-flex;
  font-size: var(--font-size-11);
  font-weight: 900;
  justify-content: center;
  min-width: 3.75rem;
  padding: 0.25rem 0.75rem;
  text-transform: uppercase;
}

.theme-preview__difficulty-pill {
  color: var(--preview-text-secondary);
  font-size: var(--font-size-11);
  font-weight: 700;
  text-transform: uppercase;
}

.theme-preview__points {
  color: var(--preview-text);
  font-family: var(--font-family-mono);
  font-size: 1.125rem;
  font-style: italic;
  font-weight: 900;
}

.custom-scrollbar::-webkit-scrollbar {
  width: 0.375rem;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: color-mix(in srgb, var(--preview-line) 88%, transparent);
  border-radius: 999px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: color-mix(in srgb, var(--preview-line-strong) 90%, transparent);
}
</style>
