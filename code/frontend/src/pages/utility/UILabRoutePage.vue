<template>
  <div class="ui-lab-root h-screen w-screen bg-[#f8fafc] text-slate-900 font-sans antialiased overflow-hidden flex flex-col">
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

    <div class="flex-1 relative overflow-hidden flex">
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

      <template v-else-if="currentLayout === 'layout2'">
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

        <main class="flex-1 flex flex-col min-w-0 bg-[#f8fafc] overflow-hidden">
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
                  <div
                    class="w-full h-full rounded-xl bg-gradient-to-br from-indigo-500 via-violet-500 to-cyan-500 flex items-center justify-center text-white text-[10px] font-black tracking-[0.18em]"
                    aria-label="server avatar"
                  >
                    SP
                  </div>
                </div>
              </div>
            </div>
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

          <div class="flex-1 overflow-y-auto p-10 custom-scrollbar">
            <div class="max-w-[1500px] mx-auto">
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
                        class="w-5 h-5"
                      />
                    </div>
                    <div class="mt-auto">
                      <div class="text-4xl font-black tracking-tight text-slate-900 mb-1">
                        {{ s.value }}
                      </div>
                      <div class="text-sm font-bold text-slate-500">
                        {{ s.trend }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </main>
      </template>

      <template v-else>
        <main class="flex-1 min-w-0 bg-slate-950 text-white flex items-center justify-center">
          <div class="text-center space-y-4">
            <Cpu class="w-12 h-12 mx-auto text-indigo-400" />
            <h2 class="text-2xl font-black tracking-tight">
              沉浸式画布方案
            </h2>
            <p class="text-sm text-slate-400">
              当前只保留实验入口，详细布局仍在迭代中。
            </p>
          </div>
        </main>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import {
  Activity,
  Book,
  CheckCircle,
  Cpu,
  Edit3,
  FileSearch,
  Globe,
  Layout,
  LogOut,
  Search,
  Settings,
  Shield,
  ShieldCheck,
  Sidebar as SidebarIcon,
  Terminal,
  Users,
  Zap,
  ChevronRight,
} from 'lucide-vue-next'

const currentLayout = ref('layout2')

const stats = [
  { label: '题目总量', value: 256, icon: Book, trend: '+12%', bgClass: 'bg-indigo-50' },
  { label: '已发布', value: 184, icon: CheckCircle, trend: '72%', bgClass: 'bg-emerald-50' },
  { label: '运行环境', value: 42, icon: Zap, trend: '正常', bgClass: 'bg-amber-50' },
  { label: '待处理', value: 12, icon: Edit3, trend: '需审核', bgClass: 'bg-rose-50' },
]

const challenges = [
  { id: 1, uuid: 'WEB-SSR-01', title: '内部笔记下载器：服务端请求伪造漏洞', points: 100 },
  { id: 2, uuid: 'PWN-HEAP-05', title: '堆溢出利用：Tcache Poisoning 核心原理', points: 850 },
  { id: 3, uuid: 'MISC-TRAF-09', title: '流量审计：异常协议识别与数据隐写提取', points: 300 },
] as const

const sidebarCollapsed = ref(false)
const activeMenu = ref('题目管理')
</script>
