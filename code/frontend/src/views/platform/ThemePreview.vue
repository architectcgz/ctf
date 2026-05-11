<script setup lang="ts">
import { ref } from 'vue'
import {
  Database, Book, CheckCircle, Zap, Edit3, Search, Filter,
  ChevronDown, Eye, MoreHorizontal, FileSearch, Trash2,
  ShieldCheck, Layout, Sidebar as SidebarIcon, Rows, Users,
  Menu, Bell, User, Settings, LogOut, ChevronRight
} from 'lucide-vue-next'
import { ChallengeCategoryPill, ChallengeDifficultyText } from '@/entities/challenge'

const currentLayout = ref('variant2') // 默认展示方案2

const stats = [
  { label: '题目总量', value: 256, icon: Book, trend: '+12%', colorClass: 'text-indigo-500' },
  { label: '已发布', value: 184, icon: CheckCircle, trend: '72%', colorClass: 'text-emerald-500' },
  { label: '运行环境', value: 42, icon: Zap, trend: '正常', colorClass: 'text-amber-500' },
  { label: '待处理', value: 12, icon: Edit3, trend: '需审核', colorClass: 'text-rose-500' },
]

const challenges = [
  { id: 1, uuid: 'WEB-SSR-01', title: '内部笔记下载器：服务端请求伪造', category: 'web', difficulty: 'easy', points: 100 },
  { id: 2, uuid: 'PWN-HEAP-05', title: '堆溢出：Tcache Poisoning 攻击实战', category: 'pwn', difficulty: 'hard', points: 850 },
  { id: 3, uuid: 'MISC-TRAF-09', title: '流量分析：隐藏协议识别与提取', category: 'misc', difficulty: 'medium', points: 300 },
] as const

const activeTab = ref('manage')
const sidebarCollapsed = ref(false)
</script>

