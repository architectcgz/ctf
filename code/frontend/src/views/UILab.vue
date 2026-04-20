<script setup lang="ts">
import { ref } from 'vue'
import {
  Database, Book, CheckCircle, Zap, Edit3, Search, Filter,
  ChevronDown, Eye, MoreHorizontal, FileSearch, Trash2,
  ShieldCheck, Layout, Sidebar as SidebarIcon, Rows, 
  Menu, Bell, User, Users, Settings, LogOut, ChevronRight,
  Terminal, Shield, Globe, Cpu, Activity
} from 'lucide-vue-next'

const currentLayout = ref('layout2') // 默认展示方案2 (对标 a.html 优化版)

// 数据模拟
const stats = [
  { label: '题目总量', value: 256, icon: Book, trend: '+12%', colorClass: 'text-indigo-600', bgClass: 'bg-indigo-50' },
  { label: '已发布', value: 184, icon: CheckCircle, trend: '72%', colorClass: 'text-emerald-600', bgClass: 'bg-emerald-50' },
  { label: '运行环境', value: 42, icon: Zap, trend: '正常', colorClass: 'text-amber-600', bgClass: 'bg-amber-50' },
  { label: '待处理', value: 12, icon: Edit3, trend: '需审核', colorClass: 'text-rose-600', bgClass: 'bg-rose-50' },
]

const challenges = [
  { id: 1, uuid: 'WEB-SSR-01', title: '内部笔记下载器：服务端请求伪造漏洞', category: 'Web', difficulty: '简单', points: 100, status: '已发布' },
  { id: 2, uuid: 'PWN-HEAP-05', title: '堆溢出利用：Tcache Poisoning 核心原理', category: 'Pwn', difficulty: '困难', points: 850, status: '已发布' },
  { id: 3, uuid: 'MISC-TRAF-09', title: '流量审计：异常协议识别与数据隐写提取', category: 'Misc', difficulty: '中等', points: 300, status: '草稿' },
]

const sidebarCollapsed = ref(false)
const activeMenu = ref('题目管理')
</script>

