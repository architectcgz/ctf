<script setup lang="ts">
import { computed } from 'vue'
import { CalendarClock, CircleCheckBig, Play, Send } from 'lucide-vue-next'

import type { TimelineEvent } from '@/api/contracts'
import AppCard from '@/components/common/AppCard.vue'
import MetricCard from '@/components/common/MetricCard.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import { formatDate, formatTime } from '@/utils/format'

import { timelineSummary, timelineTypeLabel, timelineTypeTone } from './utils'

const props = defineProps<{
  timeline: TimelineEvent[]
}>()

const solveCount = computed(() => props.timeline.filter((item) => item.type === 'solve').length)
const submitCount = computed(() => props.timeline.filter((item) => item.type === 'submit').length)
const instanceCount = computed(() =>
  props.timeline.filter((item) => item.type === 'instance' || typeof item.meta?.raw_type === 'string').length,
)
const groupedTimeline = computed(() => {
  const groups = new Map<string, TimelineEvent[]>()
  props.timeline.forEach((event) => {
    const key = new Date(event.created_at).toLocaleDateString('zh-CN')
    groups.set(key, [...(groups.get(key) || []), event])
  })
  return Array.from(groups.entries()).map(([date, events]) => ({ date, events }))
})
</script>

<template>
  <div class="space-y-6">
    <section class="grid gap-4 md:grid-cols-3">
      <MetricCard label="成功解题" :value="solveCount" hint="近期成功提交记录" accent="success" />
      <MetricCard label="提交次数" :value="submitCount" hint="近期 Flag 提交动作" accent="warning" />
      <MetricCard label="实例操作" :value="instanceCount" hint="启动、销毁等实例相关动作" accent="primary" />
    </section>

    <section class="grid gap-4 xl:grid-cols-[1.12fr_0.88fr]">
      <SectionCard title="训练时间线" subtitle="按时间顺序还原最近训练过程，不再复用主页里的简版卡片。">
        <div v-if="groupedTimeline.length === 0" class="rounded-2xl border border-dashed border-border px-4 py-12 text-center text-sm text-text-secondary">
          当前还没有训练动态。
        </div>

        <div v-else class="space-y-6">
          <section v-for="group in groupedTimeline" :key="group.date" class="space-y-4">
            <div class="text-xs font-semibold uppercase tracking-[0.18em] text-text-muted">{{ group.date }}</div>
            <div class="space-y-4">
              <AppCard
                v-for="event in group.events"
                :key="event.id"
                variant="action"
                accent="primary"
              >
                <div class="grid gap-3 md:grid-cols-[auto_1fr]">
                  <div class="flex items-start gap-3">
                    <div class="mt-1 h-3 w-3 rounded-full bg-primary" />
                    <div class="text-sm font-medium text-text-primary">{{ formatTime(event.created_at) }}</div>
                  </div>
                  <div class="space-y-3">
                    <div class="flex flex-wrap items-center justify-between gap-3">
                      <div>
                        <div class="text-base font-semibold text-text-primary">{{ event.title }}</div>
                        <div class="mt-1 text-sm text-text-secondary">{{ timelineSummary(event) }}</div>
                      </div>
                      <span
                        class="rounded-full border px-2.5 py-1 text-xs font-medium"
                        :class="timelineTypeTone(event)"
                      >
                        {{ timelineTypeLabel(event) }}
                      </span>
                    </div>
                    <div class="text-xs uppercase tracking-[0.16em] text-text-muted">记录时间：{{ formatDate(event.created_at) }}</div>
                  </div>
                </div>
              </AppCard>
            </div>
          </section>
        </div>
      </SectionCard>

      <div class="grid gap-4">
        <SectionCard title="节奏观察" subtitle="最近训练节奏的三个信号。">
          <div class="space-y-3">
            <AppCard variant="action" accent="success">
              <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
                <CircleCheckBig class="h-4 w-4 text-emerald-300" />
                成功信号
              </div>
              <div class="mt-2 text-sm leading-6 text-text-secondary">
                最近 {{ solveCount }} 次成功解题记录。若数量偏低，建议回到“训练建议”页选更适合当前阶段的题目。
              </div>
            </AppCard>
            <AppCard variant="action" accent="warning">
              <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
                <Send class="h-4 w-4 text-amber-300" />
                提交密度
              </div>
              <div class="mt-2 text-sm leading-6 text-text-secondary">
                最近 {{ submitCount }} 次提交动作。若提交多但成功少，说明方向可能跑偏，需要回看能力画像。
              </div>
            </AppCard>
            <AppCard variant="action" accent="primary">
              <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
                <Play class="h-4 w-4 text-sky-300" />
                实例节奏
              </div>
              <div class="mt-2 text-sm leading-6 text-text-secondary">
                最近 {{ instanceCount }} 次实例相关动作。实例操作多但提交少，通常代表分析阶段过长。
              </div>
            </AppCard>
          </div>
        </SectionCard>

        <SectionCard title="阅读方式" subtitle="时间线页只关注过程，不再混入总览卡片。">
          <AppCard variant="action" accent="primary">
            这页的结构被单独设计成“时间顺序 + 节奏解读”，目的就是让你看清训练过程，而不是像主页那样只看摘要。
          </AppCard>
          <div class="mt-3 flex items-center gap-2 text-sm text-text-primary">
            <CalendarClock class="h-4 w-4 text-primary" />
            从上到下看，就能还原最近一段时间的完整训练路径。
          </div>
        </SectionCard>
      </div>
    </section>
  </div>
</template>