<template>
  <div class="min-h-screen bg-[#f8fafc] text-slate-900 font-sans antialiased flex overflow-hidden">
    <!-- 1. 模拟侧边栏 (根据不同方案调整宽度) -->
    <aside 
      v-if="currentLayout !== 'variant3'"
      :class="[
        'bg-slate-900 text-slate-400 flex flex-col transition-all duration-300 z-50 shrink-0',
        currentLayout === 'variant2' ? (sidebarCollapsed ? 'w-20' : 'w-64') : 'w-64'
      ]"
    >
      <div class="h-16 flex items-center px-6 gap-3 border-b border-slate-800 shrink-0">
        <div class="w-8 h-8 bg-indigo-600 rounded-lg flex items-center justify-center shrink-0">
          <Database class="text-white w-4 h-4" />
        </div>
        <span
          v-if="!sidebarCollapsed || currentLayout === 'variant1'"
          class="font-bold text-white tracking-tight uppercase"
        >CTF<span class="text-indigo-400">Ops</span></span>
      </div>
      
      <nav class="flex-1 p-4 space-y-2">
        <div
          v-for="i in 5"
          :key="i"
          :class="['flex items-center gap-3 px-4 py-3 rounded-xl transition-all cursor-pointer hover:bg-slate-800 hover:text-white', i === 1 ? 'bg-indigo-600/10 text-indigo-400 border border-indigo-600/20' : '']"
        >
          <component
            :is="[Layout, FileSearch, ShieldCheck, Users, Settings][i-1]"
            class="w-5 h-5"
          />
          <span
            v-if="!sidebarCollapsed || currentLayout === 'variant1'"
            class="text-sm font-bold"
          >{{ ['概览中心', '题目管理', '竞赛编排', '用户治理', '系统设置'][i-1] }}</span>
        </div>
      </nav>

      <div class="p-4 border-t border-slate-800">
        <div class="flex items-center gap-3 px-4 py-3 text-slate-500">
          <LogOut class="w-5 h-5" />
          <span
            v-if="!sidebarCollapsed || currentLayout === 'variant1'"
            class="text-sm font-bold"
          >退出系统</span>
        </div>
      </div>
    </aside>

    <!-- 2. 主内容区区域 -->
    <div class="flex-1 flex flex-col min-w-0 h-screen overflow-hidden">
      <!-- 顶部控制条 (预览专用切换) -->
      <header class="bg-indigo-900 text-white px-6 py-2 flex justify-between items-center z-[60] shrink-0">
        <div class="flex items-center gap-4">
          <span class="text-[10px] font-black tracking-widest opacity-50 uppercase">Layout Switcher</span>
          <div class="flex bg-black/20 p-1 rounded-lg">
            <button
              :class="['px-3 py-1 text-[10px] font-bold rounded-md transition-all', currentLayout === 'variant1' ? 'bg-indigo-600 shadow-lg' : 'hover:bg-white/10']"
              @click="currentLayout = 'variant1'"
            >
              方案1：导航融合
            </button>
            <button
              :class="['px-3 py-1 text-[10px] font-bold rounded-md transition-all', currentLayout === 'variant2' ? 'bg-indigo-600 shadow-lg' : 'hover:bg-white/10']"
              @click="currentLayout = 'variant2'"
            >
              方案2：分层指挥
            </button>
            <button
              :class="['px-3 py-1 text-[10px] font-bold rounded-md transition-all', currentLayout === 'variant3' ? 'bg-indigo-600 shadow-lg' : 'hover:bg-white/10']"
              @click="currentLayout = 'variant3'"
            >
              方案3：沉浸画布
            </button>
          </div>
        </div>
        <div class="text-[10px] font-mono opacity-50">
          PREVIEW MODE / {{ currentLayout.toUpperCase() }}
        </div>
      </header>

      <!-- 真正的页面 Header -->
      <header class="bg-white border-b border-slate-200 shrink-0">
        <div class="px-8 h-16 flex items-center justify-between">
          <div class="flex items-center gap-4">
            <button
              v-if="currentLayout === 'variant2'"
              class="p-2 hover:bg-slate-100 rounded-lg transition-colors"
              @click="sidebarCollapsed = !sidebarCollapsed"
            >
              <SidebarIcon class="w-5 h-5 text-slate-400" />
            </button>
            <div class="flex items-center gap-2 text-sm">
              <span class="text-slate-400 font-medium">资源管理</span>
              <ChevronRight class="w-4 h-4 text-slate-300" />
              <span class="font-bold text-slate-900 text-lg tracking-tight">题目管理中心</span>
            </div>
          </div>
          <div class="flex items-center gap-4">
            <button class="p-2.5 text-slate-400 hover:bg-slate-50 rounded-xl relative">
              <Bell class="w-5 h-5" />
              <span class="absolute top-2.5 right-2.5 w-2 h-2 bg-rose-500 rounded-full border-2 border-white" />
            </button>
            <div class="w-10 h-10 rounded-2xl bg-indigo-50 border border-indigo-100 flex items-center justify-center overflow-hidden shadow-sm">
              <img
                src="https://api.dicebear.com/7.x/avataaars/svg?seed=Felix"
                alt="avatar"
              >
            </div>
          </div>
        </div>

        <!-- 方案 2 的吸顶 Tab (这是核心优化点) -->
        <div
          v-if="currentLayout === 'variant2'"
          class="px-8 flex gap-8"
        >
          <button 
            v-for="t in ['manage', 'import', 'queue']"
            :key="t"
            :class="[
              'pb-4 text-xs font-black tracking-widest uppercase transition-all border-b-2 mt-2',
              activeTab === t ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-slate-400 hover:text-slate-600'
            ]"
            @click="activeTab = t"
          >
            {{ {manage: '题库管理', import: '上传题目包', queue: '导入队列'}[t] }}
          </button>
        </div>
      </header>

      <!-- 3. 内容滚动区 -->
      <div class="flex-1 overflow-y-auto bg-[#f8fafc] p-8 custom-scrollbar">
        <!-- 如果是方案 1，Tab 放在内容顶部 -->
        <nav
          v-if="currentLayout === 'variant1'"
          class="flex gap-2 mb-8 bg-slate-200/50 p-1 rounded-2xl w-fit"
        >
          <button
            v-for="t in ['manage', 'import', 'queue']"
            :key="t"
            :class="['px-6 py-2 rounded-xl text-xs font-bold transition-all', activeTab === t ? 'bg-white text-indigo-600 shadow-sm' : 'text-slate-500 hover:text-slate-700']"
            @click="activeTab = t"
          >
            {{ {manage: '题库管理', import: '上传题目包', queue: '导入队列'}[t] }}
          </button>
        </nav>

        <div class="max-w-[1400px] mx-auto">
          <!-- 统计卡片 (a.html 风格) -->
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-10">
            <div
              v-for="s in stats"
              :key="s.label"
              class="bg-white border border-slate-200 rounded-[20px] p-6 shadow-sm hover:border-indigo-400 transition-all group relative overflow-hidden"
            >
              <div class="flex flex-col h-full relative z-10">
                <div class="flex justify-between items-start mb-6 text-slate-400 group-hover:text-indigo-500 transition-colors font-bold uppercase text-[10px] tracking-widest">
                  <span>{{ s.label }}</span>
                  <component
                    :is="s.icon"
                    class="w-4 h-4"
                  />
                </div>
                <div class="flex items-end justify-between">
                  <div class="flex flex-col">
                    <h3 class="text-4xl font-black text-slate-900 leading-none font-mono tracking-tighter">
                      {{ s.value }}
                    </h3>
                    <p class="text-[10px] font-bold text-slate-400 mt-3 tracking-wide">
                      核心指标说明
                    </p>
                  </div>
                  <div :class="['text-right font-black text-[10px] px-2.5 py-1 rounded-lg border font-mono bg-white shadow-sm', s.colorClass]">
                    {{ s.trend }}
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 工具栏 -->
          <div class="flex flex-col md:flex-row justify-between items-center gap-4 mb-6">
            <div class="relative w-full md:w-96">
              <Search class="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400 w-4 h-4" />
              <input
                type="text"
                placeholder="检索资源..."
                class="w-full bg-white border border-slate-200 pl-11 pr-4 py-3 rounded-[18px] text-sm outline-none focus:ring-4 focus:ring-indigo-500/5 focus:border-indigo-500 transition-all"
              >
            </div>
            <div class="flex items-center gap-2">
              <button class="px-4 py-2.5 bg-white border border-slate-200 rounded-xl text-xs font-bold text-slate-600 hover:bg-slate-50 transition-all flex items-center gap-2">
                <Filter class="w-4 h-4" /> 高级筛选
              </button>
              <button class="px-6 py-2.5 bg-indigo-600 text-white rounded-xl text-xs font-bold shadow-lg shadow-indigo-100 hover:bg-indigo-700 transition-all active:scale-95">
                + 新建题目
              </button>
            </div>
          </div>

          <!-- 列表 (a.html 风格) -->
          <div class="bg-white border border-slate-200 rounded-[24px] shadow-sm overflow-hidden">
            <table class="w-full text-left border-collapse table-fixed">
              <thead>
                <tr class="bg-slate-50/50 border-b border-slate-200 text-slate-400">
                  <th class="px-8 py-5 text-[10px] font-black uppercase tracking-[0.2em] w-[40%]">
                    题目资源名称
                  </th>
                  <th class="px-6 py-5 text-[10px] font-black uppercase tracking-[0.2em] text-center">
                    分类
                  </th>
                  <th class="px-6 py-5 text-[10px] font-black uppercase tracking-[0.2em] text-center">
                    难度
                  </th>
                  <th class="px-6 py-5 text-[10px] font-black uppercase tracking-[0.2em] text-center w-24">
                    分值
                  </th>
                  <th class="px-8 py-5 text-[10px] font-black uppercase tracking-[0.2em] text-right">
                    操作
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr
                  v-for="c in challenges"
                  :key="c.id"
                  class="group hover:bg-slate-50/80 transition-colors"
                >
                  <td class="px-8 py-5">
                    <div class="flex flex-col gap-1">
                      <span class="font-bold text-slate-900 group-hover:text-indigo-600 transition-colors truncate">{{ c.title }}</span>
                      <span class="text-[10px] font-mono text-slate-400 uppercase">{{ c.uuid }}</span>
                    </div>
                  </td>
                  <td class="px-6 py-5 text-center">
                    <ChallengeCategoryPill :category="c.category" />
                  </td>
                  <td class="px-6 py-5 text-center">
                    <ChallengeDifficultyText :difficulty="c.difficulty" />
                  </td>
                  <td class="px-6 py-5 text-center font-black text-slate-900 font-mono text-lg italic">
                    {{ c.points }}
                  </td>
                  <td class="px-8 py-5 text-right">
                    <div class="flex items-center justify-end gap-2">
                      <button class="p-2 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-xl transition-all">
                        <Eye class="w-4 h-4" />
                      </button>
                      <button class="p-2 text-slate-400 hover:text-slate-900 hover:bg-slate-100 rounded-xl transition-all">
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
.custom-scrollbar::-webkit-scrollbar { width: 6px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: #e2e8f0; border-radius: 10px; }
.custom-scrollbar::-webkit-scrollbar-thumb:hover { background: #cbd5e0; }
</style>