<template>
  <div class="ui-lab-root h-screen w-screen bg-[#f8fafc] text-slate-900 font-sans antialiased overflow-hidden flex flex-col">
    <!-- 顶部 Lab 控制条 (不属于 UI 方案的一部分) -->
    <div class="h-12 bg-slate-900 text-white flex items-center justify-between px-6 shrink-0 z-[100]">
      <div class="flex items-center gap-4">
        <span class="text-xs font-black tracking-[0.3em] text-indigo-400">UI_DESIGN_LAB</span>
        <div class="h-4 w-px bg-slate-700" />
        <div class="flex gap-1 p-1 bg-white/5 rounded-lg">
          <button
            :class="['px-3 py-1 text-[10px] font-bold rounded-md transition-all', currentLayout === 'layout1' ? 'bg-indigo-600' : 'hover:bg-white/10']"
            @click="currentLayout = 'layout1'"
          >
            1. 极简双轨
          </button>
          <button
            :class="['px-3 py-1 text-[10px] font-bold rounded-md transition-all', currentLayout === 'layout2' ? 'bg-indigo-600' : 'hover:bg-white/10']"
            @click="currentLayout = 'layout2'"
          >
            2. 专业工作台 (推荐)
          </button>
          <button
            :class="['px-3 py-1 text-[10px] font-bold rounded-md transition-all', currentLayout === 'layout3' ? 'bg-indigo-600' : 'hover:bg-white/10']"
            @click="currentLayout = 'layout3'"
          >
            3. 沉浸式画布
          </button>
        </div>
      </div>
      <div class="text-[10px] font-mono text-slate-500">
        EXPERIMENTAL INTERFACE / V2.0
      </div>
    </div>

    <!-- 方案展示区域 -->
    <div class="flex-1 relative overflow-hidden flex">
      <!-- ============================================================
           方案 1：极简双轨式 (The Dual Rail)
           特点：极窄侧边栏，强调空间释放
           ============================================================ -->
      <template v-if="currentLayout === 'layout1'">
        <aside class="w-16 bg-white border-r border-slate-200 flex flex-col items-center py-6 gap-8 z-50 shadow-sm">
          <div class="w-10 h-10 bg-indigo-600 rounded-2xl flex items-center justify-center text-white shadow-lg shadow-indigo-100">
            <Shield class="w-6 h-6" />
          </div>
          <nav class="flex flex-col gap-4">
            <button
              v-for="(icon, index) in [Layout, FileSearch, ShieldCheck, Users, Settings]"
              :key="`layout1-nav-${index}`"
              class="p-3 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-xl transition-all"
            >
              <component
                :is="icon"
                class="w-5 h-5"
              />
            </button>
          </nav>
        </aside>
        <main class="flex-1 flex flex-col min-w-0 bg-white">
          <header class="h-16 px-8 flex items-center justify-between border-b border-slate-100">
            <h2 class="font-black text-xl tracking-tight text-slate-800">
              题目资源管理
            </h2>
            <div class="flex items-center gap-4">
              <Search class="w-5 h-5 text-slate-300" />
              <div class="w-8 h-8 rounded-full bg-slate-100 border border-slate-200" />
            </div>
          </header>
          <div class="flex-1 p-8 overflow-y-auto">
            <div class="max-w-5xl mx-auto">
              <div class="grid grid-cols-4 gap-4 mb-8">
                <div
                  v-for="s in stats"
                  :key="s.label"
                  class="p-4 bg-slate-50 rounded-2xl border border-slate-100"
                >
                  <div class="text-[10px] font-bold text-slate-400 mb-1">
                    {{ s.label }}
                  </div>
                  <div class="text-2xl font-black">
                    {{ s.value }}
                  </div>
                </div>
              </div>
              <div class="bg-white border border-slate-200 rounded-2xl overflow-hidden shadow-sm">
                <div class="p-4 bg-slate-50/50 border-b border-slate-100 text-[10px] font-bold tracking-widest text-slate-400 uppercase">
                  Challenge Data List
                </div>
                <div
                  v-for="c in challenges"
                  :key="c.id"
                  class="p-4 border-b border-slate-50 flex items-center justify-between last:border-0 hover:bg-slate-50 transition-colors"
                >
                  <div class="flex flex-col">
                    <span class="font-bold text-sm">{{ c.title }}</span>
                    <span class="text-[10px] font-mono text-slate-400">{{ c.uuid }}</span>
                  </div>
                  <span class="text-xs font-black text-indigo-600">{{ c.points }} pts</span>
                </div>
              </div>
            </div>
          </div>
        </main>
      </template>

      <!-- ============================================================
           方案 2：专业工作台 (The Pro-Workspace) —— 核心优化版
           特点：深色侧边栏 + 精致 Header + a.html 配色与卡片
           ============================================================ -->
      <template v-else-if="currentLayout === 'layout2'">
        <!-- 左侧导航 -->
        <aside :class="['bg-[#0f172a] text-slate-400 flex flex-col transition-all duration-300 z-50', sidebarCollapsed ? 'w-20' : 'w-64']">
          <div class="h-16 flex items-center px-6 gap-3 shrink-0">
            <div class="w-8 h-8 bg-indigo-500 rounded-lg flex items-center justify-center shrink-0 shadow-lg shadow-indigo-500/20">
              <Terminal class="text-white w-5 h-5" />
            </div>
            <span
              v-if="!sidebarCollapsed"
              class="font-bold text-white tracking-tight uppercase text-sm"
            >Challenge<span class="text-indigo-400">Vault</span></span>
          </div>
          <nav class="flex-1 px-3 py-4 space-y-1">
            <div
              v-for="menu in ['概览中心', '题目管理', '竞赛编排', '用户治理', '系统镜像', '审计日志']"
              :key="menu" 
              :class="['flex items-center gap-3 px-4 py-2.5 rounded-xl transition-all cursor-pointer group', activeMenu === menu ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-600/20' : 'hover:bg-white/5 hover:text-white']"
              @click="activeMenu = menu"
            >
              <component
                :is="[Layout, Book, ShieldCheck, Users, Globe, Activity][['概览中心', '题目管理', '竞赛编排', '用户治理', '系统镜像', '审计日志'].indexOf(menu)]"
                class="w-5 h-5 shrink-0"
              />
              <span
                v-if="!sidebarCollapsed"
                class="text-sm font-bold"
              >{{ menu }}</span>
            </div>
          </nav>
          <div class="p-4 border-t border-white/5">
            <div class="flex items-center gap-3 px-4 py-3 rounded-xl hover:bg-red-500/10 hover:text-red-400 transition-colors cursor-pointer">
              <LogOut class="w-5 h-5" />
              <span
                v-if="!sidebarCollapsed"
                class="text-sm font-bold"
              >退出登录</span>
            </div>
          </div>
        </aside>

        <!-- 主内容 -->
        <main class="flex-1 flex flex-col min-w-0 bg-[#f8fafc] overflow-hidden">
          <!-- 增强版 Header -->
          <header class="bg-white/80 backdrop-blur-xl border-b border-slate-200/60 sticky top-0 z-40 shrink-0">
            <div class="px-8 h-16 flex items-center justify-between">
              <div class="flex items-center gap-4">
                <button
                  class="p-2 hover:bg-slate-100 rounded-xl transition-colors"
                  @click="sidebarCollapsed = !sidebarCollapsed"
                >
                  <SidebarIcon class="w-5 h-5 text-slate-400" />
                </button>
                <div class="flex items-center gap-3 text-sm">
                  <span class="text-slate-400 font-bold uppercase tracking-widest text-[10px]">Resources</span>
                  <ChevronRight class="w-4 h-4 text-slate-300" />
                  <span class="font-black text-slate-900 text-lg tracking-tight">题目资源库</span>
                </div>
              </div>
              <div class="flex items-center gap-4">
                <div class="flex items-center gap-2 px-3 py-1.5 bg-slate-50 border border-slate-100 rounded-xl">
                  <span class="w-2 h-2 bg-emerald-500 rounded-full animate-pulse" />
                  <span class="text-[10px] font-black text-slate-500 uppercase tracking-widest">Server_Primary</span>
                </div>
                <div class="w-10 h-10 rounded-2xl bg-indigo-50 border border-indigo-100 flex items-center justify-center overflow-hidden shadow-sm p-0.5">
                  <img
                    src="https://api.dicebear.com/7.x/avataaars/svg?seed=Felix"
                    class="rounded-xl"
                    alt="avatar"
                  >
                </div>
              </div>
            </div>
            <!-- 吸顶二级导航 -->
            <div class="px-10 flex gap-10">
              <button
                v-for="t in ['题库管理', '导入中心', '任务队列']"
                :key="t"
                class="pb-3 text-[11px] font-black tracking-[0.2em] uppercase transition-all border-b-2 mt-2"
                :class="t === '题库管理' ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-slate-400 hover:text-slate-600'"
              >
                {{ t }}
              </button>
            </div>
          </header>

          <!-- 滚动区域 -->
          <div class="flex-1 overflow-y-auto p-10 custom-scrollbar">
            <div class="max-w-[1500px] mx-auto">
              <!-- a.html 风格统计卡片 -->
              <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-12">
                <div
                  v-for="s in stats"
                  :key="s.label"
                  class="bg-white border border-slate-200/80 rounded-[24px] p-6 shadow-sm hover:border-indigo-400 hover:shadow-xl hover:shadow-indigo-500/5 transition-all group relative overflow-hidden"
                >
                  <div :class="['absolute top-0 right-0 w-20 h-20 rounded-bl-full -mr-10 -mt-10 opacity-40 transition-colors group-hover:opacity-100', s.bgClass]" />
                  <div class="flex flex-col h-full relative z-10">
                    <div class="flex justify-between items-start mb-8 text-slate-400 group-hover:text-indigo-600 transition-colors font-black uppercase text-[10px] tracking-[0.15em]">
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
                        <p class="text-[10px] font-bold text-slate-400 mt-4 tracking-wide uppercase">
                          核心运行指标
                        </p>
                      </div>
                      <div :class="['text-right font-black text-[10px] px-3 py-1 rounded-lg border shadow-sm font-mono', s.colorClass, 'bg-white border-slate-100']">
                        {{ s.trend }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- a.html 风格数据表格 -->
              <div class="bg-white border border-slate-200/80 rounded-[32px] shadow-sm overflow-hidden border-separate">
                <div class="px-8 py-5 border-b border-slate-100 bg-slate-50/30 flex justify-between items-center">
                  <div class="flex items-center gap-4">
                    <div class="relative">
                      <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-400" />
                      <input
                        type="text"
                        placeholder="快速检索..."
                        class="bg-white border border-slate-200 pl-9 pr-4 py-2 rounded-xl text-xs w-64 outline-none focus:border-indigo-500 transition-all shadow-sm"
                      >
                    </div>
                    <button class="text-xs font-bold text-slate-500 hover:text-indigo-600 flex items-center gap-1.5">
                      <Filter class="w-3.5 h-3.5" /> 展开筛选
                    </button>
                  </div>
                  <button class="bg-indigo-600 text-white px-5 py-2 rounded-xl text-xs font-black shadow-lg shadow-indigo-200 hover:bg-indigo-700 transition-all">
                    + 新建题目
                  </button>
                </div>
                <table class="w-full text-left border-collapse table-fixed">
                  <thead>
                    <tr class="text-slate-400">
                      <th class="px-8 py-5 text-[10px] font-black uppercase tracking-[0.2em] w-[45%]">
                        题目资源名称 / 唯一标识
                      </th>
                      <th class="px-6 py-5 text-[10px] font-black uppercase tracking-[0.2em] text-center">
                        分类
                      </th>
                      <th class="px-6 py-5 text-[10px] font-black uppercase tracking-[0.2em] text-center">
                        难度层级
                      </th>
                      <th class="px-6 py-5 text-[10px] font-black uppercase tracking-[0.2em] text-center w-32">
                        当前分值
                      </th>
                      <th class="px-8 py-5 text-[10px] font-black uppercase tracking-[0.2em] text-right">
                        管理操作
                      </th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-50">
                    <tr
                      v-for="c in challenges"
                      :key="c.id"
                      class="group hover:bg-indigo-50/30 transition-all"
                    >
                      <td class="px-8 py-6">
                        <div class="flex flex-col gap-1.5">
                          <span
                            class="font-bold text-slate-900 group-hover:text-indigo-600 transition-colors truncate text-sm"
                            :title="c.title"
                          >{{ c.title }}</span>
                          <span class="text-[10px] font-mono text-slate-400 uppercase tracking-wider bg-slate-100 w-fit px-1.5 py-0.5 rounded">{{ c.uuid }}</span>
                        </div>
                      </td>
                      <td class="px-6 py-6 text-center">
                        <span class="text-[10px] font-black uppercase px-3 py-1 bg-white text-indigo-600 rounded-lg border border-indigo-100 shadow-sm">{{ c.category }}</span>
                      </td>
                      <td class="px-6 py-6 text-center">
                        <div class="flex items-center justify-center gap-2">
                          <span
                            class="w-1.5 h-1.5 rounded-full"
                            :class="c.difficulty === '困难' ? 'bg-rose-500' : 'bg-emerald-500'"
                          />
                          <span class="text-[11px] font-bold text-slate-600">{{ c.difficulty }}</span>
                        </div>
                      </td>
                      <td class="px-6 py-6 text-center font-black text-slate-900 font-mono text-xl italic tracking-tighter">
                        {{ c.points }}
                      </td>
                      <td class="px-8 py-6 text-right">
                        <div class="flex items-center justify-end gap-1">
                          <button class="p-2.5 text-slate-400 hover:text-indigo-600 hover:bg-white hover:shadow-sm rounded-xl transition-all">
                            <Eye class="w-4.5 h-4.5" />
                          </button>
                          <button class="p-2.5 text-slate-400 hover:text-rose-600 hover:bg-white hover:shadow-sm rounded-xl transition-all">
                            <Trash2 class="w-4.5 h-4.5" />
                          </button>
                          <button class="p-2.5 text-slate-400 hover:text-slate-900 hover:bg-white hover:shadow-sm rounded-xl transition-all">
                            <MoreHorizontal class="w-4.5 h-4.5" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </main>
      </template>

      <!-- ============================================================
           方案 3：沉浸式画布布局 (The Immersive Canvas)
           特点：无侧边栏，全屏工作流，不对称布局
           ============================================================ -->
      <template v-else-if="currentLayout === 'layout3'">
        <main class="flex-1 bg-slate-50 p-12 overflow-y-auto">
          <div class="max-w-6xl mx-auto flex flex-col gap-12">
            <div class="flex justify-between items-end">
              <div>
                <div class="flex items-center gap-3 mb-4">
                  <div class="w-12 h-12 bg-black rounded-full flex items-center justify-center text-white">
                    <Cpu class="w-6 h-6" />
                  </div>
                  <h1 class="text-5xl font-black italic tracking-tighter uppercase">
                    Studio_Ops
                  </h1>
                </div>
                <p class="text-slate-400 font-bold tracking-[0.4em] text-xs">
                  IMMERSIVE CHALLENGE MANAGEMENT
                </p>
              </div>
              <nav class="flex gap-12 border-b border-slate-200">
                <button
                  v-for="m in ['题目', '镜像', '审计']"
                  :key="m"
                  class="pb-4 text-sm font-black uppercase tracking-widest"
                  :class="m === '题目' ? 'border-b-4 border-black text-black' : 'text-slate-300'"
                >
                  {{ m }}
                </button>
              </nav>
            </div>

            <div class="grid grid-cols-12 gap-8">
              <div class="col-span-8 space-y-8">
                <div class="bg-white p-8 rounded-[40px] shadow-2xl shadow-slate-200/50">
                  <h3 class="text-xl font-black mb-8 flex items-center gap-2">
                    <Rows class="w-5 h-5" /> 活跃任务清单
                  </h3>
                  <div class="space-y-4">
                    <div
                      v-for="c in challenges"
                      :key="c.id"
                      class="p-6 bg-slate-50 rounded-3xl flex items-center justify-between group hover:bg-black hover:text-white transition-all duration-500"
                    >
                      <div class="flex items-center gap-6">
                        <div class="text-4xl font-black opacity-10 group-hover:opacity-100">
                          0{{ c.id }}
                        </div>
                        <div class="flex flex-col">
                          <span class="font-bold text-lg">{{ c.title }}</span>
                          <span class="text-xs opacity-50">{{ c.uuid }}</span>
                        </div>
                      </div>
                      <div class="flex items-center gap-8">
                        <span class="px-4 py-1 border border-current rounded-full text-[10px] font-black tracking-widest">{{ c.category }}</span>
                        <ChevronRight class="w-6 h-6 opacity-0 group-hover:opacity-100 -translate-x-4 group-hover:translate-x-0 transition-all" />
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              <div class="col-span-4 space-y-8">
                <div class="bg-indigo-600 p-8 rounded-[40px] text-white shadow-xl shadow-indigo-200">
                  <div class="text-[10px] font-black tracking-[0.2em] mb-8 opacity-60">
                    SYSTEM STATUS
                  </div>
                  <div class="space-y-6">
                    <div
                      v-for="s in stats.slice(0, 3)"
                      :key="s.label"
                      class="flex justify-between items-end border-b border-white/10 pb-4"
                    >
                      <span class="text-sm font-bold opacity-80">{{ s.label }}</span>
                      <span class="text-3xl font-black font-mono">{{ s.value }}</span>
                    </div>
                  </div>
                </div>
                <div class="bg-white p-8 rounded-[40px] border border-slate-200">
                  <button class="w-full bg-slate-900 text-white py-4 rounded-2xl font-black tracking-widest hover:bg-indigo-600 transition-colors uppercase text-xs">
                    Deploy New Challenge
                  </button>
                </div>
              </div>
            </div>
          </div>
        </main>
      </template>
    </div>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar { width: 6px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: #e2e8f0; border-radius: 10px; }
.custom-scrollbar::-webkit-scrollbar-thumb:hover { background: #cbd5e0; }

.ui-lab-root {
  background-color: #f8fafc;
}
</style>
